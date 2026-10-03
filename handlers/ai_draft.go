package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/ent/location"
	"villum/ent/npc"
	"villum/middleware"
)

// Draft replies are large (an adventure with acts, scenes and NPCs) and
// reasoning models spend part of the budget on hidden reasoning, so the budget
// comes from the endpoint when configured and falls back to a generous default.
const (
	defaultAIDraftMaxTokens = 16000
	minAIDraftMaxTokens     = 4000
	maxAIDraftMaxTokens     = 32000
)

// aiDraftTokenBudget resolves the max_tokens value for draft replies: the
// endpoint's configured value clamped to a safe range, or the default.
func aiDraftTokenBudget(ctx context.Context, endpointID int64) *int {
	budget := defaultAIDraftMaxTokens
	if ep, err := db.GetAIEndpoint(ctx, endpointID); err == nil && ep.MaxTokens != nil {
		budget = *ep.MaxTokens
		if budget < minAIDraftMaxTokens {
			budget = minAIDraftMaxTokens
		}
		if budget > maxAIDraftMaxTokens {
			budget = maxAIDraftMaxTokens
		}
	}
	return &budget
}

// aiDraftMessage is one turn in a drafting conversation as persisted.
type aiDraftMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// aiDraftSession is a persisted, resumable AI generation conversation.
type aiDraftSession struct {
	ID         string
	UserID     int64
	EntityType string
	CampaignID *int64
	ParentID   *int64
	Status     string
	Messages   []aiDraftMessage
	Draft      json.RawMessage
	CreatedAt  string
	UpdatedAt  string
}

// aiDraftEntities maps each supported entity type to its display label and the
// JSON shape the assistant must produce when the user approves the draft.
var aiDraftEntities = map[string]struct {
	Label  string
	Schema string
}{
	"oneshot": {
		Label: "one-shot adventure",
		Schema: `{"title":"string","premise":"string","hook":"string",` +
			`"difficulty":"easy|medium|hard|deadly","estimated_minutes":180,"notes":"string",` +
			`"acts":[{"title":"string","description":"string","estimated_minutes":30,` +
			`"scenes":[{"title":"string","description":"string","scene_type":"roleplay|exploration|combat|puzzle","estimated_minutes":15}]}],` +
			`"npcs":[{"name":"string","race":"string","description":"string","role":"string"}],` +
			`"locations":[{"name":"string","type":"string","description":"string"}],` +
			`"encounters":[{"name":"string","description":"string","difficulty":"easy|medium|hard|deadly"}],` +
			`"clues":[{"title":"string","description":"string","clue_type":"direct|witness|object|location"}]}`,
	},
	"npc": {
		Label: "NPC",
		Schema: `{"name":"string","race":"string","class":"string",` +
			`"description":"string","notes":"string"}`,
	},
	"location": {
		Label:  "location",
		Schema: `{"name":"string","type":"string","description":"string"}`,
	},
	"encounter": {
		Label: "encounter",
		Schema: `{"name":"string","description":"string",` +
			`"environment":"string","difficulty":"easy|medium|hard|deadly","notes":"string"}`,
	},
	"faction": {
		Label: "faction",
		Schema: `{"name":"string","description":"string",` +
			`"type":"string","headquarters":"string"}`,
	},
	"campaign": {
		Label: "campaign",
		Schema: `{"name":"string","party_name":"string",` +
			`"description":"string","dm_notes":"string"}`,
	},
	"quest": {
		Label:  "quest",
		Schema: `{"name":"string","description":"string","objectives":"string","rewards":"string","notes":"string"}`,
	},
	"item": {
		Label: "item",
		Schema: `{"name":"string","description":"string","category":"gear",` +
			`"quantity":1,"weight":0,"price_gp":0,"is_magical":false,"attunement":false,"notes":"string"}`,
	},
}

// aiDraftEntitySupported reports whether the entity type can be drafted.
func aiDraftEntitySupported(t string) bool {
	_, ok := aiDraftEntities[t]
	return ok
}

// aiDraftDefaultPrompt seeds a new conversation when the user has no opening
// message.
func aiDraftDefaultPrompt(entityType string) string {
	label := aiDraftEntities[entityType].Label
	return "Help me create a new " + label + ". Ask me what you need to know and suggest ideas."
}

