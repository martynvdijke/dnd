package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
)

func requireFields(t *testing.T, data map[string]any, fields []string) {
	t.Helper()
	for _, f := range fields {
		if _, ok := data[f]; !ok {
			t.Fatalf("response missing field %q in %+v", f, data)
		}
	}
}

func TestResponseFields(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/characters", CreateCharacter)
		auth.GET("/characters/:id", GetCharacter)
		auth.POST("/campaigns", CreateCampaign)
		auth.POST("/roll", HandleRoll)
	})

	t.Run("create character response fields", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/characters", map[string]any{
			"name": "Field Check", "race": "Dwarf", "class": "Cleric", "level": 3,
			"str": 10, "dex": 10, "con": 10, "int": 10, "wis": 10, "cha": 10,
			"hp_max": 24, "hp_current": 24,
		})
		var data map[string]any
		testutil.ParseJSON(t, w, &data)
		requireFields(t, data, []string{"id", "name", "race", "class", "level",
			"str", "dex", "con", "int", "wis", "cha", "hp_max", "hp_current"})
	})

	t.Run("get character response fields", func(t *testing.T) {
		testutil.SeedCharacter(t, 1, 1, "Getter", "Human", "Rogue")
		w := testutil.Get(t, r, "/api/characters/1")
		var data map[string]any
		testutil.ParseJSON(t, w, &data)
		requireFields(t, data, []string{"id", "name", "race", "class"})
	})

	t.Run("create campaign response fields", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/campaigns", map[string]any{
			"name": "Field Campaign", "party_name": "Testers",
		})
		var data map[string]any
		testutil.ParseJSON(t, w, &data)
		requireFields(t, data, []string{"id", "name", "party_name"})
	})
}

func TestErrorMessages(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/characters", CreateCharacter)
		auth.GET("/characters/:id", GetCharacter)
		auth.POST("/roll", HandleRoll)
	})

	hasError := func(t *testing.T, w *httptest.ResponseRecorder) {
		t.Helper()
		var data map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &data); err != nil {
			t.Fatalf("json decode: %v", err)
		}
		errMsg, ok := data["error"]
		if !ok || errMsg == "" {
			t.Fatalf("response missing non-empty error field: %+v", data)
		}
	}

	t.Run("empty character body has error", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/characters", map[string]any{})
		hasError(t, w)
	})

	t.Run("empty dice expression has error", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/roll", map[string]any{"expression": ""})
		hasError(t, w)
	})

	t.Run("non-existent character has error", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/characters/99999")
		hasError(t, w)
	})

	t.Run("missing name has error", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/characters", map[string]any{
			"race": "Human", "class": "Fighter",
		})
		hasError(t, w)
	})
}

func TestListEndpointsReturnArrays(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/characters", ListCharacters)
		auth.GET("/campaigns", ListCampaigns)
		auth.GET("/encounters", ListEncounters)
		auth.POST("/encounters", CreateEncounter)
		auth.GET("/dice-rolls", GetDiceRolls)
		auth.POST("/roll", HandleRoll)
	})

	// Seed some data so lists are non-empty
	testutil.SeedCharacter(t, 1, 1, "ArrTest", "Human", "Fighter")
	testutil.PostJSON(t, r, "/api/encounters", map[string]any{"name": "ArrEnc", "difficulty": "easy"})
	testutil.PostJSON(t, r, "/api/roll", map[string]any{"expression": "1d20"})

	tests := []struct {
		name string
		path string
	}{
		{"characters", "/api/characters"},
		{"campaigns", "/api/campaigns"},
		{"encounters", "/api/encounters"},
		{"dice rolls", "/api/dice-rolls"},
	}

	for _, tt := range tests {
		t.Run(tt.name+" returns array", func(t *testing.T) {
			w := testutil.Get(t, r, tt.path)
			var arr []any
			if err := json.Unmarshal(w.Body.Bytes(), &arr); err != nil {
				t.Fatalf("expected JSON array, got error: %v (body: %s)", err, w.Body.String())
			}
			if len(arr) < 1 {
				t.Logf("array is empty (may be expected if no data)")
			}
		})
	}
}
