package handlers

import (
	"fmt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"strings"
	"villum/models"
)

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
