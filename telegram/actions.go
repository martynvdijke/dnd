package telegram

import (
	"database/sql"
	"fmt"
	"math"
	"strconv"
	"strings"

	"villum/db"
	"villum/dice"
	"villum/middleware"
	"villum/search"
)

func requireActionCharacter(c *cmdContext) (int64, string, botReply, bool) {
	uid, ok := linkedUser(c)
	if !ok {
		return 0, "", botReply{}, false
	}
	choice, hasClaim, usable := claimedCharacter(c, uid)
	if !hasClaim {
		return 0, "", botReply{Text: "No character claimed. Use /claim to pick one."}, false
	}
	if !usable {
		return 0, "", botReply{Text: "Your claimed character is no longer available (use /claim)."}, false
	}
	return choice.ID, choice.Name, botReply{}, true
}

func formatHPReply(name string, s characterSummary) string {
	base := fmt.Sprintf("<b>%s</b>: %d/%d HP", escapeHTML(name), s.HpCurrent, s.HpMax)
	if s.TempHp > 0 {
		base += fmt.Sprintf(" (+%d temp)", s.TempHp)
	}
	return base
}

func runHP(c *cmdContext) botReply {
	charID, charName, refusal, ok := requireActionCharacter(c)
	if !ok {
		if refusal.Text != "" {
			return refusal
		}
		return botReply{}
	}
	uid, _ := linkedUser(c) // already linked; retrieve again for loadSummary
	// linkedUser already validated; get uid directly
	uid2, _, _, _ := lookupUID(c.tgUserID)
	if uid2 != 0 {
		uid = uid2
	}
	if len(c.args) == 0 {
		s, found := loadCharacterSummary(charID, uid)
		if !found {
			return botReply{Text: "Character not found."}
		}
		return botReply{Text: formatHPReply(s.Name, s)}
	}
	arg := strings.TrimSpace(strings.Join(c.args, " "))
	if arg == "" {
		return botReply{Text: "Usage: /hp [+N|-N|N]"}
	}
	// Validate
	var isAdjust bool
	var delta int
	var isSet bool
	var setVal int
	if strings.HasPrefix(arg, "+") || strings.HasPrefix(arg, "-") {
		if len(arg) == 1 {
			return botReply{Text: "Usage: /hp [+N|-N|N]"}
		}
		v, err := strconv.Atoi(arg)
		if err != nil {
			return botReply{Text: "Usage: /hp [+N|-N|N]"}
		}
		isAdjust = true
		delta = v
	} else {
		v, err := strconv.Atoi(arg)
		if err != nil {
			return botReply{Text: "Usage: /hp [+N|-N|N]"}
		}
		isSet = true
		setVal = v
	}

	// Load current values
	var hpMax, hpCurrent, tempHp int
	err := db.DB.QueryRow(`SELECT hp_max, hp_current, temp_hp FROM characters WHERE id=?`, charID).Scan(&hpMax, &hpCurrent, &tempHp)
	if err != nil {
		return botReply{Text: "Character not found."}
	}

	if isAdjust {
		if delta < 0 {
			dmg := -delta
			// temp absorbs first
			if tempHp > 0 {
				if dmg <= tempHp {
					tempHp -= dmg
					dmg = 0
				} else {
					dmg -= tempHp
					tempHp = 0
				}
			}
			hpCurrent -= dmg
			if hpCurrent < 0 {
				hpCurrent = 0
			}
		} else {
			hpCurrent += delta
			if hpCurrent > hpMax {
				hpCurrent = hpMax
			}
		}
	} else if isSet {
		hpCurrent = setVal
		if hpCurrent < 0 {
			hpCurrent = 0
		}
		if hpCurrent > hpMax {
			hpCurrent = hpMax
		}
	}

	if _, err := db.DB.Exec(`UPDATE characters SET hp_current=?, temp_hp=? WHERE id=?`, hpCurrent, tempHp, charID); err != nil {
		middleware.LogWarn("telegram", "hp update failed", "error", err)
		return botReply{Text: "Could not update HP right now."}
	}
	s, found := loadCharacterSummary(charID, uid)
	if !found {
		// fallback
		return botReply{Text: fmt.Sprintf("<b>%s</b>: %d/%d HP", escapeHTML(charName), hpCurrent, hpMax)}
	}
	return botReply{Text: formatHPReply(s.Name, s)}
}

func lookupUID(tgUserID int64) (int64, string, int, bool) {
	return LookupIdentityByTelegramID(tgUserID)
}

