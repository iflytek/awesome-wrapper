package mqutils

import (
"context"
"git.iflytek.com/AIaaS/xsf/utils"
"strconv"
"strings"
"math/rand"
"time"
"github.com/apache/pulsar-client-go/pulsar"
)

type PulsarProducer struct {
	producer pulsar.Producer
}

func (pp *PulsarProducer) Send(ctx context.Context, payload []byte) error {
	_, err := pp.producer.Send(ctx, &pulsar.ProducerMessage{
		Payload:            payload,
		DisableReplication: false,
	})

	return err
}

func (pp *PulsarProducer) Close() {
	pp.producer.Close()
}

type PulsarConsumer struct {
	consumer pulsar.Consumer
}

func (pc *PulsarConsumer) Recv(ctx context.Context) (Message, error) {
	msg, err := pc.consumer.Receive(ctx)
	if err != nil {
		return nil, err
	}

	return msg, nil
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
	url             string
	client          pulsar.Client
	producerCounter int32
	consumerCounter int32
}

func (pc *PulsarClient) NewNameProducer(name, topic string) (Producer, error) {
	if !strings.HasPrefix(topic, "persistent") && !strings.HasPrefix(topic, "inpersistent") {
		panic(ERR_TOPIC_FORMAT)
	}

	producer, err := pc.client.CreateProducer(pulsar.ProducerOptions{
		Topic:              topic,
		Name:               "prod-manual-" + name,
		SendTimeout:        5 * time.Second,
		MaxPendingMessages: 10000,
		CompressionType:    pulsar.LZ4,
		CompressionLevel:   pulsar.Default,
	})
	if err != nil {
		return nil, err
	}

	return &PulsarProducer{
		producer: producer,
	}, err
}

func (pc *PulsarClient) NewProducer(topic string) (Producer, error) {
	if !strings.HasPrefix(topic, "persistent") && !strings.HasPrefix(topic, "inpersistent") {
		panic(ERR_TOPIC_FORMAT)
	}

	producer, err := pc.client.CreateProducer(pulsar.ProducerOptions{
		Topic:              topic,
		Name:               "prod-named-" + getRawTopicName(topic) + "-" + strconv.FormatInt(time.Now().Unix(), 10) + generateRandomString(10),
		SendTimeout:        5 * time.Second,
		MaxPendingMessages: 10000,
		CompressionType:    pulsar.LZ4,
		CompressionLevel:   pulsar.Default,
	})
	if err != nil {
		return nil, err
	}

	return &PulsarProducer{
		producer: producer,
	}, err
}

func (pc *PulsarClient) NewConsumer(topic string) (Consumer, error) {
	if !strings.HasPrefix(topic, "persistent") && !strings.HasPrefix(topic, "inpersistent") {
		panic(ERR_TOPIC_FORMAT)
	}

	consumer, err := pc.client.Subscribe(pulsar.ConsumerOptions{
		Topic:                      topic,
		SubscriptionName:           "sub-named-" + getRawTopicName(topic),
		Type:                       pulsar.Shared, // 共享模式
		ReplicateSubscriptionState: false,         // 跨集群不同步消费状态
	})
	if err != nil {
		return nil, err
	}

	return &PulsarConsumer{consumer}, nil
}

func NewPulsarClient(mqurl string, logger *utils.Logger) (Client, error) {
	c := &PulsarClient{
		url: mqurl,
	}

	if client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:                     mqurl,
		MaxConnectionsPerBroker: 10,
		Logger:                  &LogWrapper{log: logger},
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

// "persistent://public/default/geo" -> geo
func getRawTopicName(topic string) string {
	arr := strings.Split(topic, "/")
	return arr[len(arr)-1]
}


const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateRandomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
