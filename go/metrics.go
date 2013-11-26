package main

import (
	"github.com/rcrowley/go-metrics"
	"log"
	"net/http"
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

func handleStats(w http.ResponseWriter, r *http.Request) {
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

func initMetrics() {
	defReg := metrics.DefaultRegistry
	metricCurrentReqs = metrics.NewRegisteredCounter("current http requests", defReg)
	metricHttpReqRate = metrics.NewRegisteredMeter("http requests rate", defReg)
	metricHttpReqTime = metrics.NewRegisteredTimer("http requests time", defReg)
	metricsBackupTime = metrics.NewRegisteredTimer("backup time", defReg)
}
