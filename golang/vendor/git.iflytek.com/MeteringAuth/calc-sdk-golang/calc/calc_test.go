package calc

import (
	"testing"
)

func TestInit(t *testing.T) {
	conf := &ConfigCalc{
		isNative:      false,
		nativeLogPath: "",
	}
	err := conf.init("http://10.1.87.69:6868", "guiderAllService", "gas", "calc-client", "2.1.0")
	if err != nil {
		t.Errorf("init error = %s\n", err)
	}
}

func BenchmarkCalc(b *testing.B) {

	if err := Init("http://10.1.87.69:6868", "guiderAllService", "gas", "calc-client", "2.1.0", false, "./calc.toml"); err != nil {
		b.Fatalf("calc init error : %s\n", err)
	}
	for n := 0; n < b.N; n++ {
		code, err := Calc("testCalcSDK", "calcsdk", "nothing", 1)
		if err != nil {
			b.Errorf("errocde : %d , errorInfo = %s", code, err)
		}
	}
	Fini()
}
