package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
	"villum/db"
	"villum/models"
)

func HtmxListMonsterLibrary(c *gin.Context) {
	userID, _ := c.Get("user_id")
	rows, err := db.DB.Query(`SELECT id, user_id, name, ac, hp, str, dex, con, int_, wis, cha, cr, source, is_full,
		saves, skills, damage_vulnerabilities, damage_resistances, damage_immunities, condition_immunities,
		senses, languages, special_abilities, actions, legendary_actions, description, created_at
		FROM monster_library WHERE user_id=? ORDER BY name`, userID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	var monsters []models.MonsterLibraryEntry
	for rows.Next() {
		var m models.MonsterLibraryEntry
		var isFull int
		rows.Scan(&m.ID, &m.UserID, &m.Name, &m.AC, &m.HP, &m.Str, &m.Dex, &m.Con, &m.Int, &m.Wis, &m.Cha,
			&m.CR, &m.Source, &isFull,
			&m.Saves, &m.Skills, &m.DamageVulnerabilities, &m.DamageResistances, &m.DamageImmunities, &m.ConditionImmunities,
			&m.Senses, &m.Languages, &m.SpecialAbilities, &m.Actions, &m.LegendaryActions, &m.Description, &m.CreatedAt)
		m.IsFull = isFull == 1
		monsters = append(monsters, m)
	}
	renderTemplate(c, "monster_library_list", htmxMonsterLibraryData{Monsters: monsters})
}

func HtmxMonsterLibraryForm(c *gin.Context) {
	compID := c.Query("compendium_id")
	editID := c.Query("edit_id")
	userID, _ := c.Get("user_id")

	if editID != "" {
		// Edit existing library entry
		var m models.MonsterLibraryEntry
		var isFull int
		err := db.DB.QueryRow(`SELECT id, user_id, name, ac, hp, str, dex, con, int_, wis, cha, cr, source, is_full,
			saves, skills, damage_vulnerabilities, damage_resistances, damage_immunities, condition_immunities,
			senses, languages, special_abilities, actions, legendary_actions, description, created_at
			FROM monster_library WHERE id=? AND user_id=?`, editID, userID).Scan(
			&m.ID, &m.UserID, &m.Name, &m.AC, &m.HP, &m.Str, &m.Dex, &m.Con, &m.Int, &m.Wis, &m.Cha,
			&m.CR, &m.Source, &isFull,
			&m.Saves, &m.Skills, &m.DamageVulnerabilities, &m.DamageResistances, &m.DamageImmunities, &m.ConditionImmunities,
			&m.Senses, &m.Languages, &m.SpecialAbilities, &m.Actions, &m.LegendaryActions, &m.Description, &m.CreatedAt)
		if err != nil {
			c.String(http.StatusNotFound, "not found")
			return
		}
		m.IsFull = isFull == 1
		renderTemplate(c, "monster_library_form", htmxMonsterLibraryData{Monster: &m})
		return
	}

	if compID != "" {
		// Clone from compendium
		var cm models.CompendiumMonster
		err := db.DB.QueryRow(`SELECT id,name,type,size,ac,hp,str,dex,con,int_,wis,cha,cr,source,is_full,
			saves,skills,damage_vulnerabilities,damage_resistances,damage_immunities,condition_immunities,
			senses,languages,special_abilities,actions,legendary_actions,description
			FROM compendium_monsters WHERE id=?`, compID).Scan(
			&cm.ID, &cm.Name, &cm.Type, &cm.Size, &cm.AC, &cm.HP, &cm.Str, &cm.Dex, &cm.Con, &cm.Int, &cm.Wis, &cm.Cha,
			&cm.CR, &cm.Source, &cm.IsFull,
			&cm.Saves, &cm.Skills, &cm.DamageVulnerabilities, &cm.DamageResistances, &cm.DamageImmunities, &cm.ConditionImmunities,
			&cm.Senses, &cm.Languages, &cm.SpecialAbilities, &cm.Actions, &cm.LegendaryActions, &cm.Description)
		if err != nil {
			c.String(http.StatusNotFound, "compendium monster not found")
			return
		}
		renderTemplate(c, "monster_library_form", htmxMonsterLibraryData{CompMonster: &cm})
		return
	}

	// New empty form
	renderTemplate(c, "monster_library_form", htmxMonsterLibraryData{})
}

func HtmxCreateMonsterLibrary(c *gin.Context) {
	userID, _ := c.Get("user_id")
	name := strings.TrimSpace(c.PostForm("name"))
	if name == "" {
		c.String(http.StatusBadRequest, "name is required")
		return
	}
	isFull := 0
	if c.PostForm("is_full") == "1" {
		isFull = 1
	}
	_, err := db.DB.Exec(`INSERT INTO monster_library(user_id, name, ac, hp, str, dex, con, int_, wis, cha, cr, source, is_full,
		saves, skills, damage_vulnerabilities, damage_resistances, damage_immunities, condition_immunities, senses, languages,
		special_abilities, actions, legendary_actions, description) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		userID, name, getIntParam(c, "ac", 10), getIntParam(c, "hp", 1),
		getIntParam(c, "str", 10), getIntParam(c, "dex", 10), getIntParam(c, "con", 10),
		getIntParam(c, "int", 10), getIntParam(c, "wis", 10), getIntParam(c, "cha", 10),
		c.PostForm("cr"), "homebrew", isFull,
		c.PostForm("saves"), c.PostForm("skills"),
		c.PostForm("damage_vulnerabilities"), c.PostForm("damage_resistances"), c.PostForm("damage_immunities"),
		c.PostForm("condition_immunities"), c.PostForm("senses"), c.PostForm("languages"),
		c.PostForm("special_abilities"), c.PostForm("actions"), c.PostForm("legendary_actions"),
		c.PostForm("description"))
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	// Refresh library list
	HtmxListMonsterLibrary(c)
}

func HtmxUpdateMonsterLibrary(c *gin.Context) {
	userID, _ := c.Get("user_id")
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	name := strings.TrimSpace(c.PostForm("name"))
	if name == "" {
		c.String(http.StatusBadRequest, "name is required")
		return
	}
	isFull := 0
	if c.PostForm("is_full") == "1" {
		isFull = 1
	}
	_, err := db.DB.Exec(`UPDATE monster_library SET name=?, ac=?, hp=?, str=?, dex=?, con=?, int_=?, wis=?, cha=?, cr=?, source=?, is_full=?,
		saves=?, skills=?, damage_vulnerabilities=?, damage_resistances=?, damage_immunities=?, condition_immunities=?, senses=?, languages=?,
		special_abilities=?, actions=?, legendary_actions=?, description=? WHERE id=? AND user_id=?`,
		name, getIntParam(c, "ac", 10), getIntParam(c, "hp", 1),
		getIntParam(c, "str", 10), getIntParam(c, "dex", 10), getIntParam(c, "con", 10),
		getIntParam(c, "int", 10), getIntParam(c, "wis", 10), getIntParam(c, "cha", 10),
		c.PostForm("cr"), "homebrew", isFull,
		c.PostForm("saves"), c.PostForm("skills"),
		c.PostForm("damage_vulnerabilities"), c.PostForm("damage_resistances"), c.PostForm("damage_immunities"),
		c.PostForm("condition_immunities"), c.PostForm("senses"), c.PostForm("languages"),
		c.PostForm("special_abilities"), c.PostForm("actions"), c.PostForm("legendary_actions"),
		c.PostForm("description"), id, userID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	HtmxListMonsterLibrary(c)
}

func HtmxDeleteMonsterLibrary(c *gin.Context) {
	userID, _ := c.Get("user_id")
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var ownerID int64
	err := db.DB.QueryRow("SELECT user_id FROM monster_library WHERE id=?", id).Scan(&ownerID)
	if err != nil || ownerID != userID {
		c.String(http.StatusNotFound, "not found")
		return
	}
	db.DB.Exec("DELETE FROM monster_library WHERE id=?", id)
	HtmxListMonsterLibrary(c)
}

func HtmxMonsterLibrarySection(c *gin.Context) {
	userID, _ := c.Get("user_id")
	rows, err := db.DB.Query(`SELECT id, user_id, name, ac, hp, str, dex, con, int_, wis, cha, cr, source, is_full,
		saves, skills, damage_vulnerabilities, damage_resistances, damage_immunities, condition_immunities,
		senses, languages, special_abilities, actions, legendary_actions, description, created_at
		FROM monster_library WHERE user_id=? ORDER BY name`, userID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	var monsters []models.MonsterLibraryEntry
	for rows.Next() {
		var m models.MonsterLibraryEntry
		var isFull int
		rows.Scan(&m.ID, &m.UserID, &m.Name, &m.AC, &m.HP, &m.Str, &m.Dex, &m.Con, &m.Int, &m.Wis, &m.Cha,
			&m.CR, &m.Source, &isFull,
			&m.Saves, &m.Skills, &m.DamageVulnerabilities, &m.DamageResistances, &m.DamageImmunities, &m.ConditionImmunities,
			&m.Senses, &m.Languages, &m.SpecialAbilities, &m.Actions, &m.LegendaryActions, &m.Description, &m.CreatedAt)
		m.IsFull = isFull == 1
		monsters = append(monsters, m)
	}
	renderTemplate(c, "monster_library_section", htmxMonsterLibraryData{Monsters: monsters})
}

// ─── Compendium Spell Picker (HTMX) ───

type htmxCompendiumSpellPickerData struct {
	CharacterID int64
	Query       string
	Class       string
	Level       string
	Classes     []string
	Spells      []compendiumSpellPickerItem
	Page        int
	PageSize    int
	TotalCount  int
	TotalPages  int
}

func HtmxCompendiumSpellPicker(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Query("character_id"), 10, 64)
	q := strings.TrimSpace(c.Query("q"))
	class := strings.TrimSpace(c.Query("class"))
	level := strings.TrimSpace(c.Query("level"))
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	data := htmxCompendiumSpellPickerData{
		CharacterID: charID,
		Query:       q,
		Class:       class,
		Level:       level,
		Classes:     getDistinctSpellClasses(),
		Page:        page,
		PageSize:    pageSize,
	}

	if q != "" || class != "" || level != "" {
		spells, totalCount := queryCompendiumSpellsUnion(q, class, level, pageSize, offset)
		data.Spells = spells
		data.TotalCount = totalCount
		data.TotalPages = (totalCount + pageSize - 1) / pageSize
		if data.TotalPages < 1 {
			data.TotalPages = 1
		}
	}

	renderTemplate(c, "compendium_spell_picker", data)
}

// ─── Compendium Equipment Picker (HTMX) ───

type htmxCompendiumEquipmentPickerData struct {
	CharacterID int64
	Query       string
	Items       []compendiumEquipmentPickerItem
	Page        int
	PageSize    int
	TotalCount  int
	TotalPages  int
}

func HtmxCompendiumEquipmentPicker(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Query("character_id"), 10, 64)
	q := strings.TrimSpace(c.Query("q"))
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	data := htmxCompendiumEquipmentPickerData{
		CharacterID: charID,
		Query:       q,
		Page:        page,
		PageSize:    pageSize,
	}

	if q != "" {
		items, totalCount := queryCompendiumEquipmentUnion(q, pageSize, offset)
		data.Items = items
		data.TotalCount = totalCount
		data.TotalPages = (totalCount + pageSize - 1) / pageSize
		if data.TotalPages < 1 {
			data.TotalPages = 1
		}
	}

	renderTemplate(c, "compendium_equipment_picker", data)
}

// HtmxCompendiumEquipmentPickerForOneShot renders the compendium equipment picker for one-shot items.
func HtmxCompendiumEquipmentPickerForOneShot(c *gin.Context) {
	adventureID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	q := strings.TrimSpace(c.Query("q"))
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	data := htmxCompendiumEquipmentPickerData{
		CharacterID: adventureID,
		Query:       q,
		Page:        page,
		PageSize:    pageSize,
	}

	if q != "" {
		items, totalCount := queryCompendiumEquipmentUnion(q, pageSize, offset)
		data.Items = items
		data.TotalCount = totalCount
		data.TotalPages = (totalCount + pageSize - 1) / pageSize
		if data.TotalPages < 1 {
			data.TotalPages = 1
		}
	}

	renderTemplate(c, "compendium_equipment_picker_oneshot", data)
}

// ─── Compendium Feature Picker (HTMX) ───

type htmxCompendiumFeaturePickerData struct {
	CharacterID int64
	Query       string
	Items       []compendiumFeaturePickerItem
	Page        int
	PageSize    int
	TotalCount  int
	TotalPages  int
}

// HtmxCompendiumFeaturePicker renders a picker of generic compendium entries to
// link as character features.
func HtmxCompendiumFeaturePicker(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Query("character_id"), 10, 64)
	q := strings.TrimSpace(c.Query("q"))
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	data := htmxCompendiumFeaturePickerData{
		CharacterID: charID,
		Query:       q,
		Page:        page,
		PageSize:    pageSize,
	}

	if q != "" {
		items, totalCount := queryCompendiumEntriesForFeatures(q, pageSize, offset)
		data.Items = items
		data.TotalCount = totalCount
		data.TotalPages = (totalCount + pageSize - 1) / pageSize
		if data.TotalPages < 1 {
			data.TotalPages = 1
		}
	}

	renderTemplate(c, "compendium_feature_picker", data)
}

// ─── Compendium Spell Browse (HTMX) ───

type htmxSpellBrowseData struct {
	Spells      []models.CompendiumSpell
	TotalCount  int
	Page        int
	PageSize    int
	TotalPages  int
	Query       string
	Class       string
	Level       string
	School      string
	Source      string
	Classes     []string
	Schools     []string
	SourceCount struct {
		SRD      int
		Homebrew int
		Imported int
	}
}
