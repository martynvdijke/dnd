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
	"villum/models"
)

func ListCompendiumRaces(c *gin.Context) {
	rows, err := db.DB.Query("SELECT id,name,description,speed,size,ability_bonuses,traits,languages,source_page,system,source FROM compendium_races ORDER BY name")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var out = make([]models.CompendiumRace, 0)
	for rows.Next() {
		var r models.CompendiumRace
		rows.Scan(&r.ID, &r.Name, &r.Description, &r.Speed, &r.Size, &r.AbilityBonuses, &r.Traits, &r.Languages, &r.SourcePage, &r.System, &r.Source)
		out = append(out, r)
	}
	c.JSON(http.StatusOK, out)
}

func ListCompendiumClasses(c *gin.Context) {
	rows, err := db.DB.Query("SELECT id,name,description,hit_die,primary_ability,saving_throws,proficiencies,spellcasting_ability,source_page,system,source FROM compendium_classes ORDER BY name")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var out = make([]models.CompendiumClass, 0)
	for rows.Next() {
		var cl models.CompendiumClass
		rows.Scan(&cl.ID, &cl.Name, &cl.Description, &cl.HitDie, &cl.PrimaryAbility, &cl.SavingThrows, &cl.Proficiencies, &cl.SpellcastingAbility, &cl.SourcePage, &cl.System, &cl.Source)
		out = append(out, cl)
	}
	c.JSON(http.StatusOK, out)
}

func ListCompendiumSpells(c *gin.Context) {
	query := "SELECT id,name,level,school,casting_time,range,components,duration,description,higher_levels,classes,source_page,system,source FROM compendium_spells WHERE 1=1"
	args := []interface{}{}

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
		rows.Scan(&s.ID, &s.Name, &s.Level, &s.School, &s.CastingTime, &s.Range, &s.Components, &s.Duration, &s.Description, &s.HigherLevels, &s.Classes, &s.SourcePage, &s.System, &s.Source)
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
		rows.Scan(&f.ID, &f.Name, &f.Description, &f.Prerequisites, &f.SourcePage, &f.System, &f.Source)
		out = append(out, f)
	}
	c.JSON(http.StatusOK, out)
}

func ListCompendiumBackgrounds(c *gin.Context) {
	rows, err := db.DB.Query("SELECT id,name,description,feature_name,feature_description,proficiencies,source_page,system,source FROM compendium_backgrounds ORDER BY name")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var out = make([]models.CompendiumBackground, 0)
	for rows.Next() {
		var b models.CompendiumBackground
		rows.Scan(&b.ID, &b.Name, &b.Description, &b.FeatureName, &b.FeatureDescription, &b.Proficiencies, &b.SourcePage, &b.System, &b.Source)
		out = append(out, b)
	}
	c.JSON(http.StatusOK, out)
}

func ListCompendiumEquipment(c *gin.Context) {
	query := "SELECT id,name,category,cost,weight,description,source_page,system,source FROM compendium_equipment WHERE 1=1"
	args := []interface{}{}

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
	var out = make([]models.CompendiumEquipment, 0)
	for rows.Next() {
		var e models.CompendiumEquipment
		rows.Scan(&e.ID, &e.Name, &e.Category, &e.Cost, &e.Weight, &e.Description, &e.SourcePage, &e.System, &e.Source)
		out = append(out, e)
	}
	c.JSON(http.StatusOK, out)
}

