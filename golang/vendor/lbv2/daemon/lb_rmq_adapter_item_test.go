package daemon

import "testing"

var (
	addr  string = "172.21.122.18:10600"
	topic string = "mc_iat"
	group string = "group"
)

func Test_Produce(t *testing.T) {
	RmqAdapter.Init([]string{addr})
	produceReply, err := RmqAdapter.Produce(topic, "just a test")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("produce_reply:%d\n", produceReply)
}
func Test_Consume(t *testing.T) {
	RmqAdapter.Init([]string{addr})
	produceReply, err := RmqAdapter.Consume(topic, group)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("produce_reply:%v\n,body:%v\n", produceReply, string(produceReply.GetBody()))
}
