package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"
	"time"

	xsf "git.iflytek.com/AIaaS/xsf/server"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/rep"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/rep/ver"
)

var cnt = flag.Int("l", 1, "loop times")
var dur = flag.Int("d", 100, "sleep ms")
var appid = flag.String("a", "sdka", "appid")
var sub = flag.String("sub", "sdkc", "channel")

// var ent = flag.String("f", "sdkf", "function")

func main() {
	flag.Parse()

	if err := rep.Init(*xsf.CompanionUrl,
		*xsf.Project,
		*xsf.Group,
		*xsf.Service,
		ver.Version,
		*xsf.Mode,
		"111"); err != nil {
		log.Fatalln("report init error : ", err)
	}

	addr := strconv.Itoa(*cnt)
	for x := 0; x < *cnt || *cnt < 0; x++ {
		a := make(map[string]uint, 10)
		a[*appid] = uint(x)
		err := rep.Report(*sub, a, false)
		fmt.Println("report:", *sub, a, err)

		err = rep.ReportWithAddr(*sub, "", a, addr)
		fmt.Println("report with addr:", *sub, a, addr, err)

		b := make(map[string]string, 10)
		b[*appid] = strconv.Itoa(x)
		err = rep.Sync(*sub, "", b, nil, "req_sync", addr, false)
		fmt.Println("sync:", *sub, b, addr, err)

		if *dur > 0 {
			time.Sleep(time.Duration(*dur) * time.Millisecond)
		}
	}

	rep.Fini()
	fmt.Println("rep done")
}
