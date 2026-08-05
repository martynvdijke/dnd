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
