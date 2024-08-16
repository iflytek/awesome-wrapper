package calc

import (
	"fmt"
	"git.iflytek.com/MeteringAuth/calc-sdk-golang/calc/internal/logger"
	"git.xfyun.cn/AIaaS/finder-go"
	common "git.xfyun.cn/AIaaS/finder-go/common"
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"net"
	"os"
	"time"
)

type ConfigCalc struct {
	polarisInited bool
	isNative      bool
	nativeLogPath string
	url           string
	project       string
	group         string
	service       string
	version       string
}

func (c *ConfigCalc) loadConfigFromNative() error {
	byt, err := ioutil.ReadFile(c.nativeLogPath)
	if err != nil {
		return err
	}
	_, err = toml.Decode(string(byt), &tomlconfig)
	if err != nil {
		return err
	}
	c.load()
	return nil

}

func (c *ConfigCalc) loadConfigFromPolaris() error {
	fc := common.BootConfig{
		CompanionUrl:  c.url,
		CachePath:     "./calc-cache",
		CacheService:  false,
		CacheConfig:   false,
		ExpireTimeout: 5 * time.Second,
		MeteData: &common.ServiceMeteData{
			Project: c.project,
			Group:   c.group,
			Service: c.service,
			Version: c.version,
			Address: func() string {
				hostName, _ := os.Hostname()
				addr, _ := net.LookupHost(hostName)
				if len(addr) == 0 {
					os.Exit(-1)
				}
				return addr[0]
			}(),
		},
	}

	f, err := finder.NewFinderWithLogger(fc, nil)
	if err != nil {
		return err
	}

	//if err := f.ServiceFinder.RegisterService(c.version); err != nil {
	//	return err
	//}

	h := FinderHandle{}
	configMap, err := f.ConfigFinder.UseAndSubscribeConfig([]string{configName}, h)
	if err != nil {
		return err
	}
	c.polarisInited = true

	ct, ok := configMap[configName]
	if !ok {
		return CalcFinderInitFailed.SetDetail(fmt.Sprintf("have no such config file : %s", configName))
	}

	_, err = toml.Decode(string(ct.File), &tomlconfig)
	if err != nil {
		return err
	}
	c.load()
	return nil

}
func (c *ConfigCalc) init(url, pro, gro, service, version string) error {

	c.url = url
	c.project = pro
	c.group = gro
	c.service = service
	c.version = version
	if c.isNative {
		return c.loadConfigFromNative()
	}

	if !c.polarisInited {
		return c.loadConfigFromPolaris()
	}
	return nil

}

func (c *ConfigCalc) load() {

	qSize = tomlconfig.Calc.QueueSize
	pNumber = tomlconfig.Calc.ConsumeNumber
	topic = tomlconfig.Calc.Topic
	enable = tomlconfig.Calc.Enable
	t := tomlconfig.Calc.Timeout
	timeout = time.Duration(t) * time.Millisecond

	level = tomlconfig.Log.Level
	logFormat = tomlconfig.Log.LogFormat
	logPath = tomlconfig.Log.LogPath
	consolePrint = tomlconfig.Log.ConsolePrint
}

type FinderHandle struct{}

func (FinderHandle) OnConfigFilesAdded(_ map[string]*common.Config) bool {
	return false
}

func (FinderHandle) OnConfigFilesRemoved(_ []string) bool {
	return false
}

func (FinderHandle) OnError(errorInfo common.ConfigErrInfo) {
	logger.Errorw("finderHandle onError message ", "error", errorInfo.ErrMsg)
}
func (FinderHandle) OnConfigFileChanged(config *common.Config) bool {

	logger.Infow("OnConfigFileChanged Enter")
	_, err := toml.Decode(string(config.File), &tomlconfig)
	if err != nil {
		logger.Errorw("toml decode failed when configFileChanged trigger", "error", err, "configData", string(config.File))
		return false
	}
	enableChange := tomlconfig.Calc.Enable
	if enable == enableChange {
		return true
	}
	if !enableChange {
		calcCore.Fini()
		isStop = true
		if err := logger.Flush(); err != nil {
			fmt.Printf("log flush failed when sdk fini , error = %s\n", err)
		}
		enable = enableChange

	} else {
		if err := calcCore.Init(qSize, pNumber, topic, tomlconfig.Calc.Hosts, timeout); err != nil {
			logger.Errorw("calcCore init failed when configFileChanged trigger", "error", err)
			return false
		}
		calcCore.Run()
		isStop = false
		enable = enableChange

	}
	return true
}
