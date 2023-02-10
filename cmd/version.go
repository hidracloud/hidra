package cmd

import (
	"github.com/hidracloud/hidra/v3/internal/misc"
	"github.com/hidracloud/hidra/v3/internal/utils"
	"github.com/spf13/cobra"
)

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
