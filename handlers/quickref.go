package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type QuickRefSection struct {
	Title   string          `json:"title"`
	Icon    string          `json:"icon"`
	Entries []QuickRefEntry `json:"entries"`
}

type QuickRefEntry struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Source      string `json:"source,omitempty"`
}

func GetQuickReference(c *gin.Context) {
	section := c.Query("section")

	conditions := []QuickRefEntry{
		{"Blinded", "A blinded creature can't see and automatically fails ability checks that require sight. Attack rolls against it have advantage, and its attacks have disadvantage.", "PHB p.290"},
		{"Charmed", "A charmed creature can't attack the charmer or target them with harmful abilities. The charmer has advantage on social interactions with it.", "PHB p.290"},
		{"Deafened", "A deafened creature can't hear and automatically fails ability checks that require hearing.", "PHB p.290"},
		{"Exhaustion", "Exhaustion has 6 levels. Level 1: disadvantage on ability checks. Level 2: speed halved. Level 3: disadvantage on attacks and saves. Level 4: HP max halved. Level 5: speed 0. Level 6: death.", "PHB p.291"},
		{"Frightened", "A frightened creature has disadvantage on ability checks and attack rolls while the source is visible. It can't willingly move closer.", "PHB p.290"},
		{"Grappled", "A grappled creature's speed becomes 0. It ends if the grappler is incapacitated or teleports away.", "PHB p.290"},
		{"Incapacitated", "An incapacitated creature can't take actions, bonus actions, or reactions.", "PHB p.290"},
		{"Invisible", "An invisible creature is impossible to see. Attack rolls against it have disadvantage, and its attacks have advantage.", "PHB p.291"},
		{"Paralyzed", "A paralyzed creature is incapacitated, can't move or speak. Attacks that hit are crits if within 5ft. Attack rolls against it have advantage.", "PHB p.291"},
		{"Petrified", "A petrified creature is turned to stone, incapacitated, unaware of surroundings. Resistance to all damage.", "PHB p.291"},
		{"Poisoned", "A poisoned creature has disadvantage on attack rolls and ability checks.", "PHB p.292"},
		{"Prone", "A prone creature's attacks have disadvantage. Attack rolls against it have advantage if within 5ft, disadvantage if beyond.", "PHB p.292"},
		{"Restrained", "A restrained creature's speed is 0. Attacks against it have advantage, and its attacks have disadvantage.", "PHB p.292"},
		{"Stunned", "A stunned creature is incapacitated, can't move. Attacks against it have advantage. It automatically fails STR and DEX saves.", "PHB p.292"},
		{"Unconscious", "An unconscious creature is incapacitated, drops items, and is prone. Attacks that hit are crits if within 5ft.", "PHB p.292"},
	}

	combatActions := []QuickRefEntry{
		{"Attack", "Make a weapon or unarmed strike.", "PHB p.192"},
		{"Cast a Spell", "Cast a spell with a casting time of 1 action.", "PHB p.192"},
		{"Dash", "Double your speed.", "PHB p.192"},
		{"Disengage", "Your movement doesn't provoke opportunity attacks.", "PHB p.192"},
		{"Dodge", "Until your next turn, attacks against you have disadvantage.", "PHB p.192"},
		{"Help", "Give an ally advantage on an ability check or attack.", "PHB p.192"},
		{"Hide", "Make a Stealth check to hide from enemies.", "PHB p.192"},
		{"Ready", "Prepare an action to take later in the round.", "PHB p.193"},
		{"Search", "Make a Perception or Investigation check.", "PHB p.193"},
		{"Use Object", "Interact with an object or use a special ability.", "PHB p.193"},
	}

	damageTypes := []QuickRefEntry{
		{"Acid", "Green, corrosive damage. Common in dragon breath and spells.", ""},
		{"Bludgeoning", "Blunt force from hammers, falls, and explosions.", ""},
		{"Cold", "Freezing damage from ice, frost, and cold environments.", ""},
		{"Fire", "Burning damage from flames, magma, and heat.", ""},
		{"Force", "Pure magical energy. Very few creatures resist it.", ""},
		{"Lightning", "Electrical damage from storms and shock spells.", ""},
		{"Necrotic", "Deathly, decaying energy. Common to undead.", ""},
		{"Piercing", "Sharp point damage from arrows, spears, and stabs.", ""},
		{"Poison", "Toxic damage. Many creatures have resistance or immunity.", ""},
		{"Psychic", "Mental damage that bypasses physical defenses.", ""},
		{"Radiant", "Holy, searing light. Effective against undead.", ""},
		{"Slashing", "Cutting damage from swords, claws, and blades.", ""},
		{"Thunder", "Sonic damage from loud, concussive force.", ""},
	}

	skills := []QuickRefEntry{
		{"Acrobatics (DEX)", "Stay on your feet, perform flips, navigate narrow surfaces.", "PHB p.176"},
		{"Animal Handling (WIS)", "Calm or train animals, ride mounts.", "PHB p.178"},
		{"Arcana (INT)", "Recall lore about spells, magic items, and planes.", "PHB p.177"},
		{"Athletics (STR)", "Climb, swim, jump, grapple, shove.", "PHB p.175"},
		{"Deception (CHA)", "Lie convincingly, disguise intent.", "PHB p.178"},
		{"History (INT)", "Recall lore about kingdoms, wars, and past events.", "PHB p.177"},
		{"Insight (WIS)", "Read intentions, detect lies, gauge mood.", "PHB p.178"},
		{"Intimidation (CHA)", "Coerce, threaten, browbeat.", "PHB p.179"},
		{"Investigation (INT)", "Search for clues, deduce patterns, find hidden objects.", "PHB p.178"},
		{"Medicine (WIS)", "Stabilize dying, diagnose illness.", "PHB p.178"},
		{"Nature (INT)", "Recall lore about terrain, plants, animals, weather.", "PHB p.178"},
		{"Perception (WIS)", "Notice hidden creatures, hear sounds, spot details.", "PHB p.178"},
		{"Performance (CHA)", "Entertain, sing, dance, act.", "PHB p.179"},
		{"Persuasion (CHA)", "Convince, negotiate, charm.", "PHB p.179"},
		{"Religion (INT)", "Recall lore about gods, rituals, undead.", "PHB p.178"},
		{"Sleight of Hand (DEX)", "Pick pockets, palm objects, perform tricks.", "PHB p.177"},
		{"Stealth (DEX)", "Move silently, hide, avoid notice.", "PHB p.177"},
		{"Survival (WIS)", "Track, hunt, navigate, avoid natural hazards.", "PHB p.178"},
	}

	allSections := map[string]QuickRefSection{
		"conditions":   {"Conditions", "fa-face-dizzy", conditions},
		"actions":      {"Combat Actions", "fa-crosshairs", combatActions},
		"damage-types": {"Damage Types", "fa-bolt", damageTypes},
		"skills":       {"Skills", "fa-star", skills},
	}

	if section != "" {
		if sec, ok := allSections[section]; ok {
			c.JSON(http.StatusOK, sec)
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "section not found"})
		return
	}

	result := make([]QuickRefSection, 0, len(allSections))
	for _, sec := range allSections {
		result = append(result, sec)
	}
	c.JSON(http.StatusOK, result)
}
