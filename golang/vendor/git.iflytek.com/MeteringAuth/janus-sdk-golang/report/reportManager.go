package report

import (
	"strconv"
	"time"

	xsf "git.xfyun.cn/AIaaS/xsf/client"
	"git.xfyun.cn/AIaaS/xsf/utils"
)

type ReportManager struct {
	xsfc *xsf.Client
	p    xsfParam
	addr string
}

type xsfParam struct {
	rpcTimeout time.Duration
	svc        string
	op         string
}

const (
	SVC          = "janus"
	OPREPORT     = "rep"
	OPREPORTSYNC = "rep_sync"
)

func newXsfClient(cfgName, cname string) (xsfClient *xsf.Client, err error) {
	xsfClient, err = xsf.InitClient(
		cname,
		utils.CfgMode(DefaultInitOption.cfgMode),
		utils.WithCfgName(cfgName),
		utils.WithCfgURL(DefaultInitOption.companionUrl),
		utils.WithCfgPrj(DefaultInitOption.project),
		utils.WithCfgGroup(DefaultInitOption.group),
		utils.WithCfgService(DefaultInitOption.service),
		utils.WithCfgVersion(DefaultInitOption.version),
		utils.WithCfgCacheService(DefaultInitOption.isCacheService),
		utils.WithCfgCacheConfig(DefaultInitOption.isCacheConfig),
		utils.WithCfgCachePath(DefaultInitOption.cachePath),
	)
	if err != nil {
		return
	}
	return
}

func newReportManger(addr string) (ac *ReportManager, err error) {
	xsfClient, err := newXsfClient(CfgName, ClientName)
	if err != nil {
		return
	}
	timeout, err := xsfClient.Cfg().GetInt64("janus-report", "timeout")
	if err != nil {
		timeout = 50

	}
	ac = &ReportManager{
		xsfc: xsfClient,
		p: xsfParam{
			rpcTimeout: time.Duration(timeout) * time.Millisecond,
			svc:        SVC,
			op:         OPREPORT,
		},
		addr: addr,
	}

	return
}

func (r *ReportManager) classify(d map[string]uint) (c map[string]*xsf.Req) {
	// used for classifing addr:reqParam
	if r.xsfc == nil {
		return
	}
	classifer := xsf.NewCaller(r.xsfc)
	c = make(map[string]*xsf.Req)
	for k, v := range d {
		addr, err := classifer.GetHashAddr(k, SVC)
		//fmt.Println(addr)
		if err == nil {
			req, ok := c[addr]
			if !ok {
				req = xsf.NewReq()
				count := strconv.Itoa(int(v))
				req.SetParam(k, count)
				req.SetOp(k)
				c[addr] = req
			} else {
				count := strconv.Itoa(int(v))
				req.SetParam(k, count)
			}
			//r.xsfc.Log.Debugf("addr is %v", addr)
		}
	}
	return
}
func (r *ReportManager) classify2(d map[string]string) (c map[string]*xsf.Req) {

	// used for classifing addr:reqParam
	if r.xsfc == nil {
		return
	}
	classifer := xsf.NewCaller(r.xsfc)
	c = make(map[string]*xsf.Req)
	for k, v := range d {
		addr, err := classifer.GetHashAddr(k, SVC)
		//fmt.Println(addr)
		if err == nil {
			req, ok := c[addr]
			if !ok {
				req = xsf.NewReq()
				req.SetParam(k, v)
				req.SetOp(k)
				c[addr] = req
			} else {
				req.SetParam(k, v)
			}
			r.xsfc.Log.Debugf("addr is %v", addr)
		}
	}
	return
}

func (r *ReportManager) report(channel string, authInfo map[string]uint, useEnt bool) {
	c := r.classify(authInfo)
	if c == nil {
		return
	}

	m := map[string]string{"channel": channel, "addr": r.addr}
	if useEnt {
		m["repEnt"] = "1"
	}

	for _, v := range c {
		func(req *xsf.Req) {
			caller := xsf.NewCaller(r.xsfc)
			caller.WithHashKey(req.Op())
			req.Append([]byte{}, m)
			_, errcode, err := caller.Call(r.p.svc, r.p.op, req, r.p.rpcTimeout)
			if err != nil {
				r.xsfc.Log.Errorf("report | errcode = %d, err=%s", errcode, err)
			}
		}(v)
	}
}

func (r *ReportManager) sync(authInfo map[string]string, channel, addr string, useEnt bool) {
	c := r.classify2(authInfo)
	if c == nil {
		return
	}
	m := map[string]string{"channel": channel, "addr": addr}
	if useEnt {
		m["repEnt"] = "1"
	}
	for _, v := range c {
		func(req *xsf.Req) {
			caller := xsf.NewCaller(r.xsfc)
			caller.WithHashKey(req.Op())
			req.Append([]byte{}, m)
			_, errcode, err := caller.Call(r.p.svc, OPREPORTSYNC, req, r.p.rpcTimeout)
			if err != nil {
				r.xsfc.Log.Errorf("report | errcode = %d, err=%s", errcode, err)
			}
		}(v)
	}

}

func (r *ReportManager) reportWithAddr(channel string, authInfo map[string]uint, addr string, useEnt bool) {
	c := r.classify(authInfo)
	if c == nil {
		return
	}
	m := map[string]string{"channel": channel, "addr": r.addr}
	if useEnt {
		m["repEnt"] = "1"
	}
	for _, v := range c {
		func(req *xsf.Req) {
			caller := xsf.NewCaller(r.xsfc)
			caller.WithHashKey(req.Op())
			req.Append([]byte{}, m)
			_, errcode, err := caller.Call(r.p.svc, r.p.op, req, r.p.rpcTimeout)
			if err != nil {
				r.xsfc.Log.Errorf("report | errcode = %d, err=%s", errcode, err)
			}
		}(v)
	}
}

func (r *ReportManager) fini() {
	xsf.DestroyClient(r.xsfc)
}
