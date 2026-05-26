package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
)

type OneShotShop struct {
	ID             int64   `json:"id"`
	UserID         int64   `json:"user_id"`
	CampaignID     *int64  `json:"campaign_id,omitempty"`
	OneshotID      *int64  `json:"oneshot_adventure_id,omitempty"`
	Name           string  `json:"name"`
	Description    string  `json:"description"`
	MarkupPercent  float64 `json:"markup_percent"`
	MarkupBuyPercent float64 `json:"markup_buy_percent"`
	CreatedAt      string  `json:"created_at"`
}

func ListOneShotShops(c *gin.Context) {
	adventureID := c.Param("id")
	rows, err := db.DB.Query("SELECT id, user_id, campaign_id, oneshot_adventure_id, name, description, markup_percent, markup_buy_percent, created_at FROM shops WHERE oneshot_adventure_id=? ORDER BY name", adventureID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	shops := make([]OneShotShop, 0)
	for rows.Next() {
		var s OneShotShop
		rows.Scan(&s.ID, &s.UserID, &s.CampaignID, &s.OneshotID, &s.Name, &s.Description, &s.MarkupPercent, &s.MarkupBuyPercent, &s.CreatedAt)
		shops = append(shops, s)
	}
	c.JSON(http.StatusOK, shops)
}

func CreateOneShotShop(c *gin.Context) {
	adventureID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := c.Get("user_id")
	var s OneShotShop
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := db.DB.Exec("INSERT INTO shops(user_id, oneshot_adventure_id, name, description, markup_percent, markup_buy_percent) VALUES(?,?,?,?,?,?)",
		userID, adventureID, s.Name, s.Description, s.MarkupPercent, s.MarkupBuyPercent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	db.DB.QueryRow("SELECT id, user_id, campaign_id, oneshot_adventure_id, name, description, markup_percent, markup_buy_percent, created_at FROM shops WHERE id=?", id).Scan(
		&s.ID, &s.UserID, &s.CampaignID, &s.OneshotID, &s.Name, &s.Description, &s.MarkupPercent, &s.MarkupBuyPercent, &s.CreatedAt)
	c.JSON(http.StatusCreated, s)
}

func CreateOneShotShopItem(c *gin.Context) {
	shopID, _ := strconv.ParseInt(c.Param("shopId"), 10, 64)
	var it struct {
		ItemName          string  `json:"item_name"`
		Category          string  `json:"category"`
		PriceGP           float64 `json:"price_gp"`
		QuantityAvailable int     `json:"quantity_available"`
		Description       string  `json:"description"`
		IsMagical         bool    `json:"is_magical"`
		AttunementRequired bool   `json:"attunement_required"`
		Notes             string  `json:"notes"`
	}
	if err := c.ShouldBindJSON(&it); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	isMag := 0
	if it.IsMagical { isMag = 1 }
	att := 0
	if it.AttunementRequired { att = 1 }
	_, err := db.DB.Exec("INSERT INTO shop_items(shop_id,item_name,category,price_gp,quantity_available,description,is_magical,attunement_required,notes) VALUES(?,?,?,?,?,?,?,?,?)",
		shopID, it.ItemName, it.Category, it.PriceGP, it.QuantityAvailable, it.Description, isMag, att, it.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"ok": true})
}

func DeleteOneShotShop(c *gin.Context) {
	shopID, _ := strconv.ParseInt(c.Param("shopId"), 10, 64)
	db.DB.Exec("DELETE FROM shops WHERE id=?", shopID)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
