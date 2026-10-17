package telegram

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	tgmodels "github.com/go-telegram/bot/models"

	"villum/db"
	"villum/middleware"
)

// pendingCapture holds a pending row.
type pendingCapture struct {
	ID         int64
	FileID     string
	FileUnique string
	Kind       string
	FileName   string
	Caption    string
}

// isSaveWord reports whether s begins with the /save command, tolerating the
// "/save@BotName" form Telegram sends in group chats.
func isSaveWord(s string) bool {
	if name, _, ok := splitCommand(s); ok && name == "save" {
		return true
	}
	first, _, _ := strings.Cut(strings.TrimSpace(s), " ")
	first = strings.TrimPrefix(first, "/")
	if idx := strings.Index(first, "@"); idx >= 0 {
		first = first[:idx]
	}
	return strings.EqualFold(first, "save")
}

// handleInboundMedia processes media messages according to §4.
func handleInboundMedia(ctx context.Context, msg *tgmodels.Message) bool {
	media := ExtractInboundMedia(msg)
	if media == nil {
		return false
	}
	chatID := msg.Chat.ID
	tgUserID := int64(0)
	if msg.From != nil {
		tgUserID = msg.From.ID
	}
	text := strings.TrimSpace(msg.Text)
	caption := strings.TrimSpace(media.Caption)

	// /save may arrive as the message text or as the leading words of the
	// caption. Trailing arguments are ignored here: this message carries the file
	// itself, so there is no picker choice to resolve.
	isSaveCommand := isSaveWord(text) || (caption != "" && isSaveWord(caption))

	// Try bound campaign
	if cc, ok := boundCampaign(chatID); ok {
		// In bound enabled chat
		// check auto_save or explicit save
		var autoSave bool
		var campaignID = cc.ID
		// fetch auto_save_media
		row := db.DB.QueryRow(`SELECT auto_save_media FROM campaign_telegram_settings WHERE campaign_id=?`, campaignID)
		var v sql.NullInt64
		if err := row.Scan(&v); err == nil && v.Valid && v.Int64 != 0 {
			autoSave = true
		}
		shouldSave := isSaveCommand || autoSave
		if !shouldSave {
			return true // swallowed silently
		}
		// Save immediately
		saveInboundMedia(ctx, media, campaignID, chatID)
		return true
	}

	// Not bound: DM or unbound chat - handle media storage + picker
	// If this media message itself is a /save command with caption prefix, we still treat as new media to store?
	// Spec: media in DM stores pending and replies with picker (or direct save if single campaign). Explicit /save on caption in DM?
	// We treat any media in DM/unbound as pending unless single campaign direct save.
	// But if media has /save caption, should we save directly without picker? Spec says DM: write pending then reply with picker; /save <n> consumes.
	// For single campaign, DM media saves directly. So we handle that.

	// Need member campaigns for sender
	var campaignIDs []int64
	if tgUserID != 0 {
		if uid, _, _, ok := LookupIdentityByTelegramID(tgUserID); ok {
			campaignIDs = memberCampaignIDs(uid)
		}
	}
	if len(campaignIDs) == 0 {
		// No saveable campaigns - still store pending? But picker would be empty. Just inform?
		// expire handling not needed; we still store pending but reply no campaigns
		storePendingCapture(tgUserID, chatID, media)
		if tgUserID != 0 {
			sendReply(chatID, "You have no campaigns to save to.")
		}
		return true
	}
	if len(campaignIDs) == 1 {
		// direct save
		saveInboundMedia(ctx, media, campaignIDs[0], chatID)
		return true
	}
	// multiple campaigns: store pending and send picker
	storePendingCapture(tgUserID, chatID, media)
	sendPickerReply(chatID, tgUserID, campaignIDs)
	return true
}

