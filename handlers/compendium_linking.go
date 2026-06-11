package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
)

// ─── Spell Linking (Character Sheet) ───

// LinkCompendiumSpell creates a spell row for a character linked to a compendium spell.
// POST /api/characters/:id/spells/link
func LinkCompendiumSpell(c *gin.Context) {
	charID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid character id"})
		return
	}

	compendiumSpellID, err := strconv.ParseInt(c.PostForm("compendium_spell_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "compendium_spell_id required"})
		return
	}

	// Fetch the compendium spell data to populate the inline spell
	var name, school, castingTime, range_, components, duration, description, higherLevels, classes, sourcePage string
	var level int
	err = db.DB.QueryRow(`SELECT name, level, school, casting_time, "range", components, duration, description, COALESCE(higher_levels,''), COALESCE(classes,'[]'), COALESCE(source_page,'') FROM compendium_spells WHERE id=?`, compendiumSpellID).Scan(
		&name, &level, &school, &castingTime, &range_, &components, &duration, &description, &higherLevels, &classes, &sourcePage)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "compendium spell not found"})
		return
	}

	_, err = db.DB.Exec(`INSERT INTO spells(character_id, name, level, school, casting_time, "range", components, duration, description, source, notes, compendium_spell_id)
		VALUES(?,?,?,?,?,?,?,?,?,?,'',?)`,
		charID, name, level, school, castingTime, range_, components, duration, description, sourcePage, compendiumSpellID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "linked"})
}

// UnlinkCompendiumSpell unlinks (and deletes) a spell that was linked from the compendium.
// DELETE /api/characters/:id/spells/:spellId/link
func UnlinkCompendiumSpell(c *gin.Context) {
	spellID, err := strconv.ParseInt(c.Param("spellId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid spell id"})
		return
	}

	// Only delete if linked — prevent deleting inline-created spells
	var compID int64
	err = db.DB.QueryRow("SELECT COALESCE(compendium_spell_id,0) FROM spells WHERE id=?", spellID).Scan(&compID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "spell not found"})
		return
	}
	if compID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "spell is not linked from compendium"})
		return
	}

	_, err = db.DB.Exec("DELETE FROM spells WHERE id=?", spellID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "unlinked"})
}

// ─── Inventory/Equipment Linking (Character Sheet) ───

// LinkCompendiumEquipment creates an inventory item linked to a compendium equipment entry.
// POST /api/characters/:id/inventory/link
func LinkCompendiumEquipment(c *gin.Context) {
	charID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid character id"})
		return
	}

	compendiumEquipmentID, err := strconv.ParseInt(c.PostForm("compendium_equipment_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "compendium_equipment_id required"})
		return
	}

	quantity := 1
	if q := c.PostForm("quantity"); q != "" {
		quantity, _ = strconv.Atoi(q)
		if quantity < 1 {
			quantity = 1
		}
	}

	// Fetch compendium equipment data
	var name, category, description, sourcePage string
	var weight float64
	err = db.DB.QueryRow(`SELECT name, COALESCE(category,''), COALESCE(description,''), COALESCE(weight,0), COALESCE(source_page,'') FROM compendium_equipment WHERE id=?`, compendiumEquipmentID).Scan(
		&name, &category, &description, &weight, &sourcePage)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "compendium equipment not found"})
		return
	}

	_, err = db.DB.Exec(`INSERT INTO inventory(character_id, name, quantity, weight, category, description, notes, compendium_equipment_id)
		VALUES(?,?,?,?,?,?,'',?)`,
		charID, name, quantity, weight, category, description, compendiumEquipmentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "linked"})
}

// UnlinkCompendiumEquipment unlinks (and deletes) an inventory item linked from the compendium.
// DELETE /api/characters/:id/inventory/:itemId/link
func UnlinkCompendiumEquipment(c *gin.Context) {
	itemID, err := strconv.ParseInt(c.Param("itemId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item id"})
		return
	}

	var compID int64
	err = db.DB.QueryRow("SELECT COALESCE(compendium_equipment_id,0) FROM inventory WHERE id=?", itemID).Scan(&compID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}
	if compID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "item is not linked from compendium"})
		return
	}

	_, err = db.DB.Exec("DELETE FROM inventory WHERE id=?", itemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "unlinked"})
}

// ─── Monster Linking (One-Shot Acts/Scenes) ───

// LinkCompendiumMonsterToAct creates a monster linked to a compendium monster under an act.
// POST /api/oneshots/:adventureId/acts/:actId/monsters/link
func LinkCompendiumMonsterToAct(c *gin.Context) {
	adventureID, err := strconv.ParseInt(c.Param("adventureId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid adventure id"})
		return
	}
	actID, err := strconv.ParseInt(c.Param("actId"), 10, 64)
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
	adventureID, err := strconv.ParseInt(c.Param("adventureId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid adventure id"})
		return
	}
	sceneID, err := strconv.ParseInt(c.Param("sceneId"), 10, 64)
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
	monsterID, err := strconv.ParseInt(c.Param("monsterId"), 10, 64)
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
