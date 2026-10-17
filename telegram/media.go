package telegram

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	tgbot "github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
)

// SendMediaImpl and SendMediaGroupImpl are test seams that bypass the real bot client and are nil in production.
var SendMediaImpl func(chatID int64, kind, filename, caption string, data []byte) (int64, error)
var SendMediaGroupImpl func(chatID int64, caption string, items []MediaItem) ([]int64, error)

// MediaItem is one file in an album.
type MediaItem struct {
	Filename string
	Data     []byte
	Caption  string
}

// InboundMedia is a media attachment parsed off an incoming message.
type InboundMedia struct {
	Kind         string // "photo" | "audio" | "video" | "document" | "voice"
	FileID       string
	FileUniqueID string
	FileName     string
	Caption      string
	MediaGroupID string
	Size         int64
}

// ExtractInboundMedia returns the media attachment of an incoming message, or nil.
func ExtractInboundMedia(msg *tgmodels.Message) *InboundMedia {
	if msg == nil {
		return nil
	}
	caption := msg.Caption
	groupID := msg.MediaGroupID
	if len(msg.Photo) > 0 {
		last := msg.Photo[len(msg.Photo)-1]
		return &InboundMedia{
			Kind:         "photo",
			FileID:       last.FileID,
			FileUniqueID: last.FileUniqueID,
			Size:         int64(last.FileSize),
			Caption:      caption,
			MediaGroupID: groupID,
		}
	}
	if msg.Audio != nil {
		a := msg.Audio
		return &InboundMedia{
			Kind:         "audio",
			FileID:       a.FileID,
			FileUniqueID: a.FileUniqueID,
			FileName:     a.FileName,
			Caption:      caption,
			MediaGroupID: groupID,
			Size:         a.FileSize,
		}
	}
	if msg.Video != nil {
		v := msg.Video
		return &InboundMedia{
			Kind:         "video",
			FileID:       v.FileID,
			FileUniqueID: v.FileUniqueID,
			FileName:     v.FileName,
			Caption:      caption,
			MediaGroupID: groupID,
			Size:         v.FileSize,
		}
	}
	if msg.Document != nil {
		d := msg.Document
		return &InboundMedia{
			Kind:         "document",
			FileID:       d.FileID,
			FileUniqueID: d.FileUniqueID,
			FileName:     d.FileName,
			Caption:      caption,
			MediaGroupID: groupID,
			Size:         d.FileSize,
		}
	}
	if msg.Voice != nil {
		v := msg.Voice
		return &InboundMedia{
			Kind:         "voice",
			FileID:       v.FileID,
			FileUniqueID: v.FileUniqueID,
			Caption:      caption,
			MediaGroupID: groupID,
			Size:         v.FileSize,
		}
	}
	return nil
}

func normalizeCaption(caption, filename string) string {
	c := caption
	if strings.TrimSpace(c) == "" {
		c = filename
	}
	if len([]rune(c)) > 1024 {
		runes := []rune(c)
		c = string(runes[:1024])
	}
	return c
}

func methodForFile(kind, filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp", ".gif":
		return "photo"
	case ".mp3", ".ogg", ".oga", ".opus":
		return "audio"
	case ".mp4", ".webm", ".mov":
		return "video"
	default:
		return "document"
	}
}

// SendMedia sends a single file using the Telegram method matching kind,
// falling back to sendDocument. Returns the Telegram message id.
func SendMedia(chatID int64, kind, filename, caption string, data []byte) (int64, error) {
	if SendMediaImpl != nil {
		return SendMediaImpl(chatID, kind, filename, caption, data)
	}
	caption = normalizeCaption(caption, filename)
	method := methodForFile(kind, filename)
	var msgID int64
	err := withBot(context.Background(), func(ctx context.Context, c *tgbot.Bot) error {
		var msg *tgmodels.Message
		var err error
		switch method {
		case "photo":
			msg, err = c.SendPhoto(ctx, &tgbot.SendPhotoParams{
				ChatID:  chatID,
				Photo:   &tgmodels.InputFileUpload{Filename: filename, Data: bytes.NewReader(data)},
				Caption: caption,
			})
		case "audio":
			msg, err = c.SendAudio(ctx, &tgbot.SendAudioParams{
				ChatID:  chatID,
				Audio:   &tgmodels.InputFileUpload{Filename: filename, Data: bytes.NewReader(data)},
				Caption: caption,
			})
		case "video":
			msg, err = c.SendVideo(ctx, &tgbot.SendVideoParams{
				ChatID:  chatID,
				Video:   &tgmodels.InputFileUpload{Filename: filename, Data: bytes.NewReader(data)},
				Caption: caption,
			})
		default:
			msg, err = c.SendDocument(ctx, &tgbot.SendDocumentParams{
				ChatID:   chatID,
				Document: &tgmodels.InputFileUpload{Filename: filename, Data: bytes.NewReader(data)},
				Caption:  caption,
			})
		}
		if err != nil {
			return err
		}
		if msg != nil {
			msgID = int64(msg.ID)
		}
		return nil
	})
	return msgID, err
}

