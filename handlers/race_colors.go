package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"villum/db"
)

// RaceColor is a single race-to-color mapping.
type RaceColor struct {
	RaceName string `json:"race_name"`
	Color    string `json:"color"`
}

// GetRaceColorMap returns a map of race_name -> color from the database.
// This is used by other handlers to enrich character/companion/dashboard responses.
func GetRaceColorMap() map[string]string {
	out := make(map[string]string)
	rows, err := db.DB.Query("SELECT race_name, color FROM race_colors")
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var name, color string
		rows.Scan(&name, &color)
		out[name] = color
	}
	return out
}

// ListRaceColors returns all race color mappings as a flat array.
func ListRaceColors(c *gin.Context) {
	rows, err := db.DB.Query("SELECT race_name, color FROM race_colors ORDER BY race_name")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var out []RaceColor
	for rows.Next() {
		var rc RaceColor
		rows.Scan(&rc.RaceName, &rc.Color)
		out = append(out, rc)
	}
	c.JSON(http.StatusOK, out)
}

// UpdateRaceColors replaces all race color mappings (bulk PUT).
// Expects a JSON array of {race_name, color} objects.
func UpdateRaceColors(c *gin.Context) {
	var colors []RaceColor
	if err := c.ShouldBindJSON(&colors); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Clear existing and re-insert
	if _, err := tx.Exec("DELETE FROM race_colors"); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	stmt, err := tx.Prepare("INSERT INTO race_colors(race_name, color) VALUES(?, ?)")
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer stmt.Close()

	for _, rc := range colors {
		if rc.RaceName == "" {
			continue
		}
		color := rc.Color
		if color == "" {
			color = "#6c757d"
		}
		if _, err := stmt.Exec(rc.RaceName, color); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
