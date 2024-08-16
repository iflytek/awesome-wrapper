package main

import "strings"

func main() {
	logger.Println("version:", version)

	dispatcher()
}

func dispatcher() {
	addr := strings.Replace(getHost()+":"+string(getPort(getProc())), `"`, ``, -1)
	logger.Printf("addr:%v\n", addr)

	ctx := getCtx(globalTimeout)

	healthCheck(ctx, addr)

	logger.Printf("server start ok.\n")
	logger.Println("success")
}
