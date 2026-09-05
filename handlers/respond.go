package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// BindOr400 attempts to bind JSON body into obj. On failure it writes a 400
// response with {"error": err.Error()} and returns false. Callers should
// return immediately when false is returned.
func BindOr400(c *gin.Context, obj any) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return false
	}
	return true
}

// WriteError writes a JSON error response with the given status.
func WriteError(c *gin.Context, status int, err error) {
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	c.JSON(status, gin.H{"error": msg})
}

// WriteNotFound writes a 404 JSON error response.
func WriteNotFound(c *gin.Context, msg string) {
	c.JSON(http.StatusNotFound, gin.H{"error": msg})
}

// WriteJSON writes a JSON payload with the given status.
func WriteJSON(c *gin.Context, status int, payload any) {
	c.JSON(status, payload)
}
