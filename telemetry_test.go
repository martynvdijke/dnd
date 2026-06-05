package main

import (
	"context"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// ─── parseRatio ───

func TestParseRatio(t *testing.T) {
	tests := []struct {
		input   string
		want    float64
		wantErr bool
	}{
		{"0.5", 0.5, false},
		{"1.0", 1.0, false},
		{"0.0", 0.0, false},
		{"0.25", 0.25, false},
		{"0.75", 0.75, false},
		{"-0.1", 0, true},
		{"1.5", 0, true},
		{"abc", 0, true},
		{"", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseRatio(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("got %f, want %f", got, tt.want)
			}
		})
	}
}

// ─── newSampler ───

func TestNewSampler(t *testing.T) {
	tests := []struct {
		name    string
		sampler string
		arg     string
		desc    string
	}{
		{"default (no env)", "", "", "AlwaysOn"},
		{"always_on", "always_on", "", "AlwaysOn"},
		{"always_off", "always_off", "", "AlwaysOff"},
		{"traceidratio with arg", "traceidratio", "0.5", "TraceIDRatioBased"},
		{"traceidratio default arg", "traceidratio", "", "TraceIDRatioBased"},
		{"parentbased_traceidratio", "parentbased_traceidratio", "0.25", "ParentBased"},
		{"unknown sampler", "invalid_sampler", "", "AlwaysOn"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.sampler != "" {
				os.Setenv("OTEL_TRACES_SAMPLER", tt.sampler)
				defer os.Unsetenv("OTEL_TRACES_SAMPLER")
			} else {
				os.Unsetenv("OTEL_TRACES_SAMPLER")
			}
			if tt.arg != "" {
				os.Setenv("OTEL_TRACES_SAMPLER_ARG", tt.arg)
				defer os.Unsetenv("OTEL_TRACES_SAMPLER_ARG")
			} else {
				os.Unsetenv("OTEL_TRACES_SAMPLER_ARG")
			}

			sampler := newSampler()
			desc := sampler.Description()
			if !strings.HasPrefix(desc, tt.desc) {
				t.Errorf("Description() = %q, want prefix %q", desc, tt.desc)
			}
		})
	}
}

// ─── newStdoutExporter ───

func TestNewStdoutExporter(t *testing.T) {
	exporter, err := newStdoutExporter()
	if err != nil {
		t.Fatalf("newStdoutExporter() error: %v", err)
	}
	if exporter == nil {
		t.Fatal("newStdoutExporter() returned nil")
	}

	if _, ok := exporter.(*stdouttrace.Exporter); !ok {
		t.Errorf("expected *stdouttrace.Exporter, got %T", exporter)
	}
}

// ─── newTraceExporter (stdout fallback) ───

func TestNewTraceExporterStdoutFallback(t *testing.T) {
	os.Unsetenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	os.Unsetenv("OTEL_EXPORTER_OTLP_PROTOCOL")

	ctx := context.Background()
	exporter, err := newTraceExporter(ctx)
	if err != nil {
		t.Fatalf("newTraceExporter() error: %v", err)
	}
	if exporter == nil {
		t.Fatal("newTraceExporter() returned nil")
	}

	// Should be a stdout exporter when endpoint is empty
	if _, ok := exporter.(*stdouttrace.Exporter); !ok {
		t.Errorf("expected stdout exporter, got %T", exporter)
	}
}

// ─── initTelemetry with stdout (no OTLP endpoint) ───

func TestInitTelemetryStdout(t *testing.T) {
	os.Unsetenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	os.Unsetenv("OTEL_EXPORTER_OTLP_PROTOCOL")
	os.Unsetenv("OTEL_TRACES_SAMPLER")
	os.Unsetenv("OTEL_SERVICE_NAME")

	tp, promExp, err := initTelemetry()
	if err != nil {
		t.Fatalf("initTelemetry() error: %v", err)
	}
	if tp == nil {
		t.Fatal("initTelemetry() returned nil TracerProvider")
	}

	// Cleanup
	if tp != nil {
		shutdownTelemetry(tp)
	}

	_ = promExp
}

// ─── initTelemetry with custom service name ───

