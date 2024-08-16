package finder

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	common "git.xfyun.cn/AIaaS/finder-go/common"
	errors "git.xfyun.cn/AIaaS/finder-go/errors"
	"git.xfyun.cn/AIaaS/finder-go/utils/stringutil"
	"git.xfyun.cn/AIaaS/finder-go/utils/zkutil"
	"github.com/cooleric/curator"
	"github.com/cooleric/go-zookeeper/zk"
)

type ServiceFinder struct {
	zkManager         *zkutil.ZkManager
	config            *common.BootConfig
	logger            common.Logger
	SubscribedService map[string]*common.Service
	mutex             sync.Mutex
}

func (f *ServiceFinder) RegisterService() error {
	return registerService(f, f.config.MeteData.Address)
}

func (f *ServiceFinder) RegisterServiceWithAddr(addr string) error {
	return registerService(f, addr)
}

func (f *ServiceFinder) UnRegisterService() error {
	servicePath := fmt.Sprintf("%s/%s/provider/%s", f.zkManager.MetaData.ServiceRootPath, f.config.MeteData.Service, f.config.MeteData.Address)

	return f.zkManager.RemoveInRecursive(servicePath)
}

func (f *ServiceFinder) UnRegisterServiceWithAddr(addr string) error {
	servicePath := fmt.Sprintf("%s/%s/provider/%s", f.zkManager.MetaData.ServiceRootPath, f.config.MeteData.Service, addr)

	return f.zkManager.RemoveInRecursive(servicePath)
}

func (f *ServiceFinder) UseService(name []string) (map[string]*common.Service, error) {
	var err error
	if len(name) == 0 {
		err = &errors.FinderError{
			Ret:  errors.ServiceMissName,
			Func: "UseService",
		}

		return nil, err
	}

	var addrList []string
	serviceList := make(map[string]*common.Service)
	for _, n := range name {
		servicePath := fmt.Sprintf("%s/%s/provider", f.zkManager.MetaData.ServiceRootPath, n)
		f.logger.Info("useservice:", servicePath)
		addrList, err = f.zkManager.GetChildren(servicePath)
		if err != nil {
			f.logger.Info("useservice:", err)
			service, err := GetServiceFromCache(f.config.CachePath, n)
			if err != nil {
				f.logger.Error(err)
				//todo notify
			} else {
				serviceList[n] = service
			}
		} else if len(addrList) > 0 {
			f.logger.Info("sp:", servicePath)
			f.logger.Info(addrList)
			serviceList[n] = getService(f.zkManager, servicePath, n, addrList)
			err = CacheService(f.config.CachePath, serviceList[n])
			if err != nil {
				f.logger.Error("CacheService failed")
			}
		}

		err = registerConsumer(f, n, f.config.MeteData.Address)
		if err != nil {
			f.logger.Error("registerConsumer failed,", err)
		}
	}

	return serviceList, err
}

func (f *ServiceFinder) UseAndSubscribeService(name []string, handler common.ServiceChangedHandler) (map[string]*common.Service, error) {
	var err error
	if len(name) == 0 {
		err = &errors.FinderError{
			Ret:  errors.ServiceMissName,
			Func: "UseAndSubscribeService",
		}

		return nil, err
	}

	serviceList := make(map[string]*common.Service)
	serviceChan := make(chan *common.Service)
	f.mutex.Lock()
	defer f.mutex.Unlock()
	go func(f *ServiceFinder, serviceList map[string]*common.Service, serviceChan chan *common.Service) {
		interHandle := &ServiceHandle{ChangedHandler: handler, config: f.config, zkManager: f.zkManager}
		for _, n := range name {
			if s, ok := f.SubscribedService[n]; ok {
				serviceList[n] = s
				serviceChan <- &common.Service{}

				continue
			}

			servicePath := fmt.Sprintf("%s/%s/provider", f.zkManager.MetaData.ServiceRootPath, n)
			err = f.zkManager.GetChildrenW(servicePath, func(c curator.CuratorFramework, e curator.CuratorEvent) error {
				addrList := e.Children()
				if len(addrList) > 0 {
					service := getServiceWithWatcher(f.zkManager, servicePath, n, addrList, interHandle)
					if len(service.Name) > 0 {
						err = CacheService(f.config.CachePath, service)
						if err != nil {
							f.logger.Info("CacheService failed")
						}
						serviceChan <- service
					} else {
						service, err := GetServiceFromCache(f.config.CachePath, n)
						if err != nil {
							f.logger.Info(err)
							//todo notify
							serviceChan <- &common.Service{}
						} else {
							serviceChan <- service
						}
					}

					return nil
				}
				serviceChan <- &common.Service{}
				return nil
			})
			// handleChan := ServiceHandle{ChangedHandler: handler}
			if err != nil {
				service, err := GetServiceFromCache(f.config.CachePath, n)
				if err != nil {
					f.logger.Info("GetServiceFromCache ", err)
					//todo notify
					serviceChan <- &common.Service{}
				} else {
					serviceChan <- service
				}

				continue
			}
			err = registerConsumer(f, n, f.config.MeteData.Address)
			if err != nil {
				f.logger.Error("registerConsumer failed,", err)
			}

			zkutil.ServiceEventPool.Append(common.ServiceEventPrefix+n, interHandle)
		}
	}(f, serviceList, serviceChan)

	return f.waitServiceResult(serviceList, serviceChan, len(name)), nil
}