// aiDraftSystemPrompt builds the assistant instructions for an entity type.
func aiDraftSystemPrompt(entityType string) string {
	def := aiDraftEntities[entityType]
	prompt := "You are an expert D&D 5e (2024 rules) game master assistant embedded in the villum campaign manager.\n" +
		"You are helping the user design a new " + def.Label + ".\n\n" +
		"Work conversationally and offer concrete, opinionated suggestions. When the user clearly asks " +
		"you to draft or generate something (or names an existing adventure to adapt), produce the " +
		"complete draft immediately with status \"ready\". Ask at most two focused clarifying questions, " +
		"and only when the request is genuinely ambiguous.\n\n" +
		"Reply with a single JSON object and nothing else, in exactly this shape:\n" +
		`{"status":"chatting","message":"<reply shown to the user>","draft":null}` + "\n" +
		"Use status \"chatting\" with draft=null while you are still asking questions or revising.\n" +
		"When the user approves or asks you to generate, use status \"ready\" and put the complete " +
		"object in draft.\n" +
		"The draft must match this JSON shape:\n" + def.Schema + "\n" +
		"Do not include markdown fences or commentary outside the JSON object. " +
		"Keep descriptions short (one to three sentences) and never truncate the JSON: a reply cut off mid-object is unusable."
	if entityType == "oneshot" {
		prompt += "\n\n" + aiOneShotDraftSkill
	}
	return prompt
}

// aiOneShotDraftSkill is the detailed output contract for one-shot drafts. It
// teaches the model how to map an idea (including a named published adventure)
// onto the schema within a size budget that fits a single reply.
const aiOneShotDraftSkill = `One-shot drafting contract:
- Map the user's premise onto the schema faithfully. If they name a published adventure (for example "The Wolves of Welton"), adapt its plot, characters, locations and set pieces into this schema rather than inventing a different story.
- Fields: "title" is the adventure name. "premise" is 2-4 sentences of situation and stakes. "hook" is how the party gets involved. "difficulty" is one of easy|medium|hard|deadly. "estimated_minutes" is the whole session length. "notes" holds concise DM guidance (at most about 1200 characters).
- Acts: 3-5 entries, each with "title", "description" (1-3 sentences), "estimated_minutes" and 2-4 "scenes". Each scene has "title", "description", "scene_type" (one of roleplay|exploration|combat|puzzle) and "estimated_minutes".
- "npcs": 4-8 entries with "name", "race", "description", "role".
- "locations": 4-8 entries with "name", "type", "description".
- "encounters": 3-6 entries with "name", "description", "difficulty".
- "clues": 3-8 entries with "title", "description", "clue_type" (one of direct|witness|object|location).
- Stay within these limits and write compactly so the entire JSON object fits the token budget. Every array and object must be closed, with no trailing commas and no comments.`

// toProviderMessages converts persisted turns into the provider message format.
func toProviderMessages(msgs []aiDraftMessage) []map[string]string {
	out := make([]map[string]string, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, map[string]string{"role": m.Role, "content": m.Content})
	}
	return out
}

// parseAIDraftReply extracts the status, user-facing message and draft from a
// model reply, tolerating markdown fences and stray prose.
func parseAIDraftReply(text string) (string, string, json.RawMessage) {
	var parsed struct {
		Status  string          `json:"status"`
		Message string          `json:"message"`
		Draft   json.RawMessage `json:"draft"`
	}
	// A draft may arrive as a JSON string rather than an object.
	normalize := func(raw json.RawMessage) json.RawMessage {
		trimmed := strings.TrimSpace(string(raw))
		if len(trimmed) == 0 || trimmed == "null" {
			return nil
		}
		if strings.HasPrefix(trimmed, `"`) {
			var s string
			if err := json.Unmarshal(raw, &s); err == nil {
				return json.RawMessage(s)
			}
		}
		return raw
	}

	tryParse := func(s string) bool {
		if err := json.Unmarshal([]byte(s), &parsed); err != nil {
			return false
		}
		if parsed.Status != "chatting" && parsed.Status != "ready" {
			return false
		}
		parsed.Draft = normalize(parsed.Draft)
		return true
	}

	cleaned := strings.TrimSpace(text)
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

	if tryParse(cleaned) {
		msg := parsed.Message
		if msg == "" && parsed.Status == "ready" {
			msg = "Here's the draft I put together."
		}
		return parsed.Status, msg, parsed.Draft
	}
	// Salvage an embedded object if the model wrapped it in prose.
	if i := strings.Index(cleaned, "{"); i >= 0 {
		if j := strings.LastIndex(cleaned, "}"); j > i {
			if tryParse(cleaned[i : j+1]) {
				msg := parsed.Message
				if msg == "" && parsed.Status == "ready" {
					msg = "Here's the draft I put together."
				}
				return parsed.Status, msg, parsed.Draft
			}
		}
	}
	return "chatting", strings.TrimSpace(text), nil
}

// ─── Persistence ───

