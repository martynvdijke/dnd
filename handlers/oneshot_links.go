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

// linkOneShotEncounterRow inserts an encounter link, skipping duplicates for the
// same (adventure, act, encounter) triple. actID nil means an adventure-level link.
func linkOneShotEncounterRow(adventureID, encounterID int64, actID *int64) error {
	var actArg any
	if actID != nil && *actID > 0 {
		actArg = *actID
	}
	_, err := db.DB.Exec(`INSERT INTO oneshot_adventure_encounters(adventure_id, act_id, encounter_id)
		SELECT ?,?,? WHERE NOT EXISTS (
			SELECT 1 FROM oneshot_adventure_encounters WHERE adventure_id=? AND encounter_id=? AND act_id IS ?
		)`, adventureID, actArg, encounterID, adventureID, encounterID, actArg)
	return err
}

func GetOneShotEncounters(c *gin.Context) {
	adventureID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	c.JSON(http.StatusOK, loadAdventureEncounters(adventureID))
}

func LinkOneShotEncounter(c *gin.Context) {
	adventureID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var link struct {
		EncounterID int64  `json:"encounter_id"`
		ActID       *int64 `json:"act_id"`
	}
	if err := c.ShouldBindJSON(&link); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := linkOneShotEncounterRow(adventureID, link.EncounterID, link.ActID); err != nil {
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

// Act-scoped links

func ListActEncounters(c *gin.Context) {
	actID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	c.JSON(http.StatusOK, loadActEncounters(actID))
}

func LinkActEncounter(c *gin.Context) {
	actID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var link struct {
		EncounterID int64 `json:"encounter_id"`
	}
	if err := c.ShouldBindJSON(&link); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var adventureID int64
	if err := db.DB.QueryRow("SELECT adventure_id FROM oneshot_acts WHERE id=?", actID).Scan(&adventureID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "act not found"})
		return
	}
	if err := linkOneShotEncounterRow(adventureID, link.EncounterID, &actID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func UnlinkActEncounter(c *gin.Context) {
	actID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	encounterID, _ := strconv.ParseInt(c.Param("eid"), 10, 64)
	db.DB.Exec("DELETE FROM oneshot_adventure_encounters WHERE act_id=? AND encounter_id=?", actID, encounterID)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ─── Template Generation ───
