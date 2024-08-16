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

package quiver

import (
	"os"
	"errors"
	"time"
	"fmt"
	"flume"
	"strconv"
	"io"
)

var (
	// log spill dir
	SpillDir string = "." + string(os.PathSeparator) + "spill"
	// flag to enable spill
	SpillEnable bool = true
	// max spill content size by GB
	MaxSpillContentSize int64 = 1
)

var (
	// delay time for hbase vs s3
	delayTimeMills int64 = 0

	// chan to control exit go routine
	spillExitChannel chan int
	// spill log cache channel
	spillEventChannel chan *EventData
	// spill file descriptor
	spillFD *os.File = nil
	// current spill content size
	currentContentSize int64
	// flume client
	reverseSpillFlumeClient *flumeClient
	// temp flumeEvent batch
	reverseSpillEvent *flume.ThriftFlumeEvent
	// reverseSpillHbaseClient
	reverseSpillHbaseClient *hbaseUtil
	// reverseSpillS3Client
	reverseSpillS3Client *s3Util
)

// init spill profile
func initSpillProf(flumeHost string, flumePort string) error {
	// check & create spill dir
	if SpillDir != "" {
		infof("mkdir spill dir")
		if err := os.MkdirAll(SpillDir, 0755); err != nil {
			errorf("mkdir spillDir err : %v", err)
			return err
		}
	} else {
		errorf("unvalid spill dir : %v", SpillDir)
		return errors.New("unvalid spiil dir : " + SpillDir)
	}

	// init stuff
	spillExitChannel = make(chan int)
	spillEventChannel = make(chan *EventData, BuffSize)
	// open spill file
	spillFileName := SpillDir + string(os.PathSeparator) + "event_spill"
	fp, err := os.OpenFile(spillFileName, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		errorf("open spill file %s failed : %v", spillFileName, err)
		return err
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
	// init hbase client
	reverseSpillHbaseClient = NewHbaseUtil()
	// init s3 client
	reverseSpillS3Client = NewS3Util()
	delayTimeMills = (time.Duration(DelayTime*24) * time.Hour).Nanoseconds() / 1000 / 1000

	// start background work
	globalWg.Add(1)
	go spillReverseGoroutine()
	return nil
}

// stop current spill go routine, call after pauseSpill()
func stopSpillReverse() {
	if spillExitChannel != nil {
		spillExitChannel <- 1
	}
}

// start spill go routine
func spillReverseGoroutine() {
	defer catch("spill reverse go routine error")
	defer globalWg.Done()

	infof("start spill & reverse go routine")

	// seek spill fd
	currentContentSize, _ = spillFD.Seek(0, io.SeekEnd)
	// init reverse spill flume client
	reverseSpillFlumeClient.open()

	for {
		// retrieve a event\timeout\exit
		select {
		// check spill channel
		case evTask := <-spillEventChannel:
			if currentContentSize < MaxSpillContentSize*1024*1024*1024 {
				spillEvent(evTask)
			}
			// no spill event, check reverse file
		case <-time.After(time.Second * time.Duration(LingerSec)):
			reverseSend()
			// exit signal
		case <-spillExitChannel:
			goto FinishSpill
		}
	}

FinishSpill:
	debugf("==========spill exit==========")
	select {
	case item := <-spillEventChannel:
		if currentContentSize < MaxSpillContentSize*1024*1024*1024 {
			spillEvent(item)
		}
	default:
		// nothing to do but exit
	}

	// exit
	if reverseSpillFlumeClient != nil {
		reverseSpillFlumeClient.close()
	}
	if spillFD != nil {
		spillFD.Close()
	}
	infof("exit spill reverse go routine")
}

// spill event data to local file
func spillEvent(event *EventData) {
	buf, err := Serialize(event)
	if err != nil {
		errorf("spill serialize error : %v", err)
		return
	}

	// spill data with format --> SpillContent|len(SpillContent)
	l, _ := fmt.Fprint(spillFD, string(buf))
	currentContentSize += int64(l)
	l, _ = fmt.Fprintf(spillFD, "%012d", len(buf))
	currentContentSize += int64(l)
}

// read local file and send
func reverseSend() {
	defer catch("reverse send error")

	// read file, and send to flume
	lengthBuffer := make([]byte, 12)
	spillFD.Seek(-12, io.SeekEnd)

	// start read and truncate spill file
	// 12 for max spill content length format %012d each record
	for l, err := spillFD.Read(lengthBuffer); l == 12 && err == nil; l, err = spillFD.Read(lengthBuffer) {
		length, err := strconv.Atoi(string(lengthBuffer))
		if err != nil {
			errorf("length buffer with 012d to int error : %s, %v", string(lengthBuffer), err)
			break
		}

		// seek to start position
		_, err = spillFD.Seek(int64(-(length + 12)), io.SeekEnd)
		if err != nil {
			errorf("seed to start position error : %v", err)
			break
		}

		buffer := make([]byte, length)
		if l, err := spillFD.Read(buffer); l == length && err == nil {
			if sendLog(buffer) {
				state, _ := spillFD.Stat()
				spillFD.Truncate(state.Size() - int64(length+12))
				currentContentSize -= int64(length + 12)

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

	// reset spill fd position
	spillFD.Seek(0, io.SeekEnd)
}

func sendLog(bytes []byte) bool {
	defer catch("send log")

	// upload media to hbase
	event := Deserialize(bytes)
	if event == nil {
		errorf("reverse spill de-serialize record failed")
		return false
	}

	// media dispatch storage
	sidTs, err := strconv.ParseInt(event.Sid[14:25], 16, 64)
	if err != nil {
		errorf("parse sid timestamp error : %v", err)
		return false
	}
	if CurrentTimeMillis()-sidTs < delayTimeMills {
		// send to hbase, will delete media content
		reverseSpillHbaseClient.UploadMedia(event)
	} else {
		// send to s3, will delete media content
		reverseSpillS3Client.UploadMedia(event)
	}

	// now, no media content
	buf, err := Serialize(event)
	if err != nil {
		errorf("reverse spill serialize error : %v", err)
		return false
	}

	reverseSpillEvent.Body = buf
	// set headers, see java syringe.v2 FlumeClient
	// set timestamp to the millseconds of the start of this trace
	reverseSpillEvent.Headers[TIMESTAMP] = strconv.FormatInt(event.Timestamp, 10)
	// set r.k sid for mq balancing, format: <traceId>
	reverseSpillEvent.Headers[MESSAGE_QUEUE_RECORD_KEY_KEY] = event.Sid
	// fc.evBatches[i].Headers[FLUSH_TIMESTAMP] = strconv.FormatInt(CurrentTimeMillis(), 10)

	// make sure rpc client is open before send msg
	if !reverseSpillFlumeClient.rpcClient.Transport.IsOpen() {
		errorf("flume client not open, now reset it")
		reverseSpillFlumeClient.close()
		reverseSpillFlumeClient.transport = nil
		reverseSpillFlumeClient.rpcClient = nil
		reverseSpillFlumeClient.open()
		return false
	}

	// send to flume client
	rs, err := reverseSpillFlumeClient.rpcClient.Append(reverseSpillEvent)

	if err != nil {
		errorf("append batch error : %v", err)
		reverseSpillFlumeClient.close()
		reverseSpillFlumeClient.transport = nil
		reverseSpillFlumeClient.rpcClient = nil
		reverseSpillFlumeClient.open()
		return false
	} else {
		if rs != flume.Status_OK {
			infof("reverseSpill send 1 log to flume failed, will try again later.")
			return false
		} else {
			infof("reverseSpill send 1 log successful.")
			return true
		}
	}
}
