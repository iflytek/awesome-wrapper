package main

import (
	"fmt"
	"git.iflytek.com/AIaaS/xsf/client"
	"log"
	"sync/atomic"
	"time"
)

func callConHashTest(c *xsf.Caller, tm time.Duration) {
	c.WithApiVersion(apiVersion)
	c.WithRetry(3)
	var count int64

	test := func(hashKey, svc string) {
		baseTime := time.Now()
		addr, addrErr := c.GetHashAddr(hashKey, svc)
		fmt.Printf("NO.%v dur:%v,addr:%v,addrErr:%v,hashKey:%v,svc:%v\n",
			time.Now().Sub(baseTime).String(), atomic.AddInt64(&count, 1), addr, addrErr, hashKey, svc)
	}
	{
		fmt.Println("------------------------------")
		hashKey, svc := "111", "xsf-server"
		test(hashKey, svc)
		test(hashKey, svc)
		test(hashKey, svc)
		test(hashKey, svc)
	}
	{
		fmt.Println("------------------------------")
		hashKey, svc := "432", "atmos-iat"
		test(hashKey, svc)
		test(hashKey, svc)
		test(hashKey, svc)
		test(hashKey, svc)
	}
	{
		req := xsf.NewReq()
		req.SetParam("k1", "v1")
		req.SetParam("k2", "v2")
		req.SetParam("k3", "v3")

		baseTime := time.Now()
		c.WithHashKey("classifyID")
		res, code, e := c.Call("xsf-server", "req", req, tm)
		dur := time.Now().Sub(baseTime)
		if code != 0 || e != nil {
			log.Fatalf("sse err,code:%v,e:%v,dur:%v\n", code, e, dur.Seconds())
		} else {
			fmt.Printf("F.NO.1 => handle:%s,dur:%vs\n", res.Handle(), dur.Seconds())
		}
	}
}