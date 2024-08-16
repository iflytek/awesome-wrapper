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
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"sync/atomic"
	"time"
)

var (
	// switch for self web based log metric
	WatchLogEnable = true
	WatchLogPort   = 12331

	// current consumer count to control consumer by mg
	current_consumer_count int = 1

	// flag to control mg status
	mgStatus   int32 = STATUA_NOT_RUNNING
	mgExitChan       = make(chan int)
)

func startMg() {
	mgStatus = STATUS_RUNNING
	// for default 1 consumer client, start at trace.go init
	// current_consumer_count = 1
	go mg()
	if WatchLogEnable {
		go func() {
			infof("start metric report server")
			http.HandleFunc("/metric", ReportMetric)
			// http.HandleFunc("/span/{id}", )
			if err := http.ListenAndServe(fmt.Sprintf(":%d", WatchLogPort), nil); err != nil {
				errorf("start metric report server error: %v", err)
			}
		}()
	}
}

func stopMg() {
	if atomic.CompareAndSwapInt32(&mgStatus, STATUS_RUNNING, STATUA_NOT_RUNNING) {
		<-mgExitChan
		infof("stop mg go routine")
	}
}

func mg() {
	for mgStatus == STATUS_RUNNING {
		time.Sleep(time.Second)

		// mg spill
		if spillSeQueue.len() != 0 && 1.0*float32(spillSeQueue.len())/float32(spillSeQueue.cap())-0.5 > 0.0 {
			pauseReverseSpill()
			startSpill()
			// if spill, do not check reverse spill
			continue
		}
		if CurrentTimeMillis()-lastReverseSpillTs > int64(LingerSec*60*1000) {
			lastReverseSpillTs = CurrentTimeMillis()
			pauseSpill()
			startReverseSpill()
		}
	}

	mgExitChan <- 1
}

func ReportMetric(w http.ResponseWriter, req *http.Request) {
	metric := "{"
	// in-bound speed
	// reset value
	atomic.StoreUint64(&atomicSpanInBoundCount, 0)
	// wait 1's for add
	time.Sleep(time.Second)
	inBound := atomic.LoadUint64(&atomicSpanInBoundCount)

	// consumers
	outBound := float64(0)
	metric += "\"consumers\": ["
	for i := 0; i < ConsumerClientNum; i++ {
		if consumerClients[i].cStatus == STATUS_RUNNING {
			outBound += consumerClients[i].lastDeliverySpeed
		}

		metric += fmt.Sprintf("{\"cid\":%d, \"running\":%v, \"current_item_rate\":%f, \"current_items\": %d, \"last_delivery_speed(/s)\":%f},",
			i,
			consumerClients[i].cStatus == STATUS_RUNNING,
			1.0*float64(consumerClients[i].spanSeQueue.len())/float64(consumerClients[i].spanSeQueue.cap()),
			consumerClients[i].spanSeQueue.len(),
			consumerClients[i].lastDeliverySpeed)
	}
	metric = metric[0 : len(metric)-1]
	metric += "]"

	// speed
	metric += fmt.Sprintf(", \"speed\" : {\"inbound_speed\": %d, \"outbound_speed\":%f }", inBound, outBound)

	// consumer count / index
	metric += fmt.Sprintf(", \"consumer\" : {\"current_cid\":%d, \"current_cnum\":%d }", balanceIndex, current_consumer_count)
	/* gen
	metric += fmt.Sprintf(", \"gen\" : { \"output_rate\" : %f, \"output_items\" : %d, \"output_nil_count\" : %d, \"input_rate\" : %f, \"input_items\" : %d}",
		1.0*float64(genOutputRingBuffer.Len())/float64(genOutputRingBuffer.Cap()),
		genOutputRingBuffer.Len(),
		genOutputNilCount,
		1.0*float64(genInputSeQueue.len())/float64(genInputSeQueue.cap()),
		genInputSeQueue.len(),
	)
	*/

	// spill
	if spillStatus == STATUS_RUNNING {
		metric += fmt.Sprintf(", \"spill\": {\"enable\": true, \"items_rate\":%f, \"items\": %d }",
			1.0*float64(spillSeQueue.len())/float64(spillSeQueue.cap()),
			spillSeQueue.len())
	} else {
		metric += ", \"spill\": false"
	}
	// reverse spill
	if reverseSpillStatus == STATUS_RUNNING {
		state, _ := spillFD.Stat()
		metric += fmt.Sprintf(", \"reverse\": { \"enable\" : true, \"size\": %d }", state.Size())
	} else {
		metric += ", \"reverse\": false"
	}
	// mg
	if mgStatus == STATUS_RUNNING {
		metric += ", \"mg\":true"
	} else {
		metric += ", \"mg\":false"
	}

	// param
	metric += ", \"params\" : {"
	metric += fmt.Sprintf("\"con_num\" : %d, ", ConsumerClientNum)
	metric += fmt.Sprintf("\"dump_enable\" : %v, ", DumpEnable)
	// metric += fmt.Sprintf("\"dump_dir\" : \"%s\", ", DumpDir)
	metric += fmt.Sprintf("\"flush_retry_count\" : %d, ", FlushRetryCount)
	metric += fmt.Sprintf("\"deliver_enable\" : %v, ", DeliverEnable)
	metric += fmt.Sprintf("\"force_deliver\" : %v, ", ForceDeliver)
	metric += fmt.Sprintf("\"spill_enable\" : %v, ", SpillEnable)
	// metric += fmt.Sprintf("\"spill_dir\" : \"%s\", ", SpillDir)
	metric += fmt.Sprintf("\"max_spill_size\" : %d, ", MaxSpillContentSize)
	metric += fmt.Sprintf("\"buffer_size\" : %d, ", BuffSize)
	metric += fmt.Sprintf("\"batch_size\" : %d, ", BatchSize)
	metric += fmt.Sprintf("\"low_load_ts\" : %d ", LowLoadSleepTs)
	metric += "} "

	metric += "}"
	w.Write([]byte(metric))
}
