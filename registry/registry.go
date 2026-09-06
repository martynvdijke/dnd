// Package registry is the single place where entity types declare how they
// participate in universal search, import/export (transfer), and cross-entity
// linking. New entity types onboard by adding one entry to Entities.
package registry

import "database/sql"

// Ownership describes how visibility of an entity is determined.
type Ownership int

const (
	// OwnerUser: table has a user_id column; visible to that user.
	OwnerUser Ownership = iota
	// OwnerCharacter: table has character_id; visible to the character's owner.
	OwnerCharacter
	// OwnerCampaign: table has campaign_id; visible to campaign owner + members.
	OwnerCampaign
	// OwnerUserOrCampaign: table has user_id and nullable campaign_id; visible
	// to the owner or any member of the campaign.
	OwnerUserOrCampaign
	// OwnerAdventure: table has adventure_id; visible to the adventure's owner.
	OwnerAdventure
	// OwnerGlobal: visible to every authenticated user (compendium).
	OwnerGlobal
	// OwnerCampaignShared: campaign owner sees all; members only when shared=1.
	OwnerCampaignShared
)

// EntityInfo describes one linkable/searchable/transferable entity type.
type EntityInfo struct {
	Type         string    // canonical identifier used in APIs and entity_links (e.g. "npc")
	Label        string    // plural display label (e.g. "NPCs")
	Icon         string    // FontAwesome icon class for UI chips/cards
	Table        string    // backing SQL table
	Ownership    Ownership // visibility rule
	Searchable   bool      // present in entity_search_index
	Transferable bool      // supports villum-transfer export/import
	Linkable     bool      // may appear in entity_links
}

// campaignMemberSubquery resolves campaign IDs the user owns or is a member of.
// Two placeholders, both bound to the user ID.
const campaignMemberSubquery = `(SELECT id FROM campaigns WHERE user_id = ? UNION SELECT campaign_id FROM campaign_members WHERE user_id = ?)`

// Entities is the canonical registry keyed by entity type.
var Entities = map[string]EntityInfo{
	"character":  {Type: "character", Label: "Characters", Icon: "fa-users", Table: "characters", Ownership: OwnerUser, Searchable: true, Transferable: true, Linkable: true},
	"npc":        {Type: "npc", Label: "NPCs", Icon: "fa-user-group", Table: "npcs", Ownership: OwnerUser, Searchable: true, Transferable: true, Linkable: true},
	"note":       {Type: "note", Label: "Notes", Icon: "fa-note-sticky", Table: "character_notes", Ownership: OwnerCharacter, Searchable: true, Transferable: false, Linkable: true},
	"quest":      {Type: "quest", Label: "Quests", Icon: "fa-scroll", Table: "quests", Ownership: OwnerCharacter, Searchable: true, Transferable: true, Linkable: true},
	"journal":    {Type: "journal", Label: "Journal", Icon: "fa-book-open", Table: "journal", Ownership: OwnerCharacter, Searchable: true, Transferable: true, Linkable: true},
	"session":    {Type: "session", Label: "Sessions", Icon: "fa-calendar", Table: "sessions", Ownership: OwnerCharacter, Searchable: true, Transferable: false, Linkable: true},
	"campaign":   {Type: "campaign", Label: "Campaigns", Icon: "fa-flag", Table: "campaigns", Ownership: OwnerCampaign, Searchable: true, Transferable: true, Linkable: true},
	"location":   {Type: "location", Label: "Locations", Icon: "fa-location-dot", Table: "locations", Ownership: OwnerUser, Searchable: true, Transferable: true, Linkable: true},
	"encounter":  {Type: "encounter", Label: "Encounters", Icon: "fa-dice-d20", Table: "encounter_templates", Ownership: OwnerUserOrCampaign, Searchable: true, Transferable: true, Linkable: true},
	"monster":    {Type: "monster", Label: "Monsters", Icon: "fa-dragon", Table: "monster_library", Ownership: OwnerUser, Searchable: true, Transferable: true, Linkable: true},
	"shop":       {Type: "shop", Label: "Shops", Icon: "fa-store", Table: "shops", Ownership: OwnerUserOrCampaign, Searchable: true, Transferable: true, Linkable: true},
	"faction":    {Type: "faction", Label: "Factions", Icon: "fa-chess-rook", Table: "factions", Ownership: OwnerCampaign, Searchable: true, Transferable: true, Linkable: true},
	"adventure":  {Type: "adventure", Label: "Adventures", Icon: "fa-map", Table: "oneshot_adventures", Ownership: OwnerUserOrCampaign, Searchable: true, Transferable: true, Linkable: true},
	"wiki":       {Type: "wiki", Label: "Wiki Pages", Icon: "fa-book", Table: "campaign_wiki_pages", Ownership: OwnerCampaign, Searchable: true, Transferable: false, Linkable: true},
	"timeline":   {Type: "timeline", Label: "Timeline Events", Icon: "fa-timeline", Table: "campaign_timeline_events", Ownership: OwnerCampaign, Searchable: true, Transferable: true, Linkable: true},
	"knowledge":  {Type: "knowledge", Label: "Knowledge", Icon: "fa-lightbulb", Table: "campaign_knowledge", Ownership: OwnerCampaignShared, Searchable: true, Transferable: true, Linkable: true},
	"item":       {Type: "item", Label: "Items", Icon: "fa-backpack", Table: "oneshot_items", Ownership: OwnerAdventure, Searchable: true, Transferable: false, Linkable: true},
	"compendium": {Type: "compendium", Label: "Compendium", Icon: "fa-spell-book", Table: "compendium_entries", Ownership: OwnerGlobal, Searchable: true, Transferable: false, Linkable: true},
}