// SendMediaGroup sends up to 10 items as an album. Returns the message ids.
func SendMediaGroup(chatID int64, caption string, items []MediaItem) ([]int64, error) {
	if SendMediaGroupImpl != nil {
		return SendMediaGroupImpl(chatID, caption, items)
	}
	if len(items) == 0 {
		return nil, errors.New("no media items")
	}
	if len(items) > 10 {
		return nil, errors.New("too many media items: maximum 10")
	}
	var ids []int64
	err := withBot(context.Background(), func(ctx context.Context, c *tgbot.Bot) error {
		media := make([]tgmodels.InputMedia, 0, len(items))
		for i, it := range items {
			var cap string
			if i == 0 {
				if strings.TrimSpace(caption) != "" {
					cap = normalizeCaption(caption, it.Filename)
				} else if strings.TrimSpace(it.Caption) != "" {
					cap = normalizeCaption(it.Caption, it.Filename)
				} else {
					cap = normalizeCaption("", it.Filename)
				}
			} else if strings.TrimSpace(it.Caption) != "" {
				cap = normalizeCaption(it.Caption, it.Filename)
			}
			m := methodForFile("", it.Filename)
			attachRef := fmt.Sprintf("attach://file%d", i)
			switch m {
			case "photo":
				media = append(media, &tgmodels.InputMediaPhoto{
					Media:           attachRef,
					Caption:         cap,
					MediaAttachment: bytes.NewReader(it.Data),
				})
			case "video":
				media = append(media, &tgmodels.InputMediaVideo{
					Media:           attachRef,
					Caption:         cap,
					MediaAttachment: bytes.NewReader(it.Data),
				})
			default:
				media = append(media, &tgmodels.InputMediaDocument{
					Media:           attachRef,
					Caption:         cap,
					MediaAttachment: bytes.NewReader(it.Data),
				})
			}
		}
		msgs, err := c.SendMediaGroup(ctx, &tgbot.SendMediaGroupParams{
			ChatID: chatID,
			Media:  media,
		})
		if err != nil {
			return err
		}
		for _, m := range msgs {
			if m.ID != 0 {
				ids = append(ids, int64(m.ID))
			}
		}
		return nil
	})
	return ids, err
}

// DownloadFile downloads an inbound file's bytes via getFile.
func DownloadFile(ctx context.Context, fileID string) ([]byte, error) {
	var filePath string
	err := withBot(ctx, func(ctx context.Context, c *tgbot.Bot) error {
		f, err := c.GetFile(ctx, &tgbot.GetFileParams{FileID: fileID})
		if err != nil {
			return err
		}
		filePath = f.FilePath
		return nil
	})
	if err != nil {
		return nil, err
	}
	if filePath == "" {
		return nil, errors.New("empty file path")
	}
	base := GetAPIBase()
	if base == "" {
		base = defaultAPIBase
	}
	base = strings.TrimRight(base, "/")
	token := GetBotToken()
	url := fmt.Sprintf("%s/file/bot%s/%s", base, token, filePath)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed: %s", resp.Status)
	}
	const maxSize = 20 * 1024 * 1024
	limited := io.LimitReader(resp.Body, int64(maxSize)+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if len(data) > maxSize {
		return nil, fmt.Errorf("file too large: exceeds 20 MB limit")
	}
	if resp.ContentLength > maxSize {
		return nil, fmt.Errorf("file too large: exceeds 20 MB limit")
	}
	return data, nil
}
