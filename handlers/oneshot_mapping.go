package handlers

import (
	"context"
	"database/sql"
	"sort"

	"villum/db"
	"villum/ent"
	"villum/ent/oneshotact"
	"villum/ent/oneshotadventureencounter"
	"villum/ent/oneshotitem"
	"villum/ent/oneshotscene"
	"villum/models"
)

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
			if s.LocationID != 0 {
				lid := s.LocationID
				ms.LocationID = &lid
			}
			if s.EncounterID != 0 {
				eid := s.EncounterID
				ms.EncounterID = &eid
			}
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
			if it.ActID != 0 {
				aid := it.ActID
				mi.ActID = &aid
			}
			m.Items = append(m.Items, mi)
		}
	}
	if len(e.Edges.Encounters) > 0 {
		m.Encounters = make([]models.OneShotAdventureEncounter, 0)
		for _, enc := range e.Edges.Encounters {
			me := models.OneShotAdventureEncounter{
				ID: enc.ID, AdventureID: enc.AdventureID, EncounterID: enc.EncounterID,
			}
			if enc.ActID != 0 {
				aid := enc.ActID
				me.ActID = &aid
			}
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
	if e.CampaignID != 0 {
		cid := e.CampaignID
		m.CampaignID = &cid
	}
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
	rows, err := db.DB.Query("SELECT oan.id, oan.adventure_id, oan.npc_id, oan.role, oan.story_hook, oan.combat_ready, COALESCE(n.name,'') FROM oneshot_adventure_npcs oan LEFT JOIN npcs n ON oan.npc_id=n.id WHERE oan.adventure_id=?", adventureID)
	if err != nil {
		return nil
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
	return out
}

func loadAdventureLocations(adventureID int64) []models.OneShotAdventureLocation {
	rows, err := db.DB.Query("SELECT oal.id, oal.adventure_id, oal.location_id, COALESCE(l.name,'') FROM oneshot_adventure_locations oal LEFT JOIN locations l ON oal.location_id=l.id WHERE oal.adventure_id=?", adventureID)
	if err != nil {
		return nil
	}
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
	if err != nil {
		return nil
	}
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
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := make([]models.OneShotShop, 0)
	for rows.Next() {
		var s models.OneShotShop
		var campID, aID sql.NullInt64
		rows.Scan(&s.ID, &s.UserID, &campID, &s.OneshotAdventureID, &aID, &s.Name, &s.Description, &s.MarkupPercent, &s.MarkupBuyPercent, &s.CreatedAt)
		if campID.Valid {
			s.CampaignID = &campID.Int64
		}
		if aID.Valid {
			s.ActID = &aID.Int64
		}
		out = append(out, s)
	}
	return out
}

func loadAdventureItems(adventureID int64) []models.OneShotItem {
	rows, err := db.DB.Query("SELECT id, adventure_id, act_id, name, description, category, quantity, weight, price_gp, is_magical, attunement, notes, created_at FROM oneshot_items WHERE adventure_id=? ORDER BY name", adventureID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := make([]models.OneShotItem, 0)
	for rows.Next() {
		var it models.OneShotItem
		var isMag, att int
		var aID sql.NullInt64
		rows.Scan(&it.ID, &it.AdventureID, &aID, &it.Name, &it.Description, &it.Category, &it.Quantity, &it.Weight, &it.PriceGP, &isMag, &att, &it.Notes, &it.CreatedAt)
		it.IsMagical = isMag == 1
		it.Attunement = att == 1
		if aID.Valid {
			it.ActID = &aID.Int64
		}
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
		WithScenes(func(q *ent.OneShotSceneQuery) {
			q.Order(ent.Asc(oneshotscene.FieldSortOrder), ent.Asc(oneshotscene.FieldID))
		}).
		WithItems(func(q *ent.OneShotItemQuery) {
			q.Order(ent.Asc(oneshotitem.FieldName))
		}).
		WithEncounters(func(q *ent.OneShotAdventureEncounterQuery) {
			q.Order(ent.Asc(oneshotadventureencounter.FieldID))
		}).
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

	var isMiniCampaign int
	db.DB.QueryRow("SELECT COALESCE(is_mini_campaign,0) FROM oneshot_adventures WHERE id=?", adventureID).Scan(&isMiniCampaign)
	a.IsMiniCampaign = isMiniCampaign == 1
	var sortOrder int
	db.DB.QueryRow("SELECT COALESCE(sort_order,0) FROM oneshot_adventures WHERE id=?", adventureID).Scan(&sortOrder)
	a.SortOrder = sortOrder

	return &a, nil
}

// ─── One-Shot Adventure API Handlers ───

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
