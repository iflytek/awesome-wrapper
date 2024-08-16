package main

import (
	"context"
	"fmt"
	"git.iflytek.com/AIaaS/xsf/client"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

func call(cli *xsf.Client, tm time.Duration) {
	c := xsf.NewCaller(cli)
	c.WithApiVersion(apiVersion)
	c.WithRetry(1)

	req := xsf.NewReq()
	req.SetParam("k1", "v1")
	req.SetParam("k2", "v2")
	req.SetParam("k3", "v3")

	r, code, e := c.Call("xsf-server", "req", req, tm)
	if code != 0 || e != nil {
		log.Fatal("sse err", code, e)
	}

	all := r.GetAllParam()
	fmt.Printf("allParams:%+v\n", all)
}
func callWrapper(cli *xsf.Client, tm time.Duration) {
	const (
		service = "xsf-server"
		op      = "req"
	)

	{
		_ = cli.ConfigureHystrix(
			service, op,
			xsf.WithCommandMaxConcurrentRequests(1),
			xsf.WithCommandRequestVolumeThreshold(1e5),
			xsf.WithCommandSleepWindow(5e3),
			xsf.WithCommandErrorPercentThreshold(10),
			xsf.WithCommandFallback(func(ctx context.Context, err error, service string, op string, r *xsf.Req) (*xsf.Res, int32, error) {
				c := xsf.NewCaller(cli)
				c.WithApiVersion(apiVersion)
				res, code, e := c.CallCtx(ctx, service, "req", r)
				return res, code, e
			}),
		)
		go func() {
			for {
				time.Sleep(time.Second * 3)
				_ = cli.ConfigureHystrix(
					service, op,
					xsf.WithCommandMaxConcurrentRequests(100),
					xsf.WithCommandRequestVolumeThreshold(1e5),
					xsf.WithCommandSleepWindow(5e3),
					xsf.WithCommandErrorPercentThreshold(10),
					xsf.WithCommandFallback(func(ctx context.Context, err error, service string, op string, r *xsf.Req) (*xsf.Res, int32, error) {
						c := xsf.NewCaller(cli)
						c.WithApiVersion(apiVersion)
						res, code, e := c.CallCtx(ctx, service, "req", r)
						return res, code, e
					}),
				)
			}
		}()
	}
	{
		c := xsf.NewCaller(cli)
		c.WithApiVersion(apiVersion)

		req := xsf.NewReq()
		req.SetParam("k1", "v1")
		req.SetParam("k2", "v2")
		req.SetParam("k3", "v3")

		cli.Log.Infow("callWrapper", "tm", tm.String())

		var wg sync.WaitGroup
		var totalCount int64
		for concurrent := 0; concurrent < 2; concurrent++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					time.Sleep(time.Second * 1)

					ctx, _ := context.WithTimeout(context.Background(), tm)
					r, code, e := c.CallCtx(ctx, "xsf-server", "req", req)
					if code != 0 || e != nil {
						fmt.Printf("NO.%d=>callCtx err=>r:%v,code:%v,e:%v\n", atomic.AddInt64(&totalCount, 1), r, code, e)
						continue
					}
					//if r == nil {
					//	fmt.Printf("r:%v,code:%v,e:%v\n", r, code, e)
					//	panic("exiting...")
					//}
					allStr := fmt.Sprintf("%+v", r.GetAllParam())
					fmt.Printf("NO.%d=>allParams:%s\n", atomic.AddInt64(&totalCount, 1), allStr)

					//if strings.Contains(allStr, "roll") {
					//	break
					//}
				}
			}()
		}
		wg.Wait()
	}
}
