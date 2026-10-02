package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/ent/campaign"
	"villum/ent/location"
	"villum/ent/npc"
)

// draftImportResult is the response shape for both create and replace.
type draftImportResult struct {
	ID      int64               `json:"id"`
	URL     string              `json:"url"`
	Entity  string              `json:"entity_type"`
	Name    string              `json:"name"`
	Updated bool                `json:"updated"`
	Counts  *oneShotDraftCounts `json:"counts,omitempty"`
}

// HandleImportDraftJSON creates or replaces an entity from a pasted JSON
// draft. POST /api/ai/import and POST /api/oneshot-adventures/import (where the
// entity type defaults to one-shot).
//
// The pasted text may be a bare draft object or the assistant's full reply
// envelope ({"status","message","draft"}), with or without markdown fences.
// Without entity_id the entity is created; with it the existing entity is
// replaced by the draft's content.
func HandleImportDraftJSON(c *gin.Context) {
	var req struct {
		EntityType string `json:"entity_type"`
		CampaignID *int64 `json:"campaign_id"`
		ParentID   *int64 `json:"parent_id"`
		EntityID   *int64 `json:"entity_id"`
		JSON       string `json:"json"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	uid, ok := MustGetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if strings.TrimSpace(req.EntityType) == "" {
		req.EntityType = "oneshot"
	}
	if _, ok := aiDraftEntities[req.EntityType]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unsupported entity_type %q", req.EntityType)})
		return
	}
	if req.CampaignID != nil && !isCampaignMember(c, *req.CampaignID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	draft, err := extractDraftJSON(req.JSON)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	name := draftDisplayName(draft)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "the draft must include a non-empty name or title"})
		return
	}

	result, err := importDraftEntity(c, uid, req.EntityType, req.CampaignID, req.ParentID, req.EntityID, draft)
	if err != nil {
		writeAIGenError(c, err)
		return
	}
	result.Name = name
	status := http.StatusCreated
	if result.Updated {
		status = http.StatusOK
	}
	c.JSON(status, result)
}

// importDraftEntity creates the entity through the shared commit writers, or
// replaces it through the update writers when entity_id is supplied.
func importDraftEntity(c *gin.Context, uid int64, entityType string, campaignID, parentID, entityID *int64, raw json.RawMessage) (*draftImportResult, error) {
	ctx := c.Request.Context()
	if entityID != nil {
		return updateDraftEntity(c, uid, entityType, *entityID, raw)
	}
	if entityType == "oneshot" {
		id, counts, err := createOneShotFromDraft(ctx, uid, campaignID, raw)
		if err != nil {
			return nil, err
		}
		return &draftImportResult{ID: id, URL: entityURL("adventure", id), Entity: entityType, Counts: &counts}, nil
	}
	s := &aiDraftSession{UserID: uid, EntityType: entityType, CampaignID: campaignID, ParentID: parentID, Draft: raw}
	id, url, err := commitAIDraft(c, s)
	if err != nil {
		return nil, err
	}
	return &draftImportResult{ID: id, URL: url, Entity: entityType}, nil
}

// updateDraftEntity replaces the content of an existing entity from a draft.
// Every writer verifies that the caller owns or has access to the entity.
func updateDraftEntity(c *gin.Context, uid int64, entityType string, entityID int64, raw json.RawMessage) (*draftImportResult, error) {
	switch entityType {
	case "oneshot":
		counts, err := updateOneShotFromDraft(c, uid, entityID, raw)
		if err != nil {
			return nil, err
		}
		return &draftImportResult{ID: entityID, URL: entityURL("adventure", entityID), Entity: entityType, Updated: true, Counts: &counts}, nil
	case "npc":
		return updateNPCFromDraft(c, uid, entityID, raw)
	case "location":
		return updateLocationFromDraft(c, uid, entityID, raw)
	case "encounter":
		return updateEncounterFromDraft(c, uid, entityID, raw)
	case "faction":
		return updateFactionFromDraft(c, entityID, raw)
	case "campaign":
		return updateCampaignFromDraft(c, uid, entityID, raw)
	case "quest":
		return updateQuestFromDraft(c, entityID, raw)
	case "item":
		return updateItemFromDraft(c, uid, entityID, raw)
	}
	return nil, &aiGenError{Status: http.StatusBadRequest, Msg: fmt.Sprintf("unsupported entity_type %q", entityType)}
}

// updateOneShotFromDraft replaces an adventure's content: the main row updates,
// acts/scenes and clues are replaced, linked encounter templates are pruned
// when unreferenced, and same-named NPCs/locations are reused.
func updateOneShotFromDraft(c *gin.Context, uid, adventureID int64, raw json.RawMessage) (oneShotDraftCounts, error) {
	var counts oneShotDraftCounts
	var d aiOneShotDraft
	if err := json.Unmarshal(raw, &d); err != nil {
		return counts, fmt.Errorf("invalid draft: %w", err)
	}
	normalizeOneShotDraft(&d)
	ctx := c.Request.Context()

	var owner int64
	var campaignID sql.NullInt64
	err := db.DB.QueryRowContext(ctx, "SELECT user_id, campaign_id FROM oneshot_adventures WHERE id=?", adventureID).Scan(&owner, &campaignID)
	if err == sql.ErrNoRows {
		return counts, &aiGenError{Status: http.StatusNotFound, Msg: "one-shot adventure not found"}
	}
	if err != nil {
		return counts, err
	}
	if owner != uid {
		return counts, &aiGenError{Status: http.StatusForbidden, Msg: "access denied"}
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	if _, err := db.DB.ExecContext(ctx, `UPDATE oneshot_adventures SET title=?, premise=?, hook=?, estimated_minutes=?, difficulty=?, notes=?, updated_at=? WHERE id=?`,
		d.Title, d.Premise, d.Hook, d.EstimatedMinutes, d.Difficulty, d.Notes, now, adventureID); err != nil {
		return counts, err
	}

	oldEncounterIDs, err := linkedEncounterIDs(ctx, adventureID)
	if err != nil {
		return counts, err
	}
	for _, q := range []string{
		"DELETE FROM oneshot_adventure_encounters WHERE adventure_id=?",
		"DELETE FROM oneshot_adventure_npcs WHERE adventure_id=?",
		"DELETE FROM oneshot_adventure_locations WHERE adventure_id=?",
		"DELETE FROM clues WHERE adventure_id=?",
		"DELETE FROM oneshot_acts WHERE adventure_id=?",
	} {
		if _, err := db.DB.ExecContext(ctx, q, adventureID); err != nil {
			return counts, err
		}
	}

	var campaignPtr *int64
	if campaignID.Valid {
		campaignPtr = &campaignID.Int64
	}
	counts, err = insertOneShotChildren(ctx, uid, adventureID, campaignPtr, d)
	if err != nil {
		return counts, err
	}
	pruneOrphanEncounterTemplates(ctx, oldEncounterIDs)
	return counts, nil
}

func linkedEncounterIDs(ctx context.Context, adventureID int64) ([]int64, error) {
	rows, err := db.DB.QueryContext(ctx, "SELECT encounter_id FROM oneshot_adventure_encounters WHERE adventure_id=?", adventureID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// pruneOrphanEncounterTemplates removes encounter templates that were linked to
// a replaced adventure but are no longer referenced by any adventure or scene.
// Templates shared with other adventures are left alone.
func pruneOrphanEncounterTemplates(ctx context.Context, ids []int64) {
	for _, id := range ids {
		var refs int
		if err := db.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM oneshot_adventure_encounters WHERE encounter_id=?", id).Scan(&refs); err != nil || refs > 0 {
			continue
		}
		if err := db.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM oneshot_scenes WHERE encounter_id=?", id).Scan(&refs); err != nil || refs > 0 {
			continue
		}
		db.DB.ExecContext(ctx, "DELETE FROM encounter_templates WHERE id=?", id)
	}
}

func updateNPCFromDraft(c *gin.Context, uid, entityID int64, raw json.RawMessage) (*draftImportResult, error) {
	var d struct {
		Name        string `json:"name"`
		Race        string `json:"race"`
		Class       string `json:"class"`
		Description string `json:"description"`
		Notes       string `json:"notes"`
	}
	if err := json.Unmarshal(raw, &d); err != nil {
		return nil, fmt.Errorf("invalid draft: %w", err)
	}
	if strings.TrimSpace(d.Name) == "" {
		return nil, &aiGenError{Status: http.StatusBadRequest, Msg: "name is required"}
	}
	ok, err := db.Client.NPC.Query().Where(npc.ID(entityID), npc.UserID(uid)).Exist(c.Request.Context())
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, &aiGenError{Status: http.StatusForbidden, Msg: "access denied"}
	}
	if _, err := db.Client.NPC.UpdateOneID(entityID).SetName(d.Name).SetRace(d.Race).SetClass(d.Class).
		SetDescription(d.Description).SetNotes(d.Notes).Save(c.Request.Context()); err != nil {
		return nil, err
	}
	return &draftImportResult{ID: entityID, URL: entityURL("npc", entityID), Entity: "npc", Updated: true}, nil
}

func updateLocationFromDraft(c *gin.Context, uid, entityID int64, raw json.RawMessage) (*draftImportResult, error) {
	var d struct {
		Name        string `json:"name"`
		Type        string `json:"type"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal(raw, &d); err != nil {
		return nil, fmt.Errorf("invalid draft: %w", err)
	}
	if strings.TrimSpace(d.Name) == "" {
		return nil, &aiGenError{Status: http.StatusBadRequest, Msg: "name is required"}
	}
	if strings.TrimSpace(d.Type) == "" {
		d.Type = "region"
	}
	ok, err := db.Client.Location.Query().Where(location.ID(entityID), location.UserID(uid)).Exist(c.Request.Context())
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, &aiGenError{Status: http.StatusForbidden, Msg: "access denied"}
	}
	if _, err := db.Client.Location.UpdateOneID(entityID).SetName(d.Name).SetType(d.Type).
		SetDescription(d.Description).Save(c.Request.Context()); err != nil {
		return nil, err
	}
	return &draftImportResult{ID: entityID, URL: entityURL("location", entityID), Entity: "location", Updated: true}, nil
}

