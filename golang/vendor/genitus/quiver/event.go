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
	"fmt"
	"encoding/json"
	"os"
	"errors"
	"sync/atomic"
)

// create a new event
func newEvent(eventType int64, sid string) *EventData {
	event := &EventData{
		Type:      eventType,
		Sid:       sid,
		Timestamp: CurrentTimeMillis()}
	event.Tags = make(map[string]string)
	event.Outputs = make(map[string] []string)
	return event
}

// create a new event with serverName and port
func NewEventWithNamePort(eventType int64, sid string, serviceName string, port string) *EventData {
	onceType = eventType
	oncePort = port
	onceServiceName = serviceName

	var eventData = newEvent(eventType, sid)
	return eventData.
		Tag(KV{"port", port}).
			Tag(KV{"serviceName", serviceName})
}

// update sid
func (event *EventData) WithSid(sid string) *EventData {
	event.Sid = sid
	return event
}

// add uid
func (event *EventData) WithUid(uid string) *EventData {
	event.Uid = uid
	return event
}

// add syncid
func (event *EventData) WithSyncId(syncId int64) *EventData {
	event.Syncid = syncId
	return event
}

// add sub
func (event *EventData) WithSub(sub string) *EventData {
	event.Sub = sub
	return event
}

// add ttl
func (event *EventData) WithTTL(dataCleanerType int32, dataCleanerTTL int64) *EventData {
	event.Tag(KV{HBASE_TTL_KEY, dataCleanerType})
	event.Tag(KV{HBASE_TTL_VALUE_KEY, dataCleanerTTL})
	return event
}

// add name
func (event *EventData) WithName(name string) *EventData {
	event.Name = name
	return event
}

// add endpoint
func (event *EventData) WithEndpoint(endpoint string) *EventData {
	onceIP = endpoint

	event.Endpoint = endpoint
	return event
}

// add tag value
func (event *EventData) Tag(tag KV) *EventData {
	var value = ""
	switch tag.Value.(type) {
	case bool:
		if tag.Value == true {
			value = "true"
		} else {
			value = "false"
		}
	case int, int64, int32, int16, int8, uint8, uint16, uint32, uint64, uint:
		value = fmt.Sprintf("%d", tag.Value)
	case float32, float64:
		value = fmt.Sprintf("%f", tag.Value)
	case string:
		value = fmt.Sprintf("%s", tag.Value)
	default:
		errorf("unsupported tag type %v", tag.Value)
	}
	event.Tags[tag.Key] = value
	return event
}

// add a ds, default will be `vagus`
func (event *EventData) TagDS(ds string) *EventData {
	return event.Tag(KV{"ds", ds})
}

// add outputs
func (event *EventData) Output(key string, value string) *EventData {
	event.Outputs[key] = append(event.Outputs[key], value)
	return event
}

// add desc
func (event *EventData) Desc(desc string) *EventData {
	event.Descs = append(event.Descs, desc)
	return event
}

// add desc with format
func (event *EventData) Descf(format string, values ...interface{}) *EventData {
	event.Descs = append(event.Descs, fmt.Sprintf(format, values...))
	return event
}

// add media
func (event *EventData) Media(key string, typer string, header string, data []byte) *EventData {
	event.Medias = append(event.Medias, &Medium{
		Key:   key,
		Type: typer,
		Header: header,
		Data:     data})
	return event
}

// convert to string in json
func (event *EventData) ToString() string {
	if res, err := json.MarshalIndent(event, "", "  "); err == nil {
		return string(res)
	} else {
		errorf("marshal event data error with :%v", err)
		return ""
	}
}

// dump to file
// ${dir}/${endpoint}_${metric}
func (event *EventData) dump(dir string) {
	filename := fmt.Sprintf("%s%s%s", dir, string(os.PathSeparator), event.Sid)
	if fp, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666); err == nil {
		// Dump
		fmt.Fprintln(fp, event.ToString())
		fp.Close()
	} else {
		errorf("open dump file %s failed with error %v", filename, err)
	}
}

// flush dslog to buffer channel, will be send to flume.
func (event *EventData) Flush() error {
	defer catch("flush")

	// check sid
	if len(event.Sid) < 32 {
		errorf("sid is invalid, sid=%s", event.Sid)
		return errors.New("sid is invalid, ignore this unvalid event")
	}

	// dump to file
	if DumpEnable {
		event.dump(DumpDir)
	}

	debugf("delivery enable = %v", DeliverEnable)
	// save to buffer only while plunger is enabled
	if DeliverEnable {
		for i := 0; i < FlushRetryCount; i++ {
			// compute balanceIndex each time for fixed value failed
			balanceIndex := int(CurrentTimeMillis() % int64(consumerClientNum))
			debugf("current balanceIndex = %d", balanceIndex)

			select {
			case consumerClients[balanceIndex].eventBufferChannel <- event:
				debugf("flush to consumer[%d] successful", balanceIndex)
				atomic.AddInt64(&flushSuccessGauge, 1)
				return nil
			default:
				// nothing to do, but continue
			}
		}

		// add span to spill
		if SpillEnable {
			select {
			case spillEventChannel <- event:
				atomic.AddInt64(&flushSpillGauge, 1)
				return nil
			default:
				// nothing to do, but exit
			}
		}

		// drop this at the end
		atomic.AddInt64(&flushFailedGauge, 1)
		infof("drop a event due to no capacity to save, must exceed `BuffSize` or `ConsumerNum` to cover all metric")
	}

	return nil
}
