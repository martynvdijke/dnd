package handlers

import (
	"bytes"
	"fmt"
	"net/http/httptest"
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

func TestSessionPacing(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "PacingChar", "Human", "Fighter")
	testutil.SeedOneShot(t, 1, 1, "Pacing Adventure")
	testutil.SeedOneShotAct(t, 1, 1, "Act One", 1)
	testutil.SeedOneShotScene(t, 1, 1, "Scene One", 1)

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/oneshot-adventures/:id/pacing/start", StartPacingSession)
		auth.GET("/oneshot-adventures/:id/pacing", GetPacingSession)
		auth.POST("/session-pacing/:id/pause", PausePacingSession)
		auth.POST("/session-pacing/:id/resume", ResumePacingSession)
		auth.POST("/session-pacing/:id/complete", CompletePacingSession)
		auth.POST("/session-pacing/:id/next-scene", AdvanceToNextScene)
		auth.POST("/session-pacing/:id/tick", UpdatePacingTimers)
	})

	var pacingID float64

	t.Run("start pacing returns 201", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/oneshot-adventures/1/pacing/start", map[string]any{
			"current_act_id": 1, "current_scene_id": 1,
		})
		testutil.AssertStatus(t, w, 201)
		var result map[string]any
		testutil.ParseJSON(t, w, &result)
		id, ok := result["id"].(float64)
		if !ok {
			t.Fatal("start pacing response missing id")
		}
		pacingID = id
	})

	t.Run("get pacing returns 200", func(t *testing.T) {
		if pacingID == 0 {
			t.Skip("no pacing session")
		}
		w := testutil.Get(t, r, "/api/oneshot-adventures/1/pacing")
		testutil.AssertStatus(t, w, 200)
		var result map[string]any
		testutil.ParseJSON(t, w, &result)
		if result["status"] != "running" {
			t.Fatalf("expected status running, got %v", result["status"])
		}
	})

	t.Run("pause pacing returns 200", func(t *testing.T) {
		if pacingID == 0 {
			t.Skip("no pacing session")
		}
		w := testutil.PostJSON(t, r, "/api/session-pacing/"+formatInt(pacingID)+"/pause", nil)
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("resume pacing returns 200", func(t *testing.T) {
		if pacingID == 0 {
			t.Skip("no pacing session")
		}
		w := testutil.PostJSON(t, r, "/api/session-pacing/"+formatInt(pacingID)+"/resume", nil)
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("advance to next scene returns 200", func(t *testing.T) {
		if pacingID == 0 {
			t.Skip("no pacing session")
		}
		// Seed a second scene to advance to
		testutil.SeedOneShotScene(t, 2, 1, "Scene Two", 2)
		// Need to set current_scene_id to scene 1 first via DB
		db.DB.Exec("UPDATE session_pacing SET current_scene_id = 1 WHERE id = ?", int64(pacingID))

		w := testutil.PostJSON(t, r, "/api/session-pacing/"+formatInt(pacingID)+"/next-scene", nil)
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("update pacing timers returns 200", func(t *testing.T) {
		if pacingID == 0 {
			t.Skip("no pacing session")
		}
		w := testutil.PostJSON(t, r, "/api/session-pacing/"+formatInt(pacingID)+"/tick", map[string]any{
			"elapsed_seconds": 60,
		})
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("complete pacing returns 200", func(t *testing.T) {
		if pacingID == 0 {
			t.Skip("no pacing session")
		}
		w := testutil.PostJSON(t, r, "/api/session-pacing/"+formatInt(pacingID)+"/complete", nil)
		testutil.AssertStatus(t, w, 200)
	})
}

func TestClueTrackerCRUD(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedOneShot(t, 1, 1, "Clue Adventure")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/oneshot-adventures/:id/clues", ListClues)
		auth.POST("/oneshot-adventures/:id/clues", CreateClue)
		auth.PUT("/clues/:id", UpdateClue)
		auth.DELETE("/clues/:id", DeleteClue)
	})

	var clueID float64

	t.Run("list clues returns 200 empty", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/oneshot-adventures/1/clues")
		testutil.AssertStatus(t, w, 200)
		var clues []any
		testutil.ParseJSON(t, w, &clues)
		if len(clues) != 0 {
			t.Fatalf("expected empty list, got %d", len(clues))
		}
	})

	t.Run("create clue returns 201", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/oneshot-adventures/1/clues", map[string]any{
			"title":       "Secret Passage",
			"description": "A hidden door behind the bookshelf",
			"clue_type":   "location",
		})
		testutil.AssertStatus(t, w, 201)
		var result map[string]any
		testutil.ParseJSON(t, w, &result)
		id, ok := result["id"].(float64)
		if !ok {
			t.Fatal("create clue response missing id")
		}
		clueID = id
	})

	t.Run("update clue returns 200", func(t *testing.T) {
		if clueID == 0 {
			t.Skip("no clue id")
		}
		w := testutil.PutJSON(t, r, "/api/clues/"+formatInt(clueID), map[string]any{
			"title":       "Secret Passage Updated",
			"description": "Updated description",
			"clue_type":   "location",
		})
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("list after create returns 1 item", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/oneshot-adventures/1/clues")
		testutil.AssertStatus(t, w, 200)
		var clues []any
		testutil.ParseJSON(t, w, &clues)
		if len(clues) != 1 {
			t.Fatalf("expected 1 clue, got %d", len(clues))
		}
	})

	t.Run("delete clue returns 200", func(t *testing.T) {
		if clueID == 0 {
			t.Skip("no clue id")
		}
		w := testutil.Delete(t, r, "/api/clues/"+formatInt(clueID))
		testutil.AssertStatus(t, w, 200)
	})
}

func TestPrepChecklistCRUD(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedOneShot(t, 1, 1, "Checklist Adventure")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/oneshot-adventures/:id/checklist", ListPrepChecklist)
		auth.POST("/oneshot-adventures/:id/checklist", CreatePrepChecklistItem)
		auth.PUT("/prep-checklist/:cid", UpdatePrepChecklistItem)
		auth.DELETE("/prep-checklist/:cid", DeletePrepChecklistItem)
	})

	var itemID float64

	t.Run("create checklist item returns 201", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/oneshot-adventures/1/checklist", map[string]any{
			"item":     "Prepare battle map",
			"category": "setup",
		})
		testutil.AssertStatus(t, w, 201)
		var result map[string]any
		testutil.ParseJSON(t, w, &result)
		id, ok := result["id"].(float64)
		if !ok {
			t.Fatal("create item response missing id")
		}
		itemID = id
	})

	t.Run("list checklist returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/oneshot-adventures/1/checklist")
		testutil.AssertStatus(t, w, 200)
		var items []any
		testutil.ParseJSON(t, w, &items)
		if len(items) < 1 {
			t.Fatal("expected at least 1 item")
		}
	})

	t.Run("update checklist item returns 200", func(t *testing.T) {
		if itemID == 0 {
			t.Skip("no item id")
		}
		w := testutil.PutJSON(t, r, "/api/prep-checklist/"+formatInt(itemID), map[string]any{
			"item": "Prepare battle map and minis",
			"is_checked": true,
		})
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("delete checklist item returns 200", func(t *testing.T) {
		if itemID == 0 {
			t.Skip("no item id")
		}
		w := testutil.Delete(t, r, "/api/prep-checklist/"+formatInt(itemID))
		testutil.AssertStatus(t, w, 200)
	})
}

func TestDMNotesCRUD(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedOneShot(t, 1, 1, "DM Notes Adventure")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/oneshot-adventures/:id/notes", ListDmNotes)
		auth.POST("/oneshot-adventures/:id/notes", CreateDmNote)
		auth.PUT("/dm-notes/:nid", UpdateDmNote)
		auth.DELETE("/oneshot-adventures/:id/notes/:nid", DeleteDmNote)
	})

	var noteID float64

	t.Run("list notes returns 200 empty", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/oneshot-adventures/1/notes")
		testutil.AssertStatus(t, w, 200)
		var notes []any
		testutil.ParseJSON(t, w, &notes)
		if len(notes) != 0 {
			t.Fatalf("expected empty list, got %d", len(notes))
		}
	})

	t.Run("create note returns 201", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/oneshot-adventures/1/notes", map[string]any{
			"title":   "DM Session Notes",
			"content": "Players discovered the secret passage",
		})
		testutil.AssertStatus(t, w, 201)
		var result map[string]any
		testutil.ParseJSON(t, w, &result)
		id, ok := result["id"].(float64)
		if !ok {
			t.Fatal("create note response missing id")
		}
		noteID = id
	})

	t.Run("list notes returns 1 item", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/oneshot-adventures/1/notes")
		testutil.AssertStatus(t, w, 200)
		var notes []any
		testutil.ParseJSON(t, w, &notes)
		if len(notes) != 1 {
			t.Fatalf("expected 1 note, got %d", len(notes))
		}
	})

	t.Run("update note returns 200", func(t *testing.T) {
		if noteID == 0 {
			t.Skip("no note id")
		}
		w := testutil.PutJSON(t, r, "/api/dm-notes/"+formatInt(noteID), map[string]any{
			"title":   "Updated DM Notes",
			"content": "Updated with more details",
		})
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("delete note returns 200", func(t *testing.T) {
		if noteID == 0 {
			t.Skip("no note id")
		}
		w := testutil.Delete(t, r, "/api/oneshot-adventures/1/notes/"+formatInt(noteID))
		testutil.AssertStatus(t, w, 200)
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

func seedAdventure(t *testing.T, id, userID int64, title string) {
	t.Helper()
	_, err := db.DB.Exec(
		"INSERT OR IGNORE INTO oneshot_adventures(id, user_id, title) VALUES(?,?,?)",
		id, userID, title,
	)
	if err != nil {
		t.Fatalf("seed adventure: %v", err)
	}
}

func TestActTree(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "TreeChar", "Human", "Fighter")
	seedAdventure(t, 1, 1, "Tree Adventure")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/oneshot-adventures/:id/acts", CreateOneShotAct)
		auth.GET("/oneshot-adventures/:id", GetOneShotAdventure)
		auth.DELETE("/oneshot-acts/:id", DeleteOneShotAct)
		auth.PUT("/oneshot-acts/:id", UpdateOneShotAct)
	})

	// Create root act
	w := testutil.PostJSON(t, r, "/api/oneshot-adventures/1/acts", map[string]any{
		"title": "Root Act", "description": "The root act",
		"number": 1, "sort_order": 1,
	})
	testutil.AssertStatus(t, w, 201)
	var rootResp map[string]any
	testutil.ParseJSON(t, w, &rootResp)
	rootID := int64(rootResp["id"].(float64))

	// Create sub-act (child of root)
	w = testutil.PostJSON(t, r, "/api/oneshot-adventures/1/acts", map[string]any{
		"title": "Sub Act", "description": "Child act",
		"parent_act_id": rootID, "number": 1, "sort_order": 1,
	})
	testutil.AssertStatus(t, w, 201)
	var subResp map[string]any
	testutil.ParseJSON(t, w, &subResp)
	subID := int64(subResp["id"].(float64))

	// Create grandchild sub-act
	w = testutil.PostJSON(t, r, "/api/oneshot-adventures/1/acts", map[string]any{
		"title": "Grandchild Act", "description": "Grandchild act",
		"parent_act_id": subID, "number": 1, "sort_order": 1,
	})
	testutil.AssertStatus(t, w, 201)

	// Get adventure detail and verify tree
	w = testutil.Get(t, r, "/api/oneshot-adventures/1")
	testutil.AssertStatus(t, w, 200)
	var adv map[string]any
	testutil.ParseJSON(t, w, &adv)
	acts := adv["acts"].([]any)
	if len(acts) != 1 {
		t.Fatalf("expected 1 root act, got %d", len(acts))
	}
	rootAct := acts[0].(map[string]any)
	if rootAct["title"] != "Root Act" {
		t.Fatalf("expected Root Act, got %v", rootAct["title"])
	}
	children := rootAct["children"].([]any)
	if len(children) != 1 {
		t.Fatalf("expected 1 child of root act, got %d", len(children))
	}
	childAct := children[0].(map[string]any)
	if childAct["title"] != "Sub Act" {
		t.Fatalf("expected Sub Act, got %v", childAct["title"])
	}
	grandChildren := childAct["children"].([]any)
	if len(grandChildren) != 1 {
		t.Fatalf("expected 1 grandchild of sub act, got %d", len(grandChildren))
	}
	grandchildAct := grandChildren[0].(map[string]any)
	if grandchildAct["title"] != "Grandchild Act" {
		t.Fatalf("expected Grandchild Act, got %v", grandchildAct["title"])
	}
}

func TestSceneSortOrder(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "SceneChar", "Human", "Fighter")
	seedAdventure(t, 1, 1, "Scene Sort Adventure")

	// Seed act
	_, err := db.DB.Exec(
		"INSERT OR IGNORE INTO oneshot_acts(id, adventure_id, number, title, description, sort_order) VALUES(?,?,?,?,?,?)",
		1, 1, 1, "Test Act", "Act with scenes", 1,
	)
	if err != nil {
		t.Fatalf("seed act: %v", err)
	}

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/oneshot-acts/:id/scenes", CreateOneShotScene)
		auth.GET("/oneshot-adventures/:id", GetOneShotAdventure)
	})

	// Create scenes with specific sort_order
	sortOrders := []int{3, 1, 2}
	titles := []string{"Scene C", "Scene A", "Scene B"}
	sceneIDs := make([]int64, 3)
	for i := 0; i < 3; i++ {
		w := testutil.PostJSON(t, r, "/api/oneshot-acts/1/scenes", map[string]any{
			"title": titles[i], "description": "Test",
			"number": i + 1, "sort_order": sortOrders[i], "scene_type": "roleplay",
		})
		testutil.AssertStatus(t, w, 201)
		var sc map[string]any
		testutil.ParseJSON(t, w, &sc)
		sceneIDs[i] = int64(sc["id"].(float64))
	}

	// Get adventure, verify scenes are returned with sort_order
	w := testutil.Get(t, r, "/api/oneshot-adventures/1")
	testutil.AssertStatus(t, w, 200)
	var adv map[string]any
	testutil.ParseJSON(t, w, &adv)
	acts := adv["acts"].([]any)
	if len(acts) == 0 {
		t.Fatal("expected at least 1 act")
	}
	act := acts[0].(map[string]any)
	scenes := act["scenes"].([]any)
	if len(scenes) != 3 {
		t.Fatalf("expected 3 scenes, got %d", len(scenes))
	}
	// Verify sort_order values
	for _, s := range scenes {
		sc := s.(map[string]any)
		title := sc["title"].(string)
		so := int(sc["sort_order"].(float64))
		switch title {
		case "Scene A":
			if so != 1 {
				t.Fatalf("Scene A expected sort_order=1, got %d", so)
			}
		case "Scene B":
			if so != 2 {
				t.Fatalf("Scene B expected sort_order=2, got %d", so)
			}
		case "Scene C":
			if so != 3 {
				t.Fatalf("Scene C expected sort_order=3, got %d", so)
			}
		}
	}
}

