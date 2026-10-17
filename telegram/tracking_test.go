package telegram

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"villum/db"
	"villum/handlers/testutil"
)

func trackedContext(tgUID, chatID int64) *cmdContext {
	return &cmdContext{ctx: context.Background(), chatID: chatID, tgUserID: tgUID}
}

func lastSent(sent []string) string {
	if len(sent) == 0 {
		return ""
	}
	return sent[len(sent)-1]
}

func TestInventoryList(t *testing.T) {
	tgUID, chatID := setupActionCharacter(t, 20, 20, 0, 5, 14, "1d8", 5)
	defer testutil.CloseDB(t)
	db.DB.Exec(`INSERT INTO inventory (character_id, name, quantity, category, is_equipped, attunement)
		VALUES (1, 'Rope', 2, 'gear', 0, 0), (1, 'Longsword', 1, 'weapon', 1, 0)`)
	var sent []string
	recordingMock(t, &sent)
	HandleUpdate(context.Background(), messageUpdate(1, chatID, tgUID, "/inventory"))
	joined := strings.Join(sent, "\n")
	if !strings.Contains(joined, "Inventory") || !strings.Contains(joined, "Rope") || !strings.Contains(joined, "×2") {
		t.Fatalf("inventory listing missing items: %v", sent)
	}
	if !strings.Contains(joined, "Longsword") || !strings.Contains(joined, "equipped") {
		t.Fatalf("equipped marker missing: %v", sent)
	}
}

func TestInventoryEmpty(t *testing.T) {
	tgUID, chatID := setupActionCharacter(t, 20, 20, 0, 5, 14, "1d8", 5)
	defer testutil.CloseDB(t)
	var sent []string
	recordingMock(t, &sent)
	HandleUpdate(context.Background(), messageUpdate(1, chatID, tgUID, "/inventory"))
	if !strings.Contains(lastSent(sent), "Empty") {
		t.Fatalf("expected empty prompt, got %v", sent)
	}
}

func TestInventoryAliases(t *testing.T) {
	tgUID, chatID := setupActionCharacter(t, 20, 20, 0, 5, 14, "1d8", 5)
	defer testutil.CloseDB(t)
	db.DB.Exec(`INSERT INTO inventory (character_id, name, quantity, category) VALUES (1, 'Torch', 3, 'gear')`)
	for _, cmd := range []string{"/inv", "/bag"} {
		var sent []string
		recordingMock(t, &sent)
		HandleUpdate(context.Background(), messageUpdate(1, chatID, tgUID, cmd))
		if !strings.Contains(lastSent(sent), "Torch") {
			t.Fatalf("%s alias did not list inventory: %v", cmd, sent)
		}
	}
}

func TestInventoryEquipToggle(t *testing.T) {
	tgUID, chatID := setupActionCharacter(t, 20, 20, 0, 5, 14, "1d8", 5)
	defer testutil.CloseDB(t)
	db.DB.Exec(`INSERT INTO inventory (character_id, name, quantity, category) VALUES (1, 'Shield', 1, 'armor')`)
	var id int64
	if err := db.DB.QueryRow(`SELECT id FROM inventory WHERE name = 'Shield'`).Scan(&id); err != nil {
		t.Fatalf("lookup: %v", err)
	}
	c := trackedContext(tgUID, chatID)
	if _, ok := handleCallbackData(c, fmt.Sprintf("inv:eq:%d", id)); !ok {
		t.Fatal("equip callback not routed")
	}
	var equipped int
	db.DB.QueryRow(`SELECT is_equipped FROM inventory WHERE id = ?`, id).Scan(&equipped)
	if equipped != 1 {
		t.Fatalf("expected equipped=1, got %d", equipped)
	}
	handleInventoryCallback(c, fmt.Sprintf("inv:eq:%d", id))
	db.DB.QueryRow(`SELECT is_equipped FROM inventory WHERE id = ?`, id).Scan(&equipped)
	if equipped != 0 {
		t.Fatalf("expected equipped=0 after second toggle, got %d", equipped)
	}
}

