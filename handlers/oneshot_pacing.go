package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

func StartPacingSession(c *gin.Context) {
	adventureID := c.Param("id")
	userID, _ := c.Get("user_id")

	// Verify the adventure belongs to this user
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM oneshot_adventures WHERE id=? AND user_id=?)", adventureID, userID).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "adventure not found"})
		return
	}

	// Check if there's already an active session
	var existingID int64
	err = db.DB.QueryRow("SELECT id FROM session_pacing WHERE adventure_id=? AND status IN ('running','paused')", adventureID).Scan(&existingID)
	if err == nil {
		c.JSON(http.StatusOK, gin.H{"id": existingID, "message": "resumed existing session"})
		return
	}

	// Get first act and scene
	var firstActID, firstSceneID *int64
	var actID int64
	err = db.DB.QueryRow("SELECT id FROM oneshot_acts WHERE adventure_id=? ORDER BY number ASC LIMIT 1", adventureID).Scan(&actID)
	if err == nil {
		firstActID = &actID
		var sceneID int64
		err = db.DB.QueryRow("SELECT id FROM oneshot_scenes WHERE act_id=? ORDER BY number ASC LIMIT 1", actID).Scan(&sceneID)
		if err == nil {
			firstSceneID = &sceneID
		}
	}

	result, err := db.DB.Exec(
		"INSERT INTO session_pacing(adventure_id, current_act_id, current_scene_id, status, elapsed_seconds) VALUES(?,?,?,'running',0)",
		adventureID, firstActID, firstSceneID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()

	// If first scene exists, create initial scene timing
	if firstSceneID != nil {
		db.DB.Exec("INSERT INTO scene_timings(session_id, scene_id, status) VALUES(?,?,'active')", id, *firstSceneID)
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func GetPacingSession(c *gin.Context) {
	sessionID := c.Param("id")

	var s models.SessionPacing
	err := db.DB.QueryRow(`
		SELECT sp.id, sp.adventure_id, sp.current_act_id, sp.current_scene_id, sp.status, sp.elapsed_seconds, sp.started_at, COALESCE(sp.completed_at,''),
			COALESCE(oa.title,''), COALESCE(a.title,''), COALESCE(sc.title,''), COALESCE(sc.estimated_minutes,0),
			COALESCE(a.number,0), COALESCE(sc.number,0)
		FROM session_pacing sp
		LEFT JOIN oneshot_adventures oa ON oa.id = sp.adventure_id
		LEFT JOIN oneshot_acts a ON a.id = sp.current_act_id
		LEFT JOIN oneshot_scenes sc ON sc.id = sp.current_scene_id
		WHERE sp.id=?
	`, sessionID).Scan(&s.ID, &s.AdventureID, &s.CurrentActID, &s.CurrentSceneID, &s.Status, &s.ElapsedSeconds, &s.StartedAt, &s.CompletedAt,
		&s.AdventureTitle, &s.ActTitle, &s.SceneTitle, &s.SceneEstimated, &s.ActNumber, &s.SceneNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get total counts
	db.DB.QueryRow("SELECT COUNT(*) FROM oneshot_acts WHERE adventure_id=?", s.AdventureID).Scan(&s.TotalActs)
	db.DB.QueryRow("SELECT COUNT(*) FROM oneshot_scenes WHERE act_id=?", s.CurrentActID).Scan(&s.TotalScenes)

	// Get scene timings
	rows, err := db.DB.Query(`
		SELECT st.id, st.session_id, st.scene_id, st.elapsed_seconds, st.status, COALESCE(st.started_at,''), COALESCE(st.completed_at,''),
			COALESCE(sc.title,''), COALESCE(sc.scene_type,''), COALESCE(sc.estimated_minutes,0)
		FROM scene_timings st
		LEFT JOIN oneshot_scenes sc ON sc.id = st.scene_id
		WHERE st.session_id=?
		ORDER BY st.id
	`, sessionID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var st models.SceneTiming
			if err := rows.Scan(&st.ID, &st.SessionID, &st.SceneID, &st.ElapsedSeconds, &st.Status, &st.StartedAt, &st.CompletedAt,
				&st.SceneTitle, &st.SceneType, &st.EstimatedMin); err == nil {
				s.SceneTimings = append(s.SceneTimings, st)
			}
		}
	}

	c.JSON(http.StatusOK, s)
}

func UpdatePacingTimers(c *gin.Context) {
	sessionID := c.Param("id")

	// Increment elapsed_seconds for the session
	db.DB.Exec("UPDATE session_pacing SET elapsed_seconds = elapsed_seconds + 5 WHERE id=? AND status='running'", sessionID)

	// Increment current scene timing
	db.DB.Exec("UPDATE scene_timings SET elapsed_seconds = elapsed_seconds + 5 WHERE session_id=? AND status='active'", sessionID)

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func PausePacingSession(c *gin.Context) {
	sessionID := c.Param("id")
	_, err := db.DB.Exec("UPDATE session_pacing SET status='paused' WHERE id=? AND status='running'", sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "paused"})
}

func ResumePacingSession(c *gin.Context) {
	sessionID := c.Param("id")
	_, err := db.DB.Exec("UPDATE session_pacing SET status='running' WHERE id=? AND status='paused'", sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "running"})
}

func CompletePacingSession(c *gin.Context) {
	sessionID := c.Param("id")

	// Mark current scene timing as completed
	db.DB.Exec("UPDATE scene_timings SET status='completed', completed_at=datetime('now') WHERE session_id=? AND status='active'", sessionID)

	// Mark session as completed
	_, err := db.DB.Exec("UPDATE session_pacing SET status='completed', completed_at=datetime('now') WHERE id=? AND status IN ('running','paused')", sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "completed"})
}

