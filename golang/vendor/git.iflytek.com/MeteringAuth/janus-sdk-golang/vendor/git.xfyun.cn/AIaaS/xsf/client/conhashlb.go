/*
* @file	consistencyhashlb.go
* @brief  native 一致性哈希策略
*
* @author	jianjiang
* @version	1.0
* @date		2019.4
 */
package xsf

import (
	finder "git.xfyun.cn/AIaaS/finder-go/common"
	"stathat.com/c/consistent"
)

type consistencyHashLB struct {
	sd         *serviceDiscovery
	hashCircle *consistent.Consistent
}

func newConsistencyHashLB(o *conOption) *consistencyHashLB {
	conHashLb := new(consistencyHashLB)
	conHashLb.hashCircle = consistent.New()
	conHashLb.sd = newServiceDiscovery(o.fm)
	conHashLb.sd.registerCallBackFunc(conHashLb.updateHashCircle)
	return conHashLb
}

func (c *consistencyHashLB) updateHashCircle(notifyType string, instance []*finder.ServiceInstance) {
	if notifyType == string(finder.INSTANCEADDED) {
		for _, addrUnit := range instance {
			c.hashCircle.Add(addrUnit.Addr)
		}
	} else if notifyType == string(finder.INSTANCEREMOVE) {
		for _, addrUnit := range instance {
			c.hashCircle.Remove(addrUnit.Addr)
		}
	} else {
	}
}

func (c *consistencyHashLB) Find(param *LBParams) ([]string, error) {
	_, e := c.sd.findAll(param.version, param.svc, param.logId, param.log)
	if e != nil {
		return nil, e
	}
	hashKey := param.hashKey
	addr, _ := c.hashCircle.Get(hashKey)
	if addr == "" {
		return []string{}, nil
	}
	return []string{addr}, nil
}
