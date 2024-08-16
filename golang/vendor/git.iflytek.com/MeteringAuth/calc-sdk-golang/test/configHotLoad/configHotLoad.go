package main

import (
	"fmt"
	"git.iflytek.com/MeteringAuth/calc-sdk-golang/calc"
	"time"
)

/*
test usage
	1.  修改配置中心中calc.toml的enable配置
		当为false时，日志中显示calc sdk has been disabled ,并无任何计量信息，
		当为true时，计量数据发送至rmq后退出客户端，且有custom_number条producer finish日志
    2.  enable配置热加载效果


*/

func main() {
	if err := calc.Init("http://10.1.87.69:6868", "guiderAllService", "gas", "calc-client", "2.1.0", false, "./calc.toml"); err != nil {
		fmt.Println("calc init error : ", err)
		return
	}
	for i := 0; i < 1000; i++ {
		code, err := calc.Calc("testCalcSDK", "calcsdk", "nothing", 1)
		if err != nil {
			fmt.Printf("errocde : %d , errorInfo = %s", code, err)
		}
		time.Sleep(1 * time.Second)
	}
	calc.Fini()
	fmt.Println("program terminated")
}
