package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/intility/minctl/pkg/cmderrors"
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
		fmt.Println(err)

		var usageErr cmderrors.InvalidUsageError
		if errors.As(err, &usageErr) {
			_ = usageErr.Cmd.Usage()
		}

		os.Exit(1)
	}
}

// init initializes the root command and flags.
func init() {
	// use InvalidUsageErrors for flag errors to trigger usage output
	rootCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		return cmderrors.NewInvalidUsageError(cmd, err.Error())
	})

	rootCmd.CompletionOptions.HiddenDefaultCmd = false
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
}
