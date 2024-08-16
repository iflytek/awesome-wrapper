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
	"flume"
	"sync/atomic"
)

var (
	BuffSize int32 = 100000
	// log send batch size.
	// The plunger will attempt to batch records together into fewer requests.
	// This helps performance on both the client and the trocar server.
	// A small batch size will make batching less common and may reduce throughput
	// A very large batch size may use memory a bit more wastefully and get over framesize error.
	BatchSize int = 1000
	// linger micro seconds.
	// if have fewer than many records accumulated for batch sends, plunger will 'linger'
	// for the specified time waiting for more records to show up.
	LingerSec int = 5
	// kafka topic for tracar kafka sink routing
	Topic string = MESSAGE_QUEUE_KAFKA_TOPIC_VALUE
)

type consumer struct {
	// flume client
	flumeClient *flumeClient

	// exit flag
	cStatus int32
	// exit channel
	exitConsumerChan chan int
	// flume client index for balance index
	consumerIndex int

	// last deliver timestamp
	lastDeliveryTs int64
	// delivery rate
	lastDeliverySpeed float64

	// span ring buffer
	spanSeQueue   *seQueue
	curBatchIndex int
	// temp span batch
	batches []interface{}
	// temp flumeEvent batch
	evBatches []*flume.ThriftFlumeEvent
}

func initConsumer(flumeHost string, flumePort string, index int) *consumer {
	c := &consumer{}
	// init flume client
	c.flumeClient = &flumeClient{
		host: flumeHost,
		port: flumePort,
	}
	atomic.StoreInt32(&c.cStatus, STATUA_NOT_RUNNING)
	c.exitConsumerChan = make(chan int)
	c.consumerIndex = index
	c.lastDeliveryTs = CurrentTimeMillis()

	c.spanSeQueue = newSeQueue(BuffSize)
	c.curBatchIndex = 0
	c.batches = make([]interface{}, BatchSize, BatchSize)
	c.evBatches = make([]*flume.ThriftFlumeEvent, BatchSize, BatchSize)
	for i := 0; i < BatchSize; i++ {
		ev := &flume.ThriftFlumeEvent{Headers: make(map[string]string, 8)}
		ev.Headers[MESSAGE_QUEUE_KAFKA_TOPIC_KEY] = Topic
		ev.Headers[SCHEMA_VERSION_KEY] = DEFAULT_SCHEMA_VERSION_VALUE
		ev.Headers[SCHEMA_NAME_KEY] = DEFAULT_SCHEMA_NAME_VALUE
		ev.Headers[SPAN_SERIALIZATION] = "false"
		c.evBatches[i] = ev
	}

	return c
}

func (c *consumer) start() {
	if atomic.CompareAndSwapInt32(&c.cStatus, STATUA_NOT_RUNNING, STATUS_RUNNING) {
		go c.appendData()
	}
}

func (c *consumer) pause() {
	if atomic.CompareAndSwapInt32(&c.cStatus, STATUS_RUNNING, STATUA_NOT_RUNNING) {
		<-c.exitConsumerChan
		infof("pause consumer append go routine : %d", c.consumerIndex)
	}
}

// call after pause
func (c *consumer) stop() {
	if atomic.LoadInt32(&c.cStatus) == STATUS_RUNNING {
		c.pause()
	} else {
		// if exit, then exit and flush the last items
		for item := c.spanSeQueue.get(); item != nil; item = c.spanSeQueue.get() {
			// check send condition before next assign
			if c.curBatchIndex >= BatchSize {
				c.sendData()
				// reset curBatchIndex no matter sendData() result
				c.curBatchIndex = 0
			}

			// temp batches for send
			c.batches[c.curBatchIndex] = item
			c.curBatchIndex++
		}

		// if batches have data
		if c.curBatchIndex > 0 {
			c.sendData()
		}
	}
	if c.flumeClient.transport != nil {
		c.flumeClient.transport.Close()
	}
	infof("stop consumer append go routine : %d", c.consumerIndex)
}

