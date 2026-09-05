package handlers

import (
	"math/rand"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

func HtmxListClues(c *gin.Context) {
	adventureID := c.Param("id")

	rows, err := db.DB.Query("SELECT id, adventure_id, title, description, clue_type, is_red_herring, is_revealed, sort_order, notes, created_at, updated_at FROM clues WHERE adventure_id=? ORDER BY sort_order, id", adventureID)
	if err != nil {
		c.String(http.StatusInternalServerError, "query error")
		return
	}
	defer rows.Close()

	clues := make([]models.Clue, 0)
	for rows.Next() {
		var cl models.Clue
		if rows.Scan(&cl.ID, &cl.AdventureID, &cl.Title, &cl.Description, &cl.ClueType, &cl.IsRedHerring, &cl.IsRevealed, &cl.SortOrder, &cl.Notes, &cl.CreatedAt, &cl.UpdatedAt) == nil {
			loadClueRelations(&cl)
			clues = append(clues, cl)
		}
	}

	renderTemplate(c, "oneshot_clues.html", gin.H{
		"Clues":       clues,
		"AdventureID": adventureID,
	})
}

func HtmxGetClueDetail(c *gin.Context) {
	id := c.Param("id")
	var cl models.Clue
	err := db.DB.QueryRow("SELECT id, adventure_id, title, description, clue_type, is_red_herring, is_revealed, sort_order, notes, created_at, updated_at FROM clues WHERE id=?", id).
		Scan(&cl.ID, &cl.AdventureID, &cl.Title, &cl.Description, &cl.ClueType, &cl.IsRedHerring, &cl.IsRevealed, &cl.SortOrder, &cl.Notes, &cl.CreatedAt, &cl.UpdatedAt)
	if err != nil {
		c.String(http.StatusNotFound, "Clue not found")
		return
	}
	loadClueRelations(&cl)

	renderTemplate(c, "oneshot_clue_detail.html", gin.H{
		"Clue": cl,
	})
}

func HtmxNewClueForm(c *gin.Context) {
	adventureID := c.Param("id")
	renderTemplate(c, "oneshot_clue_form.html", gin.H{
		"AdventureID": adventureID,
		"Clue":        nil,
	})
}

func HtmxCreateClue(c *gin.Context) {
	adventureID := c.Param("id")
	userID, _ := c.Get("user_id")

	var exists bool
	db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM oneshot_adventures WHERE id=? AND user_id=?)", adventureID, userID).Scan(&exists)
	if !exists {
		c.String(http.StatusNotFound, "Adventure not found")
		return
	}

	title := c.PostForm("title")
	description := c.PostForm("description")
	clueType := c.PostForm("clue_type")
	if clueType == "" {
		clueType = "direct"
	}
	sortOrder, _ := strconv.Atoi(c.PostForm("sort_order"))
	notes := c.PostForm("notes")

	db.DB.Exec("INSERT INTO clues(adventure_id, title, description, clue_type, is_red_herring, sort_order, notes) VALUES(?,?,?,?,?,?,?)",
		adventureID, title, description, clueType, 0, sortOrder, notes)

	HtmxListClues(c)
}

func HtmxEditClueForm(c *gin.Context) {
	id := c.Param("id")
	var cl models.Clue
	err := db.DB.QueryRow("SELECT id, adventure_id, title, description, clue_type, is_red_herring, is_revealed, sort_order, notes, created_at, updated_at FROM clues WHERE id=?", id).
		Scan(&cl.ID, &cl.AdventureID, &cl.Title, &cl.Description, &cl.ClueType, &cl.IsRedHerring, &cl.IsRevealed, &cl.SortOrder, &cl.Notes, &cl.CreatedAt, &cl.UpdatedAt)
	if err != nil {
		c.String(http.StatusNotFound, "Clue not found")
		return
	}
	renderTemplate(c, "oneshot_clue_form.html", gin.H{
		"AdventureID": cl.AdventureID,
		"Clue":        cl,
	})
}

func HtmxUpdateClue(c *gin.Context) {
	id := c.Param("id")
	title := c.PostForm("title")
	description := c.PostForm("description")
	clueType := c.PostForm("clue_type")
	sortOrder, _ := strconv.Atoi(c.PostForm("sort_order"))
	notes := c.PostForm("notes")

	db.DB.Exec("UPDATE clues SET title=?, description=?, clue_type=?, sort_order=?, notes=?, updated_at=datetime('now') WHERE id=?",
		title, description, clueType, sortOrder, notes, id)

	HtmxGetClueDetail(c)
}

