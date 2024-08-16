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
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

var (
	onceServiceName = ""
	onceIPUint32    = ""
	onceIpUint32Len = 0
	onceIPRune      = ""
	oncePort        string
	oncePortLen     = 0

	// atomic value for new trace id
	atomicTraceId uint64 = 0

	// tag zero for new trace id
	tagZero = []byte("00000000")

	// byte buffer pool for `ToString()`
	byteBufferPool = &sync.Pool{
		New: func() interface{} {
			var buffer bytes.Buffer
			buffer.Grow(16384)
			return &buffer
		},
	}
)

// Span define span as a fixed size `map[string]string`
type Span struct {
	// identifier for a trace, set on all items within it.
	_       [7]uint64
	traceId []byte

	// span name , rpc method for example.
	_    [7]uint64
	name string

	// identifier of this span within a trace.
	// Id []byte
	// span id for address & ts
	_        [7]uint64
	spanIdTs []byte

	// short span id for compute
	_               [7]uint64
	spanIdHierarchy []byte

	// tmpMeta for TraceId+spanIdTs+spanIdHierarchy
	_       [7]uint64
	tmpMeta []byte

	// epoch microseconds of the start of this span.
	_         [7]uint64
	timestamp int64

	// measurement in microseconds of the critical path.
	_        [7]uint64
	duration int64

	// annotations
	// annotations map[string]int64
	_          [7]uint64
	annoKeys   []string
	_          [7]uint64
	annoValues []int64

	// tags
	// tags map[string]string
	_         [7]uint64
	tagKeys   []string
	_         [7]uint64
	tagValues []string

	// span type.
	_        [7]uint64
	spanType int32

	// current child id
	_              [7]uint64
	currentChildId int32

	// serialize tmpSerializeBuf
	_               [7]uint64
	tmpSerializeBuf []byte

	// record key
	_            [7]uint64
	tmpRecordKey []byte
}

// setGlobalConfig for global value pre-set
func setGlobalConfig(serviceIP string, servicePort string, serviceName string) {
	// set the global value
	onceServiceName = serviceName
	onceIPRune = strconv.FormatInt(int64(IPv4toRune(serviceIP)), 10)
	onceIPUint32 = strconv.FormatInt(int64(IPv4toUint32(onceIPRune)), 16)
	onceIpUint32Len = len(onceIPUint32)
	oncePort = servicePort
	oncePortLen = len(oncePort)
}

// initSpan init a span from pool factory
func initSpan() interface{} {
	return &Span{
		traceId:         make([]byte, 0, 32),
		spanIdTs:        make([]byte, 0, 32),
		spanIdHierarchy: make([]byte, 0, 8),
		tmpMeta:         make([]byte, 0, 40),
		annoKeys:        make([]string, 0, 4),
		annoValues:      make([]int64, 0, 4),
		tagKeys:         make([]string, 0, 10),
		tagValues:       make([]string, 0, 10),
		tmpSerializeBuf: make([]byte, 0, 1024),
		tmpRecordKey:    make([]byte, 0, 48),
	}
}

// NewSpan Creates root span (default span type is SERVER).
// @param ip:port:serverName is your server deploy info,
// @param spanType span type, [server|client]
// @param abandon sample switch, true for sample, false for non-sample
func NewSpan(spanType int32, abandon bool) *Span {
	span := getSpan()

	if abandon {
		span.traceId[25] = 'a'
	} else {
		span.traceId[25] = 'n'
	}
	span.timestamp = CurrentTimeMicros()
	span.spanType = spanType
	span.currentChildId = 0
	span.spanIdHierarchy = append(span.spanIdHierarchy, '0')

	return span
}

// Next creates child span.
func (span *Span) Next(spanType int32) *Span {
	nextSpan := getSpan()
	if span == nil || nextSpan == nil {
		return nil
	}

	nextSpan.traceId = span.traceId
	nextSpan.timestamp = CurrentTimeMicros()
	nextSpan.spanType = spanType
	nextSpan.spanIdHierarchy = append(nextSpan.spanIdHierarchy, span.spanIdHierarchy...)
	nextSpan.spanIdHierarchy = append(nextSpan.spanIdHierarchy, '.')
	// TODO if < 10, may do 1+40 = '1'
	nextSpan.spanIdHierarchy = strconv.AppendInt(nextSpan.spanIdHierarchy, int64(atomic.AddInt32(&span.currentChildId, 1)), 10)

	return nextSpan
}

