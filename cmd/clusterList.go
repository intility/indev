/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

// clusterListCmd represents the list command.
var clusterListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all clusters",
	Long:  `List all clusters that are currently running in kind on the local machine.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		command := exec.Command("kind", "get", "clusters")

		stdErrBuffer := &bytes.Buffer{}
		command.Stderr = stdErrBuffer

		out, err := command.Output()
		if err != nil {
			// get content from err buffer
			cmd.SilenceUsage = true
			return fmt.Errorf("could not run command: %w: %s", err, stdErrBuffer.String())
		}

		// otherwise, print the output from running the command
		fmt.Println(string(out))

		return nil
	},
}

func init() {
	clusterCmd.AddCommand(clusterListCmd)
}
