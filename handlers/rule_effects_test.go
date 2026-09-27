package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers/testutil"
)

func TestApplyDamageDefenses(t *testing.T) {
	tests := []struct {
		name                 string
		dmg                  int
		damageType           string
		resist, vuln, immune string
		wantDmg              int
		wantCategory         string
	}{
		{"immune zeroes", 20, "fire", "", "", "fire, poison, cold", 0, "immune"},
		{"resistant halves", 11, "cold", "cold", "", "", 5, "resistant"},
		{"vulnerable doubles", 7, "radiant", "", "radiant", "", 14, "vulnerable"},
		{"unmatched", 7, "slashing", "fire", "radiant", "poison", 7, ""},
		{"empty type", 7, "", "fire", "radiant", "poison", 7, ""},
		{"immune wins over resist", 10, "fire", "fire", "", "fire", 0, "immune"},
		{"case insensitive and spaced", 8, " Fire ", "", "", " fire , COLD ", 0, "immune"},
		{"non-positive damage", 0, "fire", "fire", "", "", 0, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotDmg, gotCat := ApplyDamageDefenses(tc.dmg, tc.damageType, tc.resist, tc.vuln, tc.immune)
			if gotDmg != tc.wantDmg || gotCat != tc.wantCategory {
				t.Fatalf("got (%d,%q) want (%d,%q)", gotDmg, gotCat, tc.wantDmg, tc.wantCategory)
			}
		})
	}
}

func TestEffectsFromConditions(t *testing.T) {
	e := EffectsFromConditions([]string{"blinded", "restrained", "poisoned"})
	if !e.DisadvAttacks || !e.AttacksAgainstAdvantage || !e.SpeedZero || !e.DisadvChecks {
		t.Fatalf("unexpected aggregate %+v", e)
	}
	p := EffectsFromConditions([]string{"paralyzed"})
	if !p.Incapacitated || !p.SpeedZero || !p.AutoFailStrDex || !p.AttacksAgainstAdvantage {
		t.Fatalf("paralyzed effects wrong %+v", p)
	}
	if EffectsFromConditions([]string{"charmed"}).Incapacitated {
		t.Fatal("charmed should have no mechanical effects")
	}
}

func TestExhaustionEffectsForLevel(t *testing.T) {
	if e := ExhaustionEffectsForLevel(0); len(e.Effects) != 0 || e.Dead {
		t.Fatalf("level 0 should be no-op %+v", e)
	}
	e3 := ExhaustionEffectsForLevel(3)
	if !e3.DisadvChecks || !e3.SpeedHalved || !e3.DisadvAttacks || !e3.DisadvSaves || e3.HPMaxHalved {
		t.Fatalf("level 3 wrong %+v", e3)
	}
	e4 := ExhaustionEffectsForLevel(4)
	if !e4.HPMaxHalved {
		t.Fatalf("level 4 should halve hp max %+v", e4)
	}
	e6 := ExhaustionEffectsForLevel(6)
	if !e6.Dead || !e6.SpeedZero {
		t.Fatalf("level 6 wrong %+v", e6)
	}
}

func TestCombineAdvantage(t *testing.T) {
	cases := []struct {
		req      string
		adv, dis bool
		want     string
	}{
		{"normal", false, false, "normal"},
		{"advantage", false, false, "advantage"},
		{"advantage", false, true, "normal"},
		{"normal", true, true, "normal"},
		{"normal", true, false, "advantage"},
		{"normal", false, true, "disadvantage"},
		{"disadvantage", true, false, "normal"},
	}
	for _, c := range cases {
		if got := CombineAdvantage(c.req, c.adv, c.dis); got != c.want {
			t.Fatalf("CombineAdvantage(%q,%v,%v)=%q want %q", c.req, c.adv, c.dis, got, c.want)
		}
	}
}

func TestEffectiveHPMax(t *testing.T) {
	if got := EffectiveHPMax(20, 3); got != 20 {
		t.Fatalf("exhaustion 3 expected 20 got %d", got)
	}
	if got := EffectiveHPMax(21, 4); got != 10 {
		t.Fatalf("exhaustion 4 expected 10 got %d", got)
	}
}

