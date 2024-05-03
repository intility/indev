package pipeline

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"

	"github.com/intility/icpctl/internal/build"
	"github.com/intility/icpctl/internal/env"
	"github.com/intility/icpctl/internal/redact"
)

func Metrics() *MetricsMiddleware {
	return &MetricsMiddleware{}
}

type MetricsMiddleware struct{}

var _ Middleware = (*MetricsMiddleware)(nil)

func (m *MetricsMiddleware) Handle(cmd *cobra.Command, args []string, next NextFunc) error {
	ctx := cmd.Context()

	subCmd, flags, err := getSubcommand(cmd, args)
	if err != nil {
		return next(cmd, args)
	}

	cmdPath := strings.Split(subCmd.CommandPath(), " ")
	metricReader := metricsdk.NewManualReader()

	provider, err := m.initMeterProvider(
		ctx,
		metricReader,
		semconv.ProcessCommandArgs(cmdPath...),
		attribute.StringSlice("flags", flags),
	)
	if err != nil {
		return next(cmd, args)
	}

	// register the command meter counter
	meter := otel.Meter("command")

	counter, _ := meter.Int64Counter(
		"executed",
		metric.WithDescription("Number of times a command was executed"),
		metric.WithUnit("{call}"),
	)

	// increment the counter
	subCmdName := strings.Join(cmdPath, ".")
	counter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("command", subCmdName),
	))

	defer func() {
		err = provider.Shutdown(context.Background())
		if err != nil {
			log.Fatalf("failed to shutdown meter tracerProvider: %v", redact.Safe(err))
		}
	}()

	histogram, _ := meter.Float64Histogram(
		"duration",
		metric.WithDescription("Duration of command execution"),
		metric.WithUnit("ms"),
	)

	startTime := time.Now()
	runErr := next(cmd, args)

	histogram.Record(cmd.Context(), float64(time.Since(startTime).Milliseconds()))

	_ = m.exportMetrics(ctx, metricReader)

	return runErr
}

func (m *MetricsMiddleware) initMeterProvider(
	ctx context.Context,
	reader metricsdk.Reader,
	attrs ...attribute.KeyValue,
) (*metricsdk.MeterProvider, error) {
	var (
		appName     = build.AppName
		appVersion  = build.Version
		otelVersion = "v1.26.0"
	)

	attrs = append(attrs,
		semconv.ServiceNameKey.String(appName),
		semconv.ServiceVersionKey.String(appVersion),
		semconv.TelemetrySDKVersionKey.String(otelVersion),
		semconv.TelemetrySDKLanguageGo,
	)

	rsrs, err := resource.New(ctx, resource.WithAttributes(attrs...))
	if err != nil {
		return nil, redact.Errorf("failed to create resource: %w", redact.Safe(err))
	}

	// initialize the metrics tracerProvider
	provider := metricsdk.NewMeterProvider(metricsdk.WithResource(rsrs), metricsdk.WithReader(reader))

	otel.SetMeterProvider(provider)

	return provider, nil
}

func (m *MetricsMiddleware) exportMetrics(ctx context.Context, reader metricsdk.Reader) error {
	exporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(env.OtelExporterEndpoint()),
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithHeaders(map[string]string{
			"Authorization": "Bearer " + env.OtelExporterToken(),
		}),
	)
	if err != nil {
		return redact.Errorf("failed to create OTLP exporter: %w", redact.Safe(err))
	}

	collectedMetrics := &metricdata.ResourceMetrics{} //nolint:exhaustruct
	_ = reader.Collect(ctx, collectedMetrics)

	err = exporter.Export(ctx, collectedMetrics)
	if err != nil {
		return redact.Errorf("failed to export metrics: %w", redact.Safe(err))
	}

	return nil
}
