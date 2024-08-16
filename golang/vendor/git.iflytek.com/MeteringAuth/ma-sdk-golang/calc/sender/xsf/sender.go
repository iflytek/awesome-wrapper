/*
 * @CopyRight: Copyright 2019 IFLYTEK Inc
 * @Author: jianjiang@iflytek.com
 * @LastEditors: Please set LastEditors
 * @Desc: send parsed data to sea [consist hash strategy]
 * @createTime: 2019-09-24 19:33:49
 * @modifyTime: 2020-06-28 16:30:27
 */
package xsf

import (
	"encoding/json"
	"time"

	xsf "git.iflytek.com/AIaaS/xsf/client"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/calc/monitor"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/ma/config"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/ma/tool"
)

// default global sender

var (
	RemoteSVC string = "sea"
	OP        string = "S"
	// resenchan bool = make(chan bool, 1)
)

type calcmsg struct {
	Appid string `json:"appid"`
}

type Producer struct {
	client     *xsf.Client
	rpcTimeout time.Duration

	q    chan *[]byte
	stop bool
}

func (p *Producer) Init(q chan *[]byte) (err error) {
	tool.CalcPrinter.Printf("use xsf: %+v\n", config.Params)

	p.client, err = tool.NewClient("sea-client")
	if err != nil {
		return
	}

	rt, err := p.client.Cfg().GetInt64("sea-client", "rpc_timeout")
	if err != nil {
		return
	}

	p.q = q
	p.rpcTimeout = time.Duration(rt) * time.Millisecond
	return
}

func (p *Producer) Run() {
	if p == nil {
		return
	}

	go func() {
		tool.CalcPrinter.Println("rpc running...")
		for {
			if p.stop {
				tool.CalcPrinter.Println("rpc stop")
				break
			}

			msg := <-p.q
			monitor.WithMsgType(monitor.MsgCntRPC)

			req := xsf.NewReq()
			req.SetParam("D", string(*msg))
			var m calcmsg
			json.Unmarshal(*msg, &m)

			caller := xsf.NewCaller(p.client)
			caller.WithHashKey(m.Appid)

			start := time.Now()
			_, ec, err := caller.Call(RemoteSVC, OP, req, p.rpcTimeout)
			cost := time.Since(start)
			monitor.WithRPCCost(cost, p.rpcTimeout)
			if err != nil {
				monitor.WithMsgErr(monitor.MsgErrRPC)
				tool.L.Errorw("calc-sdk | rpc | send data to sea failed, push back to q", "code", ec, "error", err, "to", p.rpcTimeout.String(), m.Appid, *msg)
				select {
				case p.q <- msg:
				default:
					monitor.WithMsgErr(monitor.MsgErrQFull)
					tool.L.Errorw("calc-sdk | rpc | push back to q full", m.Appid, *msg)
				}
			} else {
				if config.C.Log.Level == "debug" {
					tool.L.Debugw("calc-sdk | rpc | send message", "msg", string(*msg))
				}
			}
		}
	}()
}

func (p *Producer) Fini() {
	if p != nil {
		p.stop = true
		if p.client != nil {
			// TODO: zk deadlock here
			// xsf.DestroyClient(p.client)
		}
	}
}
