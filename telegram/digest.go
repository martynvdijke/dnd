package telegram

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"villum/db"
	"villum/middleware"
)

const digestFrequencyWeekly = "weekly"

// UpsertDigestSubscription records a weekly digest opt-in for one user and
// campaign. Re-subscribing keeps the original created_at and last_sent_at so
// it cannot be used to force an early resend.
func UpsertDigestSubscription(userID, campaignID int64, frequency string) error {
	if frequency == "" {
		frequency = digestFrequencyWeekly
	}
	_, err := db.DB.Exec(`INSERT INTO telegram_digest_subscriptions(user_id, campaign_id, frequency, created_at)
		VALUES(?, ?, ?, datetime('now'))
		ON CONFLICT(user_id, campaign_id) DO UPDATE SET frequency = excluded.frequency`,
		userID, campaignID, frequency)
	return err
}

// GetDigestSubscription reports whether a subscription exists and returns its
// last_sent_at value when it has been sent before.
func GetDigestSubscription(userID, campaignID int64) (bool, string, error) {
	var lastSent sql.NullString
	err := db.DB.QueryRow(`SELECT last_sent_at FROM telegram_digest_subscriptions
		WHERE user_id = ? AND campaign_id = ?`, userID, campaignID).Scan(&lastSent)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, "", nil
		}
		return false, "", err
	}
	if lastSent.Valid {
		return true, lastSent.String, nil
	}
	return true, "", nil
}

// DeleteDigestSubscription removes a subscription.
func DeleteDigestSubscription(userID, campaignID int64) error {
	_, err := db.DB.Exec(`DELETE FROM telegram_digest_subscriptions
		WHERE user_id = ? AND campaign_id = ?`, userID, campaignID)
	return err
}

// runDigest handles /digest (subscribe), /digest off (unsubscribe) and
// /digest status. Remaining arguments select a campaign for users in a private
// chat with several campaigns.
func runDigest(c *cmdContext) botReply {
	uid, ok := linkedUser(c)
	if !ok {
		return botReply{}
	}

	sub := "on"
	args := c.args
	if len(args) > 0 {
		switch strings.ToLower(strings.TrimSpace(args[0])) {
		case "on", "off", "status":
			sub = strings.ToLower(strings.TrimSpace(args[0]))
			args = args[1:]
		}
	}

	campaignArgs := *c
	campaignArgs.args = args
	cc, reply, ok := resolveCampaignContext(&campaignArgs, "digest")
	if !ok {
		return reply
	}

	switch sub {
	case "off":
		found, _, err := GetDigestSubscription(uid, cc.ID)
		if err != nil {
			return botReply{Text: "Could not update your digest subscription right now."}
		}
		if !found {
			return botReply{Text: fmt.Sprintf("You are not subscribed to the weekly digest for <b>%s</b>.", escapeHTML(cc.Name))}
		}
		if err := DeleteDigestSubscription(uid, cc.ID); err != nil {
			return botReply{Text: "Could not update your digest subscription right now."}
		}
		return botReply{Text: fmt.Sprintf("🔕 Weekly digest disabled for <b>%s</b>.", escapeHTML(cc.Name))}
	case "status":
		found, lastSent, err := GetDigestSubscription(uid, cc.ID)
		if err != nil {
			return botReply{Text: "Could not load your digest status right now."}
		}
		if !found {
			return botReply{Text: fmt.Sprintf("Weekly digest for <b>%s</b>: off. Send /digest to enable it.", escapeHTML(cc.Name))}
		}
		if lastSent == "" {
			return botReply{Text: fmt.Sprintf("Weekly digest for <b>%s</b>: on (not sent yet). Send /digest off to disable it.", escapeHTML(cc.Name))}
		}
		return botReply{Text: fmt.Sprintf("Weekly digest for <b>%s</b>: on (last sent %s). Send /digest off to disable it.", escapeHTML(cc.Name), escapeHTML(lastSent))}
	default:
		if err := UpsertDigestSubscription(uid, cc.ID, digestFrequencyWeekly); err != nil {
			middleware.LogWarn("telegram", "digest subscribe failed", "user_id", uid, "campaign_id", cc.ID, "error", err)
			return botReply{Text: "Could not update your digest subscription right now."}
		}
		return botReply{Text: fmt.Sprintf("🔔 Weekly digest enabled for <b>%s</b>. You will receive a summary of the last seven days.", escapeHTML(cc.Name))}
	}
}

