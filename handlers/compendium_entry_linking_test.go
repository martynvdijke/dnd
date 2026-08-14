package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers/testutil"
	"villum/middleware"
)

// seedTestEntrySchema creates (or reuses) a generic compendium schema for
// linking tests and returns its id.
func seedTestEntrySchema(t *testing.T, typeName, displayName string) int64 {
	t.Helper()
	var id int64
	if err := db.DB.QueryRow("SELECT id FROM compendium_schemas WHERE type_name = ? LIMIT 1", typeName).Scan(&id); err == nil {
		return id
	}
	res, err := db.DB.Exec("INSERT INTO compendium_schemas (type_name, display_name, fields, created_at, updated_at) VALUES (?, ?, '{}', datetime('now'), datetime('now'))", typeName, displayName)
	if err != nil {
		t.Fatalf("insert schema: %v", err)
	}
	id, err = res.LastInsertId()
	if err != nil {
		t.Fatalf("lastinsertid: %v", err)
	}
	return id
}

// seedTestEntry inserts a generic compendium entry with a rich data payload.
func seedTestEntry(t *testing.T, id, schemaID int64, data string) {
	t.Helper()
	if _, err := db.DB.Exec("INSERT INTO compendium_entries (id, schema_id, data, created_at, updated_at) VALUES (?, ?, ?, datetime('now'), datetime('now'))", id, schemaID, data); err != nil {
		t.Fatalf("insert entry: %v", err)
	}
}

const testEntryData = `{"name":"Staff of Power","category":"Wondrous","cost":"50 gp","weight":3,"description":"A powerful staff.","level":5,"school":"Evocation","casting_time":"1 action","range":"60 feet","components":"V,S","duration":"1 minute","classes":"Wizard"}`

// newHtmxPickerRouter builds the same router used by the picker regression tests.
func newHtmxPickerRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.SecurityHeaders())
	g := r.Group("/")
	g.Use(func(c *gin.Context) {
		c.Set("user_id", int64(1))
		c.Set("username", "admin")
		c.Set("role", "admin")
		c.Set("session_id", "test-session")
		c.Next()
	})
	HtmxRegisterRoutes(g)
	return r
}

func getBody(t *testing.T, r *gin.Engine, path string) string {
	t.Helper()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", path, nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	return w.Body.String()
}

// Imported compendium entries (generic schema layer) must appear in the HTMX
// pickers alongside legacy SRD rows, tagged so the templates emit the right
// link field and a schema badge.
func TestPickersIncludeGenericEntries(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "Farmes", "Elf", "Barbarian")
	testutil.SeedCompendiumSpell(t, 1001, "Acid Splash")
	testutil.SeedCompendiumEquipment(t, 2001, "Backpack")
	schemaID := seedTestEntrySchema(t, "homebrew_items", "Homebrew Items")
	seedTestEntry(t, 9001, schemaID, testEntryData)

	r := newHtmxPickerRouter()

	t.Run("equipment picker shows entry with schema badge and entry link field", func(t *testing.T) {
		body := getBody(t, r, "/htmx/compendium/equipment/picker?character_id=1&q=staff")
		for _, want := range []string{
			`Staff of Power`,
			`hx-post="/htmx/compendium/equipment/link?character_id=1"`,
			`compendium_entry_id`,
			`Homebrew Items`,
		} {
			if !strings.Contains(body, want) {
				t.Errorf("body missing %q\n--- body ---\n%s", want, body)
			}
		}
	})

	t.Run("equipment picker mixes legacy and entry rows", func(t *testing.T) {
		body := getBody(t, r, "/htmx/compendium/equipment/picker?character_id=1&q=back")
		if !strings.Contains(body, "Backpack") {
			t.Errorf("legacy row missing\n--- body ---\n%s", body)
		}
		if !strings.Contains(body, `compendium_equipment_id`) {
			t.Errorf("legacy link field missing\n--- body ---\n%s", body)
		}
	})

	t.Run("spells picker shows entry with entry link field", func(t *testing.T) {
		body := getBody(t, r, "/htmx/compendium/spells/picker?character_id=1&q=staff")
		for _, want := range []string{
			`Staff of Power`,
			`Level 5`,
			`hx-post="/htmx/compendium/spells/link?character_id=1"`,
			`compendium_entry_id`,
			`Homebrew Items`,
		} {
			if !strings.Contains(body, want) {
				t.Errorf("body missing %q\n--- body ---\n%s", want, body)
			}
		}
	})

	t.Run("feature picker shows entries", func(t *testing.T) {
		body := getBody(t, r, "/htmx/compendium/features/picker?character_id=1&q=staff")
		for _, want := range []string{
			`id="compendiumFeatureSearch"`,
			`id="compendiumFeatureResults"`,
			`Staff of Power`,
			`Lv 5`,
			`hx-post="/htmx/compendium/features/link?character_id=1"`,
			`compendium_entry_id`,
			`Homebrew Items`,
		} {
			if !strings.Contains(body, want) {
				t.Errorf("body missing %q\n--- body ---\n%s", want, body)
			}
		}
	})

	t.Run("feature picker empty state shows prompt", func(t *testing.T) {
		body := getBody(t, r, "/htmx/compendium/features/picker?character_id=1")
		if !strings.Contains(body, "Type a feature name to search") {
			t.Errorf("empty prompt missing\n--- body ---\n%s", body)
		}
	})
}

