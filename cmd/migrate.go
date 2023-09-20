package cmd

import (
	"os"

	"github.com/hidracloud/hidra/v3/internal/migrate"
	"github.com/hidracloud/hidra/v3/internal/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// migrateCmd represents the test command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate yml from v1-v2 to v3",
	Long:  `Migrate yml from v1-v2 to v3`,
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		exitCode := 0
		log.SetLevel(log.DebugLevel)
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp: true,
		})

		// Check if output path exists
		if outputPath == "" {
			log.Fatal("Output path is required")
		}

		for _, sample := range args {
			// file name
			fileName := utils.ExtractFileNameWithoutExtension(sample) + ".yml"
			// load sample config
			oldSample, err := migrate.LoadSampleV1V2ConfigFromFile(sample)

			if err != nil {
				log.Fatal("Error loading config: ", err)
			}

			// migrate sample config
			newSample := oldSample.Migrate()

			// save newSample as yaml into outputPath
			b, err := yaml.Marshal(newSample)

			if err != nil {
				log.Fatal("Error marshaling config: ", err)
			}

			// write b to outputPath/fileName
			err = os.WriteFile(outputPath+"/"+fileName, b, 0644)

			if err != nil {
				log.Fatal("Error writing config: ", err)
			}

		}

		os.Exit(exitCode)
	},
}
