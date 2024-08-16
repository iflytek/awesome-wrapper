package test

import (
	"testing"
	"genitus/quiver"
)

func Test_Serialize(t *testing.T) {
	event := GetEvent()

	buf, err := quiver.Serialize(event)
	if err != nil {
		t.Fatalf("serialize error :%v", err)
	}

	deEvent := quiver.Deserialize(buf)
	if deEvent == nil {
		t.Fatalf("deserialize error :%v", err)
	}

	//check
	if event.Sid != deEvent.Sid {
		t.Fatalf("serialize not equal deserialize")
	} else {
		t.Logf("serialize test success")
	}
}

// 200000	      7510 ns/op	   27211 B/op	     182 allocs/op
func Benchmark_Serialize(b *testing.B) {
	event := GetEvent()

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			quiver.Serialize(event)
		}
	})
}

//
func Benchmark_Deserialize(b *testing.B) {
	event := GetEvent()
	buf, _ := quiver.Serialize(event)

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			quiver.Deserialize(buf)
		}
	})
}
