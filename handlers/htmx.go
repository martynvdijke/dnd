package handlers

import (
	"embed"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

//go:embed templates/*.html
var htmxTemplatesFS embed.FS

var htmxTemplates *template.Template

func init() {
	funcMap := template.FuncMap{
		"title": strings.Title,
		"lower": strings.ToLower,
		"seq":   seq,
		"mul":   func(a, b int) int { return a * b },
		"div": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a / b
		},
		"sub":        func(a, b int) int { return a - b },
		"add":        func(a, b int) int { return a + b },
		"truncate":   truncate,
		"capitalize": strings.Title,
		"sign": func(n int) string {
			if n >= 0 {
				return "+" + strconv.Itoa(n)
			}
			return strconv.Itoa(n)
		},
		"countRevealed": func(clues any) int {
			cls, ok := clues.([]models.Clue)
			if !ok {
				return 0
			}
			n := 0
			for _, c := range cls {
				if c.IsRevealed {
					n++
				}
			}
			return n
		},
		"countHidden": func(clues any) int {
			cls, ok := clues.([]models.Clue)
			if !ok {
				return 0
			}
			n := 0
			for _, c := range cls {
				if !c.IsRevealed {
					n++
				}
			}
			return n
		},
		"countRedHerrings": func(clues any) int {
			cls, ok := clues.([]models.Clue)
			if !ok {
				return 0
			}
			n := 0
			for _, c := range cls {
				if c.IsRedHerring {
					n++
				}
			}
			return n
		},
	}
	htmxTemplates = template.Must(template.New("").Funcs(funcMap).ParseFS(htmxTemplatesFS, "templates/*.html"))
}

func seq(from, to int) []int {
	var out []int
	for i := from; i <= to; i++ {
		out = append(out, i)
	}
	return out
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func renderTemplate(c *gin.Context, name string, data any) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := htmxTemplates.ExecuteTemplate(c.Writer, name, data); err != nil {
		c.String(http.StatusInternalServerError, "template error: %v", err)
	}
}

func getInt64Param(c *gin.Context, name string) (int64, error) {
	return strconv.ParseInt(c.PostForm(name), 10, 64)
}

func getQueryInt64(c *gin.Context, name string) (int64, error) {
	return strconv.ParseInt(c.Query(name), 10, 64)
}

func getIntParam(c *gin.Context, name string, def int) int {
	v := c.PostForm(name)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func getFloatParam(c *gin.Context, name string, def float64) float64 {
	v := c.PostForm(name)
	if v == "" {
		return def
	}
	n, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return def
	}
	return n
}

// ─── Notes ───

type htmxNoteData struct {
	CharacterID int64
	Note        *models.CharacterNote
	Notes       []models.CharacterNote
	Grouped     map[string][]models.CharacterNote
}

func HtmxMediaGallery(c *gin.Context) {
	ownerType := c.Query("owner_type")
	ownerID := c.Query("owner_id")
	if ownerType == "" || ownerID == "" {
		c.String(http.StatusBadRequest, "owner_type and owner_id required")
		return
	}

	rows, err := db.DB.Query(`
		SELECT u.id, u.hash, u.ext, u.url, COALESCE(u.resized_url,''), COALESCE(u.thumbnail_url,''),
			u.owner_type, u.owner_id, COALESCE(u.created_at,''), ul.id as link_id
		FROM uploads u
		JOIN upload_links ul ON u.id = ul.upload_id
		WHERE ul.entity_type = ? AND ul.entity_id = ?
		ORDER BY u.created_at DESC`, ownerType, ownerID)
	if err != nil {
		// Fallback: query by owner_type/owner_id directly
		rows, err = db.DB.Query(`
			SELECT u.id, u.hash, u.ext, u.url, COALESCE(u.resized_url,''), COALESCE(u.thumbnail_url,''),
				u.owner_type, u.owner_id, COALESCE(u.created_at,''), 0 as link_id
			FROM uploads u
			WHERE u.owner_type = ? AND u.owner_id = ?
			ORDER BY u.created_at DESC`, ownerType, ownerID)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
	}
	defer rows.Close()

	var uploads []mediaUploadItem
	for rows.Next() {
		var item mediaUploadItem
		var ownerTypeStr, ownerIDStr string
		rows.Scan(&item.ID, &item.Hash, &item.Ext, &item.URL, &item.ResizedURL,
			&item.ThumbnailURL, &ownerTypeStr, &ownerIDStr, &item.CreatedAt, &item.LinkID)
		item.OwnerType = ownerTypeStr
		if parsed, err := strconv.ParseInt(ownerIDStr, 10, 64); err == nil {
			item.OwnerID = parsed
		} else if ownerIDStr != "" {
			// corrupt owner_id in DB — treat as 0 and continue
			item.OwnerID = 0
		}
		item.IsPDF = item.Ext == ".pdf"
		uploads = append(uploads, item)
	}

	renderTemplate(c, "media_gallery.html", htmxMediaGalleryData{
		OwnerType: ownerType,
		OwnerID:   ownerID,
		Uploads:   uploads,
	})
}

