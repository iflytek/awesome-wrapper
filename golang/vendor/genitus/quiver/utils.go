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
	"time"
	"errors"
	"strconv"
	"fmt"
)

// Get millis timestamp.
func CurrentTimeMillis() int64 {
	return int64(time.Now().UnixNano() / int64(time.Millisecond))
}

// Get micros timestamp.
func CurrentTimeMicros() int64 {
	return int64(time.Now().UnixNano() / int64(time.Microsecond))
}

// Min
func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// Max
func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

// parse sid ts to table name or bucket name
func ParseSidTs2SaveName(sid string) (savedName string) {
	// 32 = len("ath********@bj***********####000")
	if len(sid) != 32 {
		errorf("invalid sid : %s for parse save name", sid)
		return ""
	}

	ts, err := strconv.ParseInt(sid[14:25], 16, 64)
	if err != nil {
		errorf("invalid sid : %s to parse time", sid)
		return ""
	}

	tsTime := time.Unix(ts/1000, 0)
	return fmt.Sprintf("quiver-%04d-%02d-%02d-%02d", tsTime.Year(), tsTime.Month(), tsTime.Day(), tsTime.Hour())
}

// Serialize event to bytes
func Serialize(event *EventData) ([]byte, error) {
	// set medium
	mediums := make([]interface{}, len(event.Medias))
	for i, media := range event.Medias {
		mediums[i] = map[string]interface{}{
			"key":   media.Key,
			"type": media.Type,
			"header": media.Header,
			"data":     media.Data}
	}

	// set base fields
	binary, err := codec.BinaryFromNative(nil, map[string]interface{}{
		"sid":       event.Sid,
		"type":      event.Type,
		"uid":       event.Uid,
		"syncid":    event.Syncid,
		"sub":       event.Sub,
		"timestamp": event.Timestamp,
		"name":      event.Name,
		"endpoint":  event.Endpoint,
		"tags":      event.Tags,
		"outputs": event.Outputs,
		"descs":     event.Descs,
		"media":     mediums})

	if err != nil {
		return []byte(""), errors.New("EVENT_SCHEMA error")
	}

	return binary, nil
}

// Deserialize convert binary data to event
func Deserialize(data []byte) *EventData {
	// convert binary data back to native go form
	eventMap, _, err := codec.NativeFromBinary(data)
	if err != nil {
		errorf("de-serialize record error : %v", err)
		return nil
	}

	eventMapInte := eventMap.(map[string]interface{})
	// prepare the return value
	event := &EventData{Tags: make(map[string]string), Outputs:make(map[string] []string)}
	// prepare the fields
	event.Type = int64(eventMapInte["type"].(int32))
	event.Sid = eventMapInte["sid"].(string)
	event.Uid = eventMapInte["uid"].(string)
	event.Syncid = int64(eventMapInte["syncid"].(int32))
	event.Sub = eventMapInte["sub"].(string)
	event.Timestamp = eventMapInte["timestamp"].(int64)
	event.Name = eventMapInte["name"].(string)
	event.Endpoint = eventMapInte["endpoint"].(string)
	for k, v := range eventMapInte["tags"].(map[string]interface{}) {
		event.Tags[k] = v.(string)
	}
	for k, v := range eventMapInte["outputs"].(map[string] interface{}) {
		for _, i := range v.([]interface{}) {
			event.Outputs[k] = append(event.Outputs[k], i.(string))
		}
	}
	for _, item := range eventMapInte["descs"].([]interface{}) {
		event.Descs = append(event.Descs, item.(string))
	}
	for _, item := range eventMapInte["media"].([]interface{}) {
		event.Medias = append(event.Medias, &Medium{
			Key:   item.(map[string]interface{})["key"].(string),
			Type: item.(map[string]interface{})["type"].(string),
			Header: item.(map[string]interface{})["header"].(string),
			Data:     item.(map[string]interface{})["data"].([]byte),
		})
	}

	return event
}
