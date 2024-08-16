package xsf

import (
	"fmt"
	"git.iflytek.com/AIaaS/xsf/client"
	"git.iflytek.com/AIaaS/xsf/utils"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	CNAME = "lbv2"
)

type finderAddr struct {
	project, group, service, apiVersion, sub, ent string
}
type SvcUnit struct {
	Svc    string `json:"svc"`
	SubSvc string `json:"sub_svc"`
}
type MoveTo struct {
	Original SvcUnit `json:"original"`
	From     SvcUnit `json:"from"`
	Now      SvcUnit `json:"now"`
}
type hermesAdapter struct {
	mode           utils.CfgMode
	able           bool
	originalSvc    string
	originalSubSvc string
	fromSvc        string
	fromSubSvc     string
	extra          string //保存负载以外的一些信息
	addr           string
	uid            string
	total          int
	idle           int
	best           int

	lbTargets         string
	lbTargetsAnalyzed []finderAddr
	finderTtl         time.Duration //更新本地地址的时间，通过访问服务发现实现
	backend           int           //上报的的协程数，缺省4
	timeout           time.Duration //上报的超时时间，缺省一秒

	bc BootConfig

	cli    *xsf.Client
	caller *xsf.Caller

	detectors map[string]*detector

	taskInChan chan func() //任务通道，用来传送上报任务
	//taskOutChan chan callRstItem

	hermesTask int
	svcIp      string //服务端监听ip，trace用
	svcPort    int32  //服务端监听端口，trace用

	cloud string //cloud_id
}
type HermesAdapterCfgOpt func(*hermesAdapter)

func WithHermesAdapterCloudId(cloud string) HermesAdapterCfgOpt {
	return func(h *hermesAdapter) {
		h.cloud = cloud
	}
}

func WithHermesAdapterSvcIp(svcIp string) HermesAdapterCfgOpt {
	return func(h *hermesAdapter) {
		h.svcIp = svcIp
	}
}
func WithHermesAdapterTask(task int) HermesAdapterCfgOpt {
	return func(h *hermesAdapter) {
		h.hermesTask = task
	}
}
func WithHermesAdapterSvcPort(svcPort int32) HermesAdapterCfgOpt {
	return func(h *hermesAdapter) {
		h.svcPort = svcPort
	}
}
func WithHermesAdapterTimeout(timeout time.Duration) HermesAdapterCfgOpt {
	return func(h *hermesAdapter) {
		h.timeout = timeout
	}
}
func WithHermesAdapterLbTargets(lbTargets string) HermesAdapterCfgOpt {
	return func(h *hermesAdapter) {
		h.lbTargets = lbTargets
	}
}
func WithHermesAdapterBackEnd(backend int) HermesAdapterCfgOpt {
	return func(h *hermesAdapter) {
		h.backend = backend
	}
}
func WithHermesAdapterAddr(addr string) HermesAdapterCfgOpt {
	return func(h *hermesAdapter) {
		h.addr = addr
	}
}
func WithHermesAdapterMode(mode utils.CfgMode) HermesAdapterCfgOpt {
	return func(h *hermesAdapter) {
		h.mode = mode
	}
}
func WithHermesAdapterFinderTtl(finderTtl time.Duration) HermesAdapterCfgOpt {
	return func(h *hermesAdapter) {
		h.finderTtl = finderTtl
	}
}

func WithHermesAdapterBootConfig(bc BootConfig) HermesAdapterCfgOpt {
	return func(h *hermesAdapter) {
		h.bc = bc
	}
}

func (h *hermesAdapter) Init(opts ...HermesAdapterCfgOpt) (err error) {
	for _, o := range opts {
		o(h)
	}
	if !h.able {
		loggerStd.Printf("hermes not enable\n")
		return
	}
	loggerStd.Printf("hermes init->cfgName:%v,companion:%v,prj:%v,grp:%v,srv:%v,ver:%v\n",
		h.bc.CfgData.CfgName, h.bc.CfgData.CompanionUrl, h.bc.CfgData.Project, h.bc.CfgData.Group, h.bc.CfgData.Service, h.bc.CfgData.Version)
	h.cli, err = xsf.InitClient(
		CNAME,
		h.mode,
		utils.WithCfgCacheService(true),
		utils.WithCfgCacheConfig(true),
		utils.WithCfgCachePath("."),
		utils.WithCfgName(h.bc.CfgData.CfgName),
		utils.WithCfgURL(h.bc.CfgData.CompanionUrl),
		utils.WithCfgPrj(h.bc.CfgData.Project),
		utils.WithCfgGroup(h.bc.CfgData.Group),
		utils.WithCfgService(h.bc.CfgData.Service),
		utils.WithCfgVersion(h.bc.CfgData.Version),
		utils.WithCfgSvcIp(h.svcIp),
		utils.WithCfgSvcPort(h.svcPort))
	if err != nil {
		panic(fmt.Sprintf("InitClient fail err:%v", err))
		return
	}
	h.caller = xsf.NewCaller(h.cli)
	h.caller.WithApiVersion("1.0.0") //补丁 again
	h.taskInChan = make(chan func(), h.hermesTask)

	//补丁 again
	finderAddrs := func() []finderAddr {
		if !checkLbTargets(h.lbTargets) {
			panic(fmt.Sprintf("please check lb targets:%v,ref:pro1,gro1,svc1,api1,sub1,ent1;pro2,gro2,svc2,api2,sub2,ent2", h.lbTargets))
		}
		var finderAddrTmp []finderAddr
		for _, target := range strings.Split(h.lbTargets, ";") {
			tmp := strings.Split(target, ",")
			finderAddrTmp = append(finderAddrTmp, finderAddr{project: tmp[0], group: tmp[1], service: tmp[2], apiVersion: tmp[3], sub: tmp[4], ent: tmp[5]})
		}
		return finderAddrTmp
	}()
	h.lbTargetsAnalyzed = finderAddrs
	loggerStd.Printf("hermes targets %+v\n", finderAddrs)

	for _, target := range finderAddrs {
		detectorInst, detectorInstErr := newDetector(h.bc.CfgData.CompanionUrl, target.project, target.group, target.service, target.apiVersion, h.cli.Log)
		if detectorInstErr != nil {
			panic(fmt.Sprintf("init detector fail err:%v", detectorInstErr))
		}
		loggerStd.Printf("detector for %v create successfully", target)
		// 补丁 again
		if h.detectors == nil {
			h.detectors = make(map[string]*detector)
		}
		h.detectors[fmt.Sprintf("%v_%v_%v_%v", target.project, target.group, target.service, target.apiVersion)] = detectorInst
	}

	/*
		消费task
	*/
	go h.writer()
	return
}

