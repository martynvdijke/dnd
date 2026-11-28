package handlers

import (
	"database/sql"
)

// Campaign-scoped character access — the shared shapes that used to be
// hand-written in 16 handlers (TD-006).
//
// A "campaign character" is a character that is a member of a campaign through
// the campaign_characters join table. Every campaign-scoped read in the handlers
// used to spell that membership out inline, which meant the same schema names and
// the same visibility rule lived in ~30 places and drifted independently.

// campaignCharacterJoinSQL is the standard campaign-visibility join for a query
// that already scans characters aliased as "c":
//
//	FROM characters c <campaignCharacterJoinSQL> WHERE cm.campaign_id=?
//
// Child-table queries keep their own join to characters (which also preserves the
// exclusion of orphaned child rows) and then append this fragment:
//
//	FROM quests q JOIN characters c ON c.id = q.character_id
//	    <campaignCharacterJoinSQL> WHERE cm.campaign_id=? AND ...
//
// The fragment is a compile-time constant (never client input), so appending it
// with '+' is safe; see scripts/ci/check-concatenated-sql.sh.
const campaignCharacterJoinSQL = "JOIN campaign_characters cm ON cm.character_id = c.id"

// campaignCharacterRow is the shared projection for "the party of a campaign".
// It is a superset of the columns the dashboard, DM screen, recap builder and
// oneshot overview read, so all of them share one query and one ordering.
type campaignCharacterRow struct {
	ID          int64
	Name        string
	Race        string
	Class       string
	Level       int
	HPCurrent   int
	HPMax       int
	AC          int
	PortraitURL string
}

// campaignCharactersIn returns the characters belonging to a campaign, ordered by
// name. It is the read behind the campaign dashboard, the DM screen, the recap
// template builder and the oneshot overview.
func campaignCharactersIn(db *sql.DB, campaignID int64) ([]campaignCharacterRow, error) {
	rows, err := db.Query(
		"SELECT c.id, c.name, c.race, c.class, c.level, c.hp_current, c.hp_max, c.ac, COALESCE(c.portrait_url,'') "+
			"FROM characters c "+campaignCharacterJoinSQL+" WHERE cm.campaign_id=? ORDER BY c.name", campaignID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chars []campaignCharacterRow
	for rows.Next() {
		var ch campaignCharacterRow
		// Scan errors leave the row's remaining fields zeroed and the row is
		// still reported, matching the previous per-handler behaviour.
		rows.Scan(&ch.ID, &ch.Name, &ch.Race, &ch.Class, &ch.Level, &ch.HPCurrent, &ch.HPMax, &ch.AC, &ch.PortraitURL)
		chars = append(chars, ch)
	}
	return chars, rows.Err()
}

// isCampaignDMForCharacter reports whether userID is the DM of any campaign the
// character belongs to. Used to gate "dm" visibility on character notes.
//
// characterID and userID are bound exactly as given so callers can pass the
// request's own shapes (a query-string character id, the gin user_id value);
// SQLite coerces them against the integer columns, as the inline queries did.
func isCampaignDMForCharacter(db *sql.DB, characterID, userID any) bool {
	var isDM bool
	db.QueryRow(
		`SELECT COUNT(*) > 0 FROM campaign_characters cc JOIN campaign_members cm ON cm.campaign_id = cc.campaign_id WHERE cc.character_id=? AND cm.user_id=? AND cm.role='dm'`,
		characterID, userID).Scan(&isDM)
	return isDM
}

// firstCampaignIDForCharacter returns the character's lowest campaign id, i.e. the
// membership used for bookkeeping when a character belongs to several campaigns.
// It returns sql.ErrNoRows when the character is in no campaign.
func firstCampaignIDForCharacter(db *sql.DB, characterID int64) (int64, error) {
	var campaignID int64
	err := db.QueryRow(
		"SELECT campaign_id FROM campaign_characters WHERE character_id=? ORDER BY campaign_id LIMIT 1",
		characterID).Scan(&campaignID)
	return campaignID, err
}

// campaignIDsForCharacter returns every campaign the character belongs to, unordered.
func campaignIDsForCharacter(db *sql.DB, characterID int64) ([]int64, error) {
	rows, err := db.Query("SELECT campaign_id FROM campaign_characters WHERE character_id=?", characterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if rows.Scan(&id) == nil {
			ids = append(ids, id)
		}
	}
	return ids, rows.Err()
}
