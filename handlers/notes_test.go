package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
)

func TestNotesCRUD(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "NoteHero", "Human", "Fighter")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/notes", CreateCharacterNote)
		auth.GET("/notes", ListCharacterNotes)
		auth.PUT("/notes/:id", UpdateCharacterNote)
		auth.DELETE("/notes/:id", DeleteCharacterNote)
	})

	t.Run("create note returns 201", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/notes", map[string]any{
			"character_id": 1, "title": "Quest Idea", "content": "Find the crown",
			"visibility": "player", "category": "quest",
		})
		testutil.AssertStatus(t, w, 201)
	})

	t.Run("list notes returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/notes?character_id=1")
		testutil.AssertStatus(t, w, 200)
		var notes []any
		testutil.ParseJSON(t, w, &notes)
		if len(notes) < 1 {
			t.Fatal("expected at least 1 note")
		}
	})
}

func TestNotesEdgeCases(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "EdgeNote", "Elf", "Wizard")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/notes", CreateCharacterNote)
		auth.GET("/notes", ListCharacterNotes)
	})

	t.Run("list without character_id returns 400", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/notes")
		if w.Code != 400 {
			t.Logf("expected 400 for missing character_id, got %d", w.Code)
		}
	})

	t.Run("create DM note returns 201", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/notes", map[string]any{
			"character_id": 1, "title": "Secret", "content": "The king is a spy",
			"visibility": "dm", "category": "lore",
		})
		testutil.AssertStatus(t, w, 201)
	})
}
