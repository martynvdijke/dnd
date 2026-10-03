package ai

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"villum/crypto"
	"villum/db"
	"villum/middleware"
	"villum/models"
)

const DefaultSystemPrompt = "You are a helpful assistant for a D&D website called villum. You help DMs create compelling narratives, NPCs, locations, items, and other TTRPG content. Be creative and concise."

const (
	TestTimeout  = 30 * time.Second
	TextTimeout  = 60 * time.Second
	DraftTimeout = 180 * time.Second
	ImageTimeout = 120 * time.Second
	FetchTimeout = 60 * time.Second
)

// Kept for handler compatibility (unexported aliases removed, exported now)
var (
	AppVersion string
)

func SetAppVersion(v string) { AppVersion = v }

// AIGenError mirrors handlers.aiGenError
type AIGenError struct {
	Status int
	Msg    string
}

func (e *AIGenError) Error() string { return e.Msg }

// GenerateRequest for GenerateText
type GenerateRequest struct {
	EndpointID int64
	Prompt     string
	System     string
	MaxTokens  *int
	SessionID  string
}

// GenerateText performs a single-turn generation.
func GenerateText(ctx context.Context, passed *sql.DB, req GenerateRequest) (string, string, error) {
	if strings.TrimSpace(req.Prompt) == "" {
		return "", "", &AIGenError{Status: 400, Msg: "prompt is required"}
	}
	systemPrompt := req.System
	if strings.TrimSpace(systemPrompt) == "" {
		systemPrompt = DefaultSystemPrompt
	}
	messages := []map[string]string{
		{"role": "system", "content": systemPrompt},
		{"role": "user", "content": req.Prompt},
	}
	return GenerateChat(ctx, passed, req.EndpointID, messages, req.MaxTokens, req.SessionID, TextTimeout)
}

