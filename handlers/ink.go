package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/middleware"
)

// inkCapture is one handwriting capture: pages plus the transcription state and
// the entity it was promoted into.
type inkCapture struct {
	ID             int64     `json:"id"`
	CampaignID     int64     `json:"campaign_id"`
	CharacterID    *int64    `json:"character_id,omitempty"`
	OwnerUserID    int64     `json:"owner_user_id"`
	SessionKey     string    `json:"session_key"`
	Title          string    `json:"title"`
	Source         string    `json:"source"`
	Status         string    `json:"status"`
	OCRSource      string    `json:"ocr_source"`
	JournalEntryID *int64    `json:"journal_entry_id,omitempty"`
	RecapID        *int64    `json:"recap_id,omitempty"`
	Error          string    `json:"error,omitempty"`
	CreatedAt      string    `json:"created_at"`
	UpdatedAt      string    `json:"updated_at"`
	Pages          []inkPage `json:"pages"`
}

type inkPage struct {
	ID           int64  `json:"id"`
	PageIndex    int    `json:"page_index"`
	UploadID     int64  `json:"upload_id,omitempty"`
	URL          string `json:"url,omitempty"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
	Text         string `json:"text"`
	OCRSource    string `json:"ocr_source"`
	SHA256       string `json:"sha256,omitempty"`
}

const maxInkPages = 10

// loadInkCapture loads a capture and its pages. A missing capture returns an
// error that callers map to 404.
func loadInkCapture(c *gin.Context, id int64) (*inkCapture, error) {
	var cap inkCapture
	err := db.DB.QueryRow(`SELECT id, campaign_id, character_id, owner_user_id, COALESCE(session_key,''),
			COALESCE(title,''), COALESCE(source,'share'), COALESCE(status,'uploaded'), COALESCE(ocr_source,'none'),
			journal_entry_id, recap_id, COALESCE(error,''), COALESCE(created_at,''), COALESCE(updated_at,'')
		FROM ink_captures WHERE id=?`, id).
		Scan(&cap.ID, &cap.CampaignID, &cap.CharacterID, &cap.OwnerUserID, &cap.SessionKey, &cap.Title,
			&cap.Source, &cap.Status, &cap.OCRSource, &cap.JournalEntryID, &cap.RecapID, &cap.Error,
			&cap.CreatedAt, &cap.UpdatedAt)
	if err != nil {
		return nil, err
	}
	rows, err := db.DB.Query(`SELECT p.id, p.page_index, COALESCE(p.upload_id,0), COALESCE(u.url,''),
			COALESCE(u.thumbnail_url,''), COALESCE(p.ocr_text,''), COALESCE(p.ocr_source,'none'), COALESCE(p.sha256,'')
		FROM ink_pages p LEFT JOIN uploads u ON u.id=p.upload_id
		WHERE p.capture_id=? ORDER BY p.page_index, p.id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cap.Pages = []inkPage{}
	for rows.Next() {
		var p inkPage
		if err := rows.Scan(&p.ID, &p.PageIndex, &p.UploadID, &p.URL, &p.ThumbnailURL, &p.Text, &p.OCRSource, &p.SHA256); err != nil {
			return nil, err
		}
		cap.Pages = append(cap.Pages, p)
	}
	return &cap, rows.Err()
}

// canSeeInkCapture reports whether the caller may read or mutate a capture:
// its owner, the campaign DM, or an admin. Other campaign members are denied.
func canSeeInkCapture(c *gin.Context, cap *inkCapture) bool {
	role, _ := c.Get("role")
	if role == "admin" {
		return true
	}
	if uid, ok := MustGetUserID(c); ok && uid == cap.OwnerUserID {
		return true
	}
	return isCampaignDM(c, cap.CampaignID)
}

func captureText(cap *inkCapture) string {
	parts := make([]string, 0, len(cap.Pages))
	for _, p := range cap.Pages {
		if t := strings.TrimSpace(p.Text); t != "" {
			parts = append(parts, t)
		}
	}
	return strings.Join(parts, "\n\n")
}

