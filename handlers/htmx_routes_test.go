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

// The act-details modal fetches /htmx/oneshot-acts/:id/details (root HTMX
// group). It was previously registered on the /api oneshot group, so the
// request 404'd. Guard the path stays on the root table.
func TestHtmxActDetailsRouteRegisteredAtRoot(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	HtmxRegisterRoutes(r.Group(""))
	want := "GET /htmx/oneshot-acts/:id/details"
	for _, rt := range snapshotRoutes(r) {
		if rt == want {
			return
		}
	}
	t.Fatalf("route %q not registered in HtmxRegisterRoutes", want)
}
