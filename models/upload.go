package models

type Upload struct {
	ID                   int    `json:"id"`
	Hash                 string `json:"hash"`
	Ext                  string `json:"ext"`
	URL                  string `json:"url"`
	ResizedURL           string `json:"resized_url"`
	ThumbnailURL         string `json:"thumbnail_url"`
	OwnerType            string `json:"owner_type"`
	OwnerID              int64  `json:"owner_id"`
	Kind                 string `json:"kind"`
	Source               string `json:"source"`
	TelegramFileUniqueID string `json:"telegram_file_unique_id,omitempty"`
	CreatedAt            string `json:"created_at"`
}

type MediaFile struct {
	Data     []byte
	Filename string
	Kind     string
	Ext      string
}
