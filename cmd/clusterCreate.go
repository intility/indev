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

// clusterCreateCmd represents the create command
var clusterCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new cluster",
	Long:  `Create a new cluster with the specified configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		// var clusterName string

		// clusterName = Name
		// if Name == "" {
		// 	clusterName = uuid.NewString()[:6]
		// }

		// command := exec.Command("kind", "create", "cluster", "--name", clusterName)

		// fmt.Println("Creating cluster with name: kind-" + clusterName)
		command := exec.Command("init-kind.sh")

		fmt.Println("Creating cluster...")
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
	clusterCmd.AddCommand(clusterCreateCmd)
	clusterCreateCmd.Flags().StringVarP(&Name, "name", "n", "", "Name of the cluster to create")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
