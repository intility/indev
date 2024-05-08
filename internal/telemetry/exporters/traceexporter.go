package exporters

import (
	"context"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/intility/icpctl/internal/telemetry"
	"github.com/intility/icpctl/internal/ux"
)

type TraceExporter struct {
	filePath string
}

// ensure TraceExporter implements trace.SpanExporter.
var _ sdktrace.SpanExporter = (*TraceExporter)(nil)

func NewTraceExporter() (TraceExporter, error) {
	return TraceExporter{
		filePath: filepath.Join(xdg.StateHome, "icpctl", "traces"),
	}, nil
}

func (t TraceExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	err := telemetry.Trace(spans)
	if err != nil {
		ux.Ferror(os.Stderr, "failed to flush spans: %v\n", err)
		return nil
	}

	return nil
}

func (t TraceExporter) Shutdown(ctx context.Context) error {
	return nil
}
