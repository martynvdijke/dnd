package handlers

import (
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
)

func TestResourceCRUD(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "ResourceHero", "Human", "Wizard")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/characters/:id/resources", ListCharacterResources)
		auth.POST("/characters/:id/resources", CreateCharacterResource)
		auth.PUT("/resources/:id", UpdateCharacterResource)
		auth.DELETE("/resources/:id", DeleteCharacterResource)
		auth.POST("/characters/:id/recover-resources", RecoverResourcesOnRest)
	})

	var resID int64

	t.Run("list resources returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/characters/1/resources")
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("create resource returns 201", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/characters/1/resources", map[string]any{
			"name":                "Spell Slot L1",
			"current":             2,
			"max":                 4,
			"short_rest_recovery": 0,
			"long_rest_recovery":  1,
		})
		testutil.AssertStatus(t, w, 201)
		var result map[string]any
		testutil.ParseJSON(t, w, &result)
		resID = int64(result["id"].(float64))
	})

	t.Run("update resource returns 200", func(t *testing.T) {
		if resID == 0 {
			t.Skip("no resource created")
		}
		w := testutil.PutJSON(t, r, "/api/resources/"+strconv.FormatInt(resID, 10), map[string]any{
			"name":    "Spell Slot L1 (Updated)",
			"current": 3,
			"max":     4,
		})
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("recover resources on long rest returns 200", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/characters/1/recover-resources", map[string]any{
			"rest_type": "long",
		})
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("recover resources on short rest returns 200", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/characters/1/recover-resources", map[string]any{
			"rest_type": "short",
		})
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("delete resource returns 200", func(t *testing.T) {
		if resID == 0 {
			t.Skip("no resource created")
		}
		w := testutil.Delete(t, r, "/api/resources/"+strconv.FormatInt(resID, 10))
		testutil.AssertStatus(t, w, 200)
	})
}

// TestDoRestRecoversResources verifies that taking a rest via DoRest restores
// resources with rest recovery, while consumables (max=0) never refill.
func TestDoRestRecoversResources(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "RestHero", "Human", "Monk")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/characters/:id/rest", DoRest)
		auth.GET("/characters/:id/resources", ListCharacterResources)
		auth.POST("/characters/:id/resources", CreateCharacterResource)
	})

	// Ki points recover on both rests; rations (max=0) are consumables.
	testutil.PostJSON(t, r, "/api/characters/1/resources", map[string]any{
		"name": "Ki", "current": 1, "max": 5,
		"short_rest_recovery": 1, "long_rest_recovery": 5,
	})
	testutil.PostJSON(t, r, "/api/characters/1/resources", map[string]any{
		"name": "Rations", "current": 3, "max": 0,
		"short_rest_recovery": 1, "long_rest_recovery": 1,
	})

	currentByName := func() map[string]int {
		t.Helper()
		w := testutil.Get(t, r, "/api/characters/1/resources")
		testutil.AssertStatus(t, w, 200)
		var list []map[string]any
		testutil.ParseJSON(t, w, &list)
		cur := map[string]int{}
		for _, res := range list {
			cur[res["name"].(string)] = int(res["current"].(float64))
		}
		return cur
	}

	t.Run("short rest recovers short_rest_recovery only", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/characters/1/rest", map[string]any{"rest_type": "short"})
		testutil.AssertStatus(t, w, 200)
		cur := currentByName()
		if cur["Ki"] != 2 {
			t.Errorf("Ki current = %d, want 2 after short rest", cur["Ki"])
		}
		if cur["Rations"] != 3 {
			t.Errorf("Rations current = %d, want 3 (consumables never refill)", cur["Rations"])
		}
	})

	t.Run("long rest recovers to max", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/characters/1/rest", map[string]any{"rest_type": "long"})
		testutil.AssertStatus(t, w, 200)
		cur := currentByName()
		if cur["Ki"] != 5 {
			t.Errorf("Ki current = %d, want 5 (max) after long rest", cur["Ki"])
		}
		if cur["Rations"] != 3 {
			t.Errorf("Rations current = %d, want 3 (consumables never refill)", cur["Rations"])
		}
	})
}