// handleSaveCommand handles text "/save <n>" for pending captures
func handleSaveCommand(c *cmdContext) bool {
	if c.name != "save" {
		return false
	}
	if len(c.args) == 0 {
		// Bare /save without number: if there's media? Already handled. For text picker without arg, show error?
		// Spec says /save <n> consumes. If no arg, treat as invalid.
		c.reply("Usage: /save <number> to save your pending media.")
		return true
	}
	n, err := strconv.Atoi(c.args[0])
	if err != nil || n < 1 {
		c.reply("Usage: /save <number> to save your pending media.")
		return true
	}
	// need pending capture most recent
	pc, ok := mostRecentPending(c.tgUserID)
	if !ok {
		c.reply("No pending media found — it may have expired. Please send the file again.")
		return true
	}
	// Check expiry: expires_at < now
	// mostRecentPending already filtered? we filter expired there, but reply friendly
	// resolve campaign by index
	ids := memberCampaignIDsByTg(c.tgUserID)
	if len(ids) == 0 {
		c.reply("You have no campaigns to save to.")
		return true
	}
	// Need sorted campaign list same order as picker: sorted by name? Use same order as picker (campaignChoices ordering)
	choices, err := campaignChoicesByIDs(ids)
	if err != nil {
		c.reply("Could not load campaigns.")
		return true
	}
	if n < 1 || n > len(choices) {
		c.reply(fmt.Sprintf("Invalid choice %d. Use /save 1-%d.", n, len(choices)))
		return true
	}
	chosen := choices[n-1]
	// Convert pending to InboundMedia for saving
	media := &InboundMedia{
		FileID:       pc.FileID,
		FileUniqueID: pc.FileUnique,
		Kind:         pc.Kind,
		FileName:     pc.FileName,
		Caption:      pc.Caption,
	}
	saveInboundMedia(c.ctx, media, chosen.ID, c.chatID)
	// delete pending
	deletePending(pc.ID)
	return true
}

func memberCampaignIDsByTg(tgID int64) []int64 {
	uid, _, _, ok := LookupIdentityByTelegramID(tgID)
	if !ok {
		return nil
	}
	return memberCampaignIDs(uid)
}

type choice struct {
	ID   int64
	Name string
}

