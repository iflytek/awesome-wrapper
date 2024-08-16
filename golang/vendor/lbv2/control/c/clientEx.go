package main

import (
	"encoding/json"
	"fmt"
	"git.xfyun.cn/AIaaS/xsf/client"
	"git.xfyun.cn/AIaaS/xsf/utils"
	"log"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

/*
ALL
NBESTTAG
LICSUBSVC
LICSVC
*/
func clientEx(cli *xsf.Client) {
	fmt.Printf("mode:clientEx\n")
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
	nbest, e := cli.Cfg().GetInt("clientEx", "nbest")
	if e != nil {
		log.Fatalf("get param nbest error:%s\n", e.Error())
	}

	uid, e := cli.Cfg().GetInt("clientEx", "uid")
	if e != nil {
		log.Fatalf("get param uid error:%s\n", e.Error())
	}

	allInt, e := cli.Cfg().GetInt("clientEx", "all")
	allBool := false
	if e == nil {
		if allInt == 1 {
			allBool = true
		}
	}
	thCnt, e := cli.Cfg().GetInt("clientEx", "thCnt")
	if e != nil {
		log.Fatalf("get param thCnt error:%s\n", e.Error())
	}
	loopCnt, e := cli.Cfg().GetInt("clientEx", "loopCnt")
	if e != nil {
		log.Fatalf("get param loopCnt error:%s\n", e.Error())
	}
	var printRst bool
	printRstInt, e := cli.Cfg().GetInt("clientEx", "print")
	if printRstInt == 1 {
		printRst = true
	}

	f, err := os.OpenFile(countFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		log.Fatalf("open file:%s err:%v\n", countFile, err)
	}
	defer f.Close()

	fmt.Println("about to start goroutines")

	timeStart := time.Now().Unix()
	var wg = &sync.WaitGroup{}

	for i := 0; i < thCnt; i++ {
		wg.Add(1)
		go func(goIndex int) {
			defer wg.Done()
			var goroutimeStartTime = time.Now().Unix()
			c := xsf.NewCaller(cli)
			req := utils.NewReq()
			req.SetParam("nbest", strconv.Itoa(nbest))
			req.SetParam("svc", svc)
			req.SetParam("subsvc", subsvc)
			req.SetParam("uid", strconv.Itoa(uid))
			if allBool {
				req.SetParam("all", "1")
			}
			for j := 0; j < loopCnt; j++ {
				cli.Log.Infof("main | c.Call ")
				res, code, e := c.Call("lbv2", "get", req, time.Duration(tm)*time.Millisecond)
				if e != nil {
					atomic.AddInt64(&failCount, 1)
					fmt.Printf("lb error is:%v,errcode is:%v\n", e, code)
					continue
				}

				dataArr := res.GetData()
				var bestNodes []string
				for _, data := range dataArr {
					bestNode := string(data.GetData())
					bestNodes = append(bestNodes, bestNode)
				}
				if printRst {
					//fmt.Println("=========================================")
					fmt.Printf("best Nodes is:%#v,errcode is:%v,bestNodesStr:%v\n", bestNodes, code, func(in []string) string { str, _ := json.Marshal(in); return string(str) }(bestNodes))
					//cmp := func(k1, k2 []string) bool {
					//	k1Json, _ := json.Marshal(k1)
					//	k2Json, _ := json.Marshal(k2)
					//	return string(k1Json) == string(k2Json)
					//}
					//if cmp(bestNodes, []string{""}) {
					//	fmt.Println("---------------------------------------")
					//	fmt.Printf("dataArr:%#v\n", dataArr)
					//	for _, v := range dataArr {
					//		fmt.Printf("dataArr_v:%#v\n", string(v.GetData()))
					//	}
					//}
				}
			}
			var goroutimeEndTime = time.Now().Unix()
			goroutimeConstTime := goroutimeEndTime - goroutimeStartTime
			content := fmt.Sprintf("goroutine:%v loopCnt:%v cost:%v second\n", goIndex, loopCnt, goroutimeConstTime)
			f.WriteString(content)
		}(i)
	}
	wg.Wait()

	timeEnd := time.Now().Unix()
	durationTime := timeEnd - timeStart

	runCnt := thCnt * loopCnt

	//计算tps
	var tps float32
	if durationTime > 0 {
		tps = float32(runCnt) / float32(durationTime)
	}
	sucCount = int64(runCnt) - failCount
	fmt.Printf("getServer goroutine cout is:%d, every goroutine loop count is:%d, all run count:%d,cost time is:%ds, tps is:%v, succGetServer count is:%d\n",
		thCnt, loopCnt, runCnt, durationTime, tps, sucCount)

	fmt.Println("all goroutines done")
}
