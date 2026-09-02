package handlers

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"villum/crypto"
	"villum/db"
	"villum/middleware"
	"villum/models"
)

const DefaultSystemPrompt = "You are a helpful assistant for a D&D website called villum. You help DMs create compelling narratives, NPCs, locations, items, and other TTRPG content. Be creative and concise."

const (
	aiTestTimeout  = 30 * time.Second
	aiTextTimeout  = 60 * time.Second
	aiImageTimeout = 120 * time.Second
	aiFetchTimeout = 60 * time.Second
)

func newAIClient(timeout time.Duration) *http.Client {
	return &http.Client{Timeout: timeout}
}

func ListAIEndpoints(c *gin.Context) {
	endpoints, err := db.GetAIEndpoints(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list AI endpoints"})
		return
	}
	// Redact API keys
	for i := range endpoints {
		endpoints[i].EncryptedAPIKey = ""
	}
	c.JSON(http.StatusOK, endpoints)
}

func HandleListEnabledAIEndpoints(c *gin.Context) {
	endpointType := c.DefaultQuery("type", "")
	if endpointType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type query parameter is required (text or image)"})
		return
	}
	if endpointType != "text" && endpointType != "image" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type must be 'text' or 'image'"})
		return
	}
	if !aiEnabled(c.Request.Context()) {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}
	endpoints, err := db.ListEnabledAIEndpoints(c.Request.Context(), endpointType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list AI endpoints"})
		return
	}
	c.JSON(http.StatusOK, endpoints)
}

func GetAIEndpoint(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid endpoint ID"})
		return
	}
	endpoint, err := db.GetAIEndpoint(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "AI endpoint not found"})
		return
	}
	endpoint.EncryptedAPIKey = ""
	c.JSON(http.StatusOK, endpoint)
}

type createAIEndpointRequest struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	BaseURL     string   `json:"base_url"`
	APIKey      string   `json:"api_key"`
	Model       string   `json:"model"`
	Tags        []string `json:"tags"`
	Enabled     bool     `json:"enabled"`
	Temperature *float64 `json:"temperature"`
	MaxTokens   *int     `json:"max_tokens"`
	ImageSize   *string  `json:"image_size"`
}

func CreateAIEndpoint(c *gin.Context) {
	var req createAIEndpointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.Name == "" || req.Type == "" || req.BaseURL == "" || req.APIKey == "" || req.Model == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name, type, base_url, api_key, and model are required"})
		return
	}

	if req.Type != "text" && req.Type != "image" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type must be 'text' or 'image'"})
		return
	}

	unique, err := db.CheckAIEndpointNameUnique(c.Request.Context(), req.Name, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check name uniqueness"})
		return
	}
	if !unique {
		c.JSON(http.StatusConflict, gin.H{"error": "an endpoint with this name already exists"})
		return
	}

	encryptedKey, err := crypto.Encrypt(req.APIKey)
	if err != nil {
		middleware.LogError("ai", "failed to encrypt API key", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store API key securely"})
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)
	endpoint := &models.AIEndpoint{
		Name:            req.Name,
		Type:            req.Type,
		BaseURL:         req.BaseURL,
		EncryptedAPIKey: encryptedKey,
		Model:           req.Model,
		Tags:            req.Tags,
		Enabled:         req.Enabled,
		Temperature:     req.Temperature,
		MaxTokens:       req.MaxTokens,
		ImageSize:       req.ImageSize,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	created, err := db.CreateAIEndpoint(c.Request.Context(), endpoint)
	if err != nil {
		middleware.LogError("ai", "failed to create endpoint", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create AI endpoint"})
		return
	}
	created.EncryptedAPIKey = ""
	c.JSON(http.StatusCreated, created)
}

type updateAIEndpointRequest struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	BaseURL     string   `json:"base_url"`
	APIKey      string   `json:"api_key"`
	Model       string   `json:"model"`
	Tags        []string `json:"tags"`
	Enabled     bool     `json:"enabled"`
	Temperature *float64 `json:"temperature"`
	MaxTokens   *int     `json:"max_tokens"`
	ImageSize   *string  `json:"image_size"`
}

func UpdateAIEndpoint(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid endpoint ID"})
		return
	}

	var req updateAIEndpointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.Name == "" || req.Type == "" || req.BaseURL == "" || req.Model == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name, type, base_url, and model are required"})
		return
	}

	unique, err := db.CheckAIEndpointNameUnique(c.Request.Context(), req.Name, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check name uniqueness"})
		return
	}
	if !unique {
		c.JSON(http.StatusConflict, gin.H{"error": "an endpoint with this name already exists"})
		return
	}

	encryptedKey := ""
	if req.APIKey != "" {
		encryptedKey, err = crypto.Encrypt(req.APIKey)
		if err != nil {
			middleware.LogError("ai", "failed to encrypt API key", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store API key securely"})
			return
		}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	endpoint := &models.AIEndpoint{
		Name:            req.Name,
		Type:            req.Type,
		BaseURL:         req.BaseURL,
		EncryptedAPIKey: encryptedKey,
		Model:           req.Model,
		Tags:            req.Tags,
		Enabled:         req.Enabled,
		Temperature:     req.Temperature,
		MaxTokens:       req.MaxTokens,
		ImageSize:       req.ImageSize,
		UpdatedAt:       now,
	}

	updated, err := db.UpdateAIEndpoint(c.Request.Context(), id, endpoint)
	if err != nil {
		middleware.LogError("ai", "failed to update endpoint", "endpoint_id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update AI endpoint"})
		return
	}
	updated.EncryptedAPIKey = ""
	c.JSON(http.StatusOK, updated)
}

