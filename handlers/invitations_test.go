package handlers

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers/testutil"
	"villum/middleware"
)

func TestCampaignInvitations(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "dmuser", "user")
	testutil.SeedCampaign(t, 1, "InviteCamp", "Party", 1)
	testutil.SeedUser(t, 2, "player1", "user")
	testutil.SeedUser(t, 3, "outsider", "user")
	db.DB.Exec("UPDATE users SET email=? WHERE id=2", "player1@example.com")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/campaigns/:id/invitations", ListCampaignInvitations)
		auth.POST("/campaigns/:id/invitations", CreateCampaignInvitation)
		auth.DELETE("/campaigns/:id/invitations/:inviteId", RevokeCampaignInvitation)
	})

	public := gin.New()
	public.GET("/invite/:token", InviteAcceptPage)
	public.POST("/invite/:token", InviteAccept)

	var token string

	t.Run("create invitation returns token and link", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/campaigns/1/invitations", map[string]any{
			"email": "player1@example.com", "role": "player",
		})
		testutil.AssertStatus(t, w, 201)
		var res map[string]any
		testutil.ParseJSON(t, w, &res)
		var ok bool
		token, ok = res["token"].(string)
		if !ok || token == "" {
			t.Fatalf("expected non-empty token, got %v", res["token"])
		}
		if !strings.Contains(res["url"].(string), "/invite/"+token) {
			t.Fatalf("expected url to contain invite path, got %v", res["url"])
		}
		// No SMTP configured in tests → email not sent, link still returned.
		if res["email_sent"] != false {
			t.Fatalf("expected email_sent=false, got %v", res["email_sent"])
		}
	})

	t.Run("list pending invitations", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/campaigns/1/invitations")
		testutil.AssertStatus(t, w, 200)
		var list []map[string]any
		testutil.ParseJSON(t, w, &list)
		if len(list) != 1 {
			t.Fatalf("expected 1 pending invitation, got %d", len(list))
		}
		if list[0]["email"] != "player1@example.com" {
			t.Fatalf("unexpected email %v", list[0]["email"])
		}
	})

	t.Run("public page without session prompts sign in", func(t *testing.T) {
		w := testutil.Get(t, public, "/invite/"+token)
		testutil.AssertStatus(t, w, 200)
		if !strings.Contains(w.Body.String(), "Sign in") {
			t.Fatalf("expected sign-in prompt, got %q", w.Body.String())
		}
	})

	t.Run("accept without session redirects to login", func(t *testing.T) {
		req := testutil.NewTestRequest("POST", "/invite/"+token, nil)
		w := httptest.NewRecorder()
		public.ServeHTTP(w, req)
		if w.Code != http.StatusSeeOther {
			t.Fatalf("expected 303, got %d", w.Code)
		}
		if loc := w.Header().Get("Location"); !strings.HasPrefix(loc, "/login?next=") {
			t.Fatalf("expected login redirect, got %q", loc)
		}
	})

	t.Run("accept with matching session joins campaign", func(t *testing.T) {
		sid := middleware.Store.Create(2, "player1", "user", "127.0.0.1")
		req := testutil.NewTestRequest("POST", "/invite/"+token, nil)
		req.AddCookie(&http.Cookie{Name: "session", Value: sid})
		w := httptest.NewRecorder()
		public.ServeHTTP(w, req)
		if w.Code != http.StatusSeeOther {
			t.Fatalf("expected 303, got %d: %s", w.Code, w.Body.String())
		}
		var role string
		db.DB.QueryRow("SELECT role FROM campaign_members WHERE campaign_id=1 AND user_id=2").Scan(&role)
		if role != "player" {
			t.Fatalf("expected player membership, got %q", role)
		}
		var accepted int
		db.DB.QueryRow("SELECT COUNT(*) FROM campaign_invitations WHERE token=? AND accepted_at IS NOT NULL", token).Scan(&accepted)
		if accepted != 1 {
			t.Fatalf("expected invitation marked accepted")
		}
	})

	t.Run("accept from mismatched email is rejected", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/campaigns/1/invitations", map[string]any{
			"email": "someone-else@example.com", "role": "player",
		})
		testutil.AssertStatus(t, w, 201)
		var res map[string]any
		testutil.ParseJSON(t, w, &res)
		other := res["token"].(string)

		sid := middleware.Store.Create(3, "outsider", "user", "127.0.0.1")
		db.DB.Exec("UPDATE users SET email=? WHERE id=3", "outsider@example.com")
		req := testutil.NewTestRequest("POST", "/invite/"+other, nil)
		req.AddCookie(&http.Cookie{Name: "session", Value: sid})
		w2 := httptest.NewRecorder()
		public.ServeHTTP(w2, req)
		if w2.Code != http.StatusForbidden {
			t.Fatalf("expected 403 for email mismatch, got %d", w2.Code)
		}
	})

	t.Run("revoke removes pending invitation", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/campaigns/1/invitations", map[string]any{
			"email": "temp@example.com", "role": "dm",
		})
		testutil.AssertStatus(t, w, 201)
		var res map[string]any
		testutil.ParseJSON(t, w, &res)
		id := int64(res["id"].(float64))

		w2 := testutil.Delete(t, r, "/api/campaigns/1/invitations/"+strconv.FormatInt(id, 10))
		testutil.AssertStatus(t, w2, 200)
		var remaining int
		db.DB.QueryRow("SELECT COUNT(*) FROM campaign_invitations WHERE email='temp@example.com'").Scan(&remaining)
		if remaining != 0 {
			t.Fatalf("expected invitation revoked, %d remain", remaining)
		}
	})

	t.Run("non-dm cannot invite", func(t *testing.T) {
		r3 := testutil.NewRouterWithUser(func(auth *gin.RouterGroup) {
			auth.POST("/campaigns/:id/invitations", CreateCampaignInvitation)
		}, 3, "user")
		w := testutil.PostJSON(t, r3, "/api/campaigns/1/invitations", map[string]any{
			"email": "x@example.com", "role": "player",
		})
		testutil.AssertStatus(t, w, 403)
	})
}