// FromMeta creates span from tmpMeta.
// @param tmpMeta tmpMeta info retrieve with rpc
// @param ip:port:serverName deploy server info
// @spanType span type, [server|client]
func FromMeta(meta string, spanType int32) *Span {
	// tmpMeta should like `c0a8380115174896423630105n008088#xxx`, at least > 34
	mLen := len(meta)
	if mLen < 34 {
		return nil
	}

	span := getSpan()
	if span == nil {
		return nil
	}

	span.tmpMeta = []byte(meta)
	// check traceId
	if mLen >= 32 {
		span.traceId = span.tmpMeta[0:32]
	} else {
		errorf("invalid tmpMeta traceId in %s", span.tmpMeta)
		return nil
	}
	// check spanIdTs
	if mLen > 33 && mLen >= 33+int(span.tmpMeta[32]) {
		span.spanIdTs = span.tmpMeta[33 : 33+int(span.tmpMeta[32])]
	} else {
		errorf("invalid spanIdTs in %s", span.tmpMeta)
		return nil
	}
	// check spanHierarchy
	if mLen >= 33+int(span.tmpMeta[32]) {
		span.spanIdHierarchy = span.tmpMeta[33+int(span.tmpMeta[32]):]
	} else {
		errorf("invalid spanIdHierarchy in %s", span.tmpMeta)
		return nil
	}

	span.timestamp = CurrentTimeMicros()
	span.spanType = spanType
	span.currentChildId = 0

	return span
}

// WithName set span name(rpc method name).
func (span *Span) WithName(name string) *Span {
	if span == nil {
		return nil
	}

	span.name = name
	return span
}

// Start records the start timestamp of rpc span.
func (span *Span) Start() *Span {
	if span == nil {
		return nil
	}

	span.timestamp = CurrentTimeMicros()
	span.annoKeys = append(span.annoKeys, getStartAnnatationType(span.spanType))
	span.annoValues = append(span.annoValues, span.timestamp)
	return span
}

// End records the duration of rpc span.
func (span *Span) End() *Span {
	if span == nil {
		return nil
	}

	ts := CurrentTimeMicros()
	span.duration = ts - span.timestamp
	span.annoKeys = append(span.annoKeys, getEndAnnatationType(span.spanType))
	span.annoValues = append(span.annoValues, ts)
	return span
}

// Send records the message send timestamp of mq span.
func (span *Span) Send() *Span {
	if span == nil {
		return nil
	}

	span.timestamp = CurrentTimeMicros()
	span.duration = 0
	span.annoKeys = append(span.annoKeys, getStartAnnatationType(span.spanType))
	span.annoValues = append(span.annoValues, span.timestamp)
	return span
}

// Recv records the message receive timestamp of mq span.
func (span *Span) Recv() *Span {
	if span == nil {
		return nil
	}

	span.timestamp = CurrentTimeMicros()
	span.duration = 0
	span.annoKeys = append(span.annoKeys, getEndAnnatationType(span.spanType))
	span.annoValues = append(span.annoValues, span.timestamp)
	return span
}

// WithTag set custom tag.
func (span *Span) WithTag(key string, value string) *Span {
	if span == nil {
		return nil
	}

	span.tagKeys = append(span.tagKeys, key)
	span.tagValues = append(span.tagValues, value)
	return span
}

// WithRetTag set ret
func (span *Span) WithRetTag(value string) *Span {
	if span == nil {
		return nil
	}

	span.tagKeys = append(span.tagKeys, "ret")
	span.tagValues = append(span.tagValues, value)
	return span
}

// WithErrorTag set error
func (span *Span) WithErrorTag(value string) *Span {
	if span == nil {
		return nil
	}

	span.tagKeys = append(span.tagKeys, "error")
	span.tagValues = append(span.tagValues, value)
	return span
}

// WithLocalComponent set local component.
func (span *Span) WithLocalComponent() *Span {
	if span == nil {
		return nil
	}

	span.tagKeys = append(span.tagKeys, "lc")
	span.tagValues = append(span.tagValues, "true")
	return span
}

// WithClientAddr set client address.
func (span *Span) WithClientAddr() *Span {
	if span == nil {
		return nil
	}

	span.tagKeys = append(span.tagKeys, "ca")
	span.tagValues = append(span.tagValues, "true")
	return span
}

// WithServerAddr set server address.
func (span *Span) WithServerAddr() *Span {
	if span == nil {
		return nil
	}

	span.tagKeys = append(span.tagKeys, "sa")
	span.tagValues = append(span.tagValues, "true")
	return span
}

// WithMessageAddr set message address.
func (span *Span) WithMessageAddr() *Span {
	if span == nil {
		return nil
	}

	span.tagKeys = append(span.tagKeys, "ma")
	span.tagValues = append(span.tagValues, "true")
	return span
}

// WithDescf set desc with format, `fmt.Sprintf` will cause performance
// Deprecated: this function will cause performance duo to `fmt.Printf()`
func (span *Span) WithDescf(format string, values ...interface{}) *Span {
	if span == nil {
		return nil
	}

	return span
}