func HtmxDeleteClue(c *gin.Context) {
	id := c.Param("id")
	// Get adventure_id for redirect
	var advID int64
	db.DB.QueryRow("SELECT adventure_id FROM clues WHERE id=?", id).Scan(&advID)
	db.DB.Exec("DELETE FROM clues WHERE id=?", id)
	if advID > 0 {
		HtmxListClues(c)
		return
	}
	c.String(http.StatusOK, "Deleted")
}

func HtmxRevealClue(c *gin.Context) {
	id := c.Param("id")
	db.DB.Exec("UPDATE clues SET is_revealed=1, updated_at=datetime('now') WHERE id=?", id)
	HtmxGetClueDetail(c)
}

func HtmxHideClue(c *gin.Context) {
	id := c.Param("id")
	db.DB.Exec("UPDATE clues SET is_revealed=0, updated_at=datetime('now') WHERE id=?", id)
	HtmxGetClueDetail(c)
}

// ─── Pregenerated Characters ───

func HtmxListPregens(c *gin.Context) {
	userID, _ := c.Get("user_id")
	rows, err := db.DB.Query("SELECT id, user_id, name, race, class, subclass, level, background, alignment, str, dex, con, int, wis, cha, hp, ac, speed, skills, equipment, spells, features, personality, backstory, portrait_url, notes, created_at, updated_at FROM pregen_characters WHERE user_id=? ORDER BY level DESC, name ASC", userID)
	if err != nil {
		c.String(http.StatusInternalServerError, "query error")
		return
	}
	defer rows.Close()
	chars := make([]models.PregeneratedCharacter, 0)
	for rows.Next() {
		var ch models.PregeneratedCharacter
		if err := rows.Scan(&ch.ID, &ch.UserID, &ch.Name, &ch.Race, &ch.Class, &ch.Subclass, &ch.Level, &ch.Background, &ch.Alignment,
			&ch.Str, &ch.Dex, &ch.Con, &ch.Int, &ch.Wis, &ch.Cha, &ch.HP, &ch.AC, &ch.Speed,
			&ch.Skills, &ch.Equipment, &ch.Spells, &ch.Features, &ch.Personality, &ch.Backstory,
			&ch.PortraitURL, &ch.Notes, &ch.CreatedAt, &ch.UpdatedAt); err == nil {
			chars = append(chars, ch)
		}
	}
	renderTemplate(c, "oneshot_pregens.html", gin.H{"Pregens": chars})
}

func HtmxGeneratePregen(c *gin.Context) {
	userID, _ := c.Get("user_id")
	name := c.Query("name")
	genLevel, _ := strconv.Atoi(c.DefaultQuery("level", "3"))
	if genLevel < 1 || genLevel > 20 {
		genLevel = 3
	}
	genRace := c.Query("race")
	genClass := c.Query("class")

	races := []string{"dwarf", "elf", "halfling", "human", "dragonborn", "gnome", "half-elf", "half-orc", "tiefling"}
	classes := []string{"barbarian", "fighter", "paladin", "cleric", "druid", "wizard", "sorcerer", "warlock", "bard", "rogue", "monk", "ranger", "artificer"}
	names := []string{"Aldric", "Briar", "Cassian", "Dorn", "Elara", "Finn", "Grom", "Halia", "Ivy", "Jax", "Kira", "Lark", "Mira", "Nyx", "Orin", "Piper", "Quinn", "Rook", "Sage", "Talon", "Una", "Vex", "Wren", "Xara", "Zephyr"}

	if genClass == "" {
		genClass = classes[rand.Intn(len(classes))]
	}
	if genRace == "" {
		genRace = races[rand.Intn(len(races))]
	}
	if name == "" {
		name = names[rand.Intn(len(names))]
	}

	var str, dex, con, intel, wis, cha int
	switch genClass {
	case "barbarian", "fighter", "paladin":
		str, dex, con, intel, wis, cha = 16, 12, 14, 8, 10, 10
	case "cleric", "druid":
		str, dex, con, intel, wis, cha = 12, 10, 14, 10, 16, 8
	case "wizard":
		str, dex, con, intel, wis, cha = 8, 12, 14, 16, 10, 10
	case "sorcerer", "warlock", "bard":
		str, dex, con, intel, wis, cha = 8, 12, 14, 10, 10, 16
	case "rogue", "ranger", "monk":
		str, dex, con, intel, wis, cha = 10, 16, 14, 10, 12, 8
	case "artificer":
		str, dex, con, intel, wis, cha = 8, 12, 14, 16, 10, 10
	default:
		str, dex, con, intel, wis, cha = 10, 10, 10, 10, 10, 10
	}

	hp := 8 + (con-10)/2 + (genLevel-1)*5
	ac := 12 + (dex-10)/2

	_, err := db.DB.Exec(`
		INSERT INTO pregen_characters(user_id, name, race, class, level, str, dex, con, int, wis, cha, hp, ac, speed)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,30)`,
		userID, name, genRace, genClass, genLevel, str, dex, con, intel, wis, cha, hp, ac)
	if err != nil {
		c.String(http.StatusInternalServerError, "insert error")
		return
	}

	HtmxListPregens(c)
}

