package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	xsf "git.iflytek.com/AIaaS/xsf/server"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/licc"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/licc/ver"
)

var cnt = flag.Int("l", 1, "loop times")
var dur = flag.Int("d", 100, "sleep ms")
var appid = flag.String("a", "testaqc", "appid")
var uid = flag.String("uid", "testaqcuid", "uid")
var sub = flag.String("sub", "testaqcsub", "channel")
var ent = flag.String("f", "testfunc", "function")
var channel = flag.String("init", "testaqcchannel", "init channel")

func main() {
	flag.Parse()

	funcs := strings.Split(*ent, ";")

	if err := licc.Init(*xsf.CompanionUrl,
		*xsf.Project,
		*xsf.Group,
		*xsf.Service,
		ver.Version,
		*xsf.Mode,
		strings.Split(*channel, ";")); err != nil {
		log.Fatalln(err)
		return
	}

	fmt.Println(*channel, *appid, *uid, *sub, *channel, funcs)

	for x := 0; x < *cnt || *cnt < 0; x++ {
		fmt.Println(x + 1)

		result, logInfo, err := licc.Check(*appid, "", *sub, funcs, time.Now().String())
		fmt.Printf("check result: %+v , logInfo: %s , err: %v\n", result, logInfo, err)

		result, logInfo, err = licc.Check(*appid, *uid, *sub, funcs, time.Now().String())
		fmt.Printf("check did result: %+v , logInfo: %s , err: %v\n", result, logInfo, err)

		result, err = licc.GetAcfLimits(*appid, *sub, funcs, time.Now().String())
		fmt.Printf("check acf limit result: %+v ,  err: %v\n", result, err)
		if *dur > 0 {
			time.Sleep(time.Duration(*dur) * time.Millisecond)
		}
	}

	licc.Fini()
	fmt.Println("licc done")
}
