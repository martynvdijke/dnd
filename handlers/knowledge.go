package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"villum/db"
)

var validStatuses = map[string]bool{"rumor": true, "confirmed": true, "revealed": true, "false": true}

type Knowledge struct {
	ID            int64  `json:"id"`
	CampaignID    int64  `json:"campaign_id"`
	Title         string `json:"title"`
	Content       string `json:"content"`
	Source        string `json:"source"`
	Status        string `json:"status"`
	Shared        bool   `json:"shared"`
	StatusHistory string `json:"status_history"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

func knowledgeVisible(campaignID, userID int64, role string) bool {
	if role == "admin" {
		return true
	}
	var ownerID int64
	err := db.DB.QueryRow(`SELECT user_id FROM campaigns WHERE id=?`, campaignID).Scan(&ownerID)
	if err != nil {
		return false
	}
	if ownerID == userID {
		return true
	}
	var cnt int
	db.DB.QueryRow(`SELECT COUNT(*) FROM campaign_members WHERE campaign_id=? AND user_id=?`, campaignID, userID).Scan(&cnt)
	return cnt > 0
}

func isOwner(campaignID, userID int64, role string) bool {
	if role == "admin" {
		return true
	}
	var ownerID int64
	if err := db.DB.QueryRow(`SELECT user_id FROM campaigns WHERE id=?`, campaignID).Scan(&ownerID); err != nil {
		return false
	}
	if ownerID == userID {
		return true
	}
	var cnt int
	db.DB.QueryRow(`SELECT COUNT(*) FROM campaign_members WHERE campaign_id=? AND user_id=? AND role='dm'`, campaignID, userID).Scan(&cnt)
	return cnt > 0
}

func knowledgeRowToStruct(id int64) (*Knowledge, error) {
	var k Knowledge
	var shared int
	var sh string
	err := db.DB.QueryRow(`SELECT id,campaign_id,title,content,source,status,shared,status_history,created_at,updated_at FROM campaign_knowledge WHERE id=?`, id).Scan(&k.ID, &k.CampaignID, &k.Title, &k.Content, &k.Source, &k.Status, &shared, &sh, &k.CreatedAt, &k.UpdatedAt)
	if err != nil {
		return nil, err
	}
	k.Shared = shared != 0
	k.StatusHistory = sh
	return &k, nil
}

func ListKnowledge(c *gin.Context) {
	campaignID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := c.Get("user_id")
	uid := userID.(int64)
	role, _ := c.Get("role")
	r, _ := role.(string)
	if !knowledgeVisible(campaignID, uid, r) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	owner := isOwner(campaignID, uid, r)
	filter := c.Query("status")
	query := `SELECT id,campaign_id,title,content,source,status,shared,status_history,created_at,updated_at FROM campaign_knowledge WHERE campaign_id=?`
	args := []any{campaignID}
	if filter != "" {
		query += ` AND status=?`
		args = append(args, filter)
	}
	if !owner {
		query += ` AND shared=1`
	}
	query += ` ORDER BY created_at DESC`
	rows, err := db.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := []Knowledge{}
	for rows.Next() {
		var k Knowledge
		var shared int
		var sh string
		rows.Scan(&k.ID, &k.CampaignID, &k.Title, &k.Content, &k.Source, &k.Status, &shared, &sh, &k.CreatedAt, &k.UpdatedAt)
		k.Shared = shared != 0
		k.StatusHistory = sh
		out = append(out, k)
	}
	c.JSON(http.StatusOK, out)
}

func CreateKnowledge(c *gin.Context) {
	campaignID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := c.Get("user_id")
	uid := userID.(int64)
	role, _ := c.Get("role")
	r, _ := role.(string)
	if !isOwner(campaignID, uid, r) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var body struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		Source  string `json:"source"`
		Status  string `json:"status"`
		Shared  *bool  `json:"shared"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	if strings.TrimSpace(body.Title) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title required"})
		return
	}
	status := body.Status
	if status == "" {
		status = "rumor"
	}
	if !validStatuses[status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
		return
	}
	shared := 0
	if body.Shared != nil && *body.Shared {
		shared = 1
	}
	now := time.Now().UTC().Format(time.RFC3339)
	hist, _ := json.Marshal([]map[string]string{{"status": status, "at": now}})
	res, err := db.DB.Exec(`INSERT INTO campaign_knowledge(campaign_id,title,content,source,status,shared,status_history,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?)`, campaignID, body.Title, body.Content, body.Source, status, shared, string(hist), now, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := res.LastInsertId()
	k, _ := knowledgeRowToStruct(id)
	c.JSON(http.StatusCreated, k)
}

func GetKnowledge(c *gin.Context) {
	kid, _ := strconv.ParseInt(c.Param("kid"), 10, 64)
	userID, _ := c.Get("user_id")
	uid := userID.(int64)
	role, _ := c.Get("role")
	r, _ := role.(string)
	k, err := knowledgeRowToStruct(kid)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	owner := isOwner(k.CampaignID, uid, r)
	if !k.Shared && !owner {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, k)
}

