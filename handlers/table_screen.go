package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"villum/db"
)

// Table screen models: a read-only snapshot rendered on a shared/second screen.
type tableCombatant struct {
	ID         int64    `json:"id"`
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	Initiative int      `json:"initiative"`
	HPCurrent  int      `json:"hp_current"`
	HPMax      int      `json:"hp_max"`
	AC         int      `json:"ac"`
	Conditions []string `json:"conditions"`
	IsCurrent  bool     `json:"is_current"`
}

type tableCharacter struct {
	ID         int64    `json:"id"`
	Name       string   `json:"name"`
	Class      string   `json:"class"`
	Level      int      `json:"level"`
	HPCurrent  int      `json:"hp_current"`
	HPMax      int      `json:"hp_max"`
	AC         int      `json:"ac"`
	Conditions []string `json:"conditions"`
}

type tableHandout struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Status  string `json:"status"`
}

type tableRoll struct {
	Name       string `json:"name"`
	Expression string `json:"expression"`
	Total      int    `json:"total"`
	Text       string `json:"text"`
	Timestamp  string `json:"timestamp"`
}

// splitCondList accepts the two shapes conditions are stored in: a JSON array
// (combat_entries.condition_ids) or a comma-joined list (GROUP_CONCAT).
func splitCondList(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "[]" {
		return nil
	}
	var arr []string
	if err := json.Unmarshal([]byte(raw), &arr); err == nil {
		return arr
	}
	var out []string
	for _, p := range strings.Split(raw, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// TableStateForCampaign gathers the live table snapshot: initiative order,
// party vitals, revealed handouts and recent rolls. Returns nil if the campaign
// does not exist.
func TableStateForCampaign(campaignID int64) gin.H {
	var name, party string
	if err := db.DB.QueryRow("SELECT name, COALESCE(party_name,'') FROM campaigns WHERE id=?", campaignID).Scan(&name, &party); err != nil {
		return nil
	}

	var currentID int64
	_ = db.DB.QueryRow("SELECT id FROM combat_entries WHERE campaign_id=? AND is_active=1 ORDER BY turn_order DESC LIMIT 1", campaignID).Scan(&currentID)

	combat := []tableCombatant{}
	if rows, err := db.DB.Query(`
		SELECT id, name, type, initiative_roll, initiative_mod, hp_current, hp_max, ac, COALESCE(condition_ids,'[]')
		FROM combat_entries WHERE campaign_id=?
		ORDER BY is_active DESC, (initiative_roll + initiative_mod) DESC, id DESC`, campaignID); err == nil {
		defer rows.Close()
		for rows.Next() {
			var t tableCombatant
			var raw string
			var initRoll, initMod int
			rows.Scan(&t.ID, &t.Name, &t.Type, &initRoll, &initMod, &t.HPCurrent, &t.HPMax, &t.AC, &raw)
			t.Initiative = initRoll + initMod
			t.Conditions = splitCondList(raw)
			t.IsCurrent = t.ID == currentID
			combat = append(combat, t)
		}
	}

	partyChars := []tableCharacter{}
	if rows, err := db.DB.Query(`
		SELECT c.id, c.name, COALESCE(c.class,''), c.level, c.hp_current, c.hp_max, c.ac,
			COALESCE((SELECT GROUP_CONCAT(cc.name, ', ') FROM character_conditions cc WHERE cc.character_id = c.id), '')
		FROM characters c WHERE c.campaign_id=? ORDER BY c.name`, campaignID); err == nil {
		defer rows.Close()
		for rows.Next() {
			var ch tableCharacter
			var conds string
			rows.Scan(&ch.ID, &ch.Name, &ch.Class, &ch.Level, &ch.HPCurrent, &ch.HPMax, &ch.AC, &conds)
			ch.Conditions = splitCondList(conds)
			partyChars = append(partyChars, ch)
		}
	}

	handouts := []tableHandout{}
	if rows, err := db.DB.Query(`
		SELECT title, COALESCE(content,''), status FROM campaign_knowledge
		WHERE campaign_id=? AND shared=1 ORDER BY updated_at DESC, id DESC LIMIT 12`, campaignID); err == nil {
		defer rows.Close()
		for rows.Next() {
			var h tableHandout
			rows.Scan(&h.Title, &h.Content, &h.Status)
			handouts = append(handouts, h)
		}
	}

	rolls := []tableRoll{}
	if rows, err := db.DB.Query(`
		SELECT COALESCE(c.name,''), dr.expression, dr.total, COALESCE(dr.result,''), COALESCE(dr.timestamp,'')
		FROM dice_rolls dr LEFT JOIN characters c ON c.id = dr.character_id
		WHERE c.campaign_id=? ORDER BY dr.timestamp DESC, dr.id DESC LIMIT 10`, campaignID); err == nil {
		defer rows.Close()
		for rows.Next() {
			var r tableRoll
			rows.Scan(&r.Name, &r.Expression, &r.Total, &r.Text, &r.Timestamp)
			rolls = append(rolls, r)
		}
	}

	var current any
	if currentID != 0 {
		current = currentID
	}
	return gin.H{
		"campaign":        gin.H{"id": campaignID, "name": name, "party_name": party},
		"current_turn_id": current,
		"combat":          combat,
		"party":           partyChars,
		"handouts":        handouts,
		"rolls":           rolls,
		"updated_at":      time.Now().UTC().Format(time.RFC3339),
	}
}

// GetCampaignTableState is the authenticated table snapshot used by the in-app
// second screen (any campaign member).
func GetCampaignTableState(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		WriteNotFound(c, "campaign not found")
		return
	}
	if !isCampaignMember(c, id) {
		WriteError(c, http.StatusForbidden, errAccessDenied)
		return
	}
	state := TableStateForCampaign(id)
	if state == nil {
		WriteNotFound(c, "campaign not found")
		return
	}
	WriteJSON(c, http.StatusOK, state)
}
