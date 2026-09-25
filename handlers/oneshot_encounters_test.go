package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers/testutil"
)

func seedEncounterTemplate(t *testing.T, id, campaignID, userID int64, name string) {
	t.Helper()
	_, err := db.DB.Exec(
		"INSERT OR IGNORE INTO encounter_templates(id, campaign_id, user_id, name) VALUES(?,NULLIF(?,0),?,?)",
		id, campaignID, userID, name,
	)
	if err != nil {
		t.Fatalf("seed encounter: %v", err)
	}
}

func TestActEncounterLinks(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "dm", "dm")
	testutil.SeedOneShot(t, 10, 1, "Adventure")
	testutil.SeedOneShotAct(t, 100, 10, "Act One", 1)
	seedEncounterTemplate(t, 500, 0, 1, "Goblin Ambush")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/oneshot-acts/:id/encounters", ListActEncounters)
		auth.POST("/oneshot-acts/:id/encounters", LinkActEncounter)
		auth.DELETE("/oneshot-acts/:id/encounters/:eid", UnlinkActEncounter)
		auth.GET("/oneshot-adventures/:id/encounters", GetOneShotEncounters)
		auth.POST("/oneshot-adventures/:id/encounters", LinkOneShotEncounter)
	})

	t.Run("link and list act encounter", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/oneshot-acts/100/encounters", map[string]any{"encounter_id": 500})
		testutil.AssertStatus(t, w, 200)

		w = testutil.Get(t, r, "/api/oneshot-acts/100/encounters")
		testutil.AssertStatus(t, w, 200)
		var links []struct {
			EncounterID   int64  `json:"encounter_id"`
			EncounterName string `json:"encounter_name"`
			ActID         *int64 `json:"act_id"`
		}
		testutil.ParseJSON(t, w, &links)
		if len(links) != 1 || links[0].EncounterID != 500 || links[0].EncounterName != "Goblin Ambush" {
			t.Fatalf("unexpected links %+v", links)
		}
		if links[0].ActID == nil || *links[0].ActID != 100 {
			t.Fatalf("expected act_id 100, got %+v", links[0].ActID)
		}
	})

	t.Run("link is idempotent", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/oneshot-acts/100/encounters", map[string]any{"encounter_id": 500})
		testutil.AssertStatus(t, w, 200)
		if n := testutil.CountRows(t, "oneshot_adventure_encounters"); n != 1 {
			t.Fatalf("expected 1 row, got %d", n)
		}
	})

	t.Run("unlink act encounter", func(t *testing.T) {
		w := testutil.Delete(t, r, "/api/oneshot-acts/100/encounters/500")
		testutil.AssertStatus(t, w, 200)
		if n := testutil.CountRows(t, "oneshot_adventure_encounters"); n != 0 {
			t.Fatalf("expected 0 rows, got %d", n)
		}
	})

	t.Run("adventure-level link keeps nil act_id and stays out of act list", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/oneshot-adventures/10/encounters", map[string]any{"encounter_id": 500})
		testutil.AssertStatus(t, w, 200)

		w = testutil.Get(t, r, "/api/oneshot-acts/100/encounters")
		var actLinks []any
		testutil.ParseJSON(t, w, &actLinks)
		if len(actLinks) != 0 {
			t.Fatalf("expected no act links, got %d", len(actLinks))
		}

		w = testutil.Get(t, r, "/api/oneshot-adventures/10/encounters")
		var advLinks []struct {
			ActID *int64 `json:"act_id"`
		}
		testutil.ParseJSON(t, w, &advLinks)
		if len(advLinks) != 1 || advLinks[0].ActID != nil {
			t.Fatalf("expected one adventure-level link with nil act_id, got %+v", advLinks)
		}
	})
}
