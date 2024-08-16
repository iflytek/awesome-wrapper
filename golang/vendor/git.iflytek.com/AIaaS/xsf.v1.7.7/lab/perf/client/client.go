package main

import (
	"context"
	"flag"
	"fmt"
	"git.iflytek.com/AIaaS/xsf/client"
	"git.iflytek.com/AIaaS/xsf/utils"
	"github.com/cihub/seelog"
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"
	"time"
)

var (
	cfgUrl   = flag.String("u", "http://10.1.87.69:6868", "cfgUrl")
	cfgPrj   = flag.String("p", "xsf", "cfgPrj")
	cfgGroup = flag.String("g", "xsf", "cfgGroup")
	mode     = flag.Int("m", 0, "0:local,1:online")
	strategy = flag.Int("s", 0, "strategy 0:xrpc 1:janus")

	tm    = flag.Int64("tm", 1000, "timeout")
	delay = flag.Int64("delay", 0, "delay->unit:ms")
	gNum  = flag.Int64("goroutines", 1, "goroutine number")
	gCnt  = flag.Int64("count", 1, "total request per goroutine")
	gSize = flag.Int("size", 1, "message size:kb")
)
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")
var (
	stopFunc func()
)

const (
	cname = "xsf-client" //配置文件的主段名

	clientCfg = "client.toml"

	cfgService = "xsf-client" //服务发现的服务名
	cfgVersion = "0.0.0"      //配置文件的版本号

	cacheService = true
	cacheConfig  = true
	cachePath    = "./findercache" //配置缓存路径
)

func init() {

	flag.Parse()
	if logger, err := seelog.LoggerFromConfigAsString(`<seelog type="sync">
    <outputs formatid="main">
        <filter levels="trace,debug,info,warn,error,critical">
            <console/>
        </filter>
        <filter levels="trace,debug,info,warn,error,critical">
            <file path="log/cli.log"/>
        </filter>
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
		signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGPIPE)
		s := <-c
		switch s {
		case syscall.SIGINT, syscall.SIGKILL:
			{
				fmt.Printf("the program had received %v signal, will exit immediately -_-|||\n", s.String())
				stopFunc()
			}
		case syscall.SIGPIPE:
			{
				fmt.Printf("get broken pipe")
			}
		}
	}()
	//utils.StartPProf("", "8080")
}
func main() {

	flag.Parse()
	if *cpuprofile != "" {
		fmt.Printf("cpuprofile:%v\n", *cpuprofile)
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil { //监控cpu
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}
	{
		//单客户端测试
		cli, cliErr := xsf.InitClient(
			cname,
			func() utils.CfgMode {
				if *mode == 1 {
					return utils.Centre
				} else if *mode == 0 {
					return utils.Native
				}
				panic("illegal mode")
			}(),
			utils.WithCfgCacheService(cacheService),
			utils.WithCfgCacheConfig(cacheConfig),
			utils.WithCfgCachePath(cachePath),
			utils.WithCfgName(clientCfg),
			utils.WithCfgURL(*cfgUrl),
			utils.WithCfgPrj(*cfgPrj),
			utils.WithCfgGroup(*cfgGroup),
			utils.WithCfgService(cfgService),
			utils.WithCfgVersion(cfgVersion),
		)
		if cliErr != nil {
			log.Fatal("main | InitCient error:", cliErr)
		}

		ctx, cancel := context.WithCancel(context.Background())
		stopFunc = cancel

		switch *strategy {
		case 0:
			performanceXrpc(ctx, cli, time.Millisecond*time.Duration(*tm))
		case 1:
			performanceJanus(ctx, cli, time.Millisecond*time.Duration(*tm))
		case 2:
			performanceHystrix(ctx, cli, time.Millisecond*time.Duration(*tm))
		default:
			panic("illegal strategy")
		}
	}
	if *memprofile != "" {
		fmt.Printf("memprofile:%v\n", *memprofile)
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		runtime.GC()                                      // GC，获取最新的数据信息
		if err := pprof.WriteHeapProfile(f); err != nil { // 写入内存信息
			log.Fatal("could not write memory profile: ", err)
		}
		f.Close()
	}
}
