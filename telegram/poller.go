package telegram

import (
	"context"
	"time"
	"villum/middleware"
)

func StartTelegramPoller(stop <-chan struct{}) {
	mode := EffectiveMode()
	if mode != "polling" {
		return
	}
	if GetBotToken() == "" {
		return
	}
	go func() {
		backoff := time.Second
		for {
			select {
			case <-stop:
				return
			default:
			}
			offset := GetUpdateOffset()
			// long poll 25s but http timeout 30s
			updates, err := GetUpdates(offset, 25)
			if err != nil {
				middleware.LogWarn("telegram", "getUpdates failed", "error", err)
				if ra, ok := err.(*RetryAfterError); ok {
					select {
					case <-time.After(ra.After):
					case <-stop:
						return
					}
				} else {
					select {
					case <-time.After(backoff):
					case <-stop:
						return
					}
					if backoff < 30*time.Second {
						backoff *= 2
					}
				}
				continue
			}
			backoff = time.Second
			for _, u := range updates {
				HandleUpdate(context.Background(), u)
				if u.UpdateID+1 > offset {
					offset = u.UpdateID + 1
					_ = SetUpdateOffset(offset)
				}
			}
		}
	}()
}