func saveAIDraftSession(ctx context.Context, s *aiDraftSession) error {
	msgsJSON, err := json.Marshal(s.Messages)
	if err != nil {
		return err
	}
	draft := ""
	if len(s.Draft) > 0 {
		draft = string(s.Draft)
	}
	now := time.Now().UTC().Format("2006-01-02 15:04:05")
	_, err = db.DB.Exec(`INSERT INTO ai_draft_sessions
		(id, user_id, entity_type, campaign_id, parent_id, status, messages, draft, created_at, updated_at)
		VALUES(?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(id) DO UPDATE SET
			status=excluded.status, messages=excluded.messages, draft=excluded.draft, updated_at=excluded.updated_at`,
		s.ID, s.UserID, s.EntityType, s.CampaignID, s.ParentID, s.Status, string(msgsJSON), draft, now, now)
	return err
}

func loadAIDraftSession(id string) (*aiDraftSession, error) {
	var s aiDraftSession
	var msgs, draft string
	err := db.DB.QueryRow(`SELECT id, user_id, entity_type, campaign_id, parent_id, status, messages, draft, created_at, updated_at
		FROM ai_draft_sessions WHERE id=?`, id).
		Scan(&s.ID, &s.UserID, &s.EntityType, &s.CampaignID, &s.ParentID, &s.Status, &msgs, &draft, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(msgs), &s.Messages); err != nil {
		s.Messages = []aiDraftMessage{}
	}
	if strings.TrimSpace(draft) != "" {
		s.Draft = json.RawMessage(draft)
	}
	return &s, nil
}

// requireOwnedAIDraft loads a session and enforces ownership (admins bypass).
func requireOwnedAIDraft(c *gin.Context) (*aiDraftSession, bool) {
	id := c.Param("id")
	s, err := loadAIDraftSession(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "draft session not found"})
		return nil, false
	}
	uid, _ := MustGetUserID(c)
	role, _ := c.Get("role")
	if s.UserID != uid && role != "admin" {
		c.JSON(http.StatusNotFound, gin.H{"error": "draft session not found"})
		return nil, false
	}
	return s, true
}

// aiDraftView is the JSON shape returned to the client.
func aiDraftView(s *aiDraftSession, message string) gin.H {
	draft := json.RawMessage(nil)
	if len(s.Draft) > 0 {
		draft = s.Draft
	}
	return gin.H{
		"id":          s.ID,
		"entity_type": s.EntityType,
		"status":      s.Status,
		"message":     message,
		"draft":       draft,
		"messages":    s.Messages,
	}
}

