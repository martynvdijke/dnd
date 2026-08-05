package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"villum/db"
)

func CompareCharacters(c *gin.Context) {
	ids := c.Query("ids")
	if ids == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ids query param required (comma-separated)"})
		return
	}

	idStrs := strings.Split(ids, ",")
	if len(idStrs) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provide at least 2 character IDs"})
		return
	}

	type CharSummary struct {
		ID         int64  `json:"id"`
		Name       string `json:"name"`
		Race       string `json:"race"`
		Class      string `json:"class"`
		Level      int    `json:"level"`
		HPMax      int    `json:"hp_max"`
		HPCurrent  int    `json:"hp_current"`
		AC         int    `json:"ac"`
		Str        int    `json:"str"`
		Dex        int    `json:"dex"`
		Con        int    `json:"con"`
		Int        int    `json:"int"`
		Wis        int    `json:"wis"`
		Cha        int    `json:"cha"`
		XP         int    `json:"xp"`
		Background string `json:"background"`
		Alignment  string `json:"alignment"`
		Speed      int    `json:"speed"`
		Initiative int    `json:"initiative"`
	}

	result := make([]CharSummary, 0)
	for _, idStr := range idStrs {
		id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64)
		if err != nil {
			continue
		}
		var s CharSummary
		err = db.DB.QueryRow(`
			SELECT id, name, race, class, level, hp_max, hp_current, ac, str, dex, con, int, wis, cha, xp, background, alignment, speed, initiative
			FROM characters WHERE id=?`, id).Scan(
			&s.ID, &s.Name, &s.Race, &s.Class, &s.Level, &s.HPMax, &s.HPCurrent,
			&s.AC, &s.Str, &s.Dex, &s.Con, &s.Int, &s.Wis, &s.Cha,
			&s.XP, &s.Background, &s.Alignment, &s.Speed, &s.Initiative)
		if err == nil {
			result = append(result, s)
		}
	}

	if len(result) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not find enough characters"})
		return
	}
	c.JSON(http.StatusOK, result)
}
