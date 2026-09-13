package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
)

type ConcentrationCheckResult struct {
	NeedsCheck bool   `json:"needs_check"`
	DC         int    `json:"dc"`
	SpellName  string `json:"spell_name"`
	Damage     int    `json:"damage"`
}

func CheckConcentration(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var concentratingOn string
	var con int
	err := db.DB.QueryRow("SELECT concentrating_on, con FROM characters WHERE id=?", charID).Scan(&concentratingOn, &con)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}
	if !canEditCharacterID(c, charID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	if concentratingOn == "" {
		c.JSON(http.StatusOK, ConcentrationCheckResult{NeedsCheck: false})
		return
	}

	var req struct {
		Damage int `json:"damage"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dc := calcConcentrationDC(req.Damage)

	c.JSON(http.StatusOK, ConcentrationCheckResult{
		NeedsCheck: true,
		DC:         dc,
		SpellName:  concentratingOn,
		Damage:     req.Damage,
	})
}
