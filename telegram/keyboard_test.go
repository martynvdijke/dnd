package telegram

import (
	"context"
	"net/http"
	"strings"
	"testing"

	tgmodels "github.com/go-telegram/bot/models"

	"villum/db"
	"villum/handlers/testutil"
)

func TestParseCallbackID(t *testing.T) {
	id, err := parseCallbackID("claim:42", "claim:")
	if err != nil || id != 42 {
		t.Fatalf("expected 42 got %d err %v", id, err)
	}
	if _, err := parseCallbackID("claim:abc", "claim:"); err == nil {
		t.Fatalf("expected error for abc")
	}
	if _, err := parseCallbackID("claim:", "claim:"); err == nil {
		t.Fatalf("expected error for empty")
	}
}

func TestHandleCallbackData_SwitchArms(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 10, 1, "Aria", "Elf", "Ranger")
	testutil.SeedCampaign(t, 20, "Camp", "Party", 1)
	if _, err := db.DB.Exec("INSERT OR IGNORE INTO campaign_characters(campaign_id, character_id) VALUES(20,10)"); err != nil {
		t.Fatalf("attach: %v", err)
	}
	// Need identities for handlers that require linked user
	if err := UpsertIdentity(1, 999, 999, "admin"); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	// mock not needed for handleCallbackData alone, but some handlers may send replies via DB; ensure no panic
	c := &cmdContext{ctx: context.Background(), chatID: 999, tgUserID: 999, username: "admin"}

	// sub:toggle
	if _, ok := handleCallbackData(c, "sub:toggle"); !ok {
		t.Fatalf("sub:toggle should be handled")
	}

	// nav:help valid
	if _, ok := handleCallbackData(c, "nav:help"); !ok {
		t.Fatalf("nav:help should be handled")
	}
	// nav:nosuchcmd unknown -> false
	if _, ok := handleCallbackData(c, "nav:nosuchcmd"); ok {
		t.Fatalf("unknown nav should be false")
	}

	// claim:1 — needs editable character; character 10 owned by admin so ok
	if reply, ok := handleCallbackData(c, "claim:10"); !ok {
		t.Fatalf("claim:10 not handled")
	} else {
		_ = reply
	}
	// malformed claim
	if _, ok := handleCallbackData(c, "claim:abc"); ok {
		t.Fatalf("malformed claim should be false")
	}

	// sheet:10
	if _, ok := handleCallbackData(c, "sheet:10"); !ok {
		t.Fatalf("sheet:10 should be handled")
	}
	if _, ok := handleCallbackData(c, "sheet:abc"); ok {
		t.Fatalf("malformed sheet should be false")
	}

	// campaign:20
	if _, ok := handleCallbackData(c, "campaign:20"); !ok {
		t.Fatalf("campaign:20 should be handled")
	}
	if _, ok := handleCallbackData(c, "campaign:abc"); ok {
		t.Fatalf("malformed campaign should be false")
	}

	// create:level:3 requires flow active; without flow returns expired message but ok=true
	if reply, ok := handleCallbackData(c, "create:level:3"); !ok {
		t.Fatalf("create:level:3 should be ok")
	} else {
		if !strings.Contains(reply.Text, "expired") {
			t.Fatalf("expected expired for no flow, got %q", reply.Text)
		}
	}
	// create:level malformed
	if _, ok := handleCallbackData(c, "create:level:abc"); ok {
		t.Fatalf("malformed level should be false")
	}

	// create:campaign:none
	if reply, ok := handleCallbackData(c, "create:campaign:none"); !ok {
		t.Fatalf("create:campaign:none should be ok")
	} else {
		_ = reply
	}
	// create:campaign:20 (valid campaign)
	// Need active flow for this to succeed; without flow it returns expired
	// So test with active flow
	SetCharacterCreator(func(ctx context.Context, userID int64, in CreateCharacterInput) (CreatedCharacter, error) {
		return CreatedCharacter{ID: 99, Name: in.Name}, nil
	})
	defer SetCharacterCreator(nil)
	// start flow
	rc := &cmdContext{ctx: context.Background(), chatID: 999, tgUserID: 999}
	// use runCreate to create flow then handle callback
	runCreate(rc)
	// Manually set level so campaign choice is valid step
	flow := getCreateFlow(999)
	if flow != nil {
		flow.Level = 1
		flow.Step = createStepCampaign
		putCreateFlow(999, flow)
	}
	if _, ok := handleCallbackData(rc, "create:campaign:20"); !ok {
		t.Fatalf("create:campaign:20 with flow should be ok")
	}
	// malformed campaign id
	if _, ok := handleCallbackData(c, "create:campaign:abc"); ok {
		t.Fatalf("malformed create campaign should be false")
	}

	// unknown prefix
	if _, ok := handleCallbackData(c, "bogus:1"); ok {
		t.Fatalf("unknown prefix should be false")
	}
}

func TestHandleCallback(t *testing.T) {
	// nil cq
	handleCallback(context.Background(), nil)

	// cq with nil Message.Message
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	cq := &tgmodels.CallbackQuery{
		ID:   "cb1",
		From: tgmodels.User{ID: 1},
		Message: tgmodels.MaybeInaccessibleMessage{
			Message: nil,
		},
		Data: "nav:help",
	}
	// needs mock for AnswerCallback otherwise transient client will error but not panic
	mockTelegramServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write(jsonOK(true))
	})
	handleCallback(context.Background(), cq)

	// happy path
	testutil.SeedUser(t, 1, "admin", "admin")
	if err := UpsertIdentity(1, 1, 1, "admin"); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	var methods []string
	var texts []string
	mockTelegramServer(t, func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.URL.Path)
		if txt := requestFormValue(r, "text"); txt != "" {
			texts = append(texts, txt)
		}
		// also check for answerCallbackQuery vs sendMessage
		w.Write(jsonOK(map[string]any{"message_id": 1}))
	})
	// Need to handle JSON body for method path? The server path includes /bot<token>/method
	// Use string contains check
	cq2 := &tgmodels.CallbackQuery{
		ID:   "cb2",
		From: tgmodels.User{ID: 1, Username: "admin", FirstName: "A"},
		Message: tgmodels.MaybeInaccessibleMessage{
			Message: &tgmodels.Message{
				ID:   10,
				Chat: tgmodels.Chat{ID: 1},
			},
		},
		Data: "nav:help",
	}
	handleCallback(context.Background(), cq2)
	joined := strings.Join(methods, " ")
	if !strings.Contains(joined, "answerCallbackQuery") {
		t.Fatalf("expected answerCallbackQuery, got %v", methods)
	}
	if !strings.Contains(joined, "sendMessage") {
		t.Fatalf("expected sendMessage, got %v texts %v", methods, texts)
	}
	// Verify the sent message actually carried the help text.
	found := false
	for _, txt := range texts {
		if strings.Contains(txt, "commands") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected help text in sent messages, got %v", texts)
	}
}
