package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/middleware"
	"villum/models"
)

// ─── Compendium Monster Browser (HTMX) ───

type htmxCompendiumMonsterListData struct {
	Monsters    []models.CompendiumMonster
	EncounterID int64
	CampaignID  int64
	AdventureID int64
	TotalCount  int
	Page        int
	PageSize    int
	TotalPages  int
	Query       string
	CR          string
	MonsterType string
}

type htmxCompendiumMonsterDetailData struct {
	Monster     *models.CompendiumMonster
	EncounterID int64
	CampaignID  int64
	AdventureID int64
}

func HtmxCompendiumMonsterBrowser(c *gin.Context) {
	renderTemplate(c, "compendium_monster_browser", htmxCompendiumMonsterListData{})
}

func HtmxCompendiumMonsterSearch(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	cr := strings.TrimSpace(c.Query("cr"))
	monsterType := strings.TrimSpace(c.Query("type"))
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}
	offset := (page - 1) * pageSize

	baseQuery := `SELECT id,name,type,size,ac,hp,str,dex,con,int_,wis,cha,cr,source,is_full,saves,skills,damage_vulnerabilities,damage_resistances,damage_immunities,condition_immunities,senses,languages,special_abilities,actions,legendary_actions,description FROM compendium_monsters WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM compendium_monsters WHERE 1=1`
	args := []any{}

	if q != "" {
		clause := " AND name LIKE ?"
		baseQuery += clause
		countQuery += clause
		args = append(args, "%"+q+"%")
	}
	if cr != "" {
		clause := " AND cr=?"
		baseQuery += clause
		countQuery += clause
		args = append(args, cr)
	}
	if monsterType != "" {
		clause := " AND type LIKE ?"
		baseQuery += clause
		countQuery += clause
		args = append(args, "%"+monsterType+"%")
	}

	var totalCount int
	err := db.DB.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		totalCount = 0
	}

	baseQuery += " ORDER BY name LIMIT ? OFFSET ?"
	pageArgs := append(args, pageSize, offset)

	rows, err := db.DB.Query(baseQuery, pageArgs...)
	if err != nil {
		c.String(http.StatusInternalServerError, "query error")
		return
	}
	defer rows.Close()

	out := make([]models.CompendiumMonster, 0)
	for rows.Next() {
		var m models.CompendiumMonster
		var isFull int
		if err := rows.Scan(&m.ID, &m.Name, &m.Type, &m.Size, &m.AC, &m.HP,
			&m.Str, &m.Dex, &m.Con, &m.Int, &m.Wis, &m.Cha,
			&m.CR, &m.Source, &isFull,
			&m.Saves, &m.Skills, &m.DamageVulnerabilities, &m.DamageResistances, &m.DamageImmunities, &m.ConditionImmunities,
			&m.Senses, &m.Languages, &m.SpecialAbilities, &m.Actions, &m.LegendaryActions, &m.Description); err != nil {
			middleware.LogWarn("compendium", "scan failed, skipping monster", "error", err)
			continue
		}
		m.IsFull = isFull == 1
		out = append(out, m)
	}

	totalPages := (totalCount + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}

	encounterID, _ := strconv.ParseInt(c.Query("encounter_id"), 10, 64)
	campaignID, _ := strconv.ParseInt(c.Query("campaign_id"), 10, 64)
	adventureID, _ := strconv.ParseInt(c.Query("adventure_id"), 10, 64)

	renderTemplate(c, "compendium_monster_list_item", htmxCompendiumMonsterListData{
		Monsters:    out,
		EncounterID: encounterID,
		CampaignID:  campaignID,
		AdventureID: adventureID,
		TotalCount:  totalCount,
		Page:        page,
		PageSize:    pageSize,
		TotalPages:  totalPages,
		Query:       q,
		CR:          cr,
		MonsterType: monsterType,
	})
}

func HtmxCompendiumMonsterDetail(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	encounterID, _ := strconv.ParseInt(c.Query("encounter_id"), 10, 64)
	campaignID, _ := strconv.ParseInt(c.Query("campaign_id"), 10, 64)
	adventureID, _ := strconv.ParseInt(c.Query("adventure_id"), 10, 64)

	var m models.CompendiumMonster
	var isFull int
	err := db.DB.QueryRow("SELECT id,name,type,size,ac,hp,str,dex,con,int_,wis,cha,cr,source,is_full,saves,skills,damage_vulnerabilities,damage_resistances,damage_immunities,condition_immunities,senses,languages,special_abilities,actions,legendary_actions,description FROM compendium_monsters WHERE id=?", id).
		Scan(&m.ID, &m.Name, &m.Type, &m.Size, &m.AC, &m.HP,
			&m.Str, &m.Dex, &m.Con, &m.Int, &m.Wis, &m.Cha,
			&m.CR, &m.Source, &isFull,
			&m.Saves, &m.Skills, &m.DamageVulnerabilities, &m.DamageResistances, &m.DamageImmunities, &m.ConditionImmunities,
			&m.Senses, &m.Languages, &m.SpecialAbilities, &m.Actions, &m.LegendaryActions, &m.Description)
	if err != nil {
		c.String(http.StatusNotFound, "monster not found")
		return
	}
	m.IsFull = isFull == 1

	renderTemplate(c, "compendium_monster_detail", htmxCompendiumMonsterDetailData{
		Monster:     &m,
		EncounterID: encounterID,
		CampaignID:  campaignID,
		AdventureID: adventureID,
	})
}

// ─── Compendium Monster Picker (HTMX) ───

