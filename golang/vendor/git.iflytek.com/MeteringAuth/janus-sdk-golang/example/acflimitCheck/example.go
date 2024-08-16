package main

import (
	"flag"
	"fmt"
	"strings"

	"git.iflytek.com/MeteringAuth/janus-sdk-golang/authenticate"
)

var appid = flag.String("a", "", "appid")
var sub = flag.String("c", "", "channel")
var ent = flag.String("f", "", "function")
var channel = flag.String("channel", "", "init channel or c")

var u = flag.String("u", "http://10.1.87.69:6868", "url")
var p = flag.String("p", "metrics", "function")
var g = flag.String("g", "reporter", "function")
var s = flag.String("s", "janus-client", "function")
var v = flag.String("v", authenticate.Version, "function")

func main() {
	flag.Parse()
	authenticate.SetCompanionUrl(*u).SetProjectName(*p).SetGroup(*g).SetServiceName(*s).SetVersion(*v).SetCacheConfig(true).SetCacheService(true).SetCfgMode(1)
	su := *sub
	if *channel != "" {
		su = *channel
	}
	// 初始化鉴权sdk
	// 支持正则匹配
	//if err := authenticate.Init([]string{"^tts$"}); err != nil {
	if err := authenticate.Init([]string{"passivefeaonline"}); err != nil {
		fmt.Println("authenticate init error : ", err)
		return
	}
	// 发起鉴权
	funcs := strings.Split(*ent, ";")
	result, err := authenticate.GetAcfLimits(*appid, su, funcs)
	fmt.Printf("result : %+v ,  err : %s\n", result, err)

}
