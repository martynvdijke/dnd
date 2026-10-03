package search

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"
)

// SearchCompendium performs compendium recall.
func SearchCompendium(ctx context.Context, db *sql.DB, p CompendiumParams) ([]CompendiumResult, error) {
	q := strings.TrimSpace(p.Query)
	if q == "" {
		return nil, nil
	}
	var results []CompendiumResult

	if p.TypeFilter == "" || p.TypeFilter == "spell" {
		extra := ""
		args := []any{}
		if p.Class != "" {
			extra += " AND classes LIKE ?"
			args = append(args, "%\""+p.Class+"\"%")
		}
		if p.Level != "" {
			extra += " AND level=?"
			args = append(args, p.Level)
		}
		if p.School != "" {
			extra += " AND school=?"
			args = append(args, p.School)
		}
		rows, err := db.QueryContext(ctx, "SELECT id, name, level, school FROM compendium_spells WHERE name LIKE ?"+extra+" ORDER BY level, name LIMIT 10", append([]any{"%" + q + "%"}, args...)...)
		if err == nil {
			for rows.Next() {
				var r CompendiumResult
				r.Type = "spell"
				_ = rows.Scan(&r.ID, &r.Name, &r.Level, &r.Subtype)
				results = append(results, r)
			}
			rows.Close()
		}
	}
	if p.TypeFilter == "" || p.TypeFilter == "equipment" {
		extra := ""
		args := []any{}
		if p.Category != "" {
			extra += " AND category=?"
			args = append(args, p.Category)
		}
		rows, err := db.QueryContext(ctx, "SELECT id, name, category FROM compendium_equipment WHERE name LIKE ?"+extra+" ORDER BY name LIMIT 10", append([]any{"%" + q + "%"}, args...)...)
		if err == nil {
			for rows.Next() {
				var r CompendiumResult
				r.Type = "equipment"
				_ = rows.Scan(&r.ID, &r.Name, &r.Subtype)
				results = append(results, r)
			}
			rows.Close()
		}
	}
	if p.TypeFilter == "" || p.TypeFilter == "monster" {
		extra := ""
		args := []any{}
		if p.CR != "" {
			extra += " AND cr=?"
			args = append(args, p.CR)
		}
		if p.MonsterType != "" {
			extra += " AND type LIKE ?"
			args = append(args, "%"+p.MonsterType+"%")
		}
		rows, err := db.QueryContext(ctx, "SELECT id, name, cr, type FROM compendium_monsters WHERE name LIKE ?"+extra+" ORDER BY name LIMIT 10", append([]any{"%" + q + "%"}, args...)...)
		if err == nil {
			for rows.Next() {
				var r CompendiumResult
				r.Type = "monster"
				_ = rows.Scan(&r.ID, &r.Name, &r.CR, &r.Subtype)
				results = append(results, r)
			}
			rows.Close()
		}
	}
	if p.TypeFilter == "" || p.TypeFilter == "race" {
		rows, err := db.QueryContext(ctx, "SELECT id, name FROM compendium_races WHERE name LIKE ? ORDER BY name LIMIT 10", "%"+q+"%")
		if err == nil {
			for rows.Next() {
				var r CompendiumResult
				_ = rows.Scan(&r.ID, &r.Name)
				r.Type = "race"
				results = append(results, r)
			}
			rows.Close()
		}
	}
	if p.TypeFilter == "" || p.TypeFilter == "feat" {
		rows, err := db.QueryContext(ctx, "SELECT id, name FROM compendium_feats WHERE name LIKE ? ORDER BY name LIMIT 10", "%"+q+"%")
		if err == nil {
			for rows.Next() {
				var r CompendiumResult
				_ = rows.Scan(&r.ID, &r.Name)
				r.Type = "feat"
				results = append(results, r)
			}
			rows.Close()
		}
	}
	if p.TypeFilter == "" || p.TypeFilter == "background" {
		rows, err := db.QueryContext(ctx, "SELECT id, name FROM compendium_backgrounds WHERE name LIKE ? ORDER BY name LIMIT 10", "%"+q+"%")
		if err == nil {
			for rows.Next() {
				var r CompendiumResult
				_ = rows.Scan(&r.ID, &r.Name)
				r.Type = "background"
				results = append(results, r)
			}
			rows.Close()
		}
	}
	if p.TypeFilter == "" || p.TypeFilter == "class" {
		rows, err := db.QueryContext(ctx, "SELECT id, name, hit_die, primary_ability FROM compendium_classes WHERE name LIKE ? ORDER BY name LIMIT 10", "%"+q+"%")
		if err == nil {
			for rows.Next() {
				var r CompendiumResult
				_ = rows.Scan(&r.ID, &r.Name, &r.HitDie, &r.PrimaryAbility)
				r.Type = "class"
				results = append(results, r)
			}
			rows.Close()
		}
	}
	// TODO(compendium-and-ai-search 1.3): when primary recall is thin and len(q) >= 4, add bounded LIKE broadening then DefaultRerank. Ceiling: compendium scale (~thousands of rows). Upgrade path: trigram FTS index behind the reranker seam.
	// FTS over generic entries
	ftsWhere := "compendium_entries_fts MATCH ?"
	ftsArgs := []any{BuildFTS5Query(q)}
	if p.TypeFilter != "" {
		ftsWhere += " AND cs.type_name = ?"
		ftsArgs = append(ftsArgs, p.TypeFilter)
	}
	rows, err := db.QueryContext(ctx, "SELECT e.id, e.schema_id, e.data, cs.type_name FROM compendium_entries e JOIN compendium_entries_fts f ON e.id = f.rowid JOIN compendium_schemas cs ON e.schema_id = cs.id WHERE "+ftsWhere+" ORDER BY rank LIMIT 25", ftsArgs...)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var r CompendiumResult
			var schemaID int64
			var dataJSON string
			if err := rows.Scan(&r.ID, &schemaID, &dataJSON, &r.Type); err != nil {
				continue
			}
			var data map[string]any
			if err := json.Unmarshal([]byte(dataJSON), &data); err == nil {
				if n, ok := data["name"].(string); ok {
					r.Name = n
				}
			}
			results = append(results, r)
		}
	}
	if p.Reranker != nil {
		results = p.Reranker(q, results)
	}
	return results, nil
}

