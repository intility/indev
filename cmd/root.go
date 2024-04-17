package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var Name string

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "minctl",
	Short: "",
	Long:  ``,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		// lipgloss will handle TTY detection and color support
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
		_, _ = fmt.Fprintf(os.Stderr, "%s %s", style.Render("ERROR"), err.Error())
		os.Exit(1)
	}
}

// init initializes the root command and flags.
func init() {
	rootCmd.CompletionOptions.HiddenDefaultCmd = false
	rootCmd.SilenceUsage = false
	rootCmd.SilenceErrors = true
}
