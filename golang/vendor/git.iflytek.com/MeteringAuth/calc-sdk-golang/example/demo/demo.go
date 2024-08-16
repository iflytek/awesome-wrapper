package main

import (
	"fmt"
	"git.iflytek.com/MeteringAuth/calc-sdk-golang/calc"
)

func main() {
	// 初始化参数
	// @url 配置中心地址
	// @pro 配置中心项目名
	// @gro 配置中心项目分组名
	// @isNative 是否使用本地配置
	// @nativeLogPath 本地配置路径（仅当isNative为true时生效）
	if err := calc.Init("http://10.1.87.69:6868", "guiderAllService", "gas", "calc-client", "2.1.0", false, "./calc.toml"); err != nil {
		fmt.Println("calc init error : ", err)
		return
	}

	// 调用计量函数
	// @appid 应用名称
	// @channel 业务渠道
	// @funcs 功能名称
	// @c 用量
	// 该函数只计量APPID
	code, err := calc.Calc("testCalcSDK", "calcsdk", "nothing", 1)
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
	code, err = calc.CalcWithSubId("testCalcSDK", "did", "calcsdk", "nothing", 1)
	if err != nil {
		fmt.Printf("errocde : %d , errorInfo = %s", code, err)
		return
	}

	// sdk 析构
	calc.Fini()
	fmt.Println("program terminated")

}