func (c *consumer) appendData() {
	defer catch("append data error")

	atomic.StoreInt32(&c.cStatus, STATUS_RUNNING)
	infof("start consumer append go routine : %d", c.consumerIndex)

	// runtime.LockOSThread()

	c.flumeClient.open()
	c.lastDeliveryTs = CurrentTimeMillis()
	c.curBatchIndex = 0

	for item := c.spanSeQueue.get(); atomic.LoadInt32(&c.cStatus) == STATUS_RUNNING && item != nil; item = c.spanSeQueue.get() {
		// check send condition before next assign
		if c.curBatchIndex >= BatchSize || (c.curBatchIndex > 0 && CurrentTimeMillis()-c.lastDeliveryTs > int64(LingerSec*1000)) {
			c.sendData()
			// reset curBatchIndex no matter sendData() result
			c.curBatchIndex = 0
		}

		// temp batches for send
		c.batches[c.curBatchIndex] = item
		c.curBatchIndex++
	}

	// if exit, then exit and flush the last items
	for item := c.spanSeQueue.get(); item != nil; item = c.spanSeQueue.get() {
		// check send condition before next assign
		if c.curBatchIndex >= BatchSize {
			c.sendData()
			// reset curBatchIndex no matter sendData() result
			c.curBatchIndex = 0
		}

		// temp batches for send
		c.batches[c.curBatchIndex] = item
		c.curBatchIndex++
	}

	// if batches have data
	if c.curBatchIndex > 0 {
		c.sendData()
	}

	// runtime.UnlockOSThread()

	infof("exit consumer append go routine : %d", c.consumerIndex)
	if atomic.CompareAndSwapInt32(&c.cStatus, STATUS_RUNNING, STATUA_NOT_RUNNING) {
		// if consumer exit by itself
	} else {
		// if consumer exit by other command
		c.exitConsumerChan <- 1
	}
}

func (c *consumer) sendData() {
	defer catch("sendData")
	begin := CurrentTimeMillis()

	for i := 0; i < c.curBatchIndex; i++ {
		// retrieve span info
		span := c.batches[i].(*Span)
		buf, err := Serialize(span, false)
		if err != nil {
			continue
		}

		c.evBatches[i].Body = buf
		// set headers, see java syringe.v2 FlumeClient
		// set timestamp to the millseconds of the start of this trace
		c.evBatches[i].Headers[TIMESTAMP] = string(span.traceId[8:21])
		// set r.k sid for mq balancing, format: <traceId>
		span.tmpRecordKey = span.tmpRecordKey[:0]
		span.tmpRecordKey = append(span.tmpRecordKey, span.traceId...)
		span.tmpRecordKey = append(span.tmpRecordKey, '#')
		span.tmpRecordKey = append(span.tmpRecordKey, span.spanIdHierarchy...)
		span.tmpRecordKey = append(span.tmpRecordKey, '#')
		switch span.spanType {
		case CLIENT, PRODUCER:
			span.tmpRecordKey = append(span.tmpRecordKey, 'c')
		case SERVER, CONSUMER:
			span.tmpRecordKey = append(span.tmpRecordKey, 's')
		default:
			span.tmpRecordKey = append(span.tmpRecordKey, '0')
		}
		c.evBatches[i].Headers[MESSAGE_QUEUE_RECORD_KEY_KEY] = string(span.tmpRecordKey)
		// fc.evBatches[i].Headers[FLUSH_TIMESTAMP] = strconv.FormatInt(CurrentTimeMillis(), 10)
	}

	// make sure rpc client is open before send msg
	if !c.flumeClient.rpcClient.Transport.IsOpen() {
		errorf("flume client not open, now reset it")
		c.flumeClient.close()
		c.flumeClient.transport = nil
		c.flumeClient.rpcClient = nil
		c.flumeClient.open()

		// release those data
		for i := 0; i < c.curBatchIndex; i++ {
			spillSeQueue.put((c.batches[i]).(*Span))
		}
		// reset cur batch index
		c.curBatchIndex = 0

		return
	}

	rs, err := c.flumeClient.rpcClient.AppendBatch(c.evBatches[:c.curBatchIndex])
	if err != nil {
		errorf("append batch error : %v\n", err)
		c.flumeClient.close()
		c.flumeClient.transport = nil
		c.flumeClient.rpcClient = nil
		c.flumeClient.open()

		// release those data
		for i := 0; i < c.curBatchIndex; i++ {
			spillSeQueue.put((c.batches[i]).(*Span))
		}
		// reset cur batch index
		c.curBatchIndex = 0
	} else {
		if rs != flume.Status_OK {
			debugf("cid:%d send %d spans to flume failed, will try again later.", c.consumerIndex, c.curBatchIndex)
			// release those data
			for i := 0; i < c.curBatchIndex; i++ {
				spillSeQueue.put((c.batches[i]).(*Span))
			}
			// reset cur batch index
			c.curBatchIndex = 0
		} else {
			c.lastDeliveryTs = CurrentTimeMillis()
			debugf("cid:%d send %d items to flume successfully, use %d ms.",
				c.consumerIndex, c.curBatchIndex, c.lastDeliveryTs-begin)

			sub := c.lastDeliveryTs - begin
			if sub != 0 {
				c.lastDeliverySpeed = float64(c.curBatchIndex) * 1000 / float64(sub)
			} else {
				c.lastDeliverySpeed = 0
			}

			c.curBatchIndex = 0
		}
	}
}
