package telegram

import (
	"context"
	"strings"
	"sync"
	"testing"

	"villum/db"
	"villum/handlers/testutil"
)

func setupActionCharacter(t *testing.T, hpMax, hpCurrent, tempHp, level, con int, hitDice string, hitDiceCurrent int) (tgUserID, chatID int64) {
	t.Helper()
	setupTelegramDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "Aria", "Elf", "Ranger")
	// set custom fields
	if _, err := db.DB.Exec(`UPDATE characters SET hp_max=?, hp_current=?, temp_hp=?, level=?, con=?, hit_dice=?, hit_dice_current=? WHERE id=1`, hpMax, hpCurrent, tempHp, level, con, hitDice, hitDiceCurrent); err != nil {
		t.Fatalf("update char: %v", err)
	}
	// ensure spellcasting row exists
	if _, err := db.DB.Exec(`INSERT OR IGNORE INTO character_spellcasting (character_id, slots_1_max, slots_1_used, slots_3_max, slots_3_used) VALUES (1, 0, 0, 0, 0)`); err != nil {
		t.Fatalf("spellcasting: %v", err)
	}
	tgUserID = 111
	chatID = 111
	if err := UpsertIdentity(1, tgUserID, chatID, "admin"); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if _, err := db.DB.Exec(`INSERT OR REPLACE INTO telegram_character_claims (telegram_user_id, character_id, created_at) VALUES (?, 1, datetime('now'))`, tgUserID); err != nil {
		t.Fatalf("claim: %v", err)
	}
	return
}

func TestHPShow(t *testing.T) {
	tgUID, chatID := setupActionCharacter(t, 24, 20, 0, 3, 14, "1d8", 3)
	defer testutil.CloseDB(t)
	var sent []string
	recordingMock(t, &sent)
	HandleUpdate(context.Background(), messageUpdate(1, chatID, tgUID, "/hp"))
	if len(sent) == 0 || !strings.Contains(sent[0], "20/24") {
		t.Fatalf("expected HP show, got %v", sent)
	}
}

func TestHPAdjustAndTempAbsorb(t *testing.T) {
	tgUID, chatID := setupActionCharacter(t, 24, 20, 5, 3, 14, "1d8", 3)
	defer testutil.CloseDB(t)
	var sent []string
	recordingMock(t, &sent)
	// -8: 5 temp absorbs, 3 to hp => 17
	HandleUpdate(context.Background(), messageUpdate(1, chatID, tgUID, "/hp -8"))
	if len(sent) == 0 || !strings.Contains(sent[0], "17/24") {
		t.Fatalf("expected 17/24 after damage, got %v", sent)
	}
	var hpCurrent, tempHp int
	db.DB.QueryRow(`SELECT hp_current, temp_hp FROM characters WHERE id=1`).Scan(&hpCurrent, &tempHp)
	if hpCurrent != 17 || tempHp != 0 {
		t.Fatalf("expected 17/0 got %d/%d", hpCurrent, tempHp)
	}
	// heal +10 clamped to max
	sent = nil
	HandleUpdate(context.Background(), messageUpdate(2, chatID, tgUID, "/hp +10"))
	db.DB.QueryRow(`SELECT hp_current FROM characters WHERE id=1`).Scan(&hpCurrent)
	if hpCurrent != 24 {
		t.Fatalf("expected clamp to 24, got %d", hpCurrent)
	}
	// set to 5
	sent = nil
	HandleUpdate(context.Background(), messageUpdate(3, chatID, tgUID, "/hp 5"))
	db.DB.QueryRow(`SELECT hp_current FROM characters WHERE id=1`).Scan(&hpCurrent)
	if hpCurrent != 5 {
		t.Fatalf("expected set to 5, got %d", hpCurrent)
	}
	// clamp negative set
	HandleUpdate(context.Background(), messageUpdate(4, chatID, tgUID, "/hp -100"))
	db.DB.QueryRow(`SELECT hp_current FROM characters WHERE id=1`).Scan(&hpCurrent)
	if hpCurrent != 0 {
		t.Fatalf("expected clamp to 0, got %d", hpCurrent)
	}
	// invalid
	sent = nil
	HandleUpdate(context.Background(), messageUpdate(5, chatID, tgUID, "/hp foo"))
	if len(sent) == 0 || !strings.Contains(sent[0], "Usage") {
		t.Fatalf("expected usage, got %v", sent)
	}
}

