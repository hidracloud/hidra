package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/hidracloud/hidra/v3/config"
	"github.com/hidracloud/hidra/v3/internal/exporter"
	"github.com/hidracloud/hidra/v3/internal/migrate"
	"github.com/hidracloud/hidra/v3/internal/misc"
	"github.com/hidracloud/hidra/v3/internal/report"
	"github.com/hidracloud/hidra/v3/internal/runner"
	"github.com/hidracloud/hidra/v3/internal/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	log "github.com/sirupsen/logrus"
)

var (
	exitOnError bool
	outputPath  string

	// configNotFoundErr is the error returned when the config file is not found.
	configNotFoundErr = "config file not found"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "hidra",
	Short: "Don't lose your mind monitoring your services. Hidra lends you its head.",
	Long: ` /$$   /$$ /$$       /$$                   
| $$  | $$|__/      | $$                   
| $$  | $$ /$$  /$$$$$$$  /$$$$$$  /$$$$$$ 
| $$$$$$$$| $$ /$$__  $$ /$$__  $$|____  $$
| $$__  $$| $$| $$  | $$| $$  \__/ /$$$$$$$
| $$  | $$| $$| $$  | $$| $$      /$$__  $$
| $$  | $$| $$|  $$$$$$$| $$     |  $$$$$$$
|__/  |__/|__/ \_______/|__/      \_______/	

--  Hidra 2021-2022 license under GPLv3  --`,
}

// exporterCmd represents the exporter command
var exporterCmd = &cobra.Command{
	Use:        "exporter",
	Short:      "Starts the exporter",
	Long:       `Starts the exporter`,
	Args:       cobra.ExactArgs(1),
	ArgAliases: []string{"config_path"},
	Run: func(cmd *cobra.Command, args []string) {
		confPath := args[0]

		log.Infof("Starting exporter with config: %s", confPath)

		exporterConf, err := config.LoadExporterConfigFromFile(confPath)

		if err != nil {
			log.Fatal(configNotFoundErr, err)
		}

		// Set log level
		utils.SetLogLevelFromStr(exporterConf.LogLevel)

		// Set report mode
		if exporterConf.ReportConfig.Enabled {
			report.IsEnabled = true

			if exporterConf.ReportConfig.S3Config.Enabled {
				report.SetS3Configuration(&report.ReportS3Config{
					AccessKeyID:     exporterConf.ReportConfig.S3Config.AccessKeyID,
					SecretAccessKey: exporterConf.ReportConfig.S3Config.SecretAccessKey,
					Endpoint:        exporterConf.ReportConfig.S3Config.Endpoint,
					Bucket:          exporterConf.ReportConfig.S3Config.Bucket,
					Region:          exporterConf.ReportConfig.S3Config.Region,
					ForcePathStyle:  exporterConf.ReportConfig.S3Config.ForcePathStyle,
					UseSSL:          exporterConf.ReportConfig.S3Config.UseSSL,
				})
			}

			if exporterConf.ReportConfig.CallbackConfig.Enabled {
				report.SetCallbackConfiguration(&report.CallbackConfig{
					URL: exporterConf.ReportConfig.CallbackConfig.URL,
				})
			}
		}

		// Start exporter
		exporter.Initialize(exporterConf)
	},
}

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

// migrateCmd represents the test command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate yml from v1-v2 to v3",
	Long:  `Migrate yml from v1-v2 to v3`,
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		exitCode := 0
		log.SetLevel(log.DebugLevel)

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

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Hidra",
	Long:  `All software has versions. This is Hidra's`,
	Run: func(cmd *cobra.Command, args []string) {
		infoTable := [][]string{
			{"Version", misc.Version},
			{"Build date", misc.BuildDate},
			{"Branch", misc.Branch},
			{"Commit", misc.Commit},
		}

		utils.PrintTable(infoTable)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(exporterCmd)

	testCmd.PersistentFlags().BoolVar(&exitOnError, "exit-on-error", false, "Exit with error code 1 if any test fails")
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(verifyCmd)

	migrateCmd.PersistentFlags().StringVar(&outputPath, "output", "", "Output path")
	rootCmd.AddCommand(migrateCmd)
}
