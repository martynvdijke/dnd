package models

import "time"

// ─── Dynamic Compendium Schema System ───

type SchemaFieldType string

const (
	FieldTypeString      SchemaFieldType = "string"
	FieldTypeText        SchemaFieldType = "text"
	FieldTypeInteger     SchemaFieldType = "integer"
	FieldTypeFloat       SchemaFieldType = "float"
	FieldTypeBoolean     SchemaFieldType = "boolean"
	FieldTypeJSON        SchemaFieldType = "json"
	FieldTypeSelect      SchemaFieldType = "select"
	FieldTypeMultiSelect SchemaFieldType = "multi-select"
)

type SchemaField struct {
	Name       string          `json:"name"`
	Label      string          `json:"label"`
	Type       SchemaFieldType `json:"type"`
	Required   bool            `json:"required"`
	Default    any             `json:"default,omitempty"`
	Options    []string        `json:"options,omitempty"` // for select/multi-select
	Sortable   bool            `json:"sortable"`
	Searchable bool            `json:"searchable"`
}

type CompendiumSchema struct {
	ID          int64         `json:"id"`
	TypeName    string        `json:"type_name"`
	DisplayName string        `json:"display_name"`
	Fields      []SchemaField `json:"fields"`
	CreatedAt   string        `json:"created_at"`
	UpdatedAt   string        `json:"updated_at"`
	EntryCount  int           `json:"entry_count,omitempty"`
}

type CompendiumEntry struct {
	ID        int64          `json:"id"`
	SchemaID  int64          `json:"schema_id"`
	Data      map[string]any `json:"data"`
	CreatedAt string         `json:"created_at"`
	UpdatedAt string         `json:"updated_at"`
}

