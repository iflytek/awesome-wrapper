package test

import (
	"testing"
	"genitus/quiver"
	"crypto/md5"
	"encoding/hex"
	"time"
	"fmt"
)

func Test_Md5(t *testing.T) {
	event := GetEvent()

	hash := md5.New()
	hash.Write([]byte(event.Sid))
	key := hex.EncodeToString(hash.Sum(nil))
	t.Log(key[0:8] + event.Sid)
}

func Test_Date(t *testing.T) {
	now := time.Now()
	t.Log(now.Year())
	t.Log(int(now.Month()))
	t.Log(now.Day())
	t.Log(now.Date())
	t.Log(now.Hour())

	t.Log(fmt.Sprintf("%04d-%02d-%02d-%02d", now.Year(), int(now.Month()), now.Day(), now.Hour()))
}

func Test_Hbase(t *testing.T) {
	quiver.ZK_HOSTS = "172.16.53.11"
	quiver.TABLE_NAME = "table1"

	hbaseClient := quiver.NewHbaseUtil()
	event := GetEvent()

	// test parse
	content, length := hbaseClient.ParseDescContent(event)
	t.Log(content)
	t.Log(length)

	// test upload
	if err := hbaseClient.UploadMedia(event); err != nil {
		t.Fatalf("upload error :%v", err)
	}

	hbaseClient.Close()
}
