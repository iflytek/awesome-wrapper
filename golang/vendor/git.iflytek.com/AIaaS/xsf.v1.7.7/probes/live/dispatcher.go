package main

import (
	"flag"
	"strings"
)

var svcPort = flag.String("port", "", "")

func init() {
	flag.Parse()
}

func main() {
	logger.Println("version:", version)

	dispatcher()
}

func dispatcher() {
	addr := strings.Replace(getHost()+":"+*svcPort, `"`, ``, -1)
	logger.Printf("addr:%v\n", addr)

	ctx := getCtx(globalTimeout)

	healthCheck(ctx, addr)

	logger.Printf("server start ok.\n")
	logger.Println("success")
}
