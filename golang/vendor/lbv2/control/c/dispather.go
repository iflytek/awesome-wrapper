package main

import (
	"git.xfyun.cn/AIaaS/xsf/client"
	"git.xfyun.cn/AIaaS/xsf/utils"
	"log"
)

var (
	failCount int64 = 0 //失败数统计
	sucCount  int64 = 0 //成功数统计
	countFile       = "count.txt"
)

func runClient(cfgFile string) {
	defCfg := "[sfc]\ntimeout=1000\n"
	cli, e := xsf.InitClient("lb_ctl", utils.Native,
		utils.WithCfgName(cfgFile),
		utils.WithCfgDefault(defCfg),
	)

	if e != nil {
		log.Fatal("main | InitCient error:", e)
	}

	modeString, modeStringErr := cli.Cfg().GetString("dispatcher", "mode")
	if modeStringErr != nil {
		log.Fatalf("mode error")
	}
	switch modeString {
	case "client":
		{
			client(cli)
		}
	case "reporter":
		{
			reporter(cli)
		}
	case "clientEx":
		{
			clientEx(cli)
		}
	case "reporterEx":
		{
			reporterEx(cli)
		}
	default:
		{
			log.Fatalf("don't support mode:%v", modeString)
		}
	}

}
