package telegram

import (
	"fmt"
	"strings"

	"villum/models"
)

// MediaPostResult is the per-upload outcome returned to the API caller.
type MediaPostResult struct {
	UploadID          int64  `json:"upload_id"`
	Status            string `json:"status"`
	TelegramMessageID *int64 `json:"telegram_message_id,omitempty"`
	Error             string `json:"error,omitempty"`
}

// isAlbumKind reports whether an upload may join a media group album. Telegram
// groups photos and videos only, so eligibility follows the send method rather
// than the stored kind: extensions such as svg or tiff are stored as images but
// sent as documents, and those must go out individually.
func isAlbumKind(kind, filename string) bool {
	switch methodForFile(kind, filename) {
	case "photo", "video":
		return true
	default:
		return false
	}
}

// sizeLimit returns the outbound Telegram limit for a kind.
// Photos (images) ≤10 MB, all other media ≤50 MB per task spec.
func sizeLimit(kind string) int {
	if strings.ToLower(kind) == "image" {
		return 10 * 1024 * 1024
	}
	return 50 * 1024 * 1024
}

func sizeErrMsg(kind string, size int) string {
	limit := sizeLimit(kind)
	return fmt.Sprintf("file too large: %d bytes exceeds %d bytes limit for %s", size, limit, kind)
}

// PostMedia sends uploads to a Telegram chat, grouping consecutive photo/video
// items into albums of at most 10 via SendMediaGroup and sending the rest
// individually via SendMedia. It writes a telegram_media_posts row for every
// upload and returns per-upload results.
func PostMedia(targetChatID int64, targetType string, campaignID *int64, uploadIDs []int64, recapID *int64, captionOverride string) ([]MediaPostResult, error) {
	results := make([]MediaPostResult, 0, len(uploadIDs))

	// Load all uploads first to validate existence and sizes.
	type loaded struct {
		uploadID int64
		mf       *models.MediaFile
		loadErr  error
	}
	loadedList := make([]loaded, 0, len(uploadIDs))
	for _, uid := range uploadIDs {
		mf, err := LoadMedia(uid)
		loadedList = append(loadedList, loaded{uploadID: uid, mf: mf, loadErr: err})
	}

	// Process in groups: consecutive album-eligible items batched ≤10.
	i := 0
	for i < len(loadedList) {
		// If this item failed to load, record failed row and move on.
		if loadedList[i].loadErr != nil {
			uid := loadedList[i].uploadID
			rowID, _ := InsertMediaPostPending(uid, campaignID, targetType, targetChatID, "", recapID)
			msg := loadedList[i].loadErr.Error()
			if rowID != 0 {
				_ = UpdateMediaPostFailed(rowID, msg)
			}
			results = append(results, MediaPostResult{UploadID: uid, Status: "failed", Error: msg})
			i++
			continue
		}
		mf := loadedList[i].mf
		kind := mf.Kind

		// Size pre-validation before any Telegram call.
		if len(mf.Data) > sizeLimit(kind) {
			uid := loadedList[i].uploadID
			rowID, _ := InsertMediaPostPending(uid, campaignID, targetType, targetChatID, kind, recapID)
			msg := sizeErrMsg(kind, len(mf.Data))
			if rowID != 0 {
				_ = UpdateMediaPostFailed(rowID, msg)
			}
			results = append(results, MediaPostResult{UploadID: uid, Status: "failed", Error: msg})
			i++
			continue
		}

		if isAlbumKind(mf.Kind, mf.Filename) {
			// Collect consecutive album-eligible items up to 10 that also pass size check.
			group := []loaded{loadedList[i]}
			j := i + 1
			for j < len(loadedList) && len(group) < 10 {
				if loadedList[j].loadErr != nil {
					break
				}
				k2 := loadedList[j].mf.Kind
				if !isAlbumKind(loadedList[j].mf.Kind, loadedList[j].mf.Filename) {
					break
				}
				if len(loadedList[j].mf.Data) > sizeLimit(k2) {
					break // oversized will be handled as individual failed in next iteration
				}
				group = append(group, loadedList[j])
				j++
			}
			if len(group) > 1 {
				// Send as album.
				items := make([]MediaItem, 0, len(group))
				for _, g := range group {
					cap := captionOverride
					if cap == "" {
						cap = g.mf.Filename
					}
					items = append(items, MediaItem{Filename: g.mf.Filename, Data: g.mf.Data, Caption: cap})
				}
				// Insert pending rows for all items in group.
				rowIDs := make([]int64, len(group))
				for idx, g := range group {
					rid, _ := InsertMediaPostPending(g.uploadID, campaignID, targetType, targetChatID, g.mf.Kind, recapID)
					rowIDs[idx] = rid
				}
				firstCap := captionOverride
				if firstCap == "" && len(group) > 0 {
					firstCap = group[0].mf.Filename
				}
				msgIDs, err := SendMediaGroup(targetChatID, firstCap, items)
				if err != nil {
					for idx, g := range group {
						if rowIDs[idx] != 0 {
							_ = UpdateMediaPostFailed(rowIDs[idx], err.Error())
						}
						results = append(results, MediaPostResult{UploadID: g.uploadID, Status: "failed", Error: err.Error()})
					}
				} else {
					for idx, g := range group {
						var mid *int64
						if idx < len(msgIDs) {
							v := msgIDs[idx]
							mid = &v
							if rowIDs[idx] != 0 {
								_ = UpdateMediaPostSent(rowIDs[idx], v)
							}
						} else if rowIDs[idx] != 0 {
							_ = UpdateMediaPostFailed(rowIDs[idx], "missing message id from SendMediaGroup")
						}
						res := MediaPostResult{UploadID: g.uploadID, Status: "sent", TelegramMessageID: mid}
						if mid == nil {
							res.Status = "failed"
							res.Error = "missing message id from SendMediaGroup"
						}
						results = append(results, res)
					}
				}
				i = j
				continue
			}
			// Single album-eligible item: send individually via SendMedia.
			uid := loadedList[i].uploadID
			rowID, _ := InsertMediaPostPending(uid, campaignID, targetType, targetChatID, kind, recapID)
			cap := captionOverride
			if cap == "" {
				cap = mf.Filename
			}
			msgID, err := SendMedia(targetChatID, kind, mf.Filename, cap, mf.Data)
			if err != nil {
				if rowID != 0 {
					_ = UpdateMediaPostFailed(rowID, err.Error())
				}
				results = append(results, MediaPostResult{UploadID: uid, Status: "failed", Error: err.Error()})
			} else {
				if rowID != 0 {
					_ = UpdateMediaPostSent(rowID, msgID)
				}
				m := msgID
				results = append(results, MediaPostResult{UploadID: uid, Status: "sent", TelegramMessageID: &m})
			}
			i++
			continue
		}

		// Non-album kind: send individually.
		uid := loadedList[i].uploadID
		rowID, _ := InsertMediaPostPending(uid, campaignID, targetType, targetChatID, kind, recapID)
		cap := captionOverride
		if cap == "" {
			cap = mf.Filename
		}
		msgID, err := SendMedia(targetChatID, kind, mf.Filename, cap, mf.Data)
		if err != nil {
			if rowID != 0 {
				_ = UpdateMediaPostFailed(rowID, err.Error())
			}
			results = append(results, MediaPostResult{UploadID: uid, Status: "failed", Error: err.Error()})
		} else {
			if rowID != 0 {
				_ = UpdateMediaPostSent(rowID, msgID)
			}
			m := msgID
			results = append(results, MediaPostResult{UploadID: uid, Status: "sent", TelegramMessageID: &m})
		}
		i++
	}
	return results, nil
}
