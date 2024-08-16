package calc

import (
	"fmt"
	"testing"
)

func TestInit(t *testing.T) {
	if err := Init("http://10.1.87.69:6868", "llhou5", "meteringauth", "janus-client", "2.0.6", 1); err != nil {
		t.Errorf(err.Error())
	}

	appid := "testredis2"
	channel := "test_redis_update2"
	function := "testredisfunc"
	did := "useruid"

	code, err := Calc(appid, channel, function, 1)
	fmt.Println(appid, channel, function, 1, "->", code, err)

	code, err = CalcWithSubId(appid, did, channel, function, 1)
	fmt.Println(appid, did, channel, function, 1, "->", code, err)

	Fini()
	fmt.Println("program terminated")
}
