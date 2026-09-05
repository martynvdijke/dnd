package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/ent"
	"villum/ent/characterlocation"
	"villum/ent/characternpc"
	"villum/ent/location"
	"villum/ent/npc"
	"villum/models"
)

func ListLocations(c *gin.Context) {
	userID, _ := c.Get("user_id")
	locs, err := db.Client.Location.Query().Where(location.UserID(userID.(int64))).Order(location.ByName()).All(c.Request.Context())
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	var out = make([]models.Location, 0)
	for _, loc := range locs {
		l := models.Location{ID: loc.ID, UserID: loc.UserID, Name: loc.Name, Type: loc.Type, Description: loc.Description, CreatedAt: loc.CreatedAt}
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
	WriteJSON(c, http.StatusOK, out)
}

func CreateLocation(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var l models.Location
	if !BindOr400(c, &l) {
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
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	WriteJSON(c, http.StatusCreated, gin.H{"id": result.ID})
}

func UpdateLocation(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	loc, err := db.Client.Location.Get(c.Request.Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			WriteNotFound(c, "location not found")
			return
		}
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	if role != "admin" && loc.UserID != userID {
		WriteError(c, http.StatusForbidden, errAccessDenied)
		return
	}
	var l models.Location
	if !BindOr400(c, &l) {
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
	WriteJSON(c, http.StatusOK, gin.H{"ok": true})
}

func DeleteLocation(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	loc, err := db.Client.Location.Get(c.Request.Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			WriteNotFound(c, "location not found")
			return
		}
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	if role != "admin" && loc.UserID != userID {
		WriteError(c, http.StatusForbidden, errAccessDenied)
		return
	}
	db.Client.Location.DeleteOneID(id).Exec(c.Request.Context())
	WriteJSON(c, http.StatusOK, gin.H{"ok": true})
}

func LinkLocation(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var cl models.CharacterLocation
	if !BindOr400(c, &cl) {
		return
	}
	_, err := db.Client.CharacterLocation.Create().SetCharacterID(charID).SetLocationID(cl.LocationID).SetRelationship(cl.Relationship).SetNotes(cl.Notes).OnConflict(sql.ResolveWithNewValues()).ID(c.Request.Context())
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	WriteJSON(c, http.StatusOK, gin.H{"ok": true})
}

func UnlinkLocation(c *gin.Context) {
	linkID, _ := strconv.ParseInt(c.Param("lid"), 10, 64)
	cl, err := db.Client.CharacterLocation.Get(c.Request.Context(), linkID)
	if err != nil {
		WriteNotFound(c, "link not found")
		return
	}
	if !canEditCharacterID(c, cl.CharacterID) {
		WriteError(c, http.StatusForbidden, errAccessDenied)
		return
	}
	db.Client.CharacterLocation.DeleteOneID(linkID).Exec(c.Request.Context())
	WriteJSON(c, http.StatusOK, gin.H{"ok": true})
}

func GetCharacterLocations(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	cls, err := db.Client.CharacterLocation.Query().Where(characterlocation.CharacterID(charID)).WithLocation().Order(ent.Asc(characterlocation.FieldRelationship)).All(c.Request.Context())
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
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
			CharacterLocation: models.CharacterLocation{ID: cl.ID, CharacterID: cl.CharacterID, LocationID: cl.LocationID, Relationship: cl.Relationship, Notes: cl.Notes},
			LocationName: loc.Name, LocationType: loc.Type, Description: loc.Description,
		})
	}
	WriteJSON(c, http.StatusOK, out)
}

func SearchLocations(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	query := db.Client.Location.Query()
	if q != "" {
		query = query.Where(location.NameContainsFold(q))
	}
	locs, err := query.Order(location.ByName()).Limit(campaignGraphLimit).All(c.Request.Context())
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	out := make([]gin.H, 0, len(locs))
	for _, l := range locs {
		out = append(out, gin.H{"id": l.ID, "user_id": l.UserID, "name": l.Name, "type": l.Type, "description": l.Description})
	}
	WriteJSON(c, http.StatusOK, out)
}

