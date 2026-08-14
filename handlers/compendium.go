package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/middleware"
	"villum/models"
)

func ListCompendiumRaces(c *gin.Context) {
	rows, err := db.DB.Query("SELECT id,name,description,speed,size,ability_bonuses,traits,languages,source_page,system,source,category,expansion,publisher FROM compendium_races ORDER BY name")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var out = make([]models.CompendiumRace, 0)
	for rows.Next() {
		var r models.CompendiumRace
		if err := rows.Scan(&r.ID, &r.Name, &r.Description, &r.Speed, &r.Size, &r.AbilityBonuses, &r.Traits, &r.Languages, &r.SourcePage, &r.System, &r.Source, &r.Category, &r.Expansion, &r.Publisher); err != nil {
			middleware.LogWarn("compendium", "scan failed, skipping race", "error", err)
			continue
		}
		out = append(out, r)
	}
	c.JSON(http.StatusOK, out)
}

func ListCompendiumClasses(c *gin.Context) {
	rows, err := db.DB.Query("SELECT id,name,description,hit_die,primary_ability,saving_throws,proficiencies,spellcasting_ability,source_page,system,source,category,expansion,publisher FROM compendium_classes ORDER BY name")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var out = make([]models.CompendiumClass, 0)
	for rows.Next() {
		var cl models.CompendiumClass
		if err := rows.Scan(&cl.ID, &cl.Name, &cl.Description, &cl.HitDie, &cl.PrimaryAbility, &cl.SavingThrows, &cl.Proficiencies, &cl.SpellcastingAbility, &cl.SourcePage, &cl.System, &cl.Source, &cl.Category, &cl.Expansion, &cl.Publisher); err != nil {
			middleware.LogWarn("compendium", "scan failed, skipping class", "error", err)
			continue
		}
		out = append(out, cl)
	}
	c.JSON(http.StatusOK, out)
}

func ListCompendiumSpells(c *gin.Context) {
	query := "SELECT id,name,level,school,casting_time,range,components,duration,description,higher_levels,classes,source_page,system,source,publisher FROM compendium_spells WHERE 1=1"
	args := []any{}

	if cls := c.Query("class"); cls != "" {
		query += " AND classes LIKE ?"
		args = append(args, "%\""+cls+"\"%")
	}
	if lvl := c.Query("level"); lvl != "" {
		query += " AND level=?"
		args = append(args, lvl)
	}
	if school := c.Query("school"); school != "" {
		query += " AND school=?"
		args = append(args, school)
	}
	if q := c.Query("q"); q != "" {
		query += " AND name LIKE ?"
		args = append(args, "%"+q+"%")
	}
	query += " ORDER BY level, name"

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var out = make([]models.CompendiumSpell, 0)
	for rows.Next() {
		var s models.CompendiumSpell
		if err := rows.Scan(&s.ID, &s.Name, &s.Level, &s.School, &s.CastingTime, &s.Range, &s.Components, &s.Duration, &s.Description, &s.HigherLevels, &s.Classes, &s.SourcePage, &s.System, &s.Source, &s.Publisher); err != nil {
			middleware.LogWarn("compendium", "scan failed, skipping spell", "error", err)
			continue
		}
		out = append(out, s)
	}
	c.JSON(http.StatusOK, out)
}

func ListCompendiumFeats(c *gin.Context) {
	rows, err := db.DB.Query("SELECT id,name,description,prerequisites,source_page,system,source FROM compendium_feats ORDER BY name")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var out = make([]models.CompendiumFeat, 0)
	for rows.Next() {
		var f models.CompendiumFeat
		if err := rows.Scan(&f.ID, &f.Name, &f.Description, &f.Prerequisites, &f.SourcePage, &f.System, &f.Source); err != nil {
			middleware.LogWarn("compendium", "scan failed, skipping feat", "error", err)
			continue
		}
		out = append(out, f)
	}
	c.JSON(http.StatusOK, out)
}

