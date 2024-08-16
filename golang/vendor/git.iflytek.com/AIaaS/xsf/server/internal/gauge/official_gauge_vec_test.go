package gauge

import (
	"log"
	"testing"
	"time"
)

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

func Test_Official_Gauge(t *testing.T) {
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hermes_qps",
		Help: "hermes_qps",
	})

	registry := prometheus.NewRegistry()
	if err := registry.Register(gauge); nil != err {
		log.Fatal(err)
	}

	go func() {
		var cnt float64
		for {
			cnt += 1
			gauge.Set(cnt)
			time.Sleep(time.Millisecond * 500)
		}
	}()

	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	log.Println("Beginning to serve on port :80")
	log.Fatal(http.ListenAndServe(":80", nil))
}

func Test_Official_GaugeVec(t *testing.T) {
	gaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "hermes_qps",
		Help: "hermes_qps",
	}, []string{"tag1", "tag2"})

	registry := prometheus.NewRegistry()
	if err := registry.Register(gaugeVec); nil != err {
		log.Fatal(err)
	}

	go func() {
		var cnt float64
		for {
			cnt += 1
			gaugeVec.WithLabelValues("label_11", "label_21").Set(cnt)
			gaugeVec.WithLabelValues("label_21", "label_22").Set(cnt)
			time.Sleep(time.Millisecond * 500)
		}
	}()

	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	log.Println("Beginning to serve on port :80")
	log.Fatal(http.ListenAndServe(":80", nil))
}
