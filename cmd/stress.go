package cmd

import (
	"fmt"
	"time"

	"github.com/hidracloud/hidra/v3/config"
	"github.com/hidracloud/hidra/v3/internal/stress"
	"github.com/hidracloud/hidra/v3/internal/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// stressCmd represents the stress command
var stressCmd = &cobra.Command{
	Use:   "stress",
	Short: "Starts a stress test",
	Long:  `Starts a stress test`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		sample := args[0]

		sampleConf, err := config.LoadSampleConfigFromFile(sample)

		if err != nil {
			log.Fatal(configNotFoundErr, err)
		}

		// Create a channel for receiving the status of the stress test
		status := make(chan stress.Status)

		// parse duration
		duration, err := time.ParseDuration(stressDuration)

		if err != nil {
			log.Fatal(err)
		}

		go stress.Start(&stress.TestArgs{
			SampleConfig: sampleConf,
			StatusChan:   &status,
			Duration:     duration,
			Workers:      stressThreads,
		})

		// while status is not closed

		for s, ok := <-status; ok; s, ok = <-status {
			elapsedTime := time.Since(s.StartTime)

			errorRate := float64(s.Errors) / float64(s.Successes+s.Errors)
			successRate := float64(s.Successes) / float64(s.Successes+s.Errors)
			execRate := float64(s.Successes+s.Errors) / elapsedTime.Seconds()

			infoTable := [][]string{
				{"Total", fmt.Sprintf("%d", s.Successes+s.Errors)},
				{"Total time", elapsedTime.String()},
				{"Errors", fmt.Sprintf("%d", s.Errors)},
				{"Error rate", fmt.Sprintf("%f", errorRate)},
				{"Successes", fmt.Sprintf("%d", s.Successes)},
				{"Success rate", fmt.Sprintf("%f", successRate)},
				{"Execution rate", fmt.Sprintf("%f", execRate)},
				{"Time per sample", fmt.Sprintf("%f", elapsedTime.Seconds()/float64(s.Successes+s.Errors))},
				{"Last error", fmt.Sprintf("%v", s.LastError)},
			}

			fmt.Println("\033[2J")

			utils.PrintTable(infoTable)
		}

	},
}