func updateEncounterFromDraft(c *gin.Context, uid, entityID int64, raw json.RawMessage) (*draftImportResult, error) {
	var d struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Environment string `json:"environment"`
		Difficulty  string `json:"difficulty"`
		Notes       string `json:"notes"`
	}
	if err := json.Unmarshal(raw, &d); err != nil {
		return nil, fmt.Errorf("invalid draft: %w", err)
	}
	if strings.TrimSpace(d.Name) == "" {
		return nil, &aiGenError{Status: http.StatusBadRequest, Msg: "name is required"}
	}
	if !aiDifficulties[d.Difficulty] {
		d.Difficulty = "medium"
	}
	var owner int64
	err := db.DB.QueryRowContext(c.Request.Context(), "SELECT user_id FROM encounter_templates WHERE id=?", entityID).Scan(&owner)
	if err == sql.ErrNoRows {
		return nil, &aiGenError{Status: http.StatusNotFound, Msg: "encounter not found"}
	}
	if err != nil {
		return nil, err
	}
	if owner != uid {
		return nil, &aiGenError{Status: http.StatusForbidden, Msg: "access denied"}
	}
	if _, err := db.DB.ExecContext(c.Request.Context(), "UPDATE encounter_templates SET name=?, description=?, environment=?, difficulty=?, notes=? WHERE id=?",
		d.Name, d.Description, d.Environment, d.Difficulty, d.Notes, entityID); err != nil {
		return nil, err
	}
	return &draftImportResult{ID: entityID, URL: entityURL("encounter", entityID), Entity: "encounter", Updated: true}, nil
}

