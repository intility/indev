/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// generateDockerfileCmd represents the generateDockerfile command
var generateDockerfileCmd = &cobra.Command{
	Use:   "dockerfile",
	Short: "Generate a Dockerfile for the current project",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if fileExists("Dockerfile") {
			fmt.Println("Dockerfile already exists")
			os.Exit(1)
		}

		// prompt := "Generate a Dockerfile for the current project"
		// command := exec.Command("cwc", "apply", "-f")

		fmt.Println("Generating dockerfile...")
		// s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
		// s.Start()
		// out, err := command.Output()
		// if err != nil {
		// 	fmt.Println("Could not run command: ", err)
		// }
		// s.Stop()

		// fmt.Println(string(out))
	},
}

func init() {
	generateCmd.AddCommand(generateDockerfileCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// generateDockerfileCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// generateDockerfileCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
