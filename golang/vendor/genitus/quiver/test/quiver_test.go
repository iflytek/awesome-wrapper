package test

import (
	"testing"
	"genitus/quiver"
	"time"
	"fmt"
)

func Test_Quiver(t *testing.T) {
	defer quiver.Fini()

	quiver.Logger = &quiver.FmtLog{}
	quiver.DumpEnable = true
	quiver.DeliverEnable = true
	quiver.SpillEnable = true

	quiver.Init("10.1.86.58", "4545",
		2, "ak", "sk", "10.1.86.58",
		"10.1.86.58")

	for i := 0; i < 10; i++ {
		event := GetEvent()
		if event.Flush("aue", "auf", "rate") != nil {
			t.Fatalf("flush err")
		}
	}

	fmt.Println("exit")
	time.Sleep(time.Second * 20)
	fmt.Println("minute end")
}
