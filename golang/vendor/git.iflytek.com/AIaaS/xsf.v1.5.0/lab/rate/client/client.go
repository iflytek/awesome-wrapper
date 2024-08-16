package main

import (
	"flag"
	"fmt"
	"git.iflytek.com/AIaaS/xsf/client"
	"git.iflytek.com/AIaaS/xsf/utils"
	"github.com/cihub/seelog"
	"log"
	"math"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	clientCfg  = "client.toml"
	cfgUrl     = "http://10.1.87.70:6868"
	cfgPrj     = "xsf"
	cfgGroup   = "xsf"
	cname      = "xsf-client" //配置文件的主段名
	cfgService = "xsf-client" //服务发现的服务名
	cfgVersion = "2.0.0"      //配置文件的版本号
	apiVersion = "1.0.0"      //api版本号，一般不用修改

	cacheService = true
	cacheConfig  = true
	cachePath    = "./findercache" //配置缓存路径
)

var (
	Goroutines = flag.Int64("g", 1, "goroutines")
	Count      = flag.Int64("c", 1, "count")
	Tm         = flag.Int64("tm", 1000, "timeout")
	mode       = flag.Int64("mode", 0, "0:native;1:center")
	retry      = flag.Int("retry", 0, "retry")
)
var (
	ssbFail int64
)

var ssbMax int64 = math.MinInt64
var ssbMin int64 = math.MaxInt64
var ssbAllTime int64 = 0
var ssbCnt int64 = 0
var countTmp = int64(0)
var cliRetryCnt int64

func init() {
	flag.Parse()
	if logger, err := seelog.LoggerFromConfigAsString(`<seelog type="sync">
    <outputs>
        <splitter formatid="main">
            <filter levels="trace,debug,info,warn,error,critical">
                <console/>
            </filter>
        </splitter>
        <splitter formatid="main">
            <filter levels="trace">
                <file path="trace.log"/>
            </filter>
            <filter levels="debug,info,warn,error,critical">
                <file path="log/record.log"/>
            </filter>
        </splitter>
    </outputs>
    <formats>
        <format id="main" format="%Msg"/>
    </formats>
</seelog>`); err != nil {
		log.Fatal(err)
	} else {
		seelog.ReplaceLogger(logger)
	}
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGKILL) //不同的操作系统信号可能有差异. 不过syscall.SIGINT和syscall.SIGKILL各个系统是一致的, 分别对应os.Interrupt和os.Kill.
		s := <-c
		switch s {
		case syscall.SIGINT, syscall.SIGKILL:
			{
				fmt.Printf("the program had received %v signal, will exit immediately -_-|||", s.String())
				seelog.Flush()
				os.Exit(1)
			}
		case syscall.SIGPIPE:
			{
				fmt.Printf("get broken pipe")
			}
		}
	}()
}

func ssb(c *xsf.Caller, tm time.Duration) (*xsf.Res, string, int32, error) {
	req := xsf.NewReq()
	ssbBase := time.Now()

	res, code, e := c.SessionCall(xsf.CREATE, "sms", "ssb", req, tm)
	//if e != nil || code != 0 {
	//	for i := 0; i < *cliRetry; i++ {
	//		atomic.AddInt64(&cliRetryCnt, 1)
	//		res, code, e = c.SessionCall(xsf.CREATE, "sms", "ssb", req, tm)
	//		if code == 0 && e == nil {
	//			break
	//		}
	//	}
	//}
	ssbDur := time.Since(ssbBase).Nanoseconds()

	if atomic.LoadInt64(&ssbMax) < ssbDur {
		atomic.StoreInt64(&ssbMax, ssbDur)
	}
	if atomic.LoadInt64(&ssbMin) > ssbDur {
		atomic.StoreInt64(&ssbMin, ssbDur)
	}
	atomic.AddInt64(&ssbAllTime, ssbDur)
	atomic.AddInt64(&ssbCnt, 1)
	if e != nil {
		seelog.Infof("NO.%d ssb error:%v,ssbDur:%vms\n", atomic.LoadInt64(&countTmp), e, ssbDur/1e6)
	}

	var sess string
	if e == nil {
		sess = res.Session()
	}
	return res, sess, code, e
}

func auw(c *xsf.Caller, sess string, tm time.Duration) (*xsf.Res, int32, error) {
	req := xsf.NewReq()
	req.Session(sess)
	baseTime := time.Now()
	res, code, e := c.SessionCall(xsf.CONTINUE, "sms", "auw", req, tm)
	dur := time.Now().Sub(baseTime)
	if e != nil {
		seelog.Infof("dur:%v,auw error:%v", dur, e)
	}

	return res, code, e
}

