package telegram

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"villum/db"
	"villum/middleware"
)

// sendReply sends HTML text, splitting anything above Telegram's limit.
func sendReply(chatID int64, text string) {
	if strings.TrimSpace(text) == "" {
		return
	}
	for _, part := range chunkMessage(text, 4096) {
		if _, err := SendMessage(chatID, part); err != nil {
			middleware.LogWarn("telegram", "send message failed", "chat_id", chatID, "error", err)
		}
	}
}

// sendEscapedReply is for replies that embed user data: it chunks first, then
// escapes each part so a split can never cut an HTML entity in half.
func sendEscapedReply(chatID int64, text string) {
	if strings.TrimSpace(text) == "" {
		return
	}
	for _, part := range chunkMessage(text, 4000) {
		if _, err := SendMessage(chatID, escapeHTML(part)); err != nil {
			middleware.LogWarn("telegram", "send message failed", "chat_id", chatID, "error", err)
		}
	}
}

// linkedUser resolves the sender's Villum account, or answers with linking
// instructions and returns false.
func linkedUser(c *cmdContext) (int64, bool) {
	uid, _, _, ok := LookupIdentityByTelegramID(c.tgUserID)
	if !ok {
		c.reply(linkingHelpText())
		return 0, false
	}
	return uid, true
}

func runHelp(c *cmdContext) botReply {
	return botReply{Text: helpText()}
}

func runStart(c *cmdContext) botReply {
	if len(c.args) == 0 {
		return botReply{Text: welcomeText(), Keyboard: navKeyboard()}
	}
	arg := strings.TrimSpace(c.args[0])
	if strings.HasPrefix(arg, "ask_") {
		token := strings.TrimPrefix(arg, "ask_")
		if q, ok := lookupInlineAskQuery(token); ok {
			return runAskQuery(c, q)
		}
		return botReply{Text: "This link expired — run /search or /ask directly."}
	}
	code := strings.ToUpper(arg)
	userID, ok := ConsumeLinkCode(hashCode(code))
	if !ok {
		return botReply{Text: "Invalid or expired code. Generate a new one in Villum settings."}
	}
	if err := UpsertIdentity(userID, c.tgUserID, c.chatID, c.username); err != nil {
		middleware.LogWarn("telegram", "upsert identity failed", "error", err)
		return botReply{Text: "Linking failed, try again."}
	}
	return botReply{Text: "✅ Linked! Try /claim to pick your character, /create to make a new one, or /help to see everything I can do.", Keyboard: navKeyboard()}
}

// welcomeText is the friendly first-run message for /start without a code.
func welcomeText() string {
	return "👋 <b>Welcome to the Villum bot!</b>\n\n" +
		"I bring your table to Telegram: recaps, campaign overviews, party loot, open quests, location visits and character sheets.\n\n" +
		"<b>Link your account</b>\n" +
		"Easiest: send /login and enter your email.\n" +
		"Or with a code: open Villum → <b>Settings → Telegram</b>, generate a code, then send <code>/start &lt;code&gt;</code> here.\n\n" +
		"Once linked, try /claim to pick your character or /create to make one. /help lists everything."
}

func runUnlink(c *cmdContext) botReply {
	if _, _, _, ok := LookupIdentityByTelegramID(c.tgUserID); !ok {
		return botReply{Text: "Not linked. Use /start <code> to link."}
	}
	if err := clearClaimByTelegramUser(c.tgUserID); err != nil {
		middleware.LogWarn("telegram", "clear claim failed", "error", err)
	}
	if err := DeleteIdentityByTelegramID(c.tgUserID); err != nil {
		middleware.LogWarn("telegram", "unlink failed", "error", err)
		return botReply{Text: "Unlink failed, try again."}
	}
	return botReply{Text: "Unlinked."}
}

func runCancel(c *cmdContext) botReply {
	if abortCreateFlow(c.chatID) || abortAuthFlow(c.chatID) {
		return botReply{Text: "Cancelled."}
	}
	return botReply{Text: "Nothing to cancel."}
}

func runRecap(c *cmdContext) botReply {
	uid, ok := linkedUser(c)
	if !ok {
		return botReply{}
	}
	ids := memberCampaignIDs(uid)
	if len(ids) == 0 {
		return botReply{Text: "No campaigns found."}
	}

	targets := ids
	if len(c.args) > 0 {
		id, found := resolveMemberCampaign(uid, c.args)
		if !found {
			return botReply{Text: "Recap not found."}
		}
		targets = []int64{id}
	} else if len(ids) == 1 {
		targets = ids
	} else {
		return botReply{Text: campaignPickerText(uid)}
	}
	return recapReply(c.chatID, targets)
}

func runLastRecap(c *cmdContext) botReply {
	uid, ok := linkedUser(c)
	if !ok {
		return botReply{}
	}
	ids := memberCampaignIDs(uid)
	if len(ids) == 0 {
		return botReply{Text: "No campaigns found."}
	}
	return recapReply(c.chatID, ids)
}

// recapReply sends the newest recap across the given campaigns. It sends
// directly because recap bodies are user content and must be escaped.
func recapReply(chatID int64, campaignIDs []int64) botReply {
	title, content, found := latestRecap(campaignIDs)
	if !found {
		return botReply{Text: "Recap not found."}
	}
	sendEscapedReply(chatID, strings.TrimSpace(title+"\n\n"+content))
	return botReply{}
}

func runSubscribe(c *cmdContext) botReply {
	uid, ok := linkedUser(c)
	if !ok {
		return botReply{}
	}
	if err := SetDMEnabled(uid, true); err != nil {
		middleware.LogWarn("telegram", "subscribe failed", "error", err)
		return botReply{Text: "Could not update your subscription right now."}
	}
	return botReply{Text: "🔔 Subscribed. Recaps will arrive here as direct messages."}
}