func AdminCreateCompendiumEntry(c *gin.Context) {
	entryType := c.Param("type")
	switch entryType {
	case "spells":
		var s models.CompendiumSpell
		if err := c.ShouldBindJSON(&s); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if s.System == "" {
			s.System = "dnd5e"
		}
		if s.Source == "" {
			s.Source = "srd"
		}
		result, err := db.DB.Exec(`INSERT INTO compendium_spells(name,level,school,casting_time,range,components,duration,description,higher_levels,classes,system,source) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`,
			s.Name, s.Level, s.School, s.CastingTime, s.Range, s.Components, s.Duration, s.Description, s.HigherLevels, s.Classes, s.System, s.Source)
		if err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		id, _ := result.LastInsertId()
		c.JSON(http.StatusCreated, gin.H{"id": id})
	case "races":
		var r models.CompendiumRace
		if err := c.ShouldBindJSON(&r); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if r.System == "" {
			r.System = "dnd5e"
		}
		if r.Source == "" {
			r.Source = "srd"
		}
		result, err := db.DB.Exec(`INSERT INTO compendium_races(name,description,speed,size,ability_bonuses,traits,languages,system,source) VALUES(?,?,?,?,?,?,?,?,?)`,
			r.Name, r.Description, r.Speed, r.Size, r.AbilityBonuses, r.Traits, r.Languages, r.System, r.Source)
		if err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		id, _ := result.LastInsertId()
		c.JSON(http.StatusCreated, gin.H{"id": id})
	case "classes":
		var cl models.CompendiumClass
		if err := c.ShouldBindJSON(&cl); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if cl.System == "" {
			cl.System = "dnd5e"
		}
		if cl.Source == "" {
			cl.Source = "srd"
		}
		result, err := db.DB.Exec(`INSERT INTO compendium_classes(name,description,hit_die,primary_ability,saving_throws,proficiencies,spellcasting_ability,system,source) VALUES(?,?,?,?,?,?,?,?,?)`,
			cl.Name, cl.Description, cl.HitDie, cl.PrimaryAbility, cl.SavingThrows, cl.Proficiencies, cl.SpellcastingAbility, cl.System, cl.Source)
		if err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		id, _ := result.LastInsertId()
		c.JSON(http.StatusCreated, gin.H{"id": id})
	case "feats":
		var f models.CompendiumFeat
		if err := c.ShouldBindJSON(&f); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if f.System == "" {
			f.System = "dnd5e"
		}
		if f.Source == "" {
			f.Source = "srd"
		}
		result, err := db.DB.Exec(`INSERT INTO compendium_feats(name,description,prerequisites,system,source) VALUES(?,?,?,?,?)`,
			f.Name, f.Description, f.Prerequisites, f.System, f.Source)
		if err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		id, _ := result.LastInsertId()
		c.JSON(http.StatusCreated, gin.H{"id": id})
	case "backgrounds":
		var b models.CompendiumBackground
		if err := c.ShouldBindJSON(&b); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if b.System == "" {
			b.System = "dnd5e"
		}
		if b.Source == "" {
			b.Source = "srd"
		}
		result, err := db.DB.Exec(`INSERT INTO compendium_backgrounds(name,description,feature_name,feature_description,proficiencies,system,source) VALUES(?,?,?,?,?,?,?)`,
			b.Name, b.Description, b.FeatureName, b.FeatureDescription, b.Proficiencies, b.System, b.Source)
		if err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		id, _ := result.LastInsertId()
		c.JSON(http.StatusCreated, gin.H{"id": id})
	case "equipment":
		var e models.CompendiumEquipment
		if err := c.ShouldBindJSON(&e); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if e.System == "" {
			e.System = "dnd5e"
		}
		if e.Source == "" {
			e.Source = "srd"
		}
		result, err := db.DB.Exec(`INSERT INTO compendium_equipment(name,category,cost,weight,description,system,source) VALUES(?,?,?,?,?,?,?)`,
			e.Name, e.Category, e.Cost, e.Weight, e.Description, e.System, e.Source)
		if err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		id, _ := result.LastInsertId()
		c.JSON(http.StatusCreated, gin.H{"id": id})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown type: " + entryType})
	}
}