func campaignChoicesByIDs(ids []int64) ([]choice, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	rows, err := db.DB.Query(`SELECT id, name FROM campaigns WHERE id IN (`+placeholders+`) ORDER BY name`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []choice
	for rows.Next() {
		var ch choice
		if err := rows.Scan(&ch.ID, &ch.Name); err != nil {
			continue
		}
		out = append(out, ch)
	}
	return out, rows.Err()
}

func storePendingCapture(tgUserID, chatID int64, media *InboundMedia) {
	now := time.Now().UTC()
	expires := now.Add(24 * time.Hour)
	_, err := db.DB.Exec(`INSERT INTO telegram_pending_captures(telegram_user_id, chat_id, file_id, file_unique_id, kind, file_name, caption, created_at, expires_at) VALUES(?,?,?,?,?,?,?,?,?)`,
		tgUserID, chatID, media.FileID, media.FileUniqueID, media.Kind, media.FileName, media.Caption, now.Format("2006-01-02 15:04:05"), expires.Format("2006-01-02 15:04:05"))
	if err != nil {
		middleware.LogWarn("telegram", "store pending failed", "error", err)
	}
}

func sendPickerReply(chatID, tgUserID int64, campaignIDs []int64) {
	choices, err := campaignChoicesByIDs(campaignIDs)
	if err != nil || len(choices) == 0 {
		sendReply(chatID, "You have no campaigns to save to.")
		return
	}
	var b strings.Builder
	b.WriteString("Choose where to save this media:\n")
	for i, ch := range choices {
		fmt.Fprintf(&b, "%d. %s\n", i+1, ch.Name)
	}
	b.WriteString("\nReply with /save <number>")
	sendReply(chatID, strings.TrimRight(b.String(), "\n"))
}

func mostRecentPending(tgUserID int64) (*pendingCapture, bool) {
	// cleanup expired first
	db.DB.Exec(`DELETE FROM telegram_pending_captures WHERE expires_at <= datetime('now')`)
	row := db.DB.QueryRow(`SELECT id, file_id, file_unique_id, kind, file_name, caption FROM telegram_pending_captures WHERE telegram_user_id=? ORDER BY created_at DESC, id DESC LIMIT 1`, tgUserID)
	var pc pendingCapture
	if err := row.Scan(&pc.ID, &pc.FileID, &pc.FileUnique, &pc.Kind, &pc.FileName, &pc.Caption); err != nil {
		return nil, false
	}
	return &pc, true
}

func deletePending(id int64) {
	db.DB.Exec(`DELETE FROM telegram_pending_captures WHERE id=?`, id)
}

func saveInboundMedia(ctx context.Context, media *InboundMedia, campaignID, chatID int64) {
	// duplicate check
	var cnt int
	err := db.DB.QueryRow(`SELECT COUNT(*) FROM uploads WHERE telegram_file_unique_id=?`, media.FileUniqueID).Scan(&cnt)
	if err == nil && cnt > 0 {
		// duplicate - silently skip, but ack?
		sendReply(chatID, "Already saved — duplicate file ignored.")
		return
	}
	data, err := DownloadFile(ctx, media.FileID)
	if err != nil {
		if strings.Contains(err.Error(), "file too large") {
			sendReply(chatID, "File too large — Telegram limits downloads to 20 MB.")
			return
		}
		middleware.LogWarn("telegram", "download failed", "error", err)
		sendReply(chatID, "Could not download that file. Please try again.")
		return
	}
	// Determine filename
	name := media.FileName
	if strings.TrimSpace(name) == "" {
		// fallback extension based on kind
		switch media.Kind {
		case "photo":
			name = "photo.jpg"
		case "video":
			name = "video.mp4"
		case "audio":
			name = "audio.mp3"
		case "voice":
			name = "voice.ogg"
		default:
			name = "file"
		}
	}
	// caption as name when non-empty? spec says use caption as upload name
	captionTrim := strings.TrimSpace(media.Caption)
	// strip leading /save from caption for name
	if strings.HasPrefix(strings.ToLower(captionTrim), "/save") {
		captionTrim = strings.TrimSpace(captionTrim[5:])
		// also handle "/save@BotName"
		if strings.HasPrefix(captionTrim, "@") {
			// remove bot name
			if idx := strings.Index(captionTrim, " "); idx >= 0 {
				captionTrim = strings.TrimSpace(captionTrim[idx+1:])
			} else {
				captionTrim = ""
			}
		}
	}
	if captionTrim != "" {
		// Use caption as filename if it looks like a name? spec: "use the caption as the upload name"
		// If caption contains extension, use it, else append original extension
		if strings.Contains(captionTrim, ".") {
			name = captionTrim
		} else {
			ext := ""
			if idx := strings.LastIndex(name, "."); idx >= 0 {
				ext = name[idx:]
			}
			if ext != "" && !strings.HasSuffix(captionTrim, ext) {
				name = captionTrim + ext
			} else {
				name = captionTrim
			}
		}
	}
	_, err = SaveMedia(MediaSaveRequest{
		Name:                 name,
		Data:                 data,
		OwnerType:            "campaign",
		OwnerID:              campaignID,
		Source:               "telegram",
		TelegramFileUniqueID: media.FileUniqueID,
	})
	if err != nil {
		// duplicate unique constraint?
		if strings.Contains(strings.ToLower(err.Error()), "unique") || strings.Contains(err.Error(), "duplicate") {
			sendReply(chatID, "Already saved — duplicate file ignored.")
			return
		}
		middleware.LogWarn("telegram", "save media failed", "error", err)
		sendReply(chatID, "Could not save that file. Please try again.")
		return
	}
	// success ack
	campaignName := ""
	db.DB.QueryRow(`SELECT name FROM campaigns WHERE id=?`, campaignID).Scan(&campaignName)
	if campaignName != "" {
		sendReply(chatID, fmt.Sprintf("Saved to %s.", campaignName))
	} else {
		sendReply(chatID, "Saved.")
	}
}
