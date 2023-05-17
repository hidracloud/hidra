package cmd

import (
	"os"

	"github.com/hidracloud/hidra/v3/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// verify is the command to verify the config
var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify the config",
	Long:  `Verify the config`,
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		exitCode := 0
		log.SetLevel(log.DebugLevel)

		errorCount := 0
		for _, sample := range args {
			// load sample config
			sampleConf, err := config.LoadSampleConfigFromFile(sample)

			if err != nil {
				log.Errorf("❌ Problems loading %s, error: %s", sample, err)
				errorCount++
				continue
			}

			// verify sample config
			err = sampleConf.Verify()

			if err != nil {
				log.Errorf("❌ Problems verifying %s, error: %s", sample, err)
				errorCount++
				continue
			}

			log.Infof("✅ %s verified", sample)
		}

		if errorCount > 0 {
			log.Errorf("❌ %d errors found", errorCount)
			exitCode = 1
		} else {
			log.Infof("✅ No errors found")
		}

		os.Exit(exitCode)
	},
}
