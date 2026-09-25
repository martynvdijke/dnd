package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"
	"villum/db"
	"villum/handlers/testutil"
)

type spellCastResult struct {
	OK           bool `json:"ok"`
	Spell        string
	Effect       string
	DamageRolled int `json:"damage_rolled"`
	Applied      []struct {
		Target     string `json:"target"`
		Type       string `json:"type"`
		ID         int64  `json:"id"`
		HPBefore   int    `json:"hp_before"`
		HPAfter    int    `json:"hp_after"`
		HPMax      int    `json:"hp_max"`
		Damage     int    `json:"damage"`
		Healing    int    `json:"healing"`
		SaveRolled bool   `json:"save_rolled"`
		Saved      bool   `json:"saved"`
	} `json:"applied"`
}

// Covers area spell application: multi-target damage, guaranteed-save halving,
// healing, and character targets.
func TestCastSpellTargets(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)

	testutil.SeedUser(t, 1, "caster", "admin")
	testutil.SeedCampaign(t, 50, "Camp", "Testers", 1)
	testutil.SeedCharacterInCampaign(t, 10, 1, 50, "Wizard", "Elf", "Wizard")
	testutil.SeedCharacterInCampaign(t, 20, 1, 50, "Goblin Boss", "Goblin", "Fighter")

	if _, err := db.DB.Exec("INSERT INTO character_spellcasting(character_id,ability,save_dc,attack_bonus,slots_3_max,slots_3_used) VALUES(10,'int',15,5,2,0)"); err != nil {
		t.Fatalf("seed spellcasting: %v", err)
	}
	res, err := db.DB.Exec("INSERT INTO spells(character_id,name,level,school) VALUES(10,'Fireball',3,'Evocation')")
	if err != nil {
		t.Fatalf("seed spell: %v", err)
	}
	spellID, _ := res.LastInsertId()

	res2, err := db.DB.Exec("INSERT INTO combat_entries(campaign_id,character_id,name,type,hp_max,hp_current) VALUES(50,20,'Goblin Boss','monster',100,100)")
	if err != nil {
		t.Fatalf("seed combat: %v", err)
	}
	entryID, _ := res2.LastInsertId()

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/characters/:id/cast-spell", CastSpell)
	})

	// 1. Damage with a save the target can only pass (DC 1) → halved.
	w := testutil.PostJSON(t, r, "/api/characters/10/cast-spell", map[string]any{
		"spell_id": spellID, "damage": "8d6", "damage_type": "fire",
		"save_ability": "dex", "save_dc": 1,
		"targets": []map[string]any{{"type": "combat", "id": entryID}},
	})
	testutil.AssertStatus(t, w, 200)
	var out spellCastResult
	testutil.ParseJSON(t, w, &out)
	if len(out.Applied) != 1 {
		t.Fatalf("applied = %d, want 1", len(out.Applied))
	}
	if a := out.Applied[0]; !a.SaveRolled || !a.Saved || a.Damage != out.DamageRolled/2 {
		t.Fatalf("save-halved damage wrong: %+v (rolled %d)", a, out.DamageRolled)
	}

	// 2. Healing raises the target back up.
	w = testutil.PostJSON(t, r, "/api/characters/10/cast-spell", map[string]any{
		"spell_id": spellID, "damage": "2d4", "healing": true,
		"targets": []map[string]any{{"type": "combat", "id": entryID}},
	})
	testutil.AssertStatus(t, w, 200)
	testutil.ParseJSON(t, w, &out)
	if len(out.Applied) != 1 || out.Applied[0].Healing <= 0 || out.Applied[0].HPAfter <= out.Applied[0].HPBefore {
		t.Fatalf("heal did not apply: %+v", out.Applied)
	}

	// 3. Character target takes damage through the HP authority.
	w = testutil.PostJSON(t, r, "/api/characters/10/cast-spell", map[string]any{
		"spell_id": spellID, "damage": "1d6", "damage_type": "force",
		"targets": []map[string]any{{"type": "character", "id": 20}},
	})
	testutil.AssertStatus(t, w, 200)
	testutil.ParseJSON(t, w, &out)
	if len(out.Applied) != 1 || out.Applied[0].Damage <= 0 || out.Applied[0].HPAfter >= out.Applied[0].HPBefore {
		t.Fatalf("character damage did not apply: %+v", out.Applied)
	}

	// 4. No targets still returns the plain cast.
	w = testutil.PostJSON(t, r, "/api/characters/10/cast-spell", map[string]any{"spell_id": spellID})
	testutil.AssertStatus(t, w, 200)
	testutil.ParseJSON(t, w, &out)
	if len(out.Applied) != 0 {
		t.Fatalf("applied = %d without targets, want 0", len(out.Applied))
	}
}