// public property functions
// Meta gets tmpMeta string, format: <traceId>#<id>.
func (span *Span) Meta() string {
	if span == nil {
		return ""
	}

	span.tmpMeta = span.tmpMeta[:0]
	span.tmpMeta = append(span.tmpMeta, span.traceId...)
	span.tmpMeta = append(span.tmpMeta, byte(len(span.spanIdTs)))
	span.tmpMeta = append(span.tmpMeta, span.spanIdTs...)
	span.tmpMeta = append(span.tmpMeta, span.spanIdHierarchy...)
	return string(span.tmpMeta)
}

// ToString convert to string in json.
func (span *Span) ToString() string {
	if span == nil {
		return ""
	}

	buffer := byteBufferPool.Get().(*bytes.Buffer)

	// basic fields
	buffer.WriteString("{")
	buffer.WriteString("\"traceId\":\"" + string(span.traceId) + "\",")
	buffer.WriteString("\"name\":\"" + span.name + "\",")
	buffer.WriteString("\"id\":\"" + string(span.spanIdTs) + string(span.spanIdHierarchy) + "\",")
	buffer.WriteString("\"timestamp\":" + strconv.FormatInt(span.timestamp, 10) + ",")
	buffer.WriteString("\"duration\":" + strconv.FormatInt(span.duration, 10) + ",")

	// annotations
	buffer.WriteString("\"annotations\":[")
	annoBuffer := byteBufferPool.Get().(*bytes.Buffer)
	for i, _ := range span.annoKeys {
		annoBuffer.WriteString("{\"timestamp\":")
		annoBuffer.WriteString(strconv.FormatInt(span.annoValues[i], 10))
		annoBuffer.WriteString(", \"value\":\"")
		annoBuffer.WriteString(span.annoKeys[i])
		annoBuffer.WriteString("\", \"endpoint\":{\"serviceName\":\"")
		annoBuffer.WriteString(onceServiceName)
		annoBuffer.WriteString("\", \"ip\":")
		annoBuffer.WriteString(onceIPRune)
		annoBuffer.WriteString(", \"port\":")
		annoBuffer.WriteString(oncePort)
		annoBuffer.WriteString("}},")
	}
	annotationStr := annoBuffer.String()
	annoBuffer.Reset()
	byteBufferPool.Put(annoBuffer)
	if len(annotationStr) > 0 {
		buffer.WriteString(annotationStr[:len(annotationStr)-1])
	}
	buffer.WriteString("],")

	// tags
	buffer.WriteString("\"tags\":[")
	tagBuffer := byteBufferPool.Get().(*bytes.Buffer)
	for i, _ := range span.tagKeys {
		tagBuffer.WriteString("{\"key\":\"")
		tagBuffer.WriteString(strings.Replace(span.tagKeys[i], "\"", "\\\"", -1))
		tagBuffer.WriteString("\", \"value\":\"")
		tagBuffer.WriteString(strings.Replace(span.tagValues[i], "\"", "\\\"", -1))
		tagBuffer.WriteString("\", \"endpoint\":{\"serviceName\":\"")
		tagBuffer.WriteString(onceServiceName)
		tagBuffer.WriteString("\", \"ip\":")
		tagBuffer.WriteString(onceIPRune)
		tagBuffer.WriteString(", \"port\":")
		tagBuffer.WriteString(oncePort)
		tagBuffer.WriteString("}},")
	}
	tagStr := tagBuffer.String()
	tagBuffer.Reset()
	byteBufferPool.Put(tagBuffer)
	if len(tagStr) > 0 {
		buffer.WriteString(tagStr[:len(tagStr)-1])
	}
	buffer.WriteString("]}")

	result := buffer.String()
	buffer.Reset()
	byteBufferPool.Put(buffer)
	return result
}

// dump span to file.
// ${dir}/${traceId}_${spanId}
func (span *Span) dump(dir string) {
	filename := dir + string(os.PathSeparator) + string(span.traceId) + "_" + string(span.spanIdTs) + string(span.spanIdHierarchy)
	fp, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)

	if err != nil {
		errorf("open %s failed: %v\n", filename, err)
		return
	}

	// Dump span
	fmt.Fprintln(fp, span.ToString())
	fp.Close()
}

// ==== other functions ====
// gets annotation type while rpc start or message send.
func getStartAnnatationType(spanType int32) string {
	if spanType == CLIENT {
		return CLIENT_SEND
	} else if spanType == SERVER {
		return SERVER_RECV
	} else {
		return MESSAGE_SEND
	}
}

// gets annotation type while rpc end or message receive.
func getEndAnnatationType(spanType int32) string {
	if spanType == CLIENT {
		return CLIENT_RECV
	} else if spanType == SERVER {
		return SERVER_SEND
	} else {
		return MESSAGE_RECV
	}
}
