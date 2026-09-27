package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"villum/db"
)

type CombatEntry struct {
	ID              int64  `json:"id"`
	CharacterID     *int64 `json:"character_id,omitempty"`
	CampaignID      *int64 `json:"campaign_id,omitempty"`
	Name            string `json:"name"`
	Type            string `json:"type"`
	InitiativeRoll  int    `json:"initiative_roll"`
	InitiativeMod   int    `json:"initiative_mod"`
	HPMax           int    `json:"hp_max"`
	HPCurrent       int    `json:"hp_current"`
	AC              int    `json:"ac"`
	IsActive        bool   `json:"is_active"`
	TurnOrder       int    `json:"turn_order"`
	ConditionIDs    string `json:"condition_ids"`
	Notes           string `json:"notes"`
	ActionUsed      bool   `json:"action_used"`
	BonusActionUsed bool   `json:"bonus_action_used"`
	ReactionUsed    bool   `json:"reaction_used"`
	MovementUsed    int    `json:"movement_used"`
	MovementMax     int    `json:"movement_max"`
}

const combatEntryCols = "id,character_id,campaign_id,name,type,initiative_roll,initiative_mod,hp_max,hp_current,ac,is_active,turn_order,condition_ids,notes,action_used,bonus_action_used,reaction_used,movement_used,movement_max"

type rowScanner interface {
	Scan(dest ...any) error
}

func scanCombatEntry(s rowScanner) (CombatEntry, error) {
	var e CombatEntry
	var isActive, action, bonus, reaction int
	err := s.Scan(&e.ID, &e.CharacterID, &e.CampaignID, &e.Name, &e.Type,
		&e.InitiativeRoll, &e.InitiativeMod, &e.HPMax, &e.HPCurrent, &e.AC,
		&isActive, &e.TurnOrder, &e.ConditionIDs, &e.Notes,
		&action, &bonus, &reaction, &e.MovementUsed, &e.MovementMax)
	e.IsActive = isActive == 1
	e.ActionUsed = action == 1
	e.BonusActionUsed = bonus == 1
	e.ReactionUsed = reaction == 1
	return e, err
}

