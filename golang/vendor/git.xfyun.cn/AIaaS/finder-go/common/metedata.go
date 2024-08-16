package finder

import "time"

type ServiceMeteData struct {
	Project string
	Group   string
	Service string
	Version string
	Address string
}

type BootConfig struct {
	CompanionUrl     string
	CachePath        string
	CacheConfig      bool
	CacheService     bool
	TickerDuration   time.Duration
	ZkSessionTimeout time.Duration
	ZkConnectTimeout time.Duration
	ZkMaxSleepTime   time.Duration
	ZkMaxRetryNum    int
	MeteData         *ServiceMeteData
}

type ZkInfo struct {
	ZkAddr          []string
	ConfigRootPath  string
	ServiceRootPath string
}

type Config struct {
	Name string
	File []byte
}

type ServiceInstanceConfig struct {
	Weight  int  `json:"weight"`
	IsValid bool `json:"is_valid"`
}

type ConsumerInstanceConfig struct {
	IsValid bool `json:"is_valid"`
}

type ServiceInstanceChangedEvent struct {
	EventType  InstanceChangedEventType
	ServerList []*ServiceInstance
}

type ServiceInstance struct {
	Addr   string
	Config *ServiceInstanceConfig
}

type ServiceConfig struct {
	ProxyMode       string `json:"proxy_mode"`
	LoadBalanceMode string `json:"lb_mode"`
}

type Service struct {
	Name       string
	ServerList []*ServiceInstance
	Config     *ServiceConfig
}

type ConfigFeedback struct {
	PushID       string
	ServiceMete  *ServiceMeteData
	Config       string
	UpdateTime   int64
	UpdateStatus int
	LoadTime     int64
	LoadStatus   int
}

type ServiceFeedback struct {
	PushID          string
	ServiceMete     *ServiceMeteData
	Provider        string
	ProviderVersion string
	UpdateTime      int64
	UpdateStatus    int
	LoadTime        int64
	LoadStatus      int
}
