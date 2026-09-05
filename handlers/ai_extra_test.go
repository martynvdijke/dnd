package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/crypto"
	"villum/db"
	"villum/handlers/testutil"
	"villum/models"
)

func setupAITextRouter(t *testing.T) *gin.Engine {
	t.Helper()
	testutil.NewDB(t)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	g := r.Group("/api/ai")
	g.Use(func(c *gin.Context) {
		c.Set("user_id", int64(1))
		c.Set("username", "admin")
		c.Set("role", "admin")
		c.Set("session_id", "test")
		c.Next()
	})
	g.POST("/text", HandleTextGeneration)
	g.POST("/image", HandleImageGeneration)
	g.POST("/save-image", SaveGeneratedImage)
	g.GET("/enabled", GetAIEnabled)
	return r
}

func seedAIEndpoint(t *testing.T, name, typ, baseURL, model string, enabled bool) int64 {
	t.Helper()
	enc, _ := crypto.Encrypt("sk-test-key")
	ep, err := db.CreateAIEndpoint(context.Background(), &models.AIEndpoint{
		Name:            name,
		Type:            typ,
		BaseURL:         baseURL,
		EncryptedAPIKey: enc,
		Model:           model,
		Enabled:         enabled,
	})
	if err != nil {
		t.Fatalf("seed endpoint: %v", err)
	}
	return ep.ID
}

func TestHandleTextGeneration(t *testing.T) {
	// AI disabled branch
	t.Run("forbidden when ai disabled", func(t *testing.T) {
		r := setupAITextRouter(t)
		defer testutil.CloseDB(t)
		// disable ai
		db.DB.Exec("INSERT OR REPLACE INTO app_settings (key, value) VALUES ('ai_enabled','0')")
		w := testutil.PostJSON(t, r, "/api/ai/text", map[string]any{"endpoint_id": 1, "prompt": "hello"})
		if w.Code != 403 {
			t.Fatalf("expected 403, got %d %s", w.Code, w.Body.String())
		}
	})

	// Success with mocked provider
	t.Run("success", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"choices": []map[string]any{
					{"message": map[string]any{"content": "hello world"}, "finish_reason": "stop"},
				},
			})
		}))
		defer srv.Close()
		r := setupAITextRouter(t)
		defer testutil.CloseDB(t)
		id := seedAIEndpoint(t, "t1", "text", srv.URL, "gpt-4o", true)
		w := testutil.PostJSON(t, r, "/api/ai/text", map[string]any{"endpoint_id": id, "prompt": "hi"})
		if w.Code != 200 {
			t.Fatalf("expected 200, got %d %s", w.Code, w.Body.String())
		}
		var resp map[string]any
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp["text"] != "hello world" {
			t.Fatalf("expected hello world, got %v", resp)
		}
	})

	t.Run("provider non-2xx truncated", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(429)
			w.Write([]byte(strings.Repeat("x", 500)))
		}))
		defer srv.Close()
		r := setupAITextRouter(t)
		defer testutil.CloseDB(t)
		id := seedAIEndpoint(t, "t2", "text", srv.URL, "gpt-4o", true)
		w := testutil.PostJSON(t, r, "/api/ai/text", map[string]any{"endpoint_id": id, "prompt": "hi"})
		if w.Code != 502 {
			t.Fatalf("expected 502, got %d %s", w.Code, w.Body.String())
		}
		body := w.Body.String()
		if !strings.Contains(body, "429") {
			t.Fatalf("expected 429 in body, got %s", body)
		}
		// truncated response should contain ...
		if !strings.Contains(body, "...") {
			t.Fatalf("expected truncated body with ..., got %s", body)
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		r := setupAITextRouter(t)
		defer testutil.CloseDB(t)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/ai/text", strings.NewReader(`{bad`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != 400 {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("missing prompt", func(t *testing.T) {
		r := setupAITextRouter(t)
		defer testutil.CloseDB(t)
		seedAIEndpoint(t, "t3", "text", "http://example.com", "gpt-4o", true)
		w := testutil.PostJSON(t, r, "/api/ai/text", map[string]any{"endpoint_id": 1})
		if w.Code != 400 {
			t.Fatalf("expected 400, got %d %s", w.Code, w.Body.String())
		}
	})

	t.Run("endpoint not found", func(t *testing.T) {
		r := setupAITextRouter(t)
		defer testutil.CloseDB(t)
		seedAIEndpoint(t, "t4", "text", "http://example.com", "gpt-4o", true)
		w := testutil.PostJSON(t, r, "/api/ai/text", map[string]any{"endpoint_id": 9999, "prompt": "hi"})
		if w.Code != 404 {
			t.Fatalf("expected 404, got %d %s", w.Code, w.Body.String())
		}
	})
}

