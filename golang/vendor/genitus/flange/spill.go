/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package flange

import (
	"errors"
	"flume"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync/atomic"
)

var (
	// log spill dir
	SpillDir string = "." + string(os.PathSeparator) + "spill"
	// flag to enable spill
	SpillEnable bool = true
	// max spill content size in G-byte
	MaxSpillContentSize int64 = 1
)

var (
	lastReverseSpillTs int64 = 0
	// spill exit flag
	reverseSpillStatus int32 = STATUA_NOT_RUNNING
	spillStatus        int32 = STATUA_NOT_RUNNING
	// chan to control exit go routine
	spillExitChan chan int
	// spillRingBuffer for spill data in-memory saving
	spillSeQueue *seQueue
	// spill file descriptor
	spillFD *os.File = nil
	// flume client
	reverseSpillFlumeClient *flumeClient
	// temp flumeEvent batch
	reverseSpillEvent *flume.ThriftFlumeEvent
)

// init spill profile
func initSpillProf(flumeHost string, flumePort string) error {
	// check & create spill dir
	if SpillDir != "" {
		infof("mkdir spill dir\n")

		if err := os.MkdirAll(SpillDir, 0755); err != nil {
			errorf("mkdir spillDir err : %v\n", err)
			return err
		}
	} else {
		errorf("unvalid spill dir : %v\n", SpillDir)
		return errors.New("unvalid spiil dir : " + SpillDir)
	}

	lastReverseSpillTs = 0
	atomic.StoreInt32(&reverseSpillStatus, STATUA_NOT_RUNNING)
	atomic.StoreInt32(&spillStatus, STATUA_NOT_RUNNING)
	spillExitChan = make(chan int)
	spillSeQueue = newSeQueue(BuffSize)
	// open spill file
	spillFileName := SpillDir + string(os.PathSeparator) + "span_spill"
	fp, err := os.OpenFile(spillFileName, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		errorf("open %s failed : %v\n", spillFileName, err)
	}
	// seek to end, cause open for append
	fp.Seek(0, io.SeekEnd)
	spillFD = fp

	// init flume client
	reverseSpillFlumeClient = &flumeClient{
		host: flumeHost,
		port: flumePort,
	}
	reverseSpillEvent = &flume.ThriftFlumeEvent{Headers: make(map[string]string, 8)}
	reverseSpillEvent.Headers[MESSAGE_QUEUE_KAFKA_TOPIC_KEY] = Topic
	reverseSpillEvent.Headers[SCHEMA_VERSION_KEY] = DEFAULT_SCHEMA_VERSION_VALUE
	reverseSpillEvent.Headers[SCHEMA_NAME_KEY] = DEFAULT_SCHEMA_NAME_VALUE
	reverseSpillEvent.Headers[SPAN_SERIALIZATION] = "false"

	return nil
}

//==spill process==
// start a spill go routine
func startSpill() {
	if atomic.CompareAndSwapInt32(&spillStatus, STATUA_NOT_RUNNING, STATUS_RUNNING) {
		go spill()
	}
}

// pause spill go routine
func pauseSpill() {
	if atomic.CompareAndSwapInt32(&spillStatus, STATUS_RUNNING, STATUA_NOT_RUNNING) {
		<-spillExitChan
		infof("pause spill go routine\n")
	}
}

// stop current spill go routine, call after pauseSpill()
func stopSpill() {
	if atomic.LoadInt32(&spillStatus) == STATUS_RUNNING {
		pauseSpill()
	}
	if atomic.LoadInt32(&reverseSpillStatus) == STATUS_RUNNING {
		pauseReverseSpill()
	}
	if atomic.LoadInt32(&spillStatus) == STATUA_NOT_RUNNING {
		spillFD.Close()
		infof("stop spill go routine\n")
	}
}

// start spill go routine
func spill() {
	atomic.StoreInt32(&spillStatus, STATUS_RUNNING)
	infof("start spill go routine")

	// get current size before a spill loop
	currentContentSize, _ := spillFD.Seek(0, io.SeekEnd)

	// if not spillExit, then spill
	for item := spillSeQueue.get(); atomic.LoadInt32(&spillStatus) == STATUS_RUNNING && item != nil; item = spillSeQueue.get() {
		// check spill content size
		if currentContentSize >= MaxSpillContentSize*1024*1024*1024 {
			// if no space to spill, then skip write
			continue
		}

		// spill data with format --> SpillContent|len(SpillContent)
		buf, err := Serialize(item, true)
		if err != nil {
			continue
		}
		l, _ := fmt.Fprint(spillFD, string(buf))
		currentContentSize += int64(l)
		l, _ = fmt.Fprintf(spillFD, "%012d", len(buf))
		currentContentSize += int64(l)
	}

	// if exit spill, spill the least data
	for item := spillSeQueue.get(); item != nil; item = spillSeQueue.get() {
		// check spill content size
		if currentContentSize >= MaxSpillContentSize*1024*1024*1024 {
			// if no space to spill, then skip write
			continue
		}

		// spill data with format --> SpillContent|len(SpillContent)
		buf, err := Serialize(item, true)
		if err != nil {
			continue
		}
		l, _ := fmt.Fprint(spillFD, string(buf))
		currentContentSize += int64(l)
		l, _ = fmt.Fprintf(spillFD, "%012d", len(buf))
		currentContentSize += int64(l)
	}

	infof("exit spill go routine")

	if atomic.CompareAndSwapInt32(&spillStatus, STATUS_RUNNING, STATUA_NOT_RUNNING) {
		// exit by itself
	} else {
		// if spill exit by other command
		spillExitChan <- 1
	}
}