func TestActReorder(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "ReorderChar", "Human", "Fighter")
	seedAdventure(t, 1, 1, "Reorder Adventure")

	// Seed 3 acts with sequential sort_order
	for i := 1; i <= 3; i++ {
		_, err := db.DB.Exec(
			"INSERT INTO oneshot_acts(id, adventure_id, number, title, sort_order) VALUES(?,?,?,?,?)",
			int64(i), 1, i, fmt.Sprintf("Act %d", i), i,
		)
		if err != nil {
			t.Fatalf("seed act %d: %v", i, err)
		}
	}

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.PUT("/oneshot-adventures/:id/acts/reorder", ReorderActs)
		auth.GET("/oneshot-adventures/:id", GetOneShotAdventure)
	})

	// Reorder: 3, 2, 1
	w := testutil.PutJSON(t, r, "/api/oneshot-adventures/1/acts/reorder", map[string]any{
		"order": []int64{3, 2, 1},
	})
	testutil.AssertStatus(t, w, 200)

	// Verify sort_order database updated
	var so1, so2, so3 int
	db.DB.QueryRow("SELECT sort_order FROM oneshot_acts WHERE id=1").Scan(&so1)
	db.DB.QueryRow("SELECT sort_order FROM oneshot_acts WHERE id=2").Scan(&so2)
	db.DB.QueryRow("SELECT sort_order FROM oneshot_acts WHERE id=3").Scan(&so3)
	if so1 != 3 || so2 != 2 || so3 != 1 {
		t.Fatalf("expected sort_order 3,2,1 got %d,%d,%d", so1, so2, so3)
	}
}

