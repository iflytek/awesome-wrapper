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
	"github.com/linkedin/goavro"
	"log"
	"math"
	"strconv"
	"strings"
)

////////////////////////////////////////
// Encode
////////////////////////////////////////
var (
	// MaxBlockCount is the maximum number of data items allowed in a single
	// block that will be decoded from a binary stream, whether when reading
	// blocks to decode an array or a map, or when reading blocks from an OCF
	// stream. This check is to ensure decoding binary data will not cause the
	// library to over allocate RAM, potentially creating a denial of service on
	// the system.
	//
	// If a particular application needs to decode binary Avro data that
	// potentially has more data items in a single block, then this variable may
	// be modified at your discretion.
	MaxBlockCount = int64(math.MaxInt32)

	// MaxBlockSize is the maximum number of bytes that will be allocated for a
	// single block of data items when decoding from a binary stream. This check
	// is to ensure decoding binary data will not cause the library to over
	// allocate RAM, potentially creating a denial of service on the system.
	//
	// If a particular application needs to decode binary Avro data that
	// potentially has more bytes in a single block, then this variable may be
	// modified at your discretion.
	MaxBlockSize = int64(math.MaxInt32)
)

// Serialize span to bytes
func Serialize(span *Span, prefix bool) ([]byte, error) {
	// defer catch("serialize error")
	// buffer to serialize
	// NOTICE data with format [c|s|0][len(spanId)spanId][avrobytes]
	// var tmpSerializeBuf []byte // := make([]byte, 0, 512)
	span.tmpSerializeBuf = span.tmpSerializeBuf[:0]

	// if prefix set, then add spanType + spanIdHierarchy before span avro bytes
	if prefix {
		switch span.spanType {
		case CLIENT, PRODUCER:
			span.tmpSerializeBuf = append(span.tmpSerializeBuf, 'c')
		case SERVER, CONSUMER:
			span.tmpSerializeBuf = append(span.tmpSerializeBuf, 's')
		default:
			span.tmpSerializeBuf = append(span.tmpSerializeBuf, '0')
		}
		// add len(spanId) and spanId
		idLen := len(span.spanIdHierarchy)
		span.tmpSerializeBuf = append(span.tmpSerializeBuf, byte(idLen))
		span.tmpSerializeBuf = append(span.tmpSerializeBuf, span.spanIdHierarchy...)
	}

	// tmpMeta fields
	// trace id with byte array
	span.tmpSerializeBuf, _ = longBinaryFromNative(span.tmpSerializeBuf, int64(len(span.traceId)))
	span.tmpSerializeBuf = append(span.tmpSerializeBuf, span.traceId...)

	span.tmpSerializeBuf, _ = stringBinaryFromNative(span.tmpSerializeBuf, span.name)

	// id with byte array
	span.tmpSerializeBuf, _ = longBinaryFromNative(span.tmpSerializeBuf, int64(len(span.spanIdTs)+len(span.spanIdHierarchy)))
	span.tmpSerializeBuf = append(span.tmpSerializeBuf, span.spanIdTs...)
	span.tmpSerializeBuf = append(span.tmpSerializeBuf, span.spanIdHierarchy...)

	span.tmpSerializeBuf, _ = longBinaryFromNative(span.tmpSerializeBuf, span.timestamp)
	span.tmpSerializeBuf, _ = longBinaryFromNative(span.tmpSerializeBuf, span.duration)

	// endpoint
	// fmt.Println(onceServiceName + "," + onceIPRune + "," + oncePort)
	edbuf, _ := stringBinaryFromNative(nil, onceServiceName)
	ipInt, _ := strconv.Atoi(onceIPRune)
	portInt, _ := strconv.Atoi(oncePort)
	edbuf, _ = intBinaryFromNative(edbuf, int32(ipInt))
	edbuf, _ = intBinaryFromNative(edbuf, int32(portInt))

	// annotations
	annolen := int64(len(span.annoKeys))
	var alreadyEncoded, remainingInBlock int64

	for i, _ := range span.annoKeys {
		if remainingInBlock == 0 { // start a new block
			remainingInBlock = annolen - alreadyEncoded
			if remainingInBlock > MaxBlockCount {
				// limit block count to MacBlockCount
				remainingInBlock = MaxBlockCount
			}
			span.tmpSerializeBuf, _ = longBinaryFromNative(span.tmpSerializeBuf, remainingInBlock)
		}

		// annotation
		span.tmpSerializeBuf, _ = longBinaryFromNative(span.tmpSerializeBuf, span.annoValues[i])
		span.tmpSerializeBuf, _ = stringBinaryFromNative(span.tmpSerializeBuf, span.annoKeys[i])
		span.tmpSerializeBuf = append(span.tmpSerializeBuf, edbuf...)

		remainingInBlock--
		alreadyEncoded++
	}

	span.tmpSerializeBuf, _ = longBinaryFromNative(span.tmpSerializeBuf, 0) // append trailing 0 block count to signal end of Array

	// tags
	taglen := int64(len(span.tagKeys))
	alreadyEncoded, remainingInBlock = 0, 0

	for i, _ := range span.tagKeys {
		if remainingInBlock == 0 { // start a new block
			remainingInBlock = taglen - alreadyEncoded
			if remainingInBlock > MaxBlockCount {
				// limit block count to MacBlockCount
				remainingInBlock = MaxBlockCount
			}
			span.tmpSerializeBuf, _ = longBinaryFromNative(span.tmpSerializeBuf, remainingInBlock)
		}

		// annotation
		// span.tmpSerializeBuf, _ = stringBinaryFromNative(span.tmpSerializeBuf, strings.Replace(span.tagKeys[i], "\"", "\\\"", -1))
		span.tmpSerializeBuf, _ = stringBinaryFromNative(span.tmpSerializeBuf, span.tagKeys[i])
		//span.tmpSerializeBuf, _ = stringBinaryFromNative(span.tmpSerializeBuf, strings.Replace(span.tagValues[i], "\"", "\\\"", -1))
		span.tmpSerializeBuf, _ = stringBinaryFromNative(span.tmpSerializeBuf, span.tagValues[i])
		span.tmpSerializeBuf = append(span.tmpSerializeBuf, edbuf...)

		remainingInBlock--
		alreadyEncoded++
	}

	span.tmpSerializeBuf, _ = longBinaryFromNative(span.tmpSerializeBuf, 0) // append trailing 0 block count to signal end of Array

	return span.tmpSerializeBuf, nil
}

