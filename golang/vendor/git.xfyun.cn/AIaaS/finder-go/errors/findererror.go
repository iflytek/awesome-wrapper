package finder

import (
	"fmt"
)

type FinderError struct {
	Ret  ReturnCode
	Func string
	Desc string
}

func (fe *FinderError) Error() string {
	format := `An error caught in %s, %s[%s].`
	return fmt.Sprintf(format, fe.Func, fe.Desc, fe.Ret)
}
