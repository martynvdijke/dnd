package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers/testutil"
)

func TestClaimPartyItem(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/characters", CreateCharacter)
		auth.POST("/campaigns/:id/party-items", CreateCampaignPartyItem)
		auth.POST("/campaigns/:id/party-items/:itemId/claim", ClaimCampaignPartyItem)
	})

	if _, err := db.DB.Exec("INSERT INTO campaigns(id, name, user_id) VALUES(1, 'Loot Campaign', 1)"); err != nil {
		t.Fatalf("seed campaign: %v", err)
	}

	var char struct {
		ID int64 `json:"id"`
	}
	w := testutil.PostJSON(t, r, "/api/characters", map[string]any{"name": "Looter", "level": 1})
	testutil.AssertStatus(t, w, 201)
	testutil.ParseJSON(t, w, &char)
	if _, err := db.DB.Exec("INSERT INTO campaign_characters(campaign_id, character_id) VALUES(1, ?)", char.ID); err != nil {
		t.Fatalf("seed roster: %v", err)
	}

	var item struct {
		ID int64 `json:"id"`
	}
	w = testutil.PostJSON(t, r, "/api/campaigns/1/party-items", map[string]any{
		"name": "Potion of Healing", "quantity": 5, "notes": "dungeon loot",
	})
	testutil.AssertStatus(t, w, 201)
	testutil.ParseJSON(t, w, &item)

	w = testutil.PostJSON(t, r, "/api/campaigns/1/party-items/1/claim", map[string]any{
		"character_id": char.ID, "quantity": 2,
	})
	testutil.AssertStatus(t, w, 200)

	var partyQty, invQty int
	_ = db.DB.QueryRow("SELECT quantity FROM party_items WHERE id=?", item.ID).Scan(&partyQty)
	if partyQty != 3 {
		t.Fatalf("expected 3 remaining party items, got %d", partyQty)
	}
	_ = db.DB.QueryRow("SELECT COALESCE(SUM(quantity),0) FROM inventory WHERE character_id=? AND name='Potion of Healing'", char.ID).Scan(&invQty)
	if invQty != 2 {
		t.Fatalf("expected 2 claimed items in inventory, got %d", invQty)
	}

	// Claiming the remainder removes the party item entirely.
	w = testutil.PostJSON(t, r, "/api/campaigns/1/party-items/1/claim", map[string]any{"character_id": char.ID})
	testutil.AssertStatus(t, w, 200)
	var remaining int
	_ = db.DB.QueryRow("SELECT COUNT(*) FROM party_items WHERE id=?", item.ID).Scan(&remaining)
	if remaining != 0 {
		t.Fatalf("expected the party item to be removed, still %d", remaining)
	}

	// A character outside the campaign cannot claim.
	var outsider struct {
		ID int64 `json:"id"`
	}
	w = testutil.PostJSON(t, r, "/api/characters", map[string]any{"name": "Outsider", "level": 1})
	testutil.AssertStatus(t, w, 201)
	testutil.ParseJSON(t, w, &outsider)
	w = testutil.PostJSON(t, r, "/api/campaigns/1/party-items", map[string]any{"name": "Gem", "quantity": 1})
	testutil.AssertStatus(t, w, 201)
	w = testutil.PostJSON(t, r, "/api/campaigns/1/party-items/2/claim", map[string]any{"character_id": outsider.ID})
	testutil.AssertStatus(t, w, 400)
}
