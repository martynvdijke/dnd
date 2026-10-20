package handlers

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/image/draw"

	"villum/db"
	"villum/models"
)

var MediaPath = "/app/media"

func SetMediaPath(path string) {
	MediaPath = path
}

// mediaExtKinds maps supported upload extensions to their media kind.
var mediaExtKinds = map[string]string{
	".jpg": "image", ".jpeg": "image", ".png": "image", ".gif": "image",
	".webp": "image", ".avif": "image", ".svg": "image", ".bmp": "image",
	".tiff": "image", ".tif": "image", ".pdf": "pdf",
	".mp3": "audio", ".ogg": "audio", ".oga": "audio", ".wav": "audio",
	".m4a": "audio", ".aac": "audio", ".flac": "audio", ".opus": "audio",
	".mp4": "video", ".webm": "video", ".mov": "video", ".m4v": "video",
}

// mediaKindMaxSize caps uploads per kind; audio and video are bounded by what
// the Telegram Bot API accepts for the corresponding send methods.
var mediaKindMaxSize = map[string]int{
	"image": 10 * 1024 * 1024,
	"pdf":   10 * 1024 * 1024,
	"audio": 25 * 1024 * 1024,
	"video": 50 * 1024 * 1024,
}

// mediaError carries an HTTP status for validation failures from the shared
// media pipeline so callers can map them without duplicating checks.
type mediaError struct {
	status int
	msg    string
}

func (e *mediaError) Error() string { return e.msg }

func newMediaError(status int, msg string) *mediaError {
	return &mediaError{status: status, msg: msg}
}

// normalizeSaveExt maps extensions to their canonical on-disk form.
func normalizeSaveExt(ext string) string {
	switch ext {
	case ".jpeg":
		return ".jpg"
	case ".tiff":
		return ".tif"
	}
	return ext
}

// validateMedia classifies an upload by extension and cross-checks the sniffed
// content type. Extensions stay the primary signal because Go's sniffer cannot
// classify every allowed container (svg, m4a, flac, opus).
func validateMedia(ext, mimeType string) (string, *mediaError) {
	kind, ok := mediaExtKinds[ext]
	if !ok {
		return "", newMediaError(http.StatusBadRequest, "Invalid file type")
	}
	switch kind {
	case "image":
		if ext == ".svg" {
			return kind, nil
		}
		if !strings.HasPrefix(mimeType, "image/") {
			return "", newMediaError(http.StatusBadRequest, "File content does not match image type")
		}
	case "pdf":
		if mimeType != "application/pdf" && mimeType != "application/octet-stream" {
			return "", newMediaError(http.StatusBadRequest, "File content does not match PDF type")
		}
	case "audio":
		if !strings.HasPrefix(mimeType, "audio/") && mimeType != "application/ogg" && mimeType != "application/octet-stream" {
			return "", newMediaError(http.StatusBadRequest, "File content does not match audio type")
		}
	case "video":
		if !strings.HasPrefix(mimeType, "video/") && mimeType != "application/octet-stream" {
			return "", newMediaError(http.StatusBadRequest, "File content does not match video type")
		}
	}
	return kind, nil
}

// uploadByHash loads a stored upload by content hash.
func uploadByHash(hashStr string) (*models.Upload, error) {
	var u models.Upload
	err := db.DB.QueryRow(`SELECT id, hash, ext, url, COALESCE(resized_url,''), COALESCE(thumbnail_url,''),
			COALESCE(owner_type,''), COALESCE(owner_id,0), COALESCE(kind,''), COALESCE(source,''),
			COALESCE(telegram_file_unique_id,''), COALESCE(created_at,'')
		FROM uploads WHERE hash=?`, hashStr).
		Scan(&u.ID, &u.Hash, &u.Ext, &u.URL, &u.ResizedURL, &u.ThumbnailURL,
			&u.OwnerType, &u.OwnerID, &u.Kind, &u.Source, &u.TelegramFileUniqueID, &u.CreatedAt)
	return &u, err
}

// writeMediaFile stores the bytes on disk, resizing images and writing a
// thumbnail where applicable. It returns the thumbnail URL when one was made.
func writeMediaFile(dir, filename, saveExt, ext, hashStr, kind string, data []byte) (string, error) {
	uploadPath := filepath.Join(dir, filename)

	if kind != "image" || ext == ".gif" || ext == ".svg" || ext == ".tif" || ext == ".tiff" {
		if err := os.WriteFile(uploadPath, data, 0644); err != nil {
			return "", err
		}
		return "", nil
	}

	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		if err := os.WriteFile(uploadPath, data, 0644); err != nil {
			return "", err
		}
		return "", nil
	}

	size := img.Bounds().Size()
	if size.X > 10000 || size.Y > 10000 {
		return "", newMediaError(http.StatusBadRequest, "Image dimensions too large (max 10000x10000)")
	}
	if size.X > 1920 || size.Y > 1920 {
		img = resizeImage(img, 1920)
	}

	if err := saveImage(uploadPath, img, format); err != nil {
		if werr := os.WriteFile(uploadPath, data, 0644); werr != nil {
			return "", werr
		}
	}

	thumb := resizeImage(img, 300)
	thumbFilename := hashStr + "_thumb" + saveExt
	if err := saveImage(filepath.Join(dir, thumbFilename), thumb, format); err == nil {
		return fmt.Sprintf("/media/%s/%s", filepath.Base(dir), thumbFilename), nil
	}
	return "", nil
}

