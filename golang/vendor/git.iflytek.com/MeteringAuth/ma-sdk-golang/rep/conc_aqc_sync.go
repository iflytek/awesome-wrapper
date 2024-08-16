package rep

import (
	"fmt"
	xsf "git.iflytek.com/AIaaS/xsf/client"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/ma/config"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/ma/syncproto"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/ma/tool"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/rep/monitor"
	"github.com/golang/protobuf/proto"
	"time"
)

const (
	AqcOpInc         = "aqc_inc"
	AqcOpDec         = "aqc_dec"
	AqcOp            = "aqc_count"
	AsyncOp          = "async_count"
	AqcCrossSyncOp   = "aqc_conc_batch_sync"
	AsyncCrossSyncOp = "async_conc_sync"
	Aqc_Report       = "aqc_report"
)

type counterMessage struct {
	aQcSyncCounterKey
	delta int // 1 or -1
}

type counterKey struct {
	appId    string
	channel  string
	function string
}

type aQcSyncCounterKey struct {
	counterKey
	addr string
}

type AqcCounterSyncManager struct {
	xsfc *xsf.Client
	to   time.Duration
	addr string

	data chan counterMessage

	//msgBatch map[aQcSyncCounterKey]int
	//batchCount int
	workerCount int

	batchMaxSize int // 一次批量发送数据上限
}

// 用于将+1，-1 操作批量打包后，同步到lks
func NewAqcCounterSyncManager(batchMaxSize int, bufferSize int, xsfClient *xsf.Client, timeout int64, addr string, workerCount int) *AqcCounterSyncManager {
	if batchMaxSize <= 0 {
		batchMaxSize = 1000
	}
	if bufferSize <= 0 {
		bufferSize = 1e4
	}
	ac := &AqcCounterSyncManager{
		xsfc: xsfClient,
		to:   time.Duration(timeout) * time.Millisecond,
		addr: addr,
		data: make(chan counterMessage, bufferSize),
		//msgBatch:     map[aQcSyncCounterKey]int{},
		batchMaxSize: batchMaxSize,
	}
	for i := 0; i < workerCount; i++ {
		go ac.run()
	}
	return ac
}

func (a *AqcCounterSyncManager) Add(key aQcSyncCounterKey, delta int) {
	monitor.WithCommonCounter(AqcOp, monitor.ReportTotal)
	select {
	case a.data <- counterMessage{aQcSyncCounterKey: key, delta: delta}:
	default:
		monitor.WithCommonCounter(AqcOp, monitor.ReportDrop)
		tool.L.Warnf("rep-sdk | AqcCounterSyncManager | Add | channel is full")
	}
}

func (a *AqcCounterSyncManager) serializeBatchMessage(msgBatch map[aQcSyncCounterKey]int) map[string]*syncproto.AqcRequest {
	if a.xsfc == nil {
		return nil
	}
	classifer := xsf.NewCaller(a.xsfc)
	aqcreq := make(map[string]*syncproto.AqcRequest)
	//封装aqcrequest数据
	for k, v := range msgBatch {
		//判断delta是否为0，为0则不发送
		if v == 0 {
			tool.L.Debugw("rep-sdk | AqcCounterSyncManager | serializeBatchMessage skip:", "appid:", k.appId, "channel:", k.channel, "function:", k.function, "addr:", k.addr, "delta:", v)
			continue
		}
		//获取hash地址
		addr, err := classifer.GetHashAddr(k.appId, config.SVC)
		if err != nil {
			tool.L.Errorw("rep-sdk | AqcCounterSyncManager | serializeBatchMessage", "error", err, "key", k, "svc", config.SVC)
			continue
		}
		//创建req
		req, ok := aqcreq[addr]
		if !ok || req == nil {
			req = &syncproto.AqcRequest{}
			aqcreq[addr] = req
		}
		//设置数据
		req.Data = append(req.Data, &syncproto.AqcMetadata{
			Tuple: &syncproto.MetaTuple{
				AppId:    k.appId,
				Channel:  k.channel,
				Function: k.function,
			},
			Addr:  k.addr,
			Delta: int32(v),
		})
		tool.L.Debugw("rep-sdk | AqcCounterSyncManager | serializeBatchMessage", "addr", addr)
	}
	return aqcreq
}

