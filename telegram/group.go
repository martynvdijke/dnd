package telegram

import (
	"fmt"
	"strings"

	"villum/db"
)

// campaignContext identifies the campaign a group command acts on.
type campaignContext struct {
	ID   int64
	Name string
}

// boundCampaign returns the campaign bound to a chat for Telegram delivery.
func boundCampaign(chatID int64) (campaignContext, bool) {
	var cc campaignContext
	err := db.DB.QueryRow(`
		SELECT s.campaign_id, c.name
		FROM campaign_telegram_settings s
		JOIN campaigns c ON c.id = s.campaign_id
		WHERE s.chat_id = ? AND s.is_enabled = 1
		LIMIT 1`, chatID).Scan(&cc.ID, &cc.Name)
	if err != nil {
		return campaignContext{}, false
	}
	return cc, true
}

// isGroupChat reports whether a chat id belongs to a group or supergroup.
// Telegram group ids are always negative.
func isGroupChat(chatID int64) bool {
	return chatID < 0
}

// resolveCampaignContext picks the campaign for a campaign command: the chat
// binding first, then the linked sender's campaigns in private chats.
func resolveCampaignContext(c *cmdContext, command string) (campaignContext, botReply, bool) {
	if cc, ok := boundCampaign(c.chatID); ok {
		return cc, botReply{}, true
	}
	if isGroupChat(c.chatID) {
		return campaignContext{}, botReply{Text: groupBindingHelpText()}, false
	}
	uid, ok := linkedUser(c)
	if !ok {
		return campaignContext{}, botReply{}, false
	}
	if len(c.args) > 0 {
		id, ok := resolveMemberCampaign(uid, c.args)
		if !ok {
			return campaignContext{}, botReply{Text: "Campaign not found."}, false
		}
		name, _ := campaignName(id)
		return campaignContext{ID: id, Name: name}, botReply{}, true
	}
	ids := memberCampaignIDs(uid)
	if len(ids) == 0 {
		return campaignContext{}, botReply{Text: "No campaigns found."}, false
	}
	if len(ids) == 1 {
		name, _ := campaignName(ids[0])
		return campaignContext{ID: ids[0], Name: name}, botReply{}, true
	}
	return campaignContext{}, botReply{Text: campaignPickerTextFor(uid, command)}, false
}

// groupBindingHelpText explains how a group chat gets connected.
func groupBindingHelpText() string {
	return "This chat is not connected to a campaign yet.\n\n" +
		"A campaign DM can connect it in Villum under <b>Campaign → Telegram</b>. " +
		"After that, group commands like /items, /quests, /visits and /stats work here."
}

// campaignPickerTextFor is the group-command variant of the campaign picker.
func campaignPickerTextFor(uid int64, command string) string {
	choices, err := campaignChoices(uid)
	if err != nil || len(choices) == 0 {
		return "No campaigns found."
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Which campaign? Use /%s &lt;campaign_id&gt;:\n", command)
	for _, ch := range choices {
		fmt.Fprintf(&b, "%d: %s\n", ch.ID, escapeHTML(ch.Name))
	}
	return strings.TrimRight(b.String(), "\n")
}

// truncateRunes shortens text to at most n runes with an ellipsis.
func truncateRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return strings.TrimSpace(string(r[:n])) + "…"
}

const (
	itemsCap  = 40
	questsCap = 5
	visitsCap = 20
)

type partyItem struct {
	Name  string
	Qty   int
	Notes string
}

