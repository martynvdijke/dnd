package telegram

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	tgmodels "github.com/go-telegram/bot/models"

	"villum/db"
	"villum/middleware"
)

// characterSummary is the data behind /sheet and /stats.
type characterSummary struct {
	ID                int64
	UserID            int64
	Name              string
	Race              string
	Class             string
	Subclass          string
	Level             int
	Ac                int
	Initiative        int
	Speed             int
	HpMax             int
	HpCurrent         int
	TempHp            int
	Str               int
	Dex               int
	Con               int
	Int               int
	Wis               int
	Cha               int
	ProficiencyBonus  int
	PassivePerception int
	Inspiration       int
	ExhaustionLevel   int
	HitDice           string
	HitDiceCurrent    int
	CharacterType     string
	Campaigns         string
}

// characterListEntry is one row of /characters.
type characterListEntry struct {
	ID            int64
	Name          string
	Race          string
	Class         string
	Level         int
	HpCurrent     int
	HpMax         int
	CharacterType string
	Campaigns     string
	ClaimedBy     int64
}

func loadCharacterSummary(id, uid int64) (characterSummary, bool) {
	var s characterSummary
	err := db.DB.QueryRow(`SELECT c.id, c.user_id, c.name, c.race, c.class, c.subclass, c.level,
			c.ac, c.initiative, c.speed, c.hp_max, c.hp_current, c.temp_hp,
			c.str, c.dex, c.con, c.int, c.wis, c.cha,
			c.proficiency_bonus, c.passive_perception, c.inspiration, c.exhaustion_level,
			c.hit_dice, c.hit_dice_current, c.character_type,
			COALESCE((SELECT GROUP_CONCAT(camp.name, ', ') FROM campaign_characters cc
				JOIN campaigns camp ON camp.id = cc.campaign_id WHERE cc.character_id = c.id), '')
		FROM characters c
		WHERE c.id = ? AND `+editableCharacterSQL,
		id, uid, uid, uid, uid).
		Scan(&s.ID, &s.UserID, &s.Name, &s.Race, &s.Class, &s.Subclass, &s.Level,
			&s.Ac, &s.Initiative, &s.Speed, &s.HpMax, &s.HpCurrent, &s.TempHp,
			&s.Str, &s.Dex, &s.Con, &s.Int, &s.Wis, &s.Cha,
			&s.ProficiencyBonus, &s.PassivePerception, &s.Inspiration, &s.ExhaustionLevel,
			&s.HitDice, &s.HitDiceCurrent, &s.CharacterType, &s.Campaigns)
	if err != nil {
		return characterSummary{}, false
	}
	return s, true
}

