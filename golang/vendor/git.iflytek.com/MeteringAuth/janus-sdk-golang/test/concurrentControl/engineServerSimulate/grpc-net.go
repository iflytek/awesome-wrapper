package main

import (
	"concurrentControl-gen-proto"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	//"github.com/cihub/seelog"
	"context"
	"log"
	"net"
	"zaplogWrap"
)

type concurrentServiceNet struct{}

func (*concurrentServiceNet) SendMsg(ctx context.Context, msg *concurrentNet.SimpleMsg) (*concurrentNet.Reply, error) {
	//seelog.Info(msg.Appid)
	//zaplogWrap.Logger.Info(msg.Appid)
	AccumulatePoolInst.DoMockEngine(APPID)

	return &concurrentNet.Reply{Ret: true}, nil
}

func startServer() {

	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	concurrentNet.RegisterConcurrentNetaServer(grpcServer, &concurrentServiceNet{})
	if err := grpcServer.Serve(lis); err != nil {
		zaplogWrap.Logger.Error("grpc Serve listen failed", zap.Error(err))
		return
	}
}
