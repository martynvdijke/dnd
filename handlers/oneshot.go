package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/ent"
	"villum/ent/oneshotact"
	"villum/ent/oneshotactnpc"
	"villum/models"
)

// ─── Helper Functions ───

func entActToModel(e *ent.OneShotAct) models.OneShotAct {
	m := models.OneShotAct{
		ID:               e.ID,
		AdventureID:      e.AdventureID,
		Number:           e.Number,
		SortOrder:        e.SortOrder,
		Title:            e.Title,
		Description:      e.Description,
		EstimatedMinutes: e.EstimatedMinutes,
		Notes:            e.Notes,
	}
	if e.ParentActID != 0 {
		pid := e.ParentActID
		m.ParentActID = &pid
	}
	if len(e.Edges.Scenes) > 0 {
		m.Scenes = make([]models.OneShotScene, 0)
		for _, s := range e.Edges.Scenes {
			ms := models.OneShotScene{
				ID: s.ID, ActID: s.ActID, Number: s.Number, SortOrder: s.SortOrder,
				Title: s.Title, Description: s.Description, SceneType: s.SceneType,
				EstimatedMinutes: s.EstimatedMinutes, Notes: s.Notes,
			}
			if s.LocationID != 0 { lid := s.LocationID; ms.LocationID = &lid }
			if s.EncounterID != 0 { eid := s.EncounterID; ms.EncounterID = &eid }
			m.Scenes = append(m.Scenes, ms)
		}
	}
	if len(e.Edges.Items) > 0 {
		m.Items = make([]models.OneShotItem, 0)
		for _, it := range e.Edges.Items {
			mi := models.OneShotItem{
				ID: it.ID, AdventureID: it.AdventureID, Name: it.Name,
				Description: it.Description, Category: it.Category, Quantity: it.Quantity,
				Weight: it.Weight, PriceGP: it.PriceGp, IsMagical: it.IsMagical,
				Attunement: it.Attunement, Notes: it.Notes, CreatedAt: it.CreatedAt,
			}
			if it.ActID != 0 { aid := it.ActID; mi.ActID = &aid }
			m.Items = append(m.Items, mi)
		}
	}
	if len(e.Edges.Encounters) > 0 {
		m.Encounters = make([]models.OneShotAdventureEncounter, 0)
		for _, enc := range e.Edges.Encounters {
			me := models.OneShotAdventureEncounter{
				ID: enc.ID, AdventureID: enc.AdventureID, EncounterID: enc.EncounterID,
			}
			if enc.ActID != 0 { aid := enc.ActID; me.ActID = &aid }
			m.Encounters = append(m.Encounters, me)
		}
	}
	return m
}

func entAdventureToModel(e *ent.OneShotAdventure) models.OneShotAdventure {
	m := models.OneShotAdventure{
		ID: e.ID, UserID: e.UserID, Title: e.Title, Premise: e.Premise, Hook: e.Hook,
		Template: e.Template, EstimatedMinutes: e.EstimatedMinutes, Difficulty: e.Difficulty,
		Notes: e.Notes, CreatedAt: e.CreatedAt, UpdatedAt: e.UpdatedAt,
	}
	if e.CampaignID != 0 { cid := e.CampaignID; m.CampaignID = &cid }
	return m
}

func sortActsBySortOrder(acts *[]models.OneShotAct) {
	sort.Slice(*acts, func(i, j int) bool {
		if (*acts)[i].SortOrder != (*acts)[j].SortOrder {
			return (*acts)[i].SortOrder < (*acts)[j].SortOrder
		}
		return (*acts)[i].Number < (*acts)[j].Number
	})
}

func loadAdventureNPCs(adventureID int64) []models.OneShotAdventureNPC {
	rows, err := db.DB.Query("SELECT oan.id, oan.adventure_id, oan.npc_id, oan.role, COALESCE(n.name,'') FROM oneshot_adventure_npcs oan LEFT JOIN npcs n ON oan.npc_id=n.id WHERE oan.adventure_id=?", adventureID)
	if err != nil { return nil }
	defer rows.Close()
	out := make([]models.OneShotAdventureNPC, 0)
	for rows.Next() {
		var npc models.OneShotAdventureNPC
		rows.Scan(&npc.ID, &npc.AdventureID, &npc.NPCID, &npc.Role, &npc.NPCName)
		out = append(out, npc)
	}
	return out
}

func loadAdventureLocations(adventureID int64) []models.OneShotAdventureLocation {
	rows, err := db.DB.Query("SELECT oal.id, oal.adventure_id, oal.location_id, COALESCE(l.name,'') FROM oneshot_adventure_locations oal LEFT JOIN locations l ON oal.location_id=l.id WHERE oal.adventure_id=?", adventureID)
	if err != nil { return nil }
	defer rows.Close()
	out := make([]models.OneShotAdventureLocation, 0)
	for rows.Next() {
		var loc models.OneShotAdventureLocation
		rows.Scan(&loc.ID, &loc.AdventureID, &loc.LocationID, &loc.LocationName)
		out = append(out, loc)
	}
	return out
}

func loadAdventureEncounters(adventureID int64) []models.OneShotAdventureEncounter {
	rows, err := db.DB.Query("SELECT oae.id, oae.adventure_id, oae.encounter_id, COALESCE(e.name,'') FROM oneshot_adventure_encounters oae LEFT JOIN encounter_templates e ON oae.encounter_id=e.id WHERE oae.adventure_id=?", adventureID)
	if err != nil { return nil }
	defer rows.Close()
	out := make([]models.OneShotAdventureEncounter, 0)
	for rows.Next() {
		var enc models.OneShotAdventureEncounter
		rows.Scan(&enc.ID, &enc.AdventureID, &enc.EncounterID, &enc.EncounterName)
		out = append(out, enc)
	}
	return out
}

func loadAdventureShops(adventureID int64) []models.OneShotShop {
	rows, err := db.DB.Query("SELECT id, user_id, campaign_id, oneshot_adventure_id, act_id, name, description, markup_percent, markup_buy_percent, created_at FROM shops WHERE oneshot_adventure_id=? ORDER BY name", adventureID)
	if err != nil { return nil }
	defer rows.Close()
	out := make([]models.OneShotShop, 0)
	for rows.Next() {
		var s models.OneShotShop
		var campID, aID sql.NullInt64
		rows.Scan(&s.ID, &s.UserID, &campID, &s.OneshotAdventureID, &aID, &s.Name, &s.Description, &s.MarkupPercent, &s.MarkupBuyPercent, &s.CreatedAt)
		if campID.Valid { s.CampaignID = &campID.Int64 }
		if aID.Valid { s.ActID = &aID.Int64 }
		out = append(out, s)
	}
	return out
}

func loadAdventureItems(adventureID int64) []models.OneShotItem {
	rows, err := db.DB.Query("SELECT id, adventure_id, act_id, name, description, category, quantity, weight, price_gp, is_magical, attunement, notes, created_at FROM oneshot_items WHERE adventure_id=? ORDER BY name", adventureID)
	if err != nil { return nil }
	defer rows.Close()
	out := make([]models.OneShotItem, 0)
	for rows.Next() {
		var it models.OneShotItem
		var isMag, att int
		var aID sql.NullInt64
		rows.Scan(&it.ID, &it.AdventureID, &aID, &it.Name, &it.Description, &it.Category, &it.Quantity, &it.Weight, &it.PriceGP, &isMag, &att, &it.Notes, &it.CreatedAt)
		it.IsMagical = isMag == 1
		it.Attunement = att == 1
		if aID.Valid { it.ActID = &aID.Int64 }
		out = append(out, it)
	}
	return out
}

func loadAdventureDetail(ctx context.Context, adventureID int64) (*models.OneShotAdventure, error) {
	entAdv, err := db.Client.OneShotAdventure.Get(ctx, adventureID)
	if err != nil {
		return nil, err
	}
	a := entAdventureToModel(entAdv)

	entActs, err := db.Client.OneShotAct.Query().
		Where(oneshotact.AdventureID(adventureID)).
		WithScenes().
		WithItems().
		WithEncounters().
		Order(oneshotact.BySortOrder(), oneshotact.ByID()).
		All(ctx)
	if err != nil {
		return nil, err
	}

	byParent := make(map[int64][]models.OneShotAct)
	for _, ea := range entActs {
		ma := entActToModel(ea)
		if ma.ParentActID != nil && *ma.ParentActID != 0 {
			byParent[*ma.ParentActID] = append(byParent[*ma.ParentActID], ma)
		}
	}

	var fillChildren func(act *models.OneShotAct)
	fillChildren = func(act *models.OneShotAct) {
		kids := byParent[act.ID]
		for i := range kids {
			fillChildren(&kids[i])
		}
		sortActsBySortOrder(&kids)
		act.Children = kids
	}

	a.Acts = make([]models.OneShotAct, 0)
	for _, ea := range entActs {
		if ea.ParentActID == 0 {
			ma := entActToModel(ea)
			fillChildren(&ma)
			a.Acts = append(a.Acts, ma)
		}
	}

	a.NPCs = loadAdventureNPCs(adventureID)
	a.Locations = loadAdventureLocations(adventureID)
	a.Encounters = loadAdventureEncounters(adventureID)

	return &a, nil
}

// ─── One-Shot Adventure API Handlers ───

