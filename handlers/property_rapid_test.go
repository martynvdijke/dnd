package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"pgregory.net/rapid"

	"villum/db"
	"villum/handlers/testutil"
)

// slotsByLevel[level][slotIndex] = max slots of that level (full-caster progression)
var slotsByLevel = [21][9]int{
	0:  {0, 0, 0, 0, 0, 0, 0, 0, 0},
	1:  {2, 0, 0, 0, 0, 0, 0, 0, 0},
	2:  {3, 0, 0, 0, 0, 0, 0, 0, 0},
	3:  {4, 2, 0, 0, 0, 0, 0, 0, 0},
	4:  {4, 3, 0, 0, 0, 0, 0, 0, 0},
	5:  {4, 3, 2, 0, 0, 0, 0, 0, 0},
	6:  {4, 3, 3, 0, 0, 0, 0, 0, 0},
	7:  {4, 3, 3, 1, 0, 0, 0, 0, 0},
	8:  {4, 3, 3, 2, 0, 0, 0, 0, 0},
	9:  {4, 3, 3, 3, 1, 0, 0, 0, 0},
	10: {4, 3, 3, 3, 2, 0, 0, 0, 0},
	11: {4, 3, 3, 3, 2, 1, 0, 0, 0},
	12: {4, 3, 3, 3, 2, 1, 0, 0, 0},
	13: {4, 3, 3, 3, 2, 1, 1, 0, 0},
	14: {4, 3, 3, 3, 2, 1, 1, 0, 0},
	15: {4, 3, 3, 3, 2, 1, 1, 1, 0},
	16: {4, 3, 3, 3, 2, 1, 1, 1, 0},
	17: {4, 3, 3, 3, 2, 1, 1, 1, 1},
	18: {4, 3, 3, 3, 3, 1, 1, 1, 1},
	19: {4, 3, 3, 3, 3, 2, 1, 1, 1},
	20: {4, 3, 3, 3, 3, 2, 2, 1, 1},
}

func setSpellSlotsUsed(ctx context.Context, scID int64, used [9]int) error {
	u := db.Client.CharacterSpellcasting.UpdateOneID(scID)
	u = u.SetSlots1Used(used[0]).SetSlots2Used(used[1]).SetSlots3Used(used[2]).SetSlots4Used(used[3])
	u = u.SetSlots5Used(used[4]).SetSlots6Used(used[5]).SetSlots7Used(used[6]).SetSlots8Used(used[7]).SetSlots9Used(used[8])
	_, err := u.Save(ctx)
	return err
}

func readSpellSlotsUsed(ctx context.Context, scID int64) ([9]int, error) {
	sc, err := db.Client.CharacterSpellcasting.Get(ctx, scID)
	if err != nil {
		return [9]int{}, err
	}
	return [9]int{sc.Slots1Used, sc.Slots2Used, sc.Slots3Used, sc.Slots4Used, sc.Slots5Used, sc.Slots6Used, sc.Slots7Used, sc.Slots8Used, sc.Slots9Used}, nil
}

// 5.4: after a long rest all spell slots are recovered, for any level and any
// partial usage pattern; a short rest must NOT reset slots.
func TestPropertySpellSlotRecovery(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 100, 1, "Rest Prop", "Human", "Wizard")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/characters/:id/rest", DoRest)
	})

	sc, err := db.Client.CharacterSpellcasting.Create().
		SetCharacterID(100).SetAbility("").SetSaveDc(10).SetAttackBonus(0).
		Save(context.Background())
	if err != nil {
		t.Fatalf("create spellcasting: %v", err)
	}

	doRest := func(restType string) int {
		body, _ := json.Marshal(map[string]any{"rest_type": restType})
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/characters/100/rest", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(rec, req)
		return rec.Code
	}

	// short rest does NOT reset slots
	if err := setSpellSlotsUsed(context.Background(), sc.ID, [9]int{2, 1, 0, 0, 0, 0, 0, 0, 0}); err != nil {
		t.Fatalf("set used slots: %v", err)
	}
	if code := doRest("short"); code != http.StatusOK {
		t.Fatalf("short rest: expected 200, got %d", code)
	}
	got, err := readSpellSlotsUsed(context.Background(), sc.ID)
	if err != nil {
		t.Fatalf("read used slots: %v", err)
	}
	if got[0] != 2 || got[1] != 1 {
		t.Fatalf("short rest must not reset slots, got %v", got)
	}

	rapid.Check(t, func(t *rapid.T) {
		level := rapid.IntRange(1, 20).Draw(t, "level")
		var used [9]int
		for i := 0; i < 9; i++ {
			used[i] = rapid.IntRange(0, slotsByLevel[level][i]).Draw(t, fmt.Sprintf("used%d", i+1))
		}
		if err := setSpellSlotsUsed(context.Background(), sc.ID, used); err != nil {
			t.Fatalf("set used slots: %v", err)
		}
		if code := doRest("long"); code != http.StatusOK {
			t.Fatalf("long rest: expected 200, got %d", code)
		}
		after, err := readSpellSlotsUsed(context.Background(), sc.ID)
		if err != nil {
			t.Fatalf("read used slots: %v", err)
		}
		for i := 0; i < 9; i++ {
			if after[i] != 0 {
				t.Fatalf("level %d: slots%d used=%d after long rest (was %d)", level, i+1, after[i], used[i])
			}
		}
	})
}