// CreateInkCapture accepts a multipart upload of one or more page images with
// optional text sidecars exported by the device. DM captures target a campaign;
// player captures target a character the caller may edit.
func CreateInkCapture(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid multipart form"})
		return
	}
	form := c.Request.MultipartForm
	if form == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid multipart form"})
		return
	}
	files := append([]*multipart.FileHeader{}, form.File["pages"]...)
	files = append(files, form.File["pages[]"]...)
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one page image is required"})
		return
	}
	if len(files) > maxInkPages {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("a capture can hold at most %d pages", maxInkPages)})
		return
	}

	campaignID, _ := strconv.ParseInt(c.PostForm("campaign_id"), 10, 64)
	if campaignID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "campaign_id is required"})
		return
	}
	characterID, _ := strconv.ParseInt(c.PostForm("character_id"), 10, 64)
	if characterID > 0 {
		if !canEditCharacterID(c, characterID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		if !isCampaignMember(c, campaignID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
	} else if !isCampaignDM(c, campaignID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	source := strings.TrimSpace(c.PostForm("source"))
	if source == "" {
		source = "share"
	}
	if source != "share" && source != "canvas" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "source must be 'share' or 'canvas'"})
		return
	}
	uid, ok := MustGetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	result, err := db.DB.Exec(`INSERT INTO ink_captures(campaign_id, character_id, owner_user_id, session_key, title, source, status, ocr_source)
		VALUES(?,?,?,?,?,?, 'uploaded', 'none')`,
		campaignID, inkNullableID(characterID), uid, strings.TrimSpace(c.PostForm("session_key")),
		strings.TrimSpace(c.PostForm("title")), source)
	if err != nil {
		InternalError(c, err)
		return
	}
	captureID, _ := result.LastInsertId()

	rollback := func() {
		db.DB.Exec("DELETE FROM ink_pages WHERE capture_id=?", captureID)
		db.DB.Exec("DELETE FROM ink_captures WHERE id=?", captureID)
	}

	texts := append([]string{}, form.Value["text"]...)
	texts = append(texts, form.Value["text[]"]...)
	seen := map[string]bool{}
	duplicates := 0
	stored := 0
	anyText := false
	for i, fh := range files {
		data, err := readMultipartFile(fh)
		if err != nil {
			rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read page image"})
			return
		}
		sum := sha256.Sum256(data)
		hash := hex.EncodeToString(sum[:])
		if seen[hash] {
			duplicates++
			continue
		}
		seen[hash] = true
		up, _, err := SaveMediaBytes(fh.Filename, data, "ink", captureID, "ink", "")
		if err != nil {
			rollback()
			var me *mediaError
			if errors.As(err, &me) {
				c.JSON(me.status, gin.H{"error": me.msg})
				return
			}
			InternalError(c, err)
			return
		}
		if up.Kind != "image" {
			rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "ink pages must be images"})
			return
		}
		text := ""
		if i < len(texts) {
			text = strings.TrimSpace(texts[i])
		}
		ocrSource := "none"
		if text != "" {
			ocrSource = "boox"
			anyText = true
		}
		if _, err := db.DB.Exec(`INSERT INTO ink_pages(capture_id, page_index, upload_id, ocr_text, ocr_source, sha256)
			VALUES(?,?,?,?,?,?)`, captureID, stored, up.ID, text, ocrSource, hash); err != nil {
			rollback()
			InternalError(c, err)
			return
		}
		stored++
	}
	if stored == 0 {
		rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "no page images were stored"})
		return
	}
	if anyText {
		db.DB.Exec(`UPDATE ink_captures SET status='transcribed', ocr_source='boox', updated_at=? WHERE id=?`,
			time.Now().UTC().Format(time.RFC3339), captureID)
	}

	cap, err := loadInkCapture(c, captureID)
	if err != nil {
		InternalError(c, err)
		return
	}
	resp := gin.H{"capture": cap, "duplicates": duplicates}
	c.JSON(http.StatusCreated, resp)
}

func inkNullableID(v int64) any {
	if v == 0 {
		return nil
	}
	return v
}

func readMultipartFile(fh *multipart.FileHeader) ([]byte, error) {
	f, err := fh.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(io.LimitReader(f, 10*1024*1024+1))
}

