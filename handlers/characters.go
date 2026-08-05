package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"villum/db"
	"villum/ent"
	"villum/ent/campaignmember"
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

func abilityMod(score int) int {
	return int(math.Floor(float64(score-10) / 2.0))
}

func computeMods(ch *models.Character) {
	ch.StrMod = abilityMod(ch.Str)
	ch.DexMod = abilityMod(ch.Dex)
	ch.ConMod = abilityMod(ch.Con)
	ch.IntMod = abilityMod(ch.Int)
	ch.WisMod = abilityMod(ch.Wis)
	ch.ChaMod = abilityMod(ch.Cha)
	if ch.Spellcasting != nil && ch.Spellcasting.Ability != "" {
		abilMod := 0
		switch ch.Spellcasting.Ability {
		case "str":
			abilMod = ch.StrMod
		case "dex":
			abilMod = ch.DexMod
		case "con":
			abilMod = ch.ConMod
		case "int":
			abilMod = ch.IntMod
		case "wis":
			abilMod = ch.WisMod
		case "cha":
			abilMod = ch.ChaMod
		}
		ch.SpellSaveDC = 8 + ch.ProficiencyBonus + abilMod
		ch.SpellAttackBonus = ch.ProficiencyBonus + abilMod
	}
}

func ListCharacters(c *gin.Context) {
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	type CharSummary struct {
		ID          int64  `json:"id"`
		UserID      int64  `json:"user_id"`
		Name        string `json:"name"`
		Race        string `json:"race"`
		Class       string `json:"class"`
		Level       int    `json:"level"`
		HPMax       int    `json:"hp_max"`
		HPCurrent   int    `json:"hp_current"`
		PortraitURL   string `json:"portrait_url,omitempty"`
		RaceColor     string `json:"race_color,omitempty"`
		CharacterType string `json:"character_type"`
		CanEdit       bool   `json:"can_edit"`
	}
	chars := []CharSummary{}
	raceColors := GetRaceColorMap()

	if role == "admin" {
		query := c.DefaultQuery("q", "")
		if query != "" {
			rows, err := db.DB.Query(`
				SELECT c.id, c.user_id, c.name, c.race, c.class, c.level, c.hp_max, c.hp_current, COALESCE(c.portrait_url,''), COALESCE(c.character_type,'player')
				FROM characters c JOIN characters_fts fts ON c.id = fts.rowid
				WHERE characters_fts MATCH ? ORDER BY c.updated_at DESC`, query)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			defer rows.Close()
			for rows.Next() {
				var ch CharSummary
				rows.Scan(&ch.ID, &ch.UserID, &ch.Name, &ch.Race, &ch.Class, &ch.Level, &ch.HPMax, &ch.HPCurrent, &ch.PortraitURL, &ch.CharacterType)
				ch.RaceColor = raceColors[ch.Race]
				ch.CanEdit = true
				chars = append(chars, ch)
			}
		} else {
			entChars, err := db.Client.Character.Query().Order(ent.Desc(character.FieldUpdatedAt)).All(c.Request.Context())
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			for _, e := range entChars {
				ch := CharSummary{ID: e.ID, UserID: e.UserID, Name: e.Name, Race: e.Race, Class: e.Class, Level: e.Level, HPMax: e.HpMax, HPCurrent: e.HpCurrent, PortraitURL: e.PortraitURL, CharacterType: e.CharacterType, CanEdit: canEditCharacter(c, e)}
				ch.RaceColor = raceColors[ch.Race]
				chars = append(chars, ch)
			}
		}
	} else {
		uid, _ := userID.(int64)
		entChars, err := db.Client.Character.Query().Where(character.UserID(uid)).Order(ent.Desc(character.FieldUpdatedAt)).All(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for _, e := range entChars {
			ch := CharSummary{ID: e.ID, UserID: e.UserID, Name: e.Name, Race: e.Race, Class: e.Class, Level: e.Level, HPMax: e.HpMax, HPCurrent: e.HpCurrent, PortraitURL: e.PortraitURL, CharacterType: e.CharacterType, CanEdit: canEditCharacter(c, e)}
			ch.RaceColor = raceColors[ch.Race]
			chars = append(chars, ch)
		}
	}
	c.JSON(http.StatusOK, chars)
}

func ListAllCharacters(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "dm" && role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "dm or admin required"})
		return
	}

	type CharSummary struct {
		ID          int64  `json:"id"`
		UserID      int64  `json:"user_id"`
		Username    string `json:"username"`
		Name        string `json:"name"`
		Race        string `json:"race"`
		Class       string `json:"class"`
		Level       int    `json:"level"`
		HPMax       int    `json:"hp_max"`
		HPCurrent   int    `json:"hp_current"`
		PortraitURL   string `json:"portrait_url,omitempty"`
		RaceColor     string `json:"race_color,omitempty"`
		CharacterType string `json:"character_type"`
		CanEdit       bool   `json:"can_edit"`
	}
	chars := make([]CharSummary, 0)
	raceColors := GetRaceColorMap()

	rows, err := db.DB.Query(`
		SELECT c.id, c.user_id, u.username, c.name, c.race, c.class, c.level, c.hp_max, c.hp_current, COALESCE(c.portrait_url,''), COALESCE(c.character_type,'player')
		FROM characters c
		JOIN users u ON c.user_id = u.id
		ORDER BY c.updated_at DESC`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	for rows.Next() {
		var ch CharSummary
		rows.Scan(&ch.ID, &ch.UserID, &ch.Username, &ch.Name, &ch.Race, &ch.Class, &ch.Level, &ch.HPMax, &ch.HPCurrent, &ch.PortraitURL, &ch.CharacterType)
		ch.RaceColor = raceColors[ch.Race]
		ch.CanEdit = true
		chars = append(chars, ch)
	}
	c.JSON(http.StatusOK, chars)
}

