package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"git.iflytek.com/AIaaS/xsf/utils"
	"google.golang.org/grpc"
	"log"
	"os"
)

var (
	host = flag.String("h", "127.0.0.1", "the host to request")
	port = flag.Int("p", 1234, "the port to request")
	op   = flag.Int("o", 1, "0:call,1:cmdServer")
)

func main() {
	flag.Parse()

	switch *op {
	case 0:
		call()
	case 1:
		cmdServer()
	default:
		panic("wrong op!!!")
	}
}

func call() {
	conn, connErr := grpc.Dial(fmt.Sprintf("%v:%v", *host, *port), grpc.WithInsecure())
	if connErr != nil {
		log.Fatalf("did not connect: %v", connErr)
	}
	defer conn.Close()
	c := utils.NewXsfCallClient(conn)

	callResp, callRespErr := c.Call(context.Background(), &utils.ReqData{Op: "req", S: &utils.Session{T: "0a01cd9715135889525510081n008088", H: "its0008342b"}})

	fmt.Printf("callResp:%+v, callRespErr:%+v\n", callResp, callRespErr)
}

func cmdServer() {
	conn, connErr := grpc.Dial(fmt.Sprintf("%v:%v", *host, *port), grpc.WithInsecure())
	if connErr != nil {
		log.Fatalf("did not connect: %v", connErr)
	}
	defer conn.Close()

	c := utils.NewToolBoxClient(conn)

	//query, _ := json.Marshal(map[string]string{"cmd": "metrics"})
	query, _ := json.Marshal(map[string]string{"cmd": "status"})
	header, _ := json.Marshal(map[string]string{"method": "GET"})

	cmdServerResp, cmdServerRespErr := c.Cmdserver(context.Background(), &utils.Request{Query: string(query), Headers: string(header), Body: ""})

	if cmdServerResp != nil {
		toFile(cmdServerResp.Body)
	}

	fmt.Printf("cmdServerResp:%+v, cmdServerRespErr:%+v\n", cmdServerResp, cmdServerRespErr)
}
func toFile(in string) {
	f, e := os.Create("cmdClient.rst")
	if e != nil {
		panic(e)
	}
	defer f.Close()
	_, _ = f.WriteString(in)
}
