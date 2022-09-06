package cmd

import (
	"os"

	"github.com/hidracloud/hidra/v3/internal/config"
	"github.com/hidracloud/hidra/v3/internal/exporter"
	"github.com/hidracloud/hidra/v3/internal/utils"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
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

		log.Infof("Starting exporter with config: %s")

		exporterConf, err := config.LoadExporterConfigFromFile(confPath)

		if err != nil {
			log.Fatal("Error loading config: ", err)
		}

		// Set log level
		utils.SetLogLevelFromStr(exporterConf.LogLevel)

		// Start exporter
		exporter.Initialize(exporterConf)
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
}
