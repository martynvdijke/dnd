package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers/testutil"
)

// ─── Shop item ↔ compendium equipment ───

func shopLinkTestRouter() *gin.Engine {
	return testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/campaigns/:id/shops", CreateShop)
		auth.POST("/shops/:id/items", CreateShopItem)
		auth.GET("/shops/:id/items", ListShopItems)
		auth.PUT("/shop-items/:id", UpdateShopItem)
		auth.DELETE("/shop-items/:id/link", UnlinkShopItem)
		auth.POST("/shops/:id/buy", BuyItem)
		auth.GET("/characters/:id", GetCharacter)
	})
}

func TestShopItemCompendiumLink(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 5, "Shop Camp", "Party", 1)
	testutil.SeedCharacter(t, 7, 1, "Buyer", "Human", "Fighter")
	testutil.SeedCompendiumEquipment(t, 500, "Test Blade")

	if _, err := db.DB.Exec(`INSERT OR REPLACE INTO character_currency (character_id, gp) VALUES (7, 1000)`); err != nil {
		t.Fatalf("set gp: %v", err)
	}
	if _, err := db.DB.Exec(`INSERT INTO shops (id, user_id, campaign_id, name, description, markup_percent, markup_buy_percent, created_at) VALUES (9001, 1, 5, 'Linked Shop', '', 100, 50, datetime('now'))`); err != nil {
		t.Fatalf("seed shop: %v", err)
	}

	r := shopLinkTestRouter()

	// Create a shop item linked to compendium equipment 500.
	rec := shopPostJSON(t, r, "/api/shops/9001/items", map[string]any{
		"item_name":               "Something Else",
		"category":                "weapon",
		"price_gp":                42,
		"quantity_available":      3,
		"compendium_equipment_id": 500,
	})
	if rec.Code != http.StatusCreated && rec.Code != http.StatusOK {
		t.Fatalf("create linked shop item: got %d: %s", rec.Code, rec.Body.String())
	}

	// List items: name must be snapshotted from the compendium entry and the link kept.
	rec = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/shops/9001/items", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list items: got %d", rec.Code)
	}
	var items []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &items); err != nil {
		t.Fatalf("list items: bad json: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	it := items[0]
	if it["item_name"] != "Test Blade" {
		t.Fatalf("expected snapshotted name 'Test Blade', got %v", it["item_name"])
	}
	if it["compendium_equipment_id"] == nil || int64(it["compendium_equipment_id"].(float64)) != 500 {
		t.Fatalf("expected compendium_equipment_id 500, got %v", it["compendium_equipment_id"])
	}
	itemID := int64(it["id"].(float64))

	// Buy the linked item: inventory row must carry the compendium link + weight.
	rec = shopPostJSON(t, r, "/api/shops/9001/buy", map[string]any{
		"character_id": 7,
		"item_id":      itemID,
		"quantity":     1,
	})
	if rec.Code != http.StatusOK && rec.Code != http.StatusCreated {
		t.Fatalf("buy linked item: got %d: %s", rec.Code, rec.Body.String())
	}
	var name, category string
	var weight float64
	var compID *int64
	err := db.DB.QueryRow("SELECT name, category, weight, compendium_equipment_id FROM inventory WHERE character_id=7").Scan(&name, &category, &weight, &compID)
	if err != nil {
		t.Fatalf("read inventory: %v", err)
	}
	if name != "Test Blade" || category != "Adventuring Gear" || weight != 2.0 || compID == nil || *compID != 500 {
		t.Fatalf("inventory row mismatch: name=%s cat=%s weight=%v compID=%v", name, category, weight, compID)
	}

	// Unlink the shop item: reference removed, data preserved.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/shop-items/"+strconv.FormatInt(itemID, 10)+"/link", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("unlink shop item: got %d: %s", rec.Code, rec.Body.String())
	}
	var compID2 *int64
	if err := db.DB.QueryRow("SELECT compendium_equipment_id FROM shop_items WHERE id=?", itemID).Scan(&compID2); err != nil {
		t.Fatalf("read shop item: %v", err)
	}
	if compID2 != nil {
		t.Fatalf("expected unlinked shop item, got compendium_equipment_id=%v", *compID2)
	}
	var stillName string
	if err := db.DB.QueryRow("SELECT item_name FROM shop_items WHERE id=?", itemID).Scan(&stillName); err != nil {
		t.Fatalf("read shop item name: %v", err)
	}
	if stillName != "Test Blade" {
		t.Fatalf("expected preserved name after unlink, got %q", stillName)
	}
}

// ─── NPC item links ↔ compendium equipment ───

func npcLinkTestRouter() *gin.Engine {
	return testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/oneshot-adventures/:id/npc-item-links", CreateNPCItemLink)
		auth.GET("/oneshot-adventures/:id/npcs/:nid/items", ListItemsForNPC)
	})
}

