package utils

import (
	"fmt"
	finder "git.iflytek.com/AIaaS/finder-go/common"
	"sync"
	"testing"
	"time"
)

type srvCallback struct {
}

func (sd *srvCallback) OnServiceInstanceConfigChanged(name string, version string, instance string, config *finder.ServiceInstanceConfig) bool {
	fmt.Println("OnServiceInstanceConfigChanged", name, version, instance, config)
	return true
}
func (sd *srvCallback) OnServiceConfigChanged(name, version string, config *finder.ServiceConfig) bool {
	fmt.Println("OnServiceConfigChanged", name, version, config)

	return true
}
func (sd *srvCallback) OnServiceInstanceChanged(name, version string, instances []*finder.ServiceInstanceChangedEvent) bool {
	fmt.Println("OnServiceInstanceChanged", name, version, instances)
	return true
}

type cfg struct {
}

func (c *cfg) OnConfigFilesAdded(configs map[string]*finder.Config) bool {
	return true
}

func (c *cfg) OnConfigFilesRemoved(configNames []string) bool {
	return true
}

func (c *cfg) OnConfigFileChanged(con *finder.Config) bool {
	fmt.Println("OnConfigFileChanged")
	return true
}
func (c *cfg) OnError(errInfo finder.ConfigErrInfo) {
	fmt.Println("OnError")
}
func TestFinder(t *testing.T) {
	checkErr := func(err error) {
		if err != nil {
			panic(err)
		}
	}

	co := &CfgOption{}
	WithCfgTick(time.Second)(co)
	WithCfgSessionTimeOut(time.Second)(co)
	WithCfgURL("http://10.1.87.69:6868")(co)
	WithCfgCachePath(".")(co)
	WithCfgCacheConfig(true)(co)
	WithCfgCacheService(true)(co)
	WithCfgPrj("guiderAllService")(co)
	WithCfgGroup("gas")(co)
	WithCfgLog(func() *Logger {
		logger, err := NewLocalLog(
			SetCaller(true),
			SetLevel("debug"),
			SetFileName("test.log"),
			SetMaxSize(3),
			SetMaxBackups(3),
			SetMaxAge(3),
			SetAsync(false),
			SetCacheMaxCount(30000),
			SetBatchSize(1024))
		checkErr(err)
		return logger
	}())(co)

	findInst, findInstErr := NewFinder(co)
	checkErr(findInstErr)

	fmt.Println("get srvs...")

	srvs, srvsErr := findInst.UseSrvAndSub("1.0.0", "atmos-iat", &srvCallback{})
	checkErr(srvsErr)

	for k, v := range srvs {
		for _, provider := range v.ProviderList {
			fmt.Println("x:", k, provider.Addr)
		}
	}
}
func TestRegister(t *testing.T) {
	checkErr := func(err error) {
		if err != nil {
			panic(err)
		}
	}

	co := &CfgOption{
		tick:         time.Second,
		stmout:       time.Second,
		url:          "http://10.1.87.69:6868",
		cachePath:    ".",
		cacheConfig:  true,
		cacheService: true,
		prj:          "guiderAllService",
		group:        "gas",
		srv:          "lbv2",
		ver:          "2.2.7",
		log: func() *Logger {
			logger, err := NewLocalLog(
				SetCaller(true),
				SetLevel("debug"),
				SetFileName("test.log"),
				SetMaxSize(3),
				SetMaxBackups(3),
				SetMaxAge(3),
				SetAsync(false),
				SetCacheMaxCount(30000),
				SetBatchSize(1024))
			checkErr(err)
			return logger
		}()}
	findInst, findInstErr := NewFinder(co)
	checkErr(findInstErr)

	//fmt.Println("get cfg...")
	//cfgInst, cfgInstErr := findInst.UseCfgAndSub("lbv2.toml", &cfg{})
	//checkErr(cfgInstErr)
	//for k, v := range cfgInst {
	//	fmt.Println(k, v.ConfigMap["lbv2"])
	//	fmt.Println(k, v.Name)
	//}

	fmt.Println("get srvs...")

	srvs, srvsErr := findInst.UseSrvAndSub("1.0.0", "atmos-iat", &srvCallback{})
	checkErr(srvsErr)

	for k, v := range srvs {
		fmt.Println(k, v)
		for _, provider := range v.ProviderList {
			fmt.Println(k, provider.Addr)
		}
	}
	fmt.Println(findInst.RegisterSrvWithAddr("1.1.1.1", "9.9.9"))
	time.Sleep(time.Hour)
}

