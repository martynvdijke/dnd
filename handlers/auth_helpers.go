package handlers

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/ent/campaign"
	"villum/ent/campaignmember"
)

// MustGetUserID extracts user_id from gin.Context.
func MustGetUserID(c *gin.Context) (int64, bool) {
	v, ok := c.Get("user_id")
	if !ok {
		return 0, false
	}
	uid, ok := v.(int64)
	if !ok || uid == 0 {
		return 0, false
	}
	return uid, true
}

// ParseCampaignID parses :id param as campaign ID.
func ParseCampaignID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		return 0, false
	}
	return id, true
}

// IsCampaignMember reports whether userID is owner or member of campaignID.
// It mirrors isCampaignMember semantics (owner or campaign_members row).
// Admin bypass is handled by the Gin wrapper IsCampaignMemberGin.
func IsCampaignMember(ctx context.Context, campaignID, userID int64) (bool, error) {
	ca, err := db.Client.Campaign.Get(ctx, campaignID)
	if err != nil {
		return false, err
	}
	if ca.UserID == userID {
		return true, nil
	}
	count, err := db.Client.CampaignMember.Query().
		Where(campaignmember.CampaignID(campaignID), campaignmember.UserID(userID)).
		Count(ctx)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// IsCampaignMemberGin is the Gin-aware wrapper that also grants access to admins.
func IsCampaignMemberGin(c *gin.Context, campaignID int64) bool {
	role, _ := c.Get("role")
	if role == "admin" {
		return true
	}
	uid, ok := MustGetUserID(c)
	if !ok {
		return false
	}
	// Fast path: check owner via query with owner check inside IsCampaignMember
	ok2, _ := IsCampaignMember(c.Request.Context(), campaignID, uid)
	return ok2
}

// RequireCampaignMember is used by handlers that need to enforce membership.
// Returns true if allowed, otherwise writes 403 and returns false.
func RequireCampaignMember(c *gin.Context, campaignID int64) bool {
	if IsCampaignMemberGin(c, campaignID) {
		return true
	}
	WriteError(c, 403, errAccessDenied)
	return false
}

var errAccessDenied = accessDeniedError{}

type accessDeniedError struct{}

func (e accessDeniedError) Error() string { return "access denied" }

// isCampaignMemberExists checks via ent whether campaign exists and user has access
// (kept for compatibility). Prefer IsCampaignMemberGin.
func isCampaignMemberExists(c *gin.Context, campaignID int64) bool {
	return IsCampaignMemberGin(c, campaignID)
}

// ensure campaign import is used
var _ = campaign.FieldID