func ListCompendiumBackgrounds(c *gin.Context) {
	rows, err := db.DB.Query("SELECT id,name,description,feature_name,feature_description,proficiencies,source_page,system,source,category,data_list,data_bonds,data_flaws,data_ideals,data_equipment,data_starting_gold,data_personality_traits,publisher FROM compendium_backgrounds ORDER BY name")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var out = make([]models.CompendiumBackground, 0)
	for rows.Next() {
		var b models.CompendiumBackground
		if err := rows.Scan(&b.ID, &b.Name, &b.Description, &b.FeatureName, &b.FeatureDescription, &b.Proficiencies, &b.SourcePage, &b.System, &b.Source, &b.Category, &b.DataList, &b.DataBonds, &b.DataFlaws, &b.DataIdeals, &b.DataEquipment, &b.DataStartingGold, &b.DataPersonalityTraits, &b.Publisher); err != nil {
			middleware.LogWarn("compendium", "scan failed, skipping background", "error", err)
			continue
		}
		out = append(out, b)
	}
	c.JSON(http.StatusOK, out)
}

func ListCompendiumEquipment(c *gin.Context) {
	query := "SELECT id,name,category,cost,weight,description,source_page,system,source,item_type,item_rarity,publisher FROM compendium_equipment WHERE 1=1"
	args := []any{}

	if cat := c.Query("category"); cat != "" {
		query += " AND category=?"
		args = append(args, cat)
	}
	if q := c.Query("q"); q != "" {
		query += " AND name LIKE ?"
		args = append(args, "%"+q+"%")
	}
	query += " ORDER BY name"

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := make([]compendiumEquipmentPickerItem, 0)
	for rows.Next() {
		var e models.CompendiumEquipment
		if err := rows.Scan(&e.ID, &e.Name, &e.Category, &e.Cost, &e.Weight, &e.Description, &e.SourcePage, &e.System, &e.Source, &e.ItemType, &e.ItemRarity, &e.Publisher); err != nil {
			middleware.LogWarn("compendium", "scan failed, skipping equipment", "error", err)
			continue
		}
		out = append(out, compendiumEquipmentPickerItem{CompendiumEquipment: e, Source: "equipment"})
	}

	// Append generic schema entries so imported items are linkable too.
	equery := `SELECT e.id, COALESCE(json_extract(e.data,'$.name'),''),
		COALESCE(json_extract(e.data,'$.category'), json_extract(e.data,'$.type'), json_extract(e.data,'$.item_type'), json_extract(e.data,'$.subtype'),''),
		COALESCE(json_extract(e.data,'$.cost'), json_extract(e.data,'$.price'), json_extract(e.data,'$.value'),''),
		COALESCE(CAST(json_extract(e.data,'$.weight') AS REAL),0),
		COALESCE(json_extract(e.data,'$.description'), json_extract(e.data,'$.desc'),''),
		'', '', '', '', '', '', COALESCE(s.display_name,'')
		FROM compendium_entries e LEFT JOIN compendium_schemas s ON s.id=e.schema_id
		WHERE COALESCE(json_extract(e.data,'$.name'),'') <> ''`
	eargs := []any{}
	if cat := c.Query("category"); cat != "" {
		equery += " AND COALESCE(json_extract(e.data,'$.category'), json_extract(e.data,'$.type'), json_extract(e.data,'$.item_type'), json_extract(e.data,'$.subtype'),'') = ?"
		eargs = append(eargs, cat)
	}
	if q := c.Query("q"); q != "" {
		equery += " AND json_extract(e.data,'$.name') LIKE ?"
		eargs = append(eargs, "%"+q+"%")
	}
	equery += " ORDER BY json_extract(e.data,'$.name')"
	erows, err := db.DB.Query(equery, eargs...)
	if err == nil {
		defer erows.Close()
		for erows.Next() {
			var e models.CompendiumEquipment
			var schemaName string
			if err := erows.Scan(&e.ID, &e.Name, &e.Category, &e.Cost, &e.Weight, &e.Description, &e.SourcePage, &e.System, &e.Source, &e.ItemType, &e.ItemRarity, &e.Publisher, &schemaName); err != nil {
				middleware.LogWarn("compendium", "entry scan failed, skipping", "error", err)
				continue
			}
			out = append(out, compendiumEquipmentPickerItem{CompendiumEquipment: e, Source: "entry", SchemaName: schemaName})
		}
	}
	c.JSON(http.StatusOK, out)
}

