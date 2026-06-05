package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"

	"villum/ent"
)

var tracer trace.Tracer

const tracerName = "villum.db"

func init() {
	tracer = otel.Tracer(tracerName)
}

// TraceQuery wraps a database query with an OTel span.
// Use it to instrument ad-hoc SQL queries.
func TraceQuery(ctx context.Context, operation string, fn func(context.Context) error) error {
	_, span := tracer.Start(ctx, operation,
		trace.WithAttributes(
			attribute.String("db.operation", operation),
			attribute.String("db.system", "sqlite"),
		),
	)
	defer span.End()

	start := time.Now()
	err := fn(ctx)
	elapsed := time.Since(start).Seconds()

	span.SetAttributes(attribute.Float64("db.duration_ms", elapsed*1000))

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}

	return err
}

var DB *sql.DB
var Client *ent.Client

func Init(dbPath string) error {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create db dir: %w", err)
	}

	var err error
	DB, err = sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}

	DB.SetMaxOpenConns(4)

	if err := DB.Ping(); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}

	// Enable WAL and foreign keys
	if _, err := DB.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return fmt.Errorf("enable wal: %w", err)
	}
	if _, err := DB.Exec("PRAGMA foreign_keys=ON"); err != nil {
		return fmt.Errorf("enable fk: %w", err)
	}

	// Initialize ent client
	drv := entsql.OpenDB(dialect.SQLite, DB)
	Client = ent.NewClient(ent.Driver(drv))

	if err := Migrate(); err != nil {
		return err
	}

	// Auto-migrate ent schemas
	if err := Client.Schema.Create(context.Background()); err != nil {
		return fmt.Errorf("ent schema migrate: %w", err)
	}

	// Apply safe ALTER TABLE additions AFTER ent schema migrate.
	// ent.Schema.Create can recreate tables when it detects schema mismatches,
	// which would drop extra columns added by ALTER TABLE. Running ALTER after
	// ensures ent has settled on the table structure first.
	if err := ApplySafeAlters(); err != nil {
		return fmt.Errorf("apply safe alters: %w", err)
	}

	return nil
}

func Close() {
	if Client != nil {
		Client.Close()
	}
	if DB != nil {
		DB.Close()
	}
}