func GetCharacter(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	entChar, err := db.Client.Character.Query().Where(character.ID(id)).Only(c.Request.Context())
	if ent.IsNotFound(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ch := entCharacterToModel(entChar)

	// Authorization — admin, owner, or campaign DM may view; edit rights depend on character_type
	role, _ := c.Get("role")
	uidVal, _ := c.Get("user_id")
	uid, _ := uidVal.(int64)
	if role != "admin" && entChar.UserID != uid && !isDMOfCharacter(c, id) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Load sub-resources
	ctx := c.Request.Context()
	ch.Proficiencies = loadProficiencies(ctx, ch.ID)
	ch.Features = loadFeatures(ctx, ch.ID)
	ch.Spellcasting = loadSpellcasting(ctx, ch.ID)
	ch.Spells = loadSpells(ctx, ch.ID)
	ch.Inventory = loadInventory(ctx, ch.ID)
	ch.Currency = loadCurrency(ctx, ch.ID)
	ch.Classes = loadCharClasses(ctx, ch.ID)
	computeMods(ch)
	ch.CanEdit = canEditCharacter(c, entChar)

	c.JSON(http.StatusOK, ch)
}

func CreateCharacter(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var ch models.Character
	if err := c.ShouldBindJSON(&ch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if strings.TrimSpace(ch.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	if ch.CharacterType == "" {
		ch.CharacterType = "player"
	}
	if ch.CharacterType != "player" && ch.CharacterType != "linked" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "character_type must be 'player' or 'linked'"})
		return
	}
	for _, score := range []int{ch.Str, ch.Dex, ch.Con, ch.Int, ch.Wis, ch.Cha} {
		if score < 0 || score > 30 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ability scores must be between 0 and 30"})
			return
		}
	}

	// Set defaults
	if ch.Level < 1 {
		ch.Level = 1
	}
	if ch.AC < 1 {
		ch.AC = 10
	}
	if ch.Speed < 1 {
		ch.Speed = 30
	}
	if ch.ProficiencyBonus < 1 {
		if ch.Level >= 17 {
			ch.ProficiencyBonus = 6
		} else if ch.Level >= 13 {
			ch.ProficiencyBonus = 5
		} else if ch.Level >= 9 {
			ch.ProficiencyBonus = 4
		} else if ch.Level >= 5 {
			ch.ProficiencyBonus = 3
		} else {
			ch.ProficiencyBonus = 2
		}
	}
	if ch.HPMax < 1 {
		ch.HPMax = 10
		ch.HPCurrent = 10
	}
	if ch.HitDice == "" {
		ch.HitDice = "1d10"
		ch.HitDiceCurrent = 1
	}

	uid, _ := userID.(int64)
	now := time.Now().Format("2006-01-02 15:04:05")

	charCreate := db.Client.Character.Create().
		SetUserID(uid).
		SetName(ch.Name).
		SetRace(ch.Race).
		SetClass(ch.Class).
		SetSubclass(ch.Subclass).
		SetLevel(ch.Level).
		SetXp(ch.XP).
		SetBackground(ch.Background).
		SetAlignment(ch.Alignment).
		SetStr(ch.Str).
		SetDex(ch.Dex).
		SetCon(ch.Con).
		SetInt(ch.Int).
		SetWis(ch.Wis).
		SetCha(ch.Cha).
		SetAc(ch.AC).
		SetInitiative(ch.Initiative).
		SetSpeed(ch.Speed).
		SetHpMax(ch.HPMax).
		SetHpCurrent(ch.HPCurrent).
		SetTempHp(ch.TempHP).
		SetHitDice(ch.HitDice).
		SetHitDiceCurrent(ch.HitDiceCurrent).
		SetProficiencyBonus(ch.ProficiencyBonus).
		SetInspiration(ch.Inspiration).
		SetPassivePerception(ch.PassivePerception).
		SetDeathSavesSuccesses(ch.DeathSavesSuccesses).
		SetDeathSavesFailures(ch.DeathSavesFailures).
		SetConcentratingOn(ch.ConcentratingOn).
		SetPersonalityTraits(ch.PersonalityTraits).
		SetIdeals(ch.Ideals).
		SetBonds(ch.Bonds).
		SetFlaws(ch.Flaws).
		SetAppearance(ch.Appearance).
		SetBackstory(ch.Backstory).
		SetPortraitURL(ch.PortraitURL).
		SetCreatedAt(now).
		SetUpdatedAt(now).
		SetCharacterType(ch.CharacterType)

	if ch.CampaignID != nil {
		charCreate.SetCampaignID(*ch.CampaignID)
	}

	char, err := charCreate.Save(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id := char.ID

	// Create default currency entry
	db.Client.CharacterCurrency.Create().SetCharacterID(id).Save(c.Request.Context())

	ch.ID = id
	ch.UserID = uid
	c.JSON(http.StatusCreated, ch)
}