// ListInkCaptures returns captures for a campaign or a character. Players only
// see their own captures; DMs and admins see the whole campaign.
func ListInkCaptures(c *gin.Context) {
	campaignID, _ := strconv.ParseInt(c.Query("campaign_id"), 10, 64)
	characterID, _ := strconv.ParseInt(c.Query("character_id"), 10, 64)
	if campaignID == 0 && characterID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "campaign_id or character_id is required"})
		return
	}
	if characterID > 0 && !canEditCharacterID(c, characterID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	if campaignID > 0 && !isCampaignMember(c, campaignID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	role, _ := c.Get("role")
	uid, _ := MustGetUserID(c)
	where := []string{}
	args := []any{}
	if characterID > 0 {
		where = append(where, "character_id=?")
		args = append(args, characterID)
	}
	if campaignID > 0 {
		where = append(where, "campaign_id=?")
		args = append(args, campaignID)
	}
	// Privacy: an owner sees their own ink, a DM/admin sees the campaign's.
	if role != "admin" && !(campaignID > 0 && isCampaignDM(c, campaignID)) {
		where = append(where, "owner_user_id=?")
		args = append(args, uid)
	}
	query := `SELECT id FROM ink_captures WHERE ` + strings.Join(where, " AND ") + ` ORDER BY created_at DESC, id DESC LIMIT 100`
	rows, err := db.DB.Query(query, args...)
	if err != nil {
		InternalError(c, err)
		return
	}
	ids := []int64{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err == nil {
			ids = append(ids, id)
		}
	}
	rows.Close()

	out := make([]*inkCapture, 0, len(ids))
	for _, id := range ids {
		cap, err := loadInkCapture(c, id)
		if err != nil {
			continue
		}
		out = append(out, cap)
	}
	c.JSON(http.StatusOK, out)
}

func GetInkCapture(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	cap, err := loadInkCapture(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "capture not found"})
		return
	}
	if !canSeeInkCapture(c, cap) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	c.JSON(http.StatusOK, cap)
}

