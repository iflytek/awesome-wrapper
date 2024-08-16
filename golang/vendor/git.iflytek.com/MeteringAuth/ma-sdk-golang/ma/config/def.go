package config

import (
	"os"
	"sync"

	"github.com/jinzhu/configor"
)

type maCfg struct {
	Metrics struct {
		Able        int
		IDC         string `toml:"idc" default:"unknownidc"`
		SUB         string `toml:"sub" default:"unknownsub"`
		MonitorSrv  string `toml:"ma_service" default:"unknownsrv"`
		MonitorSize int    `toml:"ma_pool_size" default:"1234567"`
	}
	Calc struct {
		AsyncInit bool   `toml:"async_init"`
		InitRetry int    `toml:"init_retry"`
		Use       string `toml:"use" default:"mq"`
		QueueSize int    `toml:"queue_size" default:"2345678"`
		RMQ       struct {
			Able          bool
			ConsumeNumber int `toml:"producer_number"`
			Endpoint      []string
			Topic         string
			Timeout       int64
		}
		Pulsar struct {
			Able      bool
			IDC       string
			Appids    []string
			Topic     string
			Endpoint  string
			ThreadNum int `toml:"thread_num" default:"10"`
		}
	}
	Licc struct {
		AsyncInit bool `toml:"async_init"`
		InitRetry int  `toml:"init_retry"`
	}
	Rep struct {
		AsyncInit bool `toml:"async_init"`
		InitRetry int  `toml:"init_retry"`
	}
	Log struct {
		Level string `toml:"level"`
	}

	Conc struct {
		OnlyUseAqc     bool     `toml:"only_use_aqc"`
		WhiteAppidList []string `toml:"white_appid_list"`
		BatchMaxSize   int      `toml:"batch_max_size"`
		BufferSize     int      `toml:"buffer_size"`
		Worker         int      `toml:"worker" default:"1"`
	}
}

var C *maCfg

func initCfg(data []byte) error {
	sp.Println("---- config begin ----")

	tmpfile := "./ma-sdk.cfg.toml"
	err := os.WriteFile(tmpfile, data, os.FileMode(0644))
	if err != nil {
		return err
	}

	tmp := new(maCfg)
	if err = configor.Load(tmp, tmpfile); err != nil {
		sp.Println("---- config error ----")
		return err
	}

	m := make(map[string]struct{})
	if tmp.Calc.Pulsar.Able {
		for _, app := range tmp.Calc.Pulsar.Appids {
			m[app] = struct{}{}
		}
	}
	sp.Println("pulsar appid:", m, "able:", tmp.Calc.Pulsar.Able)
	sp.Printf("cfg: %+v\n", tmp)

	checkMu.Lock()
	checkPulsarAppids = m
	checkMu.Unlock()

	C = tmp

	sp.Println("---- config done ----")
	return nil
}

var checkPulsarAppids map[string]struct{}
var checkMu sync.RWMutex

func UsePulsar(appid string) (ok bool) {
	checkMu.RLock()
	if _, ok = checkPulsarAppids["*"]; !ok {
		_, ok = checkPulsarAppids[appid]
	}
	checkMu.RUnlock()
	return
}
