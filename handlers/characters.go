package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"vellum/db"
	"vellum/models"
)

func ListCharacters(c *gin.Context) {
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	var rows *sql.Rows
	var err error
	if role == "admin" {
		query := c.DefaultQuery("q", "")
		if query != "" {
			rows, err = db.DB.Query(`
				SELECT c.id, c.user_id, c.name, c.race, c.class, c.level, c.hp_max, c.hp_current
				FROM characters c JOIN characters_fts fts ON c.id = fts.rowid
				WHERE characters_fts MATCH ? ORDER BY c.updated_at DESC`, query)
		} else {
			rows, err = db.DB.Query(`
				SELECT c.id, c.user_id, c.name, c.race, c.class, c.level, c.hp_max, c.hp_current
				FROM characters c ORDER BY c.updated_at DESC`)
		}
	} else {
		rows, err = db.DB.Query(`
			SELECT c.id, c.user_id, c.name, c.race, c.class, c.level, c.hp_max, c.hp_current
			FROM characters c WHERE c.user_id=? ORDER BY c.updated_at DESC`, userID)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type CharSummary struct {
		ID       int64  `json:"id"`
		UserID   int64  `json:"user_id"`
		Name     string `json:"name"`
		Race     string `json:"race"`
		Class    string `json:"class"`
		Level    int    `json:"level"`
		HPMax    int    `json:"hp_max"`
		HPCurrent int   `json:"hp_current"`
	}
	chars := []CharSummary{}
	for rows.Next() {
		var c CharSummary
		rows.Scan(&c.ID, &c.UserID, &c.Name, &c.Race, &c.Class, &c.Level, &c.HPMax, &c.HPCurrent)
		chars = append(chars, c)
	}
	c.JSON(http.StatusOK, chars)
}

func GetCharacter(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	ch := &models.Character{}
	err := db.DB.QueryRow(`
		SELECT id, user_id, campaign_id, name, race, class, subclass, level, xp, background, alignment,
			str, dex, con, int, wis, cha, ac, initiative, speed,
			hp_max, hp_current, temp_hp, hit_dice, hit_dice_current,
			proficiency_bonus, inspiration, passive_perception,
			death_saves_successes, death_saves_failures, concentrating_on,
			personality_traits, ideals, bonds, flaws, appearance, backstory,
			created_at, updated_at
		FROM characters WHERE id=?`, id).Scan(
		&ch.ID, &ch.UserID, &ch.CampaignID, &ch.Name, &ch.Race, &ch.Class, &ch.Subclass, &ch.Level, &ch.XP,
		&ch.Background, &ch.Alignment,
		&ch.Str, &ch.Dex, &ch.Con, &ch.Int, &ch.Wis, &ch.Cha,
		&ch.AC, &ch.Initiative, &ch.Speed,
		&ch.HPMax, &ch.HPCurrent, &ch.TempHP, &ch.HitDice, &ch.HitDiceCurrent,
		&ch.ProficiencyBonus, &ch.Inspiration, &ch.PassivePerception,
		&ch.DeathSavesSuccesses, &ch.DeathSavesFailures, &ch.ConcentratingOn,
		&ch.PersonalityTraits, &ch.Ideals, &ch.Bonds, &ch.Flaws, &ch.Appearance, &ch.Backstory,
		&ch.CreatedAt, &ch.UpdatedAt)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Authorization
	if role != "admin" && ch.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Load sub-resources
	ch.Proficiencies = loadProficiencies(ch.ID)
	ch.Features = loadFeatures(ch.ID)
	ch.Spellcasting = loadSpellcasting(ch.ID)
	ch.Spells = loadSpells(ch.ID)
	ch.Inventory = loadInventory(ch.ID)
	ch.Currency = loadCurrency(ch.ID)

	c.JSON(http.StatusOK, ch)
}

func CreateCharacter(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var ch models.Character
	if err := c.ShouldBindJSON(&ch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set defaults
	if ch.Name == "" {
		ch.Name = "Unnamed Character"
	}
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

	campaignID := ch.CampaignID
	result, err := db.DB.Exec(`
		INSERT INTO characters(
			user_id, campaign_id, name, race, class, subclass, level, xp, background, alignment,
			str, dex, con, int, wis, cha, ac, initiative, speed,
			hp_max, hp_current, temp_hp, hit_dice, hit_dice_current,
			proficiency_bonus, inspiration, passive_perception,
			death_saves_successes, death_saves_failures, concentrating_on,
			personality_traits, ideals, bonds, flaws, appearance, backstory)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		userID, campaignID,
		ch.Name, ch.Race, ch.Class, ch.Subclass, ch.Level, ch.XP, ch.Background, ch.Alignment,
		ch.Str, ch.Dex, ch.Con, ch.Int, ch.Wis, ch.Cha,
		ch.AC, ch.Initiative, ch.Speed,
		ch.HPMax, ch.HPCurrent, ch.TempHP, ch.HitDice, ch.HitDiceCurrent,
		ch.ProficiencyBonus, ch.Inspiration, ch.PassivePerception,
		ch.DeathSavesSuccesses, ch.DeathSavesFailures, ch.ConcentratingOn,
		ch.PersonalityTraits, ch.Ideals, ch.Bonds, ch.Flaws, ch.Appearance, ch.Backstory)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, _ := result.LastInsertId()

	// Create default currency entry
	db.DB.Exec("INSERT OR IGNORE INTO character_currency(character_id) VALUES(?)", id)

	ch.ID = id
	uid, _ := userID.(int64)
	ch.UserID = uid
	c.JSON(http.StatusCreated, ch)
}

func UpdateCharacter(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	// Check ownership
	var ownerID int64
	err := db.DB.QueryRow("SELECT user_id FROM characters WHERE id=?", id).Scan(&ownerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}
	if role != "admin" && ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var ch models.Character
	if err := c.ShouldBindJSON(&ch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err = db.DB.Exec(`
		UPDATE characters SET
			name=?, race=?, class=?, subclass=?, level=?, xp=?, background=?, alignment=?,
			str=?, dex=?, con=?, int=?, wis=?, cha=?,
			ac=?, initiative=?, speed=?,
			hp_max=?, hp_current=?, temp_hp=?, hit_dice=?, hit_dice_current=?,
			proficiency_bonus=?, inspiration=?, passive_perception=?,
			death_saves_successes=?, death_saves_failures=?, concentrating_on=?,
			campaign_id=?,
			personality_traits=?, ideals=?, bonds=?, flaws=?, appearance=?, backstory=?,
			updated_at=datetime('now')
		WHERE id=?`,
		ch.Name, ch.Race, ch.Class, ch.Subclass, ch.Level, ch.XP, ch.Background, ch.Alignment,
		ch.Str, ch.Dex, ch.Con, ch.Int, ch.Wis, ch.Cha,
		ch.AC, ch.Initiative, ch.Speed,
		ch.HPMax, ch.HPCurrent, ch.TempHP, ch.HitDice, ch.HitDiceCurrent,
		ch.ProficiencyBonus, ch.Inspiration, ch.PassivePerception,
		ch.DeathSavesSuccesses, ch.DeathSavesFailures, ch.ConcentratingOn,
		ch.CampaignID,
		ch.PersonalityTraits, ch.Ideals, ch.Bonds, ch.Flaws, ch.Appearance, ch.Backstory,
		id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteCharacter(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	var ownerID int64
	err := db.DB.QueryRow("SELECT user_id FROM characters WHERE id=?", id).Scan(&ownerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}
	if role != "admin" && ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	db.DB.Exec("DELETE FROM characters WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func ExportCharacter(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	ch := &models.Character{}
	err := db.DB.QueryRow(`
		SELECT id, user_id, name, race, class, subclass, level, xp, background, alignment,
			str, dex, con, int, wis, cha, ac, initiative, speed,
			hp_max, hp_current, temp_hp, hit_dice, hit_dice_current,
			proficiency_bonus, inspiration, passive_perception,
			personality_traits, ideals, bonds, flaws, appearance, backstory
		FROM characters WHERE id=?`, id).Scan(
		&ch.ID, &ch.UserID, &ch.Name, &ch.Race, &ch.Class, &ch.Subclass,
		&ch.Level, &ch.XP, &ch.Background, &ch.Alignment,
		&ch.Str, &ch.Dex, &ch.Con, &ch.Int, &ch.Wis, &ch.Cha,
		&ch.AC, &ch.Initiative, &ch.Speed,
		&ch.HPMax, &ch.HPCurrent, &ch.TempHP, &ch.HitDice, &ch.HitDiceCurrent,
		&ch.ProficiencyBonus, &ch.Inspiration, &ch.PassivePerception,
		&ch.PersonalityTraits, &ch.Ideals, &ch.Bonds, &ch.Flaws, &ch.Appearance, &ch.Backstory)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}

	ch.Proficiencies = loadProficiencies(ch.ID)
	ch.Features = loadFeatures(ch.ID)
	ch.Spellcasting = loadSpellcasting(ch.ID)
	ch.Spells = loadSpells(ch.ID)
	ch.Inventory = loadInventory(ch.ID)
	ch.Currency = loadCurrency(ch.ID)

	format := c.DefaultQuery("format", "json")
	if format == "text" {
		c.String(http.StatusOK, characterToText(ch))
		return
	}
	c.JSON(http.StatusOK, ch)
}

func PrintCharacter(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	ch := &models.Character{}
	err := db.DB.QueryRow(`
		SELECT id, user_id, name, race, class, subclass, level, xp, background, alignment,
			str, dex, con, int, wis, cha, ac, initiative, speed,
			hp_max, hp_current, temp_hp, hit_dice, hit_dice_current,
			proficiency_bonus, inspiration, passive_perception,
			personality_traits, ideals, bonds, flaws, appearance, backstory
		FROM characters WHERE id=?`, id).Scan(
		&ch.ID, &ch.UserID, &ch.Name, &ch.Race, &ch.Class, &ch.Subclass,
		&ch.Level, &ch.XP, &ch.Background, &ch.Alignment,
		&ch.Str, &ch.Dex, &ch.Con, &ch.Int, &ch.Wis, &ch.Cha,
		&ch.AC, &ch.Initiative, &ch.Speed,
		&ch.HPMax, &ch.HPCurrent, &ch.TempHP, &ch.HitDice, &ch.HitDiceCurrent,
		&ch.ProficiencyBonus, &ch.Inspiration, &ch.PassivePerception,
		&ch.PersonalityTraits, &ch.Ideals, &ch.Bonds, &ch.Flaws, &ch.Appearance, &ch.Backstory)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}

	ch.Proficiencies = loadProficiencies(ch.ID)
	ch.Features = loadFeatures(ch.ID)
	ch.Spellcasting = loadSpellcasting(ch.ID)
	ch.Spells = loadSpells(ch.ID)
	ch.Inventory = loadInventory(ch.ID)
	ch.Currency = loadCurrency(ch.ID)

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

// checkCharacterAccess verifies the current user owns (or is admin of) the given character
func checkCharacterAccess(c *gin.Context, characterID int64) bool {
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	var ownerID int64
	err := db.DB.QueryRow("SELECT user_id FROM characters WHERE id=?", characterID).Scan(&ownerID)
	if err != nil {
		return false
	}
	return role == "admin" || ownerID == userID
}

// Internal load helpers

func loadProficiencies(characterID int64) []models.Proficiency {
	rows, err := db.DB.Query("SELECT id, character_id, type, name FROM character_proficiencies WHERE character_id=?", characterID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []models.Proficiency
	for rows.Next() {
		var p models.Proficiency
		rows.Scan(&p.ID, &p.CharacterID, &p.Type, &p.Name)
		out = append(out, p)
	}
	return out
}

func loadFeatures(characterID int64) []models.Feature {
	rows, err := db.DB.Query("SELECT id, character_id, name, description, source, level_gained FROM character_features WHERE character_id=?", characterID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []models.Feature
	for rows.Next() {
		var f models.Feature
		rows.Scan(&f.ID, &f.CharacterID, &f.Name, &f.Description, &f.Source, &f.LevelGained)
		out = append(out, f)
	}
	return out
}

func loadSpellcasting(characterID int64) *models.Spellcasting {
	sc := &models.Spellcasting{}
	err := db.DB.QueryRow(`
		SELECT character_id, ability, save_dc, attack_bonus,
			slots_1_max, slots_1_used, slots_2_max, slots_2_used,
			slots_3_max, slots_3_used, slots_4_max, slots_4_used,
			slots_5_max, slots_5_used, slots_6_max, slots_6_used,
			slots_7_max, slots_7_used, slots_8_max, slots_8_used,
			slots_9_max, slots_9_used
		FROM character_spellcasting WHERE character_id=?`, characterID).Scan(
		&sc.CharacterID, &sc.Ability, &sc.SaveDC, &sc.AttackBonus,
		&sc.Slots1Max, &sc.Slots1Used, &sc.Slots2Max, &sc.Slots2Used,
		&sc.Slots3Max, &sc.Slots3Used, &sc.Slots4Max, &sc.Slots4Used,
		&sc.Slots5Max, &sc.Slots5Used, &sc.Slots6Max, &sc.Slots6Used,
		&sc.Slots7Max, &sc.Slots7Used, &sc.Slots8Max, &sc.Slots8Used,
		&sc.Slots9Max, &sc.Slots9Used)
	if err != nil {
		return nil
	}
	return sc
}

func loadSpells(characterID int64) []models.Spell {
	rows, err := db.DB.Query(`
		SELECT id, character_id, name, level, school, casting_time, range, components, duration,
			description, prepared, always_prepared, source, notes
		FROM spells WHERE character_id=? ORDER BY level, name`, characterID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []models.Spell
	for rows.Next() {
		var s models.Spell
		rows.Scan(&s.ID, &s.CharacterID, &s.Name, &s.Level, &s.School,
			&s.CastingTime, &s.Range, &s.Components, &s.Duration,
			&s.Description, &s.Prepared, &s.AlwaysPrepared, &s.Source, &s.Notes)
		out = append(out, s)
	}
	return out
}

func loadInventory(characterID int64) []models.InventoryItem {
	rows, err := db.DB.Query(`
		SELECT id, character_id, name, quantity, weight, category,
			damage_dice, damage_type, weapon_properties,
			ac_bonus, armor_type, description,
			is_equipped, is_magical, attunement, notes
		FROM inventory WHERE character_id=? ORDER BY is_equipped DESC, name`, characterID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []models.InventoryItem
	for rows.Next() {
		var item models.InventoryItem
		rows.Scan(&item.ID, &item.CharacterID, &item.Name, &item.Quantity, &item.Weight, &item.Category,
			&item.DamageDice, &item.DamageType, &item.WeaponProperties,
			&item.ACBonus, &item.ArmorType, &item.Description,
			&item.IsEquipped, &item.IsMagical, &item.Attunement, &item.Notes)
		out = append(out, item)
	}
	return out
}

func loadCurrency(characterID int64) *models.Currency {
	cur := &models.Currency{}
	err := db.DB.QueryRow("SELECT character_id, cp, sp, ep, gp, pp FROM character_currency WHERE character_id=?", characterID).
		Scan(&cur.CharacterID, &cur.CP, &cur.SP, &cur.EP, &cur.GP, &cur.PP)
	if err != nil {
		return nil
	}
	return cur
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

	tx, err := db.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
		INSERT INTO characters(
			user_id, name, race, class, subclass, level, xp, background, alignment,
			str, dex, con, int, wis, cha, ac, initiative, speed,
			hp_max, hp_current, temp_hp, hit_dice, hit_dice_current,
			proficiency_bonus, inspiration, passive_perception,
			personality_traits, ideals, bonds, flaws, appearance, backstory)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		userID,
		imp.Name, imp.Race, imp.Class, imp.Subclass, imp.Level, imp.XP, imp.Background, imp.Alignment,
		imp.Str, imp.Dex, imp.Con, imp.Int, imp.Wis, imp.Cha,
		imp.AC, imp.Initiative, imp.Speed,
		imp.HPMax, imp.HPCurrent, imp.TempHP, imp.HitDice, 0,
		0, 0, 0,
		imp.PersonalityTraits, imp.Ideals, imp.Bonds, imp.Flaws, imp.Appearance, imp.Backstory)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	charID, _ := result.LastInsertId()

	// Currency
	tx.Exec("INSERT INTO character_currency(character_id,cp,sp,ep,gp,pp) VALUES(?,?,?,?,?,?)",
		charID, imp.Currency.CP, imp.Currency.SP, imp.Currency.EP, imp.Currency.GP, imp.Currency.PP)

	// Proficiencies
	for _, p := range imp.Proficiencies {
		tx.Exec("INSERT INTO character_proficiencies(character_id,type,name) VALUES(?,?,?)", charID, p.Type, p.Name)
	}

	// Features
	for _, f := range imp.Features {
		tx.Exec("INSERT INTO character_features(character_id,name,description,source,level_gained) VALUES(?,?,?,?,?)",
			charID, f.Name, f.Description, f.Source, f.LevelGained)
	}

	// Spellcasting
	if imp.Spellcasting != nil {
		tx.Exec(`
			INSERT INTO character_spellcasting(character_id,ability,save_dc,attack_bonus,
				slots_1_max,slots_1_used,slots_2_max,slots_2_used,
				slots_3_max,slots_3_used,slots_4_max,slots_4_used,
				slots_5_max,slots_5_used,slots_6_max,slots_6_used,
				slots_7_max,slots_7_used,slots_8_max,slots_8_used,
				slots_9_max,slots_9_used)
			VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			charID, imp.Spellcasting.Ability, imp.Spellcasting.SaveDC, imp.Spellcasting.AttackBonus,
			imp.Spellcasting.Slots1Max, imp.Spellcasting.Slots1Used,
			imp.Spellcasting.Slots2Max, imp.Spellcasting.Slots2Used,
			imp.Spellcasting.Slots3Max, imp.Spellcasting.Slots3Used,
			imp.Spellcasting.Slots4Max, imp.Spellcasting.Slots4Used,
			imp.Spellcasting.Slots5Max, imp.Spellcasting.Slots5Used,
			imp.Spellcasting.Slots6Max, imp.Spellcasting.Slots6Used,
			imp.Spellcasting.Slots7Max, imp.Spellcasting.Slots7Used,
			imp.Spellcasting.Slots8Max, imp.Spellcasting.Slots8Used,
			imp.Spellcasting.Slots9Max, imp.Spellcasting.Slots9Used)
	}

	// Spells
	for _, sp := range imp.Spells {
		tx.Exec(`
			INSERT INTO spells(character_id,name,level,school,casting_time,range,components,duration,
				description,prepared,always_prepared,source,notes)
			VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			charID, sp.Name, sp.Level, sp.School, sp.CastingTime, sp.Range, sp.Components, sp.Duration,
			sp.Description, sp.Prepared, sp.AlwaysPrepared, sp.Source, sp.Notes)
	}

	// Inventory
	for _, item := range imp.Inventory {
		tx.Exec(`
			INSERT INTO inventory(character_id,name,quantity,weight,category,
				damage_dice,damage_type,weapon_properties,
				ac_bonus,armor_type,description,
				is_equipped,is_magical,attunement,notes)
			VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			charID, item.Name, item.Quantity, item.Weight, item.Category,
			item.DamageDice, item.DamageType, item.WeaponProperties,
			item.ACBonus, item.ArmorType, item.Description,
			item.IsEquipped, item.IsMagical, item.Attunement, item.Notes)
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": charID, "name": imp.Name})
}

func UpdateCurrency(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var cur models.Currency
	if err := c.ShouldBindJSON(&cur); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Exec(`UPDATE character_currency SET cp=?,sp=?,ep=?,gp=?,pp=? WHERE character_id=?`,
		cur.CP, cur.SP, cur.EP, cur.GP, cur.PP, id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func UpdateSpellcasting(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var sc models.Spellcasting
	if err := c.ShouldBindJSON(&sc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Exec(`
		INSERT INTO character_spellcasting(character_id,ability,save_dc,attack_bonus,
			slots_1_max,slots_1_used,slots_2_max,slots_2_used,
			slots_3_max,slots_3_used,slots_4_max,slots_4_used,
			slots_5_max,slots_5_used,slots_6_max,slots_6_used,
			slots_7_max,slots_7_used,slots_8_max,slots_8_used,
			slots_9_max,slots_9_used)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(character_id) DO UPDATE SET
			ability=excluded.ability, save_dc=excluded.save_dc, attack_bonus=excluded.attack_bonus,
			slots_1_max=excluded.slots_1_max, slots_1_used=excluded.slots_1_used,
			slots_2_max=excluded.slots_2_max, slots_2_used=excluded.slots_2_used,
			slots_3_max=excluded.slots_3_max, slots_3_used=excluded.slots_3_used,
			slots_4_max=excluded.slots_4_max, slots_4_used=excluded.slots_4_used,
			slots_5_max=excluded.slots_5_max, slots_5_used=excluded.slots_5_used,
			slots_6_max=excluded.slots_6_max, slots_6_used=excluded.slots_6_used,
			slots_7_max=excluded.slots_7_max, slots_7_used=excluded.slots_7_used,
			slots_8_max=excluded.slots_8_max, slots_8_used=excluded.slots_8_used,
			slots_9_max=excluded.slots_9_max, slots_9_used=excluded.slots_9_used`,
		id, sc.Ability, sc.SaveDC, sc.AttackBonus,
		sc.Slots1Max, sc.Slots1Used, sc.Slots2Max, sc.Slots2Used,
		sc.Slots3Max, sc.Slots3Used, sc.Slots4Max, sc.Slots4Used,
		sc.Slots5Max, sc.Slots5Used, sc.Slots6Max, sc.Slots6Used,
		sc.Slots7Max, sc.Slots7Used, sc.Slots8Max, sc.Slots8Used,
		sc.Slots9Max, sc.Slots9Used)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Inventory sub-resource handlers

func CreateInventory(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var item models.InventoryItem
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := db.DB.Exec(`
		INSERT INTO inventory(character_id,name,quantity,weight,category,
			damage_dice,damage_type,weapon_properties,
			ac_bonus,armor_type,description,
			is_equipped,is_magical,attunement,notes)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		charID, item.Name, item.Quantity, item.Weight, item.Category,
		item.DamageDice, item.DamageType, item.WeaponProperties,
		item.ACBonus, item.ArmorType, item.Description,
		item.IsEquipped, item.IsMagical, item.Attunement, item.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func UpdateInventory(c *gin.Context) {
	iid, _ := strconv.ParseInt(c.Param("iid"), 10, 64)
	var item models.InventoryItem
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := db.DB.Exec(`
		UPDATE inventory SET name=?,quantity=?,weight=?,category=?,
			damage_dice=?,damage_type=?,weapon_properties=?,
			ac_bonus=?,armor_type=?,description=?,
			is_equipped=?,is_magical=?,attunement=?,notes=?
		WHERE id=?`,
		item.Name, item.Quantity, item.Weight, item.Category,
		item.DamageDice, item.DamageType, item.WeaponProperties,
		item.ACBonus, item.ArmorType, item.Description,
		item.IsEquipped, item.IsMagical, item.Attunement, item.Notes, iid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteInventory(c *gin.Context) {
	iid, _ := strconv.ParseInt(c.Param("iid"), 10, 64)
	var charID int64
	err := db.DB.QueryRow("SELECT character_id FROM inventory WHERE id=?", iid).Scan(&charID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}
	if !checkCharacterAccess(c, charID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	db.DB.Exec("DELETE FROM inventory WHERE id=?", iid)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Spell sub-resource handlers

func CreateSpell(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var sp models.Spell
	if err := c.ShouldBindJSON(&sp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := db.DB.Exec(`
		INSERT INTO spells(character_id,name,level,school,casting_time,range,components,duration,
			description,prepared,always_prepared,source,notes)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		charID, sp.Name, sp.Level, sp.School, sp.CastingTime, sp.Range, sp.Components, sp.Duration,
		sp.Description, sp.Prepared, sp.AlwaysPrepared, sp.Source, sp.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func UpdateSpell(c *gin.Context) {
	sid, _ := strconv.ParseInt(c.Param("sid"), 10, 64)
	var sp models.Spell
	if err := c.ShouldBindJSON(&sp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := db.DB.Exec(`
		UPDATE spells SET name=?,level=?,school=?,casting_time=?,range=?,components=?,duration=?,
			description=?,prepared=?,always_prepared=?,source=?,notes=?
		WHERE id=?`,
		sp.Name, sp.Level, sp.School, sp.CastingTime, sp.Range, sp.Components, sp.Duration,
		sp.Description, sp.Prepared, sp.AlwaysPrepared, sp.Source, sp.Notes, sid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteSpell(c *gin.Context) {
	sid, _ := strconv.ParseInt(c.Param("sid"), 10, 64)
	var charID int64
	err := db.DB.QueryRow("SELECT character_id FROM spells WHERE id=?", sid).Scan(&charID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "spell not found"})
		return
	}
	if !checkCharacterAccess(c, charID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	db.DB.Exec("DELETE FROM spells WHERE id=?", sid)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Feature sub-resource handlers

func CreateFeature(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var f models.Feature
	if err := c.ShouldBindJSON(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := db.DB.Exec(`
		INSERT INTO character_features(character_id,name,description,source,level_gained)
		VALUES(?,?,?,?,?)`, charID, f.Name, f.Description, f.Source, f.LevelGained)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func UpdateFeature(c *gin.Context) {
	fid, _ := strconv.ParseInt(c.Param("fid"), 10, 64)
	var f models.Feature
	if err := c.ShouldBindJSON(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := db.DB.Exec(`UPDATE character_features SET name=?,description=?,source=?,level_gained=? WHERE id=?`,
		f.Name, f.Description, f.Source, f.LevelGained, fid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteFeature(c *gin.Context) {
	fid, _ := strconv.ParseInt(c.Param("fid"), 10, 64)
	var charID int64
	err := db.DB.QueryRow("SELECT character_id FROM character_features WHERE id=?", fid).Scan(&charID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "feature not found"})
		return
	}
	if !checkCharacterAccess(c, charID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	db.DB.Exec("DELETE FROM character_features WHERE id=?", fid)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Proficiency handlers

func CreateProficiency(c *gin.Context) {
	var p models.Proficiency
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := db.DB.Exec(`INSERT INTO character_proficiencies(character_id,type,name) VALUES(?,?,?)`,
		p.CharacterID, p.Type, p.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func DeleteProficiency(c *gin.Context) {
	pid, _ := strconv.ParseInt(c.Param("pid"), 10, 64)
	var charID int64
	err := db.DB.QueryRow("SELECT character_id FROM character_proficiencies WHERE id=?", pid).Scan(&charID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "proficiency not found"})
		return
	}
	if !checkCharacterAccess(c, charID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	db.DB.Exec("DELETE FROM character_proficiencies WHERE id=?", pid)
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
		results := importCharacters(userID.(int64), chars)
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

	results := importCharacters(userID.(int64), chars)
	c.JSON(http.StatusOK, results)
}

func importCharacters(userID int64, chars []models.ImportCharacter) []gin.H {
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

		tx, err := db.DB.Begin()
		if err != nil {
			results = append(results, gin.H{"error": err.Error()})
			continue
		}

		result, err := tx.Exec(`INSERT INTO characters(user_id,name,race,class,subclass,level,xp,background,alignment,
			str,dex,con,int,wis,cha,ac,initiative,speed,hp_max,hp_current,temp_hp,hit_dice,hit_dice_current,
			proficiency_bonus,inspiration,passive_perception,personality_traits,ideals,bonds,flaws,appearance,backstory)
			VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,0,0,0,?,?,?,?,?,?)`,
			userID, imp.Name, imp.Race, imp.Class, imp.Subclass, imp.Level, imp.XP, imp.Background, imp.Alignment,
			imp.Str, imp.Dex, imp.Con, imp.Int, imp.Wis, imp.Cha,
			imp.AC, imp.Initiative, imp.Speed, imp.HPMax, imp.HPCurrent, imp.TempHP, imp.HitDice, 0,
			imp.PersonalityTraits, imp.Ideals, imp.Bonds, imp.Flaws, imp.Appearance, imp.Backstory)
		if err != nil {
			tx.Rollback()
			results = append(results, gin.H{"error": err.Error(), "name": imp.Name})
			continue
		}
		charID, _ := result.LastInsertId()
		tx.Exec("INSERT INTO character_currency(character_id) VALUES(?)", charID)

		// Proficiencies
		for _, p := range imp.Proficiencies {
			tx.Exec("INSERT INTO character_proficiencies(character_id,type,name) VALUES(?,?,?)", charID, p.Type, p.Name)
		}
		for _, f := range imp.Features {
			tx.Exec("INSERT INTO character_features(character_id,name,description,source,level_gained) VALUES(?,?,?,?,?)",
				charID, f.Name, f.Description, f.Source, f.LevelGained)
		}
		if imp.Spellcasting != nil {
			sc := imp.Spellcasting
			tx.Exec(`INSERT INTO character_spellcasting(character_id,ability,save_dc,attack_bonus,
				slots_1_max,slots_1_used,slots_2_max,slots_2_used,
				slots_3_max,slots_3_used,slots_4_max,slots_4_used,
				slots_5_max,slots_5_used,slots_6_max,slots_6_used,
				slots_7_max,slots_7_used,slots_8_max,slots_8_used,
				slots_9_max,slots_9_used) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
				charID, sc.Ability, sc.SaveDC, sc.AttackBonus,
				sc.Slots1Max, sc.Slots1Used, sc.Slots2Max, sc.Slots2Used,
				sc.Slots3Max, sc.Slots3Used, sc.Slots4Max, sc.Slots4Used,
				sc.Slots5Max, sc.Slots5Used, sc.Slots6Max, sc.Slots6Used,
				sc.Slots7Max, sc.Slots7Used, sc.Slots8Max, sc.Slots8Used,
				sc.Slots9Max, sc.Slots9Used)
		}
		for _, sp := range imp.Spells {
			tx.Exec(`INSERT INTO spells(character_id,name,level,school,casting_time,range,components,duration,
				description,prepared,always_prepared,source,notes) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`,
				charID, sp.Name, sp.Level, sp.School, sp.CastingTime, sp.Range, sp.Components, sp.Duration,
				sp.Description, sp.Prepared, sp.AlwaysPrepared, sp.Source, sp.Notes)
		}
		for _, item := range imp.Inventory {
			tx.Exec(`INSERT INTO inventory(character_id,name,quantity,weight,category,
				damage_dice,damage_type,weapon_properties,ac_bonus,armor_type,description,
				is_equipped,is_magical,attunement,notes) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
				charID, item.Name, item.Quantity, item.Weight, item.Category,
				item.DamageDice, item.DamageType, item.WeaponProperties,
				item.ACBonus, item.ArmorType, item.Description,
				item.IsEquipped, item.IsMagical, item.Attunement, item.Notes)
		}
		if err := tx.Commit(); err != nil {
			results = append(results, gin.H{"error": err.Error(), "name": imp.Name})
			continue
		}
		results = append(results, gin.H{"id": charID, "name": imp.Name, "status": "imported"})
	}
	return results
}
