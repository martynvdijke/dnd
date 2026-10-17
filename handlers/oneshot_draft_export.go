package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
)

// ExportOneShotDraft returns an existing one-shot adventure as an
// aiOneShotDraft JSON object, so the draft studio can revise it with AI.
// GET /api/oneshot-adventures/:id/draft
func ExportOneShotDraft(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	uid, ok := MustGetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	ctx := c.Request.Context()

	var owner int64
	draft := aiOneShotDraft{}
	err = db.DB.QueryRowContext(ctx, `SELECT user_id, title, premise, hook, difficulty, estimated_minutes, notes
		FROM oneshot_adventures WHERE id=?`, id).
		Scan(&owner, &draft.Title, &draft.Premise, &draft.Hook, &draft.Difficulty, &draft.EstimatedMinutes, &draft.Notes)
	if err == sql.ErrNoRows || (err == nil && owner != uid) {
		c.JSON(http.StatusNotFound, gin.H{"error": "one-shot adventure not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	actRows, err := db.DB.QueryContext(ctx, "SELECT id, title, description, estimated_minutes FROM oneshot_acts WHERE adventure_id=? ORDER BY number, id", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	type actRow struct {
		id  int64
		act aiOneShotAct
	}
	var acts []actRow
	for actRows.Next() {
		var ar actRow
		if err := actRows.Scan(&ar.id, &ar.act.Title, &ar.act.Description, &ar.act.EstimatedMinutes); err != nil {
			actRows.Close()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		acts = append(acts, ar)
	}
	actRows.Close()
	for _, ar := range acts {
		sceneRows, err := db.DB.QueryContext(ctx, "SELECT title, description, scene_type, estimated_minutes FROM oneshot_scenes WHERE act_id=? ORDER BY number, id", ar.id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for sceneRows.Next() {
			var s aiOneShotScene
			if err := sceneRows.Scan(&s.Title, &s.Description, &s.SceneType, &s.EstimatedMinutes); err != nil {
				sceneRows.Close()
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			ar.act.Scenes = append(ar.act.Scenes, s)
		}
		sceneRows.Close()
		draft.Acts = append(draft.Acts, ar.act)
	}

	if draft.NPCs, err = queryOneShotNPCs(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if draft.Locations, err = queryOneShotLocations(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if draft.Encounters, err = queryOneShotEncounters(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if draft.Clues, err = queryOneShotClues(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, draft)
}

func queryOneShotNPCs(ctx context.Context, adventureID int64) ([]aiOneShotNPC, error) {
	rows, err := db.DB.QueryContext(ctx, `SELECT n.name, n.race, n.description, l.role
		FROM oneshot_adventure_npcs l JOIN npcs n ON n.id = l.npc_id
		WHERE l.adventure_id=? ORDER BY n.name`, adventureID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []aiOneShotNPC
	for rows.Next() {
		var n aiOneShotNPC
		if err := rows.Scan(&n.Name, &n.Race, &n.Description, &n.Role); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func queryOneShotLocations(ctx context.Context, adventureID int64) ([]aiOneShotLocation, error) {
	rows, err := db.DB.QueryContext(ctx, `SELECT l.name, l.type, l.description
		FROM oneshot_adventure_locations al JOIN locations l ON l.id = al.location_id
		WHERE al.adventure_id=? ORDER BY l.name`, adventureID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []aiOneShotLocation
	for rows.Next() {
		var l aiOneShotLocation
		if err := rows.Scan(&l.Name, &l.Type, &l.Description); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func queryOneShotEncounters(ctx context.Context, adventureID int64) ([]aiOneShotEncounter, error) {
	rows, err := db.DB.QueryContext(ctx, `SELECT e.name, e.description, e.difficulty
		FROM oneshot_adventure_encounters ae JOIN encounter_templates e ON e.id = ae.encounter_id
		WHERE ae.adventure_id=? ORDER BY e.name`, adventureID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []aiOneShotEncounter
	for rows.Next() {
		var e aiOneShotEncounter
		if err := rows.Scan(&e.Name, &e.Description, &e.Difficulty); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func queryOneShotClues(ctx context.Context, adventureID int64) ([]aiOneShotClue, error) {
	rows, err := db.DB.QueryContext(ctx, `SELECT title, description, clue_type FROM clues
		WHERE adventure_id=? ORDER BY sort_order, id`, adventureID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []aiOneShotClue
	for rows.Next() {
		var cl aiOneShotClue
		if err := rows.Scan(&cl.Title, &cl.Description, &cl.ClueType); err != nil {
			return nil, err
		}
		out = append(out, cl)
	}
	return out, rows.Err()
}
