package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

const (
	// Default timeout for OTLP exporter connection and shutdown
	defaultExportTimeout = 10 * time.Second
)

var (
	// Prometheus metrics (existing, kept for backward compatibility)
	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"method", "path", "status"})

	httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request duration in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})
)

// otelMetrics holds OTel metric instruments initialized by initOTelMetrics.
type otelMetrics struct {
	requestCount    metric.Int64Counter
	requestDuration metric.Float64Histogram
}

// initOTelMetrics creates OTel metric instruments for HTTP request monitoring.
// Returns a no-op implementation if the meter provider is not set.
func initOTelMetrics() *otelMetrics {
	meter := otel.Meter("villum.http")

	requestCount, err := meter.Int64Counter(
		"otel_http_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
	)
	if err != nil {
		log.Printf("Warning: failed to create OTel request counter: %v", err)
		return &otelMetrics{}
	}

	requestDuration, err := meter.Float64Histogram(
		"otel_http_request_duration_seconds",
		metric.WithDescription("HTTP request duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		log.Printf("Warning: failed to create OTel request duration: %v", err)
		return &otelMetrics{}
	}

	return &otelMetrics{
		requestCount:    requestCount,
		requestDuration: requestDuration,
	}
}

// newOTelMetricsMiddleware returns a Gin middleware that records OTel metrics
// alongside the existing Prometheus metrics.
func newOTelMetricsMiddleware(om *otelMetrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()

		c.Next()

		status := fmt.Sprintf("%d", c.Writer.Status())
		duration := time.Since(start).Seconds()

		// Existing Prometheus metrics
		httpRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)

		// OTel metrics
		if om != nil && om.requestCount != nil {
			om.requestCount.Add(c.Request.Context(), 1,
				metric.WithAttributes(
					attribute.String("http.method", c.Request.Method),
					attribute.String("http.route", path),
					attribute.String("http.status_code", status),
				),
			)
		}
		if om != nil && om.requestDuration != nil {
			om.requestDuration.Record(c.Request.Context(), duration,
				metric.WithAttributes(
					attribute.String("http.method", c.Request.Method),
					attribute.String("http.route", path),
					attribute.String("http.status_code", status),
				),
			)
		}
	}
}

// initTelemetry initializes OpenTelemetry tracing and metrics.
// It selects the trace exporter based on environment variables:
//   - If OTEL_EXPORTER_OTLP_ENDPOINT is set, uses OTLP gRPC (default) or HTTP
//   - Otherwise, falls back to stdout exporter
//
// Returns the tracer provider, Prometheus exporter, and any initialization error.
// A nil tracer provider on error means graceful degradation to no-op tracing.
func initTelemetry() (*sdktrace.TracerProvider, *otelprom.Exporter, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultExportTimeout)
	defer cancel()

	// --- Resource ---
	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "villum"
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
		resource.WithFromEnv(), // reads OTEL_RESOURCE_ATTRIBUTES
	)
	if err != nil {
		return nil, nil, fmt.Errorf("creating resource: %w", err)
	}

	// --- Trace Exporter ---
	traceExporter, err := newTraceExporter(ctx)
	if err != nil {
		log.Printf("Warning: OTel trace exporter init failed (%v), falling back to stdout", err)
		traceExporter, err = newStdoutExporter()
		if err != nil {
			log.Printf("Warning: stdout exporter also failed (%v), running with no-op tracing", err)
			return nil, nil, nil
		}
	}

	// --- Sampler ---
	sampler := newSampler()

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter,
			sdktrace.WithBatchTimeout(5*time.Second),
			sdktrace.WithMaxExportBatchSize(512),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)
	otel.SetTracerProvider(tp)

	// --- Metrics ---
	promExporter, err := otelprom.New()
	if err != nil {
		log.Printf("Warning: failed to create OTel Prometheus exporter: %v", err)
		// Continue without OTel metrics - Prometheus client_golang metrics still work
		return tp, nil, nil
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(promExporter),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(meterProvider)

	return tp, promExporter, nil
}

// newTraceExporter creates the appropriate trace exporter based on env vars.
func newTraceExporter(ctx context.Context) (sdktrace.SpanExporter, error) {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		return newStdoutExporter()
	}

	protocol := os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL")

	switch protocol {
	case "http/protobuf":
		return newOTLPHTTPExporter(ctx, endpoint)
	default:
		// Default to gRPC
		return newOTLPRPCExporter(ctx, endpoint)
	}
}

// newOTLPRPCExporter creates an OTLP gRPC trace exporter.
func newOTLPRPCExporter(ctx context.Context, endpoint string) (sdktrace.SpanExporter, error) {
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpointURL(endpoint),
		otlptracegrpc.WithTimeout(defaultExportTimeout),
	}
	return otlptracegrpc.New(ctx, opts...)
}

// newOTLPHTTPExporter creates an OTLP HTTP/protobuf trace exporter.
func newOTLPHTTPExporter(ctx context.Context, endpoint string) (sdktrace.SpanExporter, error) {
	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpointURL(endpoint),
		otlptracehttp.WithTimeout(defaultExportTimeout),
	}
	return otlptracehttp.New(ctx, opts...)
}

// newStdoutExporter creates a stdout trace exporter for development.
func newStdoutExporter() (sdktrace.SpanExporter, error) {
	return stdouttrace.New(
		stdouttrace.WithPrettyPrint(),
	)
}

// newSampler creates a trace sampler based on OTEL_TRACES_SAMPLER env var.
func newSampler() sdktrace.Sampler {
	samplerName := os.Getenv("OTEL_TRACES_SAMPLER")
	samplerArg := os.Getenv("OTEL_TRACES_SAMPLER_ARG")

	switch samplerName {
	case "always_on":
		return sdktrace.AlwaysSample()
	case "always_off":
		return sdktrace.NeverSample()
	case "traceidratio":
		ratio := 1.0
		if samplerArg != "" {
			if parsed, err := parseRatio(samplerArg); err == nil {
				ratio = parsed
			}
		}
		return sdktrace.TraceIDRatioBased(ratio)
	case "parentbased_traceidratio":
		ratio := 1.0
		if samplerArg != "" {
			if parsed, err := parseRatio(samplerArg); err == nil {
				ratio = parsed
			}
		}
		return sdktrace.ParentBased(sdktrace.TraceIDRatioBased(ratio))
	default:
		return sdktrace.AlwaysSample()
	}
}

// parseRatio parses a float ratio from string, returning an error for invalid values.
func parseRatio(s string) (float64, error) {
	var ratio float64
	if _, err := fmt.Sscanf(s, "%f", &ratio); err != nil {
		return 0, fmt.Errorf("invalid ratio: %w", err)
	}
	if ratio < 0 || ratio > 1 {
		return 0, fmt.Errorf("ratio %f out of range [0,1]", ratio)
	}
	return ratio, nil
}

// shutdownTelemetry flushes and shuts down the tracer provider gracefully.
func shutdownTelemetry(tp *sdktrace.TracerProvider) {
	if tp == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultExportTimeout)
	defer cancel()
	if err := tp.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down tracer provider: %v", err)
	}
}
