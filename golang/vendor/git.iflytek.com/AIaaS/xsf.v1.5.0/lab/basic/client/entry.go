package main

import (
	"flag"
	"fmt"
	"git.iflytek.com/AIaaS/xsf/client"
	"git.iflytek.com/AIaaS/xsf/utils"
	"log"
	"time"
)

const (
	cname = "xsf-client" //配置文件的主段名

	clientCfg  = "client.toml"
	cfgUrl     = "http://10.1.87.69:6868"
	cfgPrj     = "xsf"
	cfgGroup   = "xsf"
	cfgService = "xsf-client" //服务发现的服务名
	cfgVersion = "0.0.0"      //配置文件的版本号
	apiVersion = "1.0.0"      //api版本号，一般不用修改

	cacheService = true
	cacheConfig  = true
	cachePath    = "./findercache" //配置缓存路径
)

var (
	tm   = flag.Int64("tm", 1000, "timeout")
	mode = flag.Int64("mode", 0, "0:native;1:center")
	c    = flag.Int64("c", 350, "concurrent")
)

func main() {
	flag.Parse()
	cli, cliErr := xsf.InitClient(
		cname,
		func() xsf.CfgMode {
			switch *mode {
			case 0:
				{
					fmt.Println("about to init native client")
					return utils.Native
				}
			case 2:
				{
					{
						fmt.Println("about to init custom client")
						return utils.Custom
					}
				}
			default:
				{
					fmt.Println("about to init centre client")
					return utils.Centre
				}
			}
		}(),
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
	if cliErr != nil {
		log.Fatal("main | InitCient error:", cliErr)
	}
	{
		//for {
		//	fmt.Println(cli.Cfg().GetRawCfg())
		//	fmt.Println(cli.Cfg().GetString("log", "level"))
		//	fmt.Println(time.Now(), "---------------------------------------------------")
		//	time.Sleep(time.Second)
		//}
	}
	//{
	//	fmt.Println(cli.Cfg().GetRawCfg())
	//	//配置读取测试
	//	val, err := cli.Cfg().GetString("trace", "2ip")
	//	if err != nil {
	//		panic(err)
	//	} else {
	//		panic(val)
	//	}
	//}
	//callTest(xsf.NewCaller(cli), time.Millisecond*time.Duration(*tm))
	//callConHashTest(xsf.NewCaller(cli), time.Millisecond*time.Duration(*tm))
	callWithAddr(xsf.NewCaller(cli), time.Second*3)
	//callTest(xsf.NewCaller(cli), time.Millisecond*time.Duration(*tm))
	//var wg sync.WaitGroup
	//for i := 0; i < int(*c); i++ {
	//	i := i
	//	wg.Add(1)
	//	go func() {
	//		defer wg.Done()
	//		fmt.Println(i, "->->->->->->->->->->->->->->->")
	//		for {
	//			//time.Sleep(40 * time.Millisecond)
	//			//callExample(xsf.NewCaller(cli), time.Millisecond*time.Duration(*tm))
	//			sessionCallExample(xsf.NewCaller(cli), time.Millisecond*time.Duration(*tm))
	//		}
	//	}()
	//}
	//wg.Wait()

	//callExample(xsf.NewCaller(cli), time.Millisecond*time.Duration(*tm))

	/*
		callExample(xsf.NewCaller(cli), time.Millisecond*time.Duration(*tm))
		sessionCallExample(xsf.NewCaller(cli), time.Millisecond*time.Duration(*tm))
		sessionCallExample(xsf.NewCaller(cli), time.Millisecond*time.Duration(*tm))
	*/
	//sessionCallExample(xsf.NewCaller(cli), time.Millisecond*time.Duration(*tm))
	//crashTest(xsf.NewCaller(cli))
	//callExample(xsf.NewCaller(cli), time.Millisecond*time.Duration(*tm))

	//var cnt int32
	//for {
	//	fmt.Println("NO.", atomic.AddInt32(&cnt, 1), "-----------------")
	//	//sessionCallExample(xsf.NewCaller(cli), time.Millisecond*time.Duration(*tm))
	//	//sessionCallWithOneShortFlag(xsf.NewCaller(cli), time.Millisecond*time.Duration(*tm))
	//	//pingTest(xsf.NewCaller(cli), time.Millisecond*time.Duration(*tm))
	//	//callExample(xsf.NewCaller(cli), time.Millisecond*time.Duration(*tm))
	//	//time.Sleep(time.Minute * 5)
	//	time.Sleep(time.Second * 1)
	//	//break
	//}
}