func TestSceneReorder(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "SceneReorderChar", "Human", "Fighter")
	seedAdventure(t, 1, 1, "Scene Reorder Adventure")

	// Seed act
	_, err := db.DB.Exec(
		"INSERT INTO oneshot_acts(id, adventure_id, number, title, sort_order) VALUES(?,?,?,?,?)",
		1, 1, 1, "Test Act", 1,
	)
	if err != nil {
		t.Fatalf("seed act: %v", err)
	}

	// Seed 3 scenes with sort_order
	for i := 1; i <= 3; i++ {
		_, err := db.DB.Exec(
			"INSERT INTO oneshot_scenes(id, act_id, number, title, scene_type, sort_order) VALUES(?,?,?,?,?,?)",
			int64(i), 1, i, fmt.Sprintf("Scene %d", i), "roleplay", i,
		)
		if err != nil {
			t.Fatalf("seed scene %d: %v", i, err)
		}
	}

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.PUT("/oneshot-acts/:id/scenes/reorder", ReorderScenes)
		auth.GET("/oneshot-adventures/:id", GetOneShotAdventure)
	})

	// Reorder: 3, 2, 1
	w := testutil.PutJSON(t, r, "/api/oneshot-acts/1/scenes/reorder", map[string]any{
		"order": []int64{3, 2, 1},
	})
	testutil.AssertStatus(t, w, 200)

	// Verify sort_order in DB
	var so1, so2, so3 int
	db.DB.QueryRow("SELECT sort_order FROM oneshot_scenes WHERE id=1").Scan(&so1)
	db.DB.QueryRow("SELECT sort_order FROM oneshot_scenes WHERE id=2").Scan(&so2)
	db.DB.QueryRow("SELECT sort_order FROM oneshot_scenes WHERE id=3").Scan(&so3)
	if so1 != 3 || so2 != 2 || so3 != 1 {
		t.Fatalf("expected sort_order 3,2,1 got %d,%d,%d", so1, so2, so3)
	}
}

