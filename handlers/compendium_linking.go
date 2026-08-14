package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
)

// ─── Spell Linking (Character Sheet) ───

// spellLinkInsert copies a compendium spell into the character's spellbook.
// Returns (httpStatus, errorMessage); message == "" on success.
func spellLinkInsert(charID, compendiumSpellID int64) (int, string) {
	var name, school, castingTime, range_, components, duration, description, higherLevels, classes, sourcePage string
	var level int
	err := db.DB.QueryRow(`SELECT name, level, school, casting_time, "range", components, duration, description, COALESCE(higher_levels,''), COALESCE(classes,'[]'), COALESCE(source_page,'') FROM compendium_spells WHERE id=?`, compendiumSpellID).Scan(
		&name, &level, &school, &castingTime, &range_, &components, &duration, &description, &higherLevels, &classes, &sourcePage)
	if err != nil {
		return http.StatusNotFound, "compendium spell not found"
	}
	_, err = db.DB.Exec(`INSERT INTO spells(character_id, name, level, school, casting_time, "range", components, duration, description, source, notes, compendium_spell_id)
		VALUES(?,?,?,?,?,?,?,?,?,?,'',?)`,
		charID, name, level, school, castingTime, range_, components, duration, description, sourcePage, compendiumSpellID)
	if err != nil {
		return http.StatusInternalServerError, err.Error()
	}
	return 0, ""
}

// unlinkSpellRef nulls the compendium references on a spell, preserving its data.
func unlinkSpellRef(spellID int64) (int, string) {
	var compID, entryID int64
	err := db.DB.QueryRow("SELECT COALESCE(compendium_spell_id,0), COALESCE(compendium_entry_id,0) FROM spells WHERE id=?", spellID).Scan(&compID, &entryID)
	if err != nil {
		return http.StatusNotFound, "spell not found"
	}
	if compID == 0 && entryID == 0 {
		return http.StatusBadRequest, "spell is not linked from compendium"
	}
	_, err = db.DB.Exec("UPDATE spells SET compendium_spell_id = NULL, compendium_entry_id = NULL WHERE id=?", spellID)
	if err != nil {
		return http.StatusInternalServerError, err.Error()
	}
	return 0, ""
}

// formLinkID returns the first non-empty numeric id posted under any of the given form keys.
func formLinkID(c *gin.Context, keys ...string) (int64, bool) {
	for _, k := range keys {
		if v := c.PostForm(k); v != "" {
			if id, err := strconv.ParseInt(v, 10, 64); err == nil && id > 0 {
				return id, true
			}
		}
	}
	return 0, false
}

// LinkCompendiumSpell creates a spell row for a character linked to a compendium spell.
// POST /api/characters/:id/spells/link
func LinkCompendiumSpell(c *gin.Context) {
	charID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid character id"})
		return
	}
	if !canEditCharacterID(c, charID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	if entryID, ok := formLinkID(c, "compendium_entry_id"); ok {
		if st, msg := entrySpellLinkInsert(charID, entryID); msg != "" {
			c.JSON(st, gin.H{"error": msg})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"status": "linked"})
		return
	}
	compendiumSpellID, err := strconv.ParseInt(c.PostForm("compendium_spell_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "compendium_spell_id required"})
		return
	}
	if st, msg := spellLinkInsert(charID, compendiumSpellID); msg != "" {
		c.JSON(st, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "linked"})
}

