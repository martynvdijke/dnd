package handlers

import (
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"errors"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
)

const trmnlTokenKey = "trmnl_token"

// getOrCreateTRMNLToken returns the site-wide TRMNL polling token, generating
// and persisting a new one on first read.
func getOrCreateTRMNLToken() (string, error) {
	var token string
	err := db.DB.QueryRow("SELECT value FROM app_settings WHERE key = ?", trmnlTokenKey).Scan(&token)
	if err == nil && token != "" {
		return token, nil
	}
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token = hex.EncodeToString(b)
	if _, err := db.DB.Exec("INSERT OR REPLACE INTO app_settings (key, value) VALUES (?, ?)", trmnlTokenKey, token); err != nil {
		return "", err
	}
	return token, nil
}

// regenerateTRMNLToken generates a fresh token, replacing the stored value.
func regenerateTRMNLToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)
	if _, err := db.DB.Exec("INSERT OR REPLACE INTO app_settings (key, value) VALUES (?, ?)", trmnlTokenKey, token); err != nil {
		return "", err
	}
	return token, nil
}

// trmnlTokenValid reports whether the request's `token` query param matches
// the stored TRMNL token, using a constant-time comparison.
func trmnlTokenValid(c *gin.Context) bool {
	stored, err := getOrCreateTRMNLToken()
	if err != nil {
		return false
	}
	given := c.Query("token")
	if given == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(given), []byte(stored)) == 1
}

