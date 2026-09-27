package handlers

import (
	"villum/db"
	"villum/models"
)

// ComputeArmorClass derives a character's AC from equipped body armor, shields,
// and the Dexterity modifier. The bool is false when no body armor is equipped,
// in which case the caller keeps the stored manual AC (unarmored characters and
// class-feature formulas like Unarmored Defense are not modeled).
func ComputeArmorClass(dexMod int, items []models.InventoryItem) (int, bool) {
	base := 0
	computed := false
	shield := 0
	for _, it := range items {
		if !it.IsEquipped {
			continue
		}
		switch it.ArmorType {
		case "light", "medium", "heavy":
			b := it.ACBonus
			if b <= 0 {
				switch it.ArmorType {
				case "light":
					b = 11
				case "medium":
					b = 13
				case "heavy":
					b = 16
				}
			}
			dex := dexMod
			switch it.ArmorType {
			case "medium":
				if dex > 2 {
					dex = 2
				}
			case "heavy":
				dex = 0
			}
			if v := b + dex; !computed || v > base {
				base = v
			}
			computed = true
		default:
			// Shield or flat AC bonus item (ring of protection etc.).
			if it.ACBonus > 0 {
				shield += it.ACBonus
			}
		}
	}
	if !computed {
		return 10, false
	}
	return base + shield, true
}

// loadEquippedArmor returns the equipped items relevant to AC.
func loadEquippedArmor(charID int64) []models.InventoryItem {
	rows, err := db.DB.Query("SELECT COALESCE(armor_type,''), COALESCE(ac_bonus,0) FROM inventory WHERE character_id=? AND is_equipped=1", charID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var items []models.InventoryItem
	for rows.Next() {
		var it models.InventoryItem
		if rows.Scan(&it.ArmorType, &it.ACBonus) == nil {
			it.IsEquipped = true
			items = append(items, it)
		}
	}
	return items
}

// recomputeCharacterAC derives and persists the character's AC when body armor
// is equipped; otherwise the manual AC is left untouched. Returns the AC and
// whether it was auto-computed.
func recomputeCharacterAC(charID int64) (int, bool) {
	var dex int
	if err := db.DB.QueryRow("SELECT COALESCE(dex,10) FROM characters WHERE id=?", charID).Scan(&dex); err != nil {
		return 0, false
	}
	ac, computed := ComputeArmorClass(abilityMod(dex), loadEquippedArmor(charID))
	if computed {
		db.DB.Exec("UPDATE characters SET ac=? WHERE id=?", ac, charID)
	}
	return ac, computed
}
