package zkutil

import (
	"sync"

	common "git.xfyun.cn/AIaaS/finder-go/common"
)

type ConfigChangedEventPool struct {
	sync.RWMutex
	pool map[string]common.InternalConfigChangedHandler
}

func NewConfigChangedEventPool() *ConfigChangedEventPool {
	p := new(ConfigChangedEventPool)
	p.pool = make(map[string]common.InternalConfigChangedHandler)

	return p
}

func (p *ConfigChangedEventPool) Get() map[string]common.InternalConfigChangedHandler {
	p.RLock()
	defer p.RUnlock()
	return p.pool
}

func (p *ConfigChangedEventPool) Contains(key string) bool {
	p.RLock()
	defer p.RUnlock()
	if _, ok := p.pool[key]; ok {
		return true
	}

	return false
}

func (p *ConfigChangedEventPool) Append(key string, value common.InternalConfigChangedHandler) {
	p.Lock()
	p.pool[key] = value
	p.Unlock()
}

func (p *ConfigChangedEventPool) Remove(key string) {
	p.Lock()
	delete(p.pool, key)
	p.Unlock()
}
