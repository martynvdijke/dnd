package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"villum/db"
)

// ApplyDamageDefenses applies damage immunities (0), resistances (halved,
// rounded down) and vulnerabilities (doubled) for a damage type. Priority is
// immunity > resistance > vulnerability. Returns the effective damage and the
// matched category ("immune", "resistant", "vulnerable") or "".
func ApplyDamageDefenses(damage int, damageType, resist, vuln, immune string) (int, string) {
	if damage <= 0 {
		return damage, ""
	}
	dt := strings.ToLower(strings.TrimSpace(damageType))
	if dt == "" {
		return damage, ""
	}
	if listContains(immune, dt) {
		return 0, "immune"
	}
	if listContains(resist, dt) {
		return damage / 2, "resistant"
	}
	if listContains(vuln, dt) {
		return damage * 2, "vulnerable"
	}
	return damage, ""
}

func listContains(csv, want string) bool {
	want = strings.ToLower(strings.TrimSpace(want))
	if want == "" {
		return false
	}
	for _, p := range strings.Split(csv, ",") {
		if strings.ToLower(strings.TrimSpace(p)) == want {
			return true
		}
	}
	return false
}

// ConditionEffects is the mechanical roll/speed effect aggregate of a set of
// active conditions.
type ConditionEffects struct {
	AdvAttacks              bool `json:"adv_attacks"`
	DisadvAttacks           bool `json:"disadv_attacks"`
	AdvSaves                bool `json:"adv_saves"`
	DisadvSaves             bool `json:"disadv_saves"`
	AdvChecks               bool `json:"adv_checks"`
	DisadvChecks            bool `json:"disadv_checks"`
	AutoFailStrDex          bool `json:"auto_fail_str_dex"`
	SpeedZero               bool `json:"speed_zero"`
	Incapacitated           bool `json:"incapacitated"`
	AttacksAgainstAdvantage bool `json:"attacks_against_advantage"`
}

// EffectsFromConditions aggregates the mechanical effects of condition types.
func EffectsFromConditions(conditions []string) ConditionEffects {
	var e ConditionEffects
	for _, raw := range conditions {
		switch strings.ToLower(strings.TrimSpace(raw)) {
		case "blinded":
			e.DisadvAttacks = true
			e.AttacksAgainstAdvantage = true
		case "frightened":
			e.DisadvChecks = true
			e.DisadvAttacks = true
		case "grappled":
			e.SpeedZero = true
		case "incapacitated":
			e.Incapacitated = true
		case "invisible":
			e.AdvAttacks = true
		case "paralyzed", "petrified", "stunned", "unconscious":
			e.Incapacitated = true
			e.SpeedZero = true
			e.AutoFailStrDex = true
			e.AttacksAgainstAdvantage = true
		case "poisoned":
			e.DisadvAttacks = true
			e.DisadvChecks = true
		case "prone":
			e.AttacksAgainstAdvantage = true
		case "restrained":
			e.DisadvAttacks = true
			e.DisadvSaves = true
			e.SpeedZero = true
			e.AttacksAgainstAdvantage = true
		}
	}
	return e
}

// ExhaustionEffects is the mechanical effect of an exhaustion level.
type ExhaustionEffects struct {
	Level         int      `json:"level"`
	DisadvChecks  bool     `json:"disadv_checks"`
	DisadvAttacks bool     `json:"disadv_attacks"`
	DisadvSaves   bool     `json:"disadv_saves"`
	SpeedHalved   bool     `json:"speed_halved"`
	SpeedZero     bool     `json:"speed_zero"`
	HPMaxHalved   bool     `json:"hp_max_halved"`
	Dead          bool     `json:"dead"`
	Effects       []string `json:"effects"`
}

