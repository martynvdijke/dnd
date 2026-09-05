package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"villum/db"
	"villum/models"
)

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
		if cat == "" {
			cat = "general"
		}
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
		Note:        &n,
	})
}

func HtmxCreateNote(c *gin.Context) {
	charID := c.PostForm("character_id")
	if !canEditCharacterID(c, int64(getIntParam(c, "character_id", 0))) {
		c.String(http.StatusForbidden, "access denied")
		return
	}
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
	if !canEditCharacterID(c, int64(getIntParam(c, "character_id", 0))) {
		c.String(http.StatusForbidden, "access denied")
		return
	}
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
	if !canEditCharacterID(c, int64(getIntParam(c, "character_id", 0))) {
		c.String(http.StatusForbidden, "access denied")
		return
	}
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
	if !canEditCharacterID(c, int64(getIntParam(c, "character_id", 0))) {
		c.String(http.StatusForbidden, "access denied")
		return
	}
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
	if !canEditCharacterID(c, int64(getIntParam(c, "character_id", 0))) {
		c.String(http.StatusForbidden, "access denied")
		return
	}
	name := c.PostForm("name")
	if name == "" {
		c.String(http.StatusBadRequest, "name required")
		return
	}
	duration := max(getIntParam(c, "duration", 0), 0)
	db.DB.Exec("INSERT INTO character_conditions(character_id,name,type,source,duration,duration_type,saving_throw,save_dc,description) VALUES(?,?,?,?,?,?,?,?,?)",
		charID, name, c.PostForm("type"), c.PostForm("source"), duration, c.PostForm("duration_type"), c.PostForm("saving_throw"), getIntParam(c, "save_dc", 0), c.PostForm("description"))
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListConditions(c)
}

func HtmxUpdateCondition(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	charID := c.PostForm("character_id")
	if !canEditCharacterID(c, int64(getIntParam(c, "character_id", 0))) {
		c.String(http.StatusForbidden, "access denied")
		return
	}
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
	if !canEditCharacterID(c, int64(getIntParam(c, "character_id", 0))) {
		c.String(http.StatusForbidden, "access denied")
		return
	}
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
	if !canEditCharacterID(c, int64(getIntParam(c, "character_id", 0))) {
		c.String(http.StatusForbidden, "access denied")
		return
	}
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
	CharacterID   int64
	Features      []models.Feature
	Proficiencies []models.Proficiency
}

func HtmxListFeatures(c *gin.Context) {
	charID := c.Query("character_id")
	if charID == "" {
		c.String(http.StatusBadRequest, "character_id required")
		return
	}
	renderHtmxFeaturesList(c, charID)
}

