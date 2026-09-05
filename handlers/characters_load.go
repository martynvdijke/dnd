package handlers

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
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

func canEditCharacter(c *gin.Context, e *ent.Character) bool {
	role, _ := c.Get("role")
	if role == "admin" {
		return true
	}
	currentUID, _ := c.Get("user_id")
	uid, _ := currentUID.(int64)
	if e.CharacterType == "linked" {
		return isDMOfCharacter(c, e.ID)
	}
	return e.UserID == uid || isDMOfCharacter(c, e.ID)
}

// canEditCharacterID loads the character and applies canEditCharacter.
func canEditCharacterID(c *gin.Context, characterID int64) bool {
	e, err := db.Client.Character.Query().
		Where(character.ID(characterID)).
		Select(character.FieldUserID, character.FieldCharacterType).
		Only(c.Request.Context())
	if err != nil {
		return false
	}
	return canEditCharacter(c, e)
}

// canEditResourceID resolves the owning character of a sub-resource row and
// applies canEditCharacterID. table must have a character_id column.
func canEditResourceID(c *gin.Context, table string, rowID int64) bool {
	var charID int64
	err := db.DB.QueryRow(fmt.Sprintf("SELECT character_id FROM %s WHERE id=?", table), rowID).Scan(&charID)
	if err != nil {
		return false
	}
	return canEditCharacterID(c, charID)
}

// checkCharacterAccess verifies the current user owns (or is admin/DM of) the given character
func checkCharacterAccess(c *gin.Context, characterID int64) bool {
	userID, _ := c.Get("user_id")
	currentUID, _ := userID.(int64)
	role, _ := c.Get("role")

	entChar, err := db.Client.Character.Query().
		Where(character.ID(characterID)).
		Select(character.FieldUserID).
		Only(c.Request.Context())
	if err != nil {
		return false
	}
	return role == "admin" || entChar.UserID == currentUID || isDMOfCharacter(c, characterID)
}

func entCharacterToModel(e *ent.Character) *models.Character {
	ch := &models.Character{
		ID:                     e.ID,
		UserID:                 e.UserID,
		Name:                   e.Name,
		Race:                   e.Race,
		Class:                  e.Class,
		Subclass:               e.Subclass,
		Level:                  e.Level,
		XP:                     e.Xp,
		Background:             e.Background,
		Alignment:              e.Alignment,
		Str:                    e.Str,
		Dex:                    e.Dex,
		Con:                    e.Con,
		Int:                    e.Int,
		Wis:                    e.Wis,
		Cha:                    e.Cha,
		AC:                     e.Ac,
		Initiative:             e.Initiative,
		Speed:                  e.Speed,
		HPMax:                  e.HpMax,
		HPCurrent:              e.HpCurrent,
		TempHP:                 e.TempHp,
		HitDice:                e.HitDice,
		HitDiceCurrent:         e.HitDiceCurrent,
		ProficiencyBonus:       e.ProficiencyBonus,
		Inspiration:            e.Inspiration,
		PassivePerception:      e.PassivePerception,
		PersonalityTraits:      e.PersonalityTraits,
		Ideals:                 e.Ideals,
		Bonds:                  e.Bonds,
		Flaws:                  e.Flaws,
		Appearance:             e.Appearance,
		Backstory:              e.Backstory,
		PortraitURL:            e.PortraitURL,
		CreatedAt:              e.CreatedAt,
		UpdatedAt:              e.UpdatedAt,
		DeathSavesSuccesses:    e.DeathSavesSuccesses,
		DeathSavesFailures:     e.DeathSavesFailures,
		ConcentratingOn:        e.ConcentratingOn,
		ExhaustionLevel:        e.ExhaustionLevel,
		CharacterType:          e.CharacterType,
		CompendiumRaceID:       e.CompendiumRaceID,
		CompendiumClassID:      e.CompendiumClassID,
		CompendiumBackgroundID: e.CompendiumBackgroundID,
	}
	if e.CampaignID != 0 {
		ch.CampaignID = &e.CampaignID
	}
	return ch
}

// Internal load helpers

func loadProficiencies(ctx context.Context, characterID int64) []models.Proficiency {
	ents, err := db.Client.CharacterProficiency.Query().Where(characterproficiency.CharacterID(characterID)).All(ctx)
	if err != nil {
		return nil
	}
	out := make([]models.Proficiency, 0, len(ents))
	for _, e := range ents {
		out = append(out, models.Proficiency{ID: e.ID, CharacterID: e.CharacterID, Type: e.Type, Name: e.Name})
	}
	return out
}

// loadEntryIDs returns map[rowID]entryID for rows of a character table that are
// linked to a generic compendium entry. The ent schemas predate the
// compendium_entry_id column, so it is read via supplementary raw SQL.
func loadEntryIDs(table string, characterID int64) map[int64]int64 {
	m := map[int64]int64{}
	rows, err := db.DB.Query("SELECT id, compendium_entry_id FROM "+table+" WHERE character_id=? AND compendium_entry_id IS NOT NULL", characterID)
	if err != nil {
		return m
	}
	defer rows.Close()
	for rows.Next() {
		var id, eid int64
		if rows.Scan(&id, &eid) == nil {
			m[id] = eid
		}
	}
	return m
}

