package main

import (
	"fmt"
	"git.xfyun.cn/AIaaS/xsf/client"
	"git.xfyun.cn/AIaaS/xsf/utils"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

/*
LICADDR
LICSVC
LICSUBSVC
LICTOTAL
LICIDLE
LICBEST
*/
func reporterEx(cli *xsf.Client) {
	fmt.Printf("mode:reporterEx\n")
	fmt.Println("about to read config")
	tm, e := cli.Cfg().GetInt("lb_ctl", "timeout")
	if e != nil {
		tm = 500
	}
	println("set time out is:", tm)
	svc, e := cli.Cfg().GetString("dispatcher", "svc")
	if e != nil {
		log.Fatalf("get param svc error:%s\n", e.Error())
	}
	subsvc, e := cli.Cfg().GetString("dispatcher", "subsvc")
	if e != nil {
		log.Fatalf("get param subsvc error:%s\n", e.Error())
	}
	loopCnt, e := cli.Cfg().GetInt("reporterEx", "loopCnt")
	if e != nil {
		log.Fatalf("get param loopCnt error:%s\n", e.Error())
	}
	intervalInt, e := cli.Cfg().GetInt("reporterEx", "interval")
	if e != nil {
		log.Fatalf("get param interval error:%s\n", e.Error())
	}
	addrString, e := cli.Cfg().GetString("reporterEx", "nodes")
	if e != nil {
		log.Fatalf("get param addr error:%s\n", e.Error())
	}
	fmt.Println("about to start goroutines")
	var wg sync.WaitGroup
	for _, addr := range strings.Split(addrString, ",") {
		wg.Add(1)
		go func(in string) {
			defer wg.Done()
			total, e := cli.Cfg().GetInt(in, "total")
			if e != nil {
				log.Fatalf("get param total from addr:%v fail", in)
			}
			idle, e := cli.Cfg().GetInt(in, "idle")
			if e != nil {
				log.Fatalf("get param idle from addr:%v fail", in)
			}
			best, e := cli.Cfg().GetInt(in, "best")
			if e != nil {
				log.Fatalf("get param best from addr:%v fail", in)
			}
			c := xsf.NewCaller(cli)
			req := utils.NewReq()
			req.SetParam("svc", svc)
			req.SetParam("subsvc", subsvc)
			req.SetParam("addr", in)
			req.SetParam("total", strconv.Itoa(total))
			req.SetParam("best", strconv.Itoa(best))

			for ix := 0; ix < loopCnt; ix++ {
				req.SetParam("idle", fmt.Sprintf("%v", idle))
				_, _, e = c.Call("lbv2", "set", req, time.Duration(tm)*time.Millisecond)
				if e != nil {
					log.Fatal(e)
				}
				if intervalInt > 0 {
					time.Sleep(time.Millisecond * time.Duration(intervalInt))
				}
			}
		}(addr)
	}
	wg.Wait()
	fmt.Println("all goroutines done")
}
