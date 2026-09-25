package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"villum/db"
)

// consumeSpellSlot increments the used counter for a spell level, only if a
// character_spellcasting row exists and has a slot free. The statement is
// chosen from a fixed switch so no client value reaches the SQL text.
func consumeSpellSlot(charID int64, level int) (bool, error) {
	var stmt string
	switch level {
	case 1:
		stmt = "UPDATE character_spellcasting SET slots_1_used = slots_1_used + 1 WHERE character_id=? AND slots_1_used < slots_1_max"
	case 2:
		stmt = "UPDATE character_spellcasting SET slots_2_used = slots_2_used + 1 WHERE character_id=? AND slots_2_used < slots_2_max"
	case 3:
		stmt = "UPDATE character_spellcasting SET slots_3_used = slots_3_used + 1 WHERE character_id=? AND slots_3_used < slots_3_max"
	case 4:
		stmt = "UPDATE character_spellcasting SET slots_4_used = slots_4_used + 1 WHERE character_id=? AND slots_4_used < slots_4_max"
	case 5:
		stmt = "UPDATE character_spellcasting SET slots_5_used = slots_5_used + 1 WHERE character_id=? AND slots_5_used < slots_5_max"
	case 6:
		stmt = "UPDATE character_spellcasting SET slots_6_used = slots_6_used + 1 WHERE character_id=? AND slots_6_used < slots_6_max"
	case 7:
		stmt = "UPDATE character_spellcasting SET slots_7_used = slots_7_used + 1 WHERE character_id=? AND slots_7_used < slots_7_max"
	case 8:
		stmt = "UPDATE character_spellcasting SET slots_8_used = slots_8_used + 1 WHERE character_id=? AND slots_8_used < slots_8_max"
	case 9:
		stmt = "UPDATE character_spellcasting SET slots_9_used = slots_9_used + 1 WHERE character_id=? AND slots_9_used < slots_9_max"
	default:
		return false, nil
	}
	res, err := db.DB.Exec(stmt, charID)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

// spellCastTarget is one recipient of a spell's damage or healing.
type spellCastTarget struct {
	Type string `json:"type"` // "combat" (combat entry) or "character"
	ID   int64  `json:"id"`
}

// castSpellRequest is the body of POST /characters/:id/cast-spell.
type castSpellRequest struct {
	SpellID     int64             `json:"spell_id"`
	Damage      string            `json:"damage"`
	DamageType  string            `json:"damage_type"`
	Healing     bool              `json:"healing"`
	SaveAbility string            `json:"save_ability"`
	SaveDC      int               `json:"save_dc"`
	Targets     []spellCastTarget `json:"targets"`
}

// spellTargetResult reports what a cast did to one target.
type spellTargetResult struct {
	Target     string `json:"target"`
	Type       string `json:"type"`
	ID         int64  `json:"id"`
	HPBefore   int    `json:"hp_before"`
	HPAfter    int    `json:"hp_after"`
	HPMax      int    `json:"hp_max"`
	Damage     int    `json:"damage"`
	Healing    int    `json:"healing"`
	SaveRolled bool   `json:"save_rolled"`
	SaveTotal  int    `json:"save_total,omitempty"`
	Saved      bool   `json:"saved"`
	SaveDC     int    `json:"save_dc,omitempty"`
	Down       bool   `json:"down,omitempty"`
}

// rollAbilitySave rolls a d20 + ability mod (+ proficiency when the character
// has a matching save proficiency) against dc. ok=false when the character is
// unknown or the ability is invalid.
func rollAbilitySave(charID int64, ability string, dc int) (total int, success, ok bool) {
	var str, dex, con, intel, wis, cha, prof, level int
	if err := db.DB.QueryRow(
		"SELECT str,dex,con,int,wis,cha,COALESCE(proficiency_bonus,0),level FROM characters WHERE id=?",
		charID,
	).Scan(&str, &dex, &con, &intel, &wis, &cha, &prof, &level); err != nil {
		return 0, false, false
	}
	var mod int
	switch strings.ToLower(strings.TrimSpace(ability)) {
	case "str":
		mod = abilityMod(str)
	case "dex":
		mod = abilityMod(dex)
	case "con":
		mod = abilityMod(con)
	case "int":
		mod = abilityMod(intel)
	case "wis":
		mod = abilityMod(wis)
	case "cha":
		mod = abilityMod(cha)
	default:
		return 0, false, false
	}
	var cnt int
	db.DB.QueryRow(
		"SELECT COUNT(*) FROM character_proficiencies WHERE character_id=? AND type='save' AND LOWER(name)=?",
		charID, strings.ToLower(strings.TrimSpace(ability)),
	).Scan(&cnt)
	if cnt > 0 {
		if prof == 0 {
			prof = 2 + (level-1)/4
		}
		mod += prof
	}
	rolled, err := getDicePool().Roll("1d20")
	if err != nil {
		return 0, false, false
	}
	v, _ := strconv.Atoi(string(rolled.Total))
	total = v + mod
	return total, total >= dc, true
}

// CastSpell casts a prepared spell: consumes a slot, broadcasts a spell_cast
// event to the campaign, triggers the WLED effect, and (when targets are
// supplied) applies damage or healing to each one, halving on a successful
// saving throw for character-backed targets.
// POST /characters/:id/cast-spell
func CastSpell(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if !canEditCharacterID(c, charID) {
		WriteError(c, http.StatusForbidden, errAccessDenied)
		return
	}
	var body castSpellRequest
	if !BindOr400(c, &body) {
		return
	}
	var name, school string
	var level int
	err := db.DB.QueryRow(
		"SELECT name, level, COALESCE(school,'') FROM spells WHERE id=? AND character_id=?",
		body.SpellID, charID,
	).Scan(&name, &level, &school)
	if err != nil {
		WriteNotFound(c, "spell not found")
		return
	}
	spellName := strings.TrimSpace(name)
	effect := SpellEffectFor(name, school)

	slotConsumed := false
	if level >= 1 && level <= 9 {
		if consumed, cerr := consumeSpellSlot(charID, level); cerr == nil {
			slotConsumed = consumed
		}
	}

	var campaignID int64
	db.DB.QueryRow("SELECT COALESCE(campaign_id,0) FROM characters WHERE id=?", charID).Scan(&campaignID)
	var campPtr *int64
	if campaignID > 0 {
		campPtr = &campaignID
	}

	// Area effect: one damage roll shared by all targets (5e style).
	damageRolled := 0
	if strings.TrimSpace(body.Damage) != "" {
		total, _, derr := rollDamageExpr(body.Damage, false)
		if derr != nil {
			WriteError(c, http.StatusBadRequest, strErr("invalid damage expression"))
			return
		}
		damageRolled = total
	}
	results := make([]spellTargetResult, 0, len(body.Targets))
	appliedAny := false
	for _, t := range body.Targets {
		res, applied := applySpellToTarget(t, damageRolled, body, spellName, campPtr)
		if !applied {
			continue
		}
		results = append(results, res)
		appliedAny = true
	}
	if appliedAny && campaignID > 0 {
		SendCombatUpdate(campaignID)
	}

	if campaignID > 0 {
		SendSpellCast(campaignID, charID, body.SpellID, spellName, school, level, effect)
	}
	NotifyWLEDEffect(effect)

	WriteJSON(c, http.StatusOK, gin.H{
		"ok":            true,
		"effect":        effect,
		"spell":         name,
		"level":         level,
		"slot_consumed": slotConsumed,
		"damage_rolled": damageRolled,
		"applied":       results,
	})
}

// applySpellToTarget resolves one target, rolls its save when applicable and
// writes the resulting HP change. applied=false means the target was skipped.
func applySpellToTarget(t spellCastTarget, damageRolled int, body castSpellRequest, spellName string, campPtr *int64) (spellTargetResult, bool) {
	res := spellTargetResult{Type: t.Type, ID: t.ID, SaveDC: body.SaveDC}

	effective := damageRolled
	rollSave := body.SaveAbility != "" && body.SaveDC > 0

	switch t.Type {
	case "combat":
		var hpCur, hpMax int
		var targetName string
		var entryCharID *int64
		if err := db.DB.QueryRow(
			"SELECT hp_current, hp_max, name, character_id FROM combat_entries WHERE id=?", t.ID,
		).Scan(&hpCur, &hpMax, &targetName, &entryCharID); err != nil {
			return res, false
		}
		res.Target = targetName
		res.HPBefore, res.HPMax = hpCur, hpMax
		if rollSave && entryCharID != nil {
			if total, saved, ok := rollAbilitySave(*entryCharID, body.SaveAbility, body.SaveDC); ok {
				res.SaveRolled, res.SaveTotal, res.Saved = true, total, saved
				if saved {
					effective /= 2
				}
			}
		}
		delta := -effective
		if body.Healing {
			delta = damageRolled
		}
		newHP := hpCur + delta
		if newHP < 0 {
			newHP = 0
		}
		if newHP > hpMax {
			newHP = hpMax
		}
		if _, err := db.DB.Exec("UPDATE combat_entries SET hp_current=? WHERE id=?", newHP, t.ID); err != nil {
			return res, false
		}
		res.HPAfter = newHP
		if applied := newHP - hpCur; applied > 0 {
			res.Healing = applied
		} else {
			res.Damage = -applied
		}
		res.Down = newHP == 0
		logCombatEvent(CombatLogEntry{
			CampaignID: campPtr, CombatEntryID: &t.ID, ActorName: spellName,
			Action: "spell", TargetName: targetName,
			Damage: res.Damage, DamageType: body.DamageType, Healing: res.Healing,
			RollExpression: body.Damage, RollTotal: damageRolled,
		})
		return res, true

	case "character":
		var hpCur, hpMax int
		var targetName string
		if err := db.DB.QueryRow(
			"SELECT hp_current, hp_max, name FROM characters WHERE id=?", t.ID,
		).Scan(&hpCur, &hpMax, &targetName); err != nil {
			return res, false
		}
		res.Target = targetName
		res.HPBefore, res.HPMax = hpCur, hpMax
		if rollSave {
			if total, saved, ok := rollAbilitySave(t.ID, body.SaveAbility, body.SaveDC); ok {
				res.SaveRolled, res.SaveTotal, res.Saved = true, total, saved
				if saved {
					effective /= 2
				}
			}
		}
		delta := -effective
		if body.Healing {
			delta = damageRolled
		}
		hr, aerr := applyHPChange(t.ID, delta, body.DamageType, "spell:"+spellName, campPtr)
		if aerr != nil {
			return res, false
		}
		res.HPAfter = hr.HPCurrent
		if applied := hr.HPCurrent - hpCur; applied > 0 {
			res.Healing = applied
		} else {
			res.Damage = -applied
		}
		res.Down = hr.HPCurrent == 0
		return res, true

	default:
		return res, false
	}
}
