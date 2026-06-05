package middleware

import (
	"context"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "villum.db"

// TraceDBQuery wraps a database operation with an OpenTelemetry span.
// It measures duration, records success/failure, and links to the parent trace.
// Usage:
//
//	err := middleware.TraceDBQuery(ctx, "SELECT character", func(ctx context.Context) error {
//	    return db.DB.QueryRow(...).Scan(...)
//	})
func TraceDBQuery(ctx context.Context, operation string, fn func(context.Context) error) error {
	tracer := otel.Tracer(tracerName)
	ctx, span := tracer.Start(ctx, operation,
		trace.WithAttributes(
			attribute.String("db.operation", operation),
			attribute.String("db.system", "sqlite"),
		),
	)
	defer span.End()

	start := time.Now()
	err := fn(ctx)
	duration := time.Since(start).Seconds()

	span.SetAttributes(attribute.Float64("db.duration_ms", duration*1000))

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		log.Printf("DB query error [%s]: %v (duration: %.2fms)", operation, err, duration*1000)
	} else {
		span.SetStatus(codes.Ok, "")
	}

	return err
}
