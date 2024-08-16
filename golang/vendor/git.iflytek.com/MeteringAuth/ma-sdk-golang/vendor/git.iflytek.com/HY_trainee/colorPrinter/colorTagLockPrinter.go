package colorprinter

import "sync"

type ctlPrinter struct {
	lock sync.Mutex
	tag  string
	cp   *cPrinter
}
