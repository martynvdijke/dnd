package db

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type jsonSeedEntry struct {
	System string `json:"system"`
	Source string `json:"source"`
}

// SeedJSONCategory loads a single compendium category from a JSON file in data/ directory.
// Returns true if the JSON file was found and loaded.
// The JSON files should be named: races.json, classes.json, spells.json,
// feats.json, backgrounds.json, equipment.json
func SeedJSONCategory(dataDir, category string) bool {
	info, err := os.Stat(dataDir)
	if err != nil || !info.IsDir() {
		return false
	}

	path := filepath.Join(dataDir, category+".json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("Warning: failed to read %s: %v", path, err)
		return false
	}

	var entries []map[string]any
	if err := json.Unmarshal(data, &entries); err != nil {
		var entry map[string]any
		if err2 := json.Unmarshal(data, &entry); err2 != nil {
			log.Printf("Warning: invalid JSON in %s: %v", path, err)
			return false
		}
		entries = []map[string]any{entry}
	}

	if len(entries) == 0 {
		return false
	}

	// Check if table already has data
	var count int
	DB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM compendium_%s", category)).Scan(&count)
	if count > 0 {
		// If force flag exists, clear and reload
		if hasForceFlag(dataDir) {
			DB.Exec(fmt.Sprintf("DELETE FROM compendium_%s", category))
		} else {
			log.Printf("Skipping %s: table already has data", path)
			return true
		}
	}

	seedFn, ok := jsonSeeders[category]
	if !ok {
		log.Printf("Warning: unknown compendium category: %s", category)
		return false
	}

	if err := seedFn(entries); err != nil {
		log.Printf("Warning: failed to seed %s: %v", path, err)
		return false
	}
	log.Printf("Seeded %d entries from %s", len(entries), path)
	return true
}

var jsonSeeders = map[string]func([]map[string]any) error{
	"races":       seedJSONRaces,
	"classes":     seedJSONClasses,
	"spells":      seedJSONSpells,
	"feats":       seedJSONFeats,
	"backgrounds": seedJSONBackgrounds,
	"equipment":   seedJSONEquipment,
	"monsters":    seedJSONMonsters,
}

