package daemon

import (
	"fmt"
	"git.xfyun.cn/AIaaS/xsf/server"
	"git.xfyun.cn/AIaaS/xsf/utils"
	"log"
	"math"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type SegIdManager struct {
	begin    int64
	end      int64
	used     *usedList     //存放不可用的id
	disposed *disposedList //存储由于节点死亡所导致的无效的segId，这部分id可重新使用
}

func (s *SegIdManager) getMin() int64 {
	if minTmp, minOk := s.disposed.getMin(); minOk {
		return minTmp
	}
	return atomic.AddInt64(&s.end, 1)
}

type usedList struct {
	m     map[int64]bool
	mRWMu sync.RWMutex
}

func (u *usedList) set(in int64) {
	u.mRWMu.Lock()
	u.m[in] = true
	u.mRWMu.Unlock()
}
func (u *usedList) get(in int64) bool {
	u.mRWMu.RLock()
	defer u.mRWMu.RUnlock()
	return u.m[in]
}

type disposedList struct {
	l     []int64 //存放
	lRwMu sync.RWMutex
}

func (d *disposedList) set(in int64) {
	d.lRwMu.Lock()
	d.l = append(d.l, in)
	d.lRwMu.Unlock()
}
func (d *disposedList) getMin() (int64, bool) {
	minIx := 0
	minVal := int64(math.MaxInt64)
	d.lRwMu.RLock()
	for ix, val := range d.l {
		if minVal > val {
			minVal = val
			minIx = ix
		}
	}
	d.lRwMu.RUnlock()
	if minVal != int64(math.MaxInt64) {
		d.lRwMu.Lock()
		d.l = append(d.l[:minIx], d.l[minIx+1:]...)
		d.lRwMu.Unlock()
		return minVal, true
	}
	return 0, false
}

//消费rmq的时间间隔
const rmqInterval = time.Second

var segIdManagerInst *SegIdManager

func init() {
	segIdManagerInst = &SegIdManager{begin: 0, used: &usedList{m: make(map[int64]bool)}, disposed: &disposedList{}}
}

type SubSvcExItem struct {
	cas       int64 //事务性 0：可修改 1：修改中
	dead      int64 //0：alive 1：dead
	addr      string
	ttl       int64 //纳秒,节点的生命周期UnixNano
	timestamp int64 //纳秒,节点最近一次上报的时间
	bestInst  int64
	idleInst  int64
	totalInst int64

	segId int64 //当前节点对应的分段id
}

type SubSvcEx struct {
	ttl             int64                    //纳秒,定时清除无效节点的时间间隔
	subSvcMapExRwMu sync.RWMutex             //读写锁，保证subSvcMap、subSvcSlice的并发安全
	subSvcMapEx     map[string]*SubSvcExItem //K:addr
	//subSvcSlice   SubSvcItemSlice 分段模式下，uid与ats为 n:1 模型，此切片即无用了

	/*
		1、通过此字典查询seg_id对应的addr
		2、通过addr查询具体的节点信息
		3、对节点信息进行处理，然后决定是否返回此节点
		4、如此循环，一直到获得nBest熟练的节点后返回或报错
	*/
	seed          int64             //uid分段的种子，如10000
	segIdAddr     map[string]string //k:seg_id,v:server_ip  用于维护seg_id与server_ip的对应关系
	segIdAddrRwMu sync.RWMutex

	threshold int64 //阈值（百分数，如20代表20%），一旦服务节点超过这个值则不再为该节点分配请求
}
type LicEx struct {
	svc          string               //大业务名，如iat
	svcMapEx     map[string]*SubSvcEx //K:type（如sms、sms-5s）
	svcMapExRwMu sync.RWMutex
	ticker       *time.Ticker
	toolbox      *xsf.ToolBox

	rmqTopic string
	rmqGroup string
}

/*
a、节点第一次上报
	1、在subSvcMap中新增加addr key，并为此节点选取合适的segId
		1.1、segId为从零递增的自然数序列（从SegIdManager实例中获取）
b、节点授权更新
	1、通过key addr索引到相关的节点数据，进行递增或递减操作
c、节点接受服务
	1、将uid转换为segId（uid对分段基数求商）
	2、通过segId索引到节点，检查授权是否过载，如未过载直接返回，否则选取一临时节点
d、通知节点同步个性化数据
	1、从rmq收取消息，然后通知ats去更新个性化数据
e、初始化从MySQL拉取分段表数据
f、新增加uid时同步数据到MySQL
g、消费rmq消息，通知ats拉取最新的个性化数据
h、单台ats的阈值(80%)达到上限，需要临时迁移到其他ats
i、uid请求服务时，如果求余后的值超过了节点的数量，则选取负载最小的几点
j、定时检查节点的有效性
*/

//定时拉取rmq消息，通知ats更新
//消息示例：svc_name=athena,common.uid=d7987022916,common.appid=59ee9f4a,common.sid=psn001b0dc8@dx00070e90f242a11201,mc.arm.ent=automotiveknife,res.iat.files=f1,res.iat.f1.table=athena_psr_bin,res.iat.f1.family=r,res.iat.f1.rowkey=IFLYTEK@uid@CESHI1.d7987022916@Hbase@personal@IFLYTEK.app_namelist@1260@text,res.iat.f1.rowkey.hash=true,res.iat.f1.qualifier=d,res.iat.f1.cmd_type=text,res.iat.f1.group=vi
func (l *LicEx) notify() {
	ticker := time.NewTicker(rmqInterval)
	parseRmqMsg := func(in string) (rst map[string]string) {
		if nil == rst {
			rst = make(map[string]string)
		}
		cutByComma := strings.Split(in, ",")
		for _, v := range cutByComma {
			tmp := strings.Split(v, "=")
			rst[tmp[0]] = tmp[1]
		}
		return
	}
	notice := func(in *MTRMessage) {
		msgMap := parseRmqMsg(string(in.GetBody()))
		rmqUid, rmqUidErr := strconv.Atoi(msgMap[RmqUid])
		if rmqUidErr != nil {
			l.toolbox.Log.Errorf("uid from rmq msg is illegal rmqUid:%v", in)
			return
		}
		nBestNodes, nBestNodesErr := l.get(withGetUid(int64(rmqUid)), withGetAll(false), withGetNBest(1), withGetSubSvc(msgMap[rmqSvcName]), withGetSubSvc(msgMap[rmqSubSvcName]))
		if nBestNodesErr != nil {
			l.toolbox.Log.Errorf("can't take nBestNodes -> withGetUid:%v, withGetAll:%v, withGetNBest:%v, withGetSubSvc:%V, withGetSubSvc:%v", rmqUid, false, 1, msgMap[rmqSvcName], msgMap[rmqSubSvcName])
			return
		}
		for _, nBestNodesItem := range nBestNodes {
			s, errcode, e := dcInst.NoticeClient(nBestNodesItem, notifyOp, msgMap[rmqSvcName], msgMap[rmqSubSvcName], msgMap[RmqUid])
			if errcode != 0 || e != nil {
				l.toolbox.Log.Errorf("dcInst.NoticeClient -> s:%s,errcode:%v,e:%v,nBestNodesItem:%v,notifyOp:%v,svc:%v,subsvc:%v,uid:%v",
					s, errcode, e, nBestNodes, notifyOp, msgMap[rmqSvcName], msgMap[rmqSubSvcName], msgMap[RmqUid])
			}
		}
	}
	for {
		select {
		//noinspection ALL,GoFunctionCall
		case <-ticker.C:
			{
				rmqMsg, rmqMsgErr := RmqManagerInst.Consume(l.rmqTopic, l.rmqGroup)
				if rmqMsgErr != nil {
					l.toolbox.Log.Errorf("RmqManagerInst.Consume() failed.")
					continue
				}
				//通知全部的ats
				notice(rmqMsg)
			}
		}
	}
}

//清除函数，定时清除无效节点
func (l *LicEx) purge() {
	doPurge := func(in *SubSvcEx) {
		/*
			1、临时取出来，减少锁的作用时间
			2、读取时
				a、先处理subSvcMapEx，取出无效节点的addr
			3、删除时
				a、先从处理segIdAddr，否则会读取到脏数据，删除
		*/
		var addrList []string
		var segIdList []string

		in.subSvcMapExRwMu.RLock()
		for k, v := range in.subSvcMapEx {
			if (time.Now().UnixNano() - v.timestamp) > v.ttl {
				v.dead = 1
				addrList = append(addrList, k)
				segIdList = append(segIdList, strconv.Itoa(int(v.segId)))
			}
		}
		in.subSvcMapExRwMu.RUnlock()

		//删除segId列表
		for _, v := range segIdList {
			in.segIdAddrRwMu.Lock()
			delete(in.segIdAddr, v)
			in.segIdAddrRwMu.Unlock()
		}

		//删除addr列表
		for _, v := range addrList {
			in.subSvcMapExRwMu.Lock()
			delete(in.subSvcMapEx, v)
			in.subSvcMapExRwMu.Unlock()
			MysqlManagerInst.DelServer(v)
		}
	}
	for {
		select {
		case <-l.ticker.C:
			{
				//临时取出来，作缓存，减少锁svcMapExRwMu作用的时间
				var subSvcList []string
				l.svcMapExRwMu.RLock()
				for k := range l.svcMapEx {
					subSvcList = append(subSvcList, k)
				}
				l.svcMapExRwMu.RUnlock()
				for _, v := range subSvcList {
					l.svcMapExRwMu.RLock()
					SubSvcExTmp := l.svcMapEx[v]
					l.svcMapExRwMu.RUnlock()
					doPurge(SubSvcExTmp)
				}
			}
		}
	}
}
func (l *LicEx) set(opt ...SetInPutOpt) (err LbErr) {
	optInst := &SetInPut{}
	for _, optFunc := range opt {
		optFunc(optInst)
	}
	if l.svc != optInst.svc {
		return ErrLbSvcIncorrect
	}
	l.svcMapExRwMu.RLock()
	SubSvcTmp, SubSvcTmpOk := l.svcMapEx[optInst.subSvc]
	l.svcMapExRwMu.RUnlock()
	if SubSvcTmpOk {
		SubSvcTmp.subSvcMapExRwMu.RLock()
		item, itemOk := SubSvcTmp.subSvcMapEx[optInst.addr]
		SubSvcTmp.subSvcMapExRwMu.RUnlock()
		if !itemOk {
			//a、节点第一次上报
			//1、在subSvcMap中新增加addr key，并为此节点选取合适的segId
			//1.1、segId为从零递增的自然数序列（从SegIdManager实例中获取）
			tmp := &SubSvcExItem{segId: segIdManagerInst.getMin(), ttl: SubSvcTmp.ttl, timestamp: time.Now().UnixNano(), addr: optInst.addr, bestInst: optInst.best, idleInst: optInst.idle, totalInst: optInst.total}
			SubSvcTmp.subSvcMapExRwMu.Lock()
			SubSvcTmp.subSvcMapEx[optInst.addr] = tmp
			SubSvcTmp.subSvcMapExRwMu.Unlock()
			segIdStr := strconv.Itoa(int(tmp.segId))
			SubSvcTmp.segIdAddrRwMu.Lock()
			/*
				1、判断segId是否已存在
				2、如不存在则添加segId并写入数据库
			*/
			if _, ok := SubSvcTmp.segIdAddr[segIdStr]; !ok {
				SubSvcTmp.segIdAddr[segIdStr] = optInst.addr
				cnt := 0
				row := RowData{segIdDb: segIdStr, typeDb: optInst.subSvc, serverIpDb: optInst.addr}
				for {
					writeOk, writeReply := MysqlManagerInst.AddNewSegIdData(row)
					debugInst.Debugf("set writeOk:%v, writeReply:%v", writeOk, writeReply)
					cnt++
					if writeOk || cnt >= 3 {
						break
					} else {
						l.toolbox.Log.Errorf("set MysqlManagerInst.AddNewSegIdData failed")
					}
				}

			}
			SubSvcTmp.segIdAddrRwMu.Unlock()

		} else {
			//b、节点授权更新
			//1、通过key addr索引到相关的节点数据，进行递增或递减操作
			for {
				if atomic.CompareAndSwapInt64(&item.cas, 0, 1) {
					atomic.StoreInt64(&item.timestamp, time.Now().UnixNano())
					atomic.StoreInt64(&item.totalInst, optInst.total)
					atomic.StoreInt64(&item.bestInst, optInst.best)
					atomic.StoreInt64(&item.idleInst, optInst.idle)
					atomic.StoreInt64(&item.cas, 0)
					break
				}
			}
		}
	}
	return
}
func (l *LicEx) get(opt ...GetInPutOpt) (nBestNodes []string, nBestNodesErr LbErr) {
	optInst := &GetInPut{}
	for _, optFunc := range opt {
		optFunc(optInst)
	}

	if l.svc != optInst.svc {
		nBestNodesErr = ErrLbSvcIncorrect
		return
	}
	l.svcMapExRwMu.RLock()
	SubSvcTmp, SubSvcTmpOk := l.svcMapEx[optInst.subSvc]
	l.svcMapExRwMu.RUnlock()
	takeAddrMin := func(in *SubSvcEx) (addrMin string, addrMinErr error) {
		//如果超过阈值则临时转到其它节点中去
		//如果segId查不到也临时转移到其它节点
		//遍历所有节点寻找负载最小者
		loadMin := int64(math.MaxInt64)
		var loadNow int64
		in.subSvcMapExRwMu.RLock()
		for _, v := range in.subSvcMapEx {
			if v.totalInst == 0 {
				//这种情况仅在第一次拉取分段表的时候存在
				addrMinErr = ErrLbNoSurvivingNode
				return
			}
			loadNow = (v.idleInst * 100) / v.totalInst
			if loadNow < loadMin {
				loadMin = loadNow
				addrMin = v.addr
			}
		}
		in.subSvcMapExRwMu.RUnlock()
		return
	}
	if SubSvcTmpOk {
		debugInst.Debugf("success take SubSvcTmp from l.svcMapEx,optInst.subSvc:%v", optInst.subSvc)
		//将uid转换为segId
		segIdTmp := optInst.uid / SubSvcTmp.seed
		//从segId字典中取出addr
		segIdStr := strconv.FormatInt(segIdTmp, 10)
		SubSvcTmp.segIdAddrRwMu.RLock()
		addrTmp, addrTmpOk := SubSvcTmp.segIdAddr[segIdStr]
		SubSvcTmp.segIdAddrRwMu.RUnlock()
		if !addrTmpOk {
			addrMin, addrMinErr := takeAddrMin(SubSvcTmp)
			if addrMinErr != nil {
				nBestNodesErr = ErrLbNoSurvivingNode
				l.toolbox.Log.Errorf("takeAddrMin error,maybe totalInst error")
				return
			}
			nBestNodes = append(nBestNodes, addrMin)

			//添加segId并写入数据库
			SubSvcTmp.segIdAddrRwMu.Lock()
			SubSvcTmp.segIdAddr[segIdStr] = addrMin
			SubSvcTmp.segIdAddrRwMu.Unlock()
			cnt := 0
			row := RowData{segIdDb: segIdStr, typeDb: optInst.subSvc, serverIpDb: addrMin}
			for {
				writeOk, _ := MysqlManagerInst.AddNewSegIdData(row)
				cnt++
				if writeOk || cnt >= 3 {
					break
				} else {
					l.toolbox.Log.Errorf("get MysqlManagerInst.AddNewSegIdData failed")
				}
			}
		} else {
			//通过addr取出节点的详细数据
			SubSvcTmp.subSvcMapExRwMu.RLock()
			addrDetails, addrDetailsOk := SubSvcTmp.subSvcMapEx[addrTmp]
			SubSvcTmp.subSvcMapExRwMu.RUnlock()
			if !addrDetailsOk {
				//当分段表第一次从数据库中拉取时，没有相关的负载信息
				nBestNodesErr = ErrLbNoSurvivingNode
				l.toolbox.Log.Errorf("get can't take addr from subSvcMapEx,maybe logic error")
				return
			}
			if addrDetails.totalInst == 0 {
				nBestNodesErr = ErrLbNoSurvivingNode
				l.toolbox.Log.Errorf("get can't take addr from subSvcMapEx,maybe logic error")
				return
			}
			debugInst.Debug("judge whether or not be overload")
			if (100 - (addrDetails.idleInst*100)/addrDetails.totalInst) > SubSvcTmp.threshold {
				debugInst.Debugf("not overload -> idleInst:%v,totalInst:%v,threshold:%v", addrDetails.idleInst, addrDetails.totalInst, SubSvcTmp.threshold)
				//此处仅判断当前节点是否过载，没过载则返回当前节点
				nBestNodes = append(nBestNodes, addrDetails.addr)
			} else {
				debugInst.Debugf("already overload -> idleInst:%v,totalInst:%v,threshold:%v", addrDetails.idleInst, addrDetails.totalInst, SubSvcTmp.threshold)
				addrMin, addrMinErr := takeAddrMin(SubSvcTmp)
				if addrMinErr != nil {
					nBestNodesErr = ErrLbNoSurvivingNode
					l.toolbox.Log.Errorf("takeAddrMin error,maybe totalInst error")
					return
				}
				nBestNodes = append(nBestNodes, addrMin)
			}
		}
	} else {
		debugInst.Debugf("fail take SubSvcTmp from l.svcMapEx,optInst.subSvc:%v", optInst.subSvc)
		//map中查不到意味着subSvc错误
		nBestNodesErr = ErrLbSubSvcIncorrect
	}
	return
}

func (l *LicEx) init(toolbox *xsf.ToolBox) {
	l.toolbox = toolbox

	baseUrlString, baseUrlErr := l.toolbox.Cfg.GetString(DB, DBBASEURL)
	if baseUrlErr != nil {
		log.Fatalf("l.toolbox.Cfg.GetString(%v, %v)", DB, DBBASEURL)
	}
	callerString, callerErr := l.toolbox.Cfg.GetString(DB, DBCALLER)
	if callerErr != nil {
		log.Fatalf("l.toolbox.Cfg.GetString(%v, %v)", DB, DBCALLER)
	}
	callerKeyString, callerKeyErr := l.toolbox.Cfg.GetString(DB, DBCALLERKEY)
	if callerKeyErr != nil {
		log.Fatalf("l.toolbox.Cfg.GetString(%v, %v)", DB, DBCALLERKEY)
	}
	timeoutInt, timeoutErr := l.toolbox.Cfg.GetInt(DB, DBTIMEOUT)
	if timeoutErr != nil {
		log.Fatalf("l.toolbox.Cfg.GetString(%v, %v)", DB, DBTIMEOUT)
	}
	tokenString, tokenErr := l.toolbox.Cfg.GetString(DB, DBTOKEN)
	if tokenErr != nil {
		log.Fatalf("l.toolbox.Cfg.GetString(%v, %v)", DB, DBTOKEN)
	}
	versionString, versionErr := l.toolbox.Cfg.GetString(DB, DBVERSION)
	if versionErr != nil {
		log.Fatalf("l.toolbox.Cfg.GetString(%v, %v)", DB, DBVERSION)
	}
	idcString, idcErr := l.toolbox.Cfg.GetString(DB, DBIDC)
	if idcErr != nil {
		log.Fatalf("l.toolbox.Cfg.GetString(%v, %v)", DB, DBIDC)
	}
	schemaString, schemaErr := l.toolbox.Cfg.GetString(DB, DBSCHEMA)
	if schemaErr != nil {
		log.Fatalf("l.toolbox.Cfg.GetString(%v, %v)", DB, DBSCHEMA)
	}
	tableString, tableErr := l.toolbox.Cfg.GetString(DB, DBTABLE)
	if tableErr != nil {
		log.Fatalf("l.toolbox.Cfg.GetString(%v, %v)", DB, DBTABLE)
	}
	MysqlManagerInst.Init(baseUrlString, callerString, callerKeyString, time.Duration(timeoutInt)*time.Millisecond, tokenString, versionString, idcString, schemaString, tableString)
	svcString, svcStringErr := l.toolbox.Cfg.GetString(BO, SVC)
	if svcStringErr != nil {
		log.Fatalf("l.toolbox.Cfg.GetString(%v, %v)", BO, SVC)
	}
	l.svc = svcString

	tickerInt64, tickerInt64Err := l.toolbox.Cfg.GetInt64(BO, TICKER)
	if tickerInt64Err != nil {
		log.Fatalf("l.toolbox.Cfg.GetString(%v, %v)", BO, TICKER)
	}
	l.ticker = time.NewTicker(time.Millisecond * time.Duration(tickerInt64))

	subSvcString, subSvcStringErr := l.toolbox.Cfg.GetString(BO, SUBSVC)
	if subSvcStringErr != nil {
		log.Fatalf("l.toolbox.Cfg.GetString(%v, %v)", BO, SUBSVC)
	}

	thresholdInt64, thresholdInt64Err := l.toolbox.Cfg.GetInt64(BO, THRESHOLD)
	if thresholdInt64Err != nil {
		log.Fatalf("l.toolbox.Cfg.GetString(%v, %v)", BO, THRESHOLD)
	}

	rmqtopicString, rmqtopicStringErr := l.toolbox.Cfg.GetString(BO, RMQTOPIC)
	if rmqtopicStringErr != nil {
		log.Fatalf("l.toolbox.Cfg.GetString(%v, %v)", BO, RMQTOPIC)
	}

	if err := dcInst.Init(); err != nil {
		log.Fatalf("dcInst.Init fail -> err:%v", err)
	}

	l.rmqTopic = rmqtopicString
	rmqgroupString, rmqgroupStringErr := l.toolbox.Cfg.GetString(BO, RMQGROUP)
	if rmqgroupStringErr != nil {
		log.Fatalf("l.toolbox.Cfg.GetString(%v, %v)", BO, RMQGROUP)
	}
	l.rmqGroup = rmqgroupString

	subSvcItems := strings.Split(subSvcString, ",")
	l.svcMapEx = make(map[string]*SubSvcEx, len(subSvcItems))
	for _, subSvc := range subSvcItems {
		ttlInt, ttlIntErr := l.toolbox.Cfg.GetInt(subSvc, TTL)
		if ttlIntErr != nil {
			log.Fatalf("l.toolbox.Cfg.GetInt(%v, %v)", subSvc, TTL)
		}
		segIdAddr, segIdAddrErr := MysqlManagerInst.GetSubSvcSegIdData(subSvc)
		if segIdAddrErr != nil {
			panic(fmt.Sprintf("pull data from mysql failed.subSvc:%v", subSvc))
		}
		l.svcMapExRwMu.Lock()
		ttl := int64(time.Millisecond) * int64(ttlInt)
		l.svcMapEx[subSvc] = &SubSvcEx{segIdAddr: segIdAddr, seed: SEED, threshold: thresholdInt64, ttl: ttl,
			subSvcMapEx: func() (res map[string]*SubSvcExItem) {
				res = make(map[string]*SubSvcExItem)
				for segId, serverIp := range segIdAddr {
					segIdInt, _ := strconv.Atoi(segId)
					tmp := &SubSvcExItem{segId: int64(segIdInt), ttl: ttl, timestamp: time.Now().UnixNano(), addr: serverIp}
					res[serverIp] = tmp
				}
				return
			}()}
		l.svcMapExRwMu.Unlock()
	}

	go l.purge() //定时清除无效节点
}
func (l *LicEx) serve(in *xsf.Req, span *xsf.Span, toolbox *xsf.ToolBox) (res *utils.Res, err error) {
	res = xsf.NewRes()

	switch in.Op() {
	case REPORTER:
		{
			//获取addr
			addrString, addrOk := in.GetParam(LICADDR)
			if !addrOk {
				res.SetError(ErrLbAddrIsIncorrect.errCode, ErrLbAddrIsIncorrect.errInfo)
				return res, nil
			}
			//获取svc
			svcString, svcOk := in.GetParam(LICSVC)
			if !svcOk {
				res.SetError(ErrLbAddrIsIncorrect.errCode, ErrLbAddrIsIncorrect.errInfo)
				return res, nil
			}

			//获取subSvc
			subSvcString, subSvcOk := in.GetParam(LICSUBSVC)
			if !subSvcOk {
				res.SetError(ErrLbAddrIsIncorrect.errCode, ErrLbAddrIsIncorrect.errInfo)
				return res, nil
			}
			//获取total
			totalString, totalOk := in.GetParam(LICTOTAL)
			if !totalOk {
				res.SetError(ErrLbTotalIsIncorrect.errCode, ErrLbTotalIsIncorrect.errInfo)
				return res, nil
			}
			totalInt, totalErr := strconv.Atoi(totalString)
			if totalErr != nil {
				toolbox.Log.Errorf("totalErr:%v", totalErr)
				res.SetError(ErrLbTotalIsIncorrect.errCode, ErrLbTotalIsIncorrect.errInfo)
				return res, nil
			}

			//获取idle
			idleString, idleOk := in.GetParam(LICIDLE)
			if !idleOk {
				res.SetError(ErrLbIdleIsIncorrect.errCode, ErrLbIdleIsIncorrect.errInfo)
				return res, nil
			}
			idleInt, idleErr := strconv.Atoi(idleString)
			if idleErr != nil {
				toolbox.Log.Errorf("idleErr:%v", idleErr)
				res.SetError(ErrLbIdleIsIncorrect.errCode, ErrLbIdleIsIncorrect.errInfo)
				return res, nil
			}

			//获取best
			bestString, bestOk := in.GetParam(LICBEST)
			if !bestOk {
				res.SetError(ErrBestIsIncorrect.errCode, ErrBestIsIncorrect.errInfo)
				return res, nil
			}
			bestInt, bestErr := strconv.Atoi(bestString)
			if bestErr != nil {
				toolbox.Log.Errorf("bestErr:%v", bestErr)
				res.SetError(ErrBestIsIncorrect.errCode, ErrBestIsIncorrect.errInfo)
				return res, nil
			}
			totalInt64, idleInt64, bestInt64 := int64(totalInt), int64(idleInt), int64(bestInt)
			setErr := l.set(withSetAddr(addrString), withSetSvc(svcString), withSetSubSvc(subSvcString), withSetTotal(totalInt64), withSetIdle(idleInt64), withSetBest(bestInt64))
			if setErr != nil {
				res.SetError(setErr.ErrorCode(), setErr.ErrInfo())
			}
		}
	case CLIENT:
		{
			//获取svc
			svcString, svcOk := in.GetParam(LICSVC)
			if !svcOk {
				res.SetError(ErrLbAddrIsIncorrect.errCode, ErrLbAddrIsIncorrect.errInfo)
				return res, nil
			}
			//获取subSvc
			subSvcString, subSvcOk := in.GetParam(LICSUBSVC)
			if !subSvcOk {
				res.SetError(ErrLbAddrIsIncorrect.errCode, ErrLbAddrIsIncorrect.errInfo)
				return res, nil
			}
			//获取nbest
			nBestString, nBestOk := in.GetParam(NBESTTAG)
			if !nBestOk {
				res.SetError(ErrLbNbestIsIncorrect.errCode, ErrLbNbestIsIncorrect.errInfo)
				return res, nil
			}

			//获取exparam，选传
			exparamString, _ := in.GetParam(EXPARAM)

			//获取all
			allString, allOk := in.GetParam(ALL)
			all := false
			if allOk {
				if allString == "1" {
					all = true
				}
			}

			//获取uid
			uidString, uidOk := in.GetParam(UID)
			var uidInt64 int64
			if !uidOk || !func() bool {
				uidInt, uidErr := strconv.Atoi(uidString)
				if uidErr != nil {
					debugInst.Debug("uid is incorrect.")
					return false
				}
				uidInt64 = int64(uidInt)
				return true
			}() {
				res.SetError(ErrLbUidIsIncorrect.errCode, ErrLbUidIsIncorrect.errInfo)
				return res, nil
			}

			nBestInt, nBestErr := strconv.Atoi(nBestString)
			if nBestErr != nil || nBestInt <= 0 {
				debugInst.Debug("nBestErr != nil || nBestInt <= 0")
				toolbox.Log.Errorf("nBestErr:%v", nBestErr)
				res.SetError(ErrLbNbestIsIncorrect.errCode, ErrLbNbestIsIncorrect.errInfo)
				return res, nil
			}

			nBestNodes, nBestNodesErr := l.get(withGetExParam(exparamString), withGetUid(uidInt64), withGetAll(all), withGetNBest(int64(nBestInt)), withGetSubSvc(subSvcString), withGetSvc(svcString))
			if nBestNodesErr != nil {
				res.SetError(nBestNodesErr.ErrorCode(), nBestNodesErr.ErrInfo())
			}
			for _, node := range nBestNodes {
				data := utils.NewData()
				data.Append([]byte(node))
				res.AppendData(data)
			}
		}
	default:
		{
			res.SetError(ErrLbInputOperation.errCode, ErrLbInputOperation.errInfo)
		}
	}
	return res, nil
}
func newLicEx(toolbox *xsf.ToolBox) *LicEx {
	licTmp := &LicEx{}
	licTmp.init(toolbox)
	rmqAddrs, rmqAddrsErr := toolbox.Cfg.GetString(BO, RMQADDRS)
	if rmqAddrsErr != nil {
		panic(fmt.Sprintf("can't get %v from %v", RMQADDRS, BO))
	}
	if !debugFlag { //调试时，暂不关注rmq
		if RmqManagerInst.Init(strings.Split(rmqAddrs, ",")) != nil {
			panic("RmqManagerInst init failed")
		}
	}
	return licTmp
}