func listCharacters(uid int64) ([]characterListEntry, error) {
	rows, err := db.DB.Query(`SELECT c.id, c.name, c.race, c.class, c.level, c.hp_current, c.hp_max, c.character_type,
			COALESCE((SELECT GROUP_CONCAT(camp.name, ', ') FROM campaign_characters cc
				JOIN campaigns camp ON camp.id = cc.campaign_id WHERE cc.character_id = c.id), ''),
			COALESCE(tcc.telegram_user_id, 0)
		FROM characters c
		LEFT JOIN telegram_character_claims tcc ON tcc.character_id = c.id
		WHERE `+editableCharacterSQL+`
		ORDER BY c.name`,
		uid, uid, uid, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []characterListEntry
	for rows.Next() {
		var e characterListEntry
		if err := rows.Scan(&e.ID, &e.Name, &e.Race, &e.Class, &e.Level, &e.HpCurrent, &e.HpMax, &e.CharacterType, &e.Campaigns, &e.ClaimedBy); err != nil {
			continue
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func runCharacters(c *cmdContext) botReply {
	uid, ok := linkedUser(c)
	if !ok {
		return botReply{}
	}
	entries, err := listCharacters(uid)
	if err != nil {
		middleware.LogWarn("telegram", "character list failed", "error", err)
		return botReply{Text: "Could not load your characters right now."}
	}
	if len(entries) == 0 {
		return botReply{Text: "You have no characters yet. Use /create to make one."}
	}
	var b strings.Builder
	b.WriteString("<b>Your characters</b>\n\n")
	var rows [][]inlineButton
	for _, e := range entries {
		line := fmt.Sprintf("• <b>%s</b> — %s %s Lv%d (HP %d/%d)",
			escapeHTML(e.Name), escapeHTML(e.Race), escapeHTML(e.Class), e.Level, e.HpCurrent, e.HpMax)
		if e.Campaigns != "" {
			line += " · " + escapeHTML(e.Campaigns)
		}
		if e.ClaimedBy == c.tgUserID {
			line += " ⭐"
		}
		b.WriteString(line + "\n")
		rows = append(rows, []inlineButton{{Text: e.Name + " sheet", CallbackData: cbSheetPrefix + strconv.FormatInt(e.ID, 10)}})
	}
	return botReply{Text: strings.TrimRight(b.String(), "\n"), Keyboard: inlineKeyboard(rows...)}
}

// resolveCharacterArg matches an id or name among the characters the user may
// edit.
func resolveCharacterArg(uid int64, args []string) (int64, bool) {
	query := strings.TrimSpace(strings.Join(args, " "))
	if query == "" {
		return 0, false
	}
	if id, err := strconv.ParseInt(query, 10, 64); err == nil {
		if _, found := loadCharacterSummary(id, uid); found {
			return id, true
		}
		return 0, false
	}
	rows, err := db.DB.Query(`SELECT c.id, c.name FROM characters c WHERE `+editableCharacterSQL, uid, uid, uid, uid)
	if err != nil {
		return 0, false
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(name), query) {
			return id, true
		}
	}
	return 0, false
}

// pickCharacter resolves the target for /sheet and /stats. When done is true
// the returned reply should be sent as-is.
func pickCharacter(c *cmdContext, uid int64) (int64, botReply, bool) {
	if len(c.args) > 0 {
		id, found := resolveCharacterArg(uid, c.args)
		if !found {
			return 0, botReply{Text: "Character not found."}, true
		}
		return id, botReply{}, false
	}
	choice, hasClaim, usable := claimedCharacter(c, uid)
	if hasClaim && usable {
		return choice.ID, botReply{}, false
	}
	if hasClaim {
		return 0, botReply{Text: "Your claimed character is no longer available (use /unclaim), or pick one with /characters."}, true
	}
	entries, err := listCharacters(uid)
	if err != nil {
		middleware.LogWarn("telegram", "character list failed", "error", err)
		return 0, botReply{Text: "Could not load your characters right now."}, true
	}
	switch len(entries) {
	case 0:
		return 0, botReply{Text: "You have no characters yet. Use /create to make one."}, true
	case 1:
		return entries[0].ID, botReply{}, false
	default:
		var rows [][]inlineButton
		for _, e := range entries {
			rows = append(rows, []inlineButton{{Text: e.Name + " sheet", CallbackData: cbSheetPrefix + strconv.FormatInt(e.ID, 10)}})
		}
		return 0, botReply{Text: "Pick a character:", Keyboard: inlineKeyboard(rows...)}, true
	}
}

func runSheet(c *cmdContext) botReply {
	uid, ok := linkedUser(c)
	if !ok {
		return botReply{}
	}
	id, reply, done := pickCharacter(c, uid)
	if done {
		return reply
	}
	return renderSheetReply(uid, id)
}

func runStats(c *cmdContext) botReply {
	// In a campaign chat, /stats shows the campaign statistics; character
	// stats stay in private chats where a claim makes sense.
	if cc, ok := boundCampaign(c.chatID); ok {
		return campaignStatsReply(cc)
	}
	uid, ok := linkedUser(c)
	if !ok {
		return botReply{}
	}
	id, reply, done := pickCharacter(c, uid)
	if done {
		return reply
	}
	return renderStatsReply(uid, id)
}

func sheetByID(c *cmdContext, id int64) botReply {
	uid, ok := linkedUser(c)
	if !ok {
		return botReply{}
	}
	return renderSheetReply(uid, id)
}

func renderSheetReply(uid, id int64) botReply {
	s, found := loadCharacterSummary(id, uid)
	if !found {
		return botReply{Text: "Character not found."}
	}
	return botReply{Text: renderSheet(s)}
}

func renderStatsReply(uid, id int64) botReply {
	s, found := loadCharacterSummary(id, uid)
	if !found {
		return botReply{Text: "Character not found."}
	}
	return botReply{Text: renderStats(s)}
}

func renderStats(s characterSummary) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("<b>%s</b> — %s %s · Level %d\n\n",
		escapeHTML(s.Name), escapeHTML(s.Race), escapeHTML(s.Class), s.Level))
	b.WriteString("STR " + abilityLine(s.Str) + "   DEX " + abilityLine(s.Dex) + "   CON " + abilityLine(s.Con) + "\n")
	b.WriteString("INT " + abilityLine(s.Int) + "   WIS " + abilityLine(s.Wis) + "   CHA " + abilityLine(s.Cha) + "\n\n")
	b.WriteString(fmt.Sprintf("HP %d/%d   AC %d   Initiative %s   Speed %d ft\n",
		s.HpCurrent, s.HpMax, s.Ac, formatMod(s.Initiative), s.Speed))
	b.WriteString(fmt.Sprintf("Prof %s   Passive Perception %d   Hit dice %d%s",
		formatMod(s.ProficiencyBonus), s.PassivePerception, s.HitDiceCurrent, escapeHTML(s.HitDice)))
	return b.String()
}

func renderSheet(s characterSummary) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("<b>%s</b> — %s %s", escapeHTML(s.Name), escapeHTML(s.Race), escapeHTML(s.Class)))
	if s.Subclass != "" {
		b.WriteString(" (" + escapeHTML(s.Subclass) + ")")
	}
	b.WriteString(fmt.Sprintf(" · Level %d\n", s.Level))
	if s.Campaigns != "" {
		b.WriteString("<i>" + escapeHTML(s.Campaigns) + "</i>\n")
	}
	b.WriteString("\n<b>Abilities</b>\n")
	b.WriteString("STR " + abilityLine(s.Str) + "   DEX " + abilityLine(s.Dex) + "   CON " + abilityLine(s.Con) + "\n")
	b.WriteString("INT " + abilityLine(s.Int) + "   WIS " + abilityLine(s.Wis) + "   CHA " + abilityLine(s.Cha) + "\n")
	b.WriteString("\n<b>Combat</b>\n")
	b.WriteString(fmt.Sprintf("HP %d/%d", s.HpCurrent, s.HpMax))
	if s.TempHp > 0 {
		b.WriteString(fmt.Sprintf(" (+%d temp)", s.TempHp))
	}
	b.WriteString(fmt.Sprintf("   AC %d   Initiative %s   Speed %d ft\n",
		s.Ac, formatMod(s.Initiative), s.Speed))
	b.WriteString(fmt.Sprintf("Prof %s   Passive Perception %d   Hit dice %d%s\n",
		formatMod(s.ProficiencyBonus), s.PassivePerception, s.HitDiceCurrent, escapeHTML(s.HitDice)))
	var flags []string
	if s.Inspiration > 0 {
		flags = append(flags, "Inspiration")
	}
	if s.ExhaustionLevel > 0 {
		flags = append(flags, fmt.Sprintf("Exhaustion %d", s.ExhaustionLevel))
	}
	if len(flags) > 0 {
		b.WriteString(strings.Join(flags, " · ") + "\n")
	}
	b.WriteString("\n<b>Other</b>\n")
	b.WriteString(fmt.Sprintf("Conditions %d · Features %d · Spells %d · Inventory %d\n",
		countCharacterRows("character_conditions", s.ID),
		countCharacterRows("character_features", s.ID),
		countCharacterRows("spells", s.ID),
		countCharacterRows("inventory", s.ID)))
	b.WriteString("Currency: " + currencySummary(s.ID) + "\n")
	b.WriteString("Manage: /inventory · /spells · /conditions · /features · /money")
	return b.String()
}

