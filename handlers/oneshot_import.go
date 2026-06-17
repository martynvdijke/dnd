package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

func ImportCompendiumMonsterToOneShot(c *gin.Context) {
	var req struct {
		CompendiumMonsterID int64  `json:"compendium_monster_id"`
		AdventureID         int64  `json:"adventure_id"`
		ActID               *int64 `json:"act_id,omitempty"`
		SceneID             *int64 `json:"scene_id,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Fetch compendium monster
	var cm models.CompendiumMonster
	var isFull int
	err := db.DB.QueryRow("SELECT id,name,type,size,ac,hp,str,dex,con,int_,wis,cha,cr,source,is_full,saves,skills,damage_vulnerabilities,damage_resistances,damage_immunities,condition_immunities,senses,languages,special_abilities,actions,legendary_actions,description FROM compendium_monsters WHERE id=?", req.CompendiumMonsterID).
		Scan(&cm.ID, &cm.Name, &cm.Type, &cm.Size, &cm.AC, &cm.HP,
			&cm.Str, &cm.Dex, &cm.Con, &cm.Int, &cm.Wis, &cm.Cha,
			&cm.CR, &cm.Source, &isFull,
			&cm.Saves, &cm.Skills, &cm.DamageVulnerabilities, &cm.DamageResistances, &cm.DamageImmunities, &cm.ConditionImmunities,
			&cm.Senses, &cm.Languages, &cm.SpecialAbilities, &cm.Actions, &cm.LegendaryActions, &cm.Description)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "compendium monster not found"})
		return
	}

	isFullVal := 0
	if isFull == 1 {
		isFullVal = 1
	}

	result, err := db.DB.Exec(`INSERT INTO oneshot_monsters(adventure_id, act_id, scene_id, name, ac, hp, str, dex, con, int_, wis, cha, cr, source, is_full,
		saves, skills, damage_vulnerabilities, damage_resistances, damage_immunities, condition_immunities, senses, languages,
		special_abilities, actions, legendary_actions, compendium_monster_id) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		req.AdventureID, req.ActID, req.SceneID, cm.Name, cm.AC, cm.HP, cm.Str, cm.Dex, cm.Con, cm.Int, cm.Wis, cm.Cha,
		cm.CR, cm.Source, isFullVal,
		cm.Saves, cm.Skills, cm.DamageVulnerabilities, cm.DamageResistances, cm.DamageImmunities, cm.ConditionImmunities,
		cm.Senses, cm.Languages, cm.SpecialAbilities, cm.Actions, cm.LegendaryActions, req.CompendiumMonsterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func ImportCompendiumMonsterToEncounter(c *gin.Context) {
	encounterID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		CompendiumMonsterID int64 `json:"compendium_monster_id"`
		Count               int   `json:"count"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Count < 1 {
		req.Count = 1
	}

	// Fetch compendium monster for stat block
	var cm models.CompendiumMonster
	var isFull int
	err := db.DB.QueryRow("SELECT id,name,ac,hp,cr,source,description FROM compendium_monsters WHERE id=?", req.CompendiumMonsterID).
		Scan(&cm.ID, &cm.Name, &cm.AC, &cm.HP, &cm.CR, &cm.Source, &cm.Description)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "compendium monster not found"})
		return
	}
	_ = isFull

	result, err := db.DB.Exec(`INSERT INTO encounter_monsters(encounter_id, name, count, cr, ac, hp, source, notes, compendium_monster_id) VALUES(?,?,?,?,?,?,?,?,?)`,
		encounterID, cm.Name, req.Count, cm.CR, cm.AC, cm.HP, cm.Source, "", req.CompendiumMonsterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func ImportLibraryMonsterToOneShot(c *gin.Context) {
	adventureID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req struct {
		LibraryID int64  `json:"library_id"`
		ActID     *int64 `json:"act_id,omitempty"`
		SceneID   *int64 `json:"scene_id,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Return pre-filled data as JSON for the form
	var m models.MonsterLibraryEntry
	var isFull int
	err := db.DB.QueryRow(`SELECT id, user_id, name, ac, hp, str, dex, con, int_, wis, cha, cr, source, is_full,
		saves, skills, damage_vulnerabilities, damage_resistances, damage_immunities, condition_immunities, senses, languages,
		special_abilities, actions, legendary_actions, description FROM monster_library WHERE id=?`, req.LibraryID).
		Scan(&m.ID, &m.UserID, &m.Name, &m.AC, &m.HP, &m.Str, &m.Dex, &m.Con, &m.Int, &m.Wis, &m.Cha,
			&m.CR, &m.Source, &isFull,
			&m.Saves, &m.Skills, &m.DamageVulnerabilities, &m.DamageResistances, &m.DamageImmunities, &m.ConditionImmunities,
			&m.Senses, &m.Languages, &m.SpecialAbilities, &m.Actions, &m.LegendaryActions, &m.Description)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "library monster not found"})
		return
	}
	m.IsFull = isFull == 1

	c.JSON(http.StatusOK, gin.H{
		"adventure_id": adventureID,
		"act_id":       req.ActID,
		"scene_id":     req.SceneID,
		"monster":      m,
	})
}

