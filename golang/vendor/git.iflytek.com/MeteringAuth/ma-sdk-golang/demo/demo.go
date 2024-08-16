package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	xsf "git.iflytek.com/AIaaS/xsf/server"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/calc"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/licc"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/rep"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/rep/ver"
)

var cnt = flag.Int("l", 1, "loop times, -1 for infinite")
var thread = flag.Int("t", 1, "thread num")
var dur = flag.Int("d", 100, "sleep ms, 0 for no sleep")
var debug = flag.Bool("debug", true, "debug mode")

var appid = flag.String("appid", "testaqc", "appid")

// var appid = flag.String("ma", "whiteappid01", "appid")
var channel = flag.String("channel", "testaqcchannel", "licc init channel, split by ';', support regex")
var sub = flag.String("sub", "testaqcsub", "channel")
var ent = flag.String("ent", "testaqcfunc", "functions, split by ';'")
var uid = flag.String("uid", "testaqcuid", "uid/did")
var addr = flag.String("maddr", "myip", "repip")

var reportflag = flag.Bool("report", false, "report flag")
var reportnum = flag.Int("rnum", 0, "report num")
var acqIncNum = flag.Int("ainum", 0, "acq inc num")
var acqDecNum = flag.Int("adnum", 0, "acq dec num")
var asyncIncNum = flag.Int("asyncinum", 0, "async inc num")
var asyncDecNum = flag.Int("asyncdnum", 0, "async dec num")

var assignAsyncInc = flag.Int("assignasyncinum", -1, "assign async inc num")
var assignAsyncDec = flag.Int("assignasyncdnum", -1, "assign async dec num")

func println(a ...any) {
	if *debug {
		fmt.Println(a...)
	}
}