func TestInventoryQuantityFloorAndIncrement(t *testing.T) {
	tgUID, chatID := setupActionCharacter(t, 20, 20, 0, 5, 14, "1d8", 5)
	defer testutil.CloseDB(t)
	db.DB.Exec(`INSERT INTO inventory (character_id, name, quantity, category) VALUES (1, 'Torch', 1, 'gear')`)
	var id int64
	db.DB.QueryRow(`SELECT id FROM inventory WHERE name = 'Torch'`).Scan(&id)
	c := trackedContext(tgUID, chatID)
	handleInventoryCallback(c, fmt.Sprintf("inv:inc:%d", id))
	var qty int
	db.DB.QueryRow(`SELECT quantity FROM inventory WHERE id = ?`, id).Scan(&qty)
	if qty != 2 {
		t.Fatalf("expected quantity 2, got %d", qty)
	}
	handleInventoryCallback(c, fmt.Sprintf("inv:dec:%d", id))
	db.DB.QueryRow(`SELECT quantity FROM inventory WHERE id = ?`, id).Scan(&qty)
	if qty != 1 {
		t.Fatalf("expected quantity 1, got %d", qty)
	}
	// floor: decrementing at one keeps it at one and does not remove the row
	handleInventoryCallback(c, fmt.Sprintf("inv:dec:%d", id))
	db.DB.QueryRow(`SELECT quantity FROM inventory WHERE id = ?`, id).Scan(&qty)
	if qty != 1 {
		t.Fatalf("expected quantity floor at 1, got %d", qty)
	}
	var count int
	db.DB.QueryRow(`SELECT COUNT(*) FROM inventory WHERE id = ?`, id).Scan(&count)
	if count != 1 {
		t.Fatal("decrement at floor must not remove the item")
	}
}

func TestInventoryRemoveConfirm(t *testing.T) {
	tgUID, chatID := setupActionCharacter(t, 20, 20, 0, 5, 14, "1d8", 5)
	defer testutil.CloseDB(t)
	db.DB.Exec(`INSERT INTO inventory (character_id, name, quantity, category) VALUES (1, 'Gem', 1, 'gem')`)
	var id int64
	db.DB.QueryRow(`SELECT id FROM inventory WHERE name = 'Gem'`).Scan(&id)
	c := trackedContext(tgUID, chatID)

	reply := handleInventoryCallback(c, fmt.Sprintf("inv:rm:%d", id))
	if !strings.Contains(reply.Text, "Remove") || !strings.Contains(reply.Text, "cannot be undone") {
		t.Fatalf("expected confirmation prompt, got %q", reply.Text)
	}
	var count int
	db.DB.QueryRow(`SELECT COUNT(*) FROM inventory WHERE id = ?`, id).Scan(&count)
	if count != 1 {
		t.Fatal("item must not be removed before confirmation")
	}
	// cancel keeps the item
	handleInventoryCallback(c, fmt.Sprintf("inv:rmx:%d", id))
	db.DB.QueryRow(`SELECT COUNT(*) FROM inventory WHERE id = ?`, id).Scan(&count)
	if count != 1 {
		t.Fatal("cancel must keep the item")
	}
	// confirm removes it
	handleInventoryCallback(c, fmt.Sprintf("inv:rmc:%d", id))
	db.DB.QueryRow(`SELECT COUNT(*) FROM inventory WHERE id = ?`, id).Scan(&count)
	if count != 0 {
		t.Fatal("confirm must remove the item")
	}
}

func TestInventoryMissingItemGone(t *testing.T) {
	tgUID, chatID := setupActionCharacter(t, 20, 20, 0, 5, 14, "1d8", 5)
	defer testutil.CloseDB(t)
	c := trackedContext(tgUID, chatID)
	reply := handleInventoryCallback(c, "inv:eq:9999")
	if !strings.Contains(reply.Text, "already gone") {
		t.Fatalf("expected gone notice, got %q", reply.Text)
	}
}

func TestAddItem(t *testing.T) {
	tgUID, chatID := setupActionCharacter(t, 20, 20, 0, 5, 14, "1d8", 5)
	defer testutil.CloseDB(t)
	var sent []string
	recordingMock(t, &sent)
	HandleUpdate(context.Background(), messageUpdate(1, chatID, tgUID, "/additem"))
	if !strings.Contains(lastSent(sent), "Usage") {
		t.Fatalf("expected usage without a name, got %v", sent)
	}
	var count int
	db.DB.QueryRow(`SELECT COUNT(*) FROM inventory WHERE character_id = 1`).Scan(&count)
	if count != 0 {
		t.Fatal("no item should be added without a name")
	}

	HandleUpdate(context.Background(), messageUpdate(2, chatID, tgUID, "/additem Shield +1 2"))
	var name, category string
	var qty int
	if err := db.DB.QueryRow(`SELECT name, quantity, category FROM inventory WHERE character_id = 1`).Scan(&name, &qty, &category); err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if name != "Shield +1" || qty != 2 || category != "gear" {
		t.Fatalf("unexpected item %q ×%d (%s)", name, qty, category)
	}

	HandleUpdate(context.Background(), messageUpdate(3, chatID, tgUID, "/additem Torch"))
	db.DB.QueryRow(`SELECT quantity FROM inventory WHERE name = 'Torch'`).Scan(&qty)
	if qty != 1 {
		t.Fatalf("default quantity should be 1, got %d", qty)
	}
}

