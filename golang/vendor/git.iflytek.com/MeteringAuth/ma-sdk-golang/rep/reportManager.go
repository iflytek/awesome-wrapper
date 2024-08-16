package rep

import (
	"strconv"
	"time"

	xsf "git.iflytek.com/AIaaS/xsf/client"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/ma/config"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/ma/tool"
)

type ReportManager struct {
	xsfc *xsf.Client
	to   time.Duration
	addr string
}

const (
	op        = "rep"
	opRepSync = "rep_sync"
	opAqcSync = "aqc_conc_batch_sync"
	opAsync   = "async_conc_sync"
)

func (r *ReportManager) classifyInt(d map[string]uint) (c map[string]*xsf.Req) {
	if r.xsfc == nil {
		return
	}
	classifer := xsf.NewCaller(r.xsfc)
	c = make(map[string]*xsf.Req)
	for k, v := range d {
		// "100I":10
		// "600M":20
		count := strconv.Itoa(int(v))
		addr, err := classifer.GetHashAddr(k, config.SVC)
		if err == nil {
			req, ok := c[addr]
			if !ok {
				req = xsf.NewReq()
				req.SetOp(k)
				c[addr] = req
			}
			req.SetParam(k, count)
			tool.L.Debugw("rep-sdk | ReportManager | classifyInt", "addr", addr)
		} else {
			tool.L.Errorw("rep-sdk | ReportManager | classifyInt", "error", err, "key", k, "svc", config.SVC)
		}
	}
	return
}

//	func (r *ReportManager) getAqcReq(appid, channel, function, addr string) *xsf.Req {
//		if r.xsfc == nil {
//			return nil
//		}
//		classifer := xsf.NewCaller(r.xsfc)
//		addr, err := classifer.GetHashAddr(appid, config.SVC)
//		if err == nil {
//			req := xsf.NewReq()
//			req.SetOp("aqc_inc")
//			req.Append()
//			tool.L.Debugw("rep-sdk | ReportManager | classifyInt", "addr", addr)
//			return req
//		} else {
//			tool.L.Errorw("rep-sdk | ReportManager | classifyInt", "error", err, "key", appid, "svc", config.SVC)
//			return nil
//		}
//	}
func (r *ReportManager) classifyString(d map[string]string) (c map[string]*xsf.Req) {
	if r.xsfc == nil {
		return
	}
	classifer := xsf.NewCaller(r.xsfc)
	c = make(map[string]*xsf.Req)
	for k, v := range d {
		// "100I":"10"
		// "600M":"20"
		addr, err := classifer.GetHashAddr(k, config.SVC)
		if err == nil {
			req, ok := c[addr]
			if !ok {
				req = xsf.NewReq()
				req.SetOp(k)
				c[addr] = req
			}
			req.SetParam(k, v)
			tool.L.Debugw("rep-sdk | ReportManager | classifyString", "addr", addr)
		} else {
			tool.L.Errorw("rep-sdk | ReportManager | classifyString", "error", err, "key", k, "svc", config.SVC)
		}
	}
	return
}

func (r *ReportManager) report(channel, ent string, authInfo map[string]uint, aqcreport bool) {
	c := r.classifyInt(authInfo)
	if c == nil {
		return
	}

	m := map[string]string{"channel": channel, "addr": r.addr, Aqc_Report: strconv.FormatBool(aqcreport), "op": op}
	if ent != "" {
		m["on_ent"] = ent
	}

	if config.C.Log.Level == "debug" {
		tool.L.Debugw("rep-sdk | report", "channel", channel, "params", m, "info", authInfo)
	}

	for _, v := range c {
		func(req *xsf.Req) {
			caller := xsf.NewCaller(r.xsfc)
			caller.WithHashKey(req.Op())
			req.Append([]byte{}, m)
			_, errcode, err := caller.Call(config.SVC, op, req, r.to)
			if err != nil {
				tool.L.Errorw("rep-sdk | ReportManager | report", "code", errcode, "error", err)
			}
		}(v)
	}
}

func (r *ReportManager) sync(authInfo map[string]string, channel, ent, addr string, aqcreport bool) {
	c := r.classifyString(authInfo)
	if c == nil {
		return
	}

	m := map[string]string{"channel": channel, "addr": addr, Aqc_Report: strconv.FormatBool(aqcreport), "op": op}
	if ent != "" {
		m["on_ent"] = ent
	}

	if config.C.Log.Level == "debug" {
		tool.L.Debugw("rep-sdk | sync", "channel", channel, "params", m, "info", authInfo)
	}

	for _, v := range c {
		func(req *xsf.Req) {
			caller := xsf.NewCaller(r.xsfc)
			caller.WithHashKey(req.Op())
			req.Append([]byte{}, m)
			_, errcode, err := caller.Call(config.SVC, opRepSync, req, r.to)
			if err != nil {
				tool.L.Errorw("rep-sdk | ReportManager | sync", "code", errcode, "error", err)
			}
		}(v)
	}
}

func (r *ReportManager) reportWithAddr(channel, ent string, authInfo map[string]uint, addr string, aqcreport bool) {
	c := r.classifyInt(authInfo)
	if c == nil {
		return
	}
	m := map[string]string{"channel": channel, "addr": addr, Aqc_Report: strconv.FormatBool(aqcreport), "op": op}
	if ent != "" {
		m["on_ent"] = ent
	}

	if config.C.Log.Level == "debug" {
		tool.L.Debugw("rep-sdk | reportWithAddr", "channel", channel, "params", m, "info", authInfo)
	}

	for _, v := range c {
		func(req *xsf.Req) {
			caller := xsf.NewCaller(r.xsfc)
			caller.WithHashKey(req.Op())
			req.Append([]byte{}, m)
			_, errcode, err := caller.Call(config.SVC, op, req, r.to)
			if err != nil {
				tool.L.Errorw("rep-sdk | ReportManager | reportWithAddr", "code", errcode, "error", err)
			}
		}(v)
	}
}

//// 会在doSession之前并发+1，
//// doSession执⾏完成后，并发-1
//// doSession: ⼀路会话执⾏函数
//// 精确并发控制 +1
//func (r *ReportManager) concInc(appid, channel, function, addr string) error {
//	r.getAqcReq(appid, channel, function, addr)
//	xsfreq.SetOp("aqc_inc")
//	xsfreq.SetParam([app_id,channel,func,addr])
//}
//
//// 精确并发控制 -1
//func (r *ReportManager) concDec(appid, channel, function, addr string) error {
//	xsfreq.SetOp("aqc_dec")
//	xsfreq.SetParam([app_id,channel,func,addr])
//}
//
//// 异步并发控制 并发数+1。如果expireSeconds 之后没有调⽤ConcDecWithRequestId 将对应的
//// requestId -1.那么系统会⾃动将并发数-1
//// expireSecond 最⼤ 100000s
//func (r *ReportManager) concIncWithRequestId(appid, channel, function, requestId string, expireSecond int64) error {
//	xsfreq.SetOp("async_inc")
//	xsfreq.SetParam([app_id,channel,func,request_id,expire])
//}
//
//// 异步并发控制 并发数 -1
//func (r *ReportManager) concDecWithRequestId(appid, channal, function, requestId string) error {
//	xsfreq.SetOp("async_dec")
//	xsfreq.SetParam([app_id,channel,func,request_id])
//}

func (r *ReportManager) fini() {
	// TODO: zk deadlock here
	// xsf.DestroyClient(r.xsfc)
}
