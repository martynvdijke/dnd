package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
)

type Shop struct {
	ID             int64   `json:"id"`
	UserID         int64   `json:"user_id"`
	CampaignID     *int64  `json:"campaign_id,omitempty"`
	Name           string  `json:"name"`
	Description    string  `json:"description"`
	MarkupPercent  float64 `json:"markup_percent"`
	MarkupBuyPercent float64 `json:"markup_buy_percent"`
	CreatedAt      string  `json:"created_at"`
}

type ShopItem struct {
	ID                int64   `json:"id"`
	ShopID            int64   `json:"shop_id"`
	ItemName          string  `json:"item_name"`
	Category          string  `json:"category"`
	PriceGP           float64 `json:"price_gp"`
	QuantityAvailable int     `json:"quantity_available"`
	Description       string  `json:"description"`
	IsMagical         bool    `json:"is_magical"`
	AttunementRequired bool   `json:"attunement_required"`
	Notes             string  `json:"notes"`
}

type ShopTransaction struct {
	ID              int64   `json:"id"`
	ShopID          int64   `json:"shop_id"`
	CharacterID     int64   `json:"character_id"`
	ItemName        string  `json:"item_name"`
	Quantity        int     `json:"quantity"`
	PriceGP         float64 `json:"price_gp"`
	TransactionType string  `json:"transaction_type"`
	Timestamp       string  `json:"timestamp"`
}

func ListShops(c *gin.Context) {
	campaignID := c.Query("campaign_id")
	var rows *sql.Rows
	var err error
	if campaignID != "" {
		rows, err = db.DB.Query("SELECT id,user_id,campaign_id,name,description,markup_percent,markup_buy_percent,created_at FROM shops WHERE campaign_id=? ORDER BY name", campaignID)
	} else {
		rows, err = db.DB.Query("SELECT id,user_id,campaign_id,name,description,markup_percent,markup_buy_percent,created_at FROM shops ORDER BY name")
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	shops := make([]Shop, 0)
	for rows.Next() {
		var s Shop
		rows.Scan(&s.ID, &s.UserID, &s.CampaignID, &s.Name, &s.Description, &s.MarkupPercent, &s.MarkupBuyPercent, &s.CreatedAt)
		shops = append(shops, s)
	}
	c.JSON(http.StatusOK, shops)
}

func CreateShop(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var s Shop
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := db.DB.Exec("INSERT INTO shops(user_id,campaign_id,name,description,markup_percent,markup_buy_percent) VALUES(?,?,?,?,?,?)",
		userID, s.CampaignID, s.Name, s.Description, s.MarkupPercent, s.MarkupBuyPercent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"ok": true})
}

