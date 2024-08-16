package zkutil

import (
	"sync"

	common "git.xfyun.cn/AIaaS/finder-go/common"
)

type OnServiceUpdateEvent func(common.Service) bool

type ServiceChangedEventPool struct {
	sync.RWMutex
	pool map[string]common.InternalServiceChangedHandler
}

func NewServiceChangedEventPool() *ServiceChangedEventPool {
	p := new(ServiceChangedEventPool)
	p.pool = make(map[string]common.InternalServiceChangedHandler)

	return p
}

func (p *ServiceChangedEventPool) Get() map[string]common.InternalServiceChangedHandler {
	p.RLock()
	defer p.RUnlock()
	return p.pool
}

func (p *ServiceChangedEventPool) Contains(key string) bool {
	p.RLock()
	defer p.RUnlock()
	if _, ok := p.pool[key]; ok {
		return true
	}

	return false
}

func (p *ServiceChangedEventPool) Append(key string, value common.InternalServiceChangedHandler) {
	p.Lock()
	p.pool[key] = value
	p.Unlock()
}

func (p *ServiceChangedEventPool) Remove(key string) {
	p.Lock()
	delete(p.pool, key)
	p.Unlock()
}
