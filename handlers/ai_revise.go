package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"villum/db"
)

// aiDraftReviseAddendum steers the assistant when it is revising an existing
// draft rather than creating a new one.
const aiDraftReviseAddendum = `Revision mode:
- You are editing an existing draft, not creating a new one.
- Apply only the changes the user asks for and keep every other field exactly as it was.
- Return the complete revised object in draft with status "ready".
- If the request is ambiguous, reply with status "chatting", a question in message, and draft null.`

// ReviseAIDraft applies an instruction to a draft JSON and returns the revised
// draft. POST /api/ai/revise — stateless: the client sends the current JSON on
// every call, so the same flow works before and after an import.
func ReviseAIDraft(c *gin.Context) {
	if !aiEnabled(c.Request.Context()) {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AI features are disabled"})
		return
	}
	var req struct {
		EntityType  string `json:"entity_type"`
		JSON        string `json:"json"`
		Instruction string `json:"instruction"`
		EndpointID  int64  `json:"endpoint_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if _, ok := aiDraftEntities[req.EntityType]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported entity_type"})
		return
	}
	if strings.TrimSpace(req.Instruction) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "instruction is required"})
		return
	}
	draft, err := extractDraftJSON(req.JSON)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	endpointID := req.EndpointID
	if endpointID == 0 {
		endpoints, err := db.GetEnabledAIEndpointsByType(ctx, "text")
		if err != nil || len(endpoints) == 0 {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "no enabled text endpoint is configured"})
			return
		}
		endpointID = endpoints[0].ID
	}

	system := aiDraftSystemPrompt(req.EntityType) + "\n\n" + aiDraftReviseAddendum
	user := "Current draft JSON:\n" + string(draft) + "\n\nRequested changes:\n" + strings.TrimSpace(req.Instruction)
	messages := []map[string]string{
		{"role": "system", "content": system},
		{"role": "user", "content": user},
	}
	reply, finishReason, err := generateChat(ctx, endpointID, messages,
		aiDraftTokenBudget(ctx, endpointID), "", aiDraftTimeout)
	if err != nil {
		writeAIGenError(c, err)
		return
	}
	if err := validateAIDraftReply(reply, finishReason); err != nil {
		writeAIGenError(c, err)
		return
	}
	status, message, revised := parseAIDraftReply(reply)
	c.JSON(http.StatusOK, gin.H{"status": status, "message": message, "draft": revised})
}
