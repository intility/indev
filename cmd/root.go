package cmd

import (
	"context"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/intility/icpctl/internal/pipeline"
	"github.com/intility/icpctl/internal/telemetry"
)

const telemetryUploadTimeout = time.Second * 30

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "icpctl",
	Short: "",
	Long:  ``,
}

// RunPipeline adds all child commands to the root command and sets flags appropriately.
// It only needs to happen once to the rootCmd.
func RunPipeline(ctx context.Context, args []string) int {
	pipe := pipeline.New(rootCmd)
	// pipe.AddMiddleware(pipeline.Metrics())
	pipe.AddMiddleware(pipeline.Telemetry())
	pipe.AddMiddleware(pipeline.Trace())
	pipe.AddMiddleware(pipeline.Logger())

	return pipe.Execute(ctx, args[1:])
}

func Execute(ctx context.Context, args []string) int {
	if len(args) > 1 && args[1] == "upload-telemetry" {
		// This subcommand is hidden and only run by icpctl itself as a
		// child process. We need to really make sure that we always
		// exit and don't leave orphaned processes lying around.
		ctx, cancel := context.WithCancel(ctx)

		time.AfterFunc(telemetryUploadTimeout, func() {
			cancel()
			os.Exit(0)
		})

		telemetry.Upload(ctx)

		return 0
	}

	return RunPipeline(ctx, args)
}

// init initializes the root command and flags.
func init() {
	rootCmd.CompletionOptions.HiddenDefaultCmd = false
	rootCmd.SilenceUsage = false
	rootCmd.SilenceErrors = true
}
