package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
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