func UpdateCharacter(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	// Check ownership
	entChar, err := db.Client.Character.Query().Where(character.ID(id)).Select(character.FieldUserID, character.FieldCharacterType).Only(c.Request.Context())
	if ent.IsNotFound(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !canEditCharacter(c, entChar) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var ch models.Character
	if err := c.ShouldBindJSON(&ch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if ch.CharacterType != "" && ch.CharacterType != "player" && ch.CharacterType != "linked" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "character_type must be 'player' or 'linked'"})
		return
	}
	if strings.TrimSpace(ch.Name) == "" || strings.TrimSpace(ch.Race) == "" || strings.TrimSpace(ch.Class) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name, race, and class are required"})
		return
	}
	if ch.DeathSavesSuccesses < 0 || ch.DeathSavesSuccesses > 3 || ch.DeathSavesFailures < 0 || ch.DeathSavesFailures > 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "death saves must be between 0 and 3"})
		return
	}

	upd := db.Client.Character.UpdateOneID(id).
		SetName(ch.Name).
		SetRace(ch.Race).
		SetClass(ch.Class).
		SetSubclass(ch.Subclass).
		SetLevel(ch.Level).
		SetXp(ch.XP).
		SetBackground(ch.Background).
		SetAlignment(ch.Alignment).
		SetStr(ch.Str).
		SetDex(ch.Dex).
		SetCon(ch.Con).
		SetInt(ch.Int).
		SetWis(ch.Wis).
		SetCha(ch.Cha).
		SetAc(ch.AC).
		SetInitiative(ch.Initiative).
		SetSpeed(ch.Speed).
		SetHpMax(ch.HPMax).
		SetHpCurrent(ch.HPCurrent).
		SetTempHp(ch.TempHP).
		SetHitDice(ch.HitDice).
		SetHitDiceCurrent(ch.HitDiceCurrent).
		SetProficiencyBonus(ch.ProficiencyBonus).
		SetInspiration(ch.Inspiration).
		SetPassivePerception(ch.PassivePerception).
		SetDeathSavesSuccesses(ch.DeathSavesSuccesses).
		SetDeathSavesFailures(ch.DeathSavesFailures).
		SetConcentratingOn(ch.ConcentratingOn).
		SetPersonalityTraits(ch.PersonalityTraits).
		SetIdeals(ch.Ideals).
		SetBonds(ch.Bonds).
		SetFlaws(ch.Flaws).
		SetAppearance(ch.Appearance).
		SetBackstory(ch.Backstory).
		SetPortraitURL(ch.PortraitURL).
		SetUpdatedAt(time.Now().Format("2006-01-02 15:04:05"))

	if ch.CampaignID != nil {
		upd.SetCampaignID(*ch.CampaignID)
	} else {
		upd.ClearCampaignID()
	}

	_, err = upd.Save(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Auto-calc passive perception
	charStats, err := db.Client.Character.Query().Where(character.ID(id)).Select(character.FieldWis, character.FieldProficiencyBonus).Only(c.Request.Context())
	if err == nil {
		wisMod := abilityMod(charStats.Wis)
		pp := 10 + wisMod
		perceptionCount, _ := db.Client.CharacterProficiency.Query().
			Where(
				characterproficiency.CharacterID(id),
				characterproficiency.TypeEQ("skill"),
				characterproficiency.NameEqualFold("perception"),
			).
			Count(c.Request.Context())
		if perceptionCount > 0 {
			pp += charStats.ProficiencyBonus
		}
		db.Client.Character.UpdateOneID(id).SetPassivePerception(pp).Exec(c.Request.Context())
	}

	SendCharacterUpdate(id)
	SendPartyUpdate()

	updated, err := db.Client.Character.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}
	c.JSON(http.StatusOK, entCharacterToModel(updated))
}

