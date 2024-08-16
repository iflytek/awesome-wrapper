package daemon

import (
	"encoding/json"
	"git.xfyun.cn/AIaaS/xsf/server"
	"git.xfyun.cn/AIaaS/xsf/utils"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type SubSvc struct {
	ttl         int64 //纳秒,定时清除无效节点的时间间隔
	subSvcRwMu  sync.RWMutex
	subSvcMap   map[string]*SubSvcItem //K:addr
	subSvcSlice SubSvcItemSlice
}
type Lic struct {
	svc     string
	svcMap  map[string]*SubSvc
	licRwMu sync.RWMutex

	toolbox *xsf.ToolBox
}

func (l *Lic) set(opt ...SetInPutOpt) (err LbErr) {
	optInst := &SetInPut{}
	for _, optFunc := range opt {
		optFunc(optInst)
	}
	if l.svc != optInst.svc {
		return ErrLbSvcIncorrect
	}
	l.licRwMu.RLock()
	SubSvcTmp, SubSvcTmpOk := l.svcMap[optInst.subSvc]
	l.licRwMu.RUnlock()
	if SubSvcTmpOk {
		SubSvcTmp.subSvcRwMu.RLock()
		item, itemOk := SubSvcTmp.subSvcMap[optInst.addr]
		SubSvcTmp.subSvcRwMu.RUnlock()
		if !itemOk {
			tmp := &SubSvcItem{ttl: SubSvcTmp.ttl, timestamp: time.Now().UnixNano(), addr: optInst.addr, bestInst: optInst.best, idleInst: optInst.idle, totalInst: optInst.total}
			SubSvcTmp.subSvcRwMu.Lock()
			SubSvcTmp.subSvcMap[optInst.addr] = tmp
			SubSvcTmp.subSvcSlice = append(SubSvcTmp.subSvcSlice, tmp)
			sort.Sort(SubSvcTmp.subSvcSlice)
			SubSvcTmp.subSvcRwMu.Unlock()
		} else {
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
			SubSvcTmp.subSvcRwMu.Lock()
			sort.Sort(SubSvcTmp.subSvcSlice)
			SubSvcTmp.subSvcRwMu.Unlock()
		}
	}
	return
}
func (l *Lic) get(opt ...GetInPutOpt) (nBestNodes []string, nBestNodesErr LbErr) {
	optInst := &GetInPut{}
	for _, optFunc := range opt {
		optFunc(optInst)
	}
	if l.svc != optInst.svc {
		nBestNodesErr = ErrLbSvcIncorrect
		return
	}
	l.licRwMu.RLock()
	SubSvcTmp, SubSvcTmpOk := l.svcMap[optInst.subSvc]
	l.licRwMu.RUnlock()

	if SubSvcTmpOk {
		if int64(len(SubSvcTmp.subSvcSlice)) < optInst.nBest {
			nBestNodesErr = ErrLbNoSurvivingNode
		} else {
			var nBestCount int64 = 0
			for ix := len(SubSvcTmp.subSvcSlice) - 1; ix > -1; ix-- {
				if (time.Now().UnixNano() - SubSvcTmp.subSvcSlice[ix].timestamp) > SubSvcTmp.ttl {
					for {
						if atomic.CompareAndSwapInt64(&SubSvcTmp.subSvcSlice[ix].cas, 0, 1) {
							atomic.StoreInt64(&SubSvcTmp.subSvcSlice[ix].dead, 1)
							atomic.StoreInt64(&SubSvcTmp.subSvcSlice[ix].cas, 0)
							break
						}
					}
					continue
				}
				nBestNodes = append(nBestNodes, SubSvcTmp.subSvcSlice[ix].addr)
				if nBestCount++; nBestCount == optInst.nBest {
					break
				}
			}
			if optInst.all {
				all := "allNodes:" + func() string {
					var rst []string
					for _, v := range SubSvcTmp.subSvcSlice {
						m := make(map[string]interface{}, 10)
						rst = append(rst, func() string {
							m["addr"] = v.addr
							m["idleInst"] = v.idleInst
							m["bestInst"] = v.bestInst
							m["bestInst"] = v.bestInst
							r, _ := json.Marshal(m)
							return string(r)
						}())
					}
					return strings.Join(rst, ";")
				}()
				nBestNodes = append(nBestNodes, all)
			}
		}
	} else {
		nBestNodesErr = ErrLbSubSvcIncorrect
	}
	return
}
func (l *Lic) init(toolbox *xsf.ToolBox) {
	l.toolbox = toolbox

	svcString, svcStringErr := l.toolbox.Cfg.GetString(BO, SVC)
	if svcStringErr != nil {
		log.Fatalf("l.toolbox.Cfg.GetString(%v, %v)", BO, SVC)
	}
	l.svc = svcString

	subSvcString, subSvcStringErr := l.toolbox.Cfg.GetString(BO, SUBSVC)
	if subSvcStringErr != nil {
		log.Fatalf("l.toolbox.Cfg.GetString(%v, %v)", BO, SUBSVC)
	}
	subSvcItems := strings.Split(subSvcString, ",")
	l.svcMap = make(map[string]*SubSvc, len(subSvcItems))
	for _, subSvc := range subSvcItems {
		ttlInt, ttlIntlErr := l.toolbox.Cfg.GetInt(subSvc, TTL)
		if ttlIntlErr != nil {
			log.Fatalf("l.toolbox.Cfg.GetInt(%v, %v)", subSvc, TTL)
		}
		l.licRwMu.Lock()
		l.svcMap[subSvc] = &SubSvc{ttl: int64(time.Millisecond) * int64(ttlInt), subSvcSlice: make(SubSvcItemSlice, 0, 1000), subSvcMap: make(map[string]*SubSvcItem, 1000)}
		l.licRwMu.Unlock()
	}
}
func (l *Lic) serve(in *xsf.Req, span *xsf.Span, toolbox *xsf.ToolBox) (res *utils.Res, err error) {
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

			//获取all
			allString, allOk := in.GetParam(ALL)
			all := false
			if allOk {
				if allString == "1" {
					all = true
				}
			}

			nBestInt, nBestErr := strconv.Atoi(nBestString)
			if nBestErr != nil || nBestInt <= 0 {
				toolbox.Log.Errorf("nBestErr:%v", nBestErr)
				res.SetError(ErrLbNbestIsIncorrect.errCode, ErrLbNbestIsIncorrect.errInfo)
				return res, nil
			}

			nBestNodes, nBestNodesErr := l.get(withGetAll(all), withGetNBest(int64(nBestInt)), withGetSubSvc(subSvcString), withGetSvc(svcString))
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
func newLic(toolbox *xsf.ToolBox) *Lic {
	licTmp := &Lic{}
	licTmp.init(toolbox)
	return licTmp
}
