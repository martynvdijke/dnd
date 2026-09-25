package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
)

// sceneEffectPresets are the named live-table effects a DM can attach to a scene.
// The client maps each to a CSS overlay; WLED consumes the same names.
var sceneEffectPresets = []string{"none", "fire", "smoke", "lightning", "rain", "snow", "darkness", "sparkle"}

// TriggerSceneEffect plays a scene's (or an explicitly supplied) special effect
// on the live table by broadcasting it to the campaign's members.
func TriggerSceneEffect(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req struct {
		Effect string `json:"effect"`
	}
	_ = c.ShouldBindJSON(&req)

	var sceneID, campaignID, ownerID int64
	var stored string
	err := db.DB.QueryRow(`
		SELECT s.id, COALESCE(s.special_effects,''), COALESCE(adv.campaign_id,0), adv.user_id
		FROM oneshot_scenes s
		JOIN oneshot_acts act ON s.act_id = act.id
		JOIN oneshot_adventures adv ON act.adventure_id = adv.id
		WHERE s.id = ?`, id).Scan(&sceneID, &stored, &campaignID, &ownerID)
	if err != nil {
		WriteNotFound(c, "scene not found")
		return
	}

	effect := req.Effect
	if effect == "" {
		effect = stored
	}
	if effect == "" || effect == "none" {
		WriteJSON(c, http.StatusOK, gin.H{"ok": true, "effect": ""})
		return
	}

	if campaignID > 0 {
		if !isCampaignDM(c, campaignID) {
			WriteError(c, http.StatusForbidden, errAccessDenied)
			return
		}
		SendSceneEffect(campaignID, id, effect)
	} else {
		uid, ok := MustGetUserID(c)
		if !ok || (uid != ownerID && c.GetString("role") != "admin") {
			WriteError(c, http.StatusForbidden, errAccessDenied)
			return
		}
	}

	NotifyWLEDEffect(effect)
	WriteJSON(c, http.StatusOK, gin.H{"ok": true, "effect": effect})
}
