package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
)

func TestCompareCharacters(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "Hero1", "Human", "Fighter")
	testutil.SeedCharacter(t, 2, 1, "Hero2", "Elf", "Wizard")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/characters/compare", CompareCharacters)
	})

	t.Run("compare two characters returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/characters/compare?ids=1,2")
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("compare single character returns 400", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/characters/compare?ids=1")
		if w.Code != 400 {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("compare with no ids returns 400", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/characters/compare")
		if w.Code != 400 {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("compare with non-existent ids returns 200 or 400", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/characters/compare?ids=1,999")
		if w.Code != 200 && w.Code != 400 {
			t.Fatalf("expected 200 or 400, got %d: %s", w.Code, w.Body.String())
		}
	})
}
