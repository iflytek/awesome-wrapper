package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	pb "git.iflytek.com/AIaaS/xsf/lab/grpc/common"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	port = flag.String("p", "50051", "port")
	rBuf = flag.Int("rbuf", 1024, "unit:kb")
	wBuf = flag.Int("wbuf", 1024, "unit:kb")
)

type server struct{}

func (s *server) SimpleCall(ctx context.Context, in *pb.ReqData) (*pb.ResData, error) {
	return &pb.ResData{Code: 0, ErrorInfo: "ok", Param: map[string]string{
		"intro": "received data",
		"op":    "req",
		"ip":    "127.0.0.1",
		"port":  "port",
	}}, nil
}

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", *port))
	if err != nil {
		log.Panic("failed to listen:", err)
	}

	s := grpc.NewServer(grpc.ReadBufferSize((*rBuf)*1024), grpc.WriteBufferSize((*wBuf)*1024))
	pb.RegisterGrpcCallServer(s, &server{})
	reflection.Register(s)

	fmt.Printf("about to serve:%v\n", lis.Addr().String())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
