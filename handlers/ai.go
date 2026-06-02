package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"villum/crypto"
	"villum/db"
	"villum/models"
)

const DefaultSystemPrompt = "You are a helpful assistant for a D&D website called villum. You help DMs create compelling narratives, NPCs, locations, items, and other TTRPG content. Be creative and concise."

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
		log.Printf("[ai] failed to encrypt API key: %v", err)
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
		log.Printf("[ai] failed to create endpoint: %v", err)
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
			log.Printf("[ai] failed to encrypt API key: %v", err)
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
		log.Printf("[ai] failed to update endpoint %d: %v", id, err)
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
		log.Printf("[ai] failed to decrypt API key for endpoint %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to test endpoint"})
		return
	}

	client := &http.Client{Timeout: 30 * time.Second}

	if endpoint.Type == "text" {
		payload := map[string]interface{}{
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
			c.JSON(http.StatusOK, gin.H{"success": false, "message": fmt.Sprintf("Connection failed: %s", sanitizeError(err))})
			return
		}
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			c.JSON(http.StatusOK, gin.H{"success": true, "message": "Endpoint responded successfully"})
		} else {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": fmt.Sprintf("HTTP %d: %s", resp.StatusCode, truncateResponse(string(respBody)))})
		}
	} else {
		payload := map[string]interface{}{
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
			c.JSON(http.StatusOK, gin.H{"success": false, "message": fmt.Sprintf("Connection failed: %s", sanitizeError(err))})
			return
		}
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			c.JSON(http.StatusOK, gin.H{"success": true, "message": "Endpoint responded successfully"})
		} else {
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

	// Decrypt API key from the full DB record
	fullEndpoint, err := db.GetAIEndpoint(c.Request.Context(), req.EndpointID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get endpoint"})
		return
	}

	apiKey, err := crypto.Decrypt(fullEndpoint.EncryptedAPIKey)
	if err != nil {
		log.Printf("[ai] failed to decrypt API key: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to authenticate with AI provider"})
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

	payload := map[string]interface{}{
		"model":    endpoint.Model,
		"messages": messages,
	}
	if req.MaxTokens != nil {
		payload["max_tokens"] = *req.MaxTokens
	}

	body, _ := json.Marshal(payload)
	httpReq, err := http.NewRequest("POST", endpoint.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		log.Printf("[ai] text generation request failed: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "AI provider request failed"})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 65536))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("[ai] text generation returned HTTP %d: %s", resp.StatusCode, truncateResponse(string(respBody)))
		c.JSON(http.StatusBadGateway, gin.H{"error": "AI provider returned an error"})
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
		log.Printf("[ai] failed to parse text response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse AI response"})
		return
	}

	if len(result.Choices) == 0 {
		c.JSON(http.StatusOK, textGenResponse{Text: "", Finish: "no_choices"})
		return
	}

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

	fullEndpoint, err := db.GetAIEndpoint(c.Request.Context(), req.EndpointID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get endpoint"})
		return
	}

	apiKey, err := crypto.Decrypt(fullEndpoint.EncryptedAPIKey)
	if err != nil {
		log.Printf("[ai] failed to decrypt API key: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to authenticate with AI provider"})
		return
	}

	imgSize := req.Size
	if imgSize == "" && endpoint.ImageSize != nil && *endpoint.ImageSize != "" {
		imgSize = *endpoint.ImageSize
	}
	if imgSize == "" {
		imgSize = "1024x1024"
	}

	payload := map[string]interface{}{
		"model":  endpoint.Model,
		"prompt": req.Prompt,
		"n":      req.N,
		"size":   imgSize,
	}

	body, _ := json.Marshal(payload)
	httpReq, err := http.NewRequest("POST", endpoint.BaseURL+"/images/generations", bytes.NewReader(body))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		log.Printf("[ai] image generation request failed: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "AI provider request failed"})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 262144))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("[ai] image generation returned HTTP %d: %s", resp.StatusCode, truncateResponse(string(respBody)))
		c.JSON(http.StatusBadGateway, gin.H{"error": "AI provider returned an error"})
		return
	}

	var result struct {
		Data []struct {
			URL string `json:"url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		log.Printf("[ai] failed to parse image response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse AI response"})
		return
	}

	images := make([]string, 0, len(result.Data))
	for _, d := range result.Data {
		if d.URL != "" {
			images = append(images, d.URL)
		}
	}

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
