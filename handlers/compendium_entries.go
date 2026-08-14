package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

// ─── Generic Compendium Entry Snapshot ───
//
// User-imported compendium content lives in the generic schema layer
// (compendium_entries.data JSON + compendium_schemas.display_name) rather than
// the legacy compendium_equipment / compendium_spells tables. This file wires
// that layer into character-sheet linking: pickers search it, link handlers
// snapshot its JSON into the character rows, and unlink keeps the data.

// compendiumEntryData is a best-effort snapshot of a generic compendium entry.
// Imported JSON keys vary by schema, so every field falls back across the most
// common key spellings.
type compendiumEntryData struct {
	Name        string
	Category    string
	Cost        string
	Description string
	Source      string // schema display name (e.g. "Homebrew Items")
	Weight      float64
	Level       int
	School      string
	CastingTime string
	Range       string
	Components  string
	Duration    string
	Classes     string
}

// loadCompendiumEntry reads a generic compendium entry and flattens its JSON
// data into a snapshot struct.
func loadCompendiumEntry(entryID int64) (*compendiumEntryData, error) {
	var raw, schemaName string
	err := db.DB.QueryRow(`SELECT e.data, COALESCE(s.display_name,'') FROM compendium_entries e LEFT JOIN compendium_schemas s ON s.id=e.schema_id WHERE e.id=?`, entryID).Scan(&raw, &schemaName)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return nil, err
	}
	d := &compendiumEntryData{Source: schemaName}
	d.Name = jsonEntryString(m, "name")
	d.Category = jsonEntryString(m, "category", "type", "item_type", "subtype")
	d.Cost = jsonEntryString(m, "cost", "price", "value")
	d.Description = jsonEntryString(m, "description", "desc", "text")
	d.School = jsonEntryString(m, "school")
	d.CastingTime = jsonEntryString(m, "casting_time", "castingTime", "cast_time", "time")
	d.Range = jsonEntryString(m, "range", "reach")
	d.Components = jsonEntryString(m, "components", "materials")
	d.Duration = jsonEntryString(m, "duration")
	d.Classes = jsonEntryString(m, "classes")
	if w := jsonEntryNumber(m, "weight"); w != nil {
		d.Weight = *w
	}
	if lvl := jsonEntryNumber(m, "level"); lvl != nil {
		d.Level = int(*lvl)
	}
	return d, nil
}

func jsonEntryString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		v, ok := m[k]
		if !ok {
			continue
		}
		switch t := v.(type) {
		case string:
			if t != "" {
				return t
			}
		case float64:
			return strconv.FormatFloat(t, 'f', -1, 64)
		case bool:
			return strconv.FormatBool(t)
		}
	}
	return ""
}

func jsonEntryNumber(m map[string]any, key string) *float64 {
	v, ok := m[key]
	if !ok {
		return nil
	}
	switch t := v.(type) {
	case float64:
		return &t
	case string:
		if f, err := strconv.ParseFloat(t, 64); err == nil {
			return &f
		}
	}
	return nil
}

// ─── Unified Picker Rows ───
//
// Picker rows embed the legacy models and add a source discriminator so
// templates can render imported entries (badge + correct link form field).

type compendiumEquipmentPickerItem struct {
	models.CompendiumEquipment
	Source     string `json:"source"` // "equipment" | "entry"
	SchemaName string `json:"schema_name"`
}

type compendiumSpellPickerItem struct {
	models.CompendiumSpell
	Source     string `json:"source"` // "spell" | "entry"
	SchemaName string `json:"schema_name"`
}

type compendiumFeaturePickerItem struct {
	ID          int64
	Name        string
	Description string
	Level       int
	SchemaName  string
}

// queryCompendiumEquipmentUnion searches legacy equipment AND generic entries,
// returning merged rows (legacy first, then entries) plus the combined count.
func queryCompendiumEquipmentUnion(q string, limit, offset int) ([]compendiumEquipmentPickerItem, int) {
	like := "%" + q + "%"
	var total int
	db.DB.QueryRow(`SELECT COUNT(*) FROM compendium_equipment WHERE name LIKE ?`, like).Scan(&total)
	var entryTotal int
	db.DB.QueryRow(`SELECT COUNT(*) FROM compendium_entries WHERE json_extract(data,'$.name') LIKE ?`, like).Scan(&entryTotal)
	total += entryTotal

	rows, err := db.DB.Query(`SELECT * FROM (
		SELECT id, name, category, cost, weight, description, source_page,
			COALESCE(system,''), COALESCE(source,''), COALESCE(item_type,''), COALESCE(item_rarity,''), COALESCE(publisher,''),
			'equipment' AS src_kind, '' AS schema_name
		FROM compendium_equipment WHERE name LIKE ?
		UNION ALL
		SELECT e.id,
			COALESCE(json_extract(e.data,'$.name'),''),
			COALESCE(json_extract(e.data,'$.category'), json_extract(e.data,'$.type'), json_extract(e.data,'$.item_type'), json_extract(e.data,'$.subtype'), ''),
			COALESCE(json_extract(e.data,'$.cost'), json_extract(e.data,'$.price'), json_extract(e.data,'$.value'), ''),
			COALESCE(CAST(json_extract(e.data,'$.weight') AS REAL), 0),
			COALESCE(json_extract(e.data,'$.description'), json_extract(e.data,'$.desc'), ''),
			'', '', '', '', '', '',
			'entry', COALESCE(s.display_name,'')
		FROM compendium_entries e LEFT JOIN compendium_schemas s ON s.id=e.schema_id
		WHERE json_extract(e.data,'$.name') LIKE ?
	) ORDER BY name LIMIT ? OFFSET ?`, like, like, limit, offset)
	if err != nil {
		return nil, total
	}
	defer rows.Close()
	var items []compendiumEquipmentPickerItem
	for rows.Next() {
		var it compendiumEquipmentPickerItem
		rows.Scan(&it.ID, &it.Name, &it.Category, &it.Cost, &it.Weight, &it.Description, &it.SourcePage,
			&it.System, &it.Source, &it.ItemType, &it.ItemRarity, &it.Publisher,
			&it.Source, &it.SchemaName)
		items = append(items, it)
	}
	return items, total
}

