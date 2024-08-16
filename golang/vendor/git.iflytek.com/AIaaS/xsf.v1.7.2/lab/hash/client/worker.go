package main

import (
	"fmt"
	"git.iflytek.com/AIaaS/xsf/client"
	"git.iflytek.com/AIaaS/xsf/utils"
	"log"
	"time"
)

func callWithHash(c *xsf.Caller, tm time.Duration) {
	span := utils.NewSpan(utils.CliSpan).Start()
	defer span.End().Flush()

	span = span.WithName("callExample")
	span = span.WithTag("customKey1", "customVal1")
	span = span.WithTag("customKey2", "customVal2")
	span = span.WithTag("customKey3", "customVal3")

	c.WithApiVersion(apiVersion)
	c.WithRetry(3)
	c.WithLBParams("xsf-lbv2", "iat", nil)

	{ //第一组测试
		req := xsf.NewReq()
		req.SetParam("k1", "v1")
		req.SetParam("k2", "v2")
		req.SetParam("k3", "v3")

		req.SetTraceID(span.Meta()) //将span信息带到后端
		baseTime := time.Now()
		res, code, e := c.Call("xsf-server", "req", req, tm)
		dur := time.Now().Sub(baseTime)
		if code != 0 || e != nil {
			log.Fatalf("sse err,code:%v,e:%v,dur:%v\n", code, e, dur.Seconds())
		} else {
			fmt.Printf("F.NO.1 => handle:%s,dur:%vs\n", res.Handle(), dur.Seconds())
		}

		res, code, e = c.Call("xsf-server", "req", req, tm)
		if code != 0 || e != nil {
			log.Fatal("sse err", code, e)
		} else {
			fmt.Printf("F.NO.2 => handle:%s\n", res.Handle())
		}

		res, code, e = c.Call("xsf-server", "req", req, tm)
		if code != 0 || e != nil {
			log.Fatal("sse err", code, e)
		} else {
			fmt.Printf("F.NO.3 => handle:%s\n", res.Handle())
		}
	}

	{ //第二组测试
		//c.WithHashKey("555")
		req := xsf.NewReq()
		req.SetParam("k1", "v1")
		req.SetParam("k2", "v2")
		req.SetParam("k3", "v3")

		req.SetTraceID(span.Meta()) //将span信息带到后端

		res, code, e := c.Call("xsf-server", "req", req, tm)
		if code != 0 || e != nil {
			log.Fatal("sse err", code, e)
		} else {
			fmt.Printf("S.NO.1 => handle:%s\n", res.Handle())
		}

		res, code, e = c.Call("xsf-server", "req", req, tm)
		if code != 0 || e != nil {
			log.Fatal("sse err", code, e)
		} else {
			fmt.Printf("S.NO.2 => handle:%s\n", res.Handle())
		}

		res, code, e = c.Call("xsf-server", "req", req, tm)
		if code != 0 || e != nil {
			log.Fatal("sse err", code, e)
		} else {
			fmt.Printf("S.NO.3 => handle:%s\n", res.Handle())
		}
	}
}
