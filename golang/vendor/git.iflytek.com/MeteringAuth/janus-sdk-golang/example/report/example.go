package main

import (
	"fmt"
	"time"

	"git.iflytek.com/MeteringAuth/janus-sdk-golang/report"
)

func main() {

	// 设置配置中心/服务发现
	report.SetCompanionUrl("http://10.1.87.69:6868").SetProjectName("guiderAllService").SetGroup("gas").SetServiceName("janus").SetVersion("2.0.1").SetCacheConfig(true).SetCacheService(true).SetCfgMode(1)
	if err := report.Init("tts", "10.1.87.69:8080"); err != nil {
		fmt.Println("report init error : ", err)
	}

	// 构造上报数据
	var a = make(map[string]uint, 10)
	a["4CC5779C"] = 10
	report.Report(a)
	// 上报接口为异步接口，在此example中等待1秒，确保数据被发送
	time.Sleep(1 * time.Second)
}
