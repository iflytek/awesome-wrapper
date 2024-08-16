package core

// 2.0版本协议默认不记录
var (
	ProtocolVersion       = "2.0"
	ProtocolPulsarVersion = "3.0"
)

const (
	TypeRPC = "rpc"
	TypeMQ  = "mq"
)

type Funcs struct {
	Func  string `json:"func"`
	Time  int64  `json:"time"`
	Count int    `json:"cnt"`
}

type Subs struct {
	Sub   string  `json:"sub"`
	Funcs []Funcs `json:"funcs"`
}

type Msg struct {
	Appid string `json:"appid"`
	Ver   string `json:"ver"`
	Uid   string `json:"uid"`
	Subs  []Subs `json:"subs"`
}
