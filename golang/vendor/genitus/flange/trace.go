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
	"os"
	_ "runtime/debug"
	"sync/atomic"
)

var (
	// consumer clients
	consumerClients []*consumer = nil
	// default flume client number
	ConsumerClientNum = 8
)

var (
	// true if dump log to file.
	DumpEnable bool = false
	// log dump dir.
	DumpDir string = "." + string(os.PathSeparator) + "trace"

	// global init check with cas
	atomicInitInteger int32 = 0
	// count drop span
	atomicSpanDropCount int64 = 0
	// report for in-bound speed
	atomicSpanInBoundCount uint64 = 0
	// small then FlumeClientNum
	balanceIndex int = 0

	// retry count
	FlushRetryCount int = 10
	// true if deliver log to flume.
	DeliverEnable bool = true
	// force deliver the sample trace
	ForceDeliver bool = false
)

// initialize plunger.
// @param host:port determine the flume host:ip
// @param num number of backend-data-consumer in Flange, adjust with upstream load
func Init(flumeHost string, flumePort string, num int, serviceIP string, servicePort string, serviceName string) error {
	// check init already
	if !atomic.CompareAndSwapInt32(&atomicInitInteger, 0, 1) {
		infof("already init flange, ignore this operation")
		return nil
	}

	// set global value
	setGlobalConfig(serviceIP, servicePort, serviceName)

	// mkdir dump dir
	if DumpEnable && DumpDir != "" {
		infof("mkdir dumpdir")
		if err := os.MkdirAll(DumpDir, 0755); err != nil {
			errorf("mkdir dumpdir err : %v", err)
			return err
		}
	}

	// init spill info, not start, just wait for mg_center
	initSpillProf(flumeHost, flumePort)

	// uncheck consumer num
	if num < 1 {
		debugf("consumer num should not less than 1")
		return nil
	}
	ConsumerClientNum = num
	for i := 0; i < ConsumerClientNum; i++ {
		c := initConsumer(flumeHost, flumePort, i)
		consumerClients = append(consumerClients, c)
	}
	// start generate go routine
	startGen()
	// start mg go routine
	startMg()

	return nil
}

// flush dslog to buffer channel, will be send to flume.
func Flush(span *Span) error {
	defer catch("Flush")

	if span == nil {
		return errors.New("span is nil")
	}

	// dump to file
	if DumpEnable {
		span.dump(DumpDir)
	}

	// support for sdk sample, 'a' for abandon in trace id at index 25
	if !ForceDeliver && span.traceId[25] == 'a' {
		// release span
		return nil
	}

	// save to buffer only while plunger is enabled
	if DeliverEnable {
		// compute in-bound speed
		atomic.AddUint64(&atomicSpanInBoundCount, 1)

		// add span-tree to tree-ring-buffer with concurrent assigned balanceIndex for balance, see `flume_client.go`
		for i := 0; i < FlushRetryCount; i++ {
			// compute balanceIndex each time for fixed value failed
			balanceIndex = int(CurrentTimeMillis() % int64(ConsumerClientNum))

			if r := consumerClients[balanceIndex].spanSeQueue.put(span); r {
				return nil
			}
		}

		// add span to spill
		if SpillEnable {
			if r := spillSeQueue.put(span); r {
				return nil
			}
		}

		// print out to drop
		if atomic.AddInt64(&atomicSpanDropCount, 1) > 10000 {
			infof("%d drop span count = 10000", CurrentTimeMillis())
			atomic.StoreInt64(&atomicSpanDropCount, 0)
		}
	}

	return nil
}

// fini plunger.
func Fini() {
	defer catch("finish")

	// stop all component
	for _, c := range consumerClients {
		c.stop()
	}
	stopMg()
	stopSpill()
	stopGen()
}

// global error catch function
func catch(site string) {
	if err := recover(); err != nil {
		errorf("Error occur [%v] at [%s]", err, site)
		// debug.PrintStack()
	}
}
