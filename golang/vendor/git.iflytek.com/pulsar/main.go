package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	mq "mq/mqutils"
	"time"
)

func panic_on_error(err error) {
	if err != nil {
		panic(err)
	}
}

var mqurl = "pulsar://10.1.87.69:6650"
var topic = "persistent://public/default/geo"

// var mqurl = "amqp://guest:guest@10.1.87.70:5672/"
// var topic = "geo"

func TestBasic() {
	inst, err := mq.NewMQInstance(mqurl, "/log/server/pulsar-test.log")
	panic_on_error(err)

	inst.SetMsgProcesser(topic, time.Second*5, true,
		func(msg mq.Message) {
			fmt.Println("get message:", string(msg.Payload()))
			// hashcheck(msg.Payload())
			// inst.Ack(msg)
		},
		func(topic string, err error) bool {
			// fmt.Println("get message error:", err)
			return false
		},
	)

	// make a message of 20Mb
	// payload := make([]byte, 1024*1024*20)
	// hashcheck(payload)

	for i := 0; i < 3; i++ {
		fmt.Println("sender:", inst.SendMsg(topic, []byte("hello world")))
		time.Sleep(time.Second * 1)
	}

	fmt.Println("quit")
	inst.DestroyInstance()

	for i := 0; i < 3; i++ {
		fmt.Println("sender:", inst.SendMsg(topic, []byte("hello world")))
		time.Sleep(time.Second * 1)
	}
	fmt.Println("quit")

	time.Sleep(time.Second * 10)
}

func hashcheck(buf []byte) {
	hash := md5.New()
	hash.Write(buf)
	fmt.Println(hex.EncodeToString(hash.Sum(nil)))
}

// 手动创建生产者
func TestManualProducer() {
	topic := "persistent://public/default/123"
	inst, err := mq.NewMQInstance(mqurl, "/log/server/pulsar-test.log")
	panic_on_error(err)
	defer inst.DestroyInstance()

	producer1, err := inst.CreateProducer("p1", topic)
	panic_on_error(err)
	defer producer1.Close()

	producer2, err := inst.CreateProducer("p2", topic)
	panic_on_error(err)
	defer producer2.Close()

	producer1.Send(context.Background(), []byte("hello from p1"))
	producer2.Send(context.Background(), []byte("hello from p2"))

	inst.SetMsgProcesser(topic, time.Second*5, true,
		func(msg mq.Message) {
			fmt.Println("get message:", string(msg.Payload()))
		},
		func(topic string, err error) bool {
			return false
		},
	)
	time.Sleep(time.Second*3)
}

func main() {
	TestManualProducer()
	return
	TestBasic()
}
