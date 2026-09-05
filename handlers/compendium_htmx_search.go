package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
	"villum/db"
	"villum/middleware"
	"villum/models"
)

func HtmxCompendiumCard(c *gin.Context) {
	entityType := c.Param("type")
	entityIDStr := c.Param("id")
	entityID, err := strconv.ParseInt(entityIDStr, 10, 64)
	if err != nil {
		renderTemplate(c, "compendium_card_not_found", nil)
		return
	}

	switch entityType {
	case "spell":
		var s models.CompendiumSpell
		err := db.DB.QueryRow(`SELECT id, name, level, school, casting_time, "range", components, duration,
			description, higher_levels, classes, source_page,
			COALESCE(system,''), COALESCE(source,''), COALESCE(publisher,'')
			FROM compendium_spells WHERE id=?`, entityID).Scan(
			&s.ID, &s.Name, &s.Level, &s.School, &s.CastingTime, &s.Range, &s.Components, &s.Duration,
			&s.Description, &s.HigherLevels, &s.Classes, &s.SourcePage,
			&s.System, &s.Source, &s.Publisher)
		if err != nil {
			renderTemplate(c, "compendium_card_not_found", nil)
			return
		}
		renderTemplate(c, "compendium_spell_card", s)

	case "equipment":
		var e models.CompendiumEquipment
		err := db.DB.QueryRow(`SELECT id, name, category, cost, weight, description, source_page,
			COALESCE(system,''), COALESCE(source,''), COALESCE(item_type,''), COALESCE(item_rarity,''), COALESCE(publisher,'')
			FROM compendium_equipment WHERE id=?`, entityID).Scan(
			&e.ID, &e.Name, &e.Category, &e.Cost, &e.Weight, &e.Description, &e.SourcePage,
			&e.System, &e.Source, &e.ItemType, &e.ItemRarity, &e.Publisher)
		if err != nil {
			renderTemplate(c, "compendium_card_not_found", nil)
			return
		}
		renderTemplate(c, "compendium_equipment_card", e)

	case "monster":
		var m models.CompendiumMonster
		var isFull int
		err := db.DB.QueryRow(`SELECT id, name, type, size, ac, hp, str, dex, con, int_, wis, cha, cr,
			source, is_full, saves, skills, damage_vulnerabilities, damage_resistances, damage_immunities,
			condition_immunities, senses, languages, special_abilities, actions, legendary_actions, description,
			COALESCE(alignment,''), COALESCE(expansion,''), COALESCE(publisher,'')
			FROM compendium_monsters WHERE id=?`, entityID).Scan(
			&m.ID, &m.Name, &m.Type, &m.Size, &m.AC, &m.HP, &m.Str, &m.Dex, &m.Con, &m.Int, &m.Wis, &m.Cha, &m.CR,
			&m.Source, &isFull, &m.Saves, &m.Skills, &m.DamageVulnerabilities, &m.DamageResistances, &m.DamageImmunities,
			&m.ConditionImmunities, &m.Senses, &m.Languages, &m.SpecialAbilities, &m.Actions, &m.LegendaryActions, &m.Description,
			&m.Alignment, &m.Expansion, &m.Publisher)
		if err != nil {
			renderTemplate(c, "compendium_card_not_found", nil)
			return
		}
		m.IsFull = isFull == 1
		renderTemplate(c, "compendium_monster_card", m)

	default:
		renderTemplate(c, "compendium_card_not_found", nil)
	}
}

// ─── Compendium Global Search (HTMX) ───

type htmxCompendiumGlobalSearchItem struct {
	Type           string
	ID             int64
	Name           string
	Subtype        string
	CR             string
	Level          int
	HitDie         int
	PrimaryAbility string
}

type htmxCompendiumGlobalSearchData struct {
	Query       string
	Spells      []htmxCompendiumGlobalSearchItem
	Equipment   []htmxCompendiumGlobalSearchItem
	Monsters    []htmxCompendiumGlobalSearchItem
	Races       []htmxCompendiumGlobalSearchItem
	Classes     []htmxCompendiumGlobalSearchItem
	Feats       []htmxCompendiumGlobalSearchItem
	Backgrounds []htmxCompendiumGlobalSearchItem
}