func TestApplyHPChangeDefenses(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/characters/:id/hp", HandleCharacterHP)
	})

	t.Run("immunity negates damage", func(t *testing.T) {
		testutil.SeedCharacter(t, 30, 1, "ImmuneGuy", "Human", "Fighter")
		db.DB.Exec("UPDATE characters SET hp_max=20, hp_current=20, damage_immunities='fire' WHERE id=30")
		w := testutil.PostJSON(t, r, "/api/characters/30/hp", map[string]any{"delta": -12, "type": "fire", "source": "test"})
		testutil.AssertStatus(t, w, 200)
		var res HPChangeResult
		testutil.ParseJSON(t, w, &res)
		if res.HPCurrent != 20 || res.DefenseApplied != "immune" || res.EffectiveDamage != 0 {
			t.Fatalf("immunity failed %+v", res)
		}
	})

	t.Run("resistance halves damage", func(t *testing.T) {
		testutil.SeedCharacter(t, 31, 1, "ResistGuy", "Human", "Fighter")
		db.DB.Exec("UPDATE characters SET hp_max=20, hp_current=20, damage_resistances='cold' WHERE id=31")
		w := testutil.PostJSON(t, r, "/api/characters/31/hp", map[string]any{"delta": -11, "type": "cold", "source": "test"})
		testutil.AssertStatus(t, w, 200)
		var res HPChangeResult
		testutil.ParseJSON(t, w, &res)
		if res.HPCurrent != 15 || res.DefenseApplied != "resistant" || res.EffectiveDamage != 5 {
			t.Fatalf("resistance failed %+v", res)
		}
	})

	t.Run("vulnerability doubles damage", func(t *testing.T) {
		testutil.SeedCharacter(t, 32, 1, "VulnGuy", "Human", "Fighter")
		db.DB.Exec("UPDATE characters SET hp_max=40, hp_current=40, damage_vulnerabilities='radiant' WHERE id=32")
		w := testutil.PostJSON(t, r, "/api/characters/32/hp", map[string]any{"delta": -7, "type": "radiant", "source": "test"})
		testutil.AssertStatus(t, w, 200)
		var res HPChangeResult
		testutil.ParseJSON(t, w, &res)
		if res.HPCurrent != 26 || res.EffectiveDamage != 14 {
			t.Fatalf("vulnerability failed %+v", res)
		}
	})

	t.Run("exhaustion halves effective hp max on heal", func(t *testing.T) {
		testutil.SeedCharacter(t, 33, 1, "TiredGuy", "Human", "Fighter")
		db.DB.Exec("UPDATE characters SET hp_max=20, hp_current=2, exhaustion_level=4 WHERE id=33")
		w := testutil.PostJSON(t, r, "/api/characters/33/hp", map[string]any{"delta": 100, "type": "", "source": ""})
		testutil.AssertStatus(t, w, 200)
		var res HPChangeResult
		testutil.ParseJSON(t, w, &res)
		if res.HPCurrent != 10 || res.EffectiveHPMax != 10 {
			t.Fatalf("exhaustion hp clamp failed %+v", res)
		}
	})

	t.Run("massive damage kills at zero", func(t *testing.T) {
		testutil.SeedCharacter(t, 34, 1, "DoomedGuy", "Human", "Fighter")
		db.DB.Exec("UPDATE characters SET hp_max=10, hp_current=0, death_saves_failures=0 WHERE id=34")
		w := testutil.PostJSON(t, r, "/api/characters/34/hp", map[string]any{"delta": -10, "type": "slashing", "source": "test"})
		testutil.AssertStatus(t, w, 200)
		var res HPChangeResult
		testutil.ParseJSON(t, w, &res)
		if res.DeathSavesFailures != 3 {
			t.Fatalf("massive damage should set 3 failures got %+v", res)
		}
	})
}

