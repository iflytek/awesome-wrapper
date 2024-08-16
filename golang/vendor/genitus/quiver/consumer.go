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
	"flume"
	"strconv"
	"sync/atomic"
	"time"
)

var (
	// log buffer size.
	BuffSize int32 = 20000
	// log send batch size.
	// The plunger will attempt to batch records together into fewer requests.
	// This helps performance on both the client and the trocar server.
	// A small batch size will make batching less common and may reduce throughput
	// A very large batch size may use memory a bit more wastefully and get over framesize error.
	BatchSize int = 100
	// linger micro seconds.
	// if have fewer than many records accumulated for batch sends, plunger will 'linger'
	// for the specified time waiting for more records to show up.
	LingerSec int = 5
	// kafka topic for tracar kafka sink routing
	Topic string = MESSAGE_QUEUE_KAFKA_TOPIC_VALUE

	// hbase & s3 switch delay time, per day
	DelayTime int = 1
)

// consumer consume event channel and send to thrift batched
type consumer struct {
	// flume client
	flumeClient *flumeClient
	// s3 client
	s3Client *s3Util
	// hbase client
	hbaseClient *hbaseUtil

	// s3 vs hbase delay time
	delayTimeMills int64
	// event cache buffer
	eventBufferChannel chan *EventData
	// temp span batch
	eventQueue [] *EventData
	// temp flumeEvent batch
	evBatches [] *flume.ThriftFlumeEvent

	// exit channel
	exitChannel chan int
	// flume client index for balance index
	consumerIndex int
	// last deliver timestamp
	lastDeliveryTs int64
}

// initConsumer init a consumer with flume host&port
func initConsumer(flumeHost string, flumePort string, index int) *consumer {
	c := &consumer{}

	// init index
	c.consumerIndex = index
	// init flume client
	if flumeHost == "" || flumePort == "" {
		infof("no flume agent address provided, will use default - %d", c.consumerIndex)
		flumeHost = FLUME_DEFAULT_HOST
		flumePort = FLUME_DEFAULT_PORT
	}
	c.flumeClient = &flumeClient{
		host: flumeHost,
		port: flumePort,
	}

	// init stuff
	c.exitChannel = make(chan int)
	c.lastDeliveryTs = CurrentTimeMillis()
	c.delayTimeMills = (time.Duration(DelayTime*24) * time.Hour).Nanoseconds() / 1000 / 1000
	// init util
	c.hbaseClient = NewHbaseUtil()
	c.s3Client = NewS3Util()
	c.eventBufferChannel = make(chan *EventData, BuffSize)
	c.eventQueue = make([] *EventData, 0, BatchSize)
	c.evBatches = make([] *flume.ThriftFlumeEvent, BatchSize, BatchSize)
	for i := 0; i < BatchSize; i++ {
		ev := &flume.ThriftFlumeEvent{Headers: make(map[string]string, 8)}
		ev.Headers[MESSAGE_QUEUE_KAFKA_TOPIC_KEY] = Topic
		ev.Headers[SCHEMA_VERSION_KEY] = DEFAULT_SCHEMA_VERSION_VALUE
		ev.Headers[SCHEMA_NAME_KEY] = DEFAULT_SCHEMA_NAME_VALUE
		c.evBatches[i] = ev
	}

	// start consumer just at init it
	globalWg.Add(1)
	go c.appendDataGoroutine()

	return c
}

// stop consumer and flume client
// may not used
func (c *consumer) stop() {
	if c.exitChannel != nil {
		c.exitChannel <- 1
	}
}

// append metrics from buffer to queue,
// if reach at Batch or time limit then send.
func (c *consumer) appendDataGoroutine() {
	defer catch("append event error")
	defer globalWg.Done()

	infof("start event append go routine - %d", c.consumerIndex)

	// init flume client in append go routine
	c.flumeClient.open()

	// deal-loop to consumer
	ticker := time.NewTicker(time.Second * time.Duration(LingerSec))
	defer ticker.Stop()

	for {
		// retrieve a event/exit/timeout operation
		select {
		case <-ticker.C:
			if len(c.eventQueue) == 0 {
				// close hbase due to gohbase deal-loop check
				c.hbaseClient.Close()
				debugf("sleep %d second, wait for metric arriving - %d", LingerSec, c.consumerIndex)
			} else {
				debugf("sleep %d second, with metric queue size = %d - %d", LingerSec, len(c.eventQueue), c.consumerIndex)
				c.sendData()
			}
		case evTask := <-c.eventBufferChannel:
			debugf("consumer consume a event to buffer queue")
			c.eventQueue = append(c.eventQueue, evTask)
		case <-c.exitChannel:
			goto FiniDo
		}

		if len(c.eventQueue) >= BatchSize {
			debugf("event queue >= batch size, will try to send data")
			c.sendData()
		}
	}

FiniDo:
	debugf("==========consumer exit==========")
	// check buffer channel
	select {
	case item := <-c.eventBufferChannel:
		c.eventQueue = append(c.eventQueue, item)

		if len(c.eventQueue) >= BatchSize || CurrentTimeMillis()-c.lastDeliveryTs > int64(LingerSec*1000) {
			c.sendData()
		}
	default:
		// nothing to do but exit
	}

	// check event queue
	for len(c.eventQueue) > 0 {
		c.sendData()
	}

	// exit
	if c.flumeClient != nil && c.flumeClient.transport != nil {
		c.flumeClient.transport.Close()
	}
	debugf("exit consumer append go routine - %d", c.consumerIndex)
}

