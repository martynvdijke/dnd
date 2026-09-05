package handlers

import (
	"errors"
	"strings"
	"testing"
)

func TestSanitizeErrorAdditional(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want func(string) bool
	}{
		{"nil-like empty msg", errors.New(""), func(s string) bool { return s == "" }},
		{"exactly 100 chars no trunc", errors.New(strings.Repeat("x", 100)), func(s string) bool { return len(s) == 100 && !strings.HasSuffix(s, "...") }},
		{"101 chars truncated", errors.New(strings.Repeat("y", 101)), func(s string) bool { return len(s) == 103 && strings.HasSuffix(s, "...") }},
		{"long with newline", errors.New(strings.Repeat("a", 90) + "\n" + strings.Repeat("b", 50)), func(s string) bool {
			return !strings.Contains(s, "\n") && strings.Contains(s, " ") && strings.HasSuffix(s, "...")
		}},
		{"multiple newlines", errors.New("a\nb\nc"), func(s string) bool { return s == "a b c" }},
		{"pii-like key leaked truncated", errors.New("api_key=sk-1234 secret token " + strings.Repeat("z", 200)), func(s string) bool { return len(s) == 103 && !strings.Contains(s, "\n") }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeError(tt.err)
			if !tt.want(got) {
				t.Errorf("sanitizeError(%q) = %q failed predicate", tt.err.Error(), got)
			}
		})
	}
}

func TestTruncateResponseAdditional(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want func(string) bool
	}{
		{"empty", "", func(s string) bool { return s == "" }},
		{"199 chars", strings.Repeat("a", 199), func(s string) bool { return len(s) == 199 }},
		{"200 chars", strings.Repeat("a", 200), func(s string) bool { return len(s) == 200 && !strings.HasSuffix(s, "...") }},
		{"201 chars", strings.Repeat("a", 201), func(s string) bool { return len(s) == 203 && strings.HasSuffix(s, "...") }},
		{"300 chars", strings.Repeat("b", 300), func(s string) bool { return len(s) == 203 }},
		{"unicode short", "hello 🌍", func(s string) bool { return s == "hello 🌍" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateResponse(tt.in)
			if !tt.want(got) {
				t.Errorf("truncateResponse len %d got %q len %d", len(tt.in), got, len(got))
			}
		})
	}
}
