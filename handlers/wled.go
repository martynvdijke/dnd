package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// WLED integration: flash a Wi-Fi LED controller on the live table when a scene
// effect plays or a spell is cast. Config lives in the app_settings KV store.
const (
	settingWLEDEnabled       = "wled_enabled"
	settingWLEDBaseURL       = "wled_base_url"
	settingWLEDBrightness    = "wled_brightness"
	settingWLEDRestorePreset = "wled_restore_preset"
	wledHTTPTimeout          = 5 * time.Second
)

type wledSettings struct {
	Enabled       bool   `json:"enabled"`
	BaseURL       string `json:"base_url"`
	Brightness    int    `json:"brightness"`
	RestorePreset int    `json:"restore_preset"`
}

func loadWLEDSettings() wledSettings {
	s := wledSettings{Brightness: 200}
	if v, ok := appSetting(settingWLEDEnabled); ok {
		s.Enabled = v == "1" || v == "true"
	}
	if v, ok := appSetting(settingWLEDBaseURL); ok {
		s.BaseURL = strings.TrimRight(strings.TrimSpace(v), "/")
	}
	if v, ok := appSetting(settingWLEDBrightness); ok {
		if n, err := strconv.Atoi(v); err == nil {
			s.Brightness = n
		}
	}
	if v, ok := appSetting(settingWLEDRestorePreset); ok {
		if n, err := strconv.Atoi(v); err == nil {
			s.RestorePreset = n
		}
	}
	return s
}

// wledEffectColor maps an effect name to its RGB flash colour.
func wledEffectColor(effect string) ([3]int, bool) {
	switch effect {
	case "fire":
		return [3]int{255, 90, 0}, true
	case "smoke":
		return [3]int{90, 90, 100}, true
	case "lightning":
		return [3]int{200, 220, 255}, true
	case "rain":
		return [3]int{60, 120, 255}, true
	case "snow", "ice", "frost":
		return [3]int{180, 220, 255}, true
	case "darkness":
		return [3]int{10, 0, 20}, true
	case "sparkle", "holy", "heal":
		return [3]int{255, 200, 90}, true
	}
	return [3]int{}, false
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// SpellEffectFor picks an effect preset from a spell name (primary) and school.
func SpellEffectFor(name, school string) string {
	n := strings.ToLower(name)
	switch {
	case containsAny(n, "fire", "flame", "burning", "inferno", "meteor", "hellish", "scorch"):
		return "fire"
	case containsAny(n, "frost", "ice", "cold", "freez", "blizzard"):
		return "snow"
	case containsAny(n, "lightning", "thunder", "shock", "bolt", "storm", "jolt"):
		return "lightning"
	case containsAny(n, "acid", "poison", "toxic", "cloudkill", "venom"):
		return "smoke"
	case containsAny(n, "necro", "blight", "death", "vampir", "wither", "shadow", "dark", "doom"):
		return "darkness"
	case containsAny(n, "radiant", "heal", "cure", "restor", "holy", "bless", "divine", "sacred"):
		return "sparkle"
	case containsAny(n, "psychic", "charm", "mind", "illusion", "faerie", "hypnotic", "sleep"):
		return "sparkle"
	}
	switch strings.ToLower(school) {
	case "necromancy":
		return "darkness"
	case "conjuration":
		return "smoke"
	}
	return "sparkle"
}

func wledPost(baseURL string, payload any) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: wledHTTPTimeout}
	req, err := http.NewRequest(http.MethodPost, baseURL+"/json/state", bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("wled returned %d", resp.StatusCode)
	}
	return nil
}

func wledColorPayload(r, g, b, bri, transition int) map[string]any {
	return map[string]any{
		"on":         true,
		"bri":        bri,
		"transition": transition,
		"seg":        []map[string]any{{"col": [][3]int{{r, g, b}, {0, 0, 0}}}},
	}
}

// NotifyWLEDEffect flashes the configured strip for a named effect. Safe to call
// from request handlers: it is a no-op when unconfigured and runs async.
func NotifyWLEDEffect(effect string) {
	col, ok := wledEffectColor(effect)
	if !ok {
		return
	}
	cfg := loadWLEDSettings()
	if !cfg.Enabled || cfg.BaseURL == "" {
		return
	}
	bri := cfg.Brightness
	if bri <= 0 || bri > 255 {
		bri = 200
	}
	go func() {
		_ = wledPost(cfg.BaseURL, wledColorPayload(col[0], col[1], col[2], bri, 3))
		if effect == "lightning" {
			// Strobe: dark, then bright again.
			time.Sleep(120 * time.Millisecond)
			_ = wledPost(cfg.BaseURL, wledColorPayload(0, 0, 0, 0, 0))
			time.Sleep(90 * time.Millisecond)
			_ = wledPost(cfg.BaseURL, wledColorPayload(col[0], col[1], col[2], bri, 0))
		}
		time.Sleep(1400 * time.Millisecond)
		if cfg.RestorePreset > 0 {
			_ = wledPost(cfg.BaseURL, map[string]any{"on": true, "ps": cfg.RestorePreset})
		}
	}()
}

// ─── Admin settings ───

func GetWLEDSettings(c *gin.Context) {
	s := loadWLEDSettings()
	WriteJSON(c, http.StatusOK, gin.H{
		"enabled":        s.Enabled,
		"base_url":       s.BaseURL,
		"brightness":     s.Brightness,
		"restore_preset": s.RestorePreset,
		"configured":     s.BaseURL != "",
	})
}

func SaveWLEDSettings(c *gin.Context) {
	var body struct {
		Enabled       bool   `json:"enabled"`
		BaseURL       string `json:"base_url"`
		Brightness    int    `json:"brightness"`
		RestorePreset int    `json:"restore_preset"`
	}
	if !BindOr400(c, &body) {
		return
	}
	base := strings.TrimRight(strings.TrimSpace(body.BaseURL), "/")
	if base != "" && !strings.HasPrefix(base, "http://") && !strings.HasPrefix(base, "https://") {
		base = "http://" + base
	}
	if body.Brightness < 0 {
		body.Brightness = 0
	}
	if body.Brightness > 255 {
		body.Brightness = 255
	}
	enabled := "0"
	if body.Enabled {
		enabled = "1"
	}
	setAppSetting(settingWLEDEnabled, enabled)
	setAppSetting(settingWLEDBaseURL, base)
	setAppSetting(settingWLEDBrightness, strconv.Itoa(body.Brightness))
	setAppSetting(settingWLEDRestorePreset, strconv.Itoa(body.RestorePreset))
	WriteJSON(c, http.StatusOK, gin.H{"ok": true, "configured": base != ""})
}

func TestWLED(c *gin.Context) {
	cfg := loadWLEDSettings()
	if cfg.BaseURL == "" {
		WriteJSON(c, http.StatusOK, gin.H{"success": false, "message": "No WLED base URL configured"})
		return
	}
	if err := wledPost(cfg.BaseURL, wledColorPayload(255, 140, 0, 200, 3)); err != nil {
		WriteJSON(c, http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	WriteJSON(c, http.StatusOK, gin.H{"success": true, "message": "Sent test colour to WLED"})
}