func runUnsubscribe(c *cmdContext) botReply {
	uid, ok := linkedUser(c)
	if !ok {
		return botReply{}
	}
	if err := SetDMEnabled(uid, false); err != nil {
		middleware.LogWarn("telegram", "unsubscribe failed", "error", err)
		return botReply{Text: "Could not update your subscription right now."}
	}
	return botReply{Text: "🔕 Unsubscribed. You will no longer get recap DMs."}
}

// toggleSubscription flips the DM opt-in from the inline keyboard.
func toggleSubscription(c *cmdContext) botReply {
	uid, _, dmValue, ok := LookupIdentityByTelegramID(c.tgUserID)
	if !ok {
		return botReply{Text: linkingHelpText()}
	}
	dmEnabled := dmValue == 1
	if err := SetDMEnabled(uid, !dmEnabled); err != nil {
		middleware.LogWarn("telegram", "toggle subscription failed", "error", err)
		return botReply{Text: "Could not update your subscription right now."}
	}
	if dmEnabled {
		return botReply{Text: "🔕 Unsubscribed. You will no longer get recap DMs."}
	}
	return botReply{Text: "🔔 Subscribed. Recaps will arrive here as direct messages."}
}

func runStatus(c *cmdContext) botReply {
	uid, username, dmValue, ok := LookupIdentityByTelegramID(c.tgUserID)
	if !ok {
		return botReply{Text: linkingHelpText()}
	}
	dmEnabled := dmValue == 1
	var b strings.Builder
	b.WriteString("🤖 <b>Villum bot</b>\n")
	if username != "" {
		b.WriteString(fmt.Sprintf("Linked: @%s (Villum user #%d)\n", escapeHTML(username), uid))
	} else {
		b.WriteString(fmt.Sprintf("Linked: Villum user #%d\n", uid))
	}
	b.WriteString("Transport: " + escapeHTML(EffectiveMode()) + "\n")
	if dmEnabled {
		b.WriteString("Recap DMs: on\n")
	} else {
		b.WriteString("Recap DMs: off\n")
	}
	if choice, hasClaim, usable := claimedCharacter(c, uid); hasClaim && usable {
		b.WriteString("Claimed character: " + escapeHTML(choice.Name))
	} else if hasClaim {
		b.WriteString("Claimed character: unavailable (use /unclaim)")
	} else {
		b.WriteString("Claimed character: none (use /claim)")
	}
	return botReply{Text: b.String()}
}

// memberCampaignIDs returns campaigns the user owns or belongs to.
func memberCampaignIDs(uid int64) []int64 {
	seen := map[int64]bool{}
	var ids []int64
	rows, err := db.DB.Query(`SELECT id FROM campaigns WHERE user_id = ? UNION SELECT campaign_id FROM campaign_members WHERE user_id = ?`, uid, uid)
	if err != nil {
		middleware.LogWarn("telegram", "member campaigns query failed", "error", err)
		return nil
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			continue
		}
		if !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

// resolveMemberCampaign matches an id or a case-insensitive name among the
// user's campaigns.
func resolveMemberCampaign(uid int64, args []string) (int64, bool) {
	query := strings.TrimSpace(strings.Join(args, " "))
	if query == "" {
		return 0, false
	}
	if id, err := strconv.ParseInt(query, 10, 64); err == nil {
		for _, memberID := range memberCampaignIDs(uid) {
			if memberID == id {
				return id, true
			}
		}
		return 0, false
	}
	rows, err := db.DB.Query(`SELECT c.id, c.name FROM campaigns c LEFT JOIN campaign_members cm ON cm.campaign_id = c.id AND cm.user_id = ?
		WHERE c.user_id = ? OR cm.user_id IS NOT NULL`, uid, uid)
	if err != nil {
		middleware.LogWarn("telegram", "campaign lookup failed", "error", err)
		return 0, false
	}
	defer rows.Close()
	for rows.Next() {
		var campaignID int64
		var name string
		if err := rows.Scan(&campaignID, &name); err != nil {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(name), query) {
			return campaignID, true
		}
	}
	return 0, false
}

func campaignPickerText(uid int64) string {
	var b strings.Builder
	b.WriteString("Multiple campaigns. Use /recap <campaign_id>:\n")
	rows, err := db.DB.Query(`SELECT c.id, c.name FROM campaigns c LEFT JOIN campaign_members cm ON cm.campaign_id = c.id AND cm.user_id = ?
		WHERE c.user_id = ? OR cm.user_id IS NOT NULL ORDER BY c.name`, uid, uid)
	if err != nil {
		middleware.LogWarn("telegram", "campaign picker query failed", "error", err)
		return "Multiple campaigns."
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			continue
		}
		b.WriteString(fmt.Sprintf("%d: %s\n", id, name))
	}
	return strings.TrimRight(b.String(), "\n")
}

func latestRecap(campaignIDs []int64) (string, string, bool) {
	if len(campaignIDs) == 0 {
		return "", "", false
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(campaignIDs)), ",")
	args := make([]any, 0, len(campaignIDs))
	for _, id := range campaignIDs {
		args = append(args, id)
	}
	var title, content string
	err := db.DB.QueryRow(`SELECT title, content FROM campaign_recaps WHERE campaign_id IN (`+placeholders+`) ORDER BY created_at DESC LIMIT 1`, args...).Scan(&title, &content)
	if err != nil {
		return "", "", false
	}
	return title, content, true
}
