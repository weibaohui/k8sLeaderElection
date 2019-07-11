package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	_ "net/http/pprof"
	"time"
)

var (
	opsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "myapp_processed_ops_total",
		Help: "The total number of processed events",
	})
)
func recordMetrics() {
	go func() {
		for {
			opsProcessed.Inc()
			time.Sleep(2 * time.Second)
		}
	}()
}
func Start() {
	go recordMetrics()
	http.HandleFunc("/healthy", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprintf(writer, "ok")
	})
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":80", nil)
}
