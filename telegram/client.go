package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var httpClient = &http.Client{Timeout: 30 * time.Second}
var Version = "0.0.0-dev"

type tgResponse struct {
	OK          bool            `json:"ok"`
	Result      json.RawMessage `json:"result"`
	ErrorCode   int             `json:"error_code"`
	Description string          `json:"description"`
	Parameters  *struct {
		RetryAfter int `json:"retry_after"`
	} `json:"parameters"`
}

type Update struct {
	UpdateID int64    `json:"update_id"`
	Message  *Message `json:"message"`
}
type Message struct {
	MessageID int64  `json:"message_id"`
	From      *User  `json:"from"`
	Chat      *Chat  `json:"chat"`
	Text      string `json:"text"`
}
type User struct {
	ID        int64  `json:"id"`
	IsBot     bool   `json:"is_bot"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
}
type Chat struct {
	ID    int64  `json:"id"`
	Type  string `json:"type"`
	Title string `json:"title"`
}

func apiURL(method string) string {
	token := GetBotToken()
	base := GetAPIBase()
	return fmt.Sprintf("%s/bot%s/%s", base, token, method)
}

func doPost(method string, payload any, result any) error {
	token := GetBotToken()
	if token == "" {
		return fmt.Errorf("telegram not configured")
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", apiURL(method), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "villum/"+Version)
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	var r tgResponse
	if err := json.Unmarshal(data, &r); err != nil {
		return fmt.Errorf("telegram decode: %w", err)
	}
	if !r.OK {
		if r.Parameters != nil && r.Parameters.RetryAfter > 0 {
			return &RetryAfterError{After: time.Duration(r.Parameters.RetryAfter) * time.Second, Desc: r.Description}
		}
		if resp.StatusCode == 429 && r.Parameters != nil {
			return &RetryAfterError{After: time.Duration(r.Parameters.RetryAfter) * time.Second, Desc: r.Description}
		}
		return fmt.Errorf("telegram %s: %s (code %d)", method, r.Description, r.ErrorCode)
	}
	if result != nil && len(r.Result) > 0 {
		if err := json.Unmarshal(r.Result, result); err != nil {
			return err
		}
	}
	return nil
}

type RetryAfterError struct {
	After time.Duration
	Desc  string
}

func (e *RetryAfterError) Error() string { return fmt.Sprintf("retry after %s: %s", e.After, e.Desc) }

func GetUpdates(offset int64, timeout int) ([]Update, error) {
	var result []Update
	err := doPost("getUpdates", map[string]any{"offset": offset, "timeout": timeout}, &result)
	return result, err
}
func SendMessage(chatID int64, text string) (int64, error) {
	var res struct {
		MessageID int64 `json:"message_id"`
	}
	err := doPost("sendMessage", map[string]any{"chat_id": chatID, "text": text, "disable_web_page_preview": true}, &res)
	if err != nil {
		return 0, err
	}
	return res.MessageID, nil
}
func SendDocument(chatID int64, filename, content string) (int64, error) {
	token := GetBotToken()
	if token == "" {
		return 0, fmt.Errorf("telegram not configured")
	}
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	_ = w.WriteField("chat_id", fmt.Sprintf("%d", chatID))
	fw, _ := w.CreateFormFile("document", filename)
	_, _ = io.WriteString(fw, content)
	w.Close()
	req, err := http.NewRequest("POST", apiURL("sendDocument"), &b)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("User-Agent", "villum/"+Version)
	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	var r tgResponse
	if err := json.Unmarshal(data, &r); err != nil {
		return 0, err
	}
	if !r.OK {
		return 0, fmt.Errorf("sendDocument: %s", r.Description)
	}
	var res struct {
		MessageID int64 `json:"message_id"`
	}
	_ = json.Unmarshal(r.Result, &res)
	return res.MessageID, nil
}
func SetWebhook(url, secret string) error {
	return doPost("setWebhook", map[string]any{"url": url, "secret_token": secret}, nil)
}
func DeleteWebhook() error { return doPost("deleteWebhook", map[string]any{}, nil) }
func GetChat(chatID int64) (*Chat, error) {
	var ch Chat
	if err := doPost("getChat", map[string]any{"chat_id": chatID}, &ch); err != nil {
		return nil, err
	}
	return &ch, nil
}

func GetMe() (*User, error) {
	var u User
	if err := doPost("getMe", map[string]any{}, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

var (
	botUsernameCache string
	botUsernameOnce  sync.Once
)

func GetBotUsername() string {
	if v := os.Getenv("TELEGRAM_BOT_USERNAME"); v != "" {
		return v
	}
	botUsernameOnce.Do(func() {
		if u, err := GetMe(); err == nil && u.Username != "" {
			botUsernameCache = u.Username
		}
	})
	return botUsernameCache
}

// ResetBotUsernameCache is for tests.
func ResetBotUsernameCache() {
	botUsernameOnce = sync.Once{}
	botUsernameCache = ""
}

func runeCount(s string) int { return len([]rune(s)) }

func runeSlice(s string, start, end int) string {
	r := []rune(s)
	if start >= len(r) {
		return ""
	}
	if end > len(r) {
		end = len(r)
	}
	return string(r[start:end])
}

func chunkMessage(text string, limit int) []string {
	if runeCount(text) <= limit {
		return []string{text}
	}
	paras := strings.Split(text, "\n\n")
	var chunks []string
	var cur string
	curLen := 0
	flushCur := func() {
		if cur != "" {
			chunks = append(chunks, cur)
			cur = ""
			curLen = 0
		}
	}
	for _, ps := range paras {
		psLen := runeCount(ps)
		sepLen := 0
		if curLen > 0 {
			sepLen = 2
		}
		if curLen+sepLen+psLen <= limit {
			if cur != "" {
				cur += "\n\n" + ps
			} else {
				cur = ps
			}
			curLen += sepLen + psLen
		} else {
			if cur != "" {
				flushCur()
			}
			if psLen <= limit {
				cur = ps
				curLen = psLen
			} else {
				lines := strings.Split(ps, "\n")
				var lcur string
				lcurLen := 0
				for _, ls := range lines {
					lsLen := runeCount(ls)
					sep := 0
					if lcurLen > 0 {
						sep = 1
					}
					if lcurLen+sep+lsLen <= limit {
						if lcur != "" {
							lcur += "\n" + ls
						} else {
							lcur = ls
						}
						lcurLen += sep + lsLen
					} else {
						if lcur != "" {
							chunks = append(chunks, lcur)
							lcur = ""
							lcurLen = 0
						}
						// hard cut on rune boundaries
						for runeCount(ls) > limit {
							chunks = append(chunks, runeSlice(ls, 0, limit))
							ls = runeSlice(ls, limit, runeCount(ls))
						}
						lcur = ls
						lcurLen = runeCount(ls)
					}
				}
				if lcur != "" {
					cur = lcur
					curLen = lcurLen
				}
			}
		}
	}
	if cur != "" {
		chunks = append(chunks, cur)
	}
	return chunks
}
