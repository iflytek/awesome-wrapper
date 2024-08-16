package authenticate

import (
	"git.xfyun.cn/AIaaS/xsf/client"
	//"strings"

	"strconv"
	"time"
)

const (
	OPCHECK = "lic"
)

type AuthenticateCheck struct {
	//caller *xsf.Caller
	xsfc *xsf.Client
	p    xsfParam
}

type HasLicParam struct {
	appid   string
	funcs   string //分号分隔字符串
	channel string
	uid     string
}

func NewCheckLicManager(cfgName, cname string) (ac *AuthenticateCheck, err error) {
	xsfClient, err := newXsfClient(cfgName, cname)
	if err != nil {
		return
	}
	timeout, err_timeout := xsfClient.Cfg().GetInt64(cname, "timeout")
	if err_timeout != nil {
		timeout = 50
		//		return

	}

	coption, err:= xsfClient.Cfg().GetInt64(cname, "check_option")
	if err == nil {
		CheckOption = CtrlMode(uint32(coption))
	}

    svcName , err := xsfClient.Cfg().GetString(cname , "server_name")
    if err != nil {
        svcName = SVC
    }

	ac = &AuthenticateCheck{
		xsfc: xsfClient,
		p: xsfParam{
			rpcTimeout: time.Duration(timeout) * time.Millisecond,
			svc:        svcName,
			op:         OPCHECK,
		},
		//caller: xsf.NewCaller(xsfClient),
	}
	return
}

func (auth *AuthenticateCheck) HasLicense(appid, uid, channel, functions string, attribute CtrlMode) (authInfo map[string]string, logInfo string , err error) {
	req := xsf.NewReq()
	attributeStr := strconv.Itoa(int(attribute))
	req.SetParam("appid", appid)
	req.SetParam("uid", uid)
	req.SetParam("channel", channel)
	req.SetParam("function", functions)
	req.SetParam("attribute", attributeStr)
	req.SetParam("sid", "lic")

	xsfCaller := xsf.NewCaller(auth.xsfc)
	xsfCaller.WithHashKey(appid)
	xsfResp, errcode, err := xsfCaller.Call(auth.p.svc, auth.p.op, req, auth.p.rpcTimeout)
	if err != nil {
		//todo metrics 出来之后，要做告警 log errcode
		//fmt.Println(errcode)
		// TODO: log it
		auth.xsfc.Log.Errorf("HasLicense | errcode = %d, err=%s", errcode, err)
		return
	}
    dataMeta := xsfResp.GetData()
    if len(dataMeta) != 0 {
        logInfo = string(dataMeta[0].Data)
    }
	return xsfResp.GetAllParam(), logInfo , err
}