// queryCompendiumSpellsUnion searches legacy spells AND generic entries.
func queryCompendiumSpellsUnion(q string, limit, offset int) ([]compendiumSpellPickerItem, int) {
	like := "%" + q + "%"
	var total int
	db.DB.QueryRow(`SELECT COUNT(*) FROM compendium_spells WHERE name LIKE ?`, like).Scan(&total)
	var entryTotal int
	db.DB.QueryRow(`SELECT COUNT(*) FROM compendium_entries WHERE json_extract(data,'$.name') LIKE ?`, like).Scan(&entryTotal)
	total += entryTotal

	rows, err := db.DB.Query(`SELECT * FROM (
		SELECT id, name, level, school, casting_time, "range", components, duration, description, higher_levels, classes, source_page,
			COALESCE(system,''), COALESCE(source,''), COALESCE(publisher,''),
			'spell' AS src_kind, '' AS schema_name
		FROM compendium_spells WHERE name LIKE ?
		UNION ALL
		SELECT e.id,
			COALESCE(json_extract(e.data,'$.name'),''),
			COALESCE(CAST(json_extract(e.data,'$.level') AS INTEGER), 0),
			COALESCE(json_extract(e.data,'$.school'), ''),
			COALESCE(json_extract(e.data,'$.casting_time'), json_extract(e.data,'$.castingTime'), json_extract(e.data,'$.cast_time'), ''),
			COALESCE(json_extract(e.data,'$.range'), json_extract(e.data,'$.reach'), ''),
			COALESCE(json_extract(e.data,'$.components'), json_extract(e.data,'$.materials'), ''),
			COALESCE(json_extract(e.data,'$.duration'), ''),
			COALESCE(json_extract(e.data,'$.description'), json_extract(e.data,'$.desc'), ''),
			COALESCE(json_extract(e.data,'$.higher_levels'), json_extract(e.data,'$.higherLevels'), ''),
			COALESCE(json_extract(e.data,'$.classes'), ''),
			'', '', '', '',
			'entry', COALESCE(s.display_name,'')
		FROM compendium_entries e LEFT JOIN compendium_schemas s ON s.id=e.schema_id
		WHERE json_extract(e.data,'$.name') LIKE ?
	) ORDER BY level, name LIMIT ? OFFSET ?`, like, like, limit, offset)
	if err != nil {
		return nil, total
	}
	defer rows.Close()
	var items []compendiumSpellPickerItem
	for rows.Next() {
		var it compendiumSpellPickerItem
		rows.Scan(&it.ID, &it.Name, &it.Level, &it.School, &it.CastingTime, &it.Range, &it.Components, &it.Duration,
			&it.Description, &it.HigherLevels, &it.Classes, &it.SourcePage,
			&it.System, &it.Source, &it.Publisher,
			&it.Source, &it.SchemaName)
		items = append(items, it)
	}
	return items, total
}

// queryCompendiumEntriesForFeatures searches generic entries for the features picker.
func queryCompendiumEntriesForFeatures(q string, limit, offset int) ([]compendiumFeaturePickerItem, int) {
	like := "%" + q + "%"
	var total int
	db.DB.QueryRow(`SELECT COUNT(*) FROM compendium_entries WHERE json_extract(data,'$.name') LIKE ?`, like).Scan(&total)

	rows, err := db.DB.Query(`SELECT e.id,
		COALESCE(json_extract(e.data,'$.name'),'') AS name,
		COALESCE(json_extract(e.data,'$.description'), json_extract(e.data,'$.desc'), ''),
		COALESCE(CAST(json_extract(e.data,'$.level') AS INTEGER), 0),
		COALESCE(s.display_name,'')
		FROM compendium_entries e LEFT JOIN compendium_schemas s ON s.id=e.schema_id
		WHERE json_extract(e.data,'$.name') LIKE ? ORDER BY name LIMIT ? OFFSET ?`, like, limit, offset)
	if err != nil {
		return nil, total
	}
	defer rows.Close()
	var items []compendiumFeaturePickerItem
	for rows.Next() {
		var it compendiumFeaturePickerItem
		rows.Scan(&it.ID, &it.Name, &it.Description, &it.Level, &it.SchemaName)
		items = append(items, it)
	}
	return items, total
}

