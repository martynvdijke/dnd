package telegram

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"villum/db"
)

const cbLevelupPrefix = "lvl:"

// lvlAbilityMod converts an ability score into its D&D modifier.
func lvlAbilityMod(score int64) int64 {
	if score >= 10 {
		return (score - 10) / 2
	}
	return (score - 11) / 2
}

// lvlHitDieAverage returns the fixed per-level HP gain for a hit-dice string
// such as "1d10" (average = die/2 + 1: d6→4, d8→5, d10→6, d12→7).
func lvlHitDieAverage(hitDice string) int64 {
	die := int64(8)
	if i := strings.LastIndex(hitDice, "d"); i >= 0 && i+1 < len(hitDice) {
		if n, err := strconv.ParseInt(strings.TrimSpace(hitDice[i+1:]), 10, 64); err == nil && n > 0 {
			die = n
		}
	}
	return die/2 + 1
}

func runLevelup(c *cmdContext) botReply {
	charID, name, refusal, ok := requireActionCharacter(c)
	if !ok {
		if refusal.Text != "" {
			return refusal
		}
		return botReply{}
	}
	var level, hpMax, con, hdCurrent int64
	var hitDice string
	if err := db.DB.QueryRow(
		`SELECT level, hp_max, con, hit_dice, hit_dice_current FROM characters WHERE id = ?`, charID,
	).Scan(&level, &hpMax, &con, &hitDice, &hdCurrent); err != nil {
		if err == sql.ErrNoRows {
			return botReply{Text: "Character not found."}
		}
		return botReply{Text: "Could not load that character right now."}
	}
	gain := lvlHitDieAverage(hitDice) + lvlAbilityMod(con)
	if gain < 1 {
		gain = 1
	}
	text := fmt.Sprintf(
		"<b>%s</b> is level %d.\nLevel up to %d? HP max %d → %d (+%d).",
		escapeHTML(name), level, level+1, hpMax, hpMax+gain, gain)
	kb := inlineKeyboard([]inlineButton{
		{Text: "⬆ Level up", CallbackData: cbLevelupPrefix + "confirm"},
		{Text: "✖ Cancel", CallbackData: cbLevelupPrefix + "cancel"},
	})
	return botReply{Text: text, Keyboard: kb}
}

func handleLevelupCallback(c *cmdContext, data string) botReply {
	switch data {
	case cbLevelupPrefix + "cancel":
		return botReply{Text: "Level up cancelled."}
	case cbLevelupPrefix + "confirm":
	default:
		return botReply{Text: "Unknown action."}
	}
	charID, name, refusal, ok := requireActionCharacter(c)
	if !ok {
		if refusal.Text != "" {
			return refusal
		}
		return botReply{}
	}
	var level, hpMax, con, hdCurrent int64
	var hitDice string
	if err := db.DB.QueryRow(
		`SELECT level, hp_max, con, hit_dice, hit_dice_current FROM characters WHERE id = ?`, charID,
	).Scan(&level, &hpMax, &con, &hitDice, &hdCurrent); err != nil {
		return botReply{Text: "Could not load that character right now."}
	}
	gain := lvlHitDieAverage(hitDice) + lvlAbilityMod(con)
	if gain < 1 {
		gain = 1
	}
	newLevel := level + 1
	newHD := hdCurrent + 1
	newHpMax := hpMax + gain
	if _, err := db.DB.Exec(
		`UPDATE characters SET level = ?, hit_dice_current = ?, hp_max = ?, updated_at = datetime('now') WHERE id = ?`,
		newLevel, newHD, newHpMax, charID); err != nil {
		return botReply{Text: "Could not level up right now."}
	}
	return botReply{Text: fmt.Sprintf(
		"<b>%s</b> reached level %d! HP max %d → %d (+%d). Hit dice %d → %d.",
		escapeHTML(name), newLevel, hpMax, newHpMax, gain, hdCurrent, newHD)}
}
