package telemetry

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"

	"github.com/intility/idpctl/internal/env"
	"github.com/intility/idpctl/internal/ux"
)

const uploadTimeout = 5 * time.Second

type TraceRestorer interface {
	RestoreTraces() []*coltracepb.ExportTraceServiceRequest
}

type TraceUploader struct {
	restorer TraceRestorer
	timeout  time.Duration
}

func NewTraceUploader(restorer TraceRestorer) *TraceUploader {
	return &TraceUploader{
		restorer: restorer,
		timeout:  uploadTimeout,
	}
}

func (t *TraceUploader) Upload(ctx context.Context) error {
	traces := t.restorer.RestoreTraces()

	if len(traces) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	endpoint := env.OtelExporterEndpoint()

	ux.Finfo(os.Stdout, "uploading %d traces to %s\n", len(traces), endpoint)

	client := otlptracehttp.NewClient(
		otlptracehttp.WithEndpointURL(endpoint),
	)

	err := client.Start(ctx)
	if err != nil {
		ux.Ferror(os.Stderr, "failed to start client: %v\n", err)

		return fmt.Errorf("failed to start client: %w", err)
	}

	ux.Finfo(os.Stdout, "flushing %d traces\n", len(traces))

	for _, trace := range traces {
		spans := trace.GetResourceSpans()

		err = client.UploadTraces(ctx, spans)
		if err != nil {
			ux.Ferror(os.Stderr, "failed to upload traces: %v\n", err)
		}
	}

	err = client.Stop(ctx)
	if err != nil {
		ux.Ferror(os.Stderr, "failed to stop client: %v\n", err)
	}

	return nil
}
