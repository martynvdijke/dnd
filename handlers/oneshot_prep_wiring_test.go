package handlers

import (
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers/testutil"
)

func TestActNotesSaveDoesNotClobberAct(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedOneShot(t, 1, 1, "Notes Adventure")
	testutil.SeedOneShotAct(t, 1, 1, "Original Title", 1)
	db.DB.Exec("UPDATE oneshot_acts SET description='Original description', estimated_minutes=45, notes='old notes' WHERE id=1")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.PUT("/htmx/oneshot-acts/:id/notes", HtmxUpdateActNotes)
		auth.PUT("/htmx/oneshot-acts/:id", HtmxUpdateAct)
	})

	t.Run("notes-only save keeps the other act fields", func(t *testing.T) {
		w := putForm(t, r, "/api/htmx/oneshot-acts/1/notes", map[string]string{"notes": "fresh notes"})
		testutil.AssertStatus(t, w, 200)

		var title, description, notes string
		var minutes int
		if err := db.DB.QueryRow("SELECT title, description, estimated_minutes, notes FROM oneshot_acts WHERE id=1").
			Scan(&title, &description, &minutes, &notes); err != nil {
			t.Fatalf("load act: %v", err)
		}
		if title != "Original Title" || description != "Original description" || minutes != 45 {
			t.Fatalf("notes-only save clobbered fields: title=%q description=%q minutes=%d", title, description, minutes)
		}
		if notes != "fresh notes" {
			t.Fatalf("expected notes updated, got %q", notes)
		}
	})

	t.Run("full act form still updates act fields", func(t *testing.T) {
		w := putForm(t, r, "/api/htmx/oneshot-acts/1", map[string]string{
			"title":             "Updated Title",
			"description":       "Updated description",
			"estimated_minutes": "30",
			"number":            "1",
			"sort_order":        "1",
			"notes":             "updated notes",
		})
		testutil.AssertStatus(t, w, 200)

		var title, description, notes string
		var minutes int
		if err := db.DB.QueryRow("SELECT title, description, estimated_minutes, notes FROM oneshot_acts WHERE id=1").
			Scan(&title, &description, &minutes, &notes); err != nil {
			t.Fatalf("load act: %v", err)
		}
		if title != "Updated Title" || description != "Updated description" || minutes != 30 || notes != "updated notes" {
			t.Fatalf("act form did not persist: title=%q description=%q minutes=%d notes=%q", title, description, minutes, notes)
		}
	})
}

func TestAdventurePacingLookup(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedOneShot(t, 1, 1, "Pacing Adventure")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/oneshot-adventures/:id/pacing", GetAdventurePacing)
	})

	t.Run("no session returns 404", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/oneshot-adventures/1/pacing")
		testutil.AssertStatus(t, w, 404)
	})

	t.Run("running session is returned over completed", func(t *testing.T) {
		db.DB.Exec("INSERT INTO session_pacing(adventure_id, status, elapsed_seconds, started_at, completed_at) VALUES(1,'completed',10,datetime('now','-1 hour'),datetime('now','-30 minutes'))")
		db.DB.Exec("INSERT INTO session_pacing(adventure_id, status, elapsed_seconds, started_at) VALUES(1,'running',5,datetime('now','-10 seconds'))")

		w := testutil.Get(t, r, "/api/oneshot-adventures/1/pacing")
		testutil.AssertStatus(t, w, 200)
		var result map[string]any
		testutil.ParseJSON(t, w, &result)
		if result["status"] != "running" {
			t.Fatalf("expected running session, got %v", result["status"])
		}
		elapsed, _ := result["elapsed_seconds"].(float64)
		if elapsed < 5 {
			t.Fatalf("expected elapsed >= 5, got %v", elapsed)
		}
	})

	t.Run("completed session still returns 200", func(t *testing.T) {
		db.DB.Exec("UPDATE session_pacing SET status='completed', completed_at=datetime('now') WHERE status='running'")

		w := testutil.Get(t, r, "/api/oneshot-adventures/1/pacing")
		testutil.AssertStatus(t, w, 200)
		var result map[string]any
		testutil.ParseJSON(t, w, &result)
		if result["status"] != "completed" {
			t.Fatalf("expected completed session, got %v", result["status"])
		}
	})
}

