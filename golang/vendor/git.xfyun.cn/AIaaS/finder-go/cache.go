package finder

import (
	"encoding/json"
	"fmt"
	"log"

	common "git.xfyun.cn/AIaaS/finder-go/common"
	errors "git.xfyun.cn/AIaaS/finder-go/errors"
	"git.xfyun.cn/AIaaS/finder-go/utils/fileutil"
)

func CacheZkInfo(cachePath string, zkInfo *common.ZkInfo) error {
	cachePath = fmt.Sprintf("%s/zk_%s.findercache", cachePath, "info")
	data, err := json.Marshal(zkInfo)
	if err != nil {
		log.Println(err)
		return err
	}
	err = fileutil.WriteFile(cachePath, data)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func GetZkInfoFromCache(cachePath string) (*common.ZkInfo, error) {
	cachePath = fmt.Sprintf("%s/zk_%s.findercache", cachePath, "info")
	data, err := fileutil.ReadFile(cachePath)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	zkInfo := &common.ZkInfo{}
	err = json.Unmarshal(data, zkInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return zkInfo, nil
}

func CacheConfig(cachePath string, config *common.Config) error {
	cachePath = fmt.Sprintf("%s/config_%s.findercache", cachePath, config.Name)
	err := fileutil.WriteFile(cachePath, config.File)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func GetConfigFromCache(cachePath string, name string) (*common.Config, error) {
	cachePath = fmt.Sprintf("%s/config_%s.findercache", cachePath, name)
	exist, err := fileutil.ExistPath(cachePath)
	if err != nil {
		log.Println(err)
	}
	if !exist {
		err = &errors.FinderError{
			Ret:  errors.ConfigMissCacheFile,
			Func: "GetConfigFromCache",
		}

		log.Printf(cachePath, err)
		return nil, err
	}
	data, err := fileutil.ReadFile(cachePath)
	if err != nil {
		return nil, err
	}

	return &common.Config{Name: name, File: data}, nil
}

func CacheService(cachePath string, service *common.Service) error {
	cachePath = fmt.Sprintf("%s/service_%s.findercache", cachePath, service.Name)
	data, err := json.Marshal(service)
	if err != nil {
		log.Println(err)
		return err
	}
	err = fileutil.WriteFile(cachePath, data)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func GetServiceFromCache(cachePath string, name string) (*common.Service, error) {
	cachePath = fmt.Sprintf("%s/service_%s.findercache", cachePath, name)
	data, err := fileutil.ReadFile(cachePath)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	service := &common.Service{}
	err = json.Unmarshal(data, service)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return service, nil
}