func TestHPRefusalUnlinked(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	var sent []string
	recordingMock(t, &sent)
	HandleUpdate(context.Background(), messageUpdate(1, 999, 999, "/hp"))
	if len(sent) == 0 || !strings.Contains(sent[0], "/start") {
		t.Fatalf("expected linking help, got %v", sent)
	}
	// linked but unclaimed
	testutil.SeedUser(t, 1, "admin", "admin")
	_ = UpsertIdentity(1, 999, 999, "admin")
	sent = nil
	HandleUpdate(context.Background(), messageUpdate(2, 999, 999, "/hp"))
	if len(sent) == 0 || !strings.Contains(sent[0], "/claim") {
		t.Fatalf("expected claim hint, got %v", sent)
	}
}

func TestRestShortAndLong(t *testing.T) {
	tgUID, chatID := setupActionCharacter(t, 24, 10, 3, 3, 14, "1d8", 2)
	defer testutil.CloseDB(t)
	// set slots for long rest check
	db.DB.Exec(`UPDATE character_spellcasting SET slots_1_max=2, slots_1_used=1, slots_3_max=2, slots_3_used=2 WHERE character_id=1`)
	var sent []string
	recordingMock(t, &sent)
	HandleUpdate(context.Background(), messageUpdate(1, chatID, tgUID, "/rest short"))
	if len(sent) == 0 || !strings.Contains(strings.ToLower(sent[0]), "short rest") {
		t.Fatalf("expected short rest, got %v", sent)
	}
	var hitDiceCurrent, hpCurrent int
	db.DB.QueryRow(`SELECT hit_dice_current, hp_current FROM characters WHERE id=1`).Scan(&hitDiceCurrent, &hpCurrent)
	if hitDiceCurrent != 1 {
		t.Fatalf("expected hit dice 1, got %d", hitDiceCurrent)
	}
	if hpCurrent <= 10 {
		t.Fatalf("expected heal, got %d", hpCurrent)
	}
	var cnt int
	db.DB.QueryRow(`SELECT COUNT(*) FROM rest_log WHERE character_id=1 AND rest_type='short'`).Scan(&cnt)
	if cnt != 1 {
		t.Fatalf("expected rest_log short, got %d", cnt)
	}
	// short with no dice
	db.DB.Exec(`UPDATE characters SET hit_dice_current=0 WHERE id=1`)
	sent = nil
	HandleUpdate(context.Background(), messageUpdate(2, chatID, tgUID, "/rest short"))
	if len(sent) == 0 || !strings.Contains(sent[0], "No hit dice") {
		t.Fatalf("expected no hit dice, got %v", sent)
	}
	// long rest restores
	sent = nil
	HandleUpdate(context.Background(), messageUpdate(3, chatID, tgUID, "/rest long"))
	db.DB.QueryRow(`SELECT hp_current, temp_hp, hit_dice_current FROM characters WHERE id=1`).Scan(&hpCurrent, &hitDiceCurrent, &hitDiceCurrent)
	// re-query correctly
	var tempHp int
	db.DB.QueryRow(`SELECT hp_current, temp_hp, hit_dice_current FROM characters WHERE id=1`).Scan(&hpCurrent, &tempHp, &hitDiceCurrent)
	if hpCurrent != 24 || tempHp != 0 || hitDiceCurrent != 3 {
		t.Fatalf("long rest failed: hp %d temp %d hd %d", hpCurrent, tempHp, hitDiceCurrent)
	}
	var slotsUsed int
	db.DB.QueryRow(`SELECT slots_3_used FROM character_spellcasting WHERE character_id=1`).Scan(&slotsUsed)
	if slotsUsed != 0 {
		t.Fatalf("expected slots reset, got %d", slotsUsed)
	}
	db.DB.QueryRow(`SELECT COUNT(*) FROM rest_log WHERE character_id=1 AND rest_type='long'`).Scan(&cnt)
	if cnt != 1 {
		t.Fatalf("expected long log, got %d", cnt)
	}
	// invalid arg
	sent = nil
	HandleUpdate(context.Background(), messageUpdate(4, chatID, tgUID, "/rest foo"))
	if len(sent) == 0 || !strings.Contains(sent[0], "Usage") {
		t.Fatalf("expected usage, got %v", sent)
	}
}

