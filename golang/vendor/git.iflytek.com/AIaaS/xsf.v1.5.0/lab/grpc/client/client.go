package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/cihub/seelog"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	tm       = flag.Int64("tm", 1000, "timeout")
	gInfo    = flag.Int64("gInfo", 0, "gInfo")
	qpsStart = flag.Int64("qpsStart", 0, "qpsStart")
	gNum     = flag.Int64("goroutines", 1, "gogroutine number")
	gClose   = flag.Int64("gClose", 0, "grpc close")
	gIdle    = flag.Int64("gIdle", 0, "gIdle")
	gCnt     = flag.Int64("count", 1, "total request")
	host     = flag.String("h", "127.0.0.1", "host")
	port     = flag.String("p", "50051", "port")
	rBuf     = flag.Int("rbuf", 1024, "unit:kb")
	wBuf     = flag.Int("wbuf", 1024, "unit:kb")
	pre      = flag.Int("pre", 0, "pre")
)
var (
	stopFunc func()
)

func init() {

	flag.Parse()
	if logger, err := seelog.LoggerFromConfigAsString(`<seelog type="sync">
    <outputs formatid="main">
        <filter levels="trace,debug,info,warn,error,critical">
            <file path="log/grpcCli.log"/>
        </filter>
    </outputs>
    <formats>
        <format id="main" format="%Msg"/>
    </formats>
</seelog>`); err != nil {
		log.Fatal(err)
	} else {
		seelog.ReplaceLogger(logger)
	}

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGPIPE)
		s := <-c
		switch s {
		case syscall.SIGINT, syscall.SIGKILL:
			{
				fmt.Printf("the program had received %v signal, will exit immediately -_-|||\n", s.String())
				stopFunc()
			}
		case syscall.SIGPIPE:
			{
				fmt.Printf("get broken pipe")
			}
		}
	}()

}
func main() {

	flag.Parse()

	{
		ctx, cancel := context.WithCancel(context.Background())
		stopFunc = cancel
		performanceGrpc(ctx, time.Millisecond*time.Duration(*tm))
	}
}
