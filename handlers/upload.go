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
		".tiff": true, ".tif": true,
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

	mimeType := http.DetectContentType(data)
	if !strings.HasPrefix(mimeType, "image/") {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "File content does not match image type"})
		return
	}

	ownerType := c.PostForm("owner_type")
	ownerIDStr := c.PostForm("owner_id")
	var ownerID int64
	if ownerIDStr != "" {
		ownerID, _ = strconv.ParseInt(ownerIDStr, 10, 64)
	}
	allowedOwnerTypes := map[string]bool{"party": true, "item": true, "oneshot": true, "character": true, "npc": true, "": true}
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
		var existingURL string
		if err := db.DB.QueryRow("SELECT url FROM uploads WHERE hash = ?", hashStr).Scan(&existingURL); err == nil {
			c.JSON(http.StatusOK, gin.H{"url": existingURL, "duplicate": true})
			return
		}
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory"})
		return
	}

	uploadPath := filepath.Join(dir, filename)

	if ext != ".gif" && ext != ".svg" && ext != ".tif" && ext != ".tiff" {
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

	c.JSON(http.StatusOK, gin.H{
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
	ownerType := c.Query("owner_type")
	ownerID := c.Query("owner_id")

	query := "SELECT id, hash, ext, url, COALESCE(resized_url,''), COALESCE(thumbnail_url,''), owner_type, owner_id, COALESCE(created_at,'') FROM uploads"
	args := []interface{}{}
	if ownerType != "" && ownerID != "" {
		query += " WHERE owner_type=? AND owner_id=?"
		args = append(args, ownerType, ownerID)
	}
	query += " ORDER BY created_at DESC LIMIT 50"

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
