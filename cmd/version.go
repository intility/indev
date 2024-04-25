package cmd

import (
	"github.com/spf13/cobra"

	"github.com/intility/minctl/internal/build"
	"github.com/intility/minctl/internal/ux"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Long:  `Print the version information of icpctl.`,
	Run: func(cmd *cobra.Command, args []string) {
		ux.Fprint(cmd.OutOrStdout(), build.NameVersionString()+"\n")
	},
}