func DeleteCharacter(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	entChar, err := db.Client.Character.Query().Where(character.ID(id)).Select(character.FieldUserID, character.FieldCharacterType).Only(c.Request.Context())
	if ent.IsNotFound(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !canEditCharacter(c, entChar) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Delete child records first (Ent auto-migration creates FK constraints with NoAction,
	// so we must delete children manually before the parent character record).
	for _, table := range []string{
		"character_currency", "character_classes", "character_proficiencies",
		"character_locations", "character_npcs", "sessions", "quests", "journal",
		"rest_log", "character_conditions", "character_feats", "companions",
		"faction_reputation", "character_notes", "character_crafting",
		"character_resources", "downtime_activities", "level_up_plans",
		"character_spellcasting", "character_features", "spells", "inventory",
	} {
		db.DB.Exec("DELETE FROM "+table+" WHERE character_id = ?", id)
	}
	if err := db.Client.Character.DeleteOneID(id).Exec(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func ExportCharacter(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	entChar, err := db.Client.Character.Query().Where(character.ID(id)).Only(c.Request.Context())
	if ent.IsNotFound(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ch := entCharacterToModel(entChar)

	ctx := c.Request.Context()
	ch.Proficiencies = loadProficiencies(ctx, ch.ID)
	ch.Features = loadFeatures(ctx, ch.ID)
	ch.Spellcasting = loadSpellcasting(ctx, ch.ID)
	ch.Spells = loadSpells(ctx, ch.ID)
	ch.Inventory = loadInventory(ctx, ch.ID)
	ch.Currency = loadCurrency(ctx, ch.ID)
	computeMods(ch)

	format := c.DefaultQuery("format", "json")
	if format == "text" {
		c.String(http.StatusOK, characterToText(ch))
		return
	}
	c.JSON(http.StatusOK, ch)
}

func PrintCharacter(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	entChar, err := db.Client.Character.Query().Where(character.ID(id)).Only(c.Request.Context())
	if ent.IsNotFound(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ch := entCharacterToModel(entChar)

	ctx := c.Request.Context()
	ch.Proficiencies = loadProficiencies(ctx, ch.ID)
	ch.Features = loadFeatures(ctx, ch.ID)
	ch.Spellcasting = loadSpellcasting(ctx, ch.ID)
	ch.Spells = loadSpells(ctx, ch.ID)
	ch.Inventory = loadInventory(ctx, ch.ID)
	ch.Currency = loadCurrency(ctx, ch.ID)
	computeMods(ch)

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, characterToText(ch))
}

func characterToText(ch *models.Character) string {
	var b strings.Builder

	b.WriteString("============================================\n")
	fmt.Fprintf(&b, "  %s\n", ch.Name)
	b.WriteString("============================================\n\n")

	fmt.Fprintf(&b, "Race: %s\n", ch.Race)
	fmt.Fprintf(&b, "Class: %s", ch.Class)
	if ch.Subclass != "" {
		fmt.Fprintf(&b, " (%s)", ch.Subclass)
	}
	fmt.Fprintf(&b, "\nLevel: %d  XP: %d\n", ch.Level, ch.XP)
	fmt.Fprintf(&b, "Background: %s\n", ch.Background)
	fmt.Fprintf(&b, "Alignment: %s\n\n", ch.Alignment)

	b.WriteString("--- ABILITY SCORES ---\n")
	fmt.Fprintf(&b, "STR: %2d  DEX: %2d  CON: %2d  INT: %2d  WIS: %2d  CHA: %2d\n\n", ch.Str, ch.Dex, ch.Con, ch.Int, ch.Wis, ch.Cha)

	b.WriteString("--- COMBAT ---\n")
	fmt.Fprintf(&b, "AC: %d  Initiative: %+d  Speed: %d\n", ch.AC, ch.Initiative, ch.Speed)
	fmt.Fprintf(&b, "HP: %d/%d (Temp: %d)\n", ch.HPCurrent, ch.HPMax, ch.TempHP)
	fmt.Fprintf(&b, "Hit Dice: %s (%d remaining)\n", ch.HitDice, ch.HitDiceCurrent)
	fmt.Fprintf(&b, "Proficiency Bonus: +%d\n", ch.ProficiencyBonus)
	fmt.Fprintf(&b, "Passive Perception: %d\n\n", ch.PassivePerception)

	if ch.Spellcasting != nil && ch.Spellcasting.Ability != "" {
		b.WriteString("--- SPELLCASTING ---\n")
		fmt.Fprintf(&b, "Ability: %s  Save DC: %d  Attack Bonus: +%d\n", ch.Spellcasting.Ability, ch.Spellcasting.SaveDC, ch.Spellcasting.AttackBonus)
		slotLevels := []struct{ max, used int }{
			{ch.Spellcasting.Slots1Max, ch.Spellcasting.Slots1Used},
			{ch.Spellcasting.Slots2Max, ch.Spellcasting.Slots2Used},
			{ch.Spellcasting.Slots3Max, ch.Spellcasting.Slots3Used},
			{ch.Spellcasting.Slots4Max, ch.Spellcasting.Slots4Used},
			{ch.Spellcasting.Slots5Max, ch.Spellcasting.Slots5Used},
			{ch.Spellcasting.Slots6Max, ch.Spellcasting.Slots6Used},
			{ch.Spellcasting.Slots7Max, ch.Spellcasting.Slots7Used},
			{ch.Spellcasting.Slots8Max, ch.Spellcasting.Slots8Used},
			{ch.Spellcasting.Slots9Max, ch.Spellcasting.Slots9Used},
		}
		hasSlots := false
		for i, sl := range slotLevels {
			if sl.max > 0 {
				fmt.Fprintf(&b, "  Level %d: %d/%d slots", i+1, sl.max-sl.used, sl.max)
				hasSlots = true
			}
		}
		if !hasSlots {
			b.WriteString("  No spell slots")
		}
		b.WriteString("\n\n")
	}

	if len(ch.Spells) > 0 {
		b.WriteString("--- SPELLS ---\n")
		byLevel := make(map[int][]models.Spell)
		for _, sp := range ch.Spells {
			byLevel[sp.Level] = append(byLevel[sp.Level], sp)
		}
		for level := 0; level <= 9; level++ {
			spells, ok := byLevel[level]
			if !ok {
				continue
			}
			label := "Cantrips"
			if level > 0 {
				label = fmt.Sprintf("Level %d", level)
			}
			fmt.Fprintf(&b, "  %s:\n", label)
			for _, sp := range spells {
				prep := ""
				if sp.Prepared {
					prep = " [P]"
				}
				fmt.Fprintf(&b, "    - %s (%s)%s\n", sp.Name, sp.School, prep)
			}
		}
		b.WriteString("\n")
	}

	if len(ch.Inventory) > 0 {
		b.WriteString("--- INVENTORY ---\n")
		for _, item := range ch.Inventory {
			if item.IsEquipped {
				fmt.Fprintf(&b, "  [E] %s x%d", item.Name, item.Quantity)
			} else {
				fmt.Fprintf(&b, "  %s x%d", item.Name, item.Quantity)
			}
			if item.Category == "weapon" && item.DamageDice != "" {
				fmt.Fprintf(&b, " (%s %s)", item.DamageDice, item.DamageType)
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	if ch.Currency != nil {
		b.WriteString("--- CURRENCY ---\n")
		parts := []string{}
		if ch.Currency.PP > 0 {
			parts = append(parts, fmt.Sprintf("%d PP", ch.Currency.PP))
		}
		if ch.Currency.GP > 0 {
			parts = append(parts, fmt.Sprintf("%d GP", ch.Currency.GP))
		}
		if ch.Currency.EP > 0 {
			parts = append(parts, fmt.Sprintf("%d EP", ch.Currency.EP))
		}
		if ch.Currency.SP > 0 {
			parts = append(parts, fmt.Sprintf("%d SP", ch.Currency.SP))
		}
		if ch.Currency.CP > 0 {
			parts = append(parts, fmt.Sprintf("%d CP", ch.Currency.CP))
		}
		if len(parts) > 0 {
			fmt.Fprintf(&b, "  %s\n\n", strings.Join(parts, ", "))
		}
	}

	if len(ch.Features) > 0 {
		b.WriteString("--- FEATURES & TRAITS ---\n")
		for _, f := range ch.Features {
			fmt.Fprintf(&b, "  %s (Level %d, %s)\n", f.Name, f.LevelGained, f.Source)
			if f.Description != "" {
				fmt.Fprintf(&b, "    %s\n", f.Description)
			}
		}
		b.WriteString("\n")
	}

	if len(ch.Proficiencies) > 0 {
		b.WriteString("--- PROFICIENCIES ---\n")
		byType := make(map[string][]string)
		for _, p := range ch.Proficiencies {
			byType[p.Type] = append(byType[p.Type], p.Name)
		}
		for typ, names := range byType {
			titleCaser := cases.Title(language.English)
			fmt.Fprintf(&b, "  %s: %s\n", titleCaser.String(typ), strings.Join(names, ", "))
		}
		b.WriteString("\n")
	}

	if ch.PersonalityTraits != "" || ch.Ideals != "" || ch.Bonds != "" || ch.Flaws != "" {
		b.WriteString("--- PERSONALITY ---\n")
		if ch.PersonalityTraits != "" {
			fmt.Fprintf(&b, "  Traits: %s\n", ch.PersonalityTraits)
		}
		if ch.Ideals != "" {
			fmt.Fprintf(&b, "  Ideals: %s\n", ch.Ideals)
		}
		if ch.Bonds != "" {
			fmt.Fprintf(&b, "  Bonds: %s\n", ch.Bonds)
		}
		if ch.Flaws != "" {
			fmt.Fprintf(&b, "  Flaws: %s\n", ch.Flaws)
		}
		b.WriteString("\n")
	}

	if ch.Appearance != "" {
		b.WriteString("--- APPEARANCE ---\n")
		fmt.Fprintf(&b, "  %s\n\n", ch.Appearance)
	}

	if ch.Backstory != "" {
		b.WriteString("--- BACKSTORY ---\n")
		fmt.Fprintf(&b, "  %s\n\n", ch.Backstory)
	}

	return b.String()
}

func isDMOfCharacter(c *gin.Context, characterID int64) bool {
	currentUID, _ := c.Get("user_id")
	uid, _ := currentUID.(int64)

	entChar, err := db.Client.Character.Query().
		Where(character.ID(characterID)).
		Select(character.FieldCampaignID).
		Only(c.Request.Context())
	if err != nil {
		return false
	}

	if entChar.CampaignID != 0 {
		count, err := db.Client.CampaignMember.Query().
			Where(
				campaignmember.CampaignIDEQ(entChar.CampaignID),
				campaignmember.UserIDEQ(uid),
				campaignmember.RoleEQ("dm"),
			).
			Count(c.Request.Context())
		return err == nil && count > 0
	}
	return false
}

// isCampaignMemberOfCharacter reports whether the requester is a member of the
// character's campaign (any role).
func isCampaignMemberOfCharacter(c *gin.Context, characterID int64) bool {
	currentUID, _ := c.Get("user_id")
	uid, _ := currentUID.(int64)
	if uid == 0 {
		return false
	}
	entChar, err := db.Client.Character.Query().
		Where(character.ID(characterID)).
		Select(character.FieldCampaignID).
		Only(c.Request.Context())
	if err != nil || entChar.CampaignID == 0 {
		return false
	}
	count, err := db.Client.CampaignMember.Query().
		Where(
			campaignmember.CampaignIDEQ(entChar.CampaignID),
			campaignmember.UserIDEQ(uid),
		).
		Count(c.Request.Context())
	return err == nil && count > 0
}

// canViewCharacter: admin, owner, or any member of the character's campaign.
func canViewCharacter(c *gin.Context, characterID int64) bool {
	role, _ := c.Get("role")
	if role == "admin" {
		return true
	}
	currentUID, _ := c.Get("user_id")
	uid, _ := currentUID.(int64)
	entChar, err := db.Client.Character.Query().
		Where(character.ID(characterID)).
		Select(character.FieldUserID).
		Only(c.Request.Context())
	if err != nil {
		return false
	}
	if entChar.UserID == uid {
		return true
	}
	return isCampaignMemberOfCharacter(c, characterID)
}

// canEditCharacter enforces character_type edit rules:
//   - player: owner, admin, or campaign DM may edit
//   - linked: only admin or campaign DM may edit (owner cannot)
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
		ID:                  e.ID,
		UserID:              e.UserID,
		Name:                e.Name,
		Race:                e.Race,
		Class:               e.Class,
		Subclass:            e.Subclass,
		Level:               e.Level,
		XP:                  e.Xp,
		Background:          e.Background,
		Alignment:           e.Alignment,
		Str:                 e.Str,
		Dex:                 e.Dex,
		Con:                 e.Con,
		Int:                 e.Int,
		Wis:                 e.Wis,
		Cha:                 e.Cha,
		AC:                  e.Ac,
		Initiative:          e.Initiative,
		Speed:               e.Speed,
		HPMax:               e.HpMax,
		HPCurrent:           e.HpCurrent,
		TempHP:              e.TempHp,
		HitDice:             e.HitDice,
		HitDiceCurrent:      e.HitDiceCurrent,
		ProficiencyBonus:    e.ProficiencyBonus,
		Inspiration:         e.Inspiration,
		PassivePerception:   e.PassivePerception,
		PersonalityTraits:   e.PersonalityTraits,
		Ideals:              e.Ideals,
		Bonds:               e.Bonds,
		Flaws:               e.Flaws,
		Appearance:          e.Appearance,
		Backstory:           e.Backstory,
		PortraitURL:         e.PortraitURL,
		CreatedAt:           e.CreatedAt,
		UpdatedAt:           e.UpdatedAt,
		DeathSavesSuccesses: e.DeathSavesSuccesses,
		DeathSavesFailures:  e.DeathSavesFailures,
		ConcentratingOn:     e.ConcentratingOn,
		ExhaustionLevel:     e.ExhaustionLevel,
		CharacterType:       e.CharacterType,
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

func loadFeatures(ctx context.Context, characterID int64) []models.Feature {
	ents, err := db.Client.CharacterFeature.Query().Where(characterfeature.CharacterID(characterID)).All(ctx)
	if err != nil {
		return nil
	}
	out := make([]models.Feature, 0, len(ents))
	for _, e := range ents {
		out = append(out, models.Feature{ID: e.ID, CharacterID: e.CharacterID, Name: e.Name, Description: e.Description, Source: e.Source, LevelGained: e.LevelGained})
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
	for _, e := range ents {
		out = append(out, models.Spell{
			ID: e.ID, CharacterID: e.CharacterID, Name: e.Name, Level: e.Level, School: e.School,
			CastingTime: e.CastingTime, Range: e.Range, Components: e.Components, Duration: e.Duration,
			Description: e.Description, Prepared: e.Prepared, AlwaysPrepared: e.AlwaysPrepared, Source: e.Source, Notes: e.Notes,
			CompendiumSpellID: e.CompendiumSpellID,
		})
	}
	return out
}

func loadInventory(ctx context.Context, characterID int64) []models.InventoryItem {
	ents, err := db.Client.InventoryItem.Query().Where(inventoryitem.CharacterID(characterID)).Order(ent.Desc(inventoryitem.FieldIsEquipped), inventoryitem.ByName()).All(ctx)
	if err != nil {
		return nil
	}
	out := make([]models.InventoryItem, 0, len(ents))
	for _, e := range ents {
		out = append(out, models.InventoryItem{
			ID: e.ID, CharacterID: e.CharacterID, Name: e.Name, Quantity: e.Quantity, Weight: e.Weight,
			Category: e.Category, DamageDice: e.DamageDice, DamageType: e.DamageType, WeaponProperties: e.WeaponProperties,
			ACBonus: e.AcBonus, ArmorType: e.ArmorType, Description: e.Description,
			IsEquipped: e.IsEquipped, IsMagical: e.IsMagical, Attunement: e.Attunement, IsIdentified: e.IsIdentified, Notes: e.Notes,
			CompendiumEquipmentID: e.CompendiumEquipmentID,
		})
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

func CreateCharacterClass(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	entChar, err := db.Client.Character.Query().Where(character.ID(charID)).Select(character.FieldUserID, character.FieldCharacterType).Only(c.Request.Context())
	if ent.IsNotFound(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !canEditCharacter(c, entChar) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var cc models.CharClass
	if err := c.ShouldBindJSON(&cc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusNotFound, gin.H{"error": "class not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	charID := entCC.CharacterID

	entChar, err := db.Client.Character.Query().Where(character.ID(charID)).Select(character.FieldUserID, character.FieldCharacterType).Only(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "class not found"})
		return
	}
	if !canEditCharacter(c, entChar) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var cc models.CharClass
	if err := c.ShouldBindJSON(&cc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusNotFound, gin.H{"error": "class not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	charID := entCC.CharacterID

	entChar, err := db.Client.Character.Query().Where(character.ID(charID)).Select(character.FieldUserID, character.FieldCharacterType).Only(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "class not found"})
		return
	}
	if !canEditCharacter(c, entChar) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	db.Client.CharacterClass.DeleteOneID(id).Exec(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Import

func ImportCharacterJSON(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var imp models.ImportCharacter
	if err := c.ShouldBindJSON(&imp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON: " + err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": charID, "name": imp.Name})
}

func UpdateCurrency(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if !canEditCharacterID(c, id) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	var cur models.Currency
	if err := c.ShouldBindJSON(&cur); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if cur.CP < 0 || cur.SP < 0 || cur.EP < 0 || cur.GP < 0 || cur.PP < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "currency values cannot be negative"})
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
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	var sc models.Spellcasting
	if err := c.ShouldBindJSON(&sc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	var item models.InventoryItem
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(item.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": result.ID})
}

func UpdateInventory(c *gin.Context) {
	iid, _ := strconv.ParseInt(c.Param("iid"), 10, 64)
	if !canEditResourceID(c, "inventory", iid) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	var item models.InventoryItem
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteInventory(c *gin.Context) {
	iid, _ := strconv.ParseInt(c.Param("iid"), 10, 64)
	entItem, err := db.Client.InventoryItem.Query().Where(inventoryitem.ID(iid)).Select(inventoryitem.FieldCharacterID).Only(c.Request.Context())
	if ent.IsNotFound(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !canEditCharacterID(c, entItem.CharacterID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	db.Client.InventoryItem.DeleteOneID(iid).Exec(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Spell sub-resource handlers

func CreateSpell(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if !canEditCharacterID(c, charID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	var sp models.Spell
	if err := c.ShouldBindJSON(&sp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if sp.Level < 0 || sp.Level > 9 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "spell level must be between 0 and 9"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": result.ID})
}

func UpdateSpell(c *gin.Context) {
	sid, _ := strconv.ParseInt(c.Param("sid"), 10, 64)
	if !canEditResourceID(c, "spells", sid) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	var sp models.Spell
	if err := c.ShouldBindJSON(&sp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteSpell(c *gin.Context) {
	sid, _ := strconv.ParseInt(c.Param("sid"), 10, 64)
	entSpell, err := db.Client.Spell.Query().Where(spell.ID(sid)).Select(spell.FieldCharacterID).Only(c.Request.Context())
	if ent.IsNotFound(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "spell not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !canEditCharacterID(c, entSpell.CharacterID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	db.Client.Spell.DeleteOneID(sid).Exec(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Feature sub-resource handlers

func CreateFeature(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if !canEditCharacterID(c, charID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	var f models.Feature
	if err := c.ShouldBindJSON(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": result.ID, "name": f.Name})
}

func UpdateFeature(c *gin.Context) {
	fid, _ := strconv.ParseInt(c.Param("fid"), 10, 64)
	if !canEditResourceID(c, "character_features", fid) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	var f models.Feature
	if err := c.ShouldBindJSON(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := db.Client.CharacterFeature.UpdateOneID(fid).
		SetName(f.Name).
		SetDescription(f.Description).
		SetSource(f.Source).
		SetLevelGained(f.LevelGained).
		Save(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteFeature(c *gin.Context) {
	fid, _ := strconv.ParseInt(c.Param("fid"), 10, 64)
	entFeature, err := db.Client.CharacterFeature.Query().Where(characterfeature.ID(fid)).Select(characterfeature.FieldCharacterID).Only(c.Request.Context())
	if ent.IsNotFound(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "feature not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !canEditCharacterID(c, entFeature.CharacterID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	db.Client.CharacterFeature.DeleteOneID(fid).Exec(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Proficiency handlers

func CreateProficiency(c *gin.Context) {
	var p models.Proficiency
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !canEditCharacterID(c, p.CharacterID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	result, err := db.Client.CharacterProficiency.Create().
		SetCharacterID(p.CharacterID).
		SetType(p.Type).
		SetName(p.Name).
		Save(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": result.ID})
}

func DeleteProficiency(c *gin.Context) {
	pid, _ := strconv.ParseInt(c.Param("pid"), 10, 64)
	entProf, err := db.Client.CharacterProficiency.Query().Where(characterproficiency.ID(pid)).Select(characterproficiency.FieldCharacterID).Only(c.Request.Context())
	if ent.IsNotFound(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "proficiency not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !canEditCharacterID(c, entProf.CharacterID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	db.Client.CharacterProficiency.DeleteOneID(pid).Exec(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Import from JSON body (list)

func ImportJSON(c *gin.Context) {
	userID, _ := c.Get("user_id")
	contentType := c.GetHeader("Content-Type")

	if strings.HasPrefix(contentType, "multipart/form-data") {
		file, _, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
			return
		}
		defer file.Close()

		var chars []models.ImportCharacter
		if err := json.NewDecoder(file).Decode(&chars); err != nil {
			var single models.ImportCharacter
			file.Seek(0, 0)
			if err2 := json.NewDecoder(file).Decode(&single); err2 != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON: " + err.Error()})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
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

func UpdateExhaustion(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if !canEditCharacterID(c, charID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	var req struct {
		Level int `json:"level"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Level < 0 || req.Level > 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "exhaustion level must be 0-6"})
		return
	}
	_, err := db.Client.Character.UpdateOneID(charID).SetExhaustionLevel(req.Level).Save(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	SendCharacterUpdate(charID)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ─── Spell Preparation ───

func BatchPrepareSpells(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if !canEditCharacterID(c, charID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	var req struct {
		SpellIDs []int64 `json:"spell_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := c.Request.Context()
	// First, unprepare all spells for this character
	_, err := db.Client.Spell.Update().Where(spell.CharacterID(charID)).SetPrepared(false).Save(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Then set prepared for the specified spell IDs
	if len(req.SpellIDs) > 0 {
		_, err = db.Client.Spell.Update().Where(spell.IDIn(req.SpellIDs...), spell.CharacterID(charID)).SetPrepared(true).Save(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	SendCharacterUpdate(charID)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
