package stress

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hidracloud/hidra/v3/config"
	"github.com/hidracloud/hidra/v3/internal/runner"
)

// TestArgs represents the arguments for the test
type TestArgs struct {
	SampleConfig *config.SampleConfig
	StatusChan   *chan Status
	Duration     time.Duration
	Workers      int
}

// Status represents the status of the test
type Status struct {
	Successes uint64
	Errors    uint64
	LastError error
	StartTime time.Time
}

// Start starts the stress test
func Start(args *TestArgs) {
	wg := &sync.WaitGroup{}

	// create atomic counter
	ctx := context.TODO()

	var lastError error

	var errors, successes uint64

	startTime := time.Now()

	for i := 0; i < args.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for time.Since(startTime) < args.Duration {
				result := runner.RunSample(ctx, args.SampleConfig)

				if result.Error != nil {
					lastError = result.Error
					atomic.AddUint64(&errors, 1)
				} else {
					atomic.AddUint64(&successes, 1)
				}
			}
		}()
	}

	// Create a goroutine for sending the status
	go func() {
		for time.Since(startTime) < args.Duration {
			*args.StatusChan <- Status{
				Successes: successes,
				Errors:    errors,
				StartTime: startTime,
				LastError: lastError,
			}

			time.Sleep(time.Second)
		}
	}()

	wg.Wait()

	close(*args.StatusChan)
}
