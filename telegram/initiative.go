package telegram

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"villum/db"
)

const cbInitPrefix = "init:"

type initEntry struct {
	ID         int64
	Name       string
	Initiative int
	HP         int
	HPMax      int
	TurnOrder  int
}

func initFetchEntries(chatID int64) ([]initEntry, error) {
	rows, err := db.DB.Query(`SELECT id, name, initiative, hp, hp_max, turn_order FROM telegram_initiative WHERE chat_id = ? ORDER BY initiative DESC, id ASC`, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []initEntry
	for rows.Next() {
		var e initEntry
		if err := rows.Scan(&e.ID, &e.Name, &e.Initiative, &e.HP, &e.HPMax, &e.TurnOrder); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func renderInitiative(c *cmdContext) botReply {
	entries, err := initFetchEntries(c.chatID)
	if err != nil {
		return botReply{Text: "Could not load initiative right now."}
	}
	if len(entries) == 0 {
		return botReply{Text: "<b>⚔️ Initiative</b>\n\nNo combatants yet. Add one with <code>/initiative add &lt;name&gt; &lt;initiative&gt; [hp]</code>."}
	}
	// Determine active index: entry with turn_order = 1
	activeIdx := -1
	for i, e := range entries {
		if e.TurnOrder == 1 {
			activeIdx = i
			break
		}
	}
	var b strings.Builder
	b.WriteString("<b>⚔️ Initiative</b>\n\n")
	for i, e := range entries {
		marker := " "
		if i == activeIdx {
			marker = "▶"
		}
		hpStr := ""
		if e.HP != 0 || e.HPMax != 0 {
			if e.HPMax != 0 {
				hpStr = fmt.Sprintf(" HP %d/%d", e.HP, e.HPMax)
			} else {
				hpStr = fmt.Sprintf(" HP %d", e.HP)
			}
		}
		fmt.Fprintf(&b, "%s %d. <b>%s</b> (%d)%s\n", marker, i+1, escapeHTML(e.Name), e.Initiative, escapeHTML(hpStr))
	}
	text := strings.TrimRight(b.String(), "\n")

	// Build keyboard: per-entry remove buttons + Next / Clear
	var kbRows [][]inlineButton
	for _, e := range entries {
		cb := fmt.Sprintf("%srm:%d", cbInitPrefix, e.ID)
		// Ensure ≤64 bytes: prefix 8 + id digits <64
		if len(cb) > 64 {
			continue
		}
		kbRows = append(kbRows, []inlineButton{
			{Text: "✕ " + truncateRunes(e.Name, 20), CallbackData: cb},
		})
	}
	kbRows = append(kbRows, []inlineButton{
		{Text: "⏭ Next", CallbackData: cbInitPrefix + "next"},
		{Text: "🧹 Clear", CallbackData: cbInitPrefix + "clear"},
	})
	return botReply{Text: text, Keyboard: inlineKeyboard(kbRows...)}
}

func runInitiative(c *cmdContext) botReply {
	if len(c.args) == 0 {
		return renderInitiative(c)
	}
	sub := strings.ToLower(c.args[0])
	switch sub {
	case "show":
		return renderInitiative(c)
	case "clear":
		_, _ = db.DB.Exec(`DELETE FROM telegram_initiative WHERE chat_id = ?`, c.chatID)
		return renderInitiative(c)
	case "next":
		entries, err := initFetchEntries(c.chatID)
		if err != nil {
			return botReply{Text: "Could not load initiative right now."}
		}
		if len(entries) == 0 {
			return renderInitiative(c)
		}
		activeIdx := -1
		for i, e := range entries {
			if e.TurnOrder == 1 {
				activeIdx = i
				break
			}
		}
		nextIdx := 0
		if activeIdx >= 0 {
			nextIdx = (activeIdx + 1) % len(entries)
		}
		// Reset all to 0 then set next to 1
		_, _ = db.DB.Exec(`UPDATE telegram_initiative SET turn_order = 0 WHERE chat_id = ?`, c.chatID)
		_, _ = db.DB.Exec(`UPDATE telegram_initiative SET turn_order = 1 WHERE id = ? AND chat_id = ?`, entries[nextIdx].ID, c.chatID)
		return renderInitiative(c)
	case "add":
		if len(c.args) < 3 {
			return botReply{Text: "Usage: <code>/initiative add &lt;name&gt; &lt;initiative&gt; [hp]</code>"}
		}
		name := c.args[1]
		initVal, err := strconv.Atoi(c.args[2])
		if err != nil {
			return botReply{Text: "Usage: <code>/initiative add &lt;name&gt; &lt;initiative&gt; [hp]</code>"}
		}
		hp := 0
		if len(c.args) >= 4 {
			if v, err := strconv.Atoi(c.args[3]); err == nil {
				hp = v
			}
		}
		_, err = db.DB.Exec(`INSERT INTO telegram_initiative (chat_id, name, initiative, hp, hp_max, turn_order) VALUES (?, ?, ?, ?, ?, 0)`, c.chatID, name, initVal, hp, hp)
		if err != nil {
			return botReply{Text: "Could not add combatant right now."}
		}
		return renderInitiative(c)
	case "remove", "rm":
		if len(c.args) < 2 {
			return botReply{Text: "Usage: <code>/initiative remove &lt;name&gt;</code>"}
		}
		name := c.args[1]
		_, _ = db.DB.Exec(`DELETE FROM telegram_initiative WHERE chat_id = ? AND name = ?`, c.chatID, name)
		return renderInitiative(c)
	default:
		return renderInitiative(c)
	}
}

func handleInitiativeCallback(c *cmdContext, data string) botReply {
	// data is init:<action>[:<id>]
	if !strings.HasPrefix(data, cbInitPrefix) {
		return botReply{}
	}
	rest := strings.TrimPrefix(data, cbInitPrefix)
	// rest is "next", "clear", or "rm:<id>"
	if rest == "next" {
		entries, err := initFetchEntries(c.chatID)
		if err != nil {
			return botReply{Text: "Could not load initiative right now."}
		}
		if len(entries) > 0 {
			// sort already done by fetch, but ensure order
			sort.Slice(entries, func(i, j int) bool {
				if entries[i].Initiative != entries[j].Initiative {
					return entries[i].Initiative > entries[j].Initiative
				}
				return entries[i].ID < entries[j].ID
			})
			activeIdx := -1
			for i, e := range entries {
				if e.TurnOrder == 1 {
					activeIdx = i
					break
				}
			}
			nextIdx := 0
			if activeIdx >= 0 {
				nextIdx = (activeIdx + 1) % len(entries)
			}
			_, _ = db.DB.Exec(`UPDATE telegram_initiative SET turn_order = 0 WHERE chat_id = ?`, c.chatID)
			_, _ = db.DB.Exec(`UPDATE telegram_initiative SET turn_order = 1 WHERE id = ? AND chat_id = ?`, entries[nextIdx].ID, c.chatID)
		}
		return renderInitiative(c)
	}
	if rest == "clear" {
		_, _ = db.DB.Exec(`DELETE FROM telegram_initiative WHERE chat_id = ?`, c.chatID)
		return renderInitiative(c)
	}
	if strings.HasPrefix(rest, "rm:") {
		idStr := strings.TrimPrefix(rest, "rm:")
		// Use parseCallbackID for numeric id validation (prefix is cbInitPrefix+"rm:")
		id, err := parseCallbackID(data, cbInitPrefix+"rm:")
		if err != nil {
			// fallback parse directly
			if v, e2 := strconv.ParseInt(idStr, 10, 64); e2 == nil {
				id = v
			} else {
				return renderInitiative(c)
			}
		}
		_, _ = db.DB.Exec(`DELETE FROM telegram_initiative WHERE id = ? AND chat_id = ?`, id, c.chatID)
		return renderInitiative(c)
	}
	return renderInitiative(c)
}