//sendData send flume batch with queue
func (c *consumer) sendData() {
	defer catch("send event logs error")
	debugf("try to send data to flume")

	begin := CurrentTimeMillis()
	bound := Min(BatchSize, len(c.eventQueue))

	debugf("current send batch bound = %d", bound)
	for i := 0; i < bound; i++ {
		debugf("prepare a flume event to send")
		// retrieve event from event queue
		event := c.eventQueue[i]

		// media dispatch storage
		sidTs, err := strconv.ParseInt(event.Sid[14:25], 16, 64)
		if err != nil {
			errorf("parse sid timestamp error : %v", err)
			continue
		}

		// add support for multi-media batch upload, since v0.3.12
		debugf("event timestamp tag = %v", CurrentTimeMillis()-sidTs < c.delayTimeMills)
		if event.Tags["source"] == "aiaas" {
			// if event.tag["source"] == aiaas, then send to hbase
			if CurrentTimeMillis()-sidTs < c.delayTimeMills {
				debugf("tag.source=aiaas will send media data to hbase")
				// send to hbase, will delete media content
				err = c.hbaseClient.UploadMedia(event)
			} else {
				debugf("tag.source=aiaas will send media data to oss")
				// send to s3, will delete media content
				err = c.s3Client.UploadMedia(event)
			}
		} else {
			// if event.tag["source"] == aipaas/others_value, then send to oss
			if len(event.Medias) >= 2 {
				debugf("tag.source=!aiaas will send media data to oss by batch")
				err = c.s3Client.UploadMediaByBatch(event)
			} else if len(event.Medias) == 1 && len(event.Medias[0].Data) >= QUIVER_MULTI_MEDIA_DISPATCH_SIZE {
				debugf("tag.source=!aiaas will send media data to oss by batch")
				err = c.s3Client.UploadMediaByBatch(event)
			} else if CurrentTimeMillis()-sidTs < c.delayTimeMills {
				debugf("tag.source=!aiaas will send media data to hbase")
				// send to hbase, will delete media content
				err = c.hbaseClient.UploadMedia(event)
			} else {
				debugf("tag.source=!aiaas will send media data to oss")
				// send to s3, will delete media content
				err = c.s3Client.UploadMedia(event)
			}
		}

		if err != nil {
			errorf("hbase/oss error with %v, will check spill", err)
			// add to spill
			if SpillEnable {
				select {
				case spillEventChannel <- event:
					atomic.AddInt64(&flushSpillGauge, 1)
				default:
					// nothing to do, but exit
				}
			}
			// continue
		}

		// TODO prepare error with cause to s.v/s.n/s.t, which don't have body/r.k/flush.ts value

		// now, no media content
		buf, err := Serialize(event)
		if err != nil {
			errorf("flush serialize error : %v", err)
			continue
		}

		c.evBatches[i].Body = buf
		c.evBatches[i].Headers[TIMESTAMP] = strconv.FormatInt(event.Timestamp, 10)
		c.evBatches[i].Headers[MESSAGE_QUEUE_RECORD_KEY_KEY] = event.Sid
		c.evBatches[i].Headers[FLUSH_TIMESTAMP] = strconv.FormatInt(CurrentTimeMillis(), 10)
	}

	// make sure rpc client is open before send msg
	if !c.flumeClient.rpcClient.Transport.IsOpen() {
		errorf("flume client not open, now reset it - %d", c.consumerIndex)
		c.flumeClient.close()
		c.flumeClient.transport = nil
		c.flumeClient.rpcClient = nil
		c.flumeClient.open()
		// reset metric queue for OOM
		infof("reset event queue size due to flume client not open for OOM - %d", c.consumerIndex)
		c.eventQueue = c.eventQueue[bound: ]
		return
	}

	// send to flume client
	rs, err := c.flumeClient.rpcClient.AppendBatch(c.evBatches[:bound])

	if err != nil {
		errorf("append batch error : %v - %d", err, c.consumerIndex)
		c.flumeClient.close()
		c.flumeClient.transport = nil
		c.flumeClient.rpcClient = nil
		c.flumeClient.open()
		// reset metric queue for OOM
		infof("reset event queue size due to flume client not open for OOM - %d", c.consumerIndex)
		c.eventQueue = c.eventQueue[bound: ]
		return
	} else {
		if rs != flume.Status_OK {
			errorf("send %d events to flume failed : %v, will try again later - %d", bound, err, c.consumerIndex)
			// reset metric queue for OOM
			infof("reset event queue size due to flume client not open for OOM - %d", c.consumerIndex)
			c.eventQueue = c.eventQueue[bound: ]
		} else {
			// clean batch from logs which has been sent successfully
			atomic.AddInt64(&consumerSendGauge, int64(bound))

			infof("send %d events to flume successfully, use %d ms - %d", bound, CurrentTimeMillis()-begin, c.consumerIndex)
			c.eventQueue = c.eventQueue[bound: ]
			c.lastDeliveryTs = CurrentTimeMillis()
		}
	}
}