// ─── Monster Compendium ───

func ListCompendiumMonsters(c *gin.Context) {
	query := "SELECT id,name,type,size,ac,hp,str,dex,con,int_,wis,cha,cr,source,is_full,saves,skills,damage_vulnerabilities,damage_resistances,damage_immunities,condition_immunities,senses,languages,special_abilities,actions,legendary_actions,description,alignment,expansion,publisher FROM compendium_monsters WHERE 1=1"
	args := []any{}

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
	query += " ORDER BY name"

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	out := make([]models.CompendiumMonster, 0)
	for rows.Next() {
		var m models.CompendiumMonster
		var isFull int
		err := rows.Scan(&m.ID, &m.Name, &m.Type, &m.Size, &m.AC, &m.HP,
			&m.Str, &m.Dex, &m.Con, &m.Int, &m.Wis, &m.Cha,
			&m.CR, &m.Source, &isFull,
			&m.Saves, &m.Skills, &m.DamageVulnerabilities, &m.DamageResistances, &m.DamageImmunities, &m.ConditionImmunities,
			&m.Senses, &m.Languages, &m.SpecialAbilities, &m.Actions, &m.LegendaryActions, &m.Description,
			&m.Alignment, &m.Expansion, &m.Publisher)
		if err != nil {
			middleware.LogWarn("compendium", "scan failed, skipping monster", "error", err)
			continue
		}
		m.IsFull = isFull == 1
		out = append(out, m)
	}
	c.JSON(http.StatusOK, out)
}

func GetCompendiumMonster(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var m models.CompendiumMonster
	var isFull int
	err := db.DB.QueryRow("SELECT id,name,type,size,ac,hp,str,dex,con,int_,wis,cha,cr,source,is_full,saves,skills,damage_vulnerabilities,damage_resistances,damage_immunities,condition_immunities,senses,languages,special_abilities,actions,legendary_actions,description,alignment,expansion,publisher FROM compendium_monsters WHERE id=?", id).
		Scan(&m.ID, &m.Name, &m.Type, &m.Size, &m.AC, &m.HP,
			&m.Str, &m.Dex, &m.Con, &m.Int, &m.Wis, &m.Cha,
			&m.CR, &m.Source, &isFull,
			&m.Saves, &m.Skills, &m.DamageVulnerabilities, &m.DamageResistances, &m.DamageImmunities, &m.ConditionImmunities,
			&m.Senses, &m.Languages, &m.SpecialAbilities, &m.Actions, &m.LegendaryActions, &m.Description,
			&m.Alignment, &m.Expansion, &m.Publisher)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "monster not found"})
		return
	}
	m.IsFull = isFull == 1
	c.JSON(http.StatusOK, m)
}

