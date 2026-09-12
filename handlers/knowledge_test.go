package handlers

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers/testutil"
)

func knowledgeRoutes(rg *gin.RouterGroup) {
	rg.GET("/campaigns/:id/knowledge", ListKnowledge)
	rg.POST("/campaigns/:id/knowledge", CreateKnowledge)
	rg.GET("/knowledge/:kid", GetKnowledge)
	rg.PUT("/knowledge/:kid", UpdateKnowledge)
	rg.DELETE("/knowledge/:kid", DeleteKnowledge)
	rg.GET("/knowledge/:kid/known-by", ListKnowledgeKnownBy)
	rg.POST("/knowledge/:kid/known-by", AddKnowledgeKnownBy)
	rg.DELETE("/knowledge/:kid/known-by/:cid", RemoveKnowledgeKnownBy)
	rg.POST("/knowledge/:kid/reveal", BulkRevealKnowledge)
}

func TestKnowledge(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "owner", "player")
	testutil.SeedUser(t, 2, "member", "player")
	testutil.SeedUser(t, 3, "stranger", "player")
	testutil.SeedCampaign(t, 1, "Test Campaign", "Party", 1)
	_, err := db.DB.Exec("INSERT OR IGNORE INTO campaign_members(campaign_id,user_id,role) VALUES(?,?,?)", 1, 2, "player")
	if err != nil {
		t.Fatalf("seed member: %v", err)
	}

	ownerRouter := testutil.NewRouterWithUser(knowledgeRoutes, 1, "player")
	memberRouter := testutil.NewRouterWithUser(knowledgeRoutes, 2, "player")
	strangerRouter := testutil.NewRouterWithUser(knowledgeRoutes, 3, "player")

	// seed characters for known-by / bulk reveal
	testutil.SeedCharacterInCampaign(t, 10, 1, 1, "Alice", "Human", "Fighter")
	testutil.SeedCharacterInCampaign(t, 11, 2, 1, "Bob", "Elf", "Wizard")

	t.Run("Create as DM", func(t *testing.T) {
		w := testutil.PostJSON(t, ownerRouter, "/api/campaigns/1/knowledge", map[string]any{
			"title": "Rumor 1", "content": "Something", "source": "tavern", "status": "rumor",
		})
		testutil.AssertStatus(t, w, 201)
		var k Knowledge
		testutil.ParseJSON(t, w, &k)
		if k.ID == 0 {
			t.Fatal("expected id > 0")
		}
		if k.Status != "rumor" {
			t.Fatalf("expected status rumor got %q", k.Status)
		}
		if k.Shared {
			t.Fatal("expected shared false")
		}
		w = testutil.Get(t, ownerRouter, "/api/campaigns/1/knowledge")
		testutil.AssertStatus(t, w, 200)
		var list []Knowledge
		testutil.ParseJSON(t, w, &list)
		found := false
		for _, e := range list {
			if e.ID == k.ID {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("list did not contain created entry %d: %+v", k.ID, list)
		}
	})

	t.Run("Default status", func(t *testing.T) {
		w := testutil.PostJSON(t, ownerRouter, "/api/campaigns/1/knowledge", map[string]any{
			"title": "Default rumor", "content": "content",
		})
		testutil.AssertStatus(t, w, 201)
		var k Knowledge
		testutil.ParseJSON(t, w, &k)
		if k.Status != "rumor" {
			t.Fatalf("expected default status rumor got %q", k.Status)
		}
	})

	t.Run("Invalid status rejected", func(t *testing.T) {
		w := testutil.PostJSON(t, ownerRouter, "/api/campaigns/1/knowledge", map[string]any{
			"title": "Bad", "content": "x", "status": "bogus",
		})
		testutil.AssertStatus(t, w, 400)
	})

	t.Run("Title required", func(t *testing.T) {
		w := testutil.PostJSON(t, ownerRouter, "/api/campaigns/1/knowledge", map[string]any{
			"title": "", "content": "x",
		})
		testutil.AssertStatus(t, w, 400)
		w = testutil.PostJSON(t, ownerRouter, "/api/campaigns/1/knowledge", map[string]any{
			"title": "   ", "content": "x",
		})
		testutil.AssertStatus(t, w, 400)
	})

	t.Run("Status lifecycle history", func(t *testing.T) {
		w := testutil.PostJSON(t, ownerRouter, "/api/campaigns/1/knowledge", map[string]any{
			"title": "History entry", "content": "c", "status": "rumor",
		})
		testutil.AssertStatus(t, w, 201)
		var k Knowledge
		testutil.ParseJSON(t, w, &k)
		w = testutil.PutJSON(t, ownerRouter, "/api/knowledge/"+itoaKnowledge(k.ID), map[string]any{
			"status": "revealed",
		})
		testutil.AssertStatus(t, w, 200)
		var updated Knowledge
		testutil.ParseJSON(t, w, &updated)
		if updated.Status != "revealed" {
			t.Fatalf("expected status revealed got %q", updated.Status)
		}
		var hist []map[string]string
		if err := json.Unmarshal([]byte(updated.StatusHistory), &hist); err != nil {
			t.Fatalf("unmarshal status_history: %v body %q", err, updated.StatusHistory)
		}
		hasRumor, hasRevealed := false, false
		for _, h := range hist {
			if h["status"] == "rumor" {
				hasRumor = true
				if h["at"] == "" {
					t.Fatal("rumor history entry missing at")
				}
			}
			if h["status"] == "revealed" {
				hasRevealed = true
				if h["at"] == "" {
					t.Fatal("revealed history entry missing at")
				}
			}
		}
		if !hasRumor || !hasRevealed {
			t.Fatalf("history should contain rumor and revealed, got %+v", hist)
		}
	})

	t.Run("Known-by add/remove", func(t *testing.T) {
		w := testutil.PostJSON(t, ownerRouter, "/api/campaigns/1/knowledge", map[string]any{
			"title": "KnownBy entry", "content": "c",
		})
		testutil.AssertStatus(t, w, 201)
		var k Knowledge
		testutil.ParseJSON(t, w, &k)
		w = testutil.PostJSON(t, ownerRouter, "/api/knowledge/"+itoaKnowledge(k.ID)+"/known-by", map[string]any{
			"character_id": 10,
		})
		testutil.AssertStatus(t, w, 200)
		w = testutil.Get(t, ownerRouter, "/api/knowledge/"+itoaKnowledge(k.ID)+"/known-by")
		testutil.AssertStatus(t, w, 200)
		var ids []int64
		testutil.ParseJSON(t, w, &ids)
		if !containsID(ids, 10) {
			t.Fatalf("expected known-by to contain 10, got %v", ids)
		}
		w = testutil.Delete(t, ownerRouter, "/api/knowledge/"+itoaKnowledge(k.ID)+"/known-by/10")
		testutil.AssertStatus(t, w, 200)
		w = testutil.Get(t, ownerRouter, "/api/knowledge/"+itoaKnowledge(k.ID)+"/known-by")
		testutil.AssertStatus(t, w, 200)
		testutil.ParseJSON(t, w, &ids)
		if containsID(ids, 10) {
			t.Fatalf("expected known-by to not contain 10 after delete, got %v", ids)
		}
	})

	t.Run("Visibility split", func(t *testing.T) {
		w := testutil.PostJSON(t, ownerRouter, "/api/campaigns/1/knowledge", map[string]any{
			"title": "Shared entry", "content": "shared", "shared": true,
		})
		testutil.AssertStatus(t, w, 201)
		var shared Knowledge
		testutil.ParseJSON(t, w, &shared)
		w = testutil.PostJSON(t, ownerRouter, "/api/campaigns/1/knowledge", map[string]any{
			"title": "Unshared entry", "content": "secret", "shared": false,
		})
		testutil.AssertStatus(t, w, 201)
		var unshared Knowledge
		testutil.ParseJSON(t, w, &unshared)

		// member list sees only shared
		w = testutil.Get(t, memberRouter, "/api/campaigns/1/knowledge")
		testutil.AssertStatus(t, w, 200)
		var mList []Knowledge
		testutil.ParseJSON(t, w, &mList)
		if !containsKnowledge(mList, shared.ID) {
			t.Fatalf("member should see shared entry %d, got %v", shared.ID, ids(mList))
		}
		if containsKnowledge(mList, unshared.ID) {
			t.Fatalf("member should not see unshared entry %d", unshared.ID)
		}
		// owner list sees both
		w = testutil.Get(t, ownerRouter, "/api/campaigns/1/knowledge")
		testutil.AssertStatus(t, w, 200)
		var oList []Knowledge
		testutil.ParseJSON(t, w, &oList)
		if !containsKnowledge(oList, shared.ID) || !containsKnowledge(oList, unshared.ID) {
			t.Fatalf("owner should see both entries, got %v", ids(oList))
		}
		// member direct access
		w = testutil.Get(t, memberRouter, "/api/knowledge/"+itoaKnowledge(unshared.ID))
		testutil.AssertStatus(t, w, 404)
		w = testutil.Get(t, memberRouter, "/api/knowledge/"+itoaKnowledge(shared.ID))
		testutil.AssertStatus(t, w, 200)
		// member cannot modify
		w = testutil.PutJSON(t, memberRouter, "/api/knowledge/"+itoaKnowledge(shared.ID), map[string]any{"title": "hacked"})
		testutil.AssertStatus(t, w, 404)
		w = testutil.Delete(t, memberRouter, "/api/knowledge/"+itoaKnowledge(shared.ID))
		testutil.AssertStatus(t, w, 404)
	})

	t.Run("Bulk reveal", func(t *testing.T) {
		w := testutil.PostJSON(t, ownerRouter, "/api/campaigns/1/knowledge", map[string]any{
			"title": "To reveal", "content": "secret", "shared": false,
		})
		testutil.AssertStatus(t, w, 201)
		var k Knowledge
		testutil.ParseJSON(t, w, &k)
		w = testutil.PostJSON(t, ownerRouter, "/api/knowledge/"+itoaKnowledge(k.ID)+"/reveal", map[string]any{})
		testutil.AssertStatus(t, w, 200)
		var revealed Knowledge
		testutil.ParseJSON(t, w, &revealed)
		if !revealed.Shared {
			t.Fatal("expected shared true after bulk reveal")
		}
		w = testutil.Get(t, ownerRouter, "/api/knowledge/"+itoaKnowledge(k.ID)+"/known-by")
		testutil.AssertStatus(t, w, 200)
		var ids2 []int64
		testutil.ParseJSON(t, w, &ids2)
		if !containsID(ids2, 10) || !containsID(ids2, 11) {
			t.Fatalf("expected known-by to contain all campaign characters 10,11 got %v", ids2)
		}
	})

	t.Run("Non-member no access", func(t *testing.T) {
		w := testutil.Get(t, strangerRouter, "/api/campaigns/1/knowledge")
		testutil.AssertStatus(t, w, 404)
		w = testutil.PostJSON(t, strangerRouter, "/api/campaigns/1/knowledge", map[string]any{
			"title": "Should fail", "content": "x",
		})
		testutil.AssertStatus(t, w, 404)
	})
}

func itoaKnowledge(n int64) string {
	return strconv.FormatInt(n, 10)
}

func containsID(ids []int64, want int64) bool {
	for _, id := range ids {
		if id == want {
			return true
		}
	}
	return false
}

func containsKnowledge(list []Knowledge, id int64) bool {
	for _, k := range list {
		if k.ID == id {
			return true
		}
	}
	return false
}

func ids(list []Knowledge) []int64 {
	out := make([]int64, 0, len(list))
	for _, k := range list {
		out = append(out, k.ID)
	}
	return out
}
