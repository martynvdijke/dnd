package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/ent/campaign"
	"villum/ent/campaigncalendarevent"
	"villum/ent/campaigntimelineevent"
	"villum/ent/campaignwikipage"
	"villum/ent/character"
	"villum/ent/characterlocation"
	"villum/ent/characternpc"
	"villum/ent/encountertemplate"
	"villum/ent/faction"
	"villum/ent/factionreputation"
	"villum/ent/quest"
	"villum/ent/session"
	"villum/models"
)

const campaignGraphLimit = 50

func GetGraphData(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	gd := models.GraphData{Nodes: []models.GraphNode{}, Edges: []models.GraphEdge{}}
	char, err := db.Client.Character.Get(c.Request.Context(), charID)
	if err == nil {
		gd.Nodes = append(gd.Nodes, models.GraphNode{ID: "char_" + strconv.FormatInt(charID, 10), Label: char.Name + " (" + char.Race + " " + char.Class + ")", Group: "character", Color: "#8b0000", Size: 30, CharID: charID})
	}
	cls, _ := db.Client.CharacterLocation.Query().Where(characterlocation.CharacterID(charID)).WithLocation().All(c.Request.Context())
	for _, cl := range cls {
		loc := cl.Edges.Location
		nid := "loc_" + strconv.FormatInt(loc.ID, 10)
		gd.Nodes = append(gd.Nodes, models.GraphNode{ID: nid, Label: loc.Name + " (" + loc.Type + ")", Group: "location", Color: "#b8963e", Size: 20})
		gd.Edges = append(gd.Edges, models.GraphEdge{From: "char_" + strconv.FormatInt(charID, 10), To: nid, Label: cl.Relationship, Width: 2, Dashes: false})
	}
	cnpcs, _ := db.Client.CharacterNPC.Query().Where(characternpc.CharacterID(charID)).WithNpc().All(c.Request.Context())
	for _, cn := range cnpcs {
		npcEnt := cn.Edges.Npc
		nid := "npc_" + strconv.FormatInt(npcEnt.ID, 10)
		edgeWidth := npcEdgeWidth(cn.InteractionCount)
		gd.Nodes = append(gd.Nodes, models.GraphNode{ID: nid, Label: npcEnt.Name + " (NPC)", Group: "npc", Color: "#2c6b2f", Size: 20})
		gd.Edges = append(gd.Edges, models.GraphEdge{From: "char_" + strconv.FormatInt(charID, 10), To: nid, Label: cn.Relationship, Width: edgeWidth, Dashes: false})
	}
	quests, _ := db.Client.Quest.Query().Where(quest.CharacterID(charID)).All(c.Request.Context())
	for _, q := range quests {
		nid := "quest_" + strconv.FormatInt(q.ID, 10)
		qcolor := questColor(q.Status)
		gd.Nodes = append(gd.Nodes, models.GraphNode{ID: nid, Label: q.Name + " [" + q.Status + "]", Group: "quest", Color: qcolor, Size: 18})
		gd.Edges = append(gd.Edges, models.GraphEdge{From: "char_" + strconv.FormatInt(charID, 10), To: nid, Label: q.Status, Width: 1, Dashes: q.Status == "available"})
	}
	sessions, _ := db.Client.Session.Query().Where(session.CharacterID(charID)).Order(session.BySessionDate()).Limit(10).All(c.Request.Context())
	for _, s := range sessions {
		nid := "session_" + strconv.FormatInt(s.ID, 10)
		slabel := s.Title
		if slabel == "" {
			slabel = "Session " + s.SessionDate
		}
		gd.Nodes = append(gd.Nodes, models.GraphNode{ID: nid, Label: slabel, Group: "session", Color: "#5c3a2a", Size: 15})
		gd.Edges = append(gd.Edges, models.GraphEdge{From: "char_" + strconv.FormatInt(charID, 10), To: nid, Label: "played", Width: 1, Dashes: false})
	}
	WriteJSON(c, http.StatusOK, gd)
}

