package handlers

import (
	"fmt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"html"
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

// characterToHTML renders a self-contained, print-friendly character sheet.
// The browser is the PDF engine: the frontend opens this and calls print().
func characterToHTML(ch *models.Character) string {
	esc := html.EscapeString
	var b strings.Builder

	b.WriteString(`<!doctype html><html lang="en"><head><meta charset="utf-8">`)
	fmt.Fprintf(&b, `<title>%s</title>`, esc(ch.Name))
	b.WriteString(`<style>
body{font-family:Georgia,'Times New Roman',serif;color:#1a1a1a;margin:2rem;max-width:52rem}
h1{margin:0;font-size:1.8rem}
h2{font-size:1rem;text-transform:uppercase;letter-spacing:.08em;border-bottom:2px solid #8a6d1f;margin:1.4rem 0 .5rem;padding-bottom:.2rem}
.sub{color:#555;margin:.2rem 0 0}
.grid{display:grid;grid-template-columns:repeat(6,1fr);gap:.5rem;margin:.5rem 0}
.abil{border:1px solid #bbb;border-radius:6px;text-align:center;padding:.4rem}
.abil .v{font-size:1.3rem;font-weight:bold}
.abil .m{color:#555;font-size:.85rem}
.stats{display:flex;flex-wrap:wrap;gap:1.2rem;margin:.5rem 0}
.stats div{font-size:.95rem}
table{width:100%;border-collapse:collapse;font-size:.9rem}
td,th{border-bottom:1px solid #ddd;padding:.25rem .4rem;text-align:left}
ul{margin:.3rem 0;padding-left:1.2rem}
@media print{body{margin:0.5cm}h2{page-break-after:avoid}}
</style></head><body>`)

	fmt.Fprintf(&b, `<h1>%s</h1>`, esc(ch.Name))
	fmt.Fprintf(&b, `<p class="sub">%s %s`, esc(ch.Race), esc(ch.Class))
	if ch.Subclass != "" {
		fmt.Fprintf(&b, ` (%s)`, esc(ch.Subclass))
	}
	fmt.Fprintf(&b, ` &middot; Level %d &middot; %s %s</p>`, ch.Level, esc(ch.Background), esc(ch.Alignment))

	b.WriteString(`<h2>Ability Scores</h2><div class="grid">`)
	abils := []struct {
		label string
		score int
		mod   int
	}{
		{"STR", ch.Str, ch.StrMod}, {"DEX", ch.Dex, ch.DexMod}, {"CON", ch.Con, ch.ConMod},
		{"INT", ch.Int, ch.IntMod}, {"WIS", ch.Wis, ch.WisMod}, {"CHA", ch.Cha, ch.ChaMod},
	}
	for _, a := range abils {
		fmt.Fprintf(&b, `<div class="abil"><div>%s</div><div class="v">%d</div><div class="m">%+d</div></div>`, a.label, a.score, a.mod)
	}
	b.WriteString(`</div>`)

	b.WriteString(`<h2>Combat</h2><div class="stats">`)
	fmt.Fprintf(&b, `<div><strong>AC</strong> %d</div>`, ch.AC)
	fmt.Fprintf(&b, `<div><strong>Initiative</strong> %+d</div>`, ch.Initiative)
	fmt.Fprintf(&b, `<div><strong>Speed</strong> %d ft</div>`, ch.Speed)
	fmt.Fprintf(&b, `<div><strong>HP</strong> %d/%d</div>`, ch.HPCurrent, ch.HPMax)
	fmt.Fprintf(&b, `<div><strong>Temp HP</strong> %d</div>`, ch.TempHP)
	fmt.Fprintf(&b, `<div><strong>Hit Dice</strong> %s (%d)</div>`, esc(ch.HitDice), ch.HitDiceCurrent)
	fmt.Fprintf(&b, `<div><strong>Prof.</strong> +%d</div>`, ch.ProficiencyBonus)
	fmt.Fprintf(&b, `<div><strong>Passive Perception</strong> %d</div>`, ch.PassivePerception)
	b.WriteString(`</div>`)

	if ch.Spellcasting != nil && ch.Spellcasting.Ability != "" {
		b.WriteString(`<h2>Spellcasting</h2><div class="stats">`)
		fmt.Fprintf(&b, `<div><strong>Ability</strong> %s</div>`, esc(ch.Spellcasting.Ability))
		fmt.Fprintf(&b, `<div><strong>Save DC</strong> %d</div>`, ch.Spellcasting.SaveDC)
		fmt.Fprintf(&b, `<div><strong>Attack</strong> +%d</div>`, ch.Spellcasting.AttackBonus)
		b.WriteString(`</div>`)
	}

	if len(ch.Spells) > 0 {
		b.WriteString(`<h2>Spells</h2>`)
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
			fmt.Fprintf(&b, `<p><strong>%s:</strong> `, label)
			names := make([]string, 0, len(spells))
			for _, sp := range spells {
				n := esc(sp.Name)
				if sp.Prepared {
					n += " [P]"
				}
				names = append(names, n)
			}
			fmt.Fprintf(&b, `%s</p>`, strings.Join(names, ", "))
		}
	}

	if len(ch.Inventory) > 0 {
		b.WriteString(`<h2>Inventory</h2><table><tr><th>Item</th><th>Qty</th><th>Notes</th></tr>`)
		for _, item := range ch.Inventory {
			notes := ""
			if item.IsEquipped {
				notes = "equipped"
			}
			if item.Category == "weapon" && item.DamageDice != "" {
				notes = strings.TrimSpace(notes + " " + item.DamageDice + " " + item.DamageType)
			}
			fmt.Fprintf(&b, `<tr><td>%s</td><td>%d</td><td>%s</td></tr>`, esc(item.Name), item.Quantity, esc(notes))
		}
		b.WriteString(`</table>`)
	}

	if ch.Currency != nil {
		parts := []string{}
		for _, c := range []struct {
			label string
			val   int
		}{{"PP", ch.Currency.PP}, {"GP", ch.Currency.GP}, {"EP", ch.Currency.EP}, {"SP", ch.Currency.SP}, {"CP", ch.Currency.CP}} {
			if c.val > 0 {
				parts = append(parts, fmt.Sprintf("%d %s", c.val, c.label))
			}
		}
		if len(parts) > 0 {
			fmt.Fprintf(&b, `<h2>Currency</h2><p>%s</p>`, strings.Join(parts, ", "))
		}
	}

	if len(ch.Features) > 0 {
		b.WriteString(`<h2>Features &amp; Traits</h2><ul>`)
		for _, f := range ch.Features {
			fmt.Fprintf(&b, `<li><strong>%s</strong> (Level %d, %s)`, esc(f.Name), f.LevelGained, esc(f.Source))
			if f.Description != "" {
				fmt.Fprintf(&b, `<br>%s`, esc(f.Description))
			}
			b.WriteString(`</li>`)
		}
		b.WriteString(`</ul>`)
	}

	if len(ch.Proficiencies) > 0 {
		b.WriteString(`<h2>Proficiencies</h2>`)
		byType := make(map[string][]string)
		for _, p := range ch.Proficiencies {
			byType[p.Type] = append(byType[p.Type], p.Name)
		}
		titleCaser := cases.Title(language.English)
		for typ, names := range byType {
			fmt.Fprintf(&b, `<p><strong>%s:</strong> %s</p>`, titleCaser.String(typ), esc(strings.Join(names, ", ")))
		}
	}

	if ch.PersonalityTraits != "" || ch.Ideals != "" || ch.Bonds != "" || ch.Flaws != "" {
		b.WriteString(`<h2>Personality</h2>`)
		for _, p := range []struct{ label, val string }{
			{"Traits", ch.PersonalityTraits}, {"Ideals", ch.Ideals}, {"Bonds", ch.Bonds}, {"Flaws", ch.Flaws},
		} {
			if p.val != "" {
				fmt.Fprintf(&b, `<p><strong>%s:</strong> %s</p>`, p.label, esc(p.val))
			}
		}
	}

	if ch.Appearance != "" {
		fmt.Fprintf(&b, `<h2>Appearance</h2><p>%s</p>`, esc(ch.Appearance))
	}
	if ch.Backstory != "" {
		fmt.Fprintf(&b, `<h2>Backstory</h2><p>%s</p>`, esc(ch.Backstory))
	}

	b.WriteString(`</body></html>`)
	return b.String()
}