func SearchCompendium(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query required"})
		return
	}

	type SearchResult struct {
		Type           string `json:"type"`
		ID             int64  `json:"id"`
		Name           string `json:"name"`
		Subtype        string `json:"subtype,omitempty"`
		CR             string `json:"cr,omitempty"`
		Level          int    `json:"level,omitempty"`
		HitDie         int    `json:"hit_die,omitempty"`
		PrimaryAbility string `json:"primary_ability,omitempty"`
	}

	typeFilter := strings.TrimSpace(c.Query("type"))
	results := []SearchResult{}

	if typeFilter == "" || typeFilter == "spell" {
		extra := ""
		args := []any{}
		if cls := c.Query("class"); cls != "" {
			extra += " AND classes LIKE ?"
			args = append(args, "%\""+cls+"\"%")
		}
		if lvl := c.Query("level"); lvl != "" {
			extra += " AND level=?"
			args = append(args, lvl)
		}
		if school := c.Query("school"); school != "" {
			extra += " AND school=?"
			args = append(args, school)
		}
		rows, qErr := db.DB.Query("SELECT id, name, level, school FROM compendium_spells WHERE name LIKE ?"+extra+" ORDER BY level, name LIMIT 10", append([]any{"%" + q + "%"}, args...)...)
		if qErr != nil {
			middleware.LogWarn("compendium", "query failed", "type", "spell", "error", qErr)
		} else {
			for rows.Next() {
				var r SearchResult
				r.Type = "spell"
				if err := rows.Scan(&r.ID, &r.Name, &r.Level, &r.Subtype); err != nil {
					middleware.LogWarn("compendium", "scan failed, skipping row", "type", "spell", "error", err)
					continue
				}
				results = append(results, r)
			}
			rows.Close()
		}
	}

	if typeFilter == "" || typeFilter == "equipment" {
		extra := ""
		args := []any{}
		if cat := c.Query("category"); cat != "" {
			extra += " AND category=?"
			args = append(args, cat)
		}
		rows, qErr := db.DB.Query("SELECT id, name, category FROM compendium_equipment WHERE name LIKE ?"+extra+" ORDER BY name LIMIT 10", append([]any{"%" + q + "%"}, args...)...)
		if qErr != nil {
			middleware.LogWarn("compendium", "query failed", "type", "equipment", "error", qErr)
		} else {
			for rows.Next() {
				var r SearchResult
				r.Type = "equipment"
				if err := rows.Scan(&r.ID, &r.Name, &r.Subtype); err != nil {
					middleware.LogWarn("compendium", "scan failed, skipping row", "type", "equipment", "error", err)
					continue
				}
				results = append(results, r)
			}
			rows.Close()
		}
	}

	if typeFilter == "" || typeFilter == "monster" {
		extra := ""
		args := []any{}
		if cr := c.Query("cr"); cr != "" {
			extra += " AND cr=?"
			args = append(args, cr)
		}
		if t := c.Query("monster_type"); t != "" {
			extra += " AND type LIKE ?"
			args = append(args, "%"+t+"%")
		}
		rows, qErr := db.DB.Query("SELECT id, name, cr, type FROM compendium_monsters WHERE name LIKE ?"+extra+" ORDER BY name LIMIT 10", append([]any{"%" + q + "%"}, args...)...)
		if qErr != nil {
			middleware.LogWarn("compendium", "query failed", "type", "monster", "error", qErr)
		} else {
			for rows.Next() {
				var r SearchResult
				r.Type = "monster"
				if err := rows.Scan(&r.ID, &r.Name, &r.CR, &r.Subtype); err != nil {
					middleware.LogWarn("compendium", "scan failed, skipping row", "type", "monster", "error", err)
					continue
				}
				results = append(results, r)
			}
			rows.Close()
		}
	}

	if typeFilter == "" || typeFilter == "race" {
		rows, qErr := db.DB.Query("SELECT id, name FROM compendium_races WHERE name LIKE ? ORDER BY name LIMIT 10", "%"+q+"%")
		if qErr != nil {
			middleware.LogWarn("compendium", "query failed", "type", "race", "error", qErr)
		} else {
			for rows.Next() {
				var r SearchResult
				if err := rows.Scan(&r.ID, &r.Name); err != nil {
					middleware.LogWarn("compendium", "scan failed, skipping row", "type", "race", "error", err)
					continue
				}
				r.Type = "race"
				results = append(results, r)
			}
			rows.Close()
		}
	}

	if typeFilter == "" || typeFilter == "feat" {
		rows, qErr := db.DB.Query("SELECT id, name FROM compendium_feats WHERE name LIKE ? ORDER BY name LIMIT 10", "%"+q+"%")
		if qErr != nil {
			middleware.LogWarn("compendium", "query failed", "type", "feat", "error", qErr)
		} else {
			for rows.Next() {
				var r SearchResult
				if err := rows.Scan(&r.ID, &r.Name); err != nil {
					middleware.LogWarn("compendium", "scan failed, skipping row", "type", "feat", "error", err)
					continue
				}
				r.Type = "feat"
				results = append(results, r)
			}
			rows.Close()
		}
	}

	if typeFilter == "" || typeFilter == "background" {
		rows, qErr := db.DB.Query("SELECT id, name FROM compendium_backgrounds WHERE name LIKE ? ORDER BY name LIMIT 10", "%"+q+"%")
		if qErr != nil {
			middleware.LogWarn("compendium", "query failed", "type", "background", "error", qErr)
		} else {
			for rows.Next() {
				var r SearchResult
				if err := rows.Scan(&r.ID, &r.Name); err != nil {
					middleware.LogWarn("compendium", "scan failed, skipping row", "type", "background", "error", err)
					continue
				}
				r.Type = "background"
				results = append(results, r)
			}
			rows.Close()
		}
	}

	if typeFilter == "" || typeFilter == "class" {
		rows, qErr := db.DB.Query("SELECT id, name, hit_die, primary_ability FROM compendium_classes WHERE name LIKE ? ORDER BY name LIMIT 10", "%"+q+"%")
		if qErr != nil {
			middleware.LogWarn("compendium", "query failed", "type", "class", "error", qErr)
		} else {
			for rows.Next() {
				var r SearchResult
				if err := rows.Scan(&r.ID, &r.Name, &r.HitDie, &r.PrimaryAbility); err != nil {
					middleware.LogWarn("compendium", "scan failed, skipping row", "type", "class", "error", err)
					continue
				}
				r.Type = "class"
				results = append(results, r)
			}
			rows.Close()
		}
	}

	// Unified schema-based entries (compendium_entries FTS5) — appended so players
	// can discover DM-imported content through the same search endpoint.
	ftsWhere := "compendium_entries_fts MATCH ?"
	ftsArgs := []any{q}
	if typeFilter != "" {
		ftsWhere += " AND cs.type_name = ?"
		ftsArgs = append(ftsArgs, typeFilter)
	}
	rows, qErr := db.DB.Query("SELECT e.id, e.schema_id, e.data, cs.type_name FROM compendium_entries e JOIN compendium_entries_fts f ON e.id = f.rowid JOIN compendium_schemas cs ON e.schema_id = cs.id WHERE "+ftsWhere+" ORDER BY rank LIMIT 25", ftsArgs...)
	if qErr != nil {
		middleware.LogWarn("compendium", "unified search failed", "error", qErr)
	} else {
		for rows.Next() {
			var r SearchResult
			var schemaID int64
			var dataJSON string
			if err := rows.Scan(&r.ID, &schemaID, &r.Type, &dataJSON); err != nil {
				middleware.LogWarn("compendium", "scan failed, skipping row", "type", "unified", "error", err)
				continue
			}
			var data map[string]any
			if err := json.Unmarshal([]byte(dataJSON), &data); err == nil {
				if n, ok := data["name"].(string); ok {
					r.Name = n
				}
			}
			results = append(results, r)
		}
		rows.Close()
	}

	c.JSON(http.StatusOK, results)
}

