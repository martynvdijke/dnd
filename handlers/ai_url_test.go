package handlers

import (
	"encoding/json"
	"testing"

	"villum/handlers/testutil"
)

func TestNormalizeAIBaseURL(t *testing.T) {
	cases := []struct{ in, want string }{
		{"https://api.openai.com/v1", "https://api.openai.com/v1"},
		{"https://api.openai.com/v1/", "https://api.openai.com/v1"},
		{"https://opencode.ai/zen/go/v1/chat/completions", "https://opencode.ai/zen/go/v1"},
		{"https://opencode.ai/zen/go/v1/chat/completions/", "https://opencode.ai/zen/go/v1"},
		{"https://opencode.ai/zen/go/v1/images/generations", "https://opencode.ai/zen/go/v1"},
		{" https://host/v1/completions ", "https://host/v1"},
		{"", ""},
	}
	for _, tc := range cases {
		if got := normalizeAIBaseURL(tc.in); got != tc.want {
			t.Errorf("normalizeAIBaseURL(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestAIRequestURLStripsDuplicateOperationPath(t *testing.T) {
	got := aiRequestURL("https://opencode.ai/zen/go/v1/chat/completions", "/chat/completions")
	want := "https://opencode.ai/zen/go/v1/chat/completions"
	if got != want {
		t.Fatalf("aiRequestURL = %q, want %q", got, want)
	}
	got = aiRequestURL("https://api.openai.com/v1/", "/images/generations")
	want = "https://api.openai.com/v1/images/generations"
	if got != want {
		t.Fatalf("aiRequestURL = %q, want %q", got, want)
	}
}

func TestValidateAIBaseURL(t *testing.T) {
	cases := []struct {
		in    string
		ok    bool
		wantN string
	}{
		{"https://api.openai.com/v1", true, "https://api.openai.com/v1"},
		{"https://opencode.ai/zen/go/v1/chat/completions", true, "https://opencode.ai/zen/go/v1"},
		{"http://localhost:8080/v1", true, "http://localhost:8080/v1"},
		{"not-a-url", false, "not-a-url"},
		{"ftp://example.com/v1", false, "ftp://example.com/v1"},
		{"", false, ""},
	}
	for _, tc := range cases {
		got, ok := validateAIBaseURL(tc.in)
		if ok != tc.ok {
			t.Errorf("validateAIBaseURL(%q) ok = %v, want %v", tc.in, ok, tc.ok)
		}
		if got != tc.wantN {
			t.Errorf("validateAIBaseURL(%q) = %q, want %q", tc.in, got, tc.wantN)
		}
	}
}

// TestCreateAIEndpointNormalizesBaseURL verifies the stored base URL has any
// pasted operation path stripped so request building cannot double it.
func TestCreateAIEndpointNormalizesBaseURL(t *testing.T) {
	r := setupAdminAIRouter(t)
	defer cleanupAIEndpoints()
	defer testutil.CloseDB(t)

	w := testutil.PostJSON(t, r, "/api/admin/ai-endpoints", map[string]any{
		"name":     "zen",
		"type":     "text",
		"base_url": "https://opencode.ai/zen/go/v1/chat/completions",
		"api_key":  "sk-test",
		"model":    "deepseek-v4-flash",
		"enabled":  true,
	})
	if w.Code != 201 {
		t.Fatalf("create status = %d, body %s", w.Code, w.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got := created["base_url"]; got != "https://opencode.ai/zen/go/v1" {
		t.Fatalf("stored base_url = %v, want normalized", got)
	}
}

// TestCreateAIEndpointRejectsInvalidBaseURL guards against storing a
// non-absolute URL that would fail at request time.
func TestCreateAIEndpointRejectsInvalidBaseURL(t *testing.T) {
	r := setupAdminAIRouter(t)
	defer cleanupAIEndpoints()
	defer testutil.CloseDB(t)

	w := testutil.PostJSON(t, r, "/api/admin/ai-endpoints", map[string]any{
		"name":     "bad",
		"type":     "text",
		"base_url": "not-a-url",
		"api_key":  "sk-test",
		"model":    "gpt-4o",
		"enabled":  true,
	})
	if w.Code != 400 {
		t.Fatalf("expected 400 for invalid base_url, got %d body %s", w.Code, w.Body.String())
	}
}
