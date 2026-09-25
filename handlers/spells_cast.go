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

// CastSpell casts a prepared spell: consumes a slot, broadcasts a spell_cast
// event to the campaign, and triggers the WLED effect. POST /characters/:id/cast-spell
func CastSpell(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if !canEditCharacterID(c, charID) {
		WriteError(c, http.StatusForbidden, errAccessDenied)
		return
	}
	var body struct {
		SpellID int64 `json:"spell_id"`
	}
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

	effect := SpellEffectFor(name, school)

	slotConsumed := false
	if level >= 1 && level <= 9 {
		if consumed, cerr := consumeSpellSlot(charID, level); cerr == nil {
			slotConsumed = consumed
		}
	}

	var campaignID int64
	db.DB.QueryRow("SELECT COALESCE(campaign_id,0) FROM characters WHERE id=?", charID).Scan(&campaignID)
	if campaignID > 0 {
		SendSpellCast(campaignID, charID, body.SpellID, strings.TrimSpace(name), school, level, effect)
	}
	NotifyWLEDEffect(effect)

	WriteJSON(c, http.StatusOK, gin.H{
		"ok":            true,
		"effect":        effect,
		"spell":         name,
		"level":         level,
		"slot_consumed": slotConsumed,
	})
}
