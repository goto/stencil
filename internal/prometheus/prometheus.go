package prometheus

import (
	"log"
	"net/http"
	"net/http/pprof"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func MetricsHandler(mux *http.ServeMux) {
	log.Println("enabling profile metrics")

	mux.Handle("/metrics", promhttp.Handler())

	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
}
