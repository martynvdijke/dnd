package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers/testutil"
)

func seedOneShotAct(t *testing.T, adventureID, actID int64) {
	t.Helper()
	_, err := 	db.DB.Exec(
		"INSERT OR IGNORE INTO oneshot_acts(id, adventure_id, number, title, description, estimated_minutes, notes) VALUES(?,?,1,?,?,30,'')",
		actID, adventureID, "Act One", "First act",
	)
	if err != nil {
		t.Fatalf("seed oneshot act: %v", err)
	}
}

func seedNPC(t *testing.T, npcID int64) {
	t.Helper()
	_, err := db.DB.Exec(
		"INSERT OR IGNORE INTO npcs(id, user_id, name) VALUES(?,?,'TestNPC')",
		npcID, 1,
	)
	if err != nil {
		t.Fatalf("seed npc: %v", err)
	}
}

func seedOneShotNPC(t *testing.T, adventureID int64) {
	t.Helper()
	_, err := db.DB.Exec(
		"INSERT OR IGNORE INTO oneshot_adventure_npcs(adventure_id, npc_id, role) VALUES(?,1,'helper')",
		adventureID,
	)
	if err != nil {
		t.Fatalf("seed oneshot npc: %v", err)
	}
}

func TestActNPCCRUD(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "TestChar", "Human", "Fighter")

	// Seed a oneshot adventure
	var advID int64 = 1
	_, err := 	db.DB.Exec(
		"INSERT OR IGNORE INTO oneshot_adventures(id, user_id, title, premise, hook, template, estimated_minutes, difficulty) VALUES(?,?,'Test Adventure','Premise','Hook','custom',60,'medium')",
		advID, 1,
	)
	if err != nil {
		t.Fatalf("seed oneshot adventure: %v", err)
	}

	seedNPC(t, 1)
	seedOneShotAct(t, advID, 1)
	seedOneShotNPC(t, advID)

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/oneshot-acts/:id/npcs", ListActNPCs)
		auth.POST("/oneshot-acts/:id/npcs", CreateActNPC)
		auth.DELETE("/oneshot-acts/:id/npcs/:nid", DeleteActNPC)
	})

	t.Run("list act npcs returns 200 with empty list", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/oneshot-acts/1/npcs")
		testutil.AssertStatus(t, w, 200)
		var npcs []any
		testutil.ParseJSON(t, w, &npcs)
		if len(npcs) != 0 {
			t.Fatalf("expected empty list, got %d items", len(npcs))
		}
	})

	t.Run("create inline act npc returns 201", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/oneshot-acts/1/npcs", map[string]any{
			"name": "Goblin Guard",
			"role": "guard",
			"notes": "At the entrance",
		})
		testutil.AssertStatus(t, w, 201)
		var npc map[string]any
		testutil.ParseJSON(t, w, &npc)
		testutil.AssertField(t, npc, "name", "Goblin Guard")
		testutil.AssertField(t, npc, "role", "guard")
		testutil.AssertField(t, npc, "is_inline", true)
	})

	t.Run("create linked npc returns 201", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/oneshot-acts/1/npcs", map[string]any{
			"npc_id": 1,
			"name":   "Linked NPC",
			"role":   "helper",
		})
		testutil.AssertStatus(t, w, 201)
		var npc map[string]any
		testutil.ParseJSON(t, w, &npc)
		if npc["npc_id"] == nil {
			t.Fatal("expected npc_id to be set")
		}
		testutil.AssertField(t, npc, "is_inline", false)
	})

	t.Run("list act npcs returns 2 items", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/oneshot-acts/1/npcs")
		testutil.AssertStatus(t, w, 200)
		var npcs []any
		testutil.ParseJSON(t, w, &npcs)
		if len(npcs) != 2 {
			t.Fatalf("expected 2 npcs, got %d: %s", len(npcs), w.Body.String())
		}
	})

	t.Run("delete act npc returns 200", func(t *testing.T) {
		w := testutil.Delete(t, r, "/api/oneshot-acts/1/npcs/1")
		testutil.AssertStatus(t, w, 200)
		var resp map[string]any
		testutil.ParseJSON(t, w, &resp)
		testutil.AssertField(t, resp, "ok", true)
	})

	t.Run("list after delete returns 1 item", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/oneshot-acts/1/npcs")
		testutil.AssertStatus(t, w, 200)
		var npcs []any
		testutil.ParseJSON(t, w, &npcs)
		if len(npcs) != 1 {
			t.Fatalf("expected 1 npc after delete, got %d: %s", len(npcs), w.Body.String())
		}
	})
}

