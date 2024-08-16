package acflimit

import (
	"fmt"
	"time"

	xsf "git.iflytek.com/AIaaS/xsf/client"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/licc/monitor"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/ma/config"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/ma/tool"
)

const (
	op = "gal"
)

type AcfLimitCheck struct {
	xsfc *xsf.Client
	to   time.Duration
}

func (c *AcfLimitCheck) String() string {
	return fmt.Sprintf("acflimit client: dur:%v", c.to.String())
}

func NewAcfLimitCheckManager() (ac *AcfLimitCheck, err error) {
	cname := config.CliLiccTag
	xsfClient, err := tool.NewClient(cname)
	if err != nil {
		return
	}

	timeout, err_timeout := xsfClient.Cfg().GetInt64(cname, config.CliDurTag)
	if err_timeout != nil {
		timeout = 50
	}

	ac = &AcfLimitCheck{
		xsfc: xsfClient,
		to:   time.Duration(timeout) * time.Millisecond,
	}

	tool.LiccPrinter.Println("acf timeout", ac.to.String())
	return
}

func (auth *AcfLimitCheck) GetAcfLimits(appid, channel, functions string, tag string) (authInfo map[string]string, err error) {
	req := xsf.NewReq()
	req.SetParam("appid", appid)
	req.SetParam("channel", channel)
	req.SetParam("function", functions)
	req.SetParam("tag", tag)

	xsfCaller := xsf.NewCaller(auth.xsfc)
	xsfCaller.WithHashKey(appid)

	start := time.Now()
	xsfResp, errcode, err := xsfCaller.Call(config.SVC, op, req, auth.to)
	cost := time.Since(start)
	monitor.WithCost(op, cost)
	monitor.WithCallErr(op, errcode)
	if err != nil {
		tool.L.Errorw("licc-sdk | getAcfLimits", "code", errcode, "error", err, "cost", cost.String(), "tag", tag)
		return
	}
	return xsfResp.GetAllParam(), err
}
