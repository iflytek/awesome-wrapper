package rep

import (
	"fmt"
	"testing"

	"git.iflytek.com/MeteringAuth/ma-sdk-golang/rep/ver"
)

func TestInit(t *testing.T) {
	if err := Init("http://10.1.87.70:6868", "metrics", "reporter", "ma-client", ver.Version, 1, "myip"); err != nil {
		t.Errorf(err.Error())
	}

	addr := "reqip"
	appid := "testa"
	sub := "testc"

	a := make(map[string]uint, 10)
	a[appid] = 12
	err := Report(sub, a, false)
	fmt.Println("report:", sub, a, err)

	err = ReportWithAddr(sub, "", a, addr)
	fmt.Println("report with addr:", sub, a, addr, err)

	b := make(map[string]string, 10)
	b[appid] = "13"
	err = Sync(sub, "", b, nil, "req_sync", addr, false)
	fmt.Println("sync:", sub, b, addr, err)
}