func HtmxPregenCard(c *gin.Context) {
	id := c.Param("id")
	var ch models.PregeneratedCharacter
	err := db.DB.QueryRow("SELECT id, user_id, name, race, class, subclass, level, background, alignment, str, dex, con, int, wis, cha, hp, ac, speed, skills, equipment, spells, features, personality, backstory, portrait_url, notes, created_at, updated_at FROM pregen_characters WHERE id=?", id).
		Scan(&ch.ID, &ch.UserID, &ch.Name, &ch.Race, &ch.Class, &ch.Subclass, &ch.Level, &ch.Background, &ch.Alignment,
			&ch.Str, &ch.Dex, &ch.Con, &ch.Int, &ch.Wis, &ch.Cha, &ch.HP, &ch.AC, &ch.Speed,
			&ch.Skills, &ch.Equipment, &ch.Spells, &ch.Features, &ch.Personality, &ch.Backstory,
			&ch.PortraitURL, &ch.Notes, &ch.CreatedAt, &ch.UpdatedAt)
	if err != nil {
		c.String(http.StatusNotFound, "Pregen not found")
		return
	}
	renderTemplate(c, "oneshot_pregen_card.html", gin.H{"C": ch})
}

// ─── Prep Dashboard Handlers ───

func HtmxGetPrepDashboard(c *gin.Context) {
	adventureID := c.Param("id")
	userID, _ := c.Get("user_id")

	// Load adventure
	var adv models.OneShotAdventure
	err := db.DB.QueryRow("SELECT id, user_id, campaign_id, title, premise, hook, template, estimated_minutes, difficulty, notes, created_at, updated_at FROM oneshot_adventures WHERE id=? AND user_id=?", adventureID, userID).
		Scan(&adv.ID, &adv.UserID, &adv.CampaignID, &adv.Title, &adv.Premise, &adv.Hook, &adv.Template, &adv.EstimatedMinutes, &adv.Difficulty, &adv.Notes, &adv.CreatedAt, &adv.UpdatedAt)
	if err != nil {
		c.String(http.StatusNotFound, "Adventure not found")
		return
	}

	// Load acts with scenes
	actRows, _ := db.DB.Query("SELECT id, adventure_id, number, title, description, estimated_minutes FROM oneshot_acts WHERE adventure_id=? ORDER BY number", adventureID)
	if actRows != nil {
		defer actRows.Close()
		for actRows.Next() {
			var act models.OneShotAct
			if err := actRows.Scan(&act.ID, &act.AdventureID, &act.Number, &act.Title, &act.Description, &act.EstimatedMinutes); err != nil {
				continue
			}
			// Load scenes for this act
			sceneRows, _ := db.DB.Query("SELECT id, act_id, number, title, description, scene_type, location_id, encounter_id, estimated_minutes, notes FROM oneshot_scenes WHERE act_id=? ORDER BY number", act.ID)
			if sceneRows != nil {
				for sceneRows.Next() {
					var sc models.OneShotScene
					if err := sceneRows.Scan(&sc.ID, &sc.ActID, &sc.Number, &sc.Title, &sc.Description, &sc.SceneType, &sc.LocationID, &sc.EncounterID, &sc.EstimatedMinutes, &sc.Notes); err == nil {
						act.Scenes = append(act.Scenes, sc)
					}
				}
				sceneRows.Close()
			}
			adv.Acts = append(adv.Acts, act)
		}
	}

	// Load linked NPCs
	npcRows, _ := db.DB.Query(`
		SELECT oan.id, oan.adventure_id, oan.npc_id, oan.role, COALESCE(n.name,'')
		FROM oneshot_adventure_npcs oan
		LEFT JOIN npcs n ON n.id = oan.npc_id
		WHERE oan.adventure_id=?
	`, adventureID)
	if npcRows != nil {
		defer npcRows.Close()
		for npcRows.Next() {
			var n models.OneShotAdventureNPC
			if err := npcRows.Scan(&n.ID, &n.AdventureID, &n.NPCID, &n.Role, &n.NPCName); err == nil {
				adv.NPCs = append(adv.NPCs, n)
			}
		}
	}

	// Load linked locations
	locRows, _ := db.DB.Query(`
		SELECT oal.id, oal.adventure_id, oal.location_id, COALESCE(l.name,'')
		FROM oneshot_adventure_locations oal
		LEFT JOIN locations l ON l.id = oal.location_id
		WHERE oal.adventure_id=?
	`, adventureID)
	if locRows != nil {
		defer locRows.Close()
		for locRows.Next() {
			var l models.OneShotAdventureLocation
			if err := locRows.Scan(&l.ID, &l.AdventureID, &l.LocationID, &l.LocationName); err == nil {
				adv.Locations = append(adv.Locations, l)
			}
		}
	}

	// Load linked encounters
	encRows, _ := db.DB.Query(`
		SELECT oae.id, oae.adventure_id, oae.encounter_id, COALESCE(e.name,'')
		FROM oneshot_adventure_encounters oae
		LEFT JOIN encounter_templates e ON e.id = oae.encounter_id
		WHERE oae.adventure_id=?
	`, adventureID)
	if encRows != nil {
		defer encRows.Close()
		for encRows.Next() {
			var e models.OneShotAdventureEncounter
			if err := encRows.Scan(&e.ID, &e.AdventureID, &e.EncounterID, &e.EncounterName); err == nil {
				adv.Encounters = append(adv.Encounters, e)
			}
		}
	}

	// Load clues for this adventure
	var clues []models.Clue
	clueRows, _ := db.DB.Query("SELECT id, adventure_id, title, description, clue_type, is_red_herring, is_revealed, sort_order, notes, created_at, updated_at FROM clues WHERE adventure_id=? ORDER BY is_revealed DESC, id", adventureID)
	if clueRows != nil {
		defer clueRows.Close()
		for clueRows.Next() {
			var cl models.Clue
			if err := clueRows.Scan(&cl.ID, &cl.AdventureID, &cl.Title, &cl.Description, &cl.ClueType, &cl.IsRedHerring, &cl.IsRevealed, &cl.SortOrder, &cl.Notes, &cl.CreatedAt, &cl.UpdatedAt); err == nil {
				clues = append(clues, cl)
			}
		}
	}

	// Load pregens for this user (not adventure-specific, but user-specific)
	var pregens []models.PregeneratedCharacter
	pregenRows, _ := db.DB.Query("SELECT id, user_id, name, race, class, subclass, level, background, alignment, str, dex, con, int, wis, cha, hp, ac, speed, skills, equipment, spells, features, personality, backstory, portrait_url, notes, created_at, updated_at FROM pregen_characters WHERE user_id=? ORDER BY updated_at DESC", userID)
	if pregenRows != nil {
		defer pregenRows.Close()
		for pregenRows.Next() {
			var p models.PregeneratedCharacter
			if err := pregenRows.Scan(&p.ID, &p.UserID, &p.Name, &p.Race, &p.Class, &p.Subclass, &p.Level, &p.Background, &p.Alignment,
				&p.Str, &p.Dex, &p.Con, &p.Int, &p.Wis, &p.Cha, &p.HP, &p.AC, &p.Speed,
				&p.Skills, &p.Equipment, &p.Spells, &p.Features, &p.Personality, &p.Backstory,
				&p.PortraitURL, &p.Notes, &p.CreatedAt, &p.UpdatedAt); err == nil {
				pregens = append(pregens, p)
			}
		}
	}

	// Load checklist
	var checklist []models.PrepChecklistItem
	clRows, _ := db.DB.Query("SELECT id, adventure_id, item, category, is_checked, sort_order FROM prep_checklist WHERE adventure_id=? ORDER BY sort_order, id", adventureID)
	if clRows != nil {
		defer clRows.Close()
		for clRows.Next() {
			var item models.PrepChecklistItem
			var checked int
			if err := clRows.Scan(&item.ID, &item.AdventureID, &item.Item, &item.Category, &checked, &item.SortOrder); err == nil {
				item.IsChecked = checked == 1
				checklist = append(checklist, item)
			}
		}
	}

	// Check for pacing session
	var sessionID *int64
	var pacing *models.SessionPacing
	var sid int64
	var pacingStatus string
	err = db.DB.QueryRow("SELECT id, status FROM session_pacing WHERE adventure_id=? AND status IN ('running','paused') ORDER BY id DESC LIMIT 1", adventureID).Scan(&sid, &pacingStatus)
	if err == nil {
		sessionID = &sid
		pacing = &models.SessionPacing{ID: sid, Status: pacingStatus}
		db.DB.QueryRow("SELECT COALESCE(SUM(elapsed_seconds),0) FROM scene_timings WHERE session_id=?", sid).Scan(&pacing.ElapsedSeconds)
	}

	prepData := models.PrepDashboardData{
		Adventure: adv,
		Acts:      adv.Acts,
		Clues:     clues,
		Pregens:   pregens,
		Checklist: checklist,
		Pacing:    pacing,
		SessionID: sessionID,
	}

	renderTemplate(c, "oneshot_prep_dashboard.html", gin.H{
		"D": prepData,
	})
}

