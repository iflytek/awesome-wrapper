package daemon

/*
Message Queue
*/
type MqManager interface {
	Consume(topic, group string) (*MTRMessage, error)
	Init([]string) error
}

var RmqManagerInst RmqManager

type RmqManager struct {
	rmqAdapter
}

func (r *RmqManager) Init(addr []string) error {
	return r.rmqAdapter.Init(addr)
}
func (r *RmqManager) Consume(topic, group string) (ConsumeR *MTRMessage, ConsumeE error) {
	return r.rmqAdapter.Consume(topic, group)
}
