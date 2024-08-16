package main

import (
	"errors"
	"fmt"
	"git.iflytek.com/AIaaS/xsf/server"
	"git.iflytek.com/AIaaS/xsf/utils"
	"io"
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

	c.tool = toolbox
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

//业务服务接口
func (c *server) Call(in *xsf.Req, span *xsf.Span) (*utils.Res, error) {
	res := xsf.NewRes()
	res.SetParam("intro", "received data")
	res.SetParam("op", "req")
	res.SetParam("ip", c.tool.NetManager.GetIp())
	res.SetParam("port", strconv.Itoa(c.tool.NetManager.GetPort()))
	return res, nil
}
