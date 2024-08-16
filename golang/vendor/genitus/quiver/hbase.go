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
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/tsuna/gohbase"
	"github.com/tsuna/gohbase/hrpc"
	"hash"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	HBASE_INTERVAL   = 100
	HBASE_QUEUE_SIZE = 5
	HBASE_DEF_TTL = 1

	ZK_HOSTS = "127.0.0.1"
	// generate by sid ts
	// TABLE_NAME    = "quiver"
	COLUMN_FAMILY = "r"
	QUALIFIER     = "a"

	// hbase_value_format = "0000"
	hbaseClientInitOnce sync.Once
	client gohbase.Client
)

type hbaseUtil struct {
	hasher hash.Hash
	buffer bytes.Buffer
}

// set logrus default output to DevNull, and hook to quiver logger
func init() {
	// add hook to quiver logger
	logrus.AddHook(&lHook{})
	// set output to os.DevNull
	nullFile, err := os.OpenFile(os.DevNull, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		errorf("can't open os.DevNull %v", err)
	} else {
		logrus.SetOutput(nullFile)
	}
}

// NewHbaseUtil create a new hbase client
func NewHbaseUtil() *hbaseUtil {
	hbaseUtil := &hbaseUtil{}
	hbaseClientInitOnce.Do(func() {
		client = gohbase.NewClient(
			ZK_HOSTS,
			gohbase.FlushInterval(time.Duration(HBASE_INTERVAL)*time.Millisecond),
			gohbase.RpcQueueSize(consumerClientNum/2),
		)
		go reconnectHBase()
	})
	hbaseUtil.hasher = md5.New()
	return hbaseUtil
}

// Close close a hbase client
func (this *hbaseUtil) Close() {

}

// UploadMedia update event data to hbase
func (this *hbaseUtil) UploadMedia(event *EventData) error {
	// check client connection
	if client == nil {
		errorf("hbase client is not init well.")
		return errors.New("hbase client is not init well")
	}

	// send each media to hbase
	tableName := ""
	if tableName = ParseSidTs2SaveName(event.Sid); tableName == "" {
		errorf("parse event sid error : invalid sid : " + event.Sid)
		return errors.New("parse event sid error : invalid sid : " + event.Sid)
	}
	debugf("parse event sid to table name : " + tableName)

	// compute sid md5 prefix
	this.hasher.Reset()
	this.hasher.Write([]byte(event.Sid))
	key := hex.EncodeToString(this.hasher.Sum(nil))
	// debugf("hbase key=%s", key)

	// process each media
	for _, mediaItem := range event.Medias {
		this.buffer.Reset()
		// save data with format `[header] \n [raw_media_bytes]`
		this.buffer.WriteString(mediaItem.Header)
		this.buffer.WriteString("\n")
		this.buffer.Write(mediaItem.Data)
		// remove origin event media data bytes
		mediaItem.Data = []byte(fmt.Sprintf("%d", len(mediaItem.Data)))

		// prepare hbase put request, buffer.String() for deep-copy
		values := map[string]map[string][]byte{COLUMN_FAMILY: {QUALIFIER: []byte(this.buffer.String())}}

		// key[0:8] may add key/type for classify, with hbase RowFilter
		var putRequest *hrpc.Mutate
		var err error
		ctx, cancel := context.WithTimeout(context.Background(), time.Second * time.Duration(2 * LingerSec))
		if event.Tags[HBASE_TTL_KEY] == strconv.Itoa(HBASE_TTL_WORK_VALUE) {
			if ttlValue, err := strconv.ParseInt(event.Tags[HBASE_TTL_VALUE_KEY], 10, 64); err == nil {
				putRequest, err = hrpc.NewPutStr(ctx, tableName, key[0:8]+event.Sid, values,
					hrpc.TimestampUint64(uint64(event.Timestamp)),
					hrpc.TTL(time.Millisecond * time.Duration(ttlValue - event.Timestamp)))
			} else if err != nil {
				errorf("retrieve hbase ttl error --> %v, or ttlValue < curMS.", err)
			}
		} else {
			putRequest, err = hrpc.NewPutStr(ctx, tableName, key[0:8]+event.Sid, values)
		}

		if err != nil {
			errorf("error with hbass new put str : %v", err)
			cancel()
			return errors.New("error with hbass new put str : " + err.Error())
		}

		_, err = client.Put(putRequest)
		cancel()
		if err != nil {
			errorf("error with hbase put : %v", err)
			return errors.New("error with hbase put : " + err.Error())
		}
		debugf("upload a event to hbase successful")

		atomic.AddInt64(&consumerHBaseSendGauge, 1)
	}

	return nil
}

// goroutine to check hbase connection, and re-connect if put error
func reconnectHBase() {
	defer catch("reconnect hbase error")

	for {
		time.Sleep(time.Minute * 5)

		if client == nil {
			errorf("hbase client is not init well.")
			continue
		}

		tableName := "quiver-" + time.Now().Format("2006-01-02-15")
		values := map[string]map[string][]byte{COLUMN_FAMILY: {QUALIFIER: []byte("hell")}}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second * 60)
		putRequest, err := hrpc.NewPutStr(ctx, tableName, tableName, values)
		if err != nil {
			errorf("error with hbass new put str in check goroutine: %v", err)
		}

		_, err = client.Put(putRequest)
		cancel()
		if err != nil && strings.Contains(err.Error(), "deadline exceeded") {
			errorf("error with hbase put in check goroutine: %v", err)

			client.Close()
			client = gohbase.NewClient(
				ZK_HOSTS,
				gohbase.FlushInterval(time.Duration(HBASE_INTERVAL)*time.Millisecond),
				gohbase.RpcQueueSize(consumerClientNum/2),
			)
		}
	}
}

// logrus hook to redirect log
type lHook struct {
}

func (this *lHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (this *lHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		errorf("unable to read entry %v", err)
		return err
	}

	switch entry.Level {
	case logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel:
		errorf(line)
	case logrus.InfoLevel:
		infof(line)
	case logrus.DebugLevel:
		debugf(line)
	}

	return nil
}
