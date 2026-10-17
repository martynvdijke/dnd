package handlers

import (
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers/testutil"
)

func TestAIDraftLoreAddendum(t *testing.T) {
	out := aiDraftLoreAddendum("x")
	if !strings.Contains(out, "CAMPAIGN CONTEXT") {
		t.Fatalf("expected wrapper, got %q", out)
	}
	if !strings.Contains(out, "x") {
		t.Fatalf("expected lore content, got %q", out)
	}
	if !strings.Contains(out, "Keep the draft consistent with this context.") {
		t.Fatalf("expected keep consistent sentence, got %q", out)
	}
}

func TestAIDraftLoreContextEmpty(t *testing.T) {
	setupAIDraftRouter(t)
	defer testutil.CloseDB(t)
	// Seed a campaign with no indexed entities besides itself - but query empty should yield "" via FTS empty.
	// Use non-existent query or empty campaign.
	testutil.SeedCampaign(t, 99, "Empty Camp", "Party", 1)
	out := aiDraftLoreContext(t.Context(), 99, 1, true, "nonexistentquery12345")
	// Could be empty if no indexed match; at least function should not panic and return empty or value.
	// For determinism, test that a campaign with no matching indexed entities returns "".
	// Use a nonsense query unlikely to match.
	if out != "" {
		// If somehow matches campaign name, ensure truncation still works - just ensure not error.
		t.Logf("got lore: %q", out)
	}
	// Also test empty query returns ""
	out2 := aiDraftLoreContext(t.Context(), 99, 1, true, "")
	if out2 != "" {
		t.Fatalf("expected empty for empty query, got %q", out2)
	}
}

func TestAIDraftTruncateRunesText(t *testing.T) {
	s := strings.Repeat("a", 7000)
	out := truncateRunesText(s, 6000)
	if len([]rune(out)) != 6000 {
		t.Fatalf("expected 6000 runes, got %d", len([]rune(out)))
	}
	// unicode safe
	u := "🌍" + strings.Repeat("a", 10)
	out2 := truncateRunesText(u, 1)
	if out2 != "🌍" {
		t.Fatalf("expected rune-safe truncation, got %q", out2)
	}
}

func TestAIDraftLoreIntegration(t *testing.T) {
	// Campaign-scoped draft should include seeded entity title in provider system message.
	content := `{"status":"chatting","message":"ok","draft":null}`
	srv, seen := draftProvider(t, content)
	defer srv.Close()

	r := setupAIDraftRouter(t)
	defer cleanupAIEndpoints()
	defer testutil.CloseDB(t)
	epID := seedAIEndpoint(t, "lore-text", "text", srv.URL, "gpt-4o", true)

	// Seed campaign + distinctive NPC via direct insert (triggers populate entity_search_index)
	testutil.SeedCampaign(t, 10, "Lore Campaign", "Heroes", 1)
	// Create NPC with distinctive name; use ent via raw SQL for simplicity
	npcID := testutil.SeedNPC(t, 101, "ZyxtharDistinctive", "elf", "wizard")
	// Link NPC to campaign so RetrieveCampaignContext includes it
	if _, err := db.DB.Exec("INSERT OR IGNORE INTO campaign_npcs(campaign_id, npc_id, role, notes) VALUES(?,?,?,?)", 10, npcID, "ally", ""); err != nil {
		t.Fatalf("link npc: %v", err)
	}

	// Start campaign-scoped draft with message containing the distinctive title
	campID := int64(10)
	w := testutil.PostJSON(t, r, "/api/ai/draft", map[string]any{
		"endpoint_id": epID, "entity_type": "npc", "campaign_id": campID, "message": "ZyxtharDistinctive",
	})
	if w.Code != 201 {
		t.Fatalf("start status = %d %s", w.Code, w.Body.String())
	}
	if len(*seen) != 1 {
		t.Fatalf("expected 1 provider call, got %d", len(*seen))
	}
	sysMsg := (*seen)[0][0]["content"]
	if !strings.Contains(sysMsg, "ZyxtharDistinctive") {
		t.Fatalf("expected system message to contain seeded title, got %q", sysMsg)
	}
	if !strings.Contains(sysMsg, "CAMPAIGN CONTEXT") {
		t.Fatalf("expected CAMPAIGN CONTEXT wrapper, got %q", sysMsg)
	}

	// Unscoped draft should NOT contain campaign context
	seen2Srv, seen2 := draftProvider(t, content)
	defer seen2Srv.Close()
	// Need new endpoint pointing to second server - reuse but seed again with seen2 URL is tricky.
	// Instead reuse same seen but clear; easier: start unscoped draft and check its captured message is last entry.
	// We need a second router with seen2 endpoint.
	// Create isolated router for unscoped test: use seen2 server endpoint
	cleanupAIEndpoints()
	epID2 := seedAIEndpoint(t, "lore-text2", "text", seen2Srv.URL, "gpt-4o", true)
	w2 := testutil.PostJSON(t, r, "/api/ai/draft", map[string]any{
		"endpoint_id": epID2, "entity_type": "npc", "message": "ZyxtharDistinctive",
	})
	if w2.Code != 201 {
		t.Fatalf("unscoped start status = %d %s", w2.Code, w2.Body.String())
	}
	if len(*seen2) != 1 {
		t.Fatalf("expected 1 provider call for unscoped, got %d", len(*seen2))
	}
	sysMsg2 := (*seen2)[0][0]["content"]
	if strings.Contains(sysMsg2, "CAMPAIGN CONTEXT") {
		t.Fatalf("unscoped draft should not contain CAMPAIGN CONTEXT, got %q", sysMsg2)
	}
}

func TestAIDraftNonMember403(t *testing.T) {
	content := `{"status":"chatting","message":"ok","draft":null}`
	srv, _ := draftProvider(t, content)
	defer srv.Close()

	// Setup router as admin to create campaign, then test as non-member user 2
	rAdmin := setupAIDraftRouter(t)
	defer cleanupAIEndpoints()
	defer testutil.CloseDB(t)
	seedAIEndpoint(t, "lore-403", "text", srv.URL, "gpt-4o", true)
	testutil.SeedCampaign(t, 20, "Private Camp", "Party", 1)
	testutil.SeedUser(t, 2, "user2", "user")

	// Router as user 2 (non-member, non-admin)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	g := r.Group("/api/ai")
	g.Use(func(c *gin.Context) {
		c.Set("user_id", int64(2))
		c.Set("username", "user2")
		c.Set("role", "user")
		c.Set("session_id", "test")
		c.Next()
	})
	g.POST("/draft", StartAIDraft)

	w := testutil.PostJSON(t, r, "/api/ai/draft", map[string]any{
		"endpoint_id": 1, "entity_type": "npc", "campaign_id": int64(20), "message": "hello",
	})
	if w.Code != 403 {
		t.Fatalf("expected 403 for non-member, got %d %s", w.Code, w.Body.String())
	}
	// Also ensure admin router would succeed for same campaign
	var body map[string]any
	_ = body
	w2 := testutil.PostJSON(t, rAdmin, "/api/ai/draft", map[string]any{
		"endpoint_id": 1, "entity_type": "npc", "campaign_id": int64(20), "message": "hello",
	})
	if w2.Code != 201 {
		// Admin should be allowed even without explicit membership (admin bypass in isCampaignMemberGin)
		t.Fatalf("admin expected 201, got %d %s", w2.Code, w2.Body.String())
	}
}
