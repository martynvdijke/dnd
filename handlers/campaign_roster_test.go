package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers/testutil"
)

// seedRosterFixtures sets up: campaign 1 (owned by user 2, member user 3),
// campaign 2 (owned by user 2), and characters across owners:
//   - char 1: owned by user 2 (external to member user 3), unassigned
//   - char 2: owned by user 3 (member), unassigned
//   - char 3: owned by user 4 (stranger, not a member), unassigned
//   - char 4: owned by user 2, assigned to campaign 2
func seedRosterFixtures(t *testing.T) {
	t.Helper()
	testutil.NewDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedUser(t, 2, "owner", "player")
	testutil.SeedUser(t, 3, "member", "player")
	testutil.SeedUser(t, 4, "stranger", "player")
	testutil.SeedCampaign(t, 1, "Campaign One", "Party One", 2)
	testutil.SeedCampaign(t, 2, "Campaign Two", "Party Two", 2)
	testutil.SeedCampaignMember(t, 1, 3, "player")
	testutil.SeedCharacter(t, 1, 2, "OwnerHero", "Elf", "Wizard")
	testutil.SeedCharacter(t, 2, 3, "MemberHero", "Dwarf", "Cleric")
	testutil.SeedCharacter(t, 3, 4, "StrangerHero", "Human", "Fighter")
	testutil.SeedCharacter(t, 5, 3, "MemberHero2", "Elf", "Ranger")
	testutil.SeedCharacterInCampaign(t, 4, 2, 2, "OtherCampHero", "Halfling", "Rogue")
}

func rosterRouter(routes func(*gin.RouterGroup), userID int64, role string) *gin.Engine {
	return testutil.NewRouterWithUser(routes, userID, role)
}

func rosterRoutes(auth *gin.RouterGroup) {
	auth.POST("/campaigns/:id/characters", AddCampaignCharacter)
	auth.DELETE("/campaigns/:id/characters/:characterId", RemoveCampaignCharacter)
	auth.GET("/campaigns/:id/character-candidates", ListCampaignCharacterCandidates)
}

func characterCampaignID(t *testing.T, charID int64) int64 {
	t.Helper()
	var cid int64
	if err := db.DB.QueryRow("SELECT COALESCE(campaign_id,0) FROM characters WHERE id=?", charID).Scan(&cid); err != nil {
		t.Fatalf("read campaign_id of char %d: %v", charID, err)
	}
	return cid
}

func TestAddCampaignCharacter(t *testing.T) {
	seedRosterFixtures(t)
	defer testutil.CloseDB(t)

	t.Run("member adds own character", func(t *testing.T) {
		r := rosterRouter(rosterRoutes, 3, "player")
		w := testutil.PostJSON(t, r, "/api/campaigns/1/characters", map[string]any{"character_id": 2})
		testutil.AssertStatus(t, w, 201)
		if got := characterCampaignID(t, 2); got != 1 {
			t.Fatalf("expected char 2 campaign_id 1, got %d", got)
		}
	})

	t.Run("member adds another member's character", func(t *testing.T) {
		r := rosterRouter(rosterRoutes, 3, "player")
		w := testutil.PostJSON(t, r, "/api/campaigns/1/characters", map[string]any{"character_id": 1})
		testutil.AssertStatus(t, w, 201)
		if got := characterCampaignID(t, 1); got != 1 {
			t.Fatalf("expected char 1 campaign_id 1, got %d", got)
		}
	})

	t.Run("non-member is rejected", func(t *testing.T) {
		r := rosterRouter(rosterRoutes, 4, "player")
		w := testutil.PostJSON(t, r, "/api/campaigns/1/characters", map[string]any{"character_id": 5})
		testutil.AssertStatus(t, w, 403)
		if got := characterCampaignID(t, 5); got != 0 {
			t.Fatalf("expected char 5 unassigned, got campaign_id %d", got)
		}
	})

	t.Run("stranger-owned character is rejected", func(t *testing.T) {
		r := rosterRouter(rosterRoutes, 3, "player")
		w := testutil.PostJSON(t, r, "/api/campaigns/1/characters", map[string]any{"character_id": 3})
		testutil.AssertStatus(t, w, 400)
	})

	t.Run("character in another campaign is rejected", func(t *testing.T) {
		r := rosterRouter(rosterRoutes, 3, "player")
		w := testutil.PostJSON(t, r, "/api/campaigns/1/characters", map[string]any{"character_id": 4})
		testutil.AssertStatus(t, w, 409)
		if got := characterCampaignID(t, 4); got != 2 {
			t.Fatalf("expected char 4 to stay in campaign 2, got %d", got)
		}
	})
}

