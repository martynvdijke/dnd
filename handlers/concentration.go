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

	dc := 10
	if req.Damage >= 22 {
		dc = 15
	} else if req.Damage >= 12 {
		dc = 12
	} else if req.Damage >= 8 {
		dc = 11
	} else if req.Damage >= 4 {
		dc = 10
	}
	// Half the damage taken rounded down, minimum 10 (per D&D 5e rules: DC = 10 or half damage, whichever is higher)
	halfDmg := req.Damage / 2
	if halfDmg > dc {
		dc = halfDmg
	}
	if dc < 10 {
		dc = 10
	}

	c.JSON(http.StatusOK, ConcentrationCheckResult{
		NeedsCheck: true,
		DC:         dc,
		SpellName:  concentratingOn,
		Damage:     req.Damage,
	})
}
