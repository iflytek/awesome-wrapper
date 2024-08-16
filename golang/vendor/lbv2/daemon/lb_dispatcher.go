package daemon

import (
	"flag"
	"fmt"
	"git.xfyun.cn/AIaaS/xsf/server"
	"git.xfyun.cn/AIaaS/xsf/utils"
	"log"
	"os"
)

func Usage() {
	fmt.Fprint(os.Stderr, "Usage of ", os.Args[0], ":\n")
	flag.PrintDefaults()
	fmt.Fprint(os.Stderr, "\n")
}

func RunServer() (err error) {
	flag.Usage = Usage
	flag.Parse()

	var serverInst xsf.XsfServer

	if err = serverInst.Run(xsf.BootConfig{CfgMode: utils.CfgMode(-1), CfgData: xsf.CfgMeta{CfgName: "", Project: "",
		Group: "", Service: "", Version: LBVERSION, CompanionUrl: ""}}, &Server{}); err != nil {
		log.Fatal(err)
	}
	return
}

type killed struct {
}

func (k *killed) Closeout() {
	fmt.Println("lb is be killed")
}

type healthChecker struct {
}

func (h *healthChecker) Check() (err error) {
	return nil
}

type Server struct {
	toolbox *xsf.ToolBox
	LbHandle
}

func (s *Server) Init(toolbox *xsf.ToolBox) (err error) {
	debugInt, debugErr := toolbox.Cfg.GetInt(BO, DEBUG)
	fmt.Printf("Server.Init -> debugInt:%v, debugErr:%v\n", debugInt, debugErr)
	if debugErr == nil {
		if debugInt == 1 {
			fmt.Printf("Server.Init -> debugInt:%v\n", debugInt)
			debugInst.Init(true)
		}
	}
	debugInst.Debug("about to Server.init")
	s.toolbox = toolbox
	xsf.AddKillerCheck("Server", &killed{})
	xsf.AddHealthCheck("Server", &healthChecker{})
	err = s.LbHandle.Init(s.toolbox)
	if err != nil {
		s.toolbox.Log.Errorf("lb handle init error:%v", err.Error())
		return err
	}
	return nil
}

func (s *Server) Finit() error {
	return nil
}

func (s *Server) Call(in *xsf.Req, span *xsf.Span) (res *utils.Res, err error) {
	res = xsf.NewRes()
	switch s.strategy {
	case lic:
		{
			res, err = s.worker.serve(in, span, s.toolbox)
		}
	case licEx:
		{
			res, err = s.worker.serve(in, span, s.toolbox)
		}
	default:
		{
			s.toolbox.Log.Errorf("LbErrInputStrategy")
			res.SetError(ErrLbStrategyIsNotSupport.errCode, ErrLbStrategyIsNotSupport.errInfo)
		}
	}
	return res, err
}
