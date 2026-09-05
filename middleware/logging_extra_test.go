package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestAppLogger(t *testing.T) {
	buf := NewLogBuffer(100)
	al := NewAppLogger(buf, slog.LevelDebug)
	// suppress stderr
	al.Handler().stderr = &bytes.Buffer{}

	t.Run("Handler/Buffer/Logger not nil", func(t *testing.T) {
		if al.Handler() == nil || al.Buffer() == nil || al.Logger() == nil {
			t.Fatal("nil handler/buffer/logger")
		}
	})
	t.Run("MinLevel SetMinLevel", func(t *testing.T) {
		al.SetMinLevel(slog.LevelWarn)
		if al.MinLevel() != slog.LevelWarn {
			t.Fatalf("expected warn, got %v", al.MinLevel())
		}
		al.SetMinLevel(slog.LevelDebug)
	})
	t.Run("Debug Info Warn Error log to buffer", func(t *testing.T) {
		buf.Clear()
		al.SetMinLevel(slog.LevelDebug)
		al.Debug("test", "debug msg", "k", "v")
		al.Info("test", "info msg")
		al.Warn("test", "warn msg")
		al.Error("test", "error msg")
		if buf.Len() != 4 {
			t.Fatalf("expected 4, got %d", buf.Len())
		}
		// level filtering
		buf.Clear()
		al.SetMinLevel(slog.LevelError)
		al.Debug("test", "d")
		al.Info("test", "i")
		al.Warn("test", "w")
		al.Error("test", "e")
		if buf.Len() != 1 {
			t.Fatalf("expected 1 error entry, got %d", buf.Len())
		}
		al.SetMinLevel(slog.LevelDebug)
	})
	t.Run("Buffer output contains message", func(t *testing.T) {
		buf.Clear()
		var stderr bytes.Buffer
		al.Handler().stderr = &stderr
		al.Info("mysource", "hello world", "foo", "bar")
		entries := buf.Query("", "mysource", time.Time{}, 0)
		if len(entries) != 1 || entries[0].Message != "hello world" {
			t.Fatalf("unexpected entries %v", entries)
		}
		if !strings.Contains(stderr.String(), "hello world") {
			t.Fatalf("stderr missing msg: %q", stderr.String())
		}
	})
}

func TestInitAppLogger(t *testing.T) {
	t.Run("creates singleton", func(t *testing.T) {
		al := InitAppLogger(10, slog.LevelInfo)
		if AppLog == nil || al == nil {
			t.Fatal("expected AppLog not nil")
		}
		if al.Buffer().Cap() != 10 {
			t.Fatalf("expected cap 10, got %d", al.Buffer().Cap())
		}
		// cleanup
		AppLog = nil
	})
	t.Run("FromEnv", func(t *testing.T) {
		t.Setenv("LOG_BUFFER_SIZE", "42")
		t.Setenv("LOG_LEVEL", "debug")
		al := InitAppLoggerFromEnv()
		if al.Buffer().Cap() != 42 {
			t.Fatalf("expected 42, got %d", al.Buffer().Cap())
		}
		if al.MinLevel() != slog.LevelDebug {
			t.Fatalf("expected debug, got %v", al.MinLevel())
		}
		AppLog = nil
	})
	t.Run("FromEnv defaults", func(t *testing.T) {
		t.Setenv("LOG_BUFFER_SIZE", "")
		t.Setenv("LOG_LEVEL", "")
		al := InitAppLoggerFromEnv()
		if al.Buffer().Cap() != 20000 {
			t.Fatalf("expected 20000, got %d", al.Buffer().Cap())
		}
		AppLog = nil
	})
}

func TestNilSafeWrappers(t *testing.T) {
	AppLog = nil
	// should not panic
	LogDebug("s", "m")
	LogInfo("s", "m")
	LogWarn("s", "m")
	LogError("s", "m")
	// with AppLog
	buf := NewLogBuffer(10)
	al := NewAppLogger(buf, slog.LevelDebug)
	al.Handler().stderr = &bytes.Buffer{}
	AppLog = al
	LogDebug("s", "debug")
	LogInfo("s", "info")
	if buf.Len() != 2 {
		t.Fatalf("expected 2, got %d", buf.Len())
	}
	AppLog = nil
}

func TestRequestLogger(t *testing.T) {
	buf := NewLogBuffer(100)
	al := NewAppLogger(buf, slog.LevelDebug)
	al.Handler().stderr = &bytes.Buffer{}
	AppLog = al
	defer func() { AppLog = nil }()

	gin.SetMode(gin.TestMode)

	t.Run("logs 2xx as info", func(t *testing.T) {
		buf.Clear()
		r := gin.New()
		r.Use(RequestLogger())
		r.GET("/ok", func(c *gin.Context) { c.String(200, "ok") })
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/ok", nil)
		r.ServeHTTP(w, req)
		entries := buf.Query("info", "http", time.Time{}, 0)
		if len(entries) != 1 {
			t.Fatalf("expected 1 info http entry, got %d %v", len(entries), buf.Query("", "", time.Time{}, 0))
		}
	})
	t.Run("logs 4xx as warn", func(t *testing.T) {
		buf.Clear()
		r := gin.New()
		r.Use(RequestLogger())
		r.GET("/bad", func(c *gin.Context) { c.String(400, "bad") })
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/bad", nil)
		r.ServeHTTP(w, req)
		if len(buf.Query("warn", "http", time.Time{}, 0)) != 1 {
			t.Fatalf("expected warn, got %v", buf.Query("", "", time.Time{}, 0))
		}
	})
	t.Run("logs 5xx as error", func(t *testing.T) {
		buf.Clear()
		r := gin.New()
		r.Use(RequestLogger())
		r.GET("/err", func(c *gin.Context) { c.String(500, "err") })
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/err", nil)
		r.ServeHTTP(w, req)
		if len(buf.Query("error", "http", time.Time{}, 0)) != 1 {
			t.Fatalf("expected error, got %v", buf.Query("", "", time.Time{}, 0))
		}
	})
	t.Run("skips websocket", func(t *testing.T) {
		buf.Clear()
		r := gin.New()
		r.Use(RequestLogger())
		r.GET("/ws", func(c *gin.Context) { c.String(200, "ok") })
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/ws", nil)
		req.Header.Set("Upgrade", "websocket")
		r.ServeHTTP(w, req)
		if buf.Len() != 0 {
			t.Fatalf("expected 0 for websocket, got %d", buf.Len())
		}
	})
	t.Run("nil AppLog no panic", func(t *testing.T) {
		AppLog = nil
		r := gin.New()
		r.Use(RequestLogger())
		r.GET("/ok", func(c *gin.Context) { c.String(200, "ok") })
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/ok", nil)
		r.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		// restore
		AppLog = al
	})
	_ = http.StatusOK
}