func TestActLevelShopsAndItems(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "ShopChar", "Human", "Fighter")
	seedAdventure(t, 1, 1, "Shop Adventure")

	// Seed act
	_, err := db.DB.Exec(
		"INSERT INTO oneshot_acts(id, adventure_id, number, title, sort_order) VALUES(?,?,?,?,?)",
		1, 1, 1, "Shop Act", 1,
	)
	if err != nil {
		t.Fatalf("seed act: %v", err)
	}

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/oneshot-adventures/:id/shops", CreateOneShotShop)
		auth.POST("/oneshot-adventures/:id/items", CreateOneShotItem)
		auth.GET("/oneshot-adventures/:id", GetOneShotAdventure)
	})

	// Create act-level shop
	w := testutil.PostJSON(t, r, "/api/oneshot-adventures/1/shops", map[string]any{
		"name": "Act Shop", "description": "A shop for the act",
		"act_id": 1,
	})
	testutil.AssertStatus(t, w, 201)
	var shop map[string]any
	testutil.ParseJSON(t, w, &shop)
	if shop["act_id"] == nil {
		t.Fatal("expected shop.act_id to be set")
	}
	actID := int(shop["act_id"].(float64))
	if actID != 1 {
		t.Fatalf("expected shop act_id=1, got %d", actID)
	}

	// Create act-level item
	w = testutil.PostJSON(t, r, "/api/oneshot-adventures/1/items", map[string]any{
		"name": "Act Item", "description": "An item for the act",
		"category": "weapon", "quantity": 1, "act_id": 1,
	})
	testutil.AssertStatus(t, w, 201)
	var item map[string]any
	testutil.ParseJSON(t, w, &item)
	if item["act_id"] == nil {
		t.Fatalf("expected item.act_id to be set, got response: %+v", item)
	}
	itemActID := int(item["act_id"].(float64))
	if itemActID != 1 {
		t.Fatalf("expected item act_id=1, got %d", itemActID)
	}

	// Get adventure and verify item appears in act's items edge
	w = testutil.Get(t, r, "/api/oneshot-adventures/1")
	testutil.AssertStatus(t, w, 200)
	var adv map[string]any
	testutil.ParseJSON(t, w, &adv)
	acts := adv["acts"].([]any)
	if len(acts) == 0 {
		t.Fatal("expected at least 1 act")
	}
	firstAct := acts[0].(map[string]any)
	actItems := firstAct["items"].([]any)
	if len(actItems) == 0 {
		t.Fatalf("expected at least 1 item in act, got 0")
	}
	firstItem := actItems[0].(map[string]any)
	if firstItem["name"] != "Act Item" {
		t.Fatalf("expected item name 'Act Item', got %v", firstItem["name"])
	}
}

