package lmt

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	xsf "git.iflytek.com/AIaaS/xsf/client"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/licc/monitor"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/ma/config"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/ma/tool"
)

const (
	op = "glfreg"
)

type FuncsInfo struct {
	funcs     map[string]bool
	wildFuncs map[string]bool
}

func (f *FuncsInfo) String() string {
	return fmt.Sprintf("f:%v,wf:%v", f.funcs, f.wildFuncs)
}

type LimitFuncsManager struct {
	sync.RWMutex
	limitFuncsR map[string]*FuncsInfo
	limitFuncsW map[string]map[string]*FuncsInfo
	Md5Map      map[string]string
	channel     []string
	// xsfc       *xsf.Client
	caller     *xsf.Caller
	updateTime time.Duration
	to         time.Duration
}

func (c *LimitFuncsManager) String() string {
	return fmt.Sprintf("limit client: sub: %v, up: %v, dur: %v", c.channel, c.updateTime.String(), c.to.String())
}

type LimitResData struct {
	Channel string   `json:"channel"`
	Funcs   []string `json:"function"`
}

func NewLimitFuncsManager(c []string) (lfm *LimitFuncsManager, err error) {
	cname := config.CliLmtTag
	xsfClient, err := tool.NewClient(cname)
	if err != nil {
		return
	}
	updatetime, errUdt := xsfClient.Cfg().GetInt64(cname, config.CliLmtUpTag)
	if errUdt != nil {
		updatetime = 10 * 60 * 1e3
	}

	timeout, errTo := xsfClient.Cfg().GetInt64(cname, config.CliDurTag)
	if errTo != nil {
		timeout = 5000
	}

	lfm = &LimitFuncsManager{
		channel:     c,
		updateTime:  time.Duration(updatetime) * time.Millisecond,
		to:          time.Duration(timeout) * time.Millisecond,
		caller:      xsf.NewCaller(xsfClient),
		limitFuncsR: make(map[string]*FuncsInfo),
		limitFuncsW: make(map[string]map[string]*FuncsInfo),
		Md5Map:      make(map[string]string),
	}

	tool.LiccPrinter.Println("lmt updatetime", lfm.updateTime.String())
	tool.LiccPrinter.Println("lmt timeout", lfm.to.String())

	lfm.Run()
	return
}

func (l *LimitFuncsManager) Run() {
	for _, c := range l.channel {
		l.update(c)
	}
	l.refresh()

	go func() {
		t1 := time.NewTicker(l.updateTime)
		for {
			<-t1.C
			for _, c := range l.channel {
				l.update(c)
			}
			l.refresh()
		}
	}()
}

func (l *LimitFuncsManager) refresh() {
	l.Lock()
	l.limitFuncsR = make(map[string]*FuncsInfo)
	for c, m := range l.limitFuncsW {
		for k, v := range m {
			l.limitFuncsR[k] = v
			if config.C.Log.Level == "debug" {
				tool.L.Debugw("licc-sdk | refresh", "query", c, "channel", k, "function", v)
			}
		}
	}
	l.Unlock()
}

func (l *LimitFuncsManager) update(c string) {
	req := xsf.NewReq()
	md5 := l.Md5Map[c]
	tool.L.Infow("licc-sdk | update lmt begin", "channel", c, "md5", md5)

	req.SetParam("MD5", md5)
	req.SetParam("channel", c)
	// req.SetParam("sid", op)

	start := time.Now()
	xsfResp, errcode, err := l.caller.Call(config.SVC, op, req, l.to)
	cost := time.Since(start)
	monitor.WithCost(op, cost)
	monitor.WithCallErr(op, errcode)

	if err != nil {
		tool.L.Errorw("licc-sdk | update lmt call error", "code", errcode, "error", err)
		return
	}

	resMd5, ok := xsfResp.GetParam("MD5")
	tool.L.Infow("licc-sdk | update lmt", "get md5", resMd5, "ok", ok)
	if !ok || resMd5 == md5 {
		tool.L.Debugw("licc-sdk | update lmt | get nothing or md5 is same")
		return
	}

	dm := xsfResp.GetData()
	if len(dm) <= 0 {
		tool.L.Errorw("licc-sdk | update lmt | lack of limit resouce in dataMeta")
		return
	}
	tool.L.Debugw("licc-sdk | update lmt", "limtfuncs", string(dm[0].Data))

	var lmr []LimitResData
	if err = json.Unmarshal(dm[0].Data, &lmr); err != nil {
		tool.L.Errorw("licc-sdk | update lmt | resp json unmarshal error", "error", err, "data", string(dm[0].Data))
		return
	}

	delete(l.limitFuncsW, c)
	tm := make(map[string]*FuncsInfo)
	for _, vlmr := range lmr {
		// normal limit functions
		normalLimitFuncs := make(map[string]bool)
		// wild limit functions
		wildLimitFuncs := make(map[string]bool)
		for _, v := range vlmr.Funcs {
			// 通配符检测
			if strings.Contains(v, "*") {
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

	l.limitFuncsW[c] = tm
	// 更新资源Md5
	l.Md5Map[c] = resMd5
}

/*过滤受限资源*/
func (l *LimitFuncsManager) Filter(c string, funcs []string) (limitRes []string) {
	l.RLock()
	limitFuncsOfChannel, ok := l.limitFuncsR[c]
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
