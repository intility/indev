/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"minctl/pkg/cli"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kind/pkg/log"
)

func operation(operation string, duration time.Duration) {
	spinner := cli.NewSpinner(os.Stderr)
	logger := cli.NewLogger(spinner, log.Level(1))
	status := cli.StatusForLogger(logger)
	status.Start(operation)
	time.Sleep(duration)
	status.End(true)
}

// appPortForwardCmd represents the appPortForward command
var appPortForwardCmd = &cobra.Command{
	Use:   "port-forward",
	Short: "Set up port forwarding for the current application",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		command := exec.Command("kubectl", "port-forward", "deployment/"+Deployment, "1234:8080")

		fmt.Println("Forwarding port 8080 -> 1234")

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
	appCmd.AddCommand(appPortForwardCmd)
	appPortForwardCmd.Flags().StringVarP(&Deployment, "deployment", "d", "", "Name of the kubernetes deployment from which to forward the port")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// appPortForwardCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// appPortForwardCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