// ─── User-Facing Schema-Based Compendium Entries ───

type userCompendiumSchemaEntry struct {
	ID          int64                 `json:"id"`
	TypeName    string                `json:"type_name"`
	DisplayName string                `json:"display_name"`
	EntryCount  int                   `json:"entry_count"`
	Entries     []userCompendiumEntry `json:"entries"`
}

type userCompendiumEntry struct {
	ID        int64          `json:"id"`
	Data      map[string]any `json:"data"`
	CreatedAt string         `json:"created_at"`
}

// ListUserCompendiumEntriesBySchema returns all schemas with non-zero entry counts
// and their first 20 entries each. Accessible to any authenticated user.
func ListUserCompendiumEntriesBySchema(c *gin.Context) {
	rows, err := db.DB.Query("SELECT id, type_name, display_name FROM compendium_schemas ORDER BY display_name")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	result := make([]userCompendiumSchemaEntry, 0)

	for rows.Next() {
		var s userCompendiumSchemaEntry
		if err := rows.Scan(&s.ID, &s.TypeName, &s.DisplayName); err != nil {
			continue
		}

		// Count entries for this schema
		var count int
		db.DB.QueryRow("SELECT COUNT(*) FROM compendium_entries WHERE schema_id=?", s.ID).Scan(&count)
		if count == 0 {
			continue // skip empty schemas
		}
		s.EntryCount = count

		// Fetch first 20 entries
		entryRows, err := db.DB.Query(
			"SELECT id, data, created_at FROM compendium_entries WHERE schema_id=? ORDER BY created_at DESC LIMIT 20",
			s.ID)
		if err != nil {
			continue
		}

		s.Entries = make([]userCompendiumEntry, 0)
		for entryRows.Next() {
			var e userCompendiumEntry
			var dataJSON, createdAt string
			if err := entryRows.Scan(&e.ID, &dataJSON, &createdAt); err != nil {
				continue
			}
			e.Data = make(map[string]any)
			json.Unmarshal([]byte(dataJSON), &e.Data)
			e.CreatedAt = createdAt
			s.Entries = append(s.Entries, e)
		}
		entryRows.Close()

		result = append(result, s)
	}

	c.JSON(http.StatusOK, gin.H{"schemas": result})
}

