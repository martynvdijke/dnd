package handlers

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"villum/db"
	"villum/handlers/testutil"
	"villum/models"
)

func TestDnDBeyondAdapter(t *testing.T) {
	src := dndBeyondSource{}
	payload := `{
		"name":"Test Hero",
		"stats":[{"id":1,"value":15},{"id":2,"value":14},{"id":3,"value":13},{"id":4,"value":12},{"id":5,"value":10},{"id":6,"value":8}],
		"classes":[{"definition":{"name":"Fighter"},"level":3}],
		"baseHitPoints":20,"bonusHitPoints":5,"removedHitPoints":5,
		"armorClass":16,
		"race":{"fullName":"Human"},
		"background":{"definition":{"name":"Soldier"}},
		"alignmentName":"Lawful Good",
		"currentXp":300,
		"inventory":[{"definition":{"name":"Longsword","filterType":"Weapon","damage":{"diceString":"1d8","damageType":"slashing"},"properties":[{"name":"Versatile"}]},"quantity":1,"equipped":true}]
	}`
	chars, err := src.Parse([]byte(payload))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(chars) != 1 {
		t.Fatalf("len %d", len(chars))
	}
	c := chars[0]
	if c.Name != "Test Hero" || c.Str != 15 || c.Dex != 14 || c.Con != 13 || c.Int != 12 || c.Wis != 10 || c.Cha != 8 {
		t.Fatalf("stats mismatch %+v", c)
	}
	if c.Class != "Fighter" || c.Level != 3 {
		t.Fatalf("class/level %+v", c)
	}
	if c.HPMax != 25 || c.HPCurrent != 20 {
		t.Fatalf("hp %+v", c)
	}
	if c.AC != 16 || c.Race != "Human" || c.Background != "Soldier" || c.Alignment != "Lawful Good" {
		t.Fatalf("ac/race/bg %+v", c)
	}
	if len(c.Inventory) != 1 || c.Inventory[0].Name != "Longsword" || c.Inventory[0].DamageDice != "1d8" {
		t.Fatalf("inv %+v", c.Inventory)
	}
	// wrapper
	wrapped := `{"data":` + payload + `}`
	chars2, err := src.Parse([]byte(wrapped))
	if err != nil || chars2[0].Name != "Test Hero" {
		t.Fatalf("wrapped failed %v", err)
	}
	// invalid json
	if _, err := src.Parse([]byte("not json")); err == nil {
		t.Fatal("expected error")
	}
}

func TestFoundryActorAdapter(t *testing.T) {
	src := foundryActorSource{}
	payload := `{
		"name":"Foundry Hero",
		"system":{
			"abilities":{"str":{"value":16},"dex":{"value":14},"con":{"value":15},"int":{"value":10},"wis":{"value":12},"cha":{"value":8}},
			"attributes":{"hp":{"value":18,"max":22},"ac":{"value":15},"speed":{"value":"30"}},
			"details":{"level":5},
			"currency":{"pp":1,"gp":10,"ep":2,"sp":5,"cp":20}
		},
		"items":[{"name":"Battleaxe","type":"weapon","system":{"damage":{"parts":[["1d8","slashing"]]},"properties":["versatile"],"quantity":1}}]
	}`
	chars, err := src.Parse([]byte(payload))
	if err != nil {
		t.Fatalf("parse %v", err)
	}
	c := chars[0]
	if c.Name != "Foundry Hero" || c.Str != 16 || c.Dex != 14 || c.HPMax != 22 || c.HPCurrent != 18 || c.AC != 15 || c.Level != 5 {
		t.Fatalf("mismatch %+v", c)
	}
	if len(c.Inventory) != 1 || c.Inventory[0].DamageDice != "1d8" || c.Inventory[0].DamageType != "slashing" {
		t.Fatalf("inv %+v", c.Inventory)
	}
	// wrapper
	wrapped := `{"actor":` + payload + `}`
	chars2, err := src.Parse([]byte(wrapped))
	if err != nil || chars2[0].Name != "Foundry Hero" {
		t.Fatalf("wrapped %v", err)
	}
}

func TestFiveEToolsAdapter(t *testing.T) {
	src := fiveEToolsSource{}
	j1 := `{"monster":[{"name":"A"},{"name":"B"}]}`
	entries, err := src.Parse([]byte(j1))
	if err != nil || len(entries) != 2 {
		t.Fatalf("monster %v %d", err, len(entries))
	}
	j2 := `[{"name":"A"},{"name":"B"}]`
	entries, err = src.Parse([]byte(j2))
	if err != nil || len(entries) != 2 {
		t.Fatalf("bare array %v", err)
	}
	j3 := `{"data":[{"name":"X"}]}`
	entries, err = src.Parse([]byte(j3))
	if err != nil || len(entries) != 1 {
		t.Fatalf("data %v", err)
	}
	j4 := `{"name":"Solo"}`
	entries, err = src.Parse([]byte(j4))
	if err != nil || len(entries) != 1 || entries[0]["name"] != "Solo" {
		t.Fatalf("single %v", err)
	}
}

