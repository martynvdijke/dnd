package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/ent"
	"villum/ent/campaign"
	"villum/ent/campaigncalendarevent"
	"villum/ent/campaignmember"
	"villum/ent/campaigntimelineevent"
	"villum/ent/campaignwikipage"
	"villum/ent/character"
	"villum/ent/characterlocation"
	"villum/ent/characternpc"
	"villum/ent/characterspellcasting"
	"villum/ent/encountertemplate"
	"villum/ent/faction"
	"villum/ent/factionreputation"
	"villum/ent/journalentry"
	"villum/ent/location"
	"villum/ent/npc"
	"villum/ent/quest"
	"villum/ent/session"
	"villum/ent/user"
	"villum/models"
)

type CampaignMemberResponse struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

type CampaignGroup struct {
	ID        int64         `json:"id"`
	Name      string        `json:"name"`
	PartyName string        `json:"party_name"`
	OwnerName string        `json:"owner_name"`
	Members   []PartyMember `json:"members"`
}

// ─── Locations ───

func ListLocations(c *gin.Context) {
	userID, _ := c.Get("user_id")
	locs, err := db.Client.Location.Query().Where(location.UserID(userID.(int64))).Order(location.ByName()).All(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var out = make([]models.Location, 0)
	for _, loc := range locs {
		l := models.Location{
			ID:          loc.ID,
			UserID:      loc.UserID,
			Name:        loc.Name,
			Type:        loc.Type,
			Description: loc.Description,
			CreatedAt:   loc.CreatedAt,
		}
		if loc.ParentID != 0 {
			l.ParentID = &loc.ParentID
		}
		if loc.Latitude != 0 {
			l.Latitude = &loc.Latitude
		}
		if loc.Longitude != 0 {
			l.Longitude = &loc.Longitude
		}
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
	create := db.Client.Location.Create().SetUserID(userID.(int64)).SetName(l.Name).SetType(l.Type).SetDescription(l.Description)
	if l.ParentID != nil {
		create.SetParentID(*l.ParentID)
	}
	if l.Latitude != nil {
		create.SetLatitude(*l.Latitude)
	}
	if l.Longitude != nil {
		create.SetLongitude(*l.Longitude)
	}
	result, err := create.Save(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": result.ID})
}

func UpdateLocation(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	loc, err := db.Client.Location.Get(c.Request.Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "location not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if role != "admin" && loc.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	var l models.Location
	if err := c.ShouldBindJSON(&l); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	update := db.Client.Location.UpdateOneID(id).SetName(l.Name).SetType(l.Type).SetDescription(l.Description)
	if l.ParentID != nil {
		update.SetParentID(*l.ParentID)
	}
	if l.Latitude != nil {
		update.SetLatitude(*l.Latitude)
	}
	if l.Longitude != nil {
		update.SetLongitude(*l.Longitude)
	}
	update.Save(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteLocation(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	loc, err := db.Client.Location.Get(c.Request.Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "location not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if role != "admin" && loc.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	db.Client.Location.DeleteOneID(id).Exec(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func LinkLocation(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var cl models.CharacterLocation
	if err := c.ShouldBindJSON(&cl); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := db.Client.CharacterLocation.Create().
		SetCharacterID(charID).
		SetLocationID(cl.LocationID).
		SetRelationship(cl.Relationship).
		SetNotes(cl.Notes).
		OnConflict(sql.ResolveWithNewValues()).
		ID(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func UnlinkLocation(c *gin.Context) {
	linkID, _ := strconv.ParseInt(c.Param("lid"), 10, 64)
	cl, err := db.Client.CharacterLocation.Get(c.Request.Context(), linkID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
		return
	}
	if !canEditCharacterID(c, cl.CharacterID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	db.Client.CharacterLocation.DeleteOneID(linkID).Exec(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func GetCharacterLocations(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	cls, err := db.Client.CharacterLocation.Query().
		Where(characterlocation.CharacterID(charID)).
		WithLocation().
		Order(ent.Asc(characterlocation.FieldRelationship)).
		All(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	type LocLink struct {
		models.CharacterLocation
		LocationName string `json:"location_name"`
		LocationType string `json:"location_type"`
		Description  string `json:"description"`
	}
	var out = make([]LocLink, 0)
	for _, cl := range cls {
		loc := cl.Edges.Location
		out = append(out, LocLink{
			CharacterLocation: models.CharacterLocation{
				ID:           cl.ID,
				CharacterID:  cl.CharacterID,
				LocationID:   cl.LocationID,
				Relationship: cl.Relationship,
				Notes:        cl.Notes,
			},
			LocationName: loc.Name,
			LocationType: loc.Type,
			Description:  loc.Description,
		})
	}
	c.JSON(http.StatusOK, out)
}

// ─── NPCs ───

func ListNPCs(c *gin.Context) {
	userID, _ := c.Get("user_id")
	npcs, err := db.Client.NPC.Query().Where(npc.UserID(userID.(int64))).Order(npc.ByName()).All(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var out = make([]models.NPC, 0)
	raceColors := GetRaceColorMap()
	for _, n := range npcs {
		out = append(out, models.NPC{
			ID:          n.ID,
			UserID:      n.UserID,
			Name:        n.Name,
			Race:        n.Race,
			Class:       n.Class,
			Description: n.Description,
			Notes:       n.Notes,
			Str:         n.Str,
			Dex:         n.Dex,
			Con:         n.Con,
			Int:         n.Int,
			Wis:         n.Wis,
			Cha:         n.Cha,
			HPMax:       n.HpMax,
			HPCurrent:   n.HpCurrent,
			IsAlive:     n.IsAlive,
			IsFull:      n.IsFull,
			AC:          n.Ac,
			Speed:       n.Speed,
			Skills:      n.Skills,
			Saves:       n.Saves,
			Features:    n.Features,
			Actions:     n.Actions,
			Backstory:   n.Backstory,
			PortraitURL: n.PortraitURL,
			CreatedAt:   n.CreatedAt,
			RaceColor:   raceColors[n.Race],
		})
	}
	c.JSON(http.StatusOK, out)
}

// SearchNPCs searches NPCs across all users (DM-visible). GET /api/npcs/search?q=
func SearchNPCs(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	query := db.Client.NPC.Query()
	if q != "" {
		query = query.Where(npc.NameContainsFold(q))
	}
	npcs, err := query.Order(npc.ByName()).Limit(50).All(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]gin.H, 0, len(npcs))
	for _, n := range npcs {
		out = append(out, gin.H{
			"id":           n.ID,
			"user_id":      n.UserID,
			"name":         n.Name,
			"race":         n.Race,
			"class":        n.Class,
			"description":  n.Description,
			"portrait_url": n.PortraitURL,
		})
	}
	c.JSON(http.StatusOK, out)
}

// SearchLocations searches locations across all users (DM-visible). GET /api/locations/search?q=
func SearchLocations(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	query := db.Client.Location.Query()
	if q != "" {
		query = query.Where(location.NameContainsFold(q))
	}
	locs, err := query.Order(location.ByName()).Limit(50).All(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]gin.H, 0, len(locs))
	for _, l := range locs {
		out = append(out, gin.H{
			"id":          l.ID,
			"user_id":     l.UserID,
			"name":        l.Name,
			"type":        l.Type,
			"description": l.Description,
		})
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
	q := db.Client.NPC.Create().
		SetUserID(userID.(int64)).
		SetName(n.Name).
		SetRace(n.Race).
		SetClass(n.Class).
		SetDescription(n.Description).
		SetNotes(n.Notes).
		SetStr(n.Str).
		SetDex(n.Dex).
		SetCon(n.Con).
		SetInt(n.Int).
		SetWis(n.Wis).
		SetCha(n.Cha).
		SetHpMax(n.HPMax).
		SetHpCurrent(n.HPCurrent).
		SetIsAlive(n.IsAlive).
		SetIsFull(n.IsFull).
		SetAc(n.AC).
		SetSpeed(n.Speed).
		SetSkills(n.Skills).
		SetSaves(n.Saves).
		SetFeatures(n.Features).
		SetActions(n.Actions).
		SetBackstory(n.Backstory).
		SetPortraitURL(n.PortraitURL)
	result, err := q.Save(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": result.ID})
}

func UpdateNPC(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	npcEnt, err := db.Client.NPC.Get(c.Request.Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "NPC not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if role != "admin" && npcEnt.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	var n models.NPC
	if err := c.ShouldBindJSON(&n); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	up := db.Client.NPC.UpdateOneID(id).
		SetName(n.Name).
		SetRace(n.Race).
		SetClass(n.Class).
		SetDescription(n.Description).
		SetNotes(n.Notes).
		SetStr(n.Str).
		SetDex(n.Dex).
		SetCon(n.Con).
		SetInt(n.Int).
		SetWis(n.Wis).
		SetCha(n.Cha).
		SetHpMax(n.HPMax).
		SetHpCurrent(n.HPCurrent).
		SetIsAlive(n.IsAlive).
		SetIsFull(n.IsFull).
		SetAc(n.AC).
		SetSpeed(n.Speed).
		SetSkills(n.Skills).
		SetSaves(n.Saves).
		SetFeatures(n.Features).
		SetActions(n.Actions).
		SetBackstory(n.Backstory).
		SetPortraitURL(n.PortraitURL)
	up.Save(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteNPC(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	npcEnt, err := db.Client.NPC.Get(c.Request.Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "NPC not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if role != "admin" && npcEnt.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	db.Client.NPC.DeleteOneID(id).Exec(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func LinkNPC(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var cn models.CharacterNPC
	if err := c.ShouldBindJSON(&cn); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := db.Client.CharacterNPC.Create().
		SetCharacterID(charID).
		SetNpcID(cn.NPCID).
		SetRelationship(cn.Relationship).
		SetNotes(cn.Notes).
		OnConflict(sql.ResolveWithNewValues()).
		ID(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func UnlinkNPC(c *gin.Context) {
	linkID, _ := strconv.ParseInt(c.Param("nid"), 10, 64)
	cn, err := db.Client.CharacterNPC.Get(c.Request.Context(), linkID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
		return
	}
	if !canEditCharacterID(c, cn.CharacterID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	db.Client.CharacterNPC.DeleteOneID(linkID).Exec(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func GetCharacterNPCs(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	cnpcs, err := db.Client.CharacterNPC.Query().
		Where(characternpc.CharacterID(charID)).
		WithNpc().
		Order(ent.Desc(characternpc.FieldInteractionCount)).
		All(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	type NPCLink struct {
		models.CharacterNPC
		NPCName      string `json:"npc_name"`
		NPCRace      string `json:"npc_race"`
		NPCRaceColor string `json:"npc_race_color,omitempty"`
		NPCClass     string `json:"npc_class"`
		NPHPMax      int    `json:"npc_hp_max"`
		NPHPCurr     int    `json:"npc_hp_current"`
		NPCAlive     bool   `json:"npc_is_alive"`
	}
	var out = make([]NPCLink, 0)
	raceColors := GetRaceColorMap()
	for _, cn := range cnpcs {
		npcEnt := cn.Edges.Npc
		out = append(out, NPCLink{
			CharacterNPC: models.CharacterNPC{
				ID:               cn.ID,
				CharacterID:      cn.CharacterID,
				NPCID:            cn.NpcID,
				Relationship:     cn.Relationship,
				Notes:            cn.Notes,
				InteractionCount: cn.InteractionCount,
				LastInteracted:   cn.LastInteracted,
			},
			NPCName:      npcEnt.Name,
			NPCRace:      npcEnt.Race,
			NPCRaceColor: raceColors[npcEnt.Race],
			NPCClass:     npcEnt.Class,
			NPHPMax:      npcEnt.HpMax,
			NPHPCurr:     npcEnt.HpCurrent,
			NPCAlive:     npcEnt.IsAlive,
		})
	}
	c.JSON(http.StatusOK, out)
}

// ─── Sessions ───

func ListSessions(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	sessions, err := db.Client.Session.Query().
		Where(session.CharacterID(charID)).
		Order(ent.Desc(session.FieldSessionDate)).
		All(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var out = make([]models.Session, 0)
	for _, s := range sessions {
		out = append(out, models.Session{
			ID:              s.ID,
			CharacterID:     s.CharacterID,
			SessionDate:     s.SessionDate,
			Title:           s.Title,
			Notes:           s.Notes,
			XPEarned:        s.XpEarned,
			GoldEarned:      s.GoldEarned,
			ImportantEvents: s.ImportantEvents,
			CreatedAt:       s.CreatedAt,
		})
	}
	c.JSON(http.StatusOK, out)
}

func CreateSession(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if !canEditCharacterID(c, charID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	var s models.Session
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := db.Client.Session.Create().
		SetCharacterID(charID).
		SetSessionDate(s.SessionDate).
		SetTitle(s.Title).
		SetNotes(s.Notes).
		SetXpEarned(s.XPEarned).
		SetGoldEarned(s.GoldEarned).
		SetImportantEvents(s.ImportantEvents).
		Save(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": result.ID})
}

func UpdateSession(c *gin.Context) {
	sid, _ := strconv.ParseInt(c.Param("sid"), 10, 64)
	sess, err := db.Client.Session.Get(c.Request.Context(), sid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	if !canEditCharacterID(c, sess.CharacterID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	var s models.Session
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.Client.Session.UpdateOneID(sid).
		SetSessionDate(s.SessionDate).
		SetTitle(s.Title).
		SetNotes(s.Notes).
		SetXpEarned(s.XPEarned).
		SetGoldEarned(s.GoldEarned).
		SetImportantEvents(s.ImportantEvents).
		Save(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteSession(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("sid"), 10, 64)
	sess, err := db.Client.Session.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	if !canEditCharacterID(c, sess.CharacterID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	db.Client.Session.DeleteOneID(id).Exec(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ─── Quests ───

func ListQuests(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	quests, err := db.Client.Quest.Query().
		Where(quest.CharacterID(charID)).
		Order(quest.ByStatus(), quest.ByName()).
		All(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var out = make([]models.Quest, 0)
	for _, q := range quests {
		out = append(out, models.Quest{
			ID:          q.ID,
			CharacterID: q.CharacterID,
			Name:        q.Name,
			Description: q.Description,
			Status:      q.Status,
			Objectives:  q.Objectives,
			Rewards:     q.Rewards,
			Notes:       q.Notes,
			CreatedAt:   q.CreatedAt,
			UpdatedAt:   q.UpdatedAt,
		})
	}
	c.JSON(http.StatusOK, out)
}

func CreateQuest(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if !canEditCharacterID(c, charID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	var q models.Quest
	if err := c.ShouldBindJSON(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if q.Status == "" {
		q.Status = "active"
	}
	result, err := db.Client.Quest.Create().
		SetCharacterID(charID).
		SetName(q.Name).
		SetDescription(q.Description).
		SetStatus(q.Status).
		SetObjectives(q.Objectives).
		SetRewards(q.Rewards).
		SetNotes(q.Notes).
		Save(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": result.ID})
}

func UpdateQuest(c *gin.Context) {
	qid, _ := strconv.ParseInt(c.Param("qid"), 10, 64)
	if !canEditResourceID(c, "quests", qid) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	var q models.Quest
	if err := c.ShouldBindJSON(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.Client.Quest.UpdateOneID(qid).
		SetName(q.Name).
		SetDescription(q.Description).
		SetStatus(q.Status).
		SetObjectives(q.Objectives).
		SetRewards(q.Rewards).
		SetNotes(q.Notes).
		SetUpdatedAt(time.Now().Format("2006-01-02 15:04:05")).
		Save(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteQuest(c *gin.Context) {
	qid, _ := strconv.ParseInt(c.Param("qid"), 10, 64)
	q, err := db.Client.Quest.Get(c.Request.Context(), qid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "quest not found"})
		return
	}
	if !canEditCharacterID(c, q.CharacterID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	db.Client.Quest.DeleteOneID(qid).Exec(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ─── Journal ───

func ListJournal(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	entries, err := db.Client.JournalEntry.Query().
		Where(journalentry.CharacterID(charID)).
		Order(ent.Desc(journalentry.FieldEntryDate)).
		All(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var out = make([]models.JournalEntry, 0)
	for _, j := range entries {
		out = append(out, models.JournalEntry{
			ID:          j.ID,
			CharacterID: j.CharacterID,
			Title:       j.Title,
			Entry:       j.Entry,
			EntryDate:   j.EntryDate,
			CreatedAt:   j.CreatedAt,
		})
	}
	c.JSON(http.StatusOK, out)
}

func CreateJournalEntry(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if !canEditCharacterID(c, charID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	var j models.JournalEntry
	if err := c.ShouldBindJSON(&j); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := db.Client.JournalEntry.Create().
		SetCharacterID(charID).
		SetTitle(j.Title).
		SetEntry(j.Entry).
		SetEntryDate(j.EntryDate).
		Save(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": result.ID})
}

func UpdateJournalEntry(c *gin.Context) {
	jid, _ := strconv.ParseInt(c.Param("jid"), 10, 64)
	je, err := db.Client.JournalEntry.Get(c.Request.Context(), jid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "journal entry not found"})
		return
	}
	if !canEditCharacterID(c, je.CharacterID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	var j models.JournalEntry
	if err := c.ShouldBindJSON(&j); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.Client.JournalEntry.UpdateOneID(jid).
		SetTitle(j.Title).
		SetEntry(j.Entry).
		SetEntryDate(j.EntryDate).
		Save(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteJournalEntry(c *gin.Context) {
	jid, _ := strconv.ParseInt(c.Param("jid"), 10, 64)
	je, err := db.Client.JournalEntry.Get(c.Request.Context(), jid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "journal entry not found"})
		return
	}
	if !canEditCharacterID(c, je.CharacterID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	db.Client.JournalEntry.DeleteOneID(jid).Exec(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ─── Graph Data ───

func GetGraphData(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	gd := models.GraphData{Nodes: []models.GraphNode{}, Edges: []models.GraphEdge{}}

	char, err := db.Client.Character.Get(c.Request.Context(), charID)
	if err == nil {
		gd.Nodes = append(gd.Nodes, models.GraphNode{
			ID: "char_" + strconv.FormatInt(charID, 10), Label: char.Name + " (" + char.Race + " " + char.Class + ")",
			Group: "character", Color: "#8b0000", Size: 30, CharID: charID,
		})
	}

	cls, _ := db.Client.CharacterLocation.Query().
		Where(characterlocation.CharacterID(charID)).
		WithLocation().
		All(c.Request.Context())
	for _, cl := range cls {
		loc := cl.Edges.Location
		nid := "loc_" + strconv.FormatInt(loc.ID, 10)
		gd.Nodes = append(gd.Nodes, models.GraphNode{
			ID: nid, Label: loc.Name + " (" + loc.Type + ")", Group: "location", Color: "#b8963e", Size: 20,
		})
		gd.Edges = append(gd.Edges, models.GraphEdge{
			From: "char_" + strconv.FormatInt(charID, 10), To: nid, Label: cl.Relationship, Width: 2, Dashes: false,
		})
	}

	cnpcs, _ := db.Client.CharacterNPC.Query().
		Where(characternpc.CharacterID(charID)).
		WithNpc().
		All(c.Request.Context())
	for _, cn := range cnpcs {
		npcEnt := cn.Edges.Npc
		nid := "npc_" + strconv.FormatInt(npcEnt.ID, 10)
		edgeWidth := 1
		if cn.InteractionCount > 5 {
			edgeWidth = 5
		} else if cn.InteractionCount > 2 {
			edgeWidth = 3
		} else if cn.InteractionCount > 0 {
			edgeWidth = 2
		}
		gd.Nodes = append(gd.Nodes, models.GraphNode{
			ID: nid, Label: npcEnt.Name + " (NPC)", Group: "npc", Color: "#2c6b2f", Size: 20,
		})
		gd.Edges = append(gd.Edges, models.GraphEdge{
			From: "char_" + strconv.FormatInt(charID, 10), To: nid, Label: cn.Relationship, Width: edgeWidth, Dashes: false,
		})
	}

	quests, _ := db.Client.Quest.Query().
		Where(quest.CharacterID(charID)).
		All(c.Request.Context())
	for _, q := range quests {
		nid := "quest_" + strconv.FormatInt(q.ID, 10)
		qcolor := "#b8963e"
		if q.Status == "complete" {
			qcolor = "#2d6a2d"
		} else if q.Status == "failed" || q.Status == "abandoned" {
			qcolor = "#666"
		}
		gd.Nodes = append(gd.Nodes, models.GraphNode{
			ID: nid, Label: q.Name + " [" + q.Status + "]", Group: "quest", Color: qcolor, Size: 18,
		})
		gd.Edges = append(gd.Edges, models.GraphEdge{
			From: "char_" + strconv.FormatInt(charID, 10), To: nid, Label: q.Status, Width: 1, Dashes: q.Status == "available",
		})
	}

	sessions, _ := db.Client.Session.Query().
		Where(session.CharacterID(charID)).
		Order(ent.Desc(session.FieldSessionDate)).
		Limit(10).
		All(c.Request.Context())
	for _, s := range sessions {
		nid := "session_" + strconv.FormatInt(s.ID, 10)
		slabel := s.Title
		if slabel == "" {
			slabel = "Session " + s.SessionDate
		}
		gd.Nodes = append(gd.Nodes, models.GraphNode{
			ID: nid, Label: slabel, Group: "session", Color: "#5c3a2a", Size: 15,
		})
		gd.Edges = append(gd.Edges, models.GraphEdge{
			From: "char_" + strconv.FormatInt(charID, 10), To: nid, Label: "played", Width: 1, Dashes: false,
		})
	}

	c.JSON(http.StatusOK, gd)
}

func GetCampaignGraphData(c *gin.Context) {
	campaignID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	gd := models.GraphData{Nodes: []models.GraphNode{}, Edges: []models.GraphEdge{}}
	nodeSet := map[string]bool{}

	camp, _ := db.Client.Campaign.Get(c.Request.Context(), campaignID)
	if camp != nil {
		gd.Nodes = append(gd.Nodes, models.GraphNode{
			ID: "camp_" + strconv.FormatInt(campaignID, 10), Label: camp.Name,
			Group: "campaign", Color: "#8b0000", Size: 35,
		})
	}
	nodeSet["camp_"+strconv.FormatInt(campaignID, 10)] = true

	if role != "admin" {
		exists, _ := db.Client.CampaignMember.Query().
			Where(campaignmember.And(campaignmember.CampaignID(campaignID), campaignmember.UserID(userID.(int64)))).
			Exist(c.Request.Context())
		isOwner, _ := db.Client.Campaign.Query().
			Where(campaign.And(campaign.ID(campaignID), campaign.UserID(userID.(int64)))).
			Exist(c.Request.Context())
		if !exists && !isOwner {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
	}

	chars, _ := db.Client.Character.Query().
		Where(character.CampaignID(campaignID)).
		All(c.Request.Context())
	for _, ch := range chars {
		nid := "char_" + strconv.FormatInt(ch.ID, 10)
		if !nodeSet[nid] {
			gd.Nodes = append(gd.Nodes, models.GraphNode{
				ID: nid, Label: ch.Name + " (Lvl " + strconv.Itoa(ch.Level) + " " + ch.Race + " " + ch.Class + ")",
				Group: "character", Color: "#8b0000", Size: 28, CharID: ch.ID,
			})
			nodeSet[nid] = true
		}
		gd.Edges = append(gd.Edges, models.GraphEdge{
			From: "camp_" + strconv.FormatInt(campaignID, 10), To: nid, Label: "member", Width: 2, Dashes: false,
		})
	}

	wikiPages, _ := db.Client.CampaignWikiPage.Query().
		Where(campaignwikipage.CampaignID(campaignID)).
		Order(campaignwikipage.BySortOrder(), campaignwikipage.ByTitle()).
		All(c.Request.Context())
	wikiMap := map[int64]string{}
	for _, wp := range wikiPages {
		nid := "wiki_" + strconv.FormatInt(wp.ID, 10)
		if !nodeSet[nid] {
			gd.Nodes = append(gd.Nodes, models.GraphNode{
				ID: nid, Label: wp.Title, Group: "wiki", Color: "#b8963e", Size: 18,
			})
			nodeSet[nid] = true
		}
		wikiMap[wp.ID] = nid
		gd.Edges = append(gd.Edges, models.GraphEdge{
			From: "camp_" + strconv.FormatInt(campaignID, 10), To: nid, Label: "wiki", Width: 1, Dashes: false,
		})
		if wp.ParentID != 0 {
			if pnid, ok := wikiMap[wp.ParentID]; ok {
				gd.Edges = append(gd.Edges, models.GraphEdge{
					From: pnid, To: nid, Label: "child", Width: 1, Dashes: true,
				})
			}
		}
	}

	cls, _ := db.Client.CharacterLocation.Query().
		Where(characterlocation.HasCharacterWith(character.CampaignID(campaignID))).
		WithLocation().
		All(c.Request.Context())
	for _, cl := range cls {
		loc := cl.Edges.Location
		nid := "loc_" + strconv.FormatInt(loc.ID, 10)
		if !nodeSet[nid] {
			gd.Nodes = append(gd.Nodes, models.GraphNode{
				ID: nid, Label: loc.Name + " (" + loc.Type + ")", Group: "location", Color: "#b8963e", Size: 20,
			})
			nodeSet[nid] = true
		}
	}

	for _, cl := range cls {
		from := "char_" + strconv.FormatInt(cl.CharacterID, 10)
		to := "loc_" + strconv.FormatInt(cl.LocationID, 10)
		if nodeSet[from] && nodeSet[to] {
			gd.Edges = append(gd.Edges, models.GraphEdge{
				From: from, To: to, Label: cl.Relationship, Width: 2, Dashes: false,
			})
		}
	}

	cnpcs, _ := db.Client.CharacterNPC.Query().
		Where(characternpc.HasCharacterWith(character.CampaignID(campaignID))).
		WithNpc().
		All(c.Request.Context())
	for _, cn := range cnpcs {
		npcEnt := cn.Edges.Npc
		nid := "npc_" + strconv.FormatInt(npcEnt.ID, 10)
		if !nodeSet[nid] {
			label := npcEnt.Name + " (NPC)"
			if npcEnt.Race != "" || npcEnt.Class != "" {
				label = npcEnt.Name + " (" + npcEnt.Race + " " + npcEnt.Class + ")"
			}
			gd.Nodes = append(gd.Nodes, models.GraphNode{
				ID: nid, Label: label, Group: "npc", Color: "#2d6a2d", Size: 20,
			})
			nodeSet[nid] = true
		}
	}

	for _, cn := range cnpcs {
		from := "char_" + strconv.FormatInt(cn.CharacterID, 10)
		to := "npc_" + strconv.FormatInt(cn.NpcID, 10)
		edgeWidth := 1
		if cn.InteractionCount > 5 {
			edgeWidth = 5
		} else if cn.InteractionCount > 2 {
			edgeWidth = 3
		} else if cn.InteractionCount > 0 {
			edgeWidth = 2
		}
		if nodeSet[from] && nodeSet[to] {
			gd.Edges = append(gd.Edges, models.GraphEdge{
				From: from, To: to, Label: cn.Relationship, Width: edgeWidth, Dashes: false,
			})
		}
	}

	quests, _ := db.Client.Quest.Query().
		Where(quest.HasCharacterWith(character.CampaignID(campaignID))).
		All(c.Request.Context())
	for _, q := range quests {
		nid := "quest_" + strconv.FormatInt(q.ID, 10)
		if !nodeSet[nid] {
			qcolor := "#b8963e"
			if q.Status == "complete" {
				qcolor = "#2d6a2d"
			} else if q.Status == "failed" || q.Status == "abandoned" {
				qcolor = "#666"
			}
			gd.Nodes = append(gd.Nodes, models.GraphNode{
				ID: nid, Label: q.Name + " [" + q.Status + "]", Group: "quest", Color: qcolor, Size: 18,
			})
			nodeSet[nid] = true
		}
		from := "char_" + strconv.FormatInt(q.CharacterID, 10)
		if nodeSet[from] {
			gd.Edges = append(gd.Edges, models.GraphEdge{
				From: from, To: nid, Label: q.Status, Width: 1, Dashes: q.Status == "available",
			})
		}
	}

	sessions, _ := db.Client.Session.Query().
		Where(session.HasCharacterWith(character.CampaignID(campaignID))).
		Order(ent.Desc(session.FieldSessionDate)).
		Limit(30).
		All(c.Request.Context())
	for _, s := range sessions {
		nid := "session_" + strconv.FormatInt(s.ID, 10)
		if !nodeSet[nid] {
			slabel := s.Title
			if slabel == "" {
				slabel = "Session " + s.SessionDate
			}
			gd.Nodes = append(gd.Nodes, models.GraphNode{
				ID: nid, Label: slabel, Group: "session", Color: "#5c3a2a", Size: 14,
			})
			nodeSet[nid] = true
		}
		from := "char_" + strconv.FormatInt(s.CharacterID, 10)
		if nodeSet[from] {
			gd.Edges = append(gd.Edges, models.GraphEdge{
				From: from, To: nid, Label: "played", Width: 1, Dashes: false,
			})
		}
	}

	factions, _ := db.Client.Faction.Query().
		Where(faction.CampaignIDIn(campaignID, 0)).
		All(c.Request.Context())
	factionIDs := []int64{}
	for _, f := range factions {
		nid := "faction_" + strconv.FormatInt(f.ID, 10)
		if !nodeSet[nid] {
			gd.Nodes = append(gd.Nodes, models.GraphNode{
				ID: nid, Label: f.Name + " (" + f.Type + ")", Group: "faction", Color: "#9b59b6", Size: 20,
			})
			nodeSet[nid] = true
		}
		factionIDs = append(factionIDs, f.ID)
	}

	for _, fid := range factionIDs {
		frs, _ := db.Client.FactionReputation.Query().
			Where(factionreputation.And(
				factionreputation.FactionID(fid),
				factionreputation.HasCharacterWith(character.CampaignID(campaignID)),
			)).
			All(c.Request.Context())
		for _, fr := range frs {
			from := "char_" + strconv.FormatInt(fr.CharacterID, 10)
			to := "faction_" + strconv.FormatInt(fid, 10)
			relLabel := "neutral"
			if fr.Standing >= 50 {
				relLabel = "revered"
			} else if fr.Standing >= 25 {
				relLabel = "allied"
			} else if fr.Standing <= -50 {
				relLabel = "hostile"
			} else if fr.Standing <= -25 {
				relLabel = "unfriendly"
			}
			if nodeSet[from] && nodeSet[to] {
				gd.Edges = append(gd.Edges, models.GraphEdge{
					From: from, To: to, Label: relLabel + " (" + strconv.Itoa(fr.Standing) + ")", Width: 2, Dashes: false,
				})
			}
		}
	}

	encs, _ := db.Client.EncounterTemplate.Query().
		Where(encountertemplate.CampaignID(campaignID)).
		All(c.Request.Context())
	for _, e := range encs {
		nid := "encounter_" + strconv.FormatInt(e.ID, 10)
		if !nodeSet[nid] {
			encColor := "#e67e22"
			if e.Difficulty == "hard" {
				encColor = "#c0392b"
			} else if e.Difficulty == "deadly" {
				encColor = "#7b0000"
			}
			gd.Nodes = append(gd.Nodes, models.GraphNode{
				ID: nid, Label: e.Name + " [" + e.Difficulty + "]", Group: "encounter", Color: encColor, Size: 18,
			})
			nodeSet[nid] = true
		}
		gd.Edges = append(gd.Edges, models.GraphEdge{
			From: "camp_" + strconv.FormatInt(campaignID, 10), To: nid, Label: "encounter", Width: 1, Dashes: false,
		})
	}

	tls, _ := db.Client.CampaignTimelineEvent.Query().
		Where(campaigntimelineevent.And(
			campaigntimelineevent.CampaignID(campaignID),
			campaigntimelineevent.LinkedEntityTypeNEQ(""),
		)).
		All(c.Request.Context())
	for _, tl := range tls {
		nid := "timeline_" + strconv.FormatInt(tl.ID, 10)
		if !nodeSet[nid] {
			gd.Nodes = append(gd.Nodes, models.GraphNode{
				ID: nid, Label: tl.Title + " [" + tl.EventType + "]", Group: "timeline", Color: "#5c3a2a", Size: 14,
			})
			nodeSet[nid] = true
		}
		var targetNID string
		switch tl.LinkedEntityType {
		case "character":
			targetNID = "char_" + strconv.FormatInt(tl.LinkedEntityID, 10)
		case "npc":
			targetNID = "npc_" + strconv.FormatInt(tl.LinkedEntityID, 10)
		case "location":
			targetNID = "loc_" + strconv.FormatInt(tl.LinkedEntityID, 10)
		case "wiki":
			targetNID = "wiki_" + strconv.FormatInt(tl.LinkedEntityID, 10)
		}
		if targetNID != "" && nodeSet[targetNID] {
			gd.Edges = append(gd.Edges, models.GraphEdge{
				From: targetNID, To: nid, Label: tl.EventType, Width: 1, Dashes: true,
			})
		}
		gd.Edges = append(gd.Edges, models.GraphEdge{
			From: "camp_" + strconv.FormatInt(campaignID, 10), To: nid, Label: "timeline", Width: 1, Dashes: false,
		})
	}

	cals, _ := db.Client.CampaignCalendarEvent.Query().
		Where(campaigncalendarevent.CampaignID(campaignID)).
		Limit(30).
		All(c.Request.Context())
	for _, cal := range cals {
		nid := "calendar_" + strconv.FormatInt(cal.ID, 10)
		if !nodeSet[nid] {
			gd.Nodes = append(gd.Nodes, models.GraphNode{
				ID: nid, Label: cal.Title + " [" + cal.EventType + "]", Group: "calendar", Color: "#b8963e", Size: 14,
			})
			nodeSet[nid] = true
		}
		gd.Edges = append(gd.Edges, models.GraphEdge{
			From: "camp_" + strconv.FormatInt(campaignID, 10), To: nid, Label: "event", Width: 1, Dashes: false,
		})
	}

	c.JSON(http.StatusOK, gd)
}

// ─── Party View ───

type PartyMember struct {
	ID            int64  `json:"id"`
	UserID        int64  `json:"user_id"`
	OwnerName     string `json:"owner_name"`
	Name          string `json:"name"`
	Race          string `json:"race"`
	RaceColor     string `json:"race_color"`
	Class         string `json:"class"`
	Level         int    `json:"level"`
	AC            int    `json:"ac"`
	HPMax         int    `json:"hp_max"`
	HPCurrent     int    `json:"hp_current"`
	TempHP        int    `json:"temp_hp"`
	Status        string `json:"status"`
	PortraitURL   string `json:"portrait_url"`
	CampaignID    *int64 `json:"campaign_id"`
	CharacterType string `json:"character_type"`
	DMNotes       string `json:"dm_notes,omitempty"`
	Owned         bool   `json:"owned"`
}

func GetPartyView(c *gin.Context) {
	userID, _ := c.Get("user_id")
	currentUID := userID.(int64)
	role, _ := c.Get("role")
	ctx := c.Request.Context()

	var rcMap map[string]string
	var camps []*ent.Campaign
	var err error

	if role == "admin" {
		camps, err = db.Client.Campaign.Query().WithUser().All(ctx)
	} else {
		camps, err = db.Client.Campaign.Query().
			Where(campaign.Or(
				campaign.UserID(currentUID),
				campaign.HasMembersWith(campaignmember.UserID(currentUID)),
			)).
			WithUser().
			All(ctx)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	includeAll := role == "admin"
	userSet := make(map[int64]bool)
	userSet[currentUID] = true
	for _, ca := range camps {
		userSet[ca.UserID] = true
		if !includeAll {
			members, _ := db.Client.CampaignMember.Query().
				Where(campaignmember.CampaignID(ca.ID)).
				All(ctx)
			for _, m := range members {
				userSet[m.UserID] = true
			}
		}
	}

	uidList := make([]int64, 0, len(userSet))
	for uid := range userSet {
		uidList = append(uidList, uid)
	}

	// Campaign IDs where the current user is the DM (owner or member with dm role).
	// DM notes are only exposed to admins and the campaign's DM.
	dmCampaignIDs := make(map[int64]bool)
	if role != "admin" {
		for _, ca := range camps {
			if ca.UserID == currentUID {
				dmCampaignIDs[ca.ID] = true
				continue
			}
			n, err := db.Client.CampaignMember.Query().
				Where(
					campaignmember.CampaignIDEQ(ca.ID),
					campaignmember.UserIDEQ(currentUID),
					campaignmember.RoleEQ("dm"),
				).
				Count(ctx)
			if err == nil && n > 0 {
				dmCampaignIDs[ca.ID] = true
			}
		}
	}

	var chars []*ent.Character
	if includeAll {
		chars, err = db.Client.Character.Query().WithUser().Order(character.ByCampaignID(), character.ByName()).All(ctx)
	} else if len(uidList) == 0 {
		c.JSON(http.StatusOK, []CampaignGroup{})
		return
	} else {
		chars, err = db.Client.Character.Query().
			Where(character.UserIDIn(uidList...)).
			WithUser().
			Order(character.ByCampaignID(), character.ByName()).
			All(ctx)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	campaigns := make(map[int64][]PartyMember)
	var uncategorized []PartyMember

	for _, ch := range chars {
		ownerName := ""
		if ch.Edges.User != nil {
			ownerName = ch.Edges.User.Username
		}
		// Cache race colors for this request
		var raceColor string
		if rcMap == nil {
			rcMap = GetRaceColorMap()
		}
		if rc, ok := rcMap[strings.ToLower(strings.TrimSpace(ch.Race))]; ok {
			raceColor = rc
		}
		pm := PartyMember{
			ID:            ch.ID,
			UserID:        ch.UserID,
			OwnerName:     ownerName,
			Name:          ch.Name,
			Race:          ch.Race,
			RaceColor:     raceColor,
			Class:         ch.Class,
			Level:         ch.Level,
			AC:            ch.Ac,
			HPMax:         ch.HpMax,
			HPCurrent:     ch.HpCurrent,
			TempHP:        ch.TempHp,
			Status:        "alive",
			PortraitURL:   ch.PortraitURL,
			CharacterType: ch.CharacterType,
			Owned:         canEditCharacter(c, ch),
		}
		if role == "admin" || (ch.CampaignID != 0 && dmCampaignIDs[ch.CampaignID]) {
			pm.DMNotes = ch.DmNotes
		}
		if pm.HPCurrent <= 0 {
			pm.Status = "down"
		} else if float64(pm.HPCurrent)/float64(pm.HPMax) < 0.25 {
			pm.Status = "injured"
		}
		if ch.CampaignID != 0 {
			cid := ch.CampaignID
			pm.CampaignID = &cid
			campaigns[ch.CampaignID] = append(campaigns[ch.CampaignID], pm)
		} else {
			uncategorized = append(uncategorized, pm)
		}
	}

	campNames := make(map[int64]string)
	campPartyNames := make(map[int64]string)
	campOwners := make(map[int64]string)
	for _, ca := range camps {
		campNames[ca.ID] = ca.Name
		campPartyNames[ca.ID] = ca.PartyName
		if ca.Edges.User != nil {
			campOwners[ca.ID] = ca.Edges.User.Username
		}
	}

	groups := make([]CampaignGroup, 0)
	for cid, members := range campaigns {
		groups = append(groups, CampaignGroup{ID: cid, Name: campNames[cid], PartyName: campPartyNames[cid], OwnerName: campOwners[cid], Members: members})
	}
	if len(uncategorized) > 0 {
		groups = append(groups, CampaignGroup{Name: "Uncategorized", Members: uncategorized})
	}

	c.JSON(http.StatusOK, groups)
}

// ─── NPC Interaction Logging ───

func LogNPCInteraction(c *gin.Context) {
	nid, _ := strconv.ParseInt(c.Param("nid"), 10, 64)
	cn, err := db.Client.CharacterNPC.Get(c.Request.Context(), nid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "NPC link not found"})
		return
	}
	if !canEditCharacterID(c, cn.CharacterID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	db.Client.CharacterNPC.UpdateOneID(nid).
		AddInteractionCount(1).
		SetLastInteracted(time.Now().Format("2006-01-02 15:04:05")).
		Save(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ─── Campaigns ───

func ListCampaigns(c *gin.Context) {
	userID, _ := c.Get("user_id")
	currentUID := userID.(int64)
	ctx := c.Request.Context()

	camps, err := db.Client.Campaign.Query().
		Where(campaign.Or(
			campaign.UserID(currentUID),
			campaign.HasMembersWith(campaignmember.UserID(currentUID)),
		)).
		Order(campaign.ByName()).
		All(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	type CampaignWithRole struct {
		models.Campaign
		MyRole string `json:"my_role"`
	}
	var out = make([]CampaignWithRole, 0)
	for _, ca := range camps {
		myRole := "dm"
		if ca.UserID != currentUID {
			m, err := db.Client.CampaignMember.Query().
				Where(campaignmember.And(campaignmember.CampaignID(ca.ID), campaignmember.UserID(currentUID))).
				Only(ctx)
			if err == nil {
				myRole = m.Role
			}
		}
		out = append(out, CampaignWithRole{
			Campaign: models.Campaign{
				ID:          ca.ID,
				UserID:      ca.UserID,
				Name:        ca.Name,
				PartyName:   ca.PartyName,
				Description: ca.Description,
				DMNotes:     ca.DmNotes,
				CreatedAt:   ca.CreatedAt,
			},
			MyRole: myRole,
		})
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
	if strings.TrimSpace(ca.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	result, err := db.Client.Campaign.Create().
		SetUserID(userID.(int64)).
		SetName(ca.Name).
		SetPartyName(ca.PartyName).
		SetDescription(ca.Description).
		SetDmNotes(ca.DMNotes).
		Save(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	db.Client.CampaignMember.Create().
		SetCampaignID(result.ID).
		SetUserID(userID.(int64)).
		SetRole("dm").
		OnConflict(sql.ResolveWithIgnore()).
		Exec(c.Request.Context())
	c.JSON(http.StatusCreated, gin.H{"id": result.ID, "name": ca.Name, "party_name": ca.PartyName})
}

func UpdateCampaign(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	ctx := c.Request.Context()

	ca, err := db.Client.Campaign.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "campaign not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if role != "admin" && ca.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	var camp models.Campaign
	if err := c.ShouldBindJSON(&camp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.Client.Campaign.UpdateOneID(id).
		SetName(camp.Name).
		SetPartyName(camp.PartyName).
		SetDescription(camp.Description).
		SetDmNotes(camp.DMNotes).
		Save(ctx)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteCampaign(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	ctx := c.Request.Context()

	ca, err := db.Client.Campaign.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "campaign not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if role != "admin" && ca.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	db.Client.Character.Update().Where(character.CampaignID(id)).ClearCampaignID().Save(ctx)
	db.Client.Campaign.DeleteOneID(id).Exec(ctx)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ─── Campaign Roster ───

// isCampaignMember reports whether the requester may manage the campaign:
// admin, campaign owner, or a member of the campaign. The caller is expected
// to have verified the campaign exists (so 404 vs 403 stays distinguishable).
func isCampaignMember(c *gin.Context, campaignID int64) bool {
	userID, _ := c.Get("user_id")
	currentUID, _ := userID.(int64)
	role, _ := c.Get("role")
	ctx := c.Request.Context()

	ca, err := db.Client.Campaign.Get(ctx, campaignID)
	if err != nil {
		return false
	}
	if role == "admin" || ca.UserID == currentUID {
		return true
	}
	count, err := db.Client.CampaignMember.Query().
		Where(
			campaignmember.CampaignID(campaignID),
			campaignmember.UserID(currentUID),
		).
		Count(ctx)
	return err == nil && count > 0
}

// campaignMemberUserIDs returns the user ids of the campaign's members,
// including the owner, deduplicated.
func campaignMemberUserIDs(c *gin.Context, campaignID int64) ([]int64, error) {
	ctx := c.Request.Context()
	ca, err := db.Client.Campaign.Query().
		Where(campaign.ID(campaignID)).
		Select(campaign.FieldUserID).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	ms, err := db.Client.CampaignMember.Query().
		Where(campaignmember.CampaignID(campaignID)).
		Select(campaignmember.FieldUserID).
		All(ctx)
	if err != nil {
		return nil, err
	}
	ids := []int64{ca.UserID}
	seen := map[int64]bool{ca.UserID: true}
	for _, m := range ms {
		if !seen[m.UserID] {
			seen[m.UserID] = true
			ids = append(ids, m.UserID)
		}
	}
	return ids, nil
}

type RosterCandidate struct {
	ID            int64  `json:"id"`
	UserID        int64  `json:"user_id"`
	OwnerUsername string `json:"owner_username"`
	Name          string `json:"name"`
	Race          string `json:"race"`
	Class         string `json:"class"`
	Level         int    `json:"level"`
	PortraitURL   string `json:"portrait_url,omitempty"`
	CharacterType string `json:"character_type"`
	Owned         bool   `json:"owned"`
	InRoster      bool   `json:"in_roster"`
}

// ListCampaignCharacterCandidates returns the characters a member may add to
// the campaign roster: characters owned by the caller plus characters owned by
// other members of the campaign. Characters already assigned to a different
// campaign are excluded; characters in this campaign are flagged `in_roster`.
func ListCampaignCharacterCandidates(c *gin.Context) {
	campaignID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid campaign id"})
		return
	}
	ctx := c.Request.Context()
	userID, _ := c.Get("user_id")
	currentUID, _ := userID.(int64)

	if _, err := db.Client.Campaign.Get(ctx, campaignID); ent.IsNotFound(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "campaign not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !isCampaignMember(c, campaignID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	memberIDs, err := campaignMemberUserIDs(c, campaignID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	users, err := db.Client.User.Query().
		Where(user.IDIn(memberIDs...)).
		All(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	usernames := make(map[int64]string, len(users))
	for _, u := range users {
		usernames[u.ID] = u.Username
	}

	chars, err := db.Client.Character.Query().
		Where(
			character.UserIDIn(memberIDs...),
			character.Or(
				character.CampaignIDIsNil(),
				character.CampaignIDEQ(campaignID),
			),
		).
		Order(character.ByName()).
		All(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	out := make([]RosterCandidate, 0, len(chars))
	for _, ch := range chars {
		out = append(out, RosterCandidate{
			ID:            ch.ID,
			UserID:        ch.UserID,
			OwnerUsername: usernames[ch.UserID],
			Name:          ch.Name,
			Race:          ch.Race,
			Class:         ch.Class,
			Level:         ch.Level,
			PortraitURL:   ch.PortraitURL,
			CharacterType: ch.CharacterType,
			Owned:         ch.UserID == currentUID,
			InRoster:      ch.CampaignID != 0 && ch.CampaignID == campaignID,
		})
	}
	c.JSON(http.StatusOK, out)
}

// AddCampaignCharacter attaches a character to the campaign roster by setting
// its campaign_id. Any campaign member may do this; the target character must
// be owned by the caller or by another member of the campaign, and must not be
// assigned to a different campaign.
func AddCampaignCharacter(c *gin.Context) {
	campaignID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid campaign id"})
		return
	}
	var req struct {
		CharacterID int64 `json:"character_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.CharacterID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "character_id required"})
		return
	}
	ctx := c.Request.Context()
	userID, _ := c.Get("user_id")
	currentUID, _ := userID.(int64)

	if _, err := db.Client.Campaign.Get(ctx, campaignID); ent.IsNotFound(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "campaign not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !isCampaignMember(c, campaignID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	ch, err := db.Client.Character.Get(ctx, req.CharacterID)
	if ent.IsNotFound(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if ch.CampaignID != 0 && ch.CampaignID != campaignID {
		c.JSON(http.StatusConflict, gin.H{"error": "character already assigned to another campaign"})
		return
	}
	// The target must be owned by the caller or by a campaign member.
	if ch.UserID != currentUID {
		memberIDs, err := campaignMemberUserIDs(c, campaignID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		allowed := false
		for _, id := range memberIDs {
			if id == ch.UserID {
				allowed = true
				break
			}
		}
		if !allowed {
			c.JSON(http.StatusBadRequest, gin.H{"error": "character is not owned by you or a campaign member"})
			return
		}
	}

	if err := db.Client.Character.UpdateOneID(req.CharacterID).
		SetCampaignID(campaignID).
		Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"ok": true})
}

// RemoveCampaignCharacter detaches a character from the campaign roster by
// clearing its campaign_id — but only when it currently points at this
// campaign, so a newer assignment is never clobbered.
func RemoveCampaignCharacter(c *gin.Context) {
	campaignID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid campaign id"})
		return
	}
	characterID, err := strconv.ParseInt(c.Param("characterId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid character id"})
		return
	}
	ctx := c.Request.Context()

	if _, err := db.Client.Campaign.Get(ctx, campaignID); ent.IsNotFound(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "campaign not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !isCampaignMember(c, campaignID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	ch, err := db.Client.Character.Get(ctx, characterID)
	if ent.IsNotFound(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if ch.CampaignID != campaignID {
		// No-op: the character is not in this campaign (possibly moved).
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}
	if err := db.Client.Character.UpdateOneID(characterID).
		ClearCampaignID().
		Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ─── Campaign Members ───

func ListCampaignMembers(c *gin.Context) {
	campaignID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	members, err := db.Client.CampaignMember.Query().
		Where(campaignmember.CampaignID(campaignID)).
		WithUser().
		All(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var out []CampaignMemberResponse
	for _, m := range members {
		username := ""
		if m.Edges.User != nil {
			username = m.Edges.User.Username
		}
		out = append(out, CampaignMemberResponse{
			UserID:   m.UserID,
			Username: username,
			Role:     m.Role,
		})
	}
	c.JSON(http.StatusOK, out)
}

func AddCampaignMember(c *gin.Context) {
	campaignID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	ctx := c.Request.Context()

	ca, err := db.Client.Campaign.Get(ctx, campaignID)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "campaign not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if role != "admin" && ca.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the campaign owner can add members"})
		return
	}

	var req struct {
		Username string `json:"username"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username required"})
		return
	}

	targetUser, err := db.Client.User.Query().Where(user.Username(req.Username)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_, err = db.Client.CampaignMember.Create().
		SetCampaignID(campaignID).
		SetUserID(targetUser.ID).
		SetRole("player").
		OnConflict(sql.ResolveWithIgnore()).
		ID(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func SetCampaignMemberRole(c *gin.Context) {
	campaignID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	targetID, _ := strconv.ParseInt(c.Param("userId"), 10, 64)
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	ctx := c.Request.Context()

	ca, err := db.Client.Campaign.Get(ctx, campaignID)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "campaign not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if role != "admin" && ca.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the campaign owner can change roles"})
		return
	}

	var req struct {
		Role string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || (req.Role != "dm" && req.Role != "player") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "role must be 'dm' or 'player'"})
		return
	}

	db.Client.CampaignMember.Update().
		Where(campaignmember.And(campaignmember.CampaignID(campaignID), campaignmember.UserID(targetID))).
		SetRole(req.Role).
		Save(ctx)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func RemoveCampaignMember(c *gin.Context) {
	campaignID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	targetID, _ := strconv.ParseInt(c.Param("userId"), 10, 64)
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	ctx := c.Request.Context()

	ca, err := db.Client.Campaign.Get(ctx, campaignID)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "campaign not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if role != "admin" && ca.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the campaign owner can remove members"})
		return
	}

	db.Client.CampaignMember.Delete().
		Where(campaignmember.And(campaignmember.CampaignID(campaignID), campaignmember.UserID(targetID))).
		Exec(ctx)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ─── User Search ───

func SearchUsers(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusOK, []struct{}{})
		return
	}
	users, err := db.Client.User.Query().
		Where(user.UsernameContainsFold(q)).
		Order(user.ByUsername()).
		Limit(20).
		All(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	type UserResult struct {
		ID       int64  `json:"id"`
		Username string `json:"username"`
	}
	var out []UserResult
	for _, u := range users {
		out = append(out, UserResult{ID: u.ID, Username: u.Username})
	}
	c.JSON(http.StatusOK, out)
}

// ─── Rest & Level Up ───

func DoRest(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ctx := c.Request.Context()

	var req struct {
		RestType     string `json:"rest_type"`
		HitDiceCount int    `json:"hit_dice_count"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.RestType != "short" && req.RestType != "long" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rest_type must be 'short' or 'long'"})
		return
	}

	char, err := db.Client.Character.Get(ctx, charID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}
	if !canEditCharacter(c, char) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	hpHealed := 0
	if req.RestType == "long" {
		hpHealed = char.HpMax - char.HpCurrent
		recoveredHD := max(char.Level/2, 1)
		newHD := min(char.HitDiceCurrent+recoveredHD, char.Level)
		db.Client.Character.UpdateOneID(charID).
			SetHpCurrent(char.HpMax).
			SetHitDiceCurrent(newHD).
			SetDeathSavesSuccesses(0).
			SetDeathSavesFailures(0).
			SetConcentratingOn("").
			Save(ctx)
		// Reduce exhaustion by 1 on long rest
		if char.ExhaustionLevel > 0 {
			newExhaustion := char.ExhaustionLevel - 1
			db.Client.Character.UpdateOneID(charID).SetExhaustionLevel(newExhaustion).Exec(ctx)
		}
		// Reset spell slots — create record if missing
		count, _ := db.Client.CharacterSpellcasting.Query().Where(characterspellcasting.CharacterID(charID)).Count(ctx)
		if count > 0 {
			db.Client.CharacterSpellcasting.Update().
				Where(characterspellcasting.CharacterID(charID)).
				SetSlots1Used(0).
				SetSlots2Used(0).
				SetSlots3Used(0).
				SetSlots4Used(0).
				SetSlots5Used(0).
				SetSlots6Used(0).
				SetSlots7Used(0).
				SetSlots8Used(0).
				SetSlots9Used(0).
				Save(ctx)
		} else if char.Class != "" {
			db.Client.CharacterSpellcasting.Create().
				SetCharacterID(charID).
				SetAbility("").
				SetSaveDc(10).
				SetAttackBonus(0).
				SetSlots1Used(0).
				SetSlots2Used(0).
				SetSlots3Used(0).
				SetSlots4Used(0).
				SetSlots5Used(0).
				SetSlots6Used(0).
				SetSlots7Used(0).
				SetSlots8Used(0).
				SetSlots9Used(0).
				Save(ctx)
		}
	} else {
		count := max(req.HitDiceCount, 0)
		if count > char.HitDiceCurrent {
			count = char.HitDiceCurrent
		}
		if count == 0 && char.HpMax > 0 {
			count = min(1, char.HitDiceCurrent)
		}
		hitDieSize := 10
		if len(char.HitDice) > 1 {
			dieSizeStr := char.HitDice[2:]
			if d, err2 := strconv.Atoi(dieSizeStr); err2 == nil {
				hitDieSize = d
			}
		}
		conMod := abilityMod(char.Con)
		for i := 0; i < count; i++ {
			result, err := getDicePool().Roll(fmt.Sprintf("1d%d", hitDieSize))
			roll := 1
			if err == nil {
				fmt.Sscanf(string(result.Total), "%d", &roll)
			}
			heal := max(roll+conMod, 1)
			hpHealed += heal
		}
		newHp := min(char.HpCurrent+hpHealed, char.HpMax)
		hpHealed = newHp - char.HpCurrent
		db.Client.Character.UpdateOneID(charID).
			SetHpCurrent(newHp).
			SetHitDiceCurrent(char.HitDiceCurrent - count).
			Save(ctx)
	}

	db.Client.RestLog.Create().
		SetCharacterID(charID).
		SetRestType(req.RestType).
		SetHpHealed(hpHealed).
		SetNotes("").
		Save(ctx)

	SendCharacterUpdate(charID)
	SendPartyUpdate()

	c.JSON(http.StatusOK, gin.H{"ok": true, "hp_healed": hpHealed, "rest_type": req.RestType})
}

func LevelUp(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ctx := c.Request.Context()

	char, err := db.Client.Character.Get(ctx, charID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}
	if !canEditCharacter(c, char) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	newLevel := char.Level + 1
	hitDieSize := 10
	if len(char.HitDice) > 1 {
		dieSizeStr := char.HitDice[2:]
		if d, err2 := strconv.Atoi(dieSizeStr); err2 == nil {
			hitDieSize = d
		}
	}
	conMod := abilityMod(char.Con)
	hpGain := max((hitDieSize/2+1)+conMod, 1)

	newHP := char.HpMax + hpGain
	newCur := min(char.HpCurrent+hpGain, newHP)
	db.Client.Character.UpdateOneID(charID).
		SetLevel(newLevel).
		SetHpMax(newHP).
		SetHpCurrent(newCur).
		SetHitDiceCurrent(char.HitDiceCurrent + 1).
		Save(ctx)

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
	if newProf > char.ProficiencyBonus {
		db.Client.Character.UpdateOneID(charID).SetProficiencyBonus(newProf).Save(ctx)
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "new_level": newLevel, "hp_gain": hpGain, "new_hp_max": newHP})
}