func writeAIGenError(c *gin.Context, err error) {
	if ae, ok := err.(*aiGenError); ok {
		c.JSON(ae.Status, gin.H{"error": ae.Msg})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

// validateAIDraftReply turns an empty or token-truncated provider reply into an
// actionable error instead of persisting a broken assistant turn. Reasoning
// models can spend the whole max_tokens budget on hidden reasoning, which
// surfaces as an empty content string with finish_reason "length".
func validateAIDraftReply(reply, finishReason string) error {
	if strings.TrimSpace(reply) == "" {
		return &aiGenError{Status: http.StatusBadGateway, Msg: "the AI provider returned an empty reply (the model may have spent its whole token budget on reasoning); raise max_tokens for this endpoint or ask for a smaller draft"}
	}
	if finishReason == "length" {
		return &aiGenError{Status: http.StatusBadGateway, Msg: "the AI reply was cut off because it hit the token limit; raise max_tokens for this endpoint or ask for a smaller draft"}
	}
	return nil
}

// ─── Handlers ───

// StartAIDraft begins a drafting conversation: POST /api/ai/draft
func StartAIDraft(c *gin.Context) {
	if !aiEnabled(c.Request.Context()) {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AI features are disabled"})
		return
	}
	var req struct {
		EndpointID int64  `json:"endpoint_id"`
		EntityType string `json:"entity_type"`
		CampaignID *int64 `json:"campaign_id"`
		ParentID   *int64 `json:"parent_id"`
		Message    string `json:"message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	entityType := strings.TrimSpace(req.EntityType)
	if !aiDraftEntitySupported(entityType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported entity_type"})
		return
	}
	uid, ok := MustGetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if req.CampaignID != nil && !isCampaignMember(c, *req.CampaignID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	message := strings.TrimSpace(req.Message)
	if message == "" {
		message = aiDraftDefaultPrompt(entityType)
	}

	basePrompt := aiDraftSystemPrompt(entityType)
	systemPrompt := basePrompt
	if req.CampaignID != nil && *req.CampaignID > 0 {
		role, _ := c.Get("role")
		isAdmin := role == "admin"
		lore := aiDraftLoreContext(c.Request.Context(), *req.CampaignID, uid, isAdmin, message)
		if lore != "" {
			systemPrompt = basePrompt + "\n\n" + aiDraftLoreAddendum(lore)
		}
	}
	s := &aiDraftSession{
		ID:         generateSessionID(),
		UserID:     uid,
		EntityType: entityType,
		CampaignID: req.CampaignID,
		ParentID:   req.ParentID,
		Status:     "drafting",
		Messages: []aiDraftMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: message},
		},
	}
	reply, finishReason, err := generateChat(c.Request.Context(), req.EndpointID, toProviderMessages(s.Messages),
		aiDraftTokenBudget(c.Request.Context(), req.EndpointID), "", aiDraftTimeout)
	if err != nil {
		writeAIGenError(c, err)
		return
	}
	if err := validateAIDraftReply(reply, finishReason); err != nil {
		writeAIGenError(c, err)
		return
	}
	s.Status, message, s.Draft = parseAIDraftReply(reply)
	s.Messages = append(s.Messages, aiDraftMessage{Role: "assistant", Content: reply})
	if err := saveAIDraftSession(c.Request.Context(), s); err != nil {
		middleware.LogError("ai", "failed to save draft session", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save draft"})
		return
	}
	c.JSON(http.StatusCreated, aiDraftView(s, message))
}

// AIDraftTurn continues a drafting conversation: POST /api/ai/draft/:id/turn
func AIDraftTurn(c *gin.Context) {
	if !aiEnabled(c.Request.Context()) {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AI features are disabled"})
		return
	}
	s, ok := requireOwnedAIDraft(c)
	if !ok {
		return
	}
	var req struct {
		EndpointID int64  `json:"endpoint_id"`
		Message    string `json:"message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	message := strings.TrimSpace(req.Message)
	if message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message is required"})
		return
	}
	s.Messages = append(s.Messages, aiDraftMessage{Role: "user", Content: message})
	// Refresh campaign context on every turn.
	if s.CampaignID != nil && *s.CampaignID > 0 && len(s.Messages) > 0 && s.Messages[0].Role == "system" {
		base := aiDraftSystemPrompt(s.EntityType)
		role, _ := c.Get("role")
		isAdmin := role == "admin"
		lore := aiDraftLoreContext(c.Request.Context(), *s.CampaignID, s.UserID, isAdmin, message)
		if lore != "" {
			s.Messages[0].Content = base + "\n\n" + aiDraftLoreAddendum(lore)
		} else {
			s.Messages[0].Content = base
		}
	}
	reply, finishReason, err := generateChat(c.Request.Context(), req.EndpointID, toProviderMessages(s.Messages),
		aiDraftTokenBudget(c.Request.Context(), req.EndpointID), "", aiDraftTimeout)
	if err != nil {
		// Roll back the user turn so a failed request can be retried cleanly.
		s.Messages = s.Messages[:len(s.Messages)-1]
		writeAIGenError(c, err)
		return
	}
	if err := validateAIDraftReply(reply, finishReason); err != nil {
		// A broken reply must not be persisted; keep the conversation retryable.
		s.Messages = s.Messages[:len(s.Messages)-1]
		writeAIGenError(c, err)
		return
	}
	var replyMessage string
	s.Status, replyMessage, s.Draft = parseAIDraftReply(reply)
	s.Messages = append(s.Messages, aiDraftMessage{Role: "assistant", Content: reply})
	if err := saveAIDraftSession(c.Request.Context(), s); err != nil {
		middleware.LogError("ai", "failed to save draft session", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save draft"})
		return
	}
	c.JSON(http.StatusOK, aiDraftView(s, replyMessage))
}

// GetAIDraft reconnects to a session: GET /api/ai/draft/:id
func GetAIDraft(c *gin.Context) {
	s, ok := requireOwnedAIDraft(c)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, aiDraftView(s, ""))
}

// ListAIDrafts lists the caller's sessions: GET /api/ai/draft
func ListAIDrafts(c *gin.Context) {
	uid, ok := MustGetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	entityType := strings.TrimSpace(c.Query("entity_type"))
	query := "SELECT id, entity_type, status, updated_at FROM ai_draft_sessions WHERE user_id=?"
	args := []any{uid}
	if entityType != "" {
		query += " AND entity_type=?"
		args = append(args, entityType)
	}
	query += " ORDER BY updated_at DESC LIMIT 50"
	rows, err := db.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list drafts"})
		return
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var id, et, status, updated string
		if err := rows.Scan(&id, &et, &status, &updated); err != nil {
			continue
		}
		out = append(out, gin.H{"id": id, "entity_type": et, "status": status, "updated_at": updated})
	}
	c.JSON(http.StatusOK, out)
}

// DiscardAIDraft deletes a session: DELETE /api/ai/draft/:id
func DiscardAIDraft(c *gin.Context) {
	s, ok := requireOwnedAIDraft(c)
	if !ok {
		return
	}
	if _, err := db.DB.Exec("DELETE FROM ai_draft_sessions WHERE id=?", s.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to discard draft"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// CommitAIDraft creates the entity from the approved draft and returns its id.
// POST /api/ai/draft/:id/commit
func CommitAIDraft(c *gin.Context) {
	s, ok := requireOwnedAIDraft(c)
	if !ok {
		return
	}
	if s.Status != "ready" || len(s.Draft) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "the draft is not ready to save yet"})
		return
	}
	entityID, url, err := commitAIDraft(c, s)
	if err != nil {
		middleware.LogError("ai", "failed to commit draft", "entity_type", s.EntityType, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create %s: %s", s.EntityType, err.Error())})
		return
	}
	s.Status = "committed"
	if err := saveAIDraftSession(c.Request.Context(), s); err != nil {
		middleware.LogWarn("ai", "failed to mark draft committed", "error", err)
	}
	c.JSON(http.StatusCreated, gin.H{
		"id":          s.ID,
		"entity_type": s.EntityType,
		"entity_id":   entityID,
		"url":         url,
	})
}

// ─── Entity writers ───

// commitAIDraft dispatches to the writer for the session's entity type.
func commitAIDraft(c *gin.Context, s *aiDraftSession) (int64, string, error) {
	uid, _ := MustGetUserID(c)
	ctx := c.Request.Context()
	switch s.EntityType {
	case "oneshot":
		return commitAIOneShot(ctx, uid, s)
	case "npc":
		return commitAINPC(ctx, uid, s)
	case "location":
		return commitAILocation(ctx, uid, s)
	case "encounter":
		return commitAIEncounter(ctx, uid, s)
	case "faction":
		return commitAIFaction(ctx, uid, s)
	case "campaign":
		return commitAICampaign(ctx, uid, s)
	case "quest":
		return commitAIQuest(c, s)
	case "item":
		return commitAIItem(ctx, s)
	}
	return 0, "", fmt.Errorf("unsupported entity_type %q", s.EntityType)
}

type aiOneShotScene struct {
	Title            string `json:"title"`
	Description      string `json:"description"`
	SceneType        string `json:"scene_type"`
	EstimatedMinutes int    `json:"estimated_minutes"`
}

type aiOneShotAct struct {
	Title            string           `json:"title"`
	Description      string           `json:"description"`
	EstimatedMinutes int              `json:"estimated_minutes"`
	Scenes           []aiOneShotScene `json:"scenes"`
}

type aiOneShotNPC struct {
	Name        string `json:"name"`
	Race        string `json:"race"`
	Description string `json:"description"`
	Role        string `json:"role"`
}

type aiOneShotLocation struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type aiOneShotEncounter struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Difficulty  string `json:"difficulty"`
}

