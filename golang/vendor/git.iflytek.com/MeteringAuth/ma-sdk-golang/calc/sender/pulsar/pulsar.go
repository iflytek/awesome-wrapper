package pulsar

import (
	"context"

	"git.iflytek.com/AIaaS/mqutils"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/calc/monitor"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/ma/config"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/ma/tool"
)

type Producer struct {
	client   *mqutils.MQInstance
	producer mqutils.Producer
	op       *mqutils.ProducerOption

	q         chan *[]byte
	threadNum int
	stop      bool
}

func (p *Producer) Init(q chan *[]byte) (err error) {
	if p == nil {
		return nil
	}
	cfg := config.C.Calc.Pulsar

	tool.CalcPrinter.Printf("use pulsar: %+v\n", cfg)

	if cfg.Endpoint == "" || cfg.Topic == "" {
		return config.CalcHostsOrTopicError
	}

	p.client, err = mqutils.NewMQInstance(cfg.Endpoint)
	if err != nil {
		return
	}

	p.producer, err = p.client.CreateProducer(cfg.Topic)
	if err != nil {
		return err
	}

	p.op = &mqutils.ProducerOption{
		Key:                 cfg.IDC,
		ReplicationClusters: nil,
		DisableReplication:  false,
	}

	p.q = q

	p.threadNum = cfg.ThreadNum
	return nil
}

func (p *Producer) Run() {
	if p == nil {
		return
	}

	for i := 0; i < p.threadNum; i++ {
		go func() {
			tool.CalcPrinter.Println("pulsar running...")
			for {
				if p.stop {
					tool.CalcPrinter.Println("pulsar stop")
					break
				}

				msg := <-p.q
				monitor.WithMsgType(monitor.MsgCntPulsar)

				if id, err := p.send(*msg); err != nil {
					monitor.WithMsgErr(monitor.MsgErrPulsar)
					tool.L.Errorw("calc-sdk | pulsar | send message to pulsar error, push back to q", "error", err)
					select {
					case p.q <- msg:
					default:
						monitor.WithMsgErr(monitor.MsgErrQFull)
						tool.L.Errorw("calc-sdk | pulsar | push back to q full", "msg", string(*msg))
					}
				} else {
					if config.C.Log.Level == "debug" {
						tool.L.Debugw("calc-sdk | pulsar | send message", "msg", string(*msg), "message id", id)
					}
				}
			}
		}()
	}
}

func (p *Producer) Fini() {
	if p != nil {
		p.stop = true
		p.producer.Close()
		p.client.DestroyInstance()
	}
}

func (p *Producer) send(msg []byte) (interface{}, error) {
	return p.producer.Send(context.Background(), msg, p.op)
}