func HtmxCompendiumMonsterPickerForEncounter(c *gin.Context) {
	encounterID, _ := strconv.ParseInt(c.Param("eid"), 10, 64)
	campaignID, _ := strconv.ParseInt(c.Query("campaign_id"), 10, 64)
	renderTemplate(c, "compendium_monster_browser", htmxCompendiumMonsterListData{
		EncounterID: encounterID,
		CampaignID:  campaignID,
	})
}

func HtmxCompendiumMonsterPickerForOneShot(c *gin.Context) {
	adventureID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	renderTemplate(c, "compendium_monster_browser", htmxCompendiumMonsterListData{
		AdventureID: adventureID,
	})
}

// ─── API Monster Import (HTMX) ───

type htmxAPIImportSearchResult struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	CR         string `json:"cr"`
	URL        string `json:"url"`
	ImportDone bool   `json:"import_done"`
}

type htmxAPIImportSearchData struct {
	Query   string
	Results []htmxAPIImportSearchResult
	Error   string
}

// HtmxAPIImportModal renders the API import modal
func HtmxAPIImportModal(c *gin.Context) {
	renderTemplate(c, "compendium_monster_api_import_modal", nil)
}

// HtmxAPIImportSearch searches the D&D 5e API for monsters by name
func HtmxAPIImportSearch(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		renderTemplate(c, "compendium_monster_api_import_results", htmxAPIImportSearchData{
			Query: query,
		})
		return
	}

	// Call the D&D 5e API search
	searchURL := fmt.Sprintf("https://www.dnd5eapi.co/api/monsters?name=%s", strings.ReplaceAll(query, " ", "+"))
	resp, err := http.Get(searchURL)
	if err != nil {
		renderTemplate(c, "compendium_monster_api_import_results", htmxAPIImportSearchData{
			Query: query,
			Error: "Failed to reach D&D API: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	var searchResult struct {
		Count   int                 `json:"count"`
		Results []map[string]string `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
		renderTemplate(c, "compendium_monster_api_import_results", htmxAPIImportSearchData{
			Query: query,
			Error: "Failed to parse API response: " + err.Error(),
		})
		return
	}

	if searchResult.Count == 0 {
		renderTemplate(c, "compendium_monster_api_import_results", htmxAPIImportSearchData{
			Query: query,
		})
		return
	}

	results := make([]htmxAPIImportSearchResult, 0, len(searchResult.Results))
	for _, r := range searchResult.Results {
		results = append(results, htmxAPIImportSearchResult{
			Name: r["name"],
			URL:  r["url"],
		})
	}

	renderTemplate(c, "compendium_monster_api_import_results", htmxAPIImportSearchData{
		Query:   query,
		Results: results,
	})
}

// HtmxImportAPIMonster imports a single monster from the D&D 5e API into compendium
func HtmxImportAPIMonster(c *gin.Context) {
	monsterURL := c.PostForm("url")
	monsterName := c.PostForm("name")
	confirmed := c.PostForm("confirm") == "true"

	if monsterURL == "" {
		c.String(http.StatusBadRequest, "url parameter required")
		return
	}

	// Check for duplicates (skip if confirmed)
	if !confirmed {
		var count int
		err := db.DB.QueryRow("SELECT COUNT(*) FROM compendium_monsters WHERE name=?", monsterName).Scan(&count)
		if err == nil && count > 0 {
			renderTemplate(c, "compendium_monster_api_import_confirm", map[string]string{
				"Name": monsterName,
				"URL":  monsterURL,
			})
			return
		}
	}

	// Fetch full monster detail
	detailURL := fmt.Sprintf("https://www.dnd5eapi.co%s", monsterURL)
	detailResp, err := http.Get(detailURL)
	if err != nil {
		c.String(http.StatusBadGateway, "Failed to fetch monster detail: "+err.Error())
		return
	}
	defer detailResp.Body.Close()

	if detailResp.StatusCode != 200 {
		c.String(http.StatusBadGateway, fmt.Sprintf("API returned status %d", detailResp.StatusCode))
		return
	}

	var data map[string]any
	if err := json.NewDecoder(detailResp.Body).Decode(&data); err != nil {
		c.String(http.StatusInternalServerError, "Failed to parse monster detail: "+err.Error())
		return
	}

	// Map fields to compendium_monsters columns
	name := getStrFromMap(data, "name", monsterName)
	monsterType := getStrFromMap(data, "type", "")
	size := getStrFromMap(data, "size", "Medium")
	// API returns capitalized sizes (Small, Medium, etc.) matching DB convention

	// AC - API returns array of objects like [{"type":"armor","value":15}]
	ac := 10
	if acData, ok := data["armor_class"]; ok {
		if acArr, ok := acData.([]any); ok && len(acArr) > 0 {
			if acObj, ok := acArr[0].(map[string]any); ok {
				if val, ok := acObj["value"].(float64); ok {
					ac = int(val)
				}
			}
		}
	}

	hp := 1
	if v, ok := data["hit_points"].(float64); ok {
		hp = int(v)
	}

	str := 10
	if v, ok := data["strength"].(float64); ok {
		str = int(v)
	}
	dex := 10
	if v, ok := data["dexterity"].(float64); ok {
		dex = int(v)
	}
	con := 10
	if v, ok := data["constitution"].(float64); ok {
		con = int(v)
	}
	int_ := 10
	if v, ok := data["intelligence"].(float64); ok {
		int_ = int(v)
	}
	wis := 10
	if v, ok := data["wisdom"].(float64); ok {
		wis = int(v)
	}
	cha := 10
	if v, ok := data["charisma"].(float64); ok {
		cha = int(v)
	}

	// CR - API returns float like 0.25, 0.5, 1, 2...
	cr := "0"
	if v, ok := data["challenge_rating"].(float64); ok {
		cr = float64ToCR(v)
	}

	// Saves (proficiencies where proficiency.type = "saving-throw")
	var saves strings.Builder
	profs, _ := data["proficiencies"].([]any)
	saveItems := make([]string, 0)
	for _, p := range profs {
		if pObj, ok := p.(map[string]any); ok {
			if profType, ok := pObj["proficiency"].(map[string]any); ok {
				if ref, ok := profType["name"].(string); ok && strings.HasPrefix(ref, "Saving Throw: ") {
					name := strings.TrimPrefix(ref, "Saving Throw: ")
					if val, ok := pObj["value"].(float64); ok {
						saveItems = append(saveItems, fmt.Sprintf("%s +%d", strings.ToLower(name), int(val)))
					}
				}
			}
		}
	}
	if len(saveItems) > 0 {
		saves.WriteString(strings.Join(saveItems, ", "))
	}

	// Skills (proficiencies where proficiency.type starts with "Skill: ")
	skillItems := make([]string, 0)
	for _, p := range profs {
		if pObj, ok := p.(map[string]any); ok {
			if profType, ok := pObj["proficiency"].(map[string]any); ok {
				if ref, ok := profType["name"].(string); ok && strings.HasPrefix(ref, "Skill: ") {
					name := strings.TrimPrefix(ref, "Skill: ")
					if val, ok := pObj["value"].(float64); ok {
						skillItems = append(skillItems, fmt.Sprintf("%s +%d", strings.ToLower(name), int(val)))
					}
				}
			}
		}
	}

	// Damage vulnerabilities
	dv, _ := json.Marshal(data["damage_vulnerabilities"])
	dvStr := string(dv)
	if dvStr == "null" || dvStr == "[]" {
		dvStr = ""
	} else {
		dvStr = strings.Trim(strings.ReplaceAll(strings.Trim(string(dv), "[]"), "\"", ""), " ")
	}

	// Damage resistances
	dr, _ := json.Marshal(data["damage_resistances"])
	drStr := string(dr)
	if drStr == "null" || drStr == "[]" {
		drStr = ""
	} else {
		drStr = strings.Trim(strings.ReplaceAll(strings.Trim(string(dr), "[]"), "\"", ""), " ")
	}

	// Damage immunities
	di, _ := json.Marshal(data["damage_immunities"])
	diStr := string(di)
	if diStr == "null" || diStr == "[]" {
		diStr = ""
	} else {
		diStr = strings.Trim(strings.ReplaceAll(strings.Trim(string(di), "[]"), "\"", ""), " ")
	}

	// Condition immunities
	ci, _ := json.Marshal(data["condition_immunities"])
	ciStr := string(ci)
	if ciStr == "null" || ciStr == "[]" {
		ciStr = ""
	} else {
		ciStr = strings.Trim(strings.ReplaceAll(strings.Trim(string(ci), "[]"), "\"", ""), " ")
	}

	// Senses
	var senses string
	if s, ok := data["senses"].(map[string]any); ok {
		senseParts := make([]string, 0, len(s))
		for k, v := range s {
			senseParts = append(senseParts, fmt.Sprintf("%s %v", k, v))
		}
		senses = strings.Join(senseParts, ", ")
	}

	// Languages
	languages, _ := data["languages"].(string)

	// Special abilities & actions - serialize to JSON string
	specialAbilities := jsonArrayToString(data["special_abilities"])
	actions := jsonArrayToString(data["actions"])
	legendaryActions := jsonArrayToString(data["legendary_actions"])

	// Description
	description := ""
	if desc, ok := data["desc"].([]any); ok {
		parts := make([]string, 0, len(desc))
		for _, d := range desc {
			parts = append(parts, fmt.Sprintf("%v", d))
		}
		description = strings.Join(parts, "\n")
	}

	// Insert into compendium_monsters
	_, err = db.DB.Exec(`INSERT INTO compendium_monsters(name,type,size,ac,hp,str,dex,con,int_,wis,cha,cr,source,is_full,
		saves,skills,damage_vulnerabilities,damage_resistances,damage_immunities,condition_immunities,senses,languages,
		special_abilities,actions,legendary_actions,description) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,
		?,?,?,?,?,?,?,?,?,?,?,?)`,
		name, monsterType, size, ac, hp, str, dex, con, int_, wis, cha,
		cr, "dnd5eapi", 1,
		saves.String(), strings.Join(skillItems, ", "), dvStr, drStr, diStr, ciStr,
		senses, languages, specialAbilities, actions, legendaryActions, description)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to import monster: "+err.Error())
		return
	}

	renderTemplate(c, "compendium_monster_api_import_success", map[string]string{
		"Name": name,
	})
}

// ─── Compendium Spell Detail (HTMX) ───

func HtmxCompendiumSpellDetail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid spell id")
		return
	}

	var s models.CompendiumSpell
	err = db.DB.QueryRow(`SELECT id,name,level,school,casting_time,"range",components,duration,
		description,higher_levels,classes,source_page,
		COALESCE(system,''),COALESCE(source,''),COALESCE(publisher,'')
		FROM compendium_spells WHERE id=?`, id).Scan(
		&s.ID, &s.Name, &s.Level, &s.School, &s.CastingTime, &s.Range,
		&s.Components, &s.Duration, &s.Description, &s.HigherLevels, &s.Classes,
		&s.SourcePage, &s.System, &s.Source, &s.Publisher)
	if err != nil {
		c.String(http.StatusNotFound, "spell not found")
		return
	}

	renderTemplate(c, "compendium_spell_detail_expanded", s)
}

func HtmxCompendiumSpellModal(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid spell id")
		return
	}

	var s models.CompendiumSpell
	err = db.DB.QueryRow(`SELECT id,name,level,school,casting_time,"range",components,duration,
		description,higher_levels,classes,source_page,
		COALESCE(system,''),COALESCE(source,''),COALESCE(publisher,'')
		FROM compendium_spells WHERE id=?`, id).Scan(
		&s.ID, &s.Name, &s.Level, &s.School, &s.CastingTime, &s.Range,
		&s.Components, &s.Duration, &s.Description, &s.HigherLevels, &s.Classes,
		&s.SourcePage, &s.System, &s.Source, &s.Publisher)
	if err != nil {
		c.String(http.StatusNotFound, "spell not found")
		return
	}

	renderTemplate(c, "compendium_spell_card", s)
}

// ─── Helpers ───

func getStrFromMap(m map[string]any, key, def string) string {
	if v, ok := m[key]; ok && v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return def
}

func float64ToCR(f float64) string {
	switch {
	case f == 0:
		return "0"
	case f == 0.125:
		return "1/8"
	case f == 0.25:
		return "1/4"
	case f == 0.5:
		return "1/2"
	default:
		return strconv.Itoa(int(f))
	}
}

func jsonArrayToString(v any) string {
	if v == nil {
		return ""
	}
	arr, ok := v.([]any)
	if !ok || len(arr) == 0 {
		return ""
	}
	// Build a formatted string representation
	parts := make([]string, 0, len(arr))
	for _, item := range arr {
		if obj, ok := item.(map[string]any); ok {
			// Extract name and description
			name, _ := obj["name"].(string)
			desc := ""
			if d, ok := obj["desc"].([]any); ok {
				dparts := make([]string, 0, len(d))
				for _, dd := range d {
					dparts = append(dparts, fmt.Sprintf("%v", dd))
				}
				desc = strings.Join(dparts, " ")
			}
			if name != "" {
				if desc != "" {
					parts = append(parts, fmt.Sprintf("%s. %s", name, desc))
				} else {
					parts = append(parts, name)
				}
			}
		}
	}
	return strings.Join(parts, "\n\n")
}

// ─── Monster Library (HTMX) ───

type htmxMonsterLibraryData struct {
	Monsters    []models.MonsterLibraryEntry
	Monster     *models.MonsterLibraryEntry
	CompMonster *models.CompendiumMonster
}

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
	Spells      []models.CompendiumSpell
	Page        int
	PageSize    int
	TotalCount  int
	TotalPages  int
}

func HtmxCompendiumSpellPicker(c *gin.Context) {
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

	data := htmxCompendiumSpellPickerData{
		CharacterID: charID,
		Query:       q,
		Page:        page,
		PageSize:    pageSize,
	}

	if q != "" {
		var totalCount int
		err := db.DB.QueryRow(`SELECT COUNT(*) FROM compendium_spells WHERE name LIKE ?`, "%"+q+"%").Scan(&totalCount)
		if err == nil {
			data.TotalCount = totalCount
			totalPages := (totalCount + pageSize - 1) / pageSize
			if totalPages < 1 {
				totalPages = 1
			}
			data.TotalPages = totalPages
		}

		rows, err := db.DB.Query(`SELECT id, name, level, school, casting_time, "range", components, duration,
			description, higher_levels, classes, source_page,
			COALESCE(system,''), COALESCE(source,''), COALESCE(publisher,'')
			FROM compendium_spells WHERE name LIKE ? ORDER BY level, name LIMIT ? OFFSET ?`, "%"+q+"%", pageSize, offset)
		if err == nil && rows != nil {
			defer rows.Close()
			for rows.Next() {
				var s models.CompendiumSpell
				rows.Scan(&s.ID, &s.Name, &s.Level, &s.School, &s.CastingTime, &s.Range, &s.Components, &s.Duration,
					&s.Description, &s.HigherLevels, &s.Classes, &s.SourcePage,
					&s.System, &s.Source, &s.Publisher)
				data.Spells = append(data.Spells, s)
			}
		}
	}

	renderTemplate(c, "compendium_spell_picker", data)
}

// ─── Compendium Equipment Picker (HTMX) ───

type htmxCompendiumEquipmentPickerData struct {
	CharacterID int64
	Query       string
	Items       []models.CompendiumEquipment
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
		var totalCount int
		err := db.DB.QueryRow(`SELECT COUNT(*) FROM compendium_equipment WHERE name LIKE ?`, "%"+q+"%").Scan(&totalCount)
		if err == nil {
			data.TotalCount = totalCount
			totalPages := (totalCount + pageSize - 1) / pageSize
			if totalPages < 1 {
				totalPages = 1
			}
			data.TotalPages = totalPages
		}

		rows, err := db.DB.Query(`SELECT id, name, category, cost, weight, description, source_page,
			COALESCE(system,''), COALESCE(source,''), COALESCE(item_type,''), COALESCE(item_rarity,''), COALESCE(publisher,'')
			FROM compendium_equipment WHERE name LIKE ? ORDER BY name LIMIT ? OFFSET ?`, "%"+q+"%", pageSize, offset)
		if err == nil && rows != nil {
			defer rows.Close()
			for rows.Next() {
				var e models.CompendiumEquipment
				rows.Scan(&e.ID, &e.Name, &e.Category, &e.Cost, &e.Weight, &e.Description, &e.SourcePage,
					&e.System, &e.Source, &e.ItemType, &e.ItemRarity, &e.Publisher)
				data.Items = append(data.Items, e)
			}
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
		var totalCount int
		err := db.DB.QueryRow(`SELECT COUNT(*) FROM compendium_equipment WHERE name LIKE ?`, "%"+q+"%").Scan(&totalCount)
		if err == nil {
			data.TotalCount = totalCount
			totalPages := (totalCount + pageSize - 1) / pageSize
			if totalPages < 1 {
				totalPages = 1
			}
			data.TotalPages = totalPages
		}

		rows, err := db.DB.Query(`SELECT id, name, category, cost, weight, description, source_page,
			COALESCE(system,''), COALESCE(source,''), COALESCE(item_type,''), COALESCE(item_rarity,''), COALESCE(publisher,'')
			FROM compendium_equipment WHERE name LIKE ? ORDER BY name LIMIT ? OFFSET ?`, "%"+q+"%", pageSize, offset)
		if err == nil && rows != nil {
			defer rows.Close()
			for rows.Next() {
				var e models.CompendiumEquipment
				rows.Scan(&e.ID, &e.Name, &e.Category, &e.Cost, &e.Weight, &e.Description, &e.SourcePage,
					&e.System, &e.Source, &e.ItemType, &e.ItemRarity, &e.Publisher)
				data.Items = append(data.Items, e)
			}
		}
	}

	renderTemplate(c, "compendium_equipment_picker_oneshot", data)
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

func HtmxCompendiumSpellBrowse(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	class := strings.TrimSpace(c.Query("class"))
	level := strings.TrimSpace(c.Query("level"))
	school := strings.TrimSpace(c.Query("school"))
	source := strings.TrimSpace(c.Query("source"))
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}
	offset := (page - 1) * pageSize

	// Build query
	baseQuery := `SELECT id,name,level,school,casting_time,"range",components,duration,
		description,higher_levels,classes,source_page,
		COALESCE(system,''),COALESCE(source,''),COALESCE(publisher,'') FROM compendium_spells WHERE 1=1`
	countQuery := "SELECT COUNT(*) FROM compendium_spells WHERE 1=1"
	args := []any{}

	if q != "" {
		clause := " AND name LIKE ?"
		baseQuery += clause
		countQuery += clause
		args = append(args, "%"+q+"%")
	}
	if class != "" {
		clause := " AND classes LIKE ?"
		baseQuery += clause
		countQuery += clause
		args = append(args, "%\""+class+"\"%")
	}
	if level != "" {
		// Support comma-separated levels: "1,2,3" or range "1-3"
		if strings.Contains(level, "-") {
			parts := strings.SplitN(level, "-", 2)
			from, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
			to, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err1 == nil && err2 == nil {
				clause := " AND level>=? AND level<=?"
				baseQuery += clause
				countQuery += clause
				args = append(args, from, to)
			}
		} else {
			levels := strings.Split(level, ",")
			placeholders := make([]string, len(levels))
			for i, l := range levels {
				lv, err := strconv.Atoi(strings.TrimSpace(l))
				if err == nil {
					placeholders[i] = "?"
					args = append(args, lv)
				}
			}
			validPlaceholders := []string{}
			for _, p := range placeholders {
				if p != "" {
					validPlaceholders = append(validPlaceholders, p)
				}
			}
			if len(validPlaceholders) > 0 {
				clause := " AND level IN (" + strings.Join(validPlaceholders, ",") + ")"
				baseQuery += clause
				countQuery += clause
			}
		}
	}
	if school != "" {
		clause := " AND school=?"
		baseQuery += clause
		countQuery += clause
		args = append(args, school)
	}
	if source != "" {
		switch source {
		case "srd":
			clause := " AND system='dnd5e' AND source='srd'"
			baseQuery += clause
			countQuery += clause
		case "homebrew":
			clause := " AND (source='homebrew' OR COALESCE(publisher,'')!='')"
			baseQuery += clause
			countQuery += clause
		case "imported":
			clause := " AND COALESCE(publisher,'')!='' AND source!='srd'"
			baseQuery += clause
			countQuery += clause
		}
	}

	// Get total count
	var totalCount int
	err := db.DB.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		totalCount = 0
	}

	// Get spells for current page
	baseQuery += " ORDER BY level, name LIMIT ? OFFSET ?"
	pageArgs := append(args, pageSize, offset)

	rows, err := db.DB.Query(baseQuery, pageArgs...)
	if err != nil {
		c.String(http.StatusInternalServerError, "query error: %v", err)
		return
	}
	defer rows.Close()

	spells := make([]models.CompendiumSpell, 0)
	for rows.Next() {
		var s models.CompendiumSpell
		err := rows.Scan(&s.ID, &s.Name, &s.Level, &s.School, &s.CastingTime, &s.Range,
			&s.Components, &s.Duration, &s.Description, &s.HigherLevels, &s.Classes,
			&s.SourcePage, &s.System, &s.Source, &s.Publisher)
		if err != nil {
			continue
		}
		spells = append(spells, s)
	}

	totalPages := (totalCount + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}

	// Fetch distinct classes and schools for filter dropdowns
	classes := getDistinctSpellClasses()
	schools := getDistinctSpellSchools()

	// Get source counts
	srdCount := getSpellCountBySource("srd")
	homebrewCount := getSpellCountBySource("homebrew")
	importedCount := getSpellCountBySource("imported")

	data := htmxSpellBrowseData{
		Spells:     spells,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		Query:      q,
		Class:      class,
		Level:      level,
		School:     school,
		Source:     source,
		Classes:    classes,
		Schools:    schools,
	}
	data.SourceCount.SRD = srdCount
	data.SourceCount.Homebrew = homebrewCount
	data.SourceCount.Imported = importedCount

	renderTemplate(c, "compendium_spell_browse", data)
}

func getDistinctSpellClasses() []string {
	rows, err := db.DB.Query("SELECT DISTINCT classes FROM compendium_spells WHERE classes != '' ORDER BY classes")
	if err != nil {
		return nil
	}
	defer rows.Close()

	classSet := make(map[string]bool)
	for rows.Next() {
		var classes string
		rows.Scan(&classes)
		// Parse JSON array like ["Artificer","Bard","Cleric"]
		cls := strings.Trim(classes, "[]")
		parts := strings.Split(cls, ",")
		for _, p := range parts {
			name := strings.Trim(strings.TrimSpace(p), "\"")
			if name != "" {
				classSet[name] = true
			}
		}
	}

	result := make([]string, 0, len(classSet))
	for name := range classSet {
		result = append(result, name)
	}
	sort.Strings(result)
	return result
}

func getDistinctSpellSchools() []string {
	rows, err := db.DB.Query("SELECT DISTINCT school FROM compendium_spells WHERE school != '' ORDER BY school")
	if err != nil {
		return nil
	}
	defer rows.Close()

	var schools []string
	for rows.Next() {
		var school string
		rows.Scan(&school)
		schools = append(schools, school)
	}
	return schools
}

func getSpellCountBySource(sourceType string) int {
	var query string
	switch sourceType {
	case "srd":
		query = "SELECT COUNT(*) FROM compendium_spells WHERE system='dnd5e' AND source='srd'"
	case "homebrew":
		query = "SELECT COUNT(*) FROM compendium_spells WHERE source='homebrew' OR (COALESCE(publisher,'')!='' AND source!='srd')"
	case "imported":
		query = "SELECT COUNT(*) FROM compendium_spells WHERE COALESCE(publisher,'')!='' AND source!='srd'"
	default:
		return 0
	}
	var count int
	db.DB.QueryRow(query).Scan(&count)
	return count
}

// ─── Compendium Race Browse (HTMX) ───

type htmxRaceBrowseData struct {
	Races      []models.CompendiumRace
	TotalCount int
	Page       int
	PageSize   int
	TotalPages int
	Query      string
}

func HtmxCompendiumRaceBrowse(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}
	offset := (page - 1) * pageSize

	baseQuery := `SELECT id,name,description,speed,size,ability_bonuses,traits,languages,source_page,system,source,category,expansion,publisher FROM compendium_races WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM compendium_races WHERE 1=1`
	args := []any{}

	if q != "" {
		clause := " AND name LIKE ?"
		baseQuery += clause
		countQuery += clause
		args = append(args, "%"+q+"%")
	}

	var totalCount int
	err := db.DB.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		totalCount = 0
	}

	baseQuery += " ORDER BY name LIMIT ? OFFSET ?"
	pageArgs := append(args, pageSize, offset)

	rows, err := db.DB.Query(baseQuery, pageArgs...)
	if err != nil {
		c.String(http.StatusInternalServerError, "query error")
		return
	}
	defer rows.Close()

	races := make([]models.CompendiumRace, 0)
	for rows.Next() {
		var r models.CompendiumRace
		if err := rows.Scan(&r.ID, &r.Name, &r.Description, &r.Speed, &r.Size, &r.AbilityBonuses, &r.Traits, &r.Languages, &r.SourcePage, &r.System, &r.Source, &r.Category, &r.Expansion, &r.Publisher); err != nil {
			middleware.LogWarn("compendium", "scan failed, skipping race", "error", err)
			continue
		}
		races = append(races, r)
	}

	totalPages := (totalCount + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}

	renderTemplate(c, "compendium_race_browse", htmxRaceBrowseData{
		Races:      races,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		Query:      q,
	})
}

// ─── Compendium Class Browse (HTMX) ───

type htmxClassBrowseData struct {
	Classes    []models.CompendiumClass
	TotalCount int
	Page       int
	PageSize   int
	TotalPages int
	Query      string
}

func HtmxCompendiumClassBrowse(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}
	offset := (page - 1) * pageSize

	baseQuery := `SELECT id,name,description,hit_die,primary_ability,saving_throws,proficiencies,spellcasting_ability,source_page,system,source,category,expansion,publisher FROM compendium_classes WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM compendium_classes WHERE 1=1`
	args := []any{}

	if q != "" {
		clause := " AND name LIKE ?"
		baseQuery += clause
		countQuery += clause
		args = append(args, "%"+q+"%")
	}

	var totalCount int
	err := db.DB.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		totalCount = 0
	}

	baseQuery += " ORDER BY name LIMIT ? OFFSET ?"
	pageArgs := append(args, pageSize, offset)

	rows, err := db.DB.Query(baseQuery, pageArgs...)
	if err != nil {
		c.String(http.StatusInternalServerError, "query error")
		return
	}
	defer rows.Close()

	classes := make([]models.CompendiumClass, 0)
	for rows.Next() {
		var cl models.CompendiumClass
		if err := rows.Scan(&cl.ID, &cl.Name, &cl.Description, &cl.HitDie, &cl.PrimaryAbility, &cl.SavingThrows, &cl.Proficiencies, &cl.SpellcastingAbility, &cl.SourcePage, &cl.System, &cl.Source, &cl.Category, &cl.Expansion, &cl.Publisher); err != nil {
			middleware.LogWarn("compendium", "scan failed, skipping class", "error", err)
			continue
		}
		classes = append(classes, cl)
	}

	totalPages := (totalCount + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}

	renderTemplate(c, "compendium_class_browse", htmxClassBrowseData{
		Classes:    classes,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		Query:      q,
	})
}

// ─── Compendium Equipment Browse (HTMX) ───

type htmxEquipmentBrowseData struct {
	Items      []models.CompendiumEquipment
	TotalCount int
	Page       int
	PageSize   int
	TotalPages int
	Query      string
	Category   string
	Categories []string
}

func getDistinctEquipmentCategories() []string {
	rows, err := db.DB.Query("SELECT DISTINCT category FROM compendium_equipment WHERE category != '' ORDER BY category")
	if err != nil {
		return nil
	}
	defer rows.Close()
	var categories []string
	for rows.Next() {
		var cat string
		rows.Scan(&cat)
		categories = append(categories, cat)
	}
	return categories
}

func HtmxCompendiumEquipmentBrowse(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	category := strings.TrimSpace(c.Query("category"))
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}
	offset := (page - 1) * pageSize

	baseQuery := `SELECT id,name,category,cost,weight,description,source_page,system,source,item_type,item_rarity,publisher FROM compendium_equipment WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM compendium_equipment WHERE 1=1`
	args := []any{}

	if q != "" {
		clause := " AND name LIKE ?"
		baseQuery += clause
		countQuery += clause
		args = append(args, "%"+q+"%")
	}
	if category != "" {
		clause := " AND category=?"
		baseQuery += clause
		countQuery += clause
		args = append(args, category)
	}

	var totalCount int
	err := db.DB.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		totalCount = 0
	}

	baseQuery += " ORDER BY name LIMIT ? OFFSET ?"
	pageArgs := append(args, pageSize, offset)

	rows, err := db.DB.Query(baseQuery, pageArgs...)
	if err != nil {
		c.String(http.StatusInternalServerError, "query error")
		return
	}
	defer rows.Close()

	items := make([]models.CompendiumEquipment, 0)
	for rows.Next() {
		var e models.CompendiumEquipment
		if err := rows.Scan(&e.ID, &e.Name, &e.Category, &e.Cost, &e.Weight, &e.Description, &e.SourcePage, &e.System, &e.Source, &e.ItemType, &e.ItemRarity, &e.Publisher); err != nil {
			middleware.LogWarn("compendium", "scan failed, skipping equipment", "error", err)
			continue
		}
		items = append(items, e)
	}

	totalPages := (totalCount + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}

	categories := getDistinctEquipmentCategories()

	renderTemplate(c, "compendium_equipment_browse", htmxEquipmentBrowseData{
		Items:      items,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		Query:      q,
		Category:   category,
		Categories: categories,
	})
}

// ─── Compendium Card (HTMX partial) ───

func HtmxCompendiumCard(c *gin.Context) {
	entityType := c.Param("type")
	entityIDStr := c.Param("id")
	entityID, err := strconv.ParseInt(entityIDStr, 10, 64)
	if err != nil {
		renderTemplate(c, "compendium_card_not_found", nil)
		return
	}

	switch entityType {
	case "spell":
		var s models.CompendiumSpell
		err := db.DB.QueryRow(`SELECT id, name, level, school, casting_time, "range", components, duration,
			description, higher_levels, classes, source_page,
			COALESCE(system,''), COALESCE(source,''), COALESCE(publisher,'')
			FROM compendium_spells WHERE id=?`, entityID).Scan(
			&s.ID, &s.Name, &s.Level, &s.School, &s.CastingTime, &s.Range, &s.Components, &s.Duration,
			&s.Description, &s.HigherLevels, &s.Classes, &s.SourcePage,
			&s.System, &s.Source, &s.Publisher)
		if err != nil {
			renderTemplate(c, "compendium_card_not_found", nil)
			return
		}
		renderTemplate(c, "compendium_spell_card", s)

	case "equipment":
		var e models.CompendiumEquipment
		err := db.DB.QueryRow(`SELECT id, name, category, cost, weight, description, source_page,
			COALESCE(system,''), COALESCE(source,''), COALESCE(item_type,''), COALESCE(item_rarity,''), COALESCE(publisher,'')
			FROM compendium_equipment WHERE id=?`, entityID).Scan(
			&e.ID, &e.Name, &e.Category, &e.Cost, &e.Weight, &e.Description, &e.SourcePage,
			&e.System, &e.Source, &e.ItemType, &e.ItemRarity, &e.Publisher)
		if err != nil {
			renderTemplate(c, "compendium_card_not_found", nil)
			return
		}
		renderTemplate(c, "compendium_equipment_card", e)

	case "monster":
		var m models.CompendiumMonster
		var isFull int
		err := db.DB.QueryRow(`SELECT id, name, type, size, ac, hp, str, dex, con, int_, wis, cha, cr,
			source, is_full, saves, skills, damage_vulnerabilities, damage_resistances, damage_immunities,
			condition_immunities, senses, languages, special_abilities, actions, legendary_actions, description,
			COALESCE(alignment,''), COALESCE(expansion,''), COALESCE(publisher,'')
			FROM compendium_monsters WHERE id=?`, entityID).Scan(
			&m.ID, &m.Name, &m.Type, &m.Size, &m.AC, &m.HP, &m.Str, &m.Dex, &m.Con, &m.Int, &m.Wis, &m.Cha, &m.CR,
			&m.Source, &isFull, &m.Saves, &m.Skills, &m.DamageVulnerabilities, &m.DamageResistances, &m.DamageImmunities,
			&m.ConditionImmunities, &m.Senses, &m.Languages, &m.SpecialAbilities, &m.Actions, &m.LegendaryActions, &m.Description,
			&m.Alignment, &m.Expansion, &m.Publisher)
		if err != nil {
			renderTemplate(c, "compendium_card_not_found", nil)
			return
		}
		m.IsFull = isFull == 1
		renderTemplate(c, "compendium_monster_card", m)

	default:
		renderTemplate(c, "compendium_card_not_found", nil)
	}
}

// ─── Compendium Global Search (HTMX) ───

type htmxCompendiumGlobalSearchItem struct {
	Type           string
	ID             int64
	Name           string
	Subtype        string
	CR             string
	Level          int
	HitDie         int
	PrimaryAbility string
}

type htmxCompendiumGlobalSearchData struct {
	Query       string
	Spells      []htmxCompendiumGlobalSearchItem
	Equipment   []htmxCompendiumGlobalSearchItem
	Monsters    []htmxCompendiumGlobalSearchItem
	Races       []htmxCompendiumGlobalSearchItem
	Classes     []htmxCompendiumGlobalSearchItem
	Feats       []htmxCompendiumGlobalSearchItem
	Backgrounds []htmxCompendiumGlobalSearchItem
}

func HtmxCompendiumGlobalSearch(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	data := htmxCompendiumGlobalSearchData{Query: q}

	if q == "" {
		renderTemplate(c, "compendium_global_search_results", data)
		return
	}

	like := "%" + q + "%"

	// Spells
	rows, qErr := db.DB.Query("SELECT id, name, level, school FROM compendium_spells WHERE name LIKE ? ORDER BY level, name LIMIT 10", like)
	if qErr != nil {
		middleware.LogWarn("compendium", "query failed", "type", "spell", "error", qErr)
	} else {
		for rows.Next() {
			var item htmxCompendiumGlobalSearchItem
			item.Type = "spell"
			if err := rows.Scan(&item.ID, &item.Name, &item.Level, &item.Subtype); err != nil {
				middleware.LogWarn("compendium", "scan failed, skipping row", "type", "spell", "error", err)
				continue
			}
			data.Spells = append(data.Spells, item)
		}
		rows.Close()
	}

	// Equipment
	rows, qErr = db.DB.Query("SELECT id, name, category FROM compendium_equipment WHERE name LIKE ? ORDER BY name LIMIT 10", like)
	if qErr != nil {
		middleware.LogWarn("compendium", "query failed", "type", "equipment", "error", qErr)
	} else {
		for rows.Next() {
			var item htmxCompendiumGlobalSearchItem
			item.Type = "equipment"
			if err := rows.Scan(&item.ID, &item.Name, &item.Subtype); err != nil {
				middleware.LogWarn("compendium", "scan failed, skipping row", "type", "equipment", "error", err)
				continue
			}
			data.Equipment = append(data.Equipment, item)
		}
		rows.Close()
	}

	// Monsters
	rows, qErr = db.DB.Query("SELECT id, name, cr, type FROM compendium_monsters WHERE name LIKE ? ORDER BY name LIMIT 10", like)
	if qErr != nil {
		middleware.LogWarn("compendium", "query failed", "type", "monster", "error", qErr)
	} else {
		for rows.Next() {
			var item htmxCompendiumGlobalSearchItem
			item.Type = "monster"
			if err := rows.Scan(&item.ID, &item.Name, &item.CR, &item.Subtype); err != nil {
				middleware.LogWarn("compendium", "scan failed, skipping row", "type", "monster", "error", err)
				continue
			}
			data.Monsters = append(data.Monsters, item)
		}
		rows.Close()
	}

	// Races
	rows, qErr = db.DB.Query("SELECT id, name FROM compendium_races WHERE name LIKE ? ORDER BY name LIMIT 10", like)
	if qErr != nil {
		middleware.LogWarn("compendium", "query failed", "type", "race", "error", qErr)
	} else {
		for rows.Next() {
			var item htmxCompendiumGlobalSearchItem
			item.Type = "race"
			if err := rows.Scan(&item.ID, &item.Name); err != nil {
				middleware.LogWarn("compendium", "scan failed, skipping row", "type", "race", "error", err)
				continue
			}
			data.Races = append(data.Races, item)
		}
		rows.Close()
	}

	// Classes
	rows, qErr = db.DB.Query("SELECT id, name, hit_die, primary_ability FROM compendium_classes WHERE name LIKE ? ORDER BY name LIMIT 10", like)
	if qErr != nil {
		middleware.LogWarn("compendium", "query failed", "type", "class", "error", qErr)
	} else {
		for rows.Next() {
			var item htmxCompendiumGlobalSearchItem
			item.Type = "class"
			if err := rows.Scan(&item.ID, &item.Name, &item.HitDie, &item.PrimaryAbility); err != nil {
				middleware.LogWarn("compendium", "scan failed, skipping row", "type", "class", "error", err)
				continue
			}
			data.Classes = append(data.Classes, item)
		}
		rows.Close()
	}

	// Feats
	rows, qErr = db.DB.Query("SELECT id, name FROM compendium_feats WHERE name LIKE ? ORDER BY name LIMIT 10", like)
	if qErr != nil {
		middleware.LogWarn("compendium", "query failed", "type", "feat", "error", qErr)
	} else {
		for rows.Next() {
			var item htmxCompendiumGlobalSearchItem
			item.Type = "feat"
			if err := rows.Scan(&item.ID, &item.Name); err != nil {
				middleware.LogWarn("compendium", "scan failed, skipping row", "type", "feat", "error", err)
				continue
			}
			data.Feats = append(data.Feats, item)
		}
		rows.Close()
	}

	// Backgrounds
	rows, qErr = db.DB.Query("SELECT id, name FROM compendium_backgrounds WHERE name LIKE ? ORDER BY name LIMIT 10", like)
	if qErr != nil {
		middleware.LogWarn("compendium", "query failed", "type", "background", "error", qErr)
	} else {
		for rows.Next() {
			var item htmxCompendiumGlobalSearchItem
			item.Type = "background"
			if err := rows.Scan(&item.ID, &item.Name); err != nil {
				middleware.LogWarn("compendium", "scan failed, skipping row", "type", "background", "error", err)
				continue
			}
			data.Backgrounds = append(data.Backgrounds, item)
		}
		rows.Close()
	}

	renderTemplate(c, "compendium_global_search_results", data)
}
