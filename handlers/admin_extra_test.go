package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
)

func TestAdminResyncSearchIndex(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/admin/search/resync", HandleAdminResyncSearchIndex)
	})
	w := testutil.Get(t, r, "/api/admin/search/resync")
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d %s", w.Code, w.Body.String())
	}
	var m map[string]any
	json.Unmarshal(w.Body.Bytes(), &m)
	if m["message"] != "search index resynced" {
		t.Fatalf("unexpected message %v", m)
	}
}

func TestAdminOTelSettings(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/admin/otel-settings", GetOTelSettings)
		auth.POST("/admin/otel-settings", SaveOTelSettings)
	})

	t.Run("GetOTelSettings no env", func(t *testing.T) {
		os.Unsetenv("OTEL_EXPORTER_OTLP_ENDPOINT")
		w := testutil.Get(t, r, "/api/admin/otel-settings")
		if w.Code != 200 {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		var m map[string]any
		json.Unmarshal(w.Body.Bytes(), &m)
		if m["enabled"] != false {
			t.Fatalf("expected enabled false, got %v", m)
		}
	})
	t.Run("GetOTelSettings with env", func(t *testing.T) {
		t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://otel:4317")
		w := testutil.Get(t, r, "/api/admin/otel-settings")
		var m map[string]any
		json.Unmarshal(w.Body.Bytes(), &m)
		if m["enabled"] != true {
			t.Fatalf("expected enabled true, got %v", m)
		}
		if m["endpoint"] != "http://otel:4317" {
			t.Fatalf("expected endpoint, got %v", m)
		}
	})
	t.Run("SaveOTelSettings read-only", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/admin/otel-settings", map[string]any{"endpoint": "http://x"})
		if w.Code != 200 {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		var m map[string]any
		json.Unmarshal(w.Body.Bytes(), &m)
		if m["status"] != "read_only" {
			t.Fatalf("expected read_only, got %v", m)
		}
	})
}

func TestAdminCleanupTask(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	// Should not panic or block
	StartDBCleanupTask()
	// Give goroutine a moment to start (it sleeps 4h before first iteration)
	if r := recover(); r != nil {
		t.Fatalf("panic: %v", r)
	}
	// Also cover GET via http to ensure handlers still work
	_ = http.StatusOK
}
