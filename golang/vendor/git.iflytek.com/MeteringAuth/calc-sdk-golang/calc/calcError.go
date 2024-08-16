package calc

import "fmt"

var (
	CalcSuccess             = &CalcError{Code: 0, Desc: "success"}
	CalcInvalidParams       = &CalcError{Code: 20000, Desc: "invalid user parameters"}
	CalcQueueFull           = &CalcError{Code: 20001, Desc: "message queue is fulled"}
	CalcHasNotInited        = &CalcError{Code: 20002, Desc: "sdk had not been initialized"}
	CalcServerConnectFailed = &CalcError{Code: 20003, Desc: "client can not connect to the server"}
	CalcStoped              = &CalcError{Code: 20004, Desc: "sdk has finished"}
	CalcFinderInitFailed    = &CalcError{Code: 20005, Desc: "finder init failed"}
	CalcNoSuchKeyInConfig   = &CalcError{Code: 20006, Desc: "no such key in config map"}
	CalcTypeAssertError     = &CalcError{Code: 20007, Desc: "varible type assert error"}
	CalcHostsOrTopicError   = &CalcError{Code: 20008, Desc: "hosts or topic invalid"}
)

type CalcError struct {
	Code   int
	Desc   string
	Detail string
}

func (c CalcError) Error() string {
	if c.Detail == "" {
		return c.Desc
	}
	return c.Desc + fmt.Sprintf("(Detail:%s)", c.Detail)
}

func (c CalcError) GetCode() int {
	return c.Code
}

func (c *CalcError) SetDetail(d string) *CalcError {
	c.Detail = d
	return c
}

func (c CalcError) GetDetail() string {
	return c.Detail
}
