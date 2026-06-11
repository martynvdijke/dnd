package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

// ─── Compendium Monster Browser (HTMX) ───

type htmxCompendiumMonsterListData struct {
	Monsters    []models.CompendiumMonster
	EncounterID int64
	CampaignID  int64
	AdventureID int64
}

type htmxCompendiumMonsterDetailData struct {
	Monster     *models.CompendiumMonster
	EncounterID int64
	CampaignID  int64
	AdventureID int64
}

func HtmxCompendiumMonsterBrowser(c *gin.Context) {
	renderTemplate(c, "compendium_monster_browser.html", nil)
}

func HtmxCompendiumMonsterSearch(c *gin.Context) {
	query := "SELECT id,name,type,size,ac,hp,str,dex,con,int_,wis,cha,cr,source,is_full,saves,skills,damage_vulnerabilities,damage_resistances,damage_immunities,condition_immunities,senses,languages,special_abilities,actions,legendary_actions,description FROM compendium_monsters WHERE 1=1"
	args := []interface{}{}

	if q := c.Query("q"); q != "" {
		query += " AND name LIKE ?"
		args = append(args, "%"+q+"%")
	}
	if cr := c.Query("cr"); cr != "" {
		query += " AND cr=?"
		args = append(args, cr)
	}
	if t := c.Query("type"); t != "" {
		query += " AND type LIKE ?"
		args = append(args, "%"+t+"%")
	}
	query += " ORDER BY name LIMIT 50"

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		c.String(http.StatusInternalServerError, "query error")
		return
	}
	defer rows.Close()

	out := make([]models.CompendiumMonster, 0)
	for rows.Next() {
		var m models.CompendiumMonster
		var isFull int
		rows.Scan(&m.ID, &m.Name, &m.Type, &m.Size, &m.AC, &m.HP,
			&m.Str, &m.Dex, &m.Con, &m.Int, &m.Wis, &m.Cha,
			&m.CR, &m.Source, &isFull,
			&m.Saves, &m.Skills, &m.DamageVulnerabilities, &m.DamageResistances, &m.DamageImmunities, &m.ConditionImmunities,
			&m.Senses, &m.Languages, &m.SpecialAbilities, &m.Actions, &m.LegendaryActions, &m.Description)
		m.IsFull = isFull == 1
		out = append(out, m)
	}

	encounterID, _ := strconv.ParseInt(c.Query("encounter_id"), 10, 64)
	campaignID, _ := strconv.ParseInt(c.Query("campaign_id"), 10, 64)

	renderTemplate(c, "compendium_monster_list_item.html", htmxCompendiumMonsterListData{
		Monsters:    out,
		EncounterID: encounterID,
		CampaignID:  campaignID,
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

	renderTemplate(c, "compendium_monster_detail.html", htmxCompendiumMonsterDetailData{
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
	renderTemplate(c, "compendium_monster_browser.html", htmxCompendiumMonsterListData{
		EncounterID: encounterID,
		CampaignID:  campaignID,
	})
}

func HtmxCompendiumMonsterPickerForOneShot(c *gin.Context) {
	adventureID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	renderTemplate(c, "compendium_monster_browser.html", htmxCompendiumMonsterListData{
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
		Count   int                `json:"count"`
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

	var data map[string]interface{}
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
		if acArr, ok := acData.([]interface{}); ok && len(acArr) > 0 {
			if acObj, ok := acArr[0].(map[string]interface{}); ok {
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
	profs, _ := data["proficiencies"].([]interface{})
	saveItems := make([]string, 0)
	for _, p := range profs {
		if pObj, ok := p.(map[string]interface{}); ok {
			if profType, ok := pObj["proficiency"].(map[string]interface{}); ok {
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
		if pObj, ok := p.(map[string]interface{}); ok {
			if profType, ok := pObj["proficiency"].(map[string]interface{}); ok {
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
	if s, ok := data["senses"].(map[string]interface{}); ok {
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
	if desc, ok := data["desc"].([]interface{}); ok {
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

// ─── Helpers ───

func getStrFromMap(m map[string]interface{}, key, def string) string {
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

func jsonArrayToString(v interface{}) string {
	if v == nil {
		return ""
	}
	arr, ok := v.([]interface{})
	if !ok || len(arr) == 0 {
		return ""
	}
	// Build a formatted string representation
	parts := make([]string, 0, len(arr))
	for _, item := range arr {
		if obj, ok := item.(map[string]interface{}); ok {
			// Extract name and description
			name, _ := obj["name"].(string)
			desc := ""
			if d, ok := obj["desc"].([]interface{}); ok {
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
	Monsters   []models.MonsterLibraryEntry
	Monster    *models.MonsterLibraryEntry
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
	renderTemplate(c, "monster_library_list.html", htmxMonsterLibraryData{Monsters: monsters})
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
		renderTemplate(c, "monster_library_form.html", htmxMonsterLibraryData{Monster: &m})
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
		renderTemplate(c, "monster_library_form.html", htmxMonsterLibraryData{CompMonster: &cm})
		return
	}

	// New empty form
	renderTemplate(c, "monster_library_form.html", htmxMonsterLibraryData{})
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
	renderTemplate(c, "monster_library_section.html", htmxMonsterLibraryData{Monsters: monsters})
}

// ─── Compendium Spell Picker (HTMX) ───

type htmxCompendiumSpellPickerData struct {
	CharacterID int64
	Query       string
	Spells      []models.CompendiumSpell
}

func HtmxCompendiumSpellPicker(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Query("character_id"), 10, 64)
	q := strings.TrimSpace(c.Query("q"))

	data := htmxCompendiumSpellPickerData{
		CharacterID: charID,
		Query:       q,
	}

	if q != "" {
		rows, err := db.DB.Query(`SELECT id, name, level, school, casting_time, "range", components, duration,
			description, higher_levels, classes, source_page,
			COALESCE(system,''), COALESCE(source,''), COALESCE(publisher,'')
			FROM compendium_spells WHERE name LIKE ? ORDER BY level, name LIMIT 20`, "%"+q+"%")
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
}

func HtmxCompendiumEquipmentPicker(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Query("character_id"), 10, 64)
	q := strings.TrimSpace(c.Query("q"))

	data := htmxCompendiumEquipmentPickerData{
		CharacterID: charID,
		Query:       q,
	}

	if q != "" {
		rows, err := db.DB.Query(`SELECT id, name, category, cost, weight, description, source_page,
			COALESCE(system,''), COALESCE(source,''), COALESCE(item_type,''), COALESCE(item_rarity,''), COALESCE(publisher,'')
			FROM compendium_equipment WHERE name LIKE ? ORDER BY name LIMIT 20`, "%"+q+"%")
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
