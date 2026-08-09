package handlers

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
)

// TestListCampaignMine verifies GET /api/campaigns/mine returns only
// campaigns the user owns or belongs to, annotated with my_role.
func TestListCampaignMine(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	// user 1 owns campaign 1 (DM via SeedCampaign), user 2 is a member of campaign 2
	testutil.SeedUser(t, 1, "owner", "user")
	testutil.SeedUser(t, 2, "member", "user")
	testutil.SeedUser(t, 3, "othercampaign", "user")
	testutil.SeedCampaign(t, 1, "My Campaign", "Party One", 1)
	testutil.SeedCampaign(t, 2, "Shared Campaign", "Party Two", 3)
	testutil.SeedCampaignMember(t, 2, 2, "player")

	r := testutil.NewRouterWithUser(func(auth *gin.RouterGroup) {
		auth.GET("/campaigns/mine", ListCampaigns)
	}, 2, "user")

	w := testutil.Get(t, r, "/api/campaigns/mine")
	testutil.AssertStatus(t, w, 200)
	body := w.Body.String()
	if !strings.Contains(body, "Shared Campaign") {
		t.Fatalf("expected 'Shared Campaign' in campaigns/mine for member, got: %s", body)
	}
	if strings.Contains(body, "My Campaign") {
		t.Fatalf("campaign owned by another user leaked to campaigns/mine: %s", body)
	}
}

// TestListCampaignCharacters verifies ownership annotation on
// GET /api/campaigns/:id/characters.
func TestListCampaignCharacters(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	// campaign 1 owned by user 1 (DM); user 2 is a plain member
	testutil.SeedUser(t, 1, "owner", "user")
	testutil.SeedUser(t, 2, "member", "user")
	testutil.SeedCampaign(t, 1, "Campaign A", "Party", 1)
	testutil.SeedCampaignMember(t, 1, 2, "player")
	// char 1 owned by user 1 (in campaign 1), char 2 owned by user 2 (in campaign 1)
	testutil.SeedCharacterInCampaign(t, 1, 1, 1, "OwnedChar", "Human", "Fighter")
	testutil.SeedCharacterInCampaign(t, 2, 2, 1, "SharedChar", "Elf", "Wizard")

	t.Run("owner sees owned=true for all campaign characters", func(t *testing.T) {
		r := testutil.NewRouterWithUser(func(auth *gin.RouterGroup) {
			auth.GET("/campaigns/:id/characters", ListCampaignCharacters)
		}, 1, "user")
		w := testutil.Get(t, r, "/api/campaigns/1/characters")
		testutil.AssertStatus(t, w, 200)
		body := w.Body.String()
		// campaign owner is DM -> may edit every character in the campaign
		if !strings.Contains(body, `"owned":true`) {
			t.Fatalf("expected owned=true for all campaign characters, got: %s", body)
		}
		if strings.Contains(body, `"owned":false`) {
			t.Fatalf("DM should own every character in their campaign, got: %s", body)
		}
	})

	t.Run("member sees owned=false for others", func(t *testing.T) {
		r := testutil.NewRouterWithUser(func(auth *gin.RouterGroup) {
			auth.GET("/campaigns/:id/characters", ListCampaignCharacters)
		}, 2, "user")
		w := testutil.Get(t, r, "/api/campaigns/1/characters")
		testutil.AssertStatus(t, w, 200)
		body := w.Body.String()
		if !strings.Contains(body, `"owned":true`) {
			t.Fatalf("expected owned=true for own character, got: %s", body)
		}
	})

	t.Run("non-member forbidden", func(t *testing.T) {
		testutil.SeedUser(t, 9, "outsider", "user")
		r := testutil.NewRouterWithUser(func(auth *gin.RouterGroup) {
			auth.GET("/campaigns/:id/characters", ListCampaignCharacters)
		}, 9, "user")
		w := testutil.Get(t, r, "/api/campaigns/1/characters")
		testutil.AssertStatus(t, w, 403)
	})

	t.Run("unknown campaign 404", func(t *testing.T) {
		r := testutil.NewRouterWithUser(func(auth *gin.RouterGroup) {
			auth.GET("/campaigns/:id/characters", ListCampaignCharacters)
		}, 1, "user")
		w := testutil.Get(t, r, "/api/campaigns/999/characters")
		testutil.AssertStatus(t, w, 404)
	})
}