func GetCampaignGraphData(c *gin.Context) {
	campaignID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	role, _ := c.Get("role")
	gd := models.GraphData{Nodes: []models.GraphNode{}, Edges: []models.GraphEdge{}}
	nodeSet := map[string]bool{}
	camp, _ := db.Client.Campaign.Get(c.Request.Context(), campaignID)
	if camp != nil {
		gd.Nodes = append(gd.Nodes, models.GraphNode{ID: "camp_" + strconv.FormatInt(campaignID, 10), Label: camp.Name, Group: "campaign", Color: "#8b0000", Size: 35})
	}
	nodeSet["camp_"+strconv.FormatInt(campaignID, 10)] = true
	if role != "admin" {
		if !IsCampaignMemberGin(c, campaignID) {
			WriteError(c, http.StatusForbidden, errAccessDenied)
			return
		}
	}
	// Characters
	chars, _ := db.Client.Character.Query().Where(character.CampaignID(campaignID)).All(c.Request.Context())
	for _, ch := range chars {
		nid := "char_" + strconv.FormatInt(ch.ID, 10)
		addNode(&gd, nodeSet, nid, ch.Name+" (Lvl "+strconv.Itoa(ch.Level)+" "+ch.Race+" "+ch.Class+")", "character", "#8b0000", 28, ch.ID)
		gd.Edges = append(gd.Edges, models.GraphEdge{From: "camp_" + strconv.FormatInt(campaignID, 10), To: nid, Label: "member", Width: 2})
	}
	// Wiki
	wikiPages, _ := db.Client.CampaignWikiPage.Query().Where(campaignwikipage.CampaignID(campaignID)).Order(campaignwikipage.BySortOrder(), campaignwikipage.ByTitle()).All(c.Request.Context())
	wikiMap := map[int64]string{}
	for _, wp := range wikiPages {
		nid := "wiki_" + strconv.FormatInt(wp.ID, 10)
		addNode(&gd, nodeSet, nid, wp.Title, "wiki", "#b8963e", 18, 0)
		wikiMap[wp.ID] = nid
		gd.Edges = append(gd.Edges, models.GraphEdge{From: "camp_" + strconv.FormatInt(campaignID, 10), To: nid, Label: "wiki", Width: 1})
		if wp.ParentID != 0 {
			if pnid, ok := wikiMap[wp.ParentID]; ok {
				gd.Edges = append(gd.Edges, models.GraphEdge{From: pnid, To: nid, Label: "child", Width: 1, Dashes: true})
			}
		}
	}
	// Locations
	cls, _ := db.Client.CharacterLocation.Query().Where(characterlocation.HasCharacterWith(character.CampaignID(campaignID))).WithLocation().All(c.Request.Context())
	for _, cl := range cls {
		loc := cl.Edges.Location
		nid := "loc_" + strconv.FormatInt(loc.ID, 10)
		addNode(&gd, nodeSet, nid, loc.Name+" ("+loc.Type+")", "location", "#b8963e", 20, 0)
	}
	for _, cl := range cls {
		from := "char_" + strconv.FormatInt(cl.CharacterID, 10)
		to := "loc_" + strconv.FormatInt(cl.LocationID, 10)
		if nodeSet[from] && nodeSet[to] {
			gd.Edges = append(gd.Edges, models.GraphEdge{From: from, To: to, Label: cl.Relationship, Width: 2})
		}
	}
	// NPCs
	cnpcs, _ := db.Client.CharacterNPC.Query().Where(characternpc.HasCharacterWith(character.CampaignID(campaignID))).WithNpc().All(c.Request.Context())
	for _, cn := range cnpcs {
		npcEnt := cn.Edges.Npc
		nid := "npc_" + strconv.FormatInt(npcEnt.ID, 10)
		label := npcEnt.Name + " (NPC)"
		if npcEnt.Race != "" || npcEnt.Class != "" {
			label = npcEnt.Name + " (" + npcEnt.Race + " " + npcEnt.Class + ")"
		}
		addNode(&gd, nodeSet, nid, label, "npc", "#2d6a2d", 20, 0)
	}
	for _, cn := range cnpcs {
		from := "char_" + strconv.FormatInt(cn.CharacterID, 10)
		to := "npc_" + strconv.FormatInt(cn.NpcID, 10)
		if nodeSet[from] && nodeSet[to] {
			gd.Edges = append(gd.Edges, models.GraphEdge{From: from, To: to, Label: cn.Relationship, Width: npcEdgeWidth(cn.InteractionCount)})
		}
	}
	// Quests
	quests, _ := db.Client.Quest.Query().Where(quest.HasCharacterWith(character.CampaignID(campaignID))).All(c.Request.Context())
	for _, q := range quests {
		nid := "quest_" + strconv.FormatInt(q.ID, 10)
		addNode(&gd, nodeSet, nid, q.Name+" ["+q.Status+"]", "quest", questColor(q.Status), 18, 0)
		from := "char_" + strconv.FormatInt(q.CharacterID, 10)
		if nodeSet[from] {
			gd.Edges = append(gd.Edges, models.GraphEdge{From: from, To: nid, Label: q.Status, Width: 1, Dashes: q.Status == "available"})
		}
	}
	// Sessions
	sessions, _ := db.Client.Session.Query().Where(session.HasCharacterWith(character.CampaignID(campaignID))).Order(session.BySessionDate()).Limit(30).All(c.Request.Context())
	for _, s := range sessions {
		nid := "session_" + strconv.FormatInt(s.ID, 10)
		slabel := s.Title
		if slabel == "" {
			slabel = "Session " + s.SessionDate
		}
		addNode(&gd, nodeSet, nid, slabel, "session", "#5c3a2a", 14, 0)
		from := "char_" + strconv.FormatInt(s.CharacterID, 10)
		if nodeSet[from] {
			gd.Edges = append(gd.Edges, models.GraphEdge{From: from, To: nid, Label: "played", Width: 1})
		}
	}
	// Factions
	factions, _ := db.Client.Faction.Query().Where(faction.CampaignIDIn(campaignID, 0)).All(c.Request.Context())
	factionIDs := []int64{}
	for _, f := range factions {
		nid := "faction_" + strconv.FormatInt(f.ID, 10)
		addNode(&gd, nodeSet, nid, f.Name+" ("+f.Type+")", "faction", "#9b59b6", 20, 0)
		factionIDs = append(factionIDs, f.ID)
	}
	for _, fid := range factionIDs {
		frs, _ := db.Client.FactionReputation.Query().Where(factionreputation.And(factionreputation.FactionID(fid), factionreputation.HasCharacterWith(character.CampaignID(campaignID)))).All(c.Request.Context())
		for _, fr := range frs {
			from := "char_" + strconv.FormatInt(fr.CharacterID, 10)
			to := "faction_" + strconv.FormatInt(fid, 10)
			relLabel := standingLabel(fr.Standing)
			if nodeSet[from] && nodeSet[to] {
				gd.Edges = append(gd.Edges, models.GraphEdge{From: from, To: to, Label: relLabel + " (" + strconv.Itoa(fr.Standing) + ")", Width: 2})
			}
		}
	}
	// Encounters
	encs, _ := db.Client.EncounterTemplate.Query().Where(encountertemplate.CampaignID(campaignID)).All(c.Request.Context())
	for _, e := range encs {
		nid := "encounter_" + strconv.FormatInt(e.ID, 10)
		encColor := "#e67e22"
		if e.Difficulty == "hard" {
			encColor = "#c0392b"
		} else if e.Difficulty == "deadly" {
			encColor = "#7b0000"
		}
		addNode(&gd, nodeSet, nid, e.Name+" ["+e.Difficulty+"]", "encounter", encColor, 18, 0)
		gd.Edges = append(gd.Edges, models.GraphEdge{From: "camp_" + strconv.FormatInt(campaignID, 10), To: nid, Label: "encounter", Width: 1})
	}
	// Timeline
	tls, _ := db.Client.CampaignTimelineEvent.Query().Where(campaigntimelineevent.And(campaigntimelineevent.CampaignID(campaignID), campaigntimelineevent.LinkedEntityTypeNEQ(""))).All(c.Request.Context())
	for _, tl := range tls {
		nid := "timeline_" + strconv.FormatInt(tl.ID, 10)
		addNode(&gd, nodeSet, nid, tl.Title+" ["+tl.EventType+"]", "timeline", "#5c3a2a", 14, 0)
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
			gd.Edges = append(gd.Edges, models.GraphEdge{From: targetNID, To: nid, Label: tl.EventType, Width: 1, Dashes: true})
		}
		gd.Edges = append(gd.Edges, models.GraphEdge{From: "camp_" + strconv.FormatInt(campaignID, 10), To: nid, Label: "timeline", Width: 1})
	}
	cals, _ := db.Client.CampaignCalendarEvent.Query().Where(campaigncalendarevent.CampaignID(campaignID)).Limit(30).All(c.Request.Context())
	for _, cal := range cals {
		nid := "calendar_" + strconv.FormatInt(cal.ID, 10)
		addNode(&gd, nodeSet, nid, cal.Title+" ["+cal.EventType+"]", "calendar", "#b8963e", 14, 0)
		gd.Edges = append(gd.Edges, models.GraphEdge{From: "camp_" + strconv.FormatInt(campaignID, 10), To: nid, Label: "event", Width: 1})
	}
	gd.Nodes = deduplicateGraphNodes(gd.Nodes)
	gd.Edges = deduplicateGraphEdges(gd.Edges)
	WriteJSON(c, http.StatusOK, gd)
}

