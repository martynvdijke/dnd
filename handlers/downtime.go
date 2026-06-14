package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
)

type DowntimeActivity struct {
	ID            int64   `json:"id"`
	CharacterID   int64   `json:"character_id"`
	ActivityType  string  `json:"activity_type"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	DC            int     `json:"dc"`
	DaysRequired  int     `json:"days_required"`
	DaysCompleted int     `json:"days_completed"`
	CostPerDay    float64 `json:"cost_per_day"`
	TotalCost     float64 `json:"total_cost"`
	Reward        string  `json:"reward"`
	Status        string  `json:"status"`
	Notes         string  `json:"notes"`
}

var downtimeTypes = []string{"training", "crafting", "research", "carousing", "pit_fighting", "crime", "religious", "scribing", "gambling", "other"}

func ListDowntimeActivities(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	rows, err := db.DB.Query("SELECT id,character_id,activity_type,name,description,dc,days_required,days_completed,cost_per_day,total_cost,reward,status,notes FROM downtime_activities WHERE character_id=? ORDER BY status,created_at DESC", charID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var out = make([]DowntimeActivity, 0)
	for rows.Next() {
		var a DowntimeActivity
		rows.Scan(&a.ID, &a.CharacterID, &a.ActivityType, &a.Name, &a.Description, &a.DC, &a.DaysRequired, &a.DaysCompleted, &a.CostPerDay, &a.TotalCost, &a.Reward, &a.Status, &a.Notes)
		out = append(out, a)
	}
	c.JSON(http.StatusOK, out)
}

func CreateDowntimeActivity(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var a DowntimeActivity
	if err := c.ShouldBindJSON(&a); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if a.Status == "" {
		a.Status = "in-progress"
	}
	validType := false
	for _, t := range downtimeTypes {
		if a.ActivityType == t {
			validType = true
			break
		}
	}
	if !validType {
		a.ActivityType = "other"
	}
	result, err := db.DB.Exec("INSERT INTO downtime_activities(character_id,activity_type,name,description,dc,days_required,days_completed,cost_per_day,total_cost,reward,status,notes) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)",
		charID, a.ActivityType, a.Name, a.Description, a.DC, a.DaysRequired, a.DaysCompleted, a.CostPerDay, a.TotalCost, a.Reward, a.Status, a.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func UpdateDowntimeActivity(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var a DowntimeActivity
	if err := c.ShouldBindJSON(&a); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Exec("UPDATE downtime_activities SET activity_type=?,name=?,description=?,dc=?,days_required=?,days_completed=?,cost_per_day=?,total_cost=?,reward=?,status=?,notes=?,updated_at=datetime('now') WHERE id=?",
		a.ActivityType, a.Name, a.Description, a.DC, a.DaysRequired, a.DaysCompleted, a.CostPerDay, a.TotalCost, a.Reward, a.Status, a.Notes, id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteDowntimeActivity(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	db.DB.Exec("DELETE FROM downtime_activities WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func AdvanceDowntimeDay(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var a DowntimeActivity
	err := db.DB.QueryRow("SELECT id,days_required,days_completed,dc,status FROM downtime_activities WHERE id=?", id).Scan(&a.ID, &a.DaysRequired, &a.DaysCompleted, &a.DC, &a.Status)
	if err != nil || a.Status != "in-progress" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "activity not found or not in progress"})
		return
	}

	newCompleted := a.DaysCompleted + 1
	newStatus := "in-progress"
	result, err := getDicePool().Roll("1d20")
	skillCheck := 1
	if err == nil {
		fmt.Sscanf(string(result.Total), "%d", &skillCheck)
	}
	success := skillCheck >= a.DC

	if newCompleted >= a.DaysRequired {
		if success {
			newStatus = "complete"
		} else {
			newStatus = "failed"
		}
	}

	db.DB.Exec("UPDATE downtime_activities SET days_completed=?, status=?, updated_at=datetime('now') WHERE id=?", newCompleted, newStatus, id)
	c.JSON(http.StatusOK, gin.H{
		"ok":            true,
		"days_completed": newCompleted,
		"status":        newStatus,
		"skill_check":   skillCheck,
		"success":       success,
	})
}

func GetDowntimeTypes(c *gin.Context) {
	type DtInfo struct {
		Type        string `json:"type"`
		Name        string `json:"name"`
		Description string `json:"description"`
		DefaultDC   int    `json:"default_dc"`
		DefaultDays int    `json:"default_days"`
	}
	c.JSON(http.StatusOK, []DtInfo{
		{"training", "Training", "Learn new tools, languages, or skills", 10, 30},
		{"crafting", "Crafting", "Create items, potions, or scrolls", 12, 10},
		{"research", "Research", "Study lore, spells, or mysteries", 12, 20},
		{"carousing", "Carousing", "Gather information and make connections", 10, 5},
		{"pit_fighting", "Pit Fighting", "Earn gold through combat", 12, 5},
		{"crime", "Crime", "Take risks for bigger rewards", 15, 5},
		{"religious", "Religious Service", "Perform duties for a temple or deity", 10, 10},
		{"scribing", "Scribing", "Copy spells or write scrolls", 10, 5},
		{"gambling", "Gambling", "Test your luck", 10, 3},
		{"other", "Other", "Custom downtime activity", 10, 10},
	})
}
