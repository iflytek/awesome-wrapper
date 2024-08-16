package main

import (
	"context"
	"time"
)

func getCtx(timeout time.Duration) context.Context {
	logger.Printf("globalTimeout:%v\n", timeout)
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(timeout))
	return ctx
}
