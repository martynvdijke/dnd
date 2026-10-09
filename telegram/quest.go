package telegram

import (
	"fmt"
	"strings"

	"villum/db"
)

const cbQuestPrefix = "quest:"

func handleQuestCallback(c *cmdContext, data string) botReply {
	if !strings.HasPrefix(data, cbQuestPrefix) {
		return botReply{}
	}
	action, id, err := parseCallbackAction(data, cbQuestPrefix)
	if err != nil {
		return botReply{}
	}
	var status string
	switch action {
	case "done":
		status = "complete"
	case "open":
		status = "active"
	default:
		return botReply{}
	}
	_, _ = db.DB.Exec(`UPDATE quests SET status = ?, updated_at = datetime('now') WHERE id = ?`, status, id)
	return questRenderAfter(c)
}

func questRenderAfter(c *cmdContext) botReply {
	// Re-render quest list for the context's campaign/chat
	// Use runQuests which resolves campaign context
	reply := runQuests(c)
	// If runQuests returned empty due to context issues, try generic render
	if reply.Text == "" {
		return botReply{Text: fmt.Sprintf("Updated quest.")}
	}
	return reply
}

// questListWithIDs returns open quests with ids for keyboard building.
type questRow struct {
	ID        int64
	Name      string
	Character string
	Status    string
}

func questFetchOpen(campaignID int64) ([]questRow, error) {
	rows, err := db.DB.Query(`
		SELECT q.id, q.name, ch.name, q.status
		FROM campaign_characters cc
		JOIN characters ch ON ch.id = cc.character_id
		JOIN quests q ON q.character_id = ch.id
		WHERE cc.campaign_id = ? AND q.status IN ('available', 'active')
		ORDER BY ch.name, q.name`, campaignID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []questRow
	for rows.Next() {
		var r questRow
		if err := rows.Scan(&r.ID, &r.Name, &r.Character, &r.Status); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
