package cmd

import (
	"github.com/spf13/cobra"
)

// showCmd represents the show command.
var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show account information",
	Long:  `Show information about the current logged in account.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		panic("not implemented")
	},
}

func init() {
	accountCmd.AddCommand(showCmd)
}
