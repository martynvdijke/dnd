package handlers

import (
	"embed"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

//go:embed templates/*.html
var htmxTemplatesFS embed.FS

var htmxTemplates *template.Template

func init() {
	funcMap := template.FuncMap{
		"title":   strings.Title,
		"lower":   strings.ToLower,
		"seq":     seq,
		"mul":     func(a, b int) int { return a * b },
		"div":     func(a, b int) int { if b == 0 { return 0 }; return a / b },
		"sub":     func(a, b int) int { return a - b },
		"add":     func(a, b int) int { return a + b },
		"truncate": truncate,
		"capitalize": strings.Title,
		"sign": func(n int) string { if n >= 0 { return "+" + strconv.Itoa(n) }; return strconv.Itoa(n) },
	}
	htmxTemplates = template.Must(template.New("").Funcs(funcMap).ParseFS(htmxTemplatesFS, "templates/*.html"))
}

func seq(from, to int) []int {
	var out []int
	for i := from; i <= to; i++ {
		out = append(out, i)
	}
	return out
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func renderTemplate(c *gin.Context, name string, data interface{}) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := htmxTemplates.ExecuteTemplate(c.Writer, name, data); err != nil {
		c.String(http.StatusInternalServerError, "template error: %v", err)
	}
}

func getInt64Param(c *gin.Context, name string) (int64, error) {
	return strconv.ParseInt(c.PostForm(name), 10, 64)
}

func getQueryInt64(c *gin.Context, name string) (int64, error) {
	return strconv.ParseInt(c.Query(name), 10, 64)
}

func getIntParam(c *gin.Context, name string, def int) int {
	v := c.PostForm(name)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func getFloatParam(c *gin.Context, name string, def float64) float64 {
	v := c.PostForm(name)
	if v == "" {
		return def
	}
	n, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return def
	}
	return n
}

// ─── Notes ───

type htmxNoteData struct {
	CharacterID int64
	Note        *models.CharacterNote
	Notes       []models.CharacterNote
	Grouped     map[string][]models.CharacterNote
}

func HtmxListNotes(c *gin.Context) {
	charID := c.Query("character_id")
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	if charID == "" {
		c.String(http.StatusBadRequest, "character_id required")
		return
	}
	rows, err := db.DB.Query("SELECT id, character_id, title, content, visibility, category FROM character_notes WHERE character_id=? ORDER BY created_at DESC", charID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	var notes []models.CharacterNote
	for rows.Next() {
		var n models.CharacterNote
		rows.Scan(&n.ID, &n.CharacterID, &n.Title, &n.Content, &n.Visibility, &n.Category)
		if n.Visibility == "dm" && role != "admin" {
			var ownerID int64
			db.DB.QueryRow("SELECT user_id FROM characters WHERE id=?", charID).Scan(&ownerID)
			if ownerID != userID {
				var isDM bool
				db.DB.QueryRow(`SELECT COUNT(*) > 0 FROM campaign_members cm JOIN characters c ON c.campaign_id = cm.campaign_id WHERE c.id=? AND cm.user_id=? AND cm.role='dm'`, charID, userID).Scan(&isDM)
				if !isDM {
					continue
				}
			}
		}
		notes = append(notes, n)
	}
	grouped := map[string][]models.CharacterNote{
		"general": {}, "backstory": {}, "quest": {}, "lore": {}, "dm": {}, "other": {},
	}
	for _, n := range notes {
		cat := n.Category
		if cat == "" { cat = "general" }
		if _, ok := grouped[cat]; ok {
			grouped[cat] = append(grouped[cat], n)
		} else {
			grouped["other"] = append(grouped["other"], n)
		}
	}
	cid, _ := strconv.ParseInt(charID, 10, 64)
	renderTemplate(c, "notes_list.html", htmxNoteData{
		CharacterID: cid,
		Notes:       notes,
		Grouped:     grouped,
	})
}

func HtmxNewNoteForm(c *gin.Context) {
	charID := c.Query("character_id")
	cid, _ := strconv.ParseInt(charID, 10, 64)
	renderTemplate(c, "notes_form.html", htmxNoteData{CharacterID: cid})
}

func HtmxEditNoteForm(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var n models.CharacterNote
	err := db.DB.QueryRow("SELECT id, character_id, title, content, visibility, category FROM character_notes WHERE id=?", id).Scan(&n.ID, &n.CharacterID, &n.Title, &n.Content, &n.Visibility, &n.Category)
	if err != nil {
		c.String(http.StatusNotFound, "not found")
		return
	}
	renderTemplate(c, "notes_form.html", htmxNoteData{
		CharacterID: n.CharacterID,
		Note:       &n,
	})
}

func HtmxCreateNote(c *gin.Context) {
	charID := c.PostForm("character_id")
	title := c.PostForm("title")
	content := c.PostForm("content")
	visibility := c.PostForm("visibility")
	category := c.PostForm("category")
	if title == "" {
		title = "Untitled Note"
	}
	db.DB.Exec("INSERT INTO character_notes(character_id,title,content,visibility,category) VALUES(?,?,?,?,?)", charID, title, content, visibility, category)
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListNotes(c)
}

func HtmxUpdateNote(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	charID := c.PostForm("character_id")
	title := c.PostForm("title")
	content := c.PostForm("content")
	visibility := c.PostForm("visibility")
	category := c.PostForm("category")
	db.DB.Exec("UPDATE character_notes SET title=?, content=?, visibility=?, category=?, updated_at=datetime('now') WHERE id=?", title, content, visibility, category, id)
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListNotes(c)
}

func HtmxDeleteNote(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var charID string
	db.DB.QueryRow("SELECT character_id FROM character_notes WHERE id=?", id).Scan(&charID)
	db.DB.Exec("DELETE FROM character_notes WHERE id=?", id)
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListNotes(c)
}

// ─── Feats ───

type htmxFeatData struct {
	CharacterID int64
	Feat        *models.CharacterFeat
	Feats       []models.CharacterFeat
}

func HtmxListFeats(c *gin.Context) {
	charID := c.Query("character_id")
	if charID == "" {
		c.String(http.StatusBadRequest, "character_id required")
		return
	}
	rows, err := db.DB.Query("SELECT id, character_id, name, description, prerequisites, source, level_gained FROM character_feats WHERE character_id=? ORDER BY level_gained, name", charID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	var feats []models.CharacterFeat
	for rows.Next() {
		var f models.CharacterFeat
		rows.Scan(&f.ID, &f.CharacterID, &f.Name, &f.Description, &f.Prerequisites, &f.Source, &f.LevelGained)
		feats = append(feats, f)
	}
	cid, _ := strconv.ParseInt(charID, 10, 64)
	renderTemplate(c, "feats_list.html", htmxFeatData{CharacterID: cid, Feats: feats})
}

func HtmxNewFeatForm(c *gin.Context) {
	charID := c.Query("character_id")
	cid, _ := strconv.ParseInt(charID, 10, 64)
	renderTemplate(c, "feats_form.html", htmxFeatData{CharacterID: cid})
}

func HtmxEditFeatForm(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var f models.CharacterFeat
	err := db.DB.QueryRow("SELECT id, character_id, name, description, prerequisites, source, level_gained FROM character_feats WHERE id=?", id).Scan(&f.ID, &f.CharacterID, &f.Name, &f.Description, &f.Prerequisites, &f.Source, &f.LevelGained)
	if err != nil {
		c.String(http.StatusNotFound, "not found")
		return
	}
	renderTemplate(c, "feats_form.html", htmxFeatData{CharacterID: f.CharacterID, Feat: &f})
}

func HtmxCreateFeat(c *gin.Context) {
	charID := c.PostForm("character_id")
	if name := c.PostForm("name"); name == "" {
		c.String(http.StatusBadRequest, "name required")
		return
	}
	db.DB.Exec("INSERT INTO character_feats(character_id,name,description,prerequisites,source,level_gained) VALUES(?,?,?,?,?,?)",
		charID, c.PostForm("name"), c.PostForm("description"), c.PostForm("prerequisites"), c.PostForm("source"), getIntParam(c, "level_gained", 1))
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListFeats(c)
}

func HtmxUpdateFeat(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	charID := c.PostForm("character_id")
	db.DB.Exec("UPDATE character_feats SET name=?, description=?, prerequisites=?, source=?, level_gained=? WHERE id=?",
		c.PostForm("name"), c.PostForm("description"), c.PostForm("prerequisites"), c.PostForm("source"), getIntParam(c, "level_gained", 1), id)
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListFeats(c)
}

func HtmxDeleteFeat(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var charID string
	db.DB.QueryRow("SELECT character_id FROM character_feats WHERE id=?", id).Scan(&charID)
	db.DB.Exec("DELETE FROM character_feats WHERE id=?", id)
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListFeats(c)
}

// ─── Conditions ───

type htmxConditionData struct {
	CharacterID int64
	Condition   *models.Condition
	Conditions  []models.Condition
}

func HtmxListConditions(c *gin.Context) {
	charID := c.Query("character_id")
	if charID == "" {
		c.String(http.StatusBadRequest, "character_id required")
		return
	}
	rows, err := db.DB.Query("SELECT id, character_id, name, type, source, duration, duration_type, saving_throw, save_dc, description, started_at FROM character_conditions WHERE character_id=? ORDER BY started_at DESC", charID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	var conds []models.Condition
	for rows.Next() {
		var cond models.Condition
		rows.Scan(&cond.ID, &cond.CharacterID, &cond.Name, &cond.Type, &cond.Source, &cond.Duration, &cond.DurationType, &cond.SavingThrow, &cond.SaveDC, &cond.Description, &cond.StartedAt)
		conds = append(conds, cond)
	}
	cid, _ := strconv.ParseInt(charID, 10, 64)
	renderTemplate(c, "conditions_list.html", htmxConditionData{CharacterID: cid, Conditions: conds})
}

func HtmxNewConditionForm(c *gin.Context) {
	charID := c.Query("character_id")
	cid, _ := strconv.ParseInt(charID, 10, 64)
	renderTemplate(c, "conditions_form.html", htmxConditionData{CharacterID: cid})
}

func HtmxEditConditionForm(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var cond models.Condition
	err := db.DB.QueryRow("SELECT id, character_id, name, type, source, duration, duration_type, saving_throw, save_dc, description, started_at FROM character_conditions WHERE id=?", id).Scan(&cond.ID, &cond.CharacterID, &cond.Name, &cond.Type, &cond.Source, &cond.Duration, &cond.DurationType, &cond.SavingThrow, &cond.SaveDC, &cond.Description, &cond.StartedAt)
	if err != nil {
		c.String(http.StatusNotFound, "not found")
		return
	}
	renderTemplate(c, "conditions_form.html", htmxConditionData{CharacterID: cond.CharacterID, Condition: &cond})
}

func HtmxCreateCondition(c *gin.Context) {
	charID := c.PostForm("character_id")
	name := c.PostForm("name")
	if name == "" {
		c.String(http.StatusBadRequest, "name required")
		return
	}
	duration := getIntParam(c, "duration", 0)
	if duration < 0 {
		duration = 0
	}
	db.DB.Exec("INSERT INTO character_conditions(character_id,name,type,source,duration,duration_type,saving_throw,save_dc,description) VALUES(?,?,?,?,?,?,?,?,?)",
		charID, name, c.PostForm("type"), c.PostForm("source"), duration, c.PostForm("duration_type"), c.PostForm("saving_throw"), getIntParam(c, "save_dc", 0), c.PostForm("description"))
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListConditions(c)
}

func HtmxUpdateCondition(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	charID := c.PostForm("character_id")
	db.DB.Exec("UPDATE character_conditions SET name=?, type=?, source=?, duration=?, duration_type=?, saving_throw=?, save_dc=?, description=? WHERE id=?",
		c.PostForm("name"), c.PostForm("type"), c.PostForm("source"), getIntParam(c, "duration", 0), c.PostForm("duration_type"), c.PostForm("saving_throw"), getIntParam(c, "save_dc", 0), c.PostForm("description"), id)
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListConditions(c)
}

func HtmxDeleteCondition(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var charID string
	db.DB.QueryRow("SELECT character_id FROM character_conditions WHERE id=?", id).Scan(&charID)
	db.DB.Exec("DELETE FROM character_conditions WHERE id=?", id)
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListConditions(c)
}

// ─── Companions ───

type htmxCompanionData struct {
	CharacterID int64
	Companion   *models.Companion
	Companions  []models.Companion
}

func HtmxListCompanions(c *gin.Context) {
	charID := c.Query("character_id")
	if charID == "" {
		c.String(http.StatusBadRequest, "character_id required")
		return
	}
	rows, err := db.DB.Query("SELECT id, character_id, name, type, race, hp_max, hp_current, ac, str, dex, con, int, wis, cha, speed, abilities, notes, portrait_url, is_alive FROM companions WHERE character_id=? ORDER BY name", charID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	var comps []models.Companion
	for rows.Next() {
		var comp models.Companion
		var isAlive int
		rows.Scan(&comp.ID, &comp.CharacterID, &comp.Name, &comp.Type, &comp.Race, &comp.HPMax, &comp.HPCurrent, &comp.AC,
			&comp.Str, &comp.Dex, &comp.Con, &comp.Int, &comp.Wis, &comp.Cha, &comp.Speed,
			&comp.Abilities, &comp.Notes, &comp.PortraitURL, &isAlive)
		comp.IsAlive = isAlive == 1
		comps = append(comps, comp)
	}
	cid, _ := strconv.ParseInt(charID, 10, 64)
	renderTemplate(c, "companions_list.html", htmxCompanionData{CharacterID: cid, Companions: comps})
}

func HtmxNewCompanionForm(c *gin.Context) {
	charID := c.Query("character_id")
	cid, _ := strconv.ParseInt(charID, 10, 64)
	renderTemplate(c, "companions_form.html", htmxCompanionData{CharacterID: cid})
}

func HtmxEditCompanionForm(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var comp models.Companion
	var isAlive int
	err := db.DB.QueryRow("SELECT id, character_id, name, type, race, hp_max, hp_current, ac, str, dex, con, int, wis, cha, speed, abilities, notes, portrait_url, is_alive FROM companions WHERE id=?", id).Scan(
		&comp.ID, &comp.CharacterID, &comp.Name, &comp.Type, &comp.Race, &comp.HPMax, &comp.HPCurrent, &comp.AC,
		&comp.Str, &comp.Dex, &comp.Con, &comp.Int, &comp.Wis, &comp.Cha, &comp.Speed,
		&comp.Abilities, &comp.Notes, &comp.PortraitURL, &isAlive)
	if err != nil {
		c.String(http.StatusNotFound, "not found")
		return
	}
	comp.IsAlive = isAlive == 1
	renderTemplate(c, "companions_form.html", htmxCompanionData{CharacterID: comp.CharacterID, Companion: &comp})
}

func HtmxCreateCompanion(c *gin.Context) {
	charID := c.PostForm("character_id")
	name := c.PostForm("name")
	if name == "" {
		c.String(http.StatusBadRequest, "name required")
		return
	}
	isAlive := 0
	if c.PostForm("is_alive") == "true" {
		isAlive = 1
	}
	hp := getIntParam(c, "hp_max", 10)
	db.DB.Exec(`INSERT INTO companions(character_id,name,type,race,hp_max,hp_current,ac,str,dex,con,int,wis,cha,speed,abilities,notes,portrait_url,is_alive) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		charID, name, c.PostForm("type"), c.PostForm("race"), hp, getIntParam(c, "hp_current", hp),
		getIntParam(c, "ac", 10), getIntParam(c, "str", 10), getIntParam(c, "dex", 10), getIntParam(c, "con", 10),
		getIntParam(c, "int", 10), getIntParam(c, "wis", 10), getIntParam(c, "cha", 10), getIntParam(c, "speed", 30),
		c.PostForm("abilities"), c.PostForm("notes"), c.PostForm("portrait_url"), isAlive)
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListCompanions(c)
}

func HtmxUpdateCompanion(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	charID := c.PostForm("character_id")
	isAlive := 0
	if c.PostForm("is_alive") == "true" {
		isAlive = 1
	}
	hp := getIntParam(c, "hp_max", 10)
	db.DB.Exec(`UPDATE companions SET name=?, type=?, race=?, hp_max=?, hp_current=?, ac=?, str=?, dex=?, con=?, int=?, wis=?, cha=?, speed=?, abilities=?, notes=?, portrait_url=?, is_alive=? WHERE id=?`,
		c.PostForm("name"), c.PostForm("type"), c.PostForm("race"), hp, getIntParam(c, "hp_current", hp),
		getIntParam(c, "ac", 10), getIntParam(c, "str", 10), getIntParam(c, "dex", 10), getIntParam(c, "con", 10),
		getIntParam(c, "int", 10), getIntParam(c, "wis", 10), getIntParam(c, "cha", 10), getIntParam(c, "speed", 30),
		c.PostForm("abilities"), c.PostForm("notes"), c.PostForm("portrait_url"), isAlive, id)
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListCompanions(c)
}

func HtmxDeleteCompanion(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var charID string
	db.DB.QueryRow("SELECT character_id FROM companions WHERE id=?", id).Scan(&charID)
	db.DB.Exec("DELETE FROM companions WHERE id=?", id)
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListCompanions(c)
}

// ─── Features ───

type htmxFeatureData struct {
	CharacterID  int64
	Features     []models.Feature
	Proficiencies []models.Proficiency
}

func HtmxListFeatures(c *gin.Context) {
	charID := c.Query("character_id")
	if charID == "" {
		c.String(http.StatusBadRequest, "character_id required")
		return
	}
	frows, err := db.DB.Query("SELECT id, character_id, name, description, source, level_gained FROM character_features WHERE character_id=? ORDER BY level_gained, name", charID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer frows.Close()
	var feats []models.Feature
	for frows.Next() {
		var f models.Feature
		frows.Scan(&f.ID, &f.CharacterID, &f.Name, &f.Description, &f.Source, &f.LevelGained)
		feats = append(feats, f)
	}
	prows, err := db.DB.Query("SELECT id, character_id, type, name FROM character_proficiencies WHERE character_id=? ORDER BY type, name", charID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer prows.Close()
	var profs []models.Proficiency
	for prows.Next() {
		var p models.Proficiency
		prows.Scan(&p.ID, &p.CharacterID, &p.Type, &p.Name)
		profs = append(profs, p)
	}
	cid, _ := strconv.ParseInt(charID, 10, 64)
	renderTemplate(c, "features_list.html", htmxFeatureData{CharacterID: cid, Features: feats, Proficiencies: profs})
}

func HtmxNewFeatureForm(c *gin.Context) {
	charID := c.Query("character_id")
	cid, _ := strconv.ParseInt(charID, 10, 64)
	renderTemplate(c, "features_form.html", htmxFeatureData{CharacterID: cid})
}

func HtmxCreateFeature(c *gin.Context) {
	charID := c.PostForm("character_id")
	db.DB.Exec("INSERT INTO character_features(character_id,name,description,source,level_gained) VALUES(?,?,?,?,?)",
		charID, c.PostForm("name"), c.PostForm("description"), c.PostForm("source"), getIntParam(c, "level_gained", 1))
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListFeatures(c)
}

func HtmxDeleteFeature(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var charID string
	db.DB.QueryRow("SELECT character_id FROM character_features WHERE id=?", id).Scan(&charID)
	db.DB.Exec("DELETE FROM character_features WHERE id=?", id)
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListFeatures(c)
}

// ─── Proficiencies ───

func HtmxNewProficiencyForm(c *gin.Context) {
	charID := c.Query("character_id")
	cid, _ := strconv.ParseInt(charID, 10, 64)
	renderTemplate(c, "proficiencies_form.html", htmxFeatureData{CharacterID: cid})
}

func HtmxCreateProficiency(c *gin.Context) {
	charID := c.PostForm("character_id")
	db.DB.Exec("INSERT INTO character_proficiencies(character_id,type,name) VALUES(?,?,?)",
		charID, c.PostForm("type"), c.PostForm("name"))
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListFeatures(c)
}

func HtmxDeleteProficiency(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var charID string
	db.DB.QueryRow("SELECT character_id FROM character_proficiencies WHERE id=?", id).Scan(&charID)
	db.DB.Exec("DELETE FROM character_proficiencies WHERE id=?", id)
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListFeatures(c)
}

// ─── Inventory ───

type htmxInventoryData struct {
	CharacterID int64
	Item        *models.InventoryItem
	Items       []models.InventoryItem
}

func HtmxListInventory(c *gin.Context) {
	charID := c.Query("character_id")
	if charID == "" {
		c.String(http.StatusBadRequest, "character_id required")
		return
	}
	rows, err := db.DB.Query("SELECT id, character_id, name, quantity, weight, category, description, is_equipped, is_magical, attunement FROM inventory WHERE character_id=? ORDER BY category, name", charID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	var items []models.InventoryItem
	for rows.Next() {
		var i models.InventoryItem
		rows.Scan(&i.ID, &i.CharacterID, &i.Name, &i.Quantity, &i.Weight, &i.Category, &i.Description, &i.IsEquipped, &i.IsMagical, &i.Attunement)
		items = append(items, i)
	}
	cid, _ := strconv.ParseInt(charID, 10, 64)
	renderTemplate(c, "inventory_list.html", htmxInventoryData{CharacterID: cid, Items: items})
}

func HtmxNewInventoryForm(c *gin.Context) {
	charID := c.Query("character_id")
	cid, _ := strconv.ParseInt(charID, 10, 64)
	renderTemplate(c, "inventory_form.html", htmxInventoryData{CharacterID: cid})
}

func HtmxEditInventoryForm(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var i models.InventoryItem
	err := db.DB.QueryRow("SELECT id, character_id, name, quantity, weight, category, description, is_equipped, is_magical, attunement FROM inventory WHERE id=?", id).Scan(
		&i.ID, &i.CharacterID, &i.Name, &i.Quantity, &i.Weight, &i.Category, &i.Description, &i.IsEquipped, &i.IsMagical, &i.Attunement)
	if err != nil {
		c.String(http.StatusNotFound, "not found")
		return
	}
	renderTemplate(c, "inventory_form.html", htmxInventoryData{CharacterID: i.CharacterID, Item: &i})
}

func HtmxCreateInventory(c *gin.Context) {
	charID := c.PostForm("character_id")
	db.DB.Exec("INSERT INTO inventory(character_id,name,description,quantity,weight,category) VALUES(?,?,?,?,?,?)",
		charID, c.PostForm("name"), c.PostForm("description"), getIntParam(c, "quantity", 1), getFloatParam(c, "weight", 0), c.PostForm("category"))
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListInventory(c)
}

func HtmxUpdateInventory(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	charID := c.PostForm("character_id")
	db.DB.Exec("UPDATE inventory SET name=?, description=?, quantity=?, weight=?, category=? WHERE id=?",
		c.PostForm("name"), c.PostForm("description"), getIntParam(c, "quantity", 1), getFloatParam(c, "weight", 0), c.PostForm("category"), id)
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListInventory(c)
}

func HtmxDeleteInventory(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var charID string
	db.DB.QueryRow("SELECT character_id FROM inventory WHERE id=?", id).Scan(&charID)
	db.DB.Exec("DELETE FROM inventory WHERE id=?", id)
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListInventory(c)
}

// ─── Spells ───

type htmxSpellData struct {
	CharacterID int64
	Spell       *models.Spell
	Spells      []models.Spell
}

func HtmxListSpells(c *gin.Context) {
	charID := c.Query("character_id")
	if charID == "" {
		c.String(http.StatusBadRequest, "character_id required")
		return
	}
	rows, err := db.DB.Query("SELECT id, character_id, name, level, school, casting_time, range, components, duration, description, prepared, always_prepared, source, notes FROM spells WHERE character_id=? ORDER BY level, name", charID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	var spells []models.Spell
	for rows.Next() {
		var s models.Spell
		rows.Scan(&s.ID, &s.CharacterID, &s.Name, &s.Level, &s.School, &s.CastingTime, &s.Range, &s.Components, &s.Duration, &s.Description, &s.Prepared, &s.AlwaysPrepared, &s.Source, &s.Notes)
		spells = append(spells, s)
	}
	cid, _ := strconv.ParseInt(charID, 10, 64)
	renderTemplate(c, "spells_list.html", htmxSpellData{CharacterID: cid, Spells: spells})
}

func HtmxNewSpellForm(c *gin.Context) {
	charID := c.Query("character_id")
	cid, _ := strconv.ParseInt(charID, 10, 64)
	renderTemplate(c, "spells_form.html", htmxSpellData{CharacterID: cid})
}

func HtmxEditSpellForm(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var s models.Spell
	err := db.DB.QueryRow("SELECT id, character_id, name, level, school, casting_time, range, components, duration, description, prepared, always_prepared, source, notes FROM spells WHERE id=?", id).Scan(
		&s.ID, &s.CharacterID, &s.Name, &s.Level, &s.School, &s.CastingTime, &s.Range, &s.Components, &s.Duration, &s.Description, &s.Prepared, &s.AlwaysPrepared, &s.Source, &s.Notes)
	if err != nil {
		c.String(http.StatusNotFound, "not found")
		return
	}
	renderTemplate(c, "spells_form.html", htmxSpellData{CharacterID: s.CharacterID, Spell: &s})
}

func HtmxCreateSpell(c *gin.Context) {
	charID := c.PostForm("character_id")
	db.DB.Exec("INSERT INTO spells(character_id,name,level,school,casting_time,range,components,duration,description) VALUES(?,?,?,?,?,?,?,?,?)",
		charID, c.PostForm("name"), getIntParam(c, "level", 0), c.PostForm("school"), c.PostForm("casting_time"), c.PostForm("range"), c.PostForm("components"), c.PostForm("duration"), c.PostForm("description"))
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListSpells(c)
}

func HtmxUpdateSpell(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	charID := c.PostForm("character_id")
	db.DB.Exec("UPDATE spells SET name=?, level=?, school=?, casting_time=?, range=?, components=?, duration=?, description=? WHERE id=?",
		c.PostForm("name"), getIntParam(c, "level", 0), c.PostForm("school"), c.PostForm("casting_time"), c.PostForm("range"), c.PostForm("components"), c.PostForm("duration"), c.PostForm("description"), id)
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListSpells(c)
}

func HtmxDeleteSpell(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var charID string
	db.DB.QueryRow("SELECT character_id FROM spells WHERE id=?", id).Scan(&charID)
	db.DB.Exec("DELETE FROM spells WHERE id=?", id)
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListSpells(c)
}

// ─── NPCs (linked to character) ───

type htmxNPCData struct {
	CharacterID int64
	NPCs        []npcsLink
	AllNPCs     []models.NPC
}

type npcsLink struct {
	models.CharacterNPC
	NPCName  string `json:"npc_name"`
	NPCRace  string `json:"npc_race"`
	NPCClass string `json:"npc_class"`
	NPHPMax  int    `json:"npc_hp_max"`
	NPHPCurr int    `json:"npc_hp_current"`
	NPCAlive bool   `json:"npc_is_alive"`
}

func HtmxListNPCs(c *gin.Context) {
	charID := c.Query("character_id")
	if charID == "" {
		c.String(http.StatusBadRequest, "character_id required")
		return
	}
	rows, err := db.DB.Query(`
		SELECT cn.id, cn.character_id, cn.npc_id, cn.relationship, cn.notes,
			cn.interaction_count, cn.last_interacted,
			n.name, n.race, n.class, n.hp_max, n.hp_current, n.is_alive
		FROM character_npcs cn JOIN npcs n ON cn.npc_id = n.id
		WHERE cn.character_id=? ORDER BY cn.interaction_count DESC`, charID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	var links []npcsLink
	for rows.Next() {
		var nl npcsLink
		rows.Scan(&nl.ID, &nl.CharacterID, &nl.NPCID, &nl.Relationship, &nl.Notes,
			&nl.InteractionCount, &nl.LastInteracted,
			&nl.NPCName, &nl.NPCRace, &nl.NPCClass, &nl.NPHPMax, &nl.NPHPCurr, &nl.NPCAlive)
		links = append(links, nl)
	}
	cid, _ := strconv.ParseInt(charID, 10, 64)
	renderTemplate(c, "npcs_list.html", htmxNPCData{CharacterID: cid, NPCs: links})
}

func HtmxNewNPCForm(c *gin.Context) {
	charID := c.Query("character_id")
	cid, _ := strconv.ParseInt(charID, 10, 64)
	renderTemplate(c, "npcs_form.html", htmxNPCData{CharacterID: cid})
}

func HtmxLinkNPCForm(c *gin.Context) {
	charID := c.Query("character_id")
	cid, _ := strconv.ParseInt(charID, 10, 64)
	userID, _ := c.Get("user_id")
	rows, err := db.DB.Query("SELECT id, name, race, class, description, notes, str, dex, con, int, wis, cha, hp_max, hp_current, is_alive, created_at FROM npcs WHERE user_id=? ORDER BY name", userID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	var all []models.NPC
	for rows.Next() {
		var n models.NPC
		rows.Scan(&n.ID, &n.Name, &n.Race, &n.Class, &n.Description, &n.Notes,
			&n.Str, &n.Dex, &n.Con, &n.Int, &n.Wis, &n.Cha, &n.HPMax, &n.HPCurrent, &n.IsAlive, &n.CreatedAt)
		_ = n.UserID
		all = append(all, n)
	}
	renderTemplate(c, "npcs_link_form.html", htmxNPCData{CharacterID: cid, AllNPCs: all})
}

func HtmxCreateNPC(c *gin.Context) {
	charID := c.PostForm("character_id")
	userID, _ := c.Get("user_id")
	name := c.PostForm("name")
	if name == "" {
		c.String(http.StatusBadRequest, "name required")
		return
	}
	result, err := db.DB.Exec("INSERT INTO npcs(user_id,name,race,class,description,notes) VALUES(?,?,?,?,?,?)", userID, name, c.PostForm("race"), c.PostForm("class"), c.PostForm("description"), c.PostForm("notes"))
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	npcID, _ := result.LastInsertId()
	db.DB.Exec("INSERT INTO character_npcs(character_id,npc_id,relationship) VALUES(?,?,?)", charID, npcID, c.PostForm("type"))
	HtmxListNPCs(c)
}

func HtmxLinkNPC(c *gin.Context) {
	charID := c.PostForm("character_id")
	npcID := c.PostForm("npc_id")
	if npcID != "" {
		db.DB.Exec("INSERT INTO character_npcs(character_id,npc_id) VALUES(?,?)", charID, npcID)
	}
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListNPCs(c)
}

func HtmxUnlinkNPC(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var charID string
	db.DB.QueryRow("SELECT character_id FROM character_npcs WHERE id=?", id).Scan(&charID)
	db.DB.Exec("DELETE FROM character_npcs WHERE id=?", id)
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListNPCs(c)
}

// ─── Locations (linked to character) ───

type htmxLocationData struct {
	CharacterID int64
	Locations   []locLink
	AllLocations []models.Location
}

type locLink struct {
	models.CharacterLocation
	LocationName string `json:"location_name"`
	LocationType string `json:"location_type"`
	Description  string `json:"description"`
}

func HtmxListLocations(c *gin.Context) {
	charID := c.Query("character_id")
	if charID == "" {
		c.String(http.StatusBadRequest, "character_id required")
		return
	}
	rows, err := db.DB.Query(`
		SELECT cl.id, cl.character_id, cl.location_id, cl.relationship, cl.notes,
			l.name, l.type, l.description
		FROM character_locations cl JOIN locations l ON cl.location_id = l.id
		WHERE cl.character_id=? ORDER BY cl.relationship`, charID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	var links []locLink
	for rows.Next() {
		var ll locLink
		rows.Scan(&ll.ID, &ll.CharacterID, &ll.LocationID, &ll.Relationship, &ll.Notes,
			&ll.LocationName, &ll.LocationType, &ll.Description)
		links = append(links, ll)
	}
	cid, _ := strconv.ParseInt(charID, 10, 64)
	renderTemplate(c, "locations_list.html", htmxLocationData{CharacterID: cid, Locations: links})
}

func HtmxNewLocationForm(c *gin.Context) {
	charID := c.Query("character_id")
	cid, _ := strconv.ParseInt(charID, 10, 64)
	renderTemplate(c, "locations_form.html", htmxLocationData{CharacterID: cid})
}

func HtmxLinkLocationForm(c *gin.Context) {
	charID := c.Query("character_id")
	cid, _ := strconv.ParseInt(charID, 10, 64)
	userID, _ := c.Get("user_id")
	rows, err := db.DB.Query("SELECT id, user_id, name, type, description, parent_id, latitude, longitude, created_at FROM locations WHERE user_id=? ORDER BY name", userID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	var all []models.Location
	for rows.Next() {
		var l models.Location
		rows.Scan(&l.ID, &l.UserID, &l.Name, &l.Type, &l.Description, &l.ParentID, &l.Latitude, &l.Longitude, &l.CreatedAt)
		all = append(all, l)
	}
	renderTemplate(c, "locations_link_form.html", htmxLocationData{CharacterID: cid, AllLocations: all})
}

func HtmxCreateLocation(c *gin.Context) {
	charID := c.PostForm("character_id")
	userID, _ := c.Get("user_id")
	name := c.PostForm("name")
	if name == "" { return }
	result, err := db.DB.Exec("INSERT INTO locations(user_id,name,type,description) VALUES(?,?,?,?)", userID, name, c.PostForm("type"), c.PostForm("description"))
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	locID, _ := result.LastInsertId()
	db.DB.Exec("INSERT INTO character_locations(character_id,location_id) VALUES(?,?)", charID, locID)
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListLocations(c)
}

func HtmxLinkLocation(c *gin.Context) {
	charID := c.PostForm("character_id")
	locID := c.PostForm("location_id")
	if locID != "" {
		db.DB.Exec("INSERT INTO character_locations(character_id,location_id) VALUES(?,?)", charID, locID)
	}
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListLocations(c)
}

func HtmxUnlinkLocation(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var charID string
	db.DB.QueryRow("SELECT character_id FROM character_locations WHERE id=?", id).Scan(&charID)
	db.DB.Exec("DELETE FROM character_locations WHERE id=?", id)
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListLocations(c)
}

// ─── Sessions ───

type htmxSessionData struct {
	CharacterID int64
	Session     *models.Session
	Sessions    []models.Session
}

func HtmxListSessions(c *gin.Context) {
	charID := c.Query("character_id")
	if charID == "" {
		c.String(http.StatusBadRequest, "character_id required")
		return
	}
	rows, err := db.DB.Query("SELECT id, character_id, session_date, title, notes, xp_earned, gold_earned, important_events, created_at FROM sessions WHERE character_id=? ORDER BY session_date DESC", charID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	var sessions []models.Session
	for rows.Next() {
		var s models.Session
		rows.Scan(&s.ID, &s.CharacterID, &s.SessionDate, &s.Title, &s.Notes, &s.XPEarned, &s.GoldEarned, &s.ImportantEvents, &s.CreatedAt)
		sessions = append(sessions, s)
	}
	cid, _ := strconv.ParseInt(charID, 10, 64)
	renderTemplate(c, "sessions_list.html", htmxSessionData{CharacterID: cid, Sessions: sessions})
}

func HtmxNewSessionForm(c *gin.Context) {
	charID := c.Query("character_id")
	cid, _ := strconv.ParseInt(charID, 10, 64)
	renderTemplate(c, "sessions_form.html", htmxSessionData{CharacterID: cid})
}

func HtmxEditSessionForm(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var s models.Session
	err := db.DB.QueryRow("SELECT id, character_id, session_date, title, notes, xp_earned, gold_earned, important_events, created_at FROM sessions WHERE id=?", id).Scan(
		&s.ID, &s.CharacterID, &s.SessionDate, &s.Title, &s.Notes, &s.XPEarned, &s.GoldEarned, &s.ImportantEvents, &s.CreatedAt)
	if err != nil {
		c.String(http.StatusNotFound, "not found")
		return
	}
	renderTemplate(c, "sessions_form.html", htmxSessionData{CharacterID: s.CharacterID, Session: &s})
}

func HtmxCreateSession(c *gin.Context) {
	charID := c.PostForm("character_id")
	db.DB.Exec("INSERT INTO sessions(character_id,session_date,title,notes,xp_earned,gold_earned,important_events) VALUES(?,?,?,?,?,?,?)",
		charID, c.PostForm("session_date"), c.PostForm("title"), c.PostForm("notes"), getIntParam(c, "xp_earned", 0), getIntParam(c, "gold_earned", 0), c.PostForm("important_events"))
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListSessions(c)
}

func HtmxUpdateSession(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	charID := c.PostForm("character_id")
	db.DB.Exec("UPDATE sessions SET session_date=?, title=?, notes=?, xp_earned=?, gold_earned=?, important_events=? WHERE id=?",
		c.PostForm("session_date"), c.PostForm("title"), c.PostForm("notes"), getIntParam(c, "xp_earned", 0), getIntParam(c, "gold_earned", 0), c.PostForm("important_events"), id)
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListSessions(c)
}

func HtmxDeleteSession(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var charID string
	db.DB.QueryRow("SELECT character_id FROM sessions WHERE id=?", id).Scan(&charID)
	db.DB.Exec("DELETE FROM sessions WHERE id=?", id)
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListSessions(c)
}

// ─── Quests ───

type htmxQuestData struct {
	CharacterID int64
	Quest       *models.Quest
	Quests      []models.Quest
}

func HtmxListQuests(c *gin.Context) {
	charID := c.Query("character_id")
	if charID == "" {
		c.String(http.StatusBadRequest, "character_id required")
		return
	}
	rows, err := db.DB.Query("SELECT id, character_id, name, description, status, objectives, rewards, notes, created_at, updated_at FROM quests WHERE character_id=? ORDER BY status, name", charID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	var quests []models.Quest
	for rows.Next() {
		var q models.Quest
		rows.Scan(&q.ID, &q.CharacterID, &q.Name, &q.Description, &q.Status, &q.Objectives, &q.Rewards, &q.Notes, &q.CreatedAt, &q.UpdatedAt)
		quests = append(quests, q)
	}
	cid, _ := strconv.ParseInt(charID, 10, 64)
	renderTemplate(c, "quests_list.html", htmxQuestData{CharacterID: cid, Quests: quests})
}

func HtmxNewQuestForm(c *gin.Context) {
	charID := c.Query("character_id")
	cid, _ := strconv.ParseInt(charID, 10, 64)
	renderTemplate(c, "quests_form.html", htmxQuestData{CharacterID: cid})
}

func HtmxEditQuestForm(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var q models.Quest
	err := db.DB.QueryRow("SELECT id, character_id, name, description, status, objectives, rewards, notes, created_at, updated_at FROM quests WHERE id=?", id).Scan(
		&q.ID, &q.CharacterID, &q.Name, &q.Description, &q.Status, &q.Objectives, &q.Rewards, &q.Notes, &q.CreatedAt, &q.UpdatedAt)
	if err != nil {
		c.String(http.StatusNotFound, "not found")
		return
	}
	renderTemplate(c, "quests_form.html", htmxQuestData{CharacterID: q.CharacterID, Quest: &q})
}

func HtmxCreateQuest(c *gin.Context) {
	charID := c.PostForm("character_id")
	status := c.PostForm("status")
	if status == "" {
		status = "active"
	}
	db.DB.Exec("INSERT INTO quests(character_id,name,description,status,objectives,rewards,notes) VALUES(?,?,?,?,?,?,?)",
		charID, c.PostForm("name"), c.PostForm("description"), status, c.PostForm("objectives"), c.PostForm("rewards"), c.PostForm("notes"))
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListQuests(c)
}

func HtmxUpdateQuest(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	charID := c.PostForm("character_id")
	status := c.PostForm("status")
	if status == "" {
		status = "active"
	}
	db.DB.Exec("UPDATE quests SET name=?, description=?, status=?, objectives=?, rewards=?, notes=?, updated_at=datetime('now') WHERE id=?",
		c.PostForm("name"), c.PostForm("description"), status, c.PostForm("objectives"), c.PostForm("rewards"), c.PostForm("notes"), id)
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListQuests(c)
}

func HtmxDeleteQuest(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var charID string
	db.DB.QueryRow("SELECT character_id FROM quests WHERE id=?", id).Scan(&charID)
	db.DB.Exec("DELETE FROM quests WHERE id=?", id)
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListQuests(c)
}

// ─── Journal ───

type htmxJournalData struct {
	CharacterID int64
	Entry       *models.JournalEntry
	Entries     []models.JournalEntry
}

func HtmxListJournal(c *gin.Context) {
	charID := c.Query("character_id")
	if charID == "" {
		c.String(http.StatusBadRequest, "character_id required")
		return
	}
	rows, err := db.DB.Query("SELECT id, character_id, title, entry, entry_date, created_at FROM journal WHERE character_id=? ORDER BY entry_date DESC", charID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	var entries []models.JournalEntry
	for rows.Next() {
		var j models.JournalEntry
		rows.Scan(&j.ID, &j.CharacterID, &j.Title, &j.Entry, &j.EntryDate, &j.CreatedAt)
		entries = append(entries, j)
	}
	cid, _ := strconv.ParseInt(charID, 10, 64)
	renderTemplate(c, "journal_list.html", htmxJournalData{CharacterID: cid, Entries: entries})
}

func HtmxNewJournalForm(c *gin.Context) {
	charID := c.Query("character_id")
	cid, _ := strconv.ParseInt(charID, 10, 64)
	renderTemplate(c, "journal_form.html", htmxJournalData{CharacterID: cid})
}

func HtmxEditJournalForm(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var j models.JournalEntry
	err := db.DB.QueryRow("SELECT id, character_id, title, entry, entry_date, created_at FROM journal WHERE id=?", id).Scan(
		&j.ID, &j.CharacterID, &j.Title, &j.Entry, &j.EntryDate, &j.CreatedAt)
	if err != nil {
		c.String(http.StatusNotFound, "not found")
		return
	}
	renderTemplate(c, "journal_form.html", htmxJournalData{CharacterID: j.CharacterID, Entry: &j})
}

func HtmxCreateJournal(c *gin.Context) {
	charID := c.PostForm("character_id")
	db.DB.Exec("INSERT INTO journal(character_id,title,entry,entry_date) VALUES(?,?,?,?)",
		charID, c.PostForm("title"), c.PostForm("entry"), c.PostForm("entry_date"))
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListJournal(c)
}

func HtmxUpdateJournal(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	charID := c.PostForm("character_id")
	db.DB.Exec("UPDATE journal SET title=?, entry=?, entry_date=? WHERE id=?",
		c.PostForm("title"), c.PostForm("entry"), c.PostForm("entry_date"), id)
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListJournal(c)
}

func HtmxDeleteJournal(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var charID string
	db.DB.QueryRow("SELECT character_id FROM journal WHERE id=?", id).Scan(&charID)
	db.DB.Exec("DELETE FROM journal WHERE id=?", id)
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListJournal(c)
}

// ─── Calendar ───

type htmxCalendarData struct {
	Event  *models.CalendarEvent
	Events []models.CalendarEvent
}

func HtmxListCalendar(c *gin.Context) {
	rows, err := db.DB.Query("SELECT id, campaign_id, title, description, event_date, event_type, color, created_at FROM campaign_calendar_events ORDER BY event_date DESC")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	var events []models.CalendarEvent
	for rows.Next() {
		var e models.CalendarEvent
		rows.Scan(&e.ID, &e.CampaignID, &e.Title, &e.Description, &e.EventDate, &e.EventType, &e.Color, &e.CreatedAt)
		events = append(events, e)
	}
	renderTemplate(c, "calendar_list.html", htmxCalendarData{Events: events})
}

func HtmxNewCalendarForm(c *gin.Context) {
	renderTemplate(c, "calendar_form.html", htmxCalendarData{})
}

func HtmxEditCalendarForm(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var e models.CalendarEvent
	err := db.DB.QueryRow("SELECT id, campaign_id, title, description, event_date, event_type, color, created_at FROM campaign_calendar_events WHERE id=?", id).Scan(
		&e.ID, &e.CampaignID, &e.Title, &e.Description, &e.EventDate, &e.EventType, &e.Color, &e.CreatedAt)
	if err != nil {
		c.String(http.StatusNotFound, "not found")
		return
	}
	renderTemplate(c, "calendar_form.html", htmxCalendarData{Event: &e})
}

func HtmxCreateCalendar(c *gin.Context) {
	db.DB.Exec("INSERT INTO campaign_calendar_events(campaign_id,title,description,event_date,event_type) VALUES(?,?,?,?,?)",
		getIntParam(c, "campaign_id", 1), c.PostForm("title"), c.PostForm("description"), c.PostForm("event_date"), c.PostForm("event_type"))
	HtmxListCalendar(c)
}

func HtmxUpdateCalendar(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	db.DB.Exec("UPDATE campaign_calendar_events SET title=?, description=?, event_date=?, event_type=?, campaign_id=? WHERE id=?",
		c.PostForm("title"), c.PostForm("description"), c.PostForm("event_date"), c.PostForm("event_type"), getIntParam(c, "campaign_id", 1), id)
	HtmxListCalendar(c)
}

func HtmxDeleteCalendar(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	db.DB.Exec("DELETE FROM campaign_calendar_events WHERE id=?", id)
	HtmxListCalendar(c)
}

// ─── Timeline ───

type htmxTimelineData struct {
	Event  *models.TimelineEvent
	Events []models.TimelineEvent
}

func HtmxListTimeline(c *gin.Context) {
	rows, err := db.DB.Query("SELECT id, campaign_id, title, description, event_date, event_type, importance, icon, linked_entity_type, linked_entity_id, created_at FROM campaign_timeline_events ORDER BY event_date DESC")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	var events []models.TimelineEvent
	for rows.Next() {
		var e models.TimelineEvent
		rows.Scan(&e.ID, &e.CampaignID, &e.Title, &e.Description, &e.EventDate, &e.EventType, &e.Importance, &e.Icon, &e.LinkedEntityType, &e.LinkedEntityID, &e.CreatedAt)
		events = append(events, e)
	}
	renderTemplate(c, "timeline_list.html", htmxTimelineData{Events: events})
}

func HtmxNewTimelineForm(c *gin.Context) {
	renderTemplate(c, "timeline_form.html", htmxTimelineData{})
}

func HtmxEditTimelineForm(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var e models.TimelineEvent
	err := db.DB.QueryRow("SELECT id, campaign_id, title, description, event_date, event_type, importance, icon, linked_entity_type, linked_entity_id, created_at FROM campaign_timeline_events WHERE id=?", id).Scan(
		&e.ID, &e.CampaignID, &e.Title, &e.Description, &e.EventDate, &e.EventType, &e.Importance, &e.Icon, &e.LinkedEntityType, &e.LinkedEntityID, &e.CreatedAt)
	if err != nil {
		c.String(http.StatusNotFound, "not found")
		return
	}
	renderTemplate(c, "timeline_form.html", htmxTimelineData{Event: &e})
}

func HtmxCreateTimeline(c *gin.Context) {
	db.DB.Exec("INSERT INTO campaign_timeline_events(campaign_id,title,description,event_date,event_type) VALUES(?,?,?,?,?)",
		getIntParam(c, "campaign_id", 1), c.PostForm("title"), c.PostForm("description"), c.PostForm("event_date"), c.PostForm("event_type"))
	HtmxListTimeline(c)
}

func HtmxUpdateTimeline(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	db.DB.Exec("UPDATE campaign_timeline_events SET title=?, description=?, event_date=?, event_type=?, campaign_id=? WHERE id=?",
		c.PostForm("title"), c.PostForm("description"), c.PostForm("event_date"), c.PostForm("event_type"), getIntParam(c, "campaign_id", 1), id)
	HtmxListTimeline(c)
}

func HtmxDeleteTimeline(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	db.DB.Exec("DELETE FROM campaign_timeline_events WHERE id=?", id)
	HtmxListTimeline(c)
}

// ─── Factions ───

type htmxFactionData struct {
	Faction  *models.Faction
	Factions []models.Faction
}

func HtmxListFactions(c *gin.Context) {
	rows, err := db.DB.Query("SELECT id, campaign_id, name, description, type, headquarters FROM factions ORDER BY name")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	var factions []models.Faction
	for rows.Next() {
		var f models.Faction
		rows.Scan(&f.ID, &f.CampaignID, &f.Name, &f.Description, &f.Type, &f.Headquarters)
		factions = append(factions, f)
	}
	renderTemplate(c, "factions_list.html", htmxFactionData{Factions: factions})
}

func HtmxNewFactionForm(c *gin.Context) {
	renderTemplate(c, "factions_form.html", htmxFactionData{})
}

func HtmxEditFactionForm(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var f models.Faction
	err := db.DB.QueryRow("SELECT id, campaign_id, name, description, type, headquarters FROM factions WHERE id=?", id).Scan(
		&f.ID, &f.CampaignID, &f.Name, &f.Description, &f.Type, &f.Headquarters)
	if err != nil {
		c.String(http.StatusNotFound, "not found")
		return
	}
	renderTemplate(c, "factions_form.html", htmxFactionData{Faction: &f})
}

func HtmxCreateFaction(c *gin.Context) {
	db.DB.Exec("INSERT INTO factions(campaign_id,name,description,type,headquarters) VALUES(?,?,?,?,?)",
		getIntParam(c, "campaign_id", 1), c.PostForm("name"), c.PostForm("description"), c.PostForm("type"), c.PostForm("headquarters"))
	HtmxListFactions(c)
}

func HtmxUpdateFaction(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	db.DB.Exec("UPDATE factions SET name=?, description=?, type=?, headquarters=? WHERE id=?",
		c.PostForm("name"), c.PostForm("description"), c.PostForm("type"), c.PostForm("headquarters"), id)
	HtmxListFactions(c)
}

func HtmxDeleteFaction(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	db.DB.Exec("DELETE FROM factions WHERE id=?", id)
	HtmxListFactions(c)
}

// ─── Settings export for use in main.go ───

func HtmxRegisterRoutes(r *gin.RouterGroup) {
	routes := []struct{ method, path string; handler gin.HandlerFunc }{
		// Notes
		{"GET", "/htmx/notes", HtmxListNotes},
		{"GET", "/htmx/notes/new", HtmxNewNoteForm},
		{"GET", "/htmx/notes/:id/edit", HtmxEditNoteForm},
		{"POST", "/htmx/notes", HtmxCreateNote},
		{"PUT", "/htmx/notes/:id", HtmxUpdateNote},
		{"DELETE", "/htmx/notes/:id", HtmxDeleteNote},

		// Feats
		{"GET", "/htmx/feats", HtmxListFeats},
		{"GET", "/htmx/feats/new", HtmxNewFeatForm},
		{"GET", "/htmx/feats/:id/edit", HtmxEditFeatForm},
		{"POST", "/htmx/feats", HtmxCreateFeat},
		{"PUT", "/htmx/feats/:id", HtmxUpdateFeat},
		{"DELETE", "/htmx/feats/:id", HtmxDeleteFeat},

		// Conditions
		{"GET", "/htmx/conditions", HtmxListConditions},
		{"GET", "/htmx/conditions/new", HtmxNewConditionForm},
		{"GET", "/htmx/conditions/:id/edit", HtmxEditConditionForm},
		{"POST", "/htmx/conditions", HtmxCreateCondition},
		{"PUT", "/htmx/conditions/:id", HtmxUpdateCondition},
		{"DELETE", "/htmx/conditions/:id", HtmxDeleteCondition},

		// Companions
		{"GET", "/htmx/companions", HtmxListCompanions},
		{"GET", "/htmx/companions/new", HtmxNewCompanionForm},
		{"GET", "/htmx/companions/:id/edit", HtmxEditCompanionForm},
		{"POST", "/htmx/companions", HtmxCreateCompanion},
		{"PUT", "/htmx/companions/:id", HtmxUpdateCompanion},
		{"DELETE", "/htmx/companions/:id", HtmxDeleteCompanion},

		// Features
		{"GET", "/htmx/features", HtmxListFeatures},
		{"GET", "/htmx/features/new", HtmxNewFeatureForm},
		{"POST", "/htmx/features", HtmxCreateFeature},
		{"DELETE", "/htmx/features/:id", HtmxDeleteFeature},

		// Proficiencies
		{"GET", "/htmx/proficiencies/new", HtmxNewProficiencyForm},
		{"POST", "/htmx/proficiencies", HtmxCreateProficiency},
		{"DELETE", "/htmx/proficiencies/:id", HtmxDeleteProficiency},

		// Inventory
		{"GET", "/htmx/inventory", HtmxListInventory},
		{"GET", "/htmx/inventory/new", HtmxNewInventoryForm},
		{"GET", "/htmx/inventory/:id/edit", HtmxEditInventoryForm},
		{"POST", "/htmx/inventory", HtmxCreateInventory},
		{"PUT", "/htmx/inventory/:id", HtmxUpdateInventory},
		{"DELETE", "/htmx/inventory/:id", HtmxDeleteInventory},

		// Spells
		{"GET", "/htmx/spells", HtmxListSpells},
		{"GET", "/htmx/spells/new", HtmxNewSpellForm},
		{"GET", "/htmx/spells/:id/edit", HtmxEditSpellForm},
		{"POST", "/htmx/spells", HtmxCreateSpell},
		{"PUT", "/htmx/spells/:id", HtmxUpdateSpell},
		{"DELETE", "/htmx/spells/:id", HtmxDeleteSpell},

		// NPCs
		{"GET", "/htmx/npcs", HtmxListNPCs},
		{"GET", "/htmx/npcs/new", HtmxNewNPCForm},
		{"GET", "/htmx/npcs/link", HtmxLinkNPCForm},
		{"POST", "/htmx/npcs", HtmxCreateNPC},
		{"POST", "/htmx/npcs/link", HtmxLinkNPC},
		{"DELETE", "/htmx/npcs/link/:id", HtmxUnlinkNPC},

		// Locations
		{"GET", "/htmx/locations", HtmxListLocations},
		{"GET", "/htmx/locations/new", HtmxNewLocationForm},
		{"GET", "/htmx/locations/link", HtmxLinkLocationForm},
		{"POST", "/htmx/locations", HtmxCreateLocation},
		{"POST", "/htmx/locations/link", HtmxLinkLocation},
		{"DELETE", "/htmx/locations/link/:id", HtmxUnlinkLocation},

		// Sessions
		{"GET", "/htmx/sessions", HtmxListSessions},
		{"GET", "/htmx/sessions/new", HtmxNewSessionForm},
		{"GET", "/htmx/sessions/:id/edit", HtmxEditSessionForm},
		{"POST", "/htmx/sessions", HtmxCreateSession},
		{"PUT", "/htmx/sessions/:id", HtmxUpdateSession},
		{"DELETE", "/htmx/sessions/:id", HtmxDeleteSession},

		// Quests
		{"GET", "/htmx/quests", HtmxListQuests},
		{"GET", "/htmx/quests/new", HtmxNewQuestForm},
		{"GET", "/htmx/quests/:id/edit", HtmxEditQuestForm},
		{"POST", "/htmx/quests", HtmxCreateQuest},
		{"PUT", "/htmx/quests/:id", HtmxUpdateQuest},
		{"DELETE", "/htmx/quests/:id", HtmxDeleteQuest},

		// Journal
		{"GET", "/htmx/journal", HtmxListJournal},
		{"GET", "/htmx/journal/new", HtmxNewJournalForm},
		{"GET", "/htmx/journal/:id/edit", HtmxEditJournalForm},
		{"POST", "/htmx/journal", HtmxCreateJournal},
		{"PUT", "/htmx/journal/:id", HtmxUpdateJournal},
		{"DELETE", "/htmx/journal/:id", HtmxDeleteJournal},

		// Calendar
		{"GET", "/htmx/calendar", HtmxListCalendar},
		{"GET", "/htmx/calendar/new", HtmxNewCalendarForm},
		{"GET", "/htmx/calendar/:id/edit", HtmxEditCalendarForm},
		{"POST", "/htmx/calendar", HtmxCreateCalendar},
		{"PUT", "/htmx/calendar/:id", HtmxUpdateCalendar},
		{"DELETE", "/htmx/calendar/:id", HtmxDeleteCalendar},

		// Timeline
		{"GET", "/htmx/timeline", HtmxListTimeline},
		{"GET", "/htmx/timeline/new", HtmxNewTimelineForm},
		{"GET", "/htmx/timeline/:id/edit", HtmxEditTimelineForm},
		{"POST", "/htmx/timeline", HtmxCreateTimeline},
		{"PUT", "/htmx/timeline/:id", HtmxUpdateTimeline},
		{"DELETE", "/htmx/timeline/:id", HtmxDeleteTimeline},

		// Factions
		{"GET", "/htmx/factions", HtmxListFactions},
		{"GET", "/htmx/factions/new", HtmxNewFactionForm},
		{"GET", "/htmx/factions/:id/edit", HtmxEditFactionForm},
		{"POST", "/htmx/factions", HtmxCreateFaction},
		{"PUT", "/htmx/factions/:id", HtmxUpdateFaction},
		{"DELETE", "/htmx/factions/:id", HtmxDeleteFaction},

		// One-Shot Adventures
		{"GET", "/htmx/oneshot-adventures", HtmxListOneShots},
		{"GET", "/htmx/oneshot-adventures/new", HtmxNewOneShotForm},
		{"GET", "/htmx/oneshot-adventures/:id", HtmxGetOneShotDetail},
		{"GET", "/htmx/oneshot-adventures/:id/edit", HtmxEditOneShotForm},
		{"POST", "/htmx/oneshot-adventures", HtmxCreateOneShot},
		{"POST", "/htmx/oneshot-adventures/generate", HtmxGenerateOneShot},
		{"PUT", "/htmx/oneshot-adventures/:id", HtmxUpdateOneShot},
		{"DELETE", "/htmx/oneshot-adventures/:id", HtmxDeleteOneShot},
		{"POST", "/htmx/oneshot-adventures/:id/acts", HtmxCreateAct},
		{"DELETE", "/htmx/oneshot-acts/:id", HtmxDeleteAct},
		{"POST", "/htmx/oneshot-acts/:id/scenes", HtmxCreateScene},
		{"DELETE", "/htmx/oneshot-scenes/:id", HtmxDeleteScene},

		// Session Pacing
		{"GET", "/htmx/session-pacing/:id", HtmxGetPacingDashboard},
	}
	for _, rt := range routes {
		r.Handle(rt.method, rt.path, rt.handler)
	}
}