func UpdateShop(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var s Shop
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Exec("UPDATE shops SET name=?,description=?,markup_percent=?,markup_buy_percent=? WHERE id=?", s.Name, s.Description, s.MarkupPercent, s.MarkupBuyPercent, id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteShop(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	db.DB.Exec("DELETE FROM shops WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func ListShopItems(c *gin.Context) {
	shopID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	rows, err := db.DB.Query("SELECT id,shop_id,item_name,category,price_gp,quantity_available,description,is_magical,attunement_required,notes FROM shop_items WHERE shop_id=? ORDER BY item_name", shopID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	type SI struct {
		ID                int64   `json:"id"`
		ShopID            int64   `json:"shop_id"`
		ItemName          string  `json:"item_name"`
		Category          string  `json:"category"`
		PriceGP           float64 `json:"price_gp"`
		QuantityAvailable int     `json:"quantity_available"`
		Description       string  `json:"description"`
		IsMagical         bool    `json:"is_magical"`
		AttunementRequired bool   `json:"attunement_required"`
		Notes             string  `json:"notes"`
	}
	var items []SI
	for rows.Next() {
		var it SI
		var isMag, att int
		rows.Scan(&it.ID, &it.ShopID, &it.ItemName, &it.Category, &it.PriceGP, &it.QuantityAvailable, &it.Description, &isMag, &att, &it.Notes)
		it.IsMagical = isMag == 1
		it.AttunementRequired = att == 1
		items = append(items, it)
	}
	c.JSON(http.StatusOK, items)
}

func CreateShopItem(c *gin.Context) {
	shopID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var it ShopItem
	if err := c.ShouldBindJSON(&it); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	isMag := 0
	if it.IsMagical {
		isMag = 1
	}
	att := 0
	if it.AttunementRequired {
		att = 1
	}
	_, err := db.DB.Exec("INSERT INTO shop_items(shop_id,item_name,category,price_gp,quantity_available,description,is_magical,attunement_required,notes) VALUES(?,?,?,?,?,?,?,?,?)",
		shopID, it.ItemName, it.Category, it.PriceGP, it.QuantityAvailable, it.Description, isMag, att, it.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"ok": true})
}

func DeleteShopItem(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	db.DB.Exec("DELETE FROM shop_items WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func BuyItem(c *gin.Context) {
	shopID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		ItemID    int64 `json:"item_id"`
		CharacterID int64 `json:"character_id"`
		Quantity  int   `json:"quantity"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Quantity <= 0 {
		req.Quantity = 1
	}

	var itemName string
	var priceGP float64
	var qtyAvail int
	err := db.DB.QueryRow("SELECT item_name, price_gp, quantity_available FROM shop_items WHERE id=? AND shop_id=?", req.ItemID, shopID).Scan(&itemName, &priceGP, &qtyAvail)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}

	// Check quantity
	if qtyAvail >= 0 && qtyAvail < req.Quantity {
		c.JSON(http.StatusBadRequest, gin.H{"error": "not enough stock"})
		return
	}

	totalPrice := priceGP * float64(req.Quantity)

	// Deduct character gold
	var currentGP float64
	db.DB.QueryRow("SELECT gp FROM character_currency WHERE character_id=?", req.CharacterID).Scan(&currentGP)
	if currentGP < totalPrice {
		c.JSON(http.StatusBadRequest, gin.H{"error": "not enough gold"})
		return
	}

	// Process transaction
	tx, _ := db.DB.Begin()
	tx.Exec("UPDATE character_currency SET gp = gp - ? WHERE character_id=?", totalPrice, req.CharacterID)
	tx.Exec("INSERT INTO inventory(character_id,name,quantity,category) VALUES(?,?,?,'gear')", req.CharacterID, itemName, req.Quantity)
	if qtyAvail >= 0 {
		tx.Exec("UPDATE shop_items SET quantity_available = quantity_available - ? WHERE id=?", req.Quantity, req.ItemID)
	}
	tx.Exec("INSERT INTO shop_transactions(shop_id,character_id,item_name,quantity,price_gp,transaction_type) VALUES(?,?,?,?,?,'buy')", shopID, req.CharacterID, itemName, req.Quantity, totalPrice)
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"ok": true, "item": itemName, "quantity": req.Quantity, "price": totalPrice})
}

func SellItem(c *gin.Context) {
	shopID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		CharacterID int64  `json:"character_id"`
		ItemName    string `json:"item_name"`
		Quantity    int    `json:"quantity"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Quantity <= 0 {
		req.Quantity = 1
	}

	var shopMarkupBuy float64
	db.DB.QueryRow("SELECT markup_buy_percent FROM shops WHERE id=?", shopID).Scan(&shopMarkupBuy)
	var basePrice float64
	db.DB.QueryRow("SELECT price_gp FROM shop_items WHERE shop_id=? AND item_name=? LIMIT 1", shopID, req.ItemName).Scan(&basePrice)

	sellPrice := basePrice * (shopMarkupBuy / 100.0) * float64(req.Quantity)

	tx, _ := db.DB.Begin()
	tx.Exec("UPDATE character_currency SET gp = gp + ? WHERE character_id=?", sellPrice, req.CharacterID)
	tx.Exec("DELETE FROM inventory WHERE character_id=? AND name=? LIMIT ?", req.CharacterID, req.ItemName, req.Quantity)
	tx.Exec("INSERT INTO shop_transactions(shop_id,character_id,item_name,quantity,price_gp,transaction_type) VALUES(?,?,?,?,?,'sell')", shopID, req.CharacterID, req.ItemName, req.Quantity, sellPrice)
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"ok": true, "item": req.ItemName, "quantity": req.Quantity, "price": sellPrice})
}

func ListShopTransactions(c *gin.Context) {
	charID := c.Query("character_id")
	var rows *sql.Rows
	var err error
	if charID != "" {
		rows, err = db.DB.Query("SELECT id,shop_id,character_id,item_name,quantity,price_gp,transaction_type,timestamp FROM shop_transactions WHERE character_id=? ORDER BY timestamp DESC LIMIT 20", charID)
	} else {
		rows, err = db.DB.Query("SELECT id,shop_id,character_id,item_name,quantity,price_gp,transaction_type,timestamp FROM shop_transactions ORDER BY timestamp DESC LIMIT 20")
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var txns []ShopTransaction
	for rows.Next() {
		var t ShopTransaction
		rows.Scan(&t.ID, &t.ShopID, &t.CharacterID, &t.ItemName, &t.Quantity, &t.PriceGP, &t.TransactionType, &t.Timestamp)
		txns = append(txns, t)
	}
	c.JSON(http.StatusOK, txns)
}
