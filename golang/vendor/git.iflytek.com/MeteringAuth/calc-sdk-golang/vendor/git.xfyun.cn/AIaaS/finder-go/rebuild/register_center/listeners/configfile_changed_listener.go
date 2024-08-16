package listeners

import (
	"git.xfyun.cn/AIaaS/finder-go/rebuild/common"
	"git.xfyun.cn/AIaaS/finder-go/rebuild/register_center"
)

type ConfigFileChangedListener struct {
	EventType register_center.EventType
	callback common.ConfigChangedCallback
}

func (c *ConfigFileChangedListener) Type() register_center.EventType {
	return c.EventType
}

func (c *ConfigFileChangedListener) OnMessage(t register_center.EventType, data interface{}) {

}

func NewConfigfileChangedEvent(){

}