func TestNewFinder(t *testing.T) {
	checkErr := func(err error) {
		if err != nil {
			panic(err)
		}
	}

	co := &CfgOption{
		tick:         time.Second,
		stmout:       time.Second,
		url:          "http://10.1.87.69:6868",
		cachePath:    ".",
		cacheConfig:  true,
		cacheService: true,
		prj:          "guiderAllService",
		group:        "gas",
		srv:          "lbv2",
		ver:          "2.2.7",
		log: func() *Logger {
			logger, err := NewLocalLog(
				SetCaller(true),
				SetLevel("debug"),
				SetFileName("test.log"),
				SetMaxSize(3),
				SetMaxBackups(3),
				SetMaxAge(3),
				SetAsync(false),
				SetCacheMaxCount(30000),
				SetBatchSize(1024))
			checkErr(err)
			return logger
		}()}
	findInst, findInstErr := NewFinder(co)
	checkErr(findInstErr)

	fmt.Println("get cfg...")
	cfgInst, cfgInstErr := findInst.UseCfgAndSub("lbv2.toml", &cfg{})
	checkErr(cfgInstErr)
	for k, v := range cfgInst {
		fmt.Println(k, v.ConfigMap["lbv2"])
		fmt.Println(k, v.Name)
	}

	fmt.Println("get srvs...")

	srvs, srvsErr := findInst.UseSrvAndSub("1.0.0", "atmos-iat", &srvCallback{})
	checkErr(srvsErr)

	for k, v := range srvs {
		fmt.Println(k, v)
		for _, provider := range v.ProviderList {
			fmt.Println(k, provider.Addr)
		}
	}
}
func TestFinderCallback(t *testing.T) {
	checkErr := func(err error) {
		if err != nil {
			panic(err)
		}
	}

	co := &CfgOption{
		tick:         time.Second,
		stmout:       time.Second,
		url:          "http://10.1.87.69:6868",
		cachePath:    ".",
		cacheConfig:  true,
		cacheService: true,
		prj:          "guiderAllService",
		group:        "gas",
		srv:          "lbv2",
		ver:          "2.2.7",
		log: func() *Logger {
			logger, err := NewLocalLog(
				SetCaller(true),
				SetLevel("debug"),
				SetFileName("test.log"),
				SetMaxSize(3),
				SetMaxBackups(3),
				SetMaxAge(3),
				SetAsync(false),
				SetCacheMaxCount(30000),
				SetBatchSize(1024))
			checkErr(err)
			return logger
		}()}
	findInst, findInstErr := NewFinder(co)
	checkErr(findInstErr)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("get cfg...")
		cfgInst, cfgInstErr := findInst.UseCfgAndSub("lbv2.toml", &cfg{})
		checkErr(cfgInstErr)

		for k, v := range cfgInst {
			fmt.Println(k, v.ConfigMap["lbv2"])
			fmt.Println(k, v.Name)
		}

		time.Sleep(time.Minute)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("get srvs...")
		srvs, srvsErr := findInst.UseSrvAndSub("1.0.0", "atmos-iat", &srvCallback{})
		checkErr(srvsErr)
		for k, v := range srvs {
			fmt.Println(k, v)
			for _, provider := range v.ProviderList {
				fmt.Println(k, provider.Addr)
			}
		}

		time.Sleep(time.Minute)
	}()

	wg.Wait()
}
func TestFinderQueryService(t *testing.T) {
	checkErr := func(err error) {
		if err != nil {
			panic(err)
		}
	}

	co := &CfgOption{
		tick:         time.Second,
		stmout:       time.Second,
		url:          "http://10.1.87.69:6868",
		cachePath:    ".",
		cacheConfig:  true,
		cacheService: true,
		prj:          "metrics",
		group:        "3s",
		srv:          "beacon",
		ver:          "0.0.0",
		log: func() *Logger {
			logger, err := NewLocalLog(
				SetCaller(true),
				SetLevel("debug"),
				SetFileName("test.log"),
				SetMaxSize(3),
				SetMaxBackups(3),
				SetMaxAge(3),
				SetAsync(false),
				SetCacheMaxCount(30000),
				SetBatchSize(1024))
			checkErr(err)
			return logger
		}()}
	findInst, findInstErr := NewFinder(co)
	checkErr(findInstErr)
	cb := &srvCallback{}
	fmt.Println(findInst.QueryServiceWatch("metrics", "3s", cb))
	go func() {

	}()
	time.Sleep(time.Minute)
}
