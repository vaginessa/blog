package main

import (
	"log"
	"net/http"

	"github.com/rcrowley/go-metrics"
)

var (
	// number of /resize requests being processed at this time
	metricCurrentReqs metrics.Counter
	// rate of http requests
	metricHttpReqRate metrics.Meter
	// how long does it take to service http request
	metricHttpReqTime metrics.Timer
	// how long does it take to backup to s3
	metricsBackupTime metrics.Timer
)

func handleMetrics(w http.ResponseWriter, r *http.Request) {
	reg, ok := metrics.DefaultRegistry.(*metrics.StandardRegistry)
	if !ok {
		log.Fatalln("metrics.DefaultRegistry type assertion failed")
	}

	json, err := reg.MarshalJSON()
	if err != nil {
		log.Fatalln("metrics.DefaultRegistry.MarshalJSON:", err)
	}

	textResponse(w, string(json))
}

func InitMetrics() {
	defReg := metrics.DefaultRegistry
	metricCurrentReqs = metrics.NewRegisteredCounter("curr_http_req", defReg)
	metricHttpReqRate = metrics.NewRegisteredMeter("http_req_rate", defReg)
	metricHttpReqTime = metrics.NewRegisteredTimer("http_req_time", defReg)
	metricsBackupTime = metrics.NewRegisteredTimer("backup_time", defReg)
}
