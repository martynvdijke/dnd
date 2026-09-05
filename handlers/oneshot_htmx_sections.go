package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/ent"
	"villum/ent/oneshotactnpc"
	"villum/middleware"
	"villum/models"
)

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

// ─── One-Shot NPCs Section (HTMX) ───

func HtmxOneShotNPCs(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	rows, err := db.DB.Query("SELECT oan.id, oan.adventure_id, oan.npc_id, oan.role, oan.story_hook, oan.combat_ready, COALESCE(n.name,'') FROM oneshot_adventure_npcs oan LEFT JOIN npcs n ON oan.npc_id=n.id WHERE oan.adventure_id=? ORDER BY oan.id", id)
	if err != nil {
		c.String(http.StatusInternalServerError, "query error")
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
	renderTemplate(c, "oneshot_npcs_section.html", npcsSectionData{NPCs: out, AdventureID: id})
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
	rows, err := db.DB.Query("SELECT id, adventure_id, COALESCE(act_id,0), COALESCE(scene_id,0), name, ac, hp, str, dex, con, int_, wis, cha, cr, source, is_full, COALESCE(saves,''), COALESCE(skills,''), COALESCE(damage_vulnerabilities,''), COALESCE(damage_resistances,''), COALESCE(damage_immunities,''), COALESCE(condition_immunities,''), COALESCE(senses,''), COALESCE(languages,''), COALESCE(special_abilities,''), COALESCE(actions,''), COALESCE(legendary_actions,''), COALESCE(library_id,0), COALESCE(compendium_monster_id,0), created_at FROM oneshot_monsters WHERE adventure_id=? ORDER BY name", id)
	if err != nil {
		middleware.LogError("oneshot", "HtmxOneShotMonsters query failed", "adventure_id", id, "error", err)
		c.String(http.StatusInternalServerError, "query error")
		return
	}
	defer rows.Close()
	out := make([]models.OneShotMonster, 0)
	for rows.Next() {
		var m models.OneShotMonster
		var actID, sceneID, libID, compID int64
		if err := rows.Scan(&m.ID, &m.AdventureID, &actID, &sceneID, &m.Name, &m.AC, &m.HP, &m.Str, &m.Dex, &m.Con, &m.Int, &m.Wis, &m.Cha, &m.CR, &m.Source, &m.IsFull, &m.Saves, &m.Skills, &m.DamageVulnerabilities, &m.DamageResistances, &m.DamageImmunities, &m.ConditionImmunities, &m.Senses, &m.Languages, &m.SpecialAbilities, &m.Actions, &m.LegendaryActions, &libID, &compID, &m.CreatedAt); err != nil {
			middleware.LogWarn("oneshot", "HtmxOneShotMonsters scan failed, skipping monster", "adventure_id", id, "error", err)
			continue
		}
		if actID > 0 {
			m.ActID = &actID
		}
		if sceneID > 0 {
			m.SceneID = &sceneID
		}
		if libID > 0 {
			m.LibraryID = &libID
		}
		if compID > 0 {
			m.CompendiumMonsterID = &compID
		}
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

type npcsSectionData struct {
	NPCs        []models.OneShotAdventureNPC
	AdventureID int64
}

type itemsSectionData struct {
	Items       []models.OneShotItem
	AdventureID int64
}

type shopsSectionData struct {
	Shops       any
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
