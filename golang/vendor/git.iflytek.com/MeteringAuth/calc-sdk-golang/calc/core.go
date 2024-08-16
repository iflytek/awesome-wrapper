/*
 * @CopyRight: Copyright 2019 IFLYTEK Inc
 * @Author: jianjiang@iflytek.com
 * @LastEditors: Please set LastEditors
 * @Desc:
 * @createTime: 2020-07-23 16:15:07
 * @modifyTime: 2020-07-24 19:16:25
 */
package calc

import (
	"encoding/json"
	"fmt"
	"git.iflytek.com/MeteringAuth/calc-sdk-golang/calc/internal/logger"
	"strings"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
	"git.iflytek.com/MeteringAuth/calc-sdk-golang/calc/internal/rmq_adapter"
)

type CalcCore struct {
	q              chan []byte
	qSize          int
	producerNumber int
	producer       []*Producer
	isStop         bool
}

func (c *CalcCore) Init(qsize, pnumber int, topic string, hosts []string, timeout time.Duration) error {
	// 初始化producer
	c.qSize = qsize
	c.producerNumber = pnumber
	c.q = make(chan []byte, qsize)
	if len(hosts) == 0 || topic == "" {
		return CalcHostsOrTopicError
	}
	for i := 0; i < c.producerNumber; i++ {
		for _, v := range hosts {
			host := strings.Split(v, ":")[0]
			port := strings.Split(v, ":")[1]
			P := &Producer{
				topic:    topic,
				q:        c.q,
				isStop:   false,
				isHealth: true,
				rmqInst: func() *RMQConn {
					r := &RMQConn{}
					if err := r.connect(host, port, timeout); err != nil {
						logger.Errorw("connect to rmq failed when calc init", "error", err)
						return nil
					}
					return r
				}(),
			}
			if P.rmqInst == nil {
				return CalcServerConnectFailed
			}
			c.producer = append(c.producer, P)
		}
	}
	return nil
}

func (c *CalcCore) Fini() {

	maxWait := time.NewTicker(10 * time.Second)
	checkInterval := time.NewTicker(1 * time.Second)

	fini := func() {
		for _, p := range c.producer {
			p.fini()
		}
	}
	defer func() {
		c.producer = make([]*Producer, 0)
	}()
	for {
		select {
		case <-maxWait.C:
			logger.Infow("wait for timeout when sdk finish")
			fini()
			return
		case <-checkInterval.C:
			logger.Infow("wait for queue clean up", "queue size", len(c.q))
			if len(c.q) == 0 {
				logger.Infow("queue have been cleaned when sdk finish")
				fini()
				return
			}
		}
	}
}

func (c *CalcCore) Set(appid, subid, channel, funcs string, cnt int) *CalcError {
	msg := c.packMessage(appid, subid, channel, funcs, cnt)
	logger.Debugw("meter message", "msg", string(msg))
	select {
	case c.q <- msg:
	default:
		return CalcQueueFull
	}
	return nil
}

func (c *CalcCore) Run() {
	for _, p := range c.producer {
		if p == nil {
			logger.Fatalw("obtain nil producer when start producer routine")
			return
		}
		p.isStop = false
		go p.run()
	}
}

func (c *CalcCore) packMessage(appid, subid, channel, funcs string, count int) []byte {
	m := Msg{
		Appid: appid,
		Ver:   ProtocolVersion,
		Uid:   subid,
		Subs: []Subs{
			{
				Sub: channel,
				Funcs: []Funcs{
					{
						Func:  funcs,
						Count: count,
						Time:  time.Now().Unix(),
					},
				},
			},
		},
	}
	msg, _ := json.Marshal(m)
	return msg
}

type Producer struct {
	q      chan []byte
	topic  string
	isStop bool

	rmqInst  *RMQConn
	isHealth bool
}

func (p *Producer) run() {
	for !p.isStop {
		msg := <-p.q
		if _, err := p.send(p.topic, msg); err != nil {
			// set back to queue when sending failed
			logger.Errorw("send message to rmq error", "error", err)
			p.q <- msg
			p.isHealth = false
			if err := p.repair(); err != nil {
				// has stoped
				break
			}
			p.isHealth = true
		}
	}
}

func (p *Producer) repair() (err error) {

	c := 0
	for !p.isStop {
		if err = p.rmqInst.reconnect(); err == nil {
			logger.Infow("reconnect success")
			return
		}
		c += 1
		slpt := func() time.Duration {
			tms := 100*c + 10
			if tms > 5000 {
				tms = 5000
			}
			return time.Duration(tms) * time.Millisecond
		}()
		time.Sleep(slpt)
	}
	return
}

func (p *Producer) send(topic string, msg []byte) (int64, error) {
	return p.rmqInst.produce(topic, msg)
}

func (p *Producer) fini() {
	logger.Infow("producer finish")
	p.isStop = true
	p.rmqInst.fini()
}

type RMQConn struct {
	transportFactory *thrift.TBufferedTransportFactory
	protocolFactory  *thrift.TBinaryProtocolFactory
	transport        *thrift.TSocket
	useTransport     *thrift.TTransport
	client           *rmq_adapter.MTRMessageServiceClient
	host             string
	port             string
	timeout          time.Duration
}

func (r *RMQConn) connect(host, port string, timeout time.Duration) error {
	var err error
	r.host = host
	r.port = port
	r.timeout = timeout
	addr := fmt.Sprintf("%v:%v", host, port)
	r.transportFactory = thrift.NewTBufferedTransportFactory(100000)
	r.protocolFactory = thrift.NewTBinaryProtocolFactoryDefault()
	r.transport, err = thrift.NewTSocketTimeout(addr, timeout)
	if err != nil {
		return err
	}
	useTransport := r.transportFactory.GetTransport(r.transport)
	r.client = rmq_adapter.NewMTRMessageServiceClientFactory(useTransport, r.protocolFactory)
	if err := r.transport.Open(); err != nil {
		return err
	}
	return err

}

func (r *RMQConn) reconnect() (err error) {
	r.fini()
	if err = r.connect(r.host, r.port, r.timeout); err != nil {
		// log it
		return
	}
	return

}

func (r *RMQConn) fini() {
	if err := r.transport.Close(); err != nil {
		logger.Errorw("rmq fini failed", "error", err)
	}
}

func (r *RMQConn) produce(topic string, data []byte) (rs int64, err error) {

	var msg rmq_adapter.MTRMessage
	msg.Topic = topic
	msg.Body = data //强转 成字节类型的切片
	msg.Protocol = rmq_adapter.MTRProtocol_PERSONALIZED

	var resendMaxTimes = 3 //重新发送三次
	for i := 0; i < resendMaxTimes; i++ {
		if rs, err = r.client.Produce(&msg, true); err != nil {
			logger.Warnw("retry when producing message to rmq", "retry count", i)
			continue
		}
		return
	}
	return

}