// HtmxLinkCompendiumSpell links a compendium spell and re-renders the spells list.
// POST /htmx/compendium/spells/link?character_id=:cid (form: compendium_spell_id)
func HtmxLinkCompendiumSpell(c *gin.Context) {
	charID := c.Query("character_id")
	cid, err := strconv.ParseInt(charID, 10, 64)
	if err != nil || charID == "" {
		c.String(http.StatusBadRequest, "invalid character id")
		return
	}
	if !canEditCharacterID(c, cid) {
		c.String(http.StatusForbidden, "access denied")
		return
	}
	if entryID, ok := formLinkID(c, "compendium_entry_id"); ok {
		if st, msg := entrySpellLinkInsert(cid, entryID); msg != "" {
			c.String(st, msg)
			return
		}
		renderHtmxSpellsList(c, charID)
		return
	}
	compID, err := strconv.ParseInt(c.PostForm("compendium_spell_id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "compendium_spell_id required")
		return
	}
	if st, msg := spellLinkInsert(cid, compID); msg != "" {
		c.String(st, msg)
		return
	}
	renderHtmxSpellsList(c, strconv.FormatInt(cid, 10))
}

// UnlinkCompendiumSpell unlinks a spell from the compendium, preserving the spell data.
// DELETE /api/characters/:id/spells/:spellId/link  (also POST /api/characters/:id/spells/:spellId/unlink)
func UnlinkCompendiumSpell(c *gin.Context) {
	spellID, err := strconv.ParseInt(c.Param("spellId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid spell id"})
		return
	}
	if !canEditResourceID(c, "spells", spellID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	if st, msg := unlinkSpellRef(spellID); msg != "" {
		c.JSON(st, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "unlinked"})
}

// HtmxUnlinkCompendiumSpell unlinks a compendium spell and re-renders the spells list.
// DELETE /htmx/spells/:id/compendium-unlink?character_id=:cid
func HtmxUnlinkCompendiumSpell(c *gin.Context) {
	spellID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid spell id")
		return
	}
	if !canEditResourceID(c, "spells", spellID) {
		c.String(http.StatusForbidden, "access denied")
		return
	}
	charID := c.Query("character_id")
	if charID == "" {
		c.String(http.StatusBadRequest, "character_id required")
		return
	}
	if st, msg := unlinkSpellRef(spellID); msg != "" {
		c.String(st, msg)
		return
	}
	renderHtmxSpellsList(c, charID)
}

// ─── Inventory/Equipment Linking (Character Sheet) ───

// itemLinkInsert copies a compendium equipment entry into the character's inventory.
// Returns (httpStatus, errorMessage); message == "" on success.
func itemLinkInsert(charID, compendiumEquipmentID int64, quantity int) (int, string) {
	var name, category, description, sourcePage string
	var weight float64
	err := db.DB.QueryRow(`SELECT name, COALESCE(category,''), COALESCE(description,''), COALESCE(weight,0), COALESCE(source_page,'') FROM compendium_equipment WHERE id=?`, compendiumEquipmentID).Scan(
		&name, &category, &description, &weight, &sourcePage)
	if err != nil {
		return http.StatusNotFound, "compendium equipment not found"
	}
	_, err = db.DB.Exec(`INSERT INTO inventory(character_id, name, quantity, weight, category, description, notes, compendium_equipment_id)
		VALUES(?,?,?,?,?,?,'',?)`,
		charID, name, quantity, weight, category, description, compendiumEquipmentID)
	if err != nil {
		return http.StatusInternalServerError, err.Error()
	}
	return 0, ""
}

// unlinkItemRef nulls the compendium references on an inventory item, preserving its data.
func unlinkItemRef(itemID int64) (int, string) {
	var compID, entryID int64
	err := db.DB.QueryRow("SELECT COALESCE(compendium_equipment_id,0), COALESCE(compendium_entry_id,0) FROM inventory WHERE id=?", itemID).Scan(&compID, &entryID)
	if err != nil {
		return http.StatusNotFound, "item not found"
	}
	if compID == 0 && entryID == 0 {
		return http.StatusBadRequest, "item is not linked from compendium"
	}
	_, err = db.DB.Exec("UPDATE inventory SET compendium_equipment_id = NULL, compendium_entry_id = NULL WHERE id=?", itemID)
	if err != nil {
		return http.StatusInternalServerError, err.Error()
	}
	return 0, ""
}

// LinkCompendiumEquipment creates an inventory item linked to a compendium equipment entry.
// POST /api/characters/:id/inventory/link
func LinkCompendiumEquipment(c *gin.Context) {
	charID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid character id"})
		return
	}
	if !canEditCharacterID(c, charID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	quantity := 1
	if q := c.PostForm("quantity"); q != "" {
		quantity, _ = strconv.Atoi(q)
		if quantity < 1 {
			quantity = 1
		}
	}
	if entryID, ok := formLinkID(c, "compendium_entry_id"); ok {
		if st, msg := entryItemLinkInsert(charID, entryID, quantity); msg != "" {
			c.JSON(st, gin.H{"error": msg})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"status": "linked"})
		return
	}
	compendiumEquipmentID, err := strconv.ParseInt(c.PostForm("compendium_equipment_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "compendium_equipment_id required"})
		return
	}
	if st, msg := itemLinkInsert(charID, compendiumEquipmentID, quantity); msg != "" {
		c.JSON(st, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "linked"})
}