func TestHTMXClueRedHerring(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedOneShot(t, 1, 1, "Clue Adventure")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/htmx/oneshot-adventures/:id/clues", HtmxCreateClue)
		auth.PUT("/htmx/clues/:id", HtmxUpdateClue)
	})

	var clueID int64

	t.Run("create marks red herring", func(t *testing.T) {
		w := testutil.PostForm(t, r, "/api/htmx/oneshot-adventures/1/clues", map[string]string{
			"title":          "False Lead",
			"description":    "Looks important",
			"clue_type":      "object",
			"is_red_herring": "1",
			"sort_order":     "1",
		})
		testutil.AssertStatus(t, w, 200)

		var dummy int
		if err := db.DB.QueryRow("SELECT id, is_red_herring FROM clues WHERE adventure_id=1 ORDER BY id DESC LIMIT 1").Scan(&clueID, &dummy); err != nil {
			t.Fatalf("load clue: %v", err)
		}
		var redHerring int
		db.DB.QueryRow("SELECT is_red_herring FROM clues WHERE id=?", clueID).Scan(&redHerring)
		if redHerring != 1 {
			t.Fatalf("expected is_red_herring=1, got %d", redHerring)
		}
	})

	t.Run("update clears red herring", func(t *testing.T) {
		if clueID == 0 {
			t.Skip("no clue")
		}
		w := putForm(t, r, "/api/htmx/clues/"+strconv.FormatInt(clueID, 10), map[string]string{
			"title":       "False Lead",
			"description": "Looks important",
			"clue_type":   "object",
		})
		testutil.AssertStatus(t, w, 200)

		var redHerring int
		db.DB.QueryRow("SELECT is_red_herring FROM clues WHERE id=?", clueID).Scan(&redHerring)
		if redHerring != 0 {
			t.Fatalf("expected is_red_herring=0 after update, got %d", redHerring)
		}
	})
}

func TestCompendiumEquipmentImportCreatesOneShotItems(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedOneShot(t, 1, 1, "Loot Adventure")
	testutil.SeedOneShotAct(t, 1, 1, "Act One", 1)
	testutil.SeedCompendiumEquipment(t, 9001, "Rope")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/oneshot-adventures/:id/import/compendium-equipment", ImportCompendiumEquipmentToOneShot)
	})

	t.Run("import at adventure level", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/oneshot-adventures/1/import/compendium-equipment", map[string]any{
			"compendium_equipment_id": 9001,
			"adventure_id":            1,
			"quantity":                2,
		})
		testutil.AssertStatus(t, w, 201)

		var name string
		var quantity int
		if err := db.DB.QueryRow("SELECT name, quantity FROM oneshot_items WHERE adventure_id=1 ORDER BY id DESC LIMIT 1").Scan(&name, &quantity); err != nil {
			t.Fatalf("load item: %v", err)
		}
		if name != "Rope" || quantity != 2 {
			t.Fatalf("unexpected item: name=%q quantity=%d", name, quantity)
		}
	})

	t.Run("import into an act", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/oneshot-adventures/1/import/compendium-equipment", map[string]any{
			"compendium_equipment_id": 9001,
			"adventure_id":            1,
			"act_id":                  1,
			"quantity":                1,
		})
		testutil.AssertStatus(t, w, 201)

		var actID int64
		if err := db.DB.QueryRow("SELECT act_id FROM oneshot_items WHERE adventure_id=1 ORDER BY id DESC LIMIT 1").Scan(&actID); err != nil {
			t.Fatalf("load item: %v", err)
		}
		if actID != 1 {
			t.Fatalf("expected act_id=1, got %d", actID)
		}
	})
}

func TestUnlinkCompendiumMonsterRoute(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedOneShot(t, 1, 1, "Monster Adventure")
	db.DB.Exec("INSERT INTO oneshot_monsters(id, adventure_id, name, ac, hp, cr, source, compendium_monster_id) VALUES(1,1,'Goblin',15,7,'1/4','compendium',42)")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.DELETE("/oneshot-monsters/:id/link", UnlinkCompendiumMonster)
	})

	w := testutil.Delete(t, r, "/api/oneshot-monsters/1/link")
	testutil.AssertStatus(t, w, 200)

	var count int
	db.DB.QueryRow("SELECT COUNT(*) FROM oneshot_monsters WHERE id=1").Scan(&count)
	if count != 0 {
		t.Fatalf("expected monster removed, count=%d", count)
	}
}

