package main

import (
	"errors"
	"fmt"
	"io"

	"git.iflytek.com/AIaaS/xsf/server"
	"git.iflytek.com/AIaaS/xsf/utils"
	"strconv"
	"time"
)

type monitor struct {
}

func (m *monitor) Query(in map[string]string, out io.Writer) {
	_, _ = out.Write([]byte( fmt.Sprintf("%+v", in)))
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

	defer func() {
		fmt.Println("about to sleep 5s")
		//time.Sleep(time.Second * 5)
		fmt.Println("end init")
	}()
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

//业务服务接口
func (c *server) Call(in *xsf.Req, span *xsf.Span) (*utils.Res, error) {
	res := xsf.NewRes()
	res.SetHandle(in.Handle())
	op := in.Op()

	if op == "ssb" {
		if SetSessionDataErr := c.tool.Cache.SetSessionData(in.Handle(), "svcData", func(sessionTag interface{}, svcData interface{}, exception ...xsf.CallBackException) {
			return
		}); SetSessionDataErr != nil {
			res.SetError(1, fmt.Sprintf("Set %s failed. ->SetErr:%v ->addr:%v", in.Handle(), SetSessionDataErr, fmt.Sprintf("%v:%v", c.tool.NetManager.GetIp(), c.tool.NetManager.GetPort())))
		} else {
			if *delay != 0 {
				c.tool.Cache.UpdateDelay()
			} else {
				_ = c.tool.Cache.Update()
			}
		}
		res.SetParam("intro", "received data")
		res.SetParam("op", "ssb")
		res.SetParam("ip", c.tool.NetManager.GetIp())
		res.SetParam("port", strconv.Itoa(c.tool.NetManager.GetPort()))
		return res, nil
	}

	if op == "auw" {
		time.Sleep(time.Millisecond * time.Duration(*dur))
		if _, GetSessionDataErr := c.tool.Cache.GetSessionData(in.Handle()); GetSessionDataErr != nil {
			res.SetError(1, fmt.Sprintf("GetSessionData failed. ->GetSessionDataErr:%v", GetSessionDataErr))
		}
		res.SetParam("intro", "received data")
		res.SetParam("op", "auw")
		res.SetParam("ip", c.tool.NetManager.GetIp())
		res.SetParam("port", strconv.Itoa(c.tool.NetManager.GetPort()))
		return res, nil
	}

	if op == "sse" {
		c.tool.Cache.DelSessionData(in.Handle())
		if *delay != 0 {
			c.tool.Cache.UpdateDelay()
		} else {
			_ = c.tool.Cache.Update()
		}
		res.SetParam("intro", "received data")
		res.SetParam("op", "sse")
		res.SetParam("ip", c.tool.NetManager.GetIp())
		res.SetParam("port", strconv.Itoa(c.tool.NetManager.GetPort()))
		return res, nil
	}

	if op == "req" {
		res.SetParam("intro", "received data")
		res.SetParam("op", "req")
		res.SetParam("ip", c.tool.NetManager.GetIp())
		res.SetParam("port", strconv.Itoa(c.tool.NetManager.GetPort()))
		return res, nil
	}

	fmt.Printf("the op -> %v is not supported.\n", op)
	res.SetError(1, fmt.Sprintf("the op -> %v is not supported.", op))
	res.SetParam("intro", "received data")
	res.SetParam("op", "illegal")
	return res, nil
}

type rateFallback struct{}

func (r *rateFallback) Call(in *xsf.Req, span *xsf.Span) (*utils.Res, error) {
	res := xsf.NewRes()
	res.SetError(1995, fmt.Sprintf("user rate fallback"))
	return res, nil
}
