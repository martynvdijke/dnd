package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

func ListCompanions(c *gin.Context) {
	charID := c.Query("character_id")
	if charID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "character_id required"})
		return
	}
	rows, err := db.DB.Query("SELECT id, character_id, name, type, race, hp_max, hp_current, ac, str, dex, con, int, wis, cha, speed, abilities, notes, portrait_url, is_alive FROM companions WHERE character_id=? ORDER BY name", charID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var out = make([]models.Companion, 0)
	for rows.Next() {
		var comp models.Companion
		var isAlive int
		rows.Scan(&comp.ID, &comp.CharacterID, &comp.Name, &comp.Type, &comp.Race, &comp.HPMax, &comp.HPCurrent, &comp.AC,
			&comp.Str, &comp.Dex, &comp.Con, &comp.Int, &comp.Wis, &comp.Cha, &comp.Speed,
			&comp.Abilities, &comp.Notes, &comp.PortraitURL, &isAlive)
		comp.IsAlive = isAlive == 1
		out = append(out, comp)
	}
	c.JSON(http.StatusOK, out)
}

func CreateCompanion(c *gin.Context) {
	var comp models.Companion
	if err := c.ShouldBindJSON(&comp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if comp.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name required"})
		return
	}
	isAlive := 0
	if comp.IsAlive || comp.ID == 0 {
		isAlive = 1
	}
	result, err := db.DB.Exec(`INSERT INTO companions(character_id,name,type,race,hp_max,hp_current,ac,str,dex,con,int,wis,cha,speed,abilities,notes,portrait_url,is_alive) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		comp.CharacterID, comp.Name, comp.Type, comp.Race, comp.HPMax, comp.HPCurrent, comp.AC,
		comp.Str, comp.Dex, comp.Con, comp.Int, comp.Wis, comp.Cha, comp.Speed,
		comp.Abilities, comp.Notes, comp.PortraitURL, isAlive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func UpdateCompanion(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var comp models.Companion
	if err := c.ShouldBindJSON(&comp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	isAlive := 0
	if comp.IsAlive {
		isAlive = 1
	}
	db.DB.Exec(`UPDATE companions SET name=?, type=?, race=?, hp_max=?, hp_current=?, ac=?, str=?, dex=?, con=?, int=?, wis=?, cha=?, speed=?, abilities=?, notes=?, portrait_url=?, is_alive=? WHERE id=?`,
		comp.Name, comp.Type, comp.Race, comp.HPMax, comp.HPCurrent, comp.AC,
		comp.Str, comp.Dex, comp.Con, comp.Int, comp.Wis, comp.Cha, comp.Speed,
		comp.Abilities, comp.Notes, comp.PortraitURL, isAlive, id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteCompanion(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	db.DB.Exec("DELETE FROM companions WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