func TestActCRUD(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "CRUDChar", "Human", "Fighter")
	seedAdventure(t, 1, 1, "CRUD Adventure")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/oneshot-adventures/:id/acts", CreateOneShotAct)
		auth.PUT("/oneshot-acts/:id", UpdateOneShotAct)
		auth.DELETE("/oneshot-acts/:id", DeleteOneShotAct)
		auth.GET("/oneshot-adventures/:id", GetOneShotAdventure)
		auth.POST("/oneshot-acts/:id/scenes", CreateOneShotScene)
		auth.DELETE("/oneshot-scenes/:id", DeleteOneShotScene)
	})

	// Create act
	w := testutil.PostJSON(t, r, "/api/oneshot-adventures/1/acts", map[string]any{
		"title": "CRUD Act", "description": "For CRUD testing",
		"number": 1, "sort_order": 1,
	})
	testutil.AssertStatus(t, w, 201)
	var actResp map[string]any
	testutil.ParseJSON(t, w, &actResp)
	actID := int64(actResp["id"].(float64))

	// Update act
	w = testutil.PutJSON(t, r, fmt.Sprintf("/api/oneshot-acts/%d", actID), map[string]any{
		"title": "Updated Act", "description": "Updated",
		"number": 1, "sort_order": 2,
	})
	testutil.AssertStatus(t, w, 200)

	// Create scene in act
	w = testutil.PostJSON(t, r, fmt.Sprintf("/api/oneshot-acts/%d/scenes", actID), map[string]any{
		"title": "Test Scene", "description": "A scene",
		"number": 1, "sort_order": 1, "scene_type": "roleplay",
	})
	testutil.AssertStatus(t, w, 201)
	var scResp map[string]any
	testutil.ParseJSON(t, w, &scResp)
	sceneID := int64(scResp["id"].(float64))

	// Delete scene
	w = testutil.Delete(t, r, fmt.Sprintf("/api/oneshot-scenes/%d", sceneID))
	testutil.AssertStatus(t, w, 200)

	// Delete act
	w = testutil.Delete(t, r, fmt.Sprintf("/api/oneshot-acts/%d", actID))
	testutil.AssertStatus(t, w, 200)
}

