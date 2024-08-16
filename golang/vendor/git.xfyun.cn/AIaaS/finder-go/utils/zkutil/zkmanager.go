package zkutil

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	common "git.xfyun.cn/AIaaS/finder-go/common"
	companion "git.xfyun.cn/AIaaS/finder-go/companion"
	"git.xfyun.cn/AIaaS/finder-go/utils/arrayutil"

	"github.com/cooleric/curator"
	"github.com/cooleric/go-zookeeper/zk"
)

var (
	hc               *http.Client
	url              string
	zkExit           chan bool
	ConfigEventPool  *ConfigChangedEventPool
	ServiceEventPool *ServiceChangedEventPool
	ConsumeEventPool *ServiceChangedEventPool
	//mutex 	         sync.Mutex
)

type OnZkSessionExpiredEvent func()

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

	ConfigEventPool = NewConfigChangedEventPool()
	ServiceEventPool = NewServiceChangedEventPool()
}

// ZkManager for operate zk
type ZkManager struct {
	MetaData    *common.ZkInfo
	watcherPool map[string]map[string]curator.BackgroundCallback
	// watcherPool 	  *WatcherPool
	tempNodePool      map[string]map[string][]byte
	checkZkInfoTicker *time.Ticker
	zkClient          curator.CuratorFramework
	expired           bool
}

// NewZkManager for create ZkManager
func NewZkManager(config *common.BootConfig) (*ZkManager, error) {
	checkConfig(config)
	zm, err := init_(config)
	if err != nil {
		return nil, err
	}
	zm.checkZkInfoTicker = time.NewTicker(config.TickerDuration)
	// 开启一个协程去检测zkinfo变化
	go watchZkAddr(zm)
	// 创建zk连接
	err = connect(zm, config.ZkMaxRetryNum, config.ZkMaxSleepTime, config.ZkConnectTimeout, config.ZkSessionTimeout)
	if err != nil {
		return nil, err
	}
	// 增加监听
	addListeners(zm)

	return zm, nil
}

func checkConfig(c *common.BootConfig) {
	if c.TickerDuration <= 0 {
		c.TickerDuration = 30 * time.Second
	}
	if c.ZkConnectTimeout <= 0 {
		c.ZkConnectTimeout = 3 * time.Second
	}
	if c.ZkSessionTimeout <= 0 {
		c.ZkSessionTimeout = 3 * time.Second
	}
	if c.ZkMaxRetryNum < 0 {
		c.ZkMaxRetryNum = 3
	}
	if c.ZkMaxSleepTime <= 0 {
		c.ZkMaxSleepTime = 15 * time.Second
	}
}

func init_(config *common.BootConfig) (*ZkManager, error) {
	url = config.CompanionUrl + fmt.Sprintf("/finder/query_zk_info?project=%s&group=%s&service=%s&version=%s", config.MeteData.Project, config.MeteData.Group, config.MeteData.Service, config.MeteData.Version)
	metadata, err := companion.GetZkInfo(hc, url)
	if err != nil {
		return nil, err
	}
	zm := &ZkManager{
		MetaData:     metadata,
		watcherPool:  initWatcherPool(),
		tempNodePool: initTempNodePool(),
	}

	return zm, nil
}

func initWatcherPool() map[string]map[string]curator.BackgroundCallback {
	watcherPool := make(map[string]map[string]curator.BackgroundCallback)
	watcherPool["GetNodeDataW"] = make(map[string]curator.BackgroundCallback)
	watcherPool["GetChildrenW"] = make(map[string]curator.BackgroundCallback)

	return watcherPool
}

func initTempNodePool() map[string]map[string][]byte {
	tempNodePool := make(map[string]map[string][]byte)
	tempNodePool["CreateTempPath"] = make(map[string][]byte)
	tempNodePool["CreateTempPathWithData"] = make(map[string][]byte)

	return tempNodePool
}

func (zm *ZkManager) CreatePath(path string) (string, error) {
	return zm.zkClient.Create().CreatingParentsIfNeeded().ForPath(path)
}

func (zm *ZkManager) CreatePathWithData(path string, data []byte) (string, error) {
	return zm.zkClient.Create().CreatingParentsIfNeeded().ForPathWithData(path, data)
}

func (zm *ZkManager) CreateTempPath(path string) (string, error) {
	//mutex.Lock()
	log.Println("zm.tempNodePool[\"CreateTempPath\"]", path)
	zm.tempNodePool["CreateTempPath"][path] = nil
	//mutex.Unlock()
	return zm.zkClient.Create().CreatingParentsIfNeeded().WithMode(curator.EPHEMERAL).ForPath(path)
}