// RetrieveSpanInfo with specificed serialize
func RetrieveSpanInfo(data []byte) (traceId string, spanId string, spanType string, sBuf []byte) {
	// NOTICE data with format [c|s|0][len(spanId)spanId][avrobytes]
	// get span id
	idLen := uint(data[1])
	spanId = string(data[2 : idLen+2])
	// get trace id
	traceId = string(data[idLen+3 : idLen+35])
	// get span serialize byte
	buf := data[idLen+2:]

	return traceId, spanId, string(data[0]), buf
}

// Deserialize bytes to span.
// Just for validating.
func Deserialize(data []byte) *Span {
	codec, err := goavro.NewCodec(SPAN_SCHEMA)
	if err != nil {
		log.Fatal("Deserialize failed.")
	}

	// Convert binary span data back to native Go form
	spanMap, _, err := codec.NativeFromBinary(data)
	// fmt.Println("span=", spanMap)
	if err != nil {
		errorf("de-serialize record error : %v", err)
		return nil
	}

	span := initSpan().(*Span)
	span.traceId = []byte(spanMap.(map[string]interface{})["traceId"].(string))
	span.name = spanMap.(map[string]interface{})["name"].(string)
	// span.Id = spanMap.(map[string]interface{})["id"].(string)
	span.tmpMeta = []byte(spanMap.(map[string]interface{})["id"].(string))
	hi := strings.Index(string(span.tmpMeta), ".")
	if hi < 0 {
		span.spanIdTs = span.tmpMeta[0 : len(span.tmpMeta)-1]
		span.spanIdHierarchy = span.tmpMeta[len(span.tmpMeta)-1:]
	} else {
		span.spanIdTs = span.tmpMeta[0 : hi-1]
		span.spanIdHierarchy = span.tmpMeta[hi-1:]
	}
	span.timestamp = spanMap.(map[string]interface{})["timestamp"].(int64)
	span.duration = spanMap.(map[string]interface{})["duration"].(int64)
	for _, item := range spanMap.(map[string]interface{})["annotations"].([]interface{}) {
		key := item.(map[string]interface{})["value"].(string)
		span.annoKeys = append(span.annoKeys, key)
		span.annoValues = append(span.annoValues, item.(map[string]interface{})["timestamp"].(int64))
		switch key {
		case CLIENT_SEND, CLIENT_RECV:
			span.spanType = CLIENT
		case SERVER_SEND, SERVER_RECV:
			span.spanType = SERVER
		}
	}
	for _, item := range spanMap.(map[string]interface{})["tags"].([]interface{}) {
		span.tagKeys = append(span.tagKeys, item.(map[string]interface{})["key"].(string))
		span.tagValues = append(span.tagValues, item.(map[string]interface{})["value"].(string))
	}

	return span
}

////////////////////////////////////////
// Binary Encode
////////////////////////////////////////

const (
	intDownShift  = uint32(31)
	intFlag       = byte(128)
	intMask       = byte(127)
	longDownShift = uint32(63)
)

func intBinaryFromNative(buf []byte, value int32) ([]byte, error) {
	encoded := uint64((uint32(value) << 1) ^ uint32(value>>intDownShift))
	return integerBinaryEncoder(buf, encoded)
}

func longBinaryFromNative(buf []byte, value int64) ([]byte, error) {
	encoded := (uint64(value) << 1) ^ uint64(value>>longDownShift)
	return integerBinaryEncoder(buf, encoded)
}

func integerBinaryEncoder(buf []byte, encoded uint64) ([]byte, error) {
	// used by both intBinaryEncoder and longBinaryEncoder
	if encoded == 0 {
		return append(buf, 0), nil
	}
	for encoded > 0 {
		b := byte(encoded) & intMask
		encoded = encoded >> 7
		if encoded != 0 {
			b |= intFlag // set high bit; we have more bytes
		}
		buf = append(buf, b)
	}
	return buf, nil
}

////////////////////////////////////////
// Binary Encode
////////////////////////////////////////

func bytesBinaryFromNative(buf []byte, datum []byte) ([]byte, error) {
	buf, _ = longBinaryFromNative(buf, int64(len(datum))) // only fails when given non integer
	return append(buf, datum...), nil                     // append datum bytes
}

func stringBinaryFromNative(buf []byte, datum string) ([]byte, error) {
	buf, _ = longBinaryFromNative(buf, int64(len(datum))) // only fails when given non integer
	return append(buf, datum...), nil                     // append datum bytes
}