func AdminUpdateCompendiumEntry(c *gin.Context) {
	entryType := c.Param("type")
	entryID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	switch entryType {
	case "spells":
		var s models.CompendiumSpell
		if err := c.ShouldBindJSON(&s); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		db.DB.Exec(`UPDATE compendium_spells SET name=?,level=?,school=?,casting_time=?,range=?,components=?,duration=?,description=?,higher_levels=?,classes=?,system=?,source=? WHERE id=?`,
			s.Name, s.Level, s.School, s.CastingTime, s.Range, s.Components, s.Duration, s.Description, s.HigherLevels, s.Classes, s.System, s.Source, entryID)
	case "races":
		var r models.CompendiumRace
		if err := c.ShouldBindJSON(&r); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		db.DB.Exec(`UPDATE compendium_races SET name=?,description=?,speed=?,size=?,ability_bonuses=?,traits=?,languages=?,system=?,source=? WHERE id=?`,
			r.Name, r.Description, r.Speed, r.Size, r.AbilityBonuses, r.Traits, r.Languages, r.System, r.Source, entryID)
	case "classes":
		var cl models.CompendiumClass
		if err := c.ShouldBindJSON(&cl); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		db.DB.Exec(`UPDATE compendium_classes SET name=?,description=?,hit_die=?,primary_ability=?,saving_throws=?,proficiencies=?,spellcasting_ability=?,system=?,source=? WHERE id=?`,
			cl.Name, cl.Description, cl.HitDie, cl.PrimaryAbility, cl.SavingThrows, cl.Proficiencies, cl.SpellcastingAbility, cl.System, cl.Source, entryID)
	case "feats":
		var f models.CompendiumFeat
		if err := c.ShouldBindJSON(&f); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		db.DB.Exec(`UPDATE compendium_feats SET name=?,description=?,prerequisites=?,system=?,source=? WHERE id=?`,
			f.Name, f.Description, f.Prerequisites, f.System, f.Source, entryID)
	case "backgrounds":
		var b models.CompendiumBackground
		if err := c.ShouldBindJSON(&b); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		db.DB.Exec(`UPDATE compendium_backgrounds SET name=?,description=?,feature_name=?,feature_description=?,proficiencies=?,system=?,source=? WHERE id=?`,
			b.Name, b.Description, b.FeatureName, b.FeatureDescription, b.Proficiencies, b.System, b.Source, entryID)
	case "equipment":
		var e models.CompendiumEquipment
		if err := c.ShouldBindJSON(&e); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		db.DB.Exec(`UPDATE compendium_equipment SET name=?,category=?,cost=?,weight=?,description=?,system=?,source=? WHERE id=?`,
			e.Name, e.Category, e.Cost, e.Weight, e.Description, e.System, e.Source, entryID)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown type"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func AdminDeleteCompendiumEntry(c *gin.Context) {
	entryType := c.Param("type")
	entryID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	switch entryType {
	case "spells":
		db.DB.Exec("DELETE FROM compendium_spells WHERE id=?", entryID)
	case "races":
		db.DB.Exec("DELETE FROM compendium_races WHERE id=?", entryID)
	case "classes":
		db.DB.Exec("DELETE FROM compendium_classes WHERE id=?", entryID)
	case "feats":
		db.DB.Exec("DELETE FROM compendium_feats WHERE id=?", entryID)
	case "backgrounds":
		db.DB.Exec("DELETE FROM compendium_backgrounds WHERE id=?", entryID)
	case "equipment":
		db.DB.Exec("DELETE FROM compendium_equipment WHERE id=?", entryID)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown type"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func SearchCompendium(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query required"})
		return
	}

	type SearchResult struct {
		Type string `json:"type"`
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}

	results := []SearchResult{}

	rows, _ := db.DB.Query("SELECT id, name FROM compendium_spells WHERE name LIKE ? LIMIT 10", "%"+q+"%")
	if rows != nil {
		for rows.Next() {
			var r SearchResult
			rows.Scan(&r.ID, &r.Name)
			r.Type = "spell"
			results = append(results, r)
		}
		rows.Close()
	}

	rows, _ = db.DB.Query("SELECT id, name FROM compendium_equipment WHERE name LIKE ? LIMIT 10", "%"+q+"%")
	if rows != nil {
		for rows.Next() {
			var r SearchResult
			rows.Scan(&r.ID, &r.Name)
			r.Type = "equipment"
			results = append(results, r)
		}
		rows.Close()
	}

	rows, _ = db.DB.Query("SELECT id, name FROM compendium_races WHERE name LIKE ? LIMIT 5", "%"+q+"%")
	if rows != nil {
		for rows.Next() {
			var r SearchResult
			rows.Scan(&r.ID, &r.Name)
			r.Type = "race"
			results = append(results, r)
		}
		rows.Close()
	}

	rows, _ = db.DB.Query("SELECT id, name FROM compendium_feats WHERE name LIKE ? LIMIT 5", "%"+q+"%")
	if rows != nil {
		for rows.Next() {
			var r SearchResult
			rows.Scan(&r.ID, &r.Name)
			r.Type = "feat"
			results = append(results, r)
		}
		rows.Close()
	}

	c.JSON(http.StatusOK, results)
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
		Count    int                    `json:"count"`
		Results  []map[string]string    `json:"results"`
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
		Index string                 `json:"index"`
		Name  string                 `json:"name"`
		URL   string                 `json:"url"`
		Data  map[string]interface{} `json:"data"`
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
		var data map[string]interface{}
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