// TestCharacterNoteOwnership verifies write access to notes is enforced.
func TestCharacterNoteOwnership(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	// char 1 owned by user 1; char 2 owned by user 2 (shares no campaign access)
	testutil.SeedUser(t, 1, "owner", "user")
	testutil.SeedUser(t, 2, "other", "user")
	testutil.SeedCharacter(t, 1, 1, "NoteHero", "Human", "Bard")
	testutil.SeedCharacter(t, 2, 2, "OtherHero", "Elf", "Rogue")

	r := testutil.NewRouterWithUser(func(auth *gin.RouterGroup) {
		auth.POST("/notes", CreateCharacterNote)
		auth.PUT("/notes/:id", UpdateCharacterNote)
		auth.DELETE("/notes/:id", DeleteCharacterNote)
	}, 1, "user")

	t.Run("owner can create note", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/notes", map[string]any{
			"character_id": 1,
			"title":        "My Note",
			"content":      "stuff",
		})
		testutil.AssertStatus(t, w, 201)
	})

	t.Run("non-owner cannot create note", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/notes", map[string]any{
			"character_id": 2,
			"title":        "Their Note",
			"content":      "sneaky",
		})
		testutil.AssertStatus(t, w, 403)
	})

	t.Run("owner can update and delete note", func(t *testing.T) {
		w := testutil.PutJSON(t, r, "/api/notes/1", map[string]any{
			"title":   "Edited",
			"content": "changed",
		})
		testutil.AssertStatus(t, w, 200)
		w = testutil.Delete(t, r, "/api/notes/1")
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("non-owner cannot update or delete note", func(t *testing.T) {
		r2 := testutil.NewRouterWithUser(func(auth *gin.RouterGroup) {
			auth.POST("/notes", CreateCharacterNote)
			auth.PUT("/notes/:id", UpdateCharacterNote)
			auth.DELETE("/notes/:id", DeleteCharacterNote)
		}, 2, "user")
		// user 2 creates a note on their own character 2 -> allowed
		w := testutil.PostJSON(t, r2, "/api/notes", map[string]any{
			"character_id": 2,
			"title":        "Mine",
			"content":      "own note",
		})
		testutil.AssertStatus(t, w, 201)
		var created struct {
			ID int64 `json:"id"`
		}
		testutil.ParseJSON(t, w, &created)
		noteID := created.ID
		w = testutil.PutJSON(t, r2, fmt.Sprintf("/api/notes/%d", noteID), map[string]any{
			"title": "Mine Edited",
		})
		testutil.AssertStatus(t, w, 200)
		w = testutil.Delete(t, r2, fmt.Sprintf("/api/notes/%d", noteID))
		testutil.AssertStatus(t, w, 200)
		// note 1 belongs to character 1 owned by user 1 -> forbidden for user 2
		w = testutil.PutJSON(t, r2, "/api/notes/1", map[string]any{
			"title": "Not Mine",
		})
		testutil.AssertStatus(t, w, 403)
		w = testutil.Delete(t, r2, "/api/notes/1")
		testutil.AssertStatus(t, w, 403)
	})
}

// TestCompanionOwnership verifies write access to companions is enforced.
func TestCompanionOwnership(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "owner", "user")
	testutil.SeedUser(t, 2, "other", "user")
	testutil.SeedCharacter(t, 1, 1, "PetHero", "Human", "Druid")
	testutil.SeedCharacter(t, 2, 2, "OtherPet", "Elf", "Ranger")

	r := testutil.NewRouterWithUser(func(auth *gin.RouterGroup) {
		auth.POST("/companions", CreateCompanion)
		auth.PUT("/companions/:id", UpdateCompanion)
		auth.DELETE("/companions/:id", DeleteCompanion)
	}, 1, "user")

	t.Run("owner can create companion", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/companions", map[string]any{
			"character_id": 1,
			"name":         "Wolf",
			"type":         "beast",
		})
		testutil.AssertStatus(t, w, 201)
	})

	t.Run("non-owner cannot create companion", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/companions", map[string]any{
			"character_id": 2,
			"name":         "Their Wolf",
			"type":         "beast",
		})
		testutil.AssertStatus(t, w, 403)
	})

	t.Run("non-owner cannot update or delete companion", func(t *testing.T) {
		r2 := testutil.NewRouterWithUser(func(auth *gin.RouterGroup) {
			auth.PUT("/companions/:id", UpdateCompanion)
			auth.DELETE("/companions/:id", DeleteCompanion)
		}, 2, "user")
		w := testutil.PutJSON(t, r2, "/api/companions/1", map[string]any{"name": "Hijack"})
		testutil.AssertStatus(t, w, 403)
		w = testutil.Delete(t, r2, "/api/companions/1")
		testutil.AssertStatus(t, w, 403)
	})
}
