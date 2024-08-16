package monitor

import (
	"fmt"
	"time"

	"git.iflytek.com/MeteringAuth/ma-sdk-golang/ma/tool"
)

func WithMsgType(msgType string) {
	select {
	case msgCntCh <- msgType:
	default:
		tool.L.Warnw("calc-sdk | WithMsgType overflow", "type", msgType, len(msgCntCh))
	}
}

func WithMsgErr(msgErr string) {
	select {
	case msgErrCh <- msgErr:
	default:
		tool.L.Warnw("calc-sdk | WithMsgErr overflow", "error", msgErr, len(msgErrCh))
	}
}

func WithMsgRmqErr(host, port string) {
	select {
	case msgErrCh <- fmt.Sprintf("%s:%s", host, port):
	default:
		tool.L.Warnw("calc-sdk | WithMsgRmqErr overflow", "host", host, "port", port, "len", len(msgErrCh))
	}

	WithMsgErr(MsgErrRMQ)
}

func WithRPCCost(cost, expect time.Duration) {}
