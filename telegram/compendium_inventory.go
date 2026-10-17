package telegram

import (
	"fmt"
	"strconv"
	"strings"

	"villum/db"
)

const cbCompendiumAddPrefix = "cadd:"

// handleCompendiumInventoryCallback handles "cadd:<token>:<idx>" taps from a
// compendium equipment detail, adding the entry to the caller's inventory.
func handleCompendiumInventoryCallback(c *cmdContext, data string) botReply {
	rest := strings.TrimPrefix(data, cbCompendiumAddPrefix)
	sep := strings.LastIndex(rest, ":")
	if sep < 0 {
		return botReply{Text: "That add link is invalid — search again."}
	}
	token := rest[:sep]
	idxStr := rest[sep+1:]
	if !searchTokenRegexp.MatchString(token) {
		return botReply{Text: "That add link is invalid — search again."}
	}
	idx, err := strconv.Atoi(idxStr)
	if err != nil {
		return botReply{Text: "That add link is invalid — search again."}
	}
	st, ok := getSearchState(token)
	if !ok {
		return botReply{Text: "This search expired, run it again."}
	}
	if idx < 0 || idx >= len(st.results) {
		return botReply{Text: "That entry is no longer available — search again."}
	}
	res := st.results[idx]
	name := res.Name
	if name == "" {
		name = "Unknown item"
	}
	charID, charName, refusal, ok := requireActionCharacter(c)
	if !ok {
		if refusal.Text != "" {
			return refusal
		}
		return botReply{}
	}
	if _, err := db.DB.Exec(
		`INSERT INTO inventory (character_id, name, quantity, category) VALUES (?, ?, 1, 'equipment')`,
		charID, name); err != nil {
		return botReply{Text: "Could not add that item right now."}
	}
	reply := renderInventory(charID, charName)
	reply.Text = fmt.Sprintf("Added <b>%s</b> to your inventory.\n\n%s", escapeHTML(name), reply.Text)
	return reply
}