// renderHtmxFeaturesList renders the htmx features list partial for a character.
func renderHtmxFeaturesList(c *gin.Context, charID string) {
	frows, err := db.DB.Query("SELECT id, character_id, name, description, source, level_gained, COALESCE(compendium_entry_id,0) FROM character_features WHERE character_id=? ORDER BY level_gained, name", charID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer frows.Close()
	var feats []models.Feature
	for frows.Next() {
		var f models.Feature
		var entryID int64
		frows.Scan(&f.ID, &f.CharacterID, &f.Name, &f.Description, &f.Source, &f.LevelGained, &entryID)
		if entryID > 0 {
			f.CompendiumEntryID = &entryID
		}
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
	if !canEditCharacterID(c, int64(getIntParam(c, "character_id", 0))) {
		c.String(http.StatusForbidden, "access denied")
		return
	}
	if entryID, ok := formLinkID(c, "compendium_entry_id"); ok {
		cid, _ := strconv.ParseInt(charID, 10, 64)
		if st, msg := featureLinkInsert(cid, entryID, getIntParam(c, "level_gained", 1)); msg != "" {
			c.String(st, msg)
			return
		}
		c.Request.URL.RawQuery = "character_id=" + charID
		HtmxListFeatures(c)
		return
	}
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
	if !canEditCharacterID(c, int64(getIntParam(c, "character_id", 0))) {
		c.String(http.StatusForbidden, "access denied")
		return
	}
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

	renderHtmxInventoryList(c, charID)
}

// renderHtmxInventoryList renders the htmx inventory list partial for a character.
func renderHtmxInventoryList(c *gin.Context, charID string) {
	rows, err := db.DB.Query("SELECT id, character_id, name, quantity, weight, category, description, is_equipped, is_magical, attunement, COALESCE(compendium_equipment_id,0), COALESCE(compendium_entry_id,0) FROM inventory WHERE character_id=? ORDER BY category, name", charID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	var items []models.InventoryItem
	for rows.Next() {
		var i models.InventoryItem
		var compID, entryID int64
		rows.Scan(&i.ID, &i.CharacterID, &i.Name, &i.Quantity, &i.Weight, &i.Category, &i.Description, &i.IsEquipped, &i.IsMagical, &i.Attunement, &compID, &entryID)
		if compID > 0 {
			i.CompendiumEquipmentID = &compID
		}
		if entryID > 0 {
			i.CompendiumEntryID = &entryID
		}
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
	if !canEditCharacterID(c, int64(getIntParam(c, "character_id", 0))) {
		c.String(http.StatusForbidden, "access denied")
		return
	}
	db.DB.Exec("INSERT INTO inventory(character_id,name,description,quantity,weight,category) VALUES(?,?,?,?,?,?)",
		charID, c.PostForm("name"), c.PostForm("description"), getIntParam(c, "quantity", 1), getFloatParam(c, "weight", 0), c.PostForm("category"))
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListInventory(c)
}

func HtmxUpdateInventory(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	charID := c.PostForm("character_id")
	if !canEditCharacterID(c, int64(getIntParam(c, "character_id", 0))) {
		c.String(http.StatusForbidden, "access denied")
		return
	}
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

	renderHtmxSpellsList(c, charID)
}

// renderHtmxSpellsList renders the htmx spells list partial for a character.
func renderHtmxSpellsList(c *gin.Context, charID string) {
	rows, err := db.DB.Query("SELECT id, character_id, name, level, school, casting_time, range, components, duration, description, prepared, always_prepared, source, notes, COALESCE(compendium_spell_id,0), COALESCE(compendium_entry_id,0) FROM spells WHERE character_id=? ORDER BY level, name", charID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	var spells []models.Spell
	for rows.Next() {
		var s models.Spell
		var compID, entryID int64
		rows.Scan(&s.ID, &s.CharacterID, &s.Name, &s.Level, &s.School, &s.CastingTime, &s.Range, &s.Components, &s.Duration, &s.Description, &s.Prepared, &s.AlwaysPrepared, &s.Source, &s.Notes, &compID, &entryID)
		if compID > 0 {
			s.CompendiumSpellID = &compID
		}
		if entryID > 0 {
			s.CompendiumEntryID = &entryID
		}
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
	if !canEditCharacterID(c, int64(getIntParam(c, "character_id", 0))) {
		c.String(http.StatusForbidden, "access denied")
		return
	}
	db.DB.Exec("INSERT INTO spells(character_id,name,level,school,casting_time,range,components,duration,description) VALUES(?,?,?,?,?,?,?,?,?)",
		charID, c.PostForm("name"), getIntParam(c, "level", 0), c.PostForm("school"), c.PostForm("casting_time"), c.PostForm("range"), c.PostForm("components"), c.PostForm("duration"), c.PostForm("description"))
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListSpells(c)
}

func HtmxUpdateSpell(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	charID := c.PostForm("character_id")
	if !canEditCharacterID(c, int64(getIntParam(c, "character_id", 0))) {
		c.String(http.StatusForbidden, "access denied")
		return
	}
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
	NPCName        string `json:"npc_name"`
	NPCRace        string `json:"npc_race"`
	NPCClass       string `json:"npc_class"`
	NPHPMax        int    `json:"npc_hp_max"`
	NPHPCurr       int    `json:"npc_hp_current"`
	NPCAlive       bool   `json:"npc_is_alive"`
	NPCPortraitURL string `json:"npc_portrait_url"`
}
