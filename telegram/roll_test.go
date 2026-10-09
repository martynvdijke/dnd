package telegram

import (
	"context"
	"strings"
	"testing"
)

func TestRunRoll(t *testing.T) {
	setupTelegramDB(t)
	defer func() {}()

	cases := []struct {
		name string
		args []string
		want string
	}{
		{"with expr", []string{"2d6+3"}, "2d6+3"},
		{"empty defaults to 1d20", nil, "1d20"},
		{"empty slice defaults", []string{}, "1d20"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := &cmdContext{ctx: context.Background(), chatID: 123, tgUserID: 999, args: tc.args}
			reply := runRoll(c)
			if reply.Text == "" {
				t.Fatalf("expected non-empty Text")
			}
			if !strings.Contains(reply.Text, tc.want) {
				t.Fatalf("expected %q in %q", tc.want, reply.Text)
			}
		})
	}
}

func TestRunRollNonsense(t *testing.T) {
	setupTelegramDB(t)
	c := &cmdContext{ctx: context.Background(), chatID: 123, tgUserID: 999, args: []string{"not-a-dice!!!"}}
	reply := runRoll(c)
	if reply.Text == "" {
		t.Fatalf("expected error Text for nonsense expression")
	}
}