// Get returns the EntityInfo for a type and whether it exists.
func Get(entityType string) (EntityInfo, bool) {
	info, ok := Entities[entityType]
	return info, ok
}

// Linkable reports whether the type may participate in entity_links.
func Linkable(entityType string) bool {
	info, ok := Entities[entityType]
	return ok && info.Linkable
}

// SearchableTypes returns the entity types included in the unified search index.
func SearchableTypes() []EntityInfo {
	out := make([]EntityInfo, 0, len(Entities))
	for _, info := range Entities {
		if info.Searchable {
			out = append(out, info)
		}
	}
	return out
}

// TransferableTypes returns the entity types supporting villum-transfer export/import.
func TransferableTypes() []EntityInfo {
	out := make([]EntityInfo, 0, len(Entities))
	for _, info := range Entities {
		if info.Transferable {
			out = append(out, info)
		}
	}
	return out
}

// VisibilityWhere returns a SQL WHERE fragment (over the entity's table) that
// restricts rows to those visible to a non-admin user. It returns the fragment
// and the number of bind parameters (all bound to user ID).
func VisibilityWhere(o Ownership) (fragment string, argCount int) {
	switch o {
	case OwnerUser:
		return `user_id = ?`, 1
	case OwnerCharacter:
		return `character_id IN (SELECT id FROM characters WHERE user_id = ?)`, 1
	case OwnerCampaign:
		return `campaign_id IN ` + campaignMemberSubquery, 2
	case OwnerCampaignShared:
		return `(campaign_id IN (SELECT id FROM campaigns WHERE user_id = ?) OR (shared = 1 AND campaign_id IN ` + campaignMemberSubquery + `))`, 3
	case OwnerUserOrCampaign:
		return `(user_id = ? OR (campaign_id IS NOT NULL AND campaign_id IN ` + campaignMemberSubquery + `))`, 3
	case OwnerAdventure:
		return `adventure_id IN (SELECT id FROM oneshot_adventures WHERE user_id = ?)`, 1
	case OwnerGlobal:
		return `1 = 1`, 0
	}
	return `1 = 0`, 0
}

// VisibleIDs filters candidate entity IDs of the given type down to those
// visible to the user. Admins (and the "dm" role, which is trusted with
// campaign content) see everything: pass admin=true to skip filtering.
// queryer is satisfied by both *sql.DB and *sql.Tx.
// Queryer is satisfied by *sql.DB and *sql.Tx.
type Queryer interface {
	Query(query string, args ...any) (*sql.Rows, error)
}

// QueryRower is satisfied by *sql.DB and *sql.Tx.
type QueryRower interface {
	QueryRow(query string, args ...any) *sql.Row
}

// Editable reports whether the user may edit the given entity (used to
// authorize creating/removing links whose source is that entity).
// Admins may edit everything. Compendium entries are never editable as
// link sources (they are read-only reference targets).
func Editable(queryer QueryRower, entityType string, id int64, userID int64, admin bool) bool {
	info, ok := Entities[entityType]
	if !ok || !info.Linkable {
		return false
	}
	if admin {
		return true
	}
	var where string
	var args []any
	switch info.Ownership {
	case OwnerUser:
		where = `user_id = ?`
		args = []any{id, userID}
	case OwnerCharacter:
		where = `character_id IN (SELECT id FROM characters WHERE user_id = ?)`
		args = []any{id, userID}
	case OwnerCampaign:
		where = `(campaign_id IN (SELECT id FROM campaigns WHERE user_id = ?) OR campaign_id IN (SELECT campaign_id FROM campaign_members WHERE user_id = ? AND role = 'dm'))`
		args = []any{id, userID, userID}
	case OwnerUserOrCampaign:
		where = `(user_id = ? OR campaign_id IN (SELECT campaign_id FROM campaign_members WHERE user_id = ? AND role = 'dm'))`
		args = []any{id, userID, userID}
	case OwnerAdventure:
		where = `adventure_id IN (SELECT id FROM oneshot_adventures WHERE user_id = ?)`
		args = []any{id, userID}
	case OwnerCampaignShared:
		where = `(campaign_id IN (SELECT id FROM campaigns WHERE user_id = ?) OR (shared = 1 AND campaign_id IN (SELECT campaign_id FROM campaign_members WHERE user_id = ?)))`
		args = []any{id, userID, userID}
	case OwnerGlobal:
		return false
	default:
		return false
	}
	var n int
	row := queryer.QueryRow(`SELECT COUNT(*) FROM `+info.Table+` WHERE id = ? AND `+where, args...)
	if err := row.Scan(&n); err != nil {
		return false
	}
	return n > 0
}

func VisibleIDs(queryer Queryer, entityType string, ids []int64, userID int64, admin bool) (map[int64]bool, error) {
	visible := make(map[int64]bool, len(ids))
	info, ok := Entities[entityType]
	if !ok || len(ids) == 0 {
		return visible, nil
	}
	if admin {
		for _, id := range ids {
			visible[id] = true
		}
		return visible, nil
	}
	where, n := VisibilityWhere(info.Ownership)
	args := make([]any, 0, len(ids)+n)
	placeholders := ""
	for i, id := range ids {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args = append(args, id)
	}
	for i := 0; i < n; i++ {
		args = append(args, userID)
	}
	rows, err := queryer.Query(
		`SELECT id FROM `+info.Table+` WHERE id IN (`+placeholders+`) AND `+where, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		visible[id] = true
	}
	return visible, rows.Err()
}