// UpdateInkCapture edits the title/session and accepts manual page text, which
// always moves the capture to `reviewed`.
func UpdateInkCapture(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	cap, err := loadInkCapture(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "capture not found"})
		return
	}
	if !canSeeInkCapture(c, cap) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var req struct {
		Title      *string `json:"title"`
		SessionKey *string `json:"session_key"`
		Pages      []struct {
			ID   int64  `json:"id"`
			Text string `json:"text"`
		} `json:"pages"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	now := time.Now().UTC().Format(time.RFC3339)
	if req.Title != nil {
		db.DB.Exec("UPDATE ink_captures SET title=?, updated_at=? WHERE id=?", strings.TrimSpace(*req.Title), now, id)
	}
	if req.SessionKey != nil {
		db.DB.Exec("UPDATE ink_captures SET session_key=?, updated_at=? WHERE id=?", strings.TrimSpace(*req.SessionKey), now, id)
	}
	edited := false
	for _, p := range req.Pages {
		res, err := db.DB.Exec(`UPDATE ink_pages SET ocr_text=?, ocr_source='manual'
			WHERE id=? AND capture_id=?`, strings.TrimSpace(p.Text), p.ID, id)
		if err == nil {
			if n, _ := res.RowsAffected(); n > 0 {
				edited = true
			}
		}
	}
	if edited {
		db.DB.Exec(`UPDATE ink_captures SET status='reviewed', ocr_source='manual', error='', updated_at=? WHERE id=?`, now, id)
	}
	cap, err = loadInkCapture(c, id)
	if err != nil {
		InternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, cap)
}

func DeleteInkCapture(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	cap, err := loadInkCapture(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "capture not found"})
		return
	}
	if !canSeeInkCapture(c, cap) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	db.DB.Exec("DELETE FROM ink_pages WHERE capture_id=?", id)
	db.DB.Exec("DELETE FROM ink_captures WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// HandleInkOCR transcribes pending pages. Sidecar text is never overwritten.
// Nothing is written unless every requested page succeeds, so a provider
// failure cannot leave half a capture transcribed.
func HandleInkOCR(c *gin.Context) {
	if !aiEnabled(c.Request.Context()) {
		c.JSON(http.StatusConflict, gin.H{"error": "AI features are disabled", "manual": true})
		return
	}
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	cap, err := loadInkCapture(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "capture not found"})
		return
	}
	if !canSeeInkCapture(c, cap) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var req struct {
		EndpointID int64 `json:"endpoint_id"`
	}
	_ = c.ShouldBindJSON(&req)

	pending := []inkPage{}
	for _, p := range cap.Pages {
		if strings.TrimSpace(p.Text) == "" {
			pending = append(pending, p)
		}
	}
	if len(pending) == 0 {
		c.JSON(http.StatusOK, cap)
		return
	}
	if _, err := enabledVisionEndpoint(c.Request.Context(), req.EndpointID); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "no vision endpoint is configured; type the text manually", "manual": true})
		return
	}

	transcribed := make(map[int64]string, len(pending))
	for _, p := range pending {
		dataURL, err := pageDataURL(p.URL)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "page image is unavailable", "manual": true})
			return
		}
		text, err := generateVision(c.Request.Context(), req.EndpointID, dataURL, "Transcribe this handwritten page for a D&D session log.", nil, "")
		if err != nil {
			status := http.StatusBadGateway
			var ge *aiGenError
			if errors.As(err, &ge) {
				status = ge.Status
			}
			c.JSON(status, gin.H{"error": err.Error(), "manual": true})
			return
		}
		transcribed[p.ID] = text
	}

	now := time.Now().UTC().Format(time.RFC3339)
	for pageID, text := range transcribed {
		if _, err := db.DB.Exec("UPDATE ink_pages SET ocr_text=?, ocr_source='vision' WHERE id=? AND capture_id=?", text, pageID, id); err != nil {
			InternalError(c, err)
			return
		}
	}
	db.DB.Exec(`UPDATE ink_captures SET status='transcribed', ocr_source='vision', error='', updated_at=? WHERE id=?`, now, id)
	middleware.LogInfo("ink", "capture transcribed", "capture_id", id, "pages", len(transcribed))

	cap, _ = loadInkCapture(c, id)
	c.JSON(http.StatusOK, cap)
}

// PromoteInkCapture turns a reviewed capture into canonical content: an
// unpublished recap (DM) or a journal entry (player). Nothing is published.
func PromoteInkCapture(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	cap, err := loadInkCapture(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "capture not found"})
		return
	}
	if !canSeeInkCapture(c, cap) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	if cap.RecapID != nil || cap.JournalEntryID != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "capture was already promoted"})
		return
	}
	var req struct {
		Target           string  `json:"target"`
		Title            string  `json:"title"`
		SessionStartDate *string `json:"session_start_date"`
		SessionEndDate   *string `json:"session_end_date"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	text := captureText(cap)
	if text == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "transcribe the capture before promoting it"})
		return
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = strings.TrimSpace(cap.Title)
	}
	if title == "" {
		title = "Handwritten session notes"
	}

	switch req.Target {
	case "recap":
		if !isCampaignDM(c, cap.CampaignID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "only the campaign DM can promote into a recap"})
			return
		}
		wordCount := len(strings.Fields(text))
		res, err := db.DB.Exec(`INSERT INTO campaign_recaps(campaign_id, title, content, session_start_date, session_end_date, word_count)
			VALUES(?,?,?,?,?,?)`, cap.CampaignID, title, text, req.SessionStartDate, req.SessionEndDate, wordCount)
		if err != nil {
			InternalError(c, err)
			return
		}
		recapID, _ := res.LastInsertId()
		db.DB.Exec(`UPDATE ink_captures SET recap_id=?, status='promoted', updated_at=? WHERE id=?`,
			recapID, time.Now().UTC().Format(time.RFC3339), id)
		linkInkPagesTo(cap, "campaign", cap.CampaignID)
		c.JSON(http.StatusCreated, gin.H{"target": "recap", "id": recapID, "recap_id": recapID})

	case "journal":
		if cap.CharacterID == nil || !canEditCharacterID(c, *cap.CharacterID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "a journal entry needs a character you can edit"})
			return
		}
		res, err := db.DB.Exec(`INSERT INTO journal(character_id, title, entry, entry_date) VALUES(?,?,?,?)`,
			*cap.CharacterID, title, text, time.Now().Format("2006-01-02"))
		if err != nil {
			InternalError(c, err)
			return
		}
		journalID, _ := res.LastInsertId()
		db.DB.Exec(`UPDATE ink_captures SET journal_entry_id=?, status='promoted', updated_at=? WHERE id=?`,
			journalID, time.Now().UTC().Format(time.RFC3339), id)
		linkInkPagesTo(cap, "character", *cap.CharacterID)
		c.JSON(http.StatusCreated, gin.H{"target": "journal", "id": journalID, "journal_entry_id": journalID})

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "target must be 'recap' or 'journal'"})
	}
}

// linkInkPagesTo moves the page uploads onto the promoted entity so they appear
// in that entity's media gallery.
func linkInkPagesTo(cap *inkCapture, ownerType string, ownerID int64) {
	for _, p := range cap.Pages {
		if p.UploadID > 0 {
			db.DB.Exec("UPDATE uploads SET owner_type=?, owner_id=? WHERE id=?", ownerType, ownerID, p.UploadID)
		}
	}
}
