package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
	"villum/db"
	"villum/ent"
	"villum/ent/character"
	"villum/ent/characterclass"
	"villum/ent/charactercurrency"
	"villum/ent/characterfeature"
	"villum/ent/characterproficiency"
	"villum/ent/characterspellcasting"
	"villum/ent/inventoryitem"
	"villum/ent/spell"
	"villum/models"
)

func CreateCharacterClass(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	entChar, err := db.Client.Character.Query().Where(character.ID(charID)).Select(character.FieldUserID, character.FieldCharacterType).Only(c.Request.Context())
	if ent.IsNotFound(err) {
		WriteNotFound(c, "character not found")
		return
	}
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	if !canEditCharacter(c, entChar) {
		WriteError(c, http.StatusForbidden, strErr("access denied"))
		return
	}

	var cc models.CharClass
	if !BindOr400(c, &cc) {
		return
	}
	if cc.Level < 1 {
		cc.Level = 1
	}
	if cc.HitDice == "" {
		cc.HitDice = "d10"
	}

	entCC, err := db.Client.CharacterClass.Create().
		SetCharacterID(charID).
		SetClass(cc.Class).
		SetSubclass(cc.Subclass).
		SetLevel(cc.Level).
		SetHitDice(cc.HitDice).
		Save(c.Request.Context())
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}

	cc.ID = entCC.ID
	cc.CharacterID = charID
	c.JSON(http.StatusCreated, cc)
}

