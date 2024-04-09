/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

// cwcCmd represents the cwc command
var cwcCmd = &cobra.Command{
	Use:   "cwc",
	Short: "Runs cwc within minctl",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		command := exec.Command("cwc")

		out, err := command.Output()
		if err != nil {
			fmt.Println("Could not run command: ", err)
		}

		fmt.Println(string(out))
	},
}

func init() {
	rootCmd.AddCommand(cwcCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cwcCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// cwcCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