type aiOneShotClue struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	ClueType    string `json:"clue_type"`
}

type aiOneShotDraft struct {
	Title            string               `json:"title"`
	Premise          string               `json:"premise"`
	Hook             string               `json:"hook"`
	Difficulty       string               `json:"difficulty"`
	EstimatedMinutes int                  `json:"estimated_minutes"`
	Notes            string               `json:"notes"`
	Acts             []aiOneShotAct       `json:"acts"`
	NPCs             []aiOneShotNPC       `json:"npcs"`
	Locations        []aiOneShotLocation  `json:"locations"`
	Encounters       []aiOneShotEncounter `json:"encounters"`
	Clues            []aiOneShotClue      `json:"clues"`
}

var aiSceneTypes = map[string]bool{"roleplay": true, "exploration": true, "combat": true, "puzzle": true}
var aiDifficulties = map[string]bool{"easy": true, "medium": true, "hard": true, "deadly": true}

// oneShotDraftCounts reports how much linked content a draft created.
type oneShotDraftCounts struct {
	Acts       int `json:"acts"`
	Scenes     int `json:"scenes"`
	NPCs       int `json:"npcs"`
	Locations  int `json:"locations"`
	Encounters int `json:"encounters"`
	Clues      int `json:"clues"`
}

// createOneShotFromDraft creates a one-shot adventure and all of its linked
// content from a draft. The AI commit path and direct JSON import share it so
// both produce identical structures.
func createOneShotFromDraft(ctx context.Context, uid int64, campaignID *int64, raw json.RawMessage) (int64, oneShotDraftCounts, error) {
	var d aiOneShotDraft
	var counts oneShotDraftCounts
	if err := json.Unmarshal(raw, &d); err != nil {
		return 0, counts, fmt.Errorf("invalid draft: %w", err)
	}
	normalizeOneShotDraft(&d)
	now := time.Now().Format("2006-01-02 15:04:05")
	res, err := db.DB.Exec(`INSERT INTO oneshot_adventures
		(user_id, campaign_id, title, premise, hook, template, estimated_minutes, difficulty, notes, created_at, updated_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?)`,
		uid, campaignID, d.Title, d.Premise, d.Hook, "custom", d.EstimatedMinutes, d.Difficulty, d.Notes, now, now)
	if err != nil {
		return 0, counts, err
	}
	adventureID, err := res.LastInsertId()
	if err != nil {
		return 0, counts, err
	}

	counts, err = insertOneShotChildren(ctx, uid, adventureID, campaignID, d)
	if err != nil {
		return 0, counts, err
	}
	return adventureID, counts, nil
}

