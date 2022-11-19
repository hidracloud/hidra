package exporter

import (
	"math"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/hidracloud/hidra/v3/internal/config"
	"github.com/hidracloud/hidra/v3/internal/utils"
)

var (
	// samplesMutex is the mutex to protect the samples
	samplesMutex *sync.RWMutex

	// configSamples is the config samples
	configSamples []*config.SampleConfig

	// penaltiesSamples is the penalties
	penaltiesSamples = make(map[string]time.Duration)

	// enqueueSamplesInProgress  is the enqueue samples in progress
	enqueueSamplesInProgress = false

	// gcInProgress is the gc in progress
	gcInProgress = false
)

// RefreshSamples refreshes the samples.
func refreshSamples(cnf *config.ExporterConfig) {
	log.Debug("Refreshing samples...")

	samplesPath, err := utils.AutoDiscoverYML(cnf.SamplesPath)

	if err != nil {
		log.Fatalf("error while refreshing samples: %s", err)
		return
	}

	samplesMutex.Lock()
	defer samplesMutex.Unlock()

	log.Debug("Samples has been updated, trying to calculate new samples...")
	configSamples = nil

	for _, samplePath := range samplesPath {
		sample, err := config.LoadSampleConfigFromFile(samplePath)

		if err != nil {
			log.Fatalf("error while loading sample: %s", err)
			return
		}
		sample.Name = utils.ExtractFileNameWithoutExtension(samplePath)

		log.Debugf("New sample loaded: %s", sample.Name)

		configSamples = append(configSamples, sample)
	}

	// user used to create similar samples near each other, so we shuffle them
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(configSamples), func(i, j int) { configSamples[i], configSamples[j] = configSamples[j], configSamples[i] })

	prometheusMetricStoreMutex.RLock()
	for _, metric := range prometheusMetricStore {
		metric.Reset()
	}

	prometheusMetricStoreMutex.RUnlock()
	if prometheusLastUpdate != nil {
		prometheusLastUpdate.Reset()
	}

	if prometheusStatusMetric != nil {
		prometheusStatusMetric.Reset()
	}

	// we need to reset the last run
	lastRunMutex.Lock()
	lastRun = make(map[string]time.Time)
	lastRunMutex.Unlock()

	log.Debug("Samples has been updated")
}

// refreshPrometheusCustomLabels refreshes the prometheus custom labels
func refreshSampleCommonTags() {
	sampleCommonTags = []string{"sample_name", "plugins", "description"}
	alreadyAdded := make(map[string]bool)

	for _, sample := range configSamples {
		for label := range sample.Tags {
			if _, ok := alreadyAdded[label]; !ok {
				sampleCommonTags = append(sampleCommonTags, label)
				alreadyAdded[label] = true
			}
		}
	}
}

// scheduleDecide decides if a sample should be scheduled
func scheduleDecide(sample *config.SampleConfig, executionDurationAvg, executionDurationStd float64) bool {
	lastRunMutex.RLock()
	lastSampleRun, exists := lastRun[sample.Name]
	lastRunMutex.RUnlock()

	if !exists {
		log.Debugf("Sample %s has never been scheduled, scheduling...", sample.Name)
		return true
	}

	if time.Since(lastSampleRun) < sample.Interval {
		log.Debugf("Sample %s has been scheduled recently, skipping...", sample.Name)
		return false
	}

	penalty, exists := penaltiesSamples[sample.Name]

	if exists {
		if time.Since(lastSampleRun) < sample.Interval+penalty {
			log.Debugf("Sample %s has been scheduled recently and has a penalty, skipping...", sample.Name)
			return false
		}

		log.Debugf("Sample %s has been scheduled recently and has a penalty, but sanction has been reset now", sample.Name)
		delete(penaltiesSamples, sample.Name)

		return true
	}

	sampleRunningTimeMutex.RLock()
	currentSampleRunningTime, exists := sampleRunningTime[sample.Name]
	sampleRunningTimeMutex.RUnlock()

	if !exists {
		log.Debugf("Sample %s has never been executed, scheduling...", sample.Name)
		return true
	}

	penaltyPoints := (float64(currentSampleRunningTime.Load()) - executionDurationAvg) / executionDurationStd

	if penaltyPoints >= 2 {
		penalty = time.Duration(math.Round(penaltyPoints)) * sample.Interval

		log.Warnf("Sample %s has been penalized for %s", sample.Name, penalty.String())
		penaltiesSamples[sample.Name] = penalty

		return false
	}

	return true
}

// EnqueueSamples enqueues the samples.
func enqueueSamples(config *config.ExporterConfig) {
	log.Debug("Enqueuing samples...")

	executionDurationAvg := float64(runningTime.Load()) / float64(len(configSamples))
	executionDurationStd := 0.0

	for _, oneDuration := range sampleRunningTime {
		executionDurationStd += math.Pow(float64(oneDuration.Load())-executionDurationAvg, 2)
	}

	executionDurationStd = math.Sqrt(executionDurationStd / float64(len(configSamples)))

	for _, sample := range configSamples {
		if scheduleDecide(sample, executionDurationAvg, executionDurationStd) {
			log.Debugf("Enqueuing sample %s", sample.Name)
			lastRunMutex.Lock()
			lastRun[sample.Name] = time.Now().Add(time.Hour * 24 * 365)
			lastRunMutex.Unlock()
			samplesJobs <- sample
		}
	}
}

// InitializeScheduler initializes the scheduler
func InitializeScheduler(cnf *config.ExporterConfig) {
	lastRun = make(map[string]time.Time)
	lastRunMutex = &sync.RWMutex{}
	samplesMutex = &sync.RWMutex{}
	refreshSamples(cnf)
	refreshSampleCommonTags()

	listenForOSSignals(cnf)

	samplesJobs = make(chan *config.SampleConfig, cnf.WorkerConfig.MaxQueueSize)
}

// signalHandler handles signals
func signalHandler(signal os.Signal, cnf *config.ExporterConfig) {
	switch signal {
	case syscall.SIGHUP:
		log.Debug("Received SIGHUP, refreshing samples...")
		refreshSamples(cnf)
	case syscall.SIGINT:
		os.Exit(0)
	case syscall.SIGTERM:
		os.Exit(0)
	case syscall.SIGQUIT:
		os.Exit(0)
	case syscall.SIGUSR1:
	case syscall.SIGURG:
	default:
		log.Warnf("Received unknown signal %s", signal)
	}
}

// listenForOSSignals listens for OS signals
func listenForOSSignals(cnf *config.ExporterConfig) {
	sigchnl := make(chan os.Signal, 1)
	signal.Notify(sigchnl)

	go func() {
		for {
			s := <-sigchnl
			signalHandler(s, cnf)
		}
	}()
}

// TickScheduler ticks the scheduler
func TickScheduler(config *config.ExporterConfig) {
	tickerEnqueueSamples := time.NewTicker(config.SchedulerConfig.EnqueueSamplesInterval)
	tickerGC := time.NewTicker(config.SchedulerConfig.GCInterval)
	go func() {
		for {
			select {
			case <-tickerEnqueueSamples.C:
				if !enqueueSamplesInProgress {
					enqueueSamplesInProgress = true
					enqueueSamples(config)
					enqueueSamplesInProgress = false
				}
			case <-tickerGC.C:
				if !gcInProgress {
					gcInProgress = true
					runtime.GC()
					gcInProgress = false
				}
			}
		}
	}()
}
