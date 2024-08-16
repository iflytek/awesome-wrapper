package main

import (
	"context"
	"git.iflytek.com/AIaaS/xsf/utils"
	"os"
	"strings"
)

func healthCheck(ctx context.Context, addr string) {
	query, header := getParams()
	cmdServerResp, cmdServerRespErr := getRpcClient(ctx, addr).Cmdserver(
		ctx,
		&utils.Request{
			Query:   string(query),
			Headers: string(header),
			Body:    "",
		},
	)
	checkErr(cmdServerRespErr)
	logger.Printf("states of calling cmdServer,cmdServerResp.Body:%v,cmdServerRespErr:%v\n",
		cmdServerResp.Body, cmdServerRespErr)

	if strings.Contains(strings.ToLower(string(cmdServerResp.Body)), "err") {
		logger.Printf("cmdServerResp.Body contain err\n")
		logger.Println("failure")
		os.Exit(1)
	}
	logger.Println("health check successfully")
}
