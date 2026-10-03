package search

import "testing"

func TestBuildFTS5Query(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"", ""},
		{"  ", ""},
		{"fire", `"fire"*`},
		{"fire ball", `"fire"* AND "ball"*`},
		{"  fire   ball  ", `"fire"* AND "ball"*`},
		{`a"b`, `"a""b"*`},
		{`x y z`, `"x"* AND "y"* AND "z"*`},
	}
	for _, tc := range tests {
		got := BuildFTS5Query(tc.in)
		if got != tc.want {
			t.Fatalf("BuildFTS5Query(%q)=%q want %q", tc.in, got, tc.want)
		}
	}
}

func TestDefaultRerank(t *testing.T) {
	in := []CompendiumResult{
		{Name: "Fireball", Type: "spell"},
		{Name: "Fire Bolt", Type: "spell"},
		{Name: "Fyreball", Type: "spell"},
	}
	out := DefaultRerank("fireball", in)
	if out[0].Name != "Fireball" {
		t.Fatalf("expected exact first, got %q", out[0].Name)
	}
	// prefix boost
	in2 := []CompendiumResult{{Name: "Fireball"}, {Name: "AFireball"}, {Name: "Fire"}}
	out2 := DefaultRerank("Fire", in2)
	if out2[0].Name != "Fire" {
		t.Fatalf("expected prefix/exact first got %v", out2[0].Name)
	}
	if len(out2) != 3 {
		t.Fatalf("len mismatch")
	}
	// empty
	if DefaultRerank("a", nil) != nil {
		// ok if returns nil
	}
}