func TestHTMXPregenCRUD(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/htmx/pregens", HtmxListPregens)
		auth.POST("/htmx/pregens", HtmxCreatePregen)
		auth.PUT("/htmx/pregens/:id", HtmxUpdatePregen)
	})

	var pregenID int64

	t.Run("create pregen from form", func(t *testing.T) {
		w := testutil.PostForm(t, r, "/api/htmx/pregens", map[string]string{
			"name":  "Testy McTest",
			"race":  "Elf",
			"class": "Rogue",
			"level": "2",
			"str":   "12",
			"dex":   "17",
			"con":   "13",
			"int":   "14",
			"wis":   "10",
			"cha":   "11",
			"hp":    "15",
			"ac":    "14",
			"speed": "30",
		})
		testutil.AssertStatus(t, w, 200)

		var name, class string
		var level int
		if err := db.DB.QueryRow("SELECT id, name, class, level FROM pregen_characters WHERE user_id=1 ORDER BY id DESC LIMIT 1").Scan(&pregenID, &name, &class, &level); err != nil {
			t.Fatalf("load pregen: %v", err)
		}
		if name != "Testy McTest" || class != "Rogue" || level != 2 {
			t.Fatalf("unexpected pregen: name=%q class=%q level=%d", name, class, level)
		}
	})

	t.Run("update pregen from form", func(t *testing.T) {
		if pregenID == 0 {
			t.Skip("no pregen")
		}
		w := putForm(t, r, "/api/htmx/pregens/"+strconv.FormatInt(pregenID, 10), map[string]string{
			"name":  "Testy Updated",
			"race":  "Elf",
			"class": "Rogue",
			"level": "3",
			"str":   "12",
			"dex":   "18",
			"con":   "13",
			"int":   "14",
			"wis":   "10",
			"cha":   "11",
			"hp":    "21",
			"ac":    "15",
			"speed": "30",
		})
		testutil.AssertStatus(t, w, 200)

		var name string
		var level, dex int
		if err := db.DB.QueryRow("SELECT name, level, dex FROM pregen_characters WHERE id=?", pregenID).Scan(&name, &level, &dex); err != nil {
			t.Fatalf("reload pregen: %v", err)
		}
		if name != "Testy Updated" || level != 3 || dex != 18 {
			t.Fatalf("update not persisted: name=%q level=%d dex=%d", name, level, dex)
		}
	})

	t.Run("list pregens returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/htmx/pregens")
		testutil.AssertStatus(t, w, 200)
	})
}

func TestHTMXPacingControlsRenderDashboard(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedOneShot(t, 1, 1, "Pacing Wiring")
	testutil.SeedOneShotAct(t, 1, 1, "Act One", 1)
	testutil.SeedOneShotScene(t, 1, 1, "Scene One", 1)

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/oneshot-adventures/:id/pacing/start", StartPacingSession)
		auth.GET("/htmx/session-pacing/:id", HtmxGetPacingDashboard)
		auth.POST("/htmx/session-pacing/:id/pause", HtmxPausePacingSession)
		auth.POST("/htmx/session-pacing/:id/resume", HtmxResumePacingSession)
		auth.POST("/htmx/session-pacing/:id/next-scene", HtmxAdvancePacingSession)
		auth.POST("/htmx/session-pacing/:id/complete", HtmxCompletePacingSession)
	})

	w := testutil.PostJSON(t, r, "/api/oneshot-adventures/1/pacing/start", map[string]any{})
	testutil.AssertStatus(t, w, 201)
	var res map[string]any
	testutil.ParseJSON(t, w, &res)
	sessionID := strconv.FormatInt(int64(res["id"].(float64)), 10)

	for _, action := range []string{"pause", "resume", "next-scene", "complete"} {
		t.Run(action+" renders the pacing dashboard", func(t *testing.T) {
			w := testutil.PostForm(t, r, "/api/htmx/session-pacing/"+sessionID+"/"+action, nil)
			testutil.AssertStatus(t, w, 200)
			body := w.Body.String()
			if !strings.Contains(body, `id="pacingDashboard"`) {
				if len(body) > 200 {
					body = body[:200]
				}
				t.Fatalf("expected pacing dashboard fragment, got: %s", body)
			}
		})
	}
}
