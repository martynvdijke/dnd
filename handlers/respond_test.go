package handlers

import (
	"bytes"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestBindOr400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("bad json returns 400", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/", bytes.NewBufferString(`{invalid}`))
		c.Request.Header.Set("Content-Type", "application/json")
		var req struct{ Name string `json:"name"` }
		ok := BindOr400(c, &req)
		if ok {
			t.Fatal("expected false on bad json")
		}
		if w.Code != 400 {
			t.Fatalf("expected 400 got %d", w.Code)
		}
	})
	t.Run("valid json returns true", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/", bytes.NewBufferString(`{"name":"ok"}`))
		c.Request.Header.Set("Content-Type", "application/json")
		var req struct{ Name string `json:"name"` }
		ok := BindOr400(c, &req)
		if !ok {
			t.Fatal("expected true on valid json")
		}
		if req.Name != "ok" {
			t.Fatalf("expected name ok got %q", req.Name)
		}
	})
}

func TestWriteError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	WriteError(c, 500, errAccessDenied)
	if w.Code != 500 {
		t.Fatalf("expected 500 got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "access denied") {
		t.Fatalf("body missing error: %s", w.Body.String())
	}
}

func TestWriteNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	WriteNotFound(c, "not found")
	if w.Code != 404 {
		t.Fatalf("expected 404 got %d", w.Code)
	}
}