// ─── Entry Linking (Character Sheet) ───

// entryItemLinkInsert snapshots a generic compendium entry into the character's inventory.
func entryItemLinkInsert(charID, entryID int64, quantity int) (int, string) {
	snap, err := loadCompendiumEntry(entryID)
	if err != nil {
		return http.StatusNotFound, "compendium entry not found"
	}
	_, err = db.DB.Exec(`INSERT INTO inventory(character_id, name, quantity, weight, category, description, notes, compendium_entry_id)
		VALUES(?,?,?,?,?,?,'',?)`,
		charID, snap.Name, quantity, snap.Weight, snap.Category, snap.Description, entryID)
	if err != nil {
		return http.StatusInternalServerError, err.Error()
	}
	return 0, ""
}

// entrySpellLinkInsert snapshots a generic compendium entry into the character's spellbook.
func entrySpellLinkInsert(charID, entryID int64) (int, string) {
	snap, err := loadCompendiumEntry(entryID)
	if err != nil {
		return http.StatusNotFound, "compendium entry not found"
	}
	_, err = db.DB.Exec(`INSERT INTO spells(character_id, name, level, school, casting_time, "range", components, duration, description, source, notes, compendium_entry_id)
		VALUES(?,?,?,?,?,?,?,?,?,?,'',?)`,
		charID, snap.Name, snap.Level, snap.School, snap.CastingTime, snap.Range, snap.Components, snap.Duration, snap.Description, snap.Source, entryID)
	if err != nil {
		return http.StatusInternalServerError, err.Error()
	}
	return 0, ""
}

// featureLinkInsert snapshots a generic compendium entry into the character's features.
func featureLinkInsert(charID, entryID int64, levelGained int) (int, string) {
	snap, err := loadCompendiumEntry(entryID)
	if err != nil {
		return http.StatusNotFound, "compendium entry not found"
	}
	_, err = db.DB.Exec(`INSERT INTO character_features(character_id, name, description, source, level_gained, compendium_entry_id)
		VALUES(?,?,?,?,?,?)`,
		charID, snap.Name, snap.Description, snap.Source, levelGained, entryID)
	if err != nil {
		return http.StatusInternalServerError, err.Error()
	}
	return 0, ""
}

// parsePriceGP converts a cost string ("50 gp", "2 sp", "25") into gold pieces.
func parsePriceGP(cost string) float64 {
	if cost == "" {
		return 0
	}
	parts := strings.Fields(cost)
	if len(parts) == 0 {
		return 0
	}
	val, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0
	}
	if len(parts) > 1 {
		switch strings.ToLower(parts[1]) {
		case "cp":
			return val / 100
		case "sp":
			return val / 10
		case "ep":
			return val * 2
		case "pp":
			return val * 10
		}
	}
	return val
}

// HtmxLinkCompendiumFeature links a generic compendium entry as a character feature.
// POST /htmx/compendium/features/link?character_id=:cid (form: compendium_entry_id, level_gained)
func HtmxLinkCompendiumFeature(c *gin.Context) {
	charID := c.Query("character_id")
	cid, err := strconv.ParseInt(charID, 10, 64)
	if err != nil || charID == "" {
		c.String(http.StatusBadRequest, "invalid character id")
		return
	}
	if !canEditCharacterID(c, cid) {
		c.String(http.StatusForbidden, "access denied")
		return
	}
	entryID, err := strconv.ParseInt(c.PostForm("compendium_entry_id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "compendium_entry_id required")
		return
	}
	level := getIntParam(c, "level_gained", 1)
	if st, msg := featureLinkInsert(cid, entryID, level); msg != "" {
		c.String(st, msg)
		return
	}
	renderHtmxFeaturesList(c, charID)
}

// HtmxUnlinkCompendiumFeature unlinks a feature from a generic entry, preserving the feature data.
// DELETE /htmx/features/:id/compendium-unlink?character_id=:cid
func HtmxUnlinkCompendiumFeature(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid feature id")
		return
	}
	if !canEditResourceID(c, "character_features", id) {
		c.String(http.StatusForbidden, "access denied")
		return
	}
	charID := c.Query("character_id")
	if charID == "" {
		c.String(http.StatusBadRequest, "character_id required")
		return
	}
	_, err = db.DB.Exec("UPDATE character_features SET compendium_entry_id = NULL WHERE id=?", id)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	renderHtmxFeaturesList(c, charID)
}
