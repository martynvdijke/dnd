package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"villum/db"
)

// copilotExtractionSystemPrompt instructs exactly one JSON object with recap + entities.
func copilotExtractionSystemPrompt() string {
	// Include per-type field contracts from aiDraftEntities
	npcSchema := aiDraftEntities["npc"].Schema
	locSchema := aiDraftEntities["location"].Schema
	encSchema := aiDraftEntities["encounter"].Schema
	facSchema := aiDraftEntities["faction"].Schema
	return "You are a D&D session transcript extraction assistant.\n" +
		"Extract a recap and campaign entities from the transcript.\n\n" +
		"Reply with exactly one JSON object and nothing else (no markdown fences, no commentary) in this shape:\n" +
		`{"status":"ready","message":"...","draft":{"recap":{"title":"...","content":"..."},"entities":{"npcs":[...],"locations":[...],"encounters":[...],"factions":[...]}}}` + "\n\n" +
		"Recap: write key events, decisions, and unresolved threads in 250-350 words. Title should be concise.\n" +
		"Entities: extract distinct NPCs, locations, encounters, and factions introduced or developed in the session. Omit empty arrays.\n" +
		"Each entity must match its schema:\n" +
		"npc: " + npcSchema + "\n" +
		"location: " + locSchema + "\n" +
		"encounter: " + encSchema + "\n" +
		"faction: " + facSchema + "\n" +
		"Keep descriptions short (1-3 sentences). Ensure valid JSON with all objects and arrays closed, no trailing commas."
}

func parseCopilotExtractReply(text string) (string, string, json.RawMessage) {
	// Reuse tolerant envelope parsing: same as parseAIDraftReply but more permissive name
	return parseAIDraftReply(text)
}

