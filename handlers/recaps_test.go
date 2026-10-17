package handlers

import (
	"bytes"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers/testutil"
)

func TestRecapCRUD(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "RecapHero", "Human", "Fighter")
	testutil.SeedCampaign(t, 1, "Recap Campaign", "Testers", 1)

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/campaigns/:id/recaps", ListCampaignRecaps)
		auth.POST("/campaigns/:id/recaps", CreateCampaignRecap)
		auth.GET("/recaps/:id", GetCampaignRecap)
		auth.PUT("/recaps/:id", UpdateCampaignRecap)
		auth.DELETE("/recaps/:id", DeleteCampaignRecap)
		auth.POST("/campaigns/:id/recaps/generate", GenerateCampaignRecap)
		auth.PUT("/recaps/:id/send", MarkRecapAsSent)
	})

	var recapID int64

	t.Run("list recaps returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/campaigns/1/recaps")
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("create recap returns 201", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/campaigns/1/recaps", map[string]any{
			"title":              "Session 1 Recap",
			"content":            "The party explored the dungeon...",
			"session_start_date": "2026-06-01",
			"session_end_date":   "2026-06-01",
		})
		testutil.AssertStatus(t, w, 201)
		var result map[string]any
		testutil.ParseJSON(t, w, &result)
		recapID = int64(result["id"].(float64))
	})

	t.Run("get recap returns 200", func(t *testing.T) {
		if recapID == 0 {
			t.Skip("no recap created")
		}
		w := testutil.Get(t, r, "/api/recaps/"+strconv.FormatInt(recapID, 10))
		testutil.AssertStatus(t, w, 200)
		var recap map[string]any
		testutil.ParseJSON(t, w, &recap)
		if recap["title"] != "Session 1 Recap" {
			t.Fatalf("expected 'Session 1 Recap', got %v", recap["title"])
		}
	})

	t.Run("update recap returns 200", func(t *testing.T) {
		if recapID == 0 {
			t.Skip("no recap created")
		}
		w := testutil.PutJSON(t, r, "/api/recaps/"+strconv.FormatInt(recapID, 10), map[string]any{
			"title":   "Updated Recap",
			"content": "Updated content",
		})
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("generate recap returns 200 or 500", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/campaigns/1/recaps/generate", nil)
		if w.Code != 200 && w.Code != 500 {
			t.Fatalf("expected 200 or 500, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("mark recap as sent returns 200", func(t *testing.T) {
		if recapID == 0 {
			t.Skip("no recap created")
		}
		w := testutil.PutJSON(t, r, "/api/recaps/"+strconv.FormatInt(recapID, 10)+"/send", nil)
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("delete recap returns 200", func(t *testing.T) {
		if recapID == 0 {
			t.Skip("no recap created")
		}
		w := testutil.Delete(t, r, "/api/recaps/"+strconv.FormatInt(recapID, 10))
		testutil.AssertStatus(t, w, 200)
	})
}

func seedRecapUpload(t *testing.T, id int64, ownerType string, ownerID int64) {
	t.Helper()
	hash := strings.Repeat("a", 64)
	_, err := db.DB.Exec(`INSERT OR IGNORE INTO uploads(id, hash, ext, url, owner_type, owner_id, kind) VALUES(?,?,?,?,?,?,?)`, id, hash+strconv.FormatInt(id, 10), ".jpg", "/media/a.jpg", ownerType, ownerID, "image")
	if err != nil {
		t.Fatalf("seed recap upload: %v", err)
	}
}

func TestCreateCampaignRecapMediaIDOR(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedUser(t, 2, "user2", "user")
	testutil.SeedCampaign(t, 1, "Camp1", "P1", 2)
	testutil.SeedCampaign(t, 2, "Camp2", "P2", 1)
	seedRecapUpload(t, 800, "campaign", 2) // accessible to user2 via campaign 2? No, campaign 2 is owned by 1
	seedRecapUpload(t, 801, "campaign", 1) // not accessible to user2
	r := testutil.NewRouterWithUser(func(auth *gin.RouterGroup) {
		auth.POST("/campaigns/:id/recaps", CreateCampaignRecap)
	}, 2, "user")
	// user2 is DM of campaign 1 (seed creates dm membership for owner 2), try to use upload from campaign 2 (owner 1) -> forbidden
	w := testutil.PostJSON(t, r, "/api/campaigns/1/recaps", map[string]any{
		"title":            "Bad Recap",
		"content":          "content",
		"media_upload_ids": []int64{800},
	})
	if w.Code != 403 {
		t.Fatalf("expected 403, got %d %s", w.Code, w.Body.String())
	}
	if testutil.CountRows(t, "campaign_recaps") != 0 {
		t.Fatalf("recap should not have been created on forbidden")
	}
	// happy path: upload owned by campaign 1
	seedRecapUpload(t, 802, "campaign", 1)
	// campaign 1 owned by 2, so upload owner campaign 1 is accessible to user2
	w2 := testutil.PostJSON(t, r, "/api/campaigns/1/recaps", map[string]any{
		"title":            "Good Recap",
		"content":          "content",
		"media_upload_ids": []int64{802},
	})
	if w2.Code != 201 {
		t.Fatalf("expected 201 happy path, got %d %s", w2.Code, w2.Body.String())
	}
}

func TestMarkRecapAsSentMediaIDORAndEmptyBody(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedUser(t, 2, "user2", "user")
	testutil.SeedCampaign(t, 1, "Camp1", "P1", 2)
	testutil.SeedCampaign(t, 2, "Camp2", "P2", 1)
	testutil.SeedRecap(t, 10, 1, "Recap10")
	seedRecapUpload(t, 810, "campaign", 2) // foreign (campaign 2 owned by 1) -> not accessible to user2? Actually campaign 2 owned by 1, not accessible.
	r := testutil.NewRouterWithUser(func(auth *gin.RouterGroup) {
		auth.PUT("/recaps/:id/send", MarkRecapAsSent)
	}, 2, "user")
	// foreign id -> 403, is_sent should remain 0
	w := testutil.PutJSON(t, r, "/api/recaps/10/send", map[string]any{
		"media_upload_ids": []int64{810},
	})
	if w.Code != 403 {
		t.Fatalf("expected 403 for foreign media, got %d %s", w.Code, w.Body.String())
	}
	var isSent int
	_ = db.DB.QueryRow("SELECT is_sent FROM campaign_recaps WHERE id=10").Scan(&isSent)
	if isSent != 0 {
		t.Fatalf("is_sent should not be set on forbidden")
	}
	// invalid JSON -> 400
	req := httptest.NewRequest("PUT", "/api/recaps/10/send", bytes.NewReader([]byte("{invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req)
	if w2.Code != 400 {
		t.Fatalf("expected 400 for invalid json, got %d %s", w2.Code, w2.Body.String())
	}
	// empty body -> 200 (tolerant)
	req3 := httptest.NewRequest("PUT", "/api/recaps/10/send", nil)
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)
	if w3.Code != 200 {
		t.Fatalf("expected 200 for empty body, got %d %s", w3.Code, w3.Body.String())
	}
}
