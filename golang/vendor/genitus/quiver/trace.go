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
	"github.com/linkedin/goavro"
	"os"
	"runtime/debug"
	"sync"
	"sync/atomic"
)

var (
	// true if dump log to file.
	DumpEnable bool = false
	// log dump dir.
	DumpDir string = "." + string(os.PathSeparator) + "event"
	// true if deliver log to flume.
	DeliverEnable bool = true
	// retry count
	FlushRetryCount int = 5
)

var (
	// consumer clients
	consumerClients []*consumer = nil
	// default flume client number
	consumerClientNum = 4
	// codec is stateless and is safe to use by multiple go routines.
	codec, _ = goavro.NewCodec(EVENT_SCHEMA)

	// wait for buffer drain
	globalWg *sync.WaitGroup
	// global init check with cas
	atomicInitInteger int32 = 0
)

// initialize plunger.
// @param host:port determine the flume host:ip
// @param num number of backend-data-consumer in Flange, adjust with upstream load
func Init(flumeHost string, flumePort string, consumerNum int,
		s3AccessKey string, s3SecretKey string, s3Endpoint string,
		hbaseZKHosts string) error {
	// check init already
	if !atomic.CompareAndSwapInt32(&atomicInitInteger, 0, 1) {
		infof("already init flange, ignore this operation")
		return nil
	}

	// init s3 info
	InitS3Info(s3AccessKey, s3SecretKey, s3Endpoint)
	ZK_HOSTS = hbaseZKHosts

	// mkdir dump dir
	if DumpEnable && DumpDir != "" {
		infof("mkdir dumpdir")
		if err := os.MkdirAll(DumpDir, 0755); err != nil {
			errorf("mkdir dumpdir err : %v", err)
			return err
		}
	}

	// init staff
	globalWg = &sync.WaitGroup{}
	consumerClientNum = consumerNum
	for i := 0; i < consumerClientNum; i++ {
		consumerClients = append(consumerClients, initConsumer(flumeHost, flumePort, i))
	}
	// init spill info, not start, just wait for mg_center
	initSpillProf(flumeHost, flumePort)
	initSDKHeader(flumeHost, flumePort)

	return nil
}

// fini plunger
func Fini() {
	defer catch("finish")

	// stop all component
	for _, consumer := range consumerClients {
		consumer.stop()
	}
	stopSpillReverse()
	globalWg.Wait()
}

// global error catch function
func catch(site string) {
	if err := recover(); err != nil {
		errorf("Error occur [%v] at [%s] with \n%s", err, site, string(debug.Stack()))
	}
}