// LoadCompendiumEntry reads a generic compendium entry.
func LoadCompendiumEntry(ctx context.Context, db *sql.DB, id int64) (*EntryData, error) {
	var raw, schemaName string
	err := db.QueryRowContext(ctx, `SELECT e.data, COALESCE(s.display_name,'') FROM compendium_entries e LEFT JOIN compendium_schemas s ON s.id=e.schema_id WHERE e.id=?`, id).Scan(&raw, &schemaName)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return nil, err
	}
	d := &EntryData{Source: schemaName}
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

// QueryCompendiumEquipmentUnion etc. - exported helpers for pickers

// EquipmentPickerItem mirrors compendiumEquipmentPickerItem
type EquipmentPickerItem struct {
	ID         int64
	Name       string
	Category   string
	Cost       string
	Weight     float64
	Description string
	SourcePage string
	System     string
	Source     string
	ItemType   string
	ItemRarity string
	Publisher  string
	SrcKind    string
	SchemaName string
}

func QueryCompendiumEquipmentUnion(ctx context.Context, db *sql.DB, q string, limit, offset int) ([]EquipmentPickerItem, int) {
	like := "%" + q + "%"
	var total int
	_ = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM compendium_equipment WHERE name LIKE ?`, like).Scan(&total)
	var entryTotal int
	_ = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM compendium_entries WHERE json_extract(data,'$.name') LIKE ?`, like).Scan(&entryTotal)
	total += entryTotal
	rows, err := db.QueryContext(ctx, `SELECT * FROM (
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
	var items []EquipmentPickerItem
	for rows.Next() {
		var it EquipmentPickerItem
		rows.Scan(&it.ID, &it.Name, &it.Category, &it.Cost, &it.Weight, &it.Description, &it.SourcePage,
			&it.System, &it.Source, &it.ItemType, &it.ItemRarity, &it.Publisher,
			&it.SrcKind, &it.SchemaName)
		items = append(items, it)
	}
	return items, total
}

type SpellPickerItem struct {
	ID          int64
	Name        string
	Level       int
	School      string
	CastingTime string
	Range       string
	Components  string
	Duration    string
	Description string
	HigherLevels string
	Classes     string
	SourcePage  string
	System      string
	Source      string
	Publisher   string
	SrcKind     string
	SchemaName  string
}

func QueryCompendiumSpellsUnion(ctx context.Context, db *sql.DB, q, class, level string, limit, offset int) ([]SpellPickerItem, int) {
	legacyWhere := []string{"name LIKE ?"}
	legacyArgs := []any{"%" + q + "%"}
	entryWhere := []string{"json_extract(data,'$.name') LIKE ?"}
	entryArgs := []any{"%" + q + "%"}
	if class != "" {
		legacyWhere = append(legacyWhere, "classes LIKE ?")
		legacyArgs = append(legacyArgs, "%"+class+"%")
		entryWhere = append(entryWhere, "json_extract(data,'$.classes') LIKE ?")
		entryArgs = append(entryArgs, "%"+class+"%")
	}
	if from, to, ok := spellPickerLevelRange(level); ok {
		legacyWhere = append(legacyWhere, "level >= ? AND level <= ?")
		legacyArgs = append(legacyArgs, from, to)
		entryWhere = append(entryWhere, "CAST(json_extract(data,'$.level') AS INTEGER) >= ? AND CAST(json_extract(data,'$.level') AS INTEGER) <= ?")
		entryArgs = append(entryArgs, from, to)
	}
	legacyFilter := strings.Join(legacyWhere, " AND ")
	entryFilter := strings.Join(entryWhere, " AND ")
	entrySpellFilter := `(
		LOWER(COALESCE(s.display_name,'')) LIKE '%spell%'
		OR json_extract(data,'$.school') IS NOT NULL
		OR json_extract(data,'$.casting_time') IS NOT NULL
		OR json_extract(data,'$.castingTime') IS NOT NULL
		OR json_extract(data,'$.classes') IS NOT NULL
	)`
	var total, entryTotal int
	_ = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM compendium_spells WHERE "+legacyFilter, legacyArgs...).Scan(&total)
	_ = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM compendium_entries e LEFT JOIN compendium_schemas s ON s.id=e.schema_id WHERE "+strings.ReplaceAll(entryFilter, "data", "e.data")+" AND "+strings.ReplaceAll(entrySpellFilter, "data", "e.data"), entryArgs...).Scan(&entryTotal)
	total += entryTotal
	query := `SELECT * FROM (
		SELECT id, name, level, school, casting_time, "range", components, duration, description, higher_levels, classes, source_page,
			COALESCE(system,''), COALESCE(source,''), COALESCE(publisher,''),
			'spell' AS src_kind, '' AS schema_name
		FROM compendium_spells WHERE ` + legacyFilter + `
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
		WHERE (` + strings.ReplaceAll(entryFilter, "data", "e.data") + `) AND (
			LOWER(COALESCE(s.display_name,'')) LIKE '%spell%'
			OR json_extract(e.data,'$.school') IS NOT NULL
			OR json_extract(e.data,'$.casting_time') IS NOT NULL
			OR json_extract(e.data,'$.castingTime') IS NOT NULL
			OR json_extract(e.data,'$.classes') IS NOT NULL
		)
	) ORDER BY level, name LIMIT ? OFFSET ?`
	args := append(append(legacyArgs, entryArgs...), limit, offset)
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, total
	}
	defer rows.Close()
	var items []SpellPickerItem
	for rows.Next() {
		var it SpellPickerItem
		rows.Scan(&it.ID, &it.Name, &it.Level, &it.School, &it.CastingTime, &it.Range, &it.Components, &it.Duration,
			&it.Description, &it.HigherLevels, &it.Classes, &it.SourcePage,
			&it.System, &it.Source, &it.Publisher,
			&it.SrcKind, &it.SchemaName)
		items = append(items, it)
	}
	return items, total
}

func spellPickerLevelRange(level string) (int, int, bool) {
	switch level {
	case "0":
		return 0, 0, true
	case "1-3":
		return 1, 3, true
	case "4-6":
		return 4, 6, true
	case "7-9":
		return 7, 9, true
	default:
		return 0, 0, false
	}
}

type FeaturePickerItem struct {
	ID          int64
	Name        string
	Description string
	Level       int
	SchemaName  string
}

func QueryCompendiumEntriesForFeatures(ctx context.Context, db *sql.DB, q string, limit, offset int) ([]FeaturePickerItem, int) {
	like := "%" + q + "%"
	var total int
	_ = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM compendium_entries WHERE json_extract(data,'$.name') LIKE ?`, like).Scan(&total)
	rows, err := db.QueryContext(ctx, `SELECT e.id,
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
	var items []FeaturePickerItem
	for rows.Next() {
		var it FeaturePickerItem
		rows.Scan(&it.ID, &it.Name, &it.Description, &it.Level, &it.SchemaName)
		items = append(items, it)
	}
	return items, total
}
