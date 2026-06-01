package handlers

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

func CreateUploadLink(c *gin.Context) {
	bodyBytes, _ := io.ReadAll(c.Request.Body)
	c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	log.Printf("[DEBUG] CreateUploadLink body: %s, Content-Type: %s", string(bodyBytes), c.Request.Header.Get("Content-Type"))

	var req struct {
		UploadID   int64  `json:"upload_id"`
		EntityType string `json:"entity_type"`
		EntityID   int64  `json:"entity_id"`
		FieldName  string `json:"field_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[DEBUG] CreateUploadLink ShouldBindJSON error: %v", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[DEBUG] CreateUploadLink parsed: upload_id=%d entity_type=%q entity_id=%d field_name=%q", req.UploadID, req.EntityType, req.EntityID, req.FieldName)
	if req.UploadID == 0 || req.EntityType == "" || req.EntityID == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "upload_id, entity_type, and entity_id are required"})
		return
	}
	result, err := db.DB.Exec(
		"INSERT INTO upload_links (upload_id, entity_type, entity_id, field_name) VALUES (?, ?, ?, ?)",
		req.UploadID, req.EntityType, req.EntityID, req.FieldName)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

func DeleteUploadLink(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	_, err = db.DB.Exec("DELETE FROM upload_links WHERE id = ?", id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
