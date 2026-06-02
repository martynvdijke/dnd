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
		var eps []interface{}
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
		var ep map[string]interface{}
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
		var eps []interface{}
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
		var ep map[string]interface{}
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
		var ep map[string]interface{}
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
		var ep map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &ep)
		tags, ok := ep["tags"].([]interface{})
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
		var result map[string]interface{}
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
