package err

import "fmt"

type ErrType string
const(
	ErrCommon ErrType = "common"
)


type Err struct {
	E string
	T ErrType
}

func (e Err) Error() string {
	return fmt.Sprintf("%s|%s",e.T,e.E)
}

type Error  = *Err

func NewErr(t ErrType ,m string)Error{
	return &Err{
		E: m,
		T: t,
	}
}

func NewCommonError(e string)Error{
	return NewErr(ErrCommon,e)
}


func d()Error{
	return &Err{}
}