// FetchFromDnDApi fetches compendium data from the D&D 5e API as fallback
func FetchFromDnDApi(c *gin.Context) {
	category := c.Param("category")
	query := c.Query("q")

	if category == "" || query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "category and q parameters required"})
		return
	}

	// Map our categories to D&D API endpoints
	apiCategory := ""
	switch category {
	case "equipment":
		apiCategory = "equipment"
	case "spells":
		apiCategory = "spells"
	case "races":
		apiCategory = "races"
	case "classes":
		apiCategory = "classes"
	case "monsters":
		apiCategory = "monsters"
	case "features":
		apiCategory = "features"
	case "magic-items":
		apiCategory = "magic-items"
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown category: " + category})
		return
	}

	// Search D&D API
	searchURL := fmt.Sprintf("https://www.dnd5eapi.co/api/%s?name=%s", apiCategory, strings.ReplaceAll(query, " ", "+"))
	resp, err := http.Get(searchURL)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to reach D&D API: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read API response: " + err.Error()})
		return
	}

	// Parse the search results
	var searchResult struct {
		Count   int                 `json:"count"`
		Results []map[string]string `json:"results"`
	}
	if err := json.Unmarshal(body, &searchResult); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse API response: " + err.Error()})
		return
	}

	if searchResult.Count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "no results found"})
		return
	}

	// For each result, fetch the full details
	type DetailResult struct {
		Index string         `json:"index"`
		Name  string         `json:"name"`
		URL   string         `json:"url"`
		Data  map[string]any `json:"data"`
	}

	details := []DetailResult{}
	for _, r := range searchResult.Results {
		url := r["url"]
		if url == "" {
			continue
		}
		detailURL := fmt.Sprintf("https://www.dnd5eapi.co%s", url)
		detailResp, err := http.Get(detailURL)
		if err != nil {
			continue
		}
		var data map[string]any
		if err := json.NewDecoder(detailResp.Body).Decode(&data); err != nil {
			detailResp.Body.Close()
			continue
		}
		detailResp.Body.Close()
		details = append(details, DetailResult{
			Index: r["index"],
			Name:  r["name"],
			URL:   url,
			Data:  data,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"count":   len(details),
		"results": details,
		"source":  "dnd5eapi",
		"system":  "dnd5e",
	})
}

func ImportFromAPI(c *gin.Context) {
	url := c.PostForm("url")
	if url == "" {
		url = c.Query("url")
	}
	if url == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url parameter required"})
		return
	}

	// Fetch JSON from external API
	resp, err := http.Get(url)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to fetch URL: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("API returned status %d", resp.StatusCode)})
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read response: " + err.Error()})
		return
	}

	// Try as array first, then single object
	var chars []models.ImportCharacter
	if err := json.Unmarshal(body, &chars); err != nil {
		var single models.ImportCharacter
		if err2 := json.Unmarshal(body, &single); err2 != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "response is not valid character JSON: " + err.Error()})
			return
		}
		chars = []models.ImportCharacter{single}
	}

	userID, _ := c.Get("user_id")
	results := importCharacters(c.Request.Context(), userID.(int64), chars)
	c.JSON(http.StatusOK, results)
}
