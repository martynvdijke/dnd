package handlers

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

// ─── Pacing clock helpers ───
//
// Elapsed time is derived from timestamps, not from a heartbeat: a running
// session's started_at marks the beginning of the current interval, and
// elapsed_seconds holds the total accrued up to the last pause/accrual.
// Readers add (now - started_at) while the session is running; mutating
// operations persist the delta before changing state. This keeps the clock
// accurate across htmx swaps, reloads and server restarts.

func pacingSecondsSince(ts string) int {
	if ts == "" {
		return 0
	}
	var secs int
	if err := db.DB.QueryRow("SELECT CAST(MAX(0, strftime('%s','now') - strftime('%s', ?)) AS INTEGER)", ts).Scan(&secs); err != nil {
		return 0
	}
	return secs
}

// accrueSessionClock adds the running interval to elapsed_seconds and
// re-anchors started_at so the delta is never counted twice.
func accrueSessionClock(sessionID any) {
	db.DB.Exec(`UPDATE session_pacing
		SET elapsed_seconds = elapsed_seconds + CAST(MAX(0, strftime('%s','now') - strftime('%s', started_at)) AS INTEGER),
		    started_at = datetime('now')
		WHERE id=? AND status='running'`, sessionID)
}

// accrueActiveSceneClock does the same for the session's active scene timing.
func accrueActiveSceneClock(sessionID any) {
	db.DB.Exec(`UPDATE scene_timings
		SET elapsed_seconds = elapsed_seconds + CAST(MAX(0, strftime('%s','now') - strftime('%s', started_at)) AS INTEGER),
		    started_at = datetime('now')
		WHERE session_id=? AND status='active'`, sessionID)
}

// resumePacingClock restarts the clock for a paused session and its active
// scene, so time spent paused is not counted.
func resumePacingClock(sessionID any) {
	db.DB.Exec("UPDATE session_pacing SET status='running', started_at=datetime('now') WHERE id=? AND status='paused'", sessionID)
	db.DB.Exec("UPDATE scene_timings SET started_at=datetime('now') WHERE session_id=? AND status='active'", sessionID)
}

// loadPacingSession loads a session with joined titles and scene timings and
// applies the effective (timestamp-aware) elapsed values for running sessions.
func loadPacingSession(sessionID any) (*models.SessionPacing, error) {
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
		return nil, err
	}

	if s.Status == "running" {
		s.ElapsedSeconds += pacingSecondsSince(s.StartedAt)
	}

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
				if st.Status == "active" && s.Status == "running" {
					st.ElapsedSeconds += pacingSecondsSince(st.StartedAt)
				}
				s.SceneTimings = append(s.SceneTimings, st)
			}
		}
	}

	return &s, nil
}

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
	var existingStatus string
	err = db.DB.QueryRow("SELECT id, status FROM session_pacing WHERE adventure_id=? AND status IN ('running','paused')", adventureID).Scan(&existingID, &existingStatus)
	if err == nil {
		// A paused session resumes where it left off; a running one is reused as-is.
		if existingStatus == "paused" {
			resumePacingClock(existingID)
		}
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
	s, err := loadPacingSession(c.Param("id"))
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Counts for the current act's scenes
	db.DB.QueryRow("SELECT COUNT(*) FROM oneshot_acts WHERE adventure_id=?", s.AdventureID).Scan(&s.TotalActs)
	db.DB.QueryRow("SELECT COUNT(*) FROM oneshot_scenes WHERE act_id=?", s.CurrentActID).Scan(&s.TotalScenes)

	c.JSON(http.StatusOK, s)
}

