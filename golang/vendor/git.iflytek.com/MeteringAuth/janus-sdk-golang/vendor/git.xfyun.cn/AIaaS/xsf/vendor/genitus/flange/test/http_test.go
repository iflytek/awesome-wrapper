package test

import (
	"testing"
	"net/http"
	"fmt"
)

func Test_Http(t *testing.T) {
	http.HandleFunc("/metric", ReportMetric)
	if err := http.ListenAndServe(":12331", nil); err != nil {
		t.Fatalf("start metric report server error: %v", err)
	}
}

func ReportMetric(w http.ResponseWriter, req *http.Request) {
	fmt.Printf("report metric")
	metric := "{"
	metric += "\"consumer\": ["
	for i := 0; i < 4; i++ {
		metric += fmt.Sprintf("{\"cid\":%d, \"current_items\":%d},", i, 50)
	}
	metric = metric[0:len(metric)-1]
	metric += "]"

	if true {
		metric += ", \"spill\": true"
	}
	if true {
		metric += ", \"reverse\": true"
	}
	if true {
		metric += ", \"mg\":true"
	}

	metric += "}"
	w.Write([]byte(metric))
}
