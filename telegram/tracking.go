package telegram

import (
	"fmt"
	"strconv"
	"strings"

	"villum/db"
)

// Tracking callbacks are compact "<prefix><action>:<id>" payloads so they stay
// well under Telegram's 64-byte callback-data limit:
//
//	inv:eq:<id>   toggle equipped
//	inv:inc:<id>  quantity +1
//	inv:dec:<id>  quantity -1 (floor 1)
//	inv:rm:<id>   ask to confirm removal
//	inv:rmc:<id>  confirm removal
//	inv:rmx:<id>  cancel removal
//	spl:t:<id>    toggle prepared
//	cond:rm:<id>  remove condition
const (
	maxConditionNameLen = 100
	maxTrackedItems     = 20
	moneyUsage          = "Usage: <code>/money</code> or <code>/money [+N|-N] [pp|gp|ep|sp|cp]</code>"
	addItemUsage        = "Usage: <code>/additem &lt;name&gt; [qty]</code>"
)

// parseCallbackAction splits "<prefix><action>:<id>" into its action and id.
func parseCallbackAction(data, prefix string) (string, int64, error) {
	rest := strings.TrimPrefix(data, prefix)
	i := strings.LastIndex(rest, ":")
	if i < 0 {
		return "", 0, fmt.Errorf("malformed callback %q", data)
	}
	id, err := strconv.ParseInt(rest[i+1:], 10, 64)
	if err != nil {
		return "", 0, err
	}
	return rest[:i], id, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// ---------------------------------------------------------------------------
// Inventory
// ---------------------------------------------------------------------------

type inventoryItem struct {
	ID         int64
	Name       string
	Category   string
	Quantity   int
	Equipped   bool
	Attunement bool
}

func loadInventory(characterID int64) ([]inventoryItem, error) {
	rows, err := db.DB.Query(
		`SELECT id, name, category, quantity, is_equipped, attunement
		 FROM inventory WHERE character_id = ?
		 ORDER BY category COLLATE NOCASE, name COLLATE NOCASE`, characterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []inventoryItem
	for rows.Next() {
		var it inventoryItem
		var equipped, attuned int
		if err := rows.Scan(&it.ID, &it.Name, &it.Category, &it.Quantity, &equipped, &attuned); err != nil {
			return nil, err
		}
		it.Equipped = equipped != 0
		it.Attunement = attuned != 0
		items = append(items, it)
	}
	return items, rows.Err()
}

func renderInventory(characterID int64, characterName string) botReply {
	items, err := loadInventory(characterID)
	if err != nil {
		return botReply{Text: "Could not load your inventory right now."}
	}
	heading := fmt.Sprintf("<b>🎒 Inventory — %s</b>\n", escapeHTML(characterName))
	if len(items) == 0 {
		return botReply{Text: heading + "\nEmpty. Add something with <code>/additem &lt;name&gt; [qty]</code>."}
	}
	var b strings.Builder
	b.WriteString(heading)
	var rows [][]inlineButton
	for i, it := range items {
		if i >= maxTrackedItems {
			fmt.Fprintf(&b, "…and %d more.\n", len(items)-maxTrackedItems)
			break
		}
		fmt.Fprintf(&b, "• %s ×%d", escapeHTML(it.Name), it.Quantity)
		if it.Equipped {
			b.WriteString(" (equipped)")
		}
		if it.Attunement {
			b.WriteString(" (attuned)")
		}
		b.WriteString("\n")
		toggle := "Equip"
		if it.Equipped {
			toggle = "Unequip"
		}
		rows = append(rows, []inlineButton{
			{Text: toggle, CallbackData: fmt.Sprintf("%seq:%d", cbInvPrefix, it.ID)},
			{Text: "−1", CallbackData: fmt.Sprintf("%sdec:%d", cbInvPrefix, it.ID)},
			{Text: "+1", CallbackData: fmt.Sprintf("%sinc:%d", cbInvPrefix, it.ID)},
			{Text: "Remove", CallbackData: fmt.Sprintf("%srm:%d", cbInvPrefix, it.ID)},
		})
	}
	return botReply{Text: strings.TrimRight(b.String(), "\n"), Keyboard: inlineKeyboard(rows...)}
}

func runInventory(c *cmdContext) botReply {
	charID, name, refusal, ok := requireActionCharacter(c)
	if !ok {
		if refusal.Text != "" {
			return refusal
		}
		return botReply{}
	}
	return renderInventory(charID, name)
}

func runAddItem(c *cmdContext) botReply {
	charID, name, refusal, ok := requireActionCharacter(c)
	if !ok {
		if refusal.Text != "" {
			return refusal
		}
		return botReply{}
	}
	if len(c.args) == 0 {
		return botReply{Text: addItemUsage}
	}
	qty := 1
	itemName := strings.Join(c.args, " ")
	if len(c.args) > 1 {
		if n, err := strconv.Atoi(c.args[len(c.args)-1]); err == nil && n > 0 {
			qty = n
			itemName = strings.Join(c.args[:len(c.args)-1], " ")
		}
	}
	itemName = strings.TrimSpace(itemName)
	if itemName == "" {
		return botReply{Text: addItemUsage}
	}
	if _, err := db.DB.Exec(
		`INSERT INTO inventory (character_id, name, quantity, category) VALUES (?, ?, ?, 'gear')`,
		charID, itemName, qty); err != nil {
		return botReply{Text: "Could not add that item right now."}
	}
	reply := renderInventory(charID, name)
	reply.Text = fmt.Sprintf("Added <b>%s</b> ×%d.\n\n%s", escapeHTML(itemName), qty, reply.Text)
	return reply
}

func handleInventoryCallback(c *cmdContext, data string) botReply {
	charID, name, refusal, ok := requireActionCharacter(c)
	if !ok {
		if refusal.Text != "" {
			return refusal
		}
		return botReply{}
	}
	action, id, err := parseCallbackAction(data, cbInvPrefix)
	if err != nil {
		return botReply{}
	}
	switch action {
	case "eq":
		if !mutateInventory(`UPDATE inventory SET is_equipped = CASE is_equipped WHEN 1 THEN 0 ELSE 1 END WHERE id = ? AND character_id = ?`, id, charID) {
			return botReply{Text: "That item is already gone."}
		}
	case "inc":
		if !mutateInventory(`UPDATE inventory SET quantity = quantity + 1 WHERE id = ? AND character_id = ?`, id, charID) {
			return botReply{Text: "That item is already gone."}
		}
	case "dec":
		if !mutateInventory(`UPDATE inventory SET quantity = CASE WHEN quantity > 1 THEN quantity - 1 ELSE 1 END WHERE id = ? AND character_id = ?`, id, charID) {
			return botReply{Text: "That item is already gone."}
		}
	case "rm":
		itemName, found := inventoryItemName(charID, id)
		if !found {
			return botReply{Text: "That item is already gone."}
		}
		return botReply{
			Text: fmt.Sprintf("Remove <b>%s</b>? This cannot be undone.", escapeHTML(itemName)),
			Keyboard: inlineKeyboard([]inlineButton{
				{Text: "Remove", CallbackData: fmt.Sprintf("%srmc:%d", cbInvPrefix, id)},
				{Text: "Cancel", CallbackData: fmt.Sprintf("%srmx:%d", cbInvPrefix, id)},
			}),
		}
	case "rmc":
		if !mutateInventory(`DELETE FROM inventory WHERE id = ? AND character_id = ?`, id, charID) {
			return botReply{Text: "That item is already gone."}
		}
	case "rmx":
		// cancel: fall through to re-render
	default:
		return botReply{}
	}
	return renderInventory(charID, name)
}

func mutateInventory(query string, id, characterID int64) bool {
	res, err := db.DB.Exec(query, id, characterID)
	if err != nil {
		return false
	}
	n, err := res.RowsAffected()
	return err == nil && n > 0
}

func inventoryItemName(characterID, id int64) (string, bool) {
	var name string
	if err := db.DB.QueryRow(`SELECT name FROM inventory WHERE id = ? AND character_id = ?`, id, characterID).Scan(&name); err != nil {
		return "", false
	}
	return name, true
}

// ---------------------------------------------------------------------------
// Spells
// ---------------------------------------------------------------------------

type spellItem struct {
	ID             int64
	Name           string
	Level          int
	Prepared       bool
	AlwaysPrepared bool
}

func loadSpells(characterID int64) ([]spellItem, error) {
	rows, err := db.DB.Query(
		`SELECT id, name, level, prepared, always_prepared
		 FROM spells WHERE character_id = ?
		 ORDER BY level, name COLLATE NOCASE`, characterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var spells []spellItem
	for rows.Next() {
		var sp spellItem
		var prepared, always int
		if err := rows.Scan(&sp.ID, &sp.Name, &sp.Level, &prepared, &always); err != nil {
			return nil, err
		}
		sp.Prepared = prepared != 0
		sp.AlwaysPrepared = always != 0
		spells = append(spells, sp)
	}
	return spells, rows.Err()
}

func renderSpellbook(characterID int64, characterName string) botReply {
	spells, err := loadSpells(characterID)
	if err != nil {
		return botReply{Text: "Could not load your spellbook right now."}
	}
	heading := fmt.Sprintf("<b>📖 Spellbook — %s</b>\n", escapeHTML(characterName))
	if len(spells) == 0 {
		return botReply{Text: heading + "\nNo spells yet."}
	}
	var b strings.Builder
	b.WriteString(heading)
	level := -1
	var rows [][]inlineButton
	for _, sp := range spells {
		if sp.Level != level {
			if sp.Level == 0 {
				b.WriteString("\n<b>Cantrips</b>\n")
			} else {
				fmt.Fprintf(&b, "\n<b>Level %d</b>\n", sp.Level)
			}
			level = sp.Level
		}
		fmt.Fprintf(&b, "• %s", escapeHTML(sp.Name))
		switch {
		case sp.AlwaysPrepared:
			b.WriteString(" ★")
		case sp.Prepared:
			b.WriteString(" ✓")
		}
		b.WriteString("\n")
		if sp.Level > 0 && !sp.AlwaysPrepared {
			label := "Prepare"
			if sp.Prepared {
				label = "Unprepare"
			}
			rows = append(rows, []inlineButton{
				{Text: label + " " + sp.Name, CallbackData: fmt.Sprintf("%st:%d", cbSpellPrefix, sp.ID)},
			})
		}
	}
	reply := botReply{Text: strings.TrimRight(b.String(), "\n")}
	if len(rows) > 0 {
		reply.Keyboard = inlineKeyboard(rows...)
	}
	return reply
}

func runSpells(c *cmdContext) botReply {
	charID, name, refusal, ok := requireActionCharacter(c)
	if !ok {
		if refusal.Text != "" {
			return refusal
		}
		return botReply{}
	}
	return renderSpellbook(charID, name)
}

func runPrepare(c *cmdContext) botReply   { return setSpellPrepared(c, true) }
func runUnprepare(c *cmdContext) botReply { return setSpellPrepared(c, false) }

func setSpellPrepared(c *cmdContext, want bool) botReply {
	charID, charName, refusal, ok := requireActionCharacter(c)
	if !ok {
		if refusal.Text != "" {
			return refusal
		}
		return botReply{}
	}
	if len(c.args) == 0 {
		return botReply{Text: "Usage: <code>/prepare &lt;spell&gt;</code> or <code>/unprepare &lt;spell&gt;</code>"}
	}
	spellName := strings.Join(c.args, " ")
	var id int64
	var level, always int
	err := db.DB.QueryRow(
		`SELECT id, level, always_prepared FROM spells WHERE character_id = ? AND lower(name) = lower(?) LIMIT 1`,
		charID, spellName).Scan(&id, &level, &always)
	if err != nil {
		return botReply{Text: fmt.Sprintf("No spell named <b>%s</b> on %s.", escapeHTML(spellName), escapeHTML(charName))}
	}
	if level == 0 {
		return botReply{Text: "Cantrips don't need preparation."}
	}
	if always != 0 {
		return botReply{Text: "That spell is always prepared."}
	}
	if _, err := db.DB.Exec(`UPDATE spells SET prepared = ? WHERE id = ? AND character_id = ?`, boolToInt(want), id, charID); err != nil {
		return botReply{Text: "Could not update that spell right now."}
	}
	verb := "Prepared"
	if !want {
		verb = "Unprepared"
	}
	return botReply{Text: fmt.Sprintf("%s <b>%s</b>.", verb, escapeHTML(spellName))}
}

func handleSpellCallback(c *cmdContext, data string) botReply {
	charID, name, refusal, ok := requireActionCharacter(c)
	if !ok {
		if refusal.Text != "" {
			return refusal
		}
		return botReply{}
	}
	action, id, err := parseCallbackAction(data, cbSpellPrefix)
	if err != nil || action != "t" {
		return botReply{}
	}
	var level, always, prepared int
	if err := db.DB.QueryRow(
		`SELECT level, always_prepared, prepared FROM spells WHERE id = ? AND character_id = ?`,
		id, charID).Scan(&level, &always, &prepared); err != nil {
		return botReply{Text: "That spell is no longer on your sheet."}
	}
	if level == 0 {
		return botReply{Text: "Cantrips don't need preparation."}
	}
	if always != 0 {
		return botReply{Text: "That spell is always prepared."}
	}
	if _, err := db.DB.Exec(`UPDATE spells SET prepared = ? WHERE id = ? AND character_id = ?`, 1-prepared, id, charID); err != nil {
		return botReply{Text: "Could not update that spell right now."}
	}
	return renderSpellbook(charID, name)
}

// ---------------------------------------------------------------------------
// Conditions and features
// ---------------------------------------------------------------------------

func renderConditions(characterID int64, characterName string) botReply {
	rows, err := db.DB.Query(
		`SELECT id, name FROM character_conditions WHERE character_id = ? ORDER BY name COLLATE NOCASE`, characterID)
	if err != nil {
		return botReply{Text: "Could not load your conditions right now."}
	}
	defer rows.Close()
	var b strings.Builder
	fmt.Fprintf(&b, "<b>⚗️ Conditions — %s</b>\n", escapeHTML(characterName))
	var kbRows [][]inlineButton
	found := false
	for rows.Next() {
		var id int64
		var cond string
		if err := rows.Scan(&id, &cond); err != nil {
			return botReply{Text: "Could not load your conditions right now."}
		}
		found = true
		fmt.Fprintf(&b, "• %s\n", escapeHTML(cond))
		kbRows = append(kbRows, []inlineButton{
			{Text: "Remove " + cond, CallbackData: fmt.Sprintf("%srm:%d", cbCondPrefix, id)},
		})
	}
	if err := rows.Err(); err != nil {
		return botReply{Text: "Could not load your conditions right now."}
	}
	if !found {
		b.WriteString("\nNone. Add one with <code>/condition &lt;name&gt;</code>.")
	}
	reply := botReply{Text: strings.TrimRight(b.String(), "\n")}
	if len(kbRows) > 0 {
		reply.Keyboard = inlineKeyboard(kbRows...)
	}
	return reply
}

func runConditions(c *cmdContext) botReply {
	charID, name, refusal, ok := requireActionCharacter(c)
	if !ok {
		if refusal.Text != "" {
			return refusal
		}
		return botReply{}
	}
	return renderConditions(charID, name)
}

func runCondition(c *cmdContext) botReply {
	charID, name, refusal, ok := requireActionCharacter(c)
	if !ok {
		if refusal.Text != "" {
			return refusal
		}
		return botReply{}
	}
	condName := strings.TrimSpace(strings.Join(c.args, " "))
	if condName == "" {
		return botReply{Text: "Usage: <code>/condition &lt;name&gt;</code>"}
	}
	if len([]rune(condName)) > maxConditionNameLen {
		return botReply{Text: "That condition name is too long."}
	}
	if _, err := db.DB.Exec(
		`INSERT INTO character_conditions (character_id, name, type) VALUES (?, ?, 'other')`,
		charID, condName); err != nil {
		return botReply{Text: "Could not add that condition right now."}
	}
	reply := renderConditions(charID, name)
	reply.Text = fmt.Sprintf("Added condition: <b>%s</b>.\n\n%s", escapeHTML(condName), reply.Text)
	return reply
}

func handleConditionCallback(c *cmdContext, data string) botReply {
	charID, name, refusal, ok := requireActionCharacter(c)
	if !ok {
		if refusal.Text != "" {
			return refusal
		}
		return botReply{}
	}
	action, id, err := parseCallbackAction(data, cbCondPrefix)
	if err != nil || action != "rm" {
		return botReply{}
	}
	if _, err := db.DB.Exec(`DELETE FROM character_conditions WHERE id = ? AND character_id = ?`, id, charID); err != nil {
		return botReply{Text: "Could not remove that condition right now."}
	}
	return renderConditions(charID, name)
}

func runFeatures(c *cmdContext) botReply {
	charID, name, refusal, ok := requireActionCharacter(c)
	if !ok {
		if refusal.Text != "" {
			return refusal
		}
		return botReply{}
	}
	rows, err := db.DB.Query(
		`SELECT name, source FROM character_features WHERE character_id = ? ORDER BY name COLLATE NOCASE`, charID)
	if err != nil {
		return botReply{Text: "Could not load your features right now."}
	}
	defer rows.Close()
	var b strings.Builder
	fmt.Fprintf(&b, "<b>✨ Features — %s</b>\n", escapeHTML(name))
	found := false
	for rows.Next() {
		var feat, source string
		if err := rows.Scan(&feat, &source); err != nil {
			return botReply{Text: "Could not load your features right now."}
		}
		found = true
		fmt.Fprintf(&b, "• %s", escapeHTML(feat))
		if strings.TrimSpace(source) != "" {
			fmt.Fprintf(&b, " (%s)", escapeHTML(source))
		}
		b.WriteString("\n")
	}
	if err := rows.Err(); err != nil {
		return botReply{Text: "Could not load your features right now."}
	}
	if !found {
		b.WriteString("\nNone.")
	}
	return botReply{Text: strings.TrimRight(b.String(), "\n")}
}

// ---------------------------------------------------------------------------
// Currency
// ---------------------------------------------------------------------------

func currencyColumn(denom string) (string, bool) {
	switch strings.ToLower(denom) {
	case "cp":
		return "cp", true
	case "sp":
		return "sp", true
	case "ep":
		return "ep", true
	case "gp":
		return "gp", true
	case "pp":
		return "pp", true
	}
	return "", false
}

func renderCurrency(characterID int64, characterName string) botReply {
	var cp, sp, ep, gp, pp int
	// A missing row simply reads as zero across the board.
	_ = db.DB.QueryRow(`SELECT cp, sp, ep, gp, pp FROM character_currency WHERE character_id = ?`, characterID).
		Scan(&cp, &sp, &ep, &gp, &pp)
	return botReply{Text: fmt.Sprintf(
		"<b>💰 Coin — %s</b>\n%d pp · %d gp · %d ep · %d sp · %d cp\n\nAdjust: <code>/money +N gp</code> or <code>/money -N sp</code>",
		escapeHTML(characterName), pp, gp, ep, sp, cp)}
}

func runMoney(c *cmdContext) botReply {
	charID, name, refusal, ok := requireActionCharacter(c)
	if !ok {
		if refusal.Text != "" {
			return refusal
		}
		return botReply{}
	}
	if len(c.args) == 0 {
		return renderCurrency(charID, name)
	}
	if len(c.args) != 2 {
		return botReply{Text: moneyUsage}
	}
	delta, err := strconv.Atoi(c.args[0])
	if err != nil {
		return botReply{Text: moneyUsage}
	}
	col, ok := currencyColumn(c.args[1])
	if !ok {
		return botReply{Text: moneyUsage}
	}
	// col is chosen from the fixed allowlist above, never client input.
	updated, err := db.DB.Exec(
		fmt.Sprintf(`UPDATE character_currency SET %s = MAX(0, %s + ?) WHERE character_id = ?`, col, col),
		delta, charID)
	if err != nil {
		return botReply{Text: "Could not update your coin right now."}
	}
	// character_currency has no UNIQUE on character_id, so only insert when no
	// row exists yet (the web creation path normally provides one).
	if n, _ := updated.RowsAffected(); n == 0 {
		clamped := delta
		if clamped < 0 {
			clamped = 0
		}
		if _, err := db.DB.Exec(
			fmt.Sprintf(`INSERT INTO character_currency (character_id, %s) VALUES (?, ?)`, col),
			charID, clamped); err != nil {
			return botReply{Text: "Could not update your coin right now."}
		}
	}
	reply := renderCurrency(charID, name)
	reply.Text = fmt.Sprintf("Updated <b>%s</b> by %+d.\n\n%s", col, delta, reply.Text)
	return reply
}
