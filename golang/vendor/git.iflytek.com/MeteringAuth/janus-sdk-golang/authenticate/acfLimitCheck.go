package authenticate

import (
	"time"

	xsf "git.xfyun.cn/AIaaS/xsf/client"
)

const (
	OPGALCHECK = "gal"
)

type AcfLimitCheck struct {
	//caller *xsf.Caller
	xsfc *xsf.Client
	p    xsfParam
}

type AcfLimitParam struct {
	appid   string
	funcs   string //分号分隔字符串
	channel string
}

func NewAcfLimitCheckManager(cfgName, cname string) (ac *AcfLimitCheck, err error) {
	xsfClient, err := newXsfClient(cfgName, cname)
	if err != nil {
		return
	}
	timeout, err_timeout := xsfClient.Cfg().GetInt64(cname, "timeout")
	if err_timeout != nil {
		timeout = 50
		//		return

	}

	svcName, err := xsfClient.Cfg().GetString(cname, "server_name")
	if err != nil {
		svcName = SVC
	}

	ac = &AcfLimitCheck{
		xsfc: xsfClient,
		p: xsfParam{
			rpcTimeout: time.Duration(timeout) * time.Millisecond,
			svc:        svcName,
			op:         OPGALCHECK,
		},
	}
	return
}

func (auth *AcfLimitCheck) getAcfLimits(appid, channel, functions string) (authInfo map[string]string, err error) {
	req := xsf.NewReq()
	req.SetParam("appid", appid)
	req.SetParam("channel", channel)
	req.SetParam("function", functions)
	req.SetParam("sid", "gal")

	xsfCaller := xsf.NewCaller(auth.xsfc)
	xsfCaller.WithHashKey(appid)
	xsfResp, errcode, err := xsfCaller.Call(auth.p.svc, auth.p.op, req, auth.p.rpcTimeout)
	if err != nil {
		//todo metrics 出来之后，要做告警 log errcode
		// TODO: log it
		auth.xsfc.Log.Errorf("getAcfLimits | errcode = %d, err=%s", errcode, err)
		return
	}
	return xsfResp.GetAllParam(), err
}
