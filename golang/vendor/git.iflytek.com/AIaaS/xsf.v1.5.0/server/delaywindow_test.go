package xsf

import (
	"fmt"
	"testing"
	"time"
)

func Test_delayWindow_Init(t *testing.T) {
	win := newDelayWindow(1000*time.Millisecond, 10)
	go func() {
		for {
			time.Sleep(time.Millisecond * 10)
			win.setDur(20 * 1e6)
		}
	}()
	go func() {
		for {
			time.Sleep(time.Second)
			max, min, avg, qps := win.getStats()
			fmt.Printf("max:%v, min:%v, avg:%v, qps:%v\n", max, min, avg, qps)
		}
	}()
	time.Sleep(time.Hour)
}