func (a *AqcCounterSyncManager) syncToLks(msgBatch map[aQcSyncCounterKey]int) {
	msg := a.serializeBatchMessage(msgBatch)
	a.sendToLksByXsf(msg, AqcOp)
	a.reset()
}

func (a *AqcCounterSyncManager) sync(protodate []byte) {
	//解析
	reqData := &syncproto.AqcRequest{}
	err := proto.Unmarshal(protodate, reqData)
	if err != nil {
		tool.L.Errorw("rep-sdk | AqcCounterSyncManager | sync", "error", err)
		return
	}
	if reqData.GetData() == nil || len(reqData.GetData()) == 0 {
		tool.L.Errorw("rep-sdk | AqcCounterSyncManager | sync", "error", "data is nil or empty")
		return
	}
	msgBatch := make(map[aQcSyncCounterKey]int)
	for _, v := range reqData.GetData() {
		if v == nil || v.Tuple == nil {
			continue
		}
		key := aQcSyncCounterKey{
			counterKey: counterKey{
				appId:    v.Tuple.AppId,
				channel:  v.Tuple.Channel,
				function: v.Tuple.Function,
			},
			addr: v.Addr,
		}
		msgBatch[key] += int(v.Delta)
	}
	//封装
	msg := a.serializeBatchMessage(msgBatch)
	//发送
	a.sendToLksByXsf(msg, AqcCrossSyncOp)
}

// todo: timeout
//func (a *AqcCounterSyncManager) sendToLksByXsf(reqs map[string]*syncproto.AqcRequest, op string) {
//	m := map[string]string{"op": op}
//	for _, req := range reqs {
//		if req == nil || len(req.Data) == 0 || req.Data[0].Tuple == nil {
//			continue
//		}
//		//创建xsf request
//		xsfreq := xsf.NewReq()
//		//插入数据
//		databyte, err := proto.Marshal(req)
//		if err != nil {
//			tool.L.Errorw("rep-sdk | AqcCounterSyncManager | sendToLksByXsf", "error", err, "req", req)
//			continue
//		}
//		xsfreq.Append(databyte, m)
//		xsfreq.SetOp(op)
//		//发送数据
//		caller := xsf.NewCaller(a.xsfc)
//		caller.WithHashKey(req.Data[0].Tuple.AppId)
//		_, errcode, err := caller.Call(config.SVC, op, xsfreq, a.to)
//		if err != nil {
//			tool.L.Errorw("rep-sdk | AqcCounterSyncManager | sendToLksByXsf", "code", errcode, "error", err, "req", xsfreq)
//		}
//	}
//}

func (a *AqcCounterSyncManager) sendToLksByXsf(reqs map[string]*syncproto.AqcRequest, op string) {
	m := map[string]string{"op": op}
	for _, req := range reqs {
		monitor.WithCommonCounter(AqcOp, monitor.SendTotal)
		st := time.Now()
		if req == nil || len(req.Data) == 0 || req.Data[0].Tuple == nil {
			tool.L.Infof("rep-sdk | AqcCounterSyncManager | sendToLksByXsf | req is nil or empty, req: %v", req)
			continue
		}

		// 创建xsf request
		xsfreq := xsf.NewReq()
		// 插入数据
		databyte, err := proto.Marshal(req)
		if err != nil {
			tool.L.Errorw("rep-sdk | AqcCounterSyncManager | sendToLksByXsf", "error", err, "req", req)
			continue
		}
		xsfreq.Append(databyte, m)
		xsfreq.SetOp(op)
		// 发送数据
		caller := xsf.NewCaller(a.xsfc)
		caller.WithHashKey(req.Data[0].Tuple.AppId)

		// 使用select结合time.After实现超时控制
		done := make(chan struct{})
		errCh := make(chan error)
		go func() {
			_, errcode, err := caller.Call(config.SVC, op, xsfreq, a.to)
			if err != nil {
				errCh <- fmt.Errorf("code: %d, error: %v, req: %v", errcode, err, xsfreq)
			}
			close(done)
		}()
		select {
		case <-done:
			// 请求完成，继续下一个请求
			monitor.WithCommonCounter(AqcOp, monitor.SendSuccess)
			tool.L.Infof("rep-sdk | AqcCounterSyncManager | sendToLksByXsf | done, req: %v", xsfreq)
		case err := <-errCh:
			tool.L.Errorw("rep-sdk | AqcCounterSyncManager | sendToLksByXsf", "error", err)
		case <-time.After(a.to):
			// 超时，直接放弃退出
			tool.L.Infof("rep-sdk | AqcCounterSyncManager | sendToLksByXsf | timeout, req: %v", xsfreq)
			continue
		}
		monitor.WithCost(AqcOp, time.Since(st))
	}
}

