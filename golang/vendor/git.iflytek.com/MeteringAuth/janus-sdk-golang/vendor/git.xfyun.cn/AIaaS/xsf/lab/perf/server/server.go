package main

import (
	"errors"
	"fmt"
	"io"
	"sort"
	"sync"
	"time"

	"git.xfyun.cn/AIaaS/xsf/server"
	"git.xfyun.cn/AIaaS/xsf/utils"
	"strconv"
)

type monitor struct {
}

func (m *monitor) Query(in map[string]string, out io.Writer) {
	_, _ = out.Write([]byte(fmt.Sprintf("%+v", in)))
}

//当接收到syscall.SIGINT, syscall.SIGKILL时，会回调这接口
type killed struct {
}

func (k *killed) Closeout() {
	fmt.Println("server be killed.")
}

type healthChecker struct {
}

//服务自检接口，cmdserver用
func (h *healthChecker) Check() error {
	return errors.New("this is check function from health check")
}

//用户业务逻辑接口
type server struct {
	tool *xsf.ToolBox
}

//业务初始化接口
func (c *server) Init(toolbox *xsf.ToolBox) error {
	fmt.Println("begin init")

	tm := time.Second * 3
	fmt.Printf("ts:%v,sleeping tm:%v\n", time.Now(), tm.String())
	time.Sleep(time.Second * 3)
	fmt.Printf("ts:%v,sleep over\n", time.Now())
	c.tool = toolbox
	fmt.Println(c.tool.Cfg.GetInt64("log", "wash"))
	xsf.AddKillerCheck("server", &killed{})
	xsf.AddHealthCheck("server", &healthChecker{})
	xsf.StoreMonitor(&monitor{})
	fmt.Println("server init success.")
	return nil
}

//业务逆初始化接口
func (c *server) Finit() error {
	fmt.Println("user logic Finit success.")
	return nil
}

var peerAddr sync.Map

type strSlice []string

func (s strSlice) Len() int {
	return len(s)
}
func (s strSlice) Less(i, j int) bool {
	return s[i] < s[j]
}
func (s strSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func init() {
	go func() {
		for {
			time.Sleep(time.Second * 2)
			var strSliceTmp strSlice
			peerAddr.Range(func(key, value interface{}) bool {
				strSliceTmp = append(strSliceTmp, key.(string))
				return true
			})
			sort.Sort(strSliceTmp)
			fmt.Printf("ts:%v,addrSlice:%v\n", time.Now().String(), strSliceTmp)
			for _, val := range strSliceTmp {
				peerAddr.Delete(val)
			}
		}
	}()
}

//业务服务接口
func (c *server) Call(in *xsf.Req, span *xsf.Span) (*utils.Res, error) {
	addr, _ := in.GetParam("peerAddr")
	peerAddr.LoadOrStore(addr, struct{}{})

	res := xsf.NewRes()
	res.SetParam("intro", "received data")
	res.SetParam("op", "req")
	res.SetParam("ip", c.tool.NetManager.GetIp())
	res.SetParam("port", strconv.Itoa(c.tool.NetManager.GetPort()))

	return res, nil
}