func runRest(c *cmdContext) botReply {
	charID, charName, refusal, ok := requireActionCharacter(c)
	if !ok {
		if refusal.Text != "" {
			return refusal
		}
		return botReply{}
	}
	uid, _, _, _ := LookupIdentityByTelegramID(c.tgUserID)
	if len(c.args) != 1 {
		return botReply{Text: "Usage: /rest short|long"}
	}
	arg := strings.ToLower(strings.TrimSpace(c.args[0]))
	if arg != "short" && arg != "long" {
		return botReply{Text: "Usage: /rest short|long"}
	}

	// Use transaction
	tx, err := db.DB.Begin()
	if err != nil {
		return botReply{Text: "Could not rest right now."}
	}
	defer func() { _ = tx.Rollback() }()

	var hpMax, hpCurrent, tempHp, hitDiceCurrent, level, con int
	var hitDice string
	err = tx.QueryRow(`SELECT hp_max, hp_current, temp_hp, hit_dice, hit_dice_current, level, con FROM characters WHERE id=?`, charID).
		Scan(&hpMax, &hpCurrent, &tempHp, &hitDice, &hitDiceCurrent, &level, &con)
	if err != nil {
		return botReply{Text: "Character not found."}
	}

	if arg == "short" {
		if hitDiceCurrent <= 0 {
			_ = tx.Rollback()
			return botReply{Text: "No hit dice left to spend."}
		}
		// Roll hit die + con mod
		heal := 0
		expr := strings.TrimSpace(hitDice)
		if expr == "" {
			expr = "1d8"
		}
		// roll
		pool, perr := dice.NewPool(1)
		if perr == nil {
			if rr, rerr := pool.Roll(expr); rerr == nil && rr != nil {
				if n, err := rr.Total.Int64(); err == nil {
					heal = int(n)
				}
			}
		}
		// add con mod
		conMod := int(math.Floor(float64(con-10) / 2.0))
		heal += conMod
		if heal < 0 {
			heal = 0
		}
		hpHealed := heal
		newHP := hpCurrent + heal
		if newHP > hpMax {
			hpHealed = hpMax - hpCurrent
			if hpHealed < 0 {
				hpHealed = 0
			}
			newHP = hpMax
		}
		if newHP < 0 {
			newHP = 0
		}
		newHitDice := hitDiceCurrent - 1
		if _, err := tx.Exec(`UPDATE characters SET hp_current=?, hit_dice_current=? WHERE id=?`, newHP, newHitDice, charID); err != nil {
			return botReply{Text: "Could not rest right now."}
		}
		if _, err := tx.Exec(`INSERT INTO rest_log (character_id, rest_type, hp_healed, slots_recovered, hit_dice_spent, notes, timestamp) VALUES (?, 'short', ?, '[]', 1, ?, datetime('now'))`, charID, hpHealed, fmt.Sprintf("Short rest: rolled %s", escapeHTML(expr))); err != nil {
			middleware.LogWarn("telegram", "rest_log insert failed", "error", err)
		}
		if err := tx.Commit(); err != nil {
			return botReply{Text: "Could not rest right now."}
		}
		s, found := loadCharacterSummary(charID, uid)
		if found {
			return botReply{Text: fmt.Sprintf("<b>%s</b> took a short rest and healed %d HP. %s", escapeHTML(s.Name), hpHealed, formatHPReply(s.Name, s))}
		}
		return botReply{Text: fmt.Sprintf("<b>%s</b> took a short rest and healed %d HP.", escapeHTML(charName), hpHealed)}
	}

	// long rest
	// compute slots recovered: sum used before
	var usedBefore [9]int
	_ = tx.QueryRow(`SELECT slots_1_used, slots_2_used, slots_3_used, slots_4_used, slots_5_used, slots_6_used, slots_7_used, slots_8_used, slots_9_used FROM character_spellcasting WHERE character_id=?`, charID).
		Scan(&usedBefore[0], &usedBefore[1], &usedBefore[2], &usedBefore[3], &usedBefore[4], &usedBefore[5], &usedBefore[6], &usedBefore[7], &usedBefore[8])
	hpHealed := hpMax - hpCurrent
	if hpHealed < 0 {
		hpHealed = 0
	}
	totalSlotsRecovered := 0
	for _, u := range usedBefore {
		totalSlotsRecovered += u
	}

	if _, err := tx.Exec(`UPDATE characters SET hp_current=?, temp_hp=0, hit_dice_current=? WHERE id=?`, hpMax, level, charID); err != nil {
		return botReply{Text: "Could not rest right now."}
	}
	// reset slots
	_, _ = tx.Exec(`UPDATE character_spellcasting SET slots_1_used=0, slots_2_used=0, slots_3_used=0, slots_4_used=0, slots_5_used=0, slots_6_used=0, slots_7_used=0, slots_8_used=0, slots_9_used=0 WHERE character_id=?`, charID)

	slotsJSON := "[" + strconv.Itoa(totalSlotsRecovered) + "]"
	if _, err := tx.Exec(`INSERT INTO rest_log (character_id, rest_type, hp_healed, slots_recovered, hit_dice_spent, notes, timestamp) VALUES (?, 'long', ?, ?, 0, ?, datetime('now'))`, charID, hpHealed, slotsJSON, "Long rest"); err != nil {
		middleware.LogWarn("telegram", "rest_log insert failed", "error", err)
	}
	if err := tx.Commit(); err != nil {
		return botReply{Text: "Could not rest right now."}
	}
	s, found := loadCharacterSummary(charID, uid)
	if found {
		return botReply{Text: fmt.Sprintf("<b>%s</b> took a long rest. %s", escapeHTML(s.Name), formatHPReply(s.Name, s))}
	}
	return botReply{Text: fmt.Sprintf("<b>%s</b> took a long rest.", escapeHTML(charName))}
}

