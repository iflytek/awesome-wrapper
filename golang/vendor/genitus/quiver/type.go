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

//Service type
const (
	// Business gateway(Speech Interface Service)
	TYPE_SIS = 0
	// IAT engine
	TYPE_IAT = 1
	// TTS engine
	TYPE_TTS = 2
	// translation engine
	// TYPE_ITS = 3
	// speech-evaluation engine
	// TYPE_ISE = 4
	// SIS performance
	TYPE_SIS_PERF = 1000
	// IAT performance
	TYPE_IAT_PERF = 1001
	// TTS performance
	TYPE_ITS_PERF = 1002
)

// RPC settings
const (
	// default flume agent host
	FLUME_DEFAULT_HOST = "127.0.0.1"
	// default flume agent port
	FLUME_DEFAULT_PORT = "4545"
)

// Flume event headers' constants
const (
	// timestamp, need set to flume event, CAN NOT USE `ts`
	TIMESTAMP = "timestamp"
	// session id
	SID = "sid"
	// schema version key
	SCHEMA_VERSION_KEY = "s.v"
	// schema name key
	SCHEMA_NAME_KEY = "s.n"
	// division name key
	DIVISION_NAME_KEY = "d.n"
	// project name key
	PROJECT_NAME_KEY = "p.n"
	// default schema version
	DEFAULT_SCHEMA_VERSION_VALUE = "1"
	// default schema name
	DEFAULT_SCHEMA_NAME_VALUE = "event"
	// key of message queue(kafka) topic
	MESSAGE_QUEUE_KAFKA_TOPIC_KEY = "k.t"
	// message queue(kafka) default topic
	MESSAGE_QUEUE_KAFKA_TOPIC_VALUE = "event"
	// key of message queue key
	MESSAGE_QUEUE_RECORD_KEY_KEY = "r.k"
	// flush timestamp
	FLUSH_TIMESTAMP = "flush.ts"

	// bring from v0.3.13
	HBASE_TTL_KEY          = "data_cleaner_type"
	HBASE_TTL_WORK_VALUE   = 1
	HBASE_TTL_NOWORK_VALUE = 0

	HBASE_TTL_VALUE_KEY = "data_cleaner_ttl"

	// TODO remember to update this while version update
	// quiver version
	QUIVER_VERSION = "v0.3.14"

	// v0.3.12 add support for ase multi-media dispatch header
	QUIVER_MULTI_MEDIA_DISPATCH_HEADER = "QUIVER_MULTI_MEDIA_DISPATCH_HEADER"
	// v0.3.12 add support for ase multi-media dispatch fix size, default to 5M
	QUIVER_MULTI_MEDIA_DISPATCH_SIZE = 5 * 1024 * 1024

	// v0.3.12 add support for ase multi-media mediun type of [audio/image/video/text]
	MULTI_MEDIUM_TYPE_AUDIO = "audio"
	MULTI_MEDIUM_TYPE_IMAGE = "image"
	MULTI_MEDIUM_TYPE_VIDEO = "video"
	MULTI_MEDIUM_TYPE_TEXT  = "text"
)

// medium data struct
type Medium struct {
	// key
	Key string `json:"key" bson:"key"`
	// type
	Type string `json:"type" bson:"type"`
	// header
	Header string `json:"header" bson:"header"`
	// data
	Data []byte `json:"data" bson:"data"`
}

// Event type for business data
type EventData struct {
	// server type
	Type int64 `json:"type" bson:"type"`
	// sid
	Sid string `json:"sid" bson:"sid"`
	// uid
	Uid string `json:"uid" bson:"uid"`
	// syncid
	Syncid int64 `json:"syncid" bson:"syncid"`
	// sub
	Sub string `json:"sub" bson:"sub"`
	// timestamp
	Timestamp int64 `json:"timestamp" bson:"timestamp"`
	// name
	Name string `json:"name" bson:"name"`
	// endpoint
	Endpoint string `json:"endpoint" bson:"endpoint"`
	// tags
	Tags map[string]string `json:"tags" bson:"tags"`
	// outputs
	Outputs map[string][]string `json:"outputs" bson:"outputs"`
	// desc
	Descs []string `json:"descs" bson:"descs"`
	// media
	Medias []*Medium `json:"media" bson:"media"`
}

// metric tags k-v pair
type KV struct {
	// tag key
	Key string
	// tag value
	Value interface{}
}
