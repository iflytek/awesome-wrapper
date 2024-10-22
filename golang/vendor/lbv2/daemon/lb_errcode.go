package daemon

import (
	"fmt"
)

var (
	ErrLbStrategyIsNotSupport = &LbErrImpl{"input lb strategy not supported now", 20001}
	ErrLbNoSurvivingNode      = &LbErrImpl{"no surviving nodes", 20002}
	ErrLbInputOperation       = &LbErrImpl{"input is incorrect or is null", 2003}
	ErrLbUnknown              = &LbErrImpl{"occur unknown error", 2004}
	ErrBestIsIncorrect        = &LbErrImpl{"input best is incorrect or is null", 2005}
	ErrLbNbestIsIncorrect     = &LbErrImpl{"input nbest is incorrect or is null", 2006}
	ErrLbUidIsIncorrect       = &LbErrImpl{"input uid is incorrect or is null", 2046}
	ErrLbExParamIsIncorrect   = &LbErrImpl{"input exparam is incorrect or is null", 2046}
	ErrLbAddrIsIncorrect      = &LbErrImpl{"input addr is incorrect or is null", 2007}
	ErrLbTotalIsIncorrect     = &LbErrImpl{"input total is incorrect or is null", 2008}
	ErrLbIdleIsIncorrect      = &LbErrImpl{"input idle is incorrect or is null", 2009}
	ErrLbLoopQueueSize        = &LbErrImpl{"input LoopQueue Size is incorrect ", 2010}
	ErrLbSvcIncorrect         = &LbErrImpl{"input svc is incorrect or is null", 2011}
	ErrLbSubSvcIncorrect      = &LbErrImpl{"input subsvc is incorrect or is null", 2012}
	ErrLbInternalError        = &LbErrImpl{"internal error", 2013}
)

type LbErr interface {
	Error() string
	ErrInfo() string
	ErrorCode() int32
}

type LbErrImpl struct {
	errInfo string
	errCode int32
}

func NewLbErrImpl(errCode int32, errInfo string) *LbErrImpl {
	return &LbErrImpl{errInfo, errCode}
}
func (l *LbErrImpl) Error() string {
	return fmt.Sprintf("errInfo:%v,errCode:%v", l.errInfo, l.errCode)
}
func (l *LbErrImpl) ErrInfo() string {
	return l.errInfo
}
func (l *LbErrImpl) ErrorCode() int32 {
	return l.errCode
}