func (a *AqcCounterSyncManager) reset() {
	//a.batchCount = 0
	//a.msgBatch = map[aQcSyncCounterKey]int{}
}

func (a *AqcCounterSyncManager) run() {
	msgBatch := map[aQcSyncCounterKey]int{}
	batchCount := 0
	for {
		msg := <-a.data
		msgBatch[msg.aQcSyncCounterKey] += msg.delta
		batchCount++

	end:
		for {
			select {
			case msg := <-a.data:
				msgBatch[msg.aQcSyncCounterKey] += msg.delta
				batchCount++
				if batchCount >= a.batchMaxSize {
					a.syncToLks(msgBatch)
					batchCount = 0
					msgBatch = map[aQcSyncCounterKey]int{}
				}
			default:
				if batchCount > 0 {
					a.syncToLks(msgBatch)
					batchCount = 0
					msgBatch = map[aQcSyncCounterKey]int{}
				}
				break end
			}
		}
	}
}

type asyncCountMessage struct {
	counterKey
	requestId string
	expire    int
	op        syncproto.AsyncOp
}

// 用于将+1，-1 操作批量打包后，同步到lks
func NewAsyncCounterSyncManager(batchMaxSize int, bufferSize int, xsfClient *xsf.Client, timeout int64, addr string) *AsyncCounterSyncManager {
	if batchMaxSize <= 0 {
		batchMaxSize = 1000
	}
	if bufferSize <= 0 {
		bufferSize = 1e4
	}
	ac := &AsyncCounterSyncManager{
		xsfc: xsfClient,
		to:   time.Duration(timeout) * time.Millisecond,
		addr: addr,
		data: make(chan asyncCountMessage, bufferSize),
		//msgBatch:     make([]asyncCountMessage, 0, 100),
		batchMaxSize: batchMaxSize,
	}
	go ac.run()
	return ac
}

// 异步并发控制数据同步管理
type AsyncCounterSyncManager struct {
	xsfc *xsf.Client
	to   time.Duration
	addr string

	//msgBatch   []asyncCountMessage
	data chan asyncCountMessage
	//batchCount int

	batchMaxSize int // 一次批量发送数据上限
}

func (a *AsyncCounterSyncManager) Add(msg asyncCountMessage) {
	monitor.WithCommonCounter(AsyncOp, monitor.ReportTotal)
	select {
	case a.data <- msg:
	default:
		monitor.WithCommonCounter(AsyncOp, monitor.ReportDrop)
		tool.L.Warnf("rep-sdk | AsyncCounterSyncManager | Add | channel is full")
	}
}

func (a *AsyncCounterSyncManager) run() {
	msgBatch := []asyncCountMessage{}
	batchCount := 0
	for {
		msg := <-a.data
		msgBatch = append(msgBatch, msg)
		batchCount++
	end:
		for {
			select {
			case msg := <-a.data:
				msgBatch = append(msgBatch, msg)
				batchCount++
				if batchCount >= a.batchMaxSize {
					a.syncToLks(msgBatch)
					batchCount = 0
					msgBatch = msgBatch[:0]
				}
			default:
				if batchCount > 0 {
					a.syncToLks(msgBatch)
					batchCount = 0
					msgBatch = msgBatch[:0]
				}
				break end
			}
		}
	}
}

func (a *AsyncCounterSyncManager) serializeBatchMessage(msgBatch []asyncCountMessage) map[string]*syncproto.AsyncRequest {
	if a.xsfc == nil {
		return nil
	}
	classifer := xsf.NewCaller(a.xsfc)
	aqcreq := make(map[string]*syncproto.AsyncRequest)
	//封装aqcrequest数据
	for _, v := range msgBatch {
		//获取hash地址
		addr, err := classifer.GetHashAddr(v.appId, config.SVC)
		if err != nil {
			tool.L.Errorw("rep-sdk | AqcCounterSyncManager | serializeBatchMessage", "error", err, "svc", config.SVC)
			continue
		}
		//创建req
		req, ok := aqcreq[addr]
		if !ok || req == nil {
			req = &syncproto.AsyncRequest{}
			aqcreq[addr] = req
		}
		//设置数据
		data := &syncproto.AsyncMetadata{
			Tuple: &syncproto.MetaTuple{
				AppId:    v.appId,
				Channel:  v.channel,
				Function: v.function,
			},
			RequestId: v.requestId,
			Expire:    int64(v.expire),
			Op:        v.op,
		}
		req.Data = append(req.Data, data)
		tool.L.Debugw("rep-sdk | AqcCounterSyncManager | serializeBatchMessage", "addr", addr)
	}
	return aqcreq
}

