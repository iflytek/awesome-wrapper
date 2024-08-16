package main

import (
	"flag"
	"git.iflytek.com/AIaaS/xsf/server"
	"log"
	"sync"
)

func main() {
	flag.Parse()

	var serverInst xsf.XsfServer

	bc := xsf.BootConfig{
		CfgMode: -1,
		CfgData: xsf.CfgMeta{
			CfgName:      "",
			Project:      "",
			Group:        "",
			Service:      "",
			Version:      "0.0.0",
			ApiVersion:   "1.0.0",
			CachePath:    "",
			CompanionUrl: ""}}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
		}()

		if err := serverInst.Run(
			bc,
			&server{}); err != nil {
			log.Panic(err)
		}
	}()
	wg.Wait()
}
