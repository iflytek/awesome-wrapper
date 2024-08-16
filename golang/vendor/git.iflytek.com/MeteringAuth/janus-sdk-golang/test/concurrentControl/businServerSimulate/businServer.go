/*
 * @CopyRight: Copyright 2019 IFLYTEK Inc
 * @Author: jianjiang@iflytek.com
 * @LastEditors: Please set LastEditors
 * @Desc:
 * @createTime: 2019-05-16 11:04:54
 * @modifyTime: 2019-05-16 16:20:50
 */

package main

import (
	"context"
	"flag"
	"net"
	"os"
	"time"

	"concurrentControl-gen-proto"

	"git.iflytek.com/MeteringAuth/janus-sdk-golang/authenticate"

	"fmt"
	"git.xfyun.cn/AIaaS/xsf/utils"
	"google.golang.org/grpc"
)

var (
	triggleElement []chan Bullet
	sidInst        = &utils.SidGenerator2{}

	SimulateNode    = 1000
	TriggleInterval = 1000

	channel = "tts"
	appid   = "4CC57793"
	uid     = "lhli"
	funcs   = []string{"test2"}
	//option = authenticate.CtrlDayFlow|authenticate.CtrlHourFlow
	option = authenticate.CtrlConcFlow
	//option = authenticate.CtrlDayFlow

	companionUrl = "http://10.1.87.69:6868"
	projectName  = "guiderAllService"
	group        = "gas"
	serviceName  = "janus"
	version      = "2.0.1"
	EngineAddr   = "10.1.87.61:9090"
)

type Bullet struct{}

func main() {

	// flag parse

	flag.IntVar(&SimulateNode, "nodeNum", SimulateNode, "the number of simulateServer")
	flag.IntVar(&TriggleInterval, "interval", TriggleInterval, "the interval duration of sendMessage")
	flag.StringVar(&appid, "appid", appid, "appid like 4CC5779A")
	flag.StringVar(&channel, "channel", channel, "channel like tts/iat")

	flag.StringVar(&companionUrl, "url", companionUrl, "address of service finder")
	flag.StringVar(&projectName, "project", projectName, "the name of project in service finder")
	flag.StringVar(&group, "group", group, "the name of group in service finder")
	flag.StringVar(&serviceName, "service", serviceName, "the name of server in service finder")
	flag.StringVar(&version, "version", version, "the version of server in service finder")

	flag.Parse()

	lanuchGatlin()
	for {
		for i := 0; i < SimulateNode; i++ {
			triggleElement[i] <- Bullet{}
			time.Sleep(time.Duration(TriggleInterval) * time.Millisecond)
		}
	}

}

func lanuchGatlin() {
	// basic trigger init
	triggleElement = make([]chan Bullet, SimulateNode)
	for i := 0; i < SimulateNode; i++ {
		triggleElement[i] = make(chan Bullet, 1)
	}
	hostName, _ := os.Hostname()
	addr, _ := net.LookupHost(hostName)
	sidInst.Init("dx", addr[0], "2222")

	// authenticate package init
	authenticate.SetCompanionUrl(companionUrl).SetProjectName(projectName).SetGroup(group).SetServiceName(serviceName).SetVersion(version).SetCacheConfig(true).SetCacheService(true).SetCfgMode(1)
	err := authenticate.Init([]string{channel})
	if err != nil {
		panic(err)
	}

	// authorization report package init

	for i := 0; i < SimulateNode; i++ {
		go func(index int) {
			DoSimulate(index)
		}(i)
	}

}

func DoSimulate(i int) {

	c, conn := NewGrpcConnect()
	for {
		sid, _ := sidInst.NewSid("jas")
		span := utils.NewSpan(utils.CliSpan).Start()
		defer span.End().Flush()
		span = span.WithName("JanusTest")
		span = span.WithTag("sid", sid)

		<-triggleElement[i]

		// do Authentication
		//_, info, err := authenticate.Check(appid, uid, channel, funcs, option, sid, span.Meta())
		info, logInfo, err := authenticate.Check(appid, uid, channel, funcs)
		if err != nil {
			panic(err)
		}

		if len(info) == 0 {
			// run to simulatetion engine
			SendMessage(c)
		} else {
			fmt.Println(info, sid, logInfo)
		}
	}
	conn.Close()
}

func SendMessage(c concurrentNet.ConcurrentNetaClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	//atomic.AddInt32(&onRoad , 1)
	_, err := c.SendMsg(ctx, &concurrentNet.SimpleMsg{Appid: appid})
	//atomic.AddInt32(&onRoad , -1)
	if err != nil {
		//		zaplogWrap.Logger.Error("send message failed  ", zap.Error(err))
		panic(err)
	}
	//if !reply.Ret {
	//zaplogWrap.Logger.Info("send message success" , zap.Bool("ret" , reply.Ret))
	//}
}

func NewGrpcConnect() (concurrentNet.ConcurrentNetaClient, *grpc.ClientConn) {
	conn, err := grpc.Dial(EngineAddr, grpc.WithInsecure())
	if err != nil {
		//	zaplogWrap.Logger.Error("Dial endpoint failed ", zap.Error(err))
		panic(err)
	}
	//defer conn.Close()
	c := concurrentNet.NewConcurrentNetaClient(conn)
	return c, conn

}
