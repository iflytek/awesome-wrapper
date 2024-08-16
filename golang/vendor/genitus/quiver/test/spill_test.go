package test

import (
	"testing"
	"os"
	"genitus/quiver"
	"fmt"
	"io"
	"strings"
	"io/ioutil"
	"strconv"
)

var (
	spillChar       = "*-*#-..#"
	test_spill_path = "spill"
)

func Test_Spill(t *testing.T) {
	f := test_spill_path + string(os.PathSeparator) + "event_spill"
	fp, err := os.OpenFile(f, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		t.Fatalf("open %s failed: %v", f, err)
	}
	fp.Seek(0, io.SeekEnd)

	for i := 0; i < 100; i++ {
		event := GetEvent()

		buf, err := quiver.Serialize(event)
		if err != nil {
			t.Fatalf("spill serialize err:%v", err)
		}

		fmt.Fprint(fp, spillChar)
		fmt.Fprintln(fp, string(buf))
	}

	t.Log("--ok--")
}

func Test_ReverseSpillBatch(t *testing.T) {
	f := test_spill_path + string(os.PathSeparator) + "event_spill"
	fp, err := os.OpenFile(f, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		t.Fatalf("open %s failed: %v", f, err)
	}

	for i := 0; i < 10; i++ {
		buf, _ := quiver.Serialize(GetEvent())
		fmt.Fprint(fp, string(buf))
		fmt.Fprintf(fp, "%012d", len(buf))
		t.Logf("write: len=%d, buf=%v", len(buf), buf)
	}

	t.Log("--start to reverse spill--")
	// start reverse spill
	lengthBuffer := make([]byte, 12)
	fp.Seek(-12, io.SeekEnd)

	for l, err := fp.Read(lengthBuffer); l == 12 && err == nil; {
		length, _ := strconv.ParseInt(string(lengthBuffer), 10, 16)

		// seek to start position
		fp.Seek(-(12 + length), io.SeekEnd)
		buffer := make([]byte, length)
		if l, err := fp.Read(buffer); int64(l) == length && err == nil {
			t.Logf("read: len=%d, buf=%v", length, buffer)

			if check(buffer) {
				state, _ := fp.Stat()
				fp.Truncate(state.Size() - int64(12+length))

				fp.Seek(-(length + 12), io.SeekCurrent)
			} else {
				// no change for file
				break
			}
		} else {
			// no change for file
			break
		}
	}

	t.Log("--ok--")
}

func Test_ReverseSpill(t *testing.T) {
	f := test_spill_path + string(os.PathSeparator) + "event_spill"
	fp, err := os.OpenFile(f, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		t.Fatalf("open %s failed: %v", f, err)
	}

	for i := 0; i < 10; i++ {
		buf, _ := quiver.Serialize(GetEvent())
		fmt.Fprint(fp, spillChar)
		fmt.Fprintln(fp, string(buf))
	}

	fp.Seek(0, io.SeekStart)
	buf, err := ioutil.ReadAll(fp)
	if err != nil {
		t.Fatalf("read file error :%v", err)
	}

	state, _ := fp.Stat()
	content := string(buf)
	for i := strings.LastIndex(content, spillChar); i >= 0; i = strings.LastIndex(content, spillChar) {
		if check(buf[i+8:]) {
			content = string(buf[0:i])
			fp.Truncate(state.Size() - int64(len(buf[i:])))
		}
	}

	t.Log("--ok--")
}

func check(bytes []byte) bool {
	e := quiver.Deserialize(bytes)
	if e != nil {
		fmt.Println(e.Sid)
		return true
	}
	return false
}
