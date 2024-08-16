package mqutils

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/apache/pulsar-client-go/pulsar/log"
	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"github.com/sirupsen/logrus"
)

var (
	defaultLogger *logrus.Logger
)

func init() {
	os.MkdirAll("/log/server", os.ModePerm) //Unix权限位
	defaultLogger = logrus.New()
	rl, _ := rotatelogs.New( // 默认 rotate 时间1天，最长暴露时间 7 day
		"/log/server/pulsar.log.%Y%m%d%H%M",
		rotatelogs.WithLinkName("/log/server/pulsar"),
		rotatelogs.WithRotationTime(time.Hour*24),
		rotatelogs.WithMaxAge(time.Hour*24*7),
	)
	defaultLogger.SetOutput(rl)
}

type PulsarProducer struct {
	producer pulsar.Producer
}

//发送消息
func (pp *PulsarProducer) Send(ctx context.Context, payload []byte, opt *ProducerOption) (interface{}, error) {
	msg := &pulsar.ProducerMessage{
		Payload: payload,
	}

	if opt != nil {
		msg.Key = opt.Key
		msg.ReplicationClusters = opt.ReplicationClusters
		msg.DisableReplication = opt.DisableReplication
	}

	return pp.producer.Send(ctx, msg)
}

//关闭生产者
func (pp *PulsarProducer) Close() {
	pp.producer.Close()
}

type PulsarConsumer struct {
	//这是其中一个属性
	consumer pulsar.Consumer
}

// TODO
func (pc *PulsarConsumer) BatchsRecv(ctx context.Context, minNum int) (msgs []Message, err error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), time.Millisecond*(200))
		defer cancel()
	}
	msgs = make([]Message, 0, minNum)
	for i := 0; i < minNum; i++ {
		select {
		case <-ctx.Done():
			return msgs, nil
		default:
			msg, err := pc.consumer.Receive(ctx)
			if err != nil && err != context.DeadlineExceeded {
				return msgs, err
			}

			if err==nil {
				msgs = append(msgs, msg)
			}
		}
	}
	return msgs, nil
}

func (pc *PulsarConsumer) Recv(ctx context.Context) (Message, error) {
	return pc.consumer.Receive(ctx)
}

func (pc *PulsarConsumer) Ack(msg Message) error {
	pc.consumer.Ack(msg.(pulsar.Message))
	return nil
}

func (pc *PulsarConsumer) Nack(msg Message) error {
	pc.consumer.Nack(msg.(pulsar.Message))
	return nil
}

func (pc *PulsarConsumer) Close() {
	pc.consumer.Close()
}

type PulsarClient struct {
	url    string
	client pulsar.Client
}

func (pc *PulsarClient) NewProducer(topic string) (Producer, error) {
	if !strings.HasPrefix(topic, "persistent") && !strings.HasPrefix(topic, "inpersistent") {
		panic(ErrTopicFormat)
	}

	producer, err := pc.client.CreateProducer(pulsar.ProducerOptions{
		Topic:                   topic,
		SendTimeout:             5 * time.Second,
		DisableBlockIfQueueFull: true,
		MaxPendingMessages:      10000,
		CompressionType:         pulsar.LZ4,
		CompressionLevel:        pulsar.Default,
	})
	if err != nil {
		return nil, err
	}

	return &PulsarProducer{
		producer: producer,
	}, err
}

func (pc *PulsarClient) NewConsumer(subname, topic string) (Consumer, error) {
	if !strings.HasPrefix(topic, "persistent") && !strings.HasPrefix(topic, "inpersistent") {
		panic(ErrTopicFormat)
	}

	consumer, err := pc.client.Subscribe(pulsar.ConsumerOptions{
		Topic:                      topic,
		SubscriptionName:           "sub-" + subname,
		Type:                       pulsar.Shared, // 共享模式
		ReplicateSubscriptionState: false,         // 跨集群不同步消费状态
	})
	if err != nil {
		return nil, err
	}
	return &PulsarConsumer{consumer}, nil
}

func NewPulsarClient(mqurl string) (Client, error) {
	c := &PulsarClient{
		url: mqurl,
	}

	if client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:                     mqurl,
		MaxConnectionsPerBroker: 3,
		Logger:                  log.NewLoggerWithLogrus(defaultLogger),
	}); err != nil {
		return nil, err
	} else {
		c.client = client
		return c, nil
	}
}

func (pc *PulsarClient) Close() {
	pc.client.Close()
}


func panicOnErr(err error) {
	if err != nil {
		//panic会立即中断当前函数流程
		panic(err)
	}
}