func TestSpellbookListAndToggles(t *testing.T) {
	tgUID, chatID := setupActionCharacter(t, 20, 20, 0, 5, 14, "1d8", 5)
	defer testutil.CloseDB(t)
	db.DB.Exec(`INSERT INTO spells (character_id, name, level, prepared, always_prepared)
		VALUES (1, 'Fire Bolt', 0, 0, 0), (1, 'Shield', 1, 0, 0), (1, 'Oath Smite', 1, 0, 1)`)
	var sent []string
	recordingMock(t, &sent)
	HandleUpdate(context.Background(), messageUpdate(1, chatID, tgUID, "/spells"))
	joined := strings.Join(sent, "\n")
	for _, want := range []string{"Spellbook", "Cantrips", "Fire Bolt", "Level 1", "Shield", "Oath Smite", "★"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("spellbook missing %q: %v", want, sent)
		}
	}

	var fireID, shieldID, oathID int64
	db.DB.QueryRow(`SELECT id FROM spells WHERE name = 'Fire Bolt'`).Scan(&fireID)
	db.DB.QueryRow(`SELECT id FROM spells WHERE name = 'Shield'`).Scan(&shieldID)
	db.DB.QueryRow(`SELECT id FROM spells WHERE name = 'Oath Smite'`).Scan(&oathID)
	c := trackedContext(tgUID, chatID)

	handleSpellCallback(c, fmt.Sprintf("spl:t:%d", shieldID))
	var prepared int
	db.DB.QueryRow(`SELECT prepared FROM spells WHERE id = ?`, shieldID).Scan(&prepared)
	if prepared != 1 {
		t.Fatalf("expected Shield prepared, got %d", prepared)
	}

	if reply := handleSpellCallback(c, fmt.Sprintf("spl:t:%d", fireID)); !strings.Contains(reply.Text, "Cantrips") {
		t.Fatalf("cantrip toggle should be refused, got %q", reply.Text)
	}
	db.DB.QueryRow(`SELECT prepared FROM spells WHERE id = ?`, fireID).Scan(&prepared)
	if prepared != 0 {
		t.Fatal("cantrip prepared flag must not change")
	}

	if reply := handleSpellCallback(c, fmt.Sprintf("spl:t:%d", oathID)); !strings.Contains(reply.Text, "always prepared") {
		t.Fatalf("always-prepared toggle should be refused, got %q", reply.Text)
	}
}

func TestPrepareByArgument(t *testing.T) {
	tgUID, chatID := setupActionCharacter(t, 20, 20, 0, 5, 14, "1d8", 5)
	defer testutil.CloseDB(t)
	db.DB.Exec(`INSERT INTO spells (character_id, name, level, prepared, always_prepared)
		VALUES (1, 'Fireball', 3, 0, 0), (1, 'Light', 0, 0, 0), (1, 'Oath Smite', 1, 0, 1)`)
	var sent []string
	recordingMock(t, &sent)

	HandleUpdate(context.Background(), messageUpdate(1, chatID, tgUID, "/prepare Fireball"))
	var prepared int
	db.DB.QueryRow(`SELECT prepared FROM spells WHERE name = 'Fireball'`).Scan(&prepared)
	if prepared != 1 {
		t.Fatalf("expected Fireball prepared, got %d", prepared)
	}
	HandleUpdate(context.Background(), messageUpdate(2, chatID, tgUID, "/unprepare fireball"))
	db.DB.QueryRow(`SELECT prepared FROM spells WHERE name = 'Fireball'`).Scan(&prepared)
	if prepared != 0 {
		t.Fatalf("expected Fireball unprepared (case-insensitive), got %d", prepared)
	}

	HandleUpdate(context.Background(), messageUpdate(3, chatID, tgUID, "/prepare UnknownXYZ"))
	if !strings.Contains(lastSent(sent), "No spell named") {
		t.Fatalf("unknown spell should be refused, got %v", sent)
	}
	HandleUpdate(context.Background(), messageUpdate(4, chatID, tgUID, "/prepare Light"))
	if !strings.Contains(lastSent(sent), "Cantrips") {
		t.Fatalf("cantrip preparation should be refused, got %v", sent)
	}
	HandleUpdate(context.Background(), messageUpdate(5, chatID, tgUID, "/prepare Oath Smite"))
	if !strings.Contains(lastSent(sent), "always prepared") {
		t.Fatalf("always-prepared spell should be refused, got %v", sent)
	}
	HandleUpdate(context.Background(), messageUpdate(6, chatID, tgUID, "/prepare"))
	if !strings.Contains(lastSent(sent), "Usage") {
		t.Fatalf("missing spell argument should show usage, got %v", sent)
	}
}

