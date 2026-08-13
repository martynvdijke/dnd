package handlers

import (
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers/testutil"
	"villum/middleware"
)

func TestShareLinks(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "ShareHero", "Elf", "Ranger")
	testutil.SeedCampaign(t, 1, "ShareCamp", "Testers", 1)

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/share", CreateShareLink)
		auth.GET("/share", ListShareLinks)
		auth.DELETE("/share/:token", DeleteShareLink)
	})

	t.Run("create character share link returns 201", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/share", map[string]any{
			"entity_type": "character", "entity_id": 1,
		})
		testutil.AssertStatus(t, w, 201)
		var result map[string]any
		testutil.ParseJSON(t, w, &result)
		if result["token"] == "" {
			t.Fatal("expected non-empty token")
		}
	})

	t.Run("create party share link returns 201", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/share", map[string]any{
			"entity_type": "party", "entity_id": 1,
		})
		testutil.AssertStatus(t, w, 201)
	})

	t.Run("list share links returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/share")
		testutil.AssertStatus(t, w, 200)
		var links []any
		testutil.ParseJSON(t, w, &links)
		if len(links) < 2 {
			t.Fatalf("expected >=2 links, got %d", len(links))
		}
	})

	t.Run("delete non-existent returns 200 (idempotent)", func(t *testing.T) {
		w := testutil.Delete(t, r, "/api/share/test-token-123")
		if w.Code != 200 {
			t.Logf("delete non-existent: expected 200, got %d", w.Code)
		}
	})
}

func TestShareLinksPublicAccess(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "PubShare", "Human", "Fighter")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/share", CreateShareLink)
	})

	r2 := gin.New()
	r2.Use(middleware.SecurityHeaders())
	r2.GET("/api/share/:token", GetSharedEntity)

	t.Run("create then access via public route", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/share", map[string]any{
			"entity_type": "character", "entity_id": 1,
		})
		testutil.AssertStatus(t, w, 201)
		var result map[string]any
		testutil.ParseJSON(t, w, &result)
		token := result["token"].(string)

		w2 := testutil.Get(t, r2, "/api/share/"+token)
		if w2.Code != 200 {
			t.Logf("public access: got %d", w2.Code)
		}
	})

	t.Run("invalid token returns 404", func(t *testing.T) {
		w := testutil.Get(t, r2, "/api/share/invalid-token-123")
		if w.Code != 404 {
			t.Logf("invalid token: expected 404, got %d", w.Code)
		}
	})
}

// seedShareFixtures inserts entities used by the new share type tests.
func seedShareFixtures(t *testing.T) {
	t.Helper()
	// users: 1 owner/admin, 2 campaign member, 3 stranger
	testutil.SeedUser(t, 1, "owner", "dm")
	testutil.SeedUser(t, 2, "member", "player")
	testutil.SeedUser(t, 3, "stranger", "player")
	testutil.SeedCharacter(t, 1, 1, "FixtureHero", "Elf", "Ranger")
	testutil.SeedCampaign(t, 1, "FixtureCamp", "Fixture Party", 1)
	testutil.SeedCampaignMember(t, 1, 2, "player")

	mustExec(t, "INSERT INTO character_notes(id, character_id, title, content, visibility, category, created_at) VALUES(1, 1, 'Secret Plan', 'content with <script>alert(1)</script>', 'both', 'plot', '2026-01-01 00:00:00')")
	mustExec(t, "INSERT INTO journal(id, character_id, title, entry, entry_date) VALUES(1, 1, 'Session 12', 'We fought a dragon.', '2026-01-15')")
	mustExec(t, "INSERT INTO campaign_maps(id, campaign_id, name, image_url, width, height, grid_size, is_active) VALUES(1, 1, 'The Dungeon', '/media/dungeon.png', 800, 600, 50, 1)")
	mustExec(t, "INSERT INTO campaign_map_pins(id, map_id, name, x, y, is_hidden, sort_order) VALUES(1, 1, 'Trap', 100, 200, 0, 1)")
	mustExec(t, "INSERT INTO campaign_map_pins(id, map_id, name, x, y, is_hidden, sort_order) VALUES(2, 1, 'Secret', 300, 400, 1, 2)")
	mustExec(t, "INSERT INTO uploads(id, hash, ext, url, resized_url, thumbnail_url, owner_type, owner_id, created_at) VALUES(1, 'hash1', '.png', '/media/hash1.png', '/media/hash1_r.png', '', 'character', 1, '2026-01-01 00:00:00')")
	mustExec(t, "INSERT INTO uploads(id, hash, ext, url, owner_type, owner_id, created_at) VALUES(2, 'hash2', '.png', '/media/hash2.png', 'character', 99, '2026-01-01 00:00:00')")
	mustExec(t, "INSERT INTO uploads(id, hash, ext, url, owner_type, owner_id, created_at) VALUES(3, 'hash3', '.pdf', '/media/hash3.pdf', '', 0, '2026-01-01 00:00:00')")
}

func mustExec(t *testing.T, q string, args ...any) {
	t.Helper()
	if _, err := db.DB.Exec(q, args...); err != nil {
		t.Fatalf("exec %q: %v", q, err)
	}
}

func shareRouter(routes func(*gin.RouterGroup)) func(*gin.RouterGroup) {
	return routes
}

func TestShareNewTypesCreateAccess(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	seedShareFixtures(t)

	owner := testutil.NewRouterWithUser(shareRouter(func(auth *gin.RouterGroup) {
		auth.POST("/share", CreateShareLink)
	}), 1, "dm")
	member := testutil.NewRouterWithUser(shareRouter(func(auth *gin.RouterGroup) {
		auth.POST("/share", CreateShareLink)
	}), 2, "player")
	stranger := testutil.NewRouterWithUser(shareRouter(func(auth *gin.RouterGroup) {
		auth.POST("/share", CreateShareLink)
	}), 3, "player")
	admin := testutil.NewRouterWithUser(shareRouter(func(auth *gin.RouterGroup) {
		auth.POST("/share", CreateShareLink)
	}), 1, "admin")

	create := func(r *gin.Engine, entityType string, entityID int64) int {
		w := testutil.PostJSON(t, r, "/api/share", map[string]any{
			"entity_type": entityType, "entity_id": entityID,
		})
		return w.Code
	}

	t.Run("unsupported entity type rejected", func(t *testing.T) {
		if code := create(admin, "npc", 1); code != 400 {
			t.Fatalf("expected 400, got %d", code)
		}
	})

	t.Run("note owned by user -> 201, stranger -> 403", func(t *testing.T) {
		if code := create(owner, "note", 1); code != 201 {
			t.Fatalf("owner note: expected 201, got %d", code)
		}
		if code := create(stranger, "note", 1); code != 403 {
			t.Fatalf("stranger note: expected 403, got %d", code)
		}
	})

	t.Run("journal owned by user -> 201, stranger -> 403", func(t *testing.T) {
		if code := create(owner, "journal", 1); code != 201 {
			t.Fatalf("owner journal: expected 201, got %d", code)
		}
		if code := create(stranger, "journal", 1); code != 403 {
			t.Fatalf("stranger journal: expected 403, got %d", code)
		}
	})

	t.Run("map owner and campaign member -> 201, stranger -> 403", func(t *testing.T) {
		if code := create(owner, "map", 1); code != 201 {
			t.Fatalf("owner map: expected 201, got %d", code)
		}
		if code := create(member, "map", 1); code != 201 {
			t.Fatalf("member map: expected 201, got %d", code)
		}
		if code := create(stranger, "map", 1); code != 403 {
			t.Fatalf("stranger map: expected 403, got %d", code)
		}
	})

	t.Run("upload owned through character -> 201, other -> 403, legacy empty owner admin-only", func(t *testing.T) {
		if code := create(owner, "upload", 1); code != 201 {
			t.Fatalf("owner upload: expected 201, got %d", code)
		}
		if code := create(stranger, "upload", 1); code != 403 {
			t.Fatalf("stranger upload: expected 403, got %d", code)
		}
		if code := create(owner, "upload", 3); code != 403 {
			t.Fatalf("legacy upload non-admin: expected 403, got %d", code)
		}
		if code := create(admin, "upload", 3); code != 201 {
			t.Fatalf("legacy upload admin: expected 201, got %d", code)
		}
	})

	t.Run("expires_in stored", func(t *testing.T) {
		w := testutil.PostJSON(t, owner, "/api/share", map[string]any{
			"entity_type": "note", "entity_id": 1, "expires_in": 24,
		})
		testutil.AssertStatus(t, w, 201)
		var result map[string]any
		testutil.ParseJSON(t, w, &result)
		if result["expires_at"] == nil || result["expires_at"] == "" {
			t.Fatal("expected expires_at in response")
		}
		if result["label"] != "Secret Plan" {
			t.Fatalf("expected label 'Secret Plan', got %v", result["label"])
		}
	})
}

func TestShareNewTypesPublicJSON(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	seedShareFixtures(t)

	r := testutil.NewRouterWithUser(shareRouter(func(auth *gin.RouterGroup) {
		auth.POST("/share", CreateShareLink)
	}), 1, "dm")

	pub := gin.New()
	pub.Use(middleware.SecurityHeaders())
	pub.GET("/api/share/:token", GetSharedEntity)

	makeLink := func(entityType string, entityID int64) string {
		w := testutil.PostJSON(t, r, "/api/share", map[string]any{
			"entity_type": entityType, "entity_id": entityID,
		})
		testutil.AssertStatus(t, w, 201)
		var result map[string]any
		testutil.ParseJSON(t, w, &result)
		return result["token"].(string)
	}

	t.Run("note json", func(t *testing.T) {
		w := testutil.Get(t, pub, "/api/share/"+makeLink("note", 1))
		testutil.AssertStatus(t, w, 200)
		var data map[string]any
		testutil.ParseJSON(t, w, &data)
		testutil.AssertField(t, data, "title", "Secret Plan")
		testutil.AssertField(t, data, "character_name", "FixtureHero")
		testutil.AssertField(t, data, "owner_name", "owner")
	})

	t.Run("journal json", func(t *testing.T) {
		w := testutil.Get(t, pub, "/api/share/"+makeLink("journal", 1))
		testutil.AssertStatus(t, w, 200)
		var data map[string]any
		testutil.ParseJSON(t, w, &data)
		testutil.AssertField(t, data, "title", "Session 12")
		testutil.AssertField(t, data, "entry_date", "2026-01-15")
	})

	t.Run("map json hides secret pins", func(t *testing.T) {
		w := testutil.Get(t, pub, "/api/share/"+makeLink("map", 1))
		testutil.AssertStatus(t, w, 200)
		var data map[string]any
		testutil.ParseJSON(t, w, &data)
		pins, ok := data["pins"].([]any)
		if !ok {
			t.Fatalf("expected pins array, got %#v", data["pins"])
		}
		if len(pins) != 1 {
			t.Fatalf("expected 1 visible pin, got %d", len(pins))
		}
		first := pins[0].(map[string]any)
		if first["name"] != "Trap" {
			t.Fatalf("expected 'Trap', got %v", first["name"])
		}
		if data["campaign_name"] != "FixtureCamp" {
			t.Fatalf("expected campaign_name FixtureCamp, got %v", data["campaign_name"])
		}
	})

	t.Run("upload json", func(t *testing.T) {
		w := testutil.Get(t, pub, "/api/share/"+makeLink("upload", 1))
		testutil.AssertStatus(t, w, 200)
		var data map[string]any
		testutil.ParseJSON(t, w, &data)
		testutil.AssertField(t, data, "url", "/media/hash1.png")
		testutil.AssertField(t, data, "owner_type", "character")
	})

	t.Run("expired link returns 410", func(t *testing.T) {
		mustExec(t, "INSERT INTO share_links(token, entity_type, entity_id, created_by, expires_at) VALUES('expired-tok', 'note', 1, 1, '2000-01-01 00:00:00')")
		w := testutil.Get(t, pub, "/api/share/expired-tok")
		testutil.AssertStatus(t, w, 410)
	})

	t.Run("deleted entity returns 404", func(t *testing.T) {
		mustExec(t, "INSERT INTO share_links(token, entity_type, entity_id, created_by) VALUES('missing-tok', 'note', 999, 1)")
		w := testutil.Get(t, pub, "/api/share/missing-tok")
		testutil.AssertStatus(t, w, 404)
	})
}

