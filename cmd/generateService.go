/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// generateServiceCmd represents the generateService command
var generateServiceCmd = &cobra.Command{
	Use:   "service",
	Short: "Generate a kubernetes service for the current project",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("generateService called")
	},
}

func init() {
	generateCmd.AddCommand(generateServiceCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// generateServiceCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// generateServiceCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
