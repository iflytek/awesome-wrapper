package core

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"git.iflytek.com/MeteringAuth/ma-sdk-golang/calc/monitor"
	sender "git.iflytek.com/MeteringAuth/ma-sdk-golang/calc/sender"
	pulsarSender "git.iflytek.com/MeteringAuth/ma-sdk-golang/calc/sender/pulsar"
	rmqSender "git.iflytek.com/MeteringAuth/ma-sdk-golang/calc/sender/rmq"
	xsfSender "git.iflytek.com/MeteringAuth/ma-sdk-golang/calc/sender/xsf"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/ma/config"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/ma/tool"
)

type CalcCore struct {
	s     []sender.MsgSender
	q     chan *[]byte
	pq    chan *[]byte // for pulsar
	pAble bool         // use pulsar
}

func (c *CalcCore) Init() (err error) {
	cfg := config.C.Calc
	tool.CalcPrinter.Printf("init calc-sdk:%+v\n", cfg)

	c.s = make([]sender.MsgSender, 0)
	c.q = make(chan *[]byte, cfg.QueueSize)
	c.pq = make(chan *[]byte, cfg.QueueSize)

	switch config.C.Calc.Use {
	case TypeRPC:
		p := &xsfSender.Producer{}
		if err = p.Init(c.q); err != nil {
			return
		}
		c.s = append(c.s, p)
	case TypeMQ:
		if cfg.RMQ.Able {
			p := &rmqSender.Producer{}
			if err = p.Init(c.q); err != nil {
				return
			}
			c.s = append(c.s, p)
		}
		if cfg.Pulsar.Able {
			p := &pulsarSender.Producer{}
			if err = p.Init(c.pq); err != nil {
				return
			}
			c.pAble = true
			c.s = append(c.s, p)
		}
	default:
		return fmt.Errorf("invalid use type:%v", cfg.Use)
	}

	return
}

func (c *CalcCore) Run() error {
	for _, s := range c.s {
		s.Run()
	}
	return nil
}

func (c *CalcCore) Fini() {
	for _, s := range c.s {
		s.Fini()
	}
}

func (c *CalcCore) Set(appid, subid, channel, funcs string, cnt int) *config.CalcError {
	if c.pAble && config.UsePulsar(appid) {
		msg := c.packMessage(appid, subid, channel, funcs, cnt, ProtocolPulsarVersion)
		select {
		case c.pq <- &msg:
			return nil
		default:
			monitor.WithMsgErr(monitor.MsgErrQFull)
			tool.L.Errorf("calc-sdk | pulsar chan is full", "len", len(c.pq))
			return config.CalcQueueFull
		}
	} else {
		msg := c.packMessage(appid, subid, channel, funcs, cnt, ProtocolVersion)
		select {
		case c.q <- &msg:
			return nil
		default:
			monitor.WithMsgErr(monitor.MsgErrQFull)
			tool.L.Errorf("calc-sdk | msg queue is full", "len", len(c.q))
			return config.CalcQueueFull
		}
	}
}

func (c *CalcCore) packMessage(appid, subid, channel, funcs string, count int, ver string) []byte {
	fs := strings.Split(funcs, ",")
	now := time.Now().Unix()

	var f []Funcs
	for _, v := range fs {
		f = append(f, Funcs{
			Func:  v,
			Count: count,
			Time:  now,
		})
	}

	m := Msg{
		Appid: appid,
		Ver:   ver,
		Uid:   subid,
		Subs: []Subs{
			{
				Sub:   channel,
				Funcs: f,
			},
		},
	}
	msg, _ := json.Marshal(m)
	return msg
}
