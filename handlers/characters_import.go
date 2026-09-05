package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"time"
	"villum/db"
	"villum/models"
)

func ImportCharacterJSON(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var imp models.ImportCharacter
	if !BindOr400(c, &imp) {
		return
	}

	// Set defaults
	if imp.Level < 1 {
		imp.Level = 1
	}
	if imp.HPMax < 1 {
		imp.HPMax = 10
		imp.HPCurrent = 10
	}
	if imp.HitDice == "" {
		imp.HitDice = "1d10"
	}

	tx, err := db.Client.Tx(c.Request.Context())
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	defer tx.Rollback()

	uid, _ := userID.(int64)
	now := time.Now().Format("2006-01-02 15:04:05")

	char, err := tx.Character.Create().
		SetUserID(uid).
		SetName(imp.Name).
		SetRace(imp.Race).
		SetClass(imp.Class).
		SetSubclass(imp.Subclass).
		SetLevel(imp.Level).
		SetXp(imp.XP).
		SetBackground(imp.Background).
		SetAlignment(imp.Alignment).
		SetStr(imp.Str).
		SetDex(imp.Dex).
		SetCon(imp.Con).
		SetInt(imp.Int).
		SetWis(imp.Wis).
		SetCha(imp.Cha).
		SetAc(imp.AC).
		SetInitiative(imp.Initiative).
		SetSpeed(imp.Speed).
		SetHpMax(imp.HPMax).
		SetHpCurrent(imp.HPCurrent).
		SetTempHp(imp.TempHP).
		SetHitDice(imp.HitDice).
		SetHitDiceCurrent(0).
		SetProficiencyBonus(0).
		SetInspiration(0).
		SetPassivePerception(0).
		SetPersonalityTraits(imp.PersonalityTraits).
		SetIdeals(imp.Ideals).
		SetBonds(imp.Bonds).
		SetFlaws(imp.Flaws).
		SetAppearance(imp.Appearance).
		SetBackstory(imp.Backstory).
		SetCreatedAt(now).
		SetUpdatedAt(now).
		Save(c.Request.Context())
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}

	charID := char.ID

	// Currency
	tx.CharacterCurrency.Create().
		SetCharacterID(charID).
		SetCp(imp.Currency.CP).
		SetSp(imp.Currency.SP).
		SetEp(imp.Currency.EP).
		SetGp(imp.Currency.GP).
		SetPp(imp.Currency.PP).
		Save(c.Request.Context())

	// Proficiencies
	for _, p := range imp.Proficiencies {
		tx.CharacterProficiency.Create().
			SetCharacterID(charID).
			SetType(p.Type).
			SetName(p.Name).
			Save(c.Request.Context())
	}

	// Features
	for _, f := range imp.Features {
		tx.CharacterFeature.Create().
			SetCharacterID(charID).
			SetName(f.Name).
			SetDescription(f.Description).
			SetSource(f.Source).
			SetLevelGained(f.LevelGained).
			Save(c.Request.Context())
	}

	// Spellcasting
	if imp.Spellcasting != nil {
		sc := imp.Spellcasting
		tx.CharacterSpellcasting.Create().
			SetCharacterID(charID).
			SetAbility(sc.Ability).
			SetSaveDc(sc.SaveDC).
			SetAttackBonus(sc.AttackBonus).
			SetSlots1Max(sc.Slots1Max).SetSlots1Used(sc.Slots1Used).
			SetSlots2Max(sc.Slots2Max).SetSlots2Used(sc.Slots2Used).
			SetSlots3Max(sc.Slots3Max).SetSlots3Used(sc.Slots3Used).
			SetSlots4Max(sc.Slots4Max).SetSlots4Used(sc.Slots4Used).
			SetSlots5Max(sc.Slots5Max).SetSlots5Used(sc.Slots5Used).
			SetSlots6Max(sc.Slots6Max).SetSlots6Used(sc.Slots6Used).
			SetSlots7Max(sc.Slots7Max).SetSlots7Used(sc.Slots7Used).
			SetSlots8Max(sc.Slots8Max).SetSlots8Used(sc.Slots8Used).
			SetSlots9Max(sc.Slots9Max).SetSlots9Used(sc.Slots9Used).
			Save(c.Request.Context())
	}

	// Spells
	for _, sp := range imp.Spells {
		tx.Spell.Create().
			SetCharacterID(charID).
			SetName(sp.Name).
			SetLevel(sp.Level).
			SetSchool(sp.School).
			SetCastingTime(sp.CastingTime).
			SetRange(sp.Range).
			SetComponents(sp.Components).
			SetDuration(sp.Duration).
			SetDescription(sp.Description).
			SetPrepared(sp.Prepared).
			SetAlwaysPrepared(sp.AlwaysPrepared).
			SetSource(sp.Source).
			SetNotes(sp.Notes).
			Save(c.Request.Context())
	}

	// Inventory
	for _, item := range imp.Inventory {
		tx.InventoryItem.Create().
			SetCharacterID(charID).
			SetName(item.Name).
			SetQuantity(item.Quantity).
			SetWeight(item.Weight).
			SetCategory(item.Category).
			SetDamageDice(item.DamageDice).
			SetDamageType(item.DamageType).
			SetWeaponProperties(item.WeaponProperties).
			SetAcBonus(item.ACBonus).
			SetArmorType(item.ArmorType).
			SetDescription(item.Description).
			SetIsEquipped(item.IsEquipped).
			SetIsMagical(item.IsMagical).
			SetAttunement(item.Attunement).
			SetIsIdentified(item.IsIdentified).
			SetNotes(item.Notes).
			Save(c.Request.Context())
	}

	if err := tx.Commit(); err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": charID, "name": imp.Name})
}

