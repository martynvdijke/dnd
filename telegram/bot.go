package telegram

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	tgbot "github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"

	"villum/middleware"
)

// supervisorInterval controls how often stored settings are re-read.
// It is a variable so tests can shorten it.
var supervisorInterval = 15 * time.Second

type botState struct {
	mu                 sync.Mutex
	client             *tgbot.Bot
	token              string
	apiBase            string
	pollCancel         context.CancelFunc
	pollWebhookCleared bool
	webhookURL         string
	webhookSecret      string
	wake               chan struct{}
	started            bool
}

var bot = &botState{wake: make(chan struct{}, 1)}

func runningClient() *tgbot.Bot {
	bot.mu.Lock()
	defer bot.mu.Unlock()
	return bot.client
}

// StartBot starts the supervisor that keeps the Telegram client aligned with
// stored settings. It returns immediately; closing stop shuts the client down.
func StartBot(stop <-chan struct{}) {
	bot.mu.Lock()
	if bot.started {
		bot.mu.Unlock()
		return
	}
	bot.started = true
	bot.mu.Unlock()
	go bot.supervise(stop)
}

// NotifySettingsChanged wakes the supervisor so saved changes apply at once.
func NotifySettingsChanged() {
	select {
	case bot.wake <- struct{}{}:
	default:
	}
}

func (b *botState) supervise(stop <-chan struct{}) {
	b.reconcile()
	ticker := time.NewTicker(supervisorInterval)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			b.mu.Lock()
			b.stopClientLocked()
			b.mu.Unlock()
			return
		case <-ticker.C:
			b.reconcile()
		case <-b.wake:
			b.reconcile()
		}
	}
}

func (b *botState) reconcile() {
	s := LoadSettings()

	b.mu.Lock()
	defer b.mu.Unlock()

	if s.Token == "" || s.EffectiveMode == "off" {
		b.stopClientLocked()
		return
	}

	if b.client != nil && (b.token != s.Token || b.apiBase != s.APIBase) {
		b.stopClientLocked()
	}
	if b.client == nil {
		client, err := newBotClient(s.Token, s.APIBase, b.onUpdate)
		if err != nil {
			middleware.LogWarn("telegram", "failed to start bot client", "error", err)
			return
		}
		b.client = client
		b.token = s.Token
		b.apiBase = s.APIBase
		registerBotCommands(client)
	}

	if s.EffectiveMode == "webhook" {
		if url := WebhookURL(); url != "" && s.WebhookSecret != "" {
			b.startWebhookLocked(s, url)
			return
		}
		// Webhook transport needs both a public HTTPS BASE_URL and a secret.
		// Without them the bot would go completely silent, so fall back to
		// long polling and surface the misconfiguration for operators.
		middleware.LogWarn("telegram",
			"webhook mode is not fully configured; falling back to polling",
			"base_url_https", WebhookURL() != "",
			"webhook_secret_set", s.WebhookSecret != "")
	}

	b.startPollingLocked()
}

// startWebhookLocked switches the supervisor to webhook transport. The client
// stays alive for outbound calls; polling is stopped.
func (b *botState) startWebhookLocked(s Settings, url string) {
	if b.pollCancel != nil {
		b.pollCancel()
		b.pollCancel = nil
	}
	// A later switch back to polling must clear this webhook again.
	b.pollWebhookCleared = false
	if b.webhookURL != url || b.webhookSecret != s.WebhookSecret {
		client := b.client
		b.webhookURL = url
		b.webhookSecret = s.WebhookSecret
		go func() {
			if _, err := client.SetWebhook(context.Background(), &tgbot.SetWebhookParams{
				URL:         url,
				SecretToken: s.WebhookSecret,
			}); err != nil {
				middleware.LogWarn("telegram", "failed to register webhook", "error", err)
			}
		}()
	}
}

// startPollingLocked switches the supervisor to long polling. Before the first
// getUpdates it always clears any webhook, including one left behind by a
// previous deployment. Telegram rejects getUpdates with 409 Conflict while a
// webhook is registered, which silently stalls the bot, so polling does not
// start until the webhook is confirmed deleted.
func (b *botState) startPollingLocked() {
	if !b.pollWebhookCleared {
		if _, err := b.client.DeleteWebhook(context.Background(), &tgbot.DeleteWebhookParams{}); err != nil {
			middleware.LogWarn("telegram", "failed to clear webhook before polling; will retry", "error", err)
			return
		}
		b.pollWebhookCleared = true
	}
	if b.pollCancel == nil {
		ctx, cancel := context.WithCancel(context.Background())
		b.pollCancel = cancel
		client := b.client
		go client.Start(ctx)
	}
	b.webhookURL = ""
	b.webhookSecret = ""
}

func (b *botState) stopClientLocked() {
	if b.pollCancel != nil {
		b.pollCancel()
		b.pollCancel = nil
	}
	b.client = nil
	b.token = ""
	b.apiBase = ""
	b.pollWebhookCleared = false
	b.webhookURL = ""
	b.webhookSecret = ""
}

func (b *botState) onUpdate(ctx context.Context, _ *tgbot.Bot, upd *tgmodels.Update) {
	if upd == nil {
		return
	}
	recordUpdateOffset(upd)
	HandleUpdate(ctx, upd)
}

func recordUpdateOffset(upd *tgmodels.Update) {
	if upd != nil && upd.ID > 0 {
		_ = SetUpdateOffset(upd.ID + 1)
	}
}

