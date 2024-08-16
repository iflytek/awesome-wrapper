package test

import (
	"testing"
	"genitus/flange"
	"fmt"
	"time"
	"sync"
	"runtime"
)

func Test_Flange(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	defer flange.Fini()
	flange.DumpEnable = false
	flange.DeliverEnable = true
	flange.SpillEnable = true
	flange.Logger = &flange.FmtLog{}
	flange.BatchSize = 1000
	flange.BuffSize = 2048
	if err := flange.Init("172.16.51.29", "4545", 4, "0.1.0.1", "8088", "sIs"); err != nil {
		fmt.Println(err.Error())
	}
	wg := &sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for i := 0; i < 1000; i++ {
				test()
			}
		}()
	}
	wg.Wait()
	fmt.Println("--------flush over-------")

	time.Sleep(time.Hour)

	t.Log("ok")
}

func test() {
	root := flange.NewSpan(flange.SERVER, false).WithName("sessionbegin").Start()
	sid := fmt.Sprintf("iam56eb0006@lc%x3319210", time.Now().UnixNano()/1000/1000)
	root.WithTag("sid", sid)
	root.WithTag("appid", "100IME")
	root.WithTag("uid", "v2077407681")
	root.WithRetTag("10030")
	root.WithTag("cancel", "success")
	root.WithTag("{\"format\" : \"json\"}", "{\"appid\":\"4f0a594c\"}")
	root.WithErrorTag("no server response")
	root.WithDescf("hello %s", "world").WithDescf("hello %d", 3)

	for i := 0; i < 200; i++ {
		spanClient := root.Next(flange.CLIENT).WithName("Jdbc").Start()
		spanClient2 := flange.FromMeta(spanClient.Meta(), flange.SERVER).WithName("jdBc").Start()
		spanClient2.WithTag("sql", "select * form sis")
		flange.Flush(spanClient2.End())
		flange.Flush(spanClient.End())
	}

	flange.Flush(root.End())
}
