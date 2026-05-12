package handlers

import (
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
)

type HPCalcResult struct {
	HPMax     int `json:"hp_max"`
	HPCurrent int `json:"hp_current"`
	Breakdown []HPBreakdown `json:"breakdown"`
}

type HPBreakdown struct {
	Source string `json:"source"`
	Level  int    `json:"level"`
	Die    string `json:"die"`
	Roll   int    `json:"roll"`
	ConMod int    `json:"con_mod"`
	Total  int    `json:"total"`
}

var hitDieValues = map[string]int{
	"d6": 6, "d8": 8, "d10": 10, "d12": 12,
}

func CalculateHP(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var con int
	err := db.DB.QueryRow("SELECT con FROM characters WHERE id=?", charID).Scan(&con)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}
	conMod := int(math.Floor(float64(con-10) / 2.0))

	// Load character classes
	classRows, err := db.DB.Query("SELECT class, level, hit_dice FROM character_classes WHERE character_id=? ORDER BY created_at", charID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer classRows.Close()

	var classes []struct {
		class   string
		level   int
		hitDice string
	}
	totalLevel := 0
	for classRows.Next() {
		var cl struct {
			class   string
			level   int
			hitDice string
		}
		classRows.Scan(&cl.class, &cl.level, &cl.hitDice)
		classes = append(classes, cl)
		totalLevel += cl.level
	}

	// If no multi-class entries, use the character's main class
	if len(classes) == 0 {
		var mainClass, hitDice string
		var level int
		db.DB.QueryRow("SELECT class, level, hit_dice FROM characters WHERE id=?", charID).Scan(&mainClass, &level, &hitDice)
		if hitDice == "" {
			hitDice = "d10"
		}
		classes = append(classes, struct {
			class   string
			level   int
			hitDice string
		}{mainClass, level, hitDice})
		totalLevel = level
	}

	breakdown := []HPBreakdown{}
	totalHP := 0

	for _, cl := range classes {
		dieVal := hitDieValues[cl.hitDice]
		if dieVal == 0 {
			dieVal = 8 // default d8
		}
		// First level gets max HP
		firstLevelHP := dieVal + max(0, conMod)
		totalHP += firstLevelHP
		breakdown = append(breakdown, HPBreakdown{
			Source: cl.class + " (1st)", Level: 1, Die: cl.hitDice,
			Roll: dieVal, ConMod: conMod, Total: firstLevelHP,
		})
		// Remaining levels get average (die/2 + 1) + CON
		for lvl := 2; lvl <= cl.level; lvl++ {
			avgHP := (dieVal/2 + 1) + max(0, conMod)
			totalHP += avgHP
			breakdown = append(breakdown, HPBreakdown{
				Source: cl.class + fmt.Sprintf(" (Lv%d)", lvl), Level: lvl, Die: cl.hitDice,
				Roll: dieVal/2 + 1, ConMod: conMod, Total: avgHP,
			})
		}
	}

	// Update the character's HP
	db.DB.Exec("UPDATE characters SET hp_max=?, hp_current=? WHERE id=? AND hp_auto_calc=1", totalHP, totalHP, charID)

	c.JSON(http.StatusOK, HPCalcResult{
		HPMax:     totalHP,
		HPCurrent: totalHP,
		Breakdown: breakdown,
	})
}
