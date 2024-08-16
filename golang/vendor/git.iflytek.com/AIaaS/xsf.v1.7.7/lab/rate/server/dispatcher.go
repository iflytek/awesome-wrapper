package main

import (
	"flag"
	"git.iflytek.com/AIaaS/xsf/server"
	"log"
	"sync"
)

var dur = flag.Int("dur", 0, "op elpased")
var delay = flag.Int("delay", 0, "delay report")

func main() {
	flag.Parse()

	//定义一个服务实例
	var serverInst xsf.XsfServer
	//定义相关的启动参数
	/*
		1、CfgMode可选值Native、Centre，native为本地配置读取模式，Centre为配置中心模式，当此值为-1时，表示有命令行传入
		2、CfgName 配置文件名
		3、Project 配置中心用 项目名
		4、Group 配置中心用 组名
		5、Service 配置中心用 服务名
		6、Version 配置中心用 配置版本名
		7、CompanionUrl 配置中心用 配置中心地址
	*/
	bc := xsf.BootConfig{
		CfgMode: -1,
		CfgData: xsf.CfgMeta{
			CfgName:      "server.toml",
			Project:      "xsf",
			Group:        "xsf",
			Service:      "xsf-server",
			Version:      "1.0.0",
			ApiVersion:   "1.0.0",
			CachePath:    "xxx",
			CompanionUrl: "http://10.1.87.70:6868"}}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
		}()
		/*
			1、启动服务
			2、若有异常直接报错，注意需用户自己实现协程等待
		*/
		if err := serverInst.Run(
			bc,
			&server{},
			xsf.SetOpRouter(generateOpRouter()),
			xsf.SetRateFallback(&rateFallback{}),
		); err != nil {
			log.Fatal(err)
		}
	}()
	wg.Wait()
}
