package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
	"villum/middleware"
)

// TestCompendiumPickerTemplatesAreModalBodyContent guards against a regression
// where the compendium spell and equipment pickers wrapped their content in a
// full Bootstrap modal markup. The buttons that load them target the page's
// shared #genericModalBody and call genericModal.show(), so the rendered
// output must be modal body content only — not another nested modal —
// otherwise the modal appears empty.
//
// See: handlers/templates/compendium_spell_picker.html
//      handlers/templates/compendium_equipment_picker.html
func TestCompendiumPickerTemplatesAreModalBodyContent(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "Farmes", "Elf", "Barbarian")
	testutil.SeedCompendiumSpell(t, 1001, "Acid Splash")
	testutil.SeedCompendiumEquipment(t, 2001, "Backpack")

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

	cases := []struct {
		name        string
		path        string
		mustContain []string
		mustNotHave []string
	}{
		{
			name: "spells picker renders body content only",
			path: "/htmx/compendium/spells/picker?character_id=1&q=acid",
			mustContain: []string{
				`id="compendiumSpellSearch"`,
				`id="compendiumSpellResults"`,
				`hx-post="/api/characters/1/spells/link"`,
				`Acid Splash`,
				// The + link button must hide the *generic* modal that's
				// actually shown, not a phantom nested modal.
				`getElementById('genericModal')`,
			},
			mustNotHave: []string{
				// Regression markers — a nested modal would render inert.
				`<div class="modal fade" id="compendiumSpellPickerModal"`,
				`modal-dialog modal-lg modal-dialog-scrollable`,
			},
		},
		{
			name: "spells picker empty state shows prompt",
			path: "/htmx/compendium/spells/picker?character_id=1",
			mustContain: []string{
				`Type a spell name to search`,
			},
			mustNotHave: []string{
				`<div class="modal fade" id="compendiumSpellPickerModal"`,
			},
		},
		{
			name: "equipment picker renders body content only",
			path: "/htmx/compendium/equipment/picker?character_id=1&q=back",
			mustContain: []string{
				`id="compendiumEquipmentSearch"`,
				`id="compendiumEquipmentResults"`,
				`hx-post="/api/characters/1/inventory/link"`,
				`Backpack`,
				`getElementById('genericModal')`,
			},
			mustNotHave: []string{
				`<div class="modal fade" id="compendiumEquipmentPickerModal"`,
				`modal-dialog modal-lg modal-dialog-scrollable`,
			},
		},
		{
			name: "equipment picker empty state shows prompt",
			path: "/htmx/compendium/equipment/picker?character_id=1",
			mustContain: []string{
				`Type an item name to search`,
			},
			mustNotHave: []string{
				`<div class="modal fade" id="compendiumEquipmentPickerModal"`,
			},
		},
		{
			name: "monster picker for one-shot is search-first (no auto-load)",
			path: "/htmx/compendium-monsters/oneshot/1",
			mustContain: []string{
				`Type a monster name to search`,
				`id="compendiumResults"`,
				`name="adventure_id" value="1"`,
			},
			mustNotHave: []string{
				// Picker context must NOT auto-load all monsters
				`hx-trigger="load"`,
			},
		},
		{
			name: "monster picker for encounter is search-first (no auto-load)",
			path: "/htmx/compendium-monsters/picker/1",
			mustContain: []string{
				`Type a monster name to search`,
				`id="compendiumResults"`,
				`name="encounter_id" value="1"`,
			},
			mustNotHave: []string{
				`hx-trigger="load"`,
			},
		},
		{
			name: "monster compendium tab auto-loads all monsters",
			path: "/htmx/compendium-monsters",
			mustContain: []string{
				`hx-trigger="load"`,
				`Loading monsters...`,
				`id="compendiumResults"`,
			},
			mustNotHave: []string{
				// Compendium tab should NOT show the search-first prompt
				`Type a monster name to search`,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tc.path, nil)
			r.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
			}
			body := w.Body.String()
			for _, want := range tc.mustContain {
				if !strings.Contains(body, want) {
					t.Errorf("body missing %q\n--- body ---\n%s", want, body)
				}
			}
			for _, banned := range tc.mustNotHave {
				if strings.Contains(body, banned) {
					t.Errorf("body unexpectedly contains %q (regression: nested modal wrapper)\n--- body ---\n%s", banned, body)
				}
			}
		})
	}
}
