package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
)

// HandleCampaignDMScreen aggregates the live state a DM needs at the table:
// campaign notes, the active session plan, combat, party status, and ambience.
func HandleCampaignDMScreen(c *gin.Context) {
	campaignID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if campaignID <= 0 {
		WriteNotFound(c, "campaign not found")
		return
	}
	if !isCampaignDM(c, campaignID) {
		WriteError(c, http.StatusForbidden, errAccessDenied)
		return
	}

	campaign := gin.H{"id": campaignID, "name": "", "party_name": "", "dm_notes": ""}
	var name, partyName, dmNotes string
	db.DB.QueryRow("SELECT name, COALESCE(party_name,''), COALESCE(dm_notes,'') FROM campaigns WHERE id=?", campaignID).
		Scan(&name, &partyName, &dmNotes)
	campaign["name"] = name
	campaign["party_name"] = partyName
	campaign["dm_notes"] = dmNotes

	// Active session plan (most recent planned/ready/in-progress).
	var session any
	var sid int64
	var sTitle, sNotes, sEncounters, sGoals string
	err := db.DB.QueryRow(`SELECT id, title, COALESCE(dm_notes,''), COALESCE(planned_encounters,'[]'), COALESCE(player_goals,'[]')
		FROM session_plans WHERE campaign_id=? AND status IN ('planned','ready','in-progress')
		ORDER BY session_date DESC, id DESC LIMIT 1`, campaignID).
		Scan(&sid, &sTitle, &sNotes, &sEncounters, &sGoals)
	if err == nil {
		session = gin.H{"id": sid, "title": sTitle, "dm_notes": sNotes, "planned_encounters": sEncounters, "player_goals": sGoals}
	}

	// Combat entries and current turn.
	entries := []gin.H{}
	currentTurn := ""
	rows, err := db.DB.Query("SELECT id, name, type, hp_current, hp_max, ac, is_active FROM combat_entries WHERE campaign_id=? ORDER BY turn_order, id", campaignID)
	if err == nil {
		func() {
			defer rows.Close()
			for rows.Next() {
				var id int64
				var ename, etype string
				var hp, hpMax, ac, isActive int
				if rows.Scan(&id, &ename, &etype, &hp, &hpMax, &ac, &isActive) != nil {
					continue
				}
				entries = append(entries, gin.H{
					"id": id, "name": ename, "type": etype,
					"hp_current": hp, "hp_max": hpMax, "ac": ac, "is_active": isActive == 1,
				})
				if isActive == 1 && currentTurn == "" {
					currentTurn = ename
				}
			}
		}()
	}

	// Party status with active conditions.
	party := []gin.H{}
	charRows, err := db.DB.Query("SELECT id, name, hp_current, hp_max, ac FROM characters WHERE campaign_id=? ORDER BY name", campaignID)
	if err == nil {
		func() {
			defer charRows.Close()
			for charRows.Next() {
				var id int64
				var cname string
				var hp, hpMax, ac int
				if charRows.Scan(&id, &cname, &hp, &hpMax, &ac) != nil {
					continue
				}
				conditions := []string{}
				condRows, cerr := db.DB.Query("SELECT name FROM character_conditions WHERE character_id=?", id)
				if cerr == nil {
					func() {
						defer condRows.Close()
						for condRows.Next() {
							var cn string
							if condRows.Scan(&cn) == nil {
								conditions = append(conditions, cn)
							}
						}
					}()
				}
				party = append(party, gin.H{
					"id": id, "name": cname, "hp_current": hp, "hp_max": hpMax, "ac": ac, "conditions": conditions,
				})
			}
		}()
	}

	ambience := gin.H{"track": "", "action": ""}
	if v, ok := lastAmbience.Load(campaignID); ok {
		if st, ok := v.(ambienceState); ok {
			ambience["track"] = st.Track
			ambience["action"] = st.Action
		}
	}

	WriteJSON(c, http.StatusOK, gin.H{
		"campaign": campaign,
		"session":  session,
		"combat": gin.H{
			"active":       len(entries) > 0,
			"current_turn": currentTurn,
			"entries":      entries,
		},
		"party":    party,
		"ambience": ambience,
	})
}
