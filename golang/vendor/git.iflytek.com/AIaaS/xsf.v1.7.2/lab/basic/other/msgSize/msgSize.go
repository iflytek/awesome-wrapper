package main

import (
	"fmt"
	"git.iflytek.com/AIaaS/xsf/utils"
	"github.com/golang/protobuf/proto"
)

func main() {
	res := &utils.ResData{
		S: &utils.Session{Sess: make(map[string]string)},
		Param: map[string]string{
			"intro": "received data",
			"op":    "req",
			"ip":    "127.0.0.1",
			"port":  "8080",
		},
		Data: []*utils.DataMeta{{Data: []byte{}, Desc: make(map[string]string)}},
	}
	r, e := proto.Marshal(res)
	if e != nil {
		panic(e)
	}
	fmt.Println(len(r))
}
