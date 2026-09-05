package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"entgo.io/ent/dialect/sql"
	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/ent"
	"villum/ent/campaign"
	"villum/ent/campaignmember"
	"villum/ent/character"
	"villum/ent/user"
	"villum/models"
)

func ListCampaigns(c *gin.Context) {
	userID, _ := c.Get("user_id")
	currentUID := userID.(int64)
	ctx := c.Request.Context()
	camps, err := db.Client.Campaign.Query().
		Where(campaign.Or(
			campaign.UserID(currentUID),
			campaign.HasMembersWith(campaignmember.UserID(currentUID)),
		)).
		Order(campaign.ByName()).
		All(ctx)
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	type CampaignWithRole struct {
		models.Campaign
		MyRole string `json:"my_role"`
	}
	var out = make([]CampaignWithRole, 0)
	for _, ca := range camps {
		myRole := "dm"
		if ca.UserID != currentUID {
			m, err := db.Client.CampaignMember.Query().
				Where(campaignmember.And(campaignmember.CampaignID(ca.ID), campaignmember.UserID(currentUID))).
				Only(ctx)
			if err == nil {
				myRole = m.Role
			}
		}
		out = append(out, CampaignWithRole{
			Campaign: models.Campaign{
				ID: ca.ID, UserID: ca.UserID, Name: ca.Name, PartyName: ca.PartyName,
				Description: ca.Description, DMNotes: ca.DmNotes, CreatedAt: ca.CreatedAt,
			},
			MyRole: myRole,
		})
	}
	WriteJSON(c, http.StatusOK, out)
}

func CreateCampaign(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var ca models.Campaign
	if !BindOr400(c, &ca) {
		return
	}
	if strings.TrimSpace(ca.Name) == "" {
		WriteError(c, http.StatusBadRequest, errNameRequired)
		return
	}
	result, err := db.Client.Campaign.Create().
		SetUserID(userID.(int64)).
		SetName(ca.Name).
		SetPartyName(ca.PartyName).
		SetDescription(ca.Description).
		SetDmNotes(ca.DMNotes).
		Save(c.Request.Context())
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	db.Client.CampaignMember.Create().
		SetCampaignID(result.ID).
		SetUserID(userID.(int64)).
		SetRole("dm").
		OnConflict(sql.ResolveWithIgnore()).
		Exec(c.Request.Context())
	WriteJSON(c, http.StatusCreated, gin.H{"id": result.ID, "name": ca.Name, "party_name": ca.PartyName})
}

var errNameRequired = strErr("name is required")

type strErr string

func (e strErr) Error() string { return string(e) }

func UpdateCampaign(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	ctx := c.Request.Context()
	ca, err := db.Client.Campaign.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			WriteNotFound(c, "campaign not found")
			return
		}
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	if role != "admin" && ca.UserID != userID {
		WriteError(c, http.StatusForbidden, errAccessDenied)
		return
	}
	var camp models.Campaign
	if !BindOr400(c, &camp) {
		return
	}
	db.Client.Campaign.UpdateOneID(id).
		SetName(camp.Name).
		SetPartyName(camp.PartyName).
		SetDescription(camp.Description).
		SetDmNotes(camp.DMNotes).
		Save(ctx)
	WriteJSON(c, http.StatusOK, gin.H{"ok": true})
}

func DeleteCampaign(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	ctx := c.Request.Context()
	ca, err := db.Client.Campaign.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			WriteNotFound(c, "campaign not found")
			return
		}
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	if role != "admin" && ca.UserID != userID {
		WriteError(c, http.StatusForbidden, errAccessDenied)
		return
	}
	db.Client.Character.Update().Where(character.CampaignID(id)).ClearCampaignID().Save(ctx)
	db.Client.Campaign.DeleteOneID(id).Exec(ctx)
	WriteJSON(c, http.StatusOK, gin.H{"ok": true})
}

// isCampaignMember reports whether the requester may manage the campaign.
func isCampaignMember(c *gin.Context, campaignID int64) bool {
	return IsCampaignMemberGin(c, campaignID)
}

func campaignMemberUserIDs(c *gin.Context, campaignID int64) ([]int64, error) {
	ctx := c.Request.Context()
	ca, err := db.Client.Campaign.Query().
		Where(campaign.ID(campaignID)).
		Select(campaign.FieldUserID).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	ms, err := db.Client.CampaignMember.Query().
		Where(campaignmember.CampaignID(campaignID)).
		Select(campaignmember.FieldUserID).
		All(ctx)
	if err != nil {
		return nil, err
	}
	ids := []int64{ca.UserID}
	seen := map[int64]bool{ca.UserID: true}
	for _, m := range ms {
		if !seen[m.UserID] {
			seen[m.UserID] = true
			ids = append(ids, m.UserID)
		}
	}
	return ids, nil
}

