package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	exitOnError    bool
	outputPath     string
	stressDuration string
	stressThreads  int

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

--  Hidra 2021-2023 license under GPLv3  --`,
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

	stressCmd.PersistentFlags().StringVar(&stressDuration, "duration", "60s", "Duration of the stress test")
	stressCmd.PersistentFlags().IntVar(&stressThreads, "threads", 1000, "Number of threads to use")
	rootCmd.AddCommand(stressCmd)

	migrateCmd.PersistentFlags().StringVar(&outputPath, "output", "", "Output path")
	rootCmd.AddCommand(migrateCmd)
}
