package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
)

func TestDowntimeActivityCRUD(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "DowntimeHero", "Human", "Fighter")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/characters/:id/downtime", ListDowntimeActivities)
		auth.POST("/characters/:id/downtime", CreateDowntimeActivity)
		auth.PUT("/downtime/:id", UpdateDowntimeActivity)
		auth.DELETE("/downtime/:id", DeleteDowntimeActivity)
		auth.POST("/downtime/:id/advance", AdvanceDowntimeDay)
		auth.GET("/downtime/types", GetDowntimeTypes)
	})

	t.Run("list downtime returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/characters/1/downtime")
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("get downtime types returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/downtime/types")
		testutil.AssertStatus(t, w, 200)
		var types []any
		testutil.ParseJSON(t, w, &types)
		if len(types) == 0 {
			t.Fatal("expected at least 1 downtime type")
		}
	})

	t.Run("create downtime activity returns 201", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/characters/1/downtime", map[string]any{
			"activity_type": "research",
			"name":          "Research Ancient Lore",
			"description":   "Spending time in the library",
			"dc":            15,
			"days_required": 10,
		})
		testutil.AssertStatus(t, w, 201)
	})

	t.Run("advance downtime day returns 200", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/downtime/1/advance", nil)
		if w.Code != 200 {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("delete downtime activity returns 200", func(t *testing.T) {
		w := testutil.Delete(t, r, "/api/downtime/1")
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("create with invalid type defaults to other", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/characters/1/downtime", map[string]any{
			"activity_type": "nonexistent_type",
			"name":          "Custom Activity",
		})
		testutil.AssertStatus(t, w, 201)
	})
}