type CompendiumEntryList struct {
	Entries    []CompendiumEntry `json:"entries"`
	Total      int               `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}

type CompendiumImportLog struct {
	ID           int64          `json:"id"`
	UserID       int64          `json:"user_id"`
	Status       string         `json:"status"`
	Files        []string       `json:"files"`
	Mapping      map[string]any `json:"mapping"`
	Summary      map[string]any `json:"summary"`
	CreatedAt    string         `json:"created_at"`
	RolledBackAt *string        `json:"rolled_back_at,omitempty"`
}

// CompendiumImportSession tracks the multi-step import state in memory/session
type CompendiumImportSession struct {
	ID           int64                       `json:"id"`
	Files        []CompendiumImportFile      `json:"files"`
	DetectedType string                      `json:"detected_type"`
	SchemaID     int64                       `json:"schema_id"`
	Mapping      map[string]string           `json:"mapping"`
	Entries      []map[string]any            `json:"entries"`
	Duplicates   []CompendiumImportDuplicate `json:"duplicates"`
	Errors       []CompendiumImportError     `json:"errors"`
}

type CompendiumImportFile struct {
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	Entries  int    `json:"entries"`
}

type CompendiumImportDuplicate struct {
	Index      int            `json:"index"`
	ExistingID int64          `json:"existing_id"`
	Existing   map[string]any `json:"existing"`
	Incoming   map[string]any `json:"incoming"`
	Resolved   string         `json:"resolved"` // skip, overwrite, create-new
}

type CompendiumImportError struct {
	Index   int    `json:"index"`
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

// Search result for cross-type search
type CompendiumSearchResult struct {
	Type     string `json:"type"`
	TypeName string `json:"type_name"`
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Snippet  string `json:"snippet"`
}

// ─── Legacy Compendium Types (unchanged) ───

func init() {
	// Parse time for schema timestamps
	_ = time.Time{}
}

type CompendiumRace struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	Speed          int    `json:"speed"`
	Size           string `json:"size"`
	AbilityBonuses string `json:"ability_bonuses"`
	Traits         string `json:"traits"`
	Languages      string `json:"languages"`
	SourcePage     string `json:"source_page"`
	System         string `json:"system,omitempty"`
	Source         string `json:"source,omitempty"`
	Category       string `json:"category,omitempty"`
	Expansion      string `json:"expansion,omitempty"`
	Publisher      string `json:"publisher,omitempty"`
}

type CompendiumClass struct {
	ID                  int64  `json:"id"`
	Name                string `json:"name"`
	Description         string `json:"description"`
	HitDie              int    `json:"hit_die"`
	PrimaryAbility      string `json:"primary_ability"`
	SavingThrows        string `json:"saving_throws"`
	Proficiencies       string `json:"proficiencies"`
	SpellcastingAbility string `json:"spellcasting_ability"`
	SourcePage          string `json:"source_page"`
	System              string `json:"system,omitempty"`
	Source              string `json:"source,omitempty"`
	Category            string `json:"category,omitempty"`
	Expansion           string `json:"expansion,omitempty"`
	Publisher           string `json:"publisher,omitempty"`
}

type CompendiumSpell struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Level        int    `json:"level"`
	School       string `json:"school"`
	CastingTime  string `json:"casting_time"`
	Range        string `json:"range"`
	Components   string `json:"components"`
	Duration     string `json:"duration"`
	Description  string `json:"description"`
	HigherLevels string `json:"higher_levels"`
	Classes      string `json:"classes"`
	SourcePage   string `json:"source_page"`
	System       string `json:"system,omitempty"`
	Source       string `json:"source,omitempty"`
	Publisher    string `json:"publisher,omitempty"`
}

type CompendiumFeat struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Prerequisites string `json:"prerequisites"`
	SourcePage    string `json:"source_page"`
	System        string `json:"system,omitempty"`
	Source        string `json:"source,omitempty"`
}

type CompendiumBackground struct {
	ID                    int64  `json:"id"`
	Name                  string `json:"name"`
	Description           string `json:"description"`
	FeatureName           string `json:"feature_name"`
	FeatureDescription    string `json:"feature_description"`
	Proficiencies         string `json:"proficiencies"`
	SourcePage            string `json:"source_page"`
	System                string `json:"system,omitempty"`
	Source                string `json:"source,omitempty"`
	Category              string `json:"category,omitempty"`
	DataList              bool   `json:"data_list,omitempty"`
	DataBonds             string `json:"data_bonds,omitempty"`
	DataFlaws             string `json:"data_flaws,omitempty"`
	DataIdeals            string `json:"data_ideals,omitempty"`
	DataEquipment         string `json:"data_equipment,omitempty"`
	DataStartingGold      int    `json:"data_starting_gold,omitempty"`
	DataPersonalityTraits string `json:"data_personality_traits,omitempty"`
	Publisher             string `json:"publisher,omitempty"`
}

type CompendiumEquipment struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	Cost        string  `json:"cost"`
	Weight      float64 `json:"weight"`
	Description string  `json:"description"`
	SourcePage  string  `json:"source_page"`
	System      string  `json:"system,omitempty"`
	Source      string  `json:"source,omitempty"`
	ItemType    string  `json:"item_type,omitempty"`
	ItemRarity  string  `json:"item_rarity,omitempty"`
	Publisher   string  `json:"publisher,omitempty"`
}

type ImportCharacter struct {
	Name       string `json:"name"`
	Race       string `json:"race"`
	Class      string `json:"class"`
	Subclass   string `json:"subclass"`
	Level      int    `json:"level"`
	XP         int    `json:"xp"`
	Background string `json:"background"`
	Alignment  string `json:"alignment"`
	Str        int    `json:"str"`
	Dex        int    `json:"dex"`
	Con        int    `json:"con"`
	Int        int    `json:"int"`
	Wis        int    `json:"wis"`
	Cha        int    `json:"cha"`
	AC         int    `json:"ac"`
	Initiative int    `json:"initiative"`
	Speed      int    `json:"speed"`
	HPMax      int    `json:"hp_max"`
	HPCurrent  int    `json:"hp_current"`
	TempHP     int    `json:"temp_hp"`
	HitDice    string `json:"hit_dice"`

	PersonalityTraits string `json:"personality_traits"`
	Ideals            string `json:"ideals"`
	Bonds             string `json:"bonds"`
	Flaws             string `json:"flaws"`
	Appearance        string `json:"appearance"`
	Backstory         string `json:"backstory"`

	Currency      Currency        `json:"currency"`
	Proficiencies []Proficiency   `json:"proficiencies"`
	Features      []Feature       `json:"features"`
	Spellcasting  *Spellcasting   `json:"spellcasting"`
	Spells        []Spell         `json:"spells"`
	Inventory     []InventoryItem `json:"inventory"`
}
