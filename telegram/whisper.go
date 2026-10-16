package telegram

import (
	"fmt"
	"strings"

	"villum/dice"
)

func runWhisper(c *cmdContext) botReply {
	if !isGroupChat(c.chatID) {
		return botReply{Text: "Whispers only work in a group chat."}
	}
	whisperExpr := whisperResolveExpr(c.args)

	pool, err := dice.NewPool(1)
	if err != nil {
		return botReply{Text: "Dice engine unavailable."}
	}
	res, err := pool.Roll(whisperExpr)
	if err != nil {
		msg := err.Error()
		if res != nil && res.Error != "" {
			msg = res.Error
		}
		return botReply{Text: fmt.Sprintf("Could not roll <b>%s</b>: %s", escapeHTML(whisperExpr), escapeHTML(msg))}
	}
	if res.Error != "" {
		return botReply{Text: fmt.Sprintf("Could not roll <b>%s</b>: %s", escapeHTML(whisperExpr), escapeHTML(res.Error))}
	}
	text := whisperFormatResult(whisperExpr, res)
	if _, err := SendEphemeralMessage(c.chatID, c.tgUserID, text); err != nil {
		return botReply{Text: escapeHTML("Could not send whisper: " + err.Error())}
	}
	return botReply{}
}

func whisperResolveExpr(args []string) string {
	expr := strings.TrimSpace(strings.Join(args, " "))
	if expr == "" {
		return "1d20"
	}
	return expr
}

func whisperFormatResult(expr string, res *dice.RollResult) string {
	total := res.Total.String()
	output := strings.TrimSpace(res.Output)
	var b strings.Builder
	fmt.Fprintf(&b, "🎲 <b>%s</b>", escapeHTML(expr))
	if output != "" {
		fmt.Fprintf(&b, "\n<code>%s</code>", escapeHTML(output))
	}
	fmt.Fprintf(&b, "\nTotal: <b>%s</b>", escapeHTML(total))
	text := b.String()
	if len([]rune(text)) > 3900 {
		text = truncateRunes(text, 3900)
	}
	return text
}