func ImportJSON(c *gin.Context) {
	userID, _ := c.Get("user_id")
	contentType := c.GetHeader("Content-Type")

	if strings.HasPrefix(contentType, "multipart/form-data") {
		file, _, err := c.Request.FormFile("file")
		if err != nil {
			WriteError(c, http.StatusBadRequest, strErr("file required"))
			return
		}
		defer file.Close()

		var chars []models.ImportCharacter
		if err := json.NewDecoder(file).Decode(&chars); err != nil {
			var single models.ImportCharacter
			file.Seek(0, 0)
			if err2 := json.NewDecoder(file).Decode(&single); err2 != nil {
				WriteError(c, http.StatusBadRequest, fmt.Errorf("invalid JSON: %w", err))
				return
			}
			chars = []models.ImportCharacter{single}
		}
		results := importCharacters(c.Request.Context(), userID.(int64), chars)
		c.JSON(http.StatusOK, results)
		return
	}

	var chars []models.ImportCharacter
	if err := c.ShouldBindJSON(&chars); err != nil {
		var single models.ImportCharacter
		if err2 := c.ShouldBindJSON(&single); err2 != nil {
			WriteError(c, http.StatusBadRequest, strErr("invalid JSON"))
			return
		}
		chars = []models.ImportCharacter{single}
	}

	results := importCharacters(c.Request.Context(), userID.(int64), chars)
	c.JSON(http.StatusOK, results)
}

