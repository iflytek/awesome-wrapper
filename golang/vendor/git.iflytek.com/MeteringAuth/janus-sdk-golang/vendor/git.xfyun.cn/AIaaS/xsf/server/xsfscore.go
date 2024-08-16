package xsf

import (
	"context"
	"fmt"
	"git.xfyun.cn/AIaaS/xsf/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"math"
	"net"
	"sync"
	"time"
)

type grpcOpt struct {
	maxConcurrentStreams  uint32
	maxReceiveMessageSize int
	maxSendMessageSize    int
	initialWindowSize     int32
	initialConnWindowSize int32
	writeBufferSize       int
	readBufferSize        int
}

func (g *grpcOpt) setMaxConcurrentStreams(in uint32) {
	g.maxConcurrentStreams = in
}
func (g *grpcOpt) setMaxReceiveMessageSize(in int) {
	g.maxReceiveMessageSize = in
}
func (g *grpcOpt) setMaxSendMessageSize(in int) {
	g.maxSendMessageSize = in
}
func (g *grpcOpt) setInitialWindowSize(in int32) {
	g.initialWindowSize = in
}
func (g *grpcOpt) setInitialConnWindowSize(in int32) {
	g.initialConnWindowSize = in
}
func (g *grpcOpt) setWriteBufferSize(in int) {
	g.writeBufferSize = in
}
func (g *grpcOpt) setReadBufferSize(in int) {
	g.readBufferSize = in
}

const maxConcurrentStreams uint32 = 0
const maxReceiveMessageSize = 1024 * 1024 * 4
const maxSendMessageSize = math.MaxInt32
const initialWindowSize int32 = 0
const initialConnWindowSize int32 = 0
const writeBufferSize = 1024 * 1024 * 2
const readBufferSize = 0
const grpcSleepTime = time.Second * 3

var grpcOptInst = &grpcOpt{
	maxConcurrentStreams:  maxConcurrentStreams,
	maxReceiveMessageSize: maxReceiveMessageSize,
	maxSendMessageSize:    maxSendMessageSize,
	initialWindowSize:     initialWindowSize,
	initialConnWindowSize: initialConnWindowSize,
	writeBufferSize:       writeBufferSize,
	readBufferSize:        readBufferSize,
}

type xsfServer struct {
	grpcserver *grpc.Server
}

func (x *xsfServer) Closeout() {
	fmt.Println("about to stop grpcserver")
	x.grpcserver.GracefulStop()
}
func (x *xsfServer) run(bc BootConfig, listener net.Listener, srv utils.XsfCallServer) (resErr error) {
	var opts []grpc.ServerOption
	//----------------------------------------------------
	// 注册interceptor
	var interceptor grpc.UnaryServerInterceptor
	fmt.Printf("about init interceptor\n")
	interceptor = func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		err = func(ctx context.Context) error {
			return nil
		}(ctx)
		if err != nil {
			return
		}
		// 继续处理请求
		return handler(ctx, req)
	}
	fmt.Printf("success init interceptor\n")

	opts = append(opts, grpc.ConnectionTimeout(time.Duration(GRPCTIMEOUT)*time.Second))
	opts = append(opts, grpc.UnaryInterceptor(interceptor))
	//----------------------------------------------------
	opts = append(opts, grpc.MaxConcurrentStreams(grpcOptInst.maxConcurrentStreams))
	opts = append(opts, grpc.MaxRecvMsgSize(grpcOptInst.maxReceiveMessageSize))
	opts = append(opts, grpc.MaxSendMsgSize(grpcOptInst.maxSendMessageSize))
	opts = append(opts, grpc.InitialWindowSize(grpcOptInst.initialWindowSize))
	opts = append(opts, grpc.InitialConnWindowSize(grpcOptInst.initialConnWindowSize))
	opts = append(opts, grpc.WriteBufferSize(grpcOptInst.writeBufferSize))
	opts = append(opts, grpc.ReadBufferSize(grpcOptInst.readBufferSize))
	//----------------------------------------------------
	fmt.Printf("about to call grpc.NewServer(opts...),maxRecv:%v,maxSend:%v\n", grpcOptInst.maxReceiveMessageSize, grpcOptInst.maxSendMessageSize)
	x.grpcserver = grpc.NewServer(opts...)
	addKillerCheck(killerNormalPriority, "grpcserver", x)

	fmt.Printf("about to call utils.RegisterXsfCallServer(x.grpcserver, srv)\n")

	utils.RegisterXsfCallServer(x.grpcserver, srv)
	utils.RegisterToolBoxServer(x.grpcserver, &ToolBoxServer{})
	fmt.Printf("about to call reflection.Register(x.grpcserver)\n")

	reflection.Register(x.grpcserver)

	fmt.Println("about to exec userCallback")
	dealUserCallBack()

	//----------------------------------------------------
	fmt.Println("about to call x.grpcserver.Serve")
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := x.grpcserver.Serve(listener); err != nil {
			resErr = err
		}
	}()

	{
		ctxTm, cancelTm := context.WithTimeout(context.Background(), grpcSleepTime)
		defer cancelTm()

		fmt.Printf("about to check if the grpc service(%v) is started\n", listener.Addr().String())
	retryEnd:
		for {
			select {
			case <-ctxTm.Done():
				{
					log.Panicf("deadlineExceed when check grpc server(%v)\n", listener.Addr().String())
				}
			default:
				{
					if healthCheck(listener.Addr().String()) {
						fmt.Printf("grpc server(%v) started successfully\n", listener.Addr().String())
						break retryEnd
					}
				}
			}
		}
	}

	//----------------------------------------------------

	fmt.Printf("about to call finderadapter.Register(%v)\n", listener.Addr().String())
	finderExist, finderRegisterErr := finderadapter.Register(listener.Addr().String(), bc.CfgData.ApiVersion)
	if finderRegisterErr != nil {
		log.Panicf("finderadapter.Register fail -> addr:%v,bc:%+v,finderRegistErr:%v\n", listener.Addr().String(), bc, finderRegisterErr)
	}

	if finderExist {
		fmt.Printf("finderadapter.Register success. -> addr:%v\n", listener.Addr().String())
	}

	fmt.Println("about to exec fcDelayInst")
	fcDelayInst.exec()

	waitCtx, waitCtxCancel := context.WithCancel(context.Background())

	addKillerCheck(killerLastPriority, "WaitForExit", &killerWrapper{callback: func() {
		waitCtxCancel()
	}})

	fmt.Println("blocking for grpcserver.Serve")

	wg.Wait()
	fmt.Println("success stop grpc graceful.")

	fmt.Println("Waiting for exit.")

	<-waitCtx.Done()

	fmt.Println("all mission complete.")

	return nil
}