// GetAdventurePacing resolves an adventure id to its latest pacing session,
// preferring a running or paused session over a completed one.
func GetAdventurePacing(c *gin.Context) {
	adventureID := c.Param("id")

	var sessionID int64
	err := db.DB.QueryRow(`
		SELECT id FROM session_pacing
		WHERE adventure_id=?
		ORDER BY CASE status WHEN 'running' THEN 0 WHEN 'paused' THEN 1 ELSE 2 END,
			COALESCE(completed_at, started_at) DESC, id DESC
		LIMIT 1
	`, adventureID).Scan(&sessionID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "no pacing session for adventure"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	s, err := loadPacingSession(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	db.DB.QueryRow("SELECT COUNT(*) FROM oneshot_acts WHERE adventure_id=?", s.AdventureID).Scan(&s.TotalActs)
	db.DB.QueryRow("SELECT COUNT(*) FROM oneshot_scenes WHERE act_id=?", s.CurrentActID).Scan(&s.TotalScenes)

	c.JSON(http.StatusOK, s)
}

// ─── Pacing mutations (shared by JSON and HTMX handlers) ───

func pausePacingSession(sessionID string) error {
	// Persist the running interval before freezing the clock.
	accrueSessionClock(sessionID)
	accrueActiveSceneClock(sessionID)
	_, err := db.DB.Exec("UPDATE session_pacing SET status='paused' WHERE id=? AND status='running'", sessionID)
	return err
}

func PausePacingSession(c *gin.Context) {
	if err := pausePacingSession(c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "paused"})
}

func resumePacingSession(sessionID string) (string, error) {
	var status string
	if err := db.DB.QueryRow("SELECT status FROM session_pacing WHERE id=?", sessionID).Scan(&status); err != nil {
		return "", err
	}
	if status == "paused" {
		resumePacingClock(sessionID)
	}
	return status, nil
}

func ResumePacingSession(c *gin.Context) {
	if _, err := resumePacingSession(c.Param("id")); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "running"})
}

func completePacingSession(sessionID string) error {
	// Persist the running interval before closing the session and its scene.
	accrueSessionClock(sessionID)
	accrueActiveSceneClock(sessionID)

	db.DB.Exec("UPDATE scene_timings SET status='completed', completed_at=datetime('now') WHERE session_id=? AND status='active'", sessionID)

	_, err := db.DB.Exec("UPDATE session_pacing SET status='completed', completed_at=datetime('now') WHERE id=? AND status IN ('running','paused')", sessionID)
	return err
}

func CompletePacingSession(c *gin.Context) {
	if err := completePacingSession(c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "completed"})
}

// errNoPacingActs is returned when a session cannot advance because the
// adventure has no acts at all.
var errNoPacingActs = errors.New("no acts in adventure")

// advancePacingSession moves the session to its next scene, act or completed
// state. It returns "advanced" or "completed".
func advancePacingSession(sessionID string) (string, error) {
	// Persist the time spent on the current scene before moving on. The
	// session keeps running, so both clocks re-anchor to now.
	accrueSessionClock(sessionID)
	accrueActiveSceneClock(sessionID)

	// Get current scene info
	var currentActID, currentSceneID int64
	err := db.DB.QueryRow("SELECT COALESCE(current_act_id,0), COALESCE(current_scene_id,0) FROM session_pacing WHERE id=?", sessionID).Scan(&currentActID, &currentSceneID)
	if err != nil {
		return "", err
	}

	if currentSceneID > 0 {
		// Mark current scene timing as completed
		db.DB.Exec("UPDATE scene_timings SET status='completed', completed_at=datetime('now') WHERE session_id=? AND scene_id=? AND status='active'",
			sessionID, currentSceneID)
	}

	// If no current act/scene, try to set first scene from adventure
	if currentActID == 0 && currentSceneID == 0 {
		var advID int64
		db.DB.QueryRow("SELECT adventure_id FROM session_pacing WHERE id=?", sessionID).Scan(&advID)
		var firstActID int64
		err = db.DB.QueryRow("SELECT id FROM oneshot_acts WHERE adventure_id=? ORDER BY number ASC LIMIT 1", advID).Scan(&firstActID)
		if err != nil {
			return "", errNoPacingActs
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
			return "advanced", nil
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
			return "advanced", nil
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
				return "advanced", nil
			}
			// Act has no scenes, advance to next act without scenes
			currentActID = nextActID
			db.DB.Exec("UPDATE session_pacing SET current_act_id=?, current_scene_id=NULL WHERE id=?", nextActID, sessionID)
		}
	}

	// No more acts - complete session
	db.DB.Exec("UPDATE session_pacing SET status='completed', completed_at=datetime('now') WHERE id=?", sessionID)
	return "completed", nil
}

func AdvanceToNextScene(c *gin.Context) {
	status, err := advancePacingSession(c.Param("id"))
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
			return
		}
		if err == errNoPacingActs {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if status == "completed" {
		c.JSON(http.StatusOK, gin.H{"status": "completed", "message": "all scenes completed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "advanced"})
}
