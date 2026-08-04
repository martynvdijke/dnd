package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers/testutil"
	"villum/middleware"
)

// newCampaignNPCsTestRouter boots a router with the htmx campaign NPC routes
// and an authenticated admin context for user 1.
func newCampaignNPCsTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	testutil.NewDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 1, "TestCamp", "Party", 1)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.SecurityHeaders())
	g := r.Group("/")
	g.Use(func(c *gin.Context) {
		c.Set("user_id", int64(1))
		c.Set("username", "admin")
		c.Set("role", "admin")
		c.Set("session_id", "test-session")
		c.Next()
	})
	HtmxRegisterRoutes(g)
	return r
}

func campaignNPCPost(t *testing.T, r *gin.Engine, path string, form url.Values) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.ServeHTTP(w, req)
	return w
}

func campaignNPCGet(t *testing.T, r *gin.Engine, path string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	r.ServeHTTP(w, req)
	return w
}

func campaignNPCDelete(t *testing.T, r *gin.Engine, path string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, path, nil)
	r.ServeHTTP(w, req)
	return w
}

// TestCampaignCreateAndLinkNPC covers the "New NPC" flow: creating an NPC
// from the campaign page and linking it to the campaign in one step.
func TestCampaignCreateAndLinkNPC(t *testing.T) {
	r := newCampaignNPCsTestRouter(t)
	defer testutil.CloseDB(t)

	w := campaignNPCPost(t, r, "/htmx/campaigns/1/npcs/create-and-link", url.Values{
		"name":        {"Elminster"},
		"race":        {"Human"},
		"class":       {"Wizard"},
		"role":        {"ally"},
		"description": {"Sage of Shadowdale"},
		"notes":       {"friend of Gandalf"},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("create-and-link status = %d, want 200 (body: %s)", w.Code, w.Body.String())
	}
	body := w.Body.String()
	for _, want := range []string{"Elminster", "Human", "Wizard", "ally", "friend of Gandalf", "1 NPCs"} {
		if !strings.Contains(body, want) {
			t.Errorf("rendered section missing %q (body: %s)", want, body)
		}
	}

	var npcCount, linkCount int
	if err := db.DB.QueryRow("SELECT COUNT(*) FROM npcs WHERE user_id = 1 AND name = 'Elminster'").Scan(&npcCount); err != nil {
		t.Fatalf("count npcs: %v", err)
	}
	if npcCount != 1 {
		t.Errorf("npcs rows = %d, want 1", npcCount)
	}
	if err := db.DB.QueryRow("SELECT COUNT(*) FROM campaign_npcs WHERE campaign_id = 1").Scan(&linkCount); err != nil {
		t.Fatalf("count campaign_npcs: %v", err)
	}
	if linkCount != 1 {
		t.Errorf("campaign_npcs rows = %d, want 1", linkCount)
	}
}

// TestCampaignCreateAndLinkNPCRequiresName verifies the 400 on missing name.
func TestCampaignCreateAndLinkNPCRequiresName(t *testing.T) {
	r := newCampaignNPCsTestRouter(t)
	defer testutil.CloseDB(t)

	w := campaignNPCPost(t, r, "/htmx/campaigns/1/npcs/create-and-link", url.Values{
		"role": {"enemy"},
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("create-and-link without name status = %d, want 400", w.Code)
	}
}

// TestCampaignNPCsSection covers listing linked NPCs on the campaign page,
// including the empty state and a linked NPC rendered with metadata.
func TestCampaignNPCsSection(t *testing.T) {
	r := newCampaignNPCsTestRouter(t)
	defer testutil.CloseDB(t)

	// Empty state first.
	w := campaignNPCGet(t, r, "/htmx/campaigns/1/npcs-section")
	if w.Code != http.StatusOK {
		t.Fatalf("npcs-section status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "No NPCs linked to this campaign.") {
		t.Errorf("empty section missing empty state (body: %s)", w.Body.String())
	}

	// Link an existing NPC, then re-render the section.
	testutil.SeedNPC(t, 2, "Drizzt", "Drow", "Ranger")
	w = campaignNPCPost(t, r, "/htmx/campaigns/1/npcs/link", url.Values{
		"npc_id": {"2"},
		"role":   {"enemy"},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("link status = %d, want 200 (body: %s)", w.Code, w.Body.String())
	}

	w = campaignNPCGet(t, r, "/htmx/campaigns/1/npcs-section")
	body := w.Body.String()
	for _, want := range []string{"Drizzt", "Drow", "Ranger", "enemy", "1 NPCs"} {
		if !strings.Contains(body, want) {
			t.Errorf("linked section missing %q (body: %s)", want, body)
		}
	}
}

// TestCampaignUnlinkNPC covers removing the campaign link while keeping the
// NPC itself intact.
func TestCampaignUnlinkNPC(t *testing.T) {
	r := newCampaignNPCsTestRouter(t)
	defer testutil.CloseDB(t)

	testutil.SeedNPC(t, 3, "Lolth", "Drow", "Cleric")
	w := campaignNPCPost(t, r, "/htmx/campaigns/1/npcs/link", url.Values{
		"npc_id": {"3"},
		"role":   {"enemy"},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("link status = %d, want 200", w.Code)
	}

	var linkID int64
	if err := db.DB.QueryRow("SELECT id FROM campaign_npcs WHERE campaign_id = 1 AND npc_id = 3").Scan(&linkID); err != nil {
		t.Fatalf("find link row: %v", err)
	}

	w = campaignNPCDelete(t, r, "/htmx/campaigns/1/npcs/"+itoa64(linkID))
	if w.Code != http.StatusOK {
		t.Fatalf("unlink status = %d, want 200 (body: %s)", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if strings.Contains(body, "Lolth") {
		t.Errorf("unlinked section still contains NPC name (body: %s)", body)
	}
	if !strings.Contains(body, "No NPCs linked to this campaign.") {
		t.Errorf("unlinked section missing empty state (body: %s)", body)
	}

	var npcAlive, linkCount int
	if err := db.DB.QueryRow("SELECT COUNT(*) FROM npcs WHERE id = 3").Scan(&npcAlive); err != nil {
		t.Fatalf("count npcs: %v", err)
	}
	if npcAlive != 1 {
		t.Errorf("unlink deleted the NPC row itself; npcs count = %d, want 1", npcAlive)
	}
	if err := db.DB.QueryRow("SELECT COUNT(*) FROM campaign_npcs WHERE campaign_id = 1").Scan(&linkCount); err != nil {
		t.Fatalf("count campaign_npcs: %v", err)
	}
	if linkCount != 0 {
		t.Errorf("campaign_npcs rows = %d, want 0", linkCount)
	}
}

// TestCampaignNPCLinkForm covers the link-form dropdown listing the user's
// NPCs so they can be linked to the campaign.
func TestCampaignNPCLinkForm(t *testing.T) {
	r := newCampaignNPCsTestRouter(t)
	defer testutil.CloseDB(t)

	testutil.SeedNPC(t, 4, "Volo", "Human", "Loremaster")
	w := campaignNPCGet(t, r, "/htmx/campaigns/1/npcs/link-form")
	if w.Code != http.StatusOK {
		t.Fatalf("link-form status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Volo") {
		t.Errorf("link-form missing user's NPC (body: %s)", w.Body.String())
	}
}


func itoa64(v int64) string {
	return strconv.FormatInt(v, 10)
}
