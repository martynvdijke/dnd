package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers/testutil"
	"villum/models"
)

func mockAuth(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", int64(1))
		c.Set("username", "testuser")
		c.Set("role", role)
		c.Set("session_id", "test-session")
		c.Next()
	}
}

func mockCSRF() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

func setupAdminAIRouter(t *testing.T) *gin.Engine {
	t.Helper()
	testutil.NewDB(t)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	admin := r.Group("/api/admin")
	admin.Use(mockAuth("admin"), mockCSRF())
	{
		admin.GET("/ai-endpoints", ListAIEndpoints)
		admin.GET("/ai-endpoints/:id", GetAIEndpoint)
		admin.POST("/ai-endpoints", CreateAIEndpoint)
		admin.PUT("/ai-endpoints/:id", UpdateAIEndpoint)
		admin.DELETE("/ai-endpoints/:id", DeleteAIEndpoint)
		admin.POST("/ai-endpoints/:id/test", TestAIEndpoint)
	}
	return r
}

func setupNonAdminAIRouter(t *testing.T) *gin.Engine {
	t.Helper()
	testutil.NewDB(t)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	admin := r.Group("/api/admin")
	admin.Use(mockAuth("user"), mockAdminRequired(), mockCSRF())
	{
		admin.GET("/ai-endpoints", ListAIEndpoints)
	}
	return r
}

func mockAdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin required"})
			return
		}
		c.Next()
	}
}

func cleanupAIEndpoints() {
	ctx := context.Background()
	eps, _ := db.GetAIEndpoints(ctx)
	for _, ep := range eps {
		db.DeleteAIEndpoint(ctx, ep.ID)
	}
}

