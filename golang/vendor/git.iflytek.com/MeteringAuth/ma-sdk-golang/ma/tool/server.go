package tool

import (
	"time"

	xsf "git.iflytek.com/AIaaS/xsf/server"
	"git.iflytek.com/AIaaS/xsf/utils"
	cp "git.iflytek.com/HY_trainee/colorPrinter"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/ma/config"
)

var (
	sp = cp.NewctPrinter("ma-sdk-server", cp.Red)
	//_server_ *xsf.XsfServer
	inited = make(chan struct{})
	errCh  = make(chan error)
)

func Init(url, pro, gro, service, version string, mode int, to time.Duration) (err error) {
	//if _server_ != nil {
	//	return nil
	//}
	//

	if to <= 0 {
		to = 5 * time.Second
	}
	//tk := time.NewTicker(to)

	//_server_ = &xsf.XsfServer{}

	if err = config.InitCfg(url, pro, gro, service, version, mode); err != nil {
		return err
	}

	L, err = utils.NewLocalLog(utils.SetFileName("/log/server/janus-client.log"), utils.SetLevel("error"), utils.SetAsync(true))
	//go func() {
	//	if err := _server_.Run(
	//		xsf.BootConfig{
	//			CfgMode: utils.CfgMode(mode),
	//			CfgData: xsf.CfgMeta{
	//				CfgName:      config.CfgName,
	//				Project:      pro,
	//				Group:        gro,
	//				Service:      service,
	//				Version:      version,
	//				CompanionUrl: url,
	//				CachePath:    config.CfgCacheDir,
	//				CallBack: func(*utils.Configure) bool {
	//					return true
	//				},
	//			},
	//		},
	//		&Server{},
	//		xsf.SetOpRouter(nil)); err != nil {
	//		errCh <- err
	//	}
	//}()

	//sp.Println("waiting for server init...")

	//select {
	//case <-tk.C:
	//	err = fmt.Errorf("timeout: %s", to.String())
	//case err = <-errCh:
	//case <-inited:
	//}
	//
	//if err != nil {
	//	sp.Println("server init failed:", err)
	//}

	return
}

type Server struct{}

func (s *Server) Call(*utils.Req, *utils.Span) (*utils.Res, error) {
	return nil, nil
}

//func (s *Server) OnMessage(stream *xsf.MessageStream) {
//	return
//}

func (s *Server) Init(tb *xsf.ToolBox) error {
	L = tb.Log

	inited <- struct{}{}
	return nil
}

func (s *Server) Finit() error {
	return nil
}
