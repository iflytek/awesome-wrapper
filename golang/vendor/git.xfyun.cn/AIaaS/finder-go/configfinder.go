package finder

import (
	common "git.xfyun.cn/AIaaS/finder-go/common"
	errors "git.xfyun.cn/AIaaS/finder-go/errors"
	"git.xfyun.cn/AIaaS/finder-go/utils/zkutil"
	"github.com/cooleric/curator"
)

var (
	configEventPrefix = "config_"
)

type ConfigFinder struct {
	config    *common.BootConfig
	zkManager *zkutil.ZkManager
	logger    common.Logger
}

func (f *ConfigFinder) UseConfig(name []string) (map[string]*common.Config, error) {
	var err error
	if len(name) == 0 {
		err = &errors.FinderError{
			Ret:  errors.ConfigMissName,
			Func: "UseConfig",
		}

		return nil, err
	}
	configFiles := make(map[string]*common.Config)
	var data []byte
	for _, n := range name {
		data, err = f.zkManager.GetNodeData(f.zkManager.MetaData.ConfigRootPath + "/" + n)
		if err != nil {
			f.logger.Error(err)
			// get config from cache
			config, err := GetConfigFromCache(f.config.CachePath, n)
			if err != nil {
				f.logger.Error(err)
				//todo
			} else {
				configFiles[n] = config
			}
		} else {
			var fData []byte
			_, fData, err = common.DecodeValue(data)
			if err != nil {
				// todo
			} else {
				config := &common.Config{Name: n, File: fData}
				configFiles[n] = config
				err = CacheConfig(f.config.CachePath, config)
				if err != nil {
					f.logger.Error(err)
				}
			}
		}
	}

	return configFiles, err
}

func (f *ConfigFinder) UseAndSubscribeConfig(name []string, handler common.ConfigChangedHandler) (map[string]*common.Config, error) {
	var err error
	if len(name) == 0 {
		err = &errors.FinderError{
			Ret:  errors.ConfigMissName,
			Func: "UseAndSubscribeConfig",
		}

		return nil, err
	}

	fileChan := make(chan *common.Config)
	for _, n := range name {
		err = f.zkManager.GetNodeDataW(f.zkManager.MetaData.ConfigRootPath+"/"+n, func(c curator.CuratorFramework, e curator.CuratorEvent) error {
			_, file, err := common.DecodeValue(e.Data())
			if err != nil {
				// get config from cache
				config, err := GetConfigFromCache(f.config.CachePath, e.Name())
				if err != nil {
					f.logger.Error(err)
					//todo
					fileChan <- &common.Config{}
				} else {
					fileChan <- config
				}

				return err
			}
			config := &common.Config{Name: e.Name(), File: file}
			err = CacheConfig(f.config.CachePath, config)
			if err != nil {
				f.logger.Error(err)
			}
			fileChan <- config
			return nil
		})
		if err != nil {
			// get config from cache
			config, err := GetConfigFromCache(f.config.CachePath, n)
			if err != nil {
				f.logger.Info(err)
				//todo
				fileChan <- &common.Config{}
			} else {
				fileChan <- config
			}
			continue
		}

		interHandle := ConfigHandle{ChangedHandler: handler, config: f.config}
		zkutil.ConfigEventPool.Append(common.ConfigEventPrefix+n, &interHandle)
	}

	return waitConfigResult(fileChan, len(name)), nil
}

func (f *ConfigFinder) UnSubscribeConfig(name string) error {
	var err error
	if len(name) == 0 {
		err = &errors.FinderError{
			Ret:  errors.ConfigMissName,
			Func: "UnSubscribeConfig",
		}
		return err
	}

	zkutil.ConfigEventPool.Remove(name)

	return nil
}

func waitConfigResult(fileChan chan *common.Config, fileNum int) map[string]*common.Config {
	configFiles := make(map[string]*common.Config)
	index := 0
	for {
		select {
		case c := <-fileChan:
			index++
			if len(c.Name) > 0 {
				configFiles[c.Name] = c
			}
			if index == fileNum {
				close(fileChan)
				return configFiles
			}
		}
	}
}