// HtmxLinkCompendiumEquipment links a compendium equipment entry and re-renders the inventory list.
// POST /htmx/compendium/equipment/link?character_id=:cid (form: compendium_equipment_id, quantity)
func HtmxLinkCompendiumEquipment(c *gin.Context) {
	charID := c.Query("character_id")
	cid, err := strconv.ParseInt(charID, 10, 64)
	if err != nil || charID == "" {
		c.String(http.StatusBadRequest, "invalid character id")
		return
	}
	if !canEditCharacterID(c, cid) {
		c.String(http.StatusForbidden, "access denied")
		return
	}
	quantity := 1
	if q := c.PostForm("quantity"); q != "" {
		quantity, _ = strconv.Atoi(q)
		if quantity < 1 {
			quantity = 1
		}
	}
	if entryID, ok := formLinkID(c, "compendium_entry_id"); ok {
		if st, msg := entryItemLinkInsert(cid, entryID, quantity); msg != "" {
			c.String(st, msg)
			return
		}
		renderHtmxInventoryList(c, charID)
		return
	}
	compID, err := strconv.ParseInt(c.PostForm("compendium_equipment_id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "compendium_equipment_id required")
		return
	}
	if st, msg := itemLinkInsert(cid, compID, quantity); msg != "" {
		c.String(st, msg)
		return
	}
	renderHtmxInventoryList(c, charID)
}

// UnlinkCompendiumEquipment unlinks an inventory item from the compendium, preserving the item data.
// DELETE /api/characters/:id/inventory/:itemId/link  (also POST /api/characters/:id/inventory/:itemId/unlink)
func UnlinkCompendiumEquipment(c *gin.Context) {
	itemID, err := strconv.ParseInt(c.Param("itemId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item id"})
		return
	}
	if !canEditResourceID(c, "inventory", itemID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	if st, msg := unlinkItemRef(itemID); msg != "" {
		c.JSON(st, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "unlinked"})
}

// HtmxUnlinkCompendiumEquipment unlinks a compendium inventory item and re-renders the inventory list.
// DELETE /htmx/inventory/:id/compendium-unlink?character_id=:cid
func HtmxUnlinkCompendiumEquipment(c *gin.Context) {
	itemID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid item id")
		return
	}
	if !canEditResourceID(c, "inventory", itemID) {
		c.String(http.StatusForbidden, "access denied")
		return
	}
	charID := c.Query("character_id")
	if charID == "" {
		c.String(http.StatusBadRequest, "character_id required")
		return
	}
	if st, msg := unlinkItemRef(itemID); msg != "" {
		c.String(st, msg)
		return
	}
	renderHtmxInventoryList(c, charID)
}

// ─── Monster Linking (One-Shot Acts/Scenes) ───

// LinkCompendiumMonsterToAct creates a monster linked to a compendium monster under an act.
// POST /api/oneshots/:adventureId/acts/:actId/monsters/link
func LinkCompendiumMonsterToAct(c *gin.Context) {
	adventureID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid adventure id"})
		return
	}
	actID, err := strconv.ParseInt(c.Param("aid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid act id"})
		return
	}

	compendiumMonsterID, err := strconv.ParseInt(c.PostForm("compendium_monster_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "compendium_monster_id required"})
		return
	}

	// Fetch compendium monster data
	var name, monsterType, size, cr, source, saves, skills, vulns, resists, immunes, condImmunes, senses, languages, abilities, actions, legendActions, description, alignment string
	var ac, hp, str, dex, con, int_, wis, cha int
	err = db.DB.QueryRow(`SELECT name, type, size, ac, hp, str, dex, con, int_, wis, cha, cr, source,
		COALESCE(saves,''), COALESCE(skills,''), COALESCE(damage_vulnerabilities,''), COALESCE(damage_resistances,''),
		COALESCE(damage_immunities,''), COALESCE(condition_immunities,''), COALESCE(senses,''), COALESCE(languages,''),
		COALESCE(special_abilities,''), COALESCE(actions,''), COALESCE(legendary_actions,''), COALESCE(description,''),
		COALESCE(alignment,'')
		FROM compendium_monsters WHERE id=?`, compendiumMonsterID).Scan(
		&name, &monsterType, &size, &ac, &hp, &str, &dex, &con, &int_, &wis, &cha, &cr, &source,
		&saves, &skills, &vulns, &resists, &immunes, &condImmunes, &senses, &languages,
		&abilities, &actions, &legendActions, &description, &alignment)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "compendium monster not found"})
		return
	}

	_, err = db.DB.Exec(`INSERT INTO oneshot_monsters(adventure_id, act_id, name, ac, hp, str, dex, con, int_, wis, cha, cr, source, is_full,
		saves, skills, damage_vulnerabilities, damage_resistances, damage_immunities, condition_immunities, senses, languages,
		special_abilities, actions, legendary_actions, compendium_monster_id)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,1,?,?,?,?,?,?,?,?,?,?,?,?)`,
		adventureID, actID, name, ac, hp, str, dex, con, int_, wis, cha, cr, source,
		saves, skills, vulns, resists, immunes, condImmunes, senses, languages,
		abilities, actions, legendActions, compendiumMonsterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "linked"})
}

