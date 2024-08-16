# finder-go 使用说明
1. 创建config

```
config := common.BootConfig{
		//companion地址
		CompanionUrl: conf.CompanionUrl,
		//缓存路径
		CachePath: "",
		//是否缓存服务信息
		CacheService: true,
		//是否缓存配置信息
		CacheConfig:   true,
		ExpireTimeout: 5 * time.Second,
		MeteData: &common.ServiceMeteData{
			Project: "test",
			Group:   "test",
			Service: "test",
			Version: "1.0.1",
			Address: "127.0.0.1:1221",
		},
	}
```
2. 创建finder

```
f, err := finder.NewFinderWithLogger(config, nil)

```
3. 使用接口

```
//实现handler接口 订阅服务 返回的map的key是ServiceName + "_" + ApiVersion。 vlaue是Service实例，主要取ProviderList。代表当前服务的提供者列表
serviceList, err := f.ServiceFinder.UseAndSubscribeServic(subscri, handler)
//注册服务
f.ServiceFinder.RegisterServiceWithAddr(addr, apiVersion)

//获取配置文件
configFiles, err := f.ConfigFinder.UseAndSubscribeConfig(name, &handler)

```

* handler接口说明

```
type ConfigChangedHandler interface {
    //配置文件发生改变后的回调
	OnConfigFileChanged(config *Config) bool
	OnError(errInfo ConfigErrInfo)
}

type ServiceChangedHandler interface {

	//服务实例上的配置信息发生变化 回调接口
	OnServiceInstanceConfigChanged(name string,apiVersion string, addr string, config *ServiceInstanceConfig) bool
	//服务整体配置信息发生变化 回调接口
	OnServiceConfigChanged(name string,apiVersion string,  config *ServiceConfig) bool
	//服务实例发生变化回调接口，ServiceInstanceChangedEvent中的EventType代表增加提供者还是减少，ServerList代表增加的实例或者减少的实例
	OnServiceInstanceChanged(name string, apiVersion string, eventList []*ServiceInstanceChangedEvent) bool
}

```

###  查询所有服务

```

f.ServiceFinder.QueryService("AIaaS", "dx")

//返回值分析
map[string][]common.ServiceInfo 
key : serviceName
value : 列表 common.ServiceInfo 

type ServiceInfo struct {
	ApiVersion   string   //版本号
	ProviderList []string //所有地址，可能为空
}
```

### 流程
* 参加[流程图](https://git.iflytek.com/AIaaS/finder-go/blob/master/%E9%85%8D%E7%BD%AE%E4%B8%AD%E5%BF%83%E6%B5%81%E7%A8%8B.png)

### 2.1.19 更新：

1. 集成时需要新增configChangeHandler 的接口实现函数：
````go
func (s *ConfigChangedHandle) OnConfigFilesAdded(configs map[string]*common.Config) bool {


	return true
}

func (s *ConfigChangedHandle) OnConfigFilesRemoved(configNames []string) bool {


	return true
}

````

2. git.iflytek.com/AIaaS/finder-go/common 包名由原来错误的finder 修正为common。


###  c语言支持，执行该脚本：
1. [create.sh](./cgo/create.sh) 生成libfinder.so  和libfinder.h 
2. 使用时包含两个头文件 libfinder.h 和 [config_center.h](./cgo/config_center.h)
3. 使用demo见[test_config.c](./cgo/example/test_config.c) 和 [test_service.c](./cgo/example/test_service.c) 