func sse(c *xsf.Caller, sess string, tm time.Duration) (*xsf.Res, int32, error) {
	req := xsf.NewReq()
	req.Session(sess)
	res, code, e := c.SessionCall(xsf.CONTINUE, "sms", "sse", req, tm)
	if e != nil {
		seelog.Info("sse error: ", e)
	}
	return res, code, e
}

func session(goroutines, count int64) {
	var cli *xsf.Client
	var e error

	switch *mode {
	case 0:
		{
			fmt.Println("about to init native client")
			cli, e = xsf.InitClient(
				cname,
				utils.Native,
				utils.WithCfgCacheService(cacheService),
				utils.WithCfgCacheConfig(cacheConfig),
				utils.WithCfgCachePath(cachePath),
				utils.WithCfgName(clientCfg),
				utils.WithCfgURL(cfgUrl),
				utils.WithCfgPrj(cfgPrj),
				utils.WithCfgGroup(cfgGroup),
				utils.WithCfgService(cfgService),
				utils.WithCfgVersion(cfgVersion),
			)
		}
	case 1:
		{
			fmt.Println("about to init centre client")
			cli, e = xsf.InitClient(
				cname,
				utils.Centre,
				utils.WithCfgCacheService(cacheService),
				utils.WithCfgCacheConfig(cacheConfig),
				utils.WithCfgCachePath(cachePath),
				utils.WithCfgName(clientCfg),
				utils.WithCfgURL(cfgUrl),
				utils.WithCfgPrj(cfgPrj),
				utils.WithCfgGroup(cfgGroup),
				utils.WithCfgService(cfgService),
				utils.WithCfgVersion(cfgVersion),
			)
		}
	}
	if e != nil {
		log.Fatal("main | InitCient error:", e)
	}

	tm := time.Millisecond * time.Duration(*Tm)
	var wg sync.WaitGroup
	for ix := int64(0); ix < goroutines; ix++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var sess string
			var code int32
			c := xsf.NewCaller(cli)

			c.WithApiVersion(apiVersion)

			c.WithLBParams("xsf-lbv2", "iat", nil)
			c.WithRetry(*retry)
			for atomic.LoadInt64(&countTmp) < atomic.LoadInt64(&count) {
				atomic.AddInt64(&countTmp, 1)
				ssbTs := time.Now()
				_, sess, code, e = ssb(c, tm)
				ssbD := time.Since(ssbTs)
				if e != nil {
					seelog.Errorf("NO.%d,ts:%d,ssbErr:sess:%s,code:%v,err:%v,ssbDur:%v\n", atomic.LoadInt64(&countTmp), time.Now().UnixNano(), sess, code, e, ssbD)
					atomic.AddInt64(&ssbFail, 1)
					continue
				}
				_, code, e = auw(c, sess, tm)
				if e != nil {
					seelog.Errorf("auwErr:sess:%s,code:%v,err:%v\n", sess, code, e)
					continue
				}
				_, code, e = sse(c, sess, tm)
				if e != nil {
					seelog.Errorf("sseErr:sess:%s,code:%v,err:%v\n", sess, code, e)
					continue
				}
			}
		}()
	}
	wg.Wait()

}
func main() {
	baseTime := time.Now()
	seelog.Criticalf("-------------Mission session start:%v---------------\n", baseTime)
	seelog.Criticalf("retry:%v\n", *retry)
	seelog.Criticalf("gotoutines:%v,count:%v,mode:%v\n", *Goroutines, *Count, *mode)
	session(*Goroutines, *Count)
	seelog.Criticalf("ssbAllTime:%vms,ssbMin:%vms,ssbMax:%vms,ssAvg:%vms\n", ssbAllTime/1e6, ssbMin/1e6, ssbMax/1e6, (float64(ssbAllTime)/float64(ssbCnt))/1e6)
	seelog.Criticalf("TotalGoroutines:%v,totalSsbOp:%v,failSsbOp:%v,successRate:%v,cliRetryCnt:%v\n", *Goroutines, ssbCnt, ssbFail, float64(ssbCnt-ssbFail)/float64(ssbCnt), cliRetryCnt)
	dur := time.Since(baseTime)
	seelog.Criticalf("-------------Mission session complete:%v---------------", dur)
	seelog.Criticalf("\n\n\n=====我是分割线======\n\n\n")

	time.Sleep(time.Minute)
}
