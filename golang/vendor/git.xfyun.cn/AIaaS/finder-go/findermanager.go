package finder

import (
	"net"
	"net/http"
	"os"
	"time"

	common "git.xfyun.cn/AIaaS/finder-go/common"
	errors "git.xfyun.cn/AIaaS/finder-go/errors"
	"git.xfyun.cn/AIaaS/finder-go/utils/fileutil"
	"git.xfyun.cn/AIaaS/finder-go/utils/stringutil"
	"git.xfyun.cn/AIaaS/finder-go/utils/zkutil"
	"git.xfyun.cn/AIaaS/finder-go/utils/netutil"
)

var (
	hc *http.Client
)

func init() {
	hc = &http.Client{
		Transport: &http.Transport{
			Dial: func(nw, addr string) (net.Conn, error) {
				deadline := time.Now().Add(1 * time.Second)
				c, err := net.DialTimeout(nw, addr, time.Second*1)
				if err != nil {
					return nil, err
				}
				c.SetDeadline(deadline)
				return c, nil
			},
		},
	}
}

// FinderManager for controll all
type FinderManager struct {
	config         *common.BootConfig
	ConfigFinder   *ConfigFinder
	ServiceFinder  *ServiceFinder
	zkManager      *zkutil.ZkManager
	InternalLogger common.Logger
}

func checkCachePath(path string) (string, error) {
	if stringutil.IsNullOrEmpty(path) {
		p, err := os.Getwd()
		if err == nil {
			p += (fileutil.GetSystemSeparator() + common.DefaultCacheDir)
			path = p
		} else {
			return path, err
		}
	}

	return path, nil
}

func createCacheDir(path string) error {
	exist, err := fileutil.ExistPath(path)
	if err == nil && !exist {
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return err
		}
	} else {
		return err
	}

	return nil
}

// NewFinder for creating an instance
func NewFinder(config common.BootConfig) (*FinderManager, error) {
	logger := common.NewDefaultLogger()

	if stringutil.IsNullOrEmpty(config.CompanionUrl){
		err := &errors.FinderError{
			Ret:  errors.MissCompanionUrl,
			Func: "NewFinder",
		}
		return nil,err
	}

	if stringutil.IsNullOrEmpty(config.MeteData.Address){
		localIP,err:=netutil.GetLocalIP(config.CompanionUrl)
		if err!=nil{
			logger.Error(err)
			return nil,err
		}
		config.MeteData.Address=localIP
	}

	// 检查缓存路径，如果传入cachePath是空，则使用默认路径
	p, err := checkCachePath(config.CachePath)
	if err != nil {
		return nil, err
	}

	// 创建缓存目录
	err = createCacheDir(p)
	if err != nil {
		return nil, err
	}
	config.CachePath = p
	// 初始化finder
	fm := new(FinderManager)
	fm.InternalLogger = logger
	fm.config = &config
	// 初始化zk
	fm.zkManager, err = zkutil.NewZkManager(fm.config)
	if err != nil {
		return nil, err
	}

	fm.ConfigFinder = &ConfigFinder{zkManager: fm.zkManager, config: fm.config, logger: fm.InternalLogger}
	fm.ServiceFinder = &ServiceFinder{zkManager: fm.zkManager, config: fm.config, logger: fm.InternalLogger, SubscribedService: make(map[string]*common.Service)}

	if err != nil {
		return nil, err
	}

	return fm, nil
}

func NewFinderWithLogger(config common.BootConfig, logger common.Logger) (*FinderManager, error) {
	if logger == nil {
		logger = common.NewDefaultLogger()
	}

	if stringutil.IsNullOrEmpty(config.CompanionUrl){
		err := &errors.FinderError{
			Ret:  errors.MissCompanionUrl,
			Func: "NewFinder",
		}
		return nil,err
	}

	if stringutil.IsNullOrEmpty(config.MeteData.Address){
		localIP,err:=netutil.GetLocalIP(config.CompanionUrl)
		if err!=nil{
			logger.Error(err)
			return nil,err
		}
		config.MeteData.Address=localIP
	}
	
	// 检查缓存路径，如果传入cachePath是空，则使用默认路径
	p, err := checkCachePath(config.CachePath)
	if err != nil {
		return nil, err
	}

	// 创建缓存目录
	err = createCacheDir(p)
	if err != nil {
		return nil, err
	}
	config.CachePath = p
	// 初始化finder
	fm := new(FinderManager)
	fm.InternalLogger = logger
	fm.config = &config
	// 初始化zk
	fm.zkManager, err = zkutil.NewZkManager(fm.config)
	if err != nil {
		return nil, err
	}

	fm.ConfigFinder = &ConfigFinder{zkManager: fm.zkManager, config: fm.config, logger: fm.InternalLogger}
	fm.ServiceFinder = &ServiceFinder{zkManager: fm.zkManager, config: fm.config, logger: fm.InternalLogger, SubscribedService: make(map[string]*common.Service)}

	if err != nil {
		return nil, err
	}

	return fm, nil
}

func DestroyFinder(finder *FinderManager) {
	finder.zkManager.Destroy()
	// todo
}

func onCfgUpdateEvent(c common.Config) int {
	return errors.ConfigSuccess
}
