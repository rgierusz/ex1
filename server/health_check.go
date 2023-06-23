package server

import (
	"github.com/rgierusz/ex1/metric"
	"log"
	"net/http"
)

func livenessHandler(w http.ResponseWriter, _ *http.Request) {
	if _, e := w.Write([]byte("ok")); e != nil {
		metric.GenericResponseErrorCounter.WithLabelValues(e.Error()).Inc()
		log.Printf("Error writing healthcheck response: %v", e)
	}
}
