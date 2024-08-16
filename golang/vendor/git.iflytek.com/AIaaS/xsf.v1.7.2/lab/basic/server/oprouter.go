package main

import (
	"fmt"
	"git.iflytek.com/AIaaS/xsf/server"
	"git.iflytek.com/AIaaS/xsf/utils"
	"time"
)

func generateOpRouter() *xsf.OpRouter {
	router := &xsf.OpRouter{}
	router.Store("op", func(in *xsf.Req, span *xsf.Span, tool *xsf.ToolBox) (*utils.Res, error) {
		res := xsf.NewRes()
		res.SetHandle(in.Handle())
		fmt.Printf("info:this is op operator. -> timestamp:%v,Handle:%v\n", time.Now(), in.Handle())
		return res, nil
	})
	return router
}
