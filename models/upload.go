package models

type Upload struct {
	ID           int    `json:"id"`
	Hash         string `json:"hash"`
	Ext          string `json:"ext"`
	URL          string `json:"url"`
	ResizedURL   string `json:"resized_url"`
	ThumbnailURL string `json:"thumbnail_url"`
	OwnerType    string `json:"owner_type"`
	OwnerID      int64  `json:"owner_id"`
	CreatedAt    string `json:"created_at"`
}
