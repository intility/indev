/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// generateDeploymentCmd represents the generateDeployment command
var generateDeploymentCmd = &cobra.Command{
	Use:   "deployment",
	Short: "Generate a kubernetes deployment for the current project",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("generateDeployment called")
	},
}

func init() {
	generateCmd.AddCommand(generateDeploymentCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// generateDeploymentCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// generateDeploymentCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