func TestActNotes(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "NoteChar", "Human", "Fighter")

	_, err := 	db.DB.Exec(
		"INSERT OR IGNORE INTO oneshot_adventures(id, user_id, title) VALUES(?,?,'Note Adventure')",
		1, 1,
	)
	if err != nil {
		t.Fatalf("seed adventure: %v", err)
	}
	seedOneShotAct(t, 1, 1)

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/oneshot-acts/:id/notes", ListActNotes)
		auth.POST("/oneshot-acts/:id/notes", CreateActNote)
	})

	t.Run("list act notes returns 200 empty", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/oneshot-acts/1/notes")
		testutil.AssertStatus(t, w, 200)
		var notes []any
		testutil.ParseJSON(t, w, &notes)
		if len(notes) != 0 {
			t.Fatalf("expected empty list, got %d", len(notes))
		}
	})

	t.Run("create act note returns 201", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/oneshot-acts/1/notes", map[string]any{
			"title":   "Test Note",
			"content": "This is a test DM note",
		})
		testutil.AssertStatus(t, w, 201)
	})

	t.Run("list after create returns 1 item", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/oneshot-acts/1/notes")
		testutil.AssertStatus(t, w, 200)
		var notes []any
		testutil.ParseJSON(t, w, &notes)
		if len(notes) != 1 {
			t.Fatalf("expected 1 note, got %d: %s", len(notes), w.Body.String())
		}
	})
}

func TestActDetailsHTMX(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "DetailChar", "Human", "Fighter")

	_, err := 	db.DB.Exec(
		"INSERT OR IGNORE INTO oneshot_adventures(id, user_id, title) VALUES(?,?,'Detail Adventure')",
		1, 1,
	)
	if err != nil {
		t.Fatalf("seed adventure: %v", err)
	}
	seedOneShotAct(t, 1, 1)

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/htmx/oneshot-acts/:id/details", HtmxActDetails)
	})

	t.Run("htmx details returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/htmx/oneshot-acts/1/details")
		testutil.AssertStatus(t, w, 200)
		body := w.Body.String()
		if len(body) == 0 {
			t.Fatal("expected non-empty body")
		}
		if !contains(body, "No NPCs") {
			t.Log("expected details to mention 'No NPCs'")
		}
	})
}

func TestActNPCEdgeCases(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/oneshot-acts/:id/npcs", ListActNPCs)
		auth.POST("/oneshot-acts/:id/npcs", CreateActNPC)
		auth.DELETE("/oneshot-acts/:id/npcs/:nid", DeleteActNPC)
		auth.GET("/oneshot-acts/:id/notes", ListActNotes)
		auth.POST("/oneshot-acts/:id/notes", CreateActNote)
	})

	t.Run("list npcs with invalid act id returns 400", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/oneshot-acts/abc/npcs")
		testutil.AssertStatus(t, w, 400)
	})

	t.Run("list notes with invalid act id returns 400", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/oneshot-acts/abc/notes")
		testutil.AssertStatus(t, w, 400)
	})

	t.Run("create npc with invalid act id returns 400", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/oneshot-acts/abc/npcs", map[string]any{"name": "Bad"})
		testutil.AssertStatus(t, w, 400)
	})

	t.Run("create note with invalid act id returns 400", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/oneshot-acts/abc/notes", map[string]any{"title": "Bad"})
		testutil.AssertStatus(t, w, 400)
	})

	t.Run("delete npc with invalid act id returns 400", func(t *testing.T) {
		w := testutil.Delete(t, r, "/api/oneshot-acts/abc/npcs/1")
		testutil.AssertStatus(t, w, 400)
	})

	t.Run("delete npc with invalid npc id returns 400", func(t *testing.T) {
		w := testutil.Delete(t, r, "/api/oneshot-acts/1/npcs/abc")
		testutil.AssertStatus(t, w, 400)
	})

	t.Run("create npc with invalid json returns 400", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/oneshot-acts/1/npcs", "not-json")
		testutil.AssertStatus(t, w, 400)
	})

	t.Run("create note with invalid json returns 400", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/oneshot-acts/1/notes", "not-json")
		testutil.AssertStatus(t, w, 400)
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