func (a *AsyncCounterSyncManager) syncToLks(msgBatch []asyncCountMessage) {
	msg := a.serializeBatchMessage(msgBatch)
	a.sendToLksByXsf(msg, AsyncOp)
	a.reset()
}

func (a *AsyncCounterSyncManager) sync(protodate []byte) {
	//解析
	reqData := &syncproto.AsyncRequest{}
	err := proto.Unmarshal(protodate, reqData)
	if err != nil {
		tool.L.Errorw("rep-sdk | AqcCounterSyncManager | sync", "error", err)
		return
	}
	if reqData.GetData() == nil || len(reqData.GetData()) == 0 {
		tool.L.Errorw("rep-sdk | AqcCounterSyncManager | sync", "error", "data is nil or empty")
		return
	}
	msgBatch := make([]asyncCountMessage, 0)
	for _, v := range reqData.GetData() {
		if v == nil || v.Tuple == nil {
			continue
		}
		msgBatch = append(msgBatch, asyncCountMessage{
			counterKey: counterKey{
				appId:    v.Tuple.AppId,
				channel:  v.Tuple.Channel,
				function: v.Tuple.Function,
			},
			requestId: v.RequestId,
			expire:    int(v.Expire),
			op:        v.Op,
		})
	}
	//封装
	msg := a.serializeBatchMessage(msgBatch)
	//发送
	a.sendToLksByXsf(msg, AsyncCrossSyncOp)
}

func (a *AsyncCounterSyncManager) sendToLksByXsf(reqsmap map[string]*syncproto.AsyncRequest, op string) {
	m := map[string]string{"op": op}
	for _, req := range reqsmap {
		monitor.WithCommonCounter(AsyncOp, monitor.SendTotal)
		st := time.Now()
		if req == nil || len(req.Data) == 0 || req.Data[0].Tuple == nil {
			continue
		}
		//创建xsf request
		xsfreq := xsf.NewReq()
		//插入数据
		databyte, err := proto.Marshal(req)
		if err != nil {
			tool.L.Errorw("rep-sdk | AsyncCounterSyncManager | sendToLksByXsf", "error", err, "req", req)
			continue
		}
		xsfreq.Append(databyte, m)
		xsfreq.SetOp(op)
		//发送数据
		caller := xsf.NewCaller(a.xsfc)
		caller.WithHashKey(req.Data[0].Tuple.AppId)

		done := make(chan struct{})
		errCh := make(chan error)

		go func() {
			_, errcode, err := caller.Call(config.SVC, op, xsfreq, a.to)
			if err != nil {
				tool.L.Errorw("rep-sdk | AsyncCounterSyncManager | sendToLksByXsf", "code", errcode, "error", err, "req", xsfreq)
			}
			if err != nil {
				errCh <- fmt.Errorf("code: %d, error: %v, req: %v", errcode, err, xsfreq)
			}
			close(done)
		}()

		select {
		case <-done:
			// 请求完成，继续下一个请求
			monitor.WithCommonCounter(AsyncOp, monitor.SendSuccess)
			tool.L.Infof("rep-sdk | AsyncCounterSyncManager | sendToLksByXsf | done, req: %v", xsfreq)
		case err := <-errCh:
			tool.L.Errorw("rep-sdk | AsyncCounterSyncManager | sendToLksByXsf", "error", err)
		case <-time.After(a.to):
			// 超时，直接放弃退出
			tool.L.Infof("rep-sdk | AsyncCounterSyncManager | sendToLksByXsf | timeout, req: %v", xsfreq)
			continue
		}
		monitor.WithCost(AsyncOp, time.Since(st))
	}
}

func (a *AsyncCounterSyncManager) reset() {
	//a.batchCount = 0
	//a.msgBatch = a.msgBatch[:0]
}
