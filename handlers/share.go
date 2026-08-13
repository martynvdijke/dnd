package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

type ShareLinkRequest struct {
	EntityType string `json:"entity_type" binding:"required"`
	EntityID   int64  `json:"entity_id" binding:"required"`
	ExpiresIn  *int   `json:"expires_in,omitempty"`
}

type ShareLinkResponse struct {
	Token      string `json:"token"`
	URL        string `json:"url"`
	EntityType string `json:"entity_type"`
	EntityID   int64  `json:"entity_id"`
	Label      string `json:"label,omitempty"`
	CreatedAt  string `json:"created_at"`
	ExpiresAt  string `json:"expires_at,omitempty"`
}

// supportedShareTypes are the entity types that can be shared via share_links.
var supportedShareTypes = map[string]bool{
	"character": true,
	"party":     true,
	"note":      true,
	"journal":   true,
	"map":       true,
	"upload":    true,
}

func generateToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// userCanAccessCharacter reports whether the user owns the character, is a DM
// of its campaign, or is an admin (mirrors checkCharacterAccess).
func userCanAccessCharacter(c *gin.Context, characterID int64) bool {
	return checkCharacterAccess(c, characterID)
}

// userCanAccessCampaign reports whether the user owns the campaign, is a
// member of it, or is an admin.
func userCanAccessCampaign(c *gin.Context, campaignID int64) bool {
	role, _ := c.Get("role")
	if role == "admin" {
		return true
	}
	userID, _ := c.Get("user_id")
	uid, _ := userID.(int64)
	var ownerID int64
	err := db.DB.QueryRow("SELECT user_id FROM campaigns WHERE id=?", campaignID).Scan(&ownerID)
	if err != nil {
		return false
	}
	if ownerID == uid {
		return true
	}
	var count int
	db.DB.QueryRow("SELECT COUNT(*) FROM campaign_members WHERE campaign_id=? AND user_id=?", campaignID, uid).Scan(&count)
	return count > 0
}

// canShareEntity verifies the user may create a share link for the entity.
// Admins bypass all ownership checks. Unsupported types are denied.
func canShareEntity(c *gin.Context, entityType string, entityID int64) bool {
	role, _ := c.Get("role")
	if role == "admin" {
		return true
	}
	userID, _ := c.Get("user_id")
	uid, _ := userID.(int64)

	switch entityType {
	case "character":
		return checkCharacterAccess(c, entityID)
	case "party":
		return userCanAccessCampaign(c, entityID)
	case "note":
		var charID int64
		err := db.DB.QueryRow("SELECT character_id FROM character_notes WHERE id=?", entityID).Scan(&charID)
		return err == nil && checkCharacterAccess(c, charID)
	case "journal":
		var charID int64
		err := db.DB.QueryRow("SELECT character_id FROM journal WHERE id=?", entityID).Scan(&charID)
		return err == nil && checkCharacterAccess(c, charID)
	case "map":
		var campaignID int64
		err := db.DB.QueryRow("SELECT campaign_id FROM campaign_maps WHERE id=?", entityID).Scan(&campaignID)
		return err == nil && userCanAccessCampaign(c, campaignID)
	case "upload":
		return canShareUpload(uid, entityID)
	}
	return false
}

