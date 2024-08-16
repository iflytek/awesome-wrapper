package test

import (
	"testing"
	"genitus/flange"
)

// 2000000	       591 ns/op	     851 B/op	       7 allocs/op
func Benchmark_NewSpan(b *testing.B) {
	defer flange.Fini()
	flange.SpillEnable = false
	flange.DumpEnable = false
	flange.DeliverEnable = false
	//flange.Logger = &flange.FmtLog{}
	flange.Init("127.0.0.1", "4545", 4, "0.0.0.1", "9090", "iat")

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = flange.NewSpan(flange.SERVER, true)
		}
	})
}

// 2000000	       600 ns/op	     887 B/op	       6 allocs/op
func Benchmark_NextSpan(b *testing.B) {
	defer flange.Fini()
	flange.SpillEnable = false
	flange.DumpEnable = false
	flange.DeliverEnable = false
	flange.Init("127.0.0.1", "4545", 4, "0.0.0.1", "9090", "iat")
	root := flange.NewSpan(flange.SERVER, false)

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = root.Next(flange.SERVER)
		}
	})
}

// 2000000	       529 ns/op	     768 B/op	       4 allocs/op
func Benchmark_FromMeta(b *testing.B) {
	defer flange.Fini()
	flange.SpillEnable = false
	flange.DumpEnable = false
	flange.DeliverEnable = false
	flange.Init("127.0.0.1", "4545", 4, "0.0.0.1", "9090", "iat")
	root := flange.NewSpan(flange.SERVER, false)
	meta := root.Meta()

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = flange.FromMeta(meta, flange.SERVER)
		}
	})
}

//300000000	         4.28 ns/op	       0 B/op	       0 allocs/op
func Benchmark_WithName(b *testing.B) {
	defer flange.Fini()
	flange.SpillEnable = false
	flange.DumpEnable = false
	flange.DeliverEnable = false
	flange.Init("127.0.0.1", "4545", 4, "0.0.0.1", "9090", "iat")
	root := flange.NewSpan(flange.SERVER, false)

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			root.WithName("sessionbegin")
		}
	})
}

//50000000	        25.4 ns/op	       0 B/op	       0 allocs/op
func Benchmark_Start(b *testing.B) {
	defer flange.Fini()
	flange.SpillEnable = false
	flange.DumpEnable = false
	flange.DeliverEnable = false
	flange.Init("127.0.0.1", "4545", 4, "0.0.0.1", "9090", "iat")
	root := flange.NewSpan(flange.SERVER, false)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		root.Start()
	}
}

//50000000	        26.2 ns/op	       0 B/op	       0 allocs/op
func Benchmark_End(b *testing.B) {
	defer flange.Fini()
	flange.SpillEnable = false
	flange.DumpEnable = false
	flange.DeliverEnable = false
	flange.Init("127.0.0.1", "4545", 4, "0.0.0.1", "9090", "iat")
	root := flange.NewSpan(flange.SERVER, false)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		root.End()
	}
}

//100000000	        13.0 ns/op	       0 B/op	       0 allocs/op
func Benchmark_WithTag(b *testing.B) {
	defer flange.Fini()
	flange.SpillEnable = false
	flange.DumpEnable = false
	flange.DeliverEnable = false
	flange.Init("127.0.0.1", "4545", 4, "0.0.0.1", "9090", "iat")
	root := flange.NewSpan(flange.SERVER, false)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		root.WithTag("appid", "100IME")
	}
}

func Benchmark_WithDescf(b *testing.B) {
	defer flange.Fini()
	flange.SpillEnable = false
	flange.DumpEnable = false
	flange.DeliverEnable = false
	flange.Init("127.0.0.1", "4545", 4, "0.0.0.1", "9090", "iat")
	root := flange.NewSpan(flange.SERVER, false)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		root.WithDescf("hello %s", "world")
	}
}

//50000000	        35.5 ns/op	      64 B/op	       1 allocs/op
func Benchmark_Meta(b *testing.B) {
	defer flange.Fini()
	flange.SpillEnable = false
	flange.DumpEnable = false
	flange.DeliverEnable = false
	flange.Init("127.0.0.1", "4545", 4, "0.0.0.1", "9090", "iat")
	root := flange.NewSpan(flange.SERVER, false)

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			root.Meta()
		}
	})
}

//10000000	       178 ns/op	     240 B/op	       5 allocs/op
func Benchmark_Serialize(b *testing.B) {
	defer flange.Fini()
	flange.SpillEnable = false
	flange.DumpEnable = false
	flange.DeliverEnable = false
	flange.Init("127.0.0.1", "4545", 4, "0.0.0.1", "9090", "iat")
	root := flange.NewSpan(flange.SERVER, false)

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			flange.Serialize(root)
		}
	})
}

// 30000000	       100 ns/op	      93 B/op	       0 allocs/op
func Benchmark_Flush(b *testing.B) {
	defer flange.Fini()
	flange.SpillEnable = false
	flange.DumpEnable = false
	flange.DeliverEnable = false
	flange.Init("127.0.0.1", "4545", 4, "0.0.0.1", "9090", "iat")
	root := flange.NewSpan(flange.SERVER, false)

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			flange.Flush(root)
		}
	})
}