// Personal NPCs (user-owned)
func ListNPCs(c *gin.Context) {
	userID, _ := c.Get("user_id")
	npcs, err := db.Client.NPC.Query().Where(npc.UserID(userID.(int64))).Order(npc.ByName()).All(c.Request.Context())
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	var out = make([]models.NPC, 0)
	raceColors := GetRaceColorMap()
	for _, n := range npcs {
		out = append(out, models.NPC{
			ID: n.ID, UserID: n.UserID, Name: n.Name, Race: n.Race, Class: n.Class,
			Description: n.Description, Notes: n.Notes, Str: n.Str, Dex: n.Dex, Con: n.Con, Int: n.Int, Wis: n.Wis, Cha: n.Cha,
			HPMax: n.HpMax, HPCurrent: n.HpCurrent, IsAlive: n.IsAlive, IsFull: n.IsFull, AC: n.Ac, Speed: n.Speed,
			Skills: n.Skills, Saves: n.Saves, Features: n.Features, Actions: n.Actions, Backstory: n.Backstory, PortraitURL: n.PortraitURL, CreatedAt: n.CreatedAt, RaceColor: raceColors[n.Race],
		})
	}
	WriteJSON(c, http.StatusOK, out)
}

func SearchNPCs(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	query := db.Client.NPC.Query()
	if q != "" {
		query = query.Where(npc.NameContainsFold(q))
	}
	npcs, err := query.Order(npc.ByName()).Limit(campaignGraphLimit).All(c.Request.Context())
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	out := make([]gin.H, 0, len(npcs))
	for _, n := range npcs {
		out = append(out, gin.H{"id": n.ID, "user_id": n.UserID, "name": n.Name, "race": n.Race, "class": n.Class, "description": n.Description, "portrait_url": n.PortraitURL})
	}
	WriteJSON(c, http.StatusOK, out)
}

func CreateNPC(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var n models.NPC
	if !BindOr400(c, &n) {
		return
	}
	result, err := db.Client.NPC.Create().SetUserID(userID.(int64)).SetName(n.Name).SetRace(n.Race).SetClass(n.Class).SetDescription(n.Description).SetNotes(n.Notes).SetStr(n.Str).SetDex(n.Dex).SetCon(n.Con).SetInt(n.Int).SetWis(n.Wis).SetCha(n.Cha).SetHpMax(n.HPMax).SetHpCurrent(n.HPCurrent).SetIsAlive(n.IsAlive).SetIsFull(n.IsFull).SetAc(n.AC).SetSpeed(n.Speed).SetSkills(n.Skills).SetSaves(n.Saves).SetFeatures(n.Features).SetActions(n.Actions).SetBackstory(n.Backstory).SetPortraitURL(n.PortraitURL).Save(c.Request.Context())
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	WriteJSON(c, http.StatusCreated, gin.H{"id": result.ID})
}

func UpdateNPC(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	npcEnt, err := db.Client.NPC.Get(c.Request.Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			WriteNotFound(c, "NPC not found")
			return
		}
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	if role != "admin" && npcEnt.UserID != userID {
		WriteError(c, http.StatusForbidden, errAccessDenied)
		return
	}
	var n models.NPC
	if !BindOr400(c, &n) {
		return
	}
	db.Client.NPC.UpdateOneID(id).SetName(n.Name).SetRace(n.Race).SetClass(n.Class).SetDescription(n.Description).SetNotes(n.Notes).SetStr(n.Str).SetDex(n.Dex).SetCon(n.Con).SetInt(n.Int).SetWis(n.Wis).SetCha(n.Cha).SetHpMax(n.HPMax).SetHpCurrent(n.HPCurrent).SetIsAlive(n.IsAlive).SetIsFull(n.IsFull).SetAc(n.AC).SetSpeed(n.Speed).SetSkills(n.Skills).SetSaves(n.Saves).SetFeatures(n.Features).SetActions(n.Actions).SetBackstory(n.Backstory).SetPortraitURL(n.PortraitURL).Save(c.Request.Context())
	WriteJSON(c, http.StatusOK, gin.H{"ok": true})
}

