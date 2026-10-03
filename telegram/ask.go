package telegram

import (
	"fmt"
	"strings"
	"time"

	"villum/ai"
	"villum/db"
)

func runAskQuery(c *cmdContext, query string) botReply {
	q := strings.TrimSpace(query)
	if q == "" {
		return botReply{Text: "Usage: /ask <question>"}
	}
	// require linked account
	uid, _, _, linked := LookupIdentityByTelegramID(c.tgUserID)
	if !linked {
		return botReply{Text: linkingHelpText()}
	}
	// accepted: a bound group already exposes campaign data (same as /items, /quests, /visits); registry.VisibleIDs still limits results to public entities. Group binding is the authorization boundary.
	// campaign in scope: bound campaign if group-bound
	var campaignID int64
	if cc, ok := boundCampaign(c.chatID); ok {
		campaignID = cc.ID
	}
	res := ai.AnswerQuestion(c.ctx, db.DB, ai.QuestionRequest{
		Query:      q,
		CampaignID: campaignID,
		UserID:     uid,
		MaxTokens:  600,
		Timeout:    25 * time.Second,
	})
	// render HTML reply
	var b strings.Builder
	if res.Degraded {
		b.WriteString(escapeHTML(res.Message) + "\n\n")
		if len(res.Sources) == 0 {
			return botReply{Text: strings.TrimSpace(b.String())}
		}
		b.WriteString("Top matches:\n")
		for i, s := range res.Sources {
			name := s.Title
			if name == "" {
				name = s.EntityType
			}
			b.WriteString(fmt.Sprintf("\n%d. <b>%s</b> — %s", i+1, escapeHTML(name), escapeHTML(s.EntityType)))
		}
		return botReply{Text: strings.TrimSpace(b.String())}
	}
	b.WriteString("🪄 " + escapeHTML(res.Answer))
	if len(res.Sources) > 0 {
		b.WriteString("\n\nSources:")
		for i, s := range res.Sources {
			name := s.Title
			if name == "" {
				name = s.EntityType
			}
			sub := s.Subtitle
			if sub != "" {
				sub = " — " + sub
			}
			b.WriteString(fmt.Sprintf("\n%d. <b>%s</b>%s", i+1, escapeHTML(name), escapeHTML(sub)))
		}
	}
	return botReply{Text: strings.TrimSpace(b.String())}
}