func handleCopilotTranscriptExtract(c *gin.Context) {
	campaignID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	tid, _ := strconv.ParseInt(c.Param("tid"), 10, 64)
	uid, ok := MustGetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if !isCampaignMember(c, campaignID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	if !aiEnabled(c.Request.Context()) {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AI is not enabled"})
		return
	}
	// Optional endpoint_id from body
	var req struct {
		EndpointID int64 `json:"endpoint_id"`
	}
	_ = c.ShouldBindJSON(&req)

	var title, content string
	err := db.DB.QueryRow("SELECT title, content FROM campaign_wiki_pages WHERE id=? AND campaign_id=?", tid, campaignID).Scan(&title, &content)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "transcript not found"})
		return
	}
	if len(content) > 12000 {
		content = content[:12000]
	}
	endpointID := req.EndpointID
	if endpointID == 0 {
		eps, err := db.GetEnabledAIEndpointsByType(c.Request.Context(), "text")
		if err != nil || len(eps) == 0 {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "no AI endpoint configured"})
			return
		}
		endpointID = eps[0].ID
	}
	system := copilotExtractionSystemPrompt()
	maxTokens := 4000
	text, _, genErr := generateText(c.Request.Context(), endpointID, content, system, &maxTokens, "")
	if genErr != nil {
		lower := strings.ToLower(genErr.Error())
		isTruncation := strings.Contains(lower, "length") || strings.Contains(lower, "truncat") || strings.Contains(lower, "token limit")
		if !isTruncation || strings.TrimSpace(text) == "" {
			if ae, ok := genErr.(*aiGenError); ok {
				c.JSON(ae.Status, gin.H{"error": ae.Msg})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": genErr.Error()})
			return
		}
		// truncation but we have text: attempt to parse below
	}

	status, _, draftRaw := parseCopilotExtractReply(text)
	_ = status

	// draftRaw may be stringified; already normalized in parseAIDraftReply
	if draftRaw == nil || strings.TrimSpace(string(draftRaw)) == "" || strings.TrimSpace(string(draftRaw)) == "null" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "could not extract a recap or entities", "message": "empty draft"})
		return
	}

	var envelope struct {
		Recap struct {
			Title   string `json:"title"`
			Content string `json:"content"`
		} `json:"recap"`
		Entities struct {
			NPCs       []json.RawMessage `json:"npcs"`
			Locations  []json.RawMessage `json:"locations"`
			Encounters []json.RawMessage `json:"encounters"`
			Factions   []json.RawMessage `json:"factions"`
		} `json:"entities"`
	}
	// draftRaw is the "draft" field value; it should contain recap+entities
	// But the envelope is {status,message,draft:{recap,entities}}
	// parseAIDraftReply already extracted draftRaw, so we parse it as the inner object.
	if err := json.Unmarshal(draftRaw, &envelope); err != nil {
		// Try parsing the whole text salvaged object if draftRaw was not inner but outer
		// Attempt to unmarshal draftRaw as the outer envelope fallback
		var outer struct {
			Draft json.RawMessage `json:"draft"`
		}
		if err2 := json.Unmarshal(draftRaw, &outer); err2 == nil && len(outer.Draft) > 0 {
			_ = json.Unmarshal(outer.Draft, &envelope)
		}
		if envelope.Recap.Title == "" && envelope.Recap.Content == "" && len(envelope.Entities.NPCs) == 0 && len(envelope.Entities.Locations) == 0 && len(envelope.Entities.Encounters) == 0 && len(envelope.Entities.Factions) == 0 {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "could not extract a recap or entities", "message": "unparseable reply"})
			return
		}
	}

	hasRecap := strings.TrimSpace(envelope.Recap.Title) != "" || strings.TrimSpace(envelope.Recap.Content) != ""
	hasEntities := len(envelope.Entities.NPCs) > 0 || len(envelope.Entities.Locations) > 0 || len(envelope.Entities.Encounters) > 0 || len(envelope.Entities.Factions) > 0
	if !hasRecap && !hasEntities {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "could not extract a recap or entities", "message": "no recap or entities found"})
		return
	}

	// For better error detail on unparseable: if both empty after lenient check
	if !hasRecap && !hasEntities {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "could not extract a recap or entities", "message": "empty extraction"})
		return
	}

	now := time.Now().UTC().Format("2006-01-02 15:04:05")
	type draftInfo struct {
		ID         string `json:"id"`
		EntityType string `json:"entity_type"`
		Name       string `json:"name"`
	}
	var drafts []draftInfo

	extractName := func(raw json.RawMessage) string {
		var m map[string]any
		if err := json.Unmarshal(raw, &m); err != nil {
			return ""
		}
		if n, ok := m["name"].(string); ok {
			return strings.TrimSpace(n)
		}
		if t, ok := m["title"].(string); ok {
			return strings.TrimSpace(t)
		}
		return ""
	}

	persist := func(et string, raws []json.RawMessage) {
		for _, raw := range raws {
			name := extractName(raw)
			if name == "" {
				continue
			}
			id := generateSessionID()
			s := &aiDraftSession{
				ID:         id,
				UserID:     uid,
				EntityType: et,
				CampaignID: &campaignID,
				Status:     "ready",
				Draft:      raw,
				Messages: []aiDraftMessage{
					{Role: "system", Content: aiDraftSystemPrompt(et)},
					{Role: "user", Content: "Extracted from session transcript: " + title},
				},
				CreatedAt: now,
				UpdatedAt: now,
			}
			if err := saveAIDraftSession(c.Request.Context(), s); err != nil {
				continue
			}
			drafts = append(drafts, draftInfo{ID: id, EntityType: et, Name: name})
		}
	}

	persist("npc", envelope.Entities.NPCs)
	persist("location", envelope.Entities.Locations)
	persist("encounter", envelope.Entities.Encounters)
	persist("faction", envelope.Entities.Factions)

	if drafts == nil {
		drafts = []draftInfo{}
	}

	// If we had entities in the reply but all were skipped (no name), treat as empty unless recap present
	if !hasRecap && len(drafts) == 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "could not extract a recap or entities", "message": "no valid entities found"})
		return
	}

	recapTitle := envelope.Recap.Title
	recapContent := envelope.Recap.Content

	c.JSON(http.StatusOK, gin.H{
		"recap":  gin.H{"title": recapTitle, "content": recapContent},
		"drafts": drafts,
		"source": gin.H{"entity_type": "wiki", "entity_id": tid, "title": title},
	})
}