func TestConditionsAddListRemove(t *testing.T) {
	tgUID, chatID := setupActionCharacter(t, 20, 20, 0, 5, 14, "1d8", 5)
	defer testutil.CloseDB(t)
	var sent []string
	recordingMock(t, &sent)

	HandleUpdate(context.Background(), messageUpdate(1, chatID, tgUID, "/condition Prone"))
	var count int
	db.DB.QueryRow(`SELECT COUNT(*) FROM character_conditions WHERE character_id = 1 AND name = 'Prone'`).Scan(&count)
	if count != 1 {
		t.Fatalf("expected Prone condition, got %d", count)
	}

	HandleUpdate(context.Background(), messageUpdate(2, chatID, tgUID, "/condition"))
	if !strings.Contains(lastSent(sent), "Usage") {
		t.Fatalf("empty condition should show usage, got %v", sent)
	}

	HandleUpdate(context.Background(), messageUpdate(3, chatID, tgUID, "/condition "+strings.Repeat("x", 101)))
	if !strings.Contains(lastSent(sent), "too long") {
		t.Fatalf("over-long condition should be refused, got %v", sent)
	}

	HandleUpdate(context.Background(), messageUpdate(4, chatID, tgUID, "/conditions"))
	if !strings.Contains(lastSent(sent), "Prone") {
		t.Fatalf("conditions list should include Prone, got %v", sent)
	}

	var id int64
	db.DB.QueryRow(`SELECT id FROM character_conditions WHERE name = 'Prone'`).Scan(&id)
	c := trackedContext(tgUID, chatID)
	handleConditionCallback(c, fmt.Sprintf("cond:rm:%d", id))
	db.DB.QueryRow(`SELECT COUNT(*) FROM character_conditions WHERE id = ?`, id).Scan(&count)
	if count != 0 {
		t.Fatal("condition should be removed")
	}
}

func TestFeaturesReadOnly(t *testing.T) {
	tgUID, chatID := setupActionCharacter(t, 20, 20, 0, 5, 14, "1d8", 5)
	defer testutil.CloseDB(t)
	db.DB.Exec(`INSERT INTO character_features (character_id, name, source)
		VALUES (1, 'Darkvision', 'Race'), (1, 'Second Wind', 'Fighter')`)
	var sent []string
	recordingMock(t, &sent)
	HandleUpdate(context.Background(), messageUpdate(1, chatID, tgUID, "/features"))
	joined := strings.Join(sent, "\n")
	if !strings.Contains(joined, "Darkvision") || !strings.Contains(joined, "Race") || !strings.Contains(joined, "Second Wind") {
		t.Fatalf("features listing incomplete: %v", sent)
	}
	if strings.Contains(joined, "Remove") {
		t.Fatalf("features must be read-only: %v", sent)
	}

	db.DB.Exec(`DELETE FROM character_features WHERE character_id = 1`)
	sent = nil
	HandleUpdate(context.Background(), messageUpdate(2, chatID, tgUID, "/features"))
	if !strings.Contains(lastSent(sent), "None") {
		t.Fatalf("expected empty features notice, got %v", sent)
	}
}

