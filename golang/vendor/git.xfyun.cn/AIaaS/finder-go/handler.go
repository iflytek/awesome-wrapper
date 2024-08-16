package finder

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	common "git.xfyun.cn/AIaaS/finder-go/common"
	companion "git.xfyun.cn/AIaaS/finder-go/companion"
	"git.xfyun.cn/AIaaS/finder-go/utils/zkutil"
)

type ServiceHandle struct {
	ChangedHandler common.ServiceChangedHandler
	config         *common.BootConfig
	zkManager      *zkutil.ZkManager
}

func (s *ServiceHandle) OnServiceInstanceConfigChanged(name string, addr string, data []byte) {
	pushID, config, err := common.DecodeValue(data)
	if err != nil {
		// todo
		return
	}

	f := &common.ServiceFeedback{
		PushID:          pushID,
		ServiceMete:     s.config.MeteData,
		Provider:        name,
		ProviderVersion: "",
		UpdateTime:      time.Now().Unix(),
		UpdateStatus:    1,
	}
	c := &common.ServiceInstanceConfig{}
	err = json.Unmarshal(config, c)
	if err != nil {
		f.LoadStatus = -1
		log.Println(err)
	} else {
		ok := s.ChangedHandler.OnServiceInstanceConfigChanged(name, addr, c)
		if ok {
			log.Println("load success:", pushID)
			f.LoadStatus = 1
		}
	}

	f.LoadTime = time.Now().Unix()
	err = pushServiceFeedback(s.config.CompanionUrl, f)
	if err != nil {
		log.Println(err)
	}
}

func (s *ServiceHandle) OnServiceConfigChanged(name string, data []byte) {
	pushID, config, err := common.DecodeValue(data)
	if err != nil {
		// todo
		return
	}

	f := &common.ServiceFeedback{
		PushID:          pushID,
		ServiceMete:     s.config.MeteData,
		Provider:        name,
		ProviderVersion: "",
		UpdateTime:      time.Now().Unix(),
		UpdateStatus:    1,
	}
	c := &common.ServiceConfig{}
	err = json.Unmarshal(config, c)
	if err != nil {
		f.LoadStatus = -1
		log.Println(err)
	} else {
		ok := s.ChangedHandler.OnServiceConfigChanged(name, c)
		if ok {
			log.Println("load success:", pushID)
			f.LoadStatus = 1
		}
	}

	f.LoadTime = time.Now().Unix()
	err = pushServiceFeedback(s.config.CompanionUrl, f)
	if err != nil {
		log.Println(err)
	}
}

func (s *ServiceHandle) OnServiceInstanceChanged(name string, addrList []string) {
	eventList := make([]*common.ServiceInstanceChangedEvent, 0)
	newInstances := []*common.ServiceInstance{}
	cachedService, err := GetServiceFromCache(s.config.CachePath, name)
	if err != nil {
		log.Println("GetServiceFromCache", name, err)
		cachedService = &common.Service{Name: name, ServerList: newInstances}
	}
	if len(addrList) > 0 {
		servicePath := fmt.Sprintf("%s/%s/provider", s.zkManager.MetaData.ServiceRootPath, name)
		if len(cachedService.ServerList) > 0 {
			oldInstances, deletedEvent := getDeletedInstEvent(addrList, cachedService.ServerList)
			if deletedEvent != nil {
				eventList = append(eventList, deletedEvent)
			}
			if oldInstances != nil {
				newInstances = append(newInstances, oldInstances...)
			}
			addedEvent := getAddedInstEvents(s.zkManager, servicePath, addrList, cachedService.ServerList)
			if addedEvent != nil {
				newInstances = append(newInstances, addedEvent.ServerList...)
				eventList = append(eventList, addedEvent)
			}
		} else {
			addedEvent := getAddedInstEvents(s.zkManager, servicePath, addrList, cachedService.ServerList)
			if addedEvent != nil {
				newInstances = append(newInstances, addedEvent.ServerList...)
				eventList = append(eventList, addedEvent)
			}
		}
	} else {
		oldInstances, deletedEvent := getDeletedInstEvent(addrList, cachedService.ServerList)
		if deletedEvent != nil {
			eventList = append(eventList, deletedEvent)
		}
		if oldInstances != nil {
			newInstances = append(newInstances, oldInstances...)
		}
	}

	cachedService.ServerList = newInstances
	err = CacheService(s.config.CachePath, cachedService)
	if err != nil {
		log.Println("CacheService failed")
	}

	ok := s.ChangedHandler.OnServiceInstanceChanged(name, eventList)
	if !ok {
		log.Println("OnServiceInstanceChanged is not ok")
	}
}

