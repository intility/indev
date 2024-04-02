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

// appDeleteCmd represents the delete command
var appDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete application resources from a cluster",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		command := exec.Command("kubectl", "delete", "-f", Path)

		fmt.Println("Deleting resources from cluster...")
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
	appCmd.AddCommand(appDeleteCmd)
	appDeleteCmd.Flags().StringVarP(&Path, "path", "p", "", "Path of the app to delete")
	appDeleteCmd.MarkFlagRequired("path")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// appDeleteCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// appDeleteCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
