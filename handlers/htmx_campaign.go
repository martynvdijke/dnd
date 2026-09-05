package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"villum/db"
	"villum/models"
)

func HtmxListNPCs(c *gin.Context) {
	charID := c.Query("character_id")
	if charID == "" {
		c.String(http.StatusBadRequest, "character_id required")
		return
	}
	rows, err := db.DB.Query(`
		SELECT cn.id, cn.character_id, cn.npc_id, cn.relationship, cn.notes,
			cn.interaction_count, cn.last_interacted,
			n.name, n.race, n.class, n.hp_max, n.hp_current, n.is_alive, COALESCE(n.portrait_url,'')
		FROM character_npcs cn JOIN npcs n ON cn.npc_id = n.id
		WHERE cn.character_id=? ORDER BY cn.interaction_count DESC`, charID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	var links []npcsLink
	for rows.Next() {
		var nl npcsLink
		rows.Scan(&nl.ID, &nl.CharacterID, &nl.NPCID, &nl.Relationship, &nl.Notes,
			&nl.InteractionCount, &nl.LastInteracted,
			&nl.NPCName, &nl.NPCRace, &nl.NPCClass, &nl.NPHPMax, &nl.NPHPCurr, &nl.NPCAlive, &nl.NPCPortraitURL)
		links = append(links, nl)
	}
	cid, _ := strconv.ParseInt(charID, 10, 64)
	renderTemplate(c, "npcs_list.html", htmxNPCData{CharacterID: cid, NPCs: links})
}

func HtmxNewNPCForm(c *gin.Context) {
	charID := c.Query("character_id")
	cid, _ := strconv.ParseInt(charID, 10, 64)
	renderTemplate(c, "npcs_form.html", htmxNPCData{CharacterID: cid})
}

func HtmxLinkNPCForm(c *gin.Context) {
	charID := c.Query("character_id")
	cid, _ := strconv.ParseInt(charID, 10, 64)
	userID, _ := c.Get("user_id")
	rows, err := db.DB.Query("SELECT id, name, race, class, description, notes, str, dex, con, int, wis, cha, hp_max, hp_current, is_alive, created_at, COALESCE(portrait_url,'') FROM npcs WHERE user_id=? ORDER BY name", userID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	var all []models.NPC
	for rows.Next() {
		var n models.NPC
		rows.Scan(&n.ID, &n.Name, &n.Race, &n.Class, &n.Description, &n.Notes,
			&n.Str, &n.Dex, &n.Con, &n.Int, &n.Wis, &n.Cha, &n.HPMax, &n.HPCurrent, &n.IsAlive, &n.CreatedAt, &n.PortraitURL)
		// n.UserID already filtered by WHERE user_id=?; no additional auth check needed here
		all = append(all, n)
	}
	renderTemplate(c, "npcs_link_form.html", htmxNPCData{CharacterID: cid, AllNPCs: all})
}

func HtmxCreateNPC(c *gin.Context) {
	charID := c.PostForm("character_id")
	if !canEditCharacterID(c, int64(getIntParam(c, "character_id", 0))) {
		c.String(http.StatusForbidden, "access denied")
		return
	}
	userID, _ := c.Get("user_id")
	name := c.PostForm("name")
	if name == "" {
		c.String(http.StatusBadRequest, "name required")
		return
	}
	portraitURL := c.PostForm("portrait_url")
	result, err := db.DB.Exec("INSERT INTO npcs(user_id,name,race,class,description,notes,portrait_url) VALUES(?,?,?,?,?,?,?)", userID, name, c.PostForm("race"), c.PostForm("class"), c.PostForm("description"), c.PostForm("notes"), portraitURL)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	npcID, _ := result.LastInsertId()
	db.DB.Exec("INSERT INTO character_npcs(character_id,npc_id,relationship) VALUES(?,?,?)", charID, npcID, c.PostForm("type"))
	HtmxListNPCs(c)
}

func HtmxLinkNPC(c *gin.Context) {
	charID := c.PostForm("character_id")
	if !canEditCharacterID(c, int64(getIntParam(c, "character_id", 0))) {
		c.String(http.StatusForbidden, "access denied")
		return
	}
	npcID := c.PostForm("npc_id")
	if npcID != "" {
		db.DB.Exec("INSERT INTO character_npcs(character_id,npc_id) VALUES(?,?)", charID, npcID)
	}
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListNPCs(c)
}

func HtmxUnlinkNPC(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var charID string
	db.DB.QueryRow("SELECT character_id FROM character_npcs WHERE id=?", id).Scan(&charID)
	db.DB.Exec("DELETE FROM character_npcs WHERE id=?", id)
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListNPCs(c)
}

// ─── Locations (linked to character) ───

type htmxLocationData struct {
	CharacterID  int64
	Locations    []locLink
	AllLocations []models.Location
}

type locLink struct {
	models.CharacterLocation
	LocationName string `json:"location_name"`
	LocationType string `json:"location_type"`
	Description  string `json:"description"`
}

func HtmxListLocations(c *gin.Context) {
	charID := c.Query("character_id")
	if charID == "" {
		c.String(http.StatusBadRequest, "character_id required")
		return
	}
	rows, err := db.DB.Query(`
		SELECT cl.id, cl.character_id, cl.location_id, cl.relationship, cl.notes,
			l.name, l.type, l.description
		FROM character_locations cl JOIN locations l ON cl.location_id = l.id
		WHERE cl.character_id=? ORDER BY cl.relationship`, charID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	var links []locLink
	for rows.Next() {
		var ll locLink
		rows.Scan(&ll.ID, &ll.CharacterID, &ll.LocationID, &ll.Relationship, &ll.Notes,
			&ll.LocationName, &ll.LocationType, &ll.Description)
		links = append(links, ll)
	}
	cid, _ := strconv.ParseInt(charID, 10, 64)
	renderTemplate(c, "locations_list.html", htmxLocationData{CharacterID: cid, Locations: links})
}

func HtmxNewLocationForm(c *gin.Context) {
	charID := c.Query("character_id")
	cid, _ := strconv.ParseInt(charID, 10, 64)
	renderTemplate(c, "locations_form.html", htmxLocationData{CharacterID: cid})
}

func HtmxLinkLocationForm(c *gin.Context) {
	charID := c.Query("character_id")
	cid, _ := strconv.ParseInt(charID, 10, 64)
	userID, _ := c.Get("user_id")
	rows, err := db.DB.Query("SELECT id, user_id, name, type, description, parent_id, latitude, longitude, created_at FROM locations WHERE user_id=? ORDER BY name", userID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	var all []models.Location
	for rows.Next() {
		var l models.Location
		rows.Scan(&l.ID, &l.UserID, &l.Name, &l.Type, &l.Description, &l.ParentID, &l.Latitude, &l.Longitude, &l.CreatedAt)
		all = append(all, l)
	}
	renderTemplate(c, "locations_link_form.html", htmxLocationData{CharacterID: cid, AllLocations: all})
}

func HtmxCreateLocation(c *gin.Context) {
	charID := c.PostForm("character_id")
	if !canEditCharacterID(c, int64(getIntParam(c, "character_id", 0))) {
		c.String(http.StatusForbidden, "access denied")
		return
	}
	userID, _ := c.Get("user_id")
	name := c.PostForm("name")
	if name == "" {
		return
	}
	result, err := db.DB.Exec("INSERT INTO locations(user_id,name,type,description) VALUES(?,?,?,?)", userID, name, c.PostForm("type"), c.PostForm("description"))
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	locID, _ := result.LastInsertId()
	db.DB.Exec("INSERT INTO character_locations(character_id,location_id) VALUES(?,?)", charID, locID)
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListLocations(c)
}

func HtmxLinkLocation(c *gin.Context) {
	charID := c.PostForm("character_id")
	if !canEditCharacterID(c, int64(getIntParam(c, "character_id", 0))) {
		c.String(http.StatusForbidden, "access denied")
		return
	}
	locID := c.PostForm("location_id")
	if locID != "" {
		db.DB.Exec("INSERT INTO character_locations(character_id,location_id) VALUES(?,?)", charID, locID)
	}
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListLocations(c)
}

func HtmxUnlinkLocation(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var charID string
	db.DB.QueryRow("SELECT character_id FROM character_locations WHERE id=?", id).Scan(&charID)
	db.DB.Exec("DELETE FROM character_locations WHERE id=?", id)
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListLocations(c)
}

// ─── Sessions ───

type htmxSessionData struct {
	CharacterID int64
	Session     *models.Session
	Sessions    []models.Session
}

func HtmxListSessions(c *gin.Context) {
	charID := c.Query("character_id")
	if charID == "" {
		c.String(http.StatusBadRequest, "character_id required")
		return
	}
	rows, err := db.DB.Query("SELECT id, character_id, session_date, title, notes, xp_earned, gold_earned, important_events, created_at FROM sessions WHERE character_id=? ORDER BY session_date DESC", charID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	var sessions []models.Session
	for rows.Next() {
		var s models.Session
		rows.Scan(&s.ID, &s.CharacterID, &s.SessionDate, &s.Title, &s.Notes, &s.XPEarned, &s.GoldEarned, &s.ImportantEvents, &s.CreatedAt)
		sessions = append(sessions, s)
	}
	cid, _ := strconv.ParseInt(charID, 10, 64)
	renderTemplate(c, "sessions_list.html", htmxSessionData{CharacterID: cid, Sessions: sessions})
}

func HtmxNewSessionForm(c *gin.Context) {
	charID := c.Query("character_id")
	cid, _ := strconv.ParseInt(charID, 10, 64)
	renderTemplate(c, "sessions_form.html", htmxSessionData{CharacterID: cid})
}

func HtmxEditSessionForm(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var s models.Session
	err := db.DB.QueryRow("SELECT id, character_id, session_date, title, notes, xp_earned, gold_earned, important_events, created_at FROM sessions WHERE id=?", id).Scan(
		&s.ID, &s.CharacterID, &s.SessionDate, &s.Title, &s.Notes, &s.XPEarned, &s.GoldEarned, &s.ImportantEvents, &s.CreatedAt)
	if err != nil {
		c.String(http.StatusNotFound, "not found")
		return
	}
	renderTemplate(c, "sessions_form.html", htmxSessionData{CharacterID: s.CharacterID, Session: &s})
}

func HtmxCreateSession(c *gin.Context) {
	charID := c.PostForm("character_id")
	if !canEditCharacterID(c, int64(getIntParam(c, "character_id", 0))) {
		c.String(http.StatusForbidden, "access denied")
		return
	}
	db.DB.Exec("INSERT INTO sessions(character_id,session_date,title,notes,xp_earned,gold_earned,important_events) VALUES(?,?,?,?,?,?,?)",
		charID, c.PostForm("session_date"), c.PostForm("title"), c.PostForm("notes"), getIntParam(c, "xp_earned", 0), getIntParam(c, "gold_earned", 0), c.PostForm("important_events"))
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListSessions(c)
}

func HtmxUpdateSession(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	charID := c.PostForm("character_id")
	if !canEditCharacterID(c, int64(getIntParam(c, "character_id", 0))) {
		c.String(http.StatusForbidden, "access denied")
		return
	}
	db.DB.Exec("UPDATE sessions SET session_date=?, title=?, notes=?, xp_earned=?, gold_earned=?, important_events=? WHERE id=?",
		c.PostForm("session_date"), c.PostForm("title"), c.PostForm("notes"), getIntParam(c, "xp_earned", 0), getIntParam(c, "gold_earned", 0), c.PostForm("important_events"), id)
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListSessions(c)
}

func HtmxDeleteSession(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var charID string
	db.DB.QueryRow("SELECT character_id FROM sessions WHERE id=?", id).Scan(&charID)
	db.DB.Exec("DELETE FROM sessions WHERE id=?", id)
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListSessions(c)
}

// ─── Quests ───

type htmxQuestData struct {
	CharacterID int64
	Quest       *models.Quest
	Quests      []models.Quest
}

func HtmxListQuests(c *gin.Context) {
	charID := c.Query("character_id")
	if charID == "" {
		c.String(http.StatusBadRequest, "character_id required")
		return
	}
	rows, err := db.DB.Query("SELECT id, character_id, name, description, status, objectives, rewards, notes, created_at, updated_at FROM quests WHERE character_id=? ORDER BY status, name", charID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	var quests []models.Quest
	for rows.Next() {
		var q models.Quest
		rows.Scan(&q.ID, &q.CharacterID, &q.Name, &q.Description, &q.Status, &q.Objectives, &q.Rewards, &q.Notes, &q.CreatedAt, &q.UpdatedAt)
		quests = append(quests, q)
	}
	cid, _ := strconv.ParseInt(charID, 10, 64)
	renderTemplate(c, "quests_list.html", htmxQuestData{CharacterID: cid, Quests: quests})
}

func HtmxNewQuestForm(c *gin.Context) {
	charID := c.Query("character_id")
	cid, _ := strconv.ParseInt(charID, 10, 64)
	renderTemplate(c, "quests_form.html", htmxQuestData{CharacterID: cid})
}

func HtmxEditQuestForm(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var q models.Quest
	err := db.DB.QueryRow("SELECT id, character_id, name, description, status, objectives, rewards, notes, created_at, updated_at FROM quests WHERE id=?", id).Scan(
		&q.ID, &q.CharacterID, &q.Name, &q.Description, &q.Status, &q.Objectives, &q.Rewards, &q.Notes, &q.CreatedAt, &q.UpdatedAt)
	if err != nil {
		c.String(http.StatusNotFound, "not found")
		return
	}
	renderTemplate(c, "quests_form.html", htmxQuestData{CharacterID: q.CharacterID, Quest: &q})
}

func HtmxCreateQuest(c *gin.Context) {
	charID := c.PostForm("character_id")
	if !canEditCharacterID(c, int64(getIntParam(c, "character_id", 0))) {
		c.String(http.StatusForbidden, "access denied")
		return
	}
	status := c.PostForm("status")
	if status == "" {
		status = "active"
	}
	db.DB.Exec("INSERT INTO quests(character_id,name,description,status,objectives,rewards,notes) VALUES(?,?,?,?,?,?,?)",
		charID, c.PostForm("name"), c.PostForm("description"), status, c.PostForm("objectives"), c.PostForm("rewards"), c.PostForm("notes"))
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListQuests(c)
}

func HtmxUpdateQuest(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	charID := c.PostForm("character_id")
	if !canEditCharacterID(c, int64(getIntParam(c, "character_id", 0))) {
		c.String(http.StatusForbidden, "access denied")
		return
	}
	status := c.PostForm("status")
	if status == "" {
		status = "active"
	}
	db.DB.Exec("UPDATE quests SET name=?, description=?, status=?, objectives=?, rewards=?, notes=?, updated_at=datetime('now') WHERE id=?",
		c.PostForm("name"), c.PostForm("description"), status, c.PostForm("objectives"), c.PostForm("rewards"), c.PostForm("notes"), id)
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListQuests(c)
}

func HtmxDeleteQuest(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var charID string
	db.DB.QueryRow("SELECT character_id FROM quests WHERE id=?", id).Scan(&charID)
	db.DB.Exec("DELETE FROM quests WHERE id=?", id)
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListQuests(c)
}

// ─── Journal ───

type htmxJournalData struct {
	CharacterID int64
	Entry       *models.JournalEntry
	Entries     []models.JournalEntry
}

func HtmxListJournal(c *gin.Context) {
	charID := c.Query("character_id")
	if charID == "" {
		c.String(http.StatusBadRequest, "character_id required")
		return
	}
	rows, err := db.DB.Query("SELECT id, character_id, title, entry, entry_date, created_at FROM journal WHERE character_id=? ORDER BY entry_date DESC", charID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	var entries []models.JournalEntry
	for rows.Next() {
		var j models.JournalEntry
		rows.Scan(&j.ID, &j.CharacterID, &j.Title, &j.Entry, &j.EntryDate, &j.CreatedAt)
		entries = append(entries, j)
	}
	cid, _ := strconv.ParseInt(charID, 10, 64)
	renderTemplate(c, "journal_list.html", htmxJournalData{CharacterID: cid, Entries: entries})
}

func HtmxNewJournalForm(c *gin.Context) {
	charID := c.Query("character_id")
	cid, _ := strconv.ParseInt(charID, 10, 64)
	renderTemplate(c, "journal_form.html", htmxJournalData{CharacterID: cid})
}

func HtmxEditJournalForm(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var j models.JournalEntry
	err := db.DB.QueryRow("SELECT id, character_id, title, entry, entry_date, created_at FROM journal WHERE id=?", id).Scan(
		&j.ID, &j.CharacterID, &j.Title, &j.Entry, &j.EntryDate, &j.CreatedAt)
	if err != nil {
		c.String(http.StatusNotFound, "not found")
		return
	}
	renderTemplate(c, "journal_form.html", htmxJournalData{CharacterID: j.CharacterID, Entry: &j})
}

func HtmxCreateJournal(c *gin.Context) {
	charID := c.PostForm("character_id")
	if !canEditCharacterID(c, int64(getIntParam(c, "character_id", 0))) {
		c.String(http.StatusForbidden, "access denied")
		return
	}
	db.DB.Exec("INSERT INTO journal(character_id,title,entry,entry_date) VALUES(?,?,?,?)",
		charID, c.PostForm("title"), c.PostForm("entry"), c.PostForm("entry_date"))
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListJournal(c)
}

func HtmxUpdateJournal(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	charID := c.PostForm("character_id")
	if !canEditCharacterID(c, int64(getIntParam(c, "character_id", 0))) {
		c.String(http.StatusForbidden, "access denied")
		return
	}
	db.DB.Exec("UPDATE journal SET title=?, entry=?, entry_date=? WHERE id=?",
		c.PostForm("title"), c.PostForm("entry"), c.PostForm("entry_date"), id)
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListJournal(c)
}

func HtmxDeleteJournal(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var charID string
	db.DB.QueryRow("SELECT character_id FROM journal WHERE id=?", id).Scan(&charID)
	db.DB.Exec("DELETE FROM journal WHERE id=?", id)
	c.Request.URL.RawQuery = "character_id=" + charID
	HtmxListJournal(c)
}

// ─── Timeline ───

type htmxTimelineData struct {
	Event  *models.TimelineEvent
	Events []models.TimelineEvent
}

func HtmxListTimeline(c *gin.Context) {
	oneshotID := c.Query("oneshot_adventure_id")
	query := "SELECT id, campaign_id, title, description, event_date, event_type, importance, icon, linked_entity_type, linked_entity_id, created_at FROM campaign_timeline_events" + " WHERE 1=1"
	var args []any
	if oneshotID != "" {
		query += " AND oneshot_adventure_id=?"
		args = append(args, oneshotID)
	}
	query += " ORDER BY event_date DESC"
	rows, err := db.DB.Query(query, args...)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	var events []models.TimelineEvent
	for rows.Next() {
		var e models.TimelineEvent
		rows.Scan(&e.ID, &e.CampaignID, &e.Title, &e.Description, &e.EventDate, &e.EventType, &e.Importance, &e.Icon, &e.LinkedEntityType, &e.LinkedEntityID, &e.CreatedAt)
		events = append(events, e)
	}
	renderTemplate(c, "timeline_list.html", htmxTimelineData{Events: events})
}

func HtmxNewTimelineForm(c *gin.Context) {
	renderTemplate(c, "timeline_form.html", htmxTimelineData{})
}

func HtmxEditTimelineForm(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var e models.TimelineEvent
	err := db.DB.QueryRow("SELECT id, campaign_id, title, description, event_date, event_type, importance, icon, linked_entity_type, linked_entity_id, created_at FROM campaign_timeline_events WHERE id=?", id).Scan(
		&e.ID, &e.CampaignID, &e.Title, &e.Description, &e.EventDate, &e.EventType, &e.Importance, &e.Icon, &e.LinkedEntityType, &e.LinkedEntityID, &e.CreatedAt)
	if err != nil {
		c.String(http.StatusNotFound, "not found")
		return
	}
	renderTemplate(c, "timeline_form.html", htmxTimelineData{Event: &e})
}

func HtmxCreateTimeline(c *gin.Context) {
	db.DB.Exec("INSERT INTO campaign_timeline_events(campaign_id,title,description,event_date,event_type) VALUES(?,?,?,?,?)",
		getIntParam(c, "campaign_id", 1), c.PostForm("title"), c.PostForm("description"), c.PostForm("event_date"), c.PostForm("event_type"))
	HtmxListTimeline(c)
}

func HtmxUpdateTimeline(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	db.DB.Exec("UPDATE campaign_timeline_events SET title=?, description=?, event_date=?, event_type=?, campaign_id=? WHERE id=?",
		c.PostForm("title"), c.PostForm("description"), c.PostForm("event_date"), c.PostForm("event_type"), getIntParam(c, "campaign_id", 1), id)
	HtmxListTimeline(c)
}

func HtmxDeleteTimeline(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	db.DB.Exec("DELETE FROM campaign_timeline_events WHERE id=?", id)
	HtmxListTimeline(c)
}

// ─── Factions ───

type htmxFactionData struct {
	Faction  *models.Faction
	Factions []models.Faction
}

func HtmxListFactions(c *gin.Context) {
	oneshotID := c.Query("oneshot_adventure_id")
	query := "SELECT id, campaign_id, name, description, type, headquarters FROM factions WHERE 1=1"
	var args []any
	if oneshotID != "" {
		query += " AND oneshot_adventure_id=?"
		args = append(args, oneshotID)
	}
	query += " ORDER BY name"
	rows, err := db.DB.Query(query, args...)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	var factions []models.Faction
	for rows.Next() {
		var f models.Faction
		rows.Scan(&f.ID, &f.CampaignID, &f.Name, &f.Description, &f.Type, &f.Headquarters)
		factions = append(factions, f)
	}
	renderTemplate(c, "factions_list.html", htmxFactionData{Factions: factions})
}

func HtmxNewFactionForm(c *gin.Context) {
	renderTemplate(c, "factions_form.html", htmxFactionData{})
}

func HtmxEditFactionForm(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var f models.Faction
	err := db.DB.QueryRow("SELECT id, campaign_id, name, description, type, headquarters FROM factions WHERE id=?", id).Scan(
		&f.ID, &f.CampaignID, &f.Name, &f.Description, &f.Type, &f.Headquarters)
	if err != nil {
		c.String(http.StatusNotFound, "not found")
		return
	}
	renderTemplate(c, "factions_form.html", htmxFactionData{Faction: &f})
}

func HtmxCreateFaction(c *gin.Context) {
	db.DB.Exec("INSERT INTO factions(campaign_id,name,description,type,headquarters) VALUES(?,?,?,?,?)",
		getIntParam(c, "campaign_id", 1), c.PostForm("name"), c.PostForm("description"), c.PostForm("type"), c.PostForm("headquarters"))
	HtmxListFactions(c)
}

func HtmxUpdateFaction(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	db.DB.Exec("UPDATE factions SET name=?, description=?, type=?, headquarters=? WHERE id=?",
		c.PostForm("name"), c.PostForm("description"), c.PostForm("type"), c.PostForm("headquarters"), id)
	HtmxListFactions(c)
}

func HtmxDeleteFaction(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	db.DB.Exec("DELETE FROM factions WHERE id=?", id)
	HtmxListFactions(c)
}

// ─── Media Gallery ───

type htmxMediaGalleryData struct {
	OwnerType string
	OwnerID   string
	Uploads   []mediaUploadItem
}

type mediaUploadItem struct {
	models.Upload
	LinkID int64 `json:"link_id"`
	IsPDF  bool  `json:"is_pdf"`
}
