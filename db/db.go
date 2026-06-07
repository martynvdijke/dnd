package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	_ "modernc.org/sqlite"
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

	// Read mmap_size from env (default 256MB) before constructing DSN
	mmapSize := int64(268435456)
	if v := os.Getenv("SQLITE_MMAP_SIZE"); v != "" {
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			mmapSize = parsed
		}
	}

	// Build DSN with connection-level PRAGMAs via modernc _pragma query params.
	// Each _pragma value is run as "PRAGMA <value>" on every new connection
	// from the pool, not just the initial one. busy_timeout is sorted first
	// by the driver to ensure the busy handler registers before other ops.
	dsn := dbPath + "?" +
		"_pragma=busy_timeout(10000)" +
		"&_pragma=synchronous(NORMAL)" +
		"&_pragma=cache_size(-64000)" +
		"&_pragma=temp_store(MEMORY)" +
		"&_pragma=mmap_size(" + strconv.FormatInt(mmapSize, 10) + ")" +
		"&_pragma=foreign_keys(ON)"

	var err error
	DB, err = sql.Open("sqlite", dsn)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}

	DB.SetMaxOpenConns(4)

	if err := DB.Ping(); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}

	// Database-level PRAGMAs (stored in the DB file, not per-connection).
	// These only need to run once per DB lifecycle.
	dbPragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA auto_vacuum=INCREMENTAL",
	}
	for _, p := range dbPragmas {
		if _, err := DB.Exec(p); err != nil {
			return fmt.Errorf("execute %s: %w", p, err)
		}
	}

	// Log DB page statistics
	var pageCount, freeListCount, pageSize int
	DB.QueryRow("PRAGMA page_count").Scan(&pageCount)
	DB.QueryRow("PRAGMA freelist_count").Scan(&freeListCount)
	DB.QueryRow("PRAGMA page_size").Scan(&pageSize)
	log.Printf("db: page_count=%d freelist_count=%d page_size=%d mmap_size=%d",
		pageCount, freeListCount, pageSize, mmapSize)

	// Start background PRAGMA optimize every 60 minutes
	go func() {
		for {
			time.Sleep(60 * time.Minute)
			if _, err := DB.Exec("PRAGMA optimize"); err != nil {
				log.Printf("db: PRAGMA optimize error: %v", err)
			} else {
				log.Printf("db: PRAGMA optimize completed")
			}
		}
	}()

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
