/*
Copyright © 2024 Callum Powell <callum.powell@intility.no>
*/
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"

	"github.com/intility/icpctl/internal/telemetry"
	"github.com/intility/icpctl/internal/telemetry/exporters"
	"github.com/intility/icpctl/internal/ux"
	"github.com/intility/icpctl/pkg/rootcommand"
)

//go:generate ../../scripts/completions.sh ../../

func main() {
	args := os.Args[1:]
	if len(args) == 1 && args[0] == "upload-telemetry" {
		err := uploadTelemetry(context.Background())
		if err != nil {
			ux.Ferror(os.Stderr, err.Error()+"\n")
			os.Exit(1)
		}
		return
	}

	err := run(os.Args[1:])
	if err != nil {
		os.Exit(1)
	}
}

func run(args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)

	go func() {
		if <-sigChan; true {
			cancel()
		}
	}()

	cmd := rootcommand.GetRootCommand()

	tracer, shutdown, _ := telemetry.InitTracer(ctx, semconv.ProcessCommandArgs(args...))
	ctx = telemetry.ContextWithTracer(ctx, tracer)

	defer func() { _ = shutdown(ctx) }()

	ctx, span := telemetry.StartSpan(ctx, "root")
	defer span.End()

	err := cmd.ExecuteContext(ctx)
	if err != nil {
		// We can introduce warnings here if needed.
		switch {
		case errors.Is(err, context.Canceled):
			ux.Fprint(cmd.OutOrStdout(), "Operation was canceled.")
		default:
			span.RecordError(err)
			ux.Ferror(cmd.ErrOrStderr(), err.Error()+"\n")
		}
	}

	scheduleTelemetryUpload(ctx, args)

	return err //nolint:wrapcheck
}

func scheduleTelemetryUpload(ctx context.Context, args []string) {
	// prevent fork-bombing
	if len(args) == 0 || args[0] == "upload-telemetry" {
		return
	}

	ctx, span := telemetry.StartSpan(ctx, "telemetry.scheduleUpload")
	defer span.End()

	exe, err := os.Executable()
	if err == nil {
		_ = exec.CommandContext(ctx, exe, "upload-telemetry").Start()
	}
}

func uploadTelemetry(ctx context.Context) error {
	sourceAdapter, err := exporters.NewTraceExporter()
	if err != nil {
		return fmt.Errorf("failed to create trace exporter: %w", err)
	}

	uploader := telemetry.NewTraceUploader(sourceAdapter)

	err = uploader.Upload(ctx)
	if err != nil {
		return fmt.Errorf("failed to upload telemetry data: %w", err)
	}

	return nil
}
