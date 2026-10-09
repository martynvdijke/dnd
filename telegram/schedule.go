package telegram

import (
	"strings"

	"villum/db"
)

const cbSchedulePrefix = "sched:"

const schedUsageText = "Usage: /schedule <option1> | <option2> [| <option3> ...]\nExample: /schedule Friday 8pm | Saturday 2pm | Sunday 5pm"
const schedQuestionDefault = "When should we play?"
const schedRemindCallback = "sched:remind"

func schedParseOptions(c *cmdContext) []string {
	raw := strings.Join(c.args, " ")
	parts := strings.Split(raw, "|")
	var opts []string
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			opts = append(opts, t)
		}
	}
	return opts
}

func runSchedule(c *cmdContext) botReply {
	opts := schedParseOptions(c)
	if len(opts) < 2 {
		return botReply{Text: schedUsageText}
	}
	if _, ok := boundCampaign(c.chatID); !ok {
		return botReply{Text: groupBindingHelpText()}
	}
	question := schedQuestionDefault
	msgID, err := SendPoll(c.chatID, question, opts, true, 0)
	if err != nil {
		return botReply{Text: "Could not post the poll: " + escapeHTML(err.Error())}
	}
	_, _ = db.DB.Exec(
		`INSERT INTO telegram_session_polls(chat_id, message_id, poll_id, question, options) VALUES(?,?,?,?,?)`,
		c.chatID, msgID, "", question, strings.Join(opts, "|"))
	kb := inlineKeyboard([]inlineButton{{Text: "🔔 Remind players", CallbackData: schedRemindCallback}})
	_, _ = sendMessageWithKeyboard(c.chatID, "Vote above — tap 🔔 to nudge everyone.", kb)
	return botReply{}
}

func handleScheduleCallback(c *cmdContext, data string) botReply {
	switch data {
	case schedRemindCallback:
		_, _ = SendMessage(c.chatID, "🔔 Reminder: please vote on the session poll above!")
		return botReply{}
	default:
		return botReply{Text: "Unknown action."}
	}
}