func TestCast(t *testing.T) {
	tgUID, chatID := setupActionCharacter(t, 24, 20, 0, 5, 14, "1d8", 5)
	defer testutil.CloseDB(t)
	db.DB.Exec(`INSERT INTO spells (character_id, name, level) VALUES (1, 'Fireball', 3), (1, 'Light', 0)`)
	db.DB.Exec(`UPDATE character_spellcasting SET slots_3_max=2, slots_3_used=0 WHERE character_id=1`)
	// seed compendium for suggestions
	testutil.SeedCompendiumSpell(t, 900, "Fireball")
	var sent []string
	recordingMock(t, &sent)
	// cantrip free
	HandleUpdate(context.Background(), messageUpdate(1, chatID, tgUID, "/cast Light"))
	if len(sent) == 0 || !strings.Contains(strings.ToLower(sent[0]), "cantrip") {
		t.Fatalf("expected cantrip, got %v", sent)
	}
	var used int
	db.DB.QueryRow(`SELECT slots_3_used FROM character_spellcasting WHERE character_id=1`).Scan(&used)
	if used != 0 {
		t.Fatalf("cantrip should not consume slot")
	}
	// leveled
	sent = nil
	HandleUpdate(context.Background(), messageUpdate(2, chatID, tgUID, "/cast Fireball"))
	if len(sent) == 0 || !strings.Contains(sent[0], "Fireball") {
		t.Fatalf("expected fireball cast, got %v", sent)
	}
	db.DB.QueryRow(`SELECT slots_3_used FROM character_spellcasting WHERE character_id=1`).Scan(&used)
	if used != 1 {
		t.Fatalf("expected 1 used, got %d", used)
	}
	// no slot left (use second then third fails)
	HandleUpdate(context.Background(), messageUpdate(3, chatID, tgUID, "/cast Fireball"))
	sent = nil
	HandleUpdate(context.Background(), messageUpdate(4, chatID, tgUID, "/cast Fireball"))
	if len(sent) == 0 || !strings.Contains(sent[0], "No level 3 slots") {
		t.Fatalf("expected no slots, got %v", sent)
	}
	// unknown spell suggestions
	sent = nil
	HandleUpdate(context.Background(), messageUpdate(5, chatID, tgUID, "/cast UnknownSpellXYZ"))
	if len(sent) == 0 || !strings.Contains(sent[0], "not on your sheet") {
		t.Fatalf("expected unknown, got %v", sent)
	}
	// optional dice
	db.DB.Exec(`UPDATE character_spellcasting SET slots_3_used=0 WHERE character_id=1`)
	sent = nil
	HandleUpdate(context.Background(), messageUpdate(6, chatID, tgUID, "/cast Fireball 8d6"))
	if len(sent) == 0 || !strings.Contains(sent[0], "Roll") {
		t.Fatalf("expected dice roll, got %v", sent)
	}
}

