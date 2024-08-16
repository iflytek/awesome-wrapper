package main

import (
	"os"
	"runtime"
)

func checkErr(err error) {
	fnName := func() string {
		pc, _, _, _ := runtime.Caller(2)
		return runtime.FuncForPC(pc).Name()
	}()
	if err != nil {
		logger.Printf("fn:%v,err:%v\n", fnName, err)
		logger.Println("failure")
		os.Exit(1)
	}
}