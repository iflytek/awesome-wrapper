package daemon

import (
	xsfc "git.xfyun.cn/AIaaS/xsf/client"
	"git.xfyun.cn/AIaaS/xsf/utils"
	"time"
)

const defCfg = "[sfc]\ntimeout=1000\ntimeout=500"
const defTimeout = time.Second

var dcInst dc

type dcOpt func(*dc)

func withDcTimeout(in time.Duration) dcOpt {
	return func(i *dc) {
		i.tm = in
	}
}

type dc struct {
	tm     time.Duration
	cli    *xsfc.Client
	caller *xsfc.Caller
}

func (d *dc) Init(opts ...dcOpt) error {
	cli, cliErr := xsfc.InitClient("dc", utils.Native, utils.WithCfgDefault(defCfg))
	if cliErr != nil {
		return cliErr
	}
	d.cli = cli
	d.caller = xsfc.NewCaller(cli)
	d.tm = defTimeout
	for _, o := range opts {
		o(d)
	}
	return nil
}
func (d *dc) call(service string, op string, addr string, r *xsfc.Req) (s *xsfc.Res, errcode int32, e error) {
	return d.caller.CallWithAddr(service, op, addr, r, d.tm)
}
func (d *dc) NoticeClient(addr string, op, svc, subsvc, uid string) (s *xsfc.Res, errcode int32, e error) {
	r := xsfc.NewReq()
	r.SetParam(NBESTTAG, "1")
	r.SetParam(SVC, "svc")
	r.SetParam(SUBSVC, subsvc)
	r.SetParam(UID, uid)
	return d.call("", op, addr, r)
}
