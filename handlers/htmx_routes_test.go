package handlers

import (
	"sort"
	"testing"

	"github.com/gin-gonic/gin"
)

func snapshotRoutes(r *gin.Engine) []string {
	gin.SetMode(gin.TestMode)
	routes := r.Routes()
	out := make([]string, 0, len(routes))
	for _, rt := range routes {
		out = append(out, rt.Method+" "+rt.Path)
	}
	sort.Strings(out)
	return out
}

func TestHtmxRegisterRoutes_Snapshot(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r1 := gin.New()
	g1 := r1.Group("")
	HtmxRegisterRoutes(g1)
	snap1 := snapshotRoutes(r1)

	r2 := gin.New()
	g2 := r2.Group("")
	HtmxRegisterRoutes(g2)
	snap2 := snapshotRoutes(r2)

	if len(snap1) != len(snap2) {
		t.Fatalf("route count mismatch %d vs %d", len(snap1), len(snap2))
	}
	for i := range snap1 {
		if snap1[i] != snap2[i] {
			t.Fatalf("route mismatch at %d: %q vs %q", i, snap1[i], snap2[i])
		}
	}
	// Ensure we have a reasonable number of routes
	if len(snap1) < 50 {
		t.Fatalf("expected many routes, got %d", len(snap1))
	}
}
