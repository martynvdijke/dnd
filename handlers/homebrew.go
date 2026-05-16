package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

var homebrewSources = []string{"homebrew", "srd"}
var homebrewSystems = []string{"dnd5e", "pf2e", "generic"}

func isHomebrewType(t string) bool {
	switch t {
	case "races", "classes", "spells", "feats", "backgrounds", "equipment":
		return true
	}
	return false
}

func homebrewTable(t string) string {
	return "compendium_" + t
}

type HomebrewRace struct {
	models.CompendiumRace
	IsHomebrew bool `json:"is_homebrew"`
}

func ListHomebrewContent(c *gin.Context) {
	contentType := c.Param("type")
	if !isHomebrewType(contentType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid type"})
		return
	}

	table := homebrewTable(contentType)
	rows, err := db.DB.Query("SELECT * FROM "+table+" WHERE source='homebrew' ORDER BY name")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var out []map[string]any
	cols, _ := rows.Columns()
	for rows.Next() {
		vals := make([]any, len(cols))
		valPtrs := make([]any, len(cols))
		for i := range cols {
			valPtrs[i] = &vals[i]
		}
		rows.Scan(valPtrs...)
		row := make(map[string]any)
		for i, col := range cols {
			if vals[i] != nil {
				row[col] = vals[i]
			}
		}
		row["is_homebrew"] = true
		out = append(out, row)
	}
	c.JSON(http.StatusOK, out)
}

func CreateHomebrewContent(c *gin.Context) {
	contentType := c.Param("type")
	if !isHomebrewType(contentType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid type"})
		return
	}
	table := homebrewTable(contentType)

	var data map[string]any
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	data["source"] = "homebrew"
	if _, ok := data["system"]; !ok {
		data["system"] = "dnd5e"
	}

	// Build dynamic INSERT
	cols := make([]string, 0, len(data))
	vals := make([]any, 0, len(data))
	placeholders := make([]string, 0, len(data))
	for k, v := range data {
		cols = append(cols, k)
		vals = append(vals, v)
		placeholders = append(placeholders, "?")
	}

	query := "INSERT INTO " + table + "("
	for i, col := range cols {
		if i > 0 {
			query += ","
		}
		query += col
	}
	query += ") VALUES("
	for i := range placeholders {
		if i > 0 {
			query += ","
		}
		query += "?"
	}
	query += ")"

	result, err := db.DB.Exec(query, vals...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func UpdateHomebrewContent(c *gin.Context) {
	contentType := c.Param("type")
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if !isHomebrewType(contentType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid type"})
		return
	}

	table := homebrewTable(contentType)
	var data map[string]any
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build dynamic UPDATE
	cols := make([]string, 0, len(data))
	vals := make([]any, 0, len(data))
	for k, v := range data {
		cols = append(cols, k+"=?")
		vals = append(vals, v)
	}
	vals = append(vals, id)

	query := "UPDATE " + table + " SET "
	for i, col := range cols {
		if i > 0 {
			query += ", "
		}
		query += col
	}
	query += " WHERE id=?"

	_, err := db.DB.Exec(query, vals...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteHomebrewContent(c *gin.Context) {
	contentType := c.Param("type")
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if !isHomebrewType(contentType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid type"})
		return
	}

	table := homebrewTable(contentType)
	_, err := db.DB.Exec("DELETE FROM "+table+" WHERE id=? AND source='homebrew'", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
