package main

import (
	"flag"
	"fmt"
	"git.iflytek.com/AIaaS/xsf/client"
	"git.iflytek.com/AIaaS/xsf/utils"
	"log"
	"sync"
	"time"
)

func init() {
	utils.StartPProf("0.0.0.0","1998")
}
var (
	cfgUrl   = flag.String("u", "http://10.1.87.69:6868", "cfgUrl")
	cfgPrj   = flag.String("p", "xsf", "cfgPrj")
	cfgGroup = flag.String("g", "xsf", "cfgGroup")
	mode     = flag.Int("m", 0, "0:local,1:online")

	tm   = flag.Int64("tm", 1000, "timeout")
	gNum = flag.Int64("goroutines", 1, "total goroutines")
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

func main() {

	flag.Parse()

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

	//{
	//	baseTime := time.Now()
	//	res, _, err := xsf.NewCaller(cli).CallWithAddr("xsf-server", "req", *target, xsf.NewReq(), time.Duration(*tm)*time.Millisecond)
	//	if err != nil {
	//		log.Fatalf("callWithAddr tm:%v,dur:%v,err:%v\n", tm, time.Since(baseTime), err)
	//	}
	//	fmt.Printf("pre call res:%+v\n", res.GetAllParam())
	//}

	{
		var wg sync.WaitGroup
		var callersMatrix []*xsf.Caller
		var reqsMatrix []*xsf.Req
		for goIx := int64(0); goIx < *gNum; goIx++ {
			callersMatrix = append(callersMatrix, xsf.NewCaller(cli))
			reqsMatrix = append(reqsMatrix, xsf.NewReq())
		}
		for goIx := int64(0); goIx < *gNum; goIx++ {
			wg.Add(1)
			go func(goIx int) {
				defer wg.Done()
				var err error
				var baseTime time.Time
				for {
					baseTime = time.Now()
					//_, _, err = caller.CallWithAddr("xsf-server", "req", *target, req, time.Duration(*tm)*time.Millisecond)
					_, _, err = callersMatrix[goIx].Call("xsf-server", "req", reqsMatrix[goIx], time.Duration(*tm)*time.Millisecond)
					if err != nil {
						fmt.Printf("callWithAddr tm:%v,dur:%v,err:%v\n", time.Duration(*tm)*time.Millisecond, time.Since(baseTime), err)
					}
				}
			}(int(goIx))
		}
		fmt.Printf("%v goroutines already started!!!\n", *gNum)
		wg.Wait()
	}
}
