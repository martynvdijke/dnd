package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
)

// 9.1-9.3: JSON response shape guarantees for major endpoints.
func TestSchemaJSONResponseShapes(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 100, 1, "Shape", "Human", "Wizard")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/characters", ListCharacters)
		auth.GET("/npcs", ListNPCs)
		auth.GET("/campaigns", ListCampaigns)
		auth.POST("/characters", CreateCharacter)
	})

	// 9.1 characters list: array of objects with expected fields
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/characters", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("characters list: expected 200, got %d", rec.Code)
	}
	var chars []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &chars); err != nil {
		t.Fatalf("characters list: not a JSON array: %v", err)
	}
	if len(chars) == 0 {
		t.Fatalf("characters list: expected at least one character")
	}
	for _, k := range []string{"id", "name", "class", "level"} {
		if _, ok := chars[0][k]; !ok {
			t.Fatalf("characters list: missing field %q in %v", k, chars[0])
		}
	}

	// 9.2 list endpoints return arrays, not null
	for _, path := range []string{"/api/npcs", "/api/campaigns"} {
		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, path, nil)
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: expected 200, got %d", path, rec.Code)
		}
		var arr []map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &arr); err != nil {
			t.Fatalf("%s: expected JSON array, got: %s", path, rec.Body.String())
		}
		if arr == nil {
			t.Fatalf("%s: response was null, expected []", path)
		}
	}

	// 9.3 validation error carries a descriptive message field
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/characters", bytes.NewReader([]byte(`{"race":"Human","class":"Wizard"}`)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("create without name: expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	var errBody map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &errBody); err != nil {
		t.Fatalf("create without name: error body not JSON: %s", rec.Body.String())
	}
	msg, hasMsg := errBody["error"]
	if !hasMsg {
		msg, hasMsg = errBody["message"]
	}
	if !hasMsg || msg == "" {
		t.Fatalf("create without name: no descriptive error/message field: %v", errBody)
	}
}
