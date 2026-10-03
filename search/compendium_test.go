package search

import (
	"context"
	"encoding/json"
	"testing"

	"villum/db"
	"villum/handlers/testutil"
)

func TestSearchCompendiumPrimaryAndPrefix(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	ctx := context.Background()
	// seed spells
	_, _ = db.DB.Exec(`INSERT INTO compendium_spells(id, name, level, school, casting_time, "range", components, duration, description, classes) VALUES(1,'Fireball',3,'Evocation','1 action','150 feet','V,S,M','Instantaneous','desc','["Wizard"]')`)
	_, _ = db.DB.Exec(`INSERT INTO compendium_spells(id, name, level, school, casting_time, "range", components, duration, description, classes) VALUES(2,'Wall of Fire',4,'Evocation','1 action','120 feet','V,S,M','1 minute','desc','["Wizard"]')`)
	results, err := SearchCompendium(ctx, db.DB, CompendiumParams{Query: "Fireball", Reranker: DefaultRerank})
	if err != nil || len(results) == 0 {
		t.Fatalf("expected Fireball result, err=%v results=%v", err, results)
	}
	found := false
	for _, r := range results {
		if r.Name == "Fireball" {
			found = true
		}
	}
	if !found {
		t.Fatalf("Fireball not found in %v", results)
	}
	// prefix ranks Fireball above Wall of Fire for query fire
	results2, _ := SearchCompendium(ctx, db.DB, CompendiumParams{Query: "fire", Reranker: DefaultRerank})
	if len(results2) < 2 {
		t.Fatalf("expected 2 results for fire, got %v", results2)
	}
	posFireball, posWall := -1, -1
	for i, r := range results2 {
		if r.Name == "Fireball" {
			posFireball = i
		}
		if r.Name == "Wall of Fire" {
			posWall = i
		}
	}
	if posFireball < 0 || posWall < 0 {
		t.Fatalf("missing Fireball or Wall of Fire in %v", results2)
	}
	if posFireball > posWall {
		t.Fatalf("expected Fireball before Wall of Fire, got Fireball at %d Wall at %d: %v", posFireball, posWall, results2)
	}
}

func TestSearchCompendiumFuzzyFallback(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	ctx := context.Background()
	_, _ = db.DB.Exec(`INSERT INTO compendium_spells(id, name, level, school, casting_time, "range", components, duration, description, classes) VALUES(10,'Fireball',3,'Evocation','1 action','150 feet','V,S,M','Instantaneous','desc','["Wizard"]')`)
	results, _ := SearchCompendium(ctx, db.DB, CompendiumParams{Query: "firebal", Reranker: DefaultRerank, FuzzyFallback: true})
	found := false
	for _, r := range results {
		if r.Name == "Fireball" {
			found = true
		}
	}
	if !found {
		t.Fatalf("fuzzy fallback failed to return Fireball for firebal, got %v", results)
	}
}

func TestSearchCompendiumLimitOffset(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	ctx := context.Background()
	for i := 1; i <= 5; i++ {
		_, _ = db.DB.Exec(`INSERT INTO compendium_spells(id, name, level, school) VALUES(?, ?, 1, 'Evocation')`, i, json.RawMessage(`"`+string(rune('A'+i))+`Spell"`))
		name := string(rune('A'+i)) + "Spell"
		_ = name
	}
	// Better: insert predictable names
	db.DB.Exec(`DELETE FROM compendium_spells`)
	for i := 1; i <= 5; i++ {
		db.DB.Exec(`INSERT INTO compendium_spells(id, name, level, school) VALUES(?, ?, 1, 'Evocation')`, i, "Alpha"+string(rune('0'+i)))
	}
	res, _ := SearchCompendium(ctx, db.DB, CompendiumParams{Query: "Alpha", Limit: 2, Offset: 1})
	if len(res) != 2 {
		t.Fatalf("expected 2 results with limit 2 offset 1, got %d %v", len(res), res)
	}
}

func TestLoadCompendiumDetail(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	ctx := context.Background()
	_, _ = db.DB.Exec(`INSERT INTO compendium_spells(id, name, level, school, casting_time, "range", components, duration, description) VALUES(9999,'TestSpell',2,'Illusion','1 action','60 feet','V','1 hour','A spell')`)
	d, err := LoadCompendiumDetail(ctx, db.DB, "spell", 9999)
	if err != nil {
		t.Fatalf("LoadCompendiumDetail spell: %v", err)
	}
	if d.Name != "TestSpell" || d.School != "Illusion" {
		t.Fatalf("unexpected detail %v", d)
	}
	// generic entry
	_, _ = db.DB.Exec(`INSERT INTO compendium_schemas(id, type_name, display_name) VALUES(1,'alchemy','Alchemy')`)
	_, _ = db.DB.Exec(`INSERT INTO compendium_entries(id, schema_id, data) VALUES(500,1,'{"name":"Generic Potion","description":"Heals","category":"potion"}')`)
	// ensure FTS not required for Load
	d2, err := LoadCompendiumDetail(ctx, db.DB, "alchemy", 500)
	if err != nil {
		t.Fatalf("generic detail: %v", err)
	}
	if d2.Name != "Generic Potion" {
		t.Fatalf("expected Generic Potion got %q", d2.Name)
	}
}