func UpdateKnowledge(c *gin.Context) {
	kid, _ := strconv.ParseInt(c.Param("kid"), 10, 64)
	userID, _ := c.Get("user_id")
	uid := userID.(int64)
	role, _ := c.Get("role")
	r, _ := role.(string)
	k, err := knowledgeRowToStruct(kid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if !isOwner(k.CampaignID, uid, r) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var body struct {
		Title   *string `json:"title"`
		Content *string `json:"content"`
		Source  *string `json:"source"`
		Status  *string `json:"status"`
		Shared  *bool   `json:"shared"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	if body.Status != nil && !validStatuses[*body.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
		return
	}
	// apply updates
	if body.Title != nil {
		k.Title = *body.Title
	}
	if body.Content != nil {
		k.Content = *body.Content
	}
	if body.Source != nil {
		k.Source = *body.Source
	}
	newHist := k.StatusHistory
	if body.Status != nil && *body.Status != k.Status {
		var hist []map[string]string
		_ = json.Unmarshal([]byte(k.StatusHistory), &hist)
		if hist == nil {
			hist = []map[string]string{}
		}
		hist = append(hist, map[string]string{"status": *body.Status, "at": time.Now().UTC().Format(time.RFC3339)})
		b, _ := json.Marshal(hist)
		newHist = string(b)
		k.Status = *body.Status
	}
	if body.Shared != nil {
		k.Shared = *body.Shared
	}
	sharedInt := 0
	if k.Shared {
		sharedInt = 1
	}
	now := time.Now().UTC().Format(time.RFC3339)
	_, err = db.DB.Exec(`UPDATE campaign_knowledge SET title=?,content=?,source=?,status=?,shared=?,status_history=?,updated_at=? WHERE id=?`, k.Title, k.Content, k.Source, k.Status, sharedInt, newHist, now, kid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	k2, _ := knowledgeRowToStruct(kid)
	c.JSON(http.StatusOK, k2)
}

func DeleteKnowledge(c *gin.Context) {
	kid, _ := strconv.ParseInt(c.Param("kid"), 10, 64)
	userID, _ := c.Get("user_id")
	uid := userID.(int64)
	role, _ := c.Get("role")
	r, _ := role.(string)
	k, err := knowledgeRowToStruct(kid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if !isOwner(k.CampaignID, uid, r) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	db.DB.Exec(`DELETE FROM campaign_knowledge WHERE id=?`, kid)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func AddKnowledgeKnownBy(c *gin.Context) {
	kid, _ := strconv.ParseInt(c.Param("kid"), 10, 64)
	userID, _ := c.Get("user_id")
	uid := userID.(int64)
	role, _ := c.Get("role")
	r, _ := role.(string)
	k, err := knowledgeRowToStruct(kid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if !isOwner(k.CampaignID, uid, r) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var body struct {
		CharacterID int64 `json:"character_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.CharacterID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "character_id required"})
		return
	}
	now := time.Now().UTC().Format(time.RFC3339)
	_, err = db.DB.Exec(`INSERT OR IGNORE INTO campaign_knowledge_known_by(knowledge_id,character_id,created_at) VALUES(?,?,?)`, kid, body.CharacterID, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func RemoveKnowledgeKnownBy(c *gin.Context) {
	kid, _ := strconv.ParseInt(c.Param("kid"), 10, 64)
	cid, _ := strconv.ParseInt(c.Param("cid"), 10, 64)
	userID, _ := c.Get("user_id")
	uid := userID.(int64)
	role, _ := c.Get("role")
	r, _ := role.(string)
	k, err := knowledgeRowToStruct(kid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if !isOwner(k.CampaignID, uid, r) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	db.DB.Exec(`DELETE FROM campaign_knowledge_known_by WHERE knowledge_id=? AND character_id=?`, kid, cid)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func ListKnowledgeKnownBy(c *gin.Context) {
	kid, _ := strconv.ParseInt(c.Param("kid"), 10, 64)
	k, err := knowledgeRowToStruct(kid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	userID, _ := c.Get("user_id")
	uid := userID.(int64)
	role, _ := c.Get("role")
	r, _ := role.(string)
	owner := isOwner(k.CampaignID, uid, r)
	if !k.Shared && !owner {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	rows, _ := db.DB.Query(`SELECT character_id FROM campaign_knowledge_known_by WHERE knowledge_id=?`, kid)
	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()
	ids := []int64{}
	if rows != nil {
		for rows.Next() {
			var id int64
			rows.Scan(&id)
			ids = append(ids, id)
		}
	}
	if ids == nil {
		ids = []int64{}
	}
	c.JSON(http.StatusOK, ids)
}

func BulkRevealKnowledge(c *gin.Context) {
	kid, _ := strconv.ParseInt(c.Param("kid"), 10, 64)
	userID, _ := c.Get("user_id")
	uid := userID.(int64)
	role, _ := c.Get("role")
	r, _ := role.(string)
	k, err := knowledgeRowToStruct(kid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if !isOwner(k.CampaignID, uid, r) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	// get all party characters (campaign characters type player)
	rows, _ := db.DB.Query(`SELECT id FROM characters WHERE campaign_id=?`, k.CampaignID)
	var pids []int64
	if rows != nil {
		for rows.Next() {
			var id int64
			rows.Scan(&id)
			pids = append(pids, id)
		}
		rows.Close()
	}
	now := time.Now().UTC().Format(time.RFC3339)
	for _, pid := range pids {
		db.DB.Exec(`INSERT OR IGNORE INTO campaign_knowledge_known_by(knowledge_id,character_id,created_at) VALUES(?,?,?)`, kid, pid, now)
	}
	// set shared true and status revealed if not already? spec says shared=true, keep status? We'll set shared true
	sharedHist := k.StatusHistory
	// if status not revealed, transition to revealed and add history
	newStatus := k.Status
	if k.Status != "revealed" {
		var hist []map[string]string
		_ = json.Unmarshal([]byte(k.StatusHistory), &hist)
		hist = append(hist, map[string]string{"status": "revealed", "at": now})
		b, _ := json.Marshal(hist)
		sharedHist = string(b)
		newStatus = "revealed"
	}
	db.DB.Exec(`UPDATE campaign_knowledge SET shared=1,status=?,status_history=?,updated_at=? WHERE id=?`, newStatus, sharedHist, now, kid)
	k2, _ := knowledgeRowToStruct(kid)
	c.JSON(http.StatusOK, k2)
}
