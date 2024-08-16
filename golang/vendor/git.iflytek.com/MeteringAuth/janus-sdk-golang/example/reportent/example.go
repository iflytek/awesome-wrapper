package main

import (
	"flag"
	"fmt"
	"time"

	"git.iflytek.com/MeteringAuth/janus-sdk-golang/report"
)

func main() {
	//./janus -m 1 -p metrics -g reporter -s janus -c server.toml -u http://10.1.87.70:6868
	am := flag.Int("m", 1, "mode")
	ap := flag.String("p", "metrics", "program")
	ag := flag.String("g", "reporter", "group")
	as := flag.String("s", "janus-client", "service")
	//ac := flag.String("c", "janus-client.toml", "cfg name")
	au := flag.String("u", "http://10.1.87.70:6868", "url")
	sub := flag.String("sub", "sub", "report sub")
	addr := flag.String("addr", "0.x.1:dd", "report ip")

	flag.Parse()
	// 设置配置中心/服务发现
	report.
		SetCompanionUrl(*au).
		SetProjectName(*ap).
		SetGroup(*ag).
		SetServiceName(*as).
		SetVersion(report.Version).
		SetCacheConfig(true).
		SetCacheService(true).
		SetCfgMode(*am)
	if err := report.Init(*sub, *addr); err != nil {
		fmt.Println("report init error : ", err)
	}

	// 构造上报数据
	var a = make(map[string]uint, 10)
	a["4CC5779C:sms-5s"] = 10
	report.ReportEnt(a)
	// 上报接口为异步接口，在此example中等待1秒，确保数据被发送
	time.Sleep(1 * time.Second)
}
