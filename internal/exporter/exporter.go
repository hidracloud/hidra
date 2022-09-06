package exporter

import (
	"net/http"
	"sync"

	"github.com/hidracloud/hidra/v3/internal/config"
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

	wg := &sync.WaitGroup{}

	log.Debug("Initializing scheduler...")

	wg.Add(1)
	InitializeScheduler(config)
	go TickScheduler(config, wg)

	log.Debug("Initializing workers...")

	wg.Add(1)
	InitializeWorker(config)
	go RunWorkers(config, wg)

	http.Handle(config.HTTPServerConfig.MetricsPath, promhttp.Handler())
	err := http.ListenAndServe(config.HTTPServerConfig.ListenAddress, nil)
	if err != nil {
		panic(err)
	}

	wg.Wait()
}