func TestMoneyShowAddClamp(t *testing.T) {
	tgUID, chatID := setupActionCharacter(t, 20, 20, 0, 5, 14, "1d8", 5)
	defer testutil.CloseDB(t)
	var sent []string
	recordingMock(t, &sent)

	HandleUpdate(context.Background(), messageUpdate(1, chatID, tgUID, "/money"))
	if !strings.Contains(lastSent(sent), "Coin") || !strings.Contains(lastSent(sent), "0 gp") {
		t.Fatalf("expected zeroed currency, got %v", sent)
	}

	HandleUpdate(context.Background(), messageUpdate(2, chatID, tgUID, "/money +10 gp"))
	var gp int
	db.DB.QueryRow(`SELECT gp FROM character_currency WHERE character_id = 1`).Scan(&gp)
	if gp != 10 {
		t.Fatalf("expected gp=10, got %d", gp)
	}

	HandleUpdate(context.Background(), messageUpdate(3, chatID, tgUID, "/money -100 gp"))
	db.DB.QueryRow(`SELECT gp FROM character_currency WHERE character_id = 1`).Scan(&gp)
	if gp != 0 {
		t.Fatalf("overspend should clamp to 0, got %d", gp)
	}

	HandleUpdate(context.Background(), messageUpdate(4, chatID, tgUID, "/money +10 xx"))
	if !strings.Contains(lastSent(sent), "Usage") {
		t.Fatalf("invalid denomination should show usage, got %v", sent)
	}
	db.DB.QueryRow(`SELECT gp FROM character_currency WHERE character_id = 1`).Scan(&gp)
	if gp != 0 {
		t.Fatalf("invalid denomination must not change coin, got %d", gp)
	}

	HandleUpdate(context.Background(), messageUpdate(5, chatID, tgUID, "/money +5"))
	if !strings.Contains(lastSent(sent), "Usage") {
		t.Fatalf("missing denomination should show usage, got %v", sent)
	}
}

func TestTrackingCallbackRouting(t *testing.T) {
	tgUID, chatID := setupActionCharacter(t, 20, 20, 0, 5, 14, "1d8", 5)
	defer testutil.CloseDB(t)
	c := trackedContext(tgUID, chatID)
	if _, ok := handleCallbackData(c, "inv:rmx:1"); !ok {
		t.Fatal("inventory callback not routed")
	}
	if _, ok := handleCallbackData(c, "spl:t:999"); !ok {
		t.Fatal("spell callback not routed")
	}
	if _, ok := handleCallbackData(c, "cond:rm:999"); !ok {
		t.Fatal("condition callback not routed")
	}
}

func TestTrackingUnclaimedIsGuided(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	if err := UpsertIdentity(1, 999, 999, "admin"); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	var sent []string
	recordingMock(t, &sent)
	HandleUpdate(context.Background(), messageUpdate(1, 999, 999, "/inventory"))
	if len(sent) == 0 || !strings.Contains(sent[0], "/claim") {
		t.Fatalf("unclaimed user should be guided to /claim, got %v", sent)
	}
}

func TestTrackingRejectsNonEditable(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "owner", "admin")
	testutil.SeedUser(t, 2, "intruder", "user")
	testutil.SeedCharacter(t, 1, 1, "Aria", "Elf", "Ranger")
	db.DB.Exec(`INSERT INTO inventory (character_id, name, quantity, category) VALUES (1, 'Secret Gem', 1, 'gear')`)
	tgUID, chatID := int64(222), int64(222)
	if err := UpsertIdentity(2, tgUID, chatID, "intruder"); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	// Bypass claimByID so the intruder holds a claim on a character they cannot edit.
	db.DB.Exec(`INSERT OR REPLACE INTO telegram_character_claims (telegram_user_id, character_id, created_at) VALUES (?, 1, datetime('now'))`, tgUID)

	var sent []string
	recordingMock(t, &sent)
	HandleUpdate(context.Background(), messageUpdate(1, chatID, tgUID, "/inventory"))
	if strings.Contains(strings.Join(sent, "\n"), "Secret Gem") {
		t.Fatalf("foreign inventory leaked: %v", sent)
	}

	var id int64
	db.DB.QueryRow(`SELECT id FROM inventory WHERE name = 'Secret Gem'`).Scan(&id)
	c := trackedContext(tgUID, chatID)
	handleInventoryCallback(c, fmt.Sprintf("inv:eq:%d", id))
	var equipped int
	db.DB.QueryRow(`SELECT is_equipped FROM inventory WHERE id = ?`, id).Scan(&equipped)
	if equipped != 0 {
		t.Fatal("mutation on a non-editable character must not change the row")
	}
}
