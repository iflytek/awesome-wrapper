package main

import (
	"context"
	"git.iflytek.com/AIaaS/xsf/utils"
	"google.golang.org/grpc"
)

func getRpcClient(ctx context.Context, addr string) utils.ToolBoxClient {
	conn, connErr := grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithBlock())
	checkErr(connErr)
	//defer conn.Close()
	return utils.NewToolBoxClient(conn)
}
