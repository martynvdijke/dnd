package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/middleware"
	"villum/models"
)

// ─── Monster Library ───

func ListMonsterLibrary(c *gin.Context) {
	userID, _ := c.Get("user_id")
	rows, err := db.DB.Query("SELECT id, user_id, name, ac, hp, str, dex, con, int_, wis, cha, cr, source, is_full, saves, skills, damage_vulnerabilities, damage_resistances, damage_immunities, condition_immunities, senses, languages, special_abilities, actions, legendary_actions, description, created_at FROM monster_library WHERE user_id=? ORDER BY name", userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := make([]models.MonsterLibraryEntry, 0)
	for rows.Next() {
		var m models.MonsterLibraryEntry
		var isFull int
		if err := rows.Scan(&m.ID, &m.UserID, &m.Name, &m.AC, &m.HP, &m.Str, &m.Dex, &m.Con, &m.Int, &m.Wis, &m.Cha, &m.CR, &m.Source, &isFull,
			&m.Saves, &m.Skills, &m.DamageVulnerabilities, &m.DamageResistances, &m.DamageImmunities, &m.ConditionImmunities,
			&m.Senses, &m.Languages, &m.SpecialAbilities, &m.Actions, &m.LegendaryActions, &m.Description, &m.CreatedAt); err != nil {
			middleware.LogWarn("oneshot", "scan failed, skipping library monster", "error", err)
			continue
		}
		m.IsFull = isFull == 1
		out = append(out, m)
	}
	c.JSON(http.StatusOK, out)
}

func CreateMonsterLibraryEntry(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var m models.MonsterLibraryEntry
	if err := c.ShouldBindJSON(&m); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	isFull := 0
	if m.IsFull {
		isFull = 1
	}
	result, err := db.DB.Exec(`INSERT INTO monster_library(user_id, name, ac, hp, str, dex, con, int_, wis, cha, cr, source, is_full,
		saves, skills, damage_vulnerabilities, damage_resistances, damage_immunities, condition_immunities, senses, languages,
		special_abilities, actions, legendary_actions, description) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		userID, m.Name, m.AC, m.HP, m.Str, m.Dex, m.Con, m.Int, m.Wis, m.Cha, m.CR, m.Source, isFull,
		m.Saves, m.Skills, m.DamageVulnerabilities, m.DamageResistances, m.DamageImmunities, m.ConditionImmunities,
		m.Senses, m.Languages, m.SpecialAbilities, m.Actions, m.LegendaryActions, m.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	db.DB.QueryRow("SELECT id, user_id, name, ac, hp, str, dex, con, int_, wis, cha, cr, source, is_full, saves, skills, damage_vulnerabilities, damage_resistances, damage_immunities, condition_immunities, senses, languages, special_abilities, actions, legendary_actions, description, created_at FROM monster_library WHERE id=?", id).Scan(
		&m.ID, &m.UserID, &m.Name, &m.AC, &m.HP, &m.Str, &m.Dex, &m.Con, &m.Int, &m.Wis, &m.Cha, &m.CR, &m.Source, &isFull,
		&m.Saves, &m.Skills, &m.DamageVulnerabilities, &m.DamageResistances, &m.DamageImmunities, &m.ConditionImmunities,
		&m.Senses, &m.Languages, &m.SpecialAbilities, &m.Actions, &m.LegendaryActions, &m.Description, &m.CreatedAt)
	m.IsFull = isFull == 1
	c.JSON(http.StatusCreated, m)
}

func UpdateMonsterLibraryEntry(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var m models.MonsterLibraryEntry
	if err := c.ShouldBindJSON(&m); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	isFull := 0
	if m.IsFull {
		isFull = 1
	}
	_, err := db.DB.Exec(`UPDATE monster_library SET name=?, ac=?, hp=?, str=?, dex=?, con=?, int_=?, wis=?, cha=?, cr=?, source=?, is_full=?,
		saves=?, skills=?, damage_vulnerabilities=?, damage_resistances=?, damage_immunities=?, condition_immunities=?, senses=?, languages=?,
		special_abilities=?, actions=?, legendary_actions=?, description=? WHERE id=?`,
		m.Name, m.AC, m.HP, m.Str, m.Dex, m.Con, m.Int, m.Wis, m.Cha, m.CR, m.Source, isFull,
		m.Saves, m.Skills, m.DamageVulnerabilities, m.DamageResistances, m.DamageImmunities, m.ConditionImmunities,
		m.Senses, m.Languages, m.SpecialAbilities, m.Actions, m.LegendaryActions, m.Description, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteMonsterLibraryEntry(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	db.DB.Exec("DELETE FROM monster_library WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ─── Act/Scene Monsters ───

func scanMonster(rows interface{ Scan(...any) error }) (models.OneShotMonster, error) {
	var m models.OneShotMonster
	var isFull int
	err := rows.Scan(&m.ID, &m.AdventureID, &m.ActID, &m.SceneID, &m.Name, &m.AC, &m.HP,
		&m.Str, &m.Dex, &m.Con, &m.Int, &m.Wis, &m.Cha, &m.CR, &m.Source, &isFull,
		&m.Saves, &m.Skills, &m.DamageVulnerabilities, &m.DamageResistances, &m.DamageImmunities, &m.ConditionImmunities,
		&m.Senses, &m.Languages, &m.SpecialAbilities, &m.Actions, &m.LegendaryActions, &m.LibraryID, &m.CompendiumMonsterID, &m.CreatedAt)
	if err != nil {
		return m, err
	}
	m.IsFull = isFull == 1
	return m, nil
}

const monsterColumns = "id, adventure_id, act_id, scene_id, name, ac, hp, str, dex, con, int_, wis, cha, cr, source, is_full, saves, skills, damage_vulnerabilities, damage_resistances, damage_immunities, condition_immunities, senses, languages, special_abilities, actions, legendary_actions, library_id, compendium_monster_id, created_at"

func ListActMonsters(c *gin.Context) {
	actID := c.Param("id")
	rows, err := db.DB.Query("SELECT "+monsterColumns+" FROM oneshot_monsters WHERE act_id=? ORDER BY name", actID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := make([]models.OneShotMonster, 0)
	for rows.Next() {
		m, err := scanMonster(rows)
		if err != nil {
			middleware.LogWarn("oneshot", "scan failed, skipping act monster", "error", err)
			continue
		}
		out = append(out, m)
	}
	c.JSON(http.StatusOK, out)
}

func CreateActMonster(c *gin.Context) {
	actID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		AdventureID           int64  `json:"adventure_id"`
		LibraryID             *int64 `json:"library_id,omitempty"`
		Name                  string `json:"name"`
		AC                    int    `json:"ac"`
		HP                    int    `json:"hp"`
		Str                   int    `json:"str"`
		Dex                   int    `json:"dex"`
		Con                   int    `json:"con"`
		Int                   int    `json:"int"`
		Wis                   int    `json:"wis"`
		Cha                   int    `json:"cha"`
		CR                    string `json:"cr"`
		Source                string `json:"source"`
		IsFull                bool   `json:"is_full"`
		Saves                 string `json:"saves"`
		Skills                string `json:"skills"`
		DamageVulnerabilities string `json:"damage_vulnerabilities"`
		DamageResistances     string `json:"damage_resistances"`
		DamageImmunities      string `json:"damage_immunities"`
		ConditionImmunities   string `json:"condition_immunities"`
		Senses                string `json:"senses"`
		Languages             string `json:"languages"`
		SpecialAbilities      string `json:"special_abilities"`
		Actions               string `json:"actions"`
		LegendaryActions      string `json:"legendary_actions"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	isFull := 0
	if req.IsFull {
		isFull = 1
	}
	adventureID := req.AdventureID
	if adventureID == 0 {
		db.DB.QueryRow("SELECT adventure_id FROM oneshot_acts WHERE id=?", actID).Scan(&adventureID)
	}
	result, err := db.DB.Exec(`INSERT INTO oneshot_monsters(adventure_id, act_id, name, ac, hp, str, dex, con, int_, wis, cha, cr, source, is_full,
		saves, skills, damage_vulnerabilities, damage_resistances, damage_immunities, condition_immunities, senses, languages,
		special_abilities, actions, legendary_actions, library_id) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		adventureID, actID, req.Name, req.AC, req.HP, req.Str, req.Dex, req.Con, req.Int, req.Wis, req.Cha, req.CR, req.Source, isFull,
		req.Saves, req.Skills, req.DamageVulnerabilities, req.DamageResistances, req.DamageImmunities, req.ConditionImmunities,
		req.Senses, req.Languages, req.SpecialAbilities, req.Actions, req.LegendaryActions, req.LibraryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	var m models.OneShotMonster
	rows := db.DB.QueryRow("SELECT "+monsterColumns+" FROM oneshot_monsters WHERE id=?", id)
	m, err = scanMonster(rows)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, m)
}

func UpdateOneShotMonster(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		Name                  string `json:"name"`
		AC                    int    `json:"ac"`
		HP                    int    `json:"hp"`
		Str                   int    `json:"str"`
		Dex                   int    `json:"dex"`
		Con                   int    `json:"con"`
		Int                   int    `json:"int"`
		Wis                   int    `json:"wis"`
		Cha                   int    `json:"cha"`
		CR                    string `json:"cr"`
		Source                string `json:"source"`
		IsFull                bool   `json:"is_full"`
		Saves                 string `json:"saves"`
		Skills                string `json:"skills"`
		DamageVulnerabilities string `json:"damage_vulnerabilities"`
		DamageResistances     string `json:"damage_resistances"`
		DamageImmunities      string `json:"damage_immunities"`
		ConditionImmunities   string `json:"condition_immunities"`
		Senses                string `json:"senses"`
		Languages             string `json:"languages"`
		SpecialAbilities      string `json:"special_abilities"`
		Actions               string `json:"actions"`
		LegendaryActions      string `json:"legendary_actions"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	isFull := 0
	if req.IsFull {
		isFull = 1
	}
	_, err := db.DB.Exec(`UPDATE oneshot_monsters SET name=?, ac=?, hp=?, str=?, dex=?, con=?, int_=?, wis=?, cha=?, cr=?, source=?, is_full=?,
		saves=?, skills=?, damage_vulnerabilities=?, damage_resistances=?, damage_immunities=?, condition_immunities=?, senses=?, languages=?,
		special_abilities=?, actions=?, legendary_actions=? WHERE id=?`,
		req.Name, req.AC, req.HP, req.Str, req.Dex, req.Con, req.Int, req.Wis, req.Cha, req.CR, req.Source, isFull,
		req.Saves, req.Skills, req.DamageVulnerabilities, req.DamageResistances, req.DamageImmunities, req.ConditionImmunities,
		req.Senses, req.Languages, req.SpecialAbilities, req.Actions, req.LegendaryActions, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteOneShotMonster(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	db.DB.Exec("DELETE FROM oneshot_monsters WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func ListSceneMonsters(c *gin.Context) {
	sceneID := c.Param("id")
	rows, err := db.DB.Query("SELECT "+monsterColumns+" FROM oneshot_monsters WHERE scene_id=? ORDER BY name", sceneID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := make([]models.OneShotMonster, 0)
	for rows.Next() {
		m, err := scanMonster(rows)
		if err != nil {
			middleware.LogWarn("oneshot", "scan failed, skipping scene monster", "error", err)
			continue
		}
		out = append(out, m)
	}
	c.JSON(http.StatusOK, out)
}

func CreateSceneMonster(c *gin.Context) {
	sceneID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		AdventureID           int64  `json:"adventure_id"`
		LibraryID             *int64 `json:"library_id,omitempty"`
		Name                  string `json:"name"`
		AC                    int    `json:"ac"`
		HP                    int    `json:"hp"`
		Str                   int    `json:"str"`
		Dex                   int    `json:"dex"`
		Con                   int    `json:"con"`
		Int                   int    `json:"int"`
		Wis                   int    `json:"wis"`
		Cha                   int    `json:"cha"`
		CR                    string `json:"cr"`
		Source                string `json:"source"`
		IsFull                bool   `json:"is_full"`
		Saves                 string `json:"saves"`
		Skills                string `json:"skills"`
		DamageVulnerabilities string `json:"damage_vulnerabilities"`
		DamageResistances     string `json:"damage_resistances"`
		DamageImmunities      string `json:"damage_immunities"`
		ConditionImmunities   string `json:"condition_immunities"`
		Senses                string `json:"senses"`
		Languages             string `json:"languages"`
		SpecialAbilities      string `json:"special_abilities"`
		Actions               string `json:"actions"`
		LegendaryActions      string `json:"legendary_actions"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	isFull := 0
	if req.IsFull {
		isFull = 1
	}
	result, err := db.DB.Exec(`INSERT INTO oneshot_monsters(adventure_id, scene_id, name, ac, hp, str, dex, con, int_, wis, cha, cr, source, is_full,
		saves, skills, damage_vulnerabilities, damage_resistances, damage_immunities, condition_immunities, senses, languages,
		special_abilities, actions, legendary_actions, library_id) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		req.AdventureID, sceneID, req.Name, req.AC, req.HP, req.Str, req.Dex, req.Con, req.Int, req.Wis, req.Cha, req.CR, req.Source, isFull,
		req.Saves, req.Skills, req.DamageVulnerabilities, req.DamageResistances, req.DamageImmunities, req.ConditionImmunities,
		req.Senses, req.Languages, req.SpecialAbilities, req.Actions, req.LegendaryActions, req.LibraryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}
