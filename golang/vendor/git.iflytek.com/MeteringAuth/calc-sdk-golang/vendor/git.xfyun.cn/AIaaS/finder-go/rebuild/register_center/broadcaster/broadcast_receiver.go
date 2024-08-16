package broadcaster

import (
	"git.xfyun.cn/AIaaS/finder-go/rebuild/register_center"
	"sync"
)

type BroadCastReceiver struct {
	listeners map[register_center.EventType]sync.Map
	lock sync.RWMutex
	//messageQueues sync.Map
}

func NewBroadCasterReceiver()*BroadCastReceiver{
	return &BroadCastReceiver{
		listeners: map[register_center.EventType]sync.Map{},
		lock:      sync.RWMutex{},
	}
}

func (b *BroadCastReceiver) RegisterListener(ls register_center.Listener) {
	lss,ok:=b.listeners[ls.Type()]
	if ok{
		lss.Store(ls,true)
		return
	}
	b.lock.Lock()
	defer b.lock.Unlock()

	lss,ok =b.listeners[ls.Type()]
	if ok{
		lss.Store(ls,true)
		return
	}
	lss = sync.Map{}
	b.listeners[ls.Type()] = lss
	lss.Store(ls,true)
	return
}

func (b *BroadCastReceiver) RemoveListener(ls register_center.Listener) {
	lss,ok:=b.listeners[ls.Type()]
	if ok{
		lss.Delete(ls)
	}
}

func (b *BroadCastReceiver) SendBroadCast(e register_center.Event) {
	lss,ok:=b.listeners[e.Type()]
	if ok{
		lss.Range(func(key, value interface{}) bool {
			key.(register_center.Listener).OnMessage(e.Type(),e.Data())
			return true
		})
	}
}


var defaultBroadCaster =  NewBroadCasterReceiver()



func RegisterListener(ls register_center.Listener){
	defaultBroadCaster.RegisterListener(ls)
}

func  RemoveListener(ls register_center.Listener) {
	defaultBroadCaster.RemoveListener(ls)
}

func SendBroadCast(e register_center.Event){
	defaultBroadCaster.SendBroadCast(e)
}