// ─── Prep Checklist Handlers ───

func HtmxRenderChecklist(c *gin.Context) {
	adventureID := c.Param("id")
	rows, _ := db.DB.Query("SELECT id, adventure_id, item, category, is_checked, sort_order FROM prep_checklist WHERE adventure_id=? ORDER BY sort_order, id", adventureID)
	if rows != nil {
		defer rows.Close()
		var items []models.PrepChecklistItem
		for rows.Next() {
			var item models.PrepChecklistItem
			var checked int
			if err := rows.Scan(&item.ID, &item.AdventureID, &item.Item, &item.Category, &checked, &item.SortOrder); err == nil {
				item.IsChecked = checked == 1
				items = append(items, item)
			}
		}
		renderTemplate(c, "oneshot_checklist.html", gin.H{"Items": items, "AdventureID": adventureID})
		return
	}
	renderTemplate(c, "oneshot_checklist.html", gin.H{"Items": []models.PrepChecklistItem{}, "AdventureID": adventureID})
}

func HtmxToggleChecklistItem(c *gin.Context) {
	id := c.Param("cid")
	var checked int
	db.DB.QueryRow("SELECT is_checked FROM prep_checklist WHERE id=?", id).Scan(&checked)
	newVal := 0
	if checked == 0 {
		newVal = 1
	}
	db.DB.Exec("UPDATE prep_checklist SET is_checked=? WHERE id=?", newVal, id)

	// Get adventure ID for re-render
	var adventureID int64
	db.DB.QueryRow("SELECT adventure_id FROM prep_checklist WHERE id=?", id).Scan(&adventureID)

	// Re-render full checklist
	rows, _ := db.DB.Query("SELECT id, adventure_id, item, category, is_checked, sort_order FROM prep_checklist WHERE adventure_id=? ORDER BY sort_order, id", adventureID)
	if rows != nil {
		defer rows.Close()
		var items []models.PrepChecklistItem
		for rows.Next() {
			var item models.PrepChecklistItem
			var ck int
			if err := rows.Scan(&item.ID, &item.AdventureID, &item.Item, &item.Category, &ck, &item.SortOrder); err == nil {
				item.IsChecked = ck == 1
				items = append(items, item)
			}
		}
		renderTemplate(c, "oneshot_checklist.html", gin.H{"Items": items, "AdventureID": adventureID})
		return
	}
	c.String(http.StatusOK, "")
}