func getDeletedInstEvent(addrList []string, insts []*common.ServiceInstance) ([]*common.ServiceInstance, *common.ServiceInstanceChangedEvent) {
	var event *common.ServiceInstanceChangedEvent
	var oldInstances []*common.ServiceInstance
	var deletedInstances []*common.ServiceInstance
	var deleted bool
	for _, inst := range insts {
		deleted = true
		for _, addr := range addrList {
			if addr == inst.Addr {
				deleted = false
				if oldInstances == nil {
					oldInstances = []*common.ServiceInstance{}
				}
				oldInstances = append(oldInstances, inst)
			}
		}
		if deleted {
			if deletedInstances == nil {
				deletedInstances = []*common.ServiceInstance{}
			}
			deletedInstances = append(deletedInstances, inst)
		}
	}

	if deletedInstances != nil {
		event = &common.ServiceInstanceChangedEvent{
			EventType:  common.INSTANCEREMOVE,
			ServerList: deletedInstances,
		}
	}

	return oldInstances, event
}

func getAddedInstEvents(zm *zkutil.ZkManager, servicePath string, addrList []string, insts []*common.ServiceInstance) *common.ServiceInstanceChangedEvent {
	var event *common.ServiceInstanceChangedEvent
	var addedInstances []*common.ServiceInstance
	var added bool
	for _, addr := range addrList {
		added = true
		for _, inst := range insts {
			if addr == inst.Addr {
				added = false
			}
		}
		if added {
			inst, err := getServiceInstance(zm, servicePath, addr)
			if err != nil {
				log.Println(err)
				// todo
				continue
			}

			if addedInstances == nil {
				addedInstances = []*common.ServiceInstance{}
			}
			addedInstances = append(addedInstances, inst)
		}
	}

	if addedInstances != nil {
		event = &common.ServiceInstanceChangedEvent{
			EventType:  common.INSTANCEADDED,
			ServerList: addedInstances,
		}
	}

	return event
}

type ConfigHandle struct {
	config         *common.BootConfig
	ChangedHandler common.ConfigChangedHandler
}

func (s *ConfigHandle) OnConfigFileChanged(name string, data []byte) {
	pushID, file, err := common.DecodeValue(data)
	if err != nil {
		// todo
	} else {
		f := &common.ConfigFeedback{
			PushID:       pushID,
			ServiceMete:  s.config.MeteData,
			Config:       name,
			UpdateTime:   time.Now().Unix(),
			UpdateStatus: 1,
		}
		c := &common.Config{
			Name: name,
			File: file,
		}

		ok := s.ChangedHandler.OnConfigFileChanged(c)
		if ok {
			err = CacheConfig(s.config.CachePath, c)
			if err != nil {
				log.Println(err)
				// todo
			}

			log.Println("load success:", pushID)
			f.LoadStatus = 1
		}
		f.LoadTime = time.Now().Unix()
		err = pushConfigFeedback(s.config.CompanionUrl, f)
		if err != nil {
			log.Println(err)
		}
	}
}

func pushConfigFeedback(companionUrl string, f *common.ConfigFeedback) error {
	url := companionUrl + "/finder/push_config_feedback"
	return companion.FeedbackForConfig(hc, url, f)
}

func pushServiceFeedback(companionUrl string, f *common.ServiceFeedback) error {
	url := companionUrl + "/finder/push_service_feedback"
	return companion.FeedbackForService(hc, url, f)
}

func pushService(companionUrl string, project string, group string, service string) error {
	url := companionUrl + "/finder/register_service_info"
	return companion.RegisterService(hc, url, project, group, service)
}
