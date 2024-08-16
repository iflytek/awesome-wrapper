package rmq

import (
	"fmt"
	"git.iflytek.com/AIaaS/thrift"
	"time"

	"git.iflytek.com/MeteringAuth/ma-sdk-golang/calc/internal/rmq_adapter"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/ma/tool"
)

type RMQConn struct {
	transportFactory *thrift.TBufferedTransportFactory
	protocolFactory  *thrift.TBinaryProtocolFactory
	transport        *thrift.TSocket
	// useTransport     *thrift.TTransport
	client   *rmq_adapter.MTRMessageServiceClient
	host     string
	port     string
	timeout  time.Duration
	isHealth bool
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

	r.isHealth = true
	return nil
}

func (r *RMQConn) reconnect() (err error) {
	r.fini()
	if err = r.connect(r.host, r.port, r.timeout); err != nil {
		tool.L.Errorw("calc-sdk | rmq connect failed", "error", err, "host", r.host, "port", r.port, "timeout", r.timeout.String())
		return
	}
	return

}

func (r *RMQConn) fini() {
	if err := r.transport.Close(); err != nil {
		tool.L.Errorw("calc-sdk | rmq fini failed", "error", err)
	}
	r.isHealth = false
}

func (r *RMQConn) produce(topic string, data []byte) (rs int64, err error) {
	var msg rmq_adapter.MTRMessage
	msg.Topic = topic
	msg.Body = data //强转 成字节类型的切片
	msg.Protocol = rmq_adapter.MTRProtocol_PERSONALIZED

	return r.client.Produce(&msg, true)

	// var resendMaxTimes = 3 //重新发送三次
	// for i := 0; i < resendMaxTimes; i++ {
	// 	if rs, err = r.client.Produce(&msg, true); err != nil {
	// 		tool.L.Warnw("retry when producing message to rmq", "retry count", i)
	// 		continue
	// 	}
	// 	return
	// }
	// return
}