func normalizeOneShotDraft(d *aiOneShotDraft) {
	if strings.TrimSpace(d.Title) == "" {
		d.Title = "Untitled One-Shot"
	}
	if !aiDifficulties[d.Difficulty] {
		d.Difficulty = "medium"
	}
	if d.EstimatedMinutes <= 0 {
		d.EstimatedMinutes = 180
	}
}

// insertOneShotChildren writes the draft's acts, scenes, NPCs, locations,
// encounters and clues for an adventure. NPCs and locations are matched by
// name within the owner's entities and reused, so re-importing a revised draft
// does not duplicate shared entities.
func insertOneShotChildren(ctx context.Context, uid, adventureID int64, campaignID *int64, d aiOneShotDraft) (oneShotDraftCounts, error) {
	var counts oneShotDraftCounts
	for i, act := range d.Acts {
		mins := act.EstimatedMinutes
		if mins <= 0 {
			mins = 30
		}
		actRes, err := db.DB.Exec("INSERT INTO oneshot_acts(adventure_id, number, title, description, estimated_minutes) VALUES(?,?,?,?,?)",
			adventureID, i+1, act.Title, act.Description, mins)
		if err != nil {
			continue
		}
		counts.Acts++
		actID, _ := actRes.LastInsertId()
		for j, scene := range act.Scenes {
			sceneType := scene.SceneType
			if !aiSceneTypes[sceneType] {
				sceneType = "roleplay"
			}
			sceneMins := scene.EstimatedMinutes
			if sceneMins <= 0 {
				sceneMins = 15
			}
			if _, err := db.DB.Exec("INSERT INTO oneshot_scenes(act_id, number, title, description, scene_type, estimated_minutes, notes) VALUES(?,?,?,?,?,?,'')",
				actID, j+1, scene.Title, scene.Description, sceneType, sceneMins); err == nil {
				counts.Scenes++
			}
		}
	}

	for _, n := range d.NPCs {
		if strings.TrimSpace(n.Name) == "" {
			continue
		}
		npcID, err := db.Client.NPC.Query().Where(npc.UserID(uid), npc.NameEQ(n.Name)).FirstID(ctx)
		if err != nil {
			created, cerr := db.Client.NPC.Create().SetUserID(uid).SetName(n.Name).SetRace(n.Race).SetDescription(n.Description).Save(ctx)
			if cerr != nil {
				continue
			}
			npcID = created.ID
		}
		db.DB.Exec("INSERT OR REPLACE INTO oneshot_adventure_npcs(adventure_id, npc_id, role, story_hook, combat_ready) VALUES(?,?,?,'',0)",
			adventureID, npcID, n.Role)
		counts.NPCs++
	}

	for _, loc := range d.Locations {
		if strings.TrimSpace(loc.Name) == "" {
			continue
		}
		typ := loc.Type
		if strings.TrimSpace(typ) == "" {
			typ = "region"
		}
		locationID, err := db.Client.Location.Query().Where(location.UserID(uid), location.NameEQ(loc.Name)).FirstID(ctx)
		if err != nil {
			created, cerr := db.Client.Location.Create().SetUserID(uid).SetName(loc.Name).SetType(typ).SetDescription(loc.Description).Save(ctx)
			if cerr != nil {
				continue
			}
			locationID = created.ID
		}
		db.DB.Exec("INSERT OR IGNORE INTO oneshot_adventure_locations(adventure_id, location_id) VALUES(?,?)", adventureID, locationID)
		counts.Locations++
	}

	for _, e := range d.Encounters {
		if strings.TrimSpace(e.Name) == "" {
			continue
		}
		diff := e.Difficulty
		if !aiDifficulties[diff] {
			diff = "medium"
		}
		res, err := db.DB.Exec("INSERT INTO encounter_templates(campaign_id,user_id,name,description,environment,difficulty,xp_budget,total_xp,notes) VALUES(?,?,?,?,?,?,0,0,'')",
			campaignID, uid, e.Name, e.Description, "", diff)
		if err != nil {
			continue
		}
		if encID, err := res.LastInsertId(); err == nil {
			db.DB.Exec("INSERT INTO oneshot_adventure_encounters(adventure_id, encounter_id) VALUES(?,?)", adventureID, encID)
			counts.Encounters++
		}
	}

	for _, cl := range d.Clues {
		if strings.TrimSpace(cl.Title) == "" {
			continue
		}
		clueType := cl.ClueType
		if clueType != "direct" && clueType != "witness" && clueType != "object" && clueType != "location" {
			clueType = "direct"
		}
		if _, err := db.DB.Exec("INSERT INTO clues(adventure_id, title, description, clue_type, is_red_herring, sort_order, notes) VALUES(?,?,?,?,0,0,'')",
			adventureID, cl.Title, cl.Description, clueType); err == nil {
			counts.Clues++
		}
	}

	return counts, nil
}