func TestActTreeParentUpdateViaAPI(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "ParentTestChar", "Human", "Fighter")
	seedAdventure(t, 1, 1, "Parent Update Adventure")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/oneshot-adventures/:id/acts", CreateOneShotAct)
		auth.GET("/oneshot-adventures/:id", GetOneShotAdventure)
		auth.PUT("/oneshot-acts/:id", UpdateOneShotAct)
	})

	// Create root act
	w := testutil.PostJSON(t, r, "/api/oneshot-adventures/1/acts", map[string]any{
		"title": "Root", "description": "Root act",
		"number": 1, "sort_order": 1,
	})
	testutil.AssertStatus(t, w, 201)
	var rootResp map[string]any
	testutil.ParseJSON(t, w, &rootResp)
	rootID := int64(rootResp["id"].(float64))

	// Create child1
	w = testutil.PostJSON(t, r, "/api/oneshot-adventures/1/acts", map[string]any{
		"title": "Child1", "description": "First child",
		"parent_act_id": rootID, "number": 1, "sort_order": 1,
	})
	testutil.AssertStatus(t, w, 201)
	var child1Resp map[string]any
	testutil.ParseJSON(t, w, &child1Resp)
	child1ID := int64(child1Resp["id"].(float64))

	// Create child2
	w = testutil.PostJSON(t, r, "/api/oneshot-adventures/1/acts", map[string]any{
		"title": "Child2", "description": "Second child",
		"parent_act_id": rootID, "number": 2, "sort_order": 2,
	})
	testutil.AssertStatus(t, w, 201)
	var child2Resp map[string]any
	testutil.ParseJSON(t, w, &child2Resp)
	child2ID := int64(child2Resp["id"].(float64))

	// Move child1 to be under child2
	w = testutil.PutJSON(t, r, fmt.Sprintf("/api/oneshot-acts/%d", child1ID), map[string]any{
		"title": "Child1", "description": "Moved",
		"parent_act_id": child2ID, "number": 1, "sort_order": 1,
	})
	testutil.AssertStatus(t, w, 200)

	// Verify tree: Root → Child2 → Child1
	w = testutil.Get(t, r, "/api/oneshot-adventures/1")
	testutil.AssertStatus(t, w, 200)
	var adv map[string]any
	testutil.ParseJSON(t, w, &adv)
	acts := adv["acts"].([]any)
	if len(acts) != 1 {
		t.Fatalf("expected 1 root act, got %d", len(acts))
	}
	rootChildren := acts[0].(map[string]any)["children"].([]any)
	if len(rootChildren) != 1 {
		t.Fatalf("expected 1 child of root, got %d", len(rootChildren))
	}
	if rootChildren[0].(map[string]any)["title"] != "Child2" {
		t.Fatalf("expected Child2 under root, got %v", rootChildren[0].(map[string]any)["title"])
	}
	grandChildren := rootChildren[0].(map[string]any)["children"].([]any)
	if len(grandChildren) != 1 {
		t.Fatalf("expected 1 grandchild, got %d", len(grandChildren))
	}
	if grandChildren[0].(map[string]any)["title"] != "Child1" {
		t.Fatalf("expected Child1 under Child2, got %v", grandChildren[0].(map[string]any)["title"])
	}
}

