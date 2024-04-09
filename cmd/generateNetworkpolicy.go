/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// generateNetworkpolicyCmd represents the generateNetworkpolicy command
var generateNetworkpolicyCmd = &cobra.Command{
	Use:   "networkpolicy",
	Short: "Generate kubernetes network policies for the current project",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("generateNetworkpolicy called")
	},
}

func init() {
	generateCmd.AddCommand(generateNetworkpolicyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// generateNetworkpolicyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// generateNetworkpolicyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