// ImportCompendiumEntryToEncounter imports a compendium entry (schema-based) to an encounter.
func ImportCompendiumEntryToEncounter(c *gin.Context) {
	encounterID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		CompendiumEntryID int64 `json:"compendium_entry_id"`
		Count             int   `json:"count"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Count < 1 {
		req.Count = 1
	}
	var dataJSON string
	err := db.DB.QueryRow("SELECT data FROM compendium_entries WHERE id=?", req.CompendiumEntryID).Scan(&dataJSON)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "compendium entry not found"})
		return
	}
	name := "Unknown"
	var data map[string]any
	if json.Unmarshal([]byte(dataJSON), &data) == nil {
		if n, ok := data["name"].(string); ok && n != "" {
			name = n
		}
	}
	result, err := db.DB.Exec(`INSERT INTO encounter_monsters(encounter_id, name, count, cr, ac, hp, source, notes, compendium_entry_id) VALUES(?,?,?,?,?,?,?,?,?)`,
		encounterID, name, req.Count, "0", 10, 1, "compendium", "", req.CompendiumEntryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// ImportCompendiumEntryToOneShot imports a compendium entry to a one-shot adventure.
func ImportCompendiumEntryToOneShot(c *gin.Context) {
	var req struct {
		CompendiumEntryID int64  `json:"compendium_entry_id"`
		AdventureID       int64  `json:"adventure_id"`
		ActID             *int64 `json:"act_id,omitempty"`
		SceneID           *int64 `json:"scene_id,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var dataJSON string
	err := db.DB.QueryRow("SELECT data FROM compendium_entries WHERE id=?", req.CompendiumEntryID).Scan(&dataJSON)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "compendium entry not found"})
		return
	}
	name := "Unknown"
	var data map[string]any
	if json.Unmarshal([]byte(dataJSON), &data) == nil {
		if n, ok := data["name"].(string); ok && n != "" {
			name = n
		}
	}
	result, err := db.DB.Exec(`INSERT INTO oneshot_monsters(adventure_id, act_id, scene_id, name, ac, hp, cr, source, compendium_entry_id) VALUES(?,?,?,?,?,?,?,?,?)`,
		req.AdventureID, req.ActID, req.SceneID, name, 10, 1, "0", "compendium", req.CompendiumEntryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// ImportCompendiumEquipmentToOneShot imports compendium equipment to a one-shot item.
func ImportCompendiumEquipmentToOneShot(c *gin.Context) {
	var req struct {
		CompendiumEquipmentID int64  `json:"compendium_equipment_id"`
		AdventureID           int64  `json:"adventure_id"`
		ActID                 *int64 `json:"act_id,omitempty"`
		SceneID               *int64 `json:"scene_id,omitempty"`
		Quantity              int    `json:"quantity"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Quantity < 1 {
		req.Quantity = 1
	}

	var e models.CompendiumEquipment
	err := db.DB.QueryRow(`SELECT id, name, category, cost, weight, description, source_page,
		COALESCE(system,''), COALESCE(source,''), COALESCE(item_type,''), COALESCE(item_rarity,''), COALESCE(publisher,'')
		FROM compendium_equipment WHERE id=?`, req.CompendiumEquipmentID).
		Scan(&e.ID, &e.Name, &e.Category, &e.Cost, &e.Weight, &e.Description, &e.SourcePage,
			&e.System, &e.Source, &e.ItemType, &e.ItemRarity, &e.Publisher)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "compendium equipment not found"})
		return
	}

	priceGP := 0.0
	if e.Cost != "" {
		parts := strings.Fields(e.Cost)
		if len(parts) > 0 {
			if val, err := strconv.ParseFloat(parts[0], 64); err == nil {
				priceGP = val
				if len(parts) > 1 {
					switch strings.ToLower(parts[1]) {
					case "cp":
						priceGP = priceGP / 100
					case "sp":
						priceGP = priceGP / 10
					case "ep":
						priceGP = priceGP * 2
					case "pp":
						priceGP = priceGP * 10
					}
				}
			}
		}
	}

	isMagical := strings.EqualFold(e.ItemRarity, "rare") || strings.EqualFold(e.ItemRarity, "very rare") || strings.EqualFold(e.ItemRarity, "legendary") || strings.EqualFold(e.ItemRarity, "artifact")

	result, err := db.DB.Exec(`INSERT INTO oneshot_items(adventure_id, act_id, scene_id, name, description, category, quantity, weight, price_gp, is_magical, compendium_equipment_id) VALUES(?,?,?,?,?,?,?,?,?,?,?)`,
		req.AdventureID, req.ActID, req.SceneID, e.Name, e.Description, e.Category, req.Quantity, e.Weight, priceGP, isMagical, req.CompendiumEquipmentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}
