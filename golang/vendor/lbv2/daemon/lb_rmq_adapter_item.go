package daemon

import (
	"git.apache.org/thrift.git/lib/go/thrift"
	"log"
)

type rmqAdapterItem struct {
	transportFactory *thrift.TBufferedTransportFactory
	protocolFactory  *thrift.TBinaryProtocolFactory
	transport        *thrift.TSocket
	useTransport     *thrift.TTransport
	client           *MTRMessageServiceClient
}

func (r *rmqAdapterItem) Init(addr string) (err error) {
	r.transportFactory = thrift.NewTBufferedTransportFactory(100000)
	r.protocolFactory = thrift.NewTBinaryProtocolFactoryDefault()
	r.transport, err = thrift.NewTSocket(addr)
	if err != nil {
		log.Fatal(err)
	}
	useTransport := r.transportFactory.GetTransport(r.transport)
	r.client = NewMTRMessageServiceClientFactory(useTransport, r.protocolFactory)
	if err := r.transport.Open(); err != nil {
		log.Fatal(err)
	}
	return
}

func (r *rmqAdapterItem) Produce(topic string, body string) (produceReply int64, produceErr error) {
	var msg MTRMessage
	msg.Topic = topic
	msg.Body = []byte(body)
	msg.Protocol = MTRProtocol_PERSONALIZED
	produceReply, produceErr = r.client.Produce(&msg, true)
	return
}
func (r *rmqAdapterItem) Consume(topic, group string) (consumeReply *MTRMessage, consumeErr error) {
	consumeReply, consumeErr = r.client.Consume(topic, group)
	return
}
