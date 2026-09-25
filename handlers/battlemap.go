package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"villum/db"
)

func battlemapIDParam(c *gin.Context, name string) int64 {
	id, _ := strconv.ParseInt(c.Param(name), 10, 64)
	return id
}

// battlemapToken is a draggable marker on the campaign's active map. When it is
// linked to a combat entry it mirrors that entry's live HP/AC so the map
// "feeds" from the encounter.
type battlemapToken struct {
	ID            int64   `json:"id"`
	CampaignID    int64   `json:"campaign_id"`
	CombatEntryID *int64  `json:"combat_entry_id,omitempty"`
	Name          string  `json:"name"`
	X             float64 `json:"x"`
	Y             float64 `json:"y"`
	Color         string  `json:"color"`
	Size          float64 `json:"size"`
	HPCurrent     *int    `json:"hp_current,omitempty"`
	HPMax         *int    `json:"hp_max,omitempty"`
	AC            *int    `json:"ac,omitempty"`
	Conditions    string  `json:"conditions,omitempty"`
}

type battlemapTokenRequest struct {
	CombatEntryID *int64   `json:"combat_entry_id"`
	Name          string   `json:"name"`
	X             *float64 `json:"x"`
	Y             *float64 `json:"y"`
	Color         string   `json:"color"`
	Size          *float64 `json:"size"`
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func scanBattlemapToken(scan func(dest ...any) error) (battlemapToken, error) {
	var t battlemapToken
	var entryID sql.NullInt64
	var hpCur, hpMax, ac sql.NullInt64
	var conds sql.NullString
	err := scan(&t.ID, &t.CampaignID, &entryID, &t.Name, &t.X, &t.Y, &t.Color, &t.Size, &hpCur, &hpMax, &ac, &conds)
	if entryID.Valid {
		t.CombatEntryID = &entryID.Int64
	}
	if hpCur.Valid {
		v := int(hpCur.Int64)
		t.HPCurrent = &v
	}
	if hpMax.Valid {
		v := int(hpMax.Int64)
		t.HPMax = &v
	}
	if ac.Valid {
		v := int(ac.Int64)
		t.AC = &v
	}
	t.Conditions = conds.String
	return t, err
}

const battlemapTokenSelect = `
SELECT t.id, t.campaign_id, t.combat_entry_id, t.name, t.x, t.y, t.color, t.size,
       ce.hp_current, ce.hp_max, ce.ac, ce.condition_ids
FROM battlemap_tokens t
LEFT JOIN combat_entries ce ON ce.id = t.combat_entry_id
WHERE t.campaign_id = ? ORDER BY t.id`

func loadBattlemapTokens(campaignID int64) []battlemapToken {
	rows, err := db.DB.Query(battlemapTokenSelect, campaignID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	tokens := []battlemapToken{}
	for rows.Next() {
		t, err := scanBattlemapToken(rows.Scan)
		if err != nil {
			continue
		}
		tokens = append(tokens, t)
	}
	return tokens
}

// campaignStateMap returns the campaign's active map metadata for the battlemap
// surface, or nil when the campaign has no map yet.
func campaignBattlemapMeta(campaignID int64) gin.H {
	var id int64
	var name, imageURL, gridUnits string
	var width, height, gridSize int
	err := db.DB.QueryRow(`
		SELECT id, name, image_url, width, height, grid_size, COALESCE(grid_units,'ft')
		FROM campaign_maps WHERE campaign_id = ?
		ORDER BY is_active DESC, id LIMIT 1`, campaignID).
		Scan(&id, &name, &imageURL, &width, &height, &gridSize, &gridUnits)
	if err != nil {
		return nil
	}
	return gin.H{"id": id, "name": name, "image_url": imageURL, "width": width,
		"height": height, "grid_size": gridSize, "grid_units": gridUnits}
}

// GetCampaignBattlemap returns the active map plus every token for the campaign.
func GetCampaignBattlemap(c *gin.Context) {
	campaignID := battlemapIDParam(c, "id")
	if campaignID <= 0 {
		WriteNotFound(c, "campaign not found")
		return
	}
	if !isCampaignMember(c, campaignID) {
		WriteError(c, http.StatusForbidden, errAccessDenied)
		return
	}
	WriteJSON(c, http.StatusOK, gin.H{
		"map":    campaignBattlemapMeta(campaignID),
		"tokens": loadBattlemapTokens(campaignID),
	})
}

// CreateBattlemapToken adds a token to the campaign battlemap.
func CreateBattlemapToken(c *gin.Context) {
	campaignID := battlemapIDParam(c, "id")
	if campaignID <= 0 {
		WriteNotFound(c, "campaign not found")
		return
	}
	if !isCampaignDM(c, campaignID) {
		WriteError(c, http.StatusForbidden, errAccessDenied)
		return
	}
	var req battlemapTokenRequest
	if !BindOr400(c, &req) {
		return
	}
	name := req.Name
	if name == "" && req.CombatEntryID != nil {
		db.DB.QueryRow("SELECT name FROM combat_entries WHERE id=? AND campaign_id=?", *req.CombatEntryID, campaignID).Scan(&name)
	}
	if name == "" {
		name = "Token"
	}
	x, y, size := 0.5, 0.5, 1.0
	if req.X != nil {
		x = clamp01(*req.X)
	}
	if req.Y != nil {
		y = clamp01(*req.Y)
	}
	if req.Size != nil && *req.Size > 0 {
		size = *req.Size
	}
	color := req.Color
	if color == "" {
		color = "#b8963e"
	}
	res, err := db.DB.Exec(`
		INSERT INTO battlemap_tokens (campaign_id, combat_entry_id, name, x, y, color, size)
		VALUES (?, ?, ?, ?, ?, ?, ?)`, campaignID, req.CombatEntryID, name, x, y, color, size)
	if err != nil {
		WriteError(c, http.StatusInternalServerError, strErr("could not create token"))
		return
	}
	id, _ := res.LastInsertId()
	var created *battlemapToken
	for _, t := range loadBattlemapTokens(campaignID) {
		if t.ID == id {
			tt := t
			created = &tt
			break
		}
	}
	SendBattlemapUpdate(campaignID)
	WriteJSON(c, http.StatusCreated, gin.H{"token": created})
}

// UpdateBattlemapToken moves/edits an existing token.
func UpdateBattlemapToken(c *gin.Context) {
	id := battlemapIDParam(c, "id")
	var campaignID int64
	if err := db.DB.QueryRow("SELECT campaign_id FROM battlemap_tokens WHERE id=?", id).Scan(&campaignID); err != nil {
		WriteNotFound(c, "token not found")
		return
	}
	if !isCampaignMember(c, campaignID) {
		WriteError(c, http.StatusForbidden, errAccessDenied)
		return
	}
	var req battlemapTokenRequest
	if !BindOr400(c, &req) {
		return
	}
	if req.X != nil {
		if _, err := db.DB.Exec("UPDATE battlemap_tokens SET x=?, updated_at=datetime('now') WHERE id=?", clamp01(*req.X), id); err != nil {
			WriteError(c, http.StatusInternalServerError, strErr("could not move token"))
			return
		}
	}
	if req.Y != nil {
		db.DB.Exec("UPDATE battlemap_tokens SET y=?, updated_at=datetime('now') WHERE id=?", clamp01(*req.Y), id)
	}
	if req.Name != "" {
		db.DB.Exec("UPDATE battlemap_tokens SET name=?, updated_at=datetime('now') WHERE id=?", req.Name, id)
	}
	if req.Color != "" {
		db.DB.Exec("UPDATE battlemap_tokens SET color=?, updated_at=datetime('now') WHERE id=?", req.Color, id)
	}
	if req.Size != nil && *req.Size > 0 {
		db.DB.Exec("UPDATE battlemap_tokens SET size=?, updated_at=datetime('now') WHERE id=?", *req.Size, id)
	}
	SendBattlemapUpdate(campaignID)
	var updated *battlemapToken
	for _, t := range loadBattlemapTokens(campaignID) {
		if t.ID == id {
			tt := t
			updated = &tt
			break
		}
	}
	WriteJSON(c, http.StatusOK, gin.H{"token": updated})
}

// DeleteBattlemapToken removes a token (and only the token, never the combat entry).
func DeleteBattlemapToken(c *gin.Context) {
	id := battlemapIDParam(c, "id")
	var campaignID int64
	if err := db.DB.QueryRow("SELECT campaign_id FROM battlemap_tokens WHERE id=?", id).Scan(&campaignID); err != nil {
		WriteNotFound(c, "token not found")
		return
	}
	if !isCampaignDM(c, campaignID) {
		WriteError(c, http.StatusForbidden, errAccessDenied)
		return
	}
	db.DB.Exec("DELETE FROM battlemap_tokens WHERE id=?", id)
	SendBattlemapUpdate(campaignID)
	WriteJSON(c, http.StatusOK, gin.H{"ok": true})
}

// SyncBattlemapTokens creates a token for every active combat entry that does
// not have one yet, laid out in a loose grid so they do not stack.
func SyncBattlemapTokens(c *gin.Context) {
	campaignID := battlemapIDParam(c, "id")
	if campaignID <= 0 {
		WriteNotFound(c, "campaign not found")
		return
	}
	if !isCampaignDM(c, campaignID) {
		WriteError(c, http.StatusForbidden, errAccessDenied)
		return
	}
	rows, err := db.DB.Query(`
		SELECT ce.id, ce.name FROM combat_entries ce
		WHERE ce.campaign_id=? AND ce.is_active=1
		  AND NOT EXISTS (SELECT 1 FROM battlemap_tokens t WHERE t.campaign_id=ce.campaign_id AND t.combat_entry_id=ce.id)
		ORDER BY ce.id`, campaignID)
	if err != nil {
		WriteError(c, http.StatusInternalServerError, strErr("could not load combatants"))
		return
	}
	type entry struct {
		id   int64
		name string
	}
	var entries []entry
	for rows.Next() {
		var e entry
		if rows.Scan(&e.id, &e.name) == nil {
			entries = append(entries, e)
		}
	}
	rows.Close()

	created := 0
	for i, e := range entries {
		x := 0.15 + float64(i%5)*0.17
		y := 0.2 + float64(i/5)*0.17
		if _, err := db.DB.Exec(`
			INSERT INTO battlemap_tokens (campaign_id, combat_entry_id, name, x, y)
			VALUES (?, ?, ?, ?, ?)`, campaignID, e.id, e.name, clamp01(x), clamp01(y)); err == nil {
			created++
		}
	}
	if created > 0 {
		SendBattlemapUpdate(campaignID)
	}
	WriteJSON(c, http.StatusOK, gin.H{"ok": true, "created": created, "tokens": loadBattlemapTokens(campaignID)})
}
