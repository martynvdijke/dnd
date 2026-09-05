package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"
	"villum/handlers/testutil"
)

func TestListLocations(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/locations", ListLocations)
	})
	w := testutil.Get(t, r, "/api/locations")
	testutil.AssertStatus(t, w, 200)
}

func TestUpdateLocation(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/locations", CreateLocation)
		auth.PUT("/locations/:id", UpdateLocation)
	})
	w := testutil.PostJSON(t, r, "/api/locations", map[string]any{"name": "Tavern", "type": "building"})
	testutil.AssertStatus(t, w, 201)
	var data map[string]any
	testutil.ParseJSON(t, w, &data)
	id := int64(data["id"].(float64))
	t.Run("success", func(t *testing.T) {
		w := testutil.PutJSON(t, r, "/api/locations/"+itoa64(id), map[string]any{"name": "Tavern2", "type": "building"})
		testutil.AssertStatus(t, w, 200)
	})
	t.Run("not found", func(t *testing.T) {
		w := testutil.PutJSON(t, r, "/api/locations/99999", map[string]any{"name": "x"})
		if w.Code != 404 {
			t.Fatalf("want 404 got %d", w.Code)
		}
	})
	t.Run("bad json", func(t *testing.T) {
		w := testutil.PutJSON(t, r, "/api/locations/"+itoa64(id), "bad")
		if w.Code != 400 {
			t.Fatalf("want 400 got %d", w.Code)
		}
	})
}

func TestSearchLocations(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/locations/search", SearchLocations)
	})
	w := testutil.Get(t, r, "/api/locations/search?q=Tavern")
	testutil.AssertStatus(t, w, 200)
	w = testutil.Get(t, r, "/api/locations/search")
	testutil.AssertStatus(t, w, 200)
}

func TestUnlinkNPC(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "Hero", "Human", "Fighter")
	// create location link via ent
	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/characters/:id/npcs/link", LinkNPC)
		auth.DELETE("/characters/npcs/:nid", UnlinkNPC)
		auth.GET("/characters/:id/npcs", GetCharacterNPCs)
	})
	// seed NPC
	testutil.SeedNPC(t, 20, "LinkNPC", "Elf", "Ranger")
	w := testutil.PostJSON(t, r, "/api/characters/1/npcs/link", map[string]any{"npc_id": 20, "relationship": "ally"})
	testutil.AssertStatus(t, w, 200)
	// get list to find link id
	w = testutil.Get(t, r, "/api/characters/1/npcs")
	testutil.AssertStatus(t, w, 200)
	// unlink not found
	w = testutil.Delete(t, r, "/api/characters/npcs/99999")
	if w.Code != 404 {
		t.Fatalf("want 404 got %d", w.Code)
	}
}

func TestListQuests(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "Hero", "Human", "Fighter")
	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/characters/:id/quests", ListQuests)
	})
	w := testutil.Get(t, r, "/api/characters/1/quests")
	testutil.AssertStatus(t, w, 200)
}

func TestListAllCharacters(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "Hero", "Human", "Fighter")
	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/characters/all", ListAllCharacters)
	})
	w := testutil.Get(t, r, "/api/characters/all")
	testutil.AssertStatus(t, w, 200)
	// non-admin forbidden
	r2 := testutil.NewRouterWithUser(func(auth *gin.RouterGroup) {
		auth.GET("/characters/all", ListAllCharacters)
	}, 2, "user")
	testutil.SeedUser(t, 2, "user2", "user")
	w = testutil.Get(t, r2, "/api/characters/all")
	if w.Code != 403 {
		t.Fatalf("want 403 got %d", w.Code)
	}
}

func TestUpdateCharacterDMNotes(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "Hero", "Human", "Fighter")
	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.PUT("/characters/:id/dm-notes", UpdateCharacterDMNotes)
	})
	w := testutil.PutJSON(t, r, "/api/characters/1/dm-notes", map[string]any{"dm_notes": "secret"})
	testutil.AssertStatus(t, w, 200)
	// forbidden for non-dm user
	r2 := testutil.NewRouterWithUser(func(auth *gin.RouterGroup) {
		auth.PUT("/characters/:id/dm-notes", UpdateCharacterDMNotes)
	}, 2, "user")
	testutil.SeedUser(t, 2, "user2", "user")
	w = testutil.PutJSON(t, r2, "/api/characters/1/dm-notes", map[string]any{"dm_notes": "x"})
	if w.Code != 403 {
		t.Fatalf("want 403 got %d", w.Code)
	}
}

func TestImportCharacterJSON(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/characters/import", ImportCharacterJSON)
	})
	t.Run("success", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/characters/import", map[string]any{"name": "Imported", "race": "Elf", "class": "Wizard", "level": 1})
		testutil.AssertStatus(t, w, 201)
	})
	t.Run("bad json", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/characters/import", "bad")
		if w.Code != 400 {
			t.Fatalf("want 400 got %d", w.Code)
		}
	})
}
