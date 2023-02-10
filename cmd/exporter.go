package cmd

import (
	"github.com/hidracloud/hidra/v3/config"
	"github.com/hidracloud/hidra/v3/internal/exporter"
	"github.com/hidracloud/hidra/v3/internal/report"
	"github.com/hidracloud/hidra/v3/internal/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

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
