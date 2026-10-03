package ai

import (
	"fmt"
	"strings"
	"testing"
)

func TestSystemPrompts(t *testing.T) {
	a := SystemPromptForCompendiumQA("ctx")
	if !strings.Contains(a, "CONTEXT") || !strings.Contains(a, "[1]") {
		t.Fatalf("prompt missing contract %q", a)
	}
	b := SystemPromptForCampaignQA("ctx2")
	if !strings.Contains(b, "CONTEXT") {
		t.Fatalf("campaign prompt wrong %q", b)
	}
}

func TestNormalizeAndValidate(t *testing.T) {
	u, ok := ValidateAIBaseURL("https://api.openai.com/v1/chat/completions")
	if !ok || u != "https://api.openai.com/v1" {
		t.Fatalf("normalize failed %q %v", u, ok)
	}
	u2, ok2 := ValidateAIBaseURL("http://example.com/")
	if !ok2 || u2 != "http://example.com" {
		t.Fatalf("trim slash %q", u2)
	}
	if _, ok := ValidateAIBaseURL("not-a-url"); ok {
		t.Fatalf("should reject")
	}
}

func TestSanitize(t *testing.T) {
	if SanitizeError(fmt.Errorf("hello\nworld")) != "hello world" {
		t.Fatalf("sanitize")
	}
	if TruncateResponse(strings.Repeat("a", 300)) != strings.Repeat("a", 200)+"..." {
		t.Fatalf("truncate")
	}
}
