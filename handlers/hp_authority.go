package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"villum/db"
)

// calcConcentrationDC mirrors concentration.go logic.
func calcConcentrationDC(damage int) int {
	dc := 10
	if damage >= 22 {
		dc = 15
	} else if damage >= 12 {
		dc = 12
	} else if damage >= 8 {
		dc = 11
	} else if damage >= 4 {
		dc = 10
	}
	halfDmg := damage / 2
	if halfDmg > dc {
		dc = halfDmg
	}
	if dc < 10 {
		dc = 10
	}
	return dc
}

type ConcentrationOutcome struct {
	Checked   bool   `json:"checked"`
	DC        int    `json:"dc"`
	Total     int    `json:"total"`
	Success   bool   `json:"success"`
	Dropped   bool   `json:"dropped"`
	SpellName string `json:"spell_name"`
}

type HPChangeResult struct {
	HPCurrent           int                   `json:"hp_current"`
	TempHP              int                   `json:"temp_hp"`
	HPMax               int                   `json:"hp_max"`
	DeathSavesSuccesses int                   `json:"death_saves_successes"`
	DeathSavesFailures  int                   `json:"death_saves_failures"`
	Concentration       *ConcentrationOutcome `json:"concentration,omitempty"`
}

// logCombatEvent inserts into combat_log_entries and returns id (0 on error).
func logCombatEvent(e CombatLogEntry) int64 {
	res, err := db.DB.Exec(`INSERT INTO combat_log_entries(campaign_id,combat_entry_id,actor_name,action,target_name,damage,damage_type,healing,condition_applied,roll_expression,roll_total,is_critical,description) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		e.CampaignID, e.CombatEntryID, e.ActorName, e.Action, e.TargetName, e.Damage, e.DamageType, e.Healing, e.ConditionApplied, e.RollExpression, e.RollTotal, e.IsCritical, e.Description)
	if err != nil {
		return 0
	}
	id, _ := res.LastInsertId()
	return id
}

// applyConditionIfNotImmune checks condition_immunities and inserts if not immune.
func applyConditionIfNotImmune(charID int64, name, ctype, source string, duration int, durationType string, saveDC int) (bool, error) {
	var immunities string
	_ = db.DB.QueryRow("SELECT COALESCE(condition_immunities,'') FROM characters WHERE id=?", charID).Scan(&immunities)
	if immunities != "" {
		parts := strings.Split(immunities, ",")
		lowerName := strings.ToLower(strings.TrimSpace(name))
		for _, p := range parts {
			if strings.ToLower(strings.TrimSpace(p)) == lowerName && lowerName != "" {
				return false, nil
			}
		}
	}
	if duration < 0 {
		duration = 0
	}
	if durationType == "" {
		durationType = "round"
	}
	_, err := db.DB.Exec("INSERT INTO character_conditions(character_id,name,type,source,duration,duration_type,saving_throw,save_dc,description) VALUES(?,?,?,?,?,?,?,?,?)",
		charID, name, ctype, source, duration, durationType, "", saveDC, "")
	if err != nil {
		return false, err
	}
	return true, nil
}

func applyHPChange(charID int64, delta int, damageType, source string, campaignID *int64) (HPChangeResult, error) {
	var hpMax, hpCurrent, tempHP, dsSucc, dsFail int
	var concentratingOn string
	var str, dex, con, intel, wis, cha, profBonus, level int
	err := db.DB.QueryRow("SELECT hp_max,hp_current,temp_hp,death_saves_successes,death_saves_failures,concentrating_on,str,dex,con,int,wis,cha,proficiency_bonus,level FROM characters WHERE id=?", charID).
		Scan(&hpMax, &hpCurrent, &tempHP, &dsSucc, &dsFail, &concentratingOn, &str, &dex, &con, &intel, &wis, &cha, &profBonus, &level)
	if err != nil {
		return HPChangeResult{}, err
	}
	prevHP := hpCurrent
	damage := 0
	if delta < 0 {
		damage = -delta
		// absorb temp HP first
		if tempHP > 0 {
			if damage <= tempHP {
				tempHP -= damage
				damage = 0
			} else {
				damage -= tempHP
				tempHP = 0
			}
		}
		hpCurrent -= damage
		if hpCurrent < 0 {
			hpCurrent = 0
		}
		if prevHP > 0 && hpCurrent == 0 {
			dsSucc = 0
			dsFail = 0
		} else if prevHP == 0 && damage > 0 {
			// damage at 0 increments failures
			dsFail++
			if dsFail > 3 {
				dsFail = 3
			}
		}
	} else if delta > 0 {
		hpCurrent += delta
		if hpCurrent > hpMax {
			hpCurrent = hpMax
		}
		dsSucc = 0
		dsFail = 0
	}

	var concOutcome *ConcentrationOutcome
	if damage > 0 && concentratingOn != "" {
		dc := calcConcentrationDC(damage)
		// roll CON save
		conMod := abilityMod(con)
		totalMod := conMod
		var cnt int
		db.DB.QueryRow("SELECT COUNT(*) FROM character_proficiencies WHERE character_id=? AND type='save' AND LOWER(name)='con'", charID).Scan(&cnt)
		if cnt > 0 {
			if profBonus == 0 {
				profBonus = 2 + (level-1)/4
			}
			totalMod += profBonus
		}
		raw, rolls, _ := rollD20Advantage("")
		total := raw + totalMod
		_ = rolls
		success := total >= dc
		concOutcome = &ConcentrationOutcome{Checked: true, DC: dc, Total: total, Success: success, SpellName: concentratingOn}
		if !success {
			db.DB.Exec("UPDATE characters SET concentrating_on='' WHERE id=?", charID)
			db.DB.Exec("DELETE FROM character_conditions WHERE character_id=? AND type='concentration'", charID)
			concOutcome.Dropped = true
			concentratingOn = ""
		}
	}

	_, err = db.DB.Exec("UPDATE characters SET hp_current=?, temp_hp=?, death_saves_successes=?, death_saves_failures=? WHERE id=?", hpCurrent, tempHP, dsSucc, dsFail, charID)
	if err != nil {
		return HPChangeResult{}, err
	}

	result := HPChangeResult{
		HPCurrent: hpCurrent, TempHP: tempHP, HPMax: hpMax,
		DeathSavesSuccesses: dsSucc, DeathSavesFailures: dsFail,
		Concentration: concOutcome,
	}

	// log
	var actorName string
	db.DB.QueryRow("SELECT name FROM characters WHERE id=?", charID).Scan(&actorName)
	action := "heal"
	dmg := 0
	heal := 0
	if delta < 0 {
		action = "damage"
		dmg = -delta
	} else {
		heal = delta
	}
	descBytes, _ := json.Marshal(result)
	logCombatEvent(CombatLogEntry{
		CampaignID: campaignID, ActorName: actorName, Action: action,
		Damage: dmg, DamageType: damageType, Healing: heal,
		Description: string(descBytes),
	})

	return result, nil
}

func HandleCharacterHP(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if !canEditCharacterID(c, charID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}
	var req struct {
		Delta  int    `json:"delta"`
		Type   string `json:"type"`
		Source string `json:"source"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var campaignID *int64
	db.DB.QueryRow("SELECT campaign_id FROM characters WHERE id=?", charID).Scan(&campaignID)
	// treat 0 as nil
	if campaignID != nil && *campaignID == 0 {
		campaignID = nil
	}
	result, err := applyHPChange(charID, req.Delta, req.Type, req.Source, campaignID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func HandleSaveVsDC(c *gin.Context) {
	var req struct {
		CharacterID       int64  `json:"character_id"`
		Ability           string `json:"ability"`
		DC                int    `json:"dc"`
		Advantage         string `json:"advantage"`
		HalfOnSave        bool   `json:"half_on_save"`
		Condition         string `json:"condition"`
		ConditionDuration *int   `json:"condition_duration"`
		CampaignID        *int64 `json:"campaign_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ability := strings.ToLower(req.Ability)
	if _, ok := savesMap[ability]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown ability: " + ability})
		return
	}
	checkReq := CheckRollRequest{
		CharacterID: req.CharacterID,
		Type:        "save",
		Name:        ability,
		Advantage:   req.Advantage,
	}
	result, err := resolveCheckRoll(req.CharacterID, checkReq)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}
	success := result.Total >= req.DC
	conditionApplied := ""
	if !success && req.Condition != "" {
		dur := 1
		if req.ConditionDuration != nil {
			dur = *req.ConditionDuration
		}
		applied, _ := applyConditionIfNotImmune(req.CharacterID, req.Condition, req.Condition, "save-vs-dc", dur, "round", req.DC)
		if applied {
			conditionApplied = req.Condition
		}
	}
	// log
	var actorName string
	db.DB.QueryRow("SELECT name FROM characters WHERE id=?", req.CharacterID).Scan(&actorName)
	campID := req.CampaignID
	if campID == nil {
		var cid *int64
		db.DB.QueryRow("SELECT campaign_id FROM characters WHERE id=?", req.CharacterID).Scan(&cid)
		if cid != nil && *cid != 0 {
			campID = cid
		}
	}
	descBytes, _ := json.Marshal(map[string]any{"ability": ability, "dc": req.DC, "success": success, "total": result.Total})
	logCombatEvent(CombatLogEntry{
		CampaignID: campID, ActorName: actorName, Action: "save",
		RollExpression: "1d20", RollTotal: result.Total,
		ConditionApplied: conditionApplied,
		Description:      string(descBytes),
	})
	c.JSON(http.StatusOK, gin.H{
		"success": success, "total": result.Total, "dc": req.DC,
		"raw": result.Raw, "rolls": result.Rolls, "modifier": result.Modifier,
		"ability": ability, "half_on_save": req.HalfOnSave,
		"condition_applied": conditionApplied,
	})
}
