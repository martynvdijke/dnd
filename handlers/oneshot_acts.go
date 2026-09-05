package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/ent"
	"villum/ent/oneshotactnpc"
	"villum/models"
)

func CreateOneShotAct(c *gin.Context) {
	adventureID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var act models.OneShotAct
	if err := c.ShouldBindJSON(&act); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if act.Number == 0 {
		db.DB.QueryRow("SELECT COALESCE(MAX(number),0)+1 FROM oneshot_acts WHERE adventure_id=?", adventureID).Scan(&act.Number)
	}
	if act.SortOrder == 0 {
		db.DB.QueryRow("SELECT COALESCE(MAX(sort_order),0)+1 FROM oneshot_acts WHERE adventure_id=?", adventureID).Scan(&act.SortOrder)
	}
	ctx := c.Request.Context()
	q := db.Client.OneShotAct.Create().
		SetAdventureID(adventureID).
		SetNumber(act.Number).
		SetSortOrder(act.SortOrder).
		SetTitle(act.Title).
		SetDescription(act.Description).
		SetEstimatedMinutes(act.EstimatedMinutes).
		SetNotes(act.Notes)
	if act.ParentActID != nil && *act.ParentActID != 0 {
		q.SetParentActID(*act.ParentActID)
	}
	result, err := q.Save(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": result.ID})
}

func UpdateOneShotAct(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var act models.OneShotAct
	if err := c.ShouldBindJSON(&act); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := c.Request.Context()
	q := db.Client.OneShotAct.UpdateOneID(id).
		SetNumber(act.Number).
		SetSortOrder(act.SortOrder).
		SetTitle(act.Title).
		SetDescription(act.Description).
		SetEstimatedMinutes(act.EstimatedMinutes).
		SetNotes(act.Notes)
	if act.ParentActID != nil && *act.ParentActID != 0 {
		q.SetParentActID(*act.ParentActID)
	} else {
		q.ClearParentActID()
	}
	if _, err := q.Save(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteOneShotAct(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ctx := c.Request.Context()
	db.Client.OneShotAct.DeleteOneID(id).Exec(ctx)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func ListActNPCs(c *gin.Context) {
	actID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid act id"})
		return
	}
	npcs, err := db.Client.OneShotActNPC.Query().Where(oneshotactnpc.ActIDEQ(actID)).Order(ent.Asc(oneshotactnpc.FieldName)).All(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]models.OneShotActNPC, len(npcs))
	for i, n := range npcs {
		out[i] = models.OneShotActNPC{
			ID:        n.ID,
			ActID:     n.ActID,
			NPCID:     nil,
			Name:      n.Name,
			Role:      n.Role,
			Notes:     n.Notes,
			IsInline:  n.IsInline,
			CreatedAt: n.CreatedAt,
		}
		if n.NpcID != 0 {
			v := n.NpcID
			out[i].NPCID = &v
		}
	}
	c.JSON(http.StatusOK, out)
}

func CreateActNPC(c *gin.Context) {
	actID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid act id"})
		return
	}
	var input struct {
		NpcID *int64 `json:"npc_id,omitempty"`
		Name  string `json:"name"`
		Role  string `json:"role"`
		Notes string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	q := db.Client.OneShotActNPC.Create().SetActID(actID).SetName(input.Name).SetRole(input.Role).SetNotes(input.Notes)
	if input.NpcID != nil {
		q.SetNpcID(*input.NpcID).SetIsInline(false)
	} else {
		q.SetIsInline(true)
	}
	n, err := q.Save(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := models.OneShotActNPC{
		ID:        n.ID,
		ActID:     n.ActID,
		Name:      n.Name,
		Role:      n.Role,
		Notes:     n.Notes,
		IsInline:  n.IsInline,
		CreatedAt: n.CreatedAt,
	}
	if n.NpcID != 0 {
		v := n.NpcID
		out.NPCID = &v
	}
	c.JSON(http.StatusCreated, out)
}

func DeleteActNPC(c *gin.Context) {
	actID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid act id"})
		return
	}
	npcID, err := strconv.ParseInt(c.Param("nid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid npc id"})
		return
	}
	err = db.Client.OneShotActNPC.DeleteOneID(npcID).Exec(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "act_id": actID})
}

// ─── Act Notes (raw SQL) ───

func ListActNotes(c *gin.Context) {
	actID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid act id"})
		return
	}
	rows, err := db.DB.Query("SELECT id, adventure_id, user_id, title, content, created_at, updated_at FROM dm_notes WHERE act_id=? ORDER BY updated_at DESC", actID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := make([]models.DmNote, 0)
	for rows.Next() {
		var n models.DmNote
		rows.Scan(&n.ID, &n.AdventureID, &n.UserID, &n.Title, &n.Content, &n.CreatedAt, &n.UpdatedAt)
		out = append(out, n)
	}
	c.JSON(http.StatusOK, out)
}

func CreateActNote(c *gin.Context) {
	actID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid act id"})
		return
	}
	var input struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	var adventureID int64
	err = db.DB.QueryRow("SELECT adventure_id FROM oneshot_acts WHERE id=?", actID).Scan(&adventureID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "act not found"})
		return
	}
	result, err := db.DB.Exec("INSERT INTO dm_notes(adventure_id, user_id, title, content, act_id) VALUES(?,?,?,?,?)", adventureID, userID, input.Title, input.Content, actID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// ─── HTMX Act Details ───

type actDetailsData struct {
	Act   models.OneShotAct
	NPCs  []models.OneShotActNPC
	Notes []models.DmNote
}