func seedJSONMonsters(entries []map[string]any) error {
	stmt, err := DB.Prepare(`INSERT INTO compendium_monsters(name,type,size,ac,hp,str,dex,con,int_,wis,cha,cr,source,is_full,
		saves,skills,damage_vulnerabilities,damage_resistances,damage_immunities,condition_immunities,senses,languages,
		special_abilities,actions,legendary_actions,description,alignment,expansion,publisher) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,
		?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, e := range entries {
		_, err := stmt.Exec(
			getStr(e, "name"), getStr(e, "type"), getStr(e, "size"), getInt(e, "ac", 10), getInt(e, "hp", 1),
			getInt(e, "str", 10), getInt(e, "dex", 10), getInt(e, "con", 10), getInt(e, "int", 10), getInt(e, "wis", 10), getInt(e, "cha", 10),
			getStrDef(e, "cr", "0"), getStrDef(e, "source", "SRD"), getInt(e, "is_full", 0),
			getStr(e, "saves"), getStr(e, "skills"), getStr(e, "damage_vulnerabilities"), getStr(e, "damage_resistances"),
			getStr(e, "damage_immunities"), getStr(e, "condition_immunities"), getStr(e, "senses"), getStr(e, "languages"),
			getStr(e, "special_abilities"), getStr(e, "actions"), getStr(e, "legendary_actions"), getStr(e, "description"),
			getStr(e, "alignment"), getStr(e, "expansion"), getStr(e, "publisher"),
		)
		if err != nil {
			return fmt.Errorf("insert monster %v: %w", e["name"], err)
		}
	}
	return nil
}

func hasForceFlag(dataDir string) bool {
	_, err := os.Stat(filepath.Join(dataDir, ".force"))
	return err == nil
}

func seedJSONRaces(entries []map[string]any) error {
	stmt, err := DB.Prepare(`INSERT INTO compendium_races(name,description,speed,size,ability_bonuses,traits,languages,source_page,system,source,category,expansion,publisher) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, e := range entries {
		_, err := stmt.Exec(
			getStr(e, "name"), getStr(e, "description"), getInt(e, "speed", 30),
			getStr(e, "size"), getStr(e, "ability_bonuses"), getStr(e, "traits"),
			getStr(e, "languages"), getStr(e, "source_page"),
			getStrDef(e, "system", "dnd5e"), getStrDef(e, "source", "custom"),
			getStr(e, "category"), getStr(e, "expansion"), getStr(e, "publisher"),
		)
		if err != nil {
			return fmt.Errorf("insert race %v: %w", e["name"], err)
		}
	}
	return nil
}

func seedJSONClasses(entries []map[string]any) error {
	stmt, err := DB.Prepare(`INSERT INTO compendium_classes(name,description,hit_die,primary_ability,saving_throws,proficiencies,spellcasting_ability,source_page,system,source,category,expansion,publisher) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, e := range entries {
		_, err := stmt.Exec(
			getStr(e, "name"), getStr(e, "description"), getInt(e, "hit_die", 8),
			getStr(e, "primary_ability"), getStr(e, "saving_throws"), getStr(e, "proficiencies"),
			getStr(e, "spellcasting_ability"), getStr(e, "source_page"),
			getStrDef(e, "system", "dnd5e"), getStrDef(e, "source", "custom"),
			getStr(e, "category"), getStr(e, "expansion"), getStr(e, "publisher"),
		)
		if err != nil {
			return fmt.Errorf("insert class %v: %w", e["name"], err)
		}
	}
	return nil
}

func seedJSONSpells(entries []map[string]any) error {
	stmt, err := DB.Prepare(`INSERT INTO compendium_spells(name,level,school,casting_time,range,components,duration,description,higher_levels,classes,source_page,system,source,publisher) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, e := range entries {
		_, err := stmt.Exec(
			getStr(e, "name"), getInt(e, "level", 0), getStr(e, "school"),
			getStr(e, "casting_time"), getStr(e, "range"), getStr(e, "components"),
			getStr(e, "duration"), getStr(e, "description"), getStr(e, "higher_levels"),
			getStr(e, "classes"), getStr(e, "source_page"),
			getStrDef(e, "system", "dnd5e"), getStrDef(e, "source", "custom"),
			getStr(e, "publisher"),
		)
		if err != nil {
			return fmt.Errorf("insert spell %v: %w", e["name"], err)
		}
	}
	return nil
}

func seedJSONFeats(entries []map[string]any) error {
	stmt, err := DB.Prepare(`INSERT INTO compendium_feats(name,description,prerequisites,source_page,system,source) VALUES(?,?,?,?,?,?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, e := range entries {
		_, err := stmt.Exec(
			getStr(e, "name"), getStr(e, "description"), getStr(e, "prerequisites"),
			getStr(e, "source_page"),
			getStrDef(e, "system", "dnd5e"), getStrDef(e, "source", "custom"),
		)
		if err != nil {
			return fmt.Errorf("insert feat %v: %w", e["name"], err)
		}
	}
	return nil
}

func seedJSONBackgrounds(entries []map[string]any) error {
	stmt, err := DB.Prepare(`INSERT INTO compendium_backgrounds(name,description,feature_name,feature_description,proficiencies,source_page,system,source,category,data_list,data_bonds,data_flaws,data_ideals,data_equipment,data_starting_gold,data_personality_traits,publisher) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, e := range entries {
		_, err := stmt.Exec(
			getStr(e, "name"), getStr(e, "description"), getStr(e, "feature_name"),
			getStr(e, "feature_description"), getStr(e, "proficiencies"),
			getStr(e, "source_page"),
			getStrDef(e, "system", "dnd5e"), getStrDef(e, "source", "custom"),
			getStr(e, "category"), getInt(e, "data_list", 0), getStr(e, "data_bonds"),
			getStr(e, "data_flaws"), getStr(e, "data_ideals"), getStr(e, "data_equipment"),
			getInt(e, "data_starting_gold", 0), getStr(e, "data_personality_traits"),
			getStr(e, "publisher"),
		)
		if err != nil {
			return fmt.Errorf("insert background %v: %w", e["name"], err)
		}
	}
	return nil
}

func seedJSONEquipment(entries []map[string]any) error {
	stmt, err := DB.Prepare(`INSERT INTO compendium_equipment(name,category,cost,weight,description,source_page,system,source,item_type,item_rarity,publisher) VALUES(?,?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, e := range entries {
		_, err := stmt.Exec(
			getStr(e, "name"), getStr(e, "category"), getStr(e, "cost"),
			getFloat(e, "weight", 0), getStr(e, "description"), getStr(e, "source_page"),
			getStrDef(e, "system", "dnd5e"), getStrDef(e, "source", "custom"),
			getStr(e, "item_type"), getStr(e, "item_rarity"), getStr(e, "publisher"),
		)
		if err != nil {
			return fmt.Errorf("insert equipment %v: %w", e["name"], err)
		}
	}
	return nil
}

// Helper functions for safely extracting values from JSON maps
func getStr(m map[string]any, key string) string {
	if v, ok := m[key]; ok && v != nil {
		if s, ok := v.(string); ok {
			return s
		}
		// Handle numbers as strings
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func getStrDef(m map[string]any, key, def string) string {
	if v := getStr(m, key); v != "" {
		return v
	}
	return def
}

func getInt(m map[string]any, key string, def int) int {
	if v, ok := m[key]; ok && v != nil {
		switch n := v.(type) {
		case float64:
			return int(n)
		case int:
			return n
		case string:
			fmt.Sscanf(n, "%d", &def)
			return def
		}
	}
	return def
}

func getFloat(m map[string]any, key string, def float64) float64 {
	if v, ok := m[key]; ok && v != nil {
		switch n := v.(type) {
		case float64:
			return n
		case int:
			return float64(n)
		case string:
			fmt.Sscanf(n, "%f", &def)
			return def
		}
	}
	return def
}
