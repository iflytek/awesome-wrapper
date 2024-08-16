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
	"strconv"
	"sync/atomic"
	"time"
	"unsafe"
)

var (
	// sleep ts in gen under low load by micro-second
	LowLoadSleepTs = 100

	// flag to control gen status
	genStatus   int32 = STATUA_NOT_RUNNING
	genExitChan       = make(chan int)
)

func startGen() {
	genStatus = STATUS_RUNNING
	go gen()
}

func stopGen() {
	if atomic.CompareAndSwapInt32(&genStatus, STATUS_RUNNING, STATUA_NOT_RUNNING) {
		<-genExitChan
		infof("stop gen go routine")
	}
}

// start gen running go routine
func gen() {
	infof("start gen go routine")

	for genStatus == STATUS_RUNNING {
		isWorking := false

		// check consumer status, whether to start ?
		for i := 0; i < ConsumerClientNum; i++ {
			if 1.0*float32(consumerClients[i].spanSeQueue.len())/float32(consumerClients[i].spanSeQueue.cap()) > 0.5 {
				consumerClients[i].start()
				isWorking = true
			}
			if consumerClients[i].spanSeQueue.len() > 0 && (CurrentTimeMillis()-consumerClients[i].lastDeliveryTs) > int64(LingerSec*1000) {
				consumerClients[i].start()
				isWorking = true
			}
		}

		if !isWorking {
			time.Sleep(time.Duration(LowLoadSleepTs) * time.Microsecond)
		}
	}

	genExitChan <- 1
}

func getSpan() *Span {
	// init a raw span
	span := initSpan().(*Span)

	microsTs := CurrentTimeMicros()
	millisTs := microsTs / 1000

	// third, set & reset span
	span.traceId = span.traceId[:0]
	{
		// new trace id logical
		span.traceId = append(span.traceId, tagZero[0:8-onceIpUint32Len]...)
		span.traceId = append(span.traceId, onceIPUint32...)

		span.traceId = strconv.AppendInt(span.traceId, millisTs, 10)
		// span.TraceId = appendInt(span.TraceId, genGlobalMillisTs)

		// replace Rand with auto-increased value
		rand := atomic.AddUint64(&atomicTraceId, 1) % 10000
		var tempRandStrLen = 0
		if rand > 999 {
			tempRandStrLen = 4
		} else if rand > 99 {
			tempRandStrLen = 3
		} else if rand > 9 {
			tempRandStrLen = 2
		} else {
			tempRandStrLen = 1
		}
		span.traceId = append(span.traceId, tagZero[0:4-tempRandStrLen]...)
		span.traceId = strconv.AppendInt(span.traceId, int64(rand), 10)
		// span.TraceId = appendInt(span.TraceId, int64(rand))

		span.traceId = append(span.traceId, 'a')

		span.traceId = append(span.traceId, tagZero[0:6-oncePortLen]...)
		span.traceId = append(span.traceId, oncePort...)
	}
	span.name = ""
	span.timestamp = microsTs
	// id with format [&address|ts|0.1.1]
	span.spanIdTs = span.spanIdTs[:0]
	{
		// span id logical
		span.spanIdTs = strconv.AppendInt(span.spanIdTs, int64(uintptr(unsafe.Pointer(&span.traceId))), 16)
		// span.spanIdTs = appendInt(span.spanIdTs, int64(uintptr(unsafe.Pointer(&span.TraceId))))
		span.spanIdTs = strconv.AppendInt(span.spanIdTs, span.timestamp, 10)
		// span.spanIdTs = appendInt(span.spanIdTs, span.Timestamp)
	}
	span.duration = 0
	span.spanType = SERVER
	span.currentChildId = 0
	span.spanIdHierarchy = span.spanIdHierarchy[:0]

	return span
}