func TestRemoveCampaignCharacter(t *testing.T) {
	seedRosterFixtures(t)
	defer testutil.CloseDB(t)
	// Two characters seeded directly into campaign 1.
	testutil.SeedCharacterInCampaign(t, 5, 2, 1, "RosterHero", "Gnome", "Bard")
	testutil.SeedCharacterInCampaign(t, 6, 2, 1, "RosterHero2", "Dragonborn", "Monk")

	t.Run("member removes roster character", func(t *testing.T) {
		r := rosterRouter(rosterRoutes, 3, "player")
		w := testutil.Delete(t, r, "/api/campaigns/1/characters/5")
		testutil.AssertStatus(t, w, 200)
		if got := characterCampaignID(t, 5); got != 0 {
			t.Fatalf("expected char 5 unassigned, got campaign_id %d", got)
		}
	})

	t.Run("non-member removal is rejected", func(t *testing.T) {
		r := rosterRouter(rosterRoutes, 4, "player")
		w := testutil.Delete(t, r, "/api/campaigns/1/characters/6")
		testutil.AssertStatus(t, w, 403)
		if got := characterCampaignID(t, 6); got != 1 {
			t.Fatalf("expected char 6 to stay in campaign 1, got %d", got)
		}
	})

	t.Run("detach does not clobber a newer assignment", func(t *testing.T) {
		// Move char 6 to campaign 2, then attempt removal from campaign 1.
		if _, err := db.DB.Exec("UPDATE characters SET campaign_id=2 WHERE id=6"); err != nil {
			t.Fatalf("move char 6: %v", err)
		}
		r := rosterRouter(rosterRoutes, 3, "player")
		w := testutil.Delete(t, r, "/api/campaigns/1/characters/6")
		testutil.AssertStatus(t, w, 200)
		if got := characterCampaignID(t, 6); got != 2 {
			t.Fatalf("expected char 6 to stay in campaign 2, got %d", got)
		}
	})
}

func TestListCampaignCharacterCandidates(t *testing.T) {
	seedRosterFixtures(t)
	defer testutil.CloseDB(t)

	t.Run("candidates include own and external characters", func(t *testing.T) {
		r := rosterRouter(rosterRoutes, 3, "player")
		w := testutil.Get(t, r, "/api/campaigns/1/character-candidates")
		testutil.AssertStatus(t, w, 200)
		var out []map[string]any
		testutil.ParseJSON(t, w, &out)

		var own, external, stranger, otherCamp bool
		for _, c := range out {
			switch int64(c["id"].(float64)) {
			case 1:
				external = true
				if c["owned"] != false {
					t.Fatal("char 1 (owner) should be marked owned:false for member")
				}
				if c["owner_username"] != "owner" {
					t.Fatalf("expected owner_username 'owner', got %v", c["owner_username"])
				}
			case 2:
				own = true
				if c["owned"] != true {
					t.Fatal("char 2 (member's own) should be marked owned:true")
				}
				if c["in_roster"] != false {
					t.Fatal("char 2 should start with in_roster:false")
				}
			case 3:
				stranger = true
			case 4:
				otherCamp = true
			}
		}
		if !own || !external {
			t.Fatalf("expected both own (char 2) and external (char 1) candidates, got %v", out)
		}
		if stranger {
			t.Fatal("stranger-owned character must not appear as a candidate")
		}
		if otherCamp {
			t.Fatal("character assigned to another campaign must not appear as a candidate")
		}
	})

	t.Run("attached characters are flagged in_roster", func(t *testing.T) {
		if _, err := db.DB.Exec("UPDATE characters SET campaign_id=1 WHERE id=1"); err != nil {
			t.Fatalf("attach char 1: %v", err)
		}
		r := rosterRouter(rosterRoutes, 3, "player")
		w := testutil.Get(t, r, "/api/campaigns/1/character-candidates")
		testutil.AssertStatus(t, w, 200)
		var out []map[string]any
		testutil.ParseJSON(t, w, &out)
		for _, c := range out {
			if int64(c["id"].(float64)) == 1 {
				if c["in_roster"] != true {
					t.Fatal("expected attached char 1 flagged in_roster:true")
				}
				return
			}
		}
		t.Fatal("attached char 1 missing from candidates")
	})

	t.Run("non-member is rejected", func(t *testing.T) {
		r := rosterRouter(rosterRoutes, 4, "player")
		w := testutil.Get(t, r, "/api/campaigns/1/character-candidates")
		testutil.AssertStatus(t, w, 403)
	})

	t.Run("unknown campaign returns 404", func(t *testing.T) {
		r := rosterRouter(rosterRoutes, 3, "player")
		w := testutil.Get(t, r, "/api/campaigns/999/character-candidates")
		testutil.AssertStatus(t, w, 404)
	})
}
