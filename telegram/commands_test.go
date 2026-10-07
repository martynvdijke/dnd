package telegram

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	tgbot "github.com/go-telegram/bot"
)

func TestBotCommandList_ContainsExpected(t *testing.T) {
	list := botCommandList()
	names := map[string]bool{}
	for _, c := range list {
		names[c.Command] = true
	}
	for _, want := range []string{"search", "spell", "item", "monster", "race", "class", "feat", "background", "ask", "help"} {
		if !names[want] {
			t.Fatalf("botCommandList missing %q, got %v", want, names)
		}
	}
}

func TestRegisterBotCommands(t *testing.T) {
	setupTelegramDB(t)
	defer func() {
		ResetBotUsernameCache()
	}()
	var capturedCommands string
	var capturedMethod string
	srv := mockTelegramServer(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "getMe") {
			w.Write(jsonOK(map[string]any{"id": 1, "is_bot": true, "first_name": "T", "username": "bot"}))
			return
		}
		capturedMethod = r.URL.Path
		capturedCommands = requestFormValue(r, "commands")
		w.Write(jsonOK(true))
	})
	client, err := tgbot.New("test-token", tgbot.WithServerURL(srv.URL))
	if err != nil {
		t.Fatalf("new bot: %v", err)
	}
	registerBotCommands(client)
	if !strings.Contains(capturedMethod, "setMyCommands") {
		t.Fatalf("expected setMyCommands, got %q", capturedMethod)
	}
	if capturedCommands == "" {
		t.Fatalf("expected commands payload")
	}
	var cmds []map[string]any
	if err := json.Unmarshal([]byte(capturedCommands), &cmds); err != nil {
		t.Fatalf("commands not JSON array: %v %q", err, capturedCommands)
	}
	names := map[string]bool{}
	for _, c := range cmds {
		if n, ok := c["command"].(string); ok {
			names[n] = true
		}
	}
	for _, want := range []string{"search", "spell", "help"} {
		if !names[want] {
			t.Fatalf("registered commands missing %q, got %v", want, names)
		}
	}

	// error path: mock returns error, should not panic
	srv2 := mockTelegramServer(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "getMe") {
			w.Write(jsonOK(map[string]any{"id": 1, "is_bot": true, "first_name": "T", "username": "bot"}))
			return
		}
		w.Write(jsonErr(500, "boom"))
	})
	client2, err := tgbot.New("test-token", tgbot.WithServerURL(srv2.URL))
	if err != nil {
		t.Fatalf("new bot2: %v", err)
	}
	// should not panic
	registerBotCommands(client2)

	// nil client should not panic
	registerBotCommands(nil)
}
