package models

type AIEndpoint struct {
	ID              int64    `json:"id"`
	Name            string   `json:"name"`
	Type            string   `json:"type"` // "text" or "image"
	BaseURL         string   `json:"base_url"`
	EncryptedAPIKey string   `json:"encrypted_api_key,omitempty"`
	Model           string   `json:"model"`
	Tags            []string `json:"tags,omitempty"`
	Enabled         bool     `json:"enabled"`
	Temperature     *float64 `json:"temperature,omitempty"`
	MaxTokens       *int     `json:"max_tokens,omitempty"`
	ImageSize       *string  `json:"image_size,omitempty"`
	CreatedAt       string   `json:"created_at"`
	UpdatedAt       string   `json:"updated_at"`
}
