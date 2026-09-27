package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
)

// HandleDeathSave rolls a death saving throw for a character at 0 HP.
func HandleDeathSave(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if !canEditCharacterID(c, charID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}
	var hpCurrent, dsSucc, dsFail int
	err := db.DB.QueryRow("SELECT hp_current, death_saves_successes, death_saves_failures FROM characters WHERE id=?", charID).
		Scan(&hpCurrent, &dsSucc, &dsFail)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}
	if hpCurrent > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "character is not dying"})
		return
	}

	roll, _, err := rollD20Advantage("")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	revived := false
	text := ""
	switch {
	case roll == 20:
		hpCurrent = 1
		dsSucc, dsFail = 0, 0
		revived = true
		text = "Natural 20 — regain 1 hit point and recover consciousness"
	case roll == 1:
		dsFail += 2
		text = "Natural 1 — two failures"
	case roll >= 10:
		dsSucc++
		text = "Success"
	default:
		dsFail++
		text = "Failure"
	}
	if dsSucc > 3 {
		dsSucc = 3
	}
	if dsFail > 3 {
		dsFail = 3
	}

	stable := !revived && hpCurrent == 0 && dsSucc >= 3
	dead := !revived && hpCurrent == 0 && dsFail >= 3
	if dead {
		db.DB.Exec("UPDATE characters SET concentrating_on='' WHERE id=?", charID)
		db.DB.Exec("DELETE FROM character_conditions WHERE character_id=? AND type='concentration'", charID)
	}
	if _, err = db.DB.Exec("UPDATE characters SET hp_current=?, death_saves_successes=?, death_saves_failures=? WHERE id=?", hpCurrent, dsSucc, dsFail, charID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	SendCharacterUpdate(charID)
	c.JSON(http.StatusOK, gin.H{
		"roll":       roll,
		"successes":  dsSucc,
		"failures":   dsFail,
		"stable":     stable,
		"dead":       dead,
		"hp_current": hpCurrent,
		"revived":    revived,
		"text":       text,
	})
}
