package main

import (
	"flag"
	"fmt"
	"git.iflytek.com/AIaaS/xsf/utils"
)

var (
	h = flag.String("host", "", "")
	n = flag.String("netcard", "", "")
)

func main() {
	flag.Parse()
	hostIn, netCardIn := *h, *n
	fmt.Printf("about to call utils.HostAdapter host:%v,netcard:%v\n", hostIn, netCardIn)
	fmt.Println(utils.HostAdapter(hostIn, netCardIn))
}