func TestNPCItemLinkCompendiumXOR(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedOneShot(t, 9001, 1, "Test Adventure")
	testutil.SeedNPC(t, 1, "Zed", "Human", "Fighter")
	testutil.SeedCompendiumEquipment(t, 501, "NPC Relic")
	if _, err := db.DB.Exec(`INSERT OR IGNORE INTO oneshot_adventure_npcs(adventure_id, npc_id, role) VALUES(9001,1,'helper')`); err != nil {
		t.Fatalf("seed oneshot npc: %v", err)
	}

	r := npcLinkTestRouter()

	// Neither source → 400.
	rec := shopPostJSON(t, r, "/api/oneshot-adventures/9001/npc-item-links", map[string]any{"npc_id": 1})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 with no source, got %d: %s", rec.Code, rec.Body.String())
	}

	// Both sources → 400.
	rec = shopPostJSON(t, r, "/api/oneshot-adventures/9001/npc-item-links", map[string]any{"npc_id": 1, "item_id": 5, "compendium_equipment_id": 501})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 with both sources, got %d: %s", rec.Code, rec.Body.String())
	}

	// Compendium source → 201 and name resolution.
	rec = shopPostJSON(t, r, "/api/oneshot-adventures/9001/npc-item-links", map[string]any{"npc_id": 1, "compendium_equipment_id": 501})
	if rec.Code != http.StatusCreated && rec.Code != http.StatusOK {
		t.Fatalf("expected 201 with compendium source, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/oneshot-adventures/9001/npcs/1/items", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list npc items: got %d", rec.Code)
	}
	var links []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &links); err != nil {
		t.Fatalf("list npc items: bad json: %v body=%s", err, rec.Body.String())
	}
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	if links[0]["item_name"] != "NPC Relic" {
		t.Fatalf("expected resolved name 'NPC Relic', got %v", links[0]["item_name"])
	}
	if links[0]["compendium_equipment_id"] == nil || int64(links[0]["compendium_equipment_id"].(float64)) != 501 {
		t.Fatalf("expected compendium_equipment_id 501, got %v", links[0]["compendium_equipment_id"])
	}
}

// ─── Character race/class/background linking ───

func charLinkTestRouter() *gin.Engine {
	return testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/characters/:id/race/link", linkCharacterRace)
		auth.DELETE("/characters/:id/race/link", unlinkCharacterRace)
		auth.POST("/characters/:id/class/link", linkCharacterClass)
		auth.DELETE("/characters/:id/class/link", unlinkCharacterClass)
		auth.POST("/characters/:id/background/link", linkCharacterBackground)
		auth.DELETE("/characters/:id/background/link", unlinkCharacterBackground)
		auth.GET("/characters/:id", GetCharacter)
	})
}

func TestCharacterIdentityLinks(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 42, 1, "Identity Hero", "Human", "Fighter")

	if _, err := db.DB.Exec(`INSERT OR IGNORE INTO compendium_races(id, name, description) VALUES(701,'Zephyrkin','wind-touched folk')`); err != nil {
		t.Fatalf("seed compendium race: %v", err)
	}
	if _, err := db.DB.Exec(`INSERT OR IGNORE INTO compendium_classes(id, name, hit_die, primary_ability) VALUES(702,'Voidblade','d10','STR')`); err != nil {
		t.Fatalf("seed compendium class: %v", err)
	}
	if _, err := db.DB.Exec(`INSERT OR IGNORE INTO compendium_backgrounds(id, name, description) VALUES(703,'Moonsailor','lunar tides')`); err != nil {
		t.Fatalf("seed compendium background: %v", err)
	}

	r := charLinkTestRouter()

	link := func(path string, form, compID string) *httptest.ResponseRecorder {
		formData := url.Values{}
		formData.Set(form, compID)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.ServeHTTP(rec, req)
		return rec
	}

	// Race link copies the name and stores the reference.
	if rec := link("/api/characters/42/race/link", "compendium_race_id", "701"); rec.Code != http.StatusOK {
		t.Fatalf("link race: got %d: %s", rec.Code, rec.Body.String())
	}
	// Class link.
	if rec := link("/api/characters/42/class/link", "compendium_class_id", "702"); rec.Code != http.StatusOK {
		t.Fatalf("link class: got %d: %s", rec.Code, rec.Body.String())
	}
	// Background link.
	if rec := link("/api/characters/42/background/link", "compendium_background_id", "703"); rec.Code != http.StatusOK {
		t.Fatalf("link background: got %d: %s", rec.Code, rec.Body.String())
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/characters/42", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("get character: got %d: %s", rec.Code, rec.Body.String())
	}
	var ch map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &ch); err != nil {
		t.Fatalf("get character: bad json: %v", err)
	}
	if ch["race"] != "Zephyrkin" || ch["class"] != "Voidblade" || ch["background"] != "Moonsailor" {
		t.Fatalf("expected snapshotted names, got race=%v class=%v background=%v", ch["race"], ch["class"], ch["background"])
	}
	if int64(ch["compendium_race_id"].(float64)) != 701 || int64(ch["compendium_class_id"].(float64)) != 702 || int64(ch["compendium_background_id"].(float64)) != 703 {
		t.Fatalf("expected link ids, got race=%v class=%v bg=%v", ch["compendium_race_id"], ch["compendium_class_id"], ch["compendium_background_id"])
	}

	// Unlink preserves the text but drops the reference.
	unlink := func(path string) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, path, nil)
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("unlink %s: got %d: %s", path, rec.Code, rec.Body.String())
		}
	}
	unlink("/api/characters/42/race/link")
	unlink("/api/characters/42/class/link")
	unlink("/api/characters/42/background/link")

	var race, class, background string
	var raceID, classID, bgID *int64
	if err := db.DB.QueryRow("SELECT race, class, background, compendium_race_id, compendium_class_id, compendium_background_id FROM characters WHERE id=42").
		Scan(&race, &class, &background, &raceID, &classID, &bgID); err != nil {
		t.Fatalf("read character: %v", err)
	}
	if race != "Zephyrkin" || class != "Voidblade" || background != "Moonsailor" {
		t.Fatalf("expected preserved text after unlink, got race=%q class=%q bg=%q", race, class, background)
	}
	if raceID != nil || classID != nil || bgID != nil {
		t.Fatalf("expected nulled link ids, got race=%v class=%v bg=%v", raceID, classID, bgID)
	}

	// Missing compendium id → 404.
	if rec := link("/api/characters/42/race/link", "compendium_race_id", "999999"); rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown compendium race, got %d", rec.Code)
	}
}