//==reverse spill process==
// start a spill go routine
func startReverseSpill() {
	if atomic.CompareAndSwapInt32(&reverseSpillStatus, STATUA_NOT_RUNNING, STATUS_RUNNING) {
		go reverseSpill()
	}
}

// pause spill go routine
func pauseReverseSpill() {
	if atomic.CompareAndSwapInt32(&reverseSpillStatus, STATUS_RUNNING, STATUA_NOT_RUNNING) {
		<-spillExitChan
		infof("pause reverse spill go routine\n")
	}
}

// start spill go routine
func reverseSpill() {
	atomic.StoreInt32(&reverseSpillStatus, STATUS_RUNNING)
	infof("start reverse spill go routine")

	reverseSpillFlumeClient.open()

	// read file, and send to flume
	lengthBuffer := make([]byte, 12)
	spillFD.Seek(-12, io.SeekEnd)

	// start read and truncate spill file
	// 12 for max spill content length format %012d each record
	for l, err := spillFD.Read(lengthBuffer); l == 12 && err == nil && atomic.LoadInt32(&reverseSpillStatus) == STATUS_RUNNING; l, err = spillFD.Read(lengthBuffer) {
		length, _ := strconv.Atoi(string(lengthBuffer))

		// seek to start position
		spillFD.Seek(int64(-(length + 12)), io.SeekEnd)
		buffer := make([]byte, length)
		if l, err := spillFD.Read(buffer); l == length && err == nil {
			if sendLog(buffer) {
				state, _ := spillFD.Stat()
				spillFD.Truncate(state.Size() - int64(length+12))

				// seek to next position
				// length for current record, 12 for next record length
				spillFD.Seek(int64(-(12 + length)), io.SeekCurrent)
			} else {
				// no change for spill file, exit for next try
				break
			}
		} else {
			// no change for spill file, exit for next try
			break
		}
	}

	spillFD.Seek(0, io.SeekEnd)
	reverseSpillFlumeClient.close()

	infof("exit reverse spill go routine")

	if atomic.CompareAndSwapInt32(&reverseSpillStatus, STATUS_RUNNING, STATUA_NOT_RUNNING) {
		// if reverse spill exit by itself
	} else {
		// if reverse spill exit by other command
		spillExitChan <- 1
	}
}

func sendLog(bytes []byte) bool {
	defer catch("reverse spill send log")

	// retrieve span info
	traceId, spanId, spanType, buf := RetrieveSpanInfo(bytes)
	infof("retrieve tid=%s, sid=%s", traceId, spanId)

	reverseSpillEvent.Body = buf
	// set headers, see java syringe.v2 FlumeClient
	// set timestamp to the millseconds of the start of this trace
	reverseSpillEvent.Headers[TIMESTAMP] = traceId[8:21]
	// set r.k sid for mq balancing, format: <traceId>
	reverseSpillEvent.Headers[MESSAGE_QUEUE_RECORD_KEY_KEY] = traceId + "#" + spanId + "#" + spanType
	// fc.evBatches[i].Headers[FLUSH_TIMESTAMP] = strconv.FormatInt(CurrentTimeMillis(), 10)

	// make sure rpc client is open before send msg
	if !reverseSpillFlumeClient.rpcClient.Transport.IsOpen() {
		errorf("reverse spill flume client is not open, now reset it")
		reverseSpillFlumeClient.close()
		reverseSpillFlumeClient.transport = nil
		reverseSpillFlumeClient.rpcClient = nil
		reverseSpillFlumeClient.open()
		return false
	}

	rs, err := reverseSpillFlumeClient.rpcClient.Append(reverseSpillEvent)
	if err != nil {
		errorf("reverseSpill send 1 log to flume failed, will try again later.")
		reverseSpillFlumeClient.close()
		reverseSpillFlumeClient.transport = nil
		reverseSpillFlumeClient.rpcClient = nil
		reverseSpillFlumeClient.open()
		return false
	} else {
		if rs != flume.Status_OK {
			debugf("reverseSpill send 1 log to flume failed, will try again later.")
			return false
		} else {
			debugf("reverseSpill send 1 log successful.")
			return true
		}
	}
}
