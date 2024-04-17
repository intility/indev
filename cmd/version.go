package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of minctl",
	Long:  `All software has versions. This is minctl's`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("minctl 0.0.1")
	},
}
