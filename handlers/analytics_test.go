package handlers

import (
	"testing"

	"villum/db"
	"villum/handlers/testutil"
)

func TestMonthsBetween(t *testing.T) {
	cases := []struct {
		first, last string
		want        int
	}{
		{"2026-01-05", "2026-03-20", 2},
		{"2026-01-05", "2026-01-20", 1},
		{"2026-01-05", "2027-01-05", 12},
		{"", "", 1},
		{"not-a-date", "2026-01-01", 1},
	}
	for _, tc := range cases {
		if got := monthsBetween(tc.first, tc.last); got != tc.want {
			t.Errorf("monthsBetween(%q,%q)=%d want %d", tc.first, tc.last, got, tc.want)
		}
	}
}

func TestLoadCharacterStatsAnalytics(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 10, 1, "Analyst", "Human", "Wizard")

	dates := []string{"2026-01-05", "2026-01-20", "2026-02-05", "2026-02-20", "2026-03-05", "2026-03-20"}
	for _, d := range dates {
		if _, err := db.DB.Exec("INSERT INTO sessions(character_id, session_date) VALUES(?,?)", 10, d); err != nil {
			t.Fatalf("seed session: %v", err)
		}
	}
	rolls := []struct {
		expr, result string
		total        int
	}{
		{"1d20", "1d20: [20] = 20", 20},
		{"1d20", "1d20: [1] = 1", 1},
		{"2d6", "2d6: [20, 1] = 21", 21},
	}
	for _, r := range rolls {
		if _, err := db.DB.Exec("INSERT INTO dice_rolls(user_id, character_id, expression, result, total) VALUES(1,10,?,?,?)", r.expr, r.result, r.total); err != nil {
			t.Fatalf("seed roll: %v", err)
		}
	}

	stats, err := loadCharacterStats(db.DB, 10)
	if err != nil {
		t.Fatalf("loadCharacterStats: %v", err)
	}
	if stats.SessionCount != 6 {
		t.Errorf("SessionCount=%d want 6", stats.SessionCount)
	}
	if stats.SessionsPerMonth != 3 {
		t.Errorf("SessionsPerMonth=%v want 3", stats.SessionsPerMonth)
	}
	if stats.DiceRolls.Natural20s != 1 {
		t.Errorf("Natural20s=%d want 1", stats.DiceRolls.Natural20s)
	}
	if stats.DiceRolls.Natural1s != 1 {
		t.Errorf("Natural1s=%d want 1", stats.DiceRolls.Natural1s)
	}
}