func abilityLine(score int) string {
	return fmt.Sprintf("%d (%s)", score, formatMod(modifierFor(score)))
}

func modifierFor(score int) int {
	return int(math.Floor(float64(score-10) / 2.0))
}

func formatMod(v int) string {
	if v >= 0 {
		return fmt.Sprintf("+%d", v)
	}
	return strconv.Itoa(v)
}

func countCharacterRows(table string, characterID int64) int {
	switch table {
	case "character_conditions", "character_features", "spells", "inventory", "character_proficiencies":
	default:
		return 0
	}
	var n int
	if err := db.DB.QueryRow("SELECT COUNT(*) FROM "+table+" WHERE character_id = ?", characterID).Scan(&n); err != nil {
		return 0
	}
	return n
}

func currencySummary(characterID int64) string {
	var cp, sp, ep, gp, pp int
	err := db.DB.QueryRow(`SELECT cp, sp, ep, gp, pp FROM character_currency WHERE character_id = ?`, characterID).
		Scan(&cp, &sp, &ep, &gp, &pp)
	if err != nil {
		return "none"
	}
	var parts []string
	for _, c := range []struct {
		label string
		value int
	}{{"pp", pp}, {"gp", gp}, {"ep", ep}, {"sp", sp}, {"cp", cp}} {
		if c.value != 0 {
			parts = append(parts, fmt.Sprintf("%d %s", c.value, c.label))
		}
	}
	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, ", ")
}

