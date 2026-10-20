package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"villum/middleware"
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

// WriteError writes a JSON error response with the given status. Server-side
// failures are logged and replaced with a generic message so database or
// filesystem details never reach the client; 4xx validation messages are
// preserved for the UI.
func WriteError(c *gin.Context, status int, err error) {
	if status >= http.StatusInternalServerError {
		InternalError(c, err)
		return
	}
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	c.JSON(status, gin.H{"error": msg})
}

// InternalError logs the underlying error server-side and returns a generic
// 500 response. Use this instead of c.JSON(500, gin.H{"error": err.Error()}).
func InternalError(c *gin.Context, err error) {
	detail := "unknown error"
	if err != nil {
		detail = err.Error()
	}
	middleware.LogError("http", "request failed",
		"method", c.Request.Method, "path", c.Request.URL.Path, "error", detail)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}

// InternalErrorText is the text/HTML counterpart of InternalError for HTMX
// partial responses. It logs the detail and returns a generic 500 body.
func InternalErrorText(c *gin.Context, err error) {
	detail := "unknown error"
	if err != nil {
		detail = err.Error()
	}
	middleware.LogError("http", "request failed",
		"method", c.Request.Method, "path", c.Request.URL.Path, "error", detail)
	c.String(http.StatusInternalServerError, "internal server error")
}

// WriteNotFound writes a 404 JSON error response.
func WriteNotFound(c *gin.Context, msg string) {
	c.JSON(http.StatusNotFound, gin.H{"error": msg})
}

// WriteJSON writes a JSON payload with the given status.
func WriteJSON(c *gin.Context, status int, payload any) {
	c.JSON(status, payload)
}
