package cmd

import (
	"github.com/spf13/cobra"
)

// listCmd represents the list command.
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all accounts",
	Long:  `List all accounts that are available to use.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		panic("not implemented")
	},
}

func init() {
	accountCmd.AddCommand(listCmd)
}
