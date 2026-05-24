package handlers

import (
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

// ─── One-Shot Adventure API Handlers ───

func ListOneShotAdventures(c *gin.Context) {
	userID, _ := c.Get("user_id")
	campaignID := c.Query("campaign_id")

	var rows *sql.Rows
	var err error
	if campaignID != "" {
		rows, err = db.DB.Query("SELECT id, user_id, campaign_id, title, premise, hook, template, estimated_minutes, difficulty, notes, created_at, updated_at FROM oneshot_adventures WHERE campaign_id=? ORDER BY updated_at DESC", campaignID)
	} else {
		rows, err = db.DB.Query("SELECT id, user_id, campaign_id, title, premise, hook, template, estimated_minutes, difficulty, notes, created_at, updated_at FROM oneshot_adventures WHERE user_id=? ORDER BY updated_at DESC", userID)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	out := make([]models.OneShotAdventure, 0)
	for rows.Next() {
		var a models.OneShotAdventure
		rows.Scan(&a.ID, &a.UserID, &a.CampaignID, &a.Title, &a.Premise, &a.Hook, &a.Template, &a.EstimatedMinutes, &a.Difficulty, &a.Notes, &a.CreatedAt, &a.UpdatedAt)
		out = append(out, a)
	}
	c.JSON(http.StatusOK, out)
}

func CreateOneShotAdventure(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var a models.OneShotAdventure
	if err := c.ShouldBindJSON(&a); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := db.DB.Exec("INSERT INTO oneshot_adventures(user_id, campaign_id, title, premise, hook, template, estimated_minutes, difficulty, notes) VALUES(?,?,?,?,?,?,?,?,?)",
		userID, a.CampaignID, a.Title, a.Premise, a.Hook, a.Template, a.EstimatedMinutes, a.Difficulty, a.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()

	// Generate acts/scenes from template if applicable
	if a.Template == "five_room_dungeon" {
		generateFiveRoomDungeonStructure(id, a.Difficulty)
	} else if a.Template == "default" {
		generateDefaultStructure(id)
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func GetOneShotAdventure(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := c.Get("user_id")

	var a models.OneShotAdventure
	err := db.DB.QueryRow("SELECT id, user_id, campaign_id, title, premise, hook, template, estimated_minutes, difficulty, notes, created_at, updated_at FROM oneshot_adventures WHERE id=? AND user_id=?", id, userID).
		Scan(&a.ID, &a.UserID, &a.CampaignID, &a.Title, &a.Premise, &a.Hook, &a.Template, &a.EstimatedMinutes, &a.Difficulty, &a.Notes, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "one-shot not found"})
		return
	}

	// Load acts with scenes
	actRows, err := db.DB.Query("SELECT id, adventure_id, number, title, description, estimated_minutes FROM oneshot_acts WHERE adventure_id=? ORDER BY number", id)
	if err == nil {
		defer actRows.Close()
		a.Acts = make([]models.OneShotAct, 0)
		for actRows.Next() {
			var act models.OneShotAct
			actRows.Scan(&act.ID, &act.AdventureID, &act.Number, &act.Title, &act.Description, &act.EstimatedMinutes)
			// Load scenes for this act
			sceneRows, err := db.DB.Query("SELECT s.id, s.act_id, s.number, s.title, s.description, s.scene_type, s.location_id, s.encounter_id, s.estimated_minutes, s.notes, COALESCE(l.name,''), COALESCE(e.name,'') FROM oneshot_scenes s LEFT JOIN locations l ON s.location_id=l.id LEFT JOIN encounter_templates e ON s.encounter_id=e.id WHERE s.act_id=? ORDER BY s.number", act.ID)
			if err == nil {
				act.Scenes = make([]models.OneShotScene, 0)
				for sceneRows.Next() {
					var sc models.OneShotScene
					sceneRows.Scan(&sc.ID, &sc.ActID, &sc.Number, &sc.Title, &sc.Description, &sc.SceneType, &sc.LocationID, &sc.EncounterID, &sc.EstimatedMinutes, &sc.Notes, &sc.LocationName, &sc.EncounterName)
					act.Scenes = append(act.Scenes, sc)
				}
				sceneRows.Close()
			}
			a.Acts = append(a.Acts, act)
		}
	}

	// Load linked NPCs
	npcRows, err := db.DB.Query("SELECT oan.id, oan.adventure_id, oan.npc_id, oan.role, COALESCE(n.name,'') FROM oneshot_adventure_npcs oan LEFT JOIN npcs n ON oan.npc_id=n.id WHERE oan.adventure_id=?", id)
	if err == nil {
		defer npcRows.Close()
		a.NPCs = make([]models.OneShotAdventureNPC, 0)
		for npcRows.Next() {
			var npc models.OneShotAdventureNPC
			npcRows.Scan(&npc.ID, &npc.AdventureID, &npc.NPCID, &npc.Role, &npc.NPCName)
			a.NPCs = append(a.NPCs, npc)
		}
	}

	// Load linked Locations
	locRows, err := db.DB.Query("SELECT oal.id, oal.adventure_id, oal.location_id, COALESCE(l.name,'') FROM oneshot_adventure_locations oal LEFT JOIN locations l ON oal.location_id=l.id WHERE oal.adventure_id=?", id)
	if err == nil {
		defer locRows.Close()
		a.Locations = make([]models.OneShotAdventureLocation, 0)
		for locRows.Next() {
			var loc models.OneShotAdventureLocation
			locRows.Scan(&loc.ID, &loc.AdventureID, &loc.LocationID, &loc.LocationName)
			a.Locations = append(a.Locations, loc)
		}
	}

	// Load linked Encounters
	encRows, err := db.DB.Query("SELECT oae.id, oae.adventure_id, oae.encounter_id, COALESCE(e.name,'') FROM oneshot_adventure_encounters oae LEFT JOIN encounter_templates e ON oae.encounter_id=e.id WHERE oae.adventure_id=?", id)
	if err == nil {
		defer encRows.Close()
		a.Encounters = make([]models.OneShotAdventureEncounter, 0)
		for encRows.Next() {
			var enc models.OneShotAdventureEncounter
			encRows.Scan(&enc.ID, &enc.AdventureID, &enc.EncounterID, &enc.EncounterName)
			a.Encounters = append(a.Encounters, enc)
		}
	}

	c.JSON(http.StatusOK, a)
}

func UpdateOneShotAdventure(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var a models.OneShotAdventure
	if err := c.ShouldBindJSON(&a); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Exec("UPDATE oneshot_adventures SET title=?, premise=?, hook=?, template=?, estimated_minutes=?, difficulty=?, notes=?, updated_at=datetime('now') WHERE id=?",
		a.Title, a.Premise, a.Hook, a.Template, a.EstimatedMinutes, a.Difficulty, a.Notes, id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteOneShotAdventure(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := c.Get("user_id")
	db.DB.Exec("DELETE FROM oneshot_adventures WHERE id=? AND user_id=?", id, userID)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ─── Acts ───

func CreateOneShotAct(c *gin.Context) {
	adventureID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var act models.OneShotAct
	if err := c.ShouldBindJSON(&act); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Auto-assign number
	if act.Number == 0 {
		db.DB.QueryRow("SELECT COALESCE(MAX(number),0)+1 FROM oneshot_acts WHERE adventure_id=?", adventureID).Scan(&act.Number)
	}
	result, err := db.DB.Exec("INSERT INTO oneshot_acts(adventure_id, number, title, description, estimated_minutes) VALUES(?,?,?,?,?)",
		adventureID, act.Number, act.Title, act.Description, act.EstimatedMinutes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func UpdateOneShotAct(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var act models.OneShotAct
	if err := c.ShouldBindJSON(&act); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Exec("UPDATE oneshot_acts SET number=?, title=?, description=?, estimated_minutes=? WHERE id=?",
		act.Number, act.Title, act.Description, act.EstimatedMinutes, id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteOneShotAct(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	db.DB.Exec("DELETE FROM oneshot_acts WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ─── Scenes ───

func CreateOneShotScene(c *gin.Context) {
	actID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var sc models.OneShotScene
	if err := c.ShouldBindJSON(&sc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if sc.Number == 0 {
		db.DB.QueryRow("SELECT COALESCE(MAX(number),0)+1 FROM oneshot_scenes WHERE act_id=?", actID).Scan(&sc.Number)
	}
	result, err := db.DB.Exec("INSERT INTO oneshot_scenes(act_id, number, title, description, scene_type, location_id, encounter_id, estimated_minutes, notes) VALUES(?,?,?,?,?,?,?,?,?)",
		actID, sc.Number, sc.Title, sc.Description, sc.SceneType, sc.LocationID, sc.EncounterID, sc.EstimatedMinutes, sc.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func UpdateOneShotScene(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var sc models.OneShotScene
	if err := c.ShouldBindJSON(&sc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Exec("UPDATE oneshot_scenes SET number=?, title=?, description=?, scene_type=?, location_id=?, encounter_id=?, estimated_minutes=?, notes=? WHERE id=?",
		sc.Number, sc.Title, sc.Description, sc.SceneType, sc.LocationID, sc.EncounterID, sc.EstimatedMinutes, sc.Notes, id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteOneShotScene(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	db.DB.Exec("DELETE FROM oneshot_scenes WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ─── NPC Links ───

func GetOneShotNPCs(c *gin.Context) {
	adventureID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	rows, err := db.DB.Query("SELECT oan.id, oan.adventure_id, oan.npc_id, oan.role, COALESCE(n.name,'') FROM oneshot_adventure_npcs oan LEFT JOIN npcs n ON oan.npc_id=n.id WHERE oan.adventure_id=?", adventureID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := make([]models.OneShotAdventureNPC, 0)
	for rows.Next() {
		var npc models.OneShotAdventureNPC
		rows.Scan(&npc.ID, &npc.AdventureID, &npc.NPCID, &npc.Role, &npc.NPCName)
		out = append(out, npc)
	}
	c.JSON(http.StatusOK, out)
}

func LinkOneShotNPC(c *gin.Context) {
	adventureID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var link struct {
		NPCID int64  `json:"npc_id"`
		Role  string `json:"role"`
	}
	if err := c.ShouldBindJSON(&link); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := db.DB.Exec("INSERT OR IGNORE INTO oneshot_adventure_npcs(adventure_id, npc_id, role) VALUES(?,?,?)", adventureID, link.NPCID, link.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func UnlinkOneShotNPC(c *gin.Context) {
	adventureID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	npcID, _ := strconv.ParseInt(c.Param("nid"), 10, 64)
	db.DB.Exec("DELETE FROM oneshot_adventure_npcs WHERE adventure_id=? AND npc_id=?", adventureID, npcID)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ─── Location Links ───

func GetOneShotLocations(c *gin.Context) {
	adventureID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	rows, err := db.DB.Query("SELECT oal.id, oal.adventure_id, oal.location_id, COALESCE(l.name,'') FROM oneshot_adventure_locations oal LEFT JOIN locations l ON oal.location_id=l.id WHERE oal.adventure_id=?", adventureID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := make([]models.OneShotAdventureLocation, 0)
	for rows.Next() {
		var loc models.OneShotAdventureLocation
		rows.Scan(&loc.ID, &loc.AdventureID, &loc.LocationID, &loc.LocationName)
		out = append(out, loc)
	}
	c.JSON(http.StatusOK, out)
}

func LinkOneShotLocation(c *gin.Context) {
	adventureID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var link struct {
		LocationID int64 `json:"location_id"`
	}
	if err := c.ShouldBindJSON(&link); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := db.DB.Exec("INSERT OR IGNORE INTO oneshot_adventure_locations(adventure_id, location_id) VALUES(?,?)", adventureID, link.LocationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func UnlinkOneShotLocation(c *gin.Context) {
	adventureID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	locationID, _ := strconv.ParseInt(c.Param("lid"), 10, 64)
	db.DB.Exec("DELETE FROM oneshot_adventure_locations WHERE adventure_id=? AND location_id=?", adventureID, locationID)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ─── Encounter Links ───

func GetOneShotEncounters(c *gin.Context) {
	adventureID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	rows, err := db.DB.Query("SELECT oae.id, oae.adventure_id, oae.encounter_id, COALESCE(e.name,'') FROM oneshot_adventure_encounters oae LEFT JOIN encounter_templates e ON oae.encounter_id=e.id WHERE oae.adventure_id=?", adventureID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := make([]models.OneShotAdventureEncounter, 0)
	for rows.Next() {
		var enc models.OneShotAdventureEncounter
		rows.Scan(&enc.ID, &enc.AdventureID, &enc.EncounterID, &enc.EncounterName)
		out = append(out, enc)
	}
	c.JSON(http.StatusOK, out)
}

func LinkOneShotEncounter(c *gin.Context) {
	adventureID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var link struct {
		EncounterID int64 `json:"encounter_id"`
	}
	if err := c.ShouldBindJSON(&link); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := db.DB.Exec("INSERT OR IGNORE INTO oneshot_adventure_encounters(adventure_id, encounter_id) VALUES(?,?)", adventureID, link.EncounterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func UnlinkOneShotEncounter(c *gin.Context) {
	adventureID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	encounterID, _ := strconv.ParseInt(c.Param("eid"), 10, 64)
	db.DB.Exec("DELETE FROM oneshot_adventure_encounters WHERE adventure_id=? AND encounter_id=?", adventureID, encounterID)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ─── Template Generation ───

type GenerateRequest struct {
	Template     string `json:"template"`
	Title        string `json:"title"`
	Difficulty   string `json:"difficulty"`
	Minutes      int    `json:"estimated_minutes"`
	CampaignID   *int64 `json:"campaign_id,omitempty"`
}

func GenerateOneShotFromTemplate(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Minutes <= 0 {
		req.Minutes = 180
	}
	if req.Difficulty == "" {
		req.Difficulty = "medium"
	}
	if req.Title == "" {
		req.Title = "Untitled One-Shot"
	}

	var premise, hook string

	switch req.Template {
	case "five_room_dungeon":
		premise, hook = generateFiveRoomDungeon(req.Difficulty)
	default:
		premise = "A new adventure begins..."
		hook = "The party is called to action."
	}

	// Create the adventure
	now := time.Now().Format("2006-01-02 15:04:05")
	result, err := db.DB.Exec("INSERT INTO oneshot_adventures(user_id, campaign_id, title, premise, hook, template, estimated_minutes, difficulty, notes, created_at, updated_at) VALUES(?,?,?,?,?,?,?,?,'',?,?)",
		userID, req.CampaignID, req.Title, premise, hook, req.Template, req.Minutes, req.Difficulty, now, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	adventureID, _ := result.LastInsertId()

	// Generate acts and scenes based on template
	switch req.Template {
	case "five_room_dungeon":
		generateFiveRoomDungeonStructure(adventureID, req.Difficulty)
	default:
		generateDefaultStructure(adventureID)
	}

	c.JSON(http.StatusCreated, gin.H{"id": adventureID})
}

func generateFiveRoomDungeon(difficulty string) (string, string) {
	hooks := map[string][]string{
		"easy": {
			"Strange mushrooms have been sprouting in the cellar of the local inn.",
			"Kobolds have been stealing livestock from nearby farms.",
			"A mysterious fog rolls in every night, carrying strange whispers.",
		},
		"medium": {
			"Cultists have taken over an ancient temple in the forest.",
			"A powerful artifact was stolen from the duke's vault.",
			"Goblins are amassing an army in the abandoned mines.",
		},
		"hard": {
			"A dragon's lair has been discovered beneath the mountain.",
			"An evil wizard is conducting dark rituals in a fallen keep.",
			"Undead legions are rising from an ancient battlefield.",
		},
		"deadly": {
			"A lich is gathering the last surviving members of an ancient order.",
			"The boundaries between planes are weakening in a cursed valley.",
			"A demonic incursion threatens to consume the realm.",
		},
	}

	premises := map[string][]string{
		"easy": {
			"A local crisis requires brave adventurers to investigate strange occurrences.",
			"The nearby village needs help with a growing monster problem.",
		},
		"medium": {
			"A villain's plot threatens the region, and only a daring group can stop it.",
			"An ancient evil stirs, and the party must uncover its source.",
		},
		"hard": {
			"A catastrophic threat looms over the land, requiring heroes of great courage.",
			"The realm faces its greatest danger in centuries.",
		},
		"deadly": {
			"The end of all things approaches. The party must make a stand against impossible odds.",
			"A world-ending threat emerges, and the fate of civilization hangs in the balance.",
		},
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	hooksList := hooks[difficulty]
	premisesList := premises[difficulty]

	return premisesList[r.Intn(len(premisesList))], hooksList[r.Intn(len(hooksList))]
}

func generateFiveRoomDungeonStructure(adventureID int64, difficulty string) {
	rooms := []struct {
		title       string
		description string
		sceneType   string
		minutes     int
	}{
		{
			title:       "Entrance & Guardian",
			description: "The party approaches the entrance. A guardian or obstacle blocks the way, testing their readiness.",
			sceneType:   "exploration",
			minutes:     20,
		},
		{
			title:       "Puzzle or Roleplaying Challenge",
			description: "A puzzle, trap, or social encounter that requires wit rather than combat to overcome.",
			sceneType:   "puzzle",
			minutes:     25,
		},
		{
			title:       "Trick or Setback",
			description: "A trick, trap, or setback that drains resources or misleads the party.",
			sceneType:   "exploration",
			minutes:     15,
		},
		{
			title:       "Climax - Big Battle",
			description: "The main confrontation. The party faces the biggest threat in the dungeon.",
			sceneType:   "combat",
			minutes:     30,
		},
		{
			title:       "Reward & Revelation",
			description: "The aftermath. Treasure, clues about the larger story, and a chance to catch their breath.",
			sceneType:   "roleplay",
			minutes:     15,
		},
	}

	diffMult := map[string]int{"easy": 1, "medium": 2, "hard": 3, "deadly": 4}

	for i, room := range rooms {
		// Create act for each room
		actResult, err := db.DB.Exec("INSERT INTO oneshot_acts(adventure_id, number, title, description, estimated_minutes) VALUES(?,?,?,?,?)",
			adventureID, i+1, room.title, room.description, room.minutes*diffMult[difficulty])
		if err != nil {
			continue
		}
		actID, _ := actResult.LastInsertId()

		// Create scene within act
		db.DB.Exec("INSERT INTO oneshot_scenes(act_id, number, title, description, scene_type, estimated_minutes, notes) VALUES(?,1,?,?,?,?,'')",
			actID, room.title, room.description, room.sceneType, room.minutes*diffMult[difficulty])
	}
}

func generateDefaultStructure(adventureID int64) {
	acts := []struct {
		title       string
		description string
		minutes     int
		scenes      []struct {
			title       string
			description string
			sceneType   string
			minutes     int
		}
	}{
		{
			title:       "Setup",
			description: "The party receives the call to adventure and gathers information.",
			minutes:     45,
			scenes: []struct {
				title       string
				description string
				sceneType   string
				minutes     int
			}{
				{"The Hook", "The party learns about the situation and decides to act.", "roleplay", 20},
				{"Preparation", "The party gathers supplies, information, and allies.", "exploration", 25},
			},
		},
		{
			title:       "Complication",
			description: "Obstacles arise and the party faces challenges.",
			minutes:     60,
			scenes: []struct {
				title       string
				description string
				sceneType   string
				minutes     int
			}{
				{"First Obstacle", "A challenge blocks the party's path.", "exploration", 20},
				{"Twist", "A unexpected development changes the situation.", "roleplay", 15},
				{"Rising Tension", "The stakes increase as the party pushes forward.", "combat", 25},
			},
		},
		{
			title:       "Climax",
			description: "The final confrontation and resolution.",
			minutes:     45,
			scenes: []struct {
				title       string
				description string
				sceneType   string
				minutes     int
			}{
				{"Final Confrontation", "The party faces the main threat.", "climax", 30},
				{"Resolution", "The aftermath and celebration.", "roleplay", 15},
			},
		},
	}

	for i, act := range acts {
		actResult, err := db.DB.Exec("INSERT INTO oneshot_acts(adventure_id, number, title, description, estimated_minutes) VALUES(?,?,?,?,?)",
			adventureID, i+1, act.title, act.description, act.minutes)
		if err != nil {
			continue
		}
		actID, _ := actResult.LastInsertId()
		for j, sc := range act.scenes {
			db.DB.Exec("INSERT INTO oneshot_scenes(act_id, number, title, description, scene_type, estimated_minutes, notes) VALUES(?,?,?,?,?,?,'')",
				actID, j+1, sc.title, sc.description, sc.sceneType, sc.minutes)
		}
	}
}

// ─── HTMX Handlers ───

type htmxOneShotData struct {
	Adventure  *models.OneShotAdventure
	Adventures []models.OneShotAdventure
	NPCs       []models.NPC
	Locations  []models.Location
	Encounters []models.EncounterTemplate
	Act        *models.OneShotAct
	Scene      *models.OneShotScene
	SceneTypes []string
	Templates  []string
	Difficulties []string
}

func HtmxListOneShots(c *gin.Context) {
	userID, _ := c.Get("user_id")
	campaignID := c.Query("campaign_id")

	var rows *sql.Rows
	var err error
	if campaignID != "" {
		rows, err = db.DB.Query("SELECT id, user_id, campaign_id, title, premise, hook, template, estimated_minutes, difficulty, notes, created_at, updated_at FROM oneshot_adventures WHERE campaign_id=? ORDER BY updated_at DESC", campaignID)
	} else {
		rows, err = db.DB.Query("SELECT id, user_id, campaign_id, title, premise, hook, template, estimated_minutes, difficulty, notes, created_at, updated_at FROM oneshot_adventures WHERE user_id=? ORDER BY updated_at DESC", userID)
	}
	if err != nil {
		c.String(http.StatusInternalServerError, "query error: %v", err)
		return
	}
	defer rows.Close()

	out := make([]models.OneShotAdventure, 0)
	for rows.Next() {
		var a models.OneShotAdventure
		rows.Scan(&a.ID, &a.UserID, &a.CampaignID, &a.Title, &a.Premise, &a.Hook, &a.Template, &a.EstimatedMinutes, &a.Difficulty, &a.Notes, &a.CreatedAt, &a.UpdatedAt)
		out = append(out, a)
	}
	renderTemplate(c, "oneshot_list.html", htmxOneShotData{Adventures: out})
}

func HtmxNewOneShotForm(c *gin.Context) {
	campaignID := c.Query("campaign_id")
	data := htmxOneShotData{
		SceneTypes:   []string{"roleplay", "combat", "exploration", "puzzle", "climax"},
		Templates:    []string{"custom", "five_room_dungeon"},
		Difficulties: []string{"easy", "medium", "hard", "deadly"},
	}
	if campaignID != "" {
		if cid, err := strconv.ParseInt(campaignID, 10, 64); err == nil {
			data.Adventure = &models.OneShotAdventure{CampaignID: &cid}
		}
	}
	renderTemplate(c, "oneshot_form.html", data)
}

func HtmxEditOneShotForm(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := c.Get("user_id")

	var a models.OneShotAdventure
	err := db.DB.QueryRow("SELECT id, user_id, campaign_id, title, premise, hook, template, estimated_minutes, difficulty, notes, created_at, updated_at FROM oneshot_adventures WHERE id=? AND user_id=?", id, userID).
		Scan(&a.ID, &a.UserID, &a.CampaignID, &a.Title, &a.Premise, &a.Hook, &a.Template, &a.EstimatedMinutes, &a.Difficulty, &a.Notes, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		c.String(http.StatusNotFound, "not found")
		return
	}
	renderTemplate(c, "oneshot_form.html", htmxOneShotData{
		Adventure:    &a,
		SceneTypes:   []string{"roleplay", "combat", "exploration", "puzzle", "climax"},
		Templates:    []string{"custom", "five_room_dungeon"},
		Difficulties: []string{"easy", "medium", "hard", "deadly"},
	})
}

func HtmxGetOneShotDetail(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	// Use GetOneShotAdventure logic via direct queries since we need the data
	var a models.OneShotAdventure
	userID, _ := c.Get("user_id")
	err := db.DB.QueryRow("SELECT id, user_id, campaign_id, title, premise, hook, template, estimated_minutes, difficulty, notes, created_at, updated_at FROM oneshot_adventures WHERE id=? AND user_id=?", id, userID).
		Scan(&a.ID, &a.UserID, &a.CampaignID, &a.Title, &a.Premise, &a.Hook, &a.Template, &a.EstimatedMinutes, &a.Difficulty, &a.Notes, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		c.String(http.StatusNotFound, "not found")
		return
	}

	// Load acts with scenes
	actRows, _ := db.DB.Query("SELECT id, adventure_id, number, title, description, estimated_minutes FROM oneshot_acts WHERE adventure_id=? ORDER BY number", id)
	if actRows != nil {
		defer actRows.Close()
		for actRows.Next() {
			var act models.OneShotAct
			actRows.Scan(&act.ID, &act.AdventureID, &act.Number, &act.Title, &act.Description, &act.EstimatedMinutes)
			sceneRows, _ := db.DB.Query("SELECT s.id, s.act_id, s.number, s.title, s.description, s.scene_type, s.location_id, s.encounter_id, s.estimated_minutes, s.notes, COALESCE(l.name,''), COALESCE(e.name,'') FROM oneshot_scenes s LEFT JOIN locations l ON s.location_id=l.id LEFT JOIN encounter_templates e ON s.encounter_id=e.id WHERE s.act_id=? ORDER BY s.number", act.ID)
			if sceneRows != nil {
				for sceneRows.Next() {
					var sc models.OneShotScene
					sceneRows.Scan(&sc.ID, &sc.ActID, &sc.Number, &sc.Title, &sc.Description, &sc.SceneType, &sc.LocationID, &sc.EncounterID, &sc.EstimatedMinutes, &sc.Notes, &sc.LocationName, &sc.EncounterName)
					act.Scenes = append(act.Scenes, sc)
				}
				sceneRows.Close()
			}
			a.Acts = append(a.Acts, act)
		}
	}

	renderTemplate(c, "oneshot_detail.html", htmxOneShotData{Adventure: &a})
}

func HtmxCreateOneShot(c *gin.Context) {
	userID, _ := c.Get("user_id")
	title := c.PostForm("title")
	premise := c.PostForm("premise")
	hook := c.PostForm("hook")
	template := c.PostForm("template")
	difficulty := c.PostForm("difficulty")
	minutes, _ := strconv.Atoi(c.PostForm("estimated_minutes"))
	notes := c.PostForm("notes")
	campaignStr := c.PostForm("campaign_id")

	if title == "" {
		title = "Untitled One-Shot"
	}
	if template == "" {
		template = "custom"
	}
	if difficulty == "" {
		difficulty = "medium"
	}
	if minutes <= 0 {
		minutes = 180
	}

	var campaignID *int64
	if campaignStr != "" {
		if cid, err := strconv.ParseInt(campaignStr, 10, 64); err == nil {
			campaignID = &cid
		}
	}

	result, err := db.DB.Exec("INSERT INTO oneshot_adventures(user_id, campaign_id, title, premise, hook, template, estimated_minutes, difficulty, notes) VALUES(?,?,?,?,?,?,?,?,?)",
		userID, campaignID, title, premise, hook, template, minutes, difficulty, notes)
	if err != nil {
		c.String(http.StatusInternalServerError, "insert error: %v", err)
		return
	}
	id, _ := result.LastInsertId()

	// If using a template, generate structure
	if template == "five_room_dungeon" {
		generateFiveRoomDungeonStructure(id, difficulty)
	} else if template == "custom" && premise != "" {
		generateDefaultStructure(id)
	}

	// Re-render the list
	HtmxListOneShots(c)
}

func HtmxUpdateOneShot(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	title := c.PostForm("title")
	premise := c.PostForm("premise")
	hook := c.PostForm("hook")
	template := c.PostForm("template")
	difficulty := c.PostForm("difficulty")
	minutes, _ := strconv.Atoi(c.PostForm("estimated_minutes"))
	notes := c.PostForm("notes")

	db.DB.Exec("UPDATE oneshot_adventures SET title=?, premise=?, hook=?, template=?, estimated_minutes=?, difficulty=?, notes=?, updated_at=datetime('now') WHERE id=?",
		title, premise, hook, template, minutes, difficulty, notes, id)

	HtmxListOneShots(c)
}

func HtmxDeleteOneShot(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := c.Get("user_id")
	db.DB.Exec("DELETE FROM oneshot_adventures WHERE id=? AND user_id=?", id, userID)
	HtmxListOneShots(c)
}

// HTMX Act handlers
func HtmxCreateAct(c *gin.Context) {
	adventureID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	title := c.PostForm("title")
	description := c.PostForm("description")
	minutes, _ := strconv.Atoi(c.PostForm("estimated_minutes"))

	if title == "" {
		title = "New Act"
	}
	if minutes <= 0 {
		minutes = 30
	}

	var number int
	db.DB.QueryRow("SELECT COALESCE(MAX(number),0)+1 FROM oneshot_acts WHERE adventure_id=?", adventureID).Scan(&number)
	db.DB.Exec("INSERT INTO oneshot_acts(adventure_id, number, title, description, estimated_minutes) VALUES(?,?,?,?,?)",
		adventureID, number, title, description, minutes)

	HtmxGetOneShotDetail(c)
}

func HtmxUpdateAct(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	title := c.PostForm("title")
	description := c.PostForm("description")
	minutes, _ := strconv.Atoi(c.PostForm("estimated_minutes"))
	number, _ := strconv.Atoi(c.PostForm("number"))

	if number <= 0 {
		db.DB.QueryRow("SELECT number FROM oneshot_acts WHERE id=?", id).Scan(&number)
	}
	db.DB.Exec("UPDATE oneshot_acts SET number=?, title=?, description=?, estimated_minutes=? WHERE id=?",
		number, title, description, minutes, id)

	// Get adventure_id for redirect
	var adventureID int64
	db.DB.QueryRow("SELECT adventure_id FROM oneshot_acts WHERE id=?", id).Scan(&adventureID)
	c.Redirect(http.StatusFound, fmt.Sprintf("/htmx/oneshot-adventures/%d", adventureID))
}

func HtmxDeleteAct(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var adventureID int64
	db.DB.QueryRow("SELECT adventure_id FROM oneshot_acts WHERE id=?", id).Scan(&adventureID)
	db.DB.Exec("DELETE FROM oneshot_acts WHERE id=?", id)

	c.Redirect(http.StatusFound, fmt.Sprintf("/htmx/oneshot-adventures/%d", adventureID))
}

// HTMX Scene handlers
func HtmxCreateScene(c *gin.Context) {
	actID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	title := c.PostForm("title")
	description := c.PostForm("description")
	sceneType := c.PostForm("scene_type")
	minutes, _ := strconv.Atoi(c.PostForm("estimated_minutes"))
	notes := c.PostForm("notes")

	if title == "" {
		title = "New Scene"
	}
	if sceneType == "" {
		sceneType = "roleplay"
	}
	if minutes <= 0 {
		minutes = 15
	}

	var number int
	db.DB.QueryRow("SELECT COALESCE(MAX(number),0)+1 FROM oneshot_scenes WHERE act_id=?", actID).Scan(&number)
	db.DB.Exec("INSERT INTO oneshot_scenes(act_id, number, title, description, scene_type, estimated_minutes, notes) VALUES(?,?,?,?,?,?,?)",
		actID, number, title, description, sceneType, minutes, notes)

	// Get adventure_id for redirect
	var adventureID int64
	db.DB.QueryRow("SELECT oa.adventure_id FROM oneshot_acts oa WHERE oa.id=?", actID).Scan(&adventureID)
	c.Redirect(http.StatusFound, fmt.Sprintf("/htmx/oneshot-adventures/%d", adventureID))
}

func HtmxUpdateScene(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	title := c.PostForm("title")
	description := c.PostForm("description")
	sceneType := c.PostForm("scene_type")
	minutes, _ := strconv.Atoi(c.PostForm("estimated_minutes"))
	notes := c.PostForm("notes")
	number, _ := strconv.Atoi(c.PostForm("number"))

	if number <= 0 {
		db.DB.QueryRow("SELECT number FROM oneshot_scenes WHERE id=?", id).Scan(&number)
	}
	db.DB.Exec("UPDATE oneshot_scenes SET number=?, title=?, description=?, scene_type=?, estimated_minutes=?, notes=? WHERE id=?",
		number, title, description, sceneType, minutes, notes, id)

	// Get adventure_id for redirect
	var adventureID int64
	db.DB.QueryRow("SELECT oa.adventure_id FROM oneshot_acts oa JOIN oneshot_scenes s ON s.act_id=oa.id WHERE s.id=?", id).Scan(&adventureID)
	c.Redirect(http.StatusFound, fmt.Sprintf("/htmx/oneshot-adventures/%d", adventureID))
}

func HtmxDeleteScene(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var adventureID int64
	db.DB.QueryRow("SELECT oa.adventure_id FROM oneshot_acts oa JOIN oneshot_scenes s ON s.act_id=oa.id WHERE s.id=?", id).Scan(&adventureID)
	db.DB.Exec("DELETE FROM oneshot_scenes WHERE id=?", id)

	ReRenderOneShotDetail(c, adventureID)
}

func ReRenderOneShotDetail(c *gin.Context, adventureID int64) {
	var a models.OneShotAdventure
	userID, _ := c.Get("user_id")
	db.DB.QueryRow("SELECT id, user_id, campaign_id, title, premise, hook, template, estimated_minutes, difficulty, notes, created_at, updated_at FROM oneshot_adventures WHERE id=? AND user_id=?", adventureID, userID).
		Scan(&a.ID, &a.UserID, &a.CampaignID, &a.Title, &a.Premise, &a.Hook, &a.Template, &a.EstimatedMinutes, &a.Difficulty, &a.Notes, &a.CreatedAt, &a.UpdatedAt)

	actRows, _ := db.DB.Query("SELECT id, adventure_id, number, title, description, estimated_minutes FROM oneshot_acts WHERE adventure_id=? ORDER BY number", adventureID)
	if actRows != nil {
		defer actRows.Close()
		for actRows.Next() {
			var act models.OneShotAct
			actRows.Scan(&act.ID, &act.AdventureID, &act.Number, &act.Title, &act.Description, &act.EstimatedMinutes)
			sceneRows, _ := db.DB.Query("SELECT s.id, s.act_id, s.number, s.title, s.description, s.scene_type, s.location_id, s.encounter_id, s.estimated_minutes, s.notes, COALESCE(l.name,''), COALESCE(e.name,'') FROM oneshot_scenes s LEFT JOIN locations l ON s.location_id=l.id LEFT JOIN encounter_templates e ON s.encounter_id=e.id WHERE s.act_id=? ORDER BY s.number", act.ID)
			if sceneRows != nil {
				for sceneRows.Next() {
					var sc models.OneShotScene
					sceneRows.Scan(&sc.ID, &sc.ActID, &sc.Number, &sc.Title, &sc.Description, &sc.SceneType, &sc.LocationID, &sc.EncounterID, &sc.EstimatedMinutes, &sc.Notes, &sc.LocationName, &sc.EncounterName)
					act.Scenes = append(act.Scenes, sc)
				}
				sceneRows.Close()
			}
			a.Acts = append(a.Acts, act)
		}
	}

	renderTemplate(c, "oneshot_detail.html", htmxOneShotData{Adventure: &a})
}

// HTMX generate from template
func HtmxGenerateOneShot(c *gin.Context) {
	userID, _ := c.Get("user_id")
	title := c.PostForm("title")
	template := c.PostForm("template")
	difficulty := c.PostForm("difficulty")
	minutes, _ := strconv.Atoi(c.PostForm("estimated_minutes"))
	campaignStr := c.PostForm("campaign_id")

	if title == "" {
		title = "Untitled One-Shot"
	}
	if template == "" {
		template = "five_room_dungeon"
	}
	if difficulty == "" {
		difficulty = "medium"
	}
	if minutes <= 0 {
		minutes = 180
	}

	var campaignID *int64
	if campaignStr != "" {
		if cid, err := strconv.ParseInt(campaignStr, 10, 64); err == nil {
			campaignID = &cid
		}
	}

	var premise, hook string
	switch template {
	case "five_room_dungeon":
		premise, hook = generateFiveRoomDungeon(difficulty)
	default:
		premise = "A new adventure begins..."
		hook = "The party is called to action."
	}

	result, err := db.DB.Exec("INSERT INTO oneshot_adventures(user_id, campaign_id, title, premise, hook, template, estimated_minutes, difficulty, notes) VALUES(?,?,?,?,?,?,?,?,'')",
		userID, campaignID, title, premise, hook, template, minutes, difficulty)
	if err != nil {
		c.String(http.StatusInternalServerError, "insert error: %v", err)
		return
	}
	id, _ := result.LastInsertId()

	if template == "five_room_dungeon" {
		generateFiveRoomDungeonStructure(id, difficulty)
	} else {
		generateDefaultStructure(id)
	}

	HtmxListOneShots(c)
}

// ─── Session Pacing API Handlers ───

func StartPacingSession(c *gin.Context) {
	adventureID := c.Param("id")
	userID, _ := c.Get("user_id")

	// Verify the adventure belongs to this user
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM oneshot_adventures WHERE id=? AND user_id=?)", adventureID, userID).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "adventure not found"})
		return
	}

	// Check if there's already an active session
	var existingID int64
	err = db.DB.QueryRow("SELECT id FROM session_pacing WHERE adventure_id=? AND status IN ('running','paused')", adventureID).Scan(&existingID)
	if err == nil {
		c.JSON(http.StatusOK, gin.H{"id": existingID, "message": "resumed existing session"})
		return
	}

	// Get first act and scene
	var firstActID, firstSceneID *int64
	var actID int64
	err = db.DB.QueryRow("SELECT id FROM oneshot_acts WHERE adventure_id=? ORDER BY number ASC LIMIT 1", adventureID).Scan(&actID)
	if err == nil {
		firstActID = &actID
		var sceneID int64
		err = db.DB.QueryRow("SELECT id FROM oneshot_scenes WHERE act_id=? ORDER BY number ASC LIMIT 1", actID).Scan(&sceneID)
		if err == nil {
			firstSceneID = &sceneID
		}
	}

	result, err := db.DB.Exec(
		"INSERT INTO session_pacing(adventure_id, current_act_id, current_scene_id, status, elapsed_seconds) VALUES(?,?,?,'running',0)",
		adventureID, firstActID, firstSceneID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()

	// If first scene exists, create initial scene timing
	if firstSceneID != nil {
		db.DB.Exec("INSERT INTO scene_timings(session_id, scene_id, status) VALUES(?,?,'active')", id, *firstSceneID)
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func GetPacingSession(c *gin.Context) {
	sessionID := c.Param("id")

	var s models.SessionPacing
	err := db.DB.QueryRow(`
		SELECT sp.id, sp.adventure_id, sp.current_act_id, sp.current_scene_id, sp.status, sp.elapsed_seconds, sp.started_at, COALESCE(sp.completed_at,''),
			COALESCE(oa.title,''), COALESCE(a.title,''), COALESCE(sc.title,''), COALESCE(sc.estimated_minutes,0),
			COALESCE(a.number,0), COALESCE(sc.number,0)
		FROM session_pacing sp
		LEFT JOIN oneshot_adventures oa ON oa.id = sp.adventure_id
		LEFT JOIN oneshot_acts a ON a.id = sp.current_act_id
		LEFT JOIN oneshot_scenes sc ON sc.id = sp.current_scene_id
		WHERE sp.id=?
	`, sessionID).Scan(&s.ID, &s.AdventureID, &s.CurrentActID, &s.CurrentSceneID, &s.Status, &s.ElapsedSeconds, &s.StartedAt, &s.CompletedAt,
		&s.AdventureTitle, &s.ActTitle, &s.SceneTitle, &s.SceneEstimated, &s.ActNumber, &s.SceneNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get total counts
	db.DB.QueryRow("SELECT COUNT(*) FROM oneshot_acts WHERE adventure_id=?", s.AdventureID).Scan(&s.TotalActs)
	db.DB.QueryRow("SELECT COUNT(*) FROM oneshot_scenes WHERE act_id=?", s.CurrentActID).Scan(&s.TotalScenes)

	// Get scene timings
	rows, err := db.DB.Query(`
		SELECT st.id, st.session_id, st.scene_id, st.elapsed_seconds, st.status, COALESCE(st.started_at,''), COALESCE(st.completed_at,''),
			COALESCE(sc.title,''), COALESCE(sc.scene_type,''), COALESCE(sc.estimated_minutes,0)
		FROM scene_timings st
		LEFT JOIN oneshot_scenes sc ON sc.id = st.scene_id
		WHERE st.session_id=?
		ORDER BY st.id
	`, sessionID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var st models.SceneTiming
			if err := rows.Scan(&st.ID, &st.SessionID, &st.SceneID, &st.ElapsedSeconds, &st.Status, &st.StartedAt, &st.CompletedAt,
				&st.SceneTitle, &st.SceneType, &st.EstimatedMin); err == nil {
				s.SceneTimings = append(s.SceneTimings, st)
			}
		}
	}

	c.JSON(http.StatusOK, s)
}

func UpdatePacingTimers(c *gin.Context) {
	sessionID := c.Param("id")

	// Increment elapsed_seconds for the session
	db.DB.Exec("UPDATE session_pacing SET elapsed_seconds = elapsed_seconds + 5 WHERE id=? AND status='running'", sessionID)

	// Increment current scene timing
	db.DB.Exec("UPDATE scene_timings SET elapsed_seconds = elapsed_seconds + 5 WHERE session_id=? AND status='active'", sessionID)

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func PausePacingSession(c *gin.Context) {
	sessionID := c.Param("id")
	_, err := db.DB.Exec("UPDATE session_pacing SET status='paused' WHERE id=? AND status='running'", sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "paused"})
}

func ResumePacingSession(c *gin.Context) {
	sessionID := c.Param("id")
	_, err := db.DB.Exec("UPDATE session_pacing SET status='running' WHERE id=? AND status='paused'", sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "running"})
}

func CompletePacingSession(c *gin.Context) {
	sessionID := c.Param("id")

	// Mark current scene timing as completed
	db.DB.Exec("UPDATE scene_timings SET status='completed', completed_at=datetime('now') WHERE session_id=? AND status='active'", sessionID)

	// Mark session as completed
	_, err := db.DB.Exec("UPDATE session_pacing SET status='completed', completed_at=datetime('now') WHERE id=? AND status IN ('running','paused')", sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "completed"})
}

func AdvanceToNextScene(c *gin.Context) {
	sessionID := c.Param("id")

	// Get current scene info
	var currentActID, currentSceneID int64
	err := db.DB.QueryRow("SELECT COALESCE(current_act_id,0), COALESCE(current_scene_id,0) FROM session_pacing WHERE id=?", sessionID).Scan(&currentActID, &currentSceneID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	if currentSceneID > 0 {
		// Mark current scene timing as completed
		db.DB.Exec(`UPDATE scene_timings SET status='completed', elapsed_seconds=COALESCE((SELECT elapsed_seconds FROM scene_timings WHERE session_id=? AND scene_id=? AND status='active' LIMIT 1), elapsed_seconds), completed_at=datetime('now') WHERE session_id=? AND scene_id=? AND status='active'`,
			sessionID, currentSceneID, sessionID, currentSceneID)
	}

	// If no current act/scene, try to set first scene from adventure
	if currentActID == 0 && currentSceneID == 0 {
		var advID int64
		db.DB.QueryRow("SELECT adventure_id FROM session_pacing WHERE id=?", sessionID).Scan(&advID)
		var firstActID int64
		err = db.DB.QueryRow("SELECT id FROM oneshot_acts WHERE adventure_id=? ORDER BY number ASC LIMIT 1", advID).Scan(&firstActID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no acts in adventure"})
			return
		}
		var firstSceneID int64
		err = db.DB.QueryRow("SELECT id FROM oneshot_scenes WHERE act_id=? ORDER BY number ASC LIMIT 1", firstActID).Scan(&firstSceneID)
		if err != nil {
			// Act has no scenes, advance to next act later
			db.DB.Exec("UPDATE session_pacing SET current_act_id=? WHERE id=?", firstActID, sessionID)
			currentActID = firstActID
		} else {
			db.DB.Exec("UPDATE session_pacing SET current_act_id=?, current_scene_id=? WHERE id=?", firstActID, firstSceneID, sessionID)
			db.DB.Exec("INSERT INTO scene_timings(session_id, scene_id, status) VALUES(?,?,'active')", sessionID, firstSceneID)
			c.JSON(http.StatusOK, gin.H{"status": "advanced"})
			return
		}
	}

	// Find next scene in same act
	if currentActID > 0 && currentSceneID > 0 {
		var nextSceneID int64
		err = db.DB.QueryRow("SELECT id FROM oneshot_scenes WHERE act_id=? AND number > (SELECT number FROM oneshot_scenes WHERE id=?) ORDER BY number ASC LIMIT 1",
			currentActID, currentSceneID).Scan(&nextSceneID)
		if err == nil {
			db.DB.Exec("UPDATE session_pacing SET current_scene_id=? WHERE id=?", nextSceneID, sessionID)
			db.DB.Exec("INSERT INTO scene_timings(session_id, scene_id, status) VALUES(?,?,'active')", sessionID, nextSceneID)
			c.JSON(http.StatusOK, gin.H{"status": "advanced"})
			return
		}
	}

	// No more scenes in this act, find next act
	if currentActID > 0 {
		var nextActID int64
		err = db.DB.QueryRow("SELECT id FROM oneshot_acts WHERE adventure_id=(SELECT adventure_id FROM session_pacing WHERE id=?) AND number > (SELECT number FROM oneshot_acts WHERE id=?) ORDER BY number ASC LIMIT 1",
			sessionID, currentActID).Scan(&nextActID)
		if err == nil {
			var firstSceneID int64
			err = db.DB.QueryRow("SELECT id FROM oneshot_scenes WHERE act_id=? ORDER BY number ASC LIMIT 1", nextActID).Scan(&firstSceneID)
			if err == nil {
				db.DB.Exec("UPDATE session_pacing SET current_act_id=?, current_scene_id=? WHERE id=?", nextActID, firstSceneID, sessionID)
				db.DB.Exec("INSERT INTO scene_timings(session_id, scene_id, status) VALUES(?,?,'active')", sessionID, firstSceneID)
				c.JSON(http.StatusOK, gin.H{"status": "advanced"})
				return
			}
			// Act has no scenes, advance to next act without scenes
			currentActID = nextActID
			db.DB.Exec("UPDATE session_pacing SET current_act_id=?, current_scene_id=NULL WHERE id=?", nextActID, sessionID)
		}
	}

	// No more acts - complete session
	db.DB.Exec("UPDATE session_pacing SET status='completed', completed_at=datetime('now') WHERE id=?", sessionID)
	c.JSON(http.StatusOK, gin.H{"status": "completed", "message": "all scenes completed"})
}

// ─── HTMX Pacing Handlers ───

func HtmxGetPacingDashboard(c *gin.Context) {
	sessionID := c.Param("id")

	var s models.SessionPacing
	err := db.DB.QueryRow(`
		SELECT sp.id, sp.adventure_id, sp.current_act_id, sp.current_scene_id, sp.status, sp.elapsed_seconds, sp.started_at, COALESCE(sp.completed_at,''),
			COALESCE(oa.title,''), COALESCE(a.title,''), COALESCE(sc.title,''), COALESCE(sc.estimated_minutes,0),
			COALESCE(a.number,0), COALESCE(sc.number,0)
		FROM session_pacing sp
		LEFT JOIN oneshot_adventures oa ON oa.id = sp.adventure_id
		LEFT JOIN oneshot_acts a ON a.id = sp.current_act_id
		LEFT JOIN oneshot_scenes sc ON sc.id = sp.current_scene_id
		WHERE sp.id=?
	`, sessionID).Scan(&s.ID, &s.AdventureID, &s.CurrentActID, &s.CurrentSceneID, &s.Status, &s.ElapsedSeconds, &s.StartedAt, &s.CompletedAt,
		&s.AdventureTitle, &s.ActTitle, &s.SceneTitle, &s.SceneEstimated, &s.ActNumber, &s.SceneNumber)
	if err != nil {
		c.String(http.StatusNotFound, "Session not found")
		return
	}

	db.DB.QueryRow("SELECT COUNT(*) FROM oneshot_acts WHERE adventure_id=?", s.AdventureID).Scan(&s.TotalActs)
	var sceneCount int
	db.DB.QueryRow("SELECT COUNT(*) FROM oneshot_scenes sc JOIN oneshot_acts a ON a.id=sc.act_id WHERE a.adventure_id=?", s.AdventureID).Scan(&sceneCount)
	_ = sceneCount

	rows, err := db.DB.Query(`
		SELECT st.id, st.session_id, st.scene_id, st.elapsed_seconds, st.status, COALESCE(st.started_at,''), COALESCE(st.completed_at,''),
			COALESCE(sc.title,''), COALESCE(sc.scene_type,''), COALESCE(sc.estimated_minutes,0)
		FROM scene_timings st
		LEFT JOIN oneshot_scenes sc ON sc.id = st.scene_id
		WHERE st.session_id=?
		ORDER BY st.id
	`, sessionID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var st models.SceneTiming
			if err := rows.Scan(&st.ID, &st.SessionID, &st.SceneID, &st.ElapsedSeconds, &st.Status, &st.StartedAt, &st.CompletedAt,
				&st.SceneTitle, &st.SceneType, &st.EstimatedMin); err == nil {
				s.SceneTimings = append(s.SceneTimings, st)
			}
		}
	}

	paceCls, pacePct, paceLbl := computePace(s.SceneEstimated, s.SceneTimings)

	c.HTML(http.StatusOK, "oneshot_pacing.html", gin.H{
		"Session":     s,
		"Elapsed":     formatDuration(s.ElapsedSeconds),
		"PaceClass":   paceCls,
		"PacePercent": pacePct,
		"PaceLabel":   paceLbl,
	})
}

func computePace(estimatedMin int, timings []models.SceneTiming) (class string, percent int, label string) {
	if estimatedMin == 0 || len(timings) == 0 {
		return "bg-success", 0, "No Estimate"
	}
	for _, t := range timings {
		if t.Status == "active" {
			elapsedMin := t.ElapsedSeconds / 60
			ratio := float64(elapsedMin) / float64(estimatedMin)
			pct := int(ratio * 100)
			if pct > 100 {
				pct = 100
			}
			if ratio >= 1.0 {
				return "bg-danger", pct, "Over Time!"
			} else if ratio >= 0.8 {
				return "bg-warning text-dark", pct, "Warning"
			}
			return "bg-success", pct, "On Track"
		}
	}
	return "bg-success", 0, "No Data"
}

func paceClass(estimatedMin int, timings []models.SceneTiming) string {
	if estimatedMin == 0 || len(timings) == 0 {
		return "bg-success"
	}
	for _, t := range timings {
		if t.Status == "active" {
			elapsedMin := t.ElapsedSeconds / 60
			ratio := float64(elapsedMin) / float64(estimatedMin)
			if ratio >= 1.0 {
				return "bg-danger"
			} else if ratio >= 0.8 {
				return "bg-warning text-dark"
			}
			return "bg-success"
		}
	}
	return "bg-success"
}

func formatDuration(seconds int) string {
	m := seconds / 60
	s := seconds % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}

// ─── Clue/Mystery Tracker API Handlers ───

func ListClues(c *gin.Context) {
	adventureID := c.Param("id")
	rows, err := db.DB.Query("SELECT id, adventure_id, title, description, clue_type, is_red_herring, is_revealed, sort_order, notes, created_at, updated_at FROM clues WHERE adventure_id=? ORDER BY sort_order, id", adventureID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	clues := make([]models.Clue, 0)
	for rows.Next() {
		var cl models.Clue
		if err := rows.Scan(&cl.ID, &cl.AdventureID, &cl.Title, &cl.Description, &cl.ClueType, &cl.IsRedHerring, &cl.IsRevealed, &cl.SortOrder, &cl.Notes, &cl.CreatedAt, &cl.UpdatedAt); err == nil {
			loadClueRelations(&cl)
			clues = append(clues, cl)
		}
	}
	c.JSON(http.StatusOK, clues)
}

func CreateClue(c *gin.Context) {
	adventureID := c.Param("id")
	userID, _ := c.Get("user_id")

	// Verify adventure belongs to user
	var exists bool
	db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM oneshot_adventures WHERE id=? AND user_id=?)", adventureID, userID).Scan(&exists)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "adventure not found"})
		return
	}

	var input struct {
		Title        string `json:"title"`
		Description  string `json:"description"`
		ClueType     string `json:"clue_type"`
		IsRedHerring bool   `json:"is_red_herring"`
		SortOrder    int    `json:"sort_order"`
		Notes        string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if input.ClueType == "" {
		input.ClueType = "direct"
	}

	result, err := db.DB.Exec("INSERT INTO clues(adventure_id, title, description, clue_type, is_red_herring, sort_order, notes) VALUES(?,?,?,?,?,?,?)",
		adventureID, input.Title, input.Description, input.ClueType, boolToInt(input.IsRedHerring), input.SortOrder, input.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func GetClue(c *gin.Context) {
	id := c.Param("id")
	var cl models.Clue
	err := db.DB.QueryRow("SELECT id, adventure_id, title, description, clue_type, is_red_herring, is_revealed, sort_order, notes, created_at, updated_at FROM clues WHERE id=?", id).
		Scan(&cl.ID, &cl.AdventureID, &cl.Title, &cl.Description, &cl.ClueType, &cl.IsRedHerring, &cl.IsRevealed, &cl.SortOrder, &cl.Notes, &cl.CreatedAt, &cl.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "clue not found"})
		return
	}
	loadClueRelations(&cl)
	c.JSON(http.StatusOK, cl)
}

func UpdateClue(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		Title        string `json:"title"`
		Description  string `json:"description"`
		ClueType     string `json:"clue_type"`
		IsRedHerring bool   `json:"is_red_herring"`
		SortOrder    int    `json:"sort_order"`
		Notes        string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := db.DB.Exec("UPDATE clues SET title=?, description=?, clue_type=?, is_red_herring=?, sort_order=?, notes=?, updated_at=datetime('now') WHERE id=?",
		input.Title, input.Description, input.ClueType, boolToInt(input.IsRedHerring), input.SortOrder, input.Notes, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func DeleteClue(c *gin.Context) {
	id := c.Param("id")
	db.DB.Exec("DELETE FROM clues WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func RevealClue(c *gin.Context) {
	id := c.Param("id")
	db.DB.Exec("UPDATE clues SET is_revealed=1, updated_at=datetime('now') WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"status": "revealed"})
}

func HideClue(c *gin.Context) {
	id := c.Param("id")
	db.DB.Exec("UPDATE clues SET is_revealed=0, updated_at=datetime('now') WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"status": "hidden"})
}

func AddClueDependency(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		DependsOnID int64 `json:"depends_on_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := db.DB.Exec("INSERT OR IGNORE INTO clue_dependencies(clue_id, depends_on_id) VALUES(?,?)", id, input.DependsOnID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "linked"})
}

func RemoveClueDependency(c *gin.Context) {
	id := c.Param("id")
	depID := c.Param("did")
	db.DB.Exec("DELETE FROM clue_dependencies WHERE clue_id=? AND depends_on_id=?", id, depID)
	c.JSON(http.StatusOK, gin.H{"status": "removed"})
}

func LinkClueNPC(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		NPCID int64 `json:"npc_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := db.DB.Exec("INSERT OR IGNORE INTO clue_npcs(clue_id, npc_id) VALUES(?,?)", id, input.NPCID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "linked"})
}

func UnlinkClueNPC(c *gin.Context) {
	id := c.Param("id")
	npcID := c.Param("nid")
	db.DB.Exec("DELETE FROM clue_npcs WHERE clue_id=? AND npc_id=?", id, npcID)
	c.JSON(http.StatusOK, gin.H{"status": "removed"})
}

func LinkClueLocation(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		LocationID int64 `json:"location_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := db.DB.Exec("INSERT OR IGNORE INTO clue_locations(clue_id, location_id) VALUES(?,?)", id, input.LocationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "linked"})
}

func UnlinkClueLocation(c *gin.Context) {
	id := c.Param("id")
	locID := c.Param("lid")
	db.DB.Exec("DELETE FROM clue_locations WHERE clue_id=? AND location_id=?", id, locID)
	c.JSON(http.StatusOK, gin.H{"status": "removed"})
}

func loadClueRelations(cl *models.Clue) {
	cl.Dependencies = make([]models.ClueDependency, 0)
	cl.DependedBy = make([]models.ClueDependency, 0)
	cl.NPCs = make([]models.ClueNPC, 0)
	cl.Locations = make([]models.ClueLocation, 0)

	// Dependencies
	depRows, err := db.DB.Query("SELECT cd.id, cd.clue_id, cd.depends_on_id, COALESCE(c.title,'') FROM clue_dependencies cd LEFT JOIN clues c ON c.id=cd.depends_on_id WHERE cd.clue_id=?", cl.ID)
	if err == nil {
		defer depRows.Close()
		for depRows.Next() {
			var d models.ClueDependency
			if depRows.Scan(&d.ID, &d.ClueID, &d.DependsOnID, &d.DependsOnTitle) == nil {
				cl.Dependencies = append(cl.Dependencies, d)
			}
		}
	}
	// Depended by
	depByRows, err := db.DB.Query("SELECT cd.id, cd.clue_id, cd.depends_on_id, COALESCE(c.title,'') FROM clue_dependencies cd LEFT JOIN clues c ON c.id=cd.clue_id WHERE cd.depends_on_id=?", cl.ID)
	if err == nil {
		defer depByRows.Close()
		for depByRows.Next() {
			var d models.ClueDependency
			if depByRows.Scan(&d.ID, &d.ClueID, &d.DependsOnID, &d.DependsOnTitle) == nil {
				cl.DependedBy = append(cl.DependedBy, d)
			}
		}
	}
	// NPCs
	npcRows, err := db.DB.Query("SELECT cn.id, cn.clue_id, cn.npc_id, COALESCE(n.name,'') FROM clue_npcs cn LEFT JOIN npcs n ON n.id=cn.npc_id WHERE cn.clue_id=?", cl.ID)
	if err == nil {
		defer npcRows.Close()
		for npcRows.Next() {
			var n models.ClueNPC
			if npcRows.Scan(&n.ID, &n.ClueID, &n.NPCID, &n.NPCName) == nil {
				cl.NPCs = append(cl.NPCs, n)
			}
		}
	}
	// Locations
	locRows, err := db.DB.Query("SELECT cl.id, cl.clue_id, cl.location_id, COALESCE(l.name,'') FROM clue_locations cl LEFT JOIN locations l ON l.id=cl.location_id WHERE cl.clue_id=?", cl.ID)
	if err == nil {
		defer locRows.Close()
		for locRows.Next() {
			var l models.ClueLocation
			if locRows.Scan(&l.ID, &l.ClueID, &l.LocationID, &l.LocationName) == nil {
				cl.Locations = append(cl.Locations, l)
			}
		}
	}
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

// ─── HTMX Clue Handlers ───

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

	c.HTML(http.StatusOK, "oneshot_clues.html", gin.H{
		"Clues": clues,
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

	c.HTML(http.StatusOK, "oneshot_clue_detail.html", gin.H{
		"Clue": cl,
	})
}

func HtmxNewClueForm(c *gin.Context) {
	adventureID := c.Param("id")
	c.HTML(http.StatusOK, "oneshot_clue_form.html", gin.H{
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
	c.HTML(http.StatusOK, "oneshot_clue_form.html", gin.H{
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

func ListPregens(c *gin.Context) {
	userID, _ := c.Get("user_id")
	rows, err := db.DB.Query("SELECT id, user_id, name, race, class, subclass, level, background, alignment, str, dex, con, int, wis, cha, hp, ac, speed, skills, equipment, spells, features, personality, backstory, portrait_url, notes, created_at, updated_at FROM pregen_characters WHERE user_id=? ORDER BY updated_at DESC", userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
	c.JSON(http.StatusOK, chars)
}

func CreatePregen(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var input struct {
		Name       string `json:"name"`
		Race       string `json:"race"`
		Class      string `json:"class"`
		Subclass   string `json:"subclass"`
		Level      int    `json:"level"`
		Background string `json:"background"`
		Alignment  string `json:"alignment"`
		Str        int    `json:"str"`
		Dex        int    `json:"dex"`
		Con        int    `json:"con"`
		Int        int    `json:"int"`
		Wis        int    `json:"wis"`
		Cha        int    `json:"cha"`
		HP         int    `json:"hp"`
		AC         int    `json:"ac"`
		Speed      int    `json:"speed"`
		Skills     string `json:"skills"`
		Equipment  string `json:"equipment"`
		Spells     string `json:"spells"`
		Features   string `json:"features"`
		Personality string `json:"personality"`
		Backstory  string `json:"backstory"`
		PortraitURL string `json:"portrait_url"`
		Notes      string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if input.Level == 0 {
		input.Level = 1
	}
	if input.Speed == 0 {
		input.Speed = 30
	}

	result, err := db.DB.Exec(`
		INSERT INTO pregen_characters(user_id, name, race, class, subclass, level, background, alignment,
			str, dex, con, int, wis, cha, hp, ac, speed, skills, equipment, spells, features, personality, backstory, portrait_url, notes)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		userID, input.Name, input.Race, input.Class, input.Subclass, input.Level, input.Background, input.Alignment,
		input.Str, input.Dex, input.Con, input.Int, input.Wis, input.Cha, input.HP, input.AC, input.Speed,
		input.Skills, input.Equipment, input.Spells, input.Features, input.Personality, input.Backstory, input.PortraitURL, input.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func GetPregen(c *gin.Context) {
	id := c.Param("id")
	var ch models.PregeneratedCharacter
	err := db.DB.QueryRow("SELECT id, user_id, name, race, class, subclass, level, background, alignment, str, dex, con, int, wis, cha, hp, ac, speed, skills, equipment, spells, features, personality, backstory, portrait_url, notes, created_at, updated_at FROM pregen_characters WHERE id=?", id).
		Scan(&ch.ID, &ch.UserID, &ch.Name, &ch.Race, &ch.Class, &ch.Subclass, &ch.Level, &ch.Background, &ch.Alignment,
			&ch.Str, &ch.Dex, &ch.Con, &ch.Int, &ch.Wis, &ch.Cha, &ch.HP, &ch.AC, &ch.Speed,
			&ch.Skills, &ch.Equipment, &ch.Spells, &ch.Features, &ch.Personality, &ch.Backstory,
			&ch.PortraitURL, &ch.Notes, &ch.CreatedAt, &ch.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "pregen not found"})
		return
	}
	c.JSON(http.StatusOK, ch)
}

func UpdatePregen(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		Name       string `json:"name"`
		Race       string `json:"race"`
		Class      string `json:"class"`
		Subclass   string `json:"subclass"`
		Level      int    `json:"level"`
		Background string `json:"background"`
		Alignment  string `json:"alignment"`
		Str        int    `json:"str"`
		Dex        int    `json:"dex"`
		Con        int    `json:"con"`
		Int        int    `json:"int"`
		Wis        int    `json:"wis"`
		Cha        int    `json:"cha"`
		HP         int    `json:"hp"`
		AC         int    `json:"ac"`
		Speed      int    `json:"speed"`
		Skills     string `json:"skills"`
		Equipment  string `json:"equipment"`
		Spells     string `json:"spells"`
		Features   string `json:"features"`
		Personality string `json:"personality"`
		Backstory  string `json:"backstory"`
		PortraitURL string `json:"portrait_url"`
		Notes      string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := db.DB.Exec(`
		UPDATE pregen_characters SET name=?, race=?, class=?, subclass=?, level=?, background=?, alignment=?,
			str=?, dex=?, con=?, int=?, wis=?, cha=?, hp=?, ac=?, speed=?, skills=?, equipment=?, spells=?, features=?,
			personality=?, backstory=?, portrait_url=?, notes=?, updated_at=datetime('now') WHERE id=?`,
		input.Name, input.Race, input.Class, input.Subclass, input.Level, input.Background, input.Alignment,
		input.Str, input.Dex, input.Con, input.Int, input.Wis, input.Cha, input.HP, input.AC, input.Speed,
		input.Skills, input.Equipment, input.Spells, input.Features, input.Personality, input.Backstory, input.PortraitURL, input.Notes, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func DeletePregen(c *gin.Context) {
	id := c.Param("id")
	db.DB.Exec("DELETE FROM pregen_characters WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// Class role assignments for party balance
var classRoles = map[string][]string{
	"barbarian":  {"tank"},
	"fighter":    {"tank", "damage"},
	"paladin":    {"tank", "healer"},
	"cleric":     {"healer"},
	"druid":      {"healer", "support"},
	"wizard":     {"damage", "support"},
	"sorcerer":   {"damage"},
	"warlock":    {"damage"},
	"bard":       {"support", "healer"},
	"rogue":      {"damage", "skill"},
	"monk":       {"damage"},
	"ranger":     {"damage", "skill"},
	"artificer":  {"support"},
}

var allRoles = []string{"tank", "healer", "damage", "support", "skill"}

func CheckPartyBalance(c *gin.Context) {
	userID, _ := c.Get("user_id")
	rows, err := db.DB.Query("SELECT id, user_id, name, race, class, subclass, level, background, alignment, str, dex, con, int, wis, cha, hp, ac, speed, skills, equipment, spells, features, personality, backstory, portrait_url, notes, created_at, updated_at FROM pregen_characters WHERE user_id=? ORDER BY updated_at DESC", userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

	roleCounts := map[string]int{}
	for _, r := range allRoles {
		roleCounts[r] = 0
	}
	for _, ch := range chars {
		class := ch.Class
		roles, ok := classRoles[class]
		if !ok {
			roles = classRoles["fighter"]
		}
		for _, r := range roles {
			roleCounts[r]++
		}
	}

	missing := make([]string, 0)
	for _, r := range allRoles {
		if roleCounts[r] == 0 {
			missing = append(missing, r)
		}
	}

	score := 0
	for _, r := range allRoles {
		if roleCounts[r] > 0 {
			score += 2
		}
	}
	score += len(chars) * 2
	if score > 20 {
		score = 20
	}

	rating := "poor"
	suggestion := "Consider creating more characters to cover missing roles."
	if len(chars) >= 3 && len(missing) <= 2 {
		rating = "fair"
		suggestion = "Your party could use some additional coverage."
	}
	if len(chars) >= 4 && len(missing) <= 1 {
		rating = "good"
		suggestion = "Your party is well-balanced for most adventures."
	}
	if len(chars) >= 4 && len(missing) == 0 {
		rating = "excellent"
		suggestion = "Your party covers all essential roles!"
	}
	if len(chars) >= 5 && len(missing) <= 1 {
		rating = "great"
		suggestion = "Your party has excellent depth and versatility."
	}

	c.JSON(http.StatusOK, models.PartyBalance{
		Characters: chars,
		Roles:      roleCounts,
		Score:      score,
		Rating:     rating,
		Missing:    missing,
		Suggestion: suggestion,
	})
}

func GeneratePregen(c *gin.Context) {
	userID, _ := c.Get("user_id")
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
	name := names[rand.Intn(len(names))]

	str, dex, con, intel, wis, cha := 10, 10, 10, 10, 10, 10
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
	}

	hp := 8 + (con-10)/2 + (genLevel-1)*5
	ac := 12 + (dex-10)/2

	result, err := db.DB.Exec(`
		INSERT INTO pregen_characters(user_id, name, race, class, level, str, dex, con, int, wis, cha, hp, ac, speed)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,30)`,
		userID, name, genRace, genClass, genLevel, str, dex, con, intel, wis, cha, hp, ac)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// ─── HTMX Pregenerated Characters ───

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
	c.HTML(http.StatusOK, "oneshot_pregens.html", gin.H{"Pregens": chars})
}

func HtmxGeneratePregen(c *gin.Context) {
	GeneratePregen(c)
	// After generating, redirect to the list
	c.Request.Method = "GET"
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
	c.HTML(http.StatusOK, "oneshot_pregen_card.html", gin.H{"C": ch})
}