func TestFoundryPackAdapter(t *testing.T) {
	src := foundryPackSource{}
	j := `{"entries":[{"name":"A"}]}`
	entries, err := src.Parse([]byte(j))
	if err != nil || len(entries) != 1 {
		t.Fatalf("entries %v", err)
	}
	if _, err := src.Parse([]byte("\x00\x01leveldb")); err == nil || !containsStr(err.Error(), "LevelDB") {
		t.Fatalf("expected LevelDB error got %v", err)
	}
	j2 := `{"items":[{"name":"B"}]}`
	entries, err = src.Parse([]byte(j2))
	if err != nil || len(entries) != 1 {
		t.Fatalf("items %v", err)
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (func() bool {
		for i := 0; i <= len(s)-len(sub); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})()
}

func TestExternalImportCharacterDryRun(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/import/external", HandleExternalImport)
	})
	payload := map[string]any{
		"name": "Dry Hero", "stats": []any{map[string]any{"id": float64(1), "value": float64(14)}}, "armorClass": float64(12),
	}
	payloadBytes, _ := json.Marshal(payload)
	req := map[string]any{"source": "dndbeyond", "kind": "character", "payload": json.RawMessage(payloadBytes), "dry_run": true}
	w := testutil.PostJSON(t, r, "/api/import/external", req)
	testutil.AssertStatus(t, w, 200)
	var resp map[string]any
	testutil.ParseJSON(t, w, &resp)
	if int(resp["created"].(float64)) != 0 {
		t.Fatalf("expected 0 created got %v", resp)
	}
	if testutil.CountRows(t, "characters") != 0 {
		t.Fatal("should not create rows on dry_run")
	}
}

func TestExternalImportCharacterCommitAndRollback(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/import/external", HandleExternalImport)
		auth.GET("/import/external/logs", HandleExternalImportLogs)
		auth.POST("/import/external/logs/:id/rollback", HandleExternalImportRollback)
	})
	payload := map[string]any{"name": "Commit Hero", "armorClass": float64(12)}
	payloadBytes, _ := json.Marshal(payload)
	req := map[string]any{"source": "dndbeyond", "kind": "character", "payload": json.RawMessage(payloadBytes)}
	w := testutil.PostJSON(t, r, "/api/import/external", req)
	testutil.AssertStatus(t, w, 200)
	var resp map[string]any
	testutil.ParseJSON(t, w, &resp)
	if int(resp["created"].(float64)) != 1 {
		t.Fatalf("created %v", resp)
	}
	logID := int64(resp["log_id"].(float64))
	if logID == 0 {
		t.Fatal("log_id missing")
	}
	if testutil.CountRows(t, "characters") != 1 {
		t.Fatal("character not inserted")
	}
	// rollback
	w = testutil.PostJSON(t, r, "/api/import/external/logs/"+strconv.FormatInt(logID, 10)+"/rollback", map[string]any{})
	testutil.AssertStatus(t, w, 200)
	if testutil.CountRows(t, "characters") != 0 {
		t.Fatal("rollback failed")
	}
	// check log status
	var status string
	db.DB.QueryRow("SELECT status FROM external_import_logs WHERE id=?", logID).Scan(&status)
	if status != "rolled_back" {
		t.Fatalf("status %s", status)
	}
	// idempotent second rollback
	w = testutil.PostJSON(t, r, "/api/import/external/logs/"+strconv.FormatInt(logID, 10)+"/rollback", map[string]any{})
	testutil.AssertStatus(t, w, 200)
}

func TestExternalImportCompendiumDedup(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	_, err := db.DB.Exec(`INSERT INTO compendium_schemas(id, type_name, display_name, fields) VALUES(1,'testtype','TestType','[{"name":"name","label":"Name","type":"string","required":true}]')`)
	if err != nil {
		t.Fatalf("seed schema %v", err)
	}
	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/import/external", HandleExternalImport)
	})
	entries := []map[string]any{{"name": "Goblin"}, {"name": "Orc"}}
	payload, _ := json.Marshal(map[string]any{"monster": entries})
	req := map[string]any{"source": "5etools", "kind": "compendium", "schema_id": float64(1), "payload": json.RawMessage(payload), "dedup_action": "skip"}
	w := testutil.PostJSON(t, r, "/api/import/external", req)
	testutil.AssertStatus(t, w, 200)
	var resp map[string]any
	testutil.ParseJSON(t, w, &resp)
	if int(resp["created"].(float64)) != 2 {
		t.Fatalf("first import %v", resp)
	}
	// second same
	w = testutil.PostJSON(t, r, "/api/import/external", req)
	testutil.AssertStatus(t, w, 200)
	testutil.ParseJSON(t, w, &resp)
	if int(resp["duplicates"].(float64)) == 0 {
		t.Fatalf("expected duplicates %v", resp)
	}
	if testutil.CountRows(t, "compendium_entries") != 2 {
		t.Fatalf("dedup created extra rows")
	}
	// also test suggestMapping
	m := suggestMapping(entries, []models.SchemaField{{Name: "name"}, {Name: "other"}})
	if len(m) != 1 || m[0].Target != "name" {
		t.Fatalf("suggest %v", m)
	}
}

func TestFetchExternalURLLoopbackRejected(t *testing.T) {
	if _, err := fetchExternalURL("http://127.0.0.1/x"); err == nil {
		t.Fatal("expected error loopback")
	}
	if _, err := fetchExternalURL("http://localhost/x"); err == nil {
		t.Fatal("expected error localhost")
	}
}
