package monitor

import (
	"time"

	"git.iflytek.com/MeteringAuth/ma-sdk-golang/ma/tool"
)

func WithCallErr(op string, code int32) {
	select {
	case opCallCh <- &callRlt{op, code}:
	default:
		tool.L.Warnw("rep-sdk | WithCallErr overflow", "op", op, "code", code, "len", len(opCallCh))
	}
}

func WithCost(op string, to time.Duration) {
	select {
	case opCostCh <- &callCost{op, to}:
	default:
		tool.L.Warnw("rep-sdk | WithCost overflow", "op", op, "cost", to.String(), len(opCostCh))
	}
}

func WithCommonCounter(op string, datatype string) {
	select {
	case opCommonCounterCh <- &commonCounter{op, datatype}:
	default:
		tool.L.Warnw("rep-sdk | WithCommonCounter overflow", "op", op, "datatype", datatype, len(opCommonCounterCh))
	}
}
