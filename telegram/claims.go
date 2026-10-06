package telegram

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"villum/db"
	"villum/middleware"
)

// characterChoice is a compact character reference used in listings.
type characterChoice struct {
	ID    int64
	Name  string
	Race  string
	Class string
	Level int
}

type characterClaim struct {
	TelegramUserID int64
	CharacterID    int64
}

// editableCharacterSQL mirrors handlers.canEditCharacter/isDMOfCharacter: a
// character is editable by an administrator, by its owner, by the owner of a
// campaign it belongs to, or by a campaign member with the dm role. Each `?`
// is the current user's id.
const editableCharacterSQL = `
	(EXISTS (SELECT 1 FROM users u WHERE u.id = ? AND u.role = 'admin')
	 OR c.user_id = ?
	 OR EXISTS (
		SELECT 1 FROM campaign_characters cc
		JOIN campaigns cap ON cap.id = cc.campaign_id
		WHERE cc.character_id = c.id
		  AND (cap.user_id = ?
		       OR EXISTS (
		           SELECT 1 FROM campaign_members cm
		           WHERE cm.campaign_id = cap.id AND cm.user_id = ? AND cm.role = 'dm')))
	)`

func getClaim(tgUserID int64) (characterClaim, bool, error) {
	var claim characterClaim
	err := db.DB.QueryRow(`SELECT telegram_user_id, character_id FROM telegram_character_claims WHERE telegram_user_id = ?`, tgUserID).
		Scan(&claim.TelegramUserID, &claim.CharacterID)
	if errors.Is(err, sql.ErrNoRows) {
		return characterClaim{}, false, nil
	}
	if err != nil {
		return characterClaim{}, false, err
	}
	return claim, true, nil
}

func clearClaimByTelegramUser(tgUserID int64) error {
	_, err := db.DB.Exec(`DELETE FROM telegram_character_claims WHERE telegram_user_id = ?`, tgUserID)
	return err
}

func characterName(id int64) (string, bool) {
	var name string
	if err := db.DB.QueryRow(`SELECT name FROM characters WHERE id = ?`, id).Scan(&name); err != nil {
		return "", false
	}
	return name, true
}

func characterEditable(id, uid int64) bool {
	var found int64
	err := db.DB.QueryRow(`SELECT c.id FROM characters c WHERE c.id = ? AND `+editableCharacterSQL,
		id, uid, uid, uid, uid).Scan(&found)
	return err == nil
}

// claimedCharacter resolves the caller's claim. hasClaim reports whether a
// row exists, usable whether the character can still be edited by the user.
func claimedCharacter(c *cmdContext, uid int64) (choice characterChoice, hasClaim bool, usable bool) {
	claim, found, err := getClaim(c.tgUserID)
	if err != nil {
		middleware.LogWarn("telegram", "claim lookup failed", "error", err)
		return characterChoice{}, false, false
	}
	if !found {
		return characterChoice{}, false, false
	}
	err = db.DB.QueryRow(`SELECT c.id, c.name, c.race, c.class, c.level FROM characters c
		WHERE c.id = ? AND `+editableCharacterSQL,
		claim.CharacterID, uid, uid, uid, uid).
		Scan(&choice.ID, &choice.Name, &choice.Race, &choice.Class, &choice.Level)
	if err != nil {
		return characterChoice{ID: claim.CharacterID}, true, false
	}
	return choice, true, true
}