// SaveMediaBytes validates, stores and records a media file. It is the single
// pipeline shared by HTTP uploads and inbound Telegram media. The boolean
// reports a duplicate that was already stored under the same content hash.
func SaveMediaBytes(name string, data []byte, ownerType string, ownerID int64, source, telegramFileUniqueID string) (*models.Upload, bool, error) {
	ext := strings.ToLower(filepath.Ext(name))
	mimeType := http.DetectContentType(data)

	kind, verr := validateMedia(ext, mimeType)
	if verr != nil {
		return nil, false, verr
	}
	if maxSize := mediaKindMaxSize[kind]; maxSize > 0 && len(data) > maxSize {
		return nil, false, newMediaError(http.StatusRequestEntityTooLarge, fmt.Sprintf("File too large (max %dMB)", maxSize/(1024*1024)))
	}

	hash := sha256.Sum256(data)
	hashStr := hex.EncodeToString(hash[:])
	saveExt := normalizeSaveExt(ext)
	subDir := hashStr[:2]
	dir := filepath.Join(MediaPath, subDir)
	filename := hashStr + saveExt
	url := fmt.Sprintf("/media/%s/%s", subDir, filename)

	result, err := db.DB.Exec(
		`INSERT OR IGNORE INTO uploads (hash, ext, url, resized_url, thumbnail_url, owner_type, owner_id, kind, source, telegram_file_unique_id)
		 VALUES (?, ?, ?, ?, '', ?, ?, ?, ?, ?)`,
		hashStr, ext, url, url, ownerType, ownerID, kind, source, telegramFileUniqueID)
	if err != nil {
		return nil, false, err
	}

	if rows, _ := result.RowsAffected(); rows == 0 {
		up, err := uploadByHash(hashStr)
		if err != nil {
			return nil, false, err
		}
		// A file already stored from a Telegram message may not have carried
		// its unique id yet (e.g. replayed update); remember it for dedupe.
		if telegramFileUniqueID != "" && up.TelegramFileUniqueID == "" {
			db.DB.Exec("UPDATE uploads SET telegram_file_unique_id=? WHERE id=?", telegramFileUniqueID, up.ID)
			up.TelegramFileUniqueID = telegramFileUniqueID
		}
		return up, true, nil
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, false, err
	}

	thumbnailURL, err := writeMediaFile(dir, filename, saveExt, ext, hashStr, kind, data)
	if err != nil {
		db.DB.Exec("DELETE FROM uploads WHERE hash=?", hashStr)
		os.Remove(filepath.Join(dir, filename))
		return nil, false, err
	}
	if thumbnailURL != "" {
		db.DB.Exec("UPDATE uploads SET thumbnail_url=? WHERE hash=?", thumbnailURL, hashStr)
	}

	up, err := uploadByHash(hashStr)
	if err != nil {
		return nil, false, err
	}
	return up, false, nil
}

func HandleUpload(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer src.Close()

	data, err := io.ReadAll(src)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	ownerType := c.PostForm("owner_type")
	ownerIDStr := c.PostForm("owner_id")
	var ownerID int64
	if ownerIDStr != "" {
		ownerID, err = strconv.ParseInt(ownerIDStr, 10, 64)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid owner_id"})
			return
		}
	}
	allowedOwnerTypes := map[string]bool{"party": true, "item": true, "oneshot": true, "character": true, "npc": true, "campaign": true, "": true}
	if !allowedOwnerTypes[ownerType] {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid owner_type"})
		return
	}

	up, duplicate, err := SaveMediaBytes(file.Filename, data, ownerType, ownerID, "upload", "")
	if err != nil {
		var me *mediaError
		if errors.As(err, &me) {
			c.AbortWithStatusJSON(me.status, gin.H{"error": me.msg})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	resp := gin.H{
		"id":            up.ID,
		"url":           up.URL,
		"resized_url":   up.URL,
		"thumbnail_url": up.ThumbnailURL,
		"hash":          up.Hash,
		"kind":          up.Kind,
	}
	if duplicate {
		resp["duplicate"] = true
	}
	c.JSON(http.StatusOK, resp)
}

func resizeImage(img image.Image, maxDim int) image.Image {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	if w <= maxDim && h <= maxDim {
		return img
	}

	ratio := float64(maxDim) / float64(max(w, h))
	newW := int(float64(w) * ratio)
	newH := int(float64(h) * ratio)

	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)

	return dst
}

