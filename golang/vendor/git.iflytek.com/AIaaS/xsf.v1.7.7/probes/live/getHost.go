package main

import (
	"git.iflytek.com/AIaaS/xsf/utils"
)

func getHost() string {
	hostIn, netCardIn := "", ""
	logger.Printf("about to call utils.HostAdapter host:%v,netcard:%v\n", hostIn, netCardIn)
	h, e := utils.HostAdapter(hostIn, netCardIn)
	checkErr(e)
	logger.Printf("host:%s\n", h)
	return h
}
