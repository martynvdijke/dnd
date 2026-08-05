package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
)

func TestNPCPortraitCRUD(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/npcs", CreateNPC)
		auth.GET("/npcs", ListNPCs)
		auth.GET("/npcs/search", SearchNPCs)
		auth.PUT("/npcs/:id", UpdateNPC)
	})

	// Create NPC with a portrait URL
	createBody, _ := json.Marshal(map[string]any{
		"name":         "Portrait Test NPC",
		"race":         "Human",
		"class":        "Wizard",
		"portrait_url": "/media/npc-portrait-1.png",
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/npcs", bytes.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var created struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil || created.ID == 0 {
		t.Fatalf("create: no id returned: %v body=%s", err, rec.Body.String())
	}

	// List includes portrait_url
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/npcs", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list: expected 200, got %d", rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("npc-portrait-1.png")) {
		t.Fatalf("list: portrait_url missing from response: %s", rec.Body.String())
	}

	// Search includes portrait_url
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/npcs/search?q=Portrait", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("search: expected 200, got %d", rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("npc-portrait-1.png")) {
		t.Fatalf("search: portrait_url missing from response: %s", rec.Body.String())
	}

	// Update changes portrait_url
	updateBody, _ := json.Marshal(map[string]any{
		"name":         "Portrait Test NPC",
		"race":         "Human",
		"class":        "Wizard",
		"portrait_url": "/media/npc-portrait-2.png",
	})
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/api/npcs/"+itoa64(created.ID), bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("update: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// List reflects the new portrait
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/npcs", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list2: expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if strings.Contains(body, "npc-portrait-1.png") {
		t.Fatalf("list2: old portrait still present: %s", body)
	}
	if !strings.Contains(body, "npc-portrait-2.png") {
		t.Fatalf("list2: new portrait missing: %s", body)
	}
}

func TestGetUploadsForEntity(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	SetMediaPath(t.TempDir())
	defer SetMediaPath("/app/media")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/upload", HandleUpload)
		auth.GET("/uploads/entity/:type/:id", GetUploadsForEntity)
		auth.POST("/upload-links", CreateUploadLink)
	})

	// upload an image
	rec := uploadTestImage(t, r, "/api/upload", "campaign", "42", uniqueTestPNG(t))
	if rec.Code != http.StatusCreated && rec.Code != http.StatusOK {
		t.Fatalf("upload: expected 2xx, got %d: %s", rec.Code, rec.Body.String())
	}
	var up struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &up); err != nil || up.ID == 0 {
		t.Fatalf("upload: no id: %v body=%s", err, rec.Body.String())
	}

	// link it to entity (campaign, 7)
	linkBody, _ := json.Marshal(map[string]any{
		"upload_id":   up.ID,
		"entity_type": "campaign",
		"entity_id":   7,
		"field_name":  "gallery",
	})
	rec = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/upload-links", bytes.NewReader(linkBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("link: expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	// entity listing returns the upload with link_id
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/uploads/entity/campaign/7", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("entity list: expected 200, got %d", rec.Code)
	}
	var out struct {
		Uploads []map[string]any `json:"uploads"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("entity list: bad json: %v", err)
	}
	if len(out.Uploads) != 1 {
		t.Fatalf("entity list: expected 1 upload, got %d: %s", len(out.Uploads), rec.Body.String())
	}
	if out.Uploads[0]["link_id"] == nil || out.Uploads[0]["link_id"] == float64(0) {
		t.Fatalf("entity list: link_id missing: %v", out.Uploads[0])
	}

	// different entity id → empty
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/uploads/entity/campaign/999", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("entity list empty: expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"uploads":[]`) {
		t.Fatalf("entity list empty: expected empty array: %s", rec.Body.String())
	}

	// missing params → gin redirects trailing slash (301), not matched
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/uploads/entity/campaign/", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusMovedPermanently {
		t.Fatalf("bad route: expected 301 redirect, got %d", rec.Code)
	}
}
