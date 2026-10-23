package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"villum/crypto"
	"villum/db"
	"villum/middleware"
	"villum/models"
)

// visionTestImageDataURL is a 1x1 transparent PNG used by connection tests so a
// vision endpoint can be checked without uploading a real page.
const visionTestImageDataURL = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNkYAAAAAYAAjCB0C8AAAAASUVORK5CYII="

const visionSystemPrompt = "You transcribe handwritten tabletop RPG notes. Output only the transcribed text, preserving line breaks and bullet points. Do not describe the page, do not add commentary, and do not invent missing words."

// enabledVisionEndpoint returns the first enabled vision endpoint, or the one
// requested when endpointID is non-zero. A requested endpoint must be enabled
// and of type vision.
func enabledVisionEndpoint(ctx context.Context, endpointID int64) (*models.AIEndpoint, error) {
	endpoints, err := db.GetEnabledAIEndpointsByType(ctx, "vision")
	if err != nil {
		return nil, &aiGenError{Status: 500, Msg: "failed to get endpoints"}
	}
	for i := range endpoints {
		if endpointID == 0 || endpoints[i].ID == endpointID {
			full, err := db.GetAIEndpoint(ctx, endpoints[i].ID)
			if err != nil {
				return nil, &aiGenError{Status: 404, Msg: "enabled vision endpoint not found"}
			}
			return full, nil
		}
	}
	return nil, &aiGenError{Status: 409, Msg: "no enabled vision endpoint is configured"}
}

// generateVision transcribes one page image with an OpenAI-compatible vision
// model. imageDataURL must be a data: URL because stored media paths are
// relative and not reachable by the provider.
func generateVision(ctx context.Context, endpointID int64, imageDataURL, prompt string, maxTokens *int, sessionID string) (string, error) {
	if strings.TrimSpace(imageDataURL) == "" {
		return "", &aiGenError{Status: 400, Msg: "image is required"}
	}
	endpoint, err := enabledVisionEndpoint(ctx, endpointID)
	if err != nil {
		return "", err
	}
	apiKey, err := crypto.Decrypt(endpoint.EncryptedAPIKey)
	if err != nil {
		middleware.LogError("ai", "failed to decrypt vision API key", "endpoint_id", endpoint.ID, "error", err)
		return "", &aiGenError{Status: 500, Msg: fmt.Sprintf("failed to authenticate with AI provider: %s", sanitizeError(err))}
	}

	if strings.TrimSpace(prompt) == "" {
		prompt = "Transcribe this handwritten page."
	}
	content := []map[string]any{
		{"type": "text", "text": prompt},
		{"type": "image_url", "image_url": map[string]any{"url": imageDataURL}},
	}
	payload := map[string]any{
		"model": endpoint.Model,
		"messages": []map[string]any{
			{"role": "system", "content": visionSystemPrompt},
			{"role": "user", "content": content},
		},
		"max_tokens": 2048,
	}
	if maxTokens != nil {
		payload["max_tokens"] = *maxTokens
	}

	sid := resolveSessionID(sessionID)
	body, _ := json.Marshal(payload)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", aiRequestURL(endpoint.BaseURL, "/chat/completions"), bytes.NewReader(body))
	if err != nil {
		return "", &aiGenError{Status: 500, Msg: fmt.Sprintf("failed to create request: %s", sanitizeError(err))}
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	setAIProviderHeaders(httpReq, sid)

	client := newAIClient(aiTextTimeout)
	resp, err := client.Do(httpReq)
	if err != nil {
		middleware.LogError("ai", "vision request failed", "endpoint_id", endpoint.ID, "model", endpoint.Model, "error", err)
		return "", &aiGenError{Status: 502, Msg: fmt.Sprintf("AI provider request failed: %s", sanitizeError(err))}
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 65536))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errMsg := fmt.Sprintf("AI provider returned HTTP %d: %s", resp.StatusCode, truncateResponse(string(respBody)))
		middleware.LogError("ai", "vision returned non-2xx", "status", resp.StatusCode, "response_preview", truncateResponse(string(respBody)))
		return "", &aiGenError{Status: 502, Msg: errMsg}
	}
	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", &aiGenError{Status: 500, Msg: "failed to parse AI response"}
	}
	if len(result.Choices) == 0 {
		return "", &aiGenError{Status: 502, Msg: "AI provider returned no choices"}
	}
	text := strings.TrimSpace(result.Choices[0].Message.Content)
	if text == "" {
		return "", &aiGenError{Status: 502, Msg: "AI provider returned an empty transcription"}
	}
	middleware.LogInfo("ai", "vision transcription succeeded", "endpoint_id", endpoint.ID, "model", endpoint.Model, "content_len", len(text))
	return text, nil
}

// pageDataURL loads a stored upload and returns it as a base64 data URL.
func pageDataURL(uploadURL string) (string, error) {
	clean := filepath.Clean(strings.TrimPrefix(uploadURL, "/media/"))
	if strings.HasPrefix(clean, "..") || filepath.IsAbs(clean) {
		return "", fmt.Errorf("invalid media path")
	}
	full := filepath.Join(MediaPath, clean)
	info, err := os.Stat(full)
	if err != nil {
		return "", err
	}
	if info.Size() > 10*1024*1024 {
		return "", fmt.Errorf("page image is too large")
	}
	data, err := os.ReadFile(full)
	if err != nil {
		return "", err
	}
	mime := http.DetectContentType(data)
	if !strings.HasPrefix(mime, "image/") {
		return "", fmt.Errorf("page is not an image")
	}
	return "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(data), nil
}
