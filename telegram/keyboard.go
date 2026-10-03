package telegram

import (
	"context"
	"strconv"
	"strings"

	tgmodels "github.com/go-telegram/bot/models"

	"villum/middleware"
)

// Callback data namespaces.
const (
	cbSubscribeToggle      = "sub:toggle"
	cbNavPrefix            = "nav:"
	cbClaimPrefix          = "claim:"
	cbSheetPrefix          = "sheet:"
	cbCampaignPrefix       = "campaign:"
	cbCreateLevelPrefix    = "create:level:"
	cbCreateCampaignPrefix = "create:campaign:"
	cbCreateCampaignNone   = "create:campaign:none"
	cbSearchPrefix         = "srch:"
	cbSearchDetailPrefix   = "sdet:"
)

type inlineButton = tgmodels.InlineKeyboardButton

func inlineKeyboard(rows ...[]inlineButton) *tgmodels.InlineKeyboardMarkup {
	return &tgmodels.InlineKeyboardMarkup{InlineKeyboard: rows}
}

// navKeyboard mirrors HEAT's navigation: quick access to the main views.
func navKeyboard() *tgmodels.InlineKeyboardMarkup {
	return inlineKeyboard(
		[]inlineButton{
			{Text: "🧙 Characters", CallbackData: cbNavPrefix + "characters"},
			{Text: "📊 Stats", CallbackData: cbNavPrefix + "stats"},
		},
		[]inlineButton{
			{Text: "📜 Sheet", CallbackData: cbNavPrefix + "sheet"},
			{Text: "🏕 Overview", CallbackData: cbNavPrefix + "overview"},
			{Text: "📖 Latest recap", CallbackData: cbNavPrefix + "lastrecap"},
		},
	)
}

func handleCallback(ctx context.Context, cq *tgmodels.CallbackQuery) {
	if cq == nil {
		return
	}
	if err := AnswerCallback(cq.ID); err != nil {
		middleware.LogWarn("telegram", "answer callback failed", "error", err)
	}
	msg := cq.Message.Message
	if msg == nil {
		return
	}
	c := &cmdContext{
		ctx:       ctx,
		chatID:    msg.Chat.ID,
		tgUserID:  cq.From.ID,
		username:  cq.From.Username,
		firstName: cq.From.FirstName,
	}
	reply, ok := handleCallbackData(c, cq.Data)
	if !ok {
		return
	}
	sendBotReply(c, reply)
}

// handleCallbackData routes a callback payload to its handler. It is split out
// from handleCallback so tests can exercise it without a client.
func handleCallbackData(c *cmdContext, data string) (botReply, bool) {
	switch {
	case data == cbSubscribeToggle:
		return toggleSubscription(c), true

	case strings.HasPrefix(data, cbNavPrefix):
		cmd, found := findCommand(strings.TrimPrefix(data, cbNavPrefix))
		if !found {
			return botReply{}, false
		}
		c.name = cmd.name
		c.args = nil
		return execute(c, cmd), true

	case strings.HasPrefix(data, cbClaimPrefix):
		id, err := parseCallbackID(data, cbClaimPrefix)
		if err != nil {
			return botReply{}, false
		}
		return claimByID(c, id), true

	case strings.HasPrefix(data, cbSheetPrefix):
		id, err := parseCallbackID(data, cbSheetPrefix)
		if err != nil {
			return botReply{}, false
		}
		return sheetByID(c, id), true

	case strings.HasPrefix(data, cbCampaignPrefix):
		id, err := parseCallbackID(data, cbCampaignPrefix)
		if err != nil {
			return botReply{}, false
		}
		return overviewCampaignByID(c, id), true

	case strings.HasPrefix(data, cbCreateLevelPrefix):
		n, err := parseCallbackID(data, cbCreateLevelPrefix)
		if err != nil {
			return botReply{}, false
		}
		return createLevelChoice(c, int(n)), true

	case data == cbCreateCampaignNone:
		return createCampaignChoice(c, 0), true

	case strings.HasPrefix(data, cbCreateCampaignPrefix):
		id, err := parseCallbackID(data, cbCreateCampaignPrefix)
		if err != nil {
			return botReply{}, false
		}
		return createCampaignChoice(c, id), true

	case strings.HasPrefix(data, cbSearchDetailPrefix):
		return handleSearchDetailCallback(c, data)
	case strings.HasPrefix(data, cbSearchPrefix):
		return handleSearchPageCallback(c, data)
	}
	return botReply{}, false
}

func parseCallbackID(data, prefix string) (int64, error) {
	return strconv.ParseInt(strings.TrimPrefix(data, prefix), 10, 64)
}