func saveImage(path string, img image.Image, format string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	switch format {
	case "png":
		return png.Encode(f, img)
	default:
		return jpeg.Encode(f, img, &jpeg.Options{Quality: 85})
	}
}

func GetUploads(c *gin.Context) {
	entityType := c.Query("entity_type")
	entityIDStr := c.Query("entity_id")
	ownerType := c.Query("owner_type")
	ownerID := c.Query("owner_id")

	query := `SELECT DISTINCT u.id, u.hash, u.ext, u.url, COALESCE(u.resized_url,''), COALESCE(u.thumbnail_url,''), u.owner_type, u.owner_id, COALESCE(u.kind,''), COALESCE(u.created_at,'') FROM uploads u`
	args := []any{}
	conditions := []string{}

	if entityType != "" && entityIDStr != "" {
		query += ` LEFT JOIN upload_links ul ON u.id = ul.upload_id`
		conditions = append(conditions, "(ul.entity_type=? AND ul.entity_id=?)")
		args = append(args, entityType, entityIDStr)
	}

	if ownerType != "" && ownerID != "" {
		conditions = append(conditions, "(u.owner_type=? AND u.owner_id=?)")
		args = append(args, ownerType, ownerID)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " OR ")
	}

	query += " ORDER BY u.created_at DESC LIMIT 50"

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		InternalError(c, err)
		c.Abort()
		return
	}
	defer rows.Close()

	uploads := make([]models.Upload, 0)
	for rows.Next() {
		var u models.Upload
		if err := rows.Scan(&u.ID, &u.Hash, &u.Ext, &u.URL, &u.ResizedURL, &u.ThumbnailURL, &u.OwnerType, &u.OwnerID, &u.Kind, &u.CreatedAt); err != nil {
			continue
		}
		uploads = append(uploads, u)
	}
	c.JSON(http.StatusOK, uploads)
}

func HandleCropUpload(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid upload id"})
		return
	}

	var req struct {
		X            int `json:"x"`
		Y            int `json:"y"`
		Width        int `json:"width"`
		Height       int `json:"height"`
		TargetWidth  int `json:"target_width"`
		TargetHeight int `json:"target_height"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Width <= 0 || req.Height <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid crop dimensions"})
		return
	}

	var hash, ext, url string
	err = db.DB.QueryRow("SELECT hash, ext, url FROM uploads WHERE id = ?", id).Scan(&hash, &ext, &url)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "upload not found"})
		return
	}

	// Normalize extension for file access
	saveExt := normalizeSaveExt(ext)

	subDir := hash[:2]
	srcPath := filepath.Join(MediaPath, subDir, hash+saveExt)

	data, err := os.ReadFile(srcPath)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to read original"})
		return
	}

	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to decode image"})
		return
	}

	// Sub-image crop
	bounds := img.Bounds()
	if req.X+req.Width > bounds.Dx() || req.Y+req.Height > bounds.Dy() {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "crop area exceeds image bounds"})
		return
	}

	cropped := image.NewRGBA(image.Rect(0, 0, req.Width, req.Height))
	draw.CatmullRom.Scale(cropped, cropped.Bounds(), img, image.Rect(req.X, req.Y, req.X+req.Width, req.Y+req.Height), draw.Over, nil)

	// Resize to target if specified
	var final image.Image = cropped
	if req.TargetWidth > 0 && req.TargetHeight > 0 {
		maxDim := max(req.TargetHeight, req.TargetWidth)
		final = resizeImage(cropped, maxDim)
	}

	cropFilename := hash + "_crop" + saveExt
	cropPath := filepath.Join(MediaPath, subDir, cropFilename)
	if err := saveImage(cropPath, final, format); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to save cropped image"})
		return
	}

	cropURL := fmt.Sprintf("/media/%s/%s", subDir, cropFilename)
	db.DB.Exec("UPDATE uploads SET resized_url = ? WHERE id = ?", cropURL, id)

	// Regenerate thumbnail
	thumb := resizeImage(final, 300)
	thumbFilename := hash + "_thumb" + saveExt
	thumbPath := filepath.Join(MediaPath, subDir, thumbFilename)
	thumbURL := ""
	if err := saveImage(thumbPath, thumb, format); err == nil {
		thumbURL = fmt.Sprintf("/media/%s/%s", subDir, thumbFilename)
		db.DB.Exec("UPDATE uploads SET thumbnail_url = ? WHERE id = ?", thumbURL, id)
	}

	c.JSON(http.StatusOK, gin.H{
		"url":           url,
		"resized_url":   cropURL,
		"thumbnail_url": thumbURL,
	})
}