// ExhaustionEffectsForLevel returns cumulative effects for exhaustion 0-6.
func ExhaustionEffectsForLevel(level int) ExhaustionEffects {
	e := ExhaustionEffects{Level: level, Effects: []string{}}
	add := func(s string) { e.Effects = append(e.Effects, s) }
	if level >= 1 {
		e.DisadvChecks = true
		add("Disadvantage on ability checks")
	}
	if level >= 2 {
		e.SpeedHalved = true
		add("Speed halved")
	}
	if level >= 3 {
		e.DisadvAttacks = true
		e.DisadvSaves = true
		add("Disadvantage on attack rolls and saving throws")
	}
	if level >= 4 {
		e.HPMaxHalved = true
		add("Hit point maximum halved")
	}
	if level >= 5 {
		e.SpeedZero = true
		add("Speed reduced to 0")
	}
	if level >= 6 {
		e.Dead = true
		add("Death")
	}
	return e
}

// CombineAdvantage merges a requested advantage state with forced
// advantage/disadvantage flags. Any advantage and disadvantage together cancel
// to normal.
func CombineAdvantage(requested string, adv, disadv bool) string {
	r := strings.ToLower(strings.TrimSpace(requested))
	hasAdv := adv || r == "advantage"
	hasDis := disadv || r == "disadvantage"
	switch {
	case hasAdv && hasDis:
		return "normal"
	case hasAdv:
		return "advantage"
	case hasDis:
		return "disadvantage"
	default:
		return "normal"
	}
}

// EffectiveHPMax applies the exhaustion level 4+ hit point maximum halving.
func EffectiveHPMax(hpMax, exhaustion int) int {
	if exhaustion >= 4 {
		return hpMax / 2
	}
	return hpMax
}

// --- DB helpers ---

// loadActiveConditionTypes returns the lowercase condition types currently on
// a character.
func loadActiveConditionTypes(charID int64) []string {
	rows, err := db.DB.Query("SELECT LOWER(COALESCE(NULLIF(type,''), name)) FROM character_conditions WHERE character_id=?", charID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var t string
		if rows.Scan(&t) == nil && t != "" {
			out = append(out, t)
		}
	}
	return out
}

func loadCharacterExhaustion(charID int64) int {
	var level int
	db.DB.QueryRow("SELECT COALESCE(exhaustion_level,0) FROM characters WHERE id=?", charID).Scan(&level)
	return level
}

func loadCharacterDefenses(charID int64) (resist, vuln, immune string) {
	db.DB.QueryRow("SELECT COALESCE(damage_resistances,''), COALESCE(damage_vulnerabilities,''), COALESCE(damage_immunities,'') FROM characters WHERE id=?", charID).
		Scan(&resist, &vuln, &immune)
	return
}

// CharacterEffects is the response for GET /api/characters/:id/effects.
type CharacterEffects struct {
	Conditions            []string          `json:"conditions"`
	ConditionEffects      ConditionEffects  `json:"condition_effects"`
	Exhaustion            ExhaustionEffects `json:"exhaustion"`
	DamageResistances     string            `json:"damage_resistances"`
	DamageVulnerabilities string            `json:"damage_vulnerabilities"`
	DamageImmunities      string            `json:"damage_immunities"`
	EffectiveHPMax        int               `json:"effective_hp_max"`
	Speed                 int               `json:"speed"`
}

// HandleCharacterEffects reports a character's active mechanical effects.
func HandleCharacterEffects(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if !canEditCharacterID(c, charID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}
	var speed, hpMax, exhaustion int
	err := db.DB.QueryRow("SELECT COALESCE(speed,0), COALESCE(hp_max,0), COALESCE(exhaustion_level,0) FROM characters WHERE id=?", charID).
		Scan(&speed, &hpMax, &exhaustion)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}
	conditions := loadActiveConditionTypes(charID)
	condEff := EffectsFromConditions(conditions)
	exEff := ExhaustionEffectsForLevel(exhaustion)
	resist, vuln, immune := loadCharacterDefenses(charID)

	effSpeed := speed
	if condEff.SpeedZero || exEff.SpeedZero {
		effSpeed = 0
	} else if exEff.SpeedHalved {
		effSpeed = speed / 2
	}

	c.JSON(http.StatusOK, CharacterEffects{
		Conditions:            conditions,
		ConditionEffects:      condEff,
		Exhaustion:            exEff,
		DamageResistances:     resist,
		DamageVulnerabilities: vuln,
		DamageImmunities:      immune,
		EffectiveHPMax:        EffectiveHPMax(hpMax, exhaustion),
		Speed:                 effSpeed,
	})
}
