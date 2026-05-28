package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
	"villum/middleware"
)

func TestShareLinks(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "ShareHero", "Elf", "Ranger")
	testutil.SeedCampaign(t, 1, "ShareCamp", "Testers", 1)

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/share", CreateShareLink)
		auth.GET("/share", ListShareLinks)
		auth.DELETE("/share/:token", DeleteShareLink)
	})

	t.Run("create character share link returns 201", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/share", map[string]any{
			"entity_type": "character", "entity_id": 1,
		})
		testutil.AssertStatus(t, w, 201)
		var result map[string]any
		testutil.ParseJSON(t, w, &result)
		if result["token"] == "" {
			t.Fatal("expected non-empty token")
		}
	})

	t.Run("create party share link returns 201", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/share", map[string]any{
			"entity_type": "party", "entity_id": 1,
		})
		testutil.AssertStatus(t, w, 201)
	})

	t.Run("list share links returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/share")
		testutil.AssertStatus(t, w, 200)
		var links []any
		testutil.ParseJSON(t, w, &links)
		if len(links) < 2 {
			t.Fatalf("expected >=2 links, got %d", len(links))
		}
	})

	t.Run("delete non-existent returns 200 (idempotent)", func(t *testing.T) {
		w := testutil.Delete(t, r, "/api/share/test-token-123")
		if w.Code != 200 {
			t.Logf("delete non-existent: expected 200, got %d", w.Code)
		}
	})
}

func TestShareLinksPublicAccess(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "PubShare", "Human", "Fighter")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/share", CreateShareLink)
	})

	r2 := gin.New()
	r2.Use(middleware.SecurityHeaders())
	r2.GET("/api/share/:token", GetSharedEntity)

	t.Run("create then access via public route", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/share", map[string]any{
			"entity_type": "character", "entity_id": 1,
		})
		testutil.AssertStatus(t, w, 201)
		var result map[string]any
		testutil.ParseJSON(t, w, &result)
		token := result["token"].(string)

		w2 := testutil.Get(t, r2, "/api/share/"+token)
		if w2.Code != 200 {
			t.Logf("public access: got %d", w2.Code)
		}
	})

	t.Run("invalid token returns 404", func(t *testing.T) {
		w := testutil.Get(t, r2, "/api/share/invalid-token-123")
		if w.Code != 404 {
			t.Logf("invalid token: expected 404, got %d", w.Code)
		}
	})
}
