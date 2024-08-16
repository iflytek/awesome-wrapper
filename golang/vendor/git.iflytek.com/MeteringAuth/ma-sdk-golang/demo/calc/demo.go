package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	xsf "git.iflytek.com/AIaaS/xsf/server"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/calc"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/calc/ver"
)

var cnt = flag.Int("l", 1, "loop times")
var appid = flag.String("appid", "sdka", "appid")
var channel = flag.String("channel", "sdkc", "channel")
var function = flag.String("function", "sdkf", "function")
var did = flag.String("did", "did", "did/uid")

func main() {
	flag.Parse()
	// 初始化参数
	// @url 配置中心地址
	// @pro 配置中心项目名
	// @gro 配置中心项目分组名
	// @isNative 是否使用本地配置
	// @nativeLogPath 本地配置路径（仅当isNative为true时生效）
	if err := calc.Init(*xsf.CompanionUrl,
		*xsf.Project,
		*xsf.Group,
		*xsf.Service,
		ver.Version,
		*xsf.Mode); err != nil {
		log.Fatal(err)
	}

	for x := 0; x < *cnt || *cnt < 0; x++ {
		// 调用计量函数
		// @appid 应用名称
		// @channel 业务渠道
		// @funcs 功能名称
		// @c 用量
		// 该函数只计量APPID
		code, err := calc.Calc(*appid, *channel, *function, 1)
		fmt.Println(*appid, *channel, *function, 1, "->", code, err)
		if err != nil {
			fmt.Printf("errocde : %d , errorInfo = %s", code, err)
			return
		}
		// 调用计量函数
		// @appid 应用名称
		// @did 设备id
		// @channel 业务渠道
		// @funcs 功能名称
		// @c 用量
		// 该函数只计量SubId
		code, err = calc.CalcWithSubId(*appid, *did, *channel, *function, 1)
		fmt.Println(*appid, *did, *channel, *function, 1, "->", code, err)
		if err != nil {
			fmt.Printf("errocde : %d , errorInfo = %s", code, err)
			return
		}
		time.Sleep(1 * time.Second)
	}

	// sdk 析构
	calc.Fini()
	fmt.Println("calc done")
}
