package telegram

import (
	"fmt"
	"strings"

	"villum/db"
)

const forumUsageText = "Usage: /topic <name>\nExample: /topic Session 12 — The Dragon's Lair"

func runTopic(c *cmdContext) botReply {
	name := strings.TrimSpace(strings.Join(c.args, " "))
	if name == "" {
		return botReply{Text: forumUsageText}
	}
	// Could optionally enforce group chat, but CreateForumTopic will fail outside groups anyway
	topic, err := CreateForumTopic(c.chatID, name)
	if err != nil {
		return botReply{Text: escapeHTML(err.Error()) + "\n\nMake sure this is a forum supergroup and the bot is an admin."}
	}
	var campaignID *int64
	if cc, ok := boundCampaign(c.chatID); ok {
		v := cc.ID
		campaignID = &v
	}
	threadID := int64(0)
	if topic != nil {
		threadID = int64(topic.MessageThreadID)
	}
	_, _ = db.DB.Exec(`INSERT INTO telegram_forum_topics(chat_id, campaign_id, name, thread_id) VALUES(?,?,?,?)`, c.chatID, campaignID, name, threadID)
	return botReply{Text: fmt.Sprintf("✅ Created topic <b>%s</b> (thread %d).", escapeHTML(name), threadID)}
}

func runTopics(c *cmdContext) botReply {
	rows, err := db.DB.Query(`SELECT name, thread_id FROM telegram_forum_topics WHERE chat_id = ? ORDER BY created_at`, c.chatID)
	if err != nil {
		return botReply{Text: "Could not load topics right now."}
	}
	defer rows.Close()
	var lines []string
	for rows.Next() {
		var name string
		var tid int64
		if err := rows.Scan(&name, &tid); err != nil {
			continue
		}
		lines = append(lines, fmt.Sprintf("• %s (thread %d)", escapeHTML(name), tid))
	}
	if len(lines) == 0 {
		return botReply{Text: "No forum topics yet. Create one with /topic <name>."}
	}
	return botReply{Text: "<b>Forum topics</b>\n" + strings.Join(lines, "\n")}
}
