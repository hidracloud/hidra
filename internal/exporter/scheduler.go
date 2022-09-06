package exporter

import (
	"math"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/hidracloud/hidra/v3/internal/config"
	"github.com/hidracloud/hidra/v3/internal/utils"
)

var (
	// samplesMutex is the mutex to protect the samples
	samplesMutex *sync.RWMutex

	// oldSamplesPath is the old samples path
	oldSamplesPath []string

	// configSamples is the config samples
	configSamples []*config.SampleConfig

	// penaltiesSamples is the penalties
	penaltiesSamples map[string]time.Duration

	// inProgress is the in progress
	inProgress map[string]bool
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

	if utils.EqualSlices(oldSamplesPath, samplesPath) {
		log.Debug("Samples are the same, skipping...")
		return
	}

	oldSamplesPath = samplesPath

	log.Debug("Samples has been updated, trying to calculate new samples...")
	configSamples = nil

	for _, samplePath := range samplesPath {
		sample, err := config.LoadSampleConfigFromFile(samplePath)

		sample.Name = utils.ExtractFileNameWithoutExtension(samplePath)

		if err != nil {
			log.Fatalf("error while loading sample: %s", err)
			return
		}

		log.Debugf("New sample loaded: %s", sample.Name)

		configSamples = append(configSamples, sample)
	}

	log.Debug("Samples has been updated")
}

// refreshPrometheusCustomLabels refreshes the prometheus custom labels
func refreshSampleCommonTags() {
	sampleCommonTags = make([]string, 0)
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
	sampleInProgress, exists := inProgress[sample.Name]

	if !exists {
		log.Debugf("Sample %s has never been scheduled, scheduling...", sample.Name)

		return true
	}

	if sampleInProgress {
		log.Debugf("Sample %s is already in progress, skipping...", sample.Name)

		return false
	}

	lastSampleRun, exists := lastRun[sample.Name]

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

	currentSampleRunningTime, exists := sampleRunningTime[sample.Name]

	if !exists {
		log.Debugf("Sample %s has never been executed, scheduling...", sample.Name)
		return true
	}

	penaltyPoints := (float64(currentSampleRunningTime.Load()) - executionDurationAvg) / executionDurationStd

	if penaltyPoints > 2 {
		penalty = time.Duration(math.Round(penaltyPoints)) * sample.Interval

		log.Info("Sample %s has been penalized for %s", sample.Name, penalty.String())
		penaltiesSamples[sample.Name] = penalty

		return false
	}

	return true
}

// EnqueueSamples enqueues the samples.
func enqueueSamples(config *config.ExporterConfig) {
	log.Debug("Enqueuing samples...")
	samplesMutex.Lock()
	defer samplesMutex.Unlock()

	executionDurationAvg := float64(runningTime.Load()) / float64(len(configSamples))
	executionDurationStd := 0.0

	for _, oneDuration := range sampleRunningTime {
		executionDurationStd += math.Pow(float64(oneDuration.Load())-executionDurationAvg, 2)
	}

	executionDurationStd = math.Sqrt(executionDurationStd / float64(len(configSamples)))

	for _, sample := range configSamples {
		if scheduleDecide(sample, executionDurationAvg, executionDurationStd) {
			log.Debugf("Enqueuing sample %s", sample.Name)
			inProgress[sample.Name] = true
			samplesJobs <- sample
		}
	}
}

// InitializeScheduler initializes the scheduler
func InitializeScheduler(config *config.ExporterConfig) {
	lastRun = make(map[string]time.Time)
	samplesMutex = &sync.RWMutex{}
	inProgress = make(map[string]bool)

	refreshSamples(config)
	refreshSampleCommonTags()
}

// TickScheduler ticks the scheduler
func TickScheduler(config *config.ExporterConfig) {
	tickerRefreshSamples := time.NewTicker(config.SchedulerConfig.RefreshSamplesInterval)
	tickerEnqueueSamples := time.NewTicker(config.SchedulerConfig.EnqueueSamplesInterval)

	go func() {
		for {
			select {
			case <-tickerRefreshSamples.C:
				refreshSamples(config)
			case <-tickerEnqueueSamples.C:
				enqueueSamples(config)
			}
		}
	}()
}
