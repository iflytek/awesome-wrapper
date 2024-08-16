package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"git.iflytek.com/MeteringAuth/janus-sdk-golang/authenticate"
	"git.iflytek.com/MeteringAuth/janus-sdk-golang/report"

	"git.xfyun.cn/AIaaS/xsf/utils"
)

type StaticAnalyser struct {
	region []int64
}

func (s *StaticAnalyser) count(t int64) {
	s.region[t/1e6/5] += 1
	//    fmt.Println("analyser : " , s.region[t/ 1e6 /5])
}

func (s *StaticAnalyser) init() {
	s.region = make([]int64, 1000)
}

// some global varible

const (
	AUTH   = "authenticate"
	REPORT = "report"
)

var (
	companionUrl = "http://10.1.87.69:6868"
	projectName  = "guiderAllService"
	group        = "gas"
	serviceName  = "janus"
	version      = "2.0.0"

	appid      = "4CC5779A"
	channel    = "janus"
	funcs      = ""
	funcsSlice []string
	uid        = ""

	threadNum = 1
	loopNum   = 1

	packageName = AUTH

	sidInst        *utils.SidGenerator2
	staticAnalyser []*StaticAnalyser

	done chan bool = make(chan bool, 1)
	wg   sync.WaitGroup

	reportData = make(map[string]uint)
	scale      = 100
	Do         func()
)

// init sdk
func Init() error {

	flag.IntVar(&threadNum, "thread_num", threadNum, "the number of concurrent")
	flag.IntVar(&loopNum, "loop_num", loopNum, "the loop times of each concurrent")

	flag.StringVar(&appid, "appid", appid, "appid like 4CC5779A")
	flag.StringVar(&channel, "channel", channel, "channel like tts/iat")
	flag.StringVar(&funcs, "funcs", funcs, "funcs like vcn.xiaoyan;ent.sms")
	flag.StringVar(&uid, "uid", uid, "user id")

	flag.StringVar(&companionUrl, "url", companionUrl, "address of service finder")
	flag.StringVar(&projectName, "project", projectName, "the name of project in service finder")
	flag.StringVar(&group, "group", group, "the name of group in service finder")
	flag.StringVar(&serviceName, "service", serviceName, "the name of server in service finder")
	flag.StringVar(&version, "version", version, "the version of server in service finder")

	flag.StringVar(&packageName, "package", packageName, "the target package of benchmark")
	flag.IntVar(&scale, "scale", scale, "the scale of appid")
	flag.Parse()

	fmt.Println(threadNum, loopNum, appid, channel, funcs, uid, companionUrl, projectName, group, serviceName, version)

	hostName, _ := os.Hostname()
	addr, _ := net.LookupHost(hostName)
	sidInst = &utils.SidGenerator2{}
	sidInst.Init("dx", addr[0], "2222")

	funcsSlice = strings.Split(funcs, ";")

	staticAnalyser = make([]*StaticAnalyser, threadNum)

	// capture <ctrl + C> signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		//fmt.Println("captrue <Ctrl^C>")
		done <- true
	}()

	if packageName == AUTH {
		Do = DoAuth
		authenticate.SetCompanionUrl(companionUrl).SetProjectName(projectName).SetGroup(group).SetServiceName(serviceName).SetVersion(version).SetCacheConfig(true).SetCacheService(true).SetCfgMode(1)
		return authenticate.Init([]string{channel})
	} else if packageName == REPORT {
		prepareReportData()
		Do = DoReport
		report.SetCompanionUrl(companionUrl).SetProjectName(projectName).SetGroup(group).SetServiceName(serviceName).SetVersion(version).SetCacheConfig(true).SetCacheService(true).SetCfgMode(1)
		return report.Init(channel, addr[0])
	}
	return nil

}

func prepareReportData() {
	for i := 0; i < scale; i++ {
		reportData["appid@"+strconv.Itoa(i)] = 666
	}
}

func DoReport() {
	report.Report(reportData)
}

//func DoAuth(appid, uid, channel string, funcs []string, option authenticate.CtrlMode) {
func DoAuth() {

	//sid, _ := sidInst.NewSid("jas")
	//	span := utils.NewSpan(utils.CliSpan).Start()
	//	defer span.End().Flush()

	//	span = span.WithName("JanusTest")
	//	span = span.WithTag("sid", sid)
	_, logInfo, err := authenticate.Check(appid, uid, channel, funcsSlice)
	fmt.Println("logInfo : ", logInfo)
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println(ret, info, err)

}

// pressure test case
func press(Doit func()) {
	for i := 0; i < threadNum; i++ {
		counter := &StaticAnalyser{make([]int64, 1000)}
		staticAnalyser[i] = counter
		wg.Add(1)
		go func(c *StaticAnalyser) {
			for j := 0; j < loopNum; j++ {
				t := time.Now().UnixNano()
				Doit()
				e := time.Now().UnixNano()
				c.count(e - t)

			}
			wg.Done()
			wg.Wait()
			done <- true
		}(counter)
		//fmt.Println("counter : " , staticAnalyser[i])
	}
}

func analysis(cost int64) {
	var queryAll int64
	var distributeAll []int64 = make([]int64, 1000)
	for _, v := range staticAnalyser {
		//   fmt.Println("v :" ,v)
		for k, r := range v.region {
			//        fmt.Println("r : " , r)
			queryAll += r
			distributeAll[k] += r
		}
	}

	fmt.Println("queryAll : ", queryAll)

	// qps
	//fmt.Println("cost : " , cost )
	fmt.Printf("cost %f ms: \n", float64(cost)/1e6)
	qps := float64(queryAll) / (float64(cost) / 1e9)
	fmt.Println("qps is : ", qps)

	// average lantency
	fmt.Printf("average lantency is %f ms: \n", (float64(cost)/1e6)/(float64(queryAll)/float64(threadNum)))

	// p99

	fmt.Println("lantency distribute data:")
	for k, v := range distributeAll {
		if v > 0 {
			fmt.Printf("%d ~ %d : %d\n", 5*k, 5*(k+1), v)
		}
	}

}

// benchmark main
func main() {
	if err := Init(); err != nil {
		panic(err)
	}
	// preview hot
	//Do(appid, uid, channel, funcsSlice, authenticate.CtrlType)
	Do()

	t := time.Now().UnixNano()
	press(Do)
	<-done
	e := time.Now().UnixNano()
	analysis(e - t)

}
