package handlers

import "github.com/gin-gonic/gin"

// Route defines a data-driven HTMX route entry.
type Route struct {
	Method  string
	Path    string
	Handler gin.HandlerFunc
}

// RegisterRoutes registers a slice of Route on the given group, applying any middleware per-route.
func RegisterRoutes(r *gin.RouterGroup, routes []Route) {
	for _, rt := range routes {
		r.Handle(rt.Method, rt.Path, rt.Handler)
	}
}
