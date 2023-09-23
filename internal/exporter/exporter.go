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

// basicAuthHandler is the basic auth handler
func basicAuthHandler(h http.Handler, username, password string, enabled bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if enabled {
			user, pass, ok := r.BasicAuth()

			if !ok || user != username || pass != password {
				w.Header().Set("WWW-Authenticate", `Basic realm="metrics"`)
				w.WriteHeader(http.StatusUnauthorized)
				_, err := w.Write([]byte("Unauthorized.\n"))

				if err != nil {
					log.Error(err)
				}

				return
			}
		}

		h.ServeHTTP(w, r)
	})
}

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

	myMux.Handle(config.HTTPServerConfig.MetricsPath, basicAuthHandler(promhttp.Handler(), config.BasicAuth.Username, config.BasicAuth.Password, config.BasicAuth.Enabled))

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