func HtmxAddChecklistItem(c *gin.Context) {
	adventureID := c.Param("id")
	item := c.PostForm("item")
	category := c.DefaultPostForm("category", "general")

	if item == "" {
		c.String(http.StatusBadRequest, "Item is required")
		return
	}

	// Get max sort order
	var maxOrder int
	db.DB.QueryRow("SELECT COALESCE(MAX(sort_order),0) FROM prep_checklist WHERE adventure_id=?", adventureID).Scan(&maxOrder)

	_, err := db.DB.Exec("INSERT INTO prep_checklist(adventure_id, item, category, sort_order) VALUES(?,?,?,?)", adventureID, item, category, maxOrder+1)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error adding item")
		return
	}

	HtmxRenderChecklist(c)
}

func HtmxDeleteChecklistItem(c *gin.Context) {
	id := c.Param("cid")

	// Get adventure ID before delete
	var adventureID int64
	db.DB.QueryRow("SELECT adventure_id FROM prep_checklist WHERE id=?", id).Scan(&adventureID)

	db.DB.Exec("DELETE FROM prep_checklist WHERE id=?", id)

	// Re-render
	rows, _ := db.DB.Query("SELECT id, adventure_id, item, category, is_checked, sort_order FROM prep_checklist WHERE adventure_id=? ORDER BY sort_order, id", adventureID)
	if rows != nil {
		defer rows.Close()
		var items []models.PrepChecklistItem
		for rows.Next() {
			var item models.PrepChecklistItem
			var ck int
			if err := rows.Scan(&item.ID, &item.AdventureID, &item.Item, &item.Category, &ck, &item.SortOrder); err == nil {
				item.IsChecked = ck == 1
				items = append(items, item)
			}
		}
		renderTemplate(c, "oneshot_checklist.html", gin.H{"Items": items, "AdventureID": adventureID})
		return
	}
	c.String(http.StatusOK, "")
}

