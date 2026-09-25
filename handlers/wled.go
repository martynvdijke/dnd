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

// WLED integration: flash Wi-Fi LED controllers on the live table when a scene
// effect plays or a spell is cast. Config lives in the app_settings KV store:
// a master enabled switch plus a JSON list of devices.
const (
	settingWLEDEnabled       = "wled_enabled"
	settingWLEDBaseURL       = "wled_base_url"       // legacy single-device key
	settingWLEDBrightness    = "wled_brightness"     // legacy single-device key
	settingWLEDRestorePreset = "wled_restore_preset" // legacy single-device key
	settingWLEDDevices       = "wled_devices"
	wledHTTPTimeout          = 5 * time.Second
)

type wledDevice struct {
	Name          string `json:"name"`
	BaseURL       string `json:"base_url"`
	Brightness    int    `json:"brightness"`
	RestorePreset int    `json:"restore_preset"`
	Enabled       bool   `json:"enabled"`
}

func normalizeWLEDURL(raw string) string {
	u := strings.TrimRight(strings.TrimSpace(raw), "/")
	if u != "" && !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
		u = "http://" + u
	}
	return u
}

func wledClampBrightness(b int) int {
	if b < 0 {
		return 0
	}
	if b > 255 {
		return 255
	}
	return b
}

// loadWLEDDevices returns the configured devices. When no JSON list exists it
// falls back to the legacy single-device keys written by earlier versions.
func loadWLEDDevices() []wledDevice {
	if v, ok := appSetting(settingWLEDDevices); ok && strings.TrimSpace(v) != "" {
		var devs []wledDevice
		if json.Unmarshal([]byte(v), &devs) == nil {
			out := make([]wledDevice, 0, len(devs))
			for _, d := range devs {
				d.BaseURL = normalizeWLEDURL(d.BaseURL)
				if d.Brightness == 0 {
					d.Brightness = 200
				}
				d.Brightness = wledClampBrightness(d.Brightness)
				if d.Name == "" {
					d.Name = "WLED"
				}
				if d.BaseURL != "" {
					out = append(out, d)
				}
			}
			return out
		}
	}
	// Legacy single-device fallback.
	base := ""
	if v, ok := appSetting(settingWLEDBaseURL); ok {
		base = normalizeWLEDURL(v)
	}
	if base == "" {
		return nil
	}
	bri := 200
	if v, ok := appSetting(settingWLEDBrightness); ok {
		if n, err := strconv.Atoi(v); err == nil {
			bri = n
		}
	}
	restore := 0
	if v, ok := appSetting(settingWLEDRestorePreset); ok {
		if n, err := strconv.Atoi(v); err == nil {
			restore = n
		}
	}
	return []wledDevice{{Name: "WLED", BaseURL: base, Brightness: wledClampBrightness(bri), RestorePreset: restore, Enabled: true}}
}

func wledEnabled() bool {
	v, ok := appSetting(settingWLEDEnabled)
	return ok && (v == "1" || v == "true")
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

func flashWLEDDevice(dev wledDevice, col [3]int, effect string) {
	bri := dev.Brightness
	if bri <= 0 || bri > 255 {
		bri = 200
	}
	_ = wledPost(dev.BaseURL, wledColorPayload(col[0], col[1], col[2], bri, 3))
	if effect == "lightning" {
		// Strobe: dark, then bright again.
		time.Sleep(120 * time.Millisecond)
		_ = wledPost(dev.BaseURL, wledColorPayload(0, 0, 0, 0, 0))
		time.Sleep(90 * time.Millisecond)
		_ = wledPost(dev.BaseURL, wledColorPayload(col[0], col[1], col[2], bri, 0))
	}
	time.Sleep(1400 * time.Millisecond)
	if dev.RestorePreset > 0 {
		_ = wledPost(dev.BaseURL, map[string]any{"on": true, "ps": dev.RestorePreset})
	}
}

// NotifyWLEDEffect flashes every enabled device for a named effect. Safe to
// call from request handlers: no-op when disabled and runs async.
func NotifyWLEDEffect(effect string) {
	col, ok := wledEffectColor(effect)
	if !ok || !wledEnabled() {
		return
	}
	devices := loadWLEDDevices()
	if len(devices) == 0 {
		return
	}
	go func() {
		for _, dev := range devices {
			if dev.Enabled && dev.BaseURL != "" {
				flashWLEDDevice(dev, col, effect)
			}
		}
	}()
}

// ─── Admin settings ───

func GetWLEDSettings(c *gin.Context) {
	devices := loadWLEDDevices()
	if devices == nil {
		devices = []wledDevice{}
	}
	WriteJSON(c, http.StatusOK, gin.H{
		"enabled":    wledEnabled(),
		"devices":    devices,
		"configured": len(devices) > 0,
	})
}

func SaveWLEDSettings(c *gin.Context) {
	var body struct {
		Enabled bool         `json:"enabled"`
		Devices []wledDevice `json:"devices"`
	}
	if !BindOr400(c, &body) {
		return
	}
	clean := make([]wledDevice, 0, len(body.Devices))
	for _, d := range body.Devices {
		d.BaseURL = normalizeWLEDURL(d.BaseURL)
		if d.BaseURL == "" {
			continue
		}
		d.Brightness = wledClampBrightness(d.Brightness)
		if d.Brightness == 0 {
			d.Brightness = 200
		}
		if d.Name == "" {
			d.Name = "WLED"
		}
		clean = append(clean, d)
	}
	raw, _ := json.Marshal(clean)
	enabled := "0"
	if body.Enabled {
		enabled = "1"
	}
	setAppSetting(settingWLEDEnabled, enabled)
	setAppSetting(settingWLEDDevices, string(raw))
	// Clear legacy keys so the list is the single source of truth.
	deleteAppSetting(settingWLEDBaseURL)
	deleteAppSetting(settingWLEDBrightness)
	deleteAppSetting(settingWLEDRestorePreset)
	WriteJSON(c, http.StatusOK, gin.H{"ok": true, "configured": len(clean) > 0})
}

func TestWLED(c *gin.Context) {
	devices := loadWLEDDevices()
	if len(devices) == 0 {
		WriteJSON(c, http.StatusOK, gin.H{"success": false, "message": "No WLED device configured"})
		return
	}
	sent, failed := 0, 0
	for _, dev := range devices {
		if !dev.Enabled || dev.BaseURL == "" {
			continue
		}
		if err := wledPost(dev.BaseURL, wledColorPayload(255, 140, 0, dev.Brightness, 3)); err != nil {
			failed++
		} else {
			sent++
		}
	}
	if sent == 0 {
		WriteJSON(c, http.StatusOK, gin.H{"success": false, "message": "No enabled device responded"})
		return
	}
	WriteJSON(c, http.StatusOK, gin.H{"success": true, "message": fmt.Sprintf("Flashed %d device(s)", sent), "failed": failed})
}
