package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

func ListPregens(c *gin.Context) {
	userID, _ := c.Get("user_id")
	rows, err := db.DB.Query("SELECT id, user_id, name, race, class, subclass, level, background, alignment, str, dex, con, int, wis, cha, hp, ac, speed, skills, equipment, spells, features, personality, backstory, portrait_url, notes, created_at, updated_at FROM pregen_characters WHERE user_id=? ORDER BY updated_at DESC", userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	chars := make([]models.PregeneratedCharacter, 0)
	for rows.Next() {
		var ch models.PregeneratedCharacter
		if err := rows.Scan(&ch.ID, &ch.UserID, &ch.Name, &ch.Race, &ch.Class, &ch.Subclass, &ch.Level, &ch.Background, &ch.Alignment,
			&ch.Str, &ch.Dex, &ch.Con, &ch.Int, &ch.Wis, &ch.Cha, &ch.HP, &ch.AC, &ch.Speed,
			&ch.Skills, &ch.Equipment, &ch.Spells, &ch.Features, &ch.Personality, &ch.Backstory,
			&ch.PortraitURL, &ch.Notes, &ch.CreatedAt, &ch.UpdatedAt); err == nil {
			chars = append(chars, ch)
		}
	}
	c.JSON(http.StatusOK, chars)
}

func CreatePregen(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var input struct {
		Name        string `json:"name"`
		Race        string `json:"race"`
		Class       string `json:"class"`
		Subclass    string `json:"subclass"`
		Level       int    `json:"level"`
		Background  string `json:"background"`
		Alignment   string `json:"alignment"`
		Str         int    `json:"str"`
		Dex         int    `json:"dex"`
		Con         int    `json:"con"`
		Int         int    `json:"int"`
		Wis         int    `json:"wis"`
		Cha         int    `json:"cha"`
		HP          int    `json:"hp"`
		AC          int    `json:"ac"`
		Speed       int    `json:"speed"`
		Skills      string `json:"skills"`
		Equipment   string `json:"equipment"`
		Spells      string `json:"spells"`
		Features    string `json:"features"`
		Personality string `json:"personality"`
		Backstory   string `json:"backstory"`
		PortraitURL string `json:"portrait_url"`
		Notes       string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if input.Level == 0 {
		input.Level = 1
	}
	if input.Speed == 0 {
		input.Speed = 30
	}

	result, err := db.DB.Exec(`
		INSERT INTO pregen_characters(user_id, name, race, class, subclass, level, background, alignment,
			str, dex, con, int, wis, cha, hp, ac, speed, skills, equipment, spells, features, personality, backstory, portrait_url, notes)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		userID, input.Name, input.Race, input.Class, input.Subclass, input.Level, input.Background, input.Alignment,
		input.Str, input.Dex, input.Con, input.Int, input.Wis, input.Cha, input.HP, input.AC, input.Speed,
		input.Skills, input.Equipment, input.Spells, input.Features, input.Personality, input.Backstory, input.PortraitURL, input.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func GetPregen(c *gin.Context) {
	id := c.Param("id")
	var ch models.PregeneratedCharacter
	err := db.DB.QueryRow("SELECT id, user_id, name, race, class, subclass, level, background, alignment, str, dex, con, int, wis, cha, hp, ac, speed, skills, equipment, spells, features, personality, backstory, portrait_url, notes, created_at, updated_at FROM pregen_characters WHERE id=?", id).
		Scan(&ch.ID, &ch.UserID, &ch.Name, &ch.Race, &ch.Class, &ch.Subclass, &ch.Level, &ch.Background, &ch.Alignment,
			&ch.Str, &ch.Dex, &ch.Con, &ch.Int, &ch.Wis, &ch.Cha, &ch.HP, &ch.AC, &ch.Speed,
			&ch.Skills, &ch.Equipment, &ch.Spells, &ch.Features, &ch.Personality, &ch.Backstory,
			&ch.PortraitURL, &ch.Notes, &ch.CreatedAt, &ch.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "pregen not found"})
		return
	}
	c.JSON(http.StatusOK, ch)
}

func UpdatePregen(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		Name        string `json:"name"`
		Race        string `json:"race"`
		Class       string `json:"class"`
		Subclass    string `json:"subclass"`
		Level       int    `json:"level"`
		Background  string `json:"background"`
		Alignment   string `json:"alignment"`
		Str         int    `json:"str"`
		Dex         int    `json:"dex"`
		Con         int    `json:"con"`
		Int         int    `json:"int"`
		Wis         int    `json:"wis"`
		Cha         int    `json:"cha"`
		HP          int    `json:"hp"`
		AC          int    `json:"ac"`
		Speed       int    `json:"speed"`
		Skills      string `json:"skills"`
		Equipment   string `json:"equipment"`
		Spells      string `json:"spells"`
		Features    string `json:"features"`
		Personality string `json:"personality"`
		Backstory   string `json:"backstory"`
		PortraitURL string `json:"portrait_url"`
		Notes       string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := db.DB.Exec(`
		UPDATE pregen_characters SET name=?, race=?, class=?, subclass=?, level=?, background=?, alignment=?,
			str=?, dex=?, con=?, int=?, wis=?, cha=?, hp=?, ac=?, speed=?, skills=?, equipment=?, spells=?, features=?,
			personality=?, backstory=?, portrait_url=?, notes=?, updated_at=datetime('now') WHERE id=?`,
		input.Name, input.Race, input.Class, input.Subclass, input.Level, input.Background, input.Alignment,
		input.Str, input.Dex, input.Con, input.Int, input.Wis, input.Cha, input.HP, input.AC, input.Speed,
		input.Skills, input.Equipment, input.Spells, input.Features, input.Personality, input.Backstory, input.PortraitURL, input.Notes, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func DeletePregen(c *gin.Context) {
	id := c.Param("id")
	db.DB.Exec("DELETE FROM pregen_characters WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// Class role assignments for party balance
var classRoles = map[string][]string{
	"barbarian": {"tank"},
	"fighter":   {"tank", "damage"},
	"paladin":   {"tank", "healer"},
	"cleric":    {"healer"},
	"druid":     {"healer", "support"},
	"wizard":    {"damage", "support"},
	"sorcerer":  {"damage"},
	"warlock":   {"damage"},
	"bard":      {"support", "healer"},
	"rogue":     {"damage", "skill"},
	"monk":      {"damage"},
	"ranger":    {"damage", "skill"},
	"artificer": {"support"},
}

var allRoles = []string{"tank", "healer", "damage", "support", "skill"}