// LinkCompendiumMonsterToScene creates a monster linked to a compendium monster under a scene.
// POST /api/oneshots/:adventureId/scenes/:sceneId/monsters/link
func LinkCompendiumMonsterToScene(c *gin.Context) {
	adventureID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid adventure id"})
		return
	}
	sceneID, err := strconv.ParseInt(c.Param("sid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid scene id"})
		return
	}

	compendiumMonsterID, err := strconv.ParseInt(c.PostForm("compendium_monster_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "compendium_monster_id required"})
		return
	}

	// Get adventure_id via the scene's act
	var actID int64
	db.DB.QueryRow("SELECT act_id FROM oneshot_scenes WHERE id=?", sceneID).Scan(&actID)

	// Fetch compendium monster data
	var name, monsterType, size, cr, source, saves, skills, vulns, resists, immunes, condImmunes, senses, languages, abilities, actions, legendActions, description, alignment string
	var ac, hp, str, dex, con, int_, wis, cha int
	err = db.DB.QueryRow(`SELECT name, type, size, ac, hp, str, dex, con, int_, wis, cha, cr, source,
		COALESCE(saves,''), COALESCE(skills,''), COALESCE(damage_vulnerabilities,''), COALESCE(damage_resistances,''),
		COALESCE(damage_immunities,''), COALESCE(condition_immunities,''), COALESCE(senses,''), COALESCE(languages,''),
		COALESCE(special_abilities,''), COALESCE(actions,''), COALESCE(legendary_actions,''), COALESCE(description,''),
		COALESCE(alignment,'')
		FROM compendium_monsters WHERE id=?`, compendiumMonsterID).Scan(
		&name, &monsterType, &size, &ac, &hp, &str, &dex, &con, &int_, &wis, &cha, &cr, &source,
		&saves, &skills, &vulns, &resists, &immunes, &condImmunes, &senses, &languages,
		&abilities, &actions, &legendActions, &description, &alignment)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "compendium monster not found"})
		return
	}

	_, err = db.DB.Exec(`INSERT INTO oneshot_monsters(adventure_id, act_id, scene_id, name, ac, hp, str, dex, con, int_, wis, cha, cr, source, is_full,
		saves, skills, damage_vulnerabilities, damage_resistances, damage_immunities, condition_immunities, senses, languages,
		special_abilities, actions, legendary_actions, compendium_monster_id)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,1,?,?,?,?,?,?,?,?,?,?,?,?)`,
		adventureID, actID, sceneID, name, ac, hp, str, dex, con, int_, wis, cha, cr, source,
		saves, skills, vulns, resists, immunes, condImmunes, senses, languages,
		abilities, actions, legendActions, compendiumMonsterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "linked"})
}

// UnlinkCompendiumMonster unlinks (and deletes) a monster linked from the compendium.
// DELETE /api/oneshots/monsters/:monsterId/link
func UnlinkCompendiumMonster(c *gin.Context) {
	monsterID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid monster id"})
		return
	}

	var compID int64
	err = db.DB.QueryRow("SELECT COALESCE(compendium_monster_id,0) FROM oneshot_monsters WHERE id=?", monsterID).Scan(&compID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "monster not found"})
		return
	}
	if compID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "monster is not linked from compendium"})
		return
	}

	_, err = db.DB.Exec("DELETE FROM oneshot_monsters WHERE id=?", monsterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "unlinked"})
}