type RosterCandidate struct {
	ID int64 `json:"id"`; UserID int64 `json:"user_id"`; OwnerUsername string `json:"owner_username"`
	Name string `json:"name"`; Race string `json:"race"`; Class string `json:"class"`; Level int `json:"level"`
	PortraitURL string `json:"portrait_url,omitempty"`; CharacterType string `json:"character_type"`; Owned bool `json:"owned"`; InRoster bool `json:"in_roster"`
}

func ListCampaignCharacterCandidates(c *gin.Context) {
	campaignID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		WriteError(c, http.StatusBadRequest, strErr("invalid campaign id"))
		return
	}
	ctx := c.Request.Context()
	userID, _ := c.Get("user_id")
	currentUID, _ := userID.(int64)
	if _, err := db.Client.Campaign.Get(ctx, campaignID); ent.IsNotFound(err) {
		WriteNotFound(c, "campaign not found")
		return
	} else if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	if !isCampaignMember(c, campaignID) {
		WriteError(c, http.StatusForbidden, errAccessDenied)
		return
	}
	memberIDs, err := campaignMemberUserIDs(c, campaignID)
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	users, err := db.Client.User.Query().Where(user.IDIn(memberIDs...)).All(ctx)
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	usernames := make(map[int64]string, len(users))
	for _, u := range users {
		usernames[u.ID] = u.Username
	}
	chars, err := db.Client.Character.Query().
		Where(character.UserIDIn(memberIDs...), character.Or(character.CampaignIDIsNil(), character.CampaignIDEQ(campaignID))).
		Order(character.ByName()).All(ctx)
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	out := make([]RosterCandidate, 0, len(chars))
	for _, ch := range chars {
		out = append(out, RosterCandidate{
			ID: ch.ID, UserID: ch.UserID, OwnerUsername: usernames[ch.UserID],
			Name: ch.Name, Race: ch.Race, Class: ch.Class, Level: ch.Level,
			PortraitURL: ch.PortraitURL, CharacterType: ch.CharacterType,
			Owned: ch.UserID == currentUID, InRoster: ch.CampaignID != 0 && ch.CampaignID == campaignID,
		})
	}
	WriteJSON(c, http.StatusOK, out)
}

func AddCampaignCharacter(c *gin.Context) {
	campaignID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		WriteError(c, http.StatusBadRequest, strErr("invalid campaign id"))
		return
	}
	var req struct{ CharacterID int64 `json:"character_id"` }
	if !BindOr400(c, &req) || req.CharacterID == 0 {
		if req.CharacterID == 0 {
			WriteError(c, http.StatusBadRequest, strErr("character_id required"))
		}
		return
	}
	ctx := c.Request.Context()
	userID, _ := c.Get("user_id")
	currentUID, _ := userID.(int64)
	if _, err := db.Client.Campaign.Get(ctx, campaignID); ent.IsNotFound(err) {
		WriteNotFound(c, "campaign not found")
		return
	} else if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	if !isCampaignMember(c, campaignID) {
		WriteError(c, http.StatusForbidden, errAccessDenied)
		return
	}
	ch, err := db.Client.Character.Get(ctx, req.CharacterID)
	if ent.IsNotFound(err) {
		WriteNotFound(c, "character not found")
		return
	}
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	if ch.CampaignID != 0 && ch.CampaignID != campaignID {
		WriteError(c, http.StatusConflict, strErr("character already assigned to another campaign"))
		return
	}
	if ch.UserID != currentUID {
		memberIDs, err := campaignMemberUserIDs(c, campaignID)
		if err != nil {
			WriteError(c, http.StatusInternalServerError, err)
			return
		}
		allowed := false
		for _, id := range memberIDs {
			if id == ch.UserID {
				allowed = true
				break
			}
		}
		if !allowed {
			WriteError(c, http.StatusBadRequest, strErr("character is not owned by you or a campaign member"))
			return
		}
	}
	if err := db.Client.Character.UpdateOneID(req.CharacterID).SetCampaignID(campaignID).Exec(ctx); err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	WriteJSON(c, http.StatusCreated, gin.H{"ok": true})
}

func RemoveCampaignCharacter(c *gin.Context) {
	campaignID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		WriteError(c, http.StatusBadRequest, strErr("invalid campaign id"))
		return
	}
	characterID, err := strconv.ParseInt(c.Param("characterId"), 10, 64)
	if err != nil {
		WriteError(c, http.StatusBadRequest, strErr("invalid character id"))
		return
	}
	ctx := c.Request.Context()
	if _, err := db.Client.Campaign.Get(ctx, campaignID); ent.IsNotFound(err) {
		WriteNotFound(c, "campaign not found")
		return
	} else if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	if !isCampaignMember(c, campaignID) {
		WriteError(c, http.StatusForbidden, errAccessDenied)
		return
	}
	ch, err := db.Client.Character.Get(ctx, characterID)
	if ent.IsNotFound(err) {
		WriteNotFound(c, "character not found")
		return
	}
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	if ch.CampaignID != campaignID {
		WriteJSON(c, http.StatusOK, gin.H{"ok": true})
		return
	}
	if err := db.Client.Character.UpdateOneID(characterID).ClearCampaignID().Exec(ctx); err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	WriteJSON(c, http.StatusOK, gin.H{"ok": true})
}

