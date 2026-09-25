package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ambienceTracks are the synthesized ambience loops the live table can play.
// The client maps each name to a Web Audio graph; keep both sides in sync.
var ambienceTracks = []string{"rain", "fire", "wind", "tavern", "combat", "forest", "dungeon", "waves"}

func validAmbienceTrack(track string) bool {
	for _, t := range ambienceTracks {
		if t == track {
			return true
		}
	}
	return false
}

// SetCampaignAmbience broadcasts a play/stop ambience command to the campaign's
// members so every connected screen (or a TV) shares the same soundscape.
func SetCampaignAmbience(c *gin.Context) {
	campaignID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if campaignID <= 0 {
		WriteNotFound(c, "campaign not found")
		return
	}
	if !isCampaignMember(c, campaignID) {
		WriteError(c, http.StatusForbidden, errAccessDenied)
		return
	}

	var req struct {
		Track  string  `json:"track"`
		Action string  `json:"action"`
		Volume float64 `json:"volume"`
	}
	if !BindOr400(c, &req) {
		return
	}
	if req.Action == "" {
		req.Action = "play"
	}
	if req.Action != "play" && req.Action != "stop" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "action must be play or stop"})
		return
	}
	if req.Action == "play" && !validAmbienceTrack(req.Track) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown ambience track"})
		return
	}
	if req.Volume <= 0 || req.Volume > 1 {
		req.Volume = 0.5
	}

	SendAmbience(campaignID, req.Track, req.Action, req.Volume)
	WriteJSON(c, http.StatusOK, gin.H{"ok": true, "track": req.Track, "action": req.Action})
}