func UpdateCharacterClass(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("ccid"), 10, 64)

	entCC, err := db.Client.CharacterClass.Query().Where(characterclass.ID(id)).Select(characterclass.FieldCharacterID).Only(c.Request.Context())
	if ent.IsNotFound(err) {
		WriteNotFound(c, "class not found")
		return
	}
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	charID := entCC.CharacterID

	entChar, err := db.Client.Character.Query().Where(character.ID(charID)).Select(character.FieldUserID, character.FieldCharacterType).Only(c.Request.Context())
	if err != nil {
		WriteNotFound(c, "class not found")
		return
	}
	if !canEditCharacter(c, entChar) {
		WriteError(c, http.StatusForbidden, strErr("access denied"))
		return
	}

	var cc models.CharClass
	if !BindOr400(c, &cc) {
		return
	}
	db.Client.CharacterClass.UpdateOneID(id).
		SetClass(cc.Class).
		SetSubclass(cc.Subclass).
		SetLevel(cc.Level).
		SetHitDice(cc.HitDice).
		Exec(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteCharacterClass(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("ccid"), 10, 64)

	entCC, err := db.Client.CharacterClass.Query().Where(characterclass.ID(id)).Select(characterclass.FieldCharacterID).Only(c.Request.Context())
	if ent.IsNotFound(err) {
		WriteNotFound(c, "class not found")
		return
	}
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	charID := entCC.CharacterID

	entChar, err := db.Client.Character.Query().Where(character.ID(charID)).Select(character.FieldUserID, character.FieldCharacterType).Only(c.Request.Context())
	if err != nil {
		WriteNotFound(c, "class not found")
		return
	}
	if !canEditCharacter(c, entChar) {
		WriteError(c, http.StatusForbidden, strErr("access denied"))
		return
	}
	db.Client.CharacterClass.DeleteOneID(id).Exec(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Import

func UpdateCurrency(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if !canEditCharacterID(c, id) {
		WriteError(c, http.StatusForbidden, strErr("access denied"))
		return
	}
	var cur models.Currency
	if !BindOr400(c, &cur) {
		return
	}
	if cur.CP < 0 || cur.SP < 0 || cur.EP < 0 || cur.GP < 0 || cur.PP < 0 {
		WriteError(c, http.StatusBadRequest, strErr("currency values cannot be negative"))
		return
	}
	db.Client.CharacterCurrency.Update().
		Where(charactercurrency.CharacterID(id)).
		SetCp(cur.CP).
		SetSp(cur.SP).
		SetEp(cur.EP).
		SetGp(cur.GP).
		SetPp(cur.PP).
		Exec(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func UpdateSpellcasting(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if !canEditCharacterID(c, id) {
		WriteError(c, http.StatusForbidden, strErr("access denied"))
		return
	}
	var sc models.Spellcasting
	if !BindOr400(c, &sc) {
		return
	}

	// Auto-calc save DC and attack bonus from ability + proficiency
	entChar, err := db.Client.Character.Query().
		Where(character.ID(id)).
		Select(character.FieldProficiencyBonus, character.FieldStr, character.FieldDex, character.FieldCon, character.FieldInt, character.FieldWis, character.FieldCha).
		Only(c.Request.Context())
	if err == nil {
		var abilMod int
		switch sc.Ability {
		case "str":
			abilMod = abilityMod(entChar.Str)
		case "dex":
			abilMod = abilityMod(entChar.Dex)
		case "con":
			abilMod = abilityMod(entChar.Con)
		case "int":
			abilMod = abilityMod(entChar.Int)
		case "wis":
			abilMod = abilityMod(entChar.Wis)
		case "cha":
			abilMod = abilityMod(entChar.Cha)
		}
		if sc.SaveDC == 0 {
			sc.SaveDC = 8 + entChar.ProficiencyBonus + abilMod
		}
		if sc.AttackBonus == 0 {
			sc.AttackBonus = entChar.ProficiencyBonus + abilMod
		}
	}

	ctx := c.Request.Context()
	existing, err := db.Client.CharacterSpellcasting.Query().
		Where(characterspellcasting.CharacterID(id)).
		Only(ctx)
	if err == nil && existing != nil {
		db.Client.CharacterSpellcasting.UpdateOneID(existing.ID).
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
			Exec(ctx)
	} else {
		db.Client.CharacterSpellcasting.Create().
			SetCharacterID(id).
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
			Exec(ctx)
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Inventory sub-resource handlers

func CreateInventory(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if !canEditCharacterID(c, charID) {
		WriteError(c, http.StatusForbidden, strErr("access denied"))
		return
	}
	var item models.InventoryItem
	if !BindOr400(c, &item) {
		return
	}
	if strings.TrimSpace(item.Name) == "" {
		WriteError(c, http.StatusBadRequest, strErr("name is required"))
		return
	}
	result, err := db.Client.InventoryItem.Create().
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
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": result.ID})
}

func UpdateInventory(c *gin.Context) {
	iid, _ := strconv.ParseInt(c.Param("iid"), 10, 64)
	if !canEditResourceID(c, "inventory", iid) {
		WriteError(c, http.StatusForbidden, strErr("access denied"))
		return
	}
	var item models.InventoryItem
	if !BindOr400(c, &item) {
		return
	}
	_, err := db.Client.InventoryItem.UpdateOneID(iid).
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
		Save(c.Request.Context())
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteInventory(c *gin.Context) {
	iid, _ := strconv.ParseInt(c.Param("iid"), 10, 64)
	entItem, err := db.Client.InventoryItem.Query().Where(inventoryitem.ID(iid)).Select(inventoryitem.FieldCharacterID).Only(c.Request.Context())
	if ent.IsNotFound(err) {
		WriteNotFound(c, "item not found")
		return
	}
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	if !canEditCharacterID(c, entItem.CharacterID) {
		WriteError(c, http.StatusForbidden, strErr("access denied"))
		return
	}
	db.Client.InventoryItem.DeleteOneID(iid).Exec(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Spell sub-resource handlers

func CreateSpell(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if !canEditCharacterID(c, charID) {
		WriteError(c, http.StatusForbidden, strErr("access denied"))
		return
	}
	var sp models.Spell
	if !BindOr400(c, &sp) {
		return
	}
	if sp.Level < 0 || sp.Level > 9 {
		WriteError(c, http.StatusBadRequest, strErr("spell level must be between 0 and 9"))
		return
	}
	result, err := db.Client.Spell.Create().
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
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": result.ID})
}

func UpdateSpell(c *gin.Context) {
	sid, _ := strconv.ParseInt(c.Param("sid"), 10, 64)
	if !canEditResourceID(c, "spells", sid) {
		WriteError(c, http.StatusForbidden, strErr("access denied"))
		return
	}
	var sp models.Spell
	if !BindOr400(c, &sp) {
		return
	}
	_, err := db.Client.Spell.UpdateOneID(sid).
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
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteSpell(c *gin.Context) {
	sid, _ := strconv.ParseInt(c.Param("sid"), 10, 64)
	entSpell, err := db.Client.Spell.Query().Where(spell.ID(sid)).Select(spell.FieldCharacterID).Only(c.Request.Context())
	if ent.IsNotFound(err) {
		WriteNotFound(c, "spell not found")
		return
	}
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	if !canEditCharacterID(c, entSpell.CharacterID) {
		WriteError(c, http.StatusForbidden, strErr("access denied"))
		return
	}
	db.Client.Spell.DeleteOneID(sid).Exec(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Feature sub-resource handlers

func CreateFeature(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if !canEditCharacterID(c, charID) {
		WriteError(c, http.StatusForbidden, strErr("access denied"))
		return
	}
	var f models.Feature
	if !BindOr400(c, &f) {
		return
	}
	result, err := db.Client.CharacterFeature.Create().
		SetCharacterID(charID).
		SetName(f.Name).
		SetDescription(f.Description).
		SetSource(f.Source).
		SetLevelGained(f.LevelGained).
		Save(c.Request.Context())
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": result.ID, "name": f.Name})
}

func UpdateFeature(c *gin.Context) {
	fid, _ := strconv.ParseInt(c.Param("fid"), 10, 64)
	if !canEditResourceID(c, "character_features", fid) {
		WriteError(c, http.StatusForbidden, strErr("access denied"))
		return
	}
	var f models.Feature
	if !BindOr400(c, &f) {
		return
	}
	_, err := db.Client.CharacterFeature.UpdateOneID(fid).
		SetName(f.Name).
		SetDescription(f.Description).
		SetSource(f.Source).
		SetLevelGained(f.LevelGained).
		Save(c.Request.Context())
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteFeature(c *gin.Context) {
	fid, _ := strconv.ParseInt(c.Param("fid"), 10, 64)
	entFeature, err := db.Client.CharacterFeature.Query().Where(characterfeature.ID(fid)).Select(characterfeature.FieldCharacterID).Only(c.Request.Context())
	if ent.IsNotFound(err) {
		WriteNotFound(c, "feature not found")
		return
	}
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	if !canEditCharacterID(c, entFeature.CharacterID) {
		WriteError(c, http.StatusForbidden, strErr("access denied"))
		return
	}
	db.Client.CharacterFeature.DeleteOneID(fid).Exec(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Proficiency handlers

func CreateProficiency(c *gin.Context) {
	var p models.Proficiency
	if !BindOr400(c, &p) {
		return
	}
	if !canEditCharacterID(c, p.CharacterID) {
		WriteError(c, http.StatusForbidden, strErr("access denied"))
		return
	}
	result, err := db.Client.CharacterProficiency.Create().
		SetCharacterID(p.CharacterID).
		SetType(p.Type).
		SetName(p.Name).
		Save(c.Request.Context())
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": result.ID})
}

func DeleteProficiency(c *gin.Context) {
	pid, _ := strconv.ParseInt(c.Param("pid"), 10, 64)
	entProf, err := db.Client.CharacterProficiency.Query().Where(characterproficiency.ID(pid)).Select(characterproficiency.FieldCharacterID).Only(c.Request.Context())
	if ent.IsNotFound(err) {
		WriteNotFound(c, "proficiency not found")
		return
	}
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	if !canEditCharacterID(c, entProf.CharacterID) {
		WriteError(c, http.StatusForbidden, strErr("access denied"))
		return
	}
	db.Client.CharacterProficiency.DeleteOneID(pid).Exec(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Import from JSON body (list)

func UpdateExhaustion(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if !canEditCharacterID(c, charID) {
		WriteError(c, http.StatusForbidden, strErr("access denied"))
		return
	}
	var req struct {
		Level int `json:"level"`
	}
	if !BindOr400(c, &req) {
		return
	}
	if req.Level < 0 || req.Level > 6 {
		WriteError(c, http.StatusBadRequest, strErr("exhaustion level must be 0-6"))
		return
	}
	_, err := db.Client.Character.UpdateOneID(charID).SetExhaustionLevel(req.Level).Save(c.Request.Context())
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	SendCharacterUpdate(charID)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ─── Spell Preparation ───

func BatchPrepareSpells(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if !canEditCharacterID(c, charID) {
		WriteError(c, http.StatusForbidden, strErr("access denied"))
		return
	}
	var req struct {
		SpellIDs []int64 `json:"spell_ids"`
	}
	if !BindOr400(c, &req) {
		return
	}
	ctx := c.Request.Context()
	// First, unprepare all spells for this character
	_, err := db.Client.Spell.Update().Where(spell.CharacterID(charID)).SetPrepared(false).Save(ctx)
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	// Then set prepared for the specified spell IDs
	if len(req.SpellIDs) > 0 {
		_, err = db.Client.Spell.Update().Where(spell.IDIn(req.SpellIDs...), spell.CharacterID(charID)).SetPrepared(true).Save(ctx)
		if err != nil {
			WriteError(c, http.StatusInternalServerError, err)
			return
		}
	}
	SendCharacterUpdate(charID)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
