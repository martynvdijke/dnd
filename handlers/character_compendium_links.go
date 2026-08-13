package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
)

// ─── Character Race/Class/Background Linking ───
//
// Linking copies the compendium entry's name into the character's existing
// text field (race/class/background) and stores the reference. Unlinking nulls
// the reference, preserving the text — the text fields remain the single
// source of display truth and stay editable via autoSaveField.

func characterIdentityLinkInsert(charID, compendiumID int64, table, textCol, linkCol string) (int, string) {
	var name string
	err := db.DB.QueryRow("SELECT name FROM "+table+" WHERE id=?", compendiumID).Scan(&name)
	if err != nil {
		return http.StatusNotFound, "compendium entry not found"
	}
	_, err = db.DB.Exec("UPDATE characters SET "+textCol+"=?, "+linkCol+"=? WHERE id=?", name, compendiumID, charID)
	if err != nil {
		return http.StatusInternalServerError, err.Error()
	}
	return 0, ""
}

func characterIdentityLinkUnlink(charID int64, linkCol string) (int, string) {
	_, err := db.DB.Exec("UPDATE characters SET "+linkCol+" = NULL WHERE id=?", charID)
	if err != nil {
		return http.StatusInternalServerError, err.Error()
	}
	return 0, ""
}

func linkCharacterRace(c *gin.Context) {
	charID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid character id"})
		return
	}
	if !canEditCharacterID(c, charID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	compID, err := strconv.ParseInt(c.PostForm("compendium_race_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "compendium_race_id required"})
		return
	}
	if st, msg := characterIdentityLinkInsert(charID, compID, "compendium_races", "race", "compendium_race_id"); msg != "" {
		c.JSON(st, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "linked"})
}

func unlinkCharacterRace(c *gin.Context) {
	charID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid character id"})
		return
	}
	if !canEditCharacterID(c, charID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	if st, msg := characterIdentityLinkUnlink(charID, "compendium_race_id"); msg != "" {
		c.JSON(st, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "unlinked"})
}

func linkCharacterClass(c *gin.Context) {
	charID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid character id"})
		return
	}
	if !canEditCharacterID(c, charID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	compID, err := strconv.ParseInt(c.PostForm("compendium_class_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "compendium_class_id required"})
		return
	}
	if st, msg := characterIdentityLinkInsert(charID, compID, "compendium_classes", "class", "compendium_class_id"); msg != "" {
		c.JSON(st, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "linked"})
}

func unlinkCharacterClass(c *gin.Context) {
	charID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid character id"})
		return
	}
	if !canEditCharacterID(c, charID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	if st, msg := characterIdentityLinkUnlink(charID, "compendium_class_id"); msg != "" {
		c.JSON(st, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "unlinked"})
}

func linkCharacterBackground(c *gin.Context) {
	charID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid character id"})
		return
	}
	if !canEditCharacterID(c, charID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	compID, err := strconv.ParseInt(c.PostForm("compendium_background_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "compendium_background_id required"})
		return
	}
	if st, msg := characterIdentityLinkInsert(charID, compID, "compendium_backgrounds", "background", "compendium_background_id"); msg != "" {
		c.JSON(st, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "linked"})
}

func unlinkCharacterBackground(c *gin.Context) {
	charID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid character id"})
		return
	}
	if !canEditCharacterID(c, charID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	if st, msg := characterIdentityLinkUnlink(charID, "compendium_background_id"); msg != "" {
		c.JSON(st, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "unlinked"})
}
