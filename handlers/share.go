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
	CreatedAt  string `json:"created_at"`
	ExpiresAt  string `json:"expires_at,omitempty"`
}

func generateToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func CreateShareLink(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req ShareLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.EntityType != "character" && req.EntityType != "party" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "entity_type must be 'character' or 'party'"})
		return
	}

	if req.EntityType == "character" {
		if !checkCharacterAccess(c, req.EntityID) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
	} else {
		var ownerID int64
		err := db.DB.QueryRow("SELECT user_id FROM campaigns WHERE id=?", req.EntityID).Scan(&ownerID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "campaign not found"})
			return
		}
		role, _ := c.Get("role")
		uid, _ := userID.(int64)
		if role != "admin" && ownerID != uid {
			var count int
			db.DB.QueryRow("SELECT COUNT(*) FROM campaign_members WHERE campaign_id=? AND user_id=?", req.EntityID, uid).Scan(&count)
			if count == 0 {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "access denied"})
				return
			}
		}
	}

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

func GetSharedEntity(c *gin.Context) {
	token := c.Param("token")

	var entityType string
	var entityID int64
	var expiresAt sql.NullString
	var createdBy int64
	err := db.DB.QueryRow("SELECT entity_type, entity_id, created_by, expires_at FROM share_links WHERE token=?", token).
		Scan(&entityType, &entityID, &createdBy, &expiresAt)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "share link not found"})
		return
	}

	if expiresAt.Valid && expiresAt.String != "" {
		expTime, err := time.Parse("2006-01-02 15:04:05", expiresAt.String)
		if err == nil && time.Now().UTC().After(expTime) {
			c.AbortWithStatusJSON(http.StatusGone, gin.H{"error": "share link has expired"})
			return
		}
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

		ch.Proficiencies = loadProficiencies(ch.ID)
		ch.Features = loadFeatures(ch.ID)
		ch.Spellcasting = loadSpellcasting(ch.ID)
		ch.Spells = loadSpells(ch.ID)
		ch.Inventory = loadInventory(ch.ID)
		ch.Currency = loadCurrency(ch.ID)
		ch.Classes = loadCharClasses(ch.ID)
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


