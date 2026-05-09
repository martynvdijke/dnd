package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

// ─── Locations ───

func ListLocations(c *gin.Context) {
	userID, _ := c.Get("user_id")
	rows, err := db.DB.Query("SELECT id,user_id,name,type,description,parent_id,latitude,longitude,created_at FROM locations WHERE user_id=? ORDER BY name", userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var out []models.Location
	for rows.Next() {
		var l models.Location
		rows.Scan(&l.ID, &l.UserID, &l.Name, &l.Type, &l.Description, &l.ParentID, &l.Latitude, &l.Longitude, &l.CreatedAt)
		out = append(out, l)
	}
	c.JSON(http.StatusOK, out)
}

func CreateLocation(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var l models.Location
	if err := c.ShouldBindJSON(&l); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := db.DB.Exec("INSERT INTO locations(user_id,name,type,description,parent_id,latitude,longitude) VALUES(?,?,?,?,?,?,?)",
		userID, l.Name, l.Type, l.Description, l.ParentID, l.Latitude, l.Longitude)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func UpdateLocation(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	var ownerID int64
	err := db.DB.QueryRow("SELECT user_id FROM locations WHERE id=?", id).Scan(&ownerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "location not found"})
		return
	}
	if role != "admin" && ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	var l models.Location
	if err := c.ShouldBindJSON(&l); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Exec("UPDATE locations SET name=?,type=?,description=?,parent_id=?,latitude=?,longitude=? WHERE id=?",
		l.Name, l.Type, l.Description, l.ParentID, l.Latitude, l.Longitude, id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteLocation(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	var ownerID int64
	err := db.DB.QueryRow("SELECT user_id FROM locations WHERE id=?", id).Scan(&ownerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "location not found"})
		return
	}
	if role != "admin" && ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	db.DB.Exec("DELETE FROM locations WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func LinkLocation(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var cl models.CharacterLocation
	if err := c.ShouldBindJSON(&cl); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := db.DB.Exec("INSERT OR REPLACE INTO character_locations(character_id,location_id,relationship,notes) VALUES(?,?,?,?)",
		charID, cl.LocationID, cl.Relationship, cl.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func UnlinkLocation(c *gin.Context) {
	linkID, _ := strconv.ParseInt(c.Param("lid"), 10, 64)
	var charID int64
	err := db.DB.QueryRow("SELECT character_id FROM character_locations WHERE id=?", linkID).Scan(&charID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
		return
	}
	if !checkCharacterAccess(c, charID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	db.DB.Exec("DELETE FROM character_locations WHERE id=?", linkID)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func GetCharacterLocations(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	rows, err := db.DB.Query(`
		SELECT cl.id, cl.character_id, cl.location_id, cl.relationship, cl.notes,
			l.name, l.type, l.description
		FROM character_locations cl JOIN locations l ON cl.location_id = l.id
		WHERE cl.character_id=? ORDER BY cl.relationship`, charID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	type LocLink struct {
		models.CharacterLocation
		LocationName string `json:"location_name"`
		LocationType string `json:"location_type"`
		Description  string `json:"description"`
	}
	var out []LocLink
	for rows.Next() {
		var ll LocLink
		rows.Scan(&ll.ID, &ll.CharacterID, &ll.LocationID, &ll.Relationship, &ll.Notes,
			&ll.LocationName, &ll.LocationType, &ll.Description)
		out = append(out, ll)
	}
	c.JSON(http.StatusOK, out)
}

// ─── NPCs ───

func ListNPCs(c *gin.Context) {
	userID, _ := c.Get("user_id")
	rows, err := db.DB.Query("SELECT id,user_id,name,race,class,description,notes,str,dex,con,int,wis,cha,hp_max,hp_current,is_alive,created_at FROM npcs WHERE user_id=? ORDER BY name", userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var out []models.NPC
	for rows.Next() {
		var n models.NPC
		rows.Scan(&n.ID, &n.UserID, &n.Name, &n.Race, &n.Class, &n.Description, &n.Notes,
			&n.Str, &n.Dex, &n.Con, &n.Int, &n.Wis, &n.Cha, &n.HPMax, &n.HPCurrent, &n.IsAlive, &n.CreatedAt)
		out = append(out, n)
	}
	c.JSON(http.StatusOK, out)
}

func CreateNPC(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var n models.NPC
	if err := c.ShouldBindJSON(&n); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := db.DB.Exec("INSERT INTO npcs(user_id,name,race,class,description,notes,str,dex,con,int,wis,cha,hp_max,hp_current,is_alive) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",
		userID, n.Name, n.Race, n.Class, n.Description, n.Notes,
		n.Str, n.Dex, n.Con, n.Int, n.Wis, n.Cha, n.HPMax, n.HPCurrent, n.IsAlive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func UpdateNPC(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	var ownerID int64
	err := db.DB.QueryRow("SELECT user_id FROM npcs WHERE id=?", id).Scan(&ownerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "NPC not found"})
		return
	}
	if role != "admin" && ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	var n models.NPC
	if err := c.ShouldBindJSON(&n); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Exec("UPDATE npcs SET name=?,race=?,class=?,description=?,notes=?,str=?,dex=?,con=?,int=?,wis=?,cha=?,hp_max=?,hp_current=?,is_alive=? WHERE id=?",
		n.Name, n.Race, n.Class, n.Description, n.Notes,
		n.Str, n.Dex, n.Con, n.Int, n.Wis, n.Cha, n.HPMax, n.HPCurrent, n.IsAlive, id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteNPC(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	var ownerID int64
	err := db.DB.QueryRow("SELECT user_id FROM npcs WHERE id=?", id).Scan(&ownerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "NPC not found"})
		return
	}
	if role != "admin" && ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	db.DB.Exec("DELETE FROM npcs WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func LinkNPC(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var cn models.CharacterNPC
	if err := c.ShouldBindJSON(&cn); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := db.DB.Exec("INSERT OR REPLACE INTO character_npcs(character_id,npc_id,relationship,notes) VALUES(?,?,?,?)",
		charID, cn.NPCID, cn.Relationship, cn.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func UnlinkNPC(c *gin.Context) {
	linkID, _ := strconv.ParseInt(c.Param("nid"), 10, 64)
	var charID int64
	err := db.DB.QueryRow("SELECT character_id FROM character_npcs WHERE id=?", linkID).Scan(&charID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
		return
	}
	if !checkCharacterAccess(c, charID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	db.DB.Exec("DELETE FROM character_npcs WHERE id=?", linkID)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func GetCharacterNPCs(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	rows, err := db.DB.Query(`
		SELECT cn.id, cn.character_id, cn.npc_id, cn.relationship, cn.notes,
			cn.interaction_count, cn.last_interacted,
			n.name, n.race, n.class, n.hp_max, n.hp_current, n.is_alive
		FROM character_npcs cn JOIN npcs n ON cn.npc_id = n.id
		WHERE cn.character_id=? ORDER BY cn.interaction_count DESC`, charID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	type NPCLink struct {
		models.CharacterNPC
		NPCName         string `json:"npc_name"`
		NPCRace         string `json:"npc_race"`
		NPCClass        string `json:"npc_class"`
		NPHPMax         int    `json:"npc_hp_max"`
		NPHPCurr        int    `json:"npc_hp_current"`
		NPCAlive        bool   `json:"npc_is_alive"`
	}
	var out []NPCLink
	for rows.Next() {
		var nl NPCLink
		rows.Scan(&nl.ID, &nl.CharacterID, &nl.NPCID, &nl.Relationship, &nl.Notes,
			&nl.InteractionCount, &nl.LastInteracted,
			&nl.NPCName, &nl.NPCRace, &nl.NPCClass, &nl.NPHPMax, &nl.NPHPCurr, &nl.NPCAlive)
		out = append(out, nl)
	}
	c.JSON(http.StatusOK, out)
}

// ─── Sessions ───

func ListSessions(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	rows, err := db.DB.Query("SELECT id,character_id,session_date,title,notes,xp_earned,gold_earned,important_events,created_at FROM sessions WHERE character_id=? ORDER BY session_date DESC", charID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var out []models.Session
	for rows.Next() {
		var s models.Session
		rows.Scan(&s.ID, &s.CharacterID, &s.SessionDate, &s.Title, &s.Notes, &s.XPEarned, &s.GoldEarned, &s.ImportantEvents, &s.CreatedAt)
		out = append(out, s)
	}
	c.JSON(http.StatusOK, out)
}

func CreateSession(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var s models.Session
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := db.DB.Exec("INSERT INTO sessions(character_id,session_date,title,notes,xp_earned,gold_earned,important_events) VALUES(?,?,?,?,?,?,?)",
		charID, s.SessionDate, s.Title, s.Notes, s.XPEarned, s.GoldEarned, s.ImportantEvents)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func UpdateSession(c *gin.Context) {
	sid, _ := strconv.ParseInt(c.Param("sid"), 10, 64)
	var charID int64
	err := db.DB.QueryRow("SELECT character_id FROM sessions WHERE id=?", sid).Scan(&charID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	if !checkCharacterAccess(c, charID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	var s models.Session
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Exec("UPDATE sessions SET session_date=?,title=?,notes=?,xp_earned=?,gold_earned=?,important_events=? WHERE id=?",
		s.SessionDate, s.Title, s.Notes, s.XPEarned, s.GoldEarned, s.ImportantEvents, sid)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteSession(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("sid"), 10, 64)
	var charID int64
	err := db.DB.QueryRow("SELECT character_id FROM sessions WHERE id=?", id).Scan(&charID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	if !checkCharacterAccess(c, charID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	db.DB.Exec("DELETE FROM sessions WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ─── Quests ───

func ListQuests(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	rows, err := db.DB.Query("SELECT id,character_id,name,description,status,objectives,rewards,notes,created_at,updated_at FROM quests WHERE character_id=? ORDER BY status,name", charID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var out []models.Quest
	for rows.Next() {
		var q models.Quest
		rows.Scan(&q.ID, &q.CharacterID, &q.Name, &q.Description, &q.Status, &q.Objectives, &q.Rewards, &q.Notes, &q.CreatedAt, &q.UpdatedAt)
		out = append(out, q)
	}
	c.JSON(http.StatusOK, out)
}

func CreateQuest(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var q models.Quest
	if err := c.ShouldBindJSON(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if q.Status == "" {
		q.Status = "active"
	}
	result, err := db.DB.Exec("INSERT INTO quests(character_id,name,description,status,objectives,rewards,notes) VALUES(?,?,?,?,?,?,?)",
		charID, q.Name, q.Description, q.Status, q.Objectives, q.Rewards, q.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func UpdateQuest(c *gin.Context) {
	qid, _ := strconv.ParseInt(c.Param("qid"), 10, 64)
	var q models.Quest
	if err := c.ShouldBindJSON(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Exec("UPDATE quests SET name=?,description=?,status=?,objectives=?,rewards=?,notes=?,updated_at=datetime('now') WHERE id=?",
		q.Name, q.Description, q.Status, q.Objectives, q.Rewards, q.Notes, qid)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteQuest(c *gin.Context) {
	qid, _ := strconv.ParseInt(c.Param("qid"), 10, 64)
	var charID int64
	err := db.DB.QueryRow("SELECT character_id FROM quests WHERE id=?", qid).Scan(&charID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "quest not found"})
		return
	}
	if !checkCharacterAccess(c, charID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	db.DB.Exec("DELETE FROM quests WHERE id=?", qid)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ─── Journal ───

func ListJournal(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	rows, err := db.DB.Query("SELECT id,character_id,title,entry,entry_date,created_at FROM journal WHERE character_id=? ORDER BY entry_date DESC", charID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var out []models.JournalEntry
	for rows.Next() {
		var j models.JournalEntry
		rows.Scan(&j.ID, &j.CharacterID, &j.Title, &j.Entry, &j.EntryDate, &j.CreatedAt)
		out = append(out, j)
	}
	c.JSON(http.StatusOK, out)
}

func CreateJournalEntry(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var j models.JournalEntry
	if err := c.ShouldBindJSON(&j); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := db.DB.Exec("INSERT INTO journal(character_id,title,entry,entry_date) VALUES(?,?,?,?)",
		charID, j.Title, j.Entry, j.EntryDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func UpdateJournalEntry(c *gin.Context) {
	jid, _ := strconv.ParseInt(c.Param("jid"), 10, 64)
	var charID int64
	err := db.DB.QueryRow("SELECT character_id FROM journal WHERE id=?", jid).Scan(&charID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "journal entry not found"})
		return
	}
	if !checkCharacterAccess(c, charID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	var j models.JournalEntry
	if err := c.ShouldBindJSON(&j); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Exec("UPDATE journal SET title=?,entry=?,entry_date=? WHERE id=?", j.Title, j.Entry, j.EntryDate, jid)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteJournalEntry(c *gin.Context) {
	jid, _ := strconv.ParseInt(c.Param("jid"), 10, 64)
	var charID int64
	err := db.DB.QueryRow("SELECT character_id FROM journal WHERE id=?", jid).Scan(&charID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "journal entry not found"})
		return
	}
	if !checkCharacterAccess(c, charID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	db.DB.Exec("DELETE FROM journal WHERE id=?", jid)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ─── Graph Data ───

func GetGraphData(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	gd := models.GraphData{Nodes: []models.GraphNode{}, Edges: []models.GraphEdge{}}

	// Character node
	charName, race, class := "", "", ""
	db.DB.QueryRow("SELECT name,race,class FROM characters WHERE id=?", charID).Scan(&charName, &race, &class)
	gd.Nodes = append(gd.Nodes, models.GraphNode{
		ID: "char_" + strconv.FormatInt(charID, 10), Label: charName + " (" + race + " " + class + ")",
		Group: "character", Color: "#8b0000", Size: 30, CharID: charID,
	})

	// Location nodes + edges
	lrows, _ := db.DB.Query(`
		SELECT cl.id, cl.location_id, cl.relationship, l.name, l.type
		FROM character_locations cl JOIN locations l ON cl.location_id = l.id
		WHERE cl.character_id=?`, charID)
	if lrows != nil {
		for lrows.Next() {
			var linkID, locID int64
			var rel, locName, locType string
			lrows.Scan(&linkID, &locID, &rel, &locName, &locType)
			nid := "loc_" + strconv.FormatInt(locID, 10)
			gd.Nodes = append(gd.Nodes, models.GraphNode{
				ID: nid, Label: locName + " (" + locType + ")", Group: "location", Color: "#b8963e", Size: 20,
			})
			gd.Edges = append(gd.Edges, models.GraphEdge{
				From: "char_" + strconv.FormatInt(charID, 10), To: nid, Label: rel, Width: 2, Dashes: false,
			})
		}
		lrows.Close()
	}

	// NPC nodes + edges
	nrows, _ := db.DB.Query(`
		SELECT cn.npc_id, cn.relationship, cn.interaction_count, n.name
		FROM character_npcs cn JOIN npcs n ON cn.npc_id = n.id
		WHERE cn.character_id=?`, charID)
	if nrows != nil {
		for nrows.Next() {
			var npcID, intCount int64
			var rel, npcName string
			nrows.Scan(&npcID, &rel, &intCount, &npcName)
			nid := "npc_" + strconv.FormatInt(npcID, 10)
			edgeWidth := 1
			if intCount > 5 {
				edgeWidth = 5
			} else if intCount > 2 {
				edgeWidth = 3
			} else if intCount > 0 {
				edgeWidth = 2
			}
			gd.Nodes = append(gd.Nodes, models.GraphNode{
				ID: nid, Label: npcName + " (NPC)", Group: "npc", Color: "#2c6b2f", Size: 20,
			})
			gd.Edges = append(gd.Edges, models.GraphEdge{
				From: "char_" + strconv.FormatInt(charID, 10), To: nid, Label: rel, Width: edgeWidth, Dashes: false,
			})
		}
		nrows.Close()
	}

	// Quest nodes + edges
	qrows, _ := db.DB.Query("SELECT id, name, status FROM quests WHERE character_id=?", charID)
	if qrows != nil {
		for qrows.Next() {
			var qid int64
			var qname, status string
			qrows.Scan(&qid, &qname, &status)
			nid := "quest_" + strconv.FormatInt(qid, 10)
			qcolor := "#b8963e"
			if status == "complete" {
				qcolor = "#2d6a2d"
			} else if status == "failed" || status == "abandoned" {
				qcolor = "#666"
			}
			gd.Nodes = append(gd.Nodes, models.GraphNode{
				ID: nid, Label: qname + " [" + status + "]", Group: "quest", Color: qcolor, Size: 18,
			})
			gd.Edges = append(gd.Edges, models.GraphEdge{
				From: "char_" + strconv.FormatInt(charID, 10), To: nid, Label: status, Width: 1, Dashes: status == "available",
			})
		}
		qrows.Close()
	}

	// Session nodes
	srows, _ := db.DB.Query("SELECT id, title, session_date FROM sessions WHERE character_id=? ORDER BY session_date DESC LIMIT 10", charID)
	if srows != nil {
		for srows.Next() {
			var sid int64
			var title, sdate string
			srows.Scan(&sid, &title, &sdate)
			nid := "session_" + strconv.FormatInt(sid, 10)
			slabel := title
			if slabel == "" {
				slabel = "Session " + sdate
			}
			gd.Nodes = append(gd.Nodes, models.GraphNode{
				ID: nid, Label: slabel, Group: "session", Color: "#5c3a2a", Size: 15,
			})
			gd.Edges = append(gd.Edges, models.GraphEdge{
				From: "char_" + strconv.FormatInt(charID, 10), To: nid, Label: "played", Width: 1, Dashes: false,
			})
		}
		srows.Close()
	}

	c.JSON(http.StatusOK, gd)
}

// ─── Party View ───

type PartyMember struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Race       string `json:"race"`
	Class      string `json:"class"`
	Level      int    `json:"level"`
	AC         int    `json:"ac"`
	HPMax      int    `json:"hp_max"`
	HPCurrent  int    `json:"hp_current"`
	TempHP     int    `json:"temp_hp"`
	Status     string `json:"status"`
	CampaignID *int64 `json:"campaign_id"`
}

func GetPartyView(c *gin.Context) {
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	var rows *sql.Rows
	var err error
	if role == "admin" {
		rows, err = db.DB.Query(`
			SELECT c.id, c.name, c.race, c.class, c.level, c.ac,
				c.hp_max, c.hp_current, c.temp_hp, c.campaign_id
			FROM characters c ORDER BY c.campaign_id, c.name`)
	} else {
		rows, err = db.DB.Query(`
			SELECT c.id, c.name, c.race, c.class, c.level, c.ac,
				c.hp_max, c.hp_current, c.temp_hp, c.campaign_id
			FROM characters c WHERE c.user_id=? ORDER BY c.campaign_id, c.name`, userID)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	campaigns := make(map[int64][]PartyMember)
	var uncategorized []PartyMember

	for rows.Next() {
		var pm PartyMember
		var cid *int64
		rows.Scan(&pm.ID, &pm.Name, &pm.Race, &pm.Class, &pm.Level, &pm.AC,
			&pm.HPMax, &pm.HPCurrent, &pm.TempHP, &cid)
		pm.CampaignID = cid
		pm.Status = "alive"
		if pm.HPCurrent <= 0 {
			pm.Status = "down"
		} else if float64(pm.HPCurrent)/float64(pm.HPMax) < 0.25 {
			pm.Status = "injured"
		}
		if cid != nil {
			campaigns[*cid] = append(campaigns[*cid], pm)
		} else {
			uncategorized = append(uncategorized, pm)
		}
	}

	// Get campaign names
	campNames := make(map[int64]string)
	for cid := range campaigns {
		var name string
		db.DB.QueryRow("SELECT name FROM campaigns WHERE id=?", cid).Scan(&name)
		campNames[cid] = name
	}

	type CampaignGroup struct {
		ID         int64          `json:"id"`
		Name       string         `json:"name"`
		Members    []PartyMember  `json:"members"`
	}
	var groups []CampaignGroup
	for cid, members := range campaigns {
		groups = append(groups, CampaignGroup{ID: cid, Name: campNames[cid], Members: members})
	}
	if len(uncategorized) > 0 {
		groups = append(groups, CampaignGroup{Name: "Uncategorized", Members: uncategorized})
	}

	c.JSON(http.StatusOK, groups)
}

// ─── NPC Interaction Logging ───

func LogNPCInteraction(c *gin.Context) {
	nid, _ := strconv.ParseInt(c.Param("nid"), 10, 64)
	var charID int64
	err := db.DB.QueryRow("SELECT character_id FROM character_npcs WHERE id=?", nid).Scan(&charID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "NPC link not found"})
		return
	}
	if !checkCharacterAccess(c, charID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	db.DB.Exec("UPDATE character_npcs SET interaction_count = interaction_count + 1, last_interacted = datetime('now') WHERE id=?", nid)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ─── Campaigns ───

func ListCampaigns(c *gin.Context) {
	userID, _ := c.Get("user_id")
	rows, err := db.DB.Query("SELECT id,user_id,name,description,dm_notes,created_at FROM campaigns WHERE user_id=? ORDER BY name", userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var out []models.Campaign
	for rows.Next() {
		var ca models.Campaign
		rows.Scan(&ca.ID, &ca.UserID, &ca.Name, &ca.Description, &ca.DMNotes, &ca.CreatedAt)
		out = append(out, ca)
	}
	c.JSON(http.StatusOK, out)
}

func CreateCampaign(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var ca models.Campaign
	if err := c.ShouldBindJSON(&ca); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := db.DB.Exec("INSERT INTO campaigns(user_id,name,description,dm_notes) VALUES(?,?,?,?)",
		userID, ca.Name, ca.Description, ca.DMNotes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id, "name": ca.Name})
}

func UpdateCampaign(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	var ownerID int64
	err := db.DB.QueryRow("SELECT user_id FROM campaigns WHERE id=?", id).Scan(&ownerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "campaign not found"})
		return
	}
	if role != "admin" && ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	var ca models.Campaign
	if err := c.ShouldBindJSON(&ca); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Exec("UPDATE campaigns SET name=?,description=?,dm_notes=? WHERE id=?", ca.Name, ca.Description, ca.DMNotes, id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteCampaign(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	var ownerID int64
	err := db.DB.QueryRow("SELECT user_id FROM campaigns WHERE id=?", id).Scan(&ownerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "campaign not found"})
		return
	}
	if role != "admin" && ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	db.DB.Exec("UPDATE characters SET campaign_id=NULL WHERE campaign_id=?", id)
	db.DB.Exec("DELETE FROM campaigns WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ─── Rest & Level Up ───

func DoRest(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		RestType    string `json:"rest_type"`
		HitDiceCount int   `json:"hit_dice_count"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.RestType != "short" && req.RestType != "long" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rest_type must be 'short' or 'long'"})
		return
	}

	var hpMax, hpCur, con, level, hdCurrent int
	var hitDice string
	err := db.DB.QueryRow("SELECT hp_max, hp_current, hit_dice, hit_dice_current, con, level FROM characters WHERE id=?", charID).
		Scan(&hpMax, &hpCur, &hitDice, &hdCurrent, &con, &level)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}

	hpHealed := 0
	if req.RestType == "long" {
		// Long rest: full heal, recover half max hit dice, reset spell slots/death saves
		hpHealed = hpMax - hpCur
		recoveredHD := level / 2
		if recoveredHD < 1 {
			recoveredHD = 1
		}
		newHD := hdCurrent + recoveredHD
		if newHD > level {
			newHD = level
		}
		db.DB.Exec("UPDATE characters SET hp_current=hp_max, hit_dice_current=?, death_saves_successes=0, death_saves_failures=0, concentrating_on='' WHERE id=?", newHD, charID)
		db.DB.Exec(`UPDATE character_spellcasting SET
			slots_1_used=0, slots_2_used=0, slots_3_used=0, slots_4_used=0,
			slots_5_used=0, slots_6_used=0, slots_7_used=0, slots_8_used=0, slots_9_used=0
			WHERE character_id=?`, charID)
	} else {
		// Short rest: spend individual hit dice
		count := req.HitDiceCount
		if count < 0 {
			count = 0
		}
		if count > hdCurrent {
			count = hdCurrent
		}
		if count == 0 && hpMax > 0 {
			// Default to spending 1 hit die if none specified
			count = 1
			if count > hdCurrent {
				count = hdCurrent
			}
		}
		hitDieSize := 10
		if len(hitDice) > 1 {
			dieSizeStr := hitDice[2:]
			if d, err2 := strconv.Atoi(dieSizeStr); err2 == nil {
				hitDieSize = d
			}
		}
		conMod := abilityMod(con)
		for i := 0; i < count; i++ {
			roll, _ := randInt(1, hitDieSize)
			heal := roll + conMod
			if heal < 1 {
				heal = 1
			}
			hpHealed += heal
		}
		newHp := hpCur + hpHealed
		if newHp > hpMax {
			newHp = hpMax
		}
		hpHealed = newHp - hpCur
		db.DB.Exec("UPDATE characters SET hp_current=?, hit_dice_current=hit_dice_current-? WHERE id=?", newHp, count, charID)
	}

	db.DB.Exec("INSERT INTO rest_log(character_id,rest_type,hp_healed,notes) VALUES(?,?,?,?)",
		charID, req.RestType, hpHealed, "")

	c.JSON(http.StatusOK, gin.H{"ok": true, "hp_healed": hpHealed, "rest_type": req.RestType})
}

func LevelUp(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var level, hpMax, hpCur, con int
	var hitDice string
	err := db.DB.QueryRow("SELECT level, hp_max, hp_current, hit_dice, con FROM characters WHERE id=?", charID).Scan(&level, &hpMax, &hpCur, &hitDice, &con)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}

	newLevel := level + 1
	// Estimate HP gain: average of hit die + CON mod
	hitDieSize := 10 // default
	if len(hitDice) > 1 {
		// Extract die size from "1dX"
		dieSizeStr := hitDice[2:]
		if d, err2 := strconv.Atoi(dieSizeStr); err2 == nil {
			hitDieSize = d
		}
	}
	conMod := abilityMod(con)
	hpGain := (hitDieSize/2 + 1) + conMod // average roll + CON
	if hpGain < 1 {
		hpGain = 1
	}

	newHP := hpMax + hpGain
	newCur := hpCur + hpGain
	if newCur > newHP {
		newCur = newHP
	}
	db.DB.Exec("UPDATE characters SET level=?, hp_max=?, hp_current=?, hit_dice_current=hit_dice_current+1 WHERE id=?", newLevel, newHP, newCur, charID)

	// Update proficiency bonus
	newProf := 2
	if newLevel >= 17 {
		newProf = 6
	} else if newLevel >= 13 {
		newProf = 5
	} else if newLevel >= 9 {
		newProf = 4
	} else if newLevel >= 5 {
		newProf = 3
	}
	db.DB.Exec("UPDATE characters SET proficiency_bonus=? WHERE id=? AND proficiency_bonus<?", newProf, charID, newProf)

	c.JSON(http.StatusOK, gin.H{"ok": true, "new_level": newLevel, "hp_gain": hpGain, "new_hp_max": newHP})
}
