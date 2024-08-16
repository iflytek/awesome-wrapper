package monitor

import (
	"strconv"
	"time"

	xsf "git.iflytek.com/AIaaS/xsf/server"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/ma/config"
)

type callRlt struct {
	op   string
	code int32
}

type callCost struct {
	op string
	to time.Duration
}

//type commonGauge struct {
//	op   string
//	code float64
//}

type commonCounter struct {
	op       string
	datatype string
}

var (
	opCallCnt *xsf.CounterVec
	opCallCh  chan *callRlt

	opCostMs *xsf.HistogramVec
	opCostCh chan *callCost

	//opCommonGauge   *xsf.GaugeVec
	//opCommonGaugeCh chan *commonGauge

	opCommonCounter   *xsf.CounterVec
	opCommonCounterCh chan *commonCounter
)

const (
	opCallName = "MARepOp"
	opCostName = "MARepCost"

	opCommonGaugeName   = "MARepCommonGauge"
	opCommonCounterName = "MARepCommonCounter"

	SendFail    = "send_fail"
	SendSuccess = "send_success"
	SendTotal   = "send_total"
	WhiteSkip   = "white_skip"
	ReportTotal = "report_total"
	ReportDrop  = "report_drop"
)

func Init() (err error) {
	cfg := config.C.Metrics

	opCallCnt = xsf.NewCounterVec(
		xsf.CounterOpts{
			Name: opCallName,
			Help: "rep-sdk error code counter",
		},
		[]string{"op", "code"},
	)
	if err = xsf.Register(opCallName, opCallCnt); err != nil {
		return
	}
	opCallCh = make(chan *callRlt, cfg.MonitorSize)

	opCostMs = xsf.NewHistogramVec(
		xsf.HistogramOpts{
			Name: opCostName,
			Help: "rep-sdk op cost",
		},
		[]string{"op"},
	)
	if err = xsf.Register(opCostName, opCostMs); err != nil {
		return
	}
	opCostCh = make(chan *callCost, cfg.MonitorSize)

	//opCommonGauge = xsf.NewGaugeVec(
	//	xsf.GaugeOpts{
	//		Name: opCommonGaugeName,
	//		Help: "rep-sdk common gauge",
	//	}, []string{"op", "type"})
	//if err = xsf.Register(opCommonGaugeName, opCommonGauge); err != nil {
	//	return
	//}
	//opCommonGaugeCh = make(chan *commonGauge, cfg.MonitorSize)

	opCommonCounter = xsf.NewCounterVec(
		xsf.CounterOpts{
			Name: opCommonCounterName,
			Help: "rep-sdk common counter",
		}, []string{"op", "type"})
	if err = xsf.Register(opCommonCounterName, opCommonCounter); err != nil {
		return
	}
	opCommonCounterCh = make(chan *commonCounter, cfg.MonitorSize)

	runmonitor()
	return
}

func runmonitor() {
	//go func() {
	//	for {
	//		c := <-opCallCh
	//		opCallCnt.WithLabelValues(c.op, strconv.Itoa(int(c.code))).Inc()
	//	}
	//}()
	//
	//go func() {
	//	for {
	//		c := <-opCostCh
	//		opCostMs.WithLabelValues(c.op).Observe(float64(c.to.Milliseconds()))
	//	}
	//}()
	go func() {
		for {
			call := <-opCallCh
			opCallCnt.WithLabelValues(call.op, strconv.Itoa(int(call.code))).Inc()

			cost := <-opCostCh
			opCostMs.WithLabelValues(cost.op).Observe(float64(cost.to.Milliseconds()))

			comCount := <-opCommonCounterCh
			opCommonCounter.WithLabelValues(comCount.op, comCount.datatype).Inc()
		}
	}()
}