func (zm *ZkManager) CreateTempPathWithData(path string, data []byte) (string, error) {
	//mutex.Lock()
	log.Println("zm.tempNodePool[\"CreateTempPathWithData\"]", path)
	zm.tempNodePool["CreateTempPathWithData"][path] = data
	//mutex.Unlock()
	return zm.zkClient.Create().CreatingParentsIfNeeded().WithMode(curator.EPHEMERAL).ForPathWithData(path, data)
}

func (zm *ZkManager) UpdateData(path string, data []byte) (*zk.Stat, error) {
	return zm.zkClient.SetData().Compressed().ForPathWithData(path, data)
}

func (zm *ZkManager) ExistsNode(path string) (*zk.Stat, error) {
	return zm.zkClient.CheckExists().ForPath(path)
}

func (zm *ZkManager) ExistsNodeW(path string) (*zk.Stat, error) {
	return zm.zkClient.CheckExists().Watched().ForPath(path)
}

func (zm *ZkManager) UpdateDataWithCheckExists(path string, data []byte) (*zk.Stat, error) {
	s, err := zm.zkClient.CheckExists().ForPath(path)
	if err != nil {
		return nil, err
	}
	if s != nil {
		return zm.zkClient.SetData().Compressed().ForPathWithData(path, data)
	}

	return s, nil
}

func (zm *ZkManager) GetNodeData(path string) ([]byte, error) {
	return zm.zkClient.GetData().ForPath(path)
}

func (zm *ZkManager) GetNodeDataW(path string, c curator.BackgroundCallback) error {
	zm.watcherPool["GetNodeDataW"][path] = c
	_, err := zm.zkClient.GetData().InBackgroundWithCallback(c).Watched().ForPath(path)
	return err
}

func (zm *ZkManager) GetNodeDataForCallback(path string, c curator.BackgroundCallback) error {
	_, err := zm.zkClient.GetData().InBackgroundWithCallback(c).Watched().ForPath(path)
	return err
}

func (zm *ZkManager) GetNodeDataWForRecover(path string) error {
	_, err := zm.zkClient.GetData().InBackground().Watched().ForPath(path)
	return err
}

func (zm *ZkManager) GetChildrenNodeUseWatch(path string, watcher curator.Watcher) ([]string, error) {
	return zm.zkClient.GetChildren().UsingWatcher(watcher).ForPath(path)
}

func (zm *ZkManager) GetChildrenW(path string, c curator.BackgroundCallback) error {
	//mutex.Lock()
	//defer mutex.Unlock()
	zm.watcherPool["GetChildrenW"][path] = c
	_, err := zm.zkClient.GetChildren().InBackgroundWithCallback(c).Watched().ForPath(path)
	return err
}

func (zm *ZkManager) GetChildrenForCallback(path string, c curator.BackgroundCallback) error {
	_, err := zm.zkClient.GetChildren().InBackgroundWithCallback(c).Watched().ForPath(path)
	return err
}

func (zm *ZkManager) GetChildrenWForRecover(path string) error {
	_, err := zm.zkClient.GetChildren().InBackground().Watched().ForPath(path)
	return err
}

func (zm *ZkManager) GetChildren(path string) ([]string, error) {
	return zm.zkClient.GetChildren().ForPath(path)
}

func (zm *ZkManager) RemoveInRecursive(path string) error {
	return zm.zkClient.Delete().DeletingChildrenIfNeeded().ForPath(path)
}

func (zm *ZkManager) AddListener(listener curator.CuratorListener) {
	zm.zkClient.CuratorListenable().AddListener(listener)
}

func (zm *ZkManager) AddConnectionListener(listener curator.ConnectionStateListener) {
	zm.zkClient.ConnectionStateListenable().AddListener(listener)
}

func (zm *ZkManager) RemoveListener(listener curator.CuratorListener) {
	zm.zkClient.CuratorListenable().RemoveListener(listener)
}

func (zm *ZkManager) Destroy() {
	zkExit <- true

	zm.checkZkInfoTicker.Stop()

	err := close(zm)
	if err != nil {

	}
}

func (zm *ZkManager) OnZkSessionExpired() {
	if !zm.expired {
		return
	}
	recoverTempNode(zm)
	recoverWatcher(zm)

	zm.expired = false
}

func recoverWatcher(zm *ZkManager) {
	for f, v := range zm.watcherPool {
		switch f {
		case "GetNodeDataW":
			for p, _ := range v {
				for {
					err := zm.GetNodeDataWForRecover(p)
					if err != nil {
						log.Println(err)
						//time.Sleep(time.Millisecond * 200)
					} else {
						break
					}
				}
			}
		case "GetChildrenW":
			for p, _ := range v {
				for {
					err := zm.GetChildrenWForRecover(p)
					if err != nil {
						log.Println(err)
						//time.Sleep(time.Millisecond * 200)
					} else {
						break
					}
				}
			}
		}
	}
}

