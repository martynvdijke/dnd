package ai

import "testing"

func TestAIHelpers(t *testing.T) {
	_ = GenerateSessionID()
	if ResolveSessionID("abc") != "abc" {
		t.Fatal("resolve")
	}
	if ResolveSessionID("  abc  ") != "abc" {
		t.Fatal("trim")
	}
	_ = ResolveSessionID("")
	_ = NormalizeAIBaseURL("https://example.com/v1/chat/completions")
	_ = AIRequestURL("https://example.com/v1", "/chat/completions")
	_ = NewAIClient(TestTimeout)
	SetAppVersion("1.0")
	_ = AppVersion
}