func runCast(c *cmdContext) botReply {
	charID, _, refusal, ok := requireActionCharacter(c)
	if !ok {
		if refusal.Text != "" {
			return refusal
		}
		return botReply{}
	}
	if len(c.args) == 0 {
		return botReply{Text: "Usage: /cast <spell> [dice]"}
	}

	// Determine spell name and optional dice
	// Try full join as spell name first; if not found try stripping last token as dice
	fullName := strings.TrimSpace(strings.Join(c.args, " "))
	spellName := fullName
	diceExpr := ""
	if len(c.args) >= 2 {
		// Heuristic: last token containing 'd' or digit may be dice; check if stripping yields a found spell
		candidateDice := c.args[len(c.args)-1]
		candidateSpell := strings.TrimSpace(strings.Join(c.args[:len(c.args)-1], " "))
		if candidateSpell != "" && looksLikeDice(candidateDice) {
			// we will try candidateSpell first for spell lookup; if not found fallback to full
			// To decide, try lookup candidateSpell exact
			if hasSpell(charID, candidateSpell) {
				spellName = candidateSpell
				diceExpr = candidateDice
			} else {
				// also check like fallback
				if hasSpellLike(charID, candidateSpell) {
					spellName = candidateSpell
					diceExpr = candidateDice
				}
			}
		}
		// If dice not determined, check if fullName itself is not found but candidateSpell is found, prefer that
		if diceExpr == "" && candidateSpell != "" {
			// if fullName not found and candidateSpell found, treat as dice
			if !hasSpell(charID, fullName) && !hasSpellLike(charID, fullName) {
				if hasSpell(charID, candidateSpell) || hasSpellLike(charID, candidateSpell) {
					if looksLikeDice(candidateDice) {
						spellName = candidateSpell
						diceExpr = candidateDice
					}
				}
			}
		}
	}

	// Lookup spell
	var spellID int64
	var spellLevel int
	var spellNameDB string
	err := db.DB.QueryRow(`SELECT id, level, name FROM spells WHERE character_id=? AND lower(name)=lower(?)`, charID, spellName).Scan(&spellID, &spellLevel, &spellNameDB)
	if err != nil {
		if err == sql.ErrNoRows {
			// LIKE fallback
			rows, qerr := db.DB.Query(`SELECT id, level, name FROM spells WHERE character_id=? AND name LIKE ? COLLATE NOCASE`, charID, "%"+spellName+"%")
			if qerr == nil {
				defer rows.Close()
				var matches []struct {
					id    int64
					level int
					name  string
				}
				for rows.Next() {
					var m struct {
						id    int64
						level int
						name  string
					}
					if err := rows.Scan(&m.id, &m.level, &m.name); err == nil {
						matches = append(matches, m)
					}
				}
				if len(matches) == 0 {
					// unknown spell -> suggestions
					return castUnknownSuggestions(c, spellName)
				}
				if len(matches) > 1 {
					// check exact among them
					foundExact := false
					for _, m := range matches {
						if strings.EqualFold(m.name, spellName) {
							spellID = m.id
							spellLevel = m.level
							spellNameDB = m.name
							foundExact = true
							break
						}
					}
					if !foundExact {
						var names []string
						for _, m := range matches {
							names = append(names, escapeHTML(m.name))
						}
						return botReply{Text: fmt.Sprintf("Multiple matches for %s: %s. Be more specific.", escapeHTML(spellName), strings.Join(names, ", "))}
					}
				} else {
					spellID = matches[0].id
					spellLevel = matches[0].level
					spellNameDB = matches[0].name
				}
			} else {
				return castUnknownSuggestions(c, spellName)
			}
		} else {
			return botReply{Text: "Could not cast right now."}
		}
	}
	// if still not found (no LIKE matches already handled), but if err was sql.ErrNoRows and qerr handling returned, ensure spellID set
	if spellID == 0 && spellNameDB == "" {
		// check if initial exact succeeded? If err != nil and not ErrNoRows, already returned; else if rows found empty, we returned unknown
		// fallback unknown
		return castUnknownSuggestions(c, spellName)
	}

	// Cantrip free
	if spellLevel == 0 {
		msg := fmt.Sprintf("<b>%s</b> cast %s (cantrip).", escapeHTML(loadCharName(charID)), escapeHTML(spellNameDB))
		if diceExpr != "" {
			msg += " " + rollDiceSuffix(diceExpr)
		}
		return botReply{Text: msg}
	}

	if spellLevel < 1 || spellLevel > 9 {
		return botReply{Text: "Could not cast right now."}
	}

	// Conditional update
	colMax := fmt.Sprintf("slots_%d_max", spellLevel)
	colUsed := fmt.Sprintf("slots_%d_used", spellLevel)
	query := fmt.Sprintf(`UPDATE character_spellcasting SET %s=%s+1 WHERE character_id=? AND %s < %s`, colUsed, colUsed, colUsed, colMax)
	res, err := db.DB.Exec(query, charID)
	if err != nil {
		middleware.LogWarn("telegram", "slot update failed", "error", err)
		return botReply{Text: "Could not cast right now."}
	}
	ra, _ := res.RowsAffected()
	if ra == 0 {
		return botReply{Text: fmt.Sprintf("No level %d slots left.", spellLevel)}
	}

	msg := fmt.Sprintf("<b>%s</b> cast %s (level %d).", escapeHTML(loadCharName(charID)), escapeHTML(spellNameDB), spellLevel)
	if diceExpr != "" {
		msg += " " + rollDiceSuffix(diceExpr)
	}
	return botReply{Text: msg}
}

