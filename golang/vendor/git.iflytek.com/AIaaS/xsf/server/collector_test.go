package xsf

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func Test_Gauge(t *testing.T) {
	gaugeVec := NewGaugeVec(GaugeOpts{
		Name: "hermes_qps",
		Help: "hermes_qps",
	}, []string{"tag"})

	metricsRegistry, metricsRegistryErr := NewRegistry()
	if metricsRegistryErr != nil {
		t.Fatal(metricsRegistryErr)
	}
	registerErr := metricsRegistry.Register("hermes_qps", gaugeVec)
	if registerErr != nil {
		t.Fatal(registerErr)
	}

	var wg sync.WaitGroup
	wg.Add(3)

	{
		//produce data
		go func() {
			defer wg.Done()
			var count float64
			for {
				count = count + 10
				gaugeVec.WithLabelValues("upLink setServer").Set(count)
				time.Sleep(1 * time.Second)
			}
		}()
		go func() {
			defer wg.Done()
			var count float64
			for {
				count = count + 15
				gaugeVec.WithLabelValues("downLink getServer").Set(count)
				time.Sleep(5 * time.Second)
			}
		}()
	}

	{
		//collect data
		go func() {
			defer wg.Done()
			for {
				fmt.Println(metricsRegistry.String())
				fmt.Println("------------------------------------")
				time.Sleep(time.Second)
			}
		}()
	}

	wg.Wait()
}

func Test_Counter(t *testing.T) {

	counterVec := NewCounterVec(CounterOpts{
		Name: "hermes_qps",
		Help: "hermes_qps",
	}, []string{"tag"})

	metricsRegistry, metricsRegistryErr := NewRegistry()
	if metricsRegistryErr != nil {
		t.Fatal(metricsRegistryErr)
	}

	registerErr := metricsRegistry.Register("hermes_qps", counterVec)
	if registerErr != nil {
		t.Fatal(registerErr)
	}

	var wg sync.WaitGroup
	wg.Add(3)

	{
		//produce data
		go func() {
			defer wg.Done()
			var count float64
			for {
				count = count + 10
				counterVec.WithLabelValues("upLink setServer").Add(count)
				time.Sleep(1 * time.Second)
			}
		}()
		go func() {
			defer wg.Done()
			var count float64
			for {
				count = count + 15
				counterVec.WithLabelValues("downLink getServer").Add(count)
				time.Sleep(5 * time.Second)
			}
		}()
	}

	{
		//collect data
		go func() {
			defer wg.Done()
			for {
				fmt.Println(metricsRegistry.String())
				fmt.Println("------------------------------------")
				time.Sleep(time.Second)
			}
		}()
	}

	wg.Wait()
}

func Test_Histogram(t *testing.T) {
	histogramVec := NewHistogramVec(HistogramOpts{
		Name:    "hermes_qps",
		Help:    "hermes_qps",
		Buckets: []float64{1, 2, 4},
	}, []string{"tag"})

	metricsRegistry, metricsRegistryErr := NewRegistry()
	if metricsRegistryErr != nil {
		t.Fatal(metricsRegistryErr)
	}
	registerErr := metricsRegistry.Register("hermes_qps", histogramVec)
	if registerErr != nil {
		t.Fatal(registerErr)
	}

	var wg sync.WaitGroup
	wg.Add(3)

	{
		//produce data
		go func() {
			defer wg.Done()
			var count float64
			for {
				count = count + 10
				histogramVec.WithLabelValues("upLink setServer").Observe(count)
				time.Sleep(1 * time.Second)
			}
		}()
		go func() {
			defer wg.Done()
			var count float64
			for {
				count = count + 15
				histogramVec.WithLabelValues("downLink getServer").Observe(count)
				time.Sleep(5 * time.Second)
			}
		}()
	}

	{
		//collect data
		go func() {
			defer wg.Done()
			for {
				fmt.Println(metricsRegistry.String())
				fmt.Println("------------------------------------")
				time.Sleep(time.Second)
			}
		}()
	}

	wg.Wait()
}


func Test_Summary(t *testing.T) {
	summarytVec := NewSummaryVec(SummaryOpts{
		Name:    "hermes_qps",
		Help:    "hermes_qps",
	}, []string{"tag"})

	metricsRegistry, metricsRegistryErr := NewRegistry()
	if metricsRegistryErr != nil {
		t.Fatal(metricsRegistryErr)
	}
	registerErr := metricsRegistry.Register("hermes_qps", summarytVec)
	if registerErr != nil {
		t.Fatal(registerErr)
	}

	var wg sync.WaitGroup
	wg.Add(3)

	{
		//produce data
		go func() {
			defer wg.Done()
			var count float64
			for {
				count = count + 10
				summarytVec.WithLabelValues("upLink setServer").Observe(count)
				time.Sleep(1 * time.Second)
			}
		}()
		go func() {
			defer wg.Done()
			var count float64
			for {
				count = count + 15
				summarytVec.WithLabelValues("downLink getServer").Observe(count)
				time.Sleep(5 * time.Second)
			}
		}()
	}

	{
		//collect data
		go func() {
			defer wg.Done()
			for {
				fmt.Println(metricsRegistry.String())
				fmt.Println("------------------------------------")
				time.Sleep(time.Second)
			}
		}()
	}

	wg.Wait()
}


func Test_Histogram_dbg(t *testing.T) {
	histogramVec := NewHistogramVec(HistogramOpts{
		Name:    "hermes_qps",
		Help:    "hermes_qps",
		Buckets: []float64{1, 2, 4},
	}, []string{"tag"})

	metricsRegistry, metricsRegistryErr := NewRegistry()
	if metricsRegistryErr != nil {
		t.Fatal(metricsRegistryErr)
	}
	registerErr := metricsRegistry.Register("hermes_qps", histogramVec)
	if registerErr != nil {
		t.Fatal(registerErr)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	{
		//produce data
		go func() {
			defer wg.Done()
			var count float64
			for {
				count = count + 10
				histogramVec.WithLabelValues("upLink setServer").Observe(count)
				time.Sleep(1 * time.Second)
			}
		}()
	}

	{
		//collect data
		go func() {
			defer wg.Done()
			for {
				fmt.Println(metricsRegistry.String())
				fmt.Println("------------------------------------")
				time.Sleep(time.Second)
			}
		}()
	}

	wg.Wait()
}
func Test_Gauge_Num(t *testing.T) {
	gaugeVec := NewGaugeVec(GaugeOpts{
		//Name: "hermes_qps",
		Name: "999",
		Help: "hermes_qps",
	}, []string{"tag"})

	metricsRegistry, metricsRegistryErr := NewRegistry()
	if metricsRegistryErr != nil {
		t.Fatal(metricsRegistryErr)
	}
	registerErr := metricsRegistry.Register("hermes_qps", gaugeVec)
	if registerErr != nil {
		t.Fatal(registerErr)
	}

	var wg sync.WaitGroup
	wg.Add(3)

	{
		//produce data
		go func() {
			defer wg.Done()
			var count float64
			for {
				count = count + 10
				gaugeVec.WithLabelValues("upLink setServer").Set(count)
				time.Sleep(1 * time.Second)
			}
		}()
		go func() {
			defer wg.Done()
			var count float64
			for {
				count = count + 15
				gaugeVec.WithLabelValues("downLink getServer").Set(count)
				time.Sleep(5 * time.Second)
			}
		}()
	}

	{
		//collect data
		go func() {
			defer wg.Done()
			for {
				fmt.Println(metricsRegistry.String())
				fmt.Println("------------------------------------")
				time.Sleep(time.Second)
			}
		}()
	}

	wg.Wait()
}