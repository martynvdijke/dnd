package handlers

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
)

func wsLiveClient(uid int64) *WSClient {
	c := &WSClient{userID: uid, role: "player", send: make(chan []byte, 8)}
	Hub.Register(c)
	return c
}

func wsLiveCleanup(c *WSClient) {
	Hub.Unregister(c)
	// drain
	for {
		select {
		case <-c.send:
		default:
			return
		}
	}
}

func wsExpectMsg(t *testing.T, c *WSClient) WSMessage {
	t.Helper()
	select {
	case raw := <-c.send:
		var m WSMessage
		if err := json.Unmarshal(raw, &m); err != nil {
			t.Fatalf("unmarshal WSMessage: %v raw=%s", err, string(raw))
		}
		return m
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for WS message")
		return WSMessage{}
	}
}

func wsExpectNoMsg(t *testing.T, c *WSClient) {
	t.Helper()
	select {
	case raw := <-c.send:
		t.Fatalf("expected no message but got %s", string(raw))
	case <-time.After(150 * time.Millisecond):
	}
}

func resetHub() {
	Hub.mu.Lock()
	Hub.clients = make(map[int64][]*WSClient)
	Hub.mu.Unlock()
}

func TestSendDiceRoll_LiveTable(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	defer resetHub()
	testutil.SeedUser(t, 1, "owner", "player")
	testutil.SeedUser(t, 2, "member", "player")
	testutil.SeedUser(t, 3, "stranger", "player")
	testutil.SeedCampaign(t, 1, "Camp", "Party", 1)
	testutil.SeedCampaignMember(t, 1, 2, "player")

	owner := wsLiveClient(1)
	member := wsLiveClient(2)
	stranger := wsLiveClient(3)
	defer wsLiveCleanup(owner)
	defer wsLiveCleanup(member)
	defer wsLiveCleanup(stranger)

	SendDiceRoll(1, 1, 10, "owner", "1d20+3", 15, "1d20+3 = 15")

	for _, c := range []*WSClient{owner, member} {
		m := wsExpectMsg(t, c)
		if m.Type != WSEventDiceRoll {
			t.Fatalf("expected type %q got %q", WSEventDiceRoll, m.Type)
		}
		var p DiceRollEvent
		if err := json.Unmarshal(m.Payload, &p); err != nil {
			t.Fatalf("unmarshal DiceRollEvent: %v", err)
		}
		if p.Expression != "1d20+3" || p.Total != 15 || p.Username != "owner" {
			t.Fatalf("unexpected payload %+v", p)
		}
	}
	wsExpectNoMsg(t, stranger)
}

func TestSendCombatUpdate_LiveTable(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	defer resetHub()
	testutil.SeedUser(t, 1, "owner", "player")
	testutil.SeedUser(t, 2, "member", "player")
	testutil.SeedUser(t, 3, "stranger", "player")
	testutil.SeedCampaign(t, 1, "Camp", "Party", 1)
	testutil.SeedCampaignMember(t, 1, 2, "player")

	owner := wsLiveClient(1)
	member := wsLiveClient(2)
	stranger := wsLiveClient(3)
	defer wsLiveCleanup(owner)
	defer wsLiveCleanup(member)
	defer wsLiveCleanup(stranger)

	SendCombatUpdate(1)

	for _, c := range []*WSClient{owner, member} {
		m := wsExpectMsg(t, c)
		if m.Type != WSEventCombatUpdate {
			t.Fatalf("expected type %q got %q", WSEventCombatUpdate, m.Type)
		}
		var p map[string]int64
		if err := json.Unmarshal(m.Payload, &p); err != nil {
			t.Fatalf("unmarshal combat payload: %v", err)
		}
		if p["campaign_id"] != 1 {
			t.Fatalf("expected campaign_id 1 got %v", p)
		}
	}
	wsExpectNoMsg(t, stranger)
}

func TestSendKnowledgeReveal_LiveTable(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	defer resetHub()
	testutil.SeedUser(t, 1, "owner", "player")
	testutil.SeedUser(t, 2, "member", "player")
	testutil.SeedUser(t, 3, "stranger", "player")
	testutil.SeedCampaign(t, 1, "Camp", "Party", 1)
	testutil.SeedCampaignMember(t, 1, 2, "player")

	t.Run("unshared emits nothing", func(t *testing.T) {
		resetHub()
		owner := wsLiveClient(1)
		member := wsLiveClient(2)
		defer wsLiveCleanup(owner)
		defer wsLiveCleanup(member)

		SendKnowledgeReveal(Knowledge{ID: 1, CampaignID: 1, Title: "Secret", Content: "x", Shared: false})
		wsExpectNoMsg(t, owner)
		wsExpectNoMsg(t, member)
	})

	t.Run("shared emits to members only", func(t *testing.T) {
		resetHub()
		owner := wsLiveClient(1)
		member := wsLiveClient(2)
		stranger := wsLiveClient(3)
		defer wsLiveCleanup(owner)
		defer wsLiveCleanup(member)
		defer wsLiveCleanup(stranger)

		SendKnowledgeReveal(Knowledge{ID: 42, CampaignID: 1, Title: "Revealed", Content: "hello", Source: "tavern", Status: "revealed", Shared: true})

		for _, c := range []*WSClient{owner, member} {
			m := wsExpectMsg(t, c)
			if m.Type != WSEventKnowledgeReveal {
				t.Fatalf("expected %q got %q", WSEventKnowledgeReveal, m.Type)
			}
			var p map[string]any
			if err := json.Unmarshal(m.Payload, &p); err != nil {
				t.Fatalf("unmarshal payload: %v", err)
			}
			if p["title"] != "Revealed" {
				t.Fatalf("unexpected payload %+v", p)
			}
		}
		wsExpectNoMsg(t, stranger)
	})
}

func TestKnowledgeLiveTable_UpdateAndBulkReveal(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	defer resetHub()

	testutil.SeedUser(t, 1, "owner", "player")
	testutil.SeedUser(t, 2, "member", "player")
	testutil.SeedUser(t, 3, "stranger", "player")
	testutil.SeedCampaign(t, 1, "Camp", "Party", 1)
	testutil.SeedCampaignMember(t, 1, 2, "player")
	testutil.SeedCharacterInCampaign(t, 10, 1, 1, "Alice", "Human", "Fighter")
	testutil.SeedCharacterInCampaign(t, 11, 2, 1, "Bob", "Elf", "Wizard")

	ownerRouter := testutil.NewRouterWithUser(knowledgeRoutes, 1, "player")

	// helper to attach WS clients and drain
	attachClients := func() (owner, member, stranger *WSClient) {
		resetHub()
		owner = wsLiveClient(1)
		member = wsLiveClient(2)
		stranger = wsLiveClient(3)
		return
	}

	// create an unshared entry
	w := testutil.PostJSON(t, ownerRouter, "/api/campaigns/1/knowledge", map[string]any{
		"title": "Secret", "content": "hidden", "shared": false,
	})
	testutil.AssertStatus(t, w, 201)
	var k Knowledge
	testutil.ParseJSON(t, w, &k)

	t.Run("PUT shared false->true emits", func(t *testing.T) {
		owner, member, stranger := attachClients()
		defer wsLiveCleanup(owner)
		defer wsLiveCleanup(member)
		defer wsLiveCleanup(stranger)

		w := testutil.PutJSON(t, ownerRouter, "/api/knowledge/"+itoaKnowledge(k.ID), map[string]any{"shared": true})
		testutil.AssertStatus(t, w, 200)

		for _, c := range []*WSClient{owner, member} {
			m := wsExpectMsg(t, c)
			if m.Type != WSEventKnowledgeReveal {
				t.Fatalf("expected %q got %q", WSEventKnowledgeReveal, m.Type)
			}
		}
		wsExpectNoMsg(t, stranger)
	})

	t.Run("PUT title only does not emit when already shared", func(t *testing.T) {
		owner, member, stranger := attachClients()
		defer wsLiveCleanup(owner)
		defer wsLiveCleanup(member)
		defer wsLiveCleanup(stranger)

		// entry is now shared=true from previous subtest
		w := testutil.PutJSON(t, ownerRouter, "/api/knowledge/"+itoaKnowledge(k.ID), map[string]any{"title": "Renamed"})
		testutil.AssertStatus(t, w, 200)

		wsExpectNoMsg(t, owner)
		wsExpectNoMsg(t, member)
		wsExpectNoMsg(t, stranger)
	})

	t.Run("PUT title only does not emit when still unshared", func(t *testing.T) {
		// create fresh unshared entry
		w := testutil.PostJSON(t, ownerRouter, "/api/campaigns/1/knowledge", map[string]any{
			"title": "Another secret", "content": "x", "shared": false,
		})
		testutil.AssertStatus(t, w, 201)
		var k2 Knowledge
		testutil.ParseJSON(t, w, &k2)

		owner, member, stranger := attachClients()
		defer wsLiveCleanup(owner)
		defer wsLiveCleanup(member)
		defer wsLiveCleanup(stranger)

		w = testutil.PutJSON(t, ownerRouter, "/api/knowledge/"+itoaKnowledge(k2.ID), map[string]any{"title": "Still secret"})
		testutil.AssertStatus(t, w, 200)

		wsExpectNoMsg(t, owner)
		wsExpectNoMsg(t, member)
		wsExpectNoMsg(t, stranger)
	})

	t.Run("BulkReveal emits", func(t *testing.T) {
		w := testutil.PostJSON(t, ownerRouter, "/api/campaigns/1/knowledge", map[string]any{
			"title": "Bulk secret", "content": "y", "shared": false,
		})
		testutil.AssertStatus(t, w, 201)
		var k3 Knowledge
		testutil.ParseJSON(t, w, &k3)

		owner, member, stranger := attachClients()
		defer wsLiveCleanup(owner)
		defer wsLiveCleanup(member)
		defer wsLiveCleanup(stranger)

		w = testutil.PostJSON(t, ownerRouter, "/api/knowledge/"+itoaKnowledge(k3.ID)+"/reveal", map[string]any{})
		testutil.AssertStatus(t, w, 200)

		for _, c := range []*WSClient{owner, member} {
			m := wsExpectMsg(t, c)
			if m.Type != WSEventKnowledgeReveal {
				t.Fatalf("expected %q got %q", WSEventKnowledgeReveal, m.Type)
			}
		}
		wsExpectNoMsg(t, stranger)
	})
}

