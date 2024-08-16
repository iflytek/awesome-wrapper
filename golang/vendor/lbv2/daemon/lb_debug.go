package daemon

import (
	"fmt"
	"path/filepath"
	"runtime"
)

var debugFlag = false
var debugInst *debugEx

func init() {
	debugInst = &debugEx{on: debugFlag}
}

type debugEx struct {
	on bool
}

func (d *debugEx) Init(on bool) {
	debugFlag = on
	d.on = on
}

func (d *debugEx) Debugf(format string, v ...interface{}) {
	if d.on {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("Pos:%v,Msg:%v\n", fmt.Sprintf("%v:%v", filepath.Base(file), line), fmt.Sprintf(format, v...))
	}
}
func (d *debugEx) Debug(a ...interface{}) {
	if d.on {
		_, file, line, _ := runtime.Caller(1)
		fmt.Println(fmt.Sprintf("Pos:%v,", fmt.Sprintf("%v:%v", filepath.Base(file), line)), fmt.Sprintln(a...))

	}
}
