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

var (
	opCallCnt *xsf.CounterVec
	opCallCh  chan *callRlt

	opCostMs *xsf.HistogramVec
	opCostCh chan *callCost
)

const (
	opCallName = "MALiccOp"
	opCostName = "MALiccCost"
)

func Init() (err error) {
	cfg := config.C.Metrics

	opCallCnt = xsf.NewCounterVec(
		xsf.CounterOpts{
			Name: opCallName,
			Help: "licc-sdk error code counter",
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
			Help: "licc-sdk op cost",
		},
		[]string{"op"},
	)
	if err = xsf.Register(opCostName, opCostMs); err != nil {
		return
	}
	opCostCh = make(chan *callCost, cfg.MonitorSize)

	runmonitor()
	return
}

func runmonitor() {
	go func() {
		for {
			c := <-opCallCh
			opCallCnt.WithLabelValues(c.op, strconv.Itoa(int(c.code))).Inc()
		}
	}()

	go func() {
		for {
			c := <-opCostCh
			opCostMs.WithLabelValues(c.op).Observe(float64(c.to.Milliseconds()))
		}
	}()
}
