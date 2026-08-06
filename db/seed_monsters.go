package db

import (
	"villum/middleware"
)

func seedMonsters() {
	var count int
	DB.QueryRow("SELECT COUNT(*) FROM compendium_monsters").Scan(&count)
	if count > 0 {
		return
	}

	type monster struct {
		name, mtype, size              string
		ac, hp                         int
		str, dex, con                  int
		int_, wis, cha                 int
		cr                             string
		source                         string
		saves, skills                  string
		vuln, resist, immun, condImmun string
		senses, langs                  string
		abilities, actions, legendary  string
		desc                           string
	}

	monsters := []monster{
		{"Goblin", "humanoid", "Small", 15, 7, 8, 14, 10, 10, 8, 8, "1/4", "MM",
			"", "Stealth +6",
			"", "", "", "",
			"darkvision 60 ft.", "Common, Goblin",
			`Nimble Escape: The goblin can take the Disengage or Hide action as a bonus action on each of its turns.`,
			`Scimitar. Melee Weapon Attack: +4 to hit, reach 5 ft., one target. Hit: 5 (1d6+2) slashing.
Shortbow. Ranged Weapon Attack: +4 to hit, range 80/320 ft., one target. Hit: 5 (1d6+2) piercing.`,
			"", "Goblins are small, black-hearted humanoids that dwell in caves, dark forests, and ruined fortresses."},

		{"Hobgoblin", "humanoid", "Medium", 18, 11, 15, 12, 12, 10, 10, 11, "1/2", "MM",
			"", "",
			"", "", "", "",
			"darkvision 60 ft.", "Common, Goblin",
			`Martial Advantage: Once per turn, the hobgoblin can deal an extra 7 (2d6) damage to a creature it hits with a weapon attack if that creature is within 5 ft. of an ally of the hobgoblin that isn't incapacitated.`,
			`Longsword. Melee Weapon Attack: +4 to hit, reach 5 ft., one target. Hit: 6 (1d8+2) slashing, or 7 (1d10+2) slashing if used with two hands.
Longbow. Ranged Weapon Attack: +3 to hit, range 150/600 ft., one target. Hit: 5 (1d8+1) piercing.`,
			"", "Hobgoblins are militaristic humanoids known for their discipline and tactical cunning."},

		{"Orc", "humanoid", "Medium", 13, 15, 16, 12, 16, 7, 11, 10, "1/2", "MM",
			"", "",
			"", "", "", "",
			"darkvision 60 ft.", "Common, Orc",
			`Aggressive: As a bonus action, the orc can move up to its speed toward a hostile creature it can see.`,
			`Greataxe. Melee Weapon Attack: +5 to hit, reach 5 ft., one target. Hit: 9 (1d12+3) slashing.
Javelin. Melee or Ranged Weapon Attack: +5 to hit, reach 5 ft. or range 30/120 ft., one target. Hit: 6 (1d6+3) piercing.`,
			"", "Orcs are savage, brutish humanoids driven by a lust for battle and the approval of their war gods."},

		{"Kobold", "humanoid", "Small", 12, 5, 7, 15, 9, 8, 7, 8, "1/8", "MM",
			"", "",
			"", "", "", "",
			"darkvision 60 ft.", "Common, Draconic",
			`Pack Tactics: The kobold has advantage on an attack roll against a creature if at least one of the kobold's allies is within 5 ft. of the creature and the ally isn't incapacitated.
Sunlight Sensitivity: While in sunlight, the kobold has disadvantage on attack rolls and Wisdom (Perception) checks that rely on sight.`,
			`Dagger. Melee Weapon Attack: +4 to hit, reach 5 ft., one target. Hit: 4 (1d4+2) piercing.
Sling. Ranged Weapon Attack: +4 to hit, range 30/120 ft., one target. Hit: 4 (1d4+2) bludgeoning.`,
			"", "Kobolds are craven, reptilian humanoids that worship dragons and serve as their minions."},

		{"Skeleton", "undead", "Medium", 13, 13, 10, 14, 15, 6, 8, 5, "1/4", "MM",
			"", "",
			"", "", "piercing, bludgeoning", "poison, exhaustion",
			"darkvision 60 ft.", "Understands all languages it knew in life but can't speak",
			"", `Shortsword. Melee Weapon Attack: +4 to hit, reach 5 ft., one target. Hit: 5 (1d6+2) piercing.
Shortbow. Ranged Weapon Attack: +4 to hit, range 80/320 ft., one target. Hit: 5 (1d6+2) piercing.`,
			"", "Skeletons are the animated bones of the dead, bound to the will of their necromantic masters."},

		{"Zombie", "undead", "Medium", 8, 22, 13, 6, 16, 3, 6, 5, "1/4", "MM",
			"", "",
			"", "", "", "poison, exhaustion",
			"darkvision 60 ft.", "Understands all languages it knew in life but can't speak",
			`Undead Fortitude: If damage reduces the zombie to 0 hit points, it must make a Constitution saving throw with a DC of 5 + the damage taken, unless the damage is radiant or from a critical hit. On a success, the zombie drops to 1 hit point instead.`,
			`Slam. Melee Weapon Attack: +3 to hit, reach 5 ft., one target. Hit: 4 (1d6+1) bludgeoning.`,
			"", "Zombies are mindless reanimated corpses driven by necromantic magic to destroy the living."},

		{"Giant Rat", "beast", "Small", 12, 7, 7, 15, 11, 2, 10, 4, "1/8", "MM",
			"", "",
			"", "", "", "",
			"darkvision 60 ft.", "",
			`Keen Smell: The giant rat has advantage on Wisdom (Perception) checks that rely on smell.
Pack Tactics: The giant rat has advantage on an attack roll against a creature if at least one of the rat's allies is within 5 ft. of the creature.`,
			`Bite. Melee Weapon Attack: +4 to hit, reach 5 ft., one target. Hit: 4 (1d4+2) piercing.`,
			"", "Giant rats are larger, more aggressive versions of common rats, often found in sewers and dungeons."},

		{"Wolf", "beast", "Medium", 13, 11, 12, 15, 12, 3, 12, 6, "1/4", "MM",
			"", "Perception +3, Stealth +4",
			"", "", "", "",
			"darkvision 60 ft.", "",
			`Keen Hearing and Smell: The wolf has advantage on Wisdom (Perception) checks that rely on hearing or smell.
Pack Tactics: The wolf has advantage on an attack roll against a creature if at least one of the wolf's allies is within 5 ft. of the creature.`,
			`Bite. Melee Weapon Attack: +4 to hit, reach 5 ft., one target. Hit: 7 (2d4+2) piercing. If the target is a creature, it must succeed on a DC 11 Strength saving throw or be knocked prone.`,
			"", "Wolves hunt in packs, using coordinated tactics to bring down larger prey."},

		{"Dire Wolf", "beast", "Large", 14, 37, 17, 15, 15, 3, 12, 7, "1", "MM",
			"Perception +3, Stealth +4", "Perception +3, Stealth +4",
			"", "", "", "",
			"darkvision 60 ft.", "",
			`Keen Hearing and Smell: The dire wolf has advantage on Wisdom (Perception) checks that rely on hearing or smell.
Pack Tactics: The dire wolf has advantage on an attack roll against a creature if at least one of the wolf's allies is within 5 ft. of the creature.`,
			`Bite. Melee Weapon Attack: +5 to hit, reach 5 ft., one target. Hit: 10 (2d6+3) piercing. If the target is a creature, it must succeed on a DC 13 Strength saving throw or be knocked prone.`,
			"", "Dire wolves are larger, more vicious relatives of common wolves, found in wild and untamed lands."},

		{"Specter", "undead", "Medium", 12, 22, 1, 14, 11, 10, 10, 11, "1", "MM",
			"", "",
			"", "necrotic, bludgeoning, piercing, slashing (from nonmagical attacks)", "", "poison, exhaustion, prone, unconscious",
			"darkvision 60 ft.", "Understands all languages it knew in life but can't speak",
			`Incorporeal Movement: The specter can move through other creatures and objects as if they were difficult terrain. It takes 5 (1d10) force damage if it ends its turn inside an object.
Sunlight Sensitivity: While in sunlight, the specter has disadvantage on attack rolls and Wisdom (Perception) checks.`,
			`Life Drain. Melee Spell Attack: +4 to hit, reach 5 ft., one creature. Hit: 10 (3d6) necrotic damage. The target must succeed on a DC 10 Constitution saving throw or its hit point maximum is reduced by an amount equal to the damage taken.`,
			"", "Specters are the restless spirits of the dead, bound to the world by unfinished business or dark magic."},

		{"Ghoul", "undead", "Medium", 12, 22, 13, 15, 10, 7, 10, 6, "1", "MM",
			"", "",
			"", "", "", "poison, exhaustion, charmed",
			"darkvision 60 ft.", "Common, Undercommon",
			"", `Bite. Melee Weapon Attack: +2 to hit, reach 5 ft., one creature. Hit: 9 (2d6+2) piercing.
Claws. Melee Weapon Attack: +4 to hit, reach 5 ft., one target. Hit: 7 (2d4+2) slashing. If the target is a creature other than an elf or undead, it must succeed on a DC 10 Constitution saving throw or be paralyzed for 1 minute.`,
			"", "Ghouls are ravenous undead that feast on the corpses of the dead and spread their curse through their bites."},

		{"Wight", "undead", "Medium", 14, 45, 15, 14, 16, 10, 13, 15, "3", "MM",
			"", "Perception +3, Stealth +4",
			"", "", "necrotic, bludgeoning, piercing, slashing (from nonmagical attacks not silvered)", "poison, exhaustion",
			"darkvision 60 ft.", "Common, Undercommon",
			`Sunlight Sensitivity: While in sunlight, the wight has disadvantage on attack rolls and Wisdom (Perception) checks.`,
			`Longsword. Melee Weapon Attack: +4 to hit, reach 5 ft., one target. Hit: 6 (1d8+2) slashing plus 6 (1d12) necrotic.
Longbow. Ranged Weapon Attack: +4 to hit, range 150/600 ft., one target. Hit: 6 (1d8+2) piercing plus 6 (1d12) necrotic.
Life Drain (Recharge 5-6). Melee Weapon Attack: +4 to hit, reach 5 ft., one creature. Hit: 8 (1d10+3) necrotic. The target must succeed on a DC 13 Constitution saving throw or its hit point maximum is reduced by the damage taken.`,
			"", "Wights are the undead remains of warriors and conquerors, driven by hate and a thirst for life."},

		{"Gelatinous Cube", "ooze", "Large", 6, 84, 14, 3, 20, 1, 6, 1, "2", "MM",
			"", "",
			"", "", "", "poison, exhaustion, prone, paralyzed, stunned, unconscious, blinded, charmed, deafened, frightened",
			"blindsight 60 ft. (blind beyond)", "",
			`Ooze Cube: The cube takes up its entire space. Other creatures can enter the space, but a creature that does so is subjected to the cube's Engulf and has disadvantage on the saving throw.
Transparent: Even when the cube is in plain sight, it takes a successful DC 15 Wisdom (Perception) check to spot it.
Engulf: The cube moves up to its speed. While doing so, it can enter Large or smaller creatures' spaces. Whenever the cube enters a creature's space, the creature must make a DC 12 Dexterity saving throw. On a failure, the creature is engulfed and takes 10 (3d6) acid damage.`,
			`Pseudopod. Melee Weapon Attack: +4 to hit, reach 5 ft., one target. Hit: 10 (3d6) acid damage.`,
			"", "Gelatinous cubes are mindless, translucent oozes that scour dungeon corridors clean of all organic matter."},

		{"Mimic", "monstrosity", "Medium", 12, 58, 17, 12, 15, 5, 13, 8, "2", "MM",
			"", "Stealth +5",
			"", "", "acid, bludgeoning, piercing, slashing (from nonmagical attacks)", "poison, prone",
			"darkvision 60 ft.", "",
			`Shapechanger: The mimic can use its action to polymorph into an object or back to its true form.
Adhesive (Object Form Only): The mimic adheres to anything that touches it. A creature adhered to the mimic is grappled (escape DC 13).
False Appearance (Object Form Only): While the mimic remains motionless, it is indistinguishable from an ordinary object.
Grappler: The mimic has advantage on attack rolls against any creature grappled by it.`,
			`Pseudopod. Melee Weapon Attack: +5 to hit, reach 5 ft., one target. Hit: 7 (1d8+3) bludgeoning plus 4 (1d8) acid.
Bite. Melee Weapon Attack: +5 to hit, reach 5 ft., one target. Hit: 8 (1d10+3) piercing plus 4 (1d8) acid.`,
			"", "Mimics are shape-shifting monsters that disguise themselves as chests, doors, and other objects to ambush unsuspecting prey."},

		{"Basilisk", "monstrosity", "Medium", 15, 52, 16, 8, 15, 2, 8, 7, "3", "MM",
			"", "Stealth +5",
			"", "", "", "",
			"darkvision 60 ft.", "",
			`Petrifying Gaze: If a creature starts its turn within 30 ft. of the basilisk and the two can see each other, the basilisk can force the creature to make a DC 12 Constitution saving throw if the basilisk isn't incapacitated. On a failed save, the creature magically begins to turn to stone and is restrained. It must repeat the saving throw at the end of its next turn. On a success, the effect ends. On a failure, the creature is petrified until freed by greater restoration or similar magic.`,
			`Bite. Melee Weapon Attack: +5 to hit, reach 5 ft., one target. Hit: 10 (2d6+3) piercing plus 7 (2d6) poison.`,
			"", "Basilisks are reptilian monsters whose gaze can turn creatures to stone, making them deadly denizens of dungeons and ruins."},

		{"Chimera", "monstrosity", "Large", 14, 114, 19, 11, 19, 3, 14, 10, "6", "MM",
			"Perception +8", "Perception +8",
			"", "", "", "",
			"darkvision 60 ft.", "Understands Draconic but can't speak",
			`Multiattack: The chimera makes three attacks: one with its bite, one with its horns, and one with its claws. When its fire breath is available, it can use the breath in place of its bite or horns.`,
			`Bite. Melee Weapon Attack: +7 to hit, reach 5 ft., one target. Hit: 11 (2d6+4) piercing plus 4 (1d8) fire.
Horns. Melee Weapon Attack: +7 to hit, reach 5 ft., one target. Hit: 10 (1d12+4) bludgeoning.
Claws. Melee Weapon Attack: +7 to hit, reach 5 ft., one target. Hit: 11 (2d6+4) slashing.
Fire Breath (Recharge 5-6). The chimera exhales fire in a 15-foot cone. Each creature in that area must make a DC 15 Dexterity saving throw, taking 31 (7d8) fire damage on a failed save, or half as much on a successful one.`,
			"", "Chimeras are monstrous hybrids of lion, goat, and dragon, known for their savage nature and fiery breath."},

		{"Young Red Dragon", "dragon", "Large", 18, 178, 23, 10, 21, 14, 11, 19, "10", "MM",
			"STR +11, DEX +5, CON +10, WIS +5", "Perception +10, Stealth +5",
			"", "", "fire", "",
			"blindsight 30 ft., darkvision 120 ft.", "Common, Draconic",
			`Multiattack: The dragon makes three attacks: one with its bite and two with its claws.`,
			`Bite. Melee Weapon Attack: +10 to hit, reach 10 ft., one target. Hit: 17 (2d10+6) piercing plus 3 (1d6) fire.
Claw. Melee Weapon Attack: +10 to hit, reach 5 ft., one target. Hit: 13 (2d6+6) slashing.
Fire Breath (Recharge 5-6). The dragon exhales fire in a 30-foot cone. Each creature in that area must make a DC 17 Dexterity saving throw, taking 56 (16d6) fire damage on a failed save, or half as much on a successful one.`,
			"", "Young red dragons are fiercely aggressive, covetous, and destructive, hoarding treasure in volcanic lairs."},

		{"Adult Red Dragon", "dragon", "Huge", 19, 256, 27, 10, 25, 16, 13, 21, "17", "MM",
			"STR +14, DEX +6, CON +13, INT +9, WIS +7, CHA +11", "Perception +13, Stealth +6",
			"", "", "fire", "",
			"blindsight 60 ft., darkvision 120 ft.", "Common, Draconic",
			`Legendary Resistance (3/Day): If the dragon fails a saving throw, it can choose to succeed instead.
Multiattack: The dragon can use its Frightful Presence. It then makes three attacks: one with its bite and two with its claws.`,
			`Bite. Melee Weapon Attack: +14 to hit, reach 10 ft., one target. Hit: 19 (2d10+8) piercing plus 7 (2d6) fire.
Claw. Melee Weapon Attack: +14 to hit, reach 5 ft., one target. Hit: 15 (2d6+8) slashing.
Tail. Melee Weapon Attack: +14 to hit, reach 15 ft., one target. Hit: 17 (2d8+8) bludgeoning.
Frightful Presence: Each creature of the dragon's choice within 120 ft. must succeed on a DC 19 Wisdom saving throw or become frightened for 1 minute.
Fire Breath (Recharge 5-6). The dragon exhales fire in a 60-foot cone. DC 21 Dexterity save, 63 (18d6) fire damage.
Legendary Actions: 3 actions: Detect, Tail Attack, Wing Attack (costs 2).`,
			"", "Adult red dragons are among the most fearsome creatures, ruling vast territories from volcanic lairs filled with treasure."},

		{"Beholder", "aberration", "Large", 18, 180, 10, 14, 18, 17, 15, 17, "13", "MM",
			"INT +9, WIS +8, CHA +9", "Perception +14, Stealth +8",
			"", "", "", "prone, blinded, charmed, deafened, frightened, poisoned, stunned",
			"darkvision 120 ft.", "Deep Speech, Undercommon",
			`Antimagic Cone: The beholder's central eye creates an area of antimagic, as in the antimagic field spell, in a 150-foot cone. At the start of each of its turns, the beholder decides which way the cone faces.
Legendary Resistance (3/Day): If the beholder fails a saving throw, it can choose to succeed instead.
Multiattack: The beholder makes three eye ray attacks.`,
			`Eye Rays: The beholder shoots three of the following magical eye rays at random (reroll duplicates), choosing one to three targets it can see within 120 ft.:
1. Charm Ray: DC 17 WIS save or charmed 1 hour.
2. Paralyzing Ray: DC 17 CON save or paralyzed 1 minute.
3. Fear Ray: DC 17 WIS save or frightened 1 minute.
4. Slowing Ray: DC 17 DEX save or speed halved 1 minute.
5. Enervation Ray: DC 17 CON save or 36 (8d8) necrotic.
6. Telekinetic Ray: DC 17 STR save or moved 30 ft.
7. Sleep Ray: DC 17 WIS save or unconscious 1 minute.
8. Petrification Ray: DC 17 DEX save or restrained, then petrified.
9. Disintegration Ray: DC 17 DEX save or 45 (10d8) force, possibly disintegrating.
10. Death Ray: DC 17 DEX save or 55 (10d10) necrotic.
Legendary Actions: 3 actions, eye rays.`,
			"", "Beholders are tentacled, spherical aberrations driven by paranoia and megalomania, each believing itself to be the perfect specimen of its kind."},

		{"Mind Flayer", "aberration", "Medium", 15, 71, 11, 12, 12, 19, 17, 17, "7", "MM",
			"INT +7, WIS +6, CHA +6", "Arcana +7, Deception +6, Insight +6, Perception +6, Persuasion +6, Stealth +4",
			"", "", "", "",
			"darkvision 120 ft.", "Deep Speech, Undercommon, telepathy 120 ft.",
			`Magic Resistance: The mind flayer has advantage on saving throws against spells and other magical effects.
Innate Spellcasting (Psionics): At will: detect thoughts, levitate. 1/day each: dominate monster, plane shift (self only).`,
			`Tentacles. Melee Weapon Attack: +7 to hit, reach 5 ft., one creature. Hit: 15 (2d10+4) psychic damage. If the target is Medium or smaller, it is grappled (escape DC 15) and must succeed on a DC 15 Intelligence saving throw or be stunned until this grapple ends.
Extract Brain: Melee Weapon Attack: +7 to hit, reach 5 ft., one incapacitated humanoid grappled by the mind flayer. Hit: The target takes 55 (10d10) piercing damage. If this damage reduces the target to 0 hit points, the mind flayer kills the target by extracting and devouring its brain.
Mind Blast (Recharge 5-6): The mind flayer magically emits psychic energy in a 60-foot cone. Each creature in that area must succeed on a DC 15 Intelligence saving throw or take 22 (4d8+4) psychic damage and be stunned for 1 minute.`,
			"", "Mind flayers, also called illithids, are horrifying psychic predators that feed on the brains of intelligent creatures and keep slaves under their mental control."},
	}

	for _, m := range monsters {
		isFull := 1
		DB.Exec(`INSERT INTO compendium_monsters(name,type,size,ac,hp,str,dex,con,int_,wis,cha,cr,source,is_full,
			saves,skills,damage_vulnerabilities,damage_resistances,damage_immunities,condition_immunities,senses,languages,
			special_abilities,actions,legendary_actions,description,alignment,expansion,publisher) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,
			?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			m.name, m.mtype, m.size, m.ac, m.hp, m.str, m.dex, m.con, m.int_, m.wis, m.cha,
			m.cr, m.source, isFull,
			m.saves, m.skills, m.vuln, m.resist, m.immun, m.condImmun,
			m.senses, m.langs, m.abilities, m.actions, m.legendary, m.desc, "", "", "")
	}
	middleware.LogInfo("seed", "seeded compendium monsters", "count", len(monsters))
}