func DeleteAIEndpoint(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid endpoint ID"})
		return
	}

	if err := db.DeleteAIEndpoint(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "AI endpoint not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func TestAIEndpoint(c *gin.Context) {
	if !aiEnabled(c.Request.Context()) {
		c.JSON(http.StatusForbidden, gin.H{"error": "AI features are disabled"})
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid endpoint ID"})
		return
	}

	endpoint, err := db.GetAIEndpoint(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "AI endpoint not found"})
		return
	}

	apiKey, err := crypto.Decrypt(endpoint.EncryptedAPIKey)
	if err != nil {
		middleware.LogError("ai", "failed to decrypt API key", "endpoint_id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to test endpoint"})
		return
	}

	middleware.LogDebug("ai", "ai endpoint test start", "endpoint_id", id, "model", endpoint.Model, "type", endpoint.Type)

	client := newAIClient(aiTestTimeout)

	if endpoint.Type == "text" {
		payload := map[string]any{
			"model": endpoint.Model,
			"messages": []map[string]string{
				{"role": "user", "content": "Say hello in one word."},
			},
			"max_tokens": 10,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", endpoint.BaseURL+"/chat/completions", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+apiKey)

		resp, err := client.Do(req)
		if err != nil {
			middleware.LogWarn("ai", "ai endpoint test connection failed", "endpoint_id", id, "type", "text", "error", sanitizeError(err))
			c.JSON(http.StatusOK, gin.H{"success": false, "message": fmt.Sprintf("Connection failed: %s", sanitizeError(err))})
			return
		}
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			middleware.LogInfo("ai", "ai endpoint test succeeded", "endpoint_id", id, "type", "text")
			c.JSON(http.StatusOK, gin.H{"success": true, "message": "Endpoint responded successfully"})
		} else {
			middleware.LogWarn("ai", "ai endpoint test returned non-2xx", "endpoint_id", id, "type", "text", "status", resp.StatusCode)
			c.JSON(http.StatusOK, gin.H{"success": false, "message": fmt.Sprintf("HTTP %d: %s", resp.StatusCode, truncateResponse(string(respBody)))})
		}
	} else {
		payload := map[string]any{
			"model":  endpoint.Model,
			"prompt": "A red dragon flying over a castle",
			"n":      1,
			"size":   "1024x1024",
		}
		if endpoint.ImageSize != nil && *endpoint.ImageSize != "" {
			payload["size"] = *endpoint.ImageSize
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", endpoint.BaseURL+"/images/generations", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+apiKey)

		resp, err := client.Do(req)
		if err != nil {
			middleware.LogWarn("ai", "ai endpoint test connection failed", "endpoint_id", id, "type", "image", "error", sanitizeError(err))
			c.JSON(http.StatusOK, gin.H{"success": false, "message": fmt.Sprintf("Connection failed: %s", sanitizeError(err))})
			return
		}
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			middleware.LogInfo("ai", "ai endpoint test succeeded", "endpoint_id", id, "type", "image")
			c.JSON(http.StatusOK, gin.H{"success": true, "message": "Endpoint responded successfully"})
		} else {
			middleware.LogWarn("ai", "ai endpoint test returned non-2xx", "endpoint_id", id, "type", "image", "status", resp.StatusCode)
			c.JSON(http.StatusOK, gin.H{"success": false, "message": fmt.Sprintf("HTTP %d: %s", resp.StatusCode, truncateResponse(string(respBody)))})
		}
	}
}

type textGenRequest struct {
	EndpointID int64  `json:"endpoint_id"`
	Prompt     string `json:"prompt"`
	System     string `json:"system,omitempty"`
	MaxTokens  *int   `json:"max_tokens,omitempty"`
}

type textGenResponse struct {
	Text   string `json:"text"`
	Finish string `json:"finish_reason,omitempty"`
}

