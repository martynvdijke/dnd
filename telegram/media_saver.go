package telegram

import (
	"errors"

	"villum/models"
)

// MediaSaveRequest describes an inbound Telegram media file to store.
type MediaSaveRequest struct {
	Name                 string
	Data                 []byte
	OwnerType            string
	OwnerID              int64
	Source               string
	TelegramFileUniqueID string
}

// MediaSaver persists a media file and returns the stored upload.
type MediaSaver func(req MediaSaveRequest) (*models.Upload, error)

var mediaSaver MediaSaver

// SetMediaSaver installs the handler-level media pipeline. Wiring happens in
// app setup because the telegram package cannot import handlers.
func SetMediaSaver(s MediaSaver) { mediaSaver = s }

// SaveMedia routes an inbound media file through the configured saver.
func SaveMedia(req MediaSaveRequest) (*models.Upload, error) {
	if mediaSaver == nil {
		return nil, errors.New("media saver not configured")
	}
	return mediaSaver(req)
}