func claimCandidates(uid int64) ([]characterChoice, error) {
	rows, err := db.DB.Query(`SELECT c.id, c.name, c.race, c.class, c.level FROM characters c
		WHERE c.character_type != 'linked' AND `+editableCharacterSQL+`
		  AND NOT EXISTS (SELECT 1 FROM telegram_character_claims tcc WHERE tcc.character_id = c.id)
		ORDER BY c.name`,
		uid, uid, uid, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []characterChoice
	for rows.Next() {
		var ch characterChoice
		if err := rows.Scan(&ch.ID, &ch.Name, &ch.Race, &ch.Class, &ch.Level); err != nil {
			continue
		}
		out = append(out, ch)
	}
	return out, rows.Err()
}

func claimByID(c *cmdContext, id int64) botReply {
	uid, ok := linkedUser(c)
	if !ok {
		return botReply{}
	}
	name, editable := characterName(id)
	if !editable || !characterEditable(id, uid) {
		return botReply{Text: "Character not found."}
	}
	var owner int64
	err := db.DB.QueryRow(`SELECT telegram_user_id FROM telegram_character_claims WHERE character_id = ?`, id).Scan(&owner)
	if err == nil && owner != c.tgUserID {
		return botReply{Text: "That character is already claimed by another Telegram user."}
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		middleware.LogWarn("telegram", "claim lookup failed", "error", err)
		return botReply{Text: "Could not claim that character right now."}
	}
	if _, err := db.DB.Exec(`INSERT INTO telegram_character_claims (telegram_user_id, character_id, created_at)
		VALUES (?, ?, datetime('now'))
		ON CONFLICT(telegram_user_id) DO UPDATE SET character_id = excluded.character_id, created_at = excluded.created_at`,
		c.tgUserID, id); err != nil {
		middleware.LogWarn("telegram", "claim failed", "error", err)
		return botReply{Text: "Could not claim that character right now."}
	}
	return botReply{Text: fmt.Sprintf("✅ You are now playing <b>%s</b>. Try /sheet or /stats.", escapeHTML(name))}
}

func runClaim(c *cmdContext) botReply {
	uid, ok := linkedUser(c)
	if !ok {
		return botReply{}
	}
	if len(c.args) > 0 {
		query := strings.TrimSpace(strings.Join(c.args, " "))
		if id, err := strconv.ParseInt(query, 10, 64); err == nil {
			return claimByID(c, id)
		}
		candidates, err := claimCandidates(uid)
		if err != nil {
			middleware.LogWarn("telegram", "claim candidates failed", "error", err)
			return botReply{Text: "Could not load your characters right now."}
		}
		for _, ch := range candidates {
			if strings.EqualFold(strings.TrimSpace(ch.Name), query) {
				return claimByID(c, ch.ID)
			}
		}
		return botReply{Text: "Character not found."}
	}

	current, hasClaim, usable := claimedCharacter(c, uid)
	candidates, err := claimCandidates(uid)
	if err != nil {
		middleware.LogWarn("telegram", "claim candidates failed", "error", err)
		return botReply{Text: "Could not load your characters right now."}
	}
	var b strings.Builder
	if hasClaim && usable {
		b.WriteString(fmt.Sprintf("Currently claimed: <b>%s</b>\n\n", escapeHTML(current.Name)))
	} else if hasClaim {
		b.WriteString("Your claimed character is no longer available (use /unclaim).\n\n")
	}
	if len(candidates) == 0 {
		b.WriteString("No claimable characters. Use /create to make one.")
		return botReply{Text: b.String()}
	}
	b.WriteString("Tap a character to claim it:")
	var rows [][]inlineButton
	for _, ch := range candidates {
		label := fmt.Sprintf("%s — %s %s Lv%d", ch.Name, ch.Race, ch.Class, ch.Level)
		rows = append(rows, []inlineButton{{Text: label, CallbackData: cbClaimPrefix + strconv.FormatInt(ch.ID, 10)}})
	}
	return botReply{Text: b.String(), Keyboard: inlineKeyboard(rows...)}
}

func runUnclaim(c *cmdContext) botReply {
	uid, ok := linkedUser(c)
	if !ok {
		return botReply{}
	}
	current, hasClaim, _ := claimedCharacter(c, uid)
	if !hasClaim {
		return botReply{Text: "You have not claimed a character."}
	}
	if err := clearClaimByTelegramUser(c.tgUserID); err != nil {
		middleware.LogWarn("telegram", "unclaim failed", "error", err)
		return botReply{Text: "Could not release the claim right now."}
	}
	name := current.Name
	if name == "" {
		name = "your character"
	}
	return botReply{Text: fmt.Sprintf("Released <b>%s</b>. Use /claim to pick another.", escapeHTML(name))}
}
