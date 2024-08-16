package sender

type MsgSender interface {
	Init(chan *[]byte) error
	Run()
	Fini()
}
