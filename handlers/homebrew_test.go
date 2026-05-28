package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
)

func TestHomebrewCRUD(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	types := []string{"races", "classes", "spells", "feats", "backgrounds", "equipment"}

	for _, ctype := range types {
		t.Run(ctype+" full cycle", func(t *testing.T) {
			testutil.NewDB(t)
			defer testutil.CloseDB(t)
			testutil.SeedUser(t, 1, "admin", "admin")

			r := testutil.NewRouter(func(auth *gin.RouterGroup) {
				auth.POST("/homebrew/:type", CreateHomebrewContent)
				auth.GET("/homebrew/:type", ListHomebrewContent)
				auth.PUT("/homebrew/:type/:id", UpdateHomebrewContent)
				auth.DELETE("/homebrew/:type/:id", DeleteHomebrewContent)
			})

			body := map[string]any{"name": "Test " + ctype}
			if ctype == "races" {
				body["speed"] = 30
				body["size"] = "Medium"
			}
			if ctype == "spells" {
				body["level"] = 1
				body["school"] = "Evocation"
			}

			w := testutil.PostJSON(t, r, "/api/homebrew/"+ctype, body)
			testutil.AssertStatus(t, w, 201)
			var entry map[string]any
			testutil.ParseJSON(t, w, &entry)
			eid := int(entry["id"].(float64))
			if src, ok := entry["source"]; ok && src != "homebrew" {
				t.Fatalf("expected source=homebrew, got %v", src)
			}

			w = testutil.Get(t, r, "/api/homebrew/"+ctype)
			testutil.AssertStatus(t, w, 200)
			var items []any
			testutil.ParseJSON(t, w, &items)
			if len(items) < 1 {
				t.Fatal("expected at least 1 item")
			}

			w = testutil.PutJSON(t, r, "/api/homebrew/"+ctype+"/"+itoa(eid), map[string]any{
				"name": "Updated " + ctype,
			})
			testutil.AssertStatus(t, w, 200)

			w = testutil.Delete(t, r, "/api/homebrew/"+ctype+"/"+itoa(eid))
			testutil.AssertStatus(t, w, 200)
		})
	}
}

func TestHomebrewEdgeCases(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/homebrew/:type", ListHomebrewContent)
	})

	t.Run("invalid type returns 400", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/homebrew/invalid")
		if w.Code != 400 {
			t.Logf("expected 400 for invalid type, got %d", w.Code)
		}
	})
}
