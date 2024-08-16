package test

import (
	"fmt"
	"genitus/flange"
	"io"
	"os"
	"testing"
	"strconv"
)

var (
	spillChar       = "*-*#-..#"
	test_spill_path = "spill"
)

func Test_Spill(t *testing.T) {
	defer flange.Fini()
	flange.Logger = &flange.FmtLog{}
	flange.Init("127.0.0.1", "4545", 4, "0.0.0.1", "9090", "iat")

	f := test_spill_path + string(os.PathSeparator) + "span_spill"
	fp, err := os.OpenFile(f, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		t.Fatalf("open %s failed: %v", f, err)
	}
	fp.Seek(0, io.SeekEnd)

	for i := 0; i < 10; i++ {
		span := GetASpan()

		buf, err := flange.Serialize(span)
		if err != nil {
			t.Fatalf("spill serialize err:%v", err)
		}

		fmt.Fprint(fp, spillChar)
		fmt.Fprintln(fp, string(buf))
	}

	t.Log("--ok--")
}

func Test_ReverseSpill(t *testing.T) {
	defer flange.Fini()
	flange.Logger = &flange.FmtLog{}
	flange.Init("172.16.51.29", "4545", 4, "0.0.0.1", "9090", "iat")

	f := test_spill_path + string(os.PathSeparator) + "span_spill"
	fp, err := os.OpenFile(f, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		t.Fatalf("open %s failed: %v", f, err)
	}

	for i := 0; i < 10; i++ {
		buf, _ := flange.Serialize(GetASpan())
		fmt.Fprint(fp, string(buf))
		fmt.Fprintf(fp, "%012d", len(buf))
	}

	lengthBuffer := make([]byte, 12)
	fp.Seek(-12, io.SeekEnd)

	for l, err := fp.Read(lengthBuffer); l == 12 && err == nil; {
		length, _ := strconv.ParseInt(string(lengthBuffer), 10, 16)

		// seek to start position
		fp.Seek(-(12 + length), io.SeekEnd)
		buffer := make([]byte, length)
		if l, err := fp.Read(buffer); int64(l) == length && err == nil {
			if check(buffer) {
				state, _ := fp.Stat()
				fp.Truncate(state.Size() - int64(12+length))

				// seek to next position
				// length for current record, 12 for next record length
				fp.Seek(-(length + 12), io.SeekCurrent)
			} else {
				// no change for spill file, exit for next try
				break
			}
		} else {
			// no change for spill file, exit for next try
			break
		}
	}

	t.Log("--ok--")
}

func check(bytes []byte) bool {
	tid, sid, _, _ := flange.RetrieveSpanInfo(bytes)
	fmt.Println(tid, sid)
	return true
}