func (f *ServiceFinder) UnSubscribeService(name string) error {
	var err error
	if len(name) == 0 {
		err = &errors.FinderError{
			Ret:  errors.ServiceMissName,
			Func: "UnSubscribeService",
		}
		return err
	}

	zkutil.ServiceEventPool.Remove(name)
	f.mutex.Lock()
	delete(f.SubscribedService, name)
	f.mutex.Unlock()

	return nil
}

func registerService(f *ServiceFinder, addr string) error {
	if stringutil.IsNullOrEmpty(addr) {
		err := &errors.FinderError{
			Ret:  errors.ServiceMissAddr,
			Func: "RegisterService",
		}

		f.logger.Error("RegisterService:", err)
		return err
	}

	data, err := getDefaultServiceItemConfig(addr)
	if err != nil {
		f.logger.Error("RegisterService->getDefaultServiceItemConfig:", err)
		return err
	}
	parentPath := fmt.Sprintf("%s/%s/provider", f.zkManager.MetaData.ServiceRootPath, f.config.MeteData.Service)
	err = register(f.zkManager, parentPath, addr, data)
	if err != nil {
		f.logger.Error("RegisterService->register:", err)
		return err
	}

	err = pushService(f.config.CompanionUrl, f.config.MeteData.Project, f.config.MeteData.Group, f.config.MeteData.Service)
	if err != nil {
		f.logger.Error("RegisterService->registerService:", err)
	}

	return nil
}

func registerConsumer(f *ServiceFinder, service string, addr string) error {
	if stringutil.IsNullOrEmpty(addr) {
		err := &errors.FinderError{
			Ret:  errors.ServiceMissAddr,
			Func: "registerConsumer",
		}

		f.logger.Error("registerConsumer:", err)
		return err
	}

	data, err := getDefaultConsumerItemConfig(addr)
	if err != nil {
		f.logger.Error("registerConsumer->getDefaultConsumerItemConfig:", err)
		return err
	}
	parentPath := fmt.Sprintf("%s/%s/consumer", f.zkManager.MetaData.ServiceRootPath, service)
	err = register(f.zkManager, parentPath, addr, data)
	if err != nil {
		f.logger.Error("registerConsumer->register:", err)
		return err
	}

	return nil
}

func register(zm *zkutil.ZkManager, parentPath string, addr string, data []byte) error {
	log.Println("call register func")
	var node *zk.Stat
	var err error
	servicePath := parentPath + "/" + addr
	log.Println("servicePath:", servicePath)
	node, err = zm.ExistsNode(servicePath)
	if err != nil {
		log.Println("ExistsNode:", err)
		return err
	}
	if node == nil {
		log.Println("begin createParentNode")
		err = createParentNode(zm, parentPath)
		if err != nil {
			log.Println("createParentNode", err)
			return err
		}

		log.Println("begin createTempNode")
		return createTempNode(zm, servicePath, data)
	}

	log.Println("exist node")
	err = zm.RemoveInRecursive(servicePath)
	if err != nil {
		log.Println("RemoveInRecursive:", err)
		return err
	}
	log.Println("begin createTempNode")
	return createTempNode(zm, servicePath, data)
}

func createParentNode(zm *zkutil.ZkManager, parentPath string) error {
	node, err := zm.ExistsNode(parentPath)
	if err != nil {
		log.Println(err)
		return err
	}

	if node == nil {
		var result string
		result, err = zm.CreatePath(parentPath)
		if err != nil {
			log.Println(err)
			return err
		}
		log.Println(result)
	}

	return nil
}

func createTempNode(zm *zkutil.ZkManager, path string, data []byte) error {
	result, err := zm.CreateTempPathWithData(path, data)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println(result)

	return nil
}

