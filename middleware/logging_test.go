package middleware

import (
	"testing"
	"time"
)

func TestNewLogBuffer(t *testing.T) {
	t.Run("default capacity", func(t *testing.T) {
		lb := NewLogBuffer(0)
		if lb.Cap() != 5000 {
			t.Errorf("expected capacity 5000, got %d", lb.Cap())
		}
	})

	t.Run("negative capacity defaults", func(t *testing.T) {
		lb := NewLogBuffer(-1)
		if lb.Cap() != 5000 {
			t.Errorf("expected capacity 5000, got %d", lb.Cap())
		}
	})

	t.Run("explicit capacity", func(t *testing.T) {
		lb := NewLogBuffer(100)
		if lb.Cap() != 100 {
			t.Errorf("expected capacity 100, got %d", lb.Cap())
		}
		if lb.Len() != 0 {
			t.Errorf("expected initial length 0, got %d", lb.Len())
		}
	})
}

func TestLogBufferAppendAndLen(t *testing.T) {
	lb := NewLogBuffer(5)

	for range 3 {
		lb.Append(LogEntry{Source: "test", Message: "msg"})
	}
	if lb.Len() != 3 {
		t.Errorf("expected length 3, got %d", lb.Len())
	}

	// Fill beyond capacity
	for range 5 {
		lb.Append(LogEntry{Source: "test", Message: "msg"})
	}
	if lb.Len() != 5 {
		t.Errorf("expected length 5 (capped), got %d", lb.Len())
	}
}

func TestLogBufferQuery(t *testing.T) {
	lb := NewLogBuffer(20)
	now := time.Now()

	lb.Append(LogEntry{Level: "info", Source: "ai", Message: "ai msg", Timestamp: now})
	lb.Append(LogEntry{Level: "error", Source: "email", Message: "email err", Timestamp: now})
	lb.Append(LogEntry{Level: "warn", Source: "ai", Message: "ai warn", Timestamp: now})
	lb.Append(LogEntry{Level: "info", Source: "backup", Message: "backup ok", Timestamp: now})

	t.Run("all entries newest first", func(t *testing.T) {
		results := lb.Query("", "", time.Time{}, 0)
		if len(results) != 4 {
			t.Errorf("expected 4 results, got %d", len(results))
		}
		if results[0].Message != "backup ok" {
			t.Errorf("expected newest first 'backup ok', got %q", results[0].Message)
		}
	})

	t.Run("filter by level", func(t *testing.T) {
		results := lb.Query("error", "", time.Time{}, 0)
		if len(results) != 1 {
			t.Errorf("expected 1 error result, got %d", len(results))
		}
		if results[0].Source != "email" {
			t.Errorf("expected source 'email', got %q", results[0].Source)
		}
	})

	t.Run("filter by source", func(t *testing.T) {
		results := lb.Query("", "ai", time.Time{}, 0)
		if len(results) != 2 {
			t.Errorf("expected 2 ai results, got %d", len(results))
		}
	})

	t.Run("filter by level and source", func(t *testing.T) {
		results := lb.Query("info", "ai", time.Time{}, 0)
		if len(results) != 1 {
			t.Errorf("expected 1 result, got %d", len(results))
		}
	})

	t.Run("limit results", func(t *testing.T) {
		results := lb.Query("", "", time.Time{}, 2)
		if len(results) != 2 {
			t.Errorf("expected 2 results, got %d", len(results))
		}
	})

	t.Run("filter by since", func(t *testing.T) {
		results := lb.Query("", "", now.Add(time.Second), 0)
		if len(results) != 0 {
			t.Errorf("expected 0 results when since is in the future, got %d", len(results))
		}
	})

	t.Run("empty buffer returns nil slice", func(t *testing.T) {
		empty := NewLogBuffer(10)
		results := empty.Query("", "", time.Time{}, 0)
		if len(results) != 0 {
			t.Errorf("expected 0 results from empty buffer, got %d", len(results))
		}
	})
}

func TestLogBufferSources(t *testing.T) {
	t.Run("returns sorted distinct sources", func(t *testing.T) {
		lb := NewLogBuffer(10)
		lb.Append(LogEntry{Source: "ai", Message: "a"})
		lb.Append(LogEntry{Source: "backup", Message: "b"})
		lb.Append(LogEntry{Source: "ai", Message: "c"})
		lb.Append(LogEntry{Source: "email", Message: "d"})

		sources := lb.Sources()
		expected := []string{"ai", "backup", "email"}
		if len(sources) != len(expected) {
			t.Fatalf("expected %d sources, got %d: %v", len(expected), len(sources), sources)
		}
		for i, s := range expected {
			if sources[i] != s {
				t.Errorf("expected source[%d]=%q, got %q", i, s, sources[i])
			}
		}
	})

	t.Run("empty buffer returns nil", func(t *testing.T) {
		lb := NewLogBuffer(10)
		if lb.Sources() != nil {
			t.Error("expected nil sources for empty buffer")
		}
	})

	t.Run("entries without source are not tracked", func(t *testing.T) {
		lb := NewLogBuffer(10)
		lb.Append(LogEntry{Source: "", Message: "no source"})
		if lb.Sources() != nil {
			t.Error("expected nil sources when only entries with empty source exist")
		}
	})
}

func TestLogBufferClear(t *testing.T) {
	lb := NewLogBuffer(10)
	lb.Append(LogEntry{Source: "ai", Message: "a"})
	lb.Append(LogEntry{Source: "email", Message: "b"})

	if lb.Len() != 2 {
		t.Errorf("expected length 2 before clear, got %d", lb.Len())
	}

	lb.Clear()

	if lb.Len() != 0 {
		t.Errorf("expected length 0 after clear, got %d", lb.Len())
	}
	if lb.Sources() != nil {
		t.Error("expected nil sources after clear")
	}

	// Should still accept entries after clear
	lb.Append(LogEntry{Source: "backup", Message: "after clear"})
	if lb.Len() != 1 {
		t.Errorf("expected length 1 after clear+append, got %d", lb.Len())
	}
}

func TestLogBufferCapAndLen(t *testing.T) {
	lb := NewLogBuffer(3)
	if lb.Cap() != 3 {
		t.Errorf("expected cap 3, got %d", lb.Cap())
	}

	// Fill to capacity
	lb.Append(LogEntry{Source: "a", Message: "1"})
	lb.Append(LogEntry{Source: "b", Message: "2"})
	lb.Append(LogEntry{Source: "c", Message: "3"})

	if lb.Len() != 3 {
		t.Errorf("expected len 3, got %d", lb.Len())
	}

	// Overflow - should keep capacity
	lb.Append(LogEntry{Source: "d", Message: "4"})
	if lb.Len() != 3 {
		t.Errorf("expected len 3 after overflow, got %d", lb.Len())
	}
}
