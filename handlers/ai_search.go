package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"villum/ai"
	"villum/db"
)

func HandleAISearch(c *gin.Context) {
	uid, ok := MustGetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	role, _ := c.Get("role")
	isAdmin := role == "admin"

	var req struct {
		Query      string `json:"query"`
		CampaignID *int64 `json:"campaign_id"`
		TypeFilter string `json:"type_filter"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	q := strings.TrimSpace(req.Query)
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query is required"})
		return
	}
	campaignID := int64(0)
	if req.CampaignID != nil {
		campaignID = *req.CampaignID
	}
	if campaignID > 0 && !isAdmin {
		allowed, _ := IsCampaignMember(c.Request.Context(), campaignID, uid)
		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "not a campaign member"})
			return
		}
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	res := ai.AnswerQuestion(ctx, db.DB, ai.QuestionRequest{
		Query:      q,
		CampaignID: campaignID,
		UserID:     uid,
		IsAdmin:    isAdmin,
		TypeFilter: req.TypeFilter,
		MaxTokens:  800,
		Timeout:    30 * time.Second,
	})

	// map to frontend expected shape: kind/id/title/subtitle/url/snippet
	type outSource struct {
		Kind     string `json:"kind"`
		ID       int64  `json:"id"`
		Title    string `json:"title"`
		Subtitle string `json:"subtitle,omitempty"`
		URL      string `json:"url,omitempty"`
		Snippet  string `json:"snippet,omitempty"`
	}
	out := make([]outSource, 0, len(res.Sources))
	for _, s := range res.Sources {
		out = append(out, outSource{
			Kind:     s.EntityType,
			ID:       s.EntityID,
			Title:    s.Title,
			Subtitle: s.Subtitle,
			URL:      s.URL,
			Snippet:  s.Snippet,
		})
	}
	if out == nil {
		out = []outSource{}
	}
	if res.Degraded {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": res.Message, "sources": out, "scope": res.Scope})
		return
	}
	c.JSON(http.StatusOK, gin.H{"answer": res.Answer, "sources": out, "scope": res.Scope})
}