func loadFeatures(ctx context.Context, characterID int64) []models.Feature {
	ents, err := db.Client.CharacterFeature.Query().Where(characterfeature.CharacterID(characterID)).All(ctx)
	if err != nil {
		return nil
	}
	out := make([]models.Feature, 0, len(ents))
	entryIDs := loadEntryIDs("character_features", characterID)
	for _, e := range ents {
		f := models.Feature{ID: e.ID, CharacterID: e.CharacterID, Name: e.Name, Description: e.Description, Source: e.Source, LevelGained: e.LevelGained}
		if eid, ok := entryIDs[e.ID]; ok {
			f.CompendiumEntryID = &eid
		}
		out = append(out, f)
	}
	return out
}

func loadSpellcasting(ctx context.Context, characterID int64) *models.Spellcasting {
	e, err := db.Client.CharacterSpellcasting.Query().Where(characterspellcasting.CharacterID(characterID)).Only(ctx)
	if err != nil {
		return nil
	}
	return &models.Spellcasting{
		CharacterID: e.CharacterID,
		Ability:     e.Ability,
		SaveDC:      e.SaveDc,
		AttackBonus: e.AttackBonus,
		Slots1Max:   e.Slots1Max, Slots1Used: e.Slots1Used,
		Slots2Max: e.Slots2Max, Slots2Used: e.Slots2Used,
		Slots3Max: e.Slots3Max, Slots3Used: e.Slots3Used,
		Slots4Max: e.Slots4Max, Slots4Used: e.Slots4Used,
		Slots5Max: e.Slots5Max, Slots5Used: e.Slots5Used,
		Slots6Max: e.Slots6Max, Slots6Used: e.Slots6Used,
		Slots7Max: e.Slots7Max, Slots7Used: e.Slots7Used,
		Slots8Max: e.Slots8Max, Slots8Used: e.Slots8Used,
		Slots9Max: e.Slots9Max, Slots9Used: e.Slots9Used,
	}
}

func loadSpells(ctx context.Context, characterID int64) []models.Spell {
	ents, err := db.Client.Spell.Query().Where(spell.CharacterID(characterID)).Order(spell.ByLevel(), spell.ByName()).All(ctx)
	if err != nil {
		return nil
	}
	out := make([]models.Spell, 0, len(ents))
	entryIDs := loadEntryIDs("spells", characterID)
	for _, e := range ents {
		s := models.Spell{
			ID: e.ID, CharacterID: e.CharacterID, Name: e.Name, Level: e.Level, School: e.School,
			CastingTime: e.CastingTime, Range: e.Range, Components: e.Components, Duration: e.Duration,
			Description: e.Description, Prepared: e.Prepared, AlwaysPrepared: e.AlwaysPrepared, Source: e.Source, Notes: e.Notes,
			CompendiumSpellID: e.CompendiumSpellID,
		}
		if eid, ok := entryIDs[e.ID]; ok {
			s.CompendiumEntryID = &eid
		}
		out = append(out, s)
	}
	return out
}

func loadInventory(ctx context.Context, characterID int64) []models.InventoryItem {
	ents, err := db.Client.InventoryItem.Query().Where(inventoryitem.CharacterID(characterID)).Order(ent.Desc(inventoryitem.FieldIsEquipped), inventoryitem.ByName()).All(ctx)
	if err != nil {
		return nil
	}
	out := make([]models.InventoryItem, 0, len(ents))
	entryIDs := loadEntryIDs("inventory", characterID)
	for _, e := range ents {
		it := models.InventoryItem{
			ID: e.ID, CharacterID: e.CharacterID, Name: e.Name, Quantity: e.Quantity, Weight: e.Weight,
			Category: e.Category, DamageDice: e.DamageDice, DamageType: e.DamageType, WeaponProperties: e.WeaponProperties,
			ACBonus: e.AcBonus, ArmorType: e.ArmorType, Description: e.Description,
			IsEquipped: e.IsEquipped, IsMagical: e.IsMagical, Attunement: e.Attunement, IsIdentified: e.IsIdentified, Notes: e.Notes,
			CompendiumEquipmentID: e.CompendiumEquipmentID,
		}
		if eid, ok := entryIDs[e.ID]; ok {
			it.CompendiumEntryID = &eid
		}
		out = append(out, it)
	}
	return out
}

func loadCurrency(ctx context.Context, characterID int64) *models.Currency {
	e, err := db.Client.CharacterCurrency.Query().Where(charactercurrency.CharacterID(characterID)).Only(ctx)
	if err != nil {
		return nil
	}
	return &models.Currency{CharacterID: e.CharacterID, CP: e.Cp, SP: e.Sp, EP: e.Ep, GP: e.Gp, PP: e.Pp}
}

func loadCharClasses(ctx context.Context, characterID int64) []models.CharClass {
	ents, err := db.Client.CharacterClass.Query().Where(characterclass.CharacterID(characterID)).Order(characterclass.ByCreatedAt()).All(ctx)
	if err != nil {
		return nil
	}
	out := make([]models.CharClass, 0, len(ents))
	for _, e := range ents {
		out = append(out, models.CharClass{ID: e.ID, CharacterID: e.CharacterID, Class: e.Class, Subclass: e.Subclass, Level: e.Level, HitDice: e.HitDice})
	}
	return out
}

// Multi-class handlers