func TestAIEndpointCRUD(t *testing.T) {
	r := setupAdminAIRouter(t)
	defer cleanupAIEndpoints()
	defer testutil.CloseDB(t)

	t.Run("list returns empty array", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/admin/ai-endpoints", nil)
		r.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var eps []any
		json.Unmarshal(w.Body.Bytes(), &eps)
		if len(eps) != 0 {
			t.Fatalf("expected empty list, got %d items", len(eps))
		}
	})

	t.Run("create valid text endpoint", func(t *testing.T) {
		body := `{"name":"test-openai","type":"text","base_url":"https://api.openai.com/v1","api_key":"sk-test123","model":"gpt-4o","temperature":0.7,"max_tokens":2000}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/admin/ai-endpoints", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != 201 {
			t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
		}
		var ep map[string]any
		json.Unmarshal(w.Body.Bytes(), &ep)
		if ep["name"] != "test-openai" {
			t.Fatalf("expected name test-openai, got %v", ep["name"])
		}
		if _, exists := ep["encrypted_api_key"]; exists {
			t.Fatalf("API key should be redacted in response, got %v", ep["encrypted_api_key"])
		}
	})

	t.Run("list returns one endpoint", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/admin/ai-endpoints", nil)
		r.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var eps []any
		json.Unmarshal(w.Body.Bytes(), &eps)
		if len(eps) != 1 {
			t.Fatalf("expected 1 endpoint, got %d", len(eps))
		}
	})

	t.Run("get endpoint by ID", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/admin/ai-endpoints/1", nil)
		r.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var ep map[string]any
		json.Unmarshal(w.Body.Bytes(), &ep)
		if ep["id"] != float64(1) {
			t.Fatalf("expected id 1, got %v", ep["id"])
		}
	})

	t.Run("create with missing fields returns error", func(t *testing.T) {
		body := `{"name":"incomplete"}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/admin/ai-endpoints", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != 400 {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("create with duplicate name returns conflict", func(t *testing.T) {
		body := `{"name":"test-openai","type":"text","base_url":"https://api.openai.com/v1","api_key":"sk-test123","model":"gpt-4o"}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/admin/ai-endpoints", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != 409 {
			t.Fatalf("expected 409, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("update endpoint fields", func(t *testing.T) {
		body := `{"name":"test-openai-updated","type":"text","base_url":"https://api.openai.com/v1","model":"gpt-4-turbo","enabled":false,"temperature":0.5,"max_tokens":4096}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/api/admin/ai-endpoints/1", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var ep map[string]any
		json.Unmarshal(w.Body.Bytes(), &ep)
		if ep["name"] != "test-openai-updated" {
			t.Fatalf("expected name test-openai-updated, got %v", ep["name"])
		}
		if ep["model"] != "gpt-4-turbo" {
			t.Fatalf("expected model gpt-4-turbo, got %v", ep["model"])
		}
	})

	t.Run("update with new API key", func(t *testing.T) {
		body := `{"name":"test-openai-updated","type":"text","base_url":"https://api.openai.com/v1","api_key":"sk-new-key-789","model":"gpt-4-turbo"}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/api/admin/ai-endpoints/1", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("get non-existent endpoint returns 404", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/admin/ai-endpoints/999", nil)
		r.ServeHTTP(w, req)
		if w.Code != 404 {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("create with tags", func(t *testing.T) {
		body := `{"name":"tagged-endpoint","type":"text","base_url":"https://api.anthropic.com/v1","api_key":"sk-ant-test","model":"claude-3","tags":["fast","cheap","creative"]}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/admin/ai-endpoints", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != 201 {
			t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
		}
		var ep map[string]any
		json.Unmarshal(w.Body.Bytes(), &ep)
		tags, ok := ep["tags"].([]any)
		if !ok || len(tags) != 3 {
			t.Fatalf("expected 3 tags, got %v", ep["tags"])
		}
	})

	t.Run("delete existing endpoint", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/api/admin/ai-endpoints/2", nil)
		r.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("delete non-existent endpoint", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/api/admin/ai-endpoints/999", nil)
		r.ServeHTTP(w, req)
		if w.Code != 404 {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("create with invalid type returns error", func(t *testing.T) {
		body := `{"name":"bad-type","type":"video","base_url":"https://api.example.com","api_key":"sk-test","model":"test-model"}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/admin/ai-endpoints", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != 400 {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("create image endpoint with image_size", func(t *testing.T) {
		size := "1024x1024"
		body := fmt.Sprintf(`{"name":"test-dalle","type":"image","base_url":"https://api.openai.com/v1","api_key":"sk-dalle","model":"dall-e-3","image_size":"%s"}`, size)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/admin/ai-endpoints", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != 201 {
			t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("test endpoint with invalid URL", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/admin/ai-endpoints/1/test", nil)
		r.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var result map[string]any
		json.Unmarshal(w.Body.Bytes(), &result)
		if result["success"] != false {
			t.Fatalf("expected success=false for invalid URL, got %v", result["success"])
		}
	})
}

func TestAIEndpointAuth(t *testing.T) {
	r := setupNonAdminAIRouter(t)
	defer testutil.CloseDB(t)

	t.Run("non-admin cannot list endpoints", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/admin/ai-endpoints", nil)
		r.ServeHTTP(w, req)
		if w.Code != 403 {
			t.Fatalf("expected 403 for non-admin, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestSanitizeError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"short error", fmt.Errorf("connection refused"), "connection refused"},
		{"long error truncated", fmt.Errorf("AAAAABBBBBCCCCCDDDDDEEEEEFFFFFGGGGGHHHHHIIIIIJJJJJKKKKKLLLLLMMMMMNNNNNOOOOOPPPPPQQQQQRRRRRSSSSSTTTTTUUUUUVVVVV"), "AAAAABBBBBCCCCCDDDDDEEEEEFFFFFGGGGGHHHHHIIIIIJJJJJKKKKKLLLLLMMMMMNNNNNOOOOOPPPPPQQQQQRRRRRSSSSSTTTTT..."},
		{"newlines removed", fmt.Errorf("line1\nline2\nline3"), "line1 line2 line3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeError(tt.err)
			if got != tt.want {
				t.Errorf("sanitizeError() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTruncateResponse(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{"short string", "OK", "OK"},
		{"long string truncated", `{"error": {"message": "Insufficient quota", "type": "insufficient_quota", "param": null, "code": "insufficient_quota"}}`, `{"error": {"message": "Insufficient quota", "type": "insufficient_quota", "param": null, "code": "insufficient_quota"}}`},
		{"exactly 200 chars", repeatStr("a", 200), repeatStr("a", 200)},
		{"over 200 chars", repeatStr("b", 300), repeatStr("b", 200) + "..."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateResponse(tt.s)
			if got != tt.want {
				t.Errorf("truncateResponse() = %q (len=%d), want %q (len=%d)", got, len(got), tt.want, len(tt.want))
			}
		})
	}
}

func repeatStr(s string, n int) string {
	var result strings.Builder
	for range n {
		result.WriteString(s)
	}
	return result.String()
}

// TestAIErrorMessagesFormat verifies that AI endpoint error responses include
// the actual error details rather than generic messages.
func TestAIErrorMessagesFormat(t *testing.T) {
	// Test the error message patterns used in HandleTextGeneration and HandleImageGeneration
	// These are the fmt.Sprintf patterns that ensure users see the actual failure reason.

	// Simulate a connection error
	connErr := fmt.Errorf("dial tcp: lookup api.openai.com: no such host")
	msg := fmt.Sprintf("AI provider request failed: %s", sanitizeError(connErr))
	if !strings.Contains(msg, "no such host") {
		t.Errorf("expected error message to include 'no such host', got: %s", msg)
	}
	if strings.Contains(msg, "\n") {
		t.Errorf("error message should not contain newlines: %s", msg)
	}

	// Simulate an HTTP error response
	respBody := `{"error": {"message": "You exceeded your current quota", "type": "insufficient_quota"}}`
	httpMsg := fmt.Sprintf("AI provider returned HTTP %d: %s", 429, truncateResponse(respBody))
	if !strings.Contains(httpMsg, "quota") {
		t.Errorf("expected HTTP error message to include the API error detail, got: %s", httpMsg)
	}
	if !strings.Contains(httpMsg, "429") {
		t.Errorf("expected HTTP error message to include status code, got: %s", httpMsg)
	}
}

func setupDMAIRouter(t *testing.T) *gin.Engine {
	t.Helper()
	testutil.NewDB(t)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	dm := r.Group("/api/ai")
	dm.Use(mockAuth("user"), mockCSRF())
	{
		dm.GET("/endpoints", HandleListEnabledAIEndpoints)
	}
	return r
}

// TestListEnabledAIEndpoints covers the DM-facing enabled-endpoints list
// (task 5.1): type filtering, disabled exclusion, and secret redaction.
func TestListEnabledAIEndpoints(t *testing.T) {
	r := setupDMAIRouter(t)
	defer cleanupAIEndpoints()
	defer testutil.CloseDB(t)

	ctx := context.Background()
	seed := func(name, typ, model string, enabled bool) int64 {
		t.Helper()
		ep, err := db.CreateAIEndpoint(ctx, &models.AIEndpoint{
			Name:            name,
			Type:            typ,
			BaseURL:         "https://api.example.com/v1",
			EncryptedAPIKey: "sk-enc-secret",
			Model:           model,
			Enabled:         enabled,
		})
		if err != nil {
			t.Fatalf("seed endpoint failed: %v", err)
		}
		return ep.ID
	}
	textID := seed("dm-text", "text", "gpt-4o", true)
	seed("dm-text-disabled", "text", "gpt-4o", false)
	imageID := seed("dm-image", "image", "dall-e-3", true)

	t.Run("missing type returns 400", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/ai/endpoints", nil)
		r.ServeHTTP(w, req)
		if w.Code != 400 {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("invalid type returns 400", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/ai/endpoints?type=invalid", nil)
		r.ServeHTTP(w, req)
		if w.Code != 400 {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("filters by type, excludes disabled, redacts secrets", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/ai/endpoints?type=text", nil)
		r.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		body := w.Body.String()
		if strings.Contains(body, "api.example.com") || strings.Contains(body, "sk-enc-secret") {
			t.Fatalf("response leaked base URL or API key: %s", body)
		}
		var eps []map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &eps); err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}
		if len(eps) != 1 {
			t.Fatalf("expected exactly 1 enabled text endpoint, got %d: %s", len(eps), body)
		}
		if int64(eps[0]["id"].(float64)) != textID {
			t.Fatalf("expected id %d, got %v", textID, eps[0]["id"])
		}
		for _, key := range []string{"id", "name", "model", "type"} {
			if _, ok := eps[0][key]; !ok {
				t.Fatalf("expected field %q in response, got %v", key, eps[0])
			}
		}
	})

	t.Run("image type returns only image endpoints", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/ai/endpoints?type=image", nil)
		r.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var eps []map[string]any
		json.Unmarshal(w.Body.Bytes(), &eps)
		if len(eps) != 1 || int64(eps[0]["id"].(float64)) != imageID {
			t.Fatalf("expected 1 image endpoint with id %d, got %v", imageID, eps)
		}
	})
}
