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
