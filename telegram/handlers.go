package telegram

import (
	"fmt"
	"strconv"
	"strings"
	"villum/db"
	"villum/middleware"
)

func sendReply(chatID int64, text string) {
	if _, err := SendMessage(chatID, text); err != nil {
		middleware.LogWarn("telegram", "sendReply failed", "error", err)
	}
}
func handleHelp(chatID int64) {
	sendReply(chatID, "Villum bot commands:\n/help - show this help\n/start <code> - link your Villum account\n/recap [campaign] - show recaps\n/lastrecap - show latest recap\n/unlink - unlink your account")
}
func handleStart(chatID, tgUserID, tgChatID int64, username, code string) {
	h := hashCode(strings.ToUpper(strings.TrimSpace(code)))
	userID, ok := ConsumeLinkCode(h)
	if !ok {
		sendReply(chatID, "Invalid or expired code. Generate a new one in Villum settings.")
		return
	}
	if err := UpsertIdentity(userID, tgUserID, tgChatID, username); err != nil {
		sendReply(chatID, "Linking failed, try again.")
		return
	}
	sendReply(chatID, "Linked! You can now use /recap and /lastrecap.")
}
func handleUnlink(chatID, tgUserID int64) {
	if _, _, _, ok := LookupIdentityByTelegramID(tgUserID); !ok {
		sendReply(chatID, "Not linked. Use /start <code> to link.")
		return
	}
	_ = DeleteIdentityByTelegramID(tgUserID)
	sendReply(chatID, "Unlinked.")
}
func handleRecap(chatID, tgUserID int64, args []string) {
	uid, _, _, ok := LookupIdentityByTelegramID(tgUserID)
	if !ok {
		handleHelp(chatID)
		return
	}
	// find campaigns for user
	campaignIDs := memberCampaignIDs(uid)
	if len(campaignIDs) == 0 {
		sendReply(chatID, "No campaigns found.")
		return
	}
	var targetID int64
	if len(args) > 0 {
		if n, err := strconv.ParseInt(args[0], 10, 64); err == nil {
			// check membership
			allowed := false
			for _, c := range campaignIDs {
				if c == n {
					allowed = true
					break
				}
			}
			if !allowed {
				sendReply(chatID, "Recap not found.")
				return
			}
			targetID = n
		} else {
			// try name match
			for _, c := range campaignIDs {
				var name string
				_ = db.DB.QueryRow("SELECT name FROM campaigns WHERE id=?", c).Scan(&name)
				if strings.EqualFold(name, args[0]) {
					targetID = c
					break
				}
			}
			if targetID == 0 {
				sendReply(chatID, "Recap not found.")
				return
			}
		}
	} else {
		if len(campaignIDs) == 1 {
			targetID = campaignIDs[0]
		} else {
			// picker
			var names []string
			for _, c := range campaignIDs {
				var n string
				_ = db.DB.QueryRow("SELECT name FROM campaigns WHERE id=?", c).Scan(&n)
				names = append(names, fmt.Sprintf("%d: %s", c, n))
			}
			sendReply(chatID, "Multiple campaigns. Use /recap <campaign_id>:\n"+strings.Join(names, "\n"))
			return
		}
	}
	// fetch latest recap
	var title, content string
	err := db.DB.QueryRow("SELECT title, content FROM campaign_recaps WHERE campaign_id=? ORDER BY created_at DESC LIMIT 1", targetID).Scan(&title, &content)
	if err != nil {
		sendReply(chatID, "Recap not found.")
		return
	}
	msg := title + "\n\n" + content
	for _, chunk := range chunkMessage(msg, 4096) {
		sendReply(chatID, chunk)
	}
}
func memberCampaignIDs(uid int64) []int64 {
	var ids []int64
	// owner
	rows, _ := db.DB.Query("SELECT id FROM campaigns WHERE user_id=?", uid)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var id int64
			_ = rows.Scan(&id)
			ids = append(ids, id)
		}
	}
	rows2, _ := db.DB.Query("SELECT campaign_id FROM campaign_members WHERE user_id=?", uid)
	if rows2 != nil {
		defer rows2.Close()
		for rows2.Next() {
			var id int64
			_ = rows2.Scan(&id)
			dup := false
			for _, v := range ids {
				if v == id {
					dup = true
					break
				}
			}
			if !dup {
				ids = append(ids, id)
			}
		}
	}
	return ids
}