func TestHandleImageGeneration(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{"url": "https://cdn.example.com/img.png"}}})
		}))
		defer srv.Close()
		r := setupAITextRouter(t)
		defer testutil.CloseDB(t)
		id := seedAIEndpoint(t, "img1", "image", srv.URL, "dall-e-3", true)
		w := testutil.PostJSON(t, r, "/api/ai/image", map[string]any{"endpoint_id": id, "prompt": "a dragon"})
		if w.Code != 200 {
			t.Fatalf("expected 200, got %d %s", w.Code, w.Body.String())
		}
		var resp map[string]any
		json.Unmarshal(w.Body.Bytes(), &resp)
		imgs, ok := resp["images"].([]any)
		if !ok || len(imgs) != 1 {
			t.Fatalf("expected 1 image, got %v", resp)
		}
	})

	t.Run("non-2xx", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
			w.Write([]byte("internal error " + strings.Repeat("y", 400)))
		}))
		defer srv.Close()
		r := setupAITextRouter(t)
		defer testutil.CloseDB(t)
		id := seedAIEndpoint(t, "img2", "image", srv.URL, "dall-e-3", true)
		w := testutil.PostJSON(t, r, "/api/ai/image", map[string]any{"endpoint_id": id, "prompt": "a dragon"})
		if w.Code != 502 {
			t.Fatalf("expected 502, got %d %s", w.Code, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), "...") {
			t.Fatalf("expected truncated error, got %s", w.Body.String())
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		r := setupAITextRouter(t)
		defer testutil.CloseDB(t)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/ai/image", strings.NewReader(`{`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != 400 {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})
}

func TestSaveGeneratedImageValidation(t *testing.T) {
	r := setupAITextRouter(t)
	defer testutil.CloseDB(t)

	t.Run("missing url", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/ai/save-image", map[string]any{})
		if w.Code != 400 {
			t.Fatalf("expected 400, got %d %s", w.Code, w.Body.String())
		}
	})
	t.Run("invalid scheme", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/ai/save-image", map[string]any{"url": "ftp://example.com/img.png"})
		if w.Code != 400 {
			t.Fatalf("expected 400, got %d %s", w.Code, w.Body.String())
		}
	})
	t.Run("loopback rejected", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/ai/save-image", map[string]any{"url": "http://127.0.0.1/img.png"})
		if w.Code != 400 {
			t.Fatalf("expected 400, got %d %s", w.Code, w.Body.String())
		}
	})
}

func TestGetAIEnabled(t *testing.T) {
	r := setupAITextRouter(t)
	defer testutil.CloseDB(t)
	// default enabled
	w := testutil.Get(t, r, "/api/ai/enabled")
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var m map[string]any
	json.Unmarshal(w.Body.Bytes(), &m)
	if _, ok := m["enabled"]; !ok {
		t.Fatalf("missing enabled field: %s", w.Body.String())
	}
	// disabled
	db.DB.Exec("INSERT OR REPLACE INTO app_settings (key, value) VALUES ('ai_enabled','0')")
	w2 := testutil.Get(t, r, "/api/ai/enabled")
	var m2 map[string]any
	json.Unmarshal(w2.Body.Bytes(), &m2)
	if m2["enabled"] != false {
		t.Fatalf("expected enabled false, got %v", m2)
	}
}
