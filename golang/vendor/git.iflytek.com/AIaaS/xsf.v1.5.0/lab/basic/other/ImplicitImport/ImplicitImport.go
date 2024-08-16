package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)
import _ "git.iflytek.com/AIaaS/xsf/server"

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		var cnt int
	end:
		for {
			c := make(chan os.Signal)
			signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGPIPE)
			s := <-c
			switch s {
			case syscall.SIGTERM, syscall.SIGINT:
				{
					fmt.Printf("ts:%v,receive syscall.SIGTERM\n", time.Now())
					cnt++
					if cnt >= 3 {
						break end
					}
				}
			}
		}
	}()
	wg.Wait()
}
