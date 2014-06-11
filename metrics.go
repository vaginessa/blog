package main

import (
	"log"
	"net/http"
	"time"

	"github.com/rcrowley/go-metrics"
	"github.com/rcrowley/go-metrics/librato"
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

	if !inProduction {
		logger.Notice("Librato stats disabled because not in production")
		return
	}
	if StringEmpty(config.LibratoToken) || StringEmpty(config.LibratoEmail) {
		logger.Notice("Librato stats disabled because no config.LibratoToken or no config.LibratoEmail")
		return
	}

	logger.Notice("Starting librato stats\n")
	go func() {
		librato.Librato(defReg, 1*time.Minute, *config.LibratoEmail, *config.LibratoToken, "blog", make([]float64, 0), time.Second*15)
	}()
}
