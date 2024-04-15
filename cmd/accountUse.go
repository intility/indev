/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// useCmd represents the use command.
var useCmd = &cobra.Command{
	Use:   "use",
	Short: "Switch to account",
	Long:  `Switch to a currently logged in account. Use 'list' to see available accounts.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		panic("not implemented")
	},
}

func init() {
	accountCmd.AddCommand(useCmd)
}
