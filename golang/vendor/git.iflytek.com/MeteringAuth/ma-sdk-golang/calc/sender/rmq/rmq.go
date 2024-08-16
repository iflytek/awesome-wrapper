package rmq

import (
	"strings"
	"time"

	"git.iflytek.com/MeteringAuth/ma-sdk-golang/calc/monitor"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/ma/config"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/ma/tool"
)

type Producer struct {
	q       chan *[]byte
	topic   string
	workers []*RMQConn
	stop    bool
}

func (p *Producer) Init(q chan *[]byte) (err error) {
	cfg := config.C.Calc.RMQ

	tool.CalcPrinter.Printf("use rmq: %+v\n", cfg)

	if len(cfg.Endpoint) == 0 || cfg.Topic == "" {
		return config.CalcHostsOrTopicError
	}

	p.topic = cfg.Topic
	p.q = q

	for i := 0; i < cfg.ConsumeNumber; i++ {
		for _, v := range cfg.Endpoint {
			host := strings.Split(v, ":")[0]
			port := strings.Split(v, ":")[1]

			r := &RMQConn{}
			r.connect(host, port, time.Duration(cfg.Timeout)*time.Millisecond)
			if err != nil {
				tool.L.Errorw("calc-sdk | connect to rmq failed when calc init", "error", err)
				return err
			}

			p.workers = append(p.workers, r)
		}
	}

	return nil
}

func (p *Producer) Run() {
	if p == nil {
		return
	}

	for _, r := range p.workers {
		go func(r *RMQConn) {
			tool.CalcPrinter.Println("rmq running...")
			for {
				if p.stop {
					tool.CalcPrinter.Println("rmq stop")
					break
				}

				if !r.isHealth {
					r.reconnect()
					continue
				}

				msg := <-p.q
				monitor.WithMsgType(monitor.MsgCntRMQ)

				if _, err := r.produce(p.topic, *msg); err != nil {
					monitor.WithMsgRmqErr(r.host, r.port)
					tool.L.Errorw("calc-sdk | rmq | send message to rmq error, push back to q", "error", err)
					select {
					case p.q <- msg:
					default:
						monitor.WithMsgErr(monitor.MsgErrQFull)
						tool.L.Errorw("calc-sdk | rmq | send message to rmq full", "msg", string(*msg))
					}
					err = r.reconnect()
					if err != nil {
						tool.L.Errorw("calc-sdk | rmq | reconnect rmq error", "host", r.host, "port", r.port, "error", err)
					} else {
						tool.L.Infow("calc-sdk | rmq | reconnect rmq success", "host", r.host, "port", r.port)
					}
				} else {
					if config.C.Log.Level == "debug" {
						tool.L.Debugw("calc-sdk | rmq | send message", "msg", string(*msg))
					}
				}
			}
		}(r)
	}
}

func (p *Producer) Fini() {
	if p == nil {
		return
	}

	p.stop = true

	for _, r := range p.workers {
		r.fini()
	}
}
