package main

import (
	"fmt"
	"git.iflytek.com/MeteringAuth/janus-sdk-golang/authenticate"
)

func main() {
	// 设置服务发现/配置中心路径
	//authenticate.SetCompanionUrl("http://10.1.87.69:6868").SetProjectName("guiderAllService").SetGroup("gas").SetServiceName("janus-client").SetVersion("2.0.6").SetCacheConfig(true).SetCacheService(true).SetCfgMode(1)
	authenticate.SetCompanionUrl("http://10.1.87.69:6868").SetProjectName("metrics").SetGroup("reporter").SetServiceName("janus-client").SetVersion("2.0.6").SetCacheConfig(true).SetCacheService(true).SetCfgMode(1)
	//authenticate.SetCompanionUrl("http://10.1.87.54:9080").SetProjectName("AIaaS").SetGroup("aipaas").SetServiceName("passivefeaonline").SetVersion("1.0.2").SetCacheConfig(true).SetCacheService(true).SetCfgMode(1)
	// 初始化鉴权sdk
	// 支持正则匹配
	//if err := authenticate.Init([]string{"^tts$"}); err != nil {
	if err := authenticate.Init([]string{"passivefeaonline"}); err != nil {
		fmt.Println("authenticate init error : ", err)
		return
	}
	// 发起鉴权
	result, logInfo, err := authenticate.Check("4CC5779A", "useruid", "passivefeaonline", []string{"business.total", "clusteringAudio"})
	fmt.Printf("result : %+v , logInfo : %s , err : %s\n", result, logInfo, err)

}
