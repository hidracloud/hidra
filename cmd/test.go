package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/hidracloud/hidra/v3/config"
	"github.com/hidracloud/hidra/v3/internal/runner"
	"github.com/hidracloud/hidra/v3/internal/utils"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Starts the test",
	Long:  `Starts the test`,
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		exitCode := 0
		log.SetLevel(log.DebugLevel)
		if !utils.IsHeadless() {
			os.Setenv("BROWSER_NO_HEADLESS", "1")
			log.Debug("Setting up browser in headless mode")
		}

		for _, sample := range args {
			// Load sample config
			sampleConf, err := config.LoadSampleConfigFromFile(sample)

			if err != nil {
				log.Fatal(configNotFoundErr, err)
			}

			ctx := context.TODO()

			result := runner.RunSample(ctx, sampleConf)

			if result.Error != nil {
				exitCode = 1
			}

			resultEmoji := "✅"
			if result.Error != nil {
				resultEmoji = "❌"
			}

			infoTable := [][]string{
				{"Sample", sample},
				{"Error", fmt.Sprintf("%v", result.Error)},
				{"Result", resultEmoji},
			}

			for _, metric := range result.Metrics {
				infoTable = append(infoTable, []string{fmt.Sprintf("%s (%s) (%v)", metric.Description, metric.Name, metric.Labels), fmt.Sprintf("%f", metric.Value)})
			}

			utils.PrintTable(infoTable)
			fmt.Println()
			fmt.Println()

			if exitCode != 0 && exitOnError {
				log.Fatal("Exit on error enabled, one test failed")
			}
		}

		os.Exit(exitCode)
	},
}
