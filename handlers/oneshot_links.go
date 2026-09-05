package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

// ─── NPC Links ───

func GetOneShotNPCs(c *gin.Context) {
	adventureID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	rows, err := db.DB.Query("SELECT oan.id, oan.adventure_id, oan.npc_id, oan.role, oan.story_hook, oan.combat_ready, COALESCE(n.name,'') FROM oneshot_adventure_npcs oan LEFT JOIN npcs n ON oan.npc_id=n.id WHERE oan.adventure_id=?", adventureID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := make([]models.OneShotAdventureNPC, 0)
	for rows.Next() {
		var npc models.OneShotAdventureNPC
		var combatReady int
		rows.Scan(&npc.ID, &npc.AdventureID, &npc.NPCID, &npc.Role, &npc.StoryHook, &combatReady, &npc.NPCName)
		npc.CombatReady = combatReady == 1
		out = append(out, npc)
	}
	c.JSON(http.StatusOK, out)
}

func LinkOneShotNPC(c *gin.Context) {
	adventureID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var link struct {
		NPCID       int64  `json:"npc_id"`
		Role        string `json:"role"`
		StoryHook   string `json:"story_hook"`
		CombatReady bool   `json:"combat_ready"`
	}
	if err := c.ShouldBindJSON(&link); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	combatReady := 0
	if link.CombatReady {
		combatReady = 1
	}
	_, err := db.DB.Exec("INSERT OR REPLACE INTO oneshot_adventure_npcs(adventure_id, npc_id, role, story_hook, combat_ready) VALUES(?,?,?,?,?)",
		adventureID, link.NPCID, link.Role, link.StoryHook, combatReady)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func UnlinkOneShotNPC(c *gin.Context) {
	adventureID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	npcID, _ := strconv.ParseInt(c.Param("nid"), 10, 64)
	db.DB.Exec("DELETE FROM oneshot_adventure_npcs WHERE adventure_id=? AND npc_id=?", adventureID, npcID)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ─── Location Links ───

func GetOneShotLocations(c *gin.Context) {
	adventureID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	rows, err := db.DB.Query("SELECT oal.id, oal.adventure_id, oal.location_id, COALESCE(l.name,'') FROM oneshot_adventure_locations oal LEFT JOIN locations l ON oal.location_id=l.id WHERE oal.adventure_id=?", adventureID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := make([]models.OneShotAdventureLocation, 0)
	for rows.Next() {
		var loc models.OneShotAdventureLocation
		rows.Scan(&loc.ID, &loc.AdventureID, &loc.LocationID, &loc.LocationName)
		out = append(out, loc)
	}
	c.JSON(http.StatusOK, out)
}

func LinkOneShotLocation(c *gin.Context) {
	adventureID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var link struct {
		LocationID int64 `json:"location_id"`
	}
	if err := c.ShouldBindJSON(&link); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := db.DB.Exec("INSERT OR IGNORE INTO oneshot_adventure_locations(adventure_id, location_id) VALUES(?,?)", adventureID, link.LocationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func UnlinkOneShotLocation(c *gin.Context) {
	adventureID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	locationID, _ := strconv.ParseInt(c.Param("lid"), 10, 64)
	db.DB.Exec("DELETE FROM oneshot_adventure_locations WHERE adventure_id=? AND location_id=?", adventureID, locationID)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ─── Encounter Links ───

func GetOneShotEncounters(c *gin.Context) {
	adventureID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	rows, err := db.DB.Query("SELECT oae.id, oae.adventure_id, oae.encounter_id, COALESCE(e.name,'') FROM oneshot_adventure_encounters oae LEFT JOIN encounter_templates e ON oae.encounter_id=e.id WHERE oae.adventure_id=?", adventureID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := make([]models.OneShotAdventureEncounter, 0)
	for rows.Next() {
		var enc models.OneShotAdventureEncounter
		rows.Scan(&enc.ID, &enc.AdventureID, &enc.EncounterID, &enc.EncounterName)
		out = append(out, enc)
	}
	c.JSON(http.StatusOK, out)
}

func LinkOneShotEncounter(c *gin.Context) {
	adventureID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var link struct {
		EncounterID int64 `json:"encounter_id"`
	}
	if err := c.ShouldBindJSON(&link); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := db.DB.Exec("INSERT OR IGNORE INTO oneshot_adventure_encounters(adventure_id, encounter_id) VALUES(?,?)", adventureID, link.EncounterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func UnlinkOneShotEncounter(c *gin.Context) {
	adventureID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	encounterID, _ := strconv.ParseInt(c.Param("eid"), 10, 64)
	db.DB.Exec("DELETE FROM oneshot_adventure_encounters WHERE adventure_id=? AND encounter_id=?", adventureID, encounterID)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ─── Template Generation ───