// ─── Session Flow Print View ───

func HtmxGetSessionFlow(c *gin.Context) {
	adventureID := c.Param("id")
	userID, _ := c.Get("user_id")

	var adv models.OneShotAdventure
	err := db.DB.QueryRow("SELECT id, user_id, campaign_id, title, premise, hook, template, estimated_minutes, difficulty, notes, created_at, updated_at FROM oneshot_adventures WHERE id=? AND user_id=?", adventureID, userID).
		Scan(&adv.ID, &adv.UserID, &adv.CampaignID, &adv.Title, &adv.Premise, &adv.Hook, &adv.Template, &adv.EstimatedMinutes, &adv.Difficulty, &adv.Notes, &adv.CreatedAt, &adv.UpdatedAt)
	if err != nil {
		c.String(http.StatusNotFound, "Adventure not found")
		return
	}

	// Load acts with scenes
	actRows, _ := db.DB.Query("SELECT id, adventure_id, number, title, description, estimated_minutes FROM oneshot_acts WHERE adventure_id=? ORDER BY number", adventureID)
	if actRows != nil {
		defer actRows.Close()
		for actRows.Next() {
			var act models.OneShotAct
			if err := actRows.Scan(&act.ID, &act.AdventureID, &act.Number, &act.Title, &act.Description, &act.EstimatedMinutes); err != nil {
				continue
			}
			sceneRows, _ := db.DB.Query("SELECT id, act_id, number, title, description, scene_type, location_id, encounter_id, estimated_minutes, notes FROM oneshot_scenes WHERE act_id=? ORDER BY number", act.ID)
			if sceneRows != nil {
				for sceneRows.Next() {
					var sc models.OneShotScene
					if err := sceneRows.Scan(&sc.ID, &sc.ActID, &sc.Number, &sc.Title, &sc.Description, &sc.SceneType, &sc.LocationID, &sc.EncounterID, &sc.EstimatedMinutes, &sc.Notes); err == nil {
						act.Scenes = append(act.Scenes, sc)
					}
				}
				sceneRows.Close()
			}
			adv.Acts = append(adv.Acts, act)
		}
	}

	// Calculate total time
	var totalTime int
	for _, a := range adv.Acts {
		totalTime += a.EstimatedMinutes
	}

	renderTemplate(c, "oneshot_session_flow.html", gin.H{
		"Adventure": adv,
		"TotalTime": totalTime,
	})
}

// ─── DM Screen / Quick Reference Handlers ───

