package licc

import (
	"fmt"
	"testing"
	"time"

	"git.iflytek.com/MeteringAuth/ma-sdk-golang/licc/ver"
)

func TestInit(t *testing.T) {
	if err := Init("http://10.1.87.70:6868", "metrics", "reporter", "ma-client", ver.Version, 1, []string{"test*"}); err != nil {
		t.Errorf(err.Error())
	}

	appid := "testa"
	sub := "testc"
	funcs := []string{"testf", "business.total"}
	did := "testd"

	result, logInfo, err := Check(appid, "", sub, funcs, time.Now().String())
	fmt.Printf("check result: %+v , logInfo: %s , err: %v\n", result, logInfo, err)

	result, logInfo, err = Check(appid, did, sub, funcs, time.Now().String())
	fmt.Printf("check did result: %+v , logInfo: %s , err: %v\n", result, logInfo, err)

	result, err = GetAcfLimits(appid, sub, funcs, time.Now().String())
	fmt.Printf("check acf limit result: %+v ,  err: %v\n", result, err)
}
