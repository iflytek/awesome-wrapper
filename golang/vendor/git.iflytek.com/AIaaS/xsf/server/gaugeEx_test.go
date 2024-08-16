package xsf

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func Test_GaugeEx(t *testing.T) {
	gaugeVecEx := NewGaugeVecEx(GaugeOpts{
		Name: "hermes_qps",
		Help: "hermes_qps",
	}, []string{"tag"})

	metricsRegistry, metricsRegistryErr := NewRegistry()
	if metricsRegistryErr != nil {
		t.Fatal(metricsRegistryErr)
	}
	registerErr := metricsRegistry.Register("hermes_qps", gaugeVecEx)
	if registerErr != nil {
		t.Fatal(registerErr)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	//produce data
	go func() {
		defer wg.Done()
		var count float64
		for {
			count = count + 10
			gaugeVecEx.WithLabelValues("upLink setServer").Set(count)
			gaugeVecEx.WithLabelValues("downLink getServer").Set(count)
			time.Sleep(3 * time.Second)
		}
	}()

	//collect data
	go func() {
		defer wg.Done()
		for {
			fmt.Println(metricsRegistry.String())
			fmt.Println("------------------------------------")
			time.Sleep(time.Second)
		}
	}()

	wg.Wait()
}
