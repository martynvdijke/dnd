package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

func deriveAttackBonus(ch models.Character, item models.InventoryItem) int {
	if item.AttackBonus != nil {
		return *item.AttackBonus
	}
	// determine ability
	ability := strings.ToLower(strings.TrimSpace(item.AttackAbility))
	if ability == "" {
		props := strings.ToLower(item.WeaponProperties)
		if strings.Contains(props, "finesse") || strings.Contains(props, "ranged") {
			ability = "dex"
		} else {
			ability = "str"
		}
	}
	mod := 0
	switch ability {
	case "dex":
		mod = abilityMod(ch.Dex)
	case "str":
		mod = abilityMod(ch.Str)
	case "con":
		mod = abilityMod(ch.Con)
	case "int":
		mod = abilityMod(ch.Int)
	case "wis":
		mod = abilityMod(ch.Wis)
	case "cha":
		mod = abilityMod(ch.Cha)
	default:
		mod = abilityMod(ch.Str)
	}
	pb := ch.ProficiencyBonus
	if pb == 0 {
		pb = 2 + (ch.Level-1)/4
	}
	bonus := mod + pb
	if item.IsMagical {
		bonus++
	}
	return bonus
}

func rollD20Advantage(adv string) (raw int, rolls []int, err error) {
	rollOne := func() (int, error) {
		result, e := getDicePool().Roll("1d20")
		if e != nil {
			return 0, e
		}
		v, _ := strconv.Atoi(string(result.Total))
		return v, nil
	}
	adv = strings.ToLower(adv)
	if adv == "advantage" || adv == "disadvantage" {
		a, e := rollOne()
		if e != nil {
			return 0, nil, e
		}
		b, e := rollOne()
		if e != nil {
			return 0, nil, e
		}
		rolls = []int{a, b}
		if adv == "advantage" {
			if b > a {
				raw = b
			} else {
				raw = a
			}
		} else {
			if b < a {
				raw = b
			} else {
				raw = a
			}
		}
		return raw, rolls, nil
	}
	v, e := rollOne()
	if e != nil {
		return 0, nil, e
	}
	return v, []int{v}, nil
}

func rollDamageExpr(expr string, crit bool) (total int, breakdown []int, err error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return 0, nil, nil
	}
	rollOnce := func(e string) (int, []int, error) {
		res, err := getDicePool().Roll(e)
		if err != nil {
			return 0, nil, err
		}
		t, _ := strconv.Atoi(string(res.Total))
		// try to extract breakdown via ToHandlerResult if needed, but just return total
		return t, []int{t}, nil
	}
	if crit {
		a, _, err := rollOnce(expr)
		if err != nil {
			return 0, nil, err
		}
		b, _, err := rollOnce(expr)
		if err != nil {
			return 0, nil, err
		}
		return a + b, []int{a, b}, nil
	}
	t, br, err := rollOnce(expr)
	if err != nil {
		return 0, nil, err
	}
	return t, br, nil
}