// partyItems lists the campaign's shared inventory ordered by name.
func partyItems(campaignID int64) ([]partyItem, error) {
	rows, err := db.DB.Query(`
		SELECT name, quantity, COALESCE(notes, '')
		FROM party_items
		WHERE campaign_id = ?
		ORDER BY name`, campaignID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []partyItem
	for rows.Next() {
		var it partyItem
		if err := rows.Scan(&it.Name, &it.Qty, &it.Notes); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

// runItems renders the campaign party inventory.
func runItems(c *cmdContext) botReply {
	cc, reply, ok := resolveCampaignContext(c, "items")
	if !ok {
		return reply
	}
	items, err := partyItems(cc.ID)
	if err != nil {
		return botReply{Text: "Could not load party items right now."}
	}
	if len(items) == 0 {
		return botReply{Text: fmt.Sprintf("<b>%s — Party items</b>\n\nNo party items yet.", escapeHTML(cc.Name))}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "<b>%s — Party items</b>\n\n", escapeHTML(cc.Name))
	shown := items
	more := 0
	if len(items) > itemsCap {
		shown = items[:itemsCap]
		more = len(items) - itemsCap
	}
	for _, it := range shown {
		line := fmt.Sprintf("• %s ×%d", escapeHTML(it.Name), it.Qty)
		if it.Notes != "" {
			line += " — " + escapeHTML(truncateRunes(it.Notes, 80))
		}
		b.WriteString(line + "\n")
	}
	if more > 0 {
		fmt.Fprintf(&b, "\n…and %d more.", more)
	}
	return botReply{Text: strings.TrimRight(b.String(), "\n")}
}

type questItem struct {
	Character  string
	Name       string
	Objectives string
	Rewards    string
}

// openQuests lists the party's available and active quests.
func openQuests(campaignID int64) ([]questItem, error) {
	rows, err := db.DB.Query(`
		SELECT ch.name, q.name, COALESCE(q.objectives, ''), COALESCE(q.rewards, '')
		FROM campaign_characters cc
		JOIN characters ch ON ch.id = cc.character_id
		JOIN quests q ON q.character_id = ch.id
		WHERE cc.campaign_id = ? AND q.status IN ('available', 'active')
		ORDER BY ch.name, q.name`, campaignID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var quests []questItem
	for rows.Next() {
		var q questItem
		if err := rows.Scan(&q.Character, &q.Name, &q.Objectives, &q.Rewards); err != nil {
			return nil, err
		}
		quests = append(quests, q)
	}
	return quests, rows.Err()
}

// runQuests renders the party's open quests grouped by character.
func runQuests(c *cmdContext) botReply {
	cc, reply, ok := resolveCampaignContext(c, "quests")
	if !ok {
		return reply
	}
	var characters int
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM campaign_characters WHERE campaign_id = ?`, cc.ID).Scan(&characters); err != nil {
		return botReply{Text: "Could not load quests right now."}
	}
	if characters == 0 {
		return botReply{Text: fmt.Sprintf("<b>%s — Quests</b>\n\nNo characters in this campaign yet.", escapeHTML(cc.Name))}
	}
	quests, err := openQuests(cc.ID)
	if err != nil {
		return botReply{Text: "Could not load quests right now."}
	}
	if len(quests) == 0 {
		return botReply{Text: fmt.Sprintf("<b>%s — Quests</b>\n\nNo open quests for the party right now.", escapeHTML(cc.Name))}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "<b>%s — Open quests</b>\n", escapeHTML(cc.Name))
	current := ""
	count := 0
	hidden := 0
	for _, q := range quests {
		if q.Character != current {
			current = q.Character
			count = 0
			fmt.Fprintf(&b, "\n<b>%s</b>\n", escapeHTML(q.Character))
		}
		if count >= questsCap {
			hidden++
			continue
		}
		count++
		line := "• " + escapeHTML(q.Name)
		if q.Objectives != "" {
			line += "\n  " + escapeHTML(truncateRunes(q.Objectives, 120))
		}
		if q.Rewards != "" {
			line += "\n  Reward: " + escapeHTML(truncateRunes(q.Rewards, 80))
		}
		b.WriteString(line + "\n")
	}
	if hidden > 0 {
		fmt.Fprintf(&b, "\n…and %d more.", hidden)
	}
	return botReply{Text: strings.TrimRight(b.String(), "\n")}
}

type visitItem struct {
	Character    string
	Location     string
	Type         string
	Relationship string
	Notes        string
}

// partyVisits lists the party's most recent location links.
func partyVisits(campaignID int64) ([]visitItem, error) {
	rows, err := db.DB.Query(`
		SELECT ch.name, l.name, COALESCE(l.type, ''), COALESCE(cl.relationship, 'visited'), COALESCE(cl.notes, '')
		FROM character_locations cl
		JOIN locations l ON l.id = cl.location_id
		JOIN characters ch ON ch.id = cl.character_id
		JOIN campaign_characters cc ON cc.character_id = cl.character_id
		WHERE cc.campaign_id = ?
		ORDER BY cl.id DESC
		LIMIT ?`, campaignID, visitsCap+1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var visits []visitItem
	for rows.Next() {
		var v visitItem
		if err := rows.Scan(&v.Character, &v.Location, &v.Type, &v.Relationship, &v.Notes); err != nil {
			return nil, err
		}
		visits = append(visits, v)
	}
	return visits, rows.Err()
}

// runVisits renders the party's latest location visits.
func runVisits(c *cmdContext) botReply {
	cc, reply, ok := resolveCampaignContext(c, "visits")
	if !ok {
		return reply
	}
	visits, err := partyVisits(cc.ID)
	if err != nil {
		return botReply{Text: "Could not load location visits right now."}
	}
	if len(visits) == 0 {
		return botReply{Text: fmt.Sprintf("<b>%s — Location visits</b>\n\nNo location visits recorded yet.", escapeHTML(cc.Name))}
	}
	more := 0
	if len(visits) > visitsCap {
		more = len(visits) - visitsCap
		visits = visits[:visitsCap]
	}
	var b strings.Builder
	fmt.Fprintf(&b, "<b>%s — Location visits</b>\n\n", escapeHTML(cc.Name))
	for _, v := range visits {
		line := fmt.Sprintf("• <b>%s</b>", escapeHTML(v.Location))
		if v.Type != "" {
			line += " (" + escapeHTML(v.Type) + ")"
		}
		line += " — " + escapeHTML(v.Character)
		if v.Relationship != "" && v.Relationship != "visited" {
			line += ", " + escapeHTML(v.Relationship)
		}
		if v.Notes != "" {
			line += "\n  " + escapeHTML(truncateRunes(v.Notes, 100))
		}
		b.WriteString(line + "\n")
	}
	if more > 0 {
		fmt.Fprintf(&b, "\n…and %d more.", more)
	}
	return botReply{Text: strings.TrimRight(b.String(), "\n")}
}

type campaignStats struct {
	Characters int
	Sessions   int
	Quests     int
	QuestsOpen int
	QuestsDone int
	NPCs       int
	Locations  int
	AvgLevel   float64
	HasLevels  bool
}

// loadCampaignStats mirrors the web campaign analytics.
func loadCampaignStats(campaignID int64) (campaignStats, error) {
	var s campaignStats
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM campaign_characters WHERE campaign_id = ?`, campaignID).Scan(&s.Characters); err != nil {
		return s, err
	}
	if err := db.DB.QueryRow(`
		SELECT COUNT(*) FROM sessions se
		JOIN campaign_characters cc ON cc.character_id = se.character_id
		WHERE cc.campaign_id = ?`, campaignID).Scan(&s.Sessions); err != nil {
		return s, err
	}
	if err := db.DB.QueryRow(`
		SELECT COUNT(*),
		       SUM(CASE WHEN q.status IN ('available', 'active') THEN 1 ELSE 0 END),
		       SUM(CASE WHEN q.status = 'complete' THEN 1 ELSE 0 END)
		FROM quests q
		JOIN campaign_characters cc ON cc.character_id = q.character_id
		WHERE cc.campaign_id = ?`, campaignID).Scan(&s.Quests, &s.QuestsOpen, &s.QuestsDone); err != nil {
		return s, err
	}
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM npcs WHERE user_id IN (SELECT user_id FROM campaign_members WHERE campaign_id = ?)`, campaignID).Scan(&s.NPCs); err != nil {
		return s, err
	}
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM locations WHERE user_id IN (SELECT user_id FROM campaign_members WHERE campaign_id = ?)`, campaignID).Scan(&s.Locations); err != nil {
		return s, err
	}
	var avg *float64
	if err := db.DB.QueryRow(`
		SELECT AVG(ch.level) FROM campaign_characters cc
		JOIN characters ch ON ch.id = cc.character_id
		WHERE cc.campaign_id = ?`, campaignID).Scan(&avg); err != nil {
		return s, err
	}
	if avg != nil {
		s.AvgLevel = *avg
		s.HasLevels = true
	}
	return s, nil
}

// campaignStatsReply renders the campaign statistics block.
func campaignStatsReply(cc campaignContext) botReply {
	s, err := loadCampaignStats(cc.ID)
	if err != nil {
		return botReply{Text: "Could not load campaign statistics right now."}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "<b>%s — Statistics</b>\n\n", escapeHTML(cc.Name))
	fmt.Fprintf(&b, "Characters: %d\n", s.Characters)
	fmt.Fprintf(&b, "Sessions: %d\n", s.Sessions)
	fmt.Fprintf(&b, "Quests: %d (%d open, %d completed)\n", s.Quests, s.QuestsOpen, s.QuestsDone)
	fmt.Fprintf(&b, "NPCs: %d\n", s.NPCs)
	fmt.Fprintf(&b, "Locations: %d\n", s.Locations)
	if s.HasLevels {
		fmt.Fprintf(&b, "Average level: %.1f", s.AvgLevel)
	} else {
		b.WriteString("Average level: —")
	}
	return botReply{Text: b.String()}
}
