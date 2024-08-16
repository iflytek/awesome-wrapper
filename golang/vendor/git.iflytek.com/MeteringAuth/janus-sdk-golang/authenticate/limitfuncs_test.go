package authenticate

import (
	"fmt"
	xsf "git.xfyun.cn/AIaaS/xsf/client"
	"log"
	"testing"
	"time"
)

var (
	lfm *LimitFuncsManager
)

func TestMain(m *testing.M) {
	DefaultInitOption = &InitOption{
		companionUrl:   "http://10.1.87.69:6868",
		project:        "metrics",
		group:          "reporter",
		service:        "janus-client",
		version:        "2.0.6",
		isCacheService: false,
		isCacheConfig:  false,
		cachePath:      "./janus-sdk-cache",
		cfgMode:        1,
	}
	xsfClient, err := newXsfClient(APP_CONFIG, CnameJanus)
	if err != nil {
		log.Fatal("new xsf client failed", err)
	}

	lfm = &LimitFuncsManager{
		channel:    []string{"passivefeaonline"},
		xsfc:       xsfClient,
		updateTime: 5000 * time.Millisecond,
		p: xsfParam{
			rpcTimeout: 200 * time.Millisecond,
			svc:        SVC,
			op:         OPGETLIMITREG,
		},
		caller:     xsf.NewCaller(xsfClient),
		limitFuncs: make(map[string]*FuncsInfo),
		Md5Map:     make(map[string]string),
	}
	m.Run()
}

func BenchmarkGetLimitFuncs(b *testing.B) {
	for i := 0; i < b.N; i++ {
		lfm.update(".*")
	}
}

func TestGetLimitFuncs(t *testing.T) {
	printLimitFuncs := func() {
		for k, v := range lfm.limitFuncs {
			fmt.Println("channel = ", k)
			for n := range v.funcs {
				fmt.Println("normal funcs = ", n)
			}
			for w := range v.wildFuncs {
				fmt.Println("wild funcs = ", w)
			}
		}
		fmt.Println("==========================================")
	}
	lfm.update(".*")
	printLimitFuncs()
	time.Sleep(10 * time.Second)
	lfm.update(".*")
	printLimitFuncs()
}
