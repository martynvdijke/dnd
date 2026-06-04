package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
)

func TestQuickReference(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/quickref", GetQuickReference)
	})

	t.Run("get full quickref returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/quickref")
		testutil.AssertStatus(t, w, 200)
		var data []any
		testutil.ParseJSON(t, w, &data)
		if len(data) == 0 {
			t.Fatal("expected non-empty quickref array")
		}
	})

	t.Run("get conditions section returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/quickref?section=conditions")
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("get actions section returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/quickref?section=actions")
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("get damage-types section returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/quickref?section=damage-types")
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("get skills section returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/quickref?section=skills")
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("get invalid section returns 404", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/quickref?section=invalid")
		if w.Code != 404 {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})
}
