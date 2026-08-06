package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers/testutil"
)

func shopTestRouter() *gin.Engine {
	return testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/shops", ListShops)
		auth.GET("/shops/:id/items", ListShopItems)
		auth.POST("/shops/:id/buy", BuyItem)
		auth.POST("/shops/:id/sell", SellItem)
		auth.PUT("/shop-items/:id", UpdateShopItem)
		auth.POST("/campaigns/:id/shops", CreateShop)
	})
}

func shopPostJSON(t *testing.T, r *gin.Engine, path string, body map[string]any) *httptest.ResponseRecorder {
	t.Helper()
	b, _ := json.Marshal(body)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	return rec
}

func TestShopCRUDWithLocation(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 5, "Shop Camp", "Party", 1)

	// location for the shop
	if _, err := db.DB.Exec(`INSERT INTO locations (id, user_id, name, type, description, created_at) VALUES (9001, 1, 'Market Square', 'town', 'busy market', datetime('now'))`); err != nil {
		t.Fatalf("seed location: %v", err)
	}

	r := shopTestRouter()
	rec := shopPostJSON(t, r, "/api/campaigns/5/shops", map[string]any{
		"name":               "Smoke & Steel",
		"description":        "Smithy",
		"location_id":        9001,
		"markup_percent":     120,
		"markup_buy_percent": 50,
	})
	if rec.Code != http.StatusCreated && rec.Code != http.StatusOK {
		t.Fatalf("create shop: expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/shops", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list shops: expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Smoke") || !strings.Contains(body, "Market Square") {
		t.Fatalf("list shops: shop missing: %s", body)
	}
	if !strings.Contains(body, "Market Square") {
		t.Fatalf("list shops: location_name missing: %s", body)
	}
}

func TestShopItemUpdate(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 5, "Shop Camp", "Party", 1)

	if _, err := db.DB.Exec(`INSERT INTO shops (id, user_id, campaign_id, name, description, markup_percent, markup_buy_percent, created_at) VALUES (9001, 1, 5, 'Test Shop', '', 100, 50, datetime('now'))`); err != nil {
		t.Fatalf("seed shop: %v", err)
	}
	if _, err := db.DB.Exec(`INSERT INTO shop_items (id, shop_id, item_name, category, price_gp, quantity_available, is_magical, attunement_required, notes) VALUES (9001, 9001, 'Iron Sword', 'weapon', 10, 5, 0, 0, '')`); err != nil {
		t.Fatalf("seed item: %v", err)
	}

	r := shopTestRouter()
	body, _ := json.Marshal(map[string]any{"price_gp": 25, "item_name": "Iron Sword", "category": "weapon", "quantity_available": 3})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/shop-items/9001", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("update item: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	// UpdateShopItem echoes {"ok":true} — verify the persisted price via the items list.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/shops/9001/items", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("items: expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "25") {
		t.Fatalf("update item: price not reflected: %s", rec.Body.String())
	}
}

func TestShopMarkupBuySell(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 5, "Shop Camp", "Party", 1)
	testutil.SeedCharacter(t, 7, 1, "Buyer", "Human", "Fighter")

	// character gp: update the seeded character to have gp
	if _, err := db.DB.Exec(`INSERT OR REPLACE INTO character_currency (character_id, gp) VALUES (7, 1000)`); err != nil {
		t.Fatalf("set gp: %v", err)
	}
	if _, err := db.DB.Exec(`INSERT INTO shops (id, user_id, campaign_id, name, description, markup_percent, markup_buy_percent, created_at) VALUES (9001, 1, 5, 'Markup Shop', '', 150, 50, datetime('now'))`); err != nil {
		t.Fatalf("seed shop: %v", err)
	}
	if _, err := db.DB.Exec(`INSERT INTO shop_items (id, shop_id, item_name, category, price_gp, quantity_available, is_magical, attunement_required, notes) VALUES (9001, 9001, 'Potion', 'consumable', 10, 5, 0, 0, '')`); err != nil {
		t.Fatalf("seed item: %v", err)
	}

	r := shopTestRouter()

	// buy: price 10 * 150/100 = 15
	rec := shopPostJSON(t, r, "/api/shops/9001/buy", map[string]any{
		"character_id": 7,
		"item_id":      9001,
		"quantity":     1,
	})
	if rec.Code != http.StatusOK && rec.Code != http.StatusCreated {
		t.Fatalf("buy: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var buyRes struct {
		Price int `json:"price"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &buyRes); err != nil {
		t.Fatalf("buy: bad json: %v body=%s", err, rec.Body.String())
	}
	if buyRes.Price != 15 {
		t.Fatalf("buy: expected price 15 (10 * 150/100), got %d", buyRes.Price)
	}

	// sell: payout 10 * 50/100 = 5
	rec = shopPostJSON(t, r, "/api/shops/9001/sell", map[string]any{
		"character_id": 7,
		"item_name":    "Potion",
		"quantity":     1,
	})
	if rec.Code != http.StatusOK && rec.Code != http.StatusCreated {
		t.Fatalf("sell: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "5") {
		t.Fatalf("sell: payout not reflected: %s", rec.Body.String())
	}
}

// keep fmt import used (debug helper)
var _ = fmt.Sprintf
