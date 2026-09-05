package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/ent"
	"villum/ent/oneshotact"
	"villum/ent/oneshotscene"
	"villum/models"
)

func HtmxNewActForm(c *gin.Context) {
	adventureID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ctx := c.Request.Context()

	ents, err := db.Client.OneShotAct.Query().
		Where(oneshotact.AdventureID(adventureID)).
		Order(oneshotact.BySortOrder(), oneshotact.ByID()).
		All(ctx)
	if err != nil {
		c.String(http.StatusInternalServerError, "query error: %v", err)
		return
	}

	acts := make([]models.OneShotAct, 0)
	for _, e := range ents {
		acts = append(acts, entActToModel(e))
	}

	a, err := loadAdventureDetail(ctx, adventureID)
	if err != nil {
		c.String(http.StatusNotFound, "not found")
		return
	}

	data := htmxOneShotData{
		Adventure: a,
		Act:       &models.OneShotAct{AdventureID: adventureID, Number: len(acts) + 1, SortOrder: len(acts) + 1, EstimatedMinutes: 30},
		Acts:      acts,
	}
	renderTemplate(c, "oneshot_act_form.html", data)
}

func HtmxEditActForm(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	actID := id
	ctx := c.Request.Context()

	entAct, err := db.Client.OneShotAct.Query().
		Where(oneshotact.ID(actID)).
		WithScenes(func(q *ent.OneShotSceneQuery) {
			q.Order(ent.Asc(oneshotscene.FieldSortOrder), ent.Asc(oneshotscene.FieldID))
		}).
		Only(ctx)
	if err != nil {
		c.String(http.StatusNotFound, "act not found: %v", err)
		return
	}
	ma := entActToModel(entAct)

	adventureID := ma.AdventureID
	ents, err := db.Client.OneShotAct.Query().
		Where(oneshotact.AdventureID(adventureID)).
		Order(oneshotact.BySortOrder(), oneshotact.ByID()).
		All(ctx)
	if err != nil {
		c.String(http.StatusInternalServerError, "query error: %v", err)
		return
	}

	acts := make([]models.OneShotAct, 0)
	for _, e := range ents {
		acts = append(acts, entActToModel(e))
	}

	a, err := loadAdventureDetail(ctx, adventureID)
	if err != nil {
		c.String(http.StatusNotFound, "not found")
		return
	}

	data := htmxOneShotData{
		Adventure: a,
		Act:       &ma,
		Acts:      acts,
	}
	renderTemplate(c, "oneshot_act_form.html", data)
}

func HtmxSceneForm(c *gin.Context) {
	actID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var adventureID int64
	db.DB.QueryRow("SELECT adventure_id FROM oneshot_acts WHERE id=?", actID).Scan(&adventureID)

	data := htmxOneShotData{
		Scene:      &models.OneShotScene{ActID: actID, SceneType: "roleplay", EstimatedMinutes: 15},
		SceneTypes: []string{"roleplay", "combat", "exploration", "puzzle", "climax"},
	}
	renderTemplate(c, "oneshot_scene_form.html", data)
}

func HtmxEditSceneForm(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ctx := c.Request.Context()

	entScene, err := db.Client.OneShotScene.Get(ctx, id)
	if err != nil {
		c.String(http.StatusNotFound, "scene not found: %v", err)
		return
	}

	ms := models.OneShotScene{
		ID: entScene.ID, ActID: entScene.ActID, Number: entScene.Number,
		SortOrder: entScene.SortOrder, Title: entScene.Title,
		Description: entScene.Description, SceneType: entScene.SceneType,
		EstimatedMinutes: entScene.EstimatedMinutes, Notes: entScene.Notes,
	}
	if entScene.LocationID != 0 {
		lid := entScene.LocationID
		ms.LocationID = &lid
	}
	if entScene.EncounterID != 0 {
		eid := entScene.EncounterID
		ms.EncounterID = &eid
	}

	data := htmxOneShotData{
		Scene:      &ms,
		SceneTypes: []string{"roleplay", "combat", "exploration", "puzzle", "climax"},
	}
	renderTemplate(c, "oneshot_scene_form.html", data)
}

// HTMX Act handlers
func HtmxCreateAct(c *gin.Context) {
	adventureID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	title := c.PostForm("title")
	description := c.PostForm("description")
	minutes, _ := strconv.Atoi(c.PostForm("estimated_minutes"))
	notes := c.PostForm("notes")
	parentActStr := c.PostForm("parent_act_id")

	if title == "" {
		title = "New Act"
	}
	if minutes <= 0 {
		minutes = 30
	}

	var number, sortOrder int
	db.DB.QueryRow("SELECT COALESCE(MAX(number),0)+1 FROM oneshot_acts WHERE adventure_id=?", adventureID).Scan(&number)
	db.DB.QueryRow("SELECT COALESCE(MAX(sort_order),0)+1 FROM oneshot_acts WHERE adventure_id=?", adventureID).Scan(&sortOrder)

	ctx := c.Request.Context()
	q := db.Client.OneShotAct.Create().
		SetAdventureID(adventureID).
		SetNumber(number).
		SetSortOrder(sortOrder).
		SetTitle(title).
		SetDescription(description).
		SetEstimatedMinutes(minutes).
		SetNotes(notes)
	if parentActStr != "" {
		if pid, err := strconv.ParseInt(parentActStr, 10, 64); err == nil {
			q.SetParentActID(pid)
		}
	}
	if _, err := q.Save(ctx); err != nil {
		c.String(http.StatusInternalServerError, "insert error: %v", err)
		return
	}

	HtmxGetOneShotDetail(c)
}