// QuickReferenceData returns the embedded 5E quick reference sections
func QuickReferenceData() []models.DmQuickRefSection {
	return []models.DmQuickRefSection{
		{
			Title: "Conditions",
			Icon:  "fa-exclamation-triangle",
			Entries: []models.DmQuickRefEntry{
				{Name: "Blinded", Description: "Auto-fail Perception (sight). Attack rolls against creature have advantage. Creature's attacks have disadvantage.", Reference: "PHB p.290"},
				{Name: "Charmed", Description: "Can't attack charmer. Charmer has advantage on social checks. Not immunity.", Reference: "PHB p.290"},
				{Name: "Deafened", Description: "Auto-fail Perception (hearing).", Reference: "PHB p.290"},
				{Name: "Exhaustion", Description: "1: Disadvantage on ability checks. 2: Speed halved. 3: Disadvantage on attacks/saves. 4: HP max halved. 5: Speed 0. 6: Death.", Reference: "PHB p.291"},
				{Name: "Frightened", Description: "Disadvantage on checks/attacks while source visible. Can't move closer to source.", Reference: "PHB p.290"},
				{Name: "Grappled", Description: "Speed 0. Ends if grappler incapacitated or target breaks free (Athletics/Acrobatics vs DC).", Reference: "PHB p.290"},
				{Name: "Incapacitated", Description: "Can't take actions, bonus actions, or reactions.", Reference: "PHB p.290"},
				{Name: "Invisible", Description: "Can't be seen without special senses. Attack rolls have advantage. Attacks against have disadvantage.", Reference: "PHB p.291"},
				{Name: "Paralyzed", Description: "Incapacitated. Can't move/speak. Auto-fail STR/DEX saves. Attacks have advantage. Auto-crit within 5ft.", Reference: "PHB p.291"},
				{Name: "Petrified", Description: "Turned to stone + weight x10. Incapacitated. Resist all damage. Immune to poison/disease.", Reference: "PHB p.291"},
				{Name: "Poisoned", Description: "Disadvantage on attack rolls and ability checks.", Reference: "PHB p.292"},
				{Name: "Prone", Description: "Can only crawl. Attacks have disadvantage. Attacks within 5ft have advantage, beyond have disadvantage.", Reference: "PHB p.292"},
				{Name: "Restrained", Description: "Speed 0. Attacks have disadvantage. Attacks against have advantage. DEX saves have disadvantage.", Reference: "PHB p.292"},
				{Name: "Stunned", Description: "Incapacitated. Can't move. Auto-fail STR/DEX saves. Attacks have advantage.", Reference: "PHB p.292"},
				{Name: "Unconscious", Description: "Incapacitated + prone. Auto-fail STR/DEX saves. Attacks have advantage. Auto-crit within 5ft.", Reference: "PHB p.292"},
			},
		},
		{
			Title: "Combat Actions",
			Icon:  "fa-fist-raised",
			Entries: []models.DmQuickRefEntry{
				{Name: "Attack", Description: "Make a melee or ranged weapon attack, or unarmed strike.", Reference: "PHB p.192"},
				{Name: "Cast a Spell", Description: "Cast a spell with casting time of 1 action. Must have components.", Reference: "PHB p.192"},
				{Name: "Dash", Description: "Gain extra movement equal to your speed.", Reference: "PHB p.192"},
				{Name: "Disengage", Description: "Your movement doesn't provoke opportunity attacks.", Reference: "PHB p.192"},
				{Name: "Dodge", Description: "Attacks against have disadvantage. DEX saves have advantage.", Reference: "PHB p.192"},
				{Name: "Help", Description: "Give ally advantage on next ability check or attack.", Reference: "PHB p.192"},
				{Name: "Hide", Description: "Make Stealth check to become unseen.", Reference: "PHB p.192"},
				{Name: "Ready", Description: "Prepare action with trigger. Use reaction to execute.", Reference: "PHB p.193"},
				{Name: "Search", Description: "Make Perception or Investigation check.", Reference: "PHB p.193"},
				{Name: "Use Object", Description: "Interact with an object (second object interaction uses action).", Reference: "PHB p.193"},
			},
		},
		{
			Title: "Difficulty Classes",
			Icon:  "fa-tachometer-alt",
			Entries: []models.DmQuickRefEntry{
				{Name: "Very Easy", Description: "DC 5"},
				{Name: "Easy", Description: "DC 10"},
				{Name: "Medium", Description: "DC 15"},
				{Name: "Hard", Description: "DC 20"},
				{Name: "Very Hard", Description: "DC 25"},
				{Name: "Nearly Impossible", Description: "DC 30"},
			},
		},
		{
			Title: "Rest & Recovery",
			Icon:  "fa-bed",
			Entries: []models.DmQuickRefEntry{
				{Name: "Short Rest", Description: "1 hour. Spend Hit Dice (any number). Regain abilities that recharge on short rest.", Reference: "PHB p.186"},
				{Name: "Long Rest", Description: "8 hours (2h light activity, 6h sleep). Regain all HP + half max Hit Dice. Regain all abilities.", Reference: "PHB p.186"},
				{Name: "Healing", Description: "Restore HP equal to hit die roll + CON mod.", Reference: "PHB p.196"},
			},
		},
		{
			Title: "Travel Pace",
			Icon:  "fa-hiking",
			Entries: []models.DmQuickRefEntry{
				{Name: "Fast", Description: "400ft/min, 4mph, 30mi/day. -5 Passive Perception. Can't stealth.", Reference: "PHB p.182"},
				{Name: "Normal", Description: "300ft/min, 3mph, 24mi/day.", Reference: "PHB p.182"},
				{Name: "Slow", Description: "200ft/min, 2mph, 18mi/day. Can stealth. Can use Perception.", Reference: "PHB p.182"},
			},
		},
		{
			Title: "Light & Vision",
			Icon:  "fa-eye",
			Entries: []models.DmQuickRefEntry{
				{Name: "Bright Light", Description: "Normal vision. Torches (20ft), lanterns (30ft).", Reference: "PHB p.183"},
				{Name: "Dim Light", Description: "Lightly obscured. Disadvantage on Perception (sight)."},
				{Name: "Darkness", Description: "Heavily obscured. Blind. Darkvision ignores up to 60ft (dim) or 120ft (gray).", Reference: "PHB p.183"},
			},
		},
		{
			Title: "Cover",
			Icon:  "fa-shield-alt",
			Entries: []models.DmQuickRefEntry{
				{Name: "Half Cover", Description: "+2 AC and DEX saves. Behind obstacle, low wall, furniture.", Reference: "PHB p.196"},
				{Name: "Three-Quarters Cover", Description: "+5 AC and DEX saves. Behind pillar, battlement.", Reference: "PHB p.196"},
				{Name: "Total Cover", Description: "Can't be targeted directly.", Reference: "PHB p.196"},
			},
		},
		{
			Title: "Skills by Ability",
			Icon:  "fa-clipboard-list",
			Entries: []models.DmQuickRefEntry{
				{Name: "Strength", Description: "Athletics"},
				{Name: "Dexterity", Description: "Acrobatics, Sleight of Hand, Stealth"},
				{Name: "Constitution", Description: "No skills (but CON saves common)"},
				{Name: "Intelligence", Description: "Arcana, History, Investigation, Nature, Religion"},
				{Name: "Wisdom", Description: "Animal Handling, Insight, Medicine, Perception, Survival"},
				{Name: "Charisma", Description: "Deception, Intimidation, Performance, Persuasion"},
			},
		},
	}
}

