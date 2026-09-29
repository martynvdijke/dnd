package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

// ─── Pregenerated Characters (HTMX) ───

func pregenFormInt(c *gin.Context, key string, fallback int) int {
	v, err := strconv.Atoi(c.PostForm(key))
	if err != nil {
		return fallback
	}
	return v
}

// pregenFromForm builds a pregen from form-encoded fields, mirroring the JSON
// CreatePregen/UpdatePregen input.
func pregenFromForm(c *gin.Context) models.PregeneratedCharacter {
	return models.PregeneratedCharacter{
		Name:        c.PostForm("name"),
		Race:        c.PostForm("race"),
		Class:       c.PostForm("class"),
		Subclass:    c.PostForm("subclass"),
		Level:       pregenFormInt(c, "level", 1),
		Background:  c.PostForm("background"),
		Alignment:   c.PostForm("alignment"),
		Str:         pregenFormInt(c, "str", 10),
		Dex:         pregenFormInt(c, "dex", 10),
		Con:         pregenFormInt(c, "con", 10),
		Int:         pregenFormInt(c, "int", 10),
		Wis:         pregenFormInt(c, "wis", 10),
		Cha:         pregenFormInt(c, "cha", 10),
		HP:          pregenFormInt(c, "hp", 1),
		AC:          pregenFormInt(c, "ac", 10),
		Speed:       pregenFormInt(c, "speed", 30),
		Skills:      c.PostForm("skills"),
		Equipment:   c.PostForm("equipment"),
		Spells:      c.PostForm("spells"),
		Features:    c.PostForm("features"),
		Personality: c.PostForm("personality"),
		Backstory:   c.PostForm("backstory"),
		PortraitURL: c.PostForm("portrait_url"),
		Notes:       c.PostForm("notes"),
	}
}

func loadPregenForUser(id any, userID any) (*models.PregeneratedCharacter, error) {
	var ch models.PregeneratedCharacter
	err := db.DB.QueryRow(`SELECT id, user_id, name, race, class, subclass, level, background, alignment,
			str, dex, con, int, wis, cha, hp, ac, speed, skills, equipment, spells, features, personality,
			backstory, portrait_url, notes, created_at, updated_at
		FROM pregen_characters WHERE id=? AND user_id=?`, id, userID).
		Scan(&ch.ID, &ch.UserID, &ch.Name, &ch.Race, &ch.Class, &ch.Subclass, &ch.Level, &ch.Background, &ch.Alignment,
			&ch.Str, &ch.Dex, &ch.Con, &ch.Int, &ch.Wis, &ch.Cha, &ch.HP, &ch.AC, &ch.Speed,
			&ch.Skills, &ch.Equipment, &ch.Spells, &ch.Features, &ch.Personality, &ch.Backstory,
			&ch.PortraitURL, &ch.Notes, &ch.CreatedAt, &ch.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &ch, nil
}

func HtmxNewPregenForm(c *gin.Context) {
	renderTemplate(c, "oneshot_pregen_form.html", gin.H{"Pregen": nil})
}

func HtmxEditPregenForm(c *gin.Context) {
	userID, _ := c.Get("user_id")
	ch, err := loadPregenForUser(c.Param("id"), userID)
	if err != nil {
		c.String(http.StatusNotFound, "Pregen not found")
		return
	}
	renderTemplate(c, "oneshot_pregen_form.html", gin.H{"Pregen": ch})
}

func HtmxCreatePregen(c *gin.Context) {
	userID, _ := c.Get("user_id")
	ch := pregenFromForm(c)

	_, err := db.DB.Exec(`
		INSERT INTO pregen_characters(user_id, name, race, class, subclass, level, background, alignment,
			str, dex, con, int, wis, cha, hp, ac, speed, skills, equipment, spells, features, personality, backstory, portrait_url, notes)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		userID, ch.Name, ch.Race, ch.Class, ch.Subclass, ch.Level, ch.Background, ch.Alignment,
		ch.Str, ch.Dex, ch.Con, ch.Int, ch.Wis, ch.Cha, ch.HP, ch.AC, ch.Speed,
		ch.Skills, ch.Equipment, ch.Spells, ch.Features, ch.Personality, ch.Backstory, ch.PortraitURL, ch.Notes)
	if err != nil {
		c.String(http.StatusInternalServerError, "insert error: %v", err)
		return
	}

	HtmxListPregens(c)
}

func HtmxUpdatePregen(c *gin.Context) {
	userID, _ := c.Get("user_id")
	ch := pregenFromForm(c)

	_, err := db.DB.Exec(`
		UPDATE pregen_characters SET name=?, race=?, class=?, subclass=?, level=?, background=?, alignment=?,
			str=?, dex=?, con=?, int=?, wis=?, cha=?, hp=?, ac=?, speed=?, skills=?, equipment=?, spells=?, features=?,
			personality=?, backstory=?, portrait_url=?, notes=?, updated_at=datetime('now') WHERE id=? AND user_id=?`,
		ch.Name, ch.Race, ch.Class, ch.Subclass, ch.Level, ch.Background, ch.Alignment,
		ch.Str, ch.Dex, ch.Con, ch.Int, ch.Wis, ch.Cha, ch.HP, ch.AC, ch.Speed,
		ch.Skills, ch.Equipment, ch.Spells, ch.Features, ch.Personality, ch.Backstory, ch.PortraitURL, ch.Notes,
		c.Param("id"), userID)
	if err != nil {
		c.String(http.StatusInternalServerError, "update error: %v", err)
		return
	}

	HtmxListPregens(c)
}

func HtmxDeletePregen(c *gin.Context) {
	userID, _ := c.Get("user_id")
	_, err := db.DB.Exec("DELETE FROM pregen_characters WHERE id=? AND user_id=?", c.Param("id"), userID)
	if err != nil && err != sql.ErrNoRows {
		c.String(http.StatusInternalServerError, "delete error: %v", err)
		return
	}
	HtmxListPregens(c)
}