func CreateCombatEntry(c *gin.Context) {
	var e CombatEntry
	if err := c.ShouldBindJSON(&e); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(e.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	if e.MovementMax <= 0 {
		e.MovementMax = 30
	}
	result, err := db.DB.Exec(`INSERT INTO combat_entries(character_id,campaign_id,name,type,initiative_roll,initiative_mod,hp_max,hp_current,ac,turn_order,condition_ids,notes,action_used,bonus_action_used,reaction_used,movement_used,movement_max,is_active) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,1)`,
		e.CharacterID, e.CampaignID, e.Name, e.Type, e.InitiativeRoll, e.InitiativeMod, e.HPMax, e.HPCurrent, e.AC, e.TurnOrder, e.ConditionIDs, e.Notes,
		boolToInt(e.ActionUsed), boolToInt(e.BonusActionUsed), boolToInt(e.ReactionUsed), e.MovementUsed, e.MovementMax)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	if e.CampaignID != nil {
		SendCombatUpdate(*e.CampaignID)
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func ListCombatEntries(c *gin.Context) {
	campaignID := c.Query("campaign_id")
	var rows *sql.Rows
	var err error

	if campaignID != "" {
		rows, err = db.DB.Query("SELECT "+combatEntryCols+" FROM combat_entries WHERE campaign_id=? ORDER BY initiative_roll DESC, turn_order", campaignID)
	} else {
		rows, err = db.DB.Query("SELECT " + combatEntryCols + " FROM combat_entries ORDER BY initiative_roll DESC, turn_order")
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var entries = make([]CombatEntry, 0)
	for rows.Next() {
		e, err := scanCombatEntry(rows)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		entries = append(entries, e)
	}
	c.JSON(http.StatusOK, entries)
}

func UpdateCombatEntry(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var e CombatEntry
	if err := c.ShouldBindJSON(&e); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var campaignID *int64
	db.DB.QueryRow("SELECT campaign_id FROM combat_entries WHERE id=?", id).Scan(&campaignID)
	db.DB.Exec(`UPDATE combat_entries SET name=?,type=?,initiative_roll=?,initiative_mod=?,hp_max=?,hp_current=?,ac=?,is_active=?,turn_order=?,condition_ids=?,notes=?,action_used=?,bonus_action_used=?,reaction_used=?,movement_used=?,movement_max=? WHERE id=?`,
		e.Name, e.Type, e.InitiativeRoll, e.InitiativeMod, e.HPMax, e.HPCurrent, e.AC, boolToInt(e.IsActive), e.TurnOrder, e.ConditionIDs, e.Notes,
		boolToInt(e.ActionUsed), boolToInt(e.BonusActionUsed), boolToInt(e.ReactionUsed), e.MovementUsed, e.MovementMax, id)
	if campaignID != nil {
		SendCombatUpdate(*campaignID)
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteCombatEntry(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var campaignID *int64
	db.DB.QueryRow("SELECT campaign_id FROM combat_entries WHERE id=?", id).Scan(&campaignID)
	db.DB.Exec("DELETE FROM combat_entries WHERE id=?", id)
	if campaignID != nil {
		SendCombatUpdate(*campaignID)
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func RollInitiative(c *gin.Context) {
	var req struct {
		CharacterID int64 `json:"character_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var name, class string
	var dex, initiative int
	err := db.DB.QueryRow("SELECT name, class, dex, initiative FROM characters WHERE id=?", req.CharacterID).
		Scan(&name, &class, &dex, &initiative)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}

	initMod := abilityMod(dex) + initiative
	result, err := getDicePool().Roll("1d20")
	d20 := 0
	if err == nil {
		d20, _ = strconv.Atoi(string(result.Total))
	}
	roll := d20 + initMod

	c.JSON(http.StatusOK, gin.H{
		"character_id": req.CharacterID,
		"name":         name,
		"d20":          d20,
		"modifier":     initMod,
		"total":        roll,
	})
}

func NextTurn(c *gin.Context) {
	campaignID := c.Query("campaign_id")
	// Find the highest turn_order, advance it (wrap around)
	var maxOrder int
	var currentEntry CombatEntry
	if campaignID != "" {
		db.DB.QueryRow("SELECT COALESCE(MAX(turn_order),0) FROM combat_entries WHERE campaign_id=? AND is_active=1", campaignID).Scan(&maxOrder)
		db.DB.Exec("UPDATE combat_entries SET turn_order = CASE WHEN turn_order >= ? THEN 0 ELSE turn_order + 1 END WHERE campaign_id=? AND is_active=1", maxOrder, campaignID)
		currentEntry, _ = scanCombatEntry(db.DB.QueryRow("SELECT "+combatEntryCols+" FROM combat_entries WHERE campaign_id=? AND is_active=1 ORDER BY turn_order DESC LIMIT 1", campaignID))
	} else {
		db.DB.QueryRow("SELECT COALESCE(MAX(turn_order),0) FROM combat_entries WHERE is_active=1").Scan(&maxOrder)
		db.DB.Exec("UPDATE combat_entries SET turn_order = CASE WHEN turn_order >= ? THEN 0 ELSE turn_order + 1 END WHERE is_active=1", maxOrder)
	}
	// Reset the action economy for the entry whose turn is starting.
	if currentEntry.ID != 0 {
		db.DB.Exec("UPDATE combat_entries SET action_used=0,bonus_action_used=0,reaction_used=0,movement_used=0 WHERE id=?", currentEntry.ID)
		currentEntry.ActionUsed = false
		currentEntry.BonusActionUsed = false
		currentEntry.ReactionUsed = false
		currentEntry.MovementUsed = 0
	}
	if cid, err := strconv.ParseInt(campaignID, 10, 64); err == nil && cid > 0 {
		SendCombatUpdate(cid)
	}
	if currentEntry.CharacterID != nil && *currentEntry.CharacterID != 0 {
		tickConditionsForCharacter(*currentEntry.CharacterID, 1, "round")
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "current_entry": currentEntry})
}

func GetCurrentTurn(c *gin.Context) {
	campaignID := c.Query("campaign_id")
	var entry CombatEntry
	var err error
	if campaignID != "" {
		entry, err = scanCombatEntry(db.DB.QueryRow("SELECT "+combatEntryCols+" FROM combat_entries WHERE campaign_id=? AND is_active=1 ORDER BY turn_order DESC LIMIT 1", campaignID))
	} else {
		entry, err = scanCombatEntry(db.DB.QueryRow("SELECT " + combatEntryCols + " FROM combat_entries WHERE is_active=1 ORDER BY turn_order DESC LIMIT 1"))
	}
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"current": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"current": entry})
}

// HandleCombatEconomy consumes action / bonus action / reaction / movement for
// a combat entry, rejecting overspend.
func HandleCombatEconomy(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		Action      bool `json:"action"`
		BonusAction bool `json:"bonus_action"`
		Reaction    bool `json:"reaction"`
		Movement    int  `json:"movement"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	entry, err := scanCombatEntry(db.DB.QueryRow("SELECT "+combatEntryCols+" FROM combat_entries WHERE id=?", id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "combat entry not found"})
		return
	}
	if req.Action && entry.ActionUsed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "action already used this turn"})
		return
	}
	if req.BonusAction && entry.BonusActionUsed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bonus action already used this turn"})
		return
	}
	if req.Reaction && entry.ReactionUsed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "reaction already used this turn"})
		return
	}
	if req.Movement < 0 || entry.MovementUsed+req.Movement > entry.MovementMax {
		c.JSON(http.StatusBadRequest, gin.H{"error": "not enough movement remaining"})
		return
	}
	db.DB.Exec("UPDATE combat_entries SET action_used=?,bonus_action_used=?,reaction_used=?,movement_used=? WHERE id=?",
		boolToInt(entry.ActionUsed || req.Action), boolToInt(entry.BonusActionUsed || req.BonusAction),
		boolToInt(entry.ReactionUsed || req.Reaction), entry.MovementUsed+req.Movement, id)
	entry.ActionUsed = entry.ActionUsed || req.Action
	entry.BonusActionUsed = entry.BonusActionUsed || req.BonusAction
	entry.ReactionUsed = entry.ReactionUsed || req.Reaction
	entry.MovementUsed += req.Movement
	if entry.CampaignID != nil {
		SendCombatUpdate(*entry.CampaignID)
	}
	c.JSON(http.StatusOK, entry)
}
