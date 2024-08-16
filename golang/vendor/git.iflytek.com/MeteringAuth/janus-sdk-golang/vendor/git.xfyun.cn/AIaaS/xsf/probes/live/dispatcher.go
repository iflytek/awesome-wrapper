package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"git.xfyun.cn/AIaaS/xsf/utils"
	"google.golang.org/grpc"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	v = flag.String("v", "0", "1:debug,2:ver")
)

const (
	debug      = "1"
	displayVer = "2"

	timeout = time.Millisecond * 200

	liveProc = "LIVE_PROC"
	livePort = "LIVE_PORT"
)

func main() {
	flag.Parse()

	if displayVer == *v {
		fmt.Println("ver:1.1.0")
		os.Exit(0)
	}

	healthProbes()
}

func healthProbes() {
	addr := func() string {
		h, e := utils.HostAdapter("", "")
		if e != nil {
			moreMsg("take host from hostname failed\n")
			os.Exit(1)
		}
		return h
	}() + ":" + func() string {
		if "" != os.Getenv(livePort) {
			moreMsg("take port from env LIVE_PORT\n")
			return os.Getenv(livePort)
		}
		proc := os.Getenv(liveProc)
		if "" == proc {
			moreMsg("the proc:%v is empty string\n", proc)
			fmt.Println("failure")
			os.Exit(1)
		}
		ports := getPort(proc)
		moreMsg("states of calling getPort,ports:%#v,proc:%v\n", ports, proc)
		if len(ports) <= 1 {
			moreMsg("the ports is empty\n")
			fmt.Println("failure")
			os.Exit(1)
		}
		return ports[0]
	}()
	moreMsg("addr:%v\n", addr)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	conn, connErr := grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithBlock())
	if connErr != nil {
		cancel()
		moreMsg("dial %v fail,err:%v\n", addr, connErr)
		fmt.Println("failure")
		os.Exit(1)
	}
	defer conn.Close()

	c := utils.NewToolBoxClient(conn)

	query, _ := json.Marshal(map[string]string{"cmd": "health"})
	header, _ := json.Marshal(map[string]string{"method": "GET"})
	moreMsg("cmdServer.cmd:%v,method:%v\n", "health", "GET")

	ctx, cancel = context.WithTimeout(context.Background(), timeout)
	cmdServerResp, cmdServerRespErr := c.Cmdserver(ctx, &utils.Request{Query: string(query), Headers: string(header), Body: ""})
	if cmdServerRespErr != nil {
		cancel()
		moreMsg("request cmdServer fail,err:%v\n", cmdServerRespErr)
		fmt.Println("failure")
		os.Exit(1)
	}
	moreMsg("states of calling cmdServer,cmdServerResp.Body:%v,cmdServerRespErr:%v\n", string(cmdServerResp.Body), cmdServerRespErr)
	if strings.Contains(strings.ToLower(string(cmdServerResp.Body)), "err") {
		moreMsg("cmdServerResp.Body contain err\n")
		fmt.Println("failure")
		os.Exit(1)
	}
	moreMsg("server start ok.\n")
	fmt.Println("success")

}
func getPort(p string) []string {
	checkErr := func(err error) {
		if err != nil {
			moreMsg("fn:%v,err:%v\n", "getPort", err)
			fmt.Println("failure")
			os.Exit(1)
		}
	}
	ssRst, ssErr := exec.Command("ss", "-tlnp").Output()
	checkErr(ssErr)

	grepCmd := exec.Command("grep", p)
	grepCmd.Stdin = bytes.NewReader(ssRst)
	grepRst, grepErr := grepCmd.Output()
	checkErr(grepErr)

	awkCmd := exec.Command("awk", "{print $4}")
	awkCmd.Stdin = bytes.NewReader(grepRst)
	awkRst, awkErr := awkCmd.Output()
	checkErr(awkErr)

	awkCmd = exec.Command("awk", "-F", ":", "{print $NF}")
	awkCmd.Stdin = bytes.NewReader(awkRst)
	awkRst, awkErr = awkCmd.Output()
	checkErr(awkErr)
	return strings.Split(string(awkRst), "\n")
}
func moreMsg(format string, a ...interface{}) {
	if debug == *v {
		fmt.Printf(format, a...)
	}
}
