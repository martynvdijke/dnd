package handlers

import (
	"fmt"
	"os"
	"path/filepath"

	"villum/db"
	"villum/models"
)

// LoadMediaBytes reads a stored upload's bytes off disk for outbound sending.
func LoadMediaBytes(uploadID int64) (*models.MediaFile, error) {
	var hash, ext, url, kind, source string
	err := db.DB.QueryRow(`SELECT hash, ext, url, COALESCE(kind,''), COALESCE(source,'') FROM uploads WHERE id=?`, uploadID).
		Scan(&hash, &ext, &url, &kind, &source)
	if err != nil {
		return nil, fmt.Errorf("upload %d not found: %w", uploadID, err)
	}
	if hash == "" || ext == "" {
		return nil, fmt.Errorf("upload %d not found", uploadID)
	}
	saveExt := normalizeSaveExt(ext)
	subDir := hash[:2]
	filename := hash + saveExt
	diskPath := filepath.Join(MediaPath, subDir, filename)
	data, err := os.ReadFile(diskPath)
	if err != nil {
		return nil, fmt.Errorf("read media file %d: %w", uploadID, err)
	}
	return &models.MediaFile{
		Data:     data,
		Filename: filename,
		Kind:     kind,
		Ext:      ext,
	}, nil
}
