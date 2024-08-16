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
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"bytes"
	"os"
	"errors"
	"sync/atomic"
)

var (
	sess *session.Session
)

// S3Util to operation
type s3Util struct {
	svc        *s3.S3
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader

	buffer bytes.Buffer
}

// init s3 info for new util
func InitS3Info(s3AccessKey string, s3SecretKey string, s3Endpoint string) error {
	// init S3 session
	creds := credentials.NewStaticCredentials(s3AccessKey, s3SecretKey, "")
	var err error
	sess, err = session.NewSession((&aws.Config{}).
			WithS3ForcePathStyle(true).
			WithDisableSSL(true).
			WithCredentials(creds).
			WithRegion("us-east-1").
			WithEndpoint(s3Endpoint))
	if err != nil {
		errorf("init s3 session failed : %v", err)
		return err
	}
	return nil
}

// NewS3Util create a new s3 util with pre-set info
func NewS3Util() *s3Util {
	return &s3Util{
		svc:        s3.New(sess),
		uploader:   s3manager.NewUploader(sess),
		downloader: s3manager.NewDownloader(sess),
	}
}

// UploadMedia update a event media data by the sid key
func (this *s3Util) UploadMedia(event *EventData) error {
	// send each media to hbase
	bucketName := ""
	if bucketName = ParseSidTs2SaveName(event.Sid); bucketName == "" {
		errorf("parse event sid error, invalid sid " + event.Sid)
		return errors.New("parse event sid error : invalid sid : " + event.Sid)
	}
	debugf("parse event sid to bucket name : " + bucketName)

	// process each media
	for _, mediaItem := range event.Medias {
		this.buffer.Reset()
		// save data with format `[header] \n [raw_media_bytes]`
		this.buffer.WriteString(mediaItem.Header)
		this.buffer.WriteString("\n")
		this.buffer.Write(mediaItem.Data)
		// remove origin event media data bytes
		mediaItem.Data = []byte(fmt.Sprintf("%d", len(mediaItem.Data)))

		// prepare s3 put request, buffer.String() for deep-copy
		br := bytes.NewReader([]byte(this.buffer.String()))
		// upload to s3
		if _, err := this.uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(bucketName),
			// event.Sid may add key/type for classify, with s3 getObject-prefix
			Key:    aws.String(event.Sid),
			Body:   br,
		}); err != nil {
			errorf("upload event to s3 failed with : %v", err)
			return errors.New("upload event to s3 failed with : " + err.Error())
		}
		debugf("upload a event to s3 successful")

		atomic.AddInt64(&consumerOSSSendGauge, 1)
	}

	return nil
}

// UploadMediaByBatch update event media data with batch by the sid key
func (this *s3Util) UploadMediaByBatch(event *EventData) error {
	bucketName := ""
	if bucketName = ParseSidTs2SaveName(event.Sid); bucketName == "" {
		errorf("parse event sid error, invalid sid " + event.Sid)
		return errors.New("parse event sid error : invalid sid : " + event.Sid)
	}
	debugf("parse event sid to bucket name : " + bucketName)

	// process media by batch
	mediasBuff, err := json.MarshalIndent(event.Medias, "", "  ")
	if err != nil {
		errorf("marshal event to json error with :%v", err)
		return errors.New("marshal event to json error with : " + err.Error())
	}

	this.buffer.Reset()
	// save data with form `[$QUIVER_MULTI_MEDIA_DISPATCH_HEADER] \n [batch_medias_json_bytes]`
	this.buffer.WriteString(QUIVER_MULTI_MEDIA_DISPATCH_HEADER)
	this.buffer.WriteString("\n")
	this.buffer.Write(mediasBuff)

	// prepare s3 put request, buffer.String() for deep-copy
	br := bytes.NewReader([]byte(this.buffer.String()))
	// upload to s3
	if _, err := this.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucketName),
		Key: aws.String(event.Sid),
		Body: br,
	}); err != nil {
		errorf("upload event to s3 failed with : %v", err)
		return errors.New("upload event to s3 failed with : " + err.Error())
	}

	debugf("upload a event to s3 successful")
	atomic.AddInt64(&consumerOSSSendGauge, 1)
	// remove origin event media data bytes
	for _, mediaItem := range event.Medias {
		mediaItem.Data = []byte(fmt.Sprintf("%d", len(mediaItem.Data)))
	}

	return nil
}

// DownloadMedia download media from specified downloader with sid, and return the aue & auf
func (this *s3Util) DownloadMedia(sid string, filePath string) error {
	bucketName := ""
	if bucketName = ParseSidTs2SaveName(sid); bucketName == "" {
		return errors.New("invalid sid " + sid)
	}

	file, err := os.Create(filePath + sid)
	if err != nil {
		errorf("failed to create file %v", err)
		return err
	}
	defer file.Close()

	// download s3 file
	if _, err := this.downloader.Download(file, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(sid),
	}); err != nil {
		errorf("failed to download item %v", err)
		return err
	}
	return nil
}