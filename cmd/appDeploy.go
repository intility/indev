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

// appDeployCmd represents the deploy command
var appDeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy application resources to a cluster",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		command := exec.Command("kubectl", "apply", "-f", Path)

		fmt.Println("Deploying resources to cluster...")
		s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
		s.Start()
		out, err := command.Output()
		if err != nil {
			fmt.Println("Could not run command: ", err)
		}
		s.Stop()

		fmt.Println(string(out))
	},
}

func init() {
	appCmd.AddCommand(appDeployCmd)
	appDeployCmd.Flags().StringVarP(&Path, "path", "p", "", "Path to the kubernetes manifest files to be deployed")
	appDeployCmd.MarkFlagRequired("path")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// appDeployCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// appDeployCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
