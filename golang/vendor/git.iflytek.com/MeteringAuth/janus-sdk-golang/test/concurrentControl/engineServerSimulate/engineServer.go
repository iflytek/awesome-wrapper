/*
 * @CopyRight: Copyright 2019 IFLYTEK Inc
 * @Author: jianjiang@iflytek.com
 * @LastEditors: Please set LastEditors
 * @Desc:
 * @createTime: 2019-05-16 16:59:30
 * @modifyTime: 2019-05-16 20:59:12
 */

package main

import (
	//"github.com/cihub/seelog"
	"flag"
	"fmt"
	"time"
	"zaplogWrap"
)

var (
	THRESHOLD    uint32        = 200
	cycleTime    time.Duration = 500 * time.Millisecond
	reportTime   int           = 500
	SLAVENUM     int           = 200
	APPID        string        = "4CC5779A"
	CHANNEL      string        = "tts"
	begin        int           = 1000
	MockDealTime int           = 5000
	reportLaunch time.Duration = time.Duration(int(reportTime) / SLAVENUM)

	companionUrl = "http://10.1.87.69:6868"
	projectName  = "guiderAllService"
	group        = "gas"
	serviceName  = "janus"
	version      = "2.0.1"
)

func main() {
	// start grpc listen
	//seelog.Info("start server ...")

	flag.IntVar(&reportTime, "reportInterval", reportTime, "report interval")
	flag.IntVar(&SLAVENUM, "nodeNum", SLAVENUM, "the node number of engines")
	flag.IntVar(&begin, "bottom", begin, "the minimum duration of engine deal")
	flag.IntVar(&MockDealTime, "randomRange", MockDealTime, "the random part duration of engine deal")
	flag.StringVar(&CHANNEL, "channel", CHANNEL, "channel like tts/iat")
	flag.StringVar(&APPID, "appid", APPID, "appid like 4CC5779A")

	flag.StringVar(&companionUrl, "url", companionUrl, "address of service finder")
	flag.StringVar(&projectName, "project", projectName, "the name of project in service finder")
	flag.StringVar(&group, "group", group, "the name of group in service finder")
	flag.StringVar(&serviceName, "service", serviceName, "the name of server in service finder")
	flag.StringVar(&version, "version", version, "the version of server in service finder")
	flag.Parse()

	fmt.Println("reportLaunch", reportLaunch)
	zaplogWrap.Logger.Info("start server")
	StartNode()
	startServer()
	return
}
