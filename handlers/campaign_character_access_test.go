package handlers

import (
	"database/sql"
	"testing"

	"villum/db"
	"villum/handlers/testutil"
)

// seedCampaignCharacterFixtures builds two campaigns with overlapping rosters:
//
//	campaign 1: Alric (char 11, owned by player 2), Zara (char 10, owned by dm 1)
//	campaign 2: Alric only
//	char 12: Outsider, in no campaign
func seedCampaignCharacterFixtures(t *testing.T) {
	t.Helper()
	testutil.SeedUser(t, 1, "dm", "admin")
	testutil.SeedUser(t, 2, "player", "player")

	testutil.SeedCampaign(t, 1, "Curse of Strahd", "The Party", 1)
	testutil.SeedCampaignMember(t, 1, 2, "player")
	testutil.SeedCampaign(t, 2, "Lost Mine", "Other Party", 1)

	testutil.SeedCharacterInCampaign(t, 10, 1, 1, "Zara", "Elf", "Wizard")
	testutil.SeedCharacterInCampaign(t, 11, 2, 1, "Alric", "Human", "Fighter")
	testutil.SeedCharacterInCampaign(t, 11, 2, 2, "Alric", "Human", "Fighter")
	testutil.SeedCharacter(t, 12, 2, "Outsider", "Halfling", "Rogue")
}

func TestCampaignCharactersIn(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	seedCampaignCharacterFixtures(t)

	chars, err := campaignCharactersIn(db.DB, 1)
	if err != nil {
		t.Fatalf("campaignCharactersIn: %v", err)
	}
	if len(chars) != 2 {
		t.Fatalf("want 2 characters in campaign 1, got %d: %+v", len(chars), chars)
	}
	// Ordered by name, as every converted caller relied on.
	if chars[0].Name != "Alric" || chars[1].Name != "Zara" {
		t.Errorf("unexpected order: %q, %q", chars[0].Name, chars[1].Name)
	}
	alric := chars[0]
	if alric.ID != 11 || alric.Race != "Human" || alric.Class != "Fighter" {
		t.Errorf("unexpected identity for Alric: %+v", alric)
	}
	// SeedCharacter writes level 1, hp 12/12, ac 10.
	if alric.Level != 1 || alric.HPCurrent != 12 || alric.HPMax != 12 || alric.AC != 10 {
		t.Errorf("unexpected stats for Alric: %+v", alric)
	}

	// A character in no campaign is invisible, and an unknown campaign is empty.
	if got, err := campaignCharactersIn(db.DB, 99); err != nil || len(got) != 0 {
		t.Errorf("campaign 99: want empty, got %+v (err=%v)", got, err)
	}
}

func TestCampaignIDsForCharacter(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	seedCampaignCharacterFixtures(t)

	ids, err := campaignIDsForCharacter(db.DB, 11)
	if err != nil {
		t.Fatalf("campaignIDsForCharacter: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("want both memberships for char 11, got %v", ids)
	}
	if ids, err := campaignIDsForCharacter(db.DB, 12); err != nil || len(ids) != 0 {
		t.Errorf("char 12: want no memberships, got %v (err=%v)", ids, err)
	}
}

func TestFirstCampaignIDForCharacter(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	seedCampaignCharacterFixtures(t)

	// Bookkeeping uses the lowest membership id, not insertion order.
	campaignID, err := firstCampaignIDForCharacter(db.DB, 11)
	if err != nil {
		t.Fatalf("firstCampaignIDForCharacter: %v", err)
	}
	if campaignID != 1 {
		t.Errorf("want campaign 1 as first membership, got %d", campaignID)
	}

	if _, err := firstCampaignIDForCharacter(db.DB, 12); err != sql.ErrNoRows {
		t.Errorf("char without campaign: want sql.ErrNoRows, got %v", err)
	}
}

func TestIsCampaignDMForCharacter(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	seedCampaignCharacterFixtures(t)

	if !isCampaignDMForCharacter(db.DB, int64(11), int64(1)) {
		t.Error("dm 1 should be seen as DM of Alric")
	}
	if isCampaignDMForCharacter(db.DB, int64(11), int64(2)) {
		t.Error("player 2 is a member, not a DM, of Alric's campaign")
	}
	if isCampaignDMForCharacter(db.DB, int64(12), int64(1)) {
		t.Error("a character in no campaign has no DM")
	}
	// The note handlers pass the query-string character id and the gin user_id
	// value straight through; SQLite coerces them against the integer columns.
	if !isCampaignDMForCharacter(db.DB, "11", int64(1)) {
		t.Error("string character id should bind the same as an integer id")
	}
}

// TestCampaignCharacterJoinSQL pins the shared visibility fragment against a
// character-owned child table: only rows whose owning character is in the
// campaign are counted.
func TestCampaignCharacterJoinSQL(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	seedCampaignCharacterFixtures(t)

	for _, q := range []struct {
		characterID int64
		name        string
	}{
		{11, "in campaign"},
		{12, "not in campaign"},
	} {
		if _, err := db.DB.Exec("INSERT INTO quests(character_id, name, status) VALUES(?,?,?)", q.characterID, q.name, "active"); err != nil {
			t.Fatalf("seed quest for char %d: %v", q.characterID, err)
		}
	}

	var count int
	err := db.DB.QueryRow(
		"SELECT COUNT(*) FROM quests q JOIN characters c ON c.id = q.character_id "+campaignCharacterJoinSQL+" WHERE cm.campaign_id=?",
		1).Scan(&count)
	if err != nil {
		t.Fatalf("count campaign quests: %v", err)
	}
	if count != 1 {
		t.Errorf("want 1 quest for campaign 1, got %d", count)
	}
}
