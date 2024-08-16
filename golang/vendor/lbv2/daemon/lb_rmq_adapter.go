package daemon

import (
	"fmt"
	"github.com/cihub/seelog"
	"sync/atomic"
)

var RmqAdapter rmqAdapter

type rmqAdapter struct {
	clientList []rmqAdapterItem
}

var rmqIx int64 = 0
var addrList []string

func (r *rmqAdapter) Init(addr []string) (err error) {
	for _, addrItem := range addr {
		addrList = append(addrList, addrItem)
		var rmqAdapterItemTmp rmqAdapterItem
		if rmqAdapterItemTmp.Init(addrItem) != nil {
			err = fmt.Errorf("rmq init fail. addr:%v", addrItem)
			return
		} else {
			r.clientList = append(r.clientList, rmqAdapterItemTmp)
		}
	}
	return
}

func (r *rmqAdapter) Produce(topic string, body string) (produceReply int64, produceErr error) {
	var produceFlag = true
	ProduceR, ProduceE := r.clientList[atomic.LoadInt64(&rmqIx)].Produce(topic, body)
	if ProduceR != 0 || ProduceE != nil {
		produceFlag = false
		r.clientList = append(r.clientList[:atomic.LoadInt64(&rmqIx)], r.clientList[atomic.LoadInt64(&rmqIx)+1:]...)
		if len(r.clientList) < 10 {
			r.Init(addrList)
		}
		atomic.StoreInt64(&rmqIx, 0)
	}
	if !produceFlag {
		ProduceR, ProduceE = r.clientList[atomic.LoadInt64(&rmqIx)].Produce(topic, body)
		if ProduceR != 0 || ProduceE != nil {
			seelog.Error("rmq write fail.")
		}
	}
	return ProduceR, ProduceE
}
func (r *rmqAdapter) Consume(topic, group string) (ConsumeR *MTRMessage, ConsumeE error) {
	var consumeFlag = true
	ConsumeR, ConsumeE = r.clientList[atomic.LoadInt64(&rmqIx)].Consume(topic, group)
	if ConsumeE != nil {
		consumeFlag = false
		r.clientList = append(r.clientList[:atomic.LoadInt64(&rmqIx)], r.clientList[atomic.LoadInt64(&rmqIx)+1:]...)
		if len(r.clientList) < 10 {
			r.Init(addrList)
		}
		atomic.StoreInt64(&rmqIx, 0)
	}
	if !consumeFlag {
		ConsumeR, ConsumeE = r.clientList[atomic.LoadInt64(&rmqIx)].Consume(topic, group)
	}
	return
}
