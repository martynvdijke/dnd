package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
)

type CombatLogEntry struct {
	ID             int64  `json:"id"`
	CampaignID     *int64 `json:"campaign_id,omitempty"`
	CombatEntryID  *int64 `json:"combat_entry_id,omitempty"`
	ActorName      string `json:"actor_name"`
	Action         string `json:"action"`
	TargetName     string `json:"target_name"`
	Damage         int    `json:"damage"`
	DamageType     string `json:"damage_type"`
	Healing        int    `json:"healing"`
	ConditionApplied string `json:"condition_applied"`
	RollExpression string `json:"roll_expression"`
	RollTotal      int    `json:"roll_total"`
	IsCritical     bool   `json:"is_critical"`
	Description    string `json:"description"`
	CreatedAt      string `json:"created_at"`
}

func ListCombatLogEntries(c *gin.Context) {
	campaignID := c.Query("campaign_id")
	limit := c.DefaultQuery("limit", "50")

	if limitInt, _ := strconv.Atoi(limit); limitInt < 1 {
		limit = "50"
	}

	rows, err := func() (*sql.Rows, error) {
		if campaignID != "" {
			return db.DB.Query("SELECT id,campaign_id,combat_entry_id,actor_name,action,target_name,damage,damage_type,healing,condition_applied,roll_expression,roll_total,is_critical,description,created_at FROM combat_log_entries WHERE campaign_id=? ORDER BY created_at DESC LIMIT "+limit, campaignID)
		}
		return db.DB.Query("SELECT id,campaign_id,combat_entry_id,actor_name,action,target_name,damage,damage_type,healing,condition_applied,roll_expression,roll_total,is_critical,description,created_at FROM combat_log_entries ORDER BY created_at DESC LIMIT "+limit)
	}()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var out = make([]CombatLogEntry, 0)
	for rows.Next() {
		var e CombatLogEntry
		rows.Scan(&e.ID, &e.CampaignID, &e.CombatEntryID, &e.ActorName, &e.Action, &e.TargetName, &e.Damage, &e.DamageType, &e.Healing, &e.ConditionApplied, &e.RollExpression, &e.RollTotal, &e.IsCritical, &e.Description, &e.CreatedAt)
		out = append(out, e)
	}
	c.JSON(http.StatusOK, out)
}

func CreateCombatLogEntry(c *gin.Context) {
	var e CombatLogEntry
	if err := c.ShouldBindJSON(&e); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := db.DB.Exec(`INSERT INTO combat_log_entries(campaign_id,combat_entry_id,actor_name,action,target_name,damage,damage_type,healing,condition_applied,roll_expression,roll_total,is_critical,description) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		e.CampaignID, e.CombatEntryID, e.ActorName, e.Action, e.TargetName, e.Damage, e.DamageType, e.Healing, e.ConditionApplied, e.RollExpression, e.RollTotal, e.IsCritical, e.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func GetCombatLogStats(c *gin.Context) {
	campaignID := c.Query("campaign_id")
	if campaignID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "campaign_id required"})
		return
	}

	var totalEntries, totalDamage, totalHealing, critCount int
	db.DB.QueryRow("SELECT COUNT(*), COALESCE(SUM(damage),0), COALESCE(SUM(healing),0), COALESCE(SUM(is_critical),0) FROM combat_log_entries WHERE campaign_id=?", campaignID).Scan(&totalEntries, &totalDamage, &totalHealing, &critCount)

	topDamagers, _ := db.DB.Query("SELECT actor_name, SUM(damage) as dmg FROM combat_log_entries WHERE campaign_id=? GROUP BY actor_name ORDER BY dmg DESC LIMIT 5", campaignID)
	type DamagerStat struct {
		Name string `json:"name"`
		Damage int `json:"damage"`
	}
	var topDmg []DamagerStat
	if topDamagers != nil {
		for topDamagers.Next() {
			var ds DamagerStat
			topDamagers.Scan(&ds.Name, &ds.Damage)
			topDmg = append(topDmg, ds)
		}
		topDamagers.Close()
	}

	c.JSON(http.StatusOK, gin.H{
		"total_entries":  totalEntries,
		"total_damage":   totalDamage,
		"total_healing":  totalHealing,
		"crit_count":     critCount,
		"top_damagers":   topDmg,
	})
}
