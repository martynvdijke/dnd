package handlers

import (
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers/testutil"
)

// TestSceneSpecialEffects covers persisting a scene effect and playing it.
func TestSceneSpecialEffects(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "dm", "admin")
	testutil.SeedUser(t, 2, "other", "user")
	testutil.SeedOneShot(t, 10, 1, "Adventure")
	testutil.SeedOneShotAct(t, 100, 10, "Act One", 1)
	testutil.SeedOneShotScene(t, 1000, 100, "Ambush", 1)

	routes := func(auth *gin.RouterGroup) {
		auth.PUT("/oneshot-scenes/:id", UpdateOneShotScene)
		auth.GET("/htmx/oneshot-scenes/:id/edit", HtmxEditSceneForm)
		auth.POST("/oneshot-scenes/:id/effect", TriggerSceneEffect)
	}

	r := testutil.NewRouter(routes)

	// Persist an effect on the scene.
	w := testutil.PutJSON(t, r, "/api/oneshot-scenes/1000", gin.H{"special_effects": "fire", "title": "Ambush"})
	testutil.AssertStatus(t, w, 200)

	var fx string
	db.DB.QueryRow("SELECT special_effects FROM oneshot_scenes WHERE id=1000").Scan(&fx)
	if fx != "fire" {
		t.Fatalf("special_effects = %q, want fire", fx)
	}

	// Edit form prefills the chosen effect.
	w = testutil.Get(t, r, "/api/htmx/oneshot-scenes/1000/edit")
	if w.Code != 200 || !strings.Contains(w.Body.String(), `value="fire" selected`) {
		t.Fatalf("edit form missing selected effect: code=%d body=%s", w.Code, w.Body.String())
	}

	// Standalone owner can play the stored effect.
	w = testutil.PostJSON(t, r, "/api/oneshot-scenes/1000/effect", gin.H{})
	testutil.AssertStatus(t, w, 200)
	if !strings.Contains(w.Body.String(), `"effect":"fire"`) {
		t.Fatalf("trigger body = %s, want fire", w.Body.String())
	}

	// A non-owner cannot trigger it.
	rOther := testutil.NewRouterWithUser(routes, 2, "user")
	w = testutil.PostJSON(t, rOther, "/api/oneshot-scenes/1000/effect", gin.H{})
	testutil.AssertStatus(t, w, 403)

	// Unknown scene is a 404.
	w = testutil.PostJSON(t, r, "/api/oneshot-scenes/9999/effect", gin.H{})
	testutil.AssertStatus(t, w, 404)
}
