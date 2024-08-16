package gauge

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"testing"
	"time"
)

import (
	"github.com/prometheus/client_golang/prometheus"
)

func Test_Custom_GaugeVec(t *testing.T) {
	registry := prometheus.NewRegistry()
	single(registry)
	multiple(registry)
	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	log.Println("Beginning to serve on port :80")
	log.Fatal(http.ListenAndServe(":80", nil))
}
func single(registry *prometheus.Registry) {
	gaugeVec := NewGauge(prometheus.GaugeOpts{
		Name: "hermes_qps_single",
		Help: "hermes_qps_single",
	})
	if err := registry.Register(gaugeVec); nil != err {
		log.Fatal(err)
	}
	go func() {
		var cnt float64
		for {
			cnt += 1
			gaugeVec.Set(cnt)
			//time.Sleep(time.Millisecond * 500)
			time.Sleep(5 * time.Second)
		}
	}()
}
func multiple(registry *prometheus.Registry) {
	gaugeVec := NewGaugeVec(prometheus.GaugeOpts{
		Name: "hermes_qps_multiple",
		Help: "hermes_qps_multiple",
	}, []string{"tag1", "tag2"})
	if err := registry.Register(gaugeVec); nil != err {
		log.Fatal(err)
	}
	go func() {
		var cnt float64
		for {
			cnt += 1
			gaugeVec.WithLabelValues("label_11", "label_21").Set(cnt)
			gaugeVec.WithLabelValues("label_21", "label_22").Set(cnt)
			//time.Sleep(time.Millisecond * 500)
			time.Sleep(5 * time.Second)
		}
	}()
}