func main() {
	flag.Parse()

	// 初始化鉴权sdk（按需）
	if err := licc.Init(*xsf.CompanionUrl,
		*xsf.Project,
		*xsf.Group,
		*xsf.Service,
		ver.Version,
		*xsf.Mode,
		strings.Split(*channel, ";")); err != nil {
		log.Fatalln("licc init error:", err)
		return
	}

	// 初始化计量sdk（按需）
	//if err := calc.Init(*xsf.CompanionUrl,
	//	*xsf.Project,
	//	*xsf.Group,
	//	*xsf.Service,
	//	ver.Version,
	//	*xsf.Mode); err != nil {
	//	log.Fatal("calc init error:", err)
	//}

	// 初始化上报sdk（按需）
	if err := rep.Init(*xsf.CompanionUrl,
		*xsf.Project,
		*xsf.Group,
		*xsf.Service,
		ver.Version,
		*xsf.Mode,
		*addr); err != nil {
		log.Fatalln("report init error:", err)
	}

	funcs := strings.Split(*ent, ";")
	fmt.Println(funcs)

	all := float64(*thread * *cnt)
	println("all:", all)

	var wg sync.WaitGroup
	if *thread <= 0 {
		*thread = 1
	}
	wg.Add(*thread)
	start := time.Now()

	for tx := 0; tx < *thread; tx++ {
		go func() {
			defer wg.Done()
			for x := 0; x < *cnt || *cnt < 0; x++ {
				println(x + 1)
				// 用户级、设备级鉴权
				//result, logInfo, err = licc.Check(*appid, *uid, *sub, funcs, time.Now().String())
				//println("check result: ", result, " err: ", err, " logInfo: ", logInfo)

				// 阈值鉴权
				//result, err = licc.GetAcfLimits(*appid, *sub, funcs, time.Now().String())
				//println("check acf limit result: ", result, " err: ", err)

				// calc
				//for _, function := range funcs {
				//	// 计量
				//	code, err := calc.Calc(*appid, *sub, function, 1)
				//	println(*appid, *sub, function, 1, "->", code, err)
				//	if err != nil {
				//		println("Calc errocde:", code, " error:", err)
				//	}
				//
				//	// 设备级计量
				//	code, err = calc.CalcWithSubId(*appid, *uid, *sub, function, 1)
				//	println(*appid, *uid, *sub, function, 1, "->", code, err)
				//	if err != nil {
				//		println("CalcWithSubId errocde:", code, " error:", err)
				//	}
				//}
				fmt.Println("reportflag:", *reportflag, "reportnum:", *reportnum, "acqIncNum:", *acqIncNum, "acqDecNum:", *acqDecNum, "asyncIncNum:", *asyncIncNum, "asyncDecNum:", *asyncDecNum)
				if *reportflag == true {
					fmt.Println("reportnum:", *reportnum)
					// 上报并发
					a := make(map[string]uint, 10)
					a[*appid] = uint(*reportnum)
					//err := rep.Report(*channel, a, false)
					//println("report:", *channel, a, err)

					// 上报并发
					//addr := "addrInReq"
					err := rep.ReportWithAddr(*channel, *ent, a, *addr)
					println("report with addr:", *sub, a, addr, err)
				}
				time.Sleep(1 * time.Second)
				// 上报并发（cross专用）
				//b := make(map[string]string, 10)
				//b[*appid] = strconv.Itoa(x)
				//err = rep.Sync(*sub, *ent, b, nil, "req_sync", addr, false)
				//println("sync:", *sub, b, addr, err)

				//精细化并发
				for i := 0; i < *acqIncNum; i++ {
					rep.ConcInc(*appid, *channel, *ent, *addr)
				}
				for i := 0; i < *acqDecNum; i++ {
					rep.ConcDec(*appid, *channel, *ent, *addr)
				}
				//异步并发
				for i := 0; i < *asyncIncNum; i++ {
					rep.ConcIncWithRequestId(*appid, *channel, *ent, strconv.Itoa(i), 30)
				}
				for i := 0; i < *asyncDecNum; i++ {
					rep.ConcDecWithRequestId(*appid, *channel, *ent, strconv.Itoa(i))
				}
				if *assignAsyncInc != -1 {
					rep.ConcIncWithRequestId(*appid, *channel, *ent, strconv.Itoa(*assignAsyncInc), 30)
				}
				if *assignAsyncDec != -1 {
					rep.ConcDecWithRequestId(*appid, *channel, *ent, strconv.Itoa(*assignAsyncDec))
				}
				time.Sleep(2 * time.Second)
				// 鉴权
				result, logInfo, err := licc.Check(*appid, "", *channel, funcs, time.Now().String())
				println("check result: ", result, " err: ", err, " logInfo: ", logInfo)
				result, logInfo, err = licc.Check(*appid, "", *channel, funcs, time.Now().String())
				println("check result: ", result, " err: ", err, " logInfo: ", logInfo)
				result, logInfo, err = licc.Check(*appid, "", *channel, funcs, time.Now().String())
				println("check result: ", result, " err: ", err, " logInfo: ", logInfo)
				//rep.ConcInc(*appid, *channel, *ent, *addr)
				//time.Sleep(1 * time.Second)
				//result, logInfo, err = licc.Check(*appid, "", *channel, funcs, time.Now().String())
				//println("check result: ", result, " err: ", err, " logInfo: ", logInfo)

				if *dur > 0 {
					time.Sleep(time.Duration(*dur) * time.Millisecond)
				}
			}
		}()
	}
	wg.Wait()
	cost := time.Since(start)
	log.Println("cost:", cost, "qps:", all/float64(cost.Milliseconds())*1000)
	//result, logInfo, err := licc.Check(*appid, "", *channel, funcs, time.Now().String())
	//println("check result: ", result, " err: ", err, " logInfo: ", logInfo)
	// 卸载上报sdk
	rep.Fini()

	// 卸载计量sdk
	calc.Fini()

	// 卸载鉴权sdk
	licc.Fini()

	log.Println("demo done")
}