func HandleCombatAttack(c *gin.Context) {
	var req struct {
		AttackerType     string `json:"attacker_type"`
		AttackerID       int64  `json:"attacker_id"`
		TargetType       string `json:"target_type"`
		TargetID         int64  `json:"target_id"`
		ItemID           *int64 `json:"item_id"`
		AttackBonus      *int   `json:"attack_bonus"`
		DamageDice       string `json:"damage_dice"`
		DamageType       string `json:"damage_type"`
		Advantage        string `json:"advantage"`
		Apply            bool   `json:"apply"`
		Condition        string `json:"condition"`
		ConditionDuration *int  `json:"condition_duration"`
		CampaignID       *int64 `json:"campaign_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// load attacker character if needed
	var attackerChar *models.Character
	var attackBonus int
	var damageDice, damageType, weaponProps string
	var isMagical bool
	var attackerName string
	if req.AttackerType == "character" {
		var ch models.Character
		err := db.DB.QueryRow("SELECT id, name, str, dex, con, int, wis, cha, level, proficiency_bonus FROM characters WHERE id=?", req.AttackerID).Scan(&ch.ID, &ch.Name, &ch.Str, &ch.Dex, &ch.Con, &ch.Int, &ch.Wis, &ch.Cha, &ch.Level, &ch.ProficiencyBonus)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "attacker not found"})
			return
		}
		attackerChar = &ch
		attackerName = ch.Name
		if req.ItemID != nil {
			var item models.InventoryItem
			var abil string
			var b sqlNullInt64
			var name, dmgDice, dmgType, wProps string
			var magical int
			err = db.DB.QueryRow("SELECT id, name, damage_dice, damage_type, weapon_properties, is_magical, attack_ability, attack_bonus FROM inventory WHERE id=?", *req.ItemID).Scan(&item.ID, &name, &dmgDice, &dmgType, &wProps, &magical, &abil, &b)
			if err == nil {
				item.Name = name
				item.DamageDice = dmgDice
				item.DamageType = dmgType
				item.WeaponProperties = wProps
				item.IsMagical = magical == 1
				item.AttackAbility = abil
				if b.Valid {
					v := int(b.Int64)
					item.AttackBonus = &v
				}
				attackBonus = deriveAttackBonus(*attackerChar, item)
				damageDice = item.DamageDice
				damageType = item.DamageType
				weaponProps = item.WeaponProperties
				isMagical = item.IsMagical
				_ = weaponProps
				_ = isMagical
			} else {
				// fallback to body values
				if req.AttackBonus != nil {
					attackBonus = *req.AttackBonus
				}
				damageDice = req.DamageDice
				damageType = req.DamageType
			}
		} else {
			if req.AttackBonus != nil {
				attackBonus = *req.AttackBonus
			}
			damageDice = req.DamageDice
			damageType = req.DamageType
		}
	} else {
		// attacker is combat entry or unspecified
		if req.AttackBonus != nil {
			attackBonus = *req.AttackBonus
		}
		damageDice = req.DamageDice
		damageType = req.DamageType
		attackerName = "Attacker"
		if req.AttackerID != 0 {
			db.DB.QueryRow("SELECT name FROM combat_entries WHERE id=?", req.AttackerID).Scan(&attackerName)
		}
	}
	// resolve target AC and name
	targetAC := 10
	targetName := "Target"
	var targetCampaignID *int64
	var combatEntryID *int64
	switch req.TargetType {
	case "combat":
		err := db.DB.QueryRow("SELECT ac, name, campaign_id FROM combat_entries WHERE id=?", req.TargetID).Scan(&targetAC, &targetName, &targetCampaignID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "target not found"})
			return
		}
		combatEntryID = &req.TargetID
		if targetCampaignID != nil && *targetCampaignID == 0 {
			targetCampaignID = nil
		}
		if req.CampaignID != nil {
			targetCampaignID = req.CampaignID
		}
	case "character":
		err := db.DB.QueryRow("SELECT ac, name, campaign_id FROM characters WHERE id=?", req.TargetID).Scan(&targetAC, &targetName, &targetCampaignID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "target not found"})
			return
		}
		if targetCampaignID != nil && *targetCampaignID == 0 {
			targetCampaignID = nil
		}
		if req.CampaignID != nil {
			targetCampaignID = req.CampaignID
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "target_type must be character or combat"})
		return
	}

	raw, rolls, err := rollD20Advantage(req.Advantage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	critical := raw == 20
	fumble := raw == 1
	attackTotal := raw + attackBonus
	hit := false
	if !fumble {
		if critical || attackTotal >= targetAC {
			hit = true
		}
	}
	damage := 0
	var breakdown []int
	if hit {
		damage, breakdown, _ = rollDamageExpr(damageDice, critical)
	}
	conditionApplied := ""
	if hit && req.Condition != "" && req.TargetType == "character" {
		dur := 1
		if req.ConditionDuration != nil {
			dur = *req.ConditionDuration
		}
		applied, _ := applyConditionIfNotImmune(req.TargetID, req.Condition, req.Condition, "attack", dur, "round", 0)
		if applied {
			conditionApplied = req.Condition
		}
	}

	var appliedResult *HPChangeResult
	var targetHP *int
	if req.Apply && hit {
		if req.TargetType == "character" {
			campID := targetCampaignID
			if req.CampaignID != nil {
				campID = req.CampaignID
			}
			res, err := applyHPChange(req.TargetID, -damage, damageType, "attack", campID)
			if err == nil {
				appliedResult = &res
			}
		} else {
			// combat entry
			var current int
			db.DB.QueryRow("SELECT hp_current FROM combat_entries WHERE id=?", req.TargetID).Scan(&current)
			newHP := current - damage
			if newHP < 0 {
				newHP = 0
			}
			db.DB.Exec("UPDATE combat_entries SET hp_current=? WHERE id=?", newHP, req.TargetID)
			targetHP = &newHP
		}
	}

	descMap := map[string]any{
		"attacker": attackerName, "target": targetName, "raw": raw, "rolls": rolls,
		"attack_bonus": attackBonus, "attack_total": attackTotal, "target_ac": targetAC,
		"hit": hit, "critical": critical, "damage": damage, "damage_type": damageType,
	}
	descBytes, _ := json.Marshal(descMap)
	campForLog := targetCampaignID
	if req.CampaignID != nil {
		campForLog = req.CampaignID
	}
	logID := logCombatEvent(CombatLogEntry{
		CampaignID: campForLog, CombatEntryID: combatEntryID,
		ActorName: attackerName, Action: "attack", TargetName: targetName,
		Damage: damage, DamageType: damageType, ConditionApplied: conditionApplied,
		RollExpression: "1d20", RollTotal: attackTotal, IsCritical: critical,
		Description: string(descBytes),
	})

	resp := gin.H{
		"attack_roll": rolls[0], "attack_rolls": rolls, "attack_bonus": attackBonus,
		"attack_total": attackTotal, "target_ac": targetAC,
		"hit": hit, "critical": critical, "fumble": fumble,
		"damage": damage, "damage_type": damageType, "damage_breakdown": breakdown,
		"condition_applied": conditionApplied, "log_id": logID,
	}
	if appliedResult != nil {
		resp["applied"] = appliedResult
	}
	if targetHP != nil {
		resp["target_hp"] = *targetHP
	}
	c.JSON(http.StatusOK, resp)
}

// sqlNullInt64 helper for nullable int
type sqlNullInt64 struct {
	Int64 int64
	Valid bool
}

func (n *sqlNullInt64) Scan(value interface{}) error {
	if value == nil {
		n.Valid = false
		return nil
	}
	switch v := value.(type) {
	case int64:
		n.Int64 = v
		n.Valid = true
	case int:
		n.Int64 = int64(v)
		n.Valid = true
	default:
		// try via string
		s := ""
		if bs, ok := value.([]byte); ok {
			s = string(bs)
		} else if str, ok := value.(string); ok {
			s = str
		}
		if s != "" {
			iv, err := strconv.Atoi(s)
			if err == nil {
				n.Int64 = int64(iv)
				n.Valid = true
			}
		}
	}
	return nil
}
