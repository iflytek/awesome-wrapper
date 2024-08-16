package licc

import (
	"fmt"
	"strconv"
	"time"

	xsf "git.iflytek.com/AIaaS/xsf/client"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/licc/monitor"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/ma/config"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/ma/tool"
)

const (
	op = "lic"
)

var (
	CheckOption = CtrlALL
)

type AuthenticateCheck struct {
	//caller *xsf.Caller
	xsfc *xsf.Client
	to   time.Duration
}

func (c *AuthenticateCheck) String() string {
	return fmt.Sprintf("licc client: dur:%v", c.to.String())
}

func NewCheckLicManager() (ac *AuthenticateCheck, err error) {
	cname := config.CliLiccTag
	xsfClient, err := tool.NewClient(cname)
	if err != nil {
		return
	}

	timeout, err_timeout := xsfClient.Cfg().GetInt64(cname, config.CliDurTag)
	if err_timeout != nil {
		timeout = 50
	}

	coption, err := xsfClient.Cfg().GetInt64(cname, config.CliLiccOpTag)
	if err == nil {
		CheckOption = CtrlMode(uint32(coption))
	}

	ac = &AuthenticateCheck{
		xsfc: xsfClient,
		to:   time.Duration(timeout) * time.Millisecond,
	}

	tool.LiccPrinter.Println("licc timeout", ac.to.String())
	return
}

func (auth *AuthenticateCheck) HasLicense(appid, uid, channel, functions string, attribute CtrlMode, tag string) (authInfo map[string]string, logInfo string, err error) {
	req := xsf.NewReq()
	if attribute == CtrlNone {
		attribute = CheckOption
	}
	attributeStr := strconv.Itoa(int(attribute))
	req.SetParam("appid", appid)
	req.SetParam("uid", uid)
	req.SetParam("channel", channel)
	req.SetParam("function", functions)
	req.SetParam("attribute", attributeStr)
	req.SetParam("tag", tag)

	start1 := time.Now()
	xsfCaller := xsf.NewCaller(auth.xsfc)
	xsfCaller.WithHashKey(appid)

	start := time.Now()
	xsfResp, errcode, err := xsfCaller.Call(config.SVC, op, req, auth.to)
	cost := time.Since(start)
	monitor.WithCost(op, cost)
	monitor.WithCallErr(op, errcode)

	if err != nil {
		tool.L.Errorw("licc-sdk | HasLicense error", "code", errcode, "error", err,
			"caller cost", start.Sub(start1).String(), "call cost", cost.String(), "tag", tag)
		return
	}

	dataMeta := xsfResp.GetData()
	if len(dataMeta) != 0 {
		logInfo = string(dataMeta[0].Data)
	}
	return xsfResp.GetAllParam(), logInfo, err
}