func HtmxCompendiumGlobalSearch(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	data := htmxCompendiumGlobalSearchData{Query: q}

	if q == "" {
		renderTemplate(c, "compendium_global_search_results", data)
		return
	}

	like := "%" + q + "%"

	// Spells
	rows, qErr := db.DB.Query("SELECT id, name, level, school FROM compendium_spells WHERE name LIKE ? ORDER BY level, name LIMIT 10", like)
	if qErr != nil {
		middleware.LogWarn("compendium", "query failed", "type", "spell", "error", qErr)
	} else {
		for rows.Next() {
			var item htmxCompendiumGlobalSearchItem
			item.Type = "spell"
			if err := rows.Scan(&item.ID, &item.Name, &item.Level, &item.Subtype); err != nil {
				middleware.LogWarn("compendium", "scan failed, skipping row", "type", "spell", "error", err)
				continue
			}
			data.Spells = append(data.Spells, item)
		}
		rows.Close()
	}

	// Equipment
	rows, qErr = db.DB.Query("SELECT id, name, category FROM compendium_equipment WHERE name LIKE ? ORDER BY name LIMIT 10", like)
	if qErr != nil {
		middleware.LogWarn("compendium", "query failed", "type", "equipment", "error", qErr)
	} else {
		for rows.Next() {
			var item htmxCompendiumGlobalSearchItem
			item.Type = "equipment"
			if err := rows.Scan(&item.ID, &item.Name, &item.Subtype); err != nil {
				middleware.LogWarn("compendium", "scan failed, skipping row", "type", "equipment", "error", err)
				continue
			}
			data.Equipment = append(data.Equipment, item)
		}
		rows.Close()
	}

	// Monsters
	rows, qErr = db.DB.Query("SELECT id, name, cr, type FROM compendium_monsters WHERE name LIKE ? ORDER BY name LIMIT 10", like)
	if qErr != nil {
		middleware.LogWarn("compendium", "query failed", "type", "monster", "error", qErr)
	} else {
		for rows.Next() {
			var item htmxCompendiumGlobalSearchItem
			item.Type = "monster"
			if err := rows.Scan(&item.ID, &item.Name, &item.CR, &item.Subtype); err != nil {
				middleware.LogWarn("compendium", "scan failed, skipping row", "type", "monster", "error", err)
				continue
			}
			data.Monsters = append(data.Monsters, item)
		}
		rows.Close()
	}

	// Races
	rows, qErr = db.DB.Query("SELECT id, name FROM compendium_races WHERE name LIKE ? ORDER BY name LIMIT 10", like)
	if qErr != nil {
		middleware.LogWarn("compendium", "query failed", "type", "race", "error", qErr)
	} else {
		for rows.Next() {
			var item htmxCompendiumGlobalSearchItem
			item.Type = "race"
			if err := rows.Scan(&item.ID, &item.Name); err != nil {
				middleware.LogWarn("compendium", "scan failed, skipping row", "type", "race", "error", err)
				continue
			}
			data.Races = append(data.Races, item)
		}
		rows.Close()
	}

	// Classes
	rows, qErr = db.DB.Query("SELECT id, name, hit_die, primary_ability FROM compendium_classes WHERE name LIKE ? ORDER BY name LIMIT 10", like)
	if qErr != nil {
		middleware.LogWarn("compendium", "query failed", "type", "class", "error", qErr)
	} else {
		for rows.Next() {
			var item htmxCompendiumGlobalSearchItem
			item.Type = "class"
			if err := rows.Scan(&item.ID, &item.Name, &item.HitDie, &item.PrimaryAbility); err != nil {
				middleware.LogWarn("compendium", "scan failed, skipping row", "type", "class", "error", err)
				continue
			}
			data.Classes = append(data.Classes, item)
		}
		rows.Close()
	}

	// Feats
	rows, qErr = db.DB.Query("SELECT id, name FROM compendium_feats WHERE name LIKE ? ORDER BY name LIMIT 10", like)
	if qErr != nil {
		middleware.LogWarn("compendium", "query failed", "type", "feat", "error", qErr)
	} else {
		for rows.Next() {
			var item htmxCompendiumGlobalSearchItem
			item.Type = "feat"
			if err := rows.Scan(&item.ID, &item.Name); err != nil {
				middleware.LogWarn("compendium", "scan failed, skipping row", "type", "feat", "error", err)
				continue
			}
			data.Feats = append(data.Feats, item)
		}
		rows.Close()
	}

	// Backgrounds
	rows, qErr = db.DB.Query("SELECT id, name FROM compendium_backgrounds WHERE name LIKE ? ORDER BY name LIMIT 10", like)
	if qErr != nil {
		middleware.LogWarn("compendium", "query failed", "type", "background", "error", qErr)
	} else {
		for rows.Next() {
			var item htmxCompendiumGlobalSearchItem
			item.Type = "background"
			if err := rows.Scan(&item.ID, &item.Name); err != nil {
				middleware.LogWarn("compendium", "scan failed, skipping row", "type", "background", "error", err)
				continue
			}
			data.Backgrounds = append(data.Backgrounds, item)
		}
		rows.Close()
	}

	renderTemplate(c, "compendium_global_search_results", data)
}

// ─── Unified Monster Picker (monster-management change) ───

type htmxMonsterPickerData struct {
	Context     string
	ContextID   int64
	Tab         string
	Query       string
	EncounterID int64
	AdventureID int64
	CampaignID  int64
	ShowRoster  bool
}

// HtmxMonsterPicker renders the shared monster picker modal body for
// oneshot / encounter / campaign contexts with Compendium, My Library,
// and (for campaigns) Campaign Roster tabs.
func HtmxMonsterPicker(c *gin.Context) {
	context := c.Param("context")
	contextID, err := strconv.ParseInt(c.Param("contextId"), 10, 64)
	if err != nil || contextID <= 0 {
		c.String(http.StatusBadRequest, "invalid context id")
		return
	}
	data := htmxMonsterPickerData{
		Context:   context,
		ContextID: contextID,
		Tab:       c.DefaultQuery("tab", "compendium"),
		Query:     c.Query("q"),
	}
	switch context {
	case "oneshot":
		data.AdventureID = contextID
	case "encounter":
		data.EncounterID = contextID
	case "campaign":
		data.CampaignID = contextID
		data.ShowRoster = true
	default:
		c.String(http.StatusBadRequest, "invalid context")
		return
	}
	renderTemplate(c, "monster_picker", data)
}