func DeleteNPC(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	npcEnt, err := db.Client.NPC.Get(c.Request.Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			WriteNotFound(c, "NPC not found")
			return
		}
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	if role != "admin" && npcEnt.UserID != userID {
		WriteError(c, http.StatusForbidden, errAccessDenied)
		return
	}
	db.Client.NPC.DeleteOneID(id).Exec(c.Request.Context())
	WriteJSON(c, http.StatusOK, gin.H{"ok": true})
}

func LinkNPC(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var cn models.CharacterNPC
	if !BindOr400(c, &cn) {
		return
	}
	_, err := db.Client.CharacterNPC.Create().SetCharacterID(charID).SetNpcID(cn.NPCID).SetRelationship(cn.Relationship).SetNotes(cn.Notes).OnConflict(sql.ResolveWithNewValues()).ID(c.Request.Context())
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	WriteJSON(c, http.StatusOK, gin.H{"ok": true})
}

func UnlinkNPC(c *gin.Context) {
	linkID, _ := strconv.ParseInt(c.Param("nid"), 10, 64)
	cn, err := db.Client.CharacterNPC.Get(c.Request.Context(), linkID)
	if err != nil {
		WriteNotFound(c, "link not found")
		return
	}
	if !canEditCharacterID(c, cn.CharacterID) {
		WriteError(c, http.StatusForbidden, errAccessDenied)
		return
	}
	db.Client.CharacterNPC.DeleteOneID(linkID).Exec(c.Request.Context())
	WriteJSON(c, http.StatusOK, gin.H{"ok": true})
}

func GetCharacterNPCs(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	cnpcs, err := db.Client.CharacterNPC.Query().Where(characternpc.CharacterID(charID)).WithNpc().Order(ent.Desc(characternpc.FieldInteractionCount)).All(c.Request.Context())
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	type NPCLink struct {
		models.CharacterNPC
		NPCName string `json:"npc_name"`; NPCRace string `json:"npc_race"`; NPCRaceColor string `json:"npc_race_color,omitempty"`; NPCClass string `json:"npc_class"`; NPHPMax int `json:"npc_hp_max"`; NPHPCurr int `json:"npc_hp_current"`; NPCAlive bool `json:"npc_is_alive"`
	}
	var out = make([]NPCLink, 0)
	raceColors := GetRaceColorMap()
	for _, cn := range cnpcs {
		npcEnt := cn.Edges.Npc
		out = append(out, NPCLink{
			CharacterNPC: models.CharacterNPC{ID: cn.ID, CharacterID: cn.CharacterID, NPCID: cn.NpcID, Relationship: cn.Relationship, Notes: cn.Notes, InteractionCount: cn.InteractionCount, LastInteracted: cn.LastInteracted},
			NPCName: npcEnt.Name, NPCRace: npcEnt.Race, NPCRaceColor: raceColors[npcEnt.Race], NPCClass: npcEnt.Class, NPHPMax: npcEnt.HpMax, NPHPCurr: npcEnt.HpCurrent, NPCAlive: npcEnt.IsAlive,
		})
	}
	WriteJSON(c, http.StatusOK, out)
}

func LogNPCInteraction(c *gin.Context) {
	nid, _ := strconv.ParseInt(c.Param("nid"), 10, 64)
	cn, err := db.Client.CharacterNPC.Get(c.Request.Context(), nid)
	if err != nil {
		WriteNotFound(c, "NPC link not found")
		return
	}
	if !canEditCharacterID(c, cn.CharacterID) {
		WriteError(c, http.StatusForbidden, errAccessDenied)
		return
	}
	db.Client.CharacterNPC.UpdateOneID(nid).AddInteractionCount(1).SetLastInteracted(cTimeNow()).Save(c.Request.Context())
	WriteJSON(c, http.StatusOK, gin.H{"ok": true})
}

func cTimeNow() string { return time.Now().Format("2006-01-02 15:04:05") }
