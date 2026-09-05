package handlers

import (
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

type GenerateRequest struct {
	Template   string `json:"template"`
	Title      string `json:"title"`
	Difficulty string `json:"difficulty"`
	Minutes    int    `json:"estimated_minutes"`
	CampaignID *int64 `json:"campaign_id,omitempty"`
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