func updateFactionFromDraft(c *gin.Context, entityID int64, raw json.RawMessage) (*draftImportResult, error) {
	var d struct {
		Name         string `json:"name"`
		Description  string `json:"description"`
		Type         string `json:"type"`
		Headquarters string `json:"headquarters"`
	}
	if err := json.Unmarshal(raw, &d); err != nil {
		return nil, fmt.Errorf("invalid draft: %w", err)
	}
	if strings.TrimSpace(d.Name) == "" {
		return nil, &aiGenError{Status: http.StatusBadRequest, Msg: "name is required"}
	}
	if strings.TrimSpace(d.Type) == "" {
		d.Type = "organization"
	}
	var campaignID sql.NullInt64
	err := db.DB.QueryRowContext(c.Request.Context(), "SELECT campaign_id FROM factions WHERE id=?", entityID).Scan(&campaignID)
	if err == sql.ErrNoRows {
		return nil, &aiGenError{Status: http.StatusNotFound, Msg: "faction not found"}
	}
	if err != nil {
		return nil, err
	}
	if !campaignID.Valid || !isCampaignMember(c, campaignID.Int64) {
		return nil, &aiGenError{Status: http.StatusForbidden, Msg: "access denied"}
	}
	if _, err := db.DB.ExecContext(c.Request.Context(), "UPDATE factions SET name=?, description=?, type=?, headquarters=? WHERE id=?",
		d.Name, d.Description, d.Type, d.Headquarters, entityID); err != nil {
		return nil, err
	}
	return &draftImportResult{ID: entityID, URL: entityURL("faction", entityID), Entity: "faction", Updated: true}, nil
}

