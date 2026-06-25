package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

func ListEncounters(c *gin.Context) {
	userID, _ := c.Get("user_id")
	campaignID := c.Query("campaign_id")

	var rows *sql.Rows
	var err error
	if campaignID != "" {
		rows, err = db.DB.Query("SELECT id, campaign_id, user_id, name, description, environment, difficulty, xp_budget, total_xp, notes, created_at FROM encounter_templates WHERE campaign_id=? ORDER BY created_at DESC", campaignID)
	} else {
		rows, err = db.DB.Query("SELECT id, campaign_id, user_id, name, description, environment, difficulty, xp_budget, total_xp, notes, created_at FROM encounter_templates WHERE user_id=? ORDER BY created_at DESC", userID)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var out = make([]models.EncounterTemplate, 0)
	for rows.Next() {
		var e models.EncounterTemplate
		rows.Scan(&e.ID, &e.CampaignID, &e.UserID, &e.Name, &e.Description, &e.Environment, &e.Difficulty, &e.XPBudget, &e.TotalXP, &e.Notes, &e.CreatedAt)
		out = append(out, e)
	}
	c.JSON(http.StatusOK, out)
}

func CreateEncounter(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var e models.EncounterTemplate
	if err := c.ShouldBindJSON(&e); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := db.DB.Exec("INSERT INTO encounter_templates(campaign_id,user_id,name,description,environment,difficulty,xp_budget,total_xp,notes) VALUES(?,?,?,?,?,?,?,?,?)",
		e.CampaignID, userID, e.Name, e.Description, e.Environment, e.Difficulty, e.XPBudget, e.TotalXP, e.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func GetEncounter(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var e models.EncounterTemplate
	err := db.DB.QueryRow("SELECT id, campaign_id, user_id, name, description, environment, difficulty, xp_budget, total_xp, notes, created_at FROM encounter_templates WHERE id=?", id).
		Scan(&e.ID, &e.CampaignID, &e.UserID, &e.Name, &e.Description, &e.Environment, &e.Difficulty, &e.XPBudget, &e.TotalXP, &e.Notes, &e.CreatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "encounter not found"})
		return
	}
	// Load monsters
	mrows, err := db.DB.Query("SELECT id, encounter_id, name, count, cr, xp, ac, hp, initiative_mod, source, notes, compendium_monster_id FROM encounter_monsters WHERE encounter_id=? ORDER BY id", id)
	if err == nil {
		defer mrows.Close()
		e.Monsters = make([]models.EncounterMonster, 0)
		for mrows.Next() {
			var m models.EncounterMonster
			var compID sql.NullInt64
			mrows.Scan(&m.ID, &m.EncounterID, &m.Name, &m.Count, &m.CR, &m.XP, &m.AC, &m.HP, &m.InitiativeMod, &m.Source, &m.Notes, &compID)
			if compID.Valid {
				v := compID.Int64
				m.CompendiumMonsterID = &v
			}
			e.Monsters = append(e.Monsters, m)
		}
	}
	c.JSON(http.StatusOK, e)
}

func UpdateEncounter(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var e models.EncounterTemplate
	if err := c.ShouldBindJSON(&e); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Exec("UPDATE encounter_templates SET name=?, description=?, environment=?, difficulty=?, xp_budget=?, total_xp=?, notes=? WHERE id=?",
		e.Name, e.Description, e.Environment, e.Difficulty, e.XPBudget, e.TotalXP, e.Notes, id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteEncounter(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	db.DB.Exec("DELETE FROM encounter_templates WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Encounter monsters

func ListCampaignEncounterMonsters(c *gin.Context) {
	encounterID := c.Param("id")
	rows, err := db.DB.Query("SELECT id, encounter_id, name, count, cr, xp, ac, hp, initiative_mod, source, notes, compendium_monster_id FROM encounter_monsters WHERE encounter_id=? ORDER BY id", encounterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := make([]models.EncounterMonster, 0)
	for rows.Next() {
		var m models.EncounterMonster
		rows.Scan(&m.ID, &m.EncounterID, &m.Name, &m.Count, &m.CR, &m.XP, &m.AC, &m.HP, &m.InitiativeMod, &m.Source, &m.Notes, &m.CompendiumMonsterID)
		out = append(out, m)
	}
	c.JSON(http.StatusOK, out)
}

func CreateCampaignEncounterMonster(c *gin.Context) {
	eid, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		Name                string `json:"name"`
		Count               int    `json:"count"`
		CR                  string `json:"cr"`
		XP                  int    `json:"xp"`
		AC                  int    `json:"ac"`
		HP                  int    `json:"hp"`
		InitiativeMod       int    `json:"initiative_mod"`
		Source              string `json:"source"`
		Notes               string `json:"notes"`
		CompendiumMonsterID *int64 `json:"compendium_monster_id,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Count < 1 {
		req.Count = 1
	}
	if req.Source == "" {
		req.Source = "homebrew"
	}

	// If compendium_monster_id is provided, pre-fill from compendium
	if req.CompendiumMonsterID != nil && *req.CompendiumMonsterID > 0 {
		var name, cr, source string
		var ac, hp int
		err := db.DB.QueryRow("SELECT name, ac, hp, cr, source FROM compendium_monsters WHERE id=?", *req.CompendiumMonsterID).
			Scan(&name, &ac, &hp, &cr, &source)
		if err == nil {
			if req.Name == "" {
				req.Name = name
			}
			if req.AC == 0 {
				req.AC = ac
			}
			if req.HP == 0 {
				req.HP = hp
			}
			if req.CR == "" {
				req.CR = cr
			}
			if req.Source == "homebrew" {
				req.Source = source
			}
		}
	}

	result, err := db.DB.Exec("INSERT INTO encounter_monsters(encounter_id,name,count,cr,xp,ac,hp,initiative_mod,source,notes,compendium_monster_id) VALUES(?,?,?,?,?,?,?,?,?,?,?)",
		eid, req.Name, req.Count, req.CR, req.XP, req.AC, req.HP, req.InitiativeMod, req.Source, req.Notes, req.CompendiumMonsterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func AddEncounterMonster(c *gin.Context) {
	eid, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		Name                string `json:"name"`
		Count               int    `json:"count"`
		CR                  string `json:"cr"`
		XP                  int    `json:"xp"`
		AC                  int    `json:"ac"`
		HP                  int    `json:"hp"`
		InitiativeMod       int    `json:"initiative_mod"`
		Source              string `json:"source"`
		Notes               string `json:"notes"`
		CompendiumMonsterID *int64 `json:"compendium_monster_id,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Count < 1 {
		req.Count = 1
	}
	if req.Source == "" {
		req.Source = "homebrew"
	}

	// If compendium_monster_id is provided, pre-fill from compendium
	if req.CompendiumMonsterID != nil && *req.CompendiumMonsterID > 0 {
		var name, cr, source string
		var ac, hp int
		err := db.DB.QueryRow("SELECT name, ac, hp, cr, source FROM compendium_monsters WHERE id=?", *req.CompendiumMonsterID).
			Scan(&name, &ac, &hp, &cr, &source)
		if err == nil {
			if req.Name == "" {
				req.Name = name
			}
			if req.AC == 0 {
				req.AC = ac
			}
			if req.HP == 0 {
				req.HP = hp
			}
			if req.CR == "" {
				req.CR = cr
			}
			if req.Source == "homebrew" {
				req.Source = source
			}
		}
	}

	result, err := db.DB.Exec("INSERT INTO encounter_monsters(encounter_id,name,count,cr,xp,ac,hp,initiative_mod,source,notes,compendium_monster_id) VALUES(?,?,?,?,?,?,?,?,?,?,?)",
		eid, req.Name, req.Count, req.CR, req.XP, req.AC, req.HP, req.InitiativeMod, req.Source, req.Notes, req.CompendiumMonsterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func UpdateEncounterMonster(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("mid"), 10, 64)
	var req struct {
		Name                string `json:"name"`
		Count               int    `json:"count"`
		CR                  string `json:"cr"`
		XP                  int    `json:"xp"`
		AC                  int    `json:"ac"`
		HP                  int    `json:"hp"`
		InitiativeMod       int    `json:"initiative_mod"`
		Source              string `json:"source"`
		Notes               string `json:"notes"`
		CompendiumMonsterID *int64 `json:"compendium_monster_id,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Exec("UPDATE encounter_monsters SET name=?, count=?, cr=?, xp=?, ac=?, hp=?, initiative_mod=?, source=?, notes=?, compendium_monster_id=? WHERE id=?",
		req.Name, req.Count, req.CR, req.XP, req.AC, req.HP, req.InitiativeMod, req.Source, req.Notes, req.CompendiumMonsterID, id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteEncounterMonster(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("mid"), 10, 64)
	db.DB.Exec("DELETE FROM encounter_monsters WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// XP thresholds for encounter difficulty by party level
var xpThresholds = map[int]map[string]int{
	1:  {"easy": 25, "medium": 50, "hard": 75, "deadly": 100},
	2:  {"easy": 50, "medium": 100, "hard": 150, "deadly": 200},
	3:  {"easy": 75, "medium": 150, "hard": 225, "deadly": 400},
	4:  {"easy": 125, "medium": 250, "hard": 375, "deadly": 500},
	5:  {"easy": 250, "medium": 500, "hard": 750, "deadly": 1100},
	6:  {"easy": 300, "medium": 600, "hard": 900, "deadly": 1400},
	7:  {"easy": 350, "medium": 750, "hard": 1100, "deadly": 1700},
	8:  {"easy": 450, "medium": 900, "hard": 1400, "deadly": 2100},
	9:  {"easy": 550, "medium": 1100, "hard": 1600, "deadly": 2400},
	10: {"easy": 600, "medium": 1200, "hard": 1900, "deadly": 2800},
	11: {"easy": 800, "medium": 1600, "hard": 2400, "deadly": 3600},
	12: {"easy": 1000, "medium": 2000, "hard": 3000, "deadly": 4500},
	13: {"easy": 1100, "medium": 2200, "hard": 3400, "deadly": 5100},
	14: {"easy": 1250, "medium": 2500, "hard": 3800, "deadly": 5700},
	15: {"easy": 1400, "medium": 2800, "hard": 4300, "deadly": 6400},
	16: {"easy": 1600, "medium": 3200, "hard": 4800, "deadly": 7200},
	17: {"easy": 2000, "medium": 3900, "hard": 5900, "deadly": 8800},
	18: {"easy": 2100, "medium": 4200, "hard": 6300, "deadly": 9500},
	19: {"easy": 2400, "medium": 4900, "hard": 7300, "deadly": 10900},
	20: {"easy": 2500, "medium": 5700, "hard": 8500, "deadly": 12700},
}

// Monster XP values by CR
var monsterXP = map[string]int{
	"0":   10,
	"1/8": 25,
	"1/4": 50,
	"1/2": 100,
	"1":   200,
	"2":   450,
	"3":   700,
	"4":   1100,
	"5":   1800,
	"6":   2300,
	"7":   2900,
	"8":   3900,
	"9":   5000,
	"10":  5900,
	"11":  7200,
	"12":  8400,
	"13":  10000,
	"14":  11500,
	"15":  13000,
	"16":  15000,
	"17":  18000,
	"18":  20000,
	"19":  22000,
	"20":  25000,
	"21":  33000,
	"22":  41000,
	"23":  50000,
	"24":  62000,
	"25":  75000,
	"26":  90000,
	"27":  105000,
	"28":  120000,
	"29":  135000,
	"30":  155000,
}

func CalculateEncounterXP(c *gin.Context) {
	var req struct {
		PartyLevels []int             `json:"party_levels"`
		Monsters    []models.EncounterMonster `json:"monsters"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	totalMonsterXP := 0
	for _, m := range req.Monsters {
		xp, ok := monsterXP[m.CR]
		if !ok {
			xp = 0
		}
		totalMonsterXP += xp * m.Count
	}

	// Calculate difficulty thresholds for the party
	easyThreshold := 0
	mediumThreshold := 0
	hardThreshold := 0
	deadlyThreshold := 0
	for _, lvl := range req.PartyLevels {
		if t, ok := xpThresholds[lvl]; ok {
			easyThreshold += t["easy"]
			mediumThreshold += t["medium"]
			hardThreshold += t["hard"]
			deadlyThreshold += t["deadly"]
		}
	}

	// Determine difficulty
	difficulty := "easy"
	if totalMonsterXP >= deadlyThreshold {
		difficulty = "deadly"
	} else if totalMonsterXP >= hardThreshold {
		difficulty = "hard"
	} else if totalMonsterXP >= mediumThreshold {
		difficulty = "medium"
	}

	// Apply encounter size multiplier
	sizeMultiplier := 1.0
	monsterCount := 0
	for _, m := range req.Monsters {
		monsterCount += m.Count
	}
	partySize := len(req.PartyLevels)
	if monsterCount > 0 && partySize > 0 {
		ratio := float64(monsterCount) / float64(partySize)
		switch {
		case ratio >= 3:
			sizeMultiplier = 2.5
		case ratio >= 2:
			sizeMultiplier = 1.5
		case ratio >= 1.5:
			sizeMultiplier = 1.0
		case ratio >= 0.5:
			sizeMultiplier = 0.5
		default:
			sizeMultiplier = 0.5
		}
	}

	adjustedXP := int(float64(totalMonsterXP) * sizeMultiplier)

	c.JSON(http.StatusOK, gin.H{
		"total_xp":        totalMonsterXP,
		"adjusted_xp":     adjustedXP,
		"difficulty":      difficulty,
		"thresholds": gin.H{
			"easy":   easyThreshold,
			"medium": mediumThreshold,
			"hard":   hardThreshold,
			"deadly": deadlyThreshold,
		},
		"party_size":      partySize,
		"monster_count":   monsterCount,
		"size_multiplier": sizeMultiplier,
	})
}

func GetMonsterXP(c *gin.Context) {
	c.JSON(http.StatusOK, monsterXP)
}
