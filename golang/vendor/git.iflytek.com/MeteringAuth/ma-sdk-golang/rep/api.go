package rep

import (
	"errors"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/ma/syncproto"
	"github.com/sirupsen/logrus"
	"time"

	"git.iflytek.com/MeteringAuth/ma-sdk-golang/ma/config"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/ma/tool"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/rep/monitor"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/rep/ver"
)

var (
	r       *ReportManager
	sm      *AqcCounterSyncManager
	asm     *AsyncCounterSyncManager
	started bool
	inited  bool
)

func Init(url, pro, gro, service, version string, mode int, addr string) (err error) {
	if inited {
		return errors.New("already inited")
	}
	inited = true

	if addr == "" {
		return errors.New("empty addr")
	}

	tool.RepPrinter.Println("version:", ver.Version)
	tool.RepPrinter.Println("config:", url, pro, gro, service, version, mode)

	if err := tool.Init(url, pro, gro, service, version, mode, 0); err != nil {
		return err
	}

	if config.C.Metrics.Able == 1 {
		err = monitor.Init()
		if err != nil {
			return
		}
	}

	asyncInit := config.C.Rep.AsyncInit
	retryTimes := config.C.Rep.InitRetry
	if asyncInit {
		tool.RepPrinter.Println("async init")
		go func() {
			lazyInit(url, pro, gro, service, version, mode, addr, retryTimes)
		}()
	} else {
		tool.RepPrinter.Println("sync init")
		err = lazyInit(url, pro, gro, service, version, mode, addr, retryTimes)
	}

	return
}

func lazyInit(url, pro, gro, service, version string, mode int, addr string, retry int) (err error) {
	err = initOnce(url, pro, gro, service, version, mode, addr)
	for i := 1; i < retry && err != nil; i++ {
		err = initOnce(url, pro, gro, service, version, mode, addr)
	}
	if err != nil {
		tool.RepPrinter.Println("failed to init, retry times:", retry, "error:", err)
	} else {
		tool.RepPrinter.Println("init done")
		tool.L.Infow("rep-sdk | init success")
		started = true
	}
	return
}

func initOnce(url, pro, gro, service, version string, mode int, addr string) (err error) {
	cname := config.CliRepTag
	xsfClient, err := tool.NewClient(cname)
	if err != nil {
		return
	}

	timeout, err := xsfClient.Cfg().GetInt64(cname, config.CliDurTag)
	if err != nil {
		timeout = 50
	}

	r = &ReportManager{
		xsfc: xsfClient,
		to:   time.Duration(timeout) * time.Millisecond,
		addr: addr,
	}

	sm = NewAqcCounterSyncManager(config.C.Conc.BatchMaxSize, config.C.Conc.BufferSize, xsfClient, timeout, addr, config.C.Conc.Worker)
	asm = NewAsyncCounterSyncManager(config.C.Conc.BatchMaxSize, config.C.Conc.BufferSize, xsfClient, timeout, addr)

	tool.RepPrinter.Println("timeout", r.to.String())

	return
}

func Fini() {
	started = false
	r.fini()
}

/*appid,auth count*/
func Report(channel string, concInfo map[string]uint, aqcreport bool) (err error) {
	if !started {
		return errors.New("Report have not inited")
	}

	start := time.Now()
	r.report(channel, "", concInfo, aqcreport)
	cost := time.Since(start)
	monitor.WithCost("rep", cost)
	return
}

func ReportEx(channel string, ent string, concInfo map[string]uint) (err error) {
	if !started {
		return errors.New("Report have not inited")
	}

	start := time.Now()
	r.report(channel, ent, concInfo, false)
	cost := time.Since(start)
	monitor.WithCost("rep", cost)
	return
}

// for cross
func Sync(channel, ent string, concInfo map[string]string, protodata []byte, op, endpoint string, aqcreport bool) (err error) {
	if !started {
		return errors.New("Report have not inited")
	}

	start := time.Now()
	if op == "aqc_count" {
		sm.sync(protodata)
	} else if op == "async_count" {
		asm.sync(protodata)
	} else {
		r.sync(concInfo, channel, ent, endpoint, aqcreport)
	}
	cost := time.Since(start)
	monitor.WithCost("sync", cost)
	return
}

func ReportWithAddr(channel, ent string, concInfo map[string]uint, addr string) (err error) {
	if !started {
		return errors.New("Report have not inited")
	}

	start := time.Now()
	r.reportWithAddr(channel, ent, concInfo, addr, config.C.Conc.OnlyUseAqc)
	cost := time.Since(start)
	monitor.WithCost("repAddr", cost)
	return
}

// 会在doSession之前并发+1，
// doSession执⾏完成后，并发-1
// doSession: ⼀路会话执⾏函数
// 精确并发控制 +1
func ConcInc(appid, channel, function, addr string) {
	if tool.IN(appid, config.C.Conc.WhiteAppidList) {
		monitor.WithCommonCounter(AqcOp, monitor.WhiteSkip)
		return
	}
	//xsfreq.SetOp("aqc_inc")
	//xsfreq.SetParam([app_id,channel,func,addr])
	sm.Add(aQcSyncCounterKey{
		counterKey: counterKey{
			appId:    appid,
			channel:  channel,
			function: function,
		},
		addr: addr,
	}, 1)
	return
}

// 精确并发控制 -1
func ConcDec(appid, channel, function, addr string) {
	if tool.IN(appid, config.C.Conc.WhiteAppidList) {
		monitor.WithCommonCounter(AqcOp, monitor.WhiteSkip)
		return
	}
	sm.Add(aQcSyncCounterKey{
		counterKey: counterKey{
			appId:    appid,
			channel:  channel,
			function: function,
		},
		addr: addr,
	}, -1)
	return
}

// 异步并发控制 并发数+1。如果expireSeconds 之后没有调⽤ConcDecWithRequestId 将对应的
// requestId -1.那么系统会⾃动将并发数-1
// expireSecond 最⼤ 100000s
func ConcIncWithRequestId(appid, channel, function, requestId string, expireSecond int) {
	if expireSecond < 0 {
		logrus.Errorf("rep-sdk | ConcIncWithRequestId | expireSecond is less than 0")
		return
	}
	asm.Add(asyncCountMessage{
		counterKey: counterKey{
			appId:    appid,
			channel:  channel,
			function: function,
		},
		requestId: requestId,
		expire:    expireSecond,
		op:        syncproto.AsyncOp_inc,
	})
}

// 异步并发控制 并发数 -1
func ConcDecWithRequestId(appid, channel, function, requestId string) {
	asm.Add(asyncCountMessage{
		counterKey: counterKey{
			appId:    appid,
			channel:  channel,
			function: function,
		},
		requestId: requestId,
		op:        syncproto.AsyncOp_dec,
	})
}