func updateCampaignFromDraft(c *gin.Context, uid, entityID int64, raw json.RawMessage) (*draftImportResult, error) {
	var d struct {
		Name        string `json:"name"`
		PartyName   string `json:"party_name"`
		Description string `json:"description"`
		DMNotes     string `json:"dm_notes"`
	}
	if err := json.Unmarshal(raw, &d); err != nil {
		return nil, fmt.Errorf("invalid draft: %w", err)
	}
	if strings.TrimSpace(d.Name) == "" {
		return nil, &aiGenError{Status: http.StatusBadRequest, Msg: "name is required"}
	}
	ok, err := db.Client.Campaign.Query().Where(campaign.ID(entityID), campaign.UserID(uid)).Exist(c.Request.Context())
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, &aiGenError{Status: http.StatusForbidden, Msg: "access denied"}
	}
	if _, err := db.Client.Campaign.UpdateOneID(entityID).SetName(d.Name).SetPartyName(d.PartyName).
		SetDescription(d.Description).SetDmNotes(d.DMNotes).Save(c.Request.Context()); err != nil {
		return nil, err
	}
	return &draftImportResult{ID: entityID, URL: entityURL("campaign", entityID), Entity: "campaign", Updated: true}, nil
}

func updateQuestFromDraft(c *gin.Context, entityID int64, raw json.RawMessage) (*draftImportResult, error) {
	var d struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Objectives  string `json:"objectives"`
		Rewards     string `json:"rewards"`
		Notes       string `json:"notes"`
	}
	if err := json.Unmarshal(raw, &d); err != nil {
		return nil, fmt.Errorf("invalid draft: %w", err)
	}
	if strings.TrimSpace(d.Name) == "" {
		return nil, &aiGenError{Status: http.StatusBadRequest, Msg: "name is required"}
	}
	var characterID int64
	err := db.DB.QueryRowContext(c.Request.Context(), "SELECT character_id FROM quests WHERE id=?", entityID).Scan(&characterID)
	if err == sql.ErrNoRows {
		return nil, &aiGenError{Status: http.StatusNotFound, Msg: "quest not found"}
	}
	if err != nil {
		return nil, err
	}
	if !canEditCharacterID(c, characterID) {
		return nil, &aiGenError{Status: http.StatusForbidden, Msg: "access denied"}
	}
	if _, err := db.Client.Quest.UpdateOneID(entityID).SetName(d.Name).SetDescription(d.Description).
		SetObjectives(d.Objectives).SetRewards(d.Rewards).SetNotes(d.Notes).Save(c.Request.Context()); err != nil {
		return nil, err
	}
	return &draftImportResult{ID: entityID, URL: entityURL("quest", entityID), Entity: "quest", Updated: true}, nil
}