func getDefaultServiceItemConfig(addr string) ([]byte, error) {
	defaultServiceInstanceConfig := common.ServiceInstanceConfig{
		Weight:  100,
		IsValid: true,
	}

	data, err := json.Marshal(defaultServiceInstanceConfig)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var encodedData []byte
	encodedData, err = common.EncodeValue("", data)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return encodedData, nil
}

func getDefaultConsumerItemConfig(addr string) ([]byte, error) {
	defaultConsumeInstanceConfig := common.ConsumerInstanceConfig{
		IsValid: true,
	}

	data, err := json.Marshal(defaultConsumeInstanceConfig)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var encodedData []byte
	encodedData, err = common.EncodeValue("", data)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return encodedData, nil
}

func getServiceInstance(zm *zkutil.ZkManager, path string, addr string) (*common.ServiceInstance, error) {
	data, err := zm.GetNodeData(path + "/" + addr)
	if err != nil {
		return nil, err
	}

	var item []byte
	_, item, err = common.DecodeValue(data)
	if err != nil {
		return nil, err
	}

	log.Println(string(item))
	serviceInstanceConfig := &common.ServiceInstanceConfig{}
	err = json.Unmarshal(item, serviceInstanceConfig)
	if err != nil {
		return nil, err
	}

	serviceInstance := new(common.ServiceInstance)
	serviceInstance.Addr = addr
	serviceInstance.Config = serviceInstanceConfig

	return serviceInstance, nil
}

func getService(zm *zkutil.ZkManager, servicePath string, name string, addrList []string) *common.Service {
	var service = &common.Service{Name: name, ServerList: make([]*common.ServiceInstance, 0), Config: &common.ServiceConfig{}}
	for _, addr := range addrList {
		serviceInstance, err := getServiceInstance(zm, servicePath, addr)
		if err != nil {
			log.Println(err)
			// todo
			continue
		}

		service.ServerList = append(service.ServerList, serviceInstance)
	}
	// todo
	service.Config.ProxyMode = "default"
	service.Config.LoadBalanceMode = "default"

	return service
}

func getServiceWithWatcher(zm *zkutil.ZkManager, servicePath string, name string, addrList []string, interHandle *ServiceHandle) *common.Service {
	var service = &common.Service{Name: name, ServerList: make([]*common.ServiceInstance, 0), Config: &common.ServiceConfig{}}
	for _, addr := range addrList {
		serviceInstance, err := getServiceInstanceWithWatcher(zm, servicePath, addr, interHandle)
		if err != nil {
			log.Println(err)
			// todo
			continue
		}

		service.ServerList = append(service.ServerList, serviceInstance)
	}
	// todo
	service.Config.ProxyMode = "default"
	service.Config.LoadBalanceMode = "default"

	return service
}

func getServiceInstanceWithWatcher(zm *zkutil.ZkManager, servicePath string, addr string, interHandle *ServiceHandle) (*common.ServiceInstance, error) {
	serviceInstanceChan := make(chan *common.ServiceInstance)
	err := zm.GetNodeDataW(servicePath+"/"+addr, func(c curator.CuratorFramework, e curator.CuratorEvent) error {
		_, item, err := common.DecodeValue(e.Data())
		if err != nil {
			serviceInstanceChan <- &common.ServiceInstance{}
			return err
		}
		serviceInstance := &common.ServiceInstance{Addr: addr, Config: new(common.ServiceInstanceConfig)}
		err = json.Unmarshal(item, serviceInstance.Config)
		if err != nil {
			serviceInstanceChan <- &common.ServiceInstance{}
			return err
		}

		serviceInstanceChan <- serviceInstance
		return nil
	})
	if err != nil {
		return nil, err
	}
	zkutil.ServiceEventPool.Append(common.ServiceProviderEventPrefix+addr, interHandle)

	return waitServiceInstanceResult(serviceInstanceChan), nil
}

func (f *ServiceFinder) waitServiceResult(serviceList map[string]*common.Service, serviceChan chan *common.Service, serviceNum int) map[string]*common.Service {
	index := 0
	for {
		select {
		case s := <-serviceChan:
			index++
			if s != nil && len(s.Name) > 0 {
				serviceList[s.Name] = s
				f.SubscribedService[s.Name] = s
			}
			if index == serviceNum {
				close(serviceChan)

				return serviceList
			}
		}

	}
}

func waitServiceInstanceResult(serviceInstanceChan chan *common.ServiceInstance) *common.ServiceInstance {
	serviceInstance := new(common.ServiceInstance)
	for {
		select {
		case s := <-serviceInstanceChan:
			if len(s.Addr) > 0 {
				serviceInstance = s
			}
			close(serviceInstanceChan)
			return serviceInstance
		}
	}
}
