package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
)

func TestFeatCRUD(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "FeatHero", "Human", "Fighter")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/feats", ListFeats)
		auth.POST("/feats", CreateFeat)
		auth.PUT("/feats/:id", UpdateFeat)
		auth.DELETE("/feats/:id", DeleteFeat)
	})

	t.Run("list feats returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/feats?character_id=1")
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("create feat returns 201", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/feats", map[string]any{
			"character_id": 1,
			"name":         "Great Weapon Master",
			"description":  "Power attack with heavy weapons",
			"prerequisites": "Strength 18",
			"source":       "PHB",
			"level_gained": 4,
		})
		testutil.AssertStatus(t, w, 201)
	})

	t.Run("update feat returns 200", func(t *testing.T) {
		w := testutil.PutJSON(t, r, "/api/feats/1", map[string]any{
			"name":        "Great Weapon Master (Updated)",
			"description": "Updated description",
		})
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("delete feat returns 200", func(t *testing.T) {
		w := testutil.Delete(t, r, "/api/feats/1")
		testutil.AssertStatus(t, w, 200)
	})
}
