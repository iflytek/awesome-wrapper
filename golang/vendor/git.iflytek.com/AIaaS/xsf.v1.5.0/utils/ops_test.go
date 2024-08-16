package utils

import (
	"testing"
	"time"
)

func TestSyncXsfInitStatus(t *testing.T) {
	t.Log("SyncXsfInitStatus")
	SyncXsfInitStatus()
	time.Sleep(time.Second * 3)
	t.Log("SyncUserInitStatus")
	SyncUserInitStatus()
	time.Sleep(time.Second * 3)
	t.Log("SyncBvtInitStatus")
	SyncBvtInitStatus()
	time.Sleep(time.Second * 3)
	t.Log("SyncFinishStatus")
	SyncFinishStatus()
	time.Sleep(time.Second * 3)
}
