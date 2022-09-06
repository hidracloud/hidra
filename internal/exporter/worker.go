package exporter

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/hidracloud/hidra/v3/internal/config"
	log "github.com/sirupsen/logrus"
)

var (
	// samplesJobs is the samples jobs
	samplesJobs chan *config.SampleConfig

	// runningTime is the running time
	runningTime *atomic.Uint64

	// runningSamples is the running samples
	sampleRunningTime map[string]*atomic.Uint64

	// lastSchedulerRun is the last scheduler run
	lastRun map[string]time.Time

	// lastRunMutex is the mutex to protect the last scheduler run
	lastRunMutex *sync.RWMutex
)

// InitializeWorker initializes the worker
func InitializeWorker(config *config.ExporterConfig) {
	runningTime = &atomic.Uint64{}
	sampleRunningTime = make(map[string]*atomic.Uint64)
	lastRun = make(map[string]time.Time)
	lastRunMutex = &sync.RWMutex{}
}

// RunWorkers runs the workers
func RunWorkers(cnf *config.ExporterConfig, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Debugf("Initializing %d workers...", cnf.WorkerConfig.ParallelJobs)

	samplesJobs = make(chan *config.SampleConfig, cnf.WorkerConfig.MaxQueueSize)

	for i := 0; i < cnf.WorkerConfig.ParallelJobs; i++ {
		go func(worker int) {
			wg.Add(1)
			defer wg.Done()
			for {
				sample := <-samplesJobs
				startTime := time.Now()
				log.Debugf("Running sample %s, with description %s from worker %d", sample.Name, sample.Description, worker)

				runningTime.Add(uint64(time.Since(startTime).Milliseconds()))

				lastRunMutex.Lock()

				if _, ok := sampleRunningTime[sample.Name]; !ok {
					sampleRunningTime[sample.Name] = &atomic.Uint64{}
				}

				sampleRunningTime[sample.Name].Add(uint64(time.Since(startTime).Milliseconds()))
				lastRun[sample.Name] = time.Now()
				inProgress[sample.Name] = false
				lastRunMutex.Unlock()
			}
		}(i)
	}
}