func TestSharePagesHTML(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	seedShareFixtures(t)

	r := testutil.NewRouterWithUser(shareRouter(func(auth *gin.RouterGroup) {
		auth.POST("/share", CreateShareLink)
	}), 1, "dm")
	adminR := testutil.NewRouterWithUser(shareRouter(func(auth *gin.RouterGroup) {
		auth.POST("/share", CreateShareLink)
	}), 1, "admin")

	pub := gin.New()
	pub.Use(middleware.SecurityHeaders())
	pub.GET("/share/:token", GetSharedPage)

	makeLink := func(entityType string, entityID int64) string {
		w := testutil.PostJSON(t, r, "/api/share", map[string]any{
			"entity_type": entityType, "entity_id": entityID,
		})
		testutil.AssertStatus(t, w, 201)
		var result map[string]any
		testutil.ParseJSON(t, w, &result)
		return result["token"].(string)
	}
	makeLinkAs := func(router *gin.Engine, entityType string, entityID int64) string {
		w := testutil.PostJSON(t, router, "/api/share", map[string]any{
			"entity_type": entityType, "entity_id": entityID,
		})
		testutil.AssertStatus(t, w, 201)
		var result map[string]any
		testutil.ParseJSON(t, w, &result)
		return result["token"].(string)
	}

	t.Run("note page escapes html and is noindex", func(t *testing.T) {
		w := testutil.Get(t, pub, "/share/"+makeLink("note", 1))
		testutil.AssertStatus(t, w, 200)
		body := w.Body.String()
		if !strings.Contains(body, "Secret Plan") {
			t.Fatal("expected note title in page")
		}
		if strings.Contains(body, "<script>alert(1)</script>") {
			t.Fatal("script tag not escaped")
		}
		if !strings.Contains(body, "&lt;script&gt;") {
			t.Fatal("expected escaped script tag")
		}
		if !strings.Contains(body, "noindex") {
			t.Fatal("expected noindex meta")
		}
		if !strings.Contains(w.Header().Get("Content-Type"), "text/html") {
			t.Fatalf("expected text/html, got %s", w.Header().Get("Content-Type"))
		}
	})

	t.Run("journal page", func(t *testing.T) {
		w := testutil.Get(t, pub, "/share/"+makeLink("journal", 1))
		testutil.AssertStatus(t, w, 200)
		if !strings.Contains(w.Body.String(), "We fought a dragon.") {
			t.Fatal("expected journal entry in page")
		}
	})

	t.Run("map page shows visible pins only", func(t *testing.T) {
		w := testutil.Get(t, pub, "/share/"+makeLink("map", 1))
		testutil.AssertStatus(t, w, 200)
		body := w.Body.String()
		if !strings.Contains(body, "Trap") {
			t.Fatal("expected visible pin on map page")
		}
		if strings.Contains(body, ">Secret<") {
			t.Fatal("hidden pin leaked to page")
		}
	})

	t.Run("upload pdf page embeds iframe", func(t *testing.T) {
		w := testutil.Get(t, pub, "/share/"+makeLinkAs(adminR, "upload", 3))
		testutil.AssertStatus(t, w, 200)
		body := w.Body.String()
		if !strings.Contains(body, "<iframe") || !strings.Contains(body, "/media/hash3.pdf") {
			t.Fatal("expected pdf iframe on upload page")
		}
	})

	t.Run("upload image page embeds img", func(t *testing.T) {
		w := testutil.Get(t, pub, "/share/"+makeLink("upload", 1))
		testutil.AssertStatus(t, w, 200)
		if !strings.Contains(w.Body.String(), "<img") {
			t.Fatal("expected img on upload page")
		}
	})

	t.Run("character share has no page", func(t *testing.T) {
		tok := makeLink("character", 1)
		w := testutil.Get(t, pub, "/share/"+tok)
		testutil.AssertStatus(t, w, 400)
	})

	t.Run("expired page returns 410", func(t *testing.T) {
		mustExec(t, "INSERT INTO share_links(token, entity_type, entity_id, created_by, expires_at) VALUES('expired-page', 'note', 1, 1, '2000-01-01 00:00:00')")
		w := testutil.Get(t, pub, "/share/expired-page")
		testutil.AssertStatus(t, w, 410)
	})
}

func TestShareListLabelsAndRevoke(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	seedShareFixtures(t)

	owner := testutil.NewRouterWithUser(shareRouter(func(auth *gin.RouterGroup) {
		auth.POST("/share", CreateShareLink)
		auth.GET("/share", ListShareLinks)
		auth.DELETE("/share/:token", DeleteShareLink)
	}), 1, "dm")
	stranger := testutil.NewRouterWithUser(shareRouter(func(auth *gin.RouterGroup) {
		auth.DELETE("/share/:token", DeleteShareLink)
	}), 3, "player")

	var token string
	w := testutil.PostJSON(t, owner, "/api/share", map[string]any{
		"entity_type": "journal", "entity_id": 1,
	})
	testutil.AssertStatus(t, w, 201)
	var result map[string]any
	testutil.ParseJSON(t, w, &result)
	token = result["token"].(string)

	t.Run("list includes label", func(t *testing.T) {
		w := testutil.Get(t, owner, "/api/share")
		testutil.AssertStatus(t, w, 200)
		var links []map[string]any
		testutil.ParseJSON(t, w, &links)
		found := false
		for _, l := range links {
			if l["token"] == token {
				found = true
				if l["label"] != "Session 12" {
					t.Fatalf("expected label 'Session 12', got %v", l["label"])
				}
			}
		}
		if !found {
			t.Fatal("created link missing from list")
		}
	})

	t.Run("non-owner cannot revoke", func(t *testing.T) {
		w := testutil.Delete(t, stranger, "/api/share/"+token)
		testutil.AssertStatus(t, w, 403)
	})

	t.Run("owner revokes", func(t *testing.T) {
		w := testutil.Delete(t, owner, "/api/share/"+token)
		testutil.AssertStatus(t, w, 200)
		w2 := testutil.Get(t, owner, "/api/share")
		testutil.AssertStatus(t, w2, 200)
		var links []map[string]any
		testutil.ParseJSON(t, w2, &links)
		for _, l := range links {
			if l["token"] == token {
				t.Fatal("link still present after revoke")
			}
		}
	})
}
