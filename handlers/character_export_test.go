package handlers

import (
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"villum/handlers/testutil"
)

func TestPrintCharacterHTML(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)

	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "Aria the Bold", "Elf", "Wizard")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/characters/:id/print", PrintCharacter)
	})

	// HTML format.
	w := testutil.Get(t, r, "/api/characters/1/print?format=html")
	testutil.AssertStatus(t, w, 200)
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Fatalf("content-type = %q, want text/html", ct)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Aria the Bold") {
		t.Fatalf("html sheet missing character name")
	}
	if !strings.Contains(body, "<!doctype html>") {
		t.Fatalf("html sheet is not a full document")
	}

	// Default format stays plain text.
	w = testutil.Get(t, r, "/api/characters/1/print")
	testutil.AssertStatus(t, w, 200)
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "text/plain") {
		t.Fatalf("default content-type = %q, want text/plain", ct)
	}
	if !strings.Contains(w.Body.String(), "Aria the Bold") {
		t.Fatalf("text sheet missing character name")
	}
}
