/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"
)

// clusterDeleteCmd represents the clusterDelete command
var clusterDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a cluster",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		clusterName := Name
		command := exec.Command("kind", "delete", "cluster", "--name", clusterName)

		fmt.Println("Deleting cluster with name:", clusterName)
		s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
		s.Start()
		out, err := command.Output()
		if err != nil {
			// if there was any error, print it here
			fmt.Println("could not run command: ", err)
		}
		s.Stop()

		// otherwise, print the output from running the command
		fmt.Println(string(out))

	},
}

func init() {
	clusterCmd.AddCommand(clusterDeleteCmd)
	clusterDeleteCmd.Flags().StringVarP(&Name, "name", "n", "", "Name of the cluster to delete")
	clusterDeleteCmd.MarkFlagRequired("name")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// clusterDeleteCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// clusterDeleteCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