func HandleTextGeneration(c *gin.Context) {
	if !aiEnabled(c.Request.Context()) {
		c.JSON(http.StatusForbidden, gin.H{"error": "AI features are disabled"})
		return
	}
	var req textGenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if req.EndpointID == 0 || req.Prompt == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "endpoint_id and prompt are required"})
		return
	}

	endpoints, err := db.GetEnabledAIEndpointsByType(c.Request.Context(), "text")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get endpoints"})
		return
	}

	var endpoint *models.AIEndpoint
	for i := range endpoints {
		if endpoints[i].ID == req.EndpointID {
			endpoint = &endpoints[i]
			break
		}
	}
	if endpoint == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "enabled text endpoint not found"})
		return
	}

	middleware.LogDebug("ai", "text generation start", "endpoint_id", req.EndpointID, "model", endpoint.Model, "prompt_length", len(req.Prompt))

	// Decrypt API key from the full DB record
	fullEndpoint, err := db.GetAIEndpoint(c.Request.Context(), req.EndpointID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get endpoint"})
		return
	}

	apiKey, err := crypto.Decrypt(fullEndpoint.EncryptedAPIKey)
	if err != nil {
		middleware.LogError("ai", "failed to decrypt API key", "endpoint_id", req.EndpointID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to authenticate with AI provider: %s", sanitizeError(err))})
		return
	}

	systemPrompt := req.System
	if systemPrompt == "" {
		systemPrompt = DefaultSystemPrompt
	}

	messages := []map[string]string{
		{"role": "system", "content": systemPrompt},
		{"role": "user", "content": req.Prompt},
	}

	payload := map[string]any{
		"model":    endpoint.Model,
		"messages": messages,
	}
	if req.MaxTokens != nil {
		payload["max_tokens"] = *req.MaxTokens
	}

	body, _ := json.Marshal(payload)
	httpReq, err := http.NewRequest("POST", endpoint.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create request: %s", sanitizeError(err))})
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	client := newAIClient(aiTextTimeout)
	resp, err := client.Do(httpReq)
	if err != nil {
		errMsg := fmt.Sprintf("AI provider request failed: %s", sanitizeError(err))
		middleware.LogError("ai", "text generation request failed", "endpoint_id", endpoint.ID, "model", endpoint.Model, "error", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": errMsg})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 65536))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errMsg := fmt.Sprintf("AI provider returned HTTP %d: %s", resp.StatusCode, truncateResponse(string(respBody)))
		middleware.LogError("ai", "text generation returned non-2xx", "status", resp.StatusCode, "response_preview", truncateResponse(string(respBody)))
		c.JSON(http.StatusBadGateway, gin.H{"error": errMsg})
		return
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		middleware.LogError("ai", "failed to parse text response", "endpoint_id", endpoint.ID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse AI response"})
		return
	}

	if len(result.Choices) == 0 {
		c.JSON(http.StatusOK, textGenResponse{Text: "", Finish: "no_choices"})
		return
	}

	middleware.LogInfo("ai", "text generation succeeded", "endpoint_id", endpoint.ID, "model", endpoint.Model, "finish_reason", result.Choices[0].FinishReason)

	c.JSON(http.StatusOK, textGenResponse{
		Text:   result.Choices[0].Message.Content,
		Finish: result.Choices[0].FinishReason,
	})
}

type imageGenRequest struct {
	EndpointID int64  `json:"endpoint_id"`
	Prompt     string `json:"prompt"`
	Size       string `json:"size,omitempty"`
	N          int    `json:"n,omitempty"`
}

type imageGenResponse struct {
	Images []string `json:"images"`
}

func HandleImageGeneration(c *gin.Context) {
	if !aiEnabled(c.Request.Context()) {
		c.JSON(http.StatusForbidden, gin.H{"error": "AI features are disabled"})
		return
	}
	var req imageGenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if req.EndpointID == 0 || req.Prompt == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "endpoint_id and prompt are required"})
		return
	}
	if req.N <= 0 {
		req.N = 1
	}
	if req.N > 4 {
		req.N = 4
	}

	endpoints, err := db.GetEnabledAIEndpointsByType(c.Request.Context(), "image")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get endpoints"})
		return
	}

	var endpoint *models.AIEndpoint
	for i := range endpoints {
		if endpoints[i].ID == req.EndpointID {
			endpoint = &endpoints[i]
			break
		}
	}
	if endpoint == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "enabled image endpoint not found"})
		return
	}

	middleware.LogDebug("ai", "image generation start", "endpoint_id", req.EndpointID, "model", endpoint.Model, "prompt_length", len(req.Prompt))

	fullEndpoint, err := db.GetAIEndpoint(c.Request.Context(), req.EndpointID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get endpoint"})
		return
	}

	apiKey, err := crypto.Decrypt(fullEndpoint.EncryptedAPIKey)
	if err != nil {
		middleware.LogError("ai", "failed to decrypt API key", "endpoint_id", req.EndpointID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to authenticate with AI provider: %s", sanitizeError(err))})
		return
	}

	imgSize := req.Size
	if imgSize == "" && endpoint.ImageSize != nil && *endpoint.ImageSize != "" {
		imgSize = *endpoint.ImageSize
	}
	if imgSize == "" {
		imgSize = "1024x1024"
	}

	payload := map[string]any{
		"model":  endpoint.Model,
		"prompt": req.Prompt,
		"n":      req.N,
		"size":   imgSize,
	}

	body, _ := json.Marshal(payload)
	httpReq, err := http.NewRequest("POST", endpoint.BaseURL+"/images/generations", bytes.NewReader(body))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create request: %s", sanitizeError(err))})
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	client := newAIClient(aiImageTimeout)
	resp, err := client.Do(httpReq)
	if err != nil {
		errMsg := fmt.Sprintf("AI provider request failed: %s", sanitizeError(err))
		middleware.LogError("ai", "image generation request failed", "endpoint_id", endpoint.ID, "model", endpoint.Model, "error", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": errMsg})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 262144))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errMsg := fmt.Sprintf("AI provider returned HTTP %d: %s", resp.StatusCode, truncateResponse(string(respBody)))
		middleware.LogError("ai", "image generation returned non-2xx", "status", resp.StatusCode, "response_preview", truncateResponse(string(respBody)))
		c.JSON(http.StatusBadGateway, gin.H{"error": errMsg})
		return
	}

	var result struct {
		Data []struct {
			URL string `json:"url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		middleware.LogError("ai", "failed to parse image response", "endpoint_id", endpoint.ID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse AI response"})
		return
	}

	images := make([]string, 0, len(result.Data))
	for _, d := range result.Data {
		if d.URL != "" {
			images = append(images, d.URL)
		}
	}

	middleware.LogInfo("ai", "image generation succeeded", "endpoint_id", endpoint.ID, "model", endpoint.Model, "image_count", len(images))

	c.JSON(http.StatusOK, imageGenResponse{Images: images})
}

func sanitizeError(err error) string {
	msg := err.Error()
	if len(msg) > 100 {
		msg = msg[:100] + "..."
	}
	// Remove any potential sensitive data patterns
	msg = strings.ReplaceAll(msg, "\n", " ")
	return msg
}

func truncateResponse(s string) string {
	if len(s) > 200 {
		return s[:200] + "..."
	}
	return s
}

type saveImageRequest struct {
	URL string `json:"url"`
}

type saveImageResponse struct {
	URL string `json:"url"`
}

// SaveGeneratedImage downloads an AI-generated image and stores it in the media library.
func SaveGeneratedImage(c *gin.Context) {
	if !aiEnabled(c.Request.Context()) {
		c.JSON(http.StatusForbidden, gin.H{"error": "AI features are disabled"})
		return
	}
	var req saveImageRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.URL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url is required"})
		return
	}

	u, err := url.Parse(req.URL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid url"})
		return
	}

	// SSRF guard: reject loopback, private, link-local, and unspecified hosts
	ips, err := net.LookupIP(u.Hostname())
	if err != nil || len(ips) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not resolve host"})
		return
	}
	for _, ip := range ips {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "url host must be publicly reachable"})
			return
		}
	}

	client := newAIClient(aiFetchTimeout)
	resp, err := client.Get(req.URL)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": sanitizeError(err)})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("download failed: %s", resp.Status)})
		return
	}

	ctype := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ctype, "image/") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url did not return an image"})
		return
	}

	ext := map[string]string{
		"image/png":  "png",
		"image/jpeg": "jpg",
		"image/jpg":  "jpg",
		"image/webp": "webp",
		"image/gif":  "gif",
	}[ctype]
	if ext == "" {
		ext = "png"
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": sanitizeError(err)})
		return
	}

	rb := make([]byte, 4)
	if _, err := rand.Read(rb); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate filename"})
		return
	}
	filename := fmt.Sprintf("ai-%d-%s.%s", time.Now().Unix(), hex.EncodeToString(rb), ext)
	dir := filepath.Join(MediaPath, "ai")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": sanitizeError(err)})
		return
	}
	if err := os.WriteFile(filepath.Join(dir, filename), data, 0o644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": sanitizeError(err)})
		return
	}

	middleware.LogInfo("ai", "generated image saved", "file", filename, "bytes", len(data))

	c.JSON(http.StatusOK, saveImageResponse{URL: "/media/ai/" + filename})
}

// GetAIEnabled exposes the site-wide AI feature toggle to authenticated
// clients (DM-facing UI reads this to hide AI controls when disabled).
func GetAIEnabled(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"enabled": aiEnabled(c.Request.Context())})
}