func TestHandleRoll_LiveTable(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	defer resetHub()

	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 1, "Camp", "Party", 1)
	testutil.SeedCampaignMember(t, 1, 1, "dm")
	// character without campaign
	testutil.SeedCharacter(t, 10, 1, "NoCamp", "Human", "Fighter")
	// character with campaign
	testutil.SeedCharacterInCampaign(t, 11, 1, 1, "CampChar", "Elf", "Wizard")

	_ = getDicePool()

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/roll", HandleRoll)
	})

	t.Run("NULL campaign_id emits nothing", func(t *testing.T) {
		resetHub()
		owner := wsLiveClient(1)
		defer wsLiveCleanup(owner)

		w := testutil.PostJSON(t, r, "/api/roll", map[string]any{
			"expression": "1d4", "character_id": 10,
		})
		testutil.AssertStatus(t, w, 200)
		wsExpectNoMsg(t, owner)
	})

	t.Run("campaign character emits dice_roll", func(t *testing.T) {
		resetHub()
		owner := wsLiveClient(1)
		defer wsLiveCleanup(owner)
		// also ensure member would receive — seed member
		testutil.SeedUser(t, 2, "member", "player")
		testutil.SeedCampaignMember(t, 1, 2, "player")
		member := wsLiveClient(2)
		defer wsLiveCleanup(member)

		w := testutil.PostJSON(t, r, "/api/roll", map[string]any{
			"expression": "1d4", "character_id": 11,
		})
		testutil.AssertStatus(t, w, 200)

		for _, c := range []*WSClient{owner, member} {
			m := wsExpectMsg(t, c)
			if m.Type != WSEventDiceRoll {
				t.Fatalf("expected %q got %q", WSEventDiceRoll, m.Type)
			}
			var p DiceRollEvent
			if err := json.Unmarshal(m.Payload, &p); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if p.Expression != "1d4" {
				t.Fatalf("expected expression 1d4 got %q", p.Expression)
			}
			if p.CharacterID != 11 {
				t.Fatalf("expected character 11 got %d", p.CharacterID)
			}
		}
	})

	t.Run("SendDiceRoll direct coverage path", func(t *testing.T) {
		resetHub()
		c := wsLiveClient(1)
		defer wsLiveCleanup(c)
		SendDiceRoll(1, 1, 11, "admin", "1d6+1", 4, "1d6+1 = 4")
		m := wsExpectMsg(t, c)
		if m.Type != WSEventDiceRoll {
			t.Fatalf("expected dice_roll got %q", m.Type)
		}
		var p DiceRollEvent
		json.Unmarshal(m.Payload, &p)
		if p.Total != 4 || p.Username != "admin" {
			t.Fatalf("payload mismatch %+v", p)
		}
	})
}
