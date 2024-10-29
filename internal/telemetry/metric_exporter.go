package telemetry

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"sync/atomic"

	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/intility/idpctl/internal/redact"
)

type MetricFileExporter struct {
	encVal atomic.Value

	shutdownOnce sync.Once

	temporalitySelector metric.TemporalitySelector
	aggregationSelector metric.AggregationSelector

	redactTimestamps bool
}

func NewMetricFileExporter() *MetricFileExporter {
	exporter := &MetricFileExporter{ //nolint:exhaustruct
		temporalitySelector: metric.DefaultTemporalitySelector,
		aggregationSelector: metric.DefaultAggregationSelector,
		redactTimestamps:    false,
	}

	encoder := json.NewEncoder(os.Stdout)
	exporter.encVal.Store(encoderHolder{encoder: encoder})

	return exporter
}

func (m *MetricFileExporter) Temporality(k metric.InstrumentKind) metricdata.Temporality {
	return m.temporalitySelector(k)
}

func (m *MetricFileExporter) Aggregation(k metric.InstrumentKind) metric.Aggregation { //nolint:ireturn
	return m.aggregationSelector(k)
}

func (m *MetricFileExporter) Export(ctx context.Context, data *metricdata.ResourceMetrics) error {
	select {
	case <-ctx.Done():
		return ctx.Err() //nolint:wrapcheck
	default: // continue
	}

	enc, ok := m.encVal.Load().(encoderHolder)
	if !ok {
		return redact.Errorf("invalid encoder type: %T", redact.Safe(enc.encoder))
	}

	return enc.Encode(data)
}

func (m *MetricFileExporter) ForceFlush(ctx context.Context) error {
	// nothing to flush
	return ctx.Err() //nolint:wrapcheck
}

// Shutdown replaces the encoder with a shutdownEncoder, which always returns
// errShutdown when Encode is called.
func (m *MetricFileExporter) Shutdown(ctx context.Context) error {
	m.shutdownOnce.Do(func() {
		m.encVal.Store(shutdownEncoder{})
	})

	return ctx.Err() //nolint:wrapcheck
}

// Encoder encodes and outputs OpenTelemetry metric data-types as human-readable text.
type Encoder interface {
	Encode(v any) error
}

// encoderHolder is the concrete type used to wrap an Encoder, so it can be
// used as a atomic.Value type.
type encoderHolder struct {
	encoder Encoder
}

func (e encoderHolder) Encode(v any) error {
	return e.encoder.Encode(v) //nolint:wrapcheck
}

// shutdownEncoder is used when the exporter is shutdown. It always returns
// errShutdown when Encode is called.
type shutdownEncoder struct{}

var errShutdown = redact.Errorf("%s", redact.Safe("exporter is shutdown"))

func (shutdownEncoder) Encode(any) error { return errShutdown }
