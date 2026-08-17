package handlers

import "testing"

func TestNormalizedAutoSaveInterval(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want int
	}{
		{"default", nil, 12},
		{"fraction", 12.5, 12},
		{"lower bound", 1.0, 5},
		{"upper bound", 999.0, 300},
		{"valid", 45.0, 45},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizedAutoSaveInterval(tt.in); got != tt.want {
				t.Fatalf("normalizedAutoSaveInterval(%v) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}
