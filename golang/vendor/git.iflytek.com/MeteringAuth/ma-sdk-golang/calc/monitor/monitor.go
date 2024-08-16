package monitor

import (
	xsf "git.iflytek.com/AIaaS/xsf/server"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/ma/config"
	"git.iflytek.com/MeteringAuth/ma-sdk-golang/ma/tool"
)

var (
	idc  string
	from string
)
var (
	msgCntGauge *xsf.CounterVec
	msgCntCh    chan string

	msgErrGauge *xsf.GaugeVec
	msgErrCh    chan string
)

const (
	msgCntName   = "MACalcMsg"
	MsgCntRPC    = "rpc"
	MsgCntRMQ    = "rmq"
	MsgCntPulsar = "pulsar"

	msgErrName   = "MACalcErr"
	MsgErrQFull  = "queuefull"
	MsgErrRPC    = "rpc"
	MsgErrRMQ    = "rmq"
	MsgErrPulsar = "pulsar"
)

func Init() (err error) {
	cfg := config.C.Metrics

	msgCntGauge = xsf.NewCounterVec(
		xsf.CounterOpts{
			Name: msgCntName,
			Help: "calc-sdk calc counter",
		},
		[]string{"idc", "from", "sender"},
	)
	if err = xsf.Register(msgCntName, msgCntGauge); err != nil {
		return
	}
	msgCntCh = make(chan string, cfg.MonitorSize)

	msgErrGauge = xsf.NewGaugeVec(
		xsf.GaugeOpts{
			Name: msgErrName,
			Help: "calc-sdk error",
		},
		[]string{"idc", "from", "error"},
	)
	if err = xsf.Register(msgErrName, msgErrGauge); err != nil {
		return
	}
	msgErrCh = make(chan string, cfg.MonitorSize)

	idc = cfg.IDC
	from = cfg.SUB

	runmonitor()
	tool.CalcPrinter.Println("monitor init", "idc", idc, "from", from)
	return
}

func runmonitor() {
	go func() {
		for {
			msgCntGauge.WithLabelValues(idc, from, <-msgCntCh).Inc()
		}
	}()

	go func() {
		for {
			msgErrGauge.WithLabelValues(idc, from, <-msgErrCh).Inc()
		}
	}()
}
