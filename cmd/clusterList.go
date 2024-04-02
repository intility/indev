/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

// clusterListCmd represents the list command
var clusterListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all clusters",
	Long:  `List all clusters that are currently running in kind on the local machine.`,
	Run: func(cmd *cobra.Command, args []string) {
		command := exec.Command("kind", "get", "clusters")

		out, err := command.Output()
		if err != nil {
			// if there was any error, print it here
			fmt.Println("could not run command: ", err)
		}

		// otherwise, print the output from running the command
		fmt.Println(string(out))
	},
}

func init() {
	clusterCmd.AddCommand(clusterListCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
