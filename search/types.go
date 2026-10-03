package search

import "fmt"

// SearchResultItem is legacy flat result.
type SearchResultItem struct {
	Type    string `json:"type"`
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Snippet string `json:"snippet,omitempty"`
	Subtype string `json:"subtype,omitempty"`
}

type SearchResults struct {
	Characters  []SearchResultItem `json:"characters"`
	NPCs        []SearchResultItem `json:"npcs"`
	Notes       []SearchResultItem `json:"notes"`
	Quests      []SearchResultItem `json:"quests"`
	Journal     []SearchResultItem `json:"journal"`
	Sessions    []SearchResultItem `json:"sessions"`
	Spells      []SearchResultItem `json:"spells"`
	Equipment   []SearchResultItem `json:"equipment"`
	Races       []SearchResultItem `json:"races"`
	Classes     []SearchResultItem `json:"classes"`
	Feats       []SearchResultItem `json:"feats"`
	Backgrounds []SearchResultItem `json:"backgrounds"`
	Campaigns   []SearchResultItem `json:"campaigns"`
	Monsters    []SearchResultItem `json:"monsters"`
}

var EntityTypeLegacyBucket = map[string]string{
	"character": "Characters",
	"npc":       "NPCs",
	"note":      "Notes",
	"quest":     "Quests",
	"journal":   "Journal",
	"session":   "Sessions",
	"campaign":  "Campaigns",
	"monster":   "Monsters",
}

func EntityTypeName(et string) string {
	switch et {
	case "character":
		return "Character"
	case "npc":
		return "NPC"
	case "note":
		return "Note"
	case "quest":
		return "Quest"
	case "journal":
		return "Journal"
	case "session":
		return "Session"
	case "campaign":
		return "Campaign"
	case "location":
		return "Location"
	case "encounter":
		return "Encounter"
	case "monster":
		return "Monster"
	case "shop":
		return "Shop"
	case "faction":
		return "Faction"
	case "adventure":
		return "Adventure"
	case "wiki":
		return "Wiki Page"
	case "timeline":
		return "Timeline Event"
	case "item":
		return "Item"
	case "knowledge":
		return "Knowledge"
	case "compendium":
		return "Compendium"
	}
	return et
}

func EmptyLegacy() SearchResults {
	return SearchResults{
		Characters: []SearchResultItem{}, NPCs: []SearchResultItem{}, Notes: []SearchResultItem{}, Quests: []SearchResultItem{},
		Journal: []SearchResultItem{}, Sessions: []SearchResultItem{}, Spells: []SearchResultItem{}, Equipment: []SearchResultItem{},
		Races: []SearchResultItem{}, Classes: []SearchResultItem{}, Feats: []SearchResultItem{}, Backgrounds: []SearchResultItem{},
		Campaigns: []SearchResultItem{}, Monsters: []SearchResultItem{},
	}
}

func LegacyBucketPtr(legacy *SearchResults, et string) *[]SearchResultItem {
	switch et {
	case "character":
		return &legacy.Characters
	case "npc":
		return &legacy.NPCs
	case "note":
		return &legacy.Notes
	case "quest":
		return &legacy.Quests
	case "journal":
		return &legacy.Journal
	case "session":
		return &legacy.Sessions
	case "campaign":
		return &legacy.Campaigns
	case "monster":
		return &legacy.Monsters
	}
	return nil
}

func EntityURL(et string, id int64) string {
	return entityURLImpl(et, id)
}
func entityURLImpl(et string, id int64) string {
	// implemented in unified.go via init alias; fallback
	switch et {
	case "character":
		return fmt.Sprintf("#/characters/%d", id)
	case "npc":
		return fmt.Sprintf("#/npcs/%d", id)
	case "note":
		return fmt.Sprintf("#/notes/%d", id)
	case "quest":
		return fmt.Sprintf("#/quests/%d", id)
	case "journal":
		return fmt.Sprintf("#/journal/%d", id)
	case "session":
		return fmt.Sprintf("#/sessions/%d", id)
	case "campaign":
		return fmt.Sprintf("#/campaigns/%d", id)
	case "location":
		return fmt.Sprintf("#/locations/%d", id)
	case "encounter":
		return fmt.Sprintf("#/encounters/%d", id)
	case "monster":
		return fmt.Sprintf("#/monsters/%d", id)
	case "shop":
		return fmt.Sprintf("#/shops/%d", id)
	case "faction":
		return fmt.Sprintf("#/factions/%d", id)
	case "adventure":
		return fmt.Sprintf("#/adventures/%d", id)
	case "wiki":
		return fmt.Sprintf("#/wiki/%d", id)
	case "timeline":
		return fmt.Sprintf("#/timeline/%d", id)
	case "item":
		return fmt.Sprintf("#/items/%d", id)
	case "recap":
		return "#/recaps"
	case "knowledge":
		return fmt.Sprintf("#/campaigns/knowledge/%d", id)
	case "compendium":
		return fmt.Sprintf("#/compendium/%d", id)
	}
	return ""
}

// UnifiedResult is the unified search result.
type UnifiedResult struct {
	EntityType string  `json:"entity_type"`
	EntityID   int64   `json:"entity_id"`
	Title      string  `json:"title"`
	Subtitle   string  `json:"subtitle"`
	Snippet    string  `json:"snippet"`
	Score      float64 `json:"score"`
	URL        string  `json:"url"`
}

// UnifiedParams are parameters for SearchUnified.
type UnifiedParams struct {
	Query       string
	TypesFilter string // comma-separated entity types, empty means all
	Limit       int
	UserID      int64
	IsAdmin     bool
}

// CompendiumResult is a single compendium search hit.
type CompendiumResult struct {
	Type           string  `json:"type"`
	ID             int64   `json:"id"`
	Name           string  `json:"name"`
	Subtype        string  `json:"subtype,omitempty"`
	CR             string  `json:"cr,omitempty"`
	Level          int     `json:"level,omitempty"`
	HitDie         int     `json:"hit_die,omitempty"`
	PrimaryAbility string  `json:"primary_ability,omitempty"`
	Score          float64 `json:"score,omitempty"`
}

// CompendiumParams are parameters for SearchCompendium.
type CompendiumParams struct {
	Query       string
	TypeFilter  string
	Class       string
	Level       string
	School      string
	Category    string
	CR          string
	MonsterType string
	Reranker    func(query string, in []CompendiumResult) []CompendiumResult
}

// EntryData is a flattened generic compendium entry.
type EntryData struct {
	Name        string
	Category    string
	Cost        string
	Description string
	Source      string
	Weight      float64
	Level       int
	School      string
	CastingTime string
	Range       string
	Components  string
	Duration    string
	Classes     string
}

// Source is a campaign RAG source (mirrors handlers copilotSource).
type Source struct {
	EntityType string `json:"entity_type"`
	EntityID   int64  `json:"entity_id"`
	Title      string `json:"title"`
	Subtitle   string `json:"subtitle"`
	URL        string `json:"url"`
	Snippet    string `json:"snippet"`
}
