package handlers

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
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

func HandleUpload(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
		".webp": true, ".avif": true, ".svg": true, ".bmp": true,
		".tiff": true, ".tif": true, ".pdf": true,
	}
	if !allowedExts[ext] {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid file type"})
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

	const maxSize = 10 * 1024 * 1024 // 10MB
	if len(data) > maxSize {
		c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{"error": "File too large (max 10MB)"})
		return
	}

	mimeType := http.DetectContentType(data)
	isPDF := ext == ".pdf" || mimeType == "application/pdf"
	if !isPDF && !strings.HasPrefix(mimeType, "image/") {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "File content does not match image type"})
		return
	}

	ownerType := c.PostForm("owner_type")
	ownerIDStr := c.PostForm("owner_id")
	var ownerID int64
	if ownerIDStr != "" {
		ownerID, _ = strconv.ParseInt(ownerIDStr, 10, 64)
	}
	allowedOwnerTypes := map[string]bool{"party": true, "item": true, "oneshot": true, "character": true, "npc": true, "campaign": true, "": true}
	if !allowedOwnerTypes[ownerType] {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid owner_type"})
		return
	}

	hash := sha256.Sum256(data)
	hashStr := hex.EncodeToString(hash[:])

	saveExt := ext
	if ext == ".jpeg" {
		saveExt = ".jpg"
	}
	if ext == ".tiff" {
		saveExt = ".tif"
	}

	subDir := hashStr[:2]
	dir := filepath.Join(MediaPath, subDir)

	filename := hashStr + saveExt
	url := fmt.Sprintf("/media/%s/%s", subDir, filename)
	var thumbnailURL string

	result, err := db.DB.Exec(
		"INSERT OR IGNORE INTO uploads (hash, ext, url, resized_url, thumbnail_url, owner_type, owner_id) VALUES (?, ?, ?, ?, ?, ?, ?)",
		hashStr, ext, url, url, thumbnailURL, ownerType, ownerID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		var existingID int64
		var existingURL string
		if err := db.DB.QueryRow("SELECT id, url FROM uploads WHERE hash = ?", hashStr).Scan(&existingID, &existingURL); err == nil {
			c.JSON(http.StatusOK, gin.H{"id": existingID, "url": existingURL, "duplicate": true})
			return
		}
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory"})
		return
	}

	uploadPath := filepath.Join(dir, filename)

	if ext != ".gif" && ext != ".svg" && ext != ".tif" && ext != ".tiff" && !isPDF {
		img, format, err := image.Decode(bytes.NewReader(data))
		if err == nil {
			size := img.Bounds().Size()
			if size.X > 10000 || size.Y > 10000 {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Image dimensions too large (max 10000x10000)"})
				return
			}
			if size.X > 1920 || size.Y > 1920 {
				img = resizeImage(img, 1920)
			}

			if err := saveImage(uploadPath, img, format); err != nil {
				if werr := os.WriteFile(uploadPath, data, 0644); werr != nil {
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
					return
				}
			}

			thumb := resizeImage(img, 300)
			thumbFilename := hashStr + "_thumb" + saveExt
			if err := saveImage(filepath.Join(dir, thumbFilename), thumb, format); err == nil {
				thumbnailURL = fmt.Sprintf("/media/%s/%s", subDir, thumbFilename)
			}
		} else {
			if err := os.WriteFile(uploadPath, data, 0644); err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
				return
			}
		}
	} else {
		if err := os.WriteFile(uploadPath, data, 0644); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}
	}

	if thumbnailURL != "" {
		db.DB.Exec("UPDATE uploads SET thumbnail_url = ? WHERE hash = ?", thumbnailURL, hashStr)
	}

	// Fetch upload ID after insert
	var uploadID int64
	db.DB.QueryRow("SELECT id FROM uploads WHERE hash = ?", hashStr).Scan(&uploadID)

	c.JSON(http.StatusOK, gin.H{
		"id":            uploadID,
		"url":           url,
		"resized_url":   url,
		"thumbnail_url": thumbnailURL,
		"hash":          hashStr,
	})
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

	query := `SELECT DISTINCT u.id, u.hash, u.ext, u.url, COALESCE(u.resized_url,''), COALESCE(u.thumbnail_url,''), u.owner_type, u.owner_id, COALESCE(u.created_at,'') FROM uploads u`
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
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	uploads := make([]models.Upload, 0)
	for rows.Next() {
		var u models.Upload
		if err := rows.Scan(&u.ID, &u.Hash, &u.Ext, &u.URL, &u.ResizedURL, &u.ThumbnailURL, &u.OwnerType, &u.OwnerID, &u.CreatedAt); err != nil {
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
	saveExt := ext
	if ext == ".jpeg" {
		saveExt = ".jpg"
	}
	if ext == ".tiff" {
		saveExt = ".tif"
	}

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
