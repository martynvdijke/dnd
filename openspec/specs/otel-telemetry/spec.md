## Purpose

Provide OpenTelemetry-based distributed tracing, metrics, and logs for the villum server, with configurable OTLP export (gRPC + HTTP), Gin request instrumentation, and trace-to-log correlation via a slog bridge. All configuration follows standard OpenTelemetry environment variables, and the server SHALL degrade gracefully when no OTLP endpoint is configured.

## Requirements

### Requirement: OTel trace export via OTLP

The system SHALL export OpenTelemetry traces to a configurable OTLP endpoint (gRPC or HTTP). The exporter SHALL support standard OTel environment variables for configuration: `OTEL_EXPORTER_OTLP_ENDPOINT`, `OTEL_EXPORTER_OTLP_PROTOCOL` (grpc/http/protobuf), `OTEL_EXPORTER_OTLP_HEADERS`, `OTEL_EXPORTER_OTLP_TIMEOUT`. If no OTLP endpoint is configured, the system SHALL fall back to the stdout exporter.

#### Scenario: Trace export via OTLP gRPC

- **WHEN** `OTEL_EXPORTER_OTLP_ENDPOINT` is set to a valid collector address and `OTEL_EXPORTER_OTLP_PROTOCOL` is unset (defaults to grpc)
- **THEN** the system SHALL initialize a gRPC OTLP trace exporter connected to that endpoint

#### Scenario: Trace export via OTLP HTTP

- **WHEN** `OTEL_EXPORTER_OTLP_ENDPOINT` is set and `OTEL_EXPORTER_OTLP_PROTOCOL` is `http/protobuf`
- **THEN** the system SHALL initialize an HTTP OTLP trace exporter connected to that endpoint

#### Scenario: Fallback to stdout exporter

- **WHEN** no `OTEL_EXPORTER_OTLP_ENDPOINT` is set
- **THEN** the system SHALL use the stdout trace exporter (existing behavior)

#### Scenario: Graceful shutdown

- **WHEN** the server shuts down
- **THEN** the tracer provider SHALL flush all pending spans and shut down within a configurable timeout

### Requirement: Gin request tracing

The system SHALL create an OpenTelemetry span for every incoming HTTP request via the `otelgin` middleware. Each span SHALL include attributes: `http.method`, `http.url`, `http.route`, `http.status_code`, and the OTel standard `http.*` semantic conventions.

#### Scenario: Automatic span creation

- **WHEN** any HTTP request is received by the Gin router
- **THEN** a root span SHALL be created for that request with method, route, and status code attributes

#### Scenario: Span status on error

- **WHEN** the handler returns a 5xx status code
- **THEN** the span status SHALL be set to `Error`

#### Scenario: Trace context propagation

- **WHEN** an incoming request includes `traceparent` or `tracestate` headers
- **THEN** the span SHALL be part of the incoming trace context (distributed tracing)

### Requirement: OTel metrics alongside Prometheus

The system SHALL expose OTel-native metrics (HTTP request count and duration) that can be scraped via the OTel Prometheus exporter. The existing Prometheus client_golang metrics SHALL remain unchanged.

#### Scenario: OTel HTTP metrics exported

- **WHEN** the `/metrics/prometheus` endpoint is scraped
- **THEN** the response SHALL include `otel_http_requests_total` and `otel_http_request_duration_seconds` metrics alongside the existing Prometheus metrics

#### Scenario: Metrics attribute consistency

- **WHEN** OTel metrics are recorded
- **THEN** they SHALL use the same label dimensions as the existing Prometheus metrics (method, path, status)

### Requirement: Configurable sampling and resources

The system SHALL support OTel standard environment variables for trace sampling and resource attributes: `OTEL_TRACES_SAMPLER` (always_on, always_off, traceidratio, parentbased_traceidratio), `OTEL_TRACES_SAMPLER_ARG` (ratio value), `OTEL_RESOURCE_ATTRIBUTES` (comma-separated key=value pairs), `OTEL_SERVICE_NAME`.

#### Scenario: Sampling configured via env vars

- **WHEN** `OTEL_TRACES_SAMPLER` is set to `traceidratio` and `OTEL_TRACES_SAMPLER_ARG` is set to `0.1`
- **THEN** the tracer provider SHALL sample approximately 10% of traces

#### Scenario: Resource attributes from environment

- **WHEN** `OTEL_RESOURCE_ATTRIBUTES` is set to `deployment.environment=production,service.version=2.9.4`
- **THEN** the OTel resource SHALL include those attributes on every span and metric

#### Scenario: Default service name

- **WHEN** `OTEL_SERVICE_NAME` is not set
- **THEN** the service name SHALL default to `villum`

### Requirement: Graceful degradation

The system SHALL start successfully even if OTel initialization partially fails. If the OTLP exporter cannot connect, the system SHALL log a warning and continue with a no-op tracer provider. A failed OTel init SHALL NOT block server startup.

#### Scenario: OTLP connection failure at startup

- **WHEN** `OTEL_EXPORTER_OTLP_ENDPOINT` points to an unreachable collector
- **THEN** the system SHALL log a warning and start with the stdout exporter as fallback, and SHALL NOT crash

#### Scenario: No OTel configuration

- **WHEN** no OTel environment variables are set
- **THEN** the system SHALL start and run with no-op tracing (no spans exported, no errors)
