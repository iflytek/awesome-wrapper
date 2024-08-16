package main

import (
	"errors"
	"fmt"
	"git.iflytek.com/AIaaS/xsf/server"
	"git.iflytek.com/AIaaS/xsf/utils"
	"strconv"
	"time"
)
//
//func startTrace() {
//	trace.AuthRequest = func(req *http.Request) (any, sensitive bool) {
//		return true, true
//	}
//	go http.ListenAndServe(":50051", nil)
//	grpclog.Infoln("Trace listen on 50051")
//}
//func init() {
//	grpc.EnableTracing = true
//	go startTrace()
//}

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

	fmt.Println("server init success.")
	return nil
}

//业务逆初始化接口
func (c *server) Finit() error {
	fmt.Println("user logic Finit success.")
	return nil
}

var totalCnt int64

func (c *server) Call(in *xsf.Req, span *xsf.Span) (*utils.Res, error) {
	res := xsf.NewRes()
	res.SetHandle(in.Handle())
	op := in.Op()

	//{
	//	time.Sleep(time.Minute)
	//}

	switch op {
	case "req":
		{
			//if atomic.AddInt64(&totalCnt, 1)%5 == 0 {
			//	res.SetError(1, fmt.Sprintf("the op -> %v testing failed...", op))
			//	res.SetParam("intro", "received data")
			//	res.SetParam("op", "illegal")
			//	return res, nil
			//}

			fmt.Printf("ts:%v,receive...req\n", time.Now())
			//time.Sleep(time.Millisecond * 1000)
			defer func() {
				fmt.Printf("ts:%v,leave...req\n", time.Now())
			}()
			res.SetParam("intro", "received data")
			res.SetParam("op", op)
			res.SetParam("ip", c.tool.NetManager.GetIp())
			res.SetParam("port", strconv.Itoa(c.tool.NetManager.GetPort()))
			data := xsf.NewData()
			res.AppendData(data)
			return res, nil
		}
	case "roll":
		{
			time.Sleep(time.Second)
			fmt.Printf("ts:%v,receive...roll\n", time.Now())
			res.SetParam("intro", "rollback")
			res.SetParam("op", op)
			res.SetParam("ip", c.tool.NetManager.GetIp())
			res.SetParam("port", strconv.Itoa(c.tool.NetManager.GetPort()))
			data := xsf.NewData()
			res.AppendData(data)
			return res, nil
		}
	default:
		{
			fmt.Printf("the op -> %v is not supported.\n", op)
			res.SetError(1, fmt.Sprintf("the op -> %v is not supported.", op))
			res.SetParam("intro", "received data")
			res.SetParam("op", "illegal")
			return res, nil
		}
	}
}