func importCharacters(ctx context.Context, userID int64, chars []models.ImportCharacter) []gin.H {
	var results []gin.H
	for _, imp := range chars {
		if imp.Name == "" {
			continue
		}
		if imp.Level < 1 {
			imp.Level = 1
		}
		if imp.HPMax < 1 {
			imp.HPMax = 10
			imp.HPCurrent = 10
		}
		if imp.HitDice == "" {
			imp.HitDice = fmt.Sprintf("1d%d", 10)
		}

		tx, err := db.Client.Tx(ctx)
		if err != nil {
			results = append(results, gin.H{"error": err.Error()})
			continue
		}

		now := time.Now().Format("2006-01-02 15:04:05")

		char, err := tx.Character.Create().
			SetUserID(userID).
			SetName(imp.Name).
			SetRace(imp.Race).
			SetClass(imp.Class).
			SetSubclass(imp.Subclass).
			SetLevel(imp.Level).
			SetXp(imp.XP).
			SetBackground(imp.Background).
			SetAlignment(imp.Alignment).
			SetStr(imp.Str).
			SetDex(imp.Dex).
			SetCon(imp.Con).
			SetInt(imp.Int).
			SetWis(imp.Wis).
			SetCha(imp.Cha).
			SetAc(imp.AC).
			SetInitiative(imp.Initiative).
			SetSpeed(imp.Speed).
			SetHpMax(imp.HPMax).
			SetHpCurrent(imp.HPCurrent).
			SetTempHp(imp.TempHP).
			SetHitDice(imp.HitDice).
			SetHitDiceCurrent(0).
			SetProficiencyBonus(0).
			SetInspiration(0).
			SetPassivePerception(0).
			SetPersonalityTraits(imp.PersonalityTraits).
			SetIdeals(imp.Ideals).
			SetBonds(imp.Bonds).
			SetFlaws(imp.Flaws).
			SetAppearance(imp.Appearance).
			SetBackstory(imp.Backstory).
			SetCreatedAt(now).
			SetUpdatedAt(now).
			Save(ctx)
		if err != nil {
			tx.Rollback()
			results = append(results, gin.H{"error": err.Error(), "name": imp.Name})
			continue
		}
		charID := char.ID

		tx.CharacterCurrency.Create().SetCharacterID(charID).Save(ctx)

		// Proficiencies
		for _, p := range imp.Proficiencies {
			tx.CharacterProficiency.Create().SetCharacterID(charID).SetType(p.Type).SetName(p.Name).Save(ctx)
		}
		for _, f := range imp.Features {
			tx.CharacterFeature.Create().
				SetCharacterID(charID).
				SetName(f.Name).
				SetDescription(f.Description).
				SetSource(f.Source).
				SetLevelGained(f.LevelGained).
				Save(ctx)
		}
		if imp.Spellcasting != nil {
			sc := imp.Spellcasting
			tx.CharacterSpellcasting.Create().
				SetCharacterID(charID).
				SetAbility(sc.Ability).
				SetSaveDc(sc.SaveDC).
				SetAttackBonus(sc.AttackBonus).
				SetSlots1Max(sc.Slots1Max).SetSlots1Used(sc.Slots1Used).
				SetSlots2Max(sc.Slots2Max).SetSlots2Used(sc.Slots2Used).
				SetSlots3Max(sc.Slots3Max).SetSlots3Used(sc.Slots3Used).
				SetSlots4Max(sc.Slots4Max).SetSlots4Used(sc.Slots4Used).
				SetSlots5Max(sc.Slots5Max).SetSlots5Used(sc.Slots5Used).
				SetSlots6Max(sc.Slots6Max).SetSlots6Used(sc.Slots6Used).
				SetSlots7Max(sc.Slots7Max).SetSlots7Used(sc.Slots7Used).
				SetSlots8Max(sc.Slots8Max).SetSlots8Used(sc.Slots8Used).
				SetSlots9Max(sc.Slots9Max).SetSlots9Used(sc.Slots9Used).
				Save(ctx)
		}
		for _, sp := range imp.Spells {
			tx.Spell.Create().
				SetCharacterID(charID).
				SetName(sp.Name).
				SetLevel(sp.Level).
				SetSchool(sp.School).
				SetCastingTime(sp.CastingTime).
				SetRange(sp.Range).
				SetComponents(sp.Components).
				SetDuration(sp.Duration).
				SetDescription(sp.Description).
				SetPrepared(sp.Prepared).
				SetAlwaysPrepared(sp.AlwaysPrepared).
				SetSource(sp.Source).
				SetNotes(sp.Notes).
				Save(ctx)
		}
		for _, item := range imp.Inventory {
			tx.InventoryItem.Create().
				SetCharacterID(charID).
				SetName(item.Name).
				SetQuantity(item.Quantity).
				SetWeight(item.Weight).
				SetCategory(item.Category).
				SetDamageDice(item.DamageDice).
				SetDamageType(item.DamageType).
				SetWeaponProperties(item.WeaponProperties).
				SetAcBonus(item.ACBonus).
				SetArmorType(item.ArmorType).
				SetDescription(item.Description).
				SetIsEquipped(item.IsEquipped).
				SetIsMagical(item.IsMagical).
				SetAttunement(item.Attunement).
				SetNotes(item.Notes).
				Save(ctx)
		}
		if err := tx.Commit(); err != nil {
			results = append(results, gin.H{"error": err.Error(), "name": imp.Name})
			continue
		}
		results = append(results, gin.H{"id": charID, "name": imp.Name, "status": "imported"})
	}
	return results
}

// ─── Exhaustion ───
