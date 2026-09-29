package telegram

import (
	"context"
	"strings"
)

func HandleUpdate(ctx context.Context, upd Update) {
	if upd.Message == nil || upd.Message.From == nil {
		return
	}
	text := strings.TrimSpace(upd.Message.Text)
	if text == "" {
		return
	}
	// strip bot mention @botname
	if idx := strings.Index(text, "@"); idx != -1 {
		// keep command before @, drop mention suffix for matching
		space := strings.Index(text, " ")
		if space == -1 {
			text = text[:idx]
		} else if idx < space {
			text = text[:idx] + text[space:]
		}
	}
	fields := strings.Fields(text)
	if len(fields) == 0 {
		return
	}
	cmd := strings.ToLower(fields[0])
	args := fields[1:]
	chatID := upd.Message.Chat.ID
	tgUserID := upd.Message.From.ID

	switch {
	case cmd == "/help":
		handleHelp(chatID)
	case cmd == "/start" && len(args) > 0:
		handleStart(chatID, tgUserID, upd.Message.Chat.ID, upd.Message.From.Username, args[0])
	case cmd == "/start" && len(args) == 0:
		handleHelp(chatID)
	case cmd == "/unlink":
		handleUnlink(chatID, tgUserID)
	case cmd == "/recap":
		handleRecap(chatID, tgUserID, args)
	case cmd == "/lastrecap":
		handleRecap(chatID, tgUserID, nil)
	default:
		// unknown command: if unlinked generic help else ignore
		if _, _, _, ok := LookupIdentityByTelegramID(tgUserID); !ok {
			handleHelp(chatID)
		}
	}
}
