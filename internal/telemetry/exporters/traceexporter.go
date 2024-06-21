package exporters

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/intility/icpctl/internal/telemetry/exporters/tracetransform"
	coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/protobuf/proto"
	"io/fs"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/adrg/xdg"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/intility/icpctl/internal/ux"
)

const (
	permissionStateFile = 0o600
	permissionStateDir  = 0o700
)

var needsFlush atomic.Bool

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
	needsFlush.Store(true)

	err := t.saveSpans(spans)
	if err != nil {
		ux.Ferror(os.Stderr, "failed to flush spans: %v\n", err)
		return nil
	}

	return nil
}

func (t TraceExporter) Shutdown(ctx context.Context) error {
	return nil
}

func (t TraceExporter) saveSpans(spans []sdktrace.ReadOnlySpan) error {
	protoSpans := tracetransform.Spans(spans)
	if len(protoSpans) == 0 {
		return nil
	}

	req := &coltracepb.ExportTraceServiceRequest{
		ResourceSpans: protoSpans,
	}

	bytes, err := proto.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal trace: %w", err)
	}

	uid := uuid.New().String()
	fPath := filepath.Join(t.filePath, uid+".json")

	err = os.WriteFile(fPath, bytes, permissionStateFile)
	if errors.Is(err, fs.ErrNotExist) {
		// XDG specifies perms 0700.
		if err = os.MkdirAll(filepath.Dir(fPath), permissionStateDir); err != nil {
			return fmt.Errorf("failed to create trace buffer directory: %w", err)
		}

		err = os.WriteFile(fPath, bytes, permissionStateFile)
	}

	if err != nil {
		return fmt.Errorf("failed to flush trace to disk: %w", err)
	}

	return nil
}

func (t TraceExporter) RestoreTraces() []*coltracepb.ExportTraceServiceRequest {
	dirEntries, err := os.ReadDir(t.filePath)
	if err != nil {
		return []*coltracepb.ExportTraceServiceRequest{}
	}

	traces := make([]*coltracepb.ExportTraceServiceRequest, 0, len(dirEntries))

	for _, entry := range dirEntries {
		if !entry.Type().IsRegular() {
			continue
		}

		path := filepath.Join(t.filePath, entry.Name())
		data, err := os.ReadFile(path)

		// Always delete the file, so we don't end up with an infinitely growing
		// backlog of traces.
		_ = os.Remove(path)

		if err != nil {
			continue
		}

		var trace coltracepb.ExportTraceServiceRequest
		if err := proto.Unmarshal(data, &trace); err != nil {
			continue
		}

		traces = append(traces, &trace)
	}

	return traces

}