func AdvanceToNextScene(c *gin.Context) {
	sessionID := c.Param("id")

	// Get current scene info
	var currentActID, currentSceneID int64
	err := db.DB.QueryRow("SELECT COALESCE(current_act_id,0), COALESCE(current_scene_id,0) FROM session_pacing WHERE id=?", sessionID).Scan(&currentActID, &currentSceneID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	if currentSceneID > 0 {
		// Mark current scene timing as completed
		db.DB.Exec(`UPDATE scene_timings SET status='completed', elapsed_seconds=COALESCE((SELECT elapsed_seconds FROM scene_timings WHERE session_id=? AND scene_id=? AND status='active' LIMIT 1), elapsed_seconds), completed_at=datetime('now') WHERE session_id=? AND scene_id=? AND status='active'`,
			sessionID, currentSceneID, sessionID, currentSceneID)
	}

	// If no current act/scene, try to set first scene from adventure
	if currentActID == 0 && currentSceneID == 0 {
		var advID int64
		db.DB.QueryRow("SELECT adventure_id FROM session_pacing WHERE id=?", sessionID).Scan(&advID)
		var firstActID int64
		err = db.DB.QueryRow("SELECT id FROM oneshot_acts WHERE adventure_id=? ORDER BY number ASC LIMIT 1", advID).Scan(&firstActID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no acts in adventure"})
			return
		}
		var firstSceneID int64
		err = db.DB.QueryRow("SELECT id FROM oneshot_scenes WHERE act_id=? ORDER BY number ASC LIMIT 1", firstActID).Scan(&firstSceneID)
		if err != nil {
			// Act has no scenes, advance to next act later
			db.DB.Exec("UPDATE session_pacing SET current_act_id=? WHERE id=?", firstActID, sessionID)
			currentActID = firstActID
		} else {
			db.DB.Exec("UPDATE session_pacing SET current_act_id=?, current_scene_id=? WHERE id=?", firstActID, firstSceneID, sessionID)
			db.DB.Exec("INSERT INTO scene_timings(session_id, scene_id, status) VALUES(?,?,'active')", sessionID, firstSceneID)
			c.JSON(http.StatusOK, gin.H{"status": "advanced"})
			return
		}
	}

	// Find next scene in same act
	if currentActID > 0 && currentSceneID > 0 {
		var nextSceneID int64
		err = db.DB.QueryRow("SELECT id FROM oneshot_scenes WHERE act_id=? AND number > (SELECT number FROM oneshot_scenes WHERE id=?) ORDER BY number ASC LIMIT 1",
			currentActID, currentSceneID).Scan(&nextSceneID)
		if err == nil {
			db.DB.Exec("UPDATE session_pacing SET current_scene_id=? WHERE id=?", nextSceneID, sessionID)
			db.DB.Exec("INSERT INTO scene_timings(session_id, scene_id, status) VALUES(?,?,'active')", sessionID, nextSceneID)
			c.JSON(http.StatusOK, gin.H{"status": "advanced"})
			return
		}
	}

	// No more scenes in this act, find next act
	if currentActID > 0 {
		var nextActID int64
		err = db.DB.QueryRow("SELECT id FROM oneshot_acts WHERE adventure_id=(SELECT adventure_id FROM session_pacing WHERE id=?) AND number > (SELECT number FROM oneshot_acts WHERE id=?) ORDER BY number ASC LIMIT 1",
			sessionID, currentActID).Scan(&nextActID)
		if err == nil {
			var firstSceneID int64
			err = db.DB.QueryRow("SELECT id FROM oneshot_scenes WHERE act_id=? ORDER BY number ASC LIMIT 1", nextActID).Scan(&firstSceneID)
			if err == nil {
				db.DB.Exec("UPDATE session_pacing SET current_act_id=?, current_scene_id=? WHERE id=?", nextActID, firstSceneID, sessionID)
				db.DB.Exec("INSERT INTO scene_timings(session_id, scene_id, status) VALUES(?,?,'active')", sessionID, firstSceneID)
				c.JSON(http.StatusOK, gin.H{"status": "advanced"})
				return
			}
			// Act has no scenes, advance to next act without scenes
			currentActID = nextActID
			db.DB.Exec("UPDATE session_pacing SET current_act_id=?, current_scene_id=NULL WHERE id=?", nextActID, sessionID)
		}
	}

	// No more acts - complete session
	db.DB.Exec("UPDATE session_pacing SET status='completed', completed_at=datetime('now') WHERE id=?", sessionID)
	c.JSON(http.StatusOK, gin.H{"status": "completed", "message": "all scenes completed"})
}
