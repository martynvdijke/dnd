package telegram

import (
	"fmt"
	"strings"

	"villum/db"
	"villum/middleware"
)

// runAnnounce posts the caller's announcement to the campaign chat bound to
// their resolved campaign. It only answers — the announcement itself goes to
// the campaign chat, which may be a different chat from the one the command
// was typed in.
func runAnnounce(c *cmdContext) botReply {
	uid, ok := linkedUser(c)
	if !ok {
		return botReply{}
	}

	text := strings.TrimSpace(strings.Join(c.args, " "))
	if text == "" {
		return botReply{Text: "Usage: /announce &lt;text&gt;"}
	}

	cc, reply, ok := resolveAnnounceCampaign(c, uid)
	if !ok {
		return reply
	}

	isDM, _ := campaignRole(uid, cc.ID)
	if !isDM && !isCampaignAdmin(uid) {
		return botReply{Text: "Only the campaign DM (or an admin) can send announcements."}
	}

	settings, found, err := GetCampaignTelegramSettings(cc.ID)
	if err != nil || !found || settings.ChatID == nil || *settings.ChatID == 0 {
		return botReply{Text: fmt.Sprintf("No Telegram chat is bound to <b>%s</b> yet. Bind one in Villum under <b>Campaign → Telegram</b>.", escapeHTML(cc.Name))}
	}
	if !settings.IsEnabled {
		return botReply{Text: fmt.Sprintf("Telegram posting is disabled for <b>%s</b>. Enable it in Villum under <b>Campaign → Telegram</b>.", escapeHTML(cc.Name))}
	}

	if err := sendAnnouncement(*settings.ChatID, text); err != nil {
		middleware.LogWarn("telegram", "announcement send failed", "campaign_id", cc.ID, "error", err)
		return botReply{Text: "Could not post the announcement right now. Try again."}
	}

	return botReply{Text: fmt.Sprintf("✅ Announcement posted to <b>%s</b>.", escapeHTML(cc.Name))}
}

// resolveAnnounceCampaign picks the campaign an announcement belongs to. A
// bound group chat wins; otherwise a linked DM may announce for their only
// campaign. Announcement text is not a campaign selector, so a DM with several
// campaigns is asked to use the campaign's group chat instead.
func resolveAnnounceCampaign(c *cmdContext, uid int64) (campaignContext, botReply, bool) {
	if cc, ok := boundCampaign(c.chatID); ok {
		return cc, botReply{}, true
	}
	if isGroupChat(c.chatID) {
		return campaignContext{}, botReply{Text: groupBindingHelpText()}, false
	}

	ids := memberCampaignIDs(uid)
	switch len(ids) {
	case 0:
		return campaignContext{}, botReply{Text: "No campaigns found."}, false
	case 1:
		id, ok := resolveMemberCampaign(uid, []string{fmt.Sprintf("%d", ids[0])})
		if !ok {
			return campaignContext{}, botReply{Text: "Campaign not found."}, false
		}
		name, _ := campaignName(id)
		return campaignContext{ID: id, Name: name}, botReply{}, true
	default:
		return campaignContext{}, botReply{Text: "You belong to several campaigns. Send /announce from the campaign's connected group chat."}, false
	}
}

// isCampaignAdmin reports whether the user has the global admin role. Admins
// may announce to a bound campaign even when they are not a member of it.
func isCampaignAdmin(uid int64) bool {
	var role string
	if err := db.DB.QueryRow(`SELECT role FROM users WHERE id = ?`, uid).Scan(&role); err != nil {
		return false
	}
	return role == "admin"
}

// sendAnnouncement escapes each chunk before sending so hostile markup is
// rendered literally and a split never cuts an HTML entity in half.
func sendAnnouncement(chatID int64, text string) error {
	for _, part := range chunkMessage(text, 4000) {
		if _, err := SendMessage(chatID, escapeHTML(part)); err != nil {
			return err
		}
	}
	return nil
}