func updateItemFromDraft(c *gin.Context, uid, entityID int64, raw json.RawMessage) (*draftImportResult, error) {
	var d struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Category    string  `json:"category"`
		Quantity    int     `json:"quantity"`
		Weight      float64 `json:"weight"`
		PriceGP     float64 `json:"price_gp"`
		IsMagical   bool    `json:"is_magical"`
		Attunement  bool    `json:"attunement"`
		Notes       string  `json:"notes"`
	}
	if err := json.Unmarshal(raw, &d); err != nil {
		return nil, fmt.Errorf("invalid draft: %w", err)
	}
	if strings.TrimSpace(d.Name) == "" {
		return nil, &aiGenError{Status: http.StatusBadRequest, Msg: "name is required"}
	}
	if strings.TrimSpace(d.Category) == "" {
		d.Category = "gear"
	}
	if d.Quantity <= 0 {
		d.Quantity = 1
	}
	ctx := c.Request.Context()
	var adventureID int64
	err := db.DB.QueryRowContext(ctx, "SELECT adventure_id FROM oneshot_items WHERE id=?", entityID).Scan(&adventureID)
	if err == sql.ErrNoRows {
		return nil, &aiGenError{Status: http.StatusNotFound, Msg: "item not found"}
	}
	if err != nil {
		return nil, err
	}
	var owner int64
	if err := db.DB.QueryRowContext(ctx, "SELECT user_id FROM oneshot_adventures WHERE id=?", adventureID).Scan(&owner); err != nil {
		return nil, err
	}
	if owner != uid {
		return nil, &aiGenError{Status: http.StatusForbidden, Msg: "access denied"}
	}
	if _, err := db.DB.ExecContext(ctx, `UPDATE oneshot_items SET name=?, description=?, category=?, quantity=?, weight=?, price_gp=?, is_magical=?, attunement=?, notes=? WHERE id=?`,
		d.Name, d.Description, d.Category, d.Quantity, d.Weight, d.PriceGP, d.IsMagical, d.Attunement, d.Notes, entityID); err != nil {
		return nil, err
	}
	return &draftImportResult{ID: entityID, URL: entityURL("item", entityID), Entity: "item", Updated: true}, nil
}

// draftDisplayName reads the user-facing name from a draft: "title" for
// one-shots, "name" for every other entity type.
func draftDisplayName(raw json.RawMessage) string {
	var probe struct {
		Title string `json:"title"`
		Name  string `json:"name"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return ""
	}
	if strings.TrimSpace(probe.Title) != "" {
		return strings.TrimSpace(probe.Title)
	}
	return strings.TrimSpace(probe.Name)
}

// extractDraftJSON normalizes pasted text into a draft JSON object. It strips
// markdown fences and surrounding prose, then unwraps the assistant envelope
// when the object carries a non-null "draft" field.
func extractDraftJSON(text string) (json.RawMessage, error) {
	cleaned := strings.TrimSpace(text)
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)
	if i := strings.Index(cleaned, "{"); i >= 0 {
		if j := strings.LastIndex(cleaned, "}"); j > i {
			cleaned = cleaned[i : j+1]
		}
	}
	if cleaned == "" || !json.Valid([]byte(cleaned)) {
		return nil, fmt.Errorf("the pasted text is not a valid JSON object")
	}
	var envelope struct {
		Status string          `json:"status"`
		Draft  json.RawMessage `json:"draft"`
	}
	if err := json.Unmarshal([]byte(cleaned), &envelope); err == nil {
		if trimmed := strings.TrimSpace(string(envelope.Draft)); trimmed != "" && trimmed != "null" {
			// A draft may arrive as a JSON string rather than an object.
			if strings.HasPrefix(trimmed, `"`) {
				var s string
				if err := json.Unmarshal(envelope.Draft, &s); err == nil {
					return json.RawMessage(s), nil
				}
			}
			return envelope.Draft, nil
		}
		// An assistant envelope whose draft is null/absent means the model is
		// still asking questions; say so instead of failing on the title check.
		if envelope.Status != "" {
			return nil, fmt.Errorf("the pasted JSON has no draft yet — the AI is still asking questions or has not generated one; answer in the chat or ask it to generate the full draft")
		}
	}
	return json.RawMessage(cleaned), nil
}
