package main

import (
	"context"
	"fmt"
	xsf "git.iflytek.com/AIaaS/xsf/client"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var perfMiniCount int64

func performanceJanus(ctx context.Context, cli *xsf.Client, tm time.Duration) {

	fmt.Printf("init goroutines\n")

	caller := xsf.NewCaller(cli)

	start := time.Now()
	var wg sync.WaitGroup
	for gIx := int64(0); gIx < *gNum; gIx++ {
		wg.Add(1)
		go Do(&wg, caller, tm)
	}

	go func() {
		select {
		case <-ctx.Done():
			{
				fmt.Printf("qps:%v\n", float64(atomic.LoadInt64(&perfMiniCount))/time.Now().Sub(start).Seconds())
				os.Exit(0)
			}
		}
	}()

	wg.Wait()
}

func Do(wg *sync.WaitGroup, caller *xsf.Caller, tm time.Duration) {
	defer wg.Done()

	req := xsf.NewReq()

	req.SetParam("channel", "igr")
	req.SetParam("appid", "4cc57799")
	req.SetParam("uid", "uid")
	req.SetParam("function", "igr_gray")
	req.SetParam("attribute", strconv.Itoa(int(0)))

	var err error

	for {
		_, _, err = caller.CallWithAddr("client", "lic", "10.1.87.61:1997", req, tm)
		if err != nil {
			panic(err)
		}
		if atomic.AddInt64(&perfMiniCount, 1) > *gCnt {
			break
		}
	}
}
