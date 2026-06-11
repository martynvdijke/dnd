package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
)

func TestListLogs(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/admin/logs", ListLogs)
	})

	t.Run("returns 200 with empty array when AppLog is nil", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/admin/logs")
		testutil.AssertStatus(t, w, 200)
		var entries []any
		testutil.ParseJSON(t, w, &entries)
		if entries == nil {
			t.Fatal("expected empty array, got nil")
		}
		if len(entries) != 0 {
			t.Fatalf("expected 0 entries, got %d", len(entries))
		}
	})

	t.Run("returns 200 with limit param", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/admin/logs?limit=10")
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("returns 200 with source filter", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/admin/logs?source=ai")
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("returns 200 with level filter", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/admin/logs?level=error")
		testutil.AssertStatus(t, w, 200)
	})
}

func TestListLogSources(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/admin/log-sources", ListLogSources)
	})

	t.Run("returns 200 with empty array when AppLog is nil", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/admin/log-sources")
		testutil.AssertStatus(t, w, 200)
		var sources []string
		testutil.ParseJSON(t, w, &sources)
		if sources == nil {
			t.Fatal("expected empty array, got nil")
		}
		if len(sources) != 0 {
			t.Fatalf("expected 0 sources, got %d: %v", len(sources), sources)
		}
	})
}

func TestLogLevel(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	// Ensure log_settings table exists for log-level tests
	InitLogSettings()

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/admin/log-level", GetLogLevel)
		auth.PUT("/admin/log-level", SetLogLevel)
	})

	t.Run("GET returns 200 with level and effective", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/admin/log-level")
		testutil.AssertStatus(t, w, 200)
		var res map[string]string
		testutil.ParseJSON(t, w, &res)
		if res["level"] == "" {
			t.Fatal("expected level field in response")
		}
		if res["effective"] == "" {
			t.Fatal("expected effective field in response")
		}
	})

	t.Run("PUT valid level returns 200", func(t *testing.T) {
		w := testutil.PutJSON(t, r, "/api/admin/log-level", map[string]string{
			"level": "debug",
		})
		testutil.AssertStatus(t, w, 200)
		var res map[string]any
		testutil.ParseJSON(t, w, &res)
		if res["status"] != "ok" {
			t.Fatalf("expected status 'ok', got %v", res["status"])
		}
	})

	t.Run("PUT invalid level returns 400", func(t *testing.T) {
		w := testutil.PutJSON(t, r, "/api/admin/log-level", map[string]string{
			"level": "invalid",
		})
		testutil.AssertStatus(t, w, 400)
	})

	t.Run("PUT empty level returns 400", func(t *testing.T) {
		w := testutil.PutJSON(t, r, "/api/admin/log-level", map[string]string{
			"level": "",
		})
		testutil.AssertStatus(t, w, 400)
	})

	t.Run("PUT and GET reflect changes", func(t *testing.T) {
		// Set to error
		w := testutil.PutJSON(t, r, "/api/admin/log-level", map[string]string{
			"level": "error",
		})
		testutil.AssertStatus(t, w, 200)

		// Read back
		w = testutil.Get(t, r, "/api/admin/log-level")
		testutil.AssertStatus(t, w, 200)
		var res map[string]string
		testutil.ParseJSON(t, w, &res)
		if res["level"] != "error" {
			t.Fatalf("expected level 'error', got %q", res["level"])
		}
	})
}

func TestLogLevelPersistence(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	InitLogSettings()

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/admin/log-level", GetLogLevel)
		auth.PUT("/admin/log-level", SetLogLevel)
	})

	// Set level
	w := testutil.PutJSON(t, r, "/api/admin/log-level", map[string]string{
		"level": "info",
	})
	testutil.AssertStatus(t, w, 200)

	// Verify via separate request
	w = testutil.Get(t, r, "/api/admin/log-level")
	testutil.AssertStatus(t, w, 200)
	var res map[string]string
	testutil.ParseJSON(t, w, &res)
	if res["level"] != "info" {
		t.Fatalf("expected level 'info', got %q", res["level"])
	}
}