func TestHtmxEditActForm(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "EditActChar", "Human", "Fighter")
	seedAdventure(t, 1, 1, "Edit Act Adventure")

	_, err := db.DB.Exec(
		"INSERT INTO oneshot_acts(id, adventure_id, number, title, description, sort_order) VALUES(?,?,?,?,?,?)",
		1, 1, 1, "Editable Act", "An act to edit", 1,
	)
	if err != nil {
		t.Fatalf("seed act: %v", err)
	}

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/htmx/oneshot-acts/:id/edit", HtmxEditActForm)
		auth.GET("/htmx/oneshot-adventures/:id", HtmxGetOneShotDetail)
		auth.GET("/htmx/oneshot-adventures/:id/new-act-form", HtmxNewActForm)
	})

	w := testutil.Get(t, r, "/api/htmx/oneshot-acts/1/edit")
	testutil.AssertStatus(t, w, 200)
	body := w.Body.String()
	if !contains(body, "Editable Act") {
		t.Fatal("expected edit form to contain 'Editable Act'")
	}
	if !contains(body, "Update Act") {
		t.Fatal("expected edit form to have 'Update Act' button")
	}
}

func TestHtmxEditSceneForm(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "EditSceneChar", "Human", "Fighter")
	seedAdventure(t, 1, 1, "Edit Scene Adventure")

	_, err := db.DB.Exec(
		"INSERT INTO oneshot_acts(id, adventure_id, number, title, sort_order) VALUES(?,?,?,?,?)",
		1, 1, 1, "Act With Scene", 1,
	)
	if err != nil {
		t.Fatalf("seed act: %v", err)
	}
	_, err = db.DB.Exec(
		"INSERT INTO oneshot_scenes(id, act_id, number, title, scene_type, sort_order) VALUES(?,?,?,?,?,?)",
		1, 1, 1, "Editable Scene", "roleplay", 1,
	)
	if err != nil {
		t.Fatalf("seed scene: %v", err)
	}

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/htmx/oneshot-scenes/:id/edit", HtmxEditSceneForm)
	})

	w := testutil.Get(t, r, "/api/htmx/oneshot-scenes/1/edit")
	testutil.AssertStatus(t, w, 200)
	body := w.Body.String()
	if !contains(body, "Editable Scene") {
		t.Fatal("expected edit form to contain 'Editable Scene'")
	}
	if !contains(body, "Update Scene") {
		t.Fatal("expected edit form to have 'Update Scene' button")
	}
}