func TestInitTelemetryCustomServiceName(t *testing.T) {
	os.Unsetenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	os.Setenv("OTEL_SERVICE_NAME", "test-villum")
	defer os.Unsetenv("OTEL_SERVICE_NAME")

	tp, _, err := initTelemetry()
	if err != nil {
		t.Fatalf("initTelemetry() error: %v", err)
	}
	if tp == nil {
		t.Fatal("initTelemetry() returned nil TracerProvider")
	}
	if tp != nil {
		shutdownTelemetry(tp)
	}
}

// ─── initTelemetry with sampler env vars ───

func TestInitTelemetryWithSampler(t *testing.T) {
	os.Unsetenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	os.Setenv("OTEL_TRACES_SAMPLER", "traceidratio")
	os.Setenv("OTEL_TRACES_SAMPLER_ARG", "0.1")
	defer os.Unsetenv("OTEL_TRACES_SAMPLER")
	defer os.Unsetenv("OTEL_TRACES_SAMPLER_ARG")

	tp, _, err := initTelemetry()
	if err != nil {
		t.Fatalf("initTelemetry() error: %v", err)
	}
	if tp == nil {
		t.Fatal("initTelemetry() returned nil TracerProvider")
	}
	if tp != nil {
		shutdownTelemetry(tp)
	}
}

// ─── OTel Metrics integration ───

func TestOTelMetricsMiddlewareIntegration(t *testing.T) {
	// Use a custom registry to avoid conflicts with other tests
	registry := prometheus.NewRegistry()

	promExporter, err := otelprom.New(otelprom.WithRegisterer(registry))
	if err != nil {
		t.Fatalf("otelprom.New: %v", err)
	}

	res := resource.NewSchemaless(semconv.ServiceNameKey.String("test-villum"))
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(promExporter),
		sdkmetric.WithResource(res),
	)
	defer func() {
		_ = meterProvider.Shutdown(context.Background())
	}()

	// Set global meter provider so initOTelMetrics picks it up
	oldMeterProvider := otel.GetMeterProvider()
	otel.SetMeterProvider(meterProvider)
	defer otel.SetMeterProvider(oldMeterProvider)

	om := initOTelMetrics()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(newOTelMetricsMiddleware(om))
	r.GET("/metrics/prometheus", gin.WrapH(promhttp.HandlerFor(registry, promhttp.HandlerOpts{})))
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})
	r.GET("/test/error", func(c *gin.Context) {
		c.String(500, "fail")
	})

	// Make several requests of different types to generate metrics
	reqs := []struct {
		method string
		path   string
	}{
		{"GET", "/test"},
		{"GET", "/test"},
		{"GET", "/test"},
		{"GET", "/test/error"},
	}
	for _, rr := range reqs {
		req := httptest.NewRequest(rr.method, rr.path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}

	// Fetch Prometheus metrics
	req := httptest.NewRequest("GET", "/metrics/prometheus", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	body := w.Body.String()
	t.Logf("Metrics response:\n%s", body)

	if !strings.Contains(body, "otel_http_requests_total") {
		t.Error("expected otel_http_requests_total in metrics output")
	}
	if !strings.Contains(body, "otel_http_request_duration_seconds") {
		t.Error("expected otel_http_request_duration_seconds in metrics output")
	}
}

// ─── Graceful degradation (no-op when exporter fails) ───

func TestInitTelemetryGracefulDegradation(t *testing.T) {
	// Set an unreachable OTLP endpoint - should degrade gracefully
	os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://127.0.0.1:1")
	os.Unsetenv("OTEL_EXPORTER_OTLP_PROTOCOL")
	defer os.Unsetenv("OTEL_EXPORTER_OTLP_ENDPOINT")

	// This should not crash. gRPC connection to port 1 will fail quickly.
	// initTelemetry should fall back to stdout exporter.
	tp, promExp, err := initTelemetry()

	// Even with OTLP failure, the function should degrade gracefully:
	// - Either returns a valid TracerProvider with stdout fallback
	// - Or logs a warning and returns nil (no-op tracing)
	t.Logf("initTelemetry (bad endpoint): tp=%v, promExp=%v, err=%v", tp != nil, promExp != nil, err)

	if tp != nil {
		shutdownTelemetry(tp)
	}
	_ = promExp
}