// canShareUpload resolves upload ownership through owner_type/owner_id.
// Legacy uploads with empty owner_type are admin-only (no ownership chain).
func canShareUpload(uid int64, uploadID int64) bool {
	var ownerType string
	var ownerID int64
	err := db.DB.QueryRow("SELECT owner_type, owner_id FROM uploads WHERE id=?", uploadID).Scan(&ownerType, &ownerID)
	if err != nil {
		return false
	}
	switch ownerType {
	case "":
		return false // no ownership chain → admin only
	}
	// Per-owner-type ownership resolution.
	switch ownerType {
	case "character":
		var n int
		db.DB.QueryRow("SELECT COUNT(*) FROM characters WHERE id=? AND user_id=?", ownerID, uid).Scan(&n)
		return n > 0
	case "npc":
		var n int
		db.DB.QueryRow("SELECT COUNT(*) FROM npcs WHERE id=? AND user_id=?", ownerID, uid).Scan(&n)
		return n > 0
	case "campaign", "party":
		var n int
		db.DB.QueryRow(`SELECT COUNT(*) FROM campaigns WHERE id=? AND (user_id=? OR id IN (SELECT campaign_id FROM campaign_members WHERE user_id=?))`, ownerID, uid, uid).Scan(&n)
		return n > 0
	case "oneshot":
		var n int
		db.DB.QueryRow("SELECT COUNT(*) FROM oneshot_adventures WHERE id=? AND user_id=?", ownerID, uid).Scan(&n)
		return n > 0
	case "item":
		var n int
		db.DB.QueryRow("SELECT COUNT(*) FROM oneshot_items oi JOIN oneshot_adventures oa ON oa.id=oi.adventure_id WHERE oi.id=? AND oa.user_id=?", ownerID, uid).Scan(&n)
		return n > 0
	}
	return false
}

func CreateShareLink(c *gin.Context) {
	var req ShareLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !supportedShareTypes[req.EntityType] {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "unsupported entity_type"})
		return
	}

	if !canShareEntity(c, req.EntityType, req.EntityID) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	userID, _ := c.Get("user_id")
	token := generateToken()

	var expiresAt *string
	if req.ExpiresIn != nil && *req.ExpiresIn > 0 {
		exp := time.Now().Add(time.Duration(*req.ExpiresIn) * time.Hour).UTC().Format("2006-01-02 15:04:05")
		expiresAt = &exp
	}

	_, err := db.DB.Exec(
		"INSERT INTO share_links (token, entity_type, entity_id, created_by, expires_at) VALUES (?, ?, ?, ?, ?)",
		token, req.EntityType, req.EntityID, userID, expiresAt)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	protocol := "http"
	if c.Request.TLS != nil {
		protocol = "https"
	}
	url := fmt.Sprintf("%s://%s/api/share/%s", protocol, c.Request.Host, token)

	resp := ShareLinkResponse{
		Token:      token,
		URL:        url,
		EntityType: req.EntityType,
		EntityID:   req.EntityID,
		Label:      shareEntityLabel(req.EntityType, req.EntityID),
		CreatedAt:  time.Now().UTC().Format("2006-01-02 15:04:05"),
	}
	if expiresAt != nil {
		resp.ExpiresAt = *expiresAt
	}

	c.JSON(http.StatusCreated, resp)
}

type SharedCharacterView struct {
	models.Character
	OwnerName string `json:"owner_name"`
}

// SharedNoteView carries note fields into the public share page template.
type SharedNoteView struct {
	Title      string
	Content    string
	Visibility string
	Category   string
	CreatedAt  string
	OwnerName  string
}

// SharedJournalView carries journal entry fields into the public share page template.
type SharedJournalView struct {
	Title     string
	Entry     string
	EntryDate string
	CreatedAt string
	OwnerName string
}

// shareEntityLabel returns a human-readable name for a shared entity, used in
// share-link listings and creation responses.
func shareEntityLabel(entityType string, entityID int64) string {
	var label string
	switch entityType {
	case "character":
		db.DB.QueryRow("SELECT name FROM characters WHERE id=?", entityID).Scan(&label)
	case "party":
		db.DB.QueryRow("SELECT name FROM campaigns WHERE id=?", entityID).Scan(&label)
	case "note":
		db.DB.QueryRow("SELECT title FROM character_notes WHERE id=?", entityID).Scan(&label)
	case "journal":
		db.DB.QueryRow("SELECT title FROM journal WHERE id=?", entityID).Scan(&label)
	case "map":
		db.DB.QueryRow("SELECT name FROM campaign_maps WHERE id=?", entityID).Scan(&label)
	case "upload":
		db.DB.QueryRow("SELECT url FROM uploads WHERE id=?", entityID).Scan(&label)
	}
	return label
}