// commitAIOneShot commits a ready session draft through the shared writer.
func commitAIOneShot(ctx context.Context, uid int64, s *aiDraftSession) (int64, string, error) {
	adventureID, _, err := createOneShotFromDraft(ctx, uid, s.CampaignID, s.Draft)
	if err != nil {
		return 0, "", err
	}
	return adventureID, entityURL("adventure", adventureID), nil
}

func commitAINPC(ctx context.Context, uid int64, s *aiDraftSession) (int64, string, error) {
	var d struct {
		Name        string `json:"name"`
		Race        string `json:"race"`
		Class       string `json:"class"`
		Description string `json:"description"`
		Notes       string `json:"notes"`
	}
	if err := json.Unmarshal(s.Draft, &d); err != nil {
		return 0, "", fmt.Errorf("invalid draft: %w", err)
	}
	if strings.TrimSpace(d.Name) == "" {
		return 0, "", fmt.Errorf("name is required")
	}
	npc, err := db.Client.NPC.Create().SetUserID(uid).SetName(d.Name).SetRace(d.Race).SetClass(d.Class).
		SetDescription(d.Description).SetNotes(d.Notes).Save(ctx)
	if err != nil {
		return 0, "", err
	}
	if s.CampaignID != nil {
		db.DB.Exec("INSERT OR IGNORE INTO campaign_npcs(campaign_id, npc_id, role, notes) VALUES(?,?,'','')", *s.CampaignID, npc.ID)
	}
	return npc.ID, entityURL("npc", npc.ID), nil
}