// The character-sheet picker API must also surface generic entries so the
// client-rendered inventory picker can link them.
func TestListCompendiumEquipmentIncludesEntries(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCompendiumEquipment(t, 2001, "Backpack")
	schemaID := seedTestEntrySchema(t, "homebrew_items", "Homebrew Items")
	seedTestEntry(t, 9001, schemaID, testEntryData)

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/compendium/equipment", ListCompendiumEquipment)
	})
	w := testutil.Get(t, r, "/api/compendium/equipment?q=staff")
	testutil.AssertStatus(t, w, 200)
	body := w.Body.String()
	for _, want := range []string{
		`"name":"Staff of Power"`,
		`"source":"entry"`,
		`"schema_name":"Homebrew Items"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("body missing %q\n--- body ---\n%s", want, body)
		}
	}

	// Unfiltered listing mixes legacy rows and entries.
	w = testutil.Get(t, r, "/api/compendium/equipment")
	testutil.AssertStatus(t, w, 200)
	body = w.Body.String()
	if !strings.Contains(body, `"source":"equipment"`) || !strings.Contains(body, `"source":"entry"`) {
		t.Errorf("expected both source kinds\n--- body ---\n%s", body)
	}
}

// Linking a generic entry copies its JSON snapshot into the inventory.
func TestLinkCompendiumEntryToInventory(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "Hero", "Elf", "Wizard")
	schemaID := seedTestEntrySchema(t, "homebrew_items", "Homebrew Items")
	seedTestEntry(t, 9001, schemaID, testEntryData)

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/characters/:id/inventory/link", LinkCompendiumEquipment)
	})
	w := testutil.PostForm(t, r, "/api/characters/1/inventory/link", map[string]string{"compendium_entry_id": "9001"})
	testutil.AssertStatus(t, w, 201)

	var name, category string
	var weight float64
	var ref any
	err := db.DB.QueryRow(
		"SELECT name, category, weight, COALESCE(compendium_entry_id, 0) FROM inventory WHERE character_id=1 AND compendium_entry_id=9001",
	).Scan(&name, &category, &weight, &ref)
	if err != nil {
		t.Fatalf("linked item row not found: %v", err)
	}
	if name != "Staff of Power" || category != "Wondrous" || weight != 3 {
		t.Fatalf("snapshot not copied: name=%q category=%q weight=%v", name, category, weight)
	}
	if fmt.Sprint(ref) != "9001" {
		t.Fatalf("compendium_entry_id not stored: %v", ref)
	}
}

// Linking a generic entry copies its JSON snapshot into the spellbook.
func TestLinkCompendiumEntryToSpells(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "Hero", "Elf", "Wizard")
	schemaID := seedTestEntrySchema(t, "homebrew_items", "Homebrew Items")
	seedTestEntry(t, 9001, schemaID, testEntryData)

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/characters/:id/spells/link", LinkCompendiumSpell)
	})
	w := testutil.PostForm(t, r, "/api/characters/1/spells/link", map[string]string{"compendium_entry_id": "9001"})
	testutil.AssertStatus(t, w, 201)

	var name, school string
	var level int
	var ref any
	err := db.DB.QueryRow(
		"SELECT name, level, school, COALESCE(compendium_entry_id, 0) FROM spells WHERE character_id=1 AND compendium_entry_id=9001",
	).Scan(&name, &level, &school, &ref)
	if err != nil {
		t.Fatalf("linked spell row not found: %v", err)
	}
	if name != "Staff of Power" || level != 5 || school != "Evocation" {
		t.Fatalf("snapshot not copied: name=%q level=%d school=%q", name, level, school)
	}
	if fmt.Sprint(ref) != "9001" {
		t.Fatalf("compendium_entry_id not stored: %v", ref)
	}
}

// Unlinking an entry-linked item/spell preserves the snapshot but clears the ref.
func TestUnlinkCompendiumEntryPreservesData(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "Hero", "Elf", "Wizard")
	schemaID := seedTestEntrySchema(t, "homebrew_items", "Homebrew Items")
	seedTestEntry(t, 9001, schemaID, testEntryData)

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/characters/:id/inventory/link", LinkCompendiumEquipment)
		auth.DELETE("/characters/:id/inventory/:itemId/link", UnlinkCompendiumEquipment)
		auth.POST("/characters/:id/spells/link", LinkCompendiumSpell)
		auth.DELETE("/characters/:id/spells/:spellId/link", UnlinkCompendiumSpell)
	})

	testutil.PostForm(t, r, "/api/characters/1/inventory/link", map[string]string{"compendium_entry_id": "9001"})
	testutil.PostForm(t, r, "/api/characters/1/spells/link", map[string]string{"compendium_entry_id": "9001"})

	w := testutil.Delete(t, r, "/api/characters/1/inventory/1/link")
	testutil.AssertStatus(t, w, 200)
	var itemCount int
	var itemRef any
	db.DB.QueryRow("SELECT COUNT(*), COALESCE(compendium_entry_id, 0) FROM inventory WHERE id=1").Scan(&itemCount, &itemRef)
	if itemCount != 1 {
		t.Fatalf("item row deleted on unlink (want preserved): count=%d", itemCount)
	}
	if fmt.Sprint(itemRef) != "0" {
		t.Fatalf("item entry ref not cleared: %v", itemRef)
	}

	w = testutil.Delete(t, r, "/api/characters/1/spells/1/link")
	testutil.AssertStatus(t, w, 200)
	var spellCount int
	var spellRef any
	db.DB.QueryRow("SELECT COUNT(*), COALESCE(compendium_entry_id, 0) FROM spells WHERE id=1").Scan(&spellCount, &spellRef)
	if spellCount != 1 {
		t.Fatalf("spell row deleted on unlink (want preserved): count=%d", spellCount)
	}
	if fmt.Sprint(spellRef) != "0" {
		t.Fatalf("spell entry ref not cleared: %v", spellRef)
	}
}

// Feature linking via HTMX: picker → link → rendered list shows the snapshot,
// the Compendium badge and the unlink route; unlink keeps the feature data.
func TestHtmxFeatureLinkAndUnlink(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "Hero", "Elf", "Wizard")
	schemaID := seedTestEntrySchema(t, "homebrew_items", "Homebrew Items")
	seedTestEntry(t, 9001, schemaID, testEntryData)

	r := newHtmxPickerRouter()

	// link via the HTMX form
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/htmx/compendium/features/link?character_id=1", strings.NewReader("compendium_entry_id=9001&level_gained=3"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("link: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	for _, want := range []string{
		`Staff of Power`,
		`Lv 3`,
		`Compendium`,
		`hx-delete="/htmx/features/1/compendium-unlink?character_id=1"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("link render missing %q\n--- body ---\n%s", want, body)
		}
	}

	// unlink: data preserved, badge + unlink route gone
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/htmx/features/1/compendium-unlink?character_id=1", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("unlink: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	body = w.Body.String()
	if !strings.Contains(body, `Staff of Power`) {
		t.Errorf("feature data lost on unlink\n--- body ---\n%s", body)
	}
	if strings.Contains(body, `hx-delete="/htmx/features/1/compendium-unlink?character_id=1"`) {
		t.Errorf("unlink route still rendered after unlink\n--- body ---\n%s", body)
	}
}
