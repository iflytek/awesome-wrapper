package test

import (
	"testing"
	"genitus/quiver"
)

func Test_S3(t *testing.T) {
	quiver.InitS3Info(
		"ewogICAgIlJHV19UT0tFTiI6IHsKICAgICAgICAidmVyc2lvbiI6IDEsCiAgICAgICAgInR5cGUiOiAibGRhcCIsCiAgICAgICAgImlkIjogImhhZG9vcDIiLAogICAgICAgICJrZXkiOiAiaGFkb29wMiIKICAgIH0KfQo=",
		"123",
		"http://10.1.86.14:8080")

	s3Client := quiver.NewS3Util()
	event := GetEvent()
	event.Tag(quiver.KV{"aue", "aue-test"})
	event.Tag(quiver.KV{"auf", "auf-test"})
	event.Tag(quiver.KV{"rate", "rate-test"})

	// test parse
	content, length := s3Client.ParseDescContent(event)
	t.Log(content)
	t.Log(length)

	// test upload
	if err := s3Client.UploadMedia(event); err != nil {
		t.Fatalf("upload error :%v", err)
	}

	// test download
	if err := s3Client.DownloadMedia(event.Sid, "s3", ""); err != nil {
		t.Fatalf("download error %v", err)
	}
}