func commitAILocation(ctx context.Context, uid int64, s *aiDraftSession) (int64, string, error) {
	var d struct {
		Name        string `json:"name"`
		Type        string `json:"type"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal(s.Draft, &d); err != nil {
		return 0, "", fmt.Errorf("invalid draft: %w", err)
	}
	if strings.TrimSpace(d.Name) == "" {
		return 0, "", fmt.Errorf("name is required")
	}
	if strings.TrimSpace(d.Type) == "" {
		d.Type = "region"
	}
	loc, err := db.Client.Location.Create().SetUserID(uid).SetName(d.Name).SetType(d.Type).SetDescription(d.Description).Save(ctx)
	if err != nil {
		return 0, "", err
	}
	return loc.ID, entityURL("location", loc.ID), nil
}

func commitAIEncounter(ctx context.Context, uid int64, s *aiDraftSession) (int64, string, error) {
	var d struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Environment string `json:"environment"`
		Difficulty  string `json:"difficulty"`
		Notes       string `json:"notes"`
	}
	if err := json.Unmarshal(s.Draft, &d); err != nil {
		return 0, "", fmt.Errorf("invalid draft: %w", err)
	}
	if strings.TrimSpace(d.Name) == "" {
		return 0, "", fmt.Errorf("name is required")
	}
	if !aiDifficulties[d.Difficulty] {
		d.Difficulty = "medium"
	}
	res, err := db.DB.Exec("INSERT INTO encounter_templates(campaign_id,user_id,name,description,environment,difficulty,xp_budget,total_xp,notes) VALUES(?,?,?,?,?,?,0,0,?)",
		s.CampaignID, uid, d.Name, d.Description, d.Environment, d.Difficulty, d.Notes)
	if err != nil {
		return 0, "", err
	}
	id, _ := res.LastInsertId()
	return id, entityURL("encounter", id), nil
}

func commitAIFaction(ctx context.Context, uid int64, s *aiDraftSession) (int64, string, error) {
	var d struct {
		Name         string `json:"name"`
		Description  string `json:"description"`
		Type         string `json:"type"`
		Headquarters string `json:"headquarters"`
	}
	if err := json.Unmarshal(s.Draft, &d); err != nil {
		return 0, "", fmt.Errorf("invalid draft: %w", err)
	}
	if strings.TrimSpace(d.Name) == "" {
		return 0, "", fmt.Errorf("name is required")
	}
	if strings.TrimSpace(d.Type) == "" {
		d.Type = "organization"
	}
	res, err := db.DB.Exec("INSERT INTO factions(campaign_id,name,description,type,headquarters) VALUES(?,?,?,?,?)",
		s.CampaignID, d.Name, d.Description, d.Type, d.Headquarters)
	if err != nil {
		return 0, "", err
	}
	id, _ := res.LastInsertId()
	return id, entityURL("faction", id), nil
}

func commitAICampaign(ctx context.Context, uid int64, s *aiDraftSession) (int64, string, error) {
	var d struct {
		Name        string `json:"name"`
		PartyName   string `json:"party_name"`
		Description string `json:"description"`
		DMNotes     string `json:"dm_notes"`
	}
	if err := json.Unmarshal(s.Draft, &d); err != nil {
		return 0, "", fmt.Errorf("invalid draft: %w", err)
	}
	if strings.TrimSpace(d.Name) == "" {
		return 0, "", fmt.Errorf("name is required")
	}
	campaign, err := db.Client.Campaign.Create().SetUserID(uid).SetName(d.Name).
		SetPartyName(d.PartyName).SetDescription(d.Description).SetDmNotes(d.DMNotes).Save(ctx)
	if err != nil {
		return 0, "", err
	}
	db.DB.Exec("INSERT OR IGNORE INTO campaign_members(campaign_id, user_id, role) VALUES(?,?,'dm')", campaign.ID, uid)
	return campaign.ID, entityURL("campaign", campaign.ID), nil
}

func commitAIQuest(c *gin.Context, s *aiDraftSession) (int64, string, error) {
	if s.ParentID == nil {
		return 0, "", fmt.Errorf("a character_id is required to create a quest")
	}
	if !canEditCharacterID(c, *s.ParentID) {
		return 0, "", fmt.Errorf("access denied")
	}
	var d struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Objectives  string `json:"objectives"`
		Rewards     string `json:"rewards"`
		Notes       string `json:"notes"`
	}
	if err := json.Unmarshal(s.Draft, &d); err != nil {
		return 0, "", fmt.Errorf("invalid draft: %w", err)
	}
	if strings.TrimSpace(d.Name) == "" {
		return 0, "", fmt.Errorf("name is required")
	}
	q, err := db.Client.Quest.Create().SetCharacterID(*s.ParentID).SetName(d.Name).
		SetDescription(d.Description).SetStatus("active").SetObjectives(d.Objectives).
		SetRewards(d.Rewards).SetNotes(d.Notes).Save(c.Request.Context())
	if err != nil {
		return 0, "", err
	}
	return q.ID, entityURL("quest", q.ID), nil
}

func commitAIItem(ctx context.Context, s *aiDraftSession) (int64, string, error) {
	if s.ParentID == nil {
		return 0, "", fmt.Errorf("an adventure_id is required to create a one-shot item")
	}
	var d struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Category    string  `json:"category"`
		Quantity    int     `json:"quantity"`
		Weight      float64 `json:"weight"`
		PriceGP     float64 `json:"price_gp"`
		IsMagical   bool    `json:"is_magical"`
		Attunement  bool    `json:"attunement"`
		Notes       string  `json:"notes"`
	}
	if err := json.Unmarshal(s.Draft, &d); err != nil {
		return 0, "", fmt.Errorf("invalid draft: %w", err)
	}
	if strings.TrimSpace(d.Name) == "" {
		return 0, "", fmt.Errorf("name is required")
	}
	if strings.TrimSpace(d.Category) == "" {
		d.Category = "gear"
	}
	if d.Quantity <= 0 {
		d.Quantity = 1
	}
	res, err := db.DB.Exec(`INSERT INTO oneshot_items(adventure_id, name, description, category, quantity, weight, price_gp, is_magical, attunement, notes)
		VALUES(?,?,?,?,?,?,?,?,?,?)`,
		*s.ParentID, d.Name, d.Description, d.Category, d.Quantity, d.Weight, d.PriceGP, d.IsMagical, d.Attunement, d.Notes)
	if err != nil {
		return 0, "", err
	}
	id, _ := res.LastInsertId()
	return id, entityURL("item", id), nil
}

// intPtr is a small helper for optional integer parameters.
func intPtr(v int) *int { return &v }
