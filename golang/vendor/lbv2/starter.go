package main

import (
	"fmt"
	"lbv2/daemon"
)

func main() {
	if err := daemon.RunServer(); err != nil {
		fmt.Println("error running server:", err)
		return
	}
}
