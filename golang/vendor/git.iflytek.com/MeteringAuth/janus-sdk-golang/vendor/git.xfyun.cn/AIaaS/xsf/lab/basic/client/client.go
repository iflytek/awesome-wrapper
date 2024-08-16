package main

import (
	"bytes"
	"flag"
	"fmt"
	"git.xfyun.cn/AIaaS/xsf/client"
	"git.xfyun.cn/AIaaS/xsf/utils"
	"log"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

const (
	cname = "xsf-client" //配置文件的主段名

	clientCfg  = "client.toml"
	cfgUrl     = "http://10.1.87.69:6868"
	cfgPrj     = "metrics"
	cfgGroup   = "3s"
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
)

func ssb(c *xsf.Caller, tm time.Duration) (*xsf.Res, string, int32, error) {
	req := xsf.NewReq()
	req.SetParam("k1", "v1")
	req.SetParam("k2", "v2")
	req.SetParam("k3", "v3")
	req.SetParam("directEngIp", "this is directEngIp")
	//c.WithLBParams("lbname", "xxx", nil)
	res, code, e := c.SessionCall(xsf.CREATE, "xsf-server", "ssb", req, tm)
	if code != 0 || e != nil {
		log.Fatal("ssb err")
	}

	var sess string
	if e == nil {
		sess = res.Session()
	}
	return res, sess, code, e
}

func auw(c *xsf.Caller, sess string, tm time.Duration) (*xsf.Res, int32, error) {
	req := xsf.NewReq()
	req.SetParam("k1", "v1")
	req.SetParam("k2", "v2")
	req.SetParam("k3", "v3")

	req.Session(sess)

	res, code, e := c.SessionCall(xsf.CONTINUE, "xsf-server", "auw", req, tm)
	res.Session()
	if code != 0 || e != nil {
		log.Fatal("auw err")
	}

	return res, code, e
}

func sse(c *xsf.Caller, sess string, tm time.Duration) (*xsf.Res, int32, error) {
	req := xsf.NewReq()
	req.SetParam("k1", "v1")
	req.SetParam("k2", "v2")
	req.SetParam("k3", "v3")

	res, code, e := c.SessionCall(xsf.CONTINUE, "xsf-server", "sse", req, tm)
	if code != 0 || e != nil {
		log.Fatal("sse err")
	}
	return res, code, e
}

func sessionCallExample(c *xsf.Caller, tm time.Duration) {

	c.WithApiVersion(apiVersion)

	_, sess, _, _ := ssb(c, tm)
	_, _, _ = auw(c, sess, tm)
	_, _, _ = sse(c, sess, tm)

}
func callExample(c *xsf.Caller, tm time.Duration) {
	c.WithRetry(1)
	span := utils.NewSpan(utils.CliSpan).Start()
	defer span.End().Flush()

	span = span.WithName("callExample")
	span = span.WithTag("customKey1", "customVal1")
	span = span.WithTag("customKey2", "customVal2")
	span = span.WithTag("customKey3", "customVal3")

	c.WithApiVersion(apiVersion)
	c.WithRetry(3)

	req := xsf.NewReq()
	req.SetParam("k1", "v1")
	req.SetParam("k2", "v2")
	req.SetParam("k3", "v3")

	req.SetTraceID(span.Meta()) //将span信息带到后端

	_, code, e := c.Call("xsf-server", "req", req, tm)
	if code != 0 || e != nil {
		log.Fatal("sse err", code, e)
	}
}
func callTest(c *xsf.Caller, tm time.Duration) {
	span := utils.NewSpan(utils.CliSpan).Start()
	defer span.End().Flush()

	span = span.WithName("callExample")
	span = span.WithTag("customKey1", "customVal1")
	span = span.WithTag("customKey2", "customVal2")
	span = span.WithTag("customKey3", "customVal3")

	c.WithApiVersion(apiVersion)
	c.WithRetry(3)
	c.WithLBParams("xsf-lbv2", "iat", nil)

	{ //第一组测试
		req := xsf.NewReq()
		req.SetParam("k1", "v1")
		req.SetParam("k2", "v2")
		req.SetParam("k3", "v3")

		req.SetTraceID(span.Meta()) //将span信息带到后端
		baseTime := time.Now()
		res, code, e := c.Call("xsf-server", "req", req, tm)
		dur := time.Now().Sub(baseTime)
		if code != 0 || e != nil {
			log.Fatalf("sse err,code:%v,e:%v,dur:%v\n", code, e, dur.Seconds())
		} else {
			fmt.Printf("F.NO.1 => handle:%s,dur:%vs\n", res.Handle(), dur.Seconds())
		}

		res, code, e = c.Call("xsf-server", "req", req, tm)
		if code != 0 || e != nil {
			log.Fatal("sse err", code, e)
		} else {
			fmt.Printf("F.NO.2 => handle:%s\n", res.Handle())
		}

		res, code, e = c.Call("xsf-server", "req", req, tm)
		if code != 0 || e != nil {
			log.Fatal("sse err", code, e)
		} else {
			fmt.Printf("F.NO.3 => handle:%s\n", res.Handle())
		}
	}

	//{ //第二组测试
	//	c.WithHashKey("555")
	//	req := xsf.NewReq()
	//	req.SetParam("k1", "v1")
	//	req.SetParam("k2", "v2")
	//	req.SetParam("k3", "v3")
	//
	//	req.SetTraceID(span.Meta()) //将span信息带到后端
	//
	//	res, code, e := c.Call("xsf-server", "req", req, tm)
	//	if code != 0 || e != nil {
	//		log.Fatal("sse err", code, e)
	//	} else {
	//		fmt.Printf("S.NO.1 => handle:%s\n", res.Handle())
	//	}
	//
	//	res, code, e = c.Call("xsf-server", "req", req, tm)
	//	if code != 0 || e != nil {
	//		log.Fatal("sse err", code, e)
	//	} else {
	//		fmt.Printf("S.NO.2 => handle:%s\n", res.Handle())
	//	}
	//
	//	res, code, e = c.Call("xsf-server", "req", req, tm)
	//	if code != 0 || e != nil {
	//		log.Fatal("sse err", code, e)
	//	} else {
	//		fmt.Printf("S.NO.3 => handle:%s\n", res.Handle())
	//	}
	//}
}
func callWithAddr(c *xsf.Caller, tm time.Duration) {
	baseTime := time.Now()

	req := xsf.NewReq()
	data := xsf.NewData()
	data.Append(bytes.Repeat([]byte("b"), 1e8))
	req.AppendData(data)

	r, code, e := c.CallWithAddr("", "req", "127.0.0.1:1234", req, tm)
	fmt.Printf("dur:%v,r:%v,code:%v,e:%v\n", time.Now().Sub(baseTime).String(), r, code, e)
}
func callConHashTest(c *xsf.Caller, tm time.Duration) {
	span := utils.NewSpan(utils.SrvSpan).Start()
	defer span.End().Flush()

	span = span.WithName("callExample")
	span = span.WithTag("customKey1", "customVal1")
	span = span.WithTag("customKey2", "customVal2")
	span = span.WithTag("customKey3", "customVal3")

	c.WithApiVersion(apiVersion)
	c.WithRetry(3)
	var count int64

	test := func(hashKey, svc string) {
		baseTime := time.Now()
		addr, addrErr := c.GetHashAddr(hashKey, svc)
		fmt.Printf("NO.%v dur:%v,addr:%v,addrErr:%v,hashKey:%v,svc:%v\n",
			time.Now().Sub(baseTime).String(), atomic.AddInt64(&count, 1), addr, addrErr, hashKey, svc)
	}
	{
		fmt.Println("------------------------------")
		hashKey, svc := "111", "xsf-server"
		test(hashKey, svc)
		test(hashKey, svc)
		test(hashKey, svc)
		test(hashKey, svc)
	}
	{
		fmt.Println("------------------------------")
		hashKey, svc := "432", "xsf-server"
		test(hashKey, svc)
		test(hashKey, svc)
		test(hashKey, svc)
		test(hashKey, svc)
	}
	//{
	//	//第一组测试
	//	req := xsf.NewReq()
	//	req.SetParam("k1", "v1")
	//	req.SetParam("k2", "v2")
	//	req.SetParam("k3", "v3")
	//
	//	req.SetTraceID(span.Meta()) //将span信息带到后端
	//	baseTime := time.Now()
	//	res, code, e := c.Call("xsf-server", "req", req, tm)
	//	dur := time.Now().Sub(baseTime)
	//	if code != 0 || e != nil {
	//		log.Fatalf("sse err,code:%v,e:%v,dur:%v\n", code, e, dur.Seconds())
	//	} else {
	//		fmt.Printf("F.NO.1 => handle:%s,dur:%vs\n", res.Handle(), dur.Seconds())
	//	}
	//
	//	res, code, e = c.Call("xsf-server", "req", req, tm)
	//	if code != 0 || e != nil {
	//		log.Fatal("sse err", code, e)
	//	} else {
	//		fmt.Printf("F.NO.2 => handle:%s\n", res.Handle())
	//	}
	//
	//	res, code, e = c.Call("xsf-server", "req", req, tm)
	//	if code != 0 || e != nil {
	//		log.Fatal("sse err", code, e)
	//	} else {
	//		fmt.Printf("F.NO.3 => handle:%s\n", res.Handle())
	//	}
	//}

}
func crashTest(c *xsf.Caller) {
	go func() {
		for {
			time.Sleep(time.Second)
			fmt.Println(runtime.NumGoroutine())
		}
	}()
	var wg sync.WaitGroup
	for ix := 0; ix < 500000; ix++ {
		wg.Add(1)
		go func() {
			for {
				_, _, e := c.CallWithAddr("iat", "req", "10.1.87.69:8080", xsf.NewReq(), time.Millisecond*200)
				if e != nil {
					fmt.Println(e)
				}
			}
		}()
	}
	wg.Wait()
}
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
	{
		//fmt.Println(cli.Cfg().GetRawCfg())
		////配置读取测试
		//val, err := cli.Cfg().GetBool("xsf-client", "k")
		//if err != nil {
		//	panic(err)
		//} else {
		//	panic(val)
		//}
	}
	//callTest(xsf.NewCaller(cli), time.Millisecond*time.Duration(*tm))
	//callConHashTest(xsf.NewCaller(cli), time.Millisecond*time.Duration(*tm))
	//callWithAddr(xsf.NewCaller(cli), time.Second*3)
	//sessionCallExample(xsf.NewCaller(cli), time.Millisecond*time.Duration(*tm))
	/*
		callExample(xsf.NewCaller(cli), time.Millisecond*time.Duration(*tm))
		//sessionCallExample(xsf.NewCaller(cli), time.Millisecond*time.Duration(*tm))
	*/
	//crashTest(xsf.NewCaller(cli))

	callExample(xsf.NewCaller(cli), time.Millisecond*time.Duration(*tm))
	callExample(xsf.NewCaller(cli), time.Millisecond*time.Duration(*tm))
	callExample(xsf.NewCaller(cli), time.Millisecond*time.Duration(*tm))
	callExample(xsf.NewCaller(cli), time.Millisecond*time.Duration(*tm))
}