func addNode(gd *models.GraphData, set map[string]bool, id, label, group, color string, size int, charID int64) {
	if set[id] {
		return
	}
	gd.Nodes = append(gd.Nodes, models.GraphNode{ID: id, Label: label, Group: group, Color: color, Size: size, CharID: charID})
	set[id] = true
}

func deduplicateGraphNodes(nodes []models.GraphNode) []models.GraphNode {
	seen := map[string]bool{}
	out := make([]models.GraphNode, 0, len(nodes))
	for _, n := range nodes {
		if !seen[n.ID] {
			seen[n.ID] = true
			out = append(out, n)
		}
	}
	return out
}

func deduplicateGraphEdges(edges []models.GraphEdge) []models.GraphEdge {
	seen := map[string]bool{}
	out := make([]models.GraphEdge, 0, len(edges))
	for _, e := range edges {
		key := e.From + "->" + e.To + ":" + e.Label
		if !seen[key] {
			seen[key] = true
			out = append(out, e)
		}
	}
	return out
}

func npcEdgeWidth(count int) int {
	if count > 5 {
		return 5
	} else if count > 2 {
		return 3
	} else if count > 0 {
		return 2
	}
	return 1
}

func questColor(status string) string {
	if status == "complete" {
		return "#2d6a2d"
	} else if status == "failed" || status == "abandoned" {
		return "#666"
	}
	return "#b8963e"
}

func standingLabel(standing int) string {
	if standing >= 50 {
		return "revered"
	} else if standing >= 25 {
		return "allied"
	} else if standing <= -50 {
		return "hostile"
	} else if standing <= -25 {
		return "unfriendly"
	}
	return "neutral"
}

var _ = campaign.FieldID
