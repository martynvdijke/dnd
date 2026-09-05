package handlers

import (
	"testing"
	"villum/ent"
	"villum/models"
)

func TestEntActToModel_Basic(t *testing.T) {
	e := &ent.OneShotAct{ID: 1, AdventureID: 2, Number: 1, SortOrder: 5, Title: "Act", Description: "desc", EstimatedMinutes: 30, Notes: "n"}
	m := entActToModel(e)
	if m.ID != 1 || m.Title != "Act" || m.SortOrder != 5 {
		t.Fatalf("unexpected mapping %+v", m)
	}
	if m.ParentActID != nil {
		t.Fatalf("expected nil parent")
	}
	e2 := &ent.OneShotAct{ID: 2, AdventureID: 2, Number: 2, ParentActID: 1, Title: "Child"}
	m2 := entActToModel(e2)
	if m2.ParentActID == nil || *m2.ParentActID != 1 {
		t.Fatalf("parent mapping failed %+v", m2)
	}
}

func TestEntAdventureToModel(t *testing.T) {
	e := &ent.OneShotAdventure{ID: 10, UserID: 1, Title: "Adv", Template: "custom"}
	m := entAdventureToModel(e)
	if m.ID != 10 || m.Title != "Adv" {
		t.Fatalf("bad %+v", m)
	}
	if m.CampaignID != nil {
		t.Fatalf("expected nil campaign")
	}
	e.CampaignID = 99
	m2 := entAdventureToModel(e)
	if m2.CampaignID == nil || *m2.CampaignID != 99 {
		t.Fatalf("campaign mapping failed")
	}
}

func TestSortActsBySortOrder(t *testing.T) {
	acts := []models.OneShotAct{{SortOrder: 3, Number: 2}, {SortOrder: 1, Number: 5}, {SortOrder: 1, Number: 1}}
	sortActsBySortOrder(&acts)
	if acts[0].Number != 1 || acts[1].Number != 5 || acts[2].SortOrder != 3 {
		t.Fatalf("sort failed %+v", acts)
	}
}
