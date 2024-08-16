package mqutils

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

type RabbitProducer struct {
	sync.Mutex
	queue   string
	channel *amqp.Channel
}

func (rp *RabbitProducer) init(conn *amqp.Connection) error {
	rp.Lock()
	defer rp.Unlock()

	ch, err := newChannel(rp.queue, conn)
	if err != nil {
		return err
	}

	rp.channel = ch
	return nil
}

func (rp *RabbitProducer) Send(ctx context.Context, payload []byte) error {
	ch := func() *amqp.Channel {
		rp.Lock()
		defer rp.Unlock()
		return rp.channel
	}()

	return ch.Publish(
		"",       // exchange
		rp.queue, // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			ContentType:  "text/plain",
			Body:         payload,
			DeliveryMode: 2,
		})
}

func (rp *RabbitProducer) Close() {
	rp.channel.Close()
}

type RabbitMessage struct {
	topic    string
	delivery amqp.Delivery
}

func (rm *RabbitMessage) Payload() []byte {
	return rm.delivery.Body
}

func (rm *RabbitMessage) Topic() string {
	return rm.topic
}

type RabbitConsumer struct {
	sync.Mutex
	queue        string
	channel      *amqp.Channel
	deliveryChan <-chan amqp.Delivery
}

func (rc *RabbitConsumer) init(conn *amqp.Connection) error {
	rc.Lock()
	defer rc.Unlock()

	if ch, err := newChannel(rc.queue, conn); err != nil {
		return err
	} else {
		rc.channel = ch
		rc.channel.Qos(10, 0, true)
		rc.deliveryChan, err = ch.Consume(
			rc.queue,        // queue
			"sub-"+rc.queue, // consumer
			false,           // auto-ack
			false,           // exclusive
			false,           // no-local
			false,           // no-wait
			nil,             // args
		)
		return err
	}
}

func (rc *RabbitConsumer) Recv(ctx context.Context) (Message, error) {
	deliveryChan := func() <-chan amqp.Delivery {
		rc.Lock()
		defer rc.Unlock()
		return rc.deliveryChan
	}()

	select {
	case msg, ok := <-deliveryChan:
		if ok {
			return &RabbitMessage{
				topic:    rc.queue,
				delivery: msg,
			}, nil
		} else {
			return nil, errors.New(ERR_CONNECT_BROKEN)
		}

	case <-ctx.Done():
		return nil, errors.New(ERR_DEADLINE)
	}
}

func (rc *RabbitConsumer) Ack(msg Message) error {
	delivery := msg.(*RabbitMessage).delivery
	return rc.channel.Ack(delivery.DeliveryTag, false)
}

func (rc *RabbitConsumer) Nack(msg Message) error {
	delivery := msg.(*RabbitMessage).delivery
	return rc.channel.Nack(delivery.DeliveryTag, false, true)
}

func (rc *RabbitConsumer) Close() {
	rc.channel.Close()
}

type RabbitClient struct {
	url        string
	cancelFunc context.CancelFunc
	conn       *amqp.Connection
	producers  map[*RabbitProducer]struct{}
	consumers  map[*RabbitConsumer]struct{}
}

func (rc *RabbitClient) NewProducer(topic string) (Producer, error) {
	producer := &RabbitProducer{
		queue: topic,
	}

	if err := producer.init(rc.conn); err != nil {
		return nil, err
	}

	rc.producers[producer] = struct{}{}

	return producer, nil
}

func (rc *RabbitClient) NewNameProducer(name, topic string) (Producer, error) {
	return rc.NewProducer(topic)
}

func (rc *RabbitClient) NewConsumer(topic string) (Consumer, error) {
	consumer := &RabbitConsumer{
		queue: topic,
	}

	if err := consumer.init(rc.conn); err != nil {
		return nil, err
	}

	rc.consumers[consumer] = struct{}{}

	return consumer, nil
}

func NewRabbitClient(mqurl string) (Client, error) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	c := &RabbitClient{
		url:        mqurl,
		cancelFunc: cancelFunc,
		producers:  make(map[*RabbitProducer]struct{}),
		consumers:  make(map[*RabbitConsumer]struct{}),
	}

	var err error
	if c.conn, err = amqp.Dial(mqurl); err != nil {
		return nil, err
	}

	// 断开重连
	go func(ctx context.Context) {
		for {
			amqpErr := make(chan *amqp.Error)
			c.conn.NotifyClose(amqpErr)

			select {
			case <-ctx.Done():
				return
			default:
				for {
					if c.conn, err = amqp.Dial(mqurl); err != nil {
						time.Sleep(time.Second)
					} else {
						for producer := range c.producers {
							producer.init(c.conn)
						}

						for consumer := range c.consumers {
							consumer.init(c.conn)
						}

						break
					}
				}
			}
		}
	}(ctx)

	return c, nil
}

func (rc *RabbitClient) Close() {
	rc.cancelFunc()
	rc.conn.Close()
}

func newChannel(topic string, conn *amqp.Connection) (*amqp.Channel, error) {
	if conn == nil || conn.IsClosed() {
		return nil, errors.New(ERR_CONNECT_BROKEN)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	if err := initChannel(ch, topic); err != nil {
		return nil, err
	}

	return ch, err
}

func initChannel(ch *amqp.Channel, queue string) (err error) {
	if _, err = ch.QueueDeclare(
		queue, // name
		true,  // durable,设置是否持久化
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	); err != nil {
		return
	}

	if err = ch.ExchangeDeclare(
		"offlineExchange", // name
		"direct",          // type
		true,              // durable
		false,             // auto-deleted
		false,             // internal
		false,             // no-wait
		nil,               // arguments
	); err != nil {
		return
	}

	if err = ch.QueueBind(
		queue,
		"ost",
		"offlineExchange",
		false,
		nil,
	); err != nil {
		return
	}

	return
}
