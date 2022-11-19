package exporter

import (
	"net/http"
	"net/http/pprof"
	"os"

	"github.com/hidracloud/hidra/v3/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

var (
	// prometheusCustomLabels is the custom labels to add to the metrics
	sampleCommonTags []string
)

// Initialize initializes the exporter
func Initialize(config *config.ExporterConfig) {
	log.Info("Initializing hidra exporter...")

	if config.HTTPServerConfig.ListenAddress == "" {
		log.Fatal("Listen address is empty, please refer to the documentation")
	}

	if config.SamplesPath == "" {
		log.Fatal("Samples path is empty")
	}

	log.Debug("Initializing scheduler...")

	InitializeScheduler(config)
	go TickScheduler(config)

	log.Debug("Initializing workers...")

	InitializeWorker(config)
	go RunWorkers(config)

	myMux := http.NewServeMux()

	myMux.Handle(config.HTTPServerConfig.MetricsPath, promhttp.Handler())
	log.Infof("Listening on %s and path %s", config.HTTPServerConfig.ListenAddress, config.HTTPServerConfig.MetricsPath)

	if os.Getenv("DEBUG") == "true" {
		log.Info("Debug mode enabled")
		myMux.HandleFunc("/debug/pprof/", pprof.Index)
		myMux.HandleFunc("/debug/pprof/{action}", pprof.Index)
		myMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	}

	err := http.ListenAndServe(config.HTTPServerConfig.ListenAddress, myMux)

	if err != nil {
		panic(err)
	}
}
