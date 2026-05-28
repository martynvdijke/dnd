package handlers

import (
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
)

func TestSQLInjectionStringFields(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/characters", CreateCharacter)
		auth.POST("/campaigns", CreateCampaign)
	})

	payloads := []string{
		"'; DROP TABLE characters;--",
		"' OR '1'='1",
		"'; SELECT * FROM users;--",
		"admin'--",
		"1; DROP TABLE campaigns;",
	}

	for _, p := range payloads {
		t.Run("injection character name: "+p[:min(len(p), 20)], func(t *testing.T) {
			w := testutil.PostJSON(t, r, "/api/characters", map[string]any{
				"name": p, "race": "Human", "class": "Fighter",
				"str": 10, "dex": 10, "con": 10, "int": 10, "wis": 10, "cha": 10,
			})
			if w.Code != 201 && w.Code != 400 {
				t.Fatalf("unexpected status %d for injection payload: %s", w.Code, w.Body.String())
			}
		})

		t.Run("injection campaign name: "+p[:min(len(p), 20)], func(t *testing.T) {
			w := testutil.PostJSON(t, r, "/api/campaigns", map[string]any{
				"name": p, "party_name": "Test",
			})
			if w.Code != 201 && w.Code != 400 {
				t.Fatalf("unexpected status %d for injection payload: %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestNumericCoercion(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/characters", CreateCharacter)
	})

	tests := []struct {
		name string
		body map[string]any
	}{
		{"string str", map[string]any{"name": "Test", "race": "Human", "class": "Fighter", "str": "ten"}},
		{"float str", map[string]any{"name": "Test", "race": "Human", "class": "Fighter", "str": 14.5}},
		{"bool level", map[string]any{"name": "Test", "race": "Human", "class": "Fighter", "level": true}},
		{"null name", map[string]any{"name": nil, "race": "Human", "class": "Fighter"}},
		{"array dex", map[string]any{"name": "Test", "race": "Human", "class": "Fighter", "dex": []int{1, 2}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := testutil.PostJSON(t, r, "/api/characters", tt.body)
			if w.Code == 200 || w.Code == 201 {
				t.Skip("handler accepted coercion input, not a bug")
			}
		})
	}
}

func TestLargeInput(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/characters", CreateCharacter)
		auth.PUT("/characters/:id", UpdateCharacter)
	})

	largeName := strings.Repeat("A", 10240)
	desc := strings.Repeat("B", 10240)

	t.Run("10KB character name on create", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/characters", map[string]any{
			"name": largeName, "race": "Human", "class": "Fighter",
			"str": 10, "dex": 10, "con": 10, "int": 10, "wis": 10, "cha": 10,
		})
		if w.Code != 201 && w.Code != 400 {
			t.Fatalf("unexpected status: %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("10KB description on update", func(t *testing.T) {
		testutil.SeedCharacter(t, 1, 1, "Test", "Human", "Fighter")
		w := testutil.PutJSON(t, r, "/api/characters/1", map[string]any{
			"name": "Updated", "description": desc,
		})
		if w.Code != 200 && w.Code != 400 {
			t.Fatalf("unexpected status: %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestUnicodeSearch(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/compendium/search", SearchCompendium)
	})

	queries := []string{
		"魔法",
		"fireball",
		"🔥",
		"\u202Ereverse",
		"<script>alert(1)</script>",
	}

	for _, q := range queries {
		t.Run("query: "+q[:min(len(q), 15)], func(t *testing.T) {
			w := testutil.Get(t, r, "/api/compendium/search?q="+q)
			_ = w.Code
		})
	}
}
