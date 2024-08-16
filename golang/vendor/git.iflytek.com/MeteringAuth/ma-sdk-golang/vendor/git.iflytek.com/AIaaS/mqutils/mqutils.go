package mqutils

import (
	"context"
	"strings"
	"sync"
	"time"
)
//定义生产者类型
type ProducerOption struct {
	Key                 string   // 生产者根据 key 区分消息来源
	ReplicationClusters []string // 对当前覆盖默认设置，手动指定要同步到哪个集群
	DisableReplication  bool     // Disable the replication for this message 不允许复制消息
}
//定义消息接口
type Message interface {
	Payload() []byte // 消息内容
	Topic() string   // 消息的 topic
	Key() string     // 对应 ProducerOption 的 Key
}

//定义生产者接口
type Producer interface {
	// for PulsarMQ, interface{} -> MessageID
	// for RabbitMQ, interface{} is invalid 无效
	Send(context.Context, []byte, *ProducerOption) (interface{}, error)
	Close()
}
//定义消费者接口
type Consumer interface {
	Recv(context.Context) (Message, error)
	BatchsRecv(ctx context.Context, minnum int) ([]Message, error)
	Ack(Message) error
	Nack(Message) error
	Close()
}
//定义客户端接口
type Client interface {
	NewProducer(topic string) (Producer, error)
	NewConsumer(subname, topic string) (Consumer, error)
	Close()
}
//定义消息队列实例
type MQInstance struct {
	mqurl     string
	client    Client
	producers sync.Map
}

// 集群模式地址 "pulsar://10.1.87.11:6650,10.1.87.12:6650,10.1.87.18:6650"
func NewMQInstance(mqurl string) (*MQInstance, error) {
	if strings.HasPrefix(mqurl, "pulsar") {
		c, err := NewPulsarClient(mqurl)
		return &MQInstance{
			mqurl: mqurl,
			client: c,
		}, err
	}

	return nil, ErrInvalidUrl
}
//销毁实例
func (instance *MQInstance) DestroyInstance() {
	instance.client.Close()
}
//创建生产者
func (instance *MQInstance) CreateProducer(topic string) (Producer, error) {
	return instance.client.NewProducer(topic)
}

// 应用需要保证订阅(同 kafka 消费者组)不会经常变化
// 消息需要有 ack，应用自己发送 ack
func (instance *MQInstance) CreateConsumer(subname, topic string) (Consumer, error) {
	return instance.client.NewConsumer(subname, topic)
}

// 发布消息，如果 topic 对应的生产者不存在，会自动创建
func (instance *MQInstance) SendMsg(topic string, payload []byte) (interface{}, error) {
	var producer interface{}
	var ok bool
	var err error
	if producer, ok = instance.producers.Load(topic); !ok {
		if producer, err = instance.client.NewProducer(topic); err == nil {
			instance.producers.Store(topic, producer)
		} else {
			return nil, err
		}
	}

	return producer.(Producer).Send(context.Background(), payload, nil)
}

func (instance *MQInstance) SetMsgProcesser(
	subname string,
	topic string,
	tmo time.Duration,
	autoack bool,
	n int, //指定起n个协程
	msgcb func(msg Message), //指定起n个协程
	errcb func(topic string, err error) bool,
) error {
	consumer, err := instance.client.NewConsumer(subname, topic)
	if err != nil {
		return err
	}
	for i := 0; i < n; i++ {
		go func() {
			quit := false
			for !quit {
				func() {
					//如果超过了tmo的时间，需要调用取消函数
					ctx, cancel := context.WithTimeout(context.Background(), tmo)
					defer cancel()

					msg, err := consumer.Recv(ctx)
					if err != nil {
						if errcb != nil && errcb(topic, err) {
							quit = true
						}
						return
					}
					msgcb(msg)
					if autoack {
						consumer.Ack(msg)
					}
				}()
			}
		}()
	}
	return nil
}