// scanDigests sends due weekly digests. Due-ness is a rolling seven-day window
// over last_sent_at, so the timestamp is claimed before the send: if the
// process dies mid-send the digest is missed rather than sent twice. A normal
// send failure rolls the claim back so the next tick retries, which makes a
// repeat the failure mode for anything short of a crash.
func scanDigests(now time.Time) {
	nowStr := now.UTC().Format("2006-01-02 15:04:05")

	type dueDigest struct {
		userID     int64
		campaignID int64
		chatID     int64
		lastSent   sql.NullString
	}

	rows, err := db.DB.Query(`
		SELECT s.user_id, s.campaign_id, ti.telegram_chat_id, s.last_sent_at
		FROM telegram_digest_subscriptions s
		JOIN telegram_identities ti ON ti.user_id = s.user_id
		JOIN campaigns c ON c.id = s.campaign_id
		LEFT JOIN campaign_members cm ON cm.campaign_id = s.campaign_id AND cm.user_id = s.user_id
		WHERE ti.telegram_chat_id IS NOT NULL
		  AND ti.telegram_chat_id != 0
		  AND (c.user_id = s.user_id OR cm.user_id IS NOT NULL)
		  AND s.frequency = 'weekly'
		  AND (s.last_sent_at IS NULL OR datetime(s.last_sent_at) <= datetime(?, '-7 days'))
	`, nowStr)
	if err != nil {
		middleware.LogWarn("telegram", "digest scan failed", "error", err)
		return
	}
	var due []dueDigest
	func() {
		defer rows.Close()
		for rows.Next() {
			var d dueDigest
			if err := rows.Scan(&d.userID, &d.campaignID, &d.chatID, &d.lastSent); err != nil {
				continue
			}
			due = append(due, d)
		}
	}()

	for _, d := range due {
		// Claim the subscription first. The guard repeats the due check so a
		// concurrent scanner cannot claim the same row.
		res, err := db.DB.Exec(`UPDATE telegram_digest_subscriptions SET last_sent_at = ?
			WHERE user_id = ? AND campaign_id = ?
			  AND (last_sent_at IS NULL OR datetime(last_sent_at) <= datetime(?, '-7 days'))`,
			nowStr, d.userID, d.campaignID, nowStr)
		if err != nil {
			middleware.LogWarn("telegram", "digest claim failed", "user_id", d.userID, "campaign_id", d.campaignID, "error", err)
			continue
		}
		if claimed, _ := res.RowsAffected(); claimed != 1 {
			continue
		}

		text := buildDigestText(d.campaignID, now.Add(-7*24*time.Hour))
		if strings.TrimSpace(text) == "" {
			name, _ := campaignName(d.campaignID)
			if name == "" {
				name = fmt.Sprintf("Campaign %d", d.campaignID)
			}
			text = fmt.Sprintf("<b>%s — Weekly digest</b>\n\nNo new activity in the last 7 days.", escapeHTML(name))
		}
		if _, err := SendMessage(d.chatID, text); err != nil {
			// Roll the claim back so a transient failure is retried on the next
			// tick instead of losing the digest for a week.
			var previous any
			if d.lastSent.Valid {
				previous = d.lastSent.String
			}
			_, _ = db.DB.Exec(`UPDATE telegram_digest_subscriptions SET last_sent_at = ?
				WHERE user_id = ? AND campaign_id = ? AND last_sent_at = ?`,
				previous, d.userID, d.campaignID, nowStr)
			middleware.LogWarn("telegram", "digest send failed", "user_id", d.userID, "campaign_id", d.campaignID, "error", err)
			continue
		}
	}
}

// buildDigestText renders the weekly summary for one campaign: recaps and
// knowledge updated in the window, plus currently open quests.
func buildDigestText(campaignID int64, since time.Time) string {
	sinceStr := since.UTC().Format("2006-01-02 15:04:05")

	name, _ := campaignName(campaignID)
	if name == "" {
		name = fmt.Sprintf("Campaign %d", campaignID)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "<b>%s — Weekly digest</b>\n", escapeHTML(name))
	hasContent := false

	recapRows, err := db.DB.Query(`SELECT title FROM campaign_recaps
		WHERE campaign_id = ? AND datetime(created_at) >= datetime(?)
		ORDER BY created_at DESC LIMIT 5`, campaignID, sinceStr)
	if err == nil {
		var titles []string
		for recapRows.Next() {
			var title string
			if err := recapRows.Scan(&title); err == nil {
				titles = append(titles, title)
			}
		}
		recapRows.Close()
		if len(titles) > 0 {
			b.WriteString("\n<b>Recaps</b>\n")
			for _, title := range titles {
				b.WriteString("• " + escapeHTML(truncateRunes(title, 100)) + "\n")
			}
			hasContent = true
		}
	}

	quests, err := openQuests(campaignID)
	if err == nil && len(quests) > 0 {
		if len(quests) > 5 {
			quests = quests[:5]
		}
		b.WriteString("\n<b>Open quests</b>\n")
		for _, quest := range quests {
			b.WriteString("• " + escapeHTML(truncateRunes(quest.Name, 80)) + "\n")
		}
		hasContent = true
	}

	knowledgeRows, err := db.DB.Query(`SELECT title FROM campaign_knowledge
		WHERE campaign_id = ? AND datetime(updated_at) >= datetime(?)
		ORDER BY updated_at DESC LIMIT 5`, campaignID, sinceStr)
	if err == nil {
		var titles []string
		for knowledgeRows.Next() {
			var title string
			if err := knowledgeRows.Scan(&title); err == nil {
				titles = append(titles, title)
			}
		}
		knowledgeRows.Close()
		if len(titles) > 0 {
			b.WriteString("\n<b>Knowledge</b>\n")
			for _, title := range titles {
				b.WriteString("• " + escapeHTML(truncateRunes(title, 80)) + "\n")
			}
			hasContent = true
		}
	}

	if !hasContent {
		return ""
	}

	text := strings.TrimRight(b.String(), "\n")
	if runeCount(text) > 4000 {
		text = runeSlice(text, 0, 4000)
	}
	return text
}
