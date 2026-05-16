package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
)

type LevelUpPlan struct {
	ID          int64            `json:"id"`
	CharacterID int64            `json:"character_id"`
	TargetLevel int              `json:"target_level"`
	PlanData    []LevelPlanEntry `json:"plan_data"`
	Notes       string           `json:"notes"`
}

type LevelPlanEntry struct {
	Level     int    `json:"level"`
	Class     string `json:"class"`
	Subclass  string `json:"subclass"`
	Feat      string `json:"feat"`
	Spell     string `json:"spell"`
	ASI       string `json:"asi"`
	ClassFeature string `json:"class_feature"`
	Notes     string `json:"notes"`
}

func GetLevelUpPlan(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var plan LevelUpPlan
	plan.CharacterID = charID
	plan.PlanData = []LevelPlanEntry{}

	var planDataJSON string
	err := db.DB.QueryRow("SELECT id,target_level,plan_data,notes FROM level_up_plans WHERE character_id=?", charID).Scan(&plan.ID, &plan.TargetLevel, &planDataJSON, &plan.Notes)
	if err != nil {
		// No plan yet, return empty
		c.JSON(http.StatusOK, plan)
		return
	}
	json.Unmarshal([]byte(planDataJSON), &plan.PlanData)
	c.JSON(http.StatusOK, plan)
}

func SaveLevelUpPlan(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var plan LevelUpPlan
	if err := c.ShouldBindJSON(&plan); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	planDataJSON, _ := json.Marshal(plan.PlanData)

	// Check if plan exists
	var existingID int64
	err := db.DB.QueryRow("SELECT id FROM level_up_plans WHERE character_id=?", charID).Scan(&existingID)
	if err != nil {
		// Create new plan
		result, err2 := db.DB.Exec("INSERT INTO level_up_plans(character_id,target_level,plan_data,notes) VALUES(?,?,?,?)",
			charID, plan.TargetLevel, string(planDataJSON), plan.Notes)
		if err2 != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err2.Error()})
			return
		}
		id, _ := result.LastInsertId()
		c.JSON(http.StatusCreated, gin.H{"id": id})
	} else {
		// Update existing
		db.DB.Exec("UPDATE level_up_plans SET target_level=?,plan_data=?,notes=?,updated_at=datetime('now') WHERE id=?",
			plan.TargetLevel, string(planDataJSON), plan.Notes, existingID)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

func DeleteLevelUpPlan(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	db.DB.Exec("DELETE FROM level_up_plans WHERE character_id=?", charID)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func GetLevelUpSuggestions(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var level, str, dex, con, intel, wis, cha int
	var cls string
	db.DB.QueryRow("SELECT level, class, str, dex, con, int, wis, cha FROM characters WHERE id=?", charID).Scan(&level, &cls, &str, &dex, &con, &intel, &wis, &cha)

	suggestions := []map[string]any{}

	for lvl := level + 1; lvl <= 20; lvl++ {
		entry := map[string]any{
			"level": lvl,
			"features": []string{},
		}

		// ASI/Feat at levels 4, 8, 12, 16, 19
		if lvl%4 == 0 && lvl <= 20 {
			entry["has_asi"] = true
			lowestStat := "str"
			lowestVal := str
			if dex < lowestVal { lowestStat = "dex"; lowestVal = dex }
			if con < lowestVal { lowestStat = "con"; lowestVal = con }
			if intel < lowestVal { lowestStat = "int"; lowestVal = intel }
			if wis < lowestVal { lowestStat = "wis"; lowestVal = wis }
			if cha < lowestVal { lowestStat = "cha"; lowestVal = cha }
			entry["asi_suggestion"] = "Increase " + lowestStat + " (+1 or feat)"
		}

		// Class feature milestones
		if lvl == level+1 {
			entry["features"] = append(entry["features"].([]string), "New level: class features, HP increase")
		}

		// Proficiency bonus increase
		if lvl == 5 || lvl == 9 || lvl == 13 || lvl == 17 {
			entry["features"] = append(entry["features"].([]string), "Proficiency bonus increases!")
		}

		suggestions = append(suggestions, entry)
	}

	c.JSON(http.StatusOK, suggestions)
}
