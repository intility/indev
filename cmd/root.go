package cmd

import (
	"context"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/intility/minctl/internal/pipeline"
	"github.com/intility/minctl/internal/telemetry"
)

const telemetryUploadTimeout = time.Second * 5

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "icpctl",
	Short: "",
	Long:  ``,
}

// RunPipeline adds all child commands to the root command and sets flags appropriately.
// This is called by Execute(). It only needs to happen once to the rootCmd.
func RunPipeline(ctx context.Context) int {
	pipe := pipeline.New(rootCmd)
	pipe.AddMiddleware(pipeline.Telemetry())
	pipe.AddMiddleware(pipeline.Logger())

	return pipe.Execute(ctx, os.Args[1:])
}

func Execute() {
	ctx := context.Background()

	if len(os.Args) > 1 && os.Args[1] == "upload-telemetry" {
		// This subcommand is hidden and only run by minctl itself as a
		// child process. We need to really make sure that we always
		// exit and don't leave orphaned processes lying around.
		time.AfterFunc(telemetryUploadTimeout, func() {
			os.Exit(0)
		})

		telemetry.Upload()

		return
	}

	exitCode := RunPipeline(ctx)
	os.Exit(exitCode)
}

// init initializes the root command and flags.
func init() {
	rootCmd.CompletionOptions.HiddenDefaultCmd = false
	rootCmd.SilenceUsage = false
	rootCmd.SilenceErrors = true
}