func HtmxDmScreen(c *gin.Context) {
	adventureID := c.Param("id")
	userID, _ := c.Get("user_id")

	// Load adventure info
	var adv models.OneShotAdventure
	err := db.DB.QueryRow(`
		SELECT id, user_id, campaign_id, title, premise, hook, template, estimated_minutes, difficulty, notes, created_at, updated_at
		FROM oneshot_adventures WHERE id=? AND user_id=?
	`, adventureID, userID).Scan(&adv.ID, &adv.UserID, &adv.CampaignID, &adv.Title, &adv.Premise, &adv.Hook, &adv.Template, &adv.EstimatedMinutes, &adv.Difficulty, &adv.Notes, &adv.CreatedAt, &adv.UpdatedAt)
	if err != nil {
		c.String(http.StatusNotFound, "Adventure not found")
		return
	}

	// Load DM notes
	rows, err := db.DB.Query("SELECT id, adventure_id, user_id, title, content, created_at, updated_at FROM dm_notes WHERE adventure_id=? AND user_id=? ORDER BY updated_at DESC", adventureID, userID)
	var notes []models.DmNote
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var n models.DmNote
			if err := rows.Scan(&n.ID, &n.AdventureID, &n.UserID, &n.Title, &n.Content, &n.CreatedAt, &n.UpdatedAt); err == nil {
				notes = append(notes, n)
			}
		}
	}
	if notes == nil {
		notes = []models.DmNote{}
	}

	renderTemplate(c, "oneshot_dm_screen.html", gin.H{
		"Adventure": adv,
		"Sections":  QuickReferenceData(),
		"Notes":     notes,
	})
}

// ─── DM Notes API ───