type callRstItem struct {
	s       *Res
	errcode int32
	e       error
	addr    string
}

func (h *hermesAdapter) writer() {
	if !h.able {
		return
	}
	h.cli.Log.Infow("about to start writer", "backend", h.backend)
	wg := sync.WaitGroup{}
	for ix := 0; ix < h.backend; ix++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range h.taskInChan {

				h.cli.Log.Debugw("receive task")

				task()

			}
		}()
	}
	wg.Wait()
	h.cli.Log.Infow("writer exiting")
}

var setServerOnce sync.Once

const reportDef = -1
const reportFailure = 0
const reportSuccess = 1

const MoveFromSvc = "moveFromSvc"
const MoveFromSubSvc = "moveFromSubSvc"
const Extra = "extra"

var reportFlag = reportDef //这部分后续优化吧,-1:not set,0:failure,1:success

func (h *hermesAdapter) setServer(from SvcUnit, addr string, getAuthInfo func() (int32, int32, int32), live string) error {
	if !h.able {
		return nil
	}
	setServerOnce.Do(func() {
		for _, lbTargetAnalyzed := range h.lbTargetsAnalyzed {
			var forLog []string
			for _, lbAddr := range h.detectors[fmt.Sprintf("%v_%v_%v_%v", lbTargetAnalyzed.project, lbTargetAnalyzed.group, lbTargetAnalyzed.service, lbTargetAnalyzed.apiVersion)].getAll() {
				forLog = append(forLog, lbAddr)
			}
			loggerStd.Printf("about to reporting:%v->%+v\n", lbTargetAnalyzed, forLog)
		}
	})
	for _, lbTargetAnalyzed := range h.lbTargetsAnalyzed {
		for _, lbAddr := range h.detectors[fmt.Sprintf("%v_%v_%v_%v", lbTargetAnalyzed.project, lbTargetAnalyzed.group, lbTargetAnalyzed.service, lbTargetAnalyzed.apiVersion)].getAll() {
			h.cli.Log.Infow("report task", "lb", fmt.Sprintf("target:%+v,addr:%v", lbTargetAnalyzed, lbAddr))
			lbAddrTmp := lbAddr
			lbTargetAnalyzedTmp := lbTargetAnalyzed
			task := func() {
				req := utils.NewReq()
				req.SetParam(HERMESLBCLOUD, h.cloud)
				req.SetParam(Extra, h.extra)
				req.SetParam(MoveFromSvc, from.Svc)
				req.SetParam(MoveFromSubSvc, from.SubSvc)
				req.SetParam("svc", lbTargetAnalyzedTmp.sub)
				req.SetParam("subsvc", lbTargetAnalyzedTmp.ent)
				req.SetParam("subsvc", strings.ReplaceAll(lbTargetAnalyzedTmp.ent, "|", ","))
				req.SetParam("addr", addr)
				req.SetParam("live", live)
				for k, v := range lbReportExtInst.getAll() {
					req.SetParam(k, v)
				}

				if getAuthInfo != nil {
					maxLic, idle, bestLic := getAuthInfo()
					req.SetParam("total", strconv.Itoa(int(maxLic)))
					req.SetParam("best", strconv.Itoa(int(bestLic)))
					req.SetParam("idle", strconv.Itoa(int(idle)))
				} else {

					req.SetParam("total", "0")
					req.SetParam("best", "0")
					req.SetParam("idle", "0")
				}
				res, errcode, e := h.caller.CallWithAddr("", xsf.LBOPSET, lbAddrTmp, req, time.Second)
				if errcode != 0 || e != nil {
					reportFlag = reportFailure
					h.cli.Log.Errorw("fn:setServer h.caller.CallWithAddr", "errcode", errcode, "err", e, "addr", lbAddrTmp)
				} else {
					h.FilterMoveTo(res)
					reportFlag = reportSuccess
				}
				h.cli.Log.Infow("report task->sending req", "req", req.Req().String())
			}

			select {
			case h.taskInChan <- task:
			default:
				{
					h.cli.Log.Warnw("taskInChan overflow")
				}
			}
		}
	}

	h.cli.Log.Debugw("create report ctx", "timeout(ns)", int(h.timeout))

	return nil
}

func (h *hermesAdapter) FilterMoveTo(res *xsf.Res) {
	return
}

func (h *hermesAdapter) report(getAuthInfo func() (int32, int32, int32)) error {
	if !h.able {
		return nil
	}
	return h.setServer(SvcUnit{h.fromSvc, h.fromSubSvc}, h.addr, getAuthInfo, "1")
}
func (h *hermesAdapter) offline() error {
	if !h.able {
		return nil
	}
	return h.setServer(SvcUnit{"svc", "subsvc"}, h.addr, nil, "0")
}