// lookupShare resolves the entity_type/entity_id/expiry for a token, returning
// false if the token is unknown or expired. expired=true distinguishes 410
// from 404.
func lookupShare(token string) (entityType string, entityID int64, expired bool, ok bool) {
	var expiresAt sql.NullString
	err := db.DB.QueryRow("SELECT entity_type, entity_id, expires_at FROM share_links WHERE token=?", token).
		Scan(&entityType, &entityID, &expiresAt)
	if err != nil {
		return "", 0, false, false
	}
	if expiresAt.Valid && expiresAt.String != "" {
		expTime, err := time.Parse("2006-01-02 15:04:05", expiresAt.String)
		if err == nil && time.Now().UTC().After(expTime) {
			return entityType, entityID, true, true
		}
	}
	return entityType, entityID, false, true
}

func GetSharedEntity(c *gin.Context) {
	entityType, entityID, expired, ok := lookupShare(c.Param("token"))
	if !ok {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "share link not found"})
		return
	}
	if expired {
		c.AbortWithStatusJSON(http.StatusGone, gin.H{"error": "share link has expired"})
		return
	}

	switch entityType {
	case "character":
		ch := &models.Character{}
		var ownerName string
		err := db.DB.QueryRow(`
			SELECT c.id, c.user_id, c.campaign_id, c.name, c.race, c.class, c.subclass, c.level, c.xp, c.background, c.alignment,
				c.str, c.dex, c.con, c.int, c.wis, c.cha, c.ac, c.initiative, c.speed,
				c.hp_max, c.hp_current, c.temp_hp, c.hit_dice, c.hit_dice_current,
				c.proficiency_bonus, c.inspiration, c.passive_perception,
				c.death_saves_successes, c.death_saves_failures, c.concentrating_on,
				c.personality_traits, c.ideals, c.bonds, c.flaws, c.appearance, c.backstory,
				c.portrait_url, c.created_at, c.updated_at,
				COALESCE(u.username, '')
			FROM characters c LEFT JOIN users u ON u.id = c.user_id
			WHERE c.id=?`, entityID).Scan(
			&ch.ID, &ch.UserID, &ch.CampaignID, &ch.Name, &ch.Race, &ch.Class, &ch.Subclass, &ch.Level, &ch.XP,
			&ch.Background, &ch.Alignment,
			&ch.Str, &ch.Dex, &ch.Con, &ch.Int, &ch.Wis, &ch.Cha,
			&ch.AC, &ch.Initiative, &ch.Speed,
			&ch.HPMax, &ch.HPCurrent, &ch.TempHP, &ch.HitDice, &ch.HitDiceCurrent,
			&ch.ProficiencyBonus, &ch.Inspiration, &ch.PassivePerception,
			&ch.DeathSavesSuccesses, &ch.DeathSavesFailures, &ch.ConcentratingOn,
			&ch.PersonalityTraits, &ch.Ideals, &ch.Bonds, &ch.Flaws, &ch.Appearance, &ch.Backstory,
			&ch.PortraitURL, &ch.CreatedAt, &ch.UpdatedAt,
			&ownerName)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "character not found"})
			return
		}

		ctx := c.Request.Context()
		ch.Proficiencies = loadProficiencies(ctx, ch.ID)
		ch.Features = loadFeatures(ctx, ch.ID)
		ch.Spellcasting = loadSpellcasting(ctx, ch.ID)
		ch.Spells = loadSpells(ctx, ch.ID)
		ch.Inventory = loadInventory(ctx, ch.ID)
		ch.Currency = loadCurrency(ctx, ch.ID)
		ch.Classes = loadCharClasses(ctx, ch.ID)
		computeMods(ch)

		c.JSON(http.StatusOK, SharedCharacterView{Character: *ch, OwnerName: ownerName})

	case "party":
		campaignID := entityID
		var campName, partyName, ownerName string
		var ownerID int64
		err := db.DB.QueryRow(`
			SELECT c.name, COALESCE(c.party_name, ''), COALESCE(u.username, ''), c.user_id
			FROM campaigns c JOIN users u ON u.id = c.user_id WHERE c.id=?`, campaignID).
			Scan(&campName, &partyName, &ownerName, &ownerID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "campaign not found"})
			return
		}

		rows, err := db.DB.Query(`
			SELECT c.id, c.user_id, COALESCE(u.username, ''), c.name, c.race, c.class,
				c.level, c.ac, c.hp_max, c.hp_current, c.temp_hp, c.campaign_id
			FROM characters c LEFT JOIN users u ON u.id = c.user_id
			WHERE c.campaign_id=? ORDER BY c.name`, campaignID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var members []PartyMember
		for rows.Next() {
			var pm PartyMember
			var cid *int64
			rows.Scan(&pm.ID, &pm.UserID, &pm.OwnerName, &pm.Name, &pm.Race, &pm.Class,
				&pm.Level, &pm.AC, &pm.HPMax, &pm.HPCurrent, &pm.TempHP, &cid)
			pm.CampaignID = cid
			pm.Status = "alive"
			if pm.HPCurrent <= 0 {
				pm.Status = "down"
			} else if float64(pm.HPCurrent)/float64(pm.HPMax) < 0.25 {
				pm.Status = "injured"
			}
			members = append(members, pm)
		}

		c.JSON(http.StatusOK, gin.H{
			"campaign": gin.H{
				"id":         campaignID,
				"name":       campName,
				"party_name": partyName,
				"owner":      ownerName,
			},
			"members": members,
		})

	case "note":
		var note struct {
			ID            int64  `json:"id"`
			Title         string `json:"title"`
			Content       string `json:"content"`
			Visibility    string `json:"visibility"`
			Category      string `json:"category"`
			CreatedAt     string `json:"created_at"`
			CharacterID   int64  `json:"character_id"`
			CharacterName string `json:"character_name"`
			OwnerName     string `json:"owner_name"`
		}
		err := db.DB.QueryRow(`
			SELECT cn.id, cn.title, cn.content, cn.visibility, cn.category, COALESCE(cn.created_at,''),
				cn.character_id, COALESCE(ch.name,''), COALESCE(u.username,'')
			FROM character_notes cn
			JOIN characters ch ON ch.id = cn.character_id
			LEFT JOIN users u ON u.id = ch.user_id
			WHERE cn.id=?`, entityID).
			Scan(&note.ID, &note.Title, &note.Content, &note.Visibility, &note.Category, &note.CreatedAt,
				&note.CharacterID, &note.CharacterName, &note.OwnerName)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "note not found"})
			return
		}
		c.JSON(http.StatusOK, note)

	case "journal":
		var entry struct {
			ID            int64  `json:"id"`
			Title         string `json:"title"`
			Entry         string `json:"entry"`
			EntryDate     string `json:"entry_date"`
			CreatedAt     string `json:"created_at"`
			CharacterID   int64  `json:"character_id"`
			CharacterName string `json:"character_name"`
			OwnerName     string `json:"owner_name"`
		}
		err := db.DB.QueryRow(`
			SELECT j.id, j.title, j.entry, COALESCE(j.entry_date,''), COALESCE(j.created_at,''),
				j.character_id, COALESCE(ch.name,''), COALESCE(u.username,'')
			FROM journal j
			JOIN characters ch ON ch.id = j.character_id
			LEFT JOIN users u ON u.id = ch.user_id
			WHERE j.id=?`, entityID).
			Scan(&entry.ID, &entry.Title, &entry.Entry, &entry.EntryDate, &entry.CreatedAt,
				&entry.CharacterID, &entry.CharacterName, &entry.OwnerName)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "journal entry not found"})
			return
		}
		c.JSON(http.StatusOK, entry)

	case "map":
		var m CampaignMap
		var campaignName string
		err := db.DB.QueryRow(`
			SELECT m.id, m.campaign_id, m.name, COALESCE(m.image_url,''), m.width, m.height, m.grid_size, m.is_active,
				COALESCE(c.name,'')
			FROM campaign_maps m LEFT JOIN campaigns c ON c.id = m.campaign_id
			WHERE m.id=?`, entityID).
			Scan(&m.ID, &m.CampaignID, &m.Name, &m.ImageURL, &m.Width, &m.Height, &m.GridSize, &m.IsActive, &campaignName)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "map not found"})
			return
		}

		rows, err := db.DB.Query(`
			SELECT id, map_id, name, type, x, y, COALESCE(icon,''), COALESCE(color,''), COALESCE(description,'')
			FROM campaign_map_pins WHERE map_id=? AND is_hidden=0 ORDER BY sort_order, name`, entityID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var pins []MapPin
		for rows.Next() {
			var p MapPin
			rows.Scan(&p.ID, &p.MapID, &p.Name, &p.Type, &p.X, &p.Y, &p.Icon, &p.Color, &p.Description)
			pins = append(pins, p)
		}

		c.JSON(http.StatusOK, gin.H{
			"map":           m,
			"campaign_name": campaignName,
			"pins":          pins,
		})

	case "upload":
		var u models.Upload
		err := db.DB.QueryRow(`
			SELECT id, hash, ext, url, COALESCE(resized_url,''), COALESCE(thumbnail_url,''), owner_type, owner_id, COALESCE(created_at,'')
			FROM uploads WHERE id=?`, entityID).
			Scan(&u.ID, &u.Hash, &u.Ext, &u.URL, &u.ResizedURL, &u.ThumbnailURL, &u.OwnerType, &u.OwnerID, &u.CreatedAt)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "upload not found"})
			return
		}
		c.JSON(http.StatusOK, u)

	default:
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "unknown entity type"})
	}
}