func recoverTempNode(zm *ZkManager) {
	for f, v := range zm.tempNodePool {
		switch f {
		case "CreateTempPath":
			log.Println("for CreateTempPath")
			for p, _ := range v {
				for {
					log.Println("recover TempNode:", p)
					node, err := zm.ExistsNode(p)
					if err != nil {
						log.Println("zm.ExistsNode:", err)
						continue
					}
					if node == nil {
						r, err := zm.CreateTempPath(p)
						if err != nil {
							log.Println(err)
							//time.Sleep(time.Millisecond * 200)
						} else {
							log.Println(r)
							break
						}
					} else {
						log.Println("zm.ExistsNode:true")
						break
					}
				}
			}
		case "CreateTempPathWithData":
			log.Println("for CreateTempPathWithData")
			for p, v := range v {
				for {
					log.Println("recover TempPathWithData:", p)
					node, err := zm.ExistsNode(p)
					if err != nil {
						log.Println("zm.ExistsNode:", err)
						continue
					}
					if node == nil {
						r, err := zm.CreateTempPathWithData(p, v)
						if err != nil {
							log.Println(err)
							//time.Sleep(time.Millisecond * 200)
						} else {
							log.Println(r)
							break
						}
					} else {
						log.Println("zm.ExistsNode:true")
						break
					}
				}
			}
		}
	}
}

func onZkInfoChanged(zm *ZkManager) {
	// todo.
}

func onEventNodeChildrenChanged(c curator.CuratorFramework, e curator.CuratorEvent) error {
	serviceName := getServiceName(e.Path(), 1)
	serviceEvent, ok := ServiceEventPool.Get()[common.ServiceEventPrefix+serviceName]
	if ok {
		serviceEvent.OnServiceInstanceChanged(serviceName, e.Children())
		return nil
	}

	return errors.New(common.ServiceEventPrefix + serviceName + " couldn't be found in ServiceEventPool")
}

func onEventNodeCreated(e *zk.Event) {

}

func onEventNodeDataChanged(c curator.CuratorFramework, e curator.CuratorEvent) error {
	configEvent, ok := ConfigEventPool.Get()[common.ConfigEventPrefix+e.Name()]
	if ok {
		configEvent.OnConfigFileChanged(e.Name(), e.Data())
		return nil
	}
	serviceEvent, ok := ServiceEventPool.Get()[common.ServiceProviderEventPrefix+e.Name()]
	if ok {
		serviceEvent.OnServiceInstanceConfigChanged(getServiceName(e.Path(), 2), e.Name(), e.Data())
		return nil
	}
	// serviceEvent, ok := ServiceEventPool.Get()[serviceConsumerEventPrefix+e.Name()]
	// if ok {
	// 	serviceEvent.OnServiceConfigChanged(getServiceName(e.Path(), 2), e.Data())
	// 	return nil
	// }
	serviceEvent, ok = ServiceEventPool.Get()[common.ServiceConfEventPrefix+e.Name()]
	if ok {
		serviceEvent.OnServiceConfigChanged(getServiceName(e.Path(), 1), e.Data())
		return nil
	}

	return nil
}

func onEventNodeDeleted(e *zk.Event) {

}

func onEventNotWatching(e *zk.Event) {

}

func getServiceName(path string, deep int) string {
	items := strings.Split(path, "/")
	if len(items) >= 3 {
		return items[len(items)-1-deep]
	}

	return ""
}

func watchZkAddr(zm *ZkManager) {
	for t := range zm.checkZkInfoTicker.C {
		//log.Println(t)
		if t.IsZero() {

		}
		metadata, err := companion.GetZkInfo(hc, url)
		if err != nil {
			// todo.
			continue
		}
		vchanged := checkAddr(metadata.ZkAddr, zm.MetaData.ZkAddr)
		if vchanged {
			zm.MetaData.ZkAddr = metadata.ZkAddr
			zm.MetaData.ConfigRootPath = metadata.ConfigRootPath
			zm.MetaData.ServiceRootPath = metadata.ServiceRootPath
			// 通知zkinfo更新，执行相关逻辑
			onZkInfoChanged(zm)
		}
	}
}

func checkAddr(n []string, o []string) bool {
	vchanged := false
	for _, nv := range o {
		if !arrayutil.Contains(nv, o) {
			vchanged = true
		}
	}

	return vchanged
}
