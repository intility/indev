/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"
)

// clusterDeleteCmd represents the clusterDelete command.
var clusterDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a cluster",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		clusterName := Name
		command := exec.Command("kind", "delete", "cluster", "--name", clusterName)
		spin := spinner.New(spinner.CharSets[11], spinnerDelay)

		spin.Start()

		out, err := command.Output()
		if err != nil {
			return fmt.Errorf("could not run command: %w", err)
		}

		spin.Stop()

		// otherwise, print the output from running the command
		fmt.Println(string(out))

		return nil
	},
}

func init() {
	clusterCmd.AddCommand(clusterDeleteCmd)
	clusterDeleteCmd.Flags().StringVarP(&Name, "name", "n", "", "Name of the cluster to delete")

	err := clusterDeleteCmd.MarkFlagRequired("name")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