// GenerateChat performs chat completion.
func GenerateChat(ctx context.Context, passed *sql.DB, endpointID int64, messages []map[string]string, maxTokens *int, sessionID string, timeout time.Duration) (string, string, error) {
	_ = passed // endpoint lookup uses global ent helpers (db.GetAIEndpoint / db.GetEnabledAIEndpointsByType)
	if endpointID == 0 {
		return "", "", &AIGenError{Status: 400, Msg: "endpoint_id is required"}
	}
	if len(messages) == 0 {
		return "", "", &AIGenError{Status: 400, Msg: "messages are required"}
	}
	endpoints, err := db.GetEnabledAIEndpointsByType(ctx, "text")
	if err != nil {
		return "", "", &AIGenError{Status: 500, Msg: "failed to get endpoints"}
	}
	var endpoint *models.AIEndpoint
	for i := range endpoints {
		if endpoints[i].ID == endpointID {
			endpoint = &endpoints[i]
			break
		}
	}
	if endpoint == nil {
		return "", "", &AIGenError{Status: 404, Msg: "enabled text endpoint not found"}
	}
	middleware.LogDebug("ai", "text generation start", "endpoint_id", endpointID, "model", endpoint.Model, "message_count", len(messages))
	fullEndpoint, err := db.GetAIEndpoint(ctx, endpointID)
	if err != nil {
		return "", "", &AIGenError{Status: 404, Msg: "enabled text endpoint not found"}
	}
	apiKey, err := crypto.Decrypt(fullEndpoint.EncryptedAPIKey)
	if err != nil {
		middleware.LogError("ai", "failed to decrypt API key", "endpoint_id", endpointID, "error", err)
		return "", "", &AIGenError{Status: 500, Msg: fmt.Sprintf("failed to authenticate with AI provider: %s", SanitizeError(err))}
	}
	payload := map[string]any{
		"model":    endpoint.Model,
		"messages": messages,
	}
	if maxTokens != nil {
		payload["max_tokens"] = *maxTokens
	}
	sid := ResolveSessionID(sessionID)
	body, _ := json.Marshal(payload)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", AIRequestURL(endpoint.BaseURL, "/chat/completions"), bytes.NewReader(body))
	if err != nil {
		return "", "", &AIGenError{Status: 500, Msg: fmt.Sprintf("failed to create request: %s", SanitizeError(err))}
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	SetAIProviderHeaders(httpReq, sid)
	eff := timeout
	if dl, ok := ctx.Deadline(); ok {
		if remaining := time.Until(dl); remaining > 0 && remaining < eff {
			eff = remaining
		}
	}
	client := NewAIClient(eff)
	resp, err := client.Do(httpReq)
	if err != nil {
		middleware.LogError("ai", "text generation request failed", "endpoint_id", endpoint.ID, "model", endpoint.Model, "error", err)
		return "", "", &AIGenError{Status: 502, Msg: fmt.Sprintf("AI provider request failed: %s", SanitizeError(err))}
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 65536))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errMsg := fmt.Sprintf("AI provider returned HTTP %d: %s", resp.StatusCode, TruncateResponse(string(respBody)))
		middleware.LogError("ai", "text generation returned non-2xx", "status", resp.StatusCode, "response_preview", TruncateResponse(string(respBody)))
		return "", "", &AIGenError{Status: 502, Msg: errMsg}
	}
	var result struct {
		Choices []struct {
			Message struct {
				Content          string `json:"content"`
				ReasoningContent string `json:"reasoning_content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		middleware.LogError("ai", "failed to parse text response", "endpoint_id", endpoint.ID, "error", err)
		return "", "", &AIGenError{Status: 500, Msg: "failed to parse AI response"}
	}
	if len(result.Choices) == 0 {
		return "", "no_choices", nil
	}
	middleware.LogInfo("ai", "text generation succeeded", "endpoint_id", endpoint.ID, "model", endpoint.Model,
		"finish_reason", result.Choices[0].FinishReason,
		"content_len", len(result.Choices[0].Message.Content),
		"reasoning_len", len(result.Choices[0].Message.ReasoningContent))
	return result.Choices[0].Message.Content, result.Choices[0].FinishReason, nil
}

func NewAIClient(timeout time.Duration) *http.Client {
	return &http.Client{Timeout: timeout}
}

func SetAIProviderHeaders(req *http.Request, sessionID string) {
	ver := AppVersion
	if ver == "" {
		ver = "0.0.0-dev"
	}
	req.Header.Set("User-Agent", "villum/"+ver)
	req.Header.Set("x-opencode-session", sessionID)
}

func GenerateSessionID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return hex.EncodeToString([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	}
	return hex.EncodeToString(b)
}

func ResolveSessionID(inbound string) string {
	if strings.TrimSpace(inbound) != "" {
		return strings.TrimSpace(inbound)
	}
	return GenerateSessionID()
}

var aiEndpointOperationPaths = []string{
	"/chat/completions",
	"/images/generations",
	"/completions",
}

func NormalizeAIBaseURL(raw string) string {
	u := strings.TrimSpace(raw)
	u = strings.TrimRight(u, "/")
	for _, suffix := range aiEndpointOperationPaths {
		if strings.HasSuffix(u, suffix) {
			u = strings.TrimRight(strings.TrimSuffix(u, suffix), "/")
			break
		}
	}
	return u
}

func AIRequestURL(baseURL, operationPath string) string {
	return NormalizeAIBaseURL(baseURL) + operationPath
}

func ValidateAIBaseURL(raw string) (string, bool) {
	u := NormalizeAIBaseURL(raw)
	parsed, err := url.Parse(u)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return u, false
	}
	return u, true
}

func IsEnabled(ctx context.Context, passed *sql.DB) bool {
	d := db.DB
	if passed != nil {
		d = passed
	}
	var value string
	if err := d.QueryRowContext(ctx, "SELECT value FROM app_settings WHERE key='ai_enabled'").Scan(&value); err != nil {
		return true
	}
	return value == "1"
}

func ResolveEndpoint(ctx context.Context, passed *sql.DB, preferredID int64) (*models.AIEndpoint, error) {
	_ = passed // endpoint lookup uses global ent helpers (db.GetAIEndpoint / db.GetEnabledAIEndpointsByType)
	if preferredID != 0 {
		ep, err := db.GetAIEndpoint(ctx, preferredID)
		if err != nil {
			return nil, err
		}
		return ep, nil
	}
	eps, err := db.GetEnabledAIEndpointsByType(ctx, "text")
	if err != nil || len(eps) == 0 {
		if err == nil {
			err = fmt.Errorf("no enabled text endpoint")
		}
		return nil, err
	}
	ep, err := db.GetAIEndpoint(ctx, eps[0].ID)
	if err != nil {
		return nil, err
	}
	return ep, nil
}

func SanitizeError(err error) string {
	msg := err.Error()
	if len(msg) > 100 {
		msg = msg[:100] + "..."
	}
	msg = strings.ReplaceAll(msg, "\n", " ")
	return msg
}

func TruncateResponse(s string) string {
	if len(s) > 200 {
		return s[:200] + "..."
	}
	return s
}
