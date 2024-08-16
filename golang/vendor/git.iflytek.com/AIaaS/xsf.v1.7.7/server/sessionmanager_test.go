package xsf

import (
	"context"
	"fmt"
	"git.iflytek.com/AIaaS/xsf/utils"
	"log"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

func TestSessionManager(t *testing.T) {
	fmt.Printf("begin:%v\n", time.Now())
	var (
		lbStrategy     = 0                                                                          //负载策略(必传)
		zkList         = []string{"192.168.86.60:2191", "192.168.86.60:2192", "192.168.86.60:2190"} //zk列表(必传)
		root           = ""                                                                         //根目录
		routerType     = "iat"                                                                      //路由类型(如：iat)(必传)
		subRouterTypes = []string{"iat_gray", "iat_hefei"}                                          //子路由类型列表(如:["iat_gray","iat_hefei"])
		redieHost      = "192.168.86.60:6379"                                                       //redis主机(必传)
		redisPasswd    = ""                                                                         //redis密码
		maxActive      = 100                                                                        //redis最大连接数
		maxIdle        = 50                                                                         //redis最大空闲连接数
		db             = 0                                                                          //redis数据库
		idleTimeOut    = time.Second * 1000                                                         //redis空闲连接数超时时间
		svc            = "192.168.86.60:2181"                                                       //引擎节点
	)

	lc := LbAdapter{}
	if InitErr := lc.Init(
		WithLbAdapterSvc(svc),
		WithLbAdapterStrategy(lbStrategy),
		WithLbAdapterZkList(zkList),
		WithLbAdapterRoot(root),
		WithLbAdapterRouterType(routerType),
		WithLbAdapterSubRouterTypes(subRouterTypes),
		WithLbAdapterSRedisHost(redieHost),
		WithLbAdapterSRedisPasswd(redisPasswd),
		WithLbAdapterMaxActive(maxActive),
		WithLbAdapterMaxIdle(maxIdle),
		WithLbAdapterDb(db),
		WithLbAdapterIdleTimeOut(idleTimeOut)); InitErr != nil {
		log.Panicf("InitErr:%v\n", InitErr)
	} else {
		fmt.Printf("Init success.\n")
	}

	callbackCnt := int64(0)
	setCnt := int64(0)
	maxlic := 10000
	bestlic := 10
	timeout := time.Second * 3
	rollTime := time.Second
	reportInterval := time.Second
	bc := BootConfig{CfgMode: utils.Native, CfgData: CfgMeta{CfgName: "test.toml", Project: "test", Group: "default", Service: "xsf", Version: "1.0.0", CompanionUrl: "http://10.1.86.228:9080"}}

	logger, loggerErr := utils.NewLocalLog(
		utils.SetLevel("info"),
		utils.SetFileName("xsfs.log"),
		utils.SetMaxSize(10),
		utils.SetMaxBackups(10),
		utils.SetBatchSize(10))
	if loggerErr != nil {
		log.Panicf("loggerErr:%v", loggerErr)
	}

	var sm SessionManager
	sm.Init(
		WithSessionManagerBc(bc),
		WithSessionManagerMaxLic(int32(maxlic)),
		WithSessionManagerBestLic(int32(bestlic)),
		WithSessionManagerTimeout(timeout),
		WithSessionManagerRollTime(rollTime),
		WithSessionManagerReportInterval(int32(reportInterval)),
		WithSessionManagerReporter(lc),
		WithSessionManagerLogger(logger))

	atomic.AddInt64(&setCnt, 1)
	sm.SetSessionData("sessionTag", "svcData", func(sessionTag interface{}, svcData interface{}, Exception ...CallBackException) {
		fmt.Printf("info:this is callback function. -> sessionTag:%v\n", sessionTag)
		atomic.AddInt64(&callbackCnt, 1)
	})

	time.Sleep(time.Millisecond * 500)
	if resp, err := sm.GetSessionData("sessionTag"); err != nil {
		t.Fatal(err)
	} else {
		fmt.Printf("resp:%v\n", resp)
	}
	sm.DelSessionDataDelay("sessionTag")
	for setIx := 0; setIx < 1000; setIx++ {
		atomic.AddInt64(&setCnt, 1)
		fmt.Println(sm.SetSessionData("sessionTag"+strconv.Itoa(setIx), "svcData"+strconv.Itoa(setIx), func(sessionTag interface{}, svcData interface{}, Exception ...CallBackException) {
			fmt.Printf("info:this is callback function. -> sessionTag:%v\n", sessionTag)
			atomic.AddInt64(&callbackCnt, 1)
		}))
	}
	fmt.Println()
	time.Sleep(time.Second * 15)
	fmt.Printf("setCnt:%v,callbackCnt:%v\n", setCnt, callbackCnt)
}
func TestSessionManager2(t *testing.T) {
	fmt.Printf("begin:%v\n", time.Now())
	var (
		lbStrategy     = 0                                                                          //负载策略(必传)
		zkList         = []string{"192.168.86.60:2191", "192.168.86.60:2192", "192.168.86.60:2190"} //zk列表(必传)
		root           = ""                                                                         //根目录
		routerType     = "iat"                                                                      //路由类型(如：iat)(必传)
		subRouterTypes = []string{"iat_gray", "iat_hefei"}                                          //子路由类型列表(如:["iat_gray","iat_hefei"])
		redieHost      = "192.168.86.60:6379"                                                       //redis主机(必传)
		redisPasswd    = ""                                                                         //redis密码
		maxActive      = 100                                                                        //redis最大连接数
		maxIdle        = 50                                                                         //redis最大空闲连接数
		db             = 0                                                                          //redis数据库
		idleTimeOut    = time.Second * 1000                                                         //redis空闲连接数超时时间
		svc            = "192.168.86.60:2181"                                                       //引擎节点
	)

	lc := LbAdapter{}
	if InitErr := lc.Init(
		WithLbAdapterSvc(svc),
		WithLbAdapterStrategy(lbStrategy),
		WithLbAdapterZkList(zkList),
		WithLbAdapterRoot(root),
		WithLbAdapterRouterType(routerType),
		WithLbAdapterSubRouterTypes(subRouterTypes),
		WithLbAdapterSRedisHost(redieHost),
		WithLbAdapterSRedisPasswd(redisPasswd),
		WithLbAdapterMaxActive(maxActive),
		WithLbAdapterMaxIdle(maxIdle),
		WithLbAdapterDb(db),
		WithLbAdapterIdleTimeOut(idleTimeOut)); InitErr != nil {
		log.Panicf("InitErr:%v\n", InitErr)
	} else {
		fmt.Printf("Init success.\n")
	}

	//callbackCnt := int64(0)
	//setCnt := int64(0)
	maxlic := 10
	bestlic := 10
	timeout := time.Second * 3
	rollTime := time.Second
	reportInterval := time.Second
	bc := BootConfig{CfgMode: utils.Native, CfgData: CfgMeta{CfgName: "test.toml", Project: "test", Group: "default", Service: "xsf", Version: "1.0.0", CompanionUrl: "http://10.1.86.228:9080"}}

	logger, loggerErr := utils.NewLocalLog(utils.SetLevel("debug"),
		utils.SetFileName("xsfs.log"),
		utils.SetMaxSize(10),
		utils.SetMaxBackups(10),
		utils.SetBatchSize(10))
	if loggerErr != nil {
		log.Panicf("loggerErr:%v", loggerErr)
	}

	var sm SessionManager
	sm.Init(
		WithSessionManagerBc(bc),
		WithSessionManagerMaxLic(int32(maxlic)),
		WithSessionManagerBestLic(int32(bestlic)),
		WithSessionManagerTimeout(timeout),
		WithSessionManagerRollTime(rollTime),
		WithSessionManagerReportInterval(int32(reportInterval)),
		WithSessionManagerReporter(lc),
		WithSessionManagerLogger(logger))

	//for setIx := 0; setIx < 1000; setIx++ {
	//	atomic.AddInt64(&setCnt, 1)
	//	fmt.Println(sm.SetSessionData("sessionTag"+strconv.Itoa(setIx), "svcData"+strconv.Itoa(setIx), func(sessionTag interface{}, svcData interface{}) {
	//		fmt.Printf("info:this is callback function. -> sessionTag:%v\n", sessionTag)
	//		atomic.AddInt64(&callbackCnt, 1)
	//	}))
	//	sm.DelSessionDataDelay("sessionTag" + strconv.Itoa(setIx))
	//}
	//time.Sleep(time.Second * 3)
	sm.UpdateDelay()
	sm.MaxLic = 999
	sm.UpdateDelay()
	time.Sleep(time.Second * 3)
}

func TestSessionManager3(t *testing.T) {
	fmt.Printf("begin:%v\n", time.Now())
	var (
		lbStrategy     = 0                                                                          //负载策略(必传)
		zkList         = []string{"192.168.86.60:2191", "192.168.86.60:2192", "192.168.86.60:2190"} //zk列表(必传)
		root           = ""                                                                         //根目录
		routerType     = "iat"                                                                      //路由类型(如：iat)(必传)
		subRouterTypes = []string{"iat_gray", "iat_hefei"}                                          //子路由类型列表(如:["iat_gray","iat_hefei"])
		redieHost      = "192.168.86.60:6379"                                                       //redis主机(必传)
		redisPasswd    = ""                                                                         //redis密码
		maxActive      = 100                                                                        //redis最大连接数
		maxIdle        = 50                                                                         //redis最大空闲连接数
		db             = 0                                                                          //redis数据库
		idleTimeOut    = time.Second * 1000                                                         //redis空闲连接数超时时间
		svc            = "192.168.86.60:2181"                                                       //引擎节点
	)

	lc := LbAdapter{}
	if InitErr := lc.Init(
		WithLbAdapterSvc(svc),
		WithLbAdapterStrategy(lbStrategy),
		WithLbAdapterZkList(zkList),
		WithLbAdapterRoot(root),
		WithLbAdapterRouterType(routerType),
		WithLbAdapterSubRouterTypes(subRouterTypes),
		WithLbAdapterSRedisHost(redieHost),
		WithLbAdapterSRedisPasswd(redisPasswd),
		WithLbAdapterMaxActive(maxActive),
		WithLbAdapterMaxIdle(maxIdle),
		WithLbAdapterDb(db),
		WithLbAdapterIdleTimeOut(idleTimeOut)); InitErr != nil {
		log.Panicf("InitErr:%v\n", InitErr)
	} else {
		fmt.Printf("Init success.\n")
	}

	//callbackCnt := int64(0)
	//setCnt := int64(0)
	maxlic := 10
	bestlic := 10
	timeout := time.Second * 3
	rollTime := time.Second * 50
	reportInterval := time.Second
	bc := BootConfig{CfgMode: utils.Native, CfgData: CfgMeta{CfgName: "test.toml", Project: "test", Group: "default", Service: "xsf", Version: "1.0.0", CompanionUrl: "http://10.1.86.228:9080"}}

	logger, loggerErr := utils.NewLocalLog(utils.SetLevel("debug"),
		utils.SetFileName("xsfs.log"),
		utils.SetMaxSize(10),
		utils.SetMaxBackups(10),
		utils.SetBatchSize(10))
	if loggerErr != nil {
		log.Panicf("loggerErr:%v", loggerErr)
	}

	var sm SessionManager
	sm.Init(
		WithSessionManagerBc(bc),
		WithSessionManagerMaxLic(int32(maxlic)),
		WithSessionManagerBestLic(int32(bestlic)),
		WithSessionManagerTimeout(timeout),
		WithSessionManagerRollTime(rollTime),
		WithSessionManagerReportInterval(int32(reportInterval)),
		WithSessionManagerReporter(lc),
		WithSessionManagerLogger(logger))

	/*
		1、expect>now
		2、expect<now
	*/
	go func() {
		for {
			time.Sleep(time.Second)
			fmt.Printf("maxLic:%v,nowLic:%v\n", sm.MaxLic, sm.NowLic)
		}
	}()
	go func() {
		for setIx := 0; setIx < 5; setIx++ {
			sm.SetSessionData(setIx, "svcData", func(sessionTag interface{}, svcData interface{}, Exception ...CallBackException) {})
		}
	}()
	go func() {
		time.Sleep(time.Second * 2)
		sm.DelSessionData(1)
		time.Sleep(time.Second)
		sm.DelSessionData(2)
		time.Sleep(time.Second)
		sm.DelSessionData(3)
		time.Sleep(time.Second)
		sm.DelSessionData(4)
	}()

	time.Sleep(time.Second)
	sm.UpdateOTF(context.Background(), WithSessionManagerMaxLicOTF(80))
	time.Sleep(time.Hour)
}
