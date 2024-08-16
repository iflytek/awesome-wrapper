package test

import (
	"testing"
	"genitus/flange"
)

func Test_Serialize(t *testing.T) {
	defer flange.Fini()
	flange.Logger = &flange.FmtLog{}
	flange.Init("127.0.0.1", "4545", 4, "0.0.0.1", "9090", "iat")

	span := GetASpan()
	buf, err := flange.Serialize(span)
	if err != nil {
		t.Fatalf("serialize error :%v", err)
	}

	tid, sid, _, buf := flange.RetrieveSpanInfo(buf)
	t.Logf("tid=%s, sid=%s", tid, sid)

	span = flange.Deserialize(buf)
	t.Logf("tid=%s", span.TraceId)
}