// ProcessUpdate feeds a webhook update into the running client. Updates that
// arrive while no client is running are dropped, matching the old behavior
// when the bot was unconfigured.
func ProcessUpdate(upd *tgmodels.Update) {
	if upd == nil {
		return
	}
	recordUpdateOffset(upd)
	client := runningClient()
	if client == nil {
		middleware.LogWarn("telegram", "dropping webhook update: bot client not running", "update_id", upd.ID)
		return
	}
	client.ProcessUpdate(context.Background(), upd)
}

// cmdContext carries everything a command needs to answer one update.
type cmdContext struct {
	ctx       context.Context
	chatID    int64
	tgUserID  int64
	username  string
	firstName string
	name      string
	args      []string
}

type botReply struct {
	Text     string
	Keyboard *tgmodels.InlineKeyboardMarkup
}

func (c *cmdContext) reply(text string) {
	sendReply(c.chatID, text)
}

func (c *cmdContext) replyWithKeyboard(text string, kb *tgmodels.InlineKeyboardMarkup) {
	sendBotReply(c, botReply{Text: text, Keyboard: kb})
}

func sendBotReply(c *cmdContext, r botReply) {
	if r.Text == "" {
		return
	}
	if r.Keyboard != nil {
		if _, err := sendMessageWithKeyboard(c.chatID, r.Text, r.Keyboard); err != nil {
			middleware.LogWarn("telegram", "send with keyboard failed", "error", err)
		}
		return
	}
	sendReply(c.chatID, r.Text)
}

// HandleUpdate dispatches one update: callback taps, text flow steps, and
// registered commands.
func HandleUpdate(ctx context.Context, upd *tgmodels.Update) {
	if upd == nil {
		return
	}
	if upd.InlineQuery != nil {
		handleInlineQuery(ctx, upd.InlineQuery)
		return
	}
	if upd.CallbackQuery != nil {
		handleCallback(ctx, upd.CallbackQuery)
		return
	}
	if upd.MyChatMember != nil {
		handleMyChatMember(upd.MyChatMember)
		return
	}
	msg := upd.Message
	if msg == nil || msg.From == nil {
		return
	}
	text := strings.TrimSpace(msg.Text)
	if text == "" {
		return
	}
	c := &cmdContext{
		ctx:       ctx,
		chatID:    msg.Chat.ID,
		tgUserID:  msg.From.ID,
		username:  msg.From.Username,
		firstName: msg.From.FirstName,
	}

	if flowActive(c.chatID) && !strings.HasPrefix(text, "/") {
		feedCreateFlow(c, text)
		return
	}

	name, args, ok := splitCommand(text)
	if !ok {
		return
	}
	if name != "cancel" && flowActive(c.chatID) {
		abortCreateFlow(c.chatID)
	}

	cmd, found := findCommand(name)
	if !found {
		if _, _, _, linked := LookupIdentityByTelegramID(c.tgUserID); linked {
			c.reply(helpText())
		} else {
			c.reply(linkingHelpText())
		}
		return
	}
	c.name = cmd.name
	c.args = args
	sendBotReply(c, execute(c, cmd))
}

// splitCommand parses "/name@BotName args" into a bare lower-case name.
func splitCommand(text string) (string, []string, bool) {
	fields := strings.Fields(strings.TrimSpace(text))
	if len(fields) == 0 || !strings.HasPrefix(fields[0], "/") {
		return "", nil, false
	}
	name := strings.ToLower(strings.TrimPrefix(fields[0], "/"))
	if i := strings.Index(name, "@"); i >= 0 {
		name = name[:i]
	}
	return name, fields[1:], true
}

// handleMyChatMember welcomes the bot when it is added to a group chat.
func handleMyChatMember(mc *tgmodels.ChatMemberUpdated) {
	if mc == nil {
		return
	}
	if mc.Chat.Type != tgmodels.ChatTypeGroup && mc.Chat.Type != tgmodels.ChatTypeSupergroup {
		return
	}
	if !botMemberStatus(mc.NewChatMember.Type) || !botWasAbsent(mc.OldChatMember.Type) {
		return
	}
	sendReply(mc.Chat.ID, groupWelcomeText(mc.Chat.ID))
}

func botMemberStatus(t tgmodels.ChatMemberType) bool {
	switch t {
	case tgmodels.ChatMemberTypeOwner, tgmodels.ChatMemberTypeAdministrator, tgmodels.ChatMemberTypeMember:
		return true
	default:
		return false
	}
}

func botWasAbsent(t tgmodels.ChatMemberType) bool {
	switch t {
	case "", tgmodels.ChatMemberTypeLeft, tgmodels.ChatMemberTypeBanned:
		return true
	default:
		return false
	}
}

// groupWelcomeText greets a group and points at the campaign binding step.
func groupWelcomeText(chatID int64) string {
	if cc, ok := boundCampaign(chatID); ok {
		return fmt.Sprintf("🤖 Thanks for adding me! This chat is connected to <b>%s</b>.\n\nTry /items, /quests, /visits or /stats.", escapeHTML(cc.Name))
	}
	return "🤖 Thanks for adding me!\n\nA campaign DM can connect this chat to a campaign in Villum under <b>Campaign → Telegram</b>. After that, /items, /quests, /visits and /stats work here.\n\nUse /help to see everything I can do."
}
