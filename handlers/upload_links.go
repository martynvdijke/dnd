package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

func CreateUploadLink(c *gin.Context) {
	var req struct {
		UploadID   int64  `json:"upload_id"`
		EntityType string `json:"entity_type"`
		EntityID   int64  `json:"entity_id"`
		FieldName  string `json:"field_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.UploadID == 0 || req.EntityType == "" || req.EntityID == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "upload_id, entity_type, and entity_id are required"})
		return
	}
	result, err := db.DB.Exec(
		"INSERT INTO upload_links (upload_id, entity_type, entity_id, field_name) VALUES (?, ?, ?, ?)",
		req.UploadID, req.EntityType, req.EntityID, req.FieldName)
	if err != nil {
		InternalError(c, err)
		c.Abort()
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, models.UploadLink{
		ID:         id,
		UploadID:   req.UploadID,
		EntityType: req.EntityType,
		EntityID:   req.EntityID,
		FieldName:  req.FieldName,
	})
}

// LinkUploadToEntity attaches an existing upload to an entity's gallery.
// Galleries list their media through upload_links, so anything that stores an
// upload without going through CreateUploadLink (for example inbound Telegram
// media) must call this or it will never appear in the gallery.
// Idempotent: upload_links has UNIQUE(upload_id, entity_type, entity_id).
func LinkUploadToEntity(uploadID int64, entityType string, entityID int64) error {
	if uploadID == 0 || entityType == "" || entityID == 0 {
		return nil
	}
	_, err := db.DB.Exec(
		"INSERT OR IGNORE INTO upload_links (upload_id, entity_type, entity_id, field_name) VALUES (?, ?, ?, '')",
		uploadID, entityType, entityID)
	return err
}

func DeleteUploadLink(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	_, err = db.DB.Exec("DELETE FROM upload_links WHERE id = ?", id)
	if err != nil {
		InternalError(c, err)
		c.Abort()
		return
	}
	c.Status(http.StatusNoContent)
}

// GetUploadsForEntity lists uploads linked to an entity (type + id), newest first.
// Returns the same item shape used by the media gallery (mediaUploadItem).
func GetUploadsForEntity(c *gin.Context) {
	entityType := c.Param("type")
	entityID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if entityType == "" || err != nil || entityID == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "entity_type and entity_id are required"})
		return
	}

	rows, err := db.DB.Query(`
		SELECT u.id, u.hash, u.ext, u.url, COALESCE(u.resized_url,''), COALESCE(u.thumbnail_url,''),
			u.owner_type, u.owner_id, COALESCE(u.created_at,''), ul.id as link_id, COALESCE(u.kind,'')
		FROM uploads u
		JOIN upload_links ul ON u.id = ul.upload_id
		WHERE ul.entity_type = ? AND ul.entity_id = ?
		ORDER BY u.created_at DESC`, entityType, entityID)
	if err != nil {
		InternalError(c, err)
		c.Abort()
		return
	}
	defer rows.Close()

	var uploads []mediaUploadItem
	for rows.Next() {
		var item mediaUploadItem
		var ownerTypeStr, ownerIDStr string
		rows.Scan(&item.ID, &item.Hash, &item.Ext, &item.URL, &item.ResizedURL,
			&item.ThumbnailURL, &ownerTypeStr, &ownerIDStr, &item.CreatedAt, &item.LinkID, &item.Kind)
		item.OwnerType = ownerTypeStr
		if parsed, err := strconv.ParseInt(ownerIDStr, 10, 64); err == nil {
			item.OwnerID = parsed
		} else if ownerIDStr != "" {
			item.OwnerID = 0
		}
		item.IsPDF = item.Ext == ".pdf"
		uploads = append(uploads, item)
	}
	if uploads == nil {
		uploads = []mediaUploadItem{}
	}
	c.JSON(http.StatusOK, gin.H{"uploads": uploads})
}
