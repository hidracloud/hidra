package attack

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"github.com/hidracloud/hidra/src/models"
	"github.com/hidracloud/hidra/src/scenarios"
)

// IterationResult is the result of an iteration
type IterationResult struct {
	Error bool
	Time  int64
}

// RunAttackMode runs the attack mode
func RunAttackMode(testFile, results string, workers, duration int) {
	log.Println("Running attack mode")

	data, err := ioutil.ReadFile(testFile)
	if err != nil {
		panic(err)
	}

	sample, err := models.ReadSampleYAML(data)
	if err != nil {
		panic(err)
	}

	spawnWorkers(workers, duration, sample, results)
}

// spawnWorkers spawns the workers
func spawnWorkers(workers, duration int, sample *models.Sample, resultsCSV string) {
	results := make([][]*IterationResult, workers)
	runners := make([]models.IScenario, workers)
	for i := 0; i < workers; i++ {
		tmpRunner, err := scenarios.InitializeScenario(sample.Scenario)

		if err != nil {
			panic(err)
		}
		runners[i] = tmpRunner
		results[i] = make([]*IterationResult, 0)
	}

	wg := sync.WaitGroup{}
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go spawnWorker(i, duration, sample, runners[i], &wg, results)
	}

	wg.Wait()

	if resultsCSV != "" {
		err := dumpResults2CSV(results, resultsCSV)
		if err != nil {
			panic(err)
		}
	}

	printResultResume(results, duration)
}

// printResultResume prints the resume of the results
func printResultResume(results [][]*IterationResult, duration int) {
	totalIterations := 0
	totalErrors := 0

	for i := 0; i < len(results); i++ {
		fmt.Printf("### Worker %d: ", i)
		fmt.Printf("%d iterations, ", len(results[i]))

		errors := 0
		time := int64(0)
		for j := 0; j < len(results[i]); j++ {
			if results[i][j].Error {
				errors += 1
			}
			time += results[i][j].Time
		}

		fmt.Printf("%d errors, ", errors)
		fmt.Printf("%d avg time ms, ", time/int64(len(results[i])))
		fmt.Printf("%d req/s, ", int64(len(results[i]))*1000/time)
		fmt.Printf("%d total time ms\n", time)

		totalErrors += errors
		totalIterations += len(results[i])
	}

	fmt.Println()
	rate := float64(totalIterations) / float64(duration)
	fmt.Printf("### Total: %d iterations, %d errors, %f req/s\n", totalIterations, totalErrors, rate)
}

// dumpResults2CSV dumps the results to a CSV file
func dumpResults2CSV(results [][]*IterationResult, file string) error {
	log.Println("Dumping results to CSV")

	// open file
	f, err := os.Create(file)

	if err != nil {
		return err
	}

	defer f.Close()

	// write header
	f.WriteString("worker,iteration,error,time\n")

	// write results
	for i := 0; i < len(results); i++ {
		for j := 0; j < len(results[i]); j++ {
			f.WriteString(fmt.Sprintf("%d,%d,%t,%d\n", i, j, results[i][j].Error, results[i][j].Time))
		}
	}

	return nil
}

// spawnWorker spawns a worker
func spawnWorker(workerID, duration int, sample *models.Sample, runner models.IScenario, wg *sync.WaitGroup, results [][]*IterationResult) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in worker: ", r)
		}
	}()

	timeout := time.After(time.Duration(duration) * time.Second)

	for {
		select {
		case <-timeout:
			wg.Done()
			return
		default:
			start := time.Now()
			sresult := scenarios.RunIScenario("", "", sample.Scenario, runner)

			result := &IterationResult{
				Time: time.Since(start).Milliseconds(),
			}

			if sresult.Error != nil {
				result.Error = true
			}

			results[workerID] = append(results[workerID], result)
		}
	}
}
