package xsf

import (
	"fmt"
	finder "git.iflytek.com/AIaaS/finder-go/common"
	"git.iflytek.com/AIaaS/xsf/utils"
	"sync"
	"time"
)

type detector struct {
	finderInst *utils.FindManger
	addrCache  *sync.Map
}

func (d *detector) OnServiceInstanceConfigChanged(name string, version string, instance string, config *finder.ServiceInstanceConfig) bool {
	fmt.Println("OnServiceInstanceConfigChanged", name, version, instance, config)
	if config.IsValid {
		loggerStd.Println("OnServiceInstanceConfigChanged insert adddr: ", instance)
		d.addrCache.Store(instance, true)
	} else {
		loggerStd.Println("OnServiceInstanceConfigChanged delete adddr: ", instance)
		d.addrCache.Delete(instance)
	}
	return true
}
func (d *detector) OnServiceConfigChanged(name, version string, config *finder.ServiceConfig) bool {
	fmt.Println("OnServiceConfigChanged", name, version, config)
	return true
}
func (d *detector) OnServiceInstanceChanged(name, version string, instances []*finder.ServiceInstanceChangedEvent) bool {
	for _, v := range instances {
		if v.EventType == finder.INSTANCEADDED {
			for _, inst := range v.ServerList {
				loggerStd.Println("OnServiceInstanceChanged insert addrs:", inst.Addr)
				d.addrCache.Store(inst.Addr, true)
			}
		} else if v.EventType == finder.INSTANCEREMOVE {
			for _, inst := range v.ServerList {
				loggerStd.Println("OnServiceInstanceChanged delete addrs:", inst.Addr)
				d.addrCache.Delete(inst.Addr)
			}
		}
	}
	return true
}

func newDetector(
	cfgUrl string,
	cfgPrj string,
	cfgGroup string,
	cfgName string,
	cfgApiVersion string,
	log *utils.Logger,
) (*detector, error) {
	co := &utils.CfgOption{}
	utils.WithCfgTick(time.Second)(co)
	utils.WithCfgSessionTimeOut(time.Second)(co)
	utils.WithCfgURL(cfgUrl)(co)
	utils.WithCfgCachePath(".")(co)
	utils.WithCfgCacheConfig(true)(co)
	utils.WithCfgCacheService(true)(co)
	utils.WithCfgPrj(cfgPrj)(co)
	utils.WithCfgGroup(cfgGroup)(co)
	utils.WithCfgLog(log)(co)
	finderInst, err := utils.NewFinder(co)
	if err != nil {
		loggerStd.Println("newFinder failed.")
		return nil, err
	}

	detectorInst := &detector{finderInst: finderInst}

	addrCache, addrCacheErr := func() (*sync.Map, error) {
		m := &sync.Map{}
		srvs, srvsErr := finderInst.UseSrvAndSub(cfgApiVersion, cfgName, detectorInst)
		if srvsErr != nil {
			loggerStd.Printf("detect %v:%v failed\n", cfgName, cfgApiVersion)
			return m, srvsErr
		}
		for _, v := range srvs {
			if len(v.ProviderList) == 0 {
				panic(fmt.Sprintf("receive addr for %v->%v->%v->%v failed\n", cfgPrj, cfgGroup, cfgName, cfgApiVersion))
			}
			var forLog []string
			for _, provider := range v.ProviderList {
				m.Store(provider.Addr, true)
				forLog = append(forLog, provider.Addr)
			}
			loggerStd.Println(fmt.Sprintf("received addr:%v for %v->%v->%v->%v successfully", forLog, cfgPrj, cfgGroup, cfgName, cfgApiVersion))
		}
		return m, nil
	}()
	if addrCacheErr != nil {
		return nil, addrCacheErr
	}
	detectorInst.addrCache = addrCache
	return detectorInst, nil
}

func (d *detector) getAll() []string {
	var addrSet []string
	d.addrCache.Range(func(key, value interface{}) bool {
		if value.(bool) {
			addrSet = append(addrSet, key.(string))
		}
		return true
	})
	return addrSet
}
