package middleware

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
)

// LogEntry represents a single structured log entry for the in-memory buffer.
type LogEntry struct {
	ID         int64          `json:"id"`
	Timestamp  time.Time      `json:"timestamp"`
	Level      string         `json:"level"`
	Source     string         `json:"source"`
	Message    string         `json:"message"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

// LogBuffer is a concurrent-safe ring buffer for log entries.
type LogBuffer struct {
	mu      sync.RWMutex
	buffer  []LogEntry
	head    int
	tail    int
	count   int
	cap     int
	nextID  int64
	sources map[string]struct{}
}

// NewLogBuffer creates a ring buffer with the given capacity.
func NewLogBuffer(capacity int) *LogBuffer {
	if capacity <= 0 {
		capacity = 5000
	}
	return &LogBuffer{
		buffer:  make([]LogEntry, capacity),
		cap:     capacity,
		sources: make(map[string]struct{}),
	}
}

// Append adds a log entry to the buffer. Thread-safe.
func (lb *LogBuffer) Append(entry LogEntry) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	entry.ID = lb.nextID
	lb.nextID++
	entry.Timestamp = time.Now()
	if entry.Source != "" {
		lb.sources[entry.Source] = struct{}{}
	}
	lb.buffer[lb.head] = entry
	lb.head = (lb.head + 1) % lb.cap
	if lb.count < lb.cap {
		lb.count++
	} else {
		lb.tail = (lb.tail + 1) % lb.cap
	}
}

// Clear empties the buffer.
func (lb *LogBuffer) Clear() {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.head = 0
	lb.tail = 0
	lb.count = 0
	lb.nextID = 0
	lb.sources = make(map[string]struct{})
}

// Query returns log entries matching filters, newest first.
// level: filter by exact level name (empty = all).
// source: filter by exact source name (empty = all).
// since: only entries after this time (zero = all).
// limit: max entries to return (<= 0 uses capacity).
func (lb *LogBuffer) Query(level, source string, since time.Time, limit int) []LogEntry {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	if limit <= 0 || limit > lb.cap {
		limit = lb.cap
	}

	result := make([]LogEntry, 0, limit)
	for i := 0; i < lb.count && len(result) < limit; i++ {
		idx := (lb.head - 1 - i + lb.cap) % lb.cap
		entry := lb.buffer[idx]

		if level != "" && entry.Level != level {
			continue
		}
		if source != "" && entry.Source != source {
			continue
		}
		if !since.IsZero() && entry.Timestamp.Before(since) {
			continue
		}
		result = append(result, entry)
	}
	return result
}

// Len returns the current number of entries in the buffer.
func (lb *LogBuffer) Len() int {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	return lb.count
}

// Cap returns the configured capacity.
func (lb *LogBuffer) Cap() int { return lb.cap }

// Sources returns the distinct source names observed, sorted alphabetically.
func (lb *LogBuffer) Sources() []string {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	if len(lb.sources) == 0 {
		return nil
	}
	result := make([]string, 0, len(lb.sources))
	for s := range lb.sources {
		result = append(result, s)
	}
	sort.Strings(result)
	return result
}

// ─── Log level helpers ───

// LogLevelFromString converts a level string to slog.Level.
// Returns slog.LevelWarn for unknown values.
func LogLevelFromString(s string) slog.Level {
	switch s {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelWarn
	}
}

// StringFromLogLevel converts slog.Level to a string.
func StringFromLogLevel(l slog.Level) string {
	switch {
	case l < slog.LevelInfo:
		return "debug"
	case l < slog.LevelWarn:
		return "info"
	case l < slog.LevelError:
		return "warn"
	default:
		return "error"
	}
}

// ─── slog.Handler ───

// logHandler is a custom slog.Handler that writes entries to both the LogBuffer
// and stderr. It filters by a configurable minimum level.
// If exportFn is set, each handled record is also sent to it asynchronously
// (used for OTel log export).
type logHandler struct {
	buffer      *LogBuffer
	stderr      io.Writer
	minLevel    slog.Level
	mu          sync.RWMutex
	exportFn    func(ctx context.Context, r slog.Record)
	presetAttrs []slog.Attr
}

func newLogHandler(buffer *LogBuffer, minLevel slog.Level) *logHandler {
	return &logHandler{
		buffer:   buffer,
		stderr:   os.Stderr,
		minLevel: minLevel,
	}
}

// SetExportFn sets an optional export function that is called asynchronously
// for each handled log record. Thread-safe.
func (h *logHandler) SetExportFn(fn func(ctx context.Context, r slog.Record)) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.exportFn = fn
}

// SetMinLevel updates the minimum level for filtering. Thread-safe.
func (h *logHandler) SetMinLevel(level slog.Level) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.minLevel = level
}

func (h *logHandler) getMinLevel() slog.Level {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.minLevel
}

func (h *logHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.getMinLevel()
}

func (h *logHandler) Handle(ctx context.Context, r slog.Record) error {
	minLvl := h.getMinLevel()
	if r.Level < minLvl {
		return nil
	}

	source := ""
	attrs := make(map[string]any)
	// Include any preset attrs from WithAttrs first (so record attrs can override).
	for _, a := range h.presetAttrs {
		if a.Key == "source" {
			source = a.Value.String()
		} else {
			attrs[a.Key] = a.Value.Any()
		}
	}
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "source" {
			source = a.Value.String()
		} else {
			attrs[a.Key] = a.Value.Any()
		}
		return true
	})

	// Attach the OpenTelemetry trace id when the record carries span context.
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		attrs["trace_id"] = sc.TraceID().String()
	}

	level := StringFromLogLevel(r.Level)

	// Write to ring buffer
	h.buffer.Append(LogEntry{
		Timestamp:  r.Time,
		Level:      level,
		Source:     source,
		Message:    r.Message,
		Attributes: attrs,
	})

	// Write to stderr
	msg := fmt.Sprintf("[%s] [%-5s] [%s] %s", r.Time.Format(time.RFC3339), level, source, r.Message)
	if len(attrs) > 0 {
		msg += " " + formatAttrs(attrs)
	}
	msg += "\n"
	_, _ = h.stderr.Write([]byte(msg))

	// Async export (e.g., OTel log export) — read under lock, call outside
	h.mu.RLock()
	fn := h.exportFn
	h.mu.RUnlock()
	if fn != nil {
		// Capture r in closure for async execution
		rec := r.Clone()
		go fn(ctx, rec)
	}

	return nil
}

func (h *logHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	child := &logHandler{
		buffer:   h.buffer,
		stderr:   h.stderr,
		minLevel: h.getMinLevel(),
	}
	h.mu.RLock()
	child.exportFn = h.exportFn
	h.mu.RUnlock()
	// Store attrs for child context — appended to each record via WithAttrs semantics.
	// Since our Handle reads attrs from the record, we wrap the child so that
	// pre-set attrs are included. slog expects WithAttrs to return a handler
	// that automatically includes these attrs in every Handle call.
	if len(attrs) > 0 {
		child.presetAttrs = append(child.presetAttrs, attrs...)
	}
	return child
}

func (h *logHandler) WithGroup(_ string) slog.Handler { return h }

// formatAttrs renders attributes as key=value pairs for stderr output.
func formatAttrs(attrs map[string]any) string {
	if len(attrs) == 0 {
		return ""
	}
	keys := make([]string, 0, len(attrs))
	for k := range attrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(fmt.Sprintf("%v", attrs[k]))
	}
	return b.String()
}

// ─── AppLogger ───

// AppLogger provides convenience methods for structured logging with source attribution.
type AppLogger struct {
	logger  *slog.Logger
	handler *logHandler
	buffer  *LogBuffer
}

// NewAppLogger creates a new AppLogger with the given buffer and minimum level.
func NewAppLogger(buffer *LogBuffer, minLevel slog.Level) *AppLogger {
	handler := newLogHandler(buffer, minLevel)
	return &AppLogger{
		logger:  slog.New(handler),
		handler: handler,
		buffer:  buffer,
	}
}

// Handler returns the underlying logHandler (for dynamic level changes).
func (l *AppLogger) Handler() *logHandler { return l.handler }

// Buffer returns the underlying LogBuffer.
func (l *AppLogger) Buffer() *LogBuffer { return l.buffer }

// Logger returns the underlying slog.Logger.
func (l *AppLogger) Logger() *slog.Logger { return l.logger }

// SetMinLevel updates the minimum log level dynamically.
func (l *AppLogger) SetMinLevel(level slog.Level) { l.handler.SetMinLevel(level) }

// MinLevel returns the current minimum log level.
func (l *AppLogger) MinLevel() slog.Level { return l.handler.getMinLevel() }

// Debug emits a debug-level log entry.
func (l *AppLogger) Debug(source, msg string, args ...any) {
	l.logger.Debug(msg, append([]any{"source", source}, args...)...)
}

// Info emits an info-level log entry.
func (l *AppLogger) Info(source, msg string, args ...any) {
	l.logger.Info(msg, append([]any{"source", source}, args...)...)
}

// Warn emits a warning-level log entry.
func (l *AppLogger) Warn(source, msg string, args ...any) {
	l.logger.Warn(msg, append([]any{"source", source}, args...)...)
}

// Error emits an error-level log entry.
func (l *AppLogger) Error(source, msg string, args ...any) {
	l.logger.Error(msg, append([]any{"source", source}, args...)...)
}

// ─── Package-level singleton ───

// AppLog is the package-level application logger. Initialized by InitAppLogger.
var AppLog *AppLogger

// InitAppLogger initializes the package-level AppLog singleton with the given
// capacity (for the ring buffer) and minimum log level.
// If cap <= 0, defaults to 5000.
func InitAppLogger(capacity int, minLevel slog.Level) *AppLogger {
	buffer := NewLogBuffer(capacity)
	AppLog = NewAppLogger(buffer, minLevel)
	return AppLog
}

// InitAppLoggerFromEnv initializes the AppLog using environment variables:
//   - LOG_BUFFER_SIZE: ring buffer capacity (default 20000)
//   - LOG_LEVEL: minimum log level (default "warn")
func InitAppLoggerFromEnv() *AppLogger {
	cap := 20000
	if v := os.Getenv("LOG_BUFFER_SIZE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cap = n
		}
	}
	minLvl := slog.LevelWarn
	if v := strings.ToLower(strings.TrimSpace(os.Getenv("LOG_LEVEL"))); v != "" {
		minLvl = LogLevelFromString(v)
	}
	return InitAppLogger(cap, minLvl)
}

// ─── Nil-safe convenience wrappers ───
// These functions safely no-op when AppLog is nil (e.g., in test contexts).

// LogDebug emits a debug-level log entry via AppLog, no-op if AppLog is nil.
func LogDebug(source, msg string, args ...any) {
	if AppLog != nil {
		AppLog.Debug(source, msg, args...)
	}
}

// LogInfo emits an info-level log entry via AppLog, no-op if AppLog is nil.
func LogInfo(source, msg string, args ...any) {
	if AppLog != nil {
		AppLog.Info(source, msg, args...)
	}
}

// LogWarn emits a warning-level log entry via AppLog, no-op if AppLog is nil.
func LogWarn(source, msg string, args ...any) {
	if AppLog != nil {
		AppLog.Warn(source, msg, args...)
	}
}

// LogError emits an error-level log entry via AppLog, no-op if AppLog is nil.
func LogError(source, msg string, args ...any) {
	if AppLog != nil {
		AppLog.Error(source, msg, args...)
	}
}

// RequestLogger returns a Gin middleware that logs HTTP requests to AppLog.
// Uses the package-level AppLog singleton; no-ops silently if nil.
// Logs at info for 2xx, warn for 4xx, error for 5xx.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip logging for WebSocket upgrades
		if c.Request.Header.Get("Upgrade") == "websocket" {
			c.Next()
			return
		}

		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		if AppLog == nil {
			return
		}

		msg := fmt.Sprintf("%s %s %d (%s)", method, path, status, latency)
		attrs := []any{
			"method", method,
			"path", path,
			"status", status,
			"latency", latency.String(),
			"ip", clientIP,
			"user_agent", userAgent,
		}

		switch {
		case status >= http.StatusInternalServerError:
			AppLog.Error("http", msg, attrs...)
		case status >= http.StatusBadRequest:
			AppLog.Warn("http", msg, attrs...)
		default:
			AppLog.Info("http", msg, attrs...)
		}
	}
}
