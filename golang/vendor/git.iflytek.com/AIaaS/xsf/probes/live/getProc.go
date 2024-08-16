package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func getProc() []byte {
	proc := strings.Replace(os.Getenv(liveProc), `"`, ``, -1)
	if len(proc) != 0 {
		logger.Printf("extrace liveProc:%v from env:%v\n", proc, liveProc)
		return []byte(proc)
	}
	logger.Printf("the proc:%v is empty string\n", proc)

	shOut, shOutErr := exec.Command("sh", "-c", "ps -p 1 | awk '/1/{print $NF}'").CombinedOutput()
	shOut=rmLineBreak(shOut)
	fmt.Printf("shOut:%s,shOutErr:%v\n", shOut, shOutErr)
	checkErr(shOutErr)
	return shOut
}