func HtmxUpdateAct(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	title := c.PostForm("title")
	description := c.PostForm("description")
	minutes, _ := strconv.Atoi(c.PostForm("estimated_minutes"))
	number, _ := strconv.Atoi(c.PostForm("number"))
	sortOrder, _ := strconv.Atoi(c.PostForm("sort_order"))
	notes := c.PostForm("notes")
	parentActStr := c.PostForm("parent_act_id")

	if number <= 0 {
		db.DB.QueryRow("SELECT number FROM oneshot_acts WHERE id=?", id).Scan(&number)
	}

	ctx := c.Request.Context()
	q := db.Client.OneShotAct.UpdateOneID(id).
		SetNumber(number).
		SetTitle(title).
		SetDescription(description).
		SetEstimatedMinutes(minutes).
		SetNotes(notes)
	if sortOrder > 0 {
		q.SetSortOrder(sortOrder)
	}
	if parentActStr != "" {
		if pid, err := strconv.ParseInt(parentActStr, 10, 64); err == nil {
			q.SetParentActID(pid)
		}
	}
	if _, err := q.Save(ctx); err != nil {
		c.String(http.StatusInternalServerError, "update error: %v", err)
		return
	}

	// Look up adventure ID from act
	var adventureID int64
	db.DB.QueryRow("SELECT adventure_id FROM oneshot_acts WHERE id=?", id).Scan(&adventureID)
	ReRenderOneShotDetail(c, adventureID)
}
func HtmxDeleteAct(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var adventureID int64
	db.DB.QueryRow("SELECT adventure_id FROM oneshot_acts WHERE id=?", id).Scan(&adventureID)
	db.DB.Exec("DELETE FROM oneshot_acts WHERE id=?", id)

	ReRenderOneShotDetail(c, adventureID)
}

// HTMX Scene handlers
func HtmxCreateScene(c *gin.Context) {
	actID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	title := c.PostForm("title")
	description := c.PostForm("description")
	sceneType := c.PostForm("scene_type")
	minutes, _ := strconv.Atoi(c.PostForm("estimated_minutes"))
	notes := c.PostForm("notes")

	if title == "" {
		title = "New Scene"
	}
	if sceneType == "" {
		sceneType = "roleplay"
	}
	if minutes <= 0 {
		minutes = 15
	}

	var number, sortOrder int
	db.DB.QueryRow("SELECT COALESCE(MAX(number),0)+1 FROM oneshot_scenes WHERE act_id=?", actID).Scan(&number)
	db.DB.QueryRow("SELECT COALESCE(MAX(sort_order),0)+1 FROM oneshot_scenes WHERE act_id=?", actID).Scan(&sortOrder)

	ctx := c.Request.Context()
	_, err := db.Client.OneShotScene.Create().
		SetActID(actID).
		SetNumber(number).
		SetSortOrder(sortOrder).
		SetTitle(title).
		SetDescription(description).
		SetSceneType(sceneType).
		SetEstimatedMinutes(minutes).
		SetNotes(notes).
		Save(ctx)
	if err != nil {
		c.String(http.StatusInternalServerError, "insert error: %v", err)
		return
	}

	// Look up adventure ID from act
	var adventureID int64
	db.DB.QueryRow("SELECT adventure_id FROM oneshot_acts WHERE id=?", actID).Scan(&adventureID)
	ReRenderOneShotDetail(c, adventureID)
}
func HtmxUpdateScene(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	title := c.PostForm("title")
	description := c.PostForm("description")
	sceneType := c.PostForm("scene_type")
	minutes, _ := strconv.Atoi(c.PostForm("estimated_minutes"))
	notes := c.PostForm("notes")
	number, _ := strconv.Atoi(c.PostForm("number"))
	sortOrder, _ := strconv.Atoi(c.PostForm("sort_order"))

	if number <= 0 {
		db.DB.QueryRow("SELECT number FROM oneshot_scenes WHERE id=?", id).Scan(&number)
	}

	ctx := c.Request.Context()
	q := db.Client.OneShotScene.UpdateOneID(id).
		SetNumber(number).
		SetTitle(title).
		SetDescription(description).
		SetSceneType(sceneType).
		SetEstimatedMinutes(minutes).
		SetNotes(notes)
	if sortOrder > 0 {
		q.SetSortOrder(sortOrder)
	}
	if _, err := q.Save(ctx); err != nil {
		c.String(http.StatusInternalServerError, "update error: %v", err)
		return
	}

	// Look up adventure ID from scene's act
	var adventureID int64
	db.DB.QueryRow("SELECT oa.adventure_id FROM oneshot_acts oa JOIN oneshot_scenes s ON s.act_id=oa.id WHERE s.id=?", id).Scan(&adventureID)
	ReRenderOneShotDetail(c, adventureID)
}
func HtmxDeleteScene(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var adventureID int64
	db.DB.QueryRow("SELECT oa.adventure_id FROM oneshot_acts oa JOIN oneshot_scenes s ON s.act_id=oa.id WHERE s.id=?", id).Scan(&adventureID)
	db.DB.Exec("DELETE FROM oneshot_scenes WHERE id=?", id)

	ReRenderOneShotDetail(c, adventureID)
}

