package main

import (
	"bytes"
	"log"
	"os"
	"time"
)

var logger = log.New(os.Stderr, "", log.LstdFlags)

const (
	liveProc = "LIVE_PROC"
	livePort = "LIVE_PORT"

	version       = "1.0.0"
	globalTimeout = time.Second
)

func rmLineBreak(in []byte) []byte {
	return bytes.TrimSuffix(in, []byte("\n"))
}
