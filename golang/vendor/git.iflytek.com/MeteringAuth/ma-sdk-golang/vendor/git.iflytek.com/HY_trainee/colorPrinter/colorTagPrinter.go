package colorprinter

import "fmt"

func NewctPrinter(tag string, color FrontColor) *ctPrinter {
	return &ctPrinter{tag, NewcPrinter(color)}
}

func NewctPrinterEx(tag string, cp *cPrinter) *ctPrinter {
	return &ctPrinter{tag, cp}
}

type ctPrinter struct {
	tag string
	cp  *cPrinter
}

func (l *ctPrinter) Print(args ...interface{}) {
	l.cp.Print(fmt.Sprintf("[%s]", l.tag))
	l.cp.Print(args...)
}

func (l *ctPrinter) Printf(format string, args ...interface{}) {
	l.cp.Print(fmt.Sprintf("[%s]", l.tag))
	l.cp.Printf(format, args...)
}

func (l *ctPrinter) Println(args ...interface{}) {
	l.cp.Print(fmt.Sprintf("[%s]", l.tag))
	l.cp.Println(args...)
}
