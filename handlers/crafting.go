package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
)

type CraftingRecipe struct {
	ID               int64   `json:"id"`
	UserID           *int64  `json:"user_id"`
	Name             string  `json:"name"`
	Description      string  `json:"description"`
	Category         string  `json:"category"`
	DifficultyDC     int     `json:"difficulty_dc"`
	CraftingTimeHours float64 `json:"crafting_time_hours"`
	RequiredTools    string  `json:"required_tools"`
	RequiredMaterials string `json:"required_materials"`
	ResultItemName   string  `json:"result_item_name"`
	ResultItemCategory string `json:"result_item_category"`
	ResultQuantity   int     `json:"result_quantity"`
	ResultDescription string `json:"result_description"`
	Notes            string  `json:"notes"`
	CreatedAt        string  `json:"created_at"`
}

type CharacterCrafting struct {
	ID                  int64   `json:"id"`
	CharacterID         int64   `json:"character_id"`
	RecipeID            *int64  `json:"recipe_id,omitempty"`
	Name                string  `json:"name"`
	ProgressHours       float64 `json:"progress_hours"`
	TotalHoursRequired  float64 `json:"total_hours_required"`
	DC                  int     `json:"dc"`
	Status              string  `json:"status"`
	MaterialsAllocated  string  `json:"materials_allocated"`
	Notes               string  `json:"notes"`
	StartedAt           string  `json:"started_at"`
}

func scanRecipes(rows *sql.Rows, err error) ([]CraftingRecipe, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var recipes = make([]CraftingRecipe, 0)
	for rows.Next() {
		var r CraftingRecipe
		var uid sql.NullInt64
		rows.Scan(&r.ID, &uid, &r.Name, &r.Description, &r.Category,
			&r.DifficultyDC, &r.CraftingTimeHours, &r.RequiredTools, &r.RequiredMaterials,
			&r.ResultItemName, &r.ResultItemCategory, &r.ResultQuantity, &r.ResultDescription, &r.Notes, &r.CreatedAt)
		if uid.Valid {
			r.UserID = &uid.Int64
		}
		recipes = append(recipes, r)
	}
	return recipes, nil
}

func scanCharCrafting(rows *sql.Rows, err error) ([]CharacterCrafting, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var projects = make([]CharacterCrafting, 0)
	for rows.Next() {
		var p CharacterCrafting
		rows.Scan(&p.ID, &p.CharacterID, &p.RecipeID, &p.Name, &p.ProgressHours,
			&p.TotalHoursRequired, &p.DC, &p.Status, &p.MaterialsAllocated, &p.Notes, &p.StartedAt)
		projects = append(projects, p)
	}
	return projects, nil
}

func ListCraftingRecipes(c *gin.Context) {
	userID, _ := c.Get("user_id")
	rows, err := db.DB.Query("SELECT id,user_id,name,description,category,difficulty_dc,crafting_time_hours,required_tools,required_materials,result_item_name,result_item_category,result_quantity,result_description,notes,created_at FROM crafting_recipes WHERE user_id IS NULL OR user_id=? ORDER BY name", userID)
	recipes, err := scanRecipes(rows, err)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, recipes)
}

func CreateCraftingRecipe(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var r CraftingRecipe
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := db.DB.Exec(`INSERT INTO crafting_recipes(user_id,name,description,category,difficulty_dc,crafting_time_hours,required_tools,required_materials,result_item_name,result_item_category,result_quantity,result_description,notes) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`,
		userID, r.Name, r.Description, r.Category, r.DifficultyDC, r.CraftingTimeHours,
		r.RequiredTools, r.RequiredMaterials, r.ResultItemName, r.ResultItemCategory,
		r.ResultQuantity, r.ResultDescription, r.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"ok": true})
}

func DeleteCraftingRecipe(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	db.DB.Exec("DELETE FROM crafting_recipes WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func ListCharacterCrafting(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	rows, err := db.DB.Query("SELECT id,character_id,recipe_id,name,progress_hours,total_hours_required,dc,status,materials_allocated,notes,started_at FROM character_crafting WHERE character_id=? ORDER BY started_at DESC", charID)
	projects, err := scanCharCrafting(rows, err)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, projects)
}