func TestHandleDeathSave(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/characters/:id/death-save", HandleDeathSave)
	})

	t.Run("not dying returns 400", func(t *testing.T) {
		testutil.SeedCharacter(t, 40, 1, "Standing", "Human", "Fighter")
		db.DB.Exec("UPDATE characters SET hp_max=10, hp_current=5 WHERE id=40")
		w := testutil.PostJSON(t, r, "/api/characters/40/death-save", map[string]any{})
		testutil.AssertStatus(t, w, 400)
	})

	t.Run("dying character rolls a save", func(t *testing.T) {
		testutil.SeedCharacter(t, 41, 1, "Down", "Human", "Fighter")
		db.DB.Exec("UPDATE characters SET hp_max=10, hp_current=0, death_saves_successes=0, death_saves_failures=0 WHERE id=41")
		w := testutil.PostJSON(t, r, "/api/characters/41/death-save", map[string]any{})
		testutil.AssertStatus(t, w, 200)
		var res struct {
			Roll      int  `json:"roll"`
			Successes int  `json:"successes"`
			Failures  int  `json:"failures"`
			Revived   bool `json:"revived"`
			HPCurrent int  `json:"hp_current"`
		}
		testutil.ParseJSON(t, w, &res)
		if res.Roll < 1 || res.Roll > 20 {
			t.Fatalf("roll out of range %d", res.Roll)
		}
		if res.Revived {
			if res.Roll != 20 || res.HPCurrent != 1 {
				t.Fatalf("revived without nat20 %+v", res)
			}
		} else if res.Successes+res.Failures == 0 {
			t.Fatalf("expected a success or failure recorded %+v", res)
		}
	})
}

func TestResolveCheckRollConditions(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 60, 1, "Pinned", "Human", "Fighter")

	db.DB.Exec("INSERT INTO character_conditions(character_id,name,type,duration,duration_type) VALUES(60,'Paralyzed','paralyzed',1,'round')")

	res, err := resolveCheckRoll(60, CheckRollRequest{CharacterID: 60, Type: "save", Name: "dex"})
	if err != nil {
		t.Fatal(err)
	}
	if !res.AutoFail || res.Total != 0 {
		t.Fatalf("dex save should auto-fail %+v", res)
	}
	resStr, _ := resolveCheckRoll(60, CheckRollRequest{CharacterID: 60, Type: "save", Name: "str"})
	if !resStr.AutoFail {
		t.Fatalf("str save should auto-fail %+v", resStr)
	}
	resCon, _ := resolveCheckRoll(60, CheckRollRequest{CharacterID: 60, Type: "save", Name: "con"})
	if resCon.AutoFail {
		t.Fatalf("con save should not auto-fail %+v", resCon)
	}

	// Poisoned gives disadvantage on checks; an explicit advantage cancels it.
	db.DB.Exec("INSERT INTO character_conditions(character_id,name,type,duration,duration_type) VALUES(60,'Poisoned','poisoned',1,'round')")
	resDis, _ := resolveCheckRoll(60, CheckRollRequest{CharacterID: 60, Type: "skill", Name: "perception"})
	if resDis.Advantage != "disadvantage" {
		t.Fatalf("poisoned should impose disadvantage got %q", resDis.Advantage)
	}
	resCancel, _ := resolveCheckRoll(60, CheckRollRequest{CharacterID: 60, Type: "skill", Name: "perception", Advantage: "advantage"})
	if resCancel.Advantage != "normal" {
		t.Fatalf("advantage should cancel disadvantage got %q", resCancel.Advantage)
	}
}

func TestHandleCharacterEffects(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/characters/:id/effects", HandleCharacterEffects)
	})

	testutil.SeedCharacter(t, 50, 1, "Sufferer", "Human", "Fighter")
	db.DB.Exec("UPDATE characters SET speed=30, hp_max=20, exhaustion_level=4, damage_immunities='poison' WHERE id=50")
	db.DB.Exec("INSERT INTO character_conditions(character_id,name,type,duration,duration_type) VALUES(50,'Blinded','blinded',2,'round')")
	db.DB.Exec("INSERT INTO character_conditions(character_id,name,type,duration,duration_type) VALUES(50,'Restrained','restrained',2,'round')")

	w := testutil.Get(t, r, "/api/characters/50/effects")
	testutil.AssertStatus(t, w, 200)
	var res CharacterEffects
	testutil.ParseJSON(t, w, &res)
	if !res.ConditionEffects.DisadvAttacks || !res.ConditionEffects.AttacksAgainstAdvantage {
		t.Fatalf("condition effects wrong %+v", res.ConditionEffects)
	}
	if res.ConditionEffects.SpeedZero != true {
		t.Fatalf("restrained should zero speed %+v", res.ConditionEffects)
	}
	if res.EffectiveHPMax != 10 {
		t.Fatalf("exhaustion 4 should halve hp max got %d", res.EffectiveHPMax)
	}
	if res.Speed != 0 || res.DamageImmunities != "poison" {
		t.Fatalf("effects response wrong %+v", res)
	}
}
