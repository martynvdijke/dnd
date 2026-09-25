package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers/testutil"
)

func TestHtmxNPCEdit(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)

	testutil.SeedUser(t, 1, "dm", "admin")
	testutil.SeedCharacter(t, 1, 1, "Hero", "Human", "Fighter")
	res, err := db.DB.Exec("INSERT INTO npcs(user_id,name,race,class,description,notes) VALUES(?,?,?,?,?,?)", 1, "Old NPC", "Elf", "Ranger", "desc", "notes")
	if err != nil {
		t.Fatalf("seed npc: %v", err)
	}
	npcID, _ := res.LastInsertId()
	if _, err := db.DB.Exec("INSERT INTO character_npcs(character_id,npc_id,relationship) VALUES(?,?,?)", 1, npcID, "ally"); err != nil {
		t.Fatalf("seed link: %v", err)
	}

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/htmx/npcs/:id/edit", HtmxEditNPCForm)
		auth.PUT("/htmx/npcs/:id", HtmxUpdateNPC)
	})

	w := testutil.Get(t, r, fmt.Sprintf("/api/htmx/npcs/%d/edit?character_id=1", npcID))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `value="Old NPC"`) {
		t.Fatalf("edit form did not prefill: code=%d body=%s", w.Code, w.Body.String())
	}

	w = putForm(t, r, fmt.Sprintf("/api/htmx/npcs/%d", npcID), map[string]string{
		"character_id": "1", "name": "New NPC", "race": "Orc", "class": "Barbarian",
		"description": "d", "notes": "n", "is_full": "0", "type": "enemy",
	})
	testutil.AssertStatus(t, w, http.StatusOK)

	var name, rel string
	if err := db.DB.QueryRow("SELECT name FROM npcs WHERE id=?", npcID).Scan(&name); err != nil {
		t.Fatalf("read npc: %v", err)
	}
	if name != "New NPC" {
		t.Errorf("npc name not updated: %q", name)
	}
	if err := db.DB.QueryRow("SELECT relationship FROM character_npcs WHERE character_id=1 AND npc_id=?", npcID).Scan(&rel); err != nil {
		t.Fatalf("read link: %v", err)
	}
	if rel != "enemy" {
		t.Errorf("relationship not updated: %q", rel)
	}
}

func TestUpdateCraftingRecipe(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)

	testutil.SeedUser(t, 1, "dm", "admin")
	testutil.SeedUser(t, 2, "bob", "user")

	routes := func(auth *gin.RouterGroup) {
		auth.POST("/crafting/recipes", CreateCraftingRecipe)
		auth.PUT("/crafting/recipes/:id", UpdateCraftingRecipe)
	}
	r := testutil.NewRouter(routes)
	w := testutil.PostJSON(t, r, "/api/crafting/recipes", map[string]any{
		"name": "Old Recipe", "category": "potion", "difficulty_dc": 10, "crafting_time_hours": 2,
	})
	testutil.AssertStatus(t, w, http.StatusCreated)

	var id int64
	if err := db.DB.QueryRow("SELECT id FROM crafting_recipes WHERE name='Old Recipe'").Scan(&id); err != nil {
		t.Fatalf("read recipe id: %v", err)
	}

	w = testutil.PutJSON(t, r, fmt.Sprintf("/api/crafting/recipes/%d", id), map[string]any{
		"name": "New Recipe", "category": "scroll", "difficulty_dc": 15, "crafting_time_hours": 4,
	})
	testutil.AssertStatus(t, w, http.StatusOK)

	var name, category string
	var dc int
	if err := db.DB.QueryRow("SELECT name,category,difficulty_dc FROM crafting_recipes WHERE id=?", id).Scan(&name, &category, &dc); err != nil {
		t.Fatalf("read recipe: %v", err)
	}
	if name != "New Recipe" || category != "scroll" || dc != 15 {
		t.Errorf("recipe not updated: name=%q category=%q dc=%d", name, category, dc)
	}

	other := testutil.NewRouterWithUser(routes, 2, "user")
	w = testutil.PutJSON(t, other, fmt.Sprintf("/api/crafting/recipes/%d", id), map[string]any{"name": "Hijacked"})
	testutil.AssertStatus(t, w, http.StatusForbidden)
}

func TestUpdateActNPC(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)

	testutil.SeedUser(t, 1, "dm", "admin")
	testutil.SeedOneShot(t, 1, 1, "Adventure")
	testutil.SeedOneShotAct(t, 1, 1, "Act One", 1)

	routes := func(auth *gin.RouterGroup) {
		auth.POST("/oneshot-acts/:id/npcs", CreateActNPC)
		auth.PUT("/oneshot-acts/:id/npcs/:nid", UpdateActNPC)
	}
	r := testutil.NewRouter(routes)
	w := testutil.PostJSON(t, r, "/api/oneshot-acts/1/npcs", map[string]any{"name": "Goblin", "role": "boss", "notes": "x"})
	testutil.AssertStatus(t, w, http.StatusCreated)
	var created struct {
		ID int64 `json:"id"`
	}
	testutil.ParseJSON(t, w, &created)

	w = testutil.PutJSON(t, r, fmt.Sprintf("/api/oneshot-acts/1/npcs/%d", created.ID), map[string]any{"name": "Goblin Chief", "role": "boss", "notes": "y"})
	testutil.AssertStatus(t, w, http.StatusOK)

	var name string
	if err := db.DB.QueryRow("SELECT name FROM oneshot_act_npcs WHERE id=?", created.ID).Scan(&name); err != nil {
		t.Fatalf("read act npc: %v", err)
	}
	if name != "Goblin Chief" {
		t.Errorf("act npc not updated: %q", name)
	}

	// act that does not exist -> update matches no row
	w = testutil.PutJSON(t, r, fmt.Sprintf("/api/oneshot-acts/999/npcs/%d", created.ID), map[string]any{"name": "Nope"})
	testutil.AssertStatus(t, w, http.StatusNotFound)
}