func CreateCharacterCrafting(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		RecipeID           *int64  `json:"recipe_id"`
		Name               string  `json:"name"`
		TotalHoursRequired float64 `json:"total_hours_required"`
		DC                 int     `json:"dc"`
		MaterialsAllocated string  `json:"materials_allocated"`
		Notes              string  `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := db.DB.Exec(`INSERT INTO character_crafting(character_id,recipe_id,name,progress_hours,total_hours_required,dc,status,materials_allocated,notes) VALUES(?,?,?,0,?,?,'in-progress',?,?)`,
		charID, req.RecipeID, req.Name, req.TotalHoursRequired, req.DC, req.MaterialsAllocated, req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"ok": true})
}

func UpdateCharacterCrafting(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		ProgressHours float64 `json:"progress_hours"`
		Status        string  `json:"status"`
		Notes         string  `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Status == "complete" {
		// Fetch the project to get result info
		var recipeID sql.NullInt64
		var name, materials string
		var totalHours float64
		db.DB.QueryRow("SELECT recipe_id,name,total_hours_required,materials_allocated FROM character_crafting WHERE id=?", id).Scan(&recipeID, &name, &totalHours, &materials)

		// Add result item to character inventory if we have a recipe
		if recipeID.Valid {
			var itemName, category string
			var qty int
			err := db.DB.QueryRow("SELECT result_item_name, result_item_category, result_quantity FROM crafting_recipes WHERE id=?", recipeID.Int64).Scan(&itemName, &category, &qty)
			if err == nil && itemName != "" {
				db.DB.Exec("INSERT INTO inventory(character_id,name,quantity,category) VALUES((SELECT character_id FROM character_crafting WHERE id=?),?,?,?)", id, itemName, qty, category)
			}
		}

		db.DB.Exec("UPDATE character_crafting SET status='complete', progress_hours=total_hours_required WHERE id=?", id)
	} else if req.Status == "abandoned" {
		db.DB.Exec("UPDATE character_crafting SET status='abandoned' WHERE id=?", id)
	} else {
		db.DB.Exec("UPDATE character_crafting SET progress_hours=?, notes=? WHERE id=?", req.ProgressHours, req.Notes, id)
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteCharacterCrafting(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	db.DB.Exec("DELETE FROM character_crafting WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func SeedCraftingRecipes() {
	var count int
	db.DB.QueryRow("SELECT COUNT(*) FROM crafting_recipes").Scan(&count)
	if count > 0 {
		return
	}

	// system recipes - use nil user_id
	recipes := []struct {
		name, desc, category string
		dc                   int
		hours                float64
		tools, materials     string
		resultName, resultCat string
		resultQty            int
		resultDesc           string
	}{
		{"Potion of Healing", "Restores 2d4+2 hit points", "potion", 10, 1, `["alchemist's supplies"]`, `[{"name":"Healing Herb Bundle","quantity":1,"consumed":true},{"name":"Glass Vial","quantity":1,"consumed":true}]`, "Potion of Healing", "potion", 1, "2d4+2 HP"},
		{"Potion of Greater Healing", "Restores 4d4+4 hit points", "potion", 12, 4, `["alchemist's supplies"]`, `[{"name":"Potion of Healing","quantity":1,"consumed":true},{"name":"Greater Healing Herb Bundle","quantity":2,"consumed":true}]`, "Potion of Greater Healing", "potion", 1, "4d4+4 HP"},
		{"Potion of Invisibility", "Become invisible for 1 hour", "potion", 15, 8, `["alchemist's supplies"]`, `[{"name":"Moonflower Petals","quantity":3,"consumed":true},{"name":"Essence of Shadow","quantity":1,"consumed":true},{"name":"Glass Vial","quantity":1,"consumed":true}]`, "Potion of Invisibility", "potion", 1, "Invisibility 1 hour"},
		{"Scroll of Fireball", "Cast Fireball (3rd level)", "scroll", 14, 8, `["calligrapher's supplies"]`, `[{"name":"Fine Parchment","quantity":1,"consumed":true},{"name":"Phoenix Ash Ink","quantity":1,"consumed":true}]`, "Scroll of Fireball", "scroll", 1, "Fireball 3rd level"},
		{"Scroll of Bless", "Cast Bless (1st level)", "scroll", 10, 4, `["calligrapher's supplies"]`, `[{"name":"Parchment","quantity":1,"consumed":true},{"name":"Holy Water Ink","quantity":1,"consumed":true}]`, "Scroll of Bless", "scroll", 1, "Bless 1st level"},
		{"Antitoxin", "Grants advantage on saves against poison", "potion", 10, 2, `["alchemist's supplies"]`, `[{"name":"Snake Venom","quantity":1,"consumed":true},{"name":"Purifying Herb","quantity":2,"consumed":true},{"name":"Glass Vial","quantity":1,"consumed":true}]`, "Antitoxin", "potion", 1, "Advantage on poison saves"},
	}

	for _, r := range recipes {
		_, err := db.DB.Exec(`INSERT INTO crafting_recipes(user_id,name,description,category,difficulty_dc,crafting_time_hours,required_tools,required_materials,result_item_name,result_item_category,result_quantity,result_description,notes,created_at) VALUES(NULL,?,?,?,?,?,?,?,?,?,?,?,'',datetime('now'))`,
			r.name, r.desc, r.category, r.dc, r.hours, r.tools, r.materials, r.resultName, r.resultCat, r.resultQty, r.resultDesc)
		if err != nil {
			log.Printf("Failed to seed recipe %s: %v", r.name, err)
		}
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	logger.Printf("Seeded %d crafting recipes", len(recipes))
}