// GetTRMNLSettings returns the current TRMNL token (auto-creating on first read).
func GetTRMNLSettings(c *gin.Context) {
	token, err := getOrCreateTRMNLToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read TRMNL token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

// SetTRMNLSettings persists TRMNL settings. A request body of
// {"regenerate": true} replaces the stored token with a new one.
func SetTRMNLSettings(c *gin.Context) {
	var req struct {
		Regenerate bool `json:"regenerate"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if !req.Regenerate {
		token, err := getOrCreateTRMNLToken()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read TRMNL token"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"token": token})
		return
	}
	token, err := regenerateTRMNLToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to regenerate TRMNL token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

type TRMNLCharacterStats struct {
	Name       string `json:"name"`
	Race       string `json:"race"`
	Class      string `json:"class"`
	Subclass   string `json:"subclass"`
	Level      int    `json:"level"`
	XP         int    `json:"xp"`
	HPCurrent  int    `json:"hp_current"`
	HPMax      int    `json:"hp_max"`
	AC         int    `json:"ac"`
	Initiative int    `json:"initiative"`
	Str        int    `json:"str"`
	Dex        int    `json:"dex"`
	Con        int    `json:"con"`
	Int        int    `json:"int"`
	Wis        int    `json:"wis"`
	Cha        int    `json:"cha"`
	StrMod     int    `json:"str_mod"`
	DexMod     int    `json:"dex_mod"`
	ConMod     int    `json:"con_mod"`
	IntMod     int    `json:"int_mod"`
	WisMod     int    `json:"wis_mod"`
	ChaMod     int    `json:"cha_mod"`
}

// abilityModifier computes the D&D ability modifier for a score
// (floor((score-10)/2), per the 5e rulebook).
func abilityModifier(score int) int {
	return int(math.Floor(float64(score-10) / 2))
}

// loadTRMNLCharacterStats loads a character's core row for the TRMNL endpoint.
func loadTRMNLCharacterStats(charID int64) (TRMNLCharacterStats, error) {
	var s TRMNLCharacterStats
	err := db.DB.QueryRow(`
		SELECT name, race, class, subclass, level, xp, hp_current, hp_max, ac, initiative,
		       str, dex, con, int, wis, cha
		FROM characters WHERE id = ?`, charID).Scan(
		&s.Name, &s.Race, &s.Class, &s.Subclass, &s.Level, &s.XP, &s.HPCurrent, &s.HPMax, &s.AC, &s.Initiative,
		&s.Str, &s.Dex, &s.Con, &s.Int, &s.Wis, &s.Cha)
	if err != nil {
		return s, err
	}
	s.StrMod = abilityModifier(s.Str)
	s.DexMod = abilityModifier(s.Dex)
	s.ConMod = abilityModifier(s.Con)
	s.IntMod = abilityModifier(s.Int)
	s.WisMod = abilityModifier(s.Wis)
	s.ChaMod = abilityModifier(s.Cha)
	return s, nil
}

// GetTRMNLCharacterStats is a public polling endpoint returning a character's
// core stats (ability scores, HP/AC, level, XP) for TRMNL e-ink displays.
func GetTRMNLCharacterStats(c *gin.Context) {
	if !trmnlTokenValid(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	charID, err := strconv.ParseInt(c.Query("character_id"), 10, 64)
	if err != nil || charID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid character_id"})
		return
	}
	stats, err := loadTRMNLCharacterStats(charID)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load character"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// TRMNLCampaignStats embeds the campaign progress stats alongside the core
// character fields so a single payload serves both template blocks.
// Level is intentionally not redeclared: it comes from the embedded
// CharacterStats (json "level").
type TRMNLCampaignStats struct {
	CharacterStats
	Name       string `json:"name"`
	Race       string `json:"race"`
	Class      string `json:"class"`
	Subclass   string `json:"subclass"`
	XP         int    `json:"xp"`
	HPCurrent  int    `json:"hp_current"`
	HPMax      int    `json:"hp_max"`
	AC         int    `json:"ac"`
	Initiative int    `json:"initiative"`
	Str        int    `json:"str"`
	Dex        int    `json:"dex"`
	Con        int    `json:"con"`
	Int        int    `json:"int"`
	Wis        int    `json:"wis"`
	Cha        int    `json:"cha"`
	StrMod     int    `json:"str_mod"`
	DexMod     int    `json:"dex_mod"`
	ConMod     int    `json:"con_mod"`
	IntMod     int    `json:"int_mod"`
	WisMod     int    `json:"wis_mod"`
	ChaMod     int    `json:"cha_mod"`
}

// GetTRMNLCampaignStats is a public polling endpoint returning campaign
// progress stats for a character (sessions, XP/gold, quests, rests, NPCs).
func GetTRMNLCampaignStats(c *gin.Context) {
	if !trmnlTokenValid(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	charID, err := strconv.ParseInt(c.Query("character_id"), 10, 64)
	if err != nil || charID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid character_id"})
		return
	}
	charStats, err := loadTRMNLCharacterStats(charID)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load character"})
		return
	}
	campaign, err := loadCharacterStats(db.DB, charID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load campaign stats"})
		return
	}
	c.JSON(http.StatusOK, mergeTRMNLCampaignStats(campaign, charStats))
}

// TRMNLCharacterRosterEntry is one character row in the party roster payload.
// It embeds the core character stats and adds the id and character type.
type TRMNLCharacterRosterEntry struct {
	ID            int64  `json:"id"`
	CharacterType string `json:"character_type"`
	TRMNLCharacterStats
}

// GetTRMNLCharacterRoster is a public polling endpoint returning the full
// party roster — every character with their core stats — for TRMNL displays.
func GetTRMNLCharacterRoster(c *gin.Context) {
	if !trmnlTokenValid(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	rows, err := db.DB.Query(`
		SELECT id, name, race, class, subclass, level, xp, hp_current, hp_max,
		       ac, initiative, str, dex, con, int, wis, cha,
		       COALESCE(character_type, 'player')
		FROM characters ORDER BY name`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load characters"})
		return
	}
	defer rows.Close()

	roster := make([]TRMNLCharacterRosterEntry, 0)
	for rows.Next() {
		var e TRMNLCharacterRosterEntry
		if err := rows.Scan(
			&e.ID, &e.Name, &e.Race, &e.Class, &e.Subclass, &e.Level, &e.XP,
			&e.HPCurrent, &e.HPMax, &e.AC, &e.Initiative,
			&e.Str, &e.Dex, &e.Con, &e.Int, &e.Wis, &e.Cha,
			&e.CharacterType,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load characters"})
			return
		}
		e.StrMod = abilityModifier(e.Str)
		e.DexMod = abilityModifier(e.Dex)
		e.ConMod = abilityModifier(e.Con)
		e.IntMod = abilityModifier(e.Int)
		e.WisMod = abilityModifier(e.Wis)
		e.ChaMod = abilityModifier(e.Cha)
		roster = append(roster, e)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load characters"})
		return
	}
	c.JSON(http.StatusOK, roster)
}

// mergeTRMNLCampaignStats combines campaign progress stats with the core
// character fields into a single flat payload.
func mergeTRMNLCampaignStats(campaign CharacterStats, char TRMNLCharacterStats) TRMNLCampaignStats {
	return TRMNLCampaignStats{
		CharacterStats: campaign,
		Name:           char.Name,
		Race:           char.Race,
		Class:          char.Class,
		Subclass:       char.Subclass,
		XP:             char.XP,
		HPCurrent:      char.HPCurrent,
		HPMax:          char.HPMax,
		AC:             char.AC,
		Initiative:     char.Initiative,
		Str:            char.Str,
		Dex:            char.Dex,
		Con:            char.Con,
		Int:            char.Int,
		Wis:            char.Wis,
		Cha:            char.Cha,
		StrMod:         char.StrMod,
		DexMod:         char.DexMod,
		ConMod:         char.ConMod,
		IntMod:         char.IntMod,
		WisMod:         char.WisMod,
		ChaMod:         char.ChaMod,
	}
}
