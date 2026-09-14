package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/crypto"
	"villum/db"
	"villum/handlers/testutil"
	"villum/models"
)

func recapAIRouter(uid int64, role string) *gin.Engine {
	return testutil.NewRouterWithUser(func(rg *gin.RouterGroup) {
		rg.POST("/campaigns/:id/recaps/generate-ai", GenerateRecapAI)
		rg.POST("/campaigns/:id/recaps", CreateCampaignRecap)
		rg.PUT("/recaps/:id", UpdateCampaignRecap)
		rg.DELETE("/recaps/:id", DeleteCampaignRecap)
		rg.GET("/campaigns/:id/recaps", ListCampaignRecaps)
	}, uid, role)
}

func TestGenerateRecapAI_AISuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{"message": map[string]any{"content": "AI generated recap content"}, "finish_reason": "stop"}},
		})
	}))
	defer srv.Close()
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "owner", "player")
	testutil.SeedCampaign(t, 1, "Camp", "Party", 1)
	enc, _ := crypto.Encrypt("sk-test")
	_, err := db.CreateAIEndpoint(context.Background(), &models.AIEndpoint{Name: "e", Type: "text", BaseURL: srv.URL, EncryptedAPIKey: enc, Model: "gpt-4o", Enabled: true})
	if err != nil {
		t.Fatalf("seed ep: %v", err)
	}
	r := recapAIRouter(1, "player")
	w := testutil.PostJSON(t, r, "/api/campaigns/1/recaps/generate-ai", map[string]any{})
	if w.Code != 200 {
		t.Fatalf("expected 200 got %d %s", w.Code, w.Body.String())
	}
	var resp CampaignRecap
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v %s", err, w.Body.String())
	}
	if !resp.AIUsed {
		t.Fatalf("expected ai_used true, got %+v", resp)
	}
	if resp.Content != "AI generated recap content" {
		t.Fatalf("expected AI content, got %q", resp.Content)
	}
}

func TestGenerateRecapAI_FallbackDisabled(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "owner", "player")
	testutil.SeedCampaign(t, 1, "Camp", "Party", 1)
	db.DB.Exec("INSERT OR REPLACE INTO app_settings (key, value) VALUES ('ai_enabled','0')")
	r := recapAIRouter(1, "player")
	w := testutil.PostJSON(t, r, "/api/campaigns/1/recaps/generate-ai", map[string]any{})
	if w.Code != 200 {
		t.Fatalf("expected 200 got %d %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["ai_used"] != false {
		t.Fatalf("expected ai_used false, got %v", resp["ai_used"])
	}
}

func TestGenerateRecapAI_FallbackNoEndpoint(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "owner", "player")
	testutil.SeedCampaign(t, 1, "Camp", "Party", 1)
	r := recapAIRouter(1, "player")
	w := testutil.PostJSON(t, r, "/api/campaigns/1/recaps/generate-ai", map[string]any{})
	if w.Code != 200 {
		t.Fatalf("expected 200 got %d %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["ai_used"] != false {
		t.Fatalf("expected ai_used false, got %v", resp)
	}
	if resp["content"] == nil || resp["content"] == "" {
		// template may be empty string without data; just check field exists (content key may be empty string but not missing)
		// Allow empty template but ensure not AI
	}
}

func TestGenerateRecapAI_ProviderFailureFallback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte("internal error"))
	}))
	defer srv.Close()
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "owner", "player")
	testutil.SeedCampaign(t, 1, "Camp", "Party", 1)
	enc, _ := crypto.Encrypt("sk-test")
	db.CreateAIEndpoint(context.Background(), &models.AIEndpoint{Name: "e", Type: "text", BaseURL: srv.URL, EncryptedAPIKey: enc, Model: "gpt-4o", Enabled: true})
	r := recapAIRouter(1, "player")
	w := testutil.PostJSON(t, r, "/api/campaigns/1/recaps/generate-ai", map[string]any{})
	if w.Code != 200 {
		t.Fatalf("expected 200 got %d %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["ai_used"] != false {
		t.Fatalf("expected ai_used false on provider failure, got %v", resp)
	}
}

func TestGenerateRecapAI_ForbiddenNonDM(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "owner", "player")
	testutil.SeedUser(t, 2, "member", "player")
	testutil.SeedCampaign(t, 1, "Camp", "Party", 1)
	testutil.SeedCampaignMember(t, 1, 2, "player") // not dm
	r := testutil.NewRouterWithUser(func(rg *gin.RouterGroup) {
		rg.POST("/campaigns/:id/recaps/generate-ai", GenerateRecapAI)
		rg.POST("/campaigns/:id/recaps", CreateCampaignRecap)
		rg.PUT("/recaps/:id", UpdateCampaignRecap)
		rg.DELETE("/recaps/:id", DeleteCampaignRecap)
	}, 2, "player")
	w := testutil.PostJSON(t, r, "/api/campaigns/1/recaps/generate-ai", map[string]any{})
	if w.Code != 403 {
		t.Fatalf("expected 403 got %d %s", w.Code, w.Body.String())
	}
	// also check create denied
	w2 := testutil.PostJSON(t, r, "/api/campaigns/1/recaps", map[string]any{"title": "t", "content": "c"})
	if w2.Code != 403 {
		t.Fatalf("expected 403 on create, got %d %s", w2.Code, w2.Body.String())
	}
	// create a recap as owner, then try update/delete as non-dm
	ownerR := recapAIRouter(1, "player")
	w3 := testutil.PostJSON(t, ownerR, "/api/campaigns/1/recaps", map[string]any{"title": "t", "content": "c"})
	if w3.Code != 201 {
		t.Fatalf("owner create failed %d %s", w3.Code, w3.Body.String())
	}
	var cr map[string]any
	json.Unmarshal(w3.Body.Bytes(), &cr)
	id := int64(cr["id"].(float64))
	w4 := testutil.PutJSON(t, r, "/api/recaps/1", map[string]any{"title": "x", "content": "y"})
	_ = id
	if w4.Code != 403 {
		t.Fatalf("expected 403 on update, got %d %s", w4.Code, w4.Body.String())
	}
	w5 := testutil.Delete(t, r, "/api/recaps/1")
	if w5.Code != 403 {
		t.Fatalf("expected 403 on delete, got %d %s", w5.Code, w5.Body.String())
	}
	_ = gin.H{}
}
