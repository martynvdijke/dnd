package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	_ "modernc.org/sqlite"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"

	"villum/ent"
)

var tracer trace.Tracer

const tracerName = "villum.db"

var dbQueryDuration metric.Float64Histogram

func init() {
	tracer = otel.Tracer(tracerName)
	hist, err := otel.Meter(tracerName).Float64Histogram(
		"db.query_duration_ms",
		metric.WithUnit("ms"),
	)
	if err == nil {
		dbQueryDuration = hist
	}
}

// TraceQuery wraps a database query with an OTel span.
// Use it to instrument ad-hoc SQL queries.
func TraceQuery(ctx context.Context, operation string, fn func(context.Context) error) error {
	return TraceQueryEx(ctx, operation, "", fn)
}

// TraceQueryEx is TraceQuery with optional SQL text for slow-query logging,
// EXPLAIN QUERY PLAN debugging, and histogram instrumentation.
func TraceQueryEx(ctx context.Context, operation, query string, fn func(context.Context) error) error {
	attrs := []attribute.KeyValue{
		attribute.String("db.operation", operation),
		attribute.String("db.system", "sqlite"),
	}
	if query != "" {
		attrs = append(attrs, attribute.String("db.query", truncate(query, 200)))
	}
	_, span := tracer.Start(ctx, operation,
		trace.WithAttributes(attrs...),
	)
	defer span.End()

	start := time.Now()
	err := fn(ctx)
	elapsed := time.Since(start).Seconds()
	ms := elapsed * 1000

	span.SetAttributes(attribute.Float64("db.duration_ms", ms))

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}

	if dbQueryDuration != nil {
		dbQueryDuration.Record(ctx, ms,
			metric.WithAttributes(
				attribute.String("operation", operation),
				attribute.String("system", "sqlite"),
			),
		)
	}

	if ms > slowQueryMS() && query != "" {
		log.Printf("db: slow query %s took %.1fms: %s", operation, ms, truncate(query, 400))
		logExplainPlan(query)
	}

	return err
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func slowQueryMS() float64 {
	if v := os.Getenv("SQLITE_SLOW_QUERY_MS"); v != "" {
		if parsed, err := strconv.ParseFloat(v, 64); err == nil {
			return parsed
		}
	}
	return 100
}

func logExplainPlan(query string) {
	rows, err := DB.Query("EXPLAIN QUERY PLAN " + query)
	if err != nil {
		log.Printf("db: explain query plan failed: %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var id, parent, notUsed int
		var detail string
		if err := rows.Scan(&id, &parent, &notUsed, &detail); err != nil {
			break
		}
		log.Printf("db: explain[%d]: %s", id, detail)
	}
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
	LogPageStats()

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

	// Incremental vacuum weekly (auto_vacuum=INCREMENTAL frees pages from
	// freelist_count without a full VACUUM).
	go func() {
		lastPages := int64(-1)
		for {
			time.Sleep(168 * time.Hour)
			var before, after int64
			DB.QueryRow("PRAGMA page_count").Scan(&before)
			if _, err := DB.Exec("PRAGMA incremental_vacuum"); err != nil {
				log.Printf("db: incremental_vacuum error: %v", err)
				continue
			}
			DB.QueryRow("PRAGMA page_count").Scan(&after)
			if lastPages >= 0 && before > after {
				log.Printf("db: incremental_vacuum freed %d pages (%d -> %d)", before-after, before, after)
			}
			lastPages = after
		}
	}()

	// Log statement-cache stats every 60 minutes
	go func() {
		for {
			time.Sleep(60 * time.Minute)
			log.Printf("db: statement cache entries=%d", stmtCacheLen())
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

	// Create unified-search-index triggers and backfill AFTER ent has settled:
	// ent.Schema.Create recreates managed tables, dropping attached triggers.
	if err := EnsureSearchIndex(); err != nil {
		return fmt.Errorf("ensure search index: %w", err)
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

// ─── Statement cache (modernc.org/sqlite has no built-in statement cache) ───

const stmtCacheMax = 1000

var (
	stmtMu    sync.Mutex
	stmtCache = make(map[string]*sql.Stmt)
	stmtOrder []string
)

// PrepareStmt returns a cached prepared statement for sql, preparing it on
// first use. Entries are evicted FIFO when the cache exceeds stmtCacheMax.
func PrepareStmt(sql string) (*sql.Stmt, error) {
	stmtMu.Lock()
	defer stmtMu.Unlock()
	if st, ok := stmtCache[sql]; ok {
		return st, nil
	}
	st, err := DB.Prepare(sql)
	if err != nil {
		return nil, err
	}
	stmtCache[sql] = st
	stmtOrder = append(stmtOrder, sql)
	for len(stmtOrder) > stmtCacheMax {
		oldest := stmtOrder[0]
		stmtOrder = stmtOrder[1:]
		if old, ok := stmtCache[oldest]; ok {
			old.Close()
			delete(stmtCache, oldest)
		}
	}
	return st, nil
}

func stmtCacheLen() int {
	stmtMu.Lock()
	defer stmtMu.Unlock()
	return len(stmtCache)
}
