package colorprinter

import "fmt"

func NewBashText(c FrontColor) *bashText {
	return &bashText{color: c}
}

type bashText struct {
	color FrontColor
}

func (b *bashText) F(val interface{}) string {
	return fmt.Sprintf("\033[%dm%v\033[0m", b.color, val)
}

func (b *bashText) String() string {
	return fmt.Sprintf("xxx")
}
