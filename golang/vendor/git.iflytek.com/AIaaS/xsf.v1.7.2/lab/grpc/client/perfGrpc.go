package main

import (
	"fmt"
	pb "git.iflytek.com/AIaaS/xsf/lab/grpc/common"
	"github.com/cihub/seelog"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func threadCnt() int {
	checkErr := func(err error) {
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	threadStats, ssErr := exec.Command("ps", "hH", strconv.Itoa(os.Getpid())).Output()
	checkErr(ssErr)
	return strings.Count(string(threadStats), "\n")
}
func performanceGrpc(ctx context.Context, tm time.Duration) {
	var (
		perfGrpc int64
	)
	if *gInfo == 1 {
		_ = seelog.Error("---------------------------------\n")
		baseTime := time.Now()
		count := 0
		go func() {
			for {
				time.Sleep(time.Millisecond * 500)
				count++
				_ = seelog.Errorf("NO.%d,elapsed:%v,threads:%v,gorouties:%v\n", count, time.Now().Sub(baseTime).String(), threadCnt(), runtime.NumGoroutine())
			}
		}()
	}
	for ix := 0; ix < *pre; ix++ {
		fmt.Printf("NO.%d pre init\n", ix)
		t, e := grpc.Dial(
			fmt.Sprintf("%v:%v", *host, *port),
			grpc.WithInsecure(),
			grpc.WithReadBufferSize((*rBuf)*1024),
			grpc.WithWriteBufferSize((*wBuf)*1024),
		)
		if e != nil {
			panic(e)
		}
		if *gClose == 1 {
			t.Close()
		}
	}
	for ix := int64(0); ix < *gIdle; ix++ {
		fmt.Printf("NO.%d idle init\n", ix)
		go func() {
			for {
				time.Sleep(time.Millisecond * 100)
			}
		}()
	}
	conn, err := grpc.Dial(
		fmt.Sprintf("%v:%v", *host, *port),
		grpc.WithInsecure(),
		grpc.WithReadBufferSize((*rBuf)*1024),
		grpc.WithWriteBufferSize((*wBuf)*1024),
	)
	if err != nil {
		log.Panic("did not connect: ", err)
	}
	defer conn.Close()
	c := pb.NewGrpcCallClient(conn)
	_, err = c.SimpleCall(context.Background(), &pb.ReqData{})

	start := time.Now()
	var wg sync.WaitGroup
	for gIx := int64(0); gIx < *gNum; gIx++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				ctxTm, _ := context.WithTimeout(context.Background(), tm)
				_, err = c.SimpleCall(ctxTm, &pb.ReqData{})
				if err != nil {
					panic(err)
				}
				if atomic.LoadInt64(&perfGrpc) == *qpsStart {
					start = time.Now()
				}
				if atomic.AddInt64(&perfGrpc, 1) > *gCnt {
					break
				}
			}
		}()
	}

	go func() {
		select {
		case <-ctx.Done():
			{
				fmt.Printf("qps:%v\n", float64(atomic.LoadInt64(&perfGrpc)-(*qpsStart))/time.Now().Sub(start).Seconds())
				os.Exit(0)
			}
		}
	}()

	wg.Wait()
	fmt.Printf("qps:%v\n", float64(atomic.LoadInt64(&perfGrpc)-(*qpsStart))/time.Now().Sub(start).Seconds())
}
