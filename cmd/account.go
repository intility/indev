/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// accountCmd represents the account command.
var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "Manage account information",
	Long:  `Manage account information`,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := cmd.Help()
		if err != nil {
			return fmt.Errorf("could not run help command: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(accountCmd)
}