// ─── Scene Dialog HTMX handlers ───

func HtmxNewDialogForm(c *gin.Context) {
	sceneID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	data := htmxOneShotData{
		Scene:  &models.OneShotScene{ID: sceneID},
		Dialog: &models.OneShotSceneDialog{SceneID: sceneID},
	}
	renderTemplate(c, "oneshot_dialog_form.html", data)
}

func HtmxEditDialogForm(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var d models.OneShotSceneDialog
	err := db.DB.QueryRow(
		"SELECT id, scene_id, sort_order, speaker, dialog_text, dm_notes, player_handout, condition FROM oneshot_scene_dialogs WHERE id=?", id,
	).Scan(&d.ID, &d.SceneID, &d.SortOrder, &d.Speaker, &d.DialogText, &d.DMNotes, &d.PlayerHandout, &d.Condition)
	if err != nil {
		c.String(http.StatusNotFound, "dialog not found")
		return
	}
	data := htmxOneShotData{
		Dialog: &d,
	}
	renderTemplate(c, "oneshot_dialog_form.html", data)
}

func renderDialogList(c *gin.Context, sceneID int64) {
	rows, err := db.DB.Query(
		"SELECT id, scene_id, sort_order, speaker, dialog_text, dm_notes, player_handout, condition, created_at FROM oneshot_scene_dialogs WHERE scene_id=? ORDER BY sort_order ASC, id ASC",
		sceneID,
	)
	if err != nil {
		c.String(http.StatusInternalServerError, "query error: %v", err)
		return
	}
	defer rows.Close()

	dialogs := make([]models.OneShotSceneDialog, 0)
	for rows.Next() {
		var d models.OneShotSceneDialog
		rows.Scan(&d.ID, &d.SceneID, &d.SortOrder, &d.Speaker, &d.DialogText, &d.DMNotes, &d.PlayerHandout, &d.Condition, &d.CreatedAt)
		dialogs = append(dialogs, d)
	}

	data := htmxOneShotData{
		Adventure: &models.OneShotAdventure{ID: sceneID},
		Dialogs:   dialogs,
	}
	renderTemplate(c, "oneshot_scene_dialogs.html", data)
}

func HtmxDialogList(c *gin.Context) {
	sceneID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	renderDialogList(c, sceneID)
}

func HtmxCreateDialog(c *gin.Context) {
	sceneID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	speaker := c.PostForm("speaker")
	dialogText := c.PostForm("dialog_text")
	dmNotes := c.PostForm("dm_notes")
	playerHandout := c.PostForm("player_handout")
	condition := c.PostForm("condition")

	var sortOrder int
	db.DB.QueryRow("SELECT COALESCE(MAX(sort_order),0)+1 FROM oneshot_scene_dialogs WHERE scene_id=?", sceneID).Scan(&sortOrder)

	_, err := db.DB.Exec(
		"INSERT INTO oneshot_scene_dialogs(scene_id, sort_order, speaker, dialog_text, dm_notes, player_handout, condition) VALUES(?,?,?,?,?,?,?)",
		sceneID, sortOrder, speaker, dialogText, dmNotes, playerHandout, condition,
	)
	if err != nil {
		c.String(http.StatusInternalServerError, "insert error: %v", err)
		return
	}

	renderDialogList(c, sceneID)
}

func HtmxUpdateDialog(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	speaker := c.PostForm("speaker")
	dialogText := c.PostForm("dialog_text")
	dmNotes := c.PostForm("dm_notes")
	playerHandout := c.PostForm("player_handout")
	condition := c.PostForm("condition")

	var sceneID int64
	db.DB.QueryRow("SELECT scene_id FROM oneshot_scene_dialogs WHERE id=?", id).Scan(&sceneID)

	_, err := db.DB.Exec(
		"UPDATE oneshot_scene_dialogs SET speaker=?, dialog_text=?, dm_notes=?, player_handout=?, condition=? WHERE id=?",
		speaker, dialogText, dmNotes, playerHandout, condition, id,
	)
	if err != nil {
		c.String(http.StatusInternalServerError, "update error: %v", err)
		return
	}

	renderDialogList(c, sceneID)
}

func HtmxDeleteDialog(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var sceneID int64
	db.DB.QueryRow("SELECT scene_id FROM oneshot_scene_dialogs WHERE id=?", id).Scan(&sceneID)
	db.DB.Exec("DELETE FROM oneshot_scene_dialogs WHERE id=?", id)

	// Re-render dialog list directly (HTMX swaps this HTML in)
	renderDialogList(c, sceneID)
}
