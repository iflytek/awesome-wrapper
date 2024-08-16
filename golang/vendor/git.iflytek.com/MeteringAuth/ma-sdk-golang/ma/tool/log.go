package tool

import (
	xsf "git.iflytek.com/AIaaS/xsf/client"
	cp "git.iflytek.com/HY_trainee/colorPrinter"
)

var (
	CalcPrinter = cp.NewctPrinter("calc-sdk", cp.Blue)
	LiccPrinter = cp.NewctPrinter("licc-sdk", cp.Yellow)
	RepPrinter  = cp.NewctPrinter("report-sdk", cp.Magenta)
)

var L *xsf.Logger
