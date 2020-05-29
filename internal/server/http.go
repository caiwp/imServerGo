package server

import (
	"net/http"

	"gogit.oa.com/March/gopkg/util"

	_ "net/http/pprof"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func runHTTP(addr string) {
	if addr == "" {
		return
	}
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(addr, nil)
	util.MustNil(err)
}