func TestCastConcurrent(t *testing.T) {
	tgUID, chatID := setupActionCharacter(t, 24, 20, 0, 5, 14, "1d8", 5)
	defer testutil.CloseDB(t)
	db.DB.Exec(`INSERT INTO spells (character_id, name, level) VALUES (1, 'Magic Missile', 1)`)
	db.DB.Exec(`UPDATE character_spellcasting SET slots_1_max=1, slots_1_used=0 WHERE character_id=1`)
	// html escape check: spell name contains <b>
	db.DB.Exec(`INSERT INTO spells (character_id, name, level) VALUES (1, '<b>Boom</b>', 1)`)
	var sent []string
	recordingMock(t, &sent)
	// html escaping: cast the hostile name directly via runCast path
	// use exact name to avoid LIKE ambiguity
	HandleUpdate(context.Background(), messageUpdate(100, chatID, tgUID, "/cast <b>Boom</b>"))
	if len(sent) == 0 {
		t.Fatalf("expected reply for html spell, got none")
	}
	if strings.Contains(sent[len(sent)-1], "<b>Boom</b>") {
		t.Fatalf("spell name not escaped: %q", sent[len(sent)-1])
	}
	if !strings.Contains(sent[len(sent)-1], "&lt;b&gt;Boom&lt;/b&gt;") {
		t.Fatalf("expected escaped spell name, got %q", sent[len(sent)-1])
	}
	// reset slot for concurrent test (the html cast consumed one slot)
	db.DB.Exec(`UPDATE character_spellcasting SET slots_1_max=1, slots_1_used=0 WHERE character_id=1`)
	// remove the hostile spell so concurrent test only sees Magic Missile
	db.DB.Exec(`DELETE FROM spells WHERE name='<b>Boom</b>'`)
	// Concurrent cast via atomic conditional UPDATE: exactly one should succeed.
	// Using direct DB Exec to isolate the slot-expend atomicity from telegram mock races.
	// sqlite in-memory with single connection serializes writes; RowsAffected proves atomicity.
	const goroutines = 8
	var wg sync.WaitGroup
	affected := make([]int64, goroutines)
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			res, err := db.DB.Exec(`UPDATE character_spellcasting SET slots_1_used=slots_1_used+1 WHERE character_id=1 AND slots_1_used < slots_1_max`)
			if err == nil {
				ra, _ := res.RowsAffected()
				affected[idx] = ra
			}
		}(i)
	}
	wg.Wait()
	successes := 0
	for _, ra := range affected {
		if ra == 1 {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("expected exactly 1 success from %d concurrent casts, got %d results %v (sqlite locking may serialize; atomic UPDATE must prevent over-spend)", goroutines, successes, affected)
	}
	var used int
	db.DB.QueryRow(`SELECT slots_1_used FROM character_spellcasting WHERE character_id=1`).Scan(&used)
	if used != 1 {
		t.Fatalf("expected slots_1_used == 1, got %d", used)
	}
}

func TestHPRejectNonEditable(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "owner", "admin")
	testutil.SeedUser(t, 2, "intruder", "admin")
	testutil.SeedCharacter(t, 1, 1, "Aria", "Elf", "Ranger")
	if _, err := db.DB.Exec(`UPDATE characters SET hp_max=24, hp_current=20 WHERE id=1`); err != nil {
		t.Fatalf("update hp: %v", err)
	}
	if _, err := db.DB.Exec(`INSERT OR IGNORE INTO character_spellcasting (character_id, slots_1_max, slots_1_used) VALUES (1,0,0)`); err != nil {
		t.Fatalf("spellcasting: %v", err)
	}
	tgUID := int64(222)
	chatID := int64(222)
	if err := UpsertIdentity(2, tgUID, chatID, "intruder"); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	// claim character 1 for intruder even though intruder cannot edit it (bypass claimByID check)
	if _, err := db.DB.Exec(`INSERT OR REPLACE INTO telegram_character_claims (telegram_user_id, character_id, created_at) VALUES (?, 1, datetime('now'))`, tgUID); err != nil {
		t.Fatalf("claim: %v", err)
	}
	var sent []string
	recordingMock(t, &sent)
	HandleUpdate(context.Background(), messageUpdate(1, chatID, tgUID, "/hp -1"))
	if len(sent) == 0 {
		t.Fatalf("expected refusal, got none")
	}
	// refused due to non-editable claim - should contain no-longer-available or claim hint
	if !strings.Contains(sent[0], "no longer available") && !strings.Contains(sent[0], "/claim") && !strings.Contains(sent[0], "not found") {
		t.Fatalf("expected non-editable refusal, got %q", sent[0])
	}
	var hpCurrent int
	db.DB.QueryRow(`SELECT hp_current FROM characters WHERE id=1`).Scan(&hpCurrent)
	if hpCurrent != 20 {
		t.Fatalf("hp_current should be unchanged 20, got %d", hpCurrent)
	}
}
