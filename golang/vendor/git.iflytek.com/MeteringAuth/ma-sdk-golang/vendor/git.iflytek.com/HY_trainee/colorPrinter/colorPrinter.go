package colorprinter

import (
	"fmt"
	"strings"
)

func NewcPrinterEx(f func(interface{}) string) *cPrinter {
	return &cPrinter{f}
}

func NewcPrinter(color FrontColor) *cPrinter {
	return &cPrinter{NewBashText(color).F}
}

type cPrinter struct {
	f func(interface{}) string
}

func (l *cPrinter) Print(val ...interface{}) (n int, err error) {
	return fmt.Print(l.f(fmt.Sprint(val...)))
}

func (l *cPrinter) Printf(layout string, val ...interface{}) (n int, err error) {
	return fmt.Print(l.f(fmt.Sprintf(layout, val...)))
}

func (l *cPrinter) Println(val ...interface{}) (n int, err error) {
	valstr := []string{}
	for _, v := range val {
		valstr = append(valstr, fmt.Sprintf("%v", v))
	}
	return fmt.Println(l.f(strings.Join(valstr, " ")))
}
