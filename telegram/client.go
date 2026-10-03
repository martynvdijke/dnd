package telegram

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	tgbot "github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
)

var Version = "0.0.0-dev"

const telegramHTTPTimeout = 30 * time.Second

// RetryAfterError mirrors Telegram's rate-limit response for callers that
// schedule retries (recap delivery).
type RetryAfterError struct {
	After time.Duration
	Desc  string
}

func (e *RetryAfterError) Error() string { return fmt.Sprintf("retry after %s: %s", e.After, e.Desc) }

// translateBotError maps library errors onto the package's retry type.
func translateBotError(err error) error {
	if err == nil {
		return nil
	}
	var rate *tgbot.TooManyRequestsError
	if errors.As(err, &rate) {
		return &RetryAfterError{After: time.Duration(rate.RetryAfter) * time.Second, Desc: rate.Message}
	}
	return err
}

func boolPtr(v bool) *bool { return &v }

// newBotClient builds a library client pointed at the configured API base.
// When handler is nil the client is only used for outbound calls.
func newBotClient(token, apiBase string, handler tgbot.HandlerFunc) (*tgbot.Bot, error) {
	if token == "" {
		return nil, errors.New("telegram not configured")
	}
	opts := []tgbot.Option{
		tgbot.WithDefaultHandler(handler),
		tgbot.WithHTTPClient(telegramHTTPTimeout, &http.Client{Timeout: telegramHTTPTimeout}),
	}
	if handler == nil {
		opts = append(opts, tgbot.WithSkipGetMe())
	} else {
		opts = append(opts, tgbot.WithAllowedUpdates(tgbot.AllowedUpdates{
			tgmodels.AllowedUpdateMessage,
			tgmodels.AllowedUpdateCallbackQuery,
			tgmodels.AllowedUpdateMyChatMember,
			tgmodels.AllowedUpdateInlineQuery,
		}))
	}
	if apiBase != "" && apiBase != defaultAPIBase {
		opts = append(opts, tgbot.WithServerURL(apiBase))
	}
	return tgbot.New(token, opts...)
}

func newTransientClient() (*tgbot.Bot, error) {
	return newBotClient(GetBotToken(), GetAPIBase(), nil)
}

// withBot runs fn against the running client, or a short-lived client built
// from stored settings when the supervisor is stopped.
func withBot(ctx context.Context, fn func(ctx context.Context, c *tgbot.Bot) error) error {
	if c := runningClient(); c != nil {
		return translateBotError(fn(ctx, c))
	}
	c, err := newTransientClient()
	if err != nil {
		return err
	}
	return translateBotError(fn(ctx, c))
}

func SendMessage(chatID int64, text string) (int64, error) {
	return sendMessage(chatID, text, nil)
}

func sendMessageWithKeyboard(chatID int64, text string, kb *tgmodels.InlineKeyboardMarkup) (int64, error) {
	return sendMessage(chatID, text, kb)
}

func sendMessage(chatID int64, text string, kb *tgmodels.InlineKeyboardMarkup) (int64, error) {
	var msgID int64
	err := withBot(context.Background(), func(ctx context.Context, c *tgbot.Bot) error {
		msg, err := c.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID:             chatID,
			Text:               text,
			ParseMode:          tgmodels.ParseModeHTML,
			LinkPreviewOptions: &tgmodels.LinkPreviewOptions{IsDisabled: boolPtr(true)},
			ReplyMarkup:        kb,
		})
		if err != nil {
			return err
		}
		if msg != nil {
			msgID = int64(msg.ID)
		}
		return nil
	})
	return msgID, err
}

func SendDocument(chatID int64, filename, content string) (int64, error) {
	var msgID int64
	err := withBot(context.Background(), func(ctx context.Context, c *tgbot.Bot) error {
		msg, err := c.SendDocument(ctx, &tgbot.SendDocumentParams{
			ChatID: chatID,
			Document: &tgmodels.InputFileUpload{
				Filename: filename,
				Data:     strings.NewReader(content),
			},
		})
		if err != nil {
			return err
		}
		if msg != nil {
			msgID = int64(msg.ID)
		}
		return nil
	})
	return msgID, err
}

func SetWebhook(url, secret string) error {
	return withBot(context.Background(), func(ctx context.Context, c *tgbot.Bot) error {
		_, err := c.SetWebhook(ctx, &tgbot.SetWebhookParams{URL: url, SecretToken: secret})
		return err
	})
}

func DeleteWebhook() error {
	return withBot(context.Background(), func(ctx context.Context, c *tgbot.Bot) error {
		_, err := c.DeleteWebhook(ctx, &tgbot.DeleteWebhookParams{})
		return err
	})
}

func GetChat(chatID int64) (*tgmodels.ChatFullInfo, error) {
	var chat *tgmodels.ChatFullInfo
	err := withBot(context.Background(), func(ctx context.Context, c *tgbot.Bot) error {
		var err error
		chat, err = c.GetChat(ctx, &tgbot.GetChatParams{ChatID: chatID})
		return err
	})
	return chat, err
}

func GetMe() (*tgmodels.User, error) {
	var user *tgmodels.User
	err := withBot(context.Background(), func(ctx context.Context, c *tgbot.Bot) error {
		var err error
		user, err = c.GetMe(ctx)
		return err
	})
	return user, err
}

// AnswerCallback acknowledges a callback tap so clients stop the spinner.
func AnswerCallback(callbackID string) error {
	return withBot(context.Background(), func(ctx context.Context, c *tgbot.Bot) error {
		_, err := c.AnswerCallbackQuery(ctx, &tgbot.AnswerCallbackQueryParams{CallbackQueryID: callbackID})
		return err
	})
}

func AnswerInlineQuery(ctx context.Context, queryID string, results []tgmodels.InlineQueryResult, cacheTime int, isPersonal bool, button *tgmodels.InlineQueryResultsButton) error {
	return withBot(ctx, func(ctx context.Context, c *tgbot.Bot) error {
		_, err := c.AnswerInlineQuery(ctx, &tgbot.AnswerInlineQueryParams{InlineQueryID: queryID, Results: results, CacheTime: cacheTime, IsPersonal: isPersonal, Button: button})
		return err
	})
}

var (
	botUsernameCache string
	botUsernameOnce  sync.Once
)

// GetBotUsername resolves the bot username once, honoring the
// TELEGRAM_BOT_USERNAME override.
func GetBotUsername() string {
	if v := os.Getenv("TELEGRAM_BOT_USERNAME"); v != "" {
		return v
	}
	botUsernameOnce.Do(func() {
		if u, err := GetMe(); err == nil && u != nil && u.Username != "" {
			botUsernameCache = u.Username
		}
	})
	return botUsernameCache
}

// ResetBotUsernameCache is for tests.
func ResetBotUsernameCache() {
	botUsernameOnce = sync.Once{}
	botUsernameCache = ""
}
