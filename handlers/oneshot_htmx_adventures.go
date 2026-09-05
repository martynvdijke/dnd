package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

type htmxOneShotData struct {
	Adventure    *models.OneShotAdventure
	Adventures   []models.OneShotAdventure
	NPCs         []models.NPC
	Locations    []models.Location
	Encounters   []models.EncounterTemplate
	Act          *models.OneShotAct
	Scene        *models.OneShotScene
	SceneTypes   []string
	Templates    []string
	Difficulties []string
	Acts         []models.OneShotAct
	Dialogs      []models.OneShotSceneDialog
	Dialog       *models.OneShotSceneDialog
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
	a, err := loadAdventureDetail(c.Request.Context(), id)
	if err != nil {
		c.String(http.StatusNotFound, "not found")
		return
	}
	renderTemplate(c, "oneshot_detail.html", htmxOneShotData{Adventure: a})
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

	isMiniCampaign := 0
	if c.PostForm("is_mini_campaign") == "1" {
		isMiniCampaign = 1
	}

	result, err := db.DB.Exec("INSERT INTO oneshot_adventures(user_id, campaign_id, title, premise, hook, template, estimated_minutes, difficulty, notes, is_mini_campaign) VALUES(?,?,?,?,?,?,?,?,?,?)",
		userID, campaignID, title, premise, hook, template, minutes, difficulty, notes, isMiniCampaign)
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
	isMiniCampaign := 0
	if c.PostForm("is_mini_campaign") == "1" {
		isMiniCampaign = 1
	}

	db.DB.Exec("UPDATE oneshot_adventures SET title=?, premise=?, hook=?, template=?, estimated_minutes=?, difficulty=?, notes=?, is_mini_campaign=?, updated_at=datetime('now') WHERE id=?",
		title, premise, hook, template, minutes, difficulty, notes, isMiniCampaign, id)

	HtmxListOneShots(c)
}

func HtmxDeleteOneShot(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := c.Get("user_id")
	db.DB.Exec("DELETE FROM oneshot_adventures WHERE id=? AND user_id=?", id, userID)
	HtmxListOneShots(c)
}

// HTMX Act form
// API: Reorder dialogs
func ReRenderOneShotDetail(c *gin.Context, adventureID int64) {
	a, err := loadAdventureDetail(c.Request.Context(), adventureID)
	if err != nil {
		c.String(http.StatusNotFound, "not found")
		return
	}
	renderTemplate(c, "oneshot_detail.html", htmxOneShotData{Adventure: a})
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