// ─── Settings export for use in main.go ───

func HtmxRegisterRoutes(r *gin.RouterGroup) {
	routes := []Route{

		// Notes
		{"GET", "/htmx/notes", HtmxListNotes},
		{"GET", "/htmx/notes/new", HtmxNewNoteForm},
		{"GET", "/htmx/notes/:id/edit", HtmxEditNoteForm},
		{"POST", "/htmx/notes", HtmxCreateNote},
		{"PUT", "/htmx/notes/:id", HtmxUpdateNote},
		{"DELETE", "/htmx/notes/:id", HtmxDeleteNote},

		// Feats
		{"GET", "/htmx/feats", HtmxListFeats},
		{"GET", "/htmx/feats/new", HtmxNewFeatForm},
		{"GET", "/htmx/feats/:id/edit", HtmxEditFeatForm},
		{"POST", "/htmx/feats", HtmxCreateFeat},
		{"PUT", "/htmx/feats/:id", HtmxUpdateFeat},
		{"DELETE", "/htmx/feats/:id", HtmxDeleteFeat},

		// Conditions
		{"GET", "/htmx/conditions", HtmxListConditions},
		{"GET", "/htmx/conditions/new", HtmxNewConditionForm},
		{"GET", "/htmx/conditions/:id/edit", HtmxEditConditionForm},
		{"POST", "/htmx/conditions", HtmxCreateCondition},
		{"PUT", "/htmx/conditions/:id", HtmxUpdateCondition},
		{"DELETE", "/htmx/conditions/:id", HtmxDeleteCondition},

		// Companions
		{"GET", "/htmx/companions", HtmxListCompanions},
		{"GET", "/htmx/companions/new", HtmxNewCompanionForm},
		{"GET", "/htmx/companions/:id/edit", HtmxEditCompanionForm},
		{"POST", "/htmx/companions", HtmxCreateCompanion},
		{"PUT", "/htmx/companions/:id", HtmxUpdateCompanion},
		{"DELETE", "/htmx/companions/:id", HtmxDeleteCompanion},

		// Features
		{"GET", "/htmx/features", HtmxListFeatures},
		{"GET", "/htmx/features/new", HtmxNewFeatureForm},
		{"POST", "/htmx/features", HtmxCreateFeature},
		{"DELETE", "/htmx/features/:id", HtmxDeleteFeature},

		// Proficiencies
		{"GET", "/htmx/proficiencies/new", HtmxNewProficiencyForm},
		{"POST", "/htmx/proficiencies", HtmxCreateProficiency},
		{"DELETE", "/htmx/proficiencies/:id", HtmxDeleteProficiency},

		// Inventory
		{"GET", "/htmx/inventory", HtmxListInventory},
		{"GET", "/htmx/inventory/new", HtmxNewInventoryForm},
		{"GET", "/htmx/inventory/:id/edit", HtmxEditInventoryForm},
		{"POST", "/htmx/inventory", HtmxCreateInventory},
		{"PUT", "/htmx/inventory/:id", HtmxUpdateInventory},
		{"DELETE", "/htmx/inventory/:id", HtmxDeleteInventory},

		// Spells
		{"GET", "/htmx/spells", HtmxListSpells},
		{"GET", "/htmx/spells/new", HtmxNewSpellForm},
		{"GET", "/htmx/spells/:id/edit", HtmxEditSpellForm},
		{"POST", "/htmx/spells", HtmxCreateSpell},
		{"PUT", "/htmx/spells/:id", HtmxUpdateSpell},
		{"DELETE", "/htmx/spells/:id", HtmxDeleteSpell},

		// NPCs
		{"GET", "/htmx/npcs", HtmxListNPCs},
		{"GET", "/htmx/npcs/new", HtmxNewNPCForm},
		{"GET", "/htmx/npcs/link", HtmxLinkNPCForm},
		{"POST", "/htmx/npcs", HtmxCreateNPC},
		{"POST", "/htmx/npcs/link", HtmxLinkNPC},
		{"DELETE", "/htmx/npcs/link/:id", HtmxUnlinkNPC},

		// Locations
		{"GET", "/htmx/locations", HtmxListLocations},
		{"GET", "/htmx/locations/new", HtmxNewLocationForm},
		{"GET", "/htmx/locations/link", HtmxLinkLocationForm},
		{"POST", "/htmx/locations", HtmxCreateLocation},
		{"POST", "/htmx/locations/link", HtmxLinkLocation},
		{"DELETE", "/htmx/locations/link/:id", HtmxUnlinkLocation},

		// Sessions
		{"GET", "/htmx/sessions", HtmxListSessions},
		{"GET", "/htmx/sessions/new", HtmxNewSessionForm},
		{"GET", "/htmx/sessions/:id/edit", HtmxEditSessionForm},
		{"POST", "/htmx/sessions", HtmxCreateSession},
		{"PUT", "/htmx/sessions/:id", HtmxUpdateSession},
		{"DELETE", "/htmx/sessions/:id", HtmxDeleteSession},

		// Quests
		{"GET", "/htmx/quests", HtmxListQuests},
		{"GET", "/htmx/quests/new", HtmxNewQuestForm},
		{"GET", "/htmx/quests/:id/edit", HtmxEditQuestForm},
		{"POST", "/htmx/quests", HtmxCreateQuest},
		{"PUT", "/htmx/quests/:id", HtmxUpdateQuest},
		{"DELETE", "/htmx/quests/:id", HtmxDeleteQuest},

		// Journal
		{"GET", "/htmx/journal", HtmxListJournal},
		{"GET", "/htmx/journal/new", HtmxNewJournalForm},
		{"GET", "/htmx/journal/:id/edit", HtmxEditJournalForm},
		{"POST", "/htmx/journal", HtmxCreateJournal},
		{"PUT", "/htmx/journal/:id", HtmxUpdateJournal},
		{"DELETE", "/htmx/journal/:id", HtmxDeleteJournal},

		// Timeline
		{"GET", "/htmx/timeline", HtmxListTimeline},
		{"GET", "/htmx/timeline/new", HtmxNewTimelineForm},
		{"GET", "/htmx/timeline/:id/edit", HtmxEditTimelineForm},
		{"POST", "/htmx/timeline", HtmxCreateTimeline},
		{"PUT", "/htmx/timeline/:id", HtmxUpdateTimeline},
		{"DELETE", "/htmx/timeline/:id", HtmxDeleteTimeline},

		// Factions
		{"GET", "/htmx/factions", HtmxListFactions},
		{"GET", "/htmx/factions/new", HtmxNewFactionForm},
		{"GET", "/htmx/factions/:id/edit", HtmxEditFactionForm},
		{"POST", "/htmx/factions", HtmxCreateFaction},
		{"PUT", "/htmx/factions/:id", HtmxUpdateFaction},
		{"DELETE", "/htmx/factions/:id", HtmxDeleteFaction},

		// Media Gallery
		{"GET", "/htmx/media-gallery", HtmxMediaGallery},

		// Campaign Overview
		{"GET", "/htmx/campaigns/:id/overview", HtmxCampaignOverview},

		// One-Shot Adventures
		{"GET", "/htmx/oneshot-adventures", HtmxListOneShots},
		{"GET", "/htmx/oneshot-adventures/new", HtmxNewOneShotForm},
		{"GET", "/htmx/oneshot-adventures/:id", HtmxGetOneShotDetail},
		{"GET", "/htmx/oneshot-adventures/:id/edit", HtmxEditOneShotForm},
		{"POST", "/htmx/oneshot-adventures", HtmxCreateOneShot},
		{"POST", "/htmx/oneshot-adventures/generate", HtmxGenerateOneShot},
		{"PUT", "/htmx/oneshot-adventures/:id", HtmxUpdateOneShot},
		{"DELETE", "/htmx/oneshot-adventures/:id", HtmxDeleteOneShot},
		{"POST", "/htmx/oneshot-adventures/:id/acts", HtmxCreateAct},
		{"DELETE", "/htmx/oneshot-acts/:id", HtmxDeleteAct},
		{"POST", "/htmx/oneshot-acts/:id/scenes", HtmxCreateScene},
		{"DELETE", "/htmx/oneshot-scenes/:id", HtmxDeleteScene},
		{"GET", "/htmx/oneshot-adventures/:id/new-act-form", HtmxNewActForm},
		{"GET", "/htmx/oneshot-acts/:id/new-scene-form", HtmxSceneForm},
		{"GET", "/htmx/oneshot-acts/:id/edit", HtmxEditActForm},
		{"PUT", "/htmx/oneshot-acts/:id", HtmxUpdateAct},
		{"GET", "/htmx/oneshot-scenes/:id/edit", HtmxEditSceneForm},
		{"PUT", "/htmx/oneshot-scenes/:id", HtmxUpdateScene},

		// Scene Dialogs
		{"GET", "/htmx/oneshot-scenes/:id/dialogs", HtmxDialogList},
		{"GET", "/htmx/oneshot-scenes/:id/dialogs/new", HtmxNewDialogForm},
		{"POST", "/htmx/oneshot-scenes/:id/dialogs", HtmxCreateDialog},
		{"GET", "/htmx/oneshot-scene-dialogs/:id/edit", HtmxEditDialogForm},
		{"PUT", "/htmx/oneshot-scene-dialogs/:id", HtmxUpdateDialog},
		{"DELETE", "/htmx/oneshot-scene-dialogs/:id", HtmxDeleteDialog},

		// Session Pacing
		{"GET", "/htmx/session-pacing/:id", HtmxGetPacingDashboard},

		// Pregenerated Characters
		{"GET", "/htmx/pregens", HtmxListPregens},
		{"GET", "/htmx/pregens/generate", HtmxGeneratePregen},
		{"GET", "/htmx/pregens/:id", HtmxPregenCard},

		// Prep Dashboard
		{"GET", "/htmx/oneshot-adventures/:id/dashboard", HtmxGetPrepDashboard},
		{"GET", "/htmx/oneshot-adventures/:id/session-flow", HtmxGetSessionFlow},
		{"GET", "/htmx/oneshot-adventures/:id/checklist", HtmxRenderChecklist},
		{"POST", "/htmx/oneshot-adventures/:id/checklist", HtmxAddChecklistItem},
		{"POST", "/htmx/oneshot-adventures/:id/checklist/:cid/toggle", HtmxToggleChecklistItem},
		{"DELETE", "/htmx/oneshot-adventures/:id/checklist/:cid", HtmxDeleteChecklistItem},

		// DM Screen / Quick Reference
		{"GET", "/htmx/oneshot-adventures/:id/npcs", HtmxOneShotNPCs},
		{"GET", "/htmx/oneshot-adventures/:id/items", HtmxOneShotItems},
		{"GET", "/htmx/oneshot-adventures/:id/shops", HtmxOneShotShops},
		{"GET", "/htmx/oneshot-adventures/:id/monsters", HtmxOneShotMonsters},
		{"GET", "/htmx/oneshot-adventures/:id/pcs", HtmxOneShotPCs},
		{"GET", "/htmx/oneshot-adventures/:id/dm-screen", HtmxDmScreen},

		// Clue Board (HTMX)
		{"GET", "/htmx/oneshot-adventures/:id/clues", HtmxListClues},
		{"GET", "/htmx/oneshot-adventures/:id/clues/new", HtmxNewClueForm},
		{"POST", "/htmx/oneshot-adventures/:id/clues", HtmxCreateClue},
		{"GET", "/htmx/clues/:id", HtmxGetClueDetail},
		{"GET", "/htmx/clues/:id/edit", HtmxEditClueForm},
		{"PUT", "/htmx/clues/:id", HtmxUpdateClue},
		{"DELETE", "/htmx/clues/:id", HtmxDeleteClue},
		{"POST", "/htmx/clues/:id/reveal", HtmxRevealClue},
		{"POST", "/htmx/clues/:id/hide", HtmxHideClue},

		// Campaign NPCs (HTMX)
		{"GET", "/htmx/campaigns/:id/npcs-section", HtmxCampaignNPCsSection},
		{"GET", "/htmx/campaigns/:id/npcs/link-form", HtmxCampaignNPCLinkForm},
		{"GET", "/htmx/campaigns/:id/npcs/create-form", HtmxCampaignNPCCreateForm},
		{"POST", "/htmx/campaigns/:id/npcs/link", HtmxCampaignLinkNPC},
		{"POST", "/htmx/campaigns/:id/npcs/create-and-link", HtmxCampaignCreateAndLinkNPC},
		{"DELETE", "/htmx/campaigns/:id/npcs/:nid", HtmxCampaignUnlinkNPC},

		// Campaign Encounters (HTMX)
		{"GET", "/htmx/campaigns/:id/encounters-section", HtmxCampaignEncountersSection},
		{"GET", "/htmx/campaigns/:id/encounters/new", HtmxNewEncounterForm},
		{"POST", "/htmx/campaigns/:id/encounters", HtmxCreateEncounter},
		{"DELETE", "/htmx/campaigns/:id/encounters/:eid", HtmxDeleteEncounter},
		{"GET", "/htmx/campaigns/:id/encounters/:eid/monsters", HtmxCampaignEncounterMonsters},
		{"GET", "/htmx/campaigns/:id/encounters/:eid/monster-list", HtmxCampaignEncounterMonsterList},
		{"GET", "/htmx/encounters/:eid/monsters/new", HtmxAddEncounterMonsterForm},
		{"POST", "/htmx/encounters/:eid/monsters", HtmxCreateEncounterMonster},
		{"GET", "/htmx/encounters/:eid/monsters/:mid/edit", HtmxEditEncounterMonsterForm},
		{"PUT", "/htmx/encounters/:eid/monsters/:mid", HtmxUpdateEncounterMonster},
		{"DELETE", "/htmx/encounters/:eid/monsters/:mid", HtmxDeleteEncounterMonster},
		{"POST", "/htmx/encounters/:eid/import-compendium", HtmxImportCompendiumMonsterToEncounter},

		// Campaign Monster Roster (HTMX)
		{"GET", "/htmx/campaigns/:id/monster-roster", HtmxCampaignMonsterRoster},
		{"POST", "/htmx/campaigns/:id/monster-roster", HtmxAddCampaignMonster},
		{"DELETE", "/htmx/campaigns/:id/monster-roster/:rid", HtmxRemoveCampaignMonster},

		// Compendium Monsters (HTMX)
		{"GET", "/htmx/compendium-monsters", HtmxCompendiumMonsterBrowser},
		{"GET", "/htmx/compendium-monsters/search", HtmxCompendiumMonsterSearch},
		{"GET", "/htmx/compendium-monsters/import-modal", HtmxAPIImportModal},
		{"GET", "/htmx/compendium-monsters/api-search", HtmxAPIImportSearch},
		{"POST", "/htmx/compendium-monsters/import", HtmxImportAPIMonster},
		{"GET", "/htmx/compendium-monsters/:id", HtmxCompendiumMonsterDetail},
		{"GET", "/htmx/compendium-monsters/picker/:eid", HtmxCompendiumMonsterPickerForEncounter},
		{"GET", "/htmx/compendium-monsters/oneshot/:id", HtmxCompendiumMonsterPickerForOneShot},

		// Import Compendium Monster to One-Shot (HTMX)
		{"POST", "/htmx/oneshot-adventures/:id/import-compendium", HtmxImportCompendiumMonsterToOneShot},

		// Monster Library (HTMX)
		{"GET", "/htmx/monster-picker/:context/:contextId", HtmxMonsterPicker},
		{"GET", "/htmx/monster-library", HtmxListMonsterLibrary},
		{"GET", "/htmx/monster-library/new", HtmxMonsterLibraryForm},
		{"GET", "/htmx/monster-library/section", HtmxMonsterLibrarySection},
		{"POST", "/htmx/monster-library", HtmxCreateMonsterLibrary},
		{"PUT", "/htmx/monster-library/:id", HtmxUpdateMonsterLibrary},
		{"DELETE", "/htmx/monster-library/:id", HtmxDeleteMonsterLibrary},

		// Compendium Admin UI (HTMX)
		{"GET", "/htmx/admin/compendium/entries/:schemaId", HtmxCompendiumEntryTable},
		{"GET", "/htmx/admin/compendium/entry/:id", HtmxCompendiumEntryEditor},
		{"GET", "/htmx/admin/compendium/entry/detail/:id", HtmxCompendiumEntryDetail},
		{"POST", "/htmx/admin/compendium/entries/:schemaId/duplicate/:id", HtmxCompendiumDuplicateEntry},

		// Compendium Browse (HTMX)
		{"GET", "/htmx/compendium/races", HtmxCompendiumRaceBrowse},
		{"GET", "/htmx/compendium/classes", HtmxCompendiumClassBrowse},
		{"GET", "/htmx/compendium/equipment", HtmxCompendiumEquipmentBrowse},

		// Compendium Spell Browse (HTMX)
		{"GET", "/htmx/compendium/spells", HtmxCompendiumSpellBrowse},
		{"GET", "/htmx/compendium/spells/:id/detail", HtmxCompendiumSpellDetail},
		{"GET", "/htmx/compendium/spells/:id/modal", HtmxCompendiumSpellModal},

		// Compendium Linking Pickers (HTMX)
		{"GET", "/htmx/compendium/spells/picker", HtmxCompendiumSpellPicker},
		{"GET", "/htmx/compendium/equipment/picker", HtmxCompendiumEquipmentPicker},
		{"GET", "/htmx/compendium/equipment/picker-oneshot/:id", HtmxCompendiumEquipmentPickerForOneShot},
		{"POST", "/htmx/compendium/spells/link", HtmxLinkCompendiumSpell},
		{"DELETE", "/htmx/spells/:id/compendium-unlink", HtmxUnlinkCompendiumSpell},
		{"POST", "/htmx/compendium/equipment/link", HtmxLinkCompendiumEquipment},
		{"DELETE", "/htmx/inventory/:id/compendium-unlink", HtmxUnlinkCompendiumEquipment},
		{"GET", "/htmx/compendium/features/picker", HtmxCompendiumFeaturePicker},
		{"POST", "/htmx/compendium/features/link", HtmxLinkCompendiumFeature},
		{"DELETE", "/htmx/features/:id/compendium-unlink", HtmxUnlinkCompendiumFeature},

		// Compendium Card (HTMX partial)
		{"GET", "/htmx/compendium/card/:type/:id", HtmxCompendiumCard},

		// Knowledge (HTMX)
		{"GET", "/htmx/campaigns/:id/knowledge", HtmxKnowledgeList},
		{"GET", "/htmx/knowledge/:kid", HtmxKnowledgeDetail},

		// Compendium Global Search (HTMX)
		{"GET", "/htmx/compendium/search", HtmxCompendiumGlobalSearch},
	}
	for _, rt := range routes {
		r.Handle(rt.Method, rt.Path, rt.Handler)
	}
}
