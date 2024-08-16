/*
 * @CopyRight: Copyright 2019 IFLYTEK Inc
 * @Author: jianjiang@iflytek.com
 * @LastEditors: Please set LastEditors
 * @Desc: 授权上报接口功能测试文件
 * @createTime: 2019-05-09 16:15:22
 * @modifyTime: 2019-06-20 16:23:46
 */

package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"

	"git.iflytek.com/MeteringAuth/janus-sdk-golang/report"
)

var (
	channel        string = "tts"
	scale          int    = 200
	randomRange    int    = 100
	reportInterval int    = 500

	companionUrl = "http://10.1.87.69:6868"
	projectName  = "guiderAllService"
	group        = "gas"
	serviceName  = "janus"
	version      = "2.0.0"
)

var (
	cnt = 0
)

func produceTestData() (d map[string]uint) {
	//	R = rand.New(rand.NewSource(time.Now().UnixNano()))
	d = make(map[string]uint)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < scale; i++ {
		r := rand.Intn(randomRange)
		d["APP@"+strconv.Itoa(i)] = uint(r)
	}
	cnt += 1
	d["4CC5779A"] = uint(cnt)
	//fmt.Println(d)
	return
}

func main() {
	// get Address
	flag.StringVar(&channel, "channel", "tts", "business channel name")
	flag.IntVar(&scale, "scale", 200, "the scale of appid")
	flag.IntVar(&randomRange, "range", 100, "the range of authorization number for each appid")
	flag.IntVar(&reportInterval, "interval", 500, "millesecond of report interval")

	flag.StringVar(&companionUrl, "url", companionUrl, "address of service finder")
	flag.StringVar(&projectName, "project", projectName, "the name of project in service finder")
	flag.StringVar(&group, "group", group, "the name of group in service finder")
	flag.StringVar(&serviceName, "service", serviceName, "the name of server in service finder")
	flag.StringVar(&version, "version", version, "the version of server in service finder")
	flag.Parse()

	fmt.Println(channel, scale, randomRange, reportInterval)

	hostName, _ := os.Hostname()
	addr, _ := net.LookupHost(hostName)
	fmt.Println("address", addr)
	// init
	report.SetCompanionUrl(companionUrl).SetProjectName(projectName).SetGroup(group).SetServiceName(serviceName).SetVersion(version).SetCacheConfig(true).SetCacheService(true).SetCfgMode(1)

	fmt.Println(report.Init(addr[0]))
	ticker := time.NewTicker(time.Duration(reportInterval) * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			report.Report(channel, produceTestData())
		}
	}
}
