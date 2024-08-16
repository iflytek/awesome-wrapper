package mqutils

import (
	"context"
	"errors"
	"git.iflytek.com/AIaaS/xsf/utils"
	"strings"
	"sync"
	"time"
)

type Message interface {
	Payload() []byte
	Topic() string
}

type Producer interface {
	Send(context.Context, []byte) error
	Close()
}

type Consumer interface {
	Recv(context.Context) (Message, error)
	Ack(Message) error
	Nack(Message) error
	Close()
}

type Client interface {
	NewProducer(string) (Producer, error)
	NewNameProducer(string, string) (Producer, error)
	NewConsumer(string) (Consumer, error)
	Close()
}

type MQInstance struct {
	mqurl     string
	client    Client
	producers sync.Map
	consumers sync.Map
}

func NewMQInstance(mqurl string, logImpl *utils.Logger) (*MQInstance, error) {
	instance := &MQInstance{
		mqurl: mqurl,
	}

	if strings.HasPrefix(mqurl, "amqp") {
		if c, err := NewRabbitClient(mqurl); err != nil {
			return nil, err
		} else {
			instance.client = c
			return instance, nil
		}
	}

	if strings.HasPrefix(mqurl, "pulsar") {
		if c, err := NewPulsarClient(mqurl, logImpl); err != nil {
			return nil, err
		} else {
			instance.client = c
			return instance, nil
		}
	}

	return nil, errors.New(ERR_INVALID_URL)
}

func (instance *MQInstance) DestroyInstance() {
	instance.client.Close()
}

// 手动创建生产者
func (instance *MQInstance) CreateProducer(name, topic string) (Producer, error) {
	return instance.client.NewNameProducer(name, topic)
}

// 发布消息，如果 topic 对应的生产者不存在，会自动创建
func (instance *MQInstance) SendMsg(topic string, payload []byte) error {
	if _, ok := instance.producers.Load(topic); !ok {
		if producer, err := instance.client.NewProducer(topic); err == nil {
			instance.producers.Store(topic, producer)
		} else {
			return err
		}
	}

	val, _ := instance.producers.Load(topic)
	sender := val.(Producer)

	return sender.Send(context.Background(), payload)
}

func (instance *MQInstance) Ack(msg Message) error {
	topic := msg.Topic()
	if consumer, ok := instance.consumers.Load(topic); !ok {
		return errors.New(ERR_NO_SUCH_CONSUMER)
	} else {
		return consumer.(Consumer).Ack(msg)
	}
}

func (instance *MQInstance) Nack(msg Message) error {
	topic := msg.Topic()
	if consumer, ok := instance.consumers.Load(topic); !ok {
		return errors.New(ERR_NO_SUCH_CONSUMER)
	} else {
		return consumer.(Consumer).Nack(msg)
	}
}

// 消费消息，如果 topic 对应的消费者不存在，会自动创建
func (instance *MQInstance) SetMsgProcesser(
	topic string,
	tmo time.Duration,
	autoack bool,
	msgcb func(msg Message),
	errcb func(topic string, err error) bool,
) <-chan struct{} {
	ch := make(chan struct{})
	consumer, err := instance.client.NewConsumer(topic)
	if err != nil {
		close(ch)
		return ch
	}
	instance.consumers.Store(topic, consumer)

	go func() {
		for {
			ctx, _ := context.WithTimeout(context.Background(), tmo)
			if msg, err := consumer.Recv(ctx); err != nil {
				if errcb != nil && errcb(topic, err) {
					break
				} else {
					continue
				}
			} else {
				msgcb(msg)
				if autoack {
					consumer.Ack(msg)
				}
			}
		}
		close(ch)
	}()

	return ch
}