// --- Campaign overview ---

type campaignChoice struct {
	ID        int64
	Name      string
	IsDM      bool
	MemberCnt int
}

func campaignChoices(uid int64) ([]campaignChoice, error) {
	rows, err := db.DB.Query(`SELECT c.id, c.name,
			CASE WHEN c.user_id = ? OR cm.role = 'dm' THEN 1 ELSE 0 END,
			(SELECT COUNT(*) FROM campaign_members m WHERE m.campaign_id = c.id)
		FROM campaigns c
		LEFT JOIN campaign_members cm ON cm.campaign_id = c.id AND cm.user_id = ?
		WHERE c.user_id = ? OR cm.user_id IS NOT NULL
		ORDER BY c.name`, uid, uid, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []campaignChoice
	for rows.Next() {
		var cc campaignChoice
		var isDM int
		if err := rows.Scan(&cc.ID, &cc.Name, &isDM, &cc.MemberCnt); err != nil {
			continue
		}
		cc.IsDM = isDM == 1
		out = append(out, cc)
	}
	return out, rows.Err()
}

// campaignRole returns membership and DM status for the user, ok=false when
// the user is not in the campaign.
func campaignRole(uid, campaignID int64) (isDM bool, ok bool) {
	var ownerID int64
	var role sql.NullString
	err := db.DB.QueryRow(`SELECT c.user_id, cm.role FROM campaigns c
		LEFT JOIN campaign_members cm ON cm.campaign_id = c.id AND cm.user_id = ?
		WHERE c.id = ?`, uid, campaignID).Scan(&ownerID, &role)
	if err != nil {
		return false, false
	}
	if ownerID == uid {
		return true, true
	}
	if !role.Valid {
		return false, false
	}
	return role.String == "dm", true
}

func runOverview(c *cmdContext) botReply {
	uid, ok := linkedUser(c)
	if !ok {
		return botReply{}
	}
	if len(c.args) > 0 {
		id, found := resolveMemberCampaign(uid, c.args)
		if !found {
			return botReply{Text: "Campaign not found."}
		}
		return overviewDetail(uid, id)
	}
	campaigns, err := campaignChoices(uid)
	if err != nil {
		middleware.LogWarn("telegram", "campaign list failed", "error", err)
		return botReply{Text: "Could not load your campaigns right now."}
	}
	if len(campaigns) == 0 {
		return botReply{Text: "No campaigns found."}
	}
	var b strings.Builder
	b.WriteString("<b>Your campaigns</b>\n\n")
	var rows [][]inlineButton
	for _, cc := range campaigns {
		role := "player"
		if cc.IsDM {
			role = "DM"
		}
		b.WriteString(fmt.Sprintf("• <b>%s</b> — %s · %d members\n", escapeHTML(cc.Name), role, cc.MemberCnt))
		rows = append(rows, []inlineButton{{Text: cc.Name, CallbackData: cbCampaignPrefix + strconv.FormatInt(cc.ID, 10)}})
	}
	return botReply{Text: strings.TrimRight(b.String(), "\n"), Keyboard: inlineKeyboard(rows...)}
}

func overviewCampaignByID(c *cmdContext, id int64) botReply {
	uid, ok := linkedUser(c)
	if !ok {
		return botReply{}
	}
	return overviewDetail(uid, id)
}

func overviewDetail(uid, campaignID int64) botReply {
	isDM, member := campaignRole(uid, campaignID)
	if !member {
		return botReply{Text: "Campaign not found."}
	}
	var name string
	var ownerID int64
	if err := db.DB.QueryRow(`SELECT name, user_id FROM campaigns WHERE id = ?`, campaignID).Scan(&name, &ownerID); err != nil {
		return botReply{Text: "Campaign not found."}
	}
	var memberCount int
	_ = db.DB.QueryRow(`SELECT COUNT(*) FROM campaign_members WHERE campaign_id = ?`, campaignID).Scan(&memberCount)
	role := "player"
	if isDM {
		role = "DM"
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("<b>%s</b>\n", escapeHTML(name)))
	b.WriteString(fmt.Sprintf("Role: %s · Members: %d\n", role, memberCount))

	rows, err := db.DB.Query(`SELECT ch.name, ch.class, ch.level, ch.hp_current, ch.hp_max, ch.character_type
		FROM campaign_characters cc JOIN characters ch ON ch.id = cc.character_id
		WHERE cc.campaign_id = ? ORDER BY ch.name`, campaignID)
	if err == nil {
		defer rows.Close()
		var party []string
		for rows.Next() {
			var chName, class, charType string
			var level, hpCurrent, hpMax int
			if err := rows.Scan(&chName, &class, &level, &hpCurrent, &hpMax, &charType); err != nil {
				continue
			}
			line := fmt.Sprintf("%s — %s Lv%d (HP %d/%d)", escapeHTML(chName), escapeHTML(class), level, hpCurrent, hpMax)
			if charType == "linked" {
				line += " [DM-run]"
			}
			party = append(party, line)
		}
		if len(party) > 0 {
			b.WriteString("\n<b>Party</b>\n• " + strings.Join(party, "\n• ") + "\n")
		}
	}

	var recapTitle, recapCreated sql.NullString
	_ = db.DB.QueryRow(`SELECT title, created_at FROM campaign_recaps WHERE campaign_id = ? ORDER BY created_at DESC LIMIT 1`, campaignID).
		Scan(&recapTitle, &recapCreated)
	if recapTitle.Valid && recapTitle.String != "" {
		when := ""
		if recapCreated.Valid && len(recapCreated.String) >= 10 {
			when = " (" + recapCreated.String[:10] + ")"
		}
		b.WriteString("\nLatest recap: " + escapeHTML(recapTitle.String) + when)
	}
	return botReply{Text: strings.TrimRight(b.String(), "\n")}
}

// --- Guided character creation ---

// CreateCharacterInput is what the create flow collects.
type CreateCharacterInput struct {
	Name       string
	Race       string
	Class      string
	Level      int
	CampaignID int64
}

// CreatedCharacter is the subset of a new character the bot reports back.
type CreatedCharacter struct {
	ID        int64
	Name      string
	Race      string
	Class     string
	Level     int
	HpMax     int
	HpCurrent int
	Ac        int
}

// CharacterCreator is wired by the server so the bot shares the HTTP
// creation path (validation, defaults, currency, campaign attach).
type CharacterCreator func(ctx context.Context, userID int64, in CreateCharacterInput) (CreatedCharacter, error)

var (
	characterCreatorMu sync.RWMutex
	characterCreator   CharacterCreator
)

// SetCharacterCreator installs the creation callback. Pass nil to disable
// creation (tests, or a server without handlers wired).
func SetCharacterCreator(fn CharacterCreator) {
	characterCreatorMu.Lock()
	defer characterCreatorMu.Unlock()
	characterCreator = fn
}

func getCharacterCreator() CharacterCreator {
	characterCreatorMu.RLock()
	defer characterCreatorMu.RUnlock()
	return characterCreator
}

const createFlowTTL = 10 * time.Minute

const (
	createStepName     = "name"
	createStepRace     = "race"
	createStepClass    = "class"
	createStepLevel    = "level"
	createStepCampaign = "campaign"
)

type createFlow struct {
	Name       string
	Race       string
	Class      string
	Level      int
	CampaignID int64
	Step       string
	UpdatedAt  time.Time
}

var (
	createFlowsMu sync.Mutex
	createFlows   = map[int64]*createFlow{}
)

func getCreateFlow(chatID int64) *createFlow {
	createFlowsMu.Lock()
	defer createFlowsMu.Unlock()
	flow := createFlows[chatID]
	if flow == nil {
		return nil
	}
	if time.Since(flow.UpdatedAt) > createFlowTTL {
		delete(createFlows, chatID)
		return nil
	}
	return flow
}

func putCreateFlow(chatID int64, flow *createFlow) {
	flow.UpdatedAt = time.Now()
	createFlowsMu.Lock()
	defer createFlowsMu.Unlock()
	createFlows[chatID] = flow
}

func flowActive(chatID int64) bool {
	return getCreateFlow(chatID) != nil
}

// abortCreateFlow reports whether a flow was dropped.
func abortCreateFlow(chatID int64) bool {
	createFlowsMu.Lock()
	defer createFlowsMu.Unlock()
	if _, ok := createFlows[chatID]; ok {
		delete(createFlows, chatID)
		return true
	}
	return false
}

func runCreate(c *cmdContext) botReply {
	if _, ok := linkedUser(c); !ok {
		return botReply{}
	}
	if getCharacterCreator() == nil {
		return botReply{Text: "Character creation is unavailable right now."}
	}
	flow := &createFlow{Step: createStepName}
	if len(c.args) > 0 {
		name := strings.TrimSpace(strings.Join(c.args, " "))
		if name != "" {
			flow.Name = name
			flow.Step = createStepRace
		}
	}
	putCreateFlow(c.chatID, flow)
	if flow.Step == createStepRace {
		return botReply{Text: fmt.Sprintf("Name: <b>%s</b>. What race is the character?", escapeHTML(flow.Name))}
	}
	return botReply{Text: "What is the character's name? (send /cancel to stop)"}
}

// feedCreateFlow consumes a text answer for the active flow.
func feedCreateFlow(c *cmdContext, text string) {
	flow := getCreateFlow(c.chatID)
	if flow == nil {
		return
	}
	answer := strings.TrimSpace(text)
	switch flow.Step {
	case createStepName:
		if answer == "" {
			c.replyWithKeyboard("What is the character's name?", nil)
			return
		}
		flow.Name = answer
		flow.Step = createStepRace
		putCreateFlow(c.chatID, flow)
		c.reply("Got it. What race is the character?")
	case createStepRace:
		if answer == "" {
			c.replyWithKeyboard("What race is the character?", nil)
			return
		}
		flow.Race = answer
		flow.Step = createStepClass
		putCreateFlow(c.chatID, flow)
		c.reply("And the class?")
	case createStepClass:
		if answer == "" {
			c.replyWithKeyboard("What class is the character?", nil)
			return
		}
		flow.Class = answer
		flow.Step = createStepLevel
		putCreateFlow(c.chatID, flow)
		c.replyWithKeyboard("What level? Tap one or send a number from 1 to 20.", levelKeyboard())
	case createStepLevel:
		level, err := strconv.Atoi(answer)
		if err != nil || level < 1 || level > 20 {
			c.reply("Please send a level between 1 and 20.")
			return
		}
		flow.Level = level
		flow.Step = createStepCampaign
		putCreateFlow(c.chatID, flow)
		uid, linked := linkedUser(c)
		if !linked {
			abortCreateFlow(c.chatID)
			return
		}
		reply := campaignStepReply(uid)
		sendBotReply(c, reply)
	case createStepCampaign:
		uid, linked := linkedUser(c)
		if !linked {
			abortCreateFlow(c.chatID)
			return
		}
		campaigns, err := campaignChoices(uid)
		if err != nil {
			middleware.LogWarn("telegram", "campaign list failed", "error", err)
			c.reply("Could not load your campaigns right now. Send /cancel and try again.")
			return
		}
		if strings.EqualFold(answer, "none") || answer == "-" {
			finishCreate(c, flow, 0)
			return
		}
		for _, cc := range campaigns {
			if strings.EqualFold(strings.TrimSpace(cc.Name), answer) {
				finishCreate(c, flow, cc.ID)
				return
			}
		}
		c.reply("Tap one of the campaigns or send \"none\".")
	default:
		abortCreateFlow(c.chatID)
	}
}

func levelKeyboard() *tgmodels.InlineKeyboardMarkup {
	levels := []int{1, 3, 5, 10, 20}
	var row []inlineButton
	for _, lvl := range levels {
		row = append(row, inlineButton{Text: strconv.Itoa(lvl), CallbackData: cbCreateLevelPrefix + strconv.Itoa(lvl)})
	}
	return inlineKeyboard(row)
}

func campaignStepReply(uid int64) botReply {
	campaigns, err := campaignChoices(uid)
	if err != nil {
		middleware.LogWarn("telegram", "campaign list failed", "error", err)
		campaigns = nil
	}
	var rows [][]inlineButton
	for _, cc := range campaigns {
		rows = append(rows, []inlineButton{{Text: cc.Name, CallbackData: cbCreateCampaignPrefix + strconv.FormatInt(cc.ID, 10)}})
	}
	rows = append(rows, []inlineButton{{Text: "No campaign", CallbackData: cbCreateCampaignNone}})
	return botReply{Text: "Attach the character to a campaign? Tap one or send \"none\".", Keyboard: inlineKeyboard(rows...)}
}

// createLevelChoice handles the level picker button.
func createLevelChoice(c *cmdContext, level int) botReply {
	flow := getCreateFlow(c.chatID)
	if flow == nil {
		return botReply{Text: "This creation flow expired. Send /create to start again."}
	}
	if level < 1 || level > 20 {
		return botReply{Text: "Please send a level between 1 and 20."}
	}
	flow.Level = level
	flow.Step = createStepCampaign
	putCreateFlow(c.chatID, flow)
	uid, ok := linkedUser(c)
	if !ok {
		abortCreateFlow(c.chatID)
		return botReply{}
	}
	return campaignStepReply(uid)
}

// createCampaignChoice handles a campaign button (0 = no campaign).
func createCampaignChoice(c *cmdContext, campaignID int64) botReply {
	flow := getCreateFlow(c.chatID)
	if flow == nil {
		return botReply{Text: "This creation flow expired. Send /create to start again."}
	}
	uid, ok := linkedUser(c)
	if !ok {
		abortCreateFlow(c.chatID)
		return botReply{}
	}
	if campaignID != 0 {
		if _, member := campaignRole(uid, campaignID); !member {
			return botReply{Text: "Campaign not found."}
		}
	}
	finishCreate(c, flow, campaignID)
	return botReply{}
}

func finishCreate(c *cmdContext, flow *createFlow, campaignID int64) {
	uid, ok := linkedUser(c)
	if !ok {
		abortCreateFlow(c.chatID)
		return
	}
	creator := getCharacterCreator()
	if creator == nil {
		abortCreateFlow(c.chatID)
		c.reply("Character creation is unavailable right now.")
		return
	}
	flow.CampaignID = campaignID
	abortCreateFlow(c.chatID)

	created, err := creator(c.ctx, uid, CreateCharacterInput{
		Name:       flow.Name,
		Race:       flow.Race,
		Class:      flow.Class,
		Level:      flow.Level,
		CampaignID: campaignID,
	})
	if err != nil {
		middleware.LogWarn("telegram", "character creation failed", "error", err)
		c.reply("Could not create the character: " + escapeHTML(err.Error()))
		return
	}
	if _, err := db.DB.Exec(`INSERT INTO telegram_character_claims (telegram_user_id, character_id, created_at)
		VALUES (?, ?, datetime('now'))
		ON CONFLICT(telegram_user_id) DO UPDATE SET character_id = excluded.character_id, created_at = excluded.created_at`,
		c.tgUserID, created.ID); err != nil {
		middleware.LogWarn("telegram", "claim after create failed", "error", err)
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("✅ Created <b>%s</b> — %s %s · Level %d (HP %d/%d, AC %d)\n",
		escapeHTML(created.Name), escapeHTML(created.Race), escapeHTML(created.Class), created.Level, created.HpCurrent, created.HpMax, created.Ac))
	if campaignID != 0 {
		if name, ok := campaignName(campaignID); ok {
			b.WriteString("Attached to " + escapeHTML(name) + ".\n")
		}
	}
	b.WriteString("It is now your claimed character. Use /sheet to see it.")
	c.reply(b.String())
}

func campaignName(id int64) (string, bool) {
	var name string
	if err := db.DB.QueryRow(`SELECT name FROM campaigns WHERE id = ?`, id).Scan(&name); err != nil {
		return "", false
	}
	return name, true
}
