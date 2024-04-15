/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command.
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an account",
	Long:  `Delete an account`,
	RunE: func(cmd *cobra.Command, args []string) error {
		panic("not implemented")
	},
}

func init() {
	accountCmd.AddCommand(deleteCmd)
}
