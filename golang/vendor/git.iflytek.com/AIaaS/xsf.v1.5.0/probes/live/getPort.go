package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func getPort(p []byte) []byte {
	port := strings.Replace(os.Getenv(livePort), `"`, ``, -1)
	if len(port) != 0 {
		logger.Printf("extrace port:%v from env:%v\n", port, livePort)
		return []byte(os.Getenv(livePort))
	}
	cmdStr := fmt.Sprintf(`ss -tlnp | grep %v | awk  '{print $4}'|awk -F ':' '{print $2}'`, string(p))
	logger.Printf("ss cmdStr:%v\n", cmdStr)
	shOut, shOutErr := exec.Command("sh", "-c", cmdStr).CombinedOutput()
	shOut = rmLineBreak(shOut)
	fmt.Printf("shOut:%s,shOutErr:%v\n", shOut, shOutErr)
	checkErr(shOutErr)
	return bytes.TrimSuffix(shOut, []byte("\n"))
}