func ListShareLinks(c *gin.Context) {
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	query := "SELECT token, entity_type, entity_id, created_by, created_at, COALESCE(expires_at, '') FROM share_links"
	var rows *sql.Rows
	var err error
	if role == "admin" {
		rows, err = db.DB.Query(query + " ORDER BY created_at DESC")
	} else {
		rows, err = db.DB.Query(query+" WHERE created_by=? ORDER BY created_at DESC", userID)
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var links []ShareLinkResponse
	for rows.Next() {
		var sl ShareLinkResponse
		var createdBy int64
		var expiresAt string
		rows.Scan(&sl.Token, &sl.EntityType, &sl.EntityID, &createdBy, &sl.CreatedAt, &expiresAt)
		sl.Label = shareEntityLabel(sl.EntityType, sl.EntityID)
		if expiresAt != "" {
			sl.ExpiresAt = expiresAt
		}
		links = append(links, sl)
	}
	if links == nil {
		links = []ShareLinkResponse{}
	}
	c.JSON(http.StatusOK, links)
}

func DeleteShareLink(c *gin.Context) {
	token := c.Param("token")
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	var createdBy int64
	err := db.DB.QueryRow("SELECT created_by FROM share_links WHERE token=?", token).Scan(&createdBy)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "share link not found"})
		return
	}

	uid, _ := userID.(int64)
	if role != "admin" && createdBy != uid {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	db.DB.Exec("DELETE FROM share_links WHERE token=?", token)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// GetSharedPage renders a public HTML page for a shared entity (note, journal,
// map, upload). Pages are server-rendered with html/template autoescaping and
// marked noindex. character/party shares remain JSON-only.
func GetSharedPage(c *gin.Context) {
	entityType, entityID, expired, ok := lookupShare(c.Param("token"))
	if !ok {
		c.String(http.StatusNotFound, "Share link not found.")
		return
	}
	if expired {
		c.String(http.StatusGone, "Share link has expired.")
		return
	}

	switch entityType {
	case "note":
		var note SharedNoteView
		err := db.DB.QueryRow(`
			SELECT cn.title, cn.content, cn.visibility, cn.category, COALESCE(cn.created_at,''), COALESCE(u.username,'')
			FROM character_notes cn JOIN characters ch ON ch.id = cn.character_id
			LEFT JOIN users u ON u.id = ch.user_id WHERE cn.id=?`, entityID).
			Scan(&note.Title, &note.Content, &note.Visibility, &note.Category, &note.CreatedAt, &note.OwnerName)
		if err != nil {
			c.String(http.StatusNotFound, "Note not found.")
			return
		}
		renderSharePage(c, "share_note.html", note)

	case "journal":
		var entry SharedJournalView
		err := db.DB.QueryRow(`
			SELECT j.title, j.entry, COALESCE(j.entry_date,''), COALESCE(j.created_at,''), COALESCE(u.username,'')
			FROM journal j JOIN characters ch ON ch.id = j.character_id
			LEFT JOIN users u ON u.id = ch.user_id WHERE j.id=?`, entityID).
			Scan(&entry.Title, &entry.Entry, &entry.EntryDate, &entry.CreatedAt, &entry.OwnerName)
		if err != nil {
			c.String(http.StatusNotFound, "Journal entry not found.")
			return
		}
		renderSharePage(c, "share_journal.html", entry)

	case "map":
		var m CampaignMap
		var campaignName string
		err := db.DB.QueryRow(`
			SELECT m.id, m.name, COALESCE(m.image_url,''), m.width, m.height, m.grid_size, COALESCE(c.name,'')
			FROM campaign_maps m LEFT JOIN campaigns c ON c.id = m.campaign_id WHERE m.id=?`, entityID).
			Scan(&m.ID, &m.Name, &m.ImageURL, &m.Width, &m.Height, &m.GridSize, &campaignName)
		if err != nil {
			c.String(http.StatusNotFound, "Map not found.")
			return
		}

		rows, err := db.DB.Query(`
			SELECT id, name, type, x, y, COALESCE(icon,''), COALESCE(color,''), COALESCE(description,'')
			FROM campaign_map_pins WHERE map_id=? AND is_hidden=0 ORDER BY sort_order, name`, entityID)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to load map pins.")
			return
		}
		defer rows.Close()

		var pins []MapPin
		for rows.Next() {
			var p MapPin
			rows.Scan(&p.ID, &p.Name, &p.Type, &p.X, &p.Y, &p.Icon, &p.Color, &p.Description)
			pins = append(pins, p)
		}

		renderSharePage(c, "share_map.html", gin.H{
			"map":           m,
			"campaign_name": campaignName,
			"pins":          pins,
		})

	case "upload":
		var u models.Upload
		err := db.DB.QueryRow(`
			SELECT id, ext, url, COALESCE(resized_url,''), COALESCE(thumbnail_url,'')
			FROM uploads WHERE id=?`, entityID).
			Scan(&u.ID, &u.Ext, &u.URL, &u.ResizedURL, &u.ThumbnailURL)
		if err != nil {
			c.String(http.StatusNotFound, "File not found.")
			return
		}
		renderSharePage(c, "share_upload.html", u)

	default:
		c.String(http.StatusBadRequest, "This share type has no public page.")
	}
}

// renderSharePage executes a share template (full HTML document) from the
// embedded template set with html/template autoescaping.
func renderSharePage(c *gin.Context, name string, data any) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := htmxTemplates.ExecuteTemplate(c.Writer, name, data); err != nil {
		c.String(http.StatusInternalServerError, "template error: %v", err)
	}
}