type CampaignMemberResponse struct {
	UserID int64 `json:"user_id"`; Username string `json:"username"`; Role string `json:"role"`
}

func ListCampaignMembers(c *gin.Context) {
	campaignID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	members, err := db.Client.CampaignMember.Query().Where(campaignmember.CampaignID(campaignID)).WithUser().All(c.Request.Context())
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	var out []CampaignMemberResponse
	for _, m := range members {
		username := ""
		if m.Edges.User != nil {
			username = m.Edges.User.Username
		}
		out = append(out, CampaignMemberResponse{UserID: m.UserID, Username: username, Role: m.Role})
	}
	WriteJSON(c, http.StatusOK, out)
}

func AddCampaignMember(c *gin.Context) {
	campaignID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	ctx := c.Request.Context()
	ca, err := db.Client.Campaign.Get(ctx, campaignID)
	if err != nil {
		if ent.IsNotFound(err) {
			WriteNotFound(c, "campaign not found")
			return
		}
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	if role != "admin" && ca.UserID != userID {
		WriteError(c, http.StatusForbidden, strErr("only the campaign owner can add members"))
		return
	}
	var req struct{ Username string `json:"username"` }
	if !BindOr400(c, &req) || req.Username == "" {
		if req.Username == "" {
			WriteError(c, http.StatusBadRequest, strErr("username required"))
		}
		return
	}
	targetUser, err := db.Client.User.Query().Where(user.Username(req.Username)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			WriteNotFound(c, "user not found")
			return
		}
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	_, err = db.Client.CampaignMember.Create().SetCampaignID(campaignID).SetUserID(targetUser.ID).SetRole("player").OnConflict(sql.ResolveWithIgnore()).ID(ctx)
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	WriteJSON(c, http.StatusOK, gin.H{"ok": true})
}

func SetCampaignMemberRole(c *gin.Context) {
	campaignID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	targetID, _ := strconv.ParseInt(c.Param("userId"), 10, 64)
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	ctx := c.Request.Context()
	ca, err := db.Client.Campaign.Get(ctx, campaignID)
	if err != nil {
		if ent.IsNotFound(err) {
			WriteNotFound(c, "campaign not found")
			return
		}
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	if role != "admin" && ca.UserID != userID {
		WriteError(c, http.StatusForbidden, strErr("only the campaign owner can change roles"))
		return
	}
	var req struct{ Role string `json:"role"` }
	if !BindOr400(c, &req) || (req.Role != "dm" && req.Role != "player") {
		if req.Role != "dm" && req.Role != "player" {
			WriteError(c, http.StatusBadRequest, strErr("role must be 'dm' or 'player'"))
		}
		return
	}
	db.Client.CampaignMember.Update().Where(campaignmember.And(campaignmember.CampaignID(campaignID), campaignmember.UserID(targetID))).SetRole(req.Role).Save(ctx)
	WriteJSON(c, http.StatusOK, gin.H{"ok": true})
}

func RemoveCampaignMember(c *gin.Context) {
	campaignID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	targetID, _ := strconv.ParseInt(c.Param("userId"), 10, 64)
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	ctx := c.Request.Context()
	ca, err := db.Client.Campaign.Get(ctx, campaignID)
	if err != nil {
		if ent.IsNotFound(err) {
			WriteNotFound(c, "campaign not found")
			return
		}
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	if role != "admin" && ca.UserID != userID {
		WriteError(c, http.StatusForbidden, strErr("only the campaign owner can remove members"))
		return
	}
	db.Client.CampaignMember.Delete().Where(campaignmember.And(campaignmember.CampaignID(campaignID), campaignmember.UserID(targetID))).Exec(ctx)
	WriteJSON(c, http.StatusOK, gin.H{"ok": true})
}

func SearchUsers(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		WriteJSON(c, http.StatusOK, []struct{}{})
		return
	}
	users, err := db.Client.User.Query().Where(user.UsernameContainsFold(q)).Order(user.ByUsername()).Limit(20).All(c.Request.Context())
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	type UserResult struct{ ID int64 `json:"id"`; Username string `json:"username"` }
	var out []UserResult
	for _, u := range users {
		out = append(out, UserResult{ID: u.ID, Username: u.Username})
	}
	WriteJSON(c, http.StatusOK, out)
}
