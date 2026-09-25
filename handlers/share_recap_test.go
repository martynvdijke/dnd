package handlers

import (
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
	"villum/middleware"
)

func TestShareRecap(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 1, "RecapCamp", "Testers", 1)
	testutil.SeedRecap(t, 10, 1, "Session One")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/share", CreateShareLink)
	})

	public := gin.New()
	public.Use(middleware.SecurityHeaders())
	public.GET("/api/share/:token", GetSharedEntity)
	public.GET("/share/:token", GetSharedPage)

	t.Run("create recap share link returns 201", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/share", map[string]any{
			"entity_type": "recap", "entity_id": 10,
		})
		testutil.AssertStatus(t, w, 201)
		var result map[string]any
		testutil.ParseJSON(t, w, &result)
		if result["token"] == "" {
			t.Fatal("expected non-empty token")
		}
		if result["label"] != "Session One" {
			t.Fatalf("expected label 'Session One', got %v", result["label"])
		}

		token := result["token"].(string)

		t.Run("public JSON returns recap", func(t *testing.T) {
			w := testutil.Get(t, public, "/api/share/"+token)
			testutil.AssertStatus(t, w, 200)
			var body map[string]any
			testutil.ParseJSON(t, w, &body)
			if body["title"] != "Session One" {
				t.Fatalf("expected title 'Session One', got %v", body["title"])
			}
			if body["campaign_name"] != "RecapCamp" {
				t.Fatalf("expected campaign_name 'RecapCamp', got %v", body["campaign_name"])
			}
		})

		t.Run("public HTML page renders recap", func(t *testing.T) {
			w := testutil.Get(t, public, "/share/"+token)
			testutil.AssertStatus(t, w, 200)
			if !strings.Contains(w.Body.String(), "Session One") {
				t.Fatalf("expected page to contain recap title, got %q", w.Body.String())
			}
		})
	})

	t.Run("non-member cannot share recap", func(t *testing.T) {
		testutil.SeedUser(t, 2, "stranger", "user")
		r2 := testutil.NewRouterWithUser(func(auth *gin.RouterGroup) {
			auth.POST("/share", CreateShareLink)
		}, 2, "user")
		w := testutil.PostJSON(t, r2, "/api/share", map[string]any{
			"entity_type": "recap", "entity_id": 10,
		})
		testutil.AssertStatus(t, w, 403)
	})
}
