package main

import (
	"context"
	"fmt"
	xsf "git.iflytek.com/AIaaS/xsf/client"
	"github.com/cihub/seelog"
	"log"
	"math"
	"os"
	"sync"
	"time"
)

//reduce atomic
func performanceHystrix(ctx context.Context, cli *xsf.Client, tm time.Duration) {
	fmt.Println("enter performanceXrpc...")

	const displayThreshold = 100

	type stats struct {
		rst    []time.Duration
		start  time.Time
		taskId int64
	}
	var statsChan = make(chan stats, (*gNum)*(*gCnt)*2)

	analysis := func(wg *sync.WaitGroup) {
		defer wg.Done()

		var statsSlice []stats

		{
			fmt.Printf("collecting stats from goroutines\n")
			var exitCnt int64
			//汇总数据集
			for statsItem := range statsChan {

				//去重
				alreadyRepeat := false
				for _, v := range statsSlice {
					if statsItem.taskId == v.taskId {
						alreadyRepeat = true
						break
					}
				}

				if !alreadyRepeat {
					statsSlice = append(statsSlice, statsItem)

					exitCnt++
					if exitCnt >= *gNum {
						break
					}
				}
			}
		}

		//保留结束时间，此处有偏差
		endTime := time.Now()
		fmt.Printf("goroutines endTime:%v\n", endTime)

		//取最小时间
		var earlyStart time.Time
		{
			for ix, val := range statsSlice {
				if ix == 0 {
					earlyStart = val.start
				}
				if earlyStart.After(val.start) {
					earlyStart = val.start
				}
			}
			fmt.Printf("goroutines earlyTime:%v\n", earlyStart)
		}

		{
			fmt.Printf("analysis data...\n")
			var (
				total   int64
				_1ms    int64
				_5ms    int64
				_10ms   int64
				_100ms  int64
				_1000ms int64
				_other  int64

				max = -0.1
				min = math.MaxFloat64
				sum = 0.0
			)
			for _, statsItem := range statsSlice {
				for _, dur := range statsItem.rst {

					if dur == 0 {
						continue
					}

					total++

					sum += dur.Seconds()

					if dur.Seconds() > max {
						max = dur.Seconds()
					}

					if dur.Seconds() < min {
						min = dur.Seconds()
					}

					//<1ms
					if dur < time.Millisecond {
						_1ms++
						continue
					}
					//1ms-5ms
					if dur < time.Millisecond*5 {
						_5ms++
						continue
					}
					//5ms-10ms
					if dur < time.Millisecond*10 {
						_10ms++
						continue
					}
					//10ms-100ms
					if dur < time.Millisecond*100 {
						_100ms++
						continue
					}
					//100ms-1000ms
					if dur < time.Millisecond*1000 {
						_1000ms++
						continue
					}
					//>1000ms
					_other++
				}
			}
			{
				fmt.Printf("---------------------------------------\n")
				fmt.Printf("start:%v\nend:%v\n", earlyStart, endTime)
				fmt.Printf("total:%v,goroutines:%v,qps:%.10f\n", total, *gNum, float64(total)/endTime.Sub(earlyStart).Seconds())
				fmt.Printf("elapsed:%.10fs,max:%.10fs,min:%.10fs,avg:%.10fs\n", endTime.Sub(earlyStart).Seconds(), max, min, sum/float64(total))
				fmt.Printf("<1ms | cnt:%v,rate:%.10f\n", _1ms, float64(_1ms)/float64(total))
				fmt.Printf("1ms-5ms | cnt:%v,rate:%.10f\n", _5ms, float64(_5ms)/float64(total))
				fmt.Printf("5ms-10ms | cnt:%v,rate:%.10f\n", _10ms, float64(_10ms)/float64(total))
				fmt.Printf("10ms-100ms | cnt:%v,rate:%.10f\n", _100ms, float64(_100ms)/float64(total))
				fmt.Printf("100ms-1000ms | cnt:%v,rate:%.10f\n", _1000ms, float64(_1000ms)/float64(total))
				fmt.Printf(">1000ms | cnt:%v,rate:%.10f\n", _other, float64(_other)/float64(total))
				fmt.Printf("<1ms:%v,1-5ms:%v,5-10ms:%v,10-100ms:%v,100-1000ms:%v,>1000ms:%v\n",
					float64(_1ms)/float64(total), float64(_5ms)/float64(total), float64(_10ms)/float64(total), float64(_100ms)/float64(total),
					float64(_1000ms)/float64(total), float64(_other)/float64(total))
			}
			{
				seelog.Flush()
				os.Exit(0)
			}
		}

	}

	do := func(wg *sync.WaitGroup, taskId int64) {
		defer wg.Done()
		c := xsf.NewCaller(cli)
		req := xsf.NewReq()
		req.SetParam("key", "val")

		var goroutineCnt int64
		var goroutineTotal = *gCnt

		rstStats := stats{
			rst:    make([]time.Duration, goroutineTotal+2),
			start:  time.Now(),
			taskId: taskId,
		}

		{
			go func() {
				select {
				case <-ctx.Done():
					if *gNum < displayThreshold {
						fmt.Printf("NO.%v goroutine receive exit signal\n", taskId)
					}

					statsChan <- rstStats
				}
			}()
		}

		{
			var startT time.Time
			var err error
			for {
				if goroutineCnt > goroutineTotal {
					break
				}
				goroutineCnt++
				startT = time.Now()
				//_, _, err = c.Call("xsf-server", "req", req, tm)
				ctx, _ := context.WithTimeout(context.Background(), tm)
				_, _, err = c.CallCtx(ctx, "xsf-server", "req", req)
				if err != nil {
					log.Fatalf("tm dur:%v,err:%v\n", time.Since(startT).String(), err)
				}
				rstStats.rst[goroutineCnt] = time.Now().Sub(startT)
			}

			if *gNum < displayThreshold {
				fmt.Printf("NO.%v goroutine complete\n", taskId)
			}

			statsChan <- rstStats
		}
	}
	{
		_ = cli.ConfigureHystrix(
			"xsf-server", "req",
			xsf.WithCommandMaxConcurrentRequests(1e4),
			xsf.WithCommandRequestVolumeThreshold(1e8),
			xsf.WithCommandSleepWindow(5e3),
			xsf.WithCommandErrorPercentThreshold(10),
			xsf.WithCommandFallback(func(ctx context.Context, err error, service string, op string, r *xsf.Req) (*xsf.Res, int32, error) {
				panic("wrong!!!")
				c := xsf.NewCaller(cli)
				res, code, e := c.CallCtx(ctx, service, "req", r)
				return res, code, e
			}),
		)
	}
	{
		fmt.Printf("begin to init goroutines\n")
		var wg sync.WaitGroup
		for gIx := int64(0); gIx < *gNum; gIx++ {
			wg.Add(1)
			if *gNum < displayThreshold {
				fmt.Printf("start NO.%v goroutine\n", gIx)
			}
			go do(&wg, gIx)
		}

		wg.Add(1)
		go analysis(&wg)

		wg.Wait()
	}
}
