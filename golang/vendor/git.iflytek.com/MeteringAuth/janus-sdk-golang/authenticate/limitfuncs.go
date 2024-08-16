package authenticate

import (
	"encoding/json"
	"strings"
	"sync"
	"time"

	xsf "git.xfyun.cn/AIaaS/xsf/client"
)

const (
	//OPGETLIMIT    = "glf"
	OPGETLIMITREG = "glfreg"
)

type FuncsInfo struct {
	funcs     map[string]bool
	wildFuncs map[string]bool
}
type LimitFuncsManager struct {
	sync.RWMutex
	limitFuncs map[string]*FuncsInfo
	Md5Map     map[string]string
	channel    []string
	xsfc       *xsf.Client
	caller     *xsf.Caller
	updateTime time.Duration
	p          xsfParam
}

type LimitResData struct {
	Channel string   `json:"channel"`
	Funcs   []string `json:"function"`
}

type xsfParam struct {
	rpcTimeout time.Duration
	svc        string
	op         string
}

func NewLimitFuncsManager(c []string, cfgName, cname string) (lfm *LimitFuncsManager, err error) {
	xsfClient, err := newXsfClient(cfgName, cname)
	if err != nil {
		return
	}
	updatetime, errUdt := xsfClient.Cfg().GetInt64(cname, "update_time")
	if errUdt != nil {
		updatetime = 1800 * 1e3
	}
	timeout, errTo := xsfClient.Cfg().GetInt64(cname, "timeout")
	if errTo != nil {
		timeout = 5000
	}

	svcName, err := xsfClient.Cfg().GetString(cname, "server_name")
	if err != nil {
		svcName = SVC
	}

	lfm = &LimitFuncsManager{
		channel:    c,
		xsfc:       xsfClient,
		updateTime: time.Duration(updatetime) * time.Millisecond,
		p: xsfParam{
			rpcTimeout: time.Duration(timeout) * time.Millisecond,
			svc:        svcName,
			op:         OPGETLIMITREG,
		},
		caller:     xsf.NewCaller(xsfClient),
		limitFuncs: make(map[string]*FuncsInfo),
		Md5Map:     make(map[string]string),
	}
	lfm.Run()
	return
}

func (l *LimitFuncsManager) Run() {
	for _, c := range l.channel {
		l.update(c)
	}
	go func() {
		t1 := time.NewTicker(l.updateTime)
		for {
			select {
			case <-t1.C:
				for _, c := range l.channel {
					l.update(c)
				}
			}
		}
	}()
}

func (l *LimitFuncsManager) update(c string) {
	/*拉取* s *xsf.Res, errcode int32, e error*/
	//l.xsfc.Log.Infof("update limitFuncs enter\n")
	defer un(enter("updateLimitFuncs", l.xsfc), l.xsfc)
	req := xsf.NewReq()
	md5 := l.Md5Map[c]
	l.xsfc.Log.Infof("update limitFuncs| cache md5 = %s", md5)

	req.SetParam("MD5", md5)
	req.SetParam("channel", c)
	req.SetParam("sid", "glfreg")
	xsfResp, errcode, err := l.caller.Call(l.p.svc, l.p.op, req, l.p.rpcTimeout)
	if err != nil {
		//todo metrics 出来之后，要做告警 log errcode
		l.xsfc.Log.Errorf("update limitFuncs | do call error , errcode = %d , err = %s", errcode, err)
		return
	}

	businCode := xsfResp.GetRes().GetCode()
	if businCode != 0 {
		l.xsfc.Log.Errorf("update limitFuncs | businCode = %s , err = %s", businCode, xsfResp.GetRes().GetErrorInfo())
		return
	}

	/*更新*/
	resMd5, ok := xsfResp.GetParam("MD5")
	l.xsfc.Log.Infof("update limitFuncs| real md5 = %s", resMd5)
	//fmt.Println("md5", resMd5)
	// 拉取资源失败或者md5未更新，直接返回
	if !ok || resMd5 == md5 {
		// log it
		l.xsfc.Log.Errorf("update limitFuncs | get nothing or md5 is same")
		return
	}
	dm := xsfResp.GetData()
	if len(dm) <= 0 {
		l.xsfc.Log.Errorf("update limitFuncs | lack of limit resouce in dataMeta")
		return
	}
	//fmt.Println("funcs", resMd5, respData, ok)
	l.xsfc.Log.Infof("update limitFuncs| real limtfuncs = %s", string(dm[0].Data))

	if ok {

		var lmr []LimitResData
		err := json.Unmarshal(dm[0].Data, &lmr)
		if err != nil {
			l.xsfc.Log.Errorf("update limitFuncs | parse limit resource data failed = %s", err)
			return
		}

		tm := make(map[string]*FuncsInfo)
		//resfuns := strings.Split(respData, ";")
		for _, vlmr := range lmr {
			// normal limit functions
			normalLimitFuncs := make(map[string]bool)
			// wild limit functions
			wildLimitFuncs := make(map[string]bool)
			for _, v := range vlmr.Funcs {
				// 通配符检测
				if strings.Contains("*", v) {
					wildLimitFuncs[v] = true
				} else {
					normalLimitFuncs[v] = true
				}
			}

			tm[vlmr.Channel] = &FuncsInfo{
				funcs:     normalLimitFuncs,
				wildFuncs: wildLimitFuncs,
			}
		}

		//清除数据，重新赋值，被删除的数据可被感知
		// 是否有性能问题？
		l.Lock()
		l.limitFuncs = make(map[string]*FuncsInfo)
		for k, v := range tm {
			l.limitFuncs[k] = v
		}
		l.Unlock()
		// 更新资源Md5
		l.Md5Map[c] = resMd5
	}
	return
}

/*过滤受限资源*/
func (l *LimitFuncsManager) filter(c string, funcs []string) (limitRes []string) {

	l.RLock()
	limitFuncsOfChannel, ok := l.limitFuncs[c]
	l.RUnlock()
	if !ok {
		return
	}

	for _, v := range funcs {
		// normal filter
		if _, ok := limitFuncsOfChannel.funcs[v]; ok {
			limitRes = append(limitRes, v)
		}
		// wild fiter
		for w := range limitFuncsOfChannel.wildFuncs {
			sameIndex := strings.Index(v, w[0:len(w)-1])
			if sameIndex == 0 { //首字母匹配到了
				limitRes = append(limitRes, v)
			}
		}
	}
	return
}

func enter(s string, xsfc *xsf.Client) string {
	xsfc.Log.Infof("enter : %s", s)
	return s
}

func un(s string, xsfc *xsf.Client) {
	xsfc.Log.Infof("leave : %s", s)
}