func ListOneShotAdventures(c *gin.Context) {
	userID, _ := c.Get("user_id")
	campaignID := c.Query("campaign_id")

	var rows *sql.Rows
	var err error
	if campaignID != "" {
		rows, err = db.DB.Query("SELECT id, user_id, campaign_id, title, premise, hook, template, estimated_minutes, difficulty, notes, created_at, updated_at FROM oneshot_adventures WHERE campaign_id=? ORDER BY updated_at DESC", campaignID)
	} else {
		rows, err = db.DB.Query("SELECT id, user_id, campaign_id, title, premise, hook, template, estimated_minutes, difficulty, notes, created_at, updated_at FROM oneshot_adventures WHERE user_id=? ORDER BY updated_at DESC", userID)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	out := make([]models.OneShotAdventure, 0)
	for rows.Next() {
		var a models.OneShotAdventure
		rows.Scan(&a.ID, &a.UserID, &a.CampaignID, &a.Title, &a.Premise, &a.Hook, &a.Template, &a.EstimatedMinutes, &a.Difficulty, &a.Notes, &a.CreatedAt, &a.UpdatedAt)
		out = append(out, a)
	}
	c.JSON(http.StatusOK, out)
}

func CreateOneShotAdventure(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var a models.OneShotAdventure
	if err := c.ShouldBindJSON(&a); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := db.DB.Exec("INSERT INTO oneshot_adventures(user_id, campaign_id, title, premise, hook, template, estimated_minutes, difficulty, notes) VALUES(?,?,?,?,?,?,?,?,?)",
		userID, a.CampaignID, a.Title, a.Premise, a.Hook, a.Template, a.EstimatedMinutes, a.Difficulty, a.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()

	// Generate acts/scenes from template if applicable
	if a.Template == "five_room_dungeon" {
		generateFiveRoomDungeonStructure(id, a.Difficulty)
	} else if a.Template == "default" {
		generateDefaultStructure(id)
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func GetOneShotAdventure(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := c.Get("user_id")
	a, err := loadAdventureDetail(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "one-shot not found"})
		return
	}
	if a.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": "one-shot not found"})
		return
	}
	c.JSON(http.StatusOK, a)
}

func UpdateOneShotAdventure(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var a models.OneShotAdventure
	if err := c.ShouldBindJSON(&a); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Exec("UPDATE oneshot_adventures SET title=?, premise=?, hook=?, template=?, estimated_minutes=?, difficulty=?, notes=?, updated_at=datetime('now') WHERE id=?",
		a.Title, a.Premise, a.Hook, a.Template, a.EstimatedMinutes, a.Difficulty, a.Notes, id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteOneShotAdventure(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := c.Get("user_id")
	db.DB.Exec("DELETE FROM oneshot_adventures WHERE id=? AND user_id=?", id, userID)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ─── Acts ───

func CreateOneShotAct(c *gin.Context) {
	adventureID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var act models.OneShotAct
	if err := c.ShouldBindJSON(&act); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if act.Number == 0 {
		db.DB.QueryRow("SELECT COALESCE(MAX(number),0)+1 FROM oneshot_acts WHERE adventure_id=?", adventureID).Scan(&act.Number)
	}
	if act.SortOrder == 0 {
		db.DB.QueryRow("SELECT COALESCE(MAX(sort_order),0)+1 FROM oneshot_acts WHERE adventure_id=?", adventureID).Scan(&act.SortOrder)
	}
	ctx := c.Request.Context()
	q := db.Client.OneShotAct.Create().
		SetAdventureID(adventureID).
		SetNumber(act.Number).
		SetSortOrder(act.SortOrder).
		SetTitle(act.Title).
		SetDescription(act.Description).
		SetEstimatedMinutes(act.EstimatedMinutes).
		SetNotes(act.Notes)
	if act.ParentActID != nil && *act.ParentActID != 0 {
		q.SetParentActID(*act.ParentActID)
	}
	result, err := q.Save(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": result.ID})
}

func UpdateOneShotAct(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var act models.OneShotAct
	if err := c.ShouldBindJSON(&act); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := c.Request.Context()
	q := db.Client.OneShotAct.UpdateOneID(id).
		SetNumber(act.Number).
		SetSortOrder(act.SortOrder).
		SetTitle(act.Title).
		SetDescription(act.Description).
		SetEstimatedMinutes(act.EstimatedMinutes).
		SetNotes(act.Notes)
	if act.ParentActID != nil && *act.ParentActID != 0 {
		q.SetParentActID(*act.ParentActID)
	} else {
		q.ClearParentActID()
	}
	if _, err := q.Save(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteOneShotAct(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ctx := c.Request.Context()
	db.Client.OneShotAct.DeleteOneID(id).Exec(ctx)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ─── Scenes ───

func CreateOneShotScene(c *gin.Context) {
	actID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var sc models.OneShotScene
	if err := c.ShouldBindJSON(&sc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if sc.Number == 0 {
		db.DB.QueryRow("SELECT COALESCE(MAX(number),0)+1 FROM oneshot_scenes WHERE act_id=?", actID).Scan(&sc.Number)
	}
	if sc.SortOrder == 0 {
		db.DB.QueryRow("SELECT COALESCE(MAX(sort_order),0)+1 FROM oneshot_scenes WHERE act_id=?", actID).Scan(&sc.SortOrder)
	}
	ctx := c.Request.Context()
	result, err := db.Client.OneShotScene.Create().
		SetActID(actID).
		SetNumber(sc.Number).
		SetSortOrder(sc.SortOrder).
		SetTitle(sc.Title).
		SetDescription(sc.Description).
		SetSceneType(sc.SceneType).
		SetNillableLocationID(sc.LocationID).
		SetNillableEncounterID(sc.EncounterID).
		SetEstimatedMinutes(sc.EstimatedMinutes).
		SetNotes(sc.Notes).
		Save(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": result.ID})
}

func UpdateOneShotScene(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var sc models.OneShotScene
	if err := c.ShouldBindJSON(&sc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := c.Request.Context()
	_, err := db.Client.OneShotScene.UpdateOneID(id).
		SetNumber(sc.Number).
		SetSortOrder(sc.SortOrder).
		SetTitle(sc.Title).
		SetDescription(sc.Description).
		SetSceneType(sc.SceneType).
		SetNillableLocationID(sc.LocationID).
		SetNillableEncounterID(sc.EncounterID).
		SetEstimatedMinutes(sc.EstimatedMinutes).
		SetNotes(sc.Notes).
		Save(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteOneShotScene(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ctx := c.Request.Context()
	db.Client.OneShotScene.DeleteOneID(id).Exec(ctx)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ─── NPC Links ───

func GetOneShotNPCs(c *gin.Context) {
	adventureID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	rows, err := db.DB.Query("SELECT oan.id, oan.adventure_id, oan.npc_id, oan.role, COALESCE(n.name,'') FROM oneshot_adventure_npcs oan LEFT JOIN npcs n ON oan.npc_id=n.id WHERE oan.adventure_id=?", adventureID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := make([]models.OneShotAdventureNPC, 0)
	for rows.Next() {
		var npc models.OneShotAdventureNPC
		rows.Scan(&npc.ID, &npc.AdventureID, &npc.NPCID, &npc.Role, &npc.NPCName)
		out = append(out, npc)
	}
	c.JSON(http.StatusOK, out)
}

func LinkOneShotNPC(c *gin.Context) {
	adventureID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var link struct {
		NPCID int64  `json:"npc_id"`
		Role  string `json:"role"`
	}
	if err := c.ShouldBindJSON(&link); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := db.DB.Exec("INSERT OR IGNORE INTO oneshot_adventure_npcs(adventure_id, npc_id, role) VALUES(?,?,?)", adventureID, link.NPCID, link.Role)
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

type GenerateRequest struct {
	Template     string `json:"template"`
	Title        string `json:"title"`
	Difficulty   string `json:"difficulty"`
	Minutes      int    `json:"estimated_minutes"`
	CampaignID   *int64 `json:"campaign_id,omitempty"`
}

func GenerateOneShotFromTemplate(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Minutes <= 0 {
		req.Minutes = 180
	}
	if req.Difficulty == "" {
		req.Difficulty = "medium"
	}
	if req.Title == "" {
		req.Title = "Untitled One-Shot"
	}

	var premise, hook string

	switch req.Template {
	case "five_room_dungeon":
		premise, hook = generateFiveRoomDungeon(req.Difficulty)
	default:
		premise = "A new adventure begins..."
		hook = "The party is called to action."
	}

	// Create the adventure
	now := time.Now().Format("2006-01-02 15:04:05")
	result, err := db.DB.Exec("INSERT INTO oneshot_adventures(user_id, campaign_id, title, premise, hook, template, estimated_minutes, difficulty, notes, created_at, updated_at) VALUES(?,?,?,?,?,?,?,?,'',?,?)",
		userID, req.CampaignID, req.Title, premise, hook, req.Template, req.Minutes, req.Difficulty, now, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	adventureID, _ := result.LastInsertId()

	// Generate acts and scenes based on template
	switch req.Template {
	case "five_room_dungeon":
		generateFiveRoomDungeonStructure(adventureID, req.Difficulty)
	default:
		generateDefaultStructure(adventureID)
	}

	c.JSON(http.StatusCreated, gin.H{"id": adventureID})
}

func generateFiveRoomDungeon(difficulty string) (string, string) {
	hooks := map[string][]string{
		"easy": {
			"Strange mushrooms have been sprouting in the cellar of the local inn.",
			"Kobolds have been stealing livestock from nearby farms.",
			"A mysterious fog rolls in every night, carrying strange whispers.",
		},
		"medium": {
			"Cultists have taken over an ancient temple in the forest.",
			"A powerful artifact was stolen from the duke's vault.",
			"Goblins are amassing an army in the abandoned mines.",
		},
		"hard": {
			"A dragon's lair has been discovered beneath the mountain.",
			"An evil wizard is conducting dark rituals in a fallen keep.",
			"Undead legions are rising from an ancient battlefield.",
		},
		"deadly": {
			"A lich is gathering the last surviving members of an ancient order.",
			"The boundaries between planes are weakening in a cursed valley.",
			"A demonic incursion threatens to consume the realm.",
		},
	}

	premises := map[string][]string{
		"easy": {
			"A local crisis requires brave adventurers to investigate strange occurrences.",
			"The nearby village needs help with a growing monster problem.",
		},
		"medium": {
			"A villain's plot threatens the region, and only a daring group can stop it.",
			"An ancient evil stirs, and the party must uncover its source.",
		},
		"hard": {
			"A catastrophic threat looms over the land, requiring heroes of great courage.",
			"The realm faces its greatest danger in centuries.",
		},
		"deadly": {
			"The end of all things approaches. The party must make a stand against impossible odds.",
			"A world-ending threat emerges, and the fate of civilization hangs in the balance.",
		},
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	hooksList := hooks[difficulty]
	premisesList := premises[difficulty]

	return premisesList[r.Intn(len(premisesList))], hooksList[r.Intn(len(hooksList))]
}

func generateFiveRoomDungeonStructure(adventureID int64, difficulty string) {
	rooms := []struct {
		title       string
		description string
		sceneType   string
		minutes     int
	}{
		{
			title:       "Entrance & Guardian",
			description: "The party approaches the entrance. A guardian or obstacle blocks the way, testing their readiness.",
			sceneType:   "exploration",
			minutes:     20,
		},
		{
			title:       "Puzzle or Roleplaying Challenge",
			description: "A puzzle, trap, or social encounter that requires wit rather than combat to overcome.",
			sceneType:   "puzzle",
			minutes:     25,
		},
		{
			title:       "Trick or Setback",
			description: "A trick, trap, or setback that drains resources or misleads the party.",
			sceneType:   "exploration",
			minutes:     15,
		},
		{
			title:       "Climax - Big Battle",
			description: "The main confrontation. The party faces the biggest threat in the dungeon.",
			sceneType:   "combat",
			minutes:     30,
		},
		{
			title:       "Reward & Revelation",
			description: "The aftermath. Treasure, clues about the larger story, and a chance to catch their breath.",
			sceneType:   "roleplay",
			minutes:     15,
		},
	}

	diffMult := map[string]int{"easy": 1, "medium": 2, "hard": 3, "deadly": 4}

	for i, room := range rooms {
		// Create act for each room
		actResult, err := db.DB.Exec("INSERT INTO oneshot_acts(adventure_id, number, title, description, estimated_minutes) VALUES(?,?,?,?,?)",
			adventureID, i+1, room.title, room.description, room.minutes*diffMult[difficulty])
		if err != nil {
			continue
		}
		actID, _ := actResult.LastInsertId()

		// Create scene within act
		db.DB.Exec("INSERT INTO oneshot_scenes(act_id, number, title, description, scene_type, estimated_minutes, notes) VALUES(?,1,?,?,?,?,'')",
			actID, room.title, room.description, room.sceneType, room.minutes*diffMult[difficulty])
	}
}

func generateDefaultStructure(adventureID int64) {
	acts := []struct {
		title       string
		description string
		minutes     int
		scenes      []struct {
			title       string
			description string
			sceneType   string
			minutes     int
		}
	}{
		{
			title:       "Setup",
			description: "The party receives the call to adventure and gathers information.",
			minutes:     45,
			scenes: []struct {
				title       string
				description string
				sceneType   string
				minutes     int
			}{
				{"The Hook", "The party learns about the situation and decides to act.", "roleplay", 20},
				{"Preparation", "The party gathers supplies, information, and allies.", "exploration", 25},
			},
		},
		{
			title:       "Complication",
			description: "Obstacles arise and the party faces challenges.",
			minutes:     60,
			scenes: []struct {
				title       string
				description string
				sceneType   string
				minutes     int
			}{
				{"First Obstacle", "A challenge blocks the party's path.", "exploration", 20},
				{"Twist", "A unexpected development changes the situation.", "roleplay", 15},
				{"Rising Tension", "The stakes increase as the party pushes forward.", "combat", 25},
			},
		},
		{
			title:       "Climax",
			description: "The final confrontation and resolution.",
			minutes:     45,
			scenes: []struct {
				title       string
				description string
				sceneType   string
				minutes     int
			}{
				{"Final Confrontation", "The party faces the main threat.", "climax", 30},
				{"Resolution", "The aftermath and celebration.", "roleplay", 15},
			},
		},
	}

	for i, act := range acts {
		actResult, err := db.DB.Exec("INSERT INTO oneshot_acts(adventure_id, number, title, description, estimated_minutes) VALUES(?,?,?,?,?)",
			adventureID, i+1, act.title, act.description, act.minutes)
		if err != nil {
			continue
		}
		actID, _ := actResult.LastInsertId()
		for j, sc := range act.scenes {
			db.DB.Exec("INSERT INTO oneshot_scenes(act_id, number, title, description, scene_type, estimated_minutes, notes) VALUES(?,?,?,?,?,?,'')",
				actID, j+1, sc.title, sc.description, sc.sceneType, sc.minutes)
		}
	}
}

// ─── HTMX Handlers ───

type htmxOneShotData struct {
	Adventure  *models.OneShotAdventure
	Adventures []models.OneShotAdventure
	NPCs       []models.NPC
	Locations  []models.Location
	Encounters []models.EncounterTemplate
	Act        *models.OneShotAct
	Scene      *models.OneShotScene
	SceneTypes []string
	Templates  []string
	Difficulties []string
	Acts       []models.OneShotAct
}

func HtmxListOneShots(c *gin.Context) {
	userID, _ := c.Get("user_id")
	campaignID := c.Query("campaign_id")

	var rows *sql.Rows
	var err error
	if campaignID != "" {
		rows, err = db.DB.Query("SELECT id, user_id, campaign_id, title, premise, hook, template, estimated_minutes, difficulty, notes, created_at, updated_at FROM oneshot_adventures WHERE campaign_id=? ORDER BY updated_at DESC", campaignID)
	} else {
		rows, err = db.DB.Query("SELECT id, user_id, campaign_id, title, premise, hook, template, estimated_minutes, difficulty, notes, created_at, updated_at FROM oneshot_adventures WHERE user_id=? ORDER BY updated_at DESC", userID)
	}
	if err != nil {
		c.String(http.StatusInternalServerError, "query error: %v", err)
		return
	}
	defer rows.Close()

	out := make([]models.OneShotAdventure, 0)
	for rows.Next() {
		var a models.OneShotAdventure
		rows.Scan(&a.ID, &a.UserID, &a.CampaignID, &a.Title, &a.Premise, &a.Hook, &a.Template, &a.EstimatedMinutes, &a.Difficulty, &a.Notes, &a.CreatedAt, &a.UpdatedAt)
		out = append(out, a)
	}
	renderTemplate(c, "oneshot_list.html", htmxOneShotData{Adventures: out})
}

func HtmxNewOneShotForm(c *gin.Context) {
	campaignID := c.Query("campaign_id")
	data := htmxOneShotData{
		SceneTypes:   []string{"roleplay", "combat", "exploration", "puzzle", "climax"},
		Templates:    []string{"custom", "five_room_dungeon"},
		Difficulties: []string{"easy", "medium", "hard", "deadly"},
	}
	if campaignID != "" {
		if cid, err := strconv.ParseInt(campaignID, 10, 64); err == nil {
			data.Adventure = &models.OneShotAdventure{CampaignID: &cid}
		}
	}
	renderTemplate(c, "oneshot_form.html", data)
}

func HtmxEditOneShotForm(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := c.Get("user_id")

	var a models.OneShotAdventure
	err := db.DB.QueryRow("SELECT id, user_id, campaign_id, title, premise, hook, template, estimated_minutes, difficulty, notes, created_at, updated_at FROM oneshot_adventures WHERE id=? AND user_id=?", id, userID).
		Scan(&a.ID, &a.UserID, &a.CampaignID, &a.Title, &a.Premise, &a.Hook, &a.Template, &a.EstimatedMinutes, &a.Difficulty, &a.Notes, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		c.String(http.StatusNotFound, "not found")
		return
	}
	renderTemplate(c, "oneshot_form.html", htmxOneShotData{
		Adventure:    &a,
		SceneTypes:   []string{"roleplay", "combat", "exploration", "puzzle", "climax"},
		Templates:    []string{"custom", "five_room_dungeon"},
		Difficulties: []string{"easy", "medium", "hard", "deadly"},
	})
}

func HtmxGetOneShotDetail(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	a, err := loadAdventureDetail(c.Request.Context(), id)
	if err != nil {
		c.String(http.StatusNotFound, "not found")
		return
	}
	renderTemplate(c, "oneshot_detail.html", htmxOneShotData{Adventure: a})
}

func HtmxCreateOneShot(c *gin.Context) {
	userID, _ := c.Get("user_id")
	title := c.PostForm("title")
	premise := c.PostForm("premise")
	hook := c.PostForm("hook")
	template := c.PostForm("template")
	difficulty := c.PostForm("difficulty")
	minutes, _ := strconv.Atoi(c.PostForm("estimated_minutes"))
	notes := c.PostForm("notes")
	campaignStr := c.PostForm("campaign_id")

	if title == "" {
		title = "Untitled One-Shot"
	}
	if template == "" {
		template = "custom"
	}
	if difficulty == "" {
		difficulty = "medium"
	}
	if minutes <= 0 {
		minutes = 180
	}

	var campaignID *int64
	if campaignStr != "" {
		if cid, err := strconv.ParseInt(campaignStr, 10, 64); err == nil {
			campaignID = &cid
		}
	}

	result, err := db.DB.Exec("INSERT INTO oneshot_adventures(user_id, campaign_id, title, premise, hook, template, estimated_minutes, difficulty, notes) VALUES(?,?,?,?,?,?,?,?,?)",
		userID, campaignID, title, premise, hook, template, minutes, difficulty, notes)
	if err != nil {
		c.String(http.StatusInternalServerError, "insert error: %v", err)
		return
	}
	id, _ := result.LastInsertId()

	// If using a template, generate structure
	if template == "five_room_dungeon" {
		generateFiveRoomDungeonStructure(id, difficulty)
	} else if template == "custom" && premise != "" {
		generateDefaultStructure(id)
	}

	// Re-render the list
	HtmxListOneShots(c)
}

func HtmxUpdateOneShot(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	title := c.PostForm("title")
	premise := c.PostForm("premise")
	hook := c.PostForm("hook")
	template := c.PostForm("template")
	difficulty := c.PostForm("difficulty")
	minutes, _ := strconv.Atoi(c.PostForm("estimated_minutes"))
	notes := c.PostForm("notes")

	db.DB.Exec("UPDATE oneshot_adventures SET title=?, premise=?, hook=?, template=?, estimated_minutes=?, difficulty=?, notes=?, updated_at=datetime('now') WHERE id=?",
		title, premise, hook, template, minutes, difficulty, notes, id)

	HtmxListOneShots(c)
}

func HtmxDeleteOneShot(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := c.Get("user_id")
	db.DB.Exec("DELETE FROM oneshot_adventures WHERE id=? AND user_id=?", id, userID)
	HtmxListOneShots(c)
}

// HTMX Act form
func HtmxNewActForm(c *gin.Context) {
	adventureID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ctx := c.Request.Context()

	ents, err := db.Client.OneShotAct.Query().
		Where(oneshotact.AdventureID(adventureID)).
		Order(oneshotact.BySortOrder(), oneshotact.ByID()).
		All(ctx)
	if err != nil {
		c.String(http.StatusInternalServerError, "query error: %v", err)
		return
	}

	acts := make([]models.OneShotAct, 0)
	for _, e := range ents {
		acts = append(acts, entActToModel(e))
	}

	a, err := loadAdventureDetail(ctx, adventureID)
	if err != nil {
		c.String(http.StatusNotFound, "not found")
		return
	}

	data := htmxOneShotData{
		Adventure: a,
		Act:       &models.OneShotAct{AdventureID: adventureID, Number: len(acts) + 1, SortOrder: len(acts) + 1, EstimatedMinutes: 30},
		Acts:      acts,
	}
	renderTemplate(c, "oneshot_act_form.html", data)
}

func HtmxSceneForm(c *gin.Context) {
	actID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var adventureID int64
	db.DB.QueryRow("SELECT adventure_id FROM oneshot_acts WHERE id=?", actID).Scan(&adventureID)

	data := htmxOneShotData{
		Scene:      &models.OneShotScene{ActID: actID, SceneType: "roleplay", EstimatedMinutes: 15},
		SceneTypes: []string{"roleplay", "combat", "exploration", "puzzle", "climax"},
	}
	renderTemplate(c, "oneshot_scene_form.html", data)
}

// HTMX Act handlers
func HtmxCreateAct(c *gin.Context) {
	adventureID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	title := c.PostForm("title")
	description := c.PostForm("description")
	minutes, _ := strconv.Atoi(c.PostForm("estimated_minutes"))
	notes := c.PostForm("notes")
	parentActStr := c.PostForm("parent_act_id")

	if title == "" {
		title = "New Act"
	}
	if minutes <= 0 {
		minutes = 30
	}

	var number, sortOrder int
	db.DB.QueryRow("SELECT COALESCE(MAX(number),0)+1 FROM oneshot_acts WHERE adventure_id=?", adventureID).Scan(&number)
	db.DB.QueryRow("SELECT COALESCE(MAX(sort_order),0)+1 FROM oneshot_acts WHERE adventure_id=?", adventureID).Scan(&sortOrder)

	ctx := c.Request.Context()
	q := db.Client.OneShotAct.Create().
		SetAdventureID(adventureID).
		SetNumber(number).
		SetSortOrder(sortOrder).
		SetTitle(title).
		SetDescription(description).
		SetEstimatedMinutes(minutes).
		SetNotes(notes)
	if parentActStr != "" {
		if pid, err := strconv.ParseInt(parentActStr, 10, 64); err == nil {
			q.SetParentActID(pid)
		}
	}
	if _, err := q.Save(ctx); err != nil {
		c.String(http.StatusInternalServerError, "insert error: %v", err)
		return
	}

	HtmxGetOneShotDetail(c)
}

func HtmxUpdateAct(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	title := c.PostForm("title")
	description := c.PostForm("description")
	minutes, _ := strconv.Atoi(c.PostForm("estimated_minutes"))
	number, _ := strconv.Atoi(c.PostForm("number"))
	sortOrder, _ := strconv.Atoi(c.PostForm("sort_order"))
	notes := c.PostForm("notes")

	if number <= 0 {
		db.DB.QueryRow("SELECT number FROM oneshot_acts WHERE id=?", id).Scan(&number)
	}

	ctx := c.Request.Context()
	q := db.Client.OneShotAct.UpdateOneID(id).
		SetNumber(number).
		SetTitle(title).
		SetDescription(description).
		SetEstimatedMinutes(minutes).
		SetNotes(notes)
	if sortOrder > 0 {
		q.SetSortOrder(sortOrder)
	}
	if _, err := q.Save(ctx); err != nil {
		c.String(http.StatusInternalServerError, "update error: %v", err)
		return
	}

	HtmxGetOneShotDetail(c)
}
func HtmxDeleteAct(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var adventureID int64
	db.DB.QueryRow("SELECT adventure_id FROM oneshot_acts WHERE id=?", id).Scan(&adventureID)
	db.DB.Exec("DELETE FROM oneshot_acts WHERE id=?", id)

	ReRenderOneShotDetail(c, adventureID)
}

// HTMX Scene handlers
func HtmxCreateScene(c *gin.Context) {
	actID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	title := c.PostForm("title")
	description := c.PostForm("description")
	sceneType := c.PostForm("scene_type")
	minutes, _ := strconv.Atoi(c.PostForm("estimated_minutes"))
	notes := c.PostForm("notes")

	if title == "" {
		title = "New Scene"
	}
	if sceneType == "" {
		sceneType = "roleplay"
	}
	if minutes <= 0 {
		minutes = 15
	}

	var number, sortOrder int
	db.DB.QueryRow("SELECT COALESCE(MAX(number),0)+1 FROM oneshot_scenes WHERE act_id=?", actID).Scan(&number)
	db.DB.QueryRow("SELECT COALESCE(MAX(sort_order),0)+1 FROM oneshot_scenes WHERE act_id=?", actID).Scan(&sortOrder)

	ctx := c.Request.Context()
	_, err := db.Client.OneShotScene.Create().
		SetActID(actID).
		SetNumber(number).
		SetSortOrder(sortOrder).
		SetTitle(title).
		SetDescription(description).
		SetSceneType(sceneType).
		SetEstimatedMinutes(minutes).
		SetNotes(notes).
		Save(ctx)
	if err != nil {
		c.String(http.StatusInternalServerError, "insert error: %v", err)
		return
	}

	HtmxGetOneShotDetail(c)
}
func HtmxUpdateScene(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	title := c.PostForm("title")
	description := c.PostForm("description")
	sceneType := c.PostForm("scene_type")
	minutes, _ := strconv.Atoi(c.PostForm("estimated_minutes"))
	notes := c.PostForm("notes")
	number, _ := strconv.Atoi(c.PostForm("number"))
	sortOrder, _ := strconv.Atoi(c.PostForm("sort_order"))

	if number <= 0 {
		db.DB.QueryRow("SELECT number FROM oneshot_scenes WHERE id=?", id).Scan(&number)
	}

	ctx := c.Request.Context()
	q := db.Client.OneShotScene.UpdateOneID(id).
		SetNumber(number).
		SetTitle(title).
		SetDescription(description).
		SetSceneType(sceneType).
		SetEstimatedMinutes(minutes).
		SetNotes(notes)
	if sortOrder > 0 {
		q.SetSortOrder(sortOrder)
	}
	if _, err := q.Save(ctx); err != nil {
		c.String(http.StatusInternalServerError, "update error: %v", err)
		return
	}

	HtmxGetOneShotDetail(c)
}
func HtmxDeleteScene(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var adventureID int64
	db.DB.QueryRow("SELECT oa.adventure_id FROM oneshot_acts oa JOIN oneshot_scenes s ON s.act_id=oa.id WHERE s.id=?", id).Scan(&adventureID)
	db.DB.Exec("DELETE FROM oneshot_scenes WHERE id=?", id)

	ReRenderOneShotDetail(c, adventureID)
}

func ReRenderOneShotDetail(c *gin.Context, adventureID int64) {
	a, err := loadAdventureDetail(c.Request.Context(), adventureID)
	if err != nil {
		c.String(http.StatusNotFound, "not found")
		return
	}
	renderTemplate(c, "oneshot_detail.html", htmxOneShotData{Adventure: a})
}

// HTMX generate from template
func HtmxGenerateOneShot(c *gin.Context) {
	userID, _ := c.Get("user_id")
	title := c.PostForm("title")
	template := c.PostForm("template")
	difficulty := c.PostForm("difficulty")
	minutes, _ := strconv.Atoi(c.PostForm("estimated_minutes"))
	campaignStr := c.PostForm("campaign_id")

	if title == "" {
		title = "Untitled One-Shot"
	}
	if template == "" {
		template = "five_room_dungeon"
	}
	if difficulty == "" {
		difficulty = "medium"
	}
	if minutes <= 0 {
		minutes = 180
	}

	var campaignID *int64
	if campaignStr != "" {
		if cid, err := strconv.ParseInt(campaignStr, 10, 64); err == nil {
			campaignID = &cid
		}
	}

	var premise, hook string
	switch template {
	case "five_room_dungeon":
		premise, hook = generateFiveRoomDungeon(difficulty)
	default:
		premise = "A new adventure begins..."
		hook = "The party is called to action."
	}

	result, err := db.DB.Exec("INSERT INTO oneshot_adventures(user_id, campaign_id, title, premise, hook, template, estimated_minutes, difficulty, notes) VALUES(?,?,?,?,?,?,?,?,'')",
		userID, campaignID, title, premise, hook, template, minutes, difficulty)
	if err != nil {
		c.String(http.StatusInternalServerError, "insert error: %v", err)
		return
	}
	id, _ := result.LastInsertId()

	if template == "five_room_dungeon" {
		generateFiveRoomDungeonStructure(id, difficulty)
	} else {
		generateDefaultStructure(id)
	}

	HtmxListOneShots(c)
}

// ─── Session Pacing API Handlers ───

func StartPacingSession(c *gin.Context) {
	adventureID := c.Param("id")
	userID, _ := c.Get("user_id")

	// Verify the adventure belongs to this user
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM oneshot_adventures WHERE id=? AND user_id=?)", adventureID, userID).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "adventure not found"})
		return
	}

	// Check if there's already an active session
	var existingID int64
	err = db.DB.QueryRow("SELECT id FROM session_pacing WHERE adventure_id=? AND status IN ('running','paused')", adventureID).Scan(&existingID)
	if err == nil {
		c.JSON(http.StatusOK, gin.H{"id": existingID, "message": "resumed existing session"})
		return
	}

	// Get first act and scene
	var firstActID, firstSceneID *int64
	var actID int64
	err = db.DB.QueryRow("SELECT id FROM oneshot_acts WHERE adventure_id=? ORDER BY number ASC LIMIT 1", adventureID).Scan(&actID)
	if err == nil {
		firstActID = &actID
		var sceneID int64
		err = db.DB.QueryRow("SELECT id FROM oneshot_scenes WHERE act_id=? ORDER BY number ASC LIMIT 1", actID).Scan(&sceneID)
		if err == nil {
			firstSceneID = &sceneID
		}
	}

	result, err := db.DB.Exec(
		"INSERT INTO session_pacing(adventure_id, current_act_id, current_scene_id, status, elapsed_seconds) VALUES(?,?,?,'running',0)",
		adventureID, firstActID, firstSceneID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()

	// If first scene exists, create initial scene timing
	if firstSceneID != nil {
		db.DB.Exec("INSERT INTO scene_timings(session_id, scene_id, status) VALUES(?,?,'active')", id, *firstSceneID)
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func GetPacingSession(c *gin.Context) {
	sessionID := c.Param("id")

	var s models.SessionPacing
	err := db.DB.QueryRow(`
		SELECT sp.id, sp.adventure_id, sp.current_act_id, sp.current_scene_id, sp.status, sp.elapsed_seconds, sp.started_at, COALESCE(sp.completed_at,''),
			COALESCE(oa.title,''), COALESCE(a.title,''), COALESCE(sc.title,''), COALESCE(sc.estimated_minutes,0),
			COALESCE(a.number,0), COALESCE(sc.number,0)
		FROM session_pacing sp
		LEFT JOIN oneshot_adventures oa ON oa.id = sp.adventure_id
		LEFT JOIN oneshot_acts a ON a.id = sp.current_act_id
		LEFT JOIN oneshot_scenes sc ON sc.id = sp.current_scene_id
		WHERE sp.id=?
	`, sessionID).Scan(&s.ID, &s.AdventureID, &s.CurrentActID, &s.CurrentSceneID, &s.Status, &s.ElapsedSeconds, &s.StartedAt, &s.CompletedAt,
		&s.AdventureTitle, &s.ActTitle, &s.SceneTitle, &s.SceneEstimated, &s.ActNumber, &s.SceneNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get total counts
	db.DB.QueryRow("SELECT COUNT(*) FROM oneshot_acts WHERE adventure_id=?", s.AdventureID).Scan(&s.TotalActs)
	db.DB.QueryRow("SELECT COUNT(*) FROM oneshot_scenes WHERE act_id=?", s.CurrentActID).Scan(&s.TotalScenes)

	// Get scene timings
	rows, err := db.DB.Query(`
		SELECT st.id, st.session_id, st.scene_id, st.elapsed_seconds, st.status, COALESCE(st.started_at,''), COALESCE(st.completed_at,''),
			COALESCE(sc.title,''), COALESCE(sc.scene_type,''), COALESCE(sc.estimated_minutes,0)
		FROM scene_timings st
		LEFT JOIN oneshot_scenes sc ON sc.id = st.scene_id
		WHERE st.session_id=?
		ORDER BY st.id
	`, sessionID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var st models.SceneTiming
			if err := rows.Scan(&st.ID, &st.SessionID, &st.SceneID, &st.ElapsedSeconds, &st.Status, &st.StartedAt, &st.CompletedAt,
				&st.SceneTitle, &st.SceneType, &st.EstimatedMin); err == nil {
				s.SceneTimings = append(s.SceneTimings, st)
			}
		}
	}

	c.JSON(http.StatusOK, s)
}

func UpdatePacingTimers(c *gin.Context) {
	sessionID := c.Param("id")

	// Increment elapsed_seconds for the session
	db.DB.Exec("UPDATE session_pacing SET elapsed_seconds = elapsed_seconds + 5 WHERE id=? AND status='running'", sessionID)

	// Increment current scene timing
	db.DB.Exec("UPDATE scene_timings SET elapsed_seconds = elapsed_seconds + 5 WHERE session_id=? AND status='active'", sessionID)

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func PausePacingSession(c *gin.Context) {
	sessionID := c.Param("id")
	_, err := db.DB.Exec("UPDATE session_pacing SET status='paused' WHERE id=? AND status='running'", sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "paused"})
}

func ResumePacingSession(c *gin.Context) {
	sessionID := c.Param("id")
	_, err := db.DB.Exec("UPDATE session_pacing SET status='running' WHERE id=? AND status='paused'", sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "running"})
}

func CompletePacingSession(c *gin.Context) {
	sessionID := c.Param("id")

	// Mark current scene timing as completed
	db.DB.Exec("UPDATE scene_timings SET status='completed', completed_at=datetime('now') WHERE session_id=? AND status='active'", sessionID)

	// Mark session as completed
	_, err := db.DB.Exec("UPDATE session_pacing SET status='completed', completed_at=datetime('now') WHERE id=? AND status IN ('running','paused')", sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "completed"})
}

func AdvanceToNextScene(c *gin.Context) {
	sessionID := c.Param("id")

	// Get current scene info
	var currentActID, currentSceneID int64
	err := db.DB.QueryRow("SELECT COALESCE(current_act_id,0), COALESCE(current_scene_id,0) FROM session_pacing WHERE id=?", sessionID).Scan(&currentActID, &currentSceneID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	if currentSceneID > 0 {
		// Mark current scene timing as completed
		db.DB.Exec(`UPDATE scene_timings SET status='completed', elapsed_seconds=COALESCE((SELECT elapsed_seconds FROM scene_timings WHERE session_id=? AND scene_id=? AND status='active' LIMIT 1), elapsed_seconds), completed_at=datetime('now') WHERE session_id=? AND scene_id=? AND status='active'`,
			sessionID, currentSceneID, sessionID, currentSceneID)
	}

	// If no current act/scene, try to set first scene from adventure
	if currentActID == 0 && currentSceneID == 0 {
		var advID int64
		db.DB.QueryRow("SELECT adventure_id FROM session_pacing WHERE id=?", sessionID).Scan(&advID)
		var firstActID int64
		err = db.DB.QueryRow("SELECT id FROM oneshot_acts WHERE adventure_id=? ORDER BY number ASC LIMIT 1", advID).Scan(&firstActID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no acts in adventure"})
			return
		}
		var firstSceneID int64
		err = db.DB.QueryRow("SELECT id FROM oneshot_scenes WHERE act_id=? ORDER BY number ASC LIMIT 1", firstActID).Scan(&firstSceneID)
		if err != nil {
			// Act has no scenes, advance to next act later
			db.DB.Exec("UPDATE session_pacing SET current_act_id=? WHERE id=?", firstActID, sessionID)
			currentActID = firstActID
		} else {
			db.DB.Exec("UPDATE session_pacing SET current_act_id=?, current_scene_id=? WHERE id=?", firstActID, firstSceneID, sessionID)
			db.DB.Exec("INSERT INTO scene_timings(session_id, scene_id, status) VALUES(?,?,'active')", sessionID, firstSceneID)
			c.JSON(http.StatusOK, gin.H{"status": "advanced"})
			return
		}
	}

	// Find next scene in same act
	if currentActID > 0 && currentSceneID > 0 {
		var nextSceneID int64
		err = db.DB.QueryRow("SELECT id FROM oneshot_scenes WHERE act_id=? AND number > (SELECT number FROM oneshot_scenes WHERE id=?) ORDER BY number ASC LIMIT 1",
			currentActID, currentSceneID).Scan(&nextSceneID)
		if err == nil {
			db.DB.Exec("UPDATE session_pacing SET current_scene_id=? WHERE id=?", nextSceneID, sessionID)
			db.DB.Exec("INSERT INTO scene_timings(session_id, scene_id, status) VALUES(?,?,'active')", sessionID, nextSceneID)
			c.JSON(http.StatusOK, gin.H{"status": "advanced"})
			return
		}
	}

	// No more scenes in this act, find next act
	if currentActID > 0 {
		var nextActID int64
		err = db.DB.QueryRow("SELECT id FROM oneshot_acts WHERE adventure_id=(SELECT adventure_id FROM session_pacing WHERE id=?) AND number > (SELECT number FROM oneshot_acts WHERE id=?) ORDER BY number ASC LIMIT 1",
			sessionID, currentActID).Scan(&nextActID)
		if err == nil {
			var firstSceneID int64
			err = db.DB.QueryRow("SELECT id FROM oneshot_scenes WHERE act_id=? ORDER BY number ASC LIMIT 1", nextActID).Scan(&firstSceneID)
			if err == nil {
				db.DB.Exec("UPDATE session_pacing SET current_act_id=?, current_scene_id=? WHERE id=?", nextActID, firstSceneID, sessionID)
				db.DB.Exec("INSERT INTO scene_timings(session_id, scene_id, status) VALUES(?,?,'active')", sessionID, firstSceneID)
				c.JSON(http.StatusOK, gin.H{"status": "advanced"})
				return
			}
			// Act has no scenes, advance to next act without scenes
			currentActID = nextActID
			db.DB.Exec("UPDATE session_pacing SET current_act_id=?, current_scene_id=NULL WHERE id=?", nextActID, sessionID)
		}
	}

	// No more acts - complete session
	db.DB.Exec("UPDATE session_pacing SET status='completed', completed_at=datetime('now') WHERE id=?", sessionID)
	c.JSON(http.StatusOK, gin.H{"status": "completed", "message": "all scenes completed"})
}

// ─── HTMX Pacing Handlers ───

func HtmxGetPacingDashboard(c *gin.Context) {
	sessionID := c.Param("id")

	var s models.SessionPacing
	err := db.DB.QueryRow(`
		SELECT sp.id, sp.adventure_id, sp.current_act_id, sp.current_scene_id, sp.status, sp.elapsed_seconds, sp.started_at, COALESCE(sp.completed_at,''),
			COALESCE(oa.title,''), COALESCE(a.title,''), COALESCE(sc.title,''), COALESCE(sc.estimated_minutes,0),
			COALESCE(a.number,0), COALESCE(sc.number,0)
		FROM session_pacing sp
		LEFT JOIN oneshot_adventures oa ON oa.id = sp.adventure_id
		LEFT JOIN oneshot_acts a ON a.id = sp.current_act_id
		LEFT JOIN oneshot_scenes sc ON sc.id = sp.current_scene_id
		WHERE sp.id=?
	`, sessionID).Scan(&s.ID, &s.AdventureID, &s.CurrentActID, &s.CurrentSceneID, &s.Status, &s.ElapsedSeconds, &s.StartedAt, &s.CompletedAt,
		&s.AdventureTitle, &s.ActTitle, &s.SceneTitle, &s.SceneEstimated, &s.ActNumber, &s.SceneNumber)
	if err != nil {
		c.String(http.StatusNotFound, "Session not found")
		return
	}

	db.DB.QueryRow("SELECT COUNT(*) FROM oneshot_acts WHERE adventure_id=?", s.AdventureID).Scan(&s.TotalActs)
	var sceneCount int
	db.DB.QueryRow("SELECT COUNT(*) FROM oneshot_scenes sc JOIN oneshot_acts a ON a.id=sc.act_id WHERE a.adventure_id=?", s.AdventureID).Scan(&sceneCount)
	_ = sceneCount

	rows, err := db.DB.Query(`
		SELECT st.id, st.session_id, st.scene_id, st.elapsed_seconds, st.status, COALESCE(st.started_at,''), COALESCE(st.completed_at,''),
			COALESCE(sc.title,''), COALESCE(sc.scene_type,''), COALESCE(sc.estimated_minutes,0)
		FROM scene_timings st
		LEFT JOIN oneshot_scenes sc ON sc.id = st.scene_id
		WHERE st.session_id=?
		ORDER BY st.id
	`, sessionID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var st models.SceneTiming
			if err := rows.Scan(&st.ID, &st.SessionID, &st.SceneID, &st.ElapsedSeconds, &st.Status, &st.StartedAt, &st.CompletedAt,
				&st.SceneTitle, &st.SceneType, &st.EstimatedMin); err == nil {
				s.SceneTimings = append(s.SceneTimings, st)
			}
		}
	}

	paceCls, pacePct, paceLbl := computePace(s.SceneEstimated, s.SceneTimings)

	renderTemplate(c, "oneshot_pacing.html", gin.H{
		"Session":     s,
		"Elapsed":     formatDuration(s.ElapsedSeconds),
		"PaceClass":   paceCls,
		"PacePercent": pacePct,
		"PaceLabel":   paceLbl,
	})
}

func computePace(estimatedMin int, timings []models.SceneTiming) (class string, percent int, label string) {
	if estimatedMin == 0 || len(timings) == 0 {
		return "bg-success", 0, "No Estimate"
	}
	for _, t := range timings {
		if t.Status == "active" {
			elapsedMin := t.ElapsedSeconds / 60
			ratio := float64(elapsedMin) / float64(estimatedMin)
			pct := int(ratio * 100)
			if pct > 100 {
				pct = 100
			}
			if ratio >= 1.0 {
				return "bg-danger", pct, "Over Time!"
			} else if ratio >= 0.8 {
				return "bg-warning text-dark", pct, "Warning"
			}
			return "bg-success", pct, "On Track"
		}
	}
	return "bg-success", 0, "No Data"
}

func paceClass(estimatedMin int, timings []models.SceneTiming) string {
	if estimatedMin == 0 || len(timings) == 0 {
		return "bg-success"
	}
	for _, t := range timings {
		if t.Status == "active" {
			elapsedMin := t.ElapsedSeconds / 60
			ratio := float64(elapsedMin) / float64(estimatedMin)
			if ratio >= 1.0 {
				return "bg-danger"
			} else if ratio >= 0.8 {
				return "bg-warning text-dark"
			}
			return "bg-success"
		}
	}
	return "bg-success"
}

func formatDuration(seconds int) string {
	m := seconds / 60
	s := seconds % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}

// ─── Clue/Mystery Tracker API Handlers ───

func ListClues(c *gin.Context) {
	adventureID := c.Param("id")
	rows, err := db.DB.Query("SELECT id, adventure_id, title, description, clue_type, is_red_herring, is_revealed, sort_order, notes, created_at, updated_at FROM clues WHERE adventure_id=? ORDER BY sort_order, id", adventureID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	clues := make([]models.Clue, 0)
	for rows.Next() {
		var cl models.Clue
		if err := rows.Scan(&cl.ID, &cl.AdventureID, &cl.Title, &cl.Description, &cl.ClueType, &cl.IsRedHerring, &cl.IsRevealed, &cl.SortOrder, &cl.Notes, &cl.CreatedAt, &cl.UpdatedAt); err == nil {
			loadClueRelations(&cl)
			clues = append(clues, cl)
		}
	}
	c.JSON(http.StatusOK, clues)
}

func CreateClue(c *gin.Context) {
	adventureID := c.Param("id")
	userID, _ := c.Get("user_id")

	// Verify adventure belongs to user
	var exists bool
	db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM oneshot_adventures WHERE id=? AND user_id=?)", adventureID, userID).Scan(&exists)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "adventure not found"})
		return
	}

	var input struct {
		Title        string `json:"title"`
		Description  string `json:"description"`
		ClueType     string `json:"clue_type"`
		IsRedHerring bool   `json:"is_red_herring"`
		SortOrder    int    `json:"sort_order"`
		Notes        string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if input.ClueType == "" {
		input.ClueType = "direct"
	}

	result, err := db.DB.Exec("INSERT INTO clues(adventure_id, title, description, clue_type, is_red_herring, sort_order, notes) VALUES(?,?,?,?,?,?,?)",
		adventureID, input.Title, input.Description, input.ClueType, boolToInt(input.IsRedHerring), input.SortOrder, input.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func GetClue(c *gin.Context) {
	id := c.Param("id")
	var cl models.Clue
	err := db.DB.QueryRow("SELECT id, adventure_id, title, description, clue_type, is_red_herring, is_revealed, sort_order, notes, created_at, updated_at FROM clues WHERE id=?", id).
		Scan(&cl.ID, &cl.AdventureID, &cl.Title, &cl.Description, &cl.ClueType, &cl.IsRedHerring, &cl.IsRevealed, &cl.SortOrder, &cl.Notes, &cl.CreatedAt, &cl.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "clue not found"})
		return
	}
	loadClueRelations(&cl)
	c.JSON(http.StatusOK, cl)
}

func UpdateClue(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		Title        string `json:"title"`
		Description  string `json:"description"`
		ClueType     string `json:"clue_type"`
		IsRedHerring bool   `json:"is_red_herring"`
		SortOrder    int    `json:"sort_order"`
		Notes        string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := db.DB.Exec("UPDATE clues SET title=?, description=?, clue_type=?, is_red_herring=?, sort_order=?, notes=?, updated_at=datetime('now') WHERE id=?",
		input.Title, input.Description, input.ClueType, boolToInt(input.IsRedHerring), input.SortOrder, input.Notes, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func DeleteClue(c *gin.Context) {
	id := c.Param("id")
	db.DB.Exec("DELETE FROM clues WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func RevealClue(c *gin.Context) {
	id := c.Param("id")
	db.DB.Exec("UPDATE clues SET is_revealed=1, updated_at=datetime('now') WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"status": "revealed"})
}

func HideClue(c *gin.Context) {
	id := c.Param("id")
	db.DB.Exec("UPDATE clues SET is_revealed=0, updated_at=datetime('now') WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"status": "hidden"})
}

func AddClueDependency(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		DependsOnID int64 `json:"depends_on_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := db.DB.Exec("INSERT OR IGNORE INTO clue_dependencies(clue_id, depends_on_id) VALUES(?,?)", id, input.DependsOnID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "linked"})
}

func RemoveClueDependency(c *gin.Context) {
	id := c.Param("id")
	depID := c.Param("did")
	db.DB.Exec("DELETE FROM clue_dependencies WHERE clue_id=? AND depends_on_id=?", id, depID)
	c.JSON(http.StatusOK, gin.H{"status": "removed"})
}

func LinkClueNPC(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		NPCID int64 `json:"npc_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := db.DB.Exec("INSERT OR IGNORE INTO clue_npcs(clue_id, npc_id) VALUES(?,?)", id, input.NPCID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "linked"})
}

func UnlinkClueNPC(c *gin.Context) {
	id := c.Param("id")
	npcID := c.Param("nid")
	db.DB.Exec("DELETE FROM clue_npcs WHERE clue_id=? AND npc_id=?", id, npcID)
	c.JSON(http.StatusOK, gin.H{"status": "removed"})
}

func LinkClueLocation(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		LocationID int64 `json:"location_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := db.DB.Exec("INSERT OR IGNORE INTO clue_locations(clue_id, location_id) VALUES(?,?)", id, input.LocationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "linked"})
}

func UnlinkClueLocation(c *gin.Context) {
	id := c.Param("id")
	locID := c.Param("lid")
	db.DB.Exec("DELETE FROM clue_locations WHERE clue_id=? AND location_id=?", id, locID)
	c.JSON(http.StatusOK, gin.H{"status": "removed"})
}

func loadClueRelations(cl *models.Clue) {
	cl.Dependencies = make([]models.ClueDependency, 0)
	cl.DependedBy = make([]models.ClueDependency, 0)
	cl.NPCs = make([]models.ClueNPC, 0)
	cl.Locations = make([]models.ClueLocation, 0)

	// Dependencies
	depRows, err := db.DB.Query("SELECT cd.id, cd.clue_id, cd.depends_on_id, COALESCE(c.title,'') FROM clue_dependencies cd LEFT JOIN clues c ON c.id=cd.depends_on_id WHERE cd.clue_id=?", cl.ID)
	if err == nil {
		defer depRows.Close()
		for depRows.Next() {
			var d models.ClueDependency
			if depRows.Scan(&d.ID, &d.ClueID, &d.DependsOnID, &d.DependsOnTitle) == nil {
				cl.Dependencies = append(cl.Dependencies, d)
			}
		}
	}
	// Depended by
	depByRows, err := db.DB.Query("SELECT cd.id, cd.clue_id, cd.depends_on_id, COALESCE(c.title,'') FROM clue_dependencies cd LEFT JOIN clues c ON c.id=cd.clue_id WHERE cd.depends_on_id=?", cl.ID)
	if err == nil {
		defer depByRows.Close()
		for depByRows.Next() {
			var d models.ClueDependency
			if depByRows.Scan(&d.ID, &d.ClueID, &d.DependsOnID, &d.DependsOnTitle) == nil {
				cl.DependedBy = append(cl.DependedBy, d)
			}
		}
	}
	// NPCs
	npcRows, err := db.DB.Query("SELECT cn.id, cn.clue_id, cn.npc_id, COALESCE(n.name,'') FROM clue_npcs cn LEFT JOIN npcs n ON n.id=cn.npc_id WHERE cn.clue_id=?", cl.ID)
	if err == nil {
		defer npcRows.Close()
		for npcRows.Next() {
			var n models.ClueNPC
			if npcRows.Scan(&n.ID, &n.ClueID, &n.NPCID, &n.NPCName) == nil {
				cl.NPCs = append(cl.NPCs, n)
			}
		}
	}
	// Locations
	locRows, err := db.DB.Query("SELECT cl.id, cl.clue_id, cl.location_id, COALESCE(l.name,'') FROM clue_locations cl LEFT JOIN locations l ON l.id=cl.location_id WHERE cl.clue_id=?", cl.ID)
	if err == nil {
		defer locRows.Close()
		for locRows.Next() {
			var l models.ClueLocation
			if locRows.Scan(&l.ID, &l.ClueID, &l.LocationID, &l.LocationName) == nil {
				cl.Locations = append(cl.Locations, l)
			}
		}
	}
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

// ─── HTMX Clue Handlers ───

func HtmxListClues(c *gin.Context) {
	adventureID := c.Param("id")

	rows, err := db.DB.Query("SELECT id, adventure_id, title, description, clue_type, is_red_herring, is_revealed, sort_order, notes, created_at, updated_at FROM clues WHERE adventure_id=? ORDER BY sort_order, id", adventureID)
	if err != nil {
		c.String(http.StatusInternalServerError, "query error")
		return
	}
	defer rows.Close()

	clues := make([]models.Clue, 0)
	for rows.Next() {
		var cl models.Clue
		if rows.Scan(&cl.ID, &cl.AdventureID, &cl.Title, &cl.Description, &cl.ClueType, &cl.IsRedHerring, &cl.IsRevealed, &cl.SortOrder, &cl.Notes, &cl.CreatedAt, &cl.UpdatedAt) == nil {
			loadClueRelations(&cl)
			clues = append(clues, cl)
		}
	}

	renderTemplate(c, "oneshot_clues.html", gin.H{
		"Clues": clues,
		"AdventureID": adventureID,
	})
}

func HtmxGetClueDetail(c *gin.Context) {
	id := c.Param("id")
	var cl models.Clue
	err := db.DB.QueryRow("SELECT id, adventure_id, title, description, clue_type, is_red_herring, is_revealed, sort_order, notes, created_at, updated_at FROM clues WHERE id=?", id).
		Scan(&cl.ID, &cl.AdventureID, &cl.Title, &cl.Description, &cl.ClueType, &cl.IsRedHerring, &cl.IsRevealed, &cl.SortOrder, &cl.Notes, &cl.CreatedAt, &cl.UpdatedAt)
	if err != nil {
		c.String(http.StatusNotFound, "Clue not found")
		return
	}
	loadClueRelations(&cl)

	renderTemplate(c, "oneshot_clue_detail.html", gin.H{
		"Clue": cl,
	})
}

func HtmxNewClueForm(c *gin.Context) {
	adventureID := c.Param("id")
	renderTemplate(c, "oneshot_clue_form.html", gin.H{
		"AdventureID": adventureID,
		"Clue":        nil,
	})
}

func HtmxCreateClue(c *gin.Context) {
	adventureID := c.Param("id")
	userID, _ := c.Get("user_id")

	var exists bool
	db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM oneshot_adventures WHERE id=? AND user_id=?)", adventureID, userID).Scan(&exists)
	if !exists {
		c.String(http.StatusNotFound, "Adventure not found")
		return
	}

	title := c.PostForm("title")
	description := c.PostForm("description")
	clueType := c.PostForm("clue_type")
	if clueType == "" {
		clueType = "direct"
	}
	sortOrder, _ := strconv.Atoi(c.PostForm("sort_order"))
	notes := c.PostForm("notes")

	db.DB.Exec("INSERT INTO clues(adventure_id, title, description, clue_type, is_red_herring, sort_order, notes) VALUES(?,?,?,?,?,?,?)",
		adventureID, title, description, clueType, 0, sortOrder, notes)

	HtmxListClues(c)
}

func HtmxEditClueForm(c *gin.Context) {
	id := c.Param("id")
	var cl models.Clue
	err := db.DB.QueryRow("SELECT id, adventure_id, title, description, clue_type, is_red_herring, is_revealed, sort_order, notes, created_at, updated_at FROM clues WHERE id=?", id).
		Scan(&cl.ID, &cl.AdventureID, &cl.Title, &cl.Description, &cl.ClueType, &cl.IsRedHerring, &cl.IsRevealed, &cl.SortOrder, &cl.Notes, &cl.CreatedAt, &cl.UpdatedAt)
	if err != nil {
		c.String(http.StatusNotFound, "Clue not found")
		return
	}
	renderTemplate(c, "oneshot_clue_form.html", gin.H{
		"AdventureID": cl.AdventureID,
		"Clue":        cl,
	})
}

func HtmxUpdateClue(c *gin.Context) {
	id := c.Param("id")
	title := c.PostForm("title")
	description := c.PostForm("description")
	clueType := c.PostForm("clue_type")
	sortOrder, _ := strconv.Atoi(c.PostForm("sort_order"))
	notes := c.PostForm("notes")

	db.DB.Exec("UPDATE clues SET title=?, description=?, clue_type=?, sort_order=?, notes=?, updated_at=datetime('now') WHERE id=?",
		title, description, clueType, sortOrder, notes, id)

	HtmxGetClueDetail(c)
}

func HtmxDeleteClue(c *gin.Context) {
	id := c.Param("id")
	// Get adventure_id for redirect
	var advID int64
	db.DB.QueryRow("SELECT adventure_id FROM clues WHERE id=?", id).Scan(&advID)
	db.DB.Exec("DELETE FROM clues WHERE id=?", id)
	if advID > 0 {
		HtmxListClues(c)
		return
	}
	c.String(http.StatusOK, "Deleted")
}

func HtmxRevealClue(c *gin.Context) {
	id := c.Param("id")
	db.DB.Exec("UPDATE clues SET is_revealed=1, updated_at=datetime('now') WHERE id=?", id)
	HtmxGetClueDetail(c)
}

func HtmxHideClue(c *gin.Context) {
	id := c.Param("id")
	db.DB.Exec("UPDATE clues SET is_revealed=0, updated_at=datetime('now') WHERE id=?", id)
	HtmxGetClueDetail(c)
}

// ─── Pregenerated Characters ───

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
		Name       string `json:"name"`
		Race       string `json:"race"`
		Class      string `json:"class"`
		Subclass   string `json:"subclass"`
		Level      int    `json:"level"`
		Background string `json:"background"`
		Alignment  string `json:"alignment"`
		Str        int    `json:"str"`
		Dex        int    `json:"dex"`
		Con        int    `json:"con"`
		Int        int    `json:"int"`
		Wis        int    `json:"wis"`
		Cha        int    `json:"cha"`
		HP         int    `json:"hp"`
		AC         int    `json:"ac"`
		Speed      int    `json:"speed"`
		Skills     string `json:"skills"`
		Equipment  string `json:"equipment"`
		Spells     string `json:"spells"`
		Features   string `json:"features"`
		Personality string `json:"personality"`
		Backstory  string `json:"backstory"`
		PortraitURL string `json:"portrait_url"`
		Notes      string `json:"notes"`
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
		Name       string `json:"name"`
		Race       string `json:"race"`
		Class      string `json:"class"`
		Subclass   string `json:"subclass"`
		Level      int    `json:"level"`
		Background string `json:"background"`
		Alignment  string `json:"alignment"`
		Str        int    `json:"str"`
		Dex        int    `json:"dex"`
		Con        int    `json:"con"`
		Int        int    `json:"int"`
		Wis        int    `json:"wis"`
		Cha        int    `json:"cha"`
		HP         int    `json:"hp"`
		AC         int    `json:"ac"`
		Speed      int    `json:"speed"`
		Skills     string `json:"skills"`
		Equipment  string `json:"equipment"`
		Spells     string `json:"spells"`
		Features   string `json:"features"`
		Personality string `json:"personality"`
		Backstory  string `json:"backstory"`
		PortraitURL string `json:"portrait_url"`
		Notes      string `json:"notes"`
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
	"barbarian":  {"tank"},
	"fighter":    {"tank", "damage"},
	"paladin":    {"tank", "healer"},
	"cleric":     {"healer"},
	"druid":      {"healer", "support"},
	"wizard":     {"damage", "support"},
	"sorcerer":   {"damage"},
	"warlock":    {"damage"},
	"bard":       {"support", "healer"},
	"rogue":      {"damage", "skill"},
	"monk":       {"damage"},
	"ranger":     {"damage", "skill"},
	"artificer":  {"support"},
}

var allRoles = []string{"tank", "healer", "damage", "support", "skill"}

func CheckPartyBalance(c *gin.Context) {
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

	roleCounts := map[string]int{}
	for _, r := range allRoles {
		roleCounts[r] = 0
	}
	for _, ch := range chars {
		class := ch.Class
		roles, ok := classRoles[class]
		if !ok {
			roles = classRoles["fighter"]
		}
		for _, r := range roles {
			roleCounts[r]++
		}
	}

	missing := make([]string, 0)
	for _, r := range allRoles {
		if roleCounts[r] == 0 {
			missing = append(missing, r)
		}
	}

	score := 0
	for _, r := range allRoles {
		if roleCounts[r] > 0 {
			score += 2
		}
	}
	score += len(chars) * 2
	if score > 20 {
		score = 20
	}

	rating := "poor"
	suggestion := "Consider creating more characters to cover missing roles."
	if len(chars) >= 3 && len(missing) <= 2 {
		rating = "fair"
		suggestion = "Your party could use some additional coverage."
	}
	if len(chars) >= 4 && len(missing) <= 1 {
		rating = "good"
		suggestion = "Your party is well-balanced for most adventures."
	}
	if len(chars) >= 4 && len(missing) == 0 {
		rating = "excellent"
		suggestion = "Your party covers all essential roles!"
	}
	if len(chars) >= 5 && len(missing) <= 1 {
		rating = "great"
		suggestion = "Your party has excellent depth and versatility."
	}

	c.JSON(http.StatusOK, models.PartyBalance{
		Characters: chars,
		Roles:      roleCounts,
		Score:      score,
		Rating:     rating,
		Missing:    missing,
		Suggestion: suggestion,
	})
}

func GeneratePregen(c *gin.Context) {
	userID, _ := c.Get("user_id")
	genLevel, _ := strconv.Atoi(c.DefaultQuery("level", "3"))
	if genLevel < 1 || genLevel > 20 {
		genLevel = 3
	}
	genRace := c.Query("race")
	genClass := c.Query("class")

	races := []string{"dwarf", "elf", "halfling", "human", "dragonborn", "gnome", "half-elf", "half-orc", "tiefling"}
	classes := []string{"barbarian", "fighter", "paladin", "cleric", "druid", "wizard", "sorcerer", "warlock", "bard", "rogue", "monk", "ranger", "artificer"}
	names := []string{"Aldric", "Briar", "Cassian", "Dorn", "Elara", "Finn", "Grom", "Halia", "Ivy", "Jax", "Kira", "Lark", "Mira", "Nyx", "Orin", "Piper", "Quinn", "Rook", "Sage", "Talon", "Una", "Vex", "Wren", "Xara", "Zephyr"}

	if genClass == "" {
		genClass = classes[rand.Intn(len(classes))]
	}
	if genRace == "" {
		genRace = races[rand.Intn(len(races))]
	}
	name := names[rand.Intn(len(names))]

	str, dex, con, intel, wis, cha := 10, 10, 10, 10, 10, 10
	switch genClass {
	case "barbarian", "fighter", "paladin":
		str, dex, con, intel, wis, cha = 16, 12, 14, 8, 10, 10
	case "cleric", "druid":
		str, dex, con, intel, wis, cha = 12, 10, 14, 10, 16, 8
	case "wizard":
		str, dex, con, intel, wis, cha = 8, 12, 14, 16, 10, 10
	case "sorcerer", "warlock", "bard":
		str, dex, con, intel, wis, cha = 8, 12, 14, 10, 10, 16
	case "rogue", "ranger", "monk":
		str, dex, con, intel, wis, cha = 10, 16, 14, 10, 12, 8
	case "artificer":
		str, dex, con, intel, wis, cha = 8, 12, 14, 16, 10, 10
	}

	hp := 8 + (con-10)/2 + (genLevel-1)*5
	ac := 12 + (dex-10)/2

	result, err := db.DB.Exec(`
		INSERT INTO pregen_characters(user_id, name, race, class, level, str, dex, con, int, wis, cha, hp, ac, speed)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,30)`,
		userID, name, genRace, genClass, genLevel, str, dex, con, intel, wis, cha, hp, ac)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// ─── HTMX Pregenerated Characters ───

func HtmxListPregens(c *gin.Context) {
	userID, _ := c.Get("user_id")
	rows, err := db.DB.Query("SELECT id, user_id, name, race, class, subclass, level, background, alignment, str, dex, con, int, wis, cha, hp, ac, speed, skills, equipment, spells, features, personality, backstory, portrait_url, notes, created_at, updated_at FROM pregen_characters WHERE user_id=? ORDER BY level DESC, name ASC", userID)
	if err != nil {
		c.String(http.StatusInternalServerError, "query error")
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
	renderTemplate(c, "oneshot_pregens.html", gin.H{"Pregens": chars})
}

func HtmxGeneratePregen(c *gin.Context) {
	userID, _ := c.Get("user_id")
	name := c.Query("name")
	genLevel, _ := strconv.Atoi(c.DefaultQuery("level", "3"))
	if genLevel < 1 || genLevel > 20 {
		genLevel = 3
	}
	genRace := c.Query("race")
	genClass := c.Query("class")

	races := []string{"dwarf", "elf", "halfling", "human", "dragonborn", "gnome", "half-elf", "half-orc", "tiefling"}
	classes := []string{"barbarian", "fighter", "paladin", "cleric", "druid", "wizard", "sorcerer", "warlock", "bard", "rogue", "monk", "ranger", "artificer"}
	names := []string{"Aldric", "Briar", "Cassian", "Dorn", "Elara", "Finn", "Grom", "Halia", "Ivy", "Jax", "Kira", "Lark", "Mira", "Nyx", "Orin", "Piper", "Quinn", "Rook", "Sage", "Talon", "Una", "Vex", "Wren", "Xara", "Zephyr"}

	if genClass == "" {
		genClass = classes[rand.Intn(len(classes))]
	}
	if genRace == "" {
		genRace = races[rand.Intn(len(races))]
	}
	if name == "" {
		name = names[rand.Intn(len(names))]
	}

	var str, dex, con, intel, wis, cha int
	switch genClass {
	case "barbarian", "fighter", "paladin":
		str, dex, con, intel, wis, cha = 16, 12, 14, 8, 10, 10
	case "cleric", "druid":
		str, dex, con, intel, wis, cha = 12, 10, 14, 10, 16, 8
	case "wizard":
		str, dex, con, intel, wis, cha = 8, 12, 14, 16, 10, 10
	case "sorcerer", "warlock", "bard":
		str, dex, con, intel, wis, cha = 8, 12, 14, 10, 10, 16
	case "rogue", "ranger", "monk":
		str, dex, con, intel, wis, cha = 10, 16, 14, 10, 12, 8
	case "artificer":
		str, dex, con, intel, wis, cha = 8, 12, 14, 16, 10, 10
	default:
		str, dex, con, intel, wis, cha = 10, 10, 10, 10, 10, 10
	}

	hp := 8 + (con-10)/2 + (genLevel-1)*5
	ac := 12 + (dex-10)/2

	_, err := db.DB.Exec(`
		INSERT INTO pregen_characters(user_id, name, race, class, level, str, dex, con, int, wis, cha, hp, ac, speed)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,30)`,
		userID, name, genRace, genClass, genLevel, str, dex, con, intel, wis, cha, hp, ac)
	if err != nil {
		c.String(http.StatusInternalServerError, "insert error")
		return
	}

	HtmxListPregens(c)
}

func HtmxPregenCard(c *gin.Context) {
	id := c.Param("id")
	var ch models.PregeneratedCharacter
	err := db.DB.QueryRow("SELECT id, user_id, name, race, class, subclass, level, background, alignment, str, dex, con, int, wis, cha, hp, ac, speed, skills, equipment, spells, features, personality, backstory, portrait_url, notes, created_at, updated_at FROM pregen_characters WHERE id=?", id).
		Scan(&ch.ID, &ch.UserID, &ch.Name, &ch.Race, &ch.Class, &ch.Subclass, &ch.Level, &ch.Background, &ch.Alignment,
			&ch.Str, &ch.Dex, &ch.Con, &ch.Int, &ch.Wis, &ch.Cha, &ch.HP, &ch.AC, &ch.Speed,
			&ch.Skills, &ch.Equipment, &ch.Spells, &ch.Features, &ch.Personality, &ch.Backstory,
			&ch.PortraitURL, &ch.Notes, &ch.CreatedAt, &ch.UpdatedAt)
	if err != nil {
		c.String(http.StatusNotFound, "Pregen not found")
		return
	}
	renderTemplate(c, "oneshot_pregen_card.html", gin.H{"C": ch})
}

// ─── Prep Dashboard Handlers ───

func HtmxGetPrepDashboard(c *gin.Context) {
	adventureID := c.Param("id")
	userID, _ := c.Get("user_id")

	// Load adventure
	var adv models.OneShotAdventure
	err := db.DB.QueryRow("SELECT id, user_id, campaign_id, title, premise, hook, template, estimated_minutes, difficulty, notes, created_at, updated_at FROM oneshot_adventures WHERE id=? AND user_id=?", adventureID, userID).
		Scan(&adv.ID, &adv.UserID, &adv.CampaignID, &adv.Title, &adv.Premise, &adv.Hook, &adv.Template, &adv.EstimatedMinutes, &adv.Difficulty, &adv.Notes, &adv.CreatedAt, &adv.UpdatedAt)
	if err != nil {
		c.String(http.StatusNotFound, "Adventure not found")
		return
	}

	// Load acts with scenes
	actRows, _ := db.DB.Query("SELECT id, adventure_id, number, title, description, estimated_minutes FROM oneshot_acts WHERE adventure_id=? ORDER BY number", adventureID)
	if actRows != nil {
		defer actRows.Close()
		for actRows.Next() {
			var act models.OneShotAct
			if err := actRows.Scan(&act.ID, &act.AdventureID, &act.Number, &act.Title, &act.Description, &act.EstimatedMinutes); err != nil {
				continue
			}
			// Load scenes for this act
			sceneRows, _ := db.DB.Query("SELECT id, act_id, number, title, description, scene_type, location_id, encounter_id, estimated_minutes, notes FROM oneshot_scenes WHERE act_id=? ORDER BY number", act.ID)
			if sceneRows != nil {
				for sceneRows.Next() {
					var sc models.OneShotScene
					if err := sceneRows.Scan(&sc.ID, &sc.ActID, &sc.Number, &sc.Title, &sc.Description, &sc.SceneType, &sc.LocationID, &sc.EncounterID, &sc.EstimatedMinutes, &sc.Notes); err == nil {
						act.Scenes = append(act.Scenes, sc)
					}
				}
				sceneRows.Close()
			}
			adv.Acts = append(adv.Acts, act)
		}
	}

	// Load linked NPCs
	npcRows, _ := db.DB.Query(`
		SELECT oan.id, oan.adventure_id, oan.npc_id, oan.role, COALESCE(n.name,'')
		FROM oneshot_adventure_npcs oan
		LEFT JOIN npcs n ON n.id = oan.npc_id
		WHERE oan.adventure_id=?
	`, adventureID)
	if npcRows != nil {
		defer npcRows.Close()
		for npcRows.Next() {
			var n models.OneShotAdventureNPC
			if err := npcRows.Scan(&n.ID, &n.AdventureID, &n.NPCID, &n.Role, &n.NPCName); err == nil {
				adv.NPCs = append(adv.NPCs, n)
			}
		}
	}

	// Load linked locations
	locRows, _ := db.DB.Query(`
		SELECT oal.id, oal.adventure_id, oal.location_id, COALESCE(l.name,'')
		FROM oneshot_adventure_locations oal
		LEFT JOIN locations l ON l.id = oal.location_id
		WHERE oal.adventure_id=?
	`, adventureID)
	if locRows != nil {
		defer locRows.Close()
		for locRows.Next() {
			var l models.OneShotAdventureLocation
			if err := locRows.Scan(&l.ID, &l.AdventureID, &l.LocationID, &l.LocationName); err == nil {
				adv.Locations = append(adv.Locations, l)
			}
		}
	}

	// Load linked encounters
	encRows, _ := db.DB.Query(`
		SELECT oae.id, oae.adventure_id, oae.encounter_id, COALESCE(e.name,'')
		FROM oneshot_adventure_encounters oae
		LEFT JOIN encounter_templates e ON e.id = oae.encounter_id
		WHERE oae.adventure_id=?
	`, adventureID)
	if encRows != nil {
		defer encRows.Close()
		for encRows.Next() {
			var e models.OneShotAdventureEncounter
			if err := encRows.Scan(&e.ID, &e.AdventureID, &e.EncounterID, &e.EncounterName); err == nil {
				adv.Encounters = append(adv.Encounters, e)
			}
		}
	}

	// Load clues for this adventure
	var clues []models.Clue
	clueRows, _ := db.DB.Query("SELECT id, adventure_id, title, description, clue_type, is_red_herring, is_revealed, sort_order, notes, created_at, updated_at FROM clues WHERE adventure_id=? ORDER BY is_revealed DESC, id", adventureID)
	if clueRows != nil {
		defer clueRows.Close()
		for clueRows.Next() {
			var cl models.Clue
			if err := clueRows.Scan(&cl.ID, &cl.AdventureID, &cl.Title, &cl.Description, &cl.ClueType, &cl.IsRedHerring, &cl.IsRevealed, &cl.SortOrder, &cl.Notes, &cl.CreatedAt, &cl.UpdatedAt); err == nil {
				clues = append(clues, cl)
			}
		}
	}

	// Load pregens for this user (not adventure-specific, but user-specific)
	var pregens []models.PregeneratedCharacter
	pregenRows, _ := db.DB.Query("SELECT id, user_id, name, race, class, subclass, level, background, alignment, str, dex, con, int, wis, cha, hp, ac, speed, skills, equipment, spells, features, personality, backstory, portrait_url, notes, created_at, updated_at FROM pregen_characters WHERE user_id=? ORDER BY updated_at DESC", userID)
	if pregenRows != nil {
		defer pregenRows.Close()
		for pregenRows.Next() {
			var p models.PregeneratedCharacter
			if err := pregenRows.Scan(&p.ID, &p.UserID, &p.Name, &p.Race, &p.Class, &p.Subclass, &p.Level, &p.Background, &p.Alignment,
				&p.Str, &p.Dex, &p.Con, &p.Int, &p.Wis, &p.Cha, &p.HP, &p.AC, &p.Speed,
				&p.Skills, &p.Equipment, &p.Spells, &p.Features, &p.Personality, &p.Backstory,
				&p.PortraitURL, &p.Notes, &p.CreatedAt, &p.UpdatedAt); err == nil {
				pregens = append(pregens, p)
			}
		}
	}

	// Load checklist
	var checklist []models.PrepChecklistItem
	clRows, _ := db.DB.Query("SELECT id, adventure_id, item, category, is_checked, sort_order FROM prep_checklist WHERE adventure_id=? ORDER BY sort_order, id", adventureID)
	if clRows != nil {
		defer clRows.Close()
		for clRows.Next() {
			var item models.PrepChecklistItem
			var checked int
			if err := clRows.Scan(&item.ID, &item.AdventureID, &item.Item, &item.Category, &checked, &item.SortOrder); err == nil {
				item.IsChecked = checked == 1
				checklist = append(checklist, item)
			}
		}
	}

	// Check for pacing session
	var sessionID *int64
	var pacing *models.SessionPacing
	var sid int64
	var pacingStatus string
	err = db.DB.QueryRow("SELECT id, status FROM session_pacing WHERE adventure_id=? AND status IN ('running','paused') ORDER BY id DESC LIMIT 1", adventureID).Scan(&sid, &pacingStatus)
	if err == nil {
		sessionID = &sid
		pacing = &models.SessionPacing{ID: sid, Status: pacingStatus}
		db.DB.QueryRow("SELECT COALESCE(SUM(elapsed_seconds),0) FROM scene_timings WHERE session_id=?", sid).Scan(&pacing.ElapsedSeconds)
	}

	prepData := models.PrepDashboardData{
		Adventure: adv,
		Acts:      adv.Acts,
		Clues:     clues,
		Pregens:   pregens,
		Checklist: checklist,
		Pacing:    pacing,
		SessionID: sessionID,
	}

	renderTemplate(c, "oneshot_prep_dashboard.html", gin.H{
		"D": prepData,
	})
}

// ─── Prep Checklist Handlers ───

func ListPrepChecklist(c *gin.Context) {
	adventureID := c.Param("id")
	rows, err := db.DB.Query("SELECT id, adventure_id, item, category, is_checked, sort_order FROM prep_checklist WHERE adventure_id=? ORDER BY sort_order, id", adventureID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var items []models.PrepChecklistItem
	for rows.Next() {
		var item models.PrepChecklistItem
		var checked int
		if err := rows.Scan(&item.ID, &item.AdventureID, &item.Item, &item.Category, &checked, &item.SortOrder); err == nil {
			item.IsChecked = checked == 1
			items = append(items, item)
		}
	}
	c.JSON(http.StatusOK, items)
}

func CreatePrepChecklistItem(c *gin.Context) {
	adventureID := c.Param("id")
	var req struct {
		Item      string `json:"item"`
		Category  string `json:"category"`
		SortOrder int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Category == "" {
		req.Category = "general"
	}

	result, err := db.DB.Exec("INSERT INTO prep_checklist(adventure_id, item, category, sort_order) VALUES(?,?,?,?)", adventureID, req.Item, req.Category, req.SortOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func UpdatePrepChecklistItem(c *gin.Context) {
	id := c.Param("cid")
	var req struct {
		Item      string `json:"item,omitempty"`
		Category  string `json:"category,omitempty"`
		IsChecked *bool  `json:"is_checked,omitempty"`
		SortOrder *int   `json:"sort_order,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Item != "" {
		db.DB.Exec("UPDATE prep_checklist SET item=? WHERE id=?", req.Item, id)
	}
	if req.Category != "" {
		db.DB.Exec("UPDATE prep_checklist SET category=? WHERE id=?", req.Category, id)
	}
	if req.IsChecked != nil {
		val := 0
		if *req.IsChecked {
			val = 1
		}
		db.DB.Exec("UPDATE prep_checklist SET is_checked=? WHERE id=?", val, id)
	}
	if req.SortOrder != nil {
		db.DB.Exec("UPDATE prep_checklist SET sort_order=? WHERE id=?", *req.SortOrder, id)
	}

	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func DeletePrepChecklistItem(c *gin.Context) {
	id := c.Param("cid")
	_, err := db.DB.Exec("DELETE FROM prep_checklist WHERE id=?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// HTMX handlers for checklist
func HtmxRenderChecklist(c *gin.Context) {
	adventureID := c.Param("id")
	rows, _ := db.DB.Query("SELECT id, adventure_id, item, category, is_checked, sort_order FROM prep_checklist WHERE adventure_id=? ORDER BY sort_order, id", adventureID)
	if rows != nil {
		defer rows.Close()
		var items []models.PrepChecklistItem
		for rows.Next() {
			var item models.PrepChecklistItem
			var checked int
			if err := rows.Scan(&item.ID, &item.AdventureID, &item.Item, &item.Category, &checked, &item.SortOrder); err == nil {
				item.IsChecked = checked == 1
				items = append(items, item)
			}
		}
		renderTemplate(c, "oneshot_checklist.html", gin.H{"Items": items, "AdventureID": adventureID})
		return
	}
	renderTemplate(c, "oneshot_checklist.html", gin.H{"Items": []models.PrepChecklistItem{}, "AdventureID": adventureID})
}

func HtmxToggleChecklistItem(c *gin.Context) {
	id := c.Param("cid")
	var checked int
	db.DB.QueryRow("SELECT is_checked FROM prep_checklist WHERE id=?", id).Scan(&checked)
	newVal := 0
	if checked == 0 {
		newVal = 1
	}
	db.DB.Exec("UPDATE prep_checklist SET is_checked=? WHERE id=?", newVal, id)

	// Get adventure ID for re-render
	var adventureID int64
	db.DB.QueryRow("SELECT adventure_id FROM prep_checklist WHERE id=?", id).Scan(&adventureID)

	// Re-render full checklist
	rows, _ := db.DB.Query("SELECT id, adventure_id, item, category, is_checked, sort_order FROM prep_checklist WHERE adventure_id=? ORDER BY sort_order, id", adventureID)
	if rows != nil {
		defer rows.Close()
		var items []models.PrepChecklistItem
		for rows.Next() {
			var item models.PrepChecklistItem
			var ck int
			if err := rows.Scan(&item.ID, &item.AdventureID, &item.Item, &item.Category, &ck, &item.SortOrder); err == nil {
				item.IsChecked = ck == 1
				items = append(items, item)
			}
		}
		renderTemplate(c, "oneshot_checklist.html", gin.H{"Items": items, "AdventureID": adventureID})
		return
	}
	c.String(http.StatusOK, "")
}

func HtmxAddChecklistItem(c *gin.Context) {
	adventureID := c.Param("id")
	item := c.PostForm("item")
	category := c.DefaultPostForm("category", "general")

	if item == "" {
		c.String(http.StatusBadRequest, "Item is required")
		return
	}

	// Get max sort order
	var maxOrder int
	db.DB.QueryRow("SELECT COALESCE(MAX(sort_order),0) FROM prep_checklist WHERE adventure_id=?", adventureID).Scan(&maxOrder)

	_, err := db.DB.Exec("INSERT INTO prep_checklist(adventure_id, item, category, sort_order) VALUES(?,?,?,?)", adventureID, item, category, maxOrder+1)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error adding item")
		return
	}

	HtmxRenderChecklist(c)
}

func HtmxDeleteChecklistItem(c *gin.Context) {
	id := c.Param("cid")

	// Get adventure ID before delete
	var adventureID int64
	db.DB.QueryRow("SELECT adventure_id FROM prep_checklist WHERE id=?", id).Scan(&adventureID)

	db.DB.Exec("DELETE FROM prep_checklist WHERE id=?", id)

	// Re-render
	rows, _ := db.DB.Query("SELECT id, adventure_id, item, category, is_checked, sort_order FROM prep_checklist WHERE adventure_id=? ORDER BY sort_order, id", adventureID)
	if rows != nil {
		defer rows.Close()
		var items []models.PrepChecklistItem
		for rows.Next() {
			var item models.PrepChecklistItem
			var ck int
			if err := rows.Scan(&item.ID, &item.AdventureID, &item.Item, &item.Category, &ck, &item.SortOrder); err == nil {
				item.IsChecked = ck == 1
				items = append(items, item)
			}
		}
		renderTemplate(c, "oneshot_checklist.html", gin.H{"Items": items, "AdventureID": adventureID})
		return
	}
	c.String(http.StatusOK, "")
}

// ─── Session Flow Print View ───

func HtmxGetSessionFlow(c *gin.Context) {
	adventureID := c.Param("id")
	userID, _ := c.Get("user_id")

	var adv models.OneShotAdventure
	err := db.DB.QueryRow("SELECT id, user_id, campaign_id, title, premise, hook, template, estimated_minutes, difficulty, notes, created_at, updated_at FROM oneshot_adventures WHERE id=? AND user_id=?", adventureID, userID).
		Scan(&adv.ID, &adv.UserID, &adv.CampaignID, &adv.Title, &adv.Premise, &adv.Hook, &adv.Template, &adv.EstimatedMinutes, &adv.Difficulty, &adv.Notes, &adv.CreatedAt, &adv.UpdatedAt)
	if err != nil {
		c.String(http.StatusNotFound, "Adventure not found")
		return
	}

	// Load acts with scenes
	actRows, _ := db.DB.Query("SELECT id, adventure_id, number, title, description, estimated_minutes FROM oneshot_acts WHERE adventure_id=? ORDER BY number", adventureID)
	if actRows != nil {
		defer actRows.Close()
		for actRows.Next() {
			var act models.OneShotAct
			if err := actRows.Scan(&act.ID, &act.AdventureID, &act.Number, &act.Title, &act.Description, &act.EstimatedMinutes); err != nil {
				continue
			}
			sceneRows, _ := db.DB.Query("SELECT id, act_id, number, title, description, scene_type, location_id, encounter_id, estimated_minutes, notes FROM oneshot_scenes WHERE act_id=? ORDER BY number", act.ID)
			if sceneRows != nil {
				for sceneRows.Next() {
					var sc models.OneShotScene
					if err := sceneRows.Scan(&sc.ID, &sc.ActID, &sc.Number, &sc.Title, &sc.Description, &sc.SceneType, &sc.LocationID, &sc.EncounterID, &sc.EstimatedMinutes, &sc.Notes); err == nil {
						act.Scenes = append(act.Scenes, sc)
					}
				}
				sceneRows.Close()
			}
			adv.Acts = append(adv.Acts, act)
		}
	}

	// Calculate total time
	var totalTime int
	for _, a := range adv.Acts {
		totalTime += a.EstimatedMinutes
		for _, s := range a.Scenes {
			_ = s
		}
	}

	renderTemplate(c, "oneshot_session_flow.html", gin.H{
		"Adventure": adv,
		"TotalTime": totalTime,
	})
}

// ─── DM Screen / Quick Reference Handlers ───

// QuickReferenceData returns the embedded 5E quick reference sections
func QuickReferenceData() []models.DmQuickRefSection {
	return []models.DmQuickRefSection{
		{
			Title: "Conditions",
			Icon:  "fa-exclamation-triangle",
			Entries: []models.DmQuickRefEntry{
				{Name: "Blinded", Description: "Auto-fail Perception (sight). Attack rolls against creature have advantage. Creature's attacks have disadvantage.", Reference: "PHB p.290"},
				{Name: "Charmed", Description: "Can't attack charmer. Charmer has advantage on social checks. Not immunity.", Reference: "PHB p.290"},
				{Name: "Deafened", Description: "Auto-fail Perception (hearing).", Reference: "PHB p.290"},
				{Name: "Exhaustion", Description: "1: Disadvantage on ability checks. 2: Speed halved. 3: Disadvantage on attacks/saves. 4: HP max halved. 5: Speed 0. 6: Death.", Reference: "PHB p.291"},
				{Name: "Frightened", Description: "Disadvantage on checks/attacks while source visible. Can't move closer to source.", Reference: "PHB p.290"},
				{Name: "Grappled", Description: "Speed 0. Ends if grappler incapacitated or target breaks free (Athletics/Acrobatics vs DC).", Reference: "PHB p.290"},
				{Name: "Incapacitated", Description: "Can't take actions, bonus actions, or reactions.", Reference: "PHB p.290"},
				{Name: "Invisible", Description: "Can't be seen without special senses. Attack rolls have advantage. Attacks against have disadvantage.", Reference: "PHB p.291"},
				{Name: "Paralyzed", Description: "Incapacitated. Can't move/speak. Auto-fail STR/DEX saves. Attacks have advantage. Auto-crit within 5ft.", Reference: "PHB p.291"},
				{Name: "Petrified", Description: "Turned to stone + weight x10. Incapacitated. Resist all damage. Immune to poison/disease.", Reference: "PHB p.291"},
				{Name: "Poisoned", Description: "Disadvantage on attack rolls and ability checks.", Reference: "PHB p.292"},
				{Name: "Prone", Description: "Can only crawl. Attacks have disadvantage. Attacks within 5ft have advantage, beyond have disadvantage.", Reference: "PHB p.292"},
				{Name: "Restrained", Description: "Speed 0. Attacks have disadvantage. Attacks against have advantage. DEX saves have disadvantage.", Reference: "PHB p.292"},
				{Name: "Stunned", Description: "Incapacitated. Can't move. Auto-fail STR/DEX saves. Attacks have advantage.", Reference: "PHB p.292"},
				{Name: "Unconscious", Description: "Incapacitated + prone. Auto-fail STR/DEX saves. Attacks have advantage. Auto-crit within 5ft.", Reference: "PHB p.292"},
			},
		},
		{
			Title: "Combat Actions",
			Icon:  "fa-fist-raised",
			Entries: []models.DmQuickRefEntry{
				{Name: "Attack", Description: "Make a melee or ranged weapon attack, or unarmed strike.", Reference: "PHB p.192"},
				{Name: "Cast a Spell", Description: "Cast a spell with casting time of 1 action. Must have components.", Reference: "PHB p.192"},
				{Name: "Dash", Description: "Gain extra movement equal to your speed.", Reference: "PHB p.192"},
				{Name: "Disengage", Description: "Your movement doesn't provoke opportunity attacks.", Reference: "PHB p.192"},
				{Name: "Dodge", Description: "Attacks against have disadvantage. DEX saves have advantage.", Reference: "PHB p.192"},
				{Name: "Help", Description: "Give ally advantage on next ability check or attack.", Reference: "PHB p.192"},
				{Name: "Hide", Description: "Make Stealth check to become unseen.", Reference: "PHB p.192"},
				{Name: "Ready", Description: "Prepare action with trigger. Use reaction to execute.", Reference: "PHB p.193"},
				{Name: "Search", Description: "Make Perception or Investigation check.", Reference: "PHB p.193"},
				{Name: "Use Object", Description: "Interact with an object (second object interaction uses action).", Reference: "PHB p.193"},
			},
		},
		{
			Title: "Difficulty Classes",
			Icon:  "fa-tachometer-alt",
			Entries: []models.DmQuickRefEntry{
				{Name: "Very Easy", Description: "DC 5"},
				{Name: "Easy", Description: "DC 10"},
				{Name: "Medium", Description: "DC 15"},
				{Name: "Hard", Description: "DC 20"},
				{Name: "Very Hard", Description: "DC 25"},
				{Name: "Nearly Impossible", Description: "DC 30"},
			},
		},
		{
			Title: "Rest & Recovery",
			Icon:  "fa-bed",
			Entries: []models.DmQuickRefEntry{
				{Name: "Short Rest", Description: "1 hour. Spend Hit Dice (any number). Regain abilities that recharge on short rest.", Reference: "PHB p.186"},
				{Name: "Long Rest", Description: "8 hours (2h light activity, 6h sleep). Regain all HP + half max Hit Dice. Regain all abilities.", Reference: "PHB p.186"},
				{Name: "Healing", Description: "Restore HP equal to hit die roll + CON mod.", Reference: "PHB p.196"},
			},
		},
		{
			Title: "Travel Pace",
			Icon:  "fa-hiking",
			Entries: []models.DmQuickRefEntry{
				{Name: "Fast", Description: "400ft/min, 4mph, 30mi/day. -5 Passive Perception. Can't stealth.", Reference: "PHB p.182"},
				{Name: "Normal", Description: "300ft/min, 3mph, 24mi/day.", Reference: "PHB p.182"},
				{Name: "Slow", Description: "200ft/min, 2mph, 18mi/day. Can stealth. Can use Perception.", Reference: "PHB p.182"},
			},
		},
		{
			Title: "Light & Vision",
			Icon:  "fa-eye",
			Entries: []models.DmQuickRefEntry{
				{Name: "Bright Light", Description: "Normal vision. Torches (20ft), lanterns (30ft).", Reference: "PHB p.183"},
				{Name: "Dim Light", Description: "Lightly obscured. Disadvantage on Perception (sight)."},
				{Name: "Darkness", Description: "Heavily obscured. Blind. Darkvision ignores up to 60ft (dim) or 120ft (gray).", Reference: "PHB p.183"},
			},
		},
		{
			Title: "Cover",
			Icon:  "fa-shield-alt",
			Entries: []models.DmQuickRefEntry{
				{Name: "Half Cover", Description: "+2 AC and DEX saves. Behind obstacle, low wall, furniture.", Reference: "PHB p.196"},
				{Name: "Three-Quarters Cover", Description: "+5 AC and DEX saves. Behind pillar, battlement.", Reference: "PHB p.196"},
				{Name: "Total Cover", Description: "Can't be targeted directly.", Reference: "PHB p.196"},
			},
		},
		{
			Title: "Skills by Ability",
			Icon:  "fa-clipboard-list",
			Entries: []models.DmQuickRefEntry{
				{Name: "Strength", Description: "Athletics"},
				{Name: "Dexterity", Description: "Acrobatics, Sleight of Hand, Stealth"},
				{Name: "Constitution", Description: "No skills (but CON saves common)"},
				{Name: "Intelligence", Description: "Arcana, History, Investigation, Nature, Religion"},
				{Name: "Wisdom", Description: "Animal Handling, Insight, Medicine, Perception, Survival"},
				{Name: "Charisma", Description: "Deception, Intimidation, Performance, Persuasion"},
			},
		},
	}
}

func HtmxDmScreen(c *gin.Context) {
	adventureID := c.Param("id")
	userID, _ := c.Get("user_id")

	// Load adventure info
	var adv models.OneShotAdventure
	err := db.DB.QueryRow(`
		SELECT id, user_id, campaign_id, title, premise, hook, template, estimated_minutes, difficulty, notes, created_at, updated_at
		FROM oneshot_adventures WHERE id=? AND user_id=?
	`, adventureID, userID).Scan(&adv.ID, &adv.UserID, &adv.CampaignID, &adv.Title, &adv.Premise, &adv.Hook, &adv.Template, &adv.EstimatedMinutes, &adv.Difficulty, &adv.Notes, &adv.CreatedAt, &adv.UpdatedAt)
	if err != nil {
		c.String(http.StatusNotFound, "Adventure not found")
		return
	}

	// Load DM notes
	rows, err := db.DB.Query("SELECT id, adventure_id, user_id, title, content, created_at, updated_at FROM dm_notes WHERE adventure_id=? AND user_id=? ORDER BY updated_at DESC", adventureID, userID)
	var notes []models.DmNote
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var n models.DmNote
			if err := rows.Scan(&n.ID, &n.AdventureID, &n.UserID, &n.Title, &n.Content, &n.CreatedAt, &n.UpdatedAt); err == nil {
				notes = append(notes, n)
			}
		}
	}
	if notes == nil {
		notes = []models.DmNote{}
	}

	renderTemplate(c, "oneshot_dm_screen.html", gin.H{
		"Adventure": adv,
		"Sections":  QuickReferenceData(),
		"Notes":     notes,
	})
}

// ─── DM Notes API ───

func ListDmNotes(c *gin.Context) {
	adventureID := c.Param("id")
	userID, _ := c.Get("user_id")

	rows, err := db.DB.Query("SELECT id, adventure_id, user_id, title, content, created_at, updated_at FROM dm_notes WHERE adventure_id=? AND user_id=? ORDER BY updated_at DESC", adventureID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var notes []models.DmNote
	for rows.Next() {
		var n models.DmNote
		if err := rows.Scan(&n.ID, &n.AdventureID, &n.UserID, &n.Title, &n.Content, &n.CreatedAt, &n.UpdatedAt); err == nil {
			notes = append(notes, n)
		}
	}
	if notes == nil {
		notes = []models.DmNote{}
	}
	c.JSON(http.StatusOK, notes)
}

func CreateDmNote(c *gin.Context) {
	adventureID := c.Param("id")
	userID, _ := c.Get("user_id")

	var input struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		ActID   *int64 `json:"act_id,omitempty"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := db.DB.Exec("INSERT INTO dm_notes(adventure_id, user_id, title, content, act_id) VALUES(?,?,?,?,?)", adventureID, userID, input.Title, input.Content, input.ActID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func UpdateDmNote(c *gin.Context) {
	noteID := c.Param("nid")
	userID, _ := c.Get("user_id")

	var input struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := db.DB.Exec("UPDATE dm_notes SET title=?, content=?, updated_at=datetime('now') WHERE id=? AND user_id=?", input.Title, input.Content, noteID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func DeleteDmNote(c *gin.Context) {
	noteID := c.Param("nid")
	userID, _ := c.Get("user_id")

	_, err := db.DB.Exec("DELETE FROM dm_notes WHERE id=? AND user_id=?", noteID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// ─── One-Shot Items Section (HTMX) ───

func HtmxOneShotItems(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	rows, err := db.DB.Query("SELECT id, adventure_id, name, description, category, quantity, weight, price_gp, is_magical, attunement, notes, created_at FROM oneshot_items WHERE adventure_id=? ORDER BY name", id)
	if err != nil {
		c.String(http.StatusInternalServerError, "query error")
		return
	}
	defer rows.Close()
	out := make([]models.OneShotItem, 0)
	for rows.Next() {
		var it models.OneShotItem
		rows.Scan(&it.ID, &it.AdventureID, &it.Name, &it.Description, &it.Category, &it.Quantity, &it.Weight, &it.PriceGP, &it.IsMagical, &it.Attunement, &it.Notes, &it.CreatedAt)
		out = append(out, it)
	}
	renderTemplate(c, "oneshot_items_section.html", itemsSectionData{Items: out, AdventureID: id})
}

// ─── One-Shot Shops Section (HTMX) ───

func HtmxOneShotShops(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	rows, err := db.DB.Query("SELECT id, name, description, markup_percent, markup_buy_percent FROM shops WHERE oneshot_adventure_id=? ORDER BY name", id)
	if err != nil {
		c.String(http.StatusInternalServerError, "query error")
		return
	}
	defer rows.Close()
	type shopRow struct {
		ID            int64
		Name          string
		Description   string
		MarkupPercent float64
	}
	out := make([]shopRow, 0)
	for rows.Next() {
		var s shopRow
		var mbp float64
		rows.Scan(&s.ID, &s.Name, &s.Description, &s.MarkupPercent, &mbp)
		out = append(out, s)
	}
	renderTemplate(c, "oneshot_shops_section.html", shopsSectionData{Shops: out, AdventureID: id})
}

// ─── One-Shot Monsters Section (HTMX) ───

func HtmxOneShotMonsters(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	rows, err := db.DB.Query("SELECT id, adventure_id, COALESCE(act_id,0), COALESCE(scene_id,0), name, ac, hp, str, dex, con, int, wis, cha, cr, source, is_full, COALESCE(saves,''), COALESCE(skills,''), COALESCE(damage_vulnerabilities,''), COALESCE(damage_resistances,''), COALESCE(damage_immunities,''), COALESCE(condition_immunities,''), COALESCE(senses,''), COALESCE(languages,''), COALESCE(special_abilities,''), COALESCE(actions,''), COALESCE(legendary_actions,''), COALESCE(library_id,0), created_at FROM oneshot_monsters WHERE adventure_id=? ORDER BY name", id)
	if err != nil {
		c.String(http.StatusInternalServerError, "query error")
		return
	}
	defer rows.Close()
	out := make([]models.OneShotMonster, 0)
	for rows.Next() {
		var m models.OneShotMonster
		var actID, sceneID, libID int64
		rows.Scan(&m.ID, &m.AdventureID, &actID, &sceneID, &m.Name, &m.AC, &m.HP, &m.Str, &m.Dex, &m.Con, &m.Int, &m.Wis, &m.Cha, &m.CR, &m.Source, &m.IsFull, &m.Saves, &m.Skills, &m.DamageVulnerabilities, &m.DamageResistances, &m.DamageImmunities, &m.ConditionImmunities, &m.Senses, &m.Languages, &m.SpecialAbilities, &m.Actions, &m.LegendaryActions, &libID, &m.CreatedAt)
		if actID > 0 { m.ActID = &actID }
		if sceneID > 0 { m.SceneID = &sceneID }
		if libID > 0 { m.LibraryID = &libID }
		out = append(out, m)
	}
	renderTemplate(c, "oneshot_monsters_section.html", monstersSectionData{Monsters: out, AdventureID: id})
}

// ─── One-Shot Player Characters Section (HTMX) ───

func HtmxOneShotPCs(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	rows, err := db.DB.Query(`SELECT opc.id, opc.adventure_id, opc.character_id, opc.role, opc.notes,
		COALESCE(ch.name,''), COALESCE(u.username,'')
		FROM oneshot_player_characters opc
		LEFT JOIN characters ch ON opc.character_id = ch.id
		LEFT JOIN users u ON ch.user_id = u.id
		WHERE opc.adventure_id=? ORDER BY ch.name`, id)
	if err != nil {
		c.String(http.StatusInternalServerError, "query error")
		return
	}
	defer rows.Close()
	out := make([]models.OneShotPlayerCharacter, 0)
	for rows.Next() {
		var pc models.OneShotPlayerCharacter
		rows.Scan(&pc.ID, &pc.AdventureID, &pc.CharacterID, &pc.Role, &pc.Notes, &pc.CharName, &pc.Username)
		out = append(out, pc)
	}
	renderTemplate(c, "oneshot_pcs_section.html", pcsSectionData{PCs: out, AdventureID: id})
}

// ─── Data structs for section templates ───

type itemsSectionData struct {
	Items       []models.OneShotItem
	AdventureID int64
}

type shopsSectionData struct {
	Shops       interface{}
	AdventureID int64
}

type monstersSectionData struct {
	Monsters    []models.OneShotMonster
	AdventureID int64
}

type pcsSectionData struct {
	PCs         []models.OneShotPlayerCharacter
	AdventureID int64
}

// ─── Act NPC CRUD (ent-backed) ───

func ListActNPCs(c *gin.Context) {
	actID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid act id"})
		return
	}
	npcs, err := db.Client.OneShotActNPC.Query().Where(oneshotactnpc.ActIDEQ(actID)).Order(ent.Asc(oneshotactnpc.FieldName)).All(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]models.OneShotActNPC, len(npcs))
	for i, n := range npcs {
		out[i] = models.OneShotActNPC{
			ID:        n.ID,
			ActID:     n.ActID,
			NPCID:     nil,
			Name:      n.Name,
			Role:      n.Role,
			Notes:     n.Notes,
			IsInline:  n.IsInline,
			CreatedAt: n.CreatedAt,
		}
		if n.NpcID != 0 {
			v := n.NpcID
			out[i].NPCID = &v
		}
	}
	c.JSON(http.StatusOK, out)
}

func CreateActNPC(c *gin.Context) {
	actID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid act id"})
		return
	}
	var input struct {
		NpcID  *int64 `json:"npc_id,omitempty"`
		Name   string `json:"name"`
		Role   string `json:"role"`
		Notes  string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	q := db.Client.OneShotActNPC.Create().SetActID(actID).SetName(input.Name).SetRole(input.Role).SetNotes(input.Notes)
	if input.NpcID != nil {
		q.SetNpcID(*input.NpcID).SetIsInline(false)
	} else {
		q.SetIsInline(true)
	}
	n, err := q.Save(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := models.OneShotActNPC{
		ID:        n.ID,
		ActID:     n.ActID,
		Name:      n.Name,
		Role:      n.Role,
		Notes:     n.Notes,
		IsInline:  n.IsInline,
		CreatedAt: n.CreatedAt,
	}
	if n.NpcID != 0 {
		v := n.NpcID
		out.NPCID = &v
	}
	c.JSON(http.StatusCreated, out)
}

func DeleteActNPC(c *gin.Context) {
	actID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid act id"})
		return
	}
	npcID, err := strconv.ParseInt(c.Param("nid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid npc id"})
		return
	}
	err = db.Client.OneShotActNPC.DeleteOneID(npcID).Exec(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "act_id": actID})
}

// ─── Act Notes (raw SQL) ───

func ListActNotes(c *gin.Context) {
	actID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid act id"})
		return
	}
	rows, err := db.DB.Query("SELECT id, adventure_id, user_id, title, content, created_at, updated_at FROM dm_notes WHERE act_id=? ORDER BY updated_at DESC", actID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := make([]models.DmNote, 0)
	for rows.Next() {
		var n models.DmNote
		rows.Scan(&n.ID, &n.AdventureID, &n.UserID, &n.Title, &n.Content, &n.CreatedAt, &n.UpdatedAt)
		out = append(out, n)
	}
	c.JSON(http.StatusOK, out)
}

func CreateActNote(c *gin.Context) {
	actID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid act id"})
		return
	}
	var input struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	var adventureID int64
	err = db.DB.QueryRow("SELECT adventure_id FROM oneshot_acts WHERE id=?", actID).Scan(&adventureID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "act not found"})
		return
	}
	result, err := db.DB.Exec("INSERT INTO dm_notes(adventure_id, user_id, title, content, act_id) VALUES(?,?,?,?,?)", adventureID, userID, input.Title, input.Content, actID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// ─── HTMX Act Details ───

type actDetailsData struct {
	Act     models.OneShotAct
	NPCs    []models.OneShotActNPC
	Notes   []models.DmNote
}

func HtmxActDetails(c *gin.Context) {
	actID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid act id")
		return
	}

	var act models.OneShotAct
	err = db.DB.QueryRow("SELECT id, adventure_id, number, title, description, estimated_minutes, notes FROM oneshot_acts WHERE id=?", actID).
		Scan(&act.ID, &act.AdventureID, &act.Number, &act.Title, &act.Description, &act.EstimatedMinutes, &act.Notes)
	if err != nil {
		c.String(http.StatusNotFound, "act not found")
		return
	}

	// Load act NPCs via ent
	npcs, err := db.Client.OneShotActNPC.Query().Where(oneshotactnpc.ActIDEQ(actID)).Order(ent.Asc(oneshotactnpc.FieldName)).All(c.Request.Context())
	npcOut := make([]models.OneShotActNPC, 0)
	if err == nil {
		for _, n := range npcs {
			m := models.OneShotActNPC{
				ID:        n.ID,
				ActID:     n.ActID,
				Name:      n.Name,
				Role:      n.Role,
				Notes:     n.Notes,
				IsInline:  n.IsInline,
				CreatedAt: n.CreatedAt,
			}
			if n.NpcID != 0 {
				v := n.NpcID
				m.NPCID = &v
			}
			npcOut = append(npcOut, m)
		}
	}

	// Load act notes via raw SQL
	noteRows, err := db.DB.Query("SELECT id, adventure_id, user_id, title, content, created_at, updated_at FROM dm_notes WHERE act_id=? ORDER BY updated_at DESC", actID)
	noteOut := make([]models.DmNote, 0)
	if err == nil {
		defer noteRows.Close()
		for noteRows.Next() {
			var n models.DmNote
			noteRows.Scan(&n.ID, &n.AdventureID, &n.UserID, &n.Title, &n.Content, &n.CreatedAt, &n.UpdatedAt)
			noteOut = append(noteOut, n)
		}
	}

	renderTemplate(c, "oneshot_act_details.html", actDetailsData{
		Act:   act,
		NPCs:  npcOut,
		Notes: noteOut,
	})
}
