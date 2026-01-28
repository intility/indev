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

	"github.com/intility/indev/internal/telemetry"
	"github.com/intility/indev/internal/telemetry/exporters"
	"github.com/intility/indev/internal/ux"
	"github.com/intility/indev/pkg/rootcommand"
)

const uploadTelemetryCommand = "upload-telemetry"

//go:generate ../../scripts/completions.sh ../../

func main() {
	args := os.Args[1:]
	if len(args) == 1 && args[0] == uploadTelemetryCommand {
		err := uploadTelemetry(context.Background())
		if err != nil {
			ux.Ferrorf(os.Stderr, "%s\n", err.Error())
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

	tracer, shutdown, _ := telemetry.InitTracer(ctx, semconv.ProcessCommandArgs(args...))
	ctx = telemetry.ContextWithTracer(ctx, tracer)

	defer func() { _ = shutdown(ctx) }()

	ctx, span := telemetry.StartSpan(ctx, "root")
	defer span.End()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)

	go func() {
		span.AddEvent("cancellation-signal-listening")

		if <-sigChan; true {
			span.AddEvent("cancellation-signal-received")
			cancel()
		}
	}()

	cmd := rootcommand.GetRootCommand()
	err := cmd.ExecuteContext(ctx)

	span.AddEvent("command-execution-finished")

	if err != nil {
		// We can introduce warnings here if needed.
		switch {
		case errors.Is(err, context.Canceled):
			ux.Fprintf(cmd.OutOrStdout(), "Operation was canceled.")
		default:
			span.RecordError(err)
			ux.Ferrorf(cmd.ErrOrStderr(), "%s\n", err.Error())
		}
	}

	scheduleTelemetryUpload(ctx, args)
	span.AddEvent("telemetry-upload-process-forked")

	return err //nolint:wrapcheck
}

func scheduleTelemetryUpload(ctx context.Context, args []string) {
	// prevent fork-bomb
	if len(args) == 0 || args[0] == uploadTelemetryCommand {
		return
	}

	exe, err := os.Executable()
	if err == nil {
		_ = exec.CommandContext(ctx, exe, uploadTelemetryCommand).Start()
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
