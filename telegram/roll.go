package telegram

import (
	"fmt"
	"strings"

	"villum/dice"
)

func runRoll(c *cmdContext) botReply {
	rollExpr := rollResolveExpr(c.args)

	pool, err := dice.NewPool(1)
	if err != nil {
		return botReply{Text: "Dice engine unavailable."}
	}
	res, err := pool.Roll(rollExpr)
	if err != nil {
		// Include res.Error if available (dice engine returns error when Error != "")
		msg := err.Error()
		if res != nil && res.Error != "" {
			msg = res.Error
		}
		if msg == "" {
			msg = err.Error()
		}
		return botReply{Text: fmt.Sprintf("Could not roll <b>%s</b>: %s", escapeHTML(rollExpr), escapeHTML(msg))}
	}
	if res.Error != "" {
		return botReply{Text: fmt.Sprintf("Could not roll <b>%s</b>: %s", escapeHTML(rollExpr), escapeHTML(res.Error))}
	}
	return botReply{Text: rollFormatResult(rollExpr, res)}
}

func rollResolveExpr(args []string) string {
	expr := strings.TrimSpace(strings.Join(args, " "))
	if expr == "" {
		return "1d20"
	}
	return expr
}

func rollFormatResult(expr string, res *dice.RollResult) string {
	total := res.Total.String()
	output := strings.TrimSpace(res.Output)
	var b strings.Builder
	fmt.Fprintf(&b, "🎲 <b>%s</b>", escapeHTML(expr))
	if output != "" {
		fmt.Fprintf(&b, "\n<code>%s</code>", escapeHTML(output))
	}
	fmt.Fprintf(&b, "\nTotal: <b>%s</b>", escapeHTML(total))
	text := b.String()
	// Stay well under Telegram's 4096 limit.
	if len([]rune(text)) > 3900 {
		text = truncateRunes(text, 3900)
	}
	return text
}
