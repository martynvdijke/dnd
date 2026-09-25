package handlers

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers/testutil"
)

func TestSpellEffectFor(t *testing.T) {
	cases := []struct {
		name, school, want string
	}{
		{"Fireball", "Evocation", "fire"},
		{"Burning Hands", "Evocation", "fire"},
		{"Call Lightning", "Conjuration", "lightning"},
		{"Cone of Cold", "Evocation", "snow"},
		{"Ice Storm", "Evocation", "snow"},
		{"Cloudkill", "Conjuration", "smoke"},
		{"Blight", "Necromancy", "darkness"},
		{"Cure Wounds", "Evocation", "sparkle"},
		{"Sacred Flame", "Evocation", "fire"},
		{"Weird", "Illusion", "sparkle"},
		{"Unknown Spell", "Conjuration", "smoke"},
		{"Mystery", "Necromancy", "darkness"},
		{"Mystery", "Divination", "sparkle"},
	}
	for _, tc := range cases {
		if got := SpellEffectFor(tc.name, tc.school); got != tc.want {
			t.Errorf("SpellEffectFor(%q,%q)=%q want %q", tc.name, tc.school, got, tc.want)
		}
	}
}

func TestWLEDSettingsRoundtrip(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/wled-settings", GetWLEDSettings)
		auth.POST("/wled-settings", SaveWLEDSettings)
	})

	var got map[string]any
	testutil.ParseJSON(t, testutil.Get(t, r, "/api/wled-settings"), &got)
	if got["enabled"] != false || got["configured"] != false {
		t.Fatalf("unexpected defaults: %v", got)
	}

	w := testutil.PostJSON(t, r, "/api/wled-settings", map[string]any{
		"enabled": true,
		"devices": []map[string]any{
			{"name": "Table", "base_url": "wled.local/", "brightness": 900, "restore_preset": 3, "enabled": true},
			{"name": "Ignored", "base_url": "", "enabled": true},
		},
	})
	testutil.AssertStatus(t, w, http.StatusOK)

	var after map[string]any
	testutil.ParseJSON(t, testutil.Get(t, r, "/api/wled-settings"), &after)
	if after["enabled"] != true {
		t.Errorf("enabled not persisted: %v", after)
	}
	devs, _ := after["devices"].([]any)
	if len(devs) != 1 {
		t.Fatalf("expected 1 stored device, got %v", after["devices"])
	}
	d, _ := devs[0].(map[string]any)
	if d["base_url"] != "http://wled.local" {
		t.Errorf("base_url normalize failed: %v", d["base_url"])
	}
	if d["brightness"] != float64(255) {
		t.Errorf("brightness clamp failed: %v", d["brightness"])
	}
	if d["restore_preset"] != float64(3) || d["name"] != "Table" {
		t.Errorf("device fields not persisted: %v", d)
	}
}

func TestWLEDLegacyFallback(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)

	if err := setAppSetting(settingWLEDBaseURL, "legacy.local"); err != nil {
		t.Fatalf("seed legacy key: %v", err)
	}
	devs := loadWLEDDevices()
	if len(devs) != 1 {
		t.Fatalf("expected legacy fallback device, got %v", devs)
	}
	if devs[0].BaseURL != "http://legacy.local" || !devs[0].Enabled {
		t.Fatalf("unexpected legacy device: %+v", devs[0])
	}
}

func TestCastSpell(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)

	testutil.SeedUser(t, 1, "caster", "user")
	testutil.SeedUser(t, 2, "stranger", "user")
	testutil.SeedCampaign(t, 50, "CastCamp", "Testers", 1)
	testutil.SeedCharacterInCampaign(t, 10, 1, 50, "Mage", "Elf", "Wizard")
	if _, err := db.DB.Exec(`INSERT INTO character_spellcasting(character_id, ability, save_dc, attack_bonus, slots_3_max, slots_3_used) VALUES(10,'int',13,5,2,0)`); err != nil {
		t.Fatalf("seed spellcasting: %v", err)
	}
	if _, err := db.DB.Exec(`INSERT INTO spells(id, character_id, name, level, school, prepared) VALUES(7,10,'Fireball',3,'Evocation',1)`); err != nil {
		t.Fatalf("seed spell: %v", err)
	}

	routes := func(auth *gin.RouterGroup) {
		auth.POST("/characters/:id/cast-spell", CastSpell)
	}

	r := testutil.NewRouterWithUser(routes, 1, "user")
	w := testutil.PostJSON(t, r, "/api/characters/10/cast-spell", map[string]any{"spell_id": 7})
	var res map[string]any
	testutil.ParseJSON(t, w, &res)
	if res["effect"] != "fire" {
		t.Fatalf("expected fire effect, got %v", res)
	}
	if res["slot_consumed"] != true {
		t.Fatalf("expected slot consumed, got %v", res)
	}
	var used int
	db.DB.QueryRow("SELECT slots_3_used FROM character_spellcasting WHERE character_id=10").Scan(&used)
	if used != 1 {
		t.Fatalf("slot used = %d, want 1", used)
	}

	// Unknown spell id is a 404.
	w = testutil.PostJSON(t, r, "/api/characters/10/cast-spell", map[string]any{"spell_id": 999})
	testutil.AssertStatus(t, w, http.StatusNotFound)

	// Non-owner cannot cast.
	r2 := testutil.NewRouterWithUser(routes, 2, "user")
	w2 := testutil.PostJSON(t, r2, "/api/characters/10/cast-spell", map[string]any{"spell_id": 7})
	testutil.AssertStatus(t, w2, http.StatusForbidden)
}