func TestDialogCRUD(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "DialogChar", "Human", "Fighter")
	seedAdventure(t, 1, 1, "Dialog Adventure")

	_, err := db.DB.Exec(
		"INSERT INTO oneshot_acts(id, adventure_id, number, title, sort_order) VALUES(?,?,?,?,?)",
		1, 1, 1, "Dialog Act", 1,
	)
	if err != nil {
		t.Fatalf("seed act: %v", err)
	}
	_, err = db.DB.Exec(
		"INSERT INTO oneshot_scenes(id, act_id, number, title, scene_type, sort_order) VALUES(?,?,?,?,?,?)",
		1, 1, 1, "Dialog Scene", "roleplay", 1,
	)
	if err != nil {
		t.Fatalf("seed scene: %v", err)
	}

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/htmx/oneshot-scenes/:id/dialogs", HtmxDialogList)
		auth.POST("/htmx/oneshot-scenes/:id/dialogs", HtmxCreateDialog)
		auth.PUT("/htmx/oneshot-scene-dialogs/:id", HtmxUpdateDialog)
		auth.DELETE("/htmx/oneshot-scene-dialogs/:id", HtmxDeleteDialog)
		auth.PUT("/oneshot-scenes/:id/dialogs/reorder", ReorderDialogs)
	})

	// Create dialog via form POST
	formBody := "speaker=Guard&dialog_text=Halt!+Who+goes+there?"
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/htmx/oneshot-scenes/1/dialogs", bytes.NewReader([]byte(formBody)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.ServeHTTP(w, req)
	testutil.AssertStatus(t, w, 200)
	body := w.Body.String()
	if !contains(body, "Guard") || !contains(body, "Halt!") {
		t.Fatal("expected dialog list to contain created dialog")
	}

	// Get dialog ID from DB
	var dialogID int64
	db.DB.QueryRow("SELECT id FROM oneshot_scene_dialogs WHERE scene_id=1").Scan(&dialogID)
	if dialogID == 0 {
		t.Fatal("expected dialog to exist in DB")
	}

	// Update dialog via form PUT
	formBody = "speaker=Captain&dialog_text=Password%3F"
	w = httptest.NewRecorder()
	req = httptest.NewRequest("PUT", fmt.Sprintf("/api/htmx/oneshot-scene-dialogs/%d", dialogID), bytes.NewReader([]byte(formBody)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.ServeHTTP(w, req)
	testutil.AssertStatus(t, w, 200)
	body = w.Body.String()
	if !contains(body, "Captain") || !contains(body, "Password?") {
		t.Fatal("expected updated dialog content")
	}

	// Delete dialog
	w = testutil.Delete(t, r, fmt.Sprintf("/api/htmx/oneshot-scene-dialogs/%d", dialogID))
	testutil.AssertStatus(t, w, 200)

	// Verify deleted
	var count int
	db.DB.QueryRow("SELECT COUNT(*) FROM oneshot_scene_dialogs WHERE id=?", dialogID).Scan(&count)
	if count != 0 {
		t.Fatal("expected dialog to be deleted")
	}
}

func TestDialogReorder(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "DialogReorderChar", "Human", "Fighter")
	seedAdventure(t, 1, 1, "Dialog Reorder Adventure")

	_, err := db.DB.Exec(
		"INSERT INTO oneshot_acts(id, adventure_id, number, title, sort_order) VALUES(?,?,?,?,?)",
		1, 1, 1, "Reorder Act", 1,
	)
	if err != nil {
		t.Fatalf("seed act: %v", err)
	}
	_, err = db.DB.Exec(
		"INSERT INTO oneshot_scenes(id, act_id, number, title, scene_type, sort_order) VALUES(?,?,?,?,?,?)",
		1, 1, 1, "Reorder Scene", "roleplay", 1,
	)
	if err != nil {
		t.Fatalf("seed scene: %v", err)
	}

	// Seed 3 dialogs
	for i := 1; i <= 3; i++ {
		_, err := db.DB.Exec(
			"INSERT INTO oneshot_scene_dialogs(id, scene_id, sort_order, speaker, dialog_text) VALUES(?,?,?,?,?)",
			i, 1, i, fmt.Sprintf("Speaker %d", i), fmt.Sprintf("Text %d", i),
		)
		if err != nil {
			t.Fatalf("seed dialog %d: %v", i, err)
		}
	}

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.PUT("/oneshot-scenes/:id/dialogs/reorder", ReorderDialogs)
	})

	// Reorder: 3, 2, 1
	w := testutil.PutJSON(t, r, "/api/oneshot-scenes/1/dialogs/reorder", map[string]any{
		"order": []int64{3, 2, 1},
	})
	testutil.AssertStatus(t, w, 200)

	// Verify sort_order
	var so1, so2, so3 int
	db.DB.QueryRow("SELECT sort_order FROM oneshot_scene_dialogs WHERE id=1").Scan(&so1)
	db.DB.QueryRow("SELECT sort_order FROM oneshot_scene_dialogs WHERE id=2").Scan(&so2)
	db.DB.QueryRow("SELECT sort_order FROM oneshot_scene_dialogs WHERE id=3").Scan(&so3)
	if so1 != 3 || so2 != 2 || so3 != 1 {
		t.Fatalf("expected sort_order 3,2,1 got %d,%d,%d", so1, so2, so3)
	}
}
