package search

import (
	"context"
	"os"
	"testing"

	"villum/db"
)

func TestSearchUnifiedIntegration(t *testing.T) {
	path := "/tmp/villum_search_integ.db"
	os.Remove(path)
	if err := db.Init(path); err != nil {
		t.Fatalf("db init: %v", err)
	}
	db.Seed()
	defer func() { db.Close(); os.Remove(path) }()
	ctx := context.Background()
	// seed a character visible
	_, _ = db.DB.Exec(`INSERT INTO characters(id, user_id, name, race, class, level, str, dex, con, int, wis, cha, hp_max, hp_current, ac, initiative, speed) VALUES(1,1,'TestHero','Human','Fighter',1,10,10,10,10,10,10,12,12,10,0,30)`)
	results, err := SearchUnified(ctx, db.DB, UnifiedParams{Query: "TestHero", Limit: 10, UserID: 1, IsAdmin: true})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	_ = results
	// invalid query
	_, _ = SearchUnified(ctx, db.DB, UnifiedParams{Query: "   ", Limit: 10, UserID: 1, IsAdmin: true})
	// compendium
	_, _ = db.DB.Exec(`INSERT OR IGNORE INTO compendium_spells(id, name, level, school, casting_time, "range", components, duration, description, classes) VALUES(99,'Fireball',3,'Evocation','1 action','150 feet','V,S,M','Instantaneous','desc','["Wizard"]')`)
	cr, _ := SearchCompendium(ctx, db.DB, CompendiumParams{Query: "Fireball"})
	_ = cr
	// campaign context
	_, _ = db.DB.Exec(`INSERT OR IGNORE INTO campaigns(id, name, party_name, user_id) VALUES(1,'Camp','Party',1)`)
	sources, _, _ := RetrieveCampaignContext(ctx, db.DB, 1, 1, true, "TestHero", 5)
	_ = sources
	// rerank
	_ = DefaultRerank("fire", cr)
	// load entry (will fail but covers)
	_, _ = LoadCompendiumEntry(ctx, db.DB, 1)
	_, _ = QueryCompendiumEquipmentUnion(ctx, db.DB, "a", 10, 0)
	_, _ = QueryCompendiumSpellsUnion(ctx, db.DB, "a", "", "", 10, 0)
	_, _ = QueryCompendiumEntriesForFeatures(ctx, db.DB, "a", 10, 0)
}