func hasSpell(charID int64, name string) bool {
	var id int64
	err := db.DB.QueryRow(`SELECT id FROM spells WHERE character_id=? AND lower(name)=lower(?)`, charID, name).Scan(&id)
	return err == nil
}

func hasSpellLike(charID int64, name string) bool {
	var id int64
	err := db.DB.QueryRow(`SELECT id FROM spells WHERE character_id=? AND name LIKE ? COLLATE NOCASE`, charID, "%"+name+"%").Scan(&id)
	return err == nil
}

func looksLikeDice(s string) bool {
	// simple heuristic: contains 'd' or is dice-like
	ls := strings.ToLower(s)
	if strings.Contains(ls, "d") {
		return true
	}
	// also pure number with + -
	for _, r := range s {
		if r >= '0' && r <= '9' {
			continue
		}
		if r == '+' || r == '-' || r == '*' || r == '/' || r == '(' || r == ')' {
			continue
		}
		return false
	}
	return false
}

func rollDiceSuffix(expr string) string {
	pool, err := dice.NewPool(1)
	if err != nil {
		return fmt.Sprintf("Dice %s failed: %s", escapeHTML(expr), escapeHTML(err.Error()))
	}
	rr, err := pool.Roll(expr)
	if err != nil {
		return fmt.Sprintf("Dice %s failed: %s", escapeHTML(expr), escapeHTML(err.Error()))
	}
	total := rr.Total.String()
	out := rr.Output
	if out == "" {
		out = total
	}
	return fmt.Sprintf("Roll %s: %s = %s", escapeHTML(expr), escapeHTML(out), escapeHTML(total))
}

func loadCharName(charID int64) string {
	var n string
	_ = db.DB.QueryRow(`SELECT name FROM characters WHERE id=?`, charID).Scan(&n)
	if n == "" {
		n = "Character"
	}
	return n
}

func castUnknownSuggestions(c *cmdContext, name string) botReply {
	results, err := search.SearchCompendium(c.ctx, db.DB, search.CompendiumParams{
		Query:      name,
		TypeFilter: "spell",
		Limit:      5,
		Reranker:   search.DefaultRerank,
	})
	if err != nil || len(results) == 0 {
		return botReply{Text: fmt.Sprintf("Spell %s not on your sheet.", escapeHTML(name))}
	}
	var names []string
	for _, r := range results {
		names = append(names, escapeHTML(r.Name))
	}
	return botReply{Text: fmt.Sprintf("Spell %s not on your sheet. Did you mean: %s?", escapeHTML(name), strings.Join(names, ", "))}
}
