package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestPublicOTelSettings verifies the mobile-facing endpoint is registered on
// the unauthenticated router and exposes only the public configuration keys.
func TestPublicOTelSettings(t *testing.T) {
	gin.SetMode(gin.TestMode)

	newEngine := func() *gin.Engine {
		r := gin.New()
		RegisterPublicRoutes(r)
		return r
	}

	get := func(t *testing.T, r *gin.Engine) map[string]any {
		t.Helper()
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/otel-settings", nil)
		r.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var m map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &m); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		return m
	}

	t.Run("no env", func(t *testing.T) {
		t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
		m := get(t, newEngine())
		if m["endpoint"] != "" {
			t.Fatalf("expected empty endpoint, got %v", m["endpoint"])
		}
		if m["enabled"] != false || m["traces_enabled"] != false {
			t.Fatalf("expected disabled flags, got %v", m)
		}
	})

	t.Run("with env", func(t *testing.T) {
		t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://otel:4318")
		m := get(t, newEngine())
		if m["endpoint"] != "http://otel:4318" {
			t.Fatalf("expected endpoint, got %v", m["endpoint"])
		}
		if m["enabled"] != true {
			t.Fatalf("expected enabled true, got %v", m["enabled"])
		}
		if m["traces_enabled"] != true {
			t.Fatalf("expected traces_enabled true, got %v", m["traces_enabled"])
		}
		if m["metrics_enabled"] != false {
			t.Fatalf("expected metrics_enabled false, got %v", m["metrics_enabled"])
		}
		if m["logs_enabled"] != false {
			t.Fatalf("expected logs_enabled false, got %v", m["logs_enabled"])
		}
	})

	t.Run("no admin fields leak", func(t *testing.T) {
		t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://otel:4318")
		m := get(t, newEngine())
		want := map[string]bool{
			"endpoint":        true,
			"enabled":         true,
			"traces_enabled":  true,
			"metrics_enabled": true,
			"logs_enabled":    true,
		}
		if len(m) != len(want) {
			t.Fatalf("expected exactly %d keys, got %d: %v", len(want), len(m), m)
		}
		for key := range want {
			if _, ok := m[key]; !ok {
				t.Fatalf("missing key %q in %v", key, m)
			}
		}
		for _, key := range []string{"id", "status", "message"} {
			if _, ok := m[key]; ok {
				t.Fatalf("admin field %q leaked: %v", key, m)
			}
		}
	})
}
