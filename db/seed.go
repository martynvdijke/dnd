package db

import (
	"log"
)

func Seed() {
	seedRaces()
	seedClasses()
	seedBackgrounds()
	seedSpells()
	seedEquipment()
	log.Println("Seed data applied")
}

func seedRaces() {
	var count int
	DB.QueryRow("SELECT COUNT(*) FROM compendium_races").Scan(&count)
	if count > 0 {
		return
	}

	races := []struct {
		name, desc, size string
		speed            int
		abilities, traits, langs string
	}{
		{"Dwarf", "Bold and hardy, dwarves are known for their skill in warfare, their ability to withstand physical and magical punishment, and their love of good ale.", "Medium", 25, `{"con":2}`, `{"darkvision":60,"dwarven_resilience":"Advantage on saving throws against poison","stonecunning":"Double proficiency bonus on History checks related to stonework","tool_proficiency":"Proficiency with smith's tools, brewer's supplies, or mason's tools"}`,
			"Common, Dwarvish"},
		{"Elf", "Elves are a magical people of otherworldly grace, living in the world but not entirely part of it.", "Medium", 30, `{"dex":2}`, `{"darkvision":60,"keen_senses":"Proficiency in Perception","fey_ancestry":"Advantage on saving throws against being charmed, magic can't put you to sleep","trance":"Meditate 4 hours instead of sleep"}`,
			"Common, Elvish"},
		{"Halfling", "The diminutive halflings survive in a world full of larger creatures by avoiding notice or, barring that, avoiding offense.", "Small", 25, `{"dex":2}`, `{"lucky":"When you roll a 1 on an attack roll, ability check, or saving throw, you can reroll","brave":"Advantage on saving throws against being frightened","halfling_nimbleness":"You can move through the space of any creature larger than you"}`,
			"Common, Halfling"},
		{"Human", "Humans are the most adaptable and ambitious people among the common races.", "Medium", 30, `{"str":1,"dex":1,"con":1,"int":1,"wis":1,"cha":1}`, `{}`, "Common, one additional language"},
		{"Dragonborn", "Dragonborn look very much like dragons standing erect in humanoid form.", "Medium", 30, `{"str":2,"cha":1}`, `{"draconic_ancestry":"Choose a dragon type","breath_weapon":"5'x30' line or 15' cone, damage type per ancestry, CON save","damage_resistance":"Resistance to damage type of your ancestry"}`,
			"Common, Draconic"},
		{"Gnome", "A gnome's energy and enthusiasm for all things makes them natural inventors, tinkerers, and alchemists.", "Small", 25, `{"int":2}`, `{"darkvision":60,"gnome_cunning":"Advantage on INT/WIS/CHA saving throws against magic"}`,
			"Common, Gnomish"},
		{"Half-Elf", "Half-elves combine what some say are the best qualities of their elf and human parents.", "Medium", 30, `{"cha":2}`, `{"darkvision":60,"fey_ancestry":"Advantage on saving throws against being charmed, magic can't put you to sleep","skill_versatility":"Choose two additional skill proficiencies"}`,
			"Common, Elvish, one additional language"},
		{"Half-Orc", "Half-orcs combine the ferocity of orcs with the determination of humans.", "Medium", 30, `{"str":2,"con":1}`, `{"darkvision":60,"savage_attacks":"When you score a critical hit with a melee weapon, roll one additional damage die","relentless_endurance":"When reduced to 0 HP but not killed, you drop to 1 HP once per long rest"}`,
			"Common, Orc"},
		{"Tiefling", "Tieflings are derived from human bloodlines, and in the broadest sense, they still look human.", "Medium", 30, `{"cha":2,"int":1}`, `{"darkvision":60,"hellish_resistance":"Resistance to fire damage","infernal_legacy":"You know the thaumaturgy cantrip"}`,
			"Common, Infernal"},
	}

	for _, r := range races {
		DB.Exec("INSERT INTO compendium_races(name,description,speed,size,ability_bonuses,traits,languages) VALUES(?,?,?,?,?,?,?)",
			r.name, r.desc, r.speed, r.size, r.abilities, r.traits, r.langs)
	}
	log.Printf("Seeded %d races", len(races))
}

func seedClasses() {
	var count int
	DB.QueryRow("SELECT COUNT(*) FROM compendium_classes").Scan(&count)
	if count > 0 {
		return
	}

	classes := []struct {
		name, desc, primary, saves, profs, spellcasting string
		hitDie                                          int
	}{
		{"Barbarian", "A fierce warrior of primitive background who can enter a battle rage.", "Strength", `["str","con"]`, `{"armor":["light","medium","shields"],"weapons":["simple","martial"],"saves":["str","con"]}`, "", 12},
		{"Bard", "An inspiring magician whose power echoes the music of creation.", "Charisma", `["dex","cha"]`, `{"armor":["light"],"weapons":["simple","hand crossbows","longswords","rapiers","shortswords"],"skills":["any 3"],"saves":["dex","cha"]}`, "cha", 8},
		{"Cleric", "A priestly champion who wields divine magic in service of a higher power.", "Wisdom", `["wis","cha"]`, `{"armor":["light","medium","shields"],"weapons":["simple"],"saves":["wis","cha"]}`, "wis", 8},
		{"Druid", "A priest of the Old Faith, wielding the powers of nature and adopting animal forms.", "Wisdom", `["int","wis"]`, `{"armor":["light","medium","shields (non-metal)"],"weapons":["clubs","daggers","darts","javelins","maces","quarterstaffs","scimitars","sickles","slings","spears"],"saves":["int","wis"]}`, "wis", 8},
		{"Fighter", "A master of martial combat, skilled with a variety of weapons and armor.", "Strength or Dexterity", `["str","con"]`, `{"armor":["all","shields"],"weapons":["simple","martial"],"saves":["str","con"]}`, "", 10},
		{"Monk", "A master of martial arts, harnessing the power of the body in pursuit of physical and spiritual perfection.", "Dexterity & Wisdom", `["str","dex"]`, `{"armor":["none"],"weapons":["simple","shortswords"],"saves":["str","dex"]}`, "", 8},
		{"Paladin", "A holy warrior bound to a sacred oath.", "Strength & Charisma", `["wis","cha"]`, `{"armor":["all","shields"],"weapons":["simple","martial"],"saves":["wis","cha"]}`, "cha", 10},
		{"Ranger", "A warrior who uses martial prowess and nature magic to protect the wilds.", "Dexterity & Wisdom", `["str","dex"]`, `{"armor":["light","medium","shields"],"weapons":["simple","martial"],"saves":["str","dex"]}`, "wis", 10},
		{"Rogue", "A scoundrel who uses stealth and trickery to overcome obstacles.", "Dexterity", `["dex","int"]`, `{"armor":["light"],"weapons":["simple","hand crossbows","longswords","rapiers","shortswords"],"skills":["any 4"],"saves":["dex","int"]}`, "", 8},
		{"Sorcerer", "A spellcaster who draws on inherent magic from a gift or bloodline.", "Charisma", `["con","cha"]`, `{"weapons":["daggers","darts","slings","quarterstaffs","light crossbows"],"saves":["con","cha"]}`, "cha", 6},
		{"Warlock", "A wielder of magic derived from a pact with an extraplanar entity.", "Charisma", `["wis","cha"]`, `{"armor":["light"],"weapons":["simple"],"saves":["wis","cha"]}`, "cha", 8},
		{"Wizard", "A scholarly magic-user capable of manipulating the structures of reality.", "Intelligence", `["int","wis"]`, `{"weapons":["daggers","darts","slings","quarterstaffs","light crossbows"],"saves":["int","wis"]}`, "int", 6},
	}

	for _, c := range classes {
		DB.Exec("INSERT INTO compendium_classes(name,description,hit_die,primary_ability,saving_throws,proficiencies,spellcasting_ability) VALUES(?,?,?,?,?,?,?)",
			c.name, c.desc, c.hitDie, c.primary, c.saves, c.profs, c.spellcasting)
	}
	log.Printf("Seeded %d classes", len(classes))
}

func seedBackgrounds() {
	var count int
	DB.QueryRow("SELECT COUNT(*) FROM compendium_backgrounds").Scan(&count)
	if count > 0 {
		return
	}

	bgs := []struct {
		name, desc, featName, featDesc, profs string
	}{
		{"Acolyte", "You have spent your life in the service of a temple.", "Shelter of the Faithful", "You and your adventuring companions can receive free healing and care at temples of your faith.",
			`{"skills":["Insight","Religion"],"languages":["two of your choice"]}`},
		{"Charlatan", "You have an uncanny knack for getting what you want.", "False Identity", "You have created a second identity with documentation and acquaintances.",
			`{"skills":["Deception","Sleight of Hand"],"tools":["Disguise kit","Forgery kit"]}`},
		{"Criminal", "You are an experienced criminal with a network of contacts.", "Criminal Contact", "You have a reliable contact who acts as your liaison to the criminal underworld.",
			`{"skills":["Deception","Stealth"],"tools":["One gaming set","Thieves' tools"]}`},
		{"Entertainer", "You thrive in front of an audience.", "By Popular Demand", "You can always find a place to perform and receive free lodging and food.",
			`{"skills":["Acrobatics","Performance"],"tools":["Disguise kit","One musical instrument"],"instrument":["one musical instrument"]}`},
		{"Folk Hero", "You understand the common folk and their struggles.", "Rustic Hospitality", "Common folk will hide and shelter you from authorities.", `{"skills":["Animal Handling","Survival"],"tools":["One artisan's tools","Vehicles (land)"]}`},
		{"Guild Artisan", "You are a member of a trade guild.", "Guild Membership", "You can request lodging and food at guild halls and get access to guild resources.",
			`{"skills":["Insight","Persuasion"],"languages":["one of your choice"],"tools":["One artisan's tools"]}`},
		{"Hermit", "You lived in seclusion for a formative part of your life.", "Discovery", "Your seclusion gave you a unique discovery that is key to your adventuring career.",
			`{"skills":["Medicine","Religion"],"tools":["Herbalism kit"],"languages":["one of your choice"]}`},
		{"Noble", "You were born into wealth and privilege.", "Position of Privilege", "You are welcome in high society and people will accommodate you.",
			`{"skills":["History","Persuasion"],"tools":["One gaming set"],"languages":["one of your choice"]}`},
		{"Outlander", "You grew up in the wilds, far from civilization.", "Wanderer", "You have an excellent memory for maps and geography, and can find food and water for yourself and up to five others.",
			`{"skills":["Athletics","Survival"],"tools":["One musical instrument"],"languages":["one of your choice"]}`},
		{"Sage", "You spent years learning the lore of the multiverse.", "Researcher", "You can usually recall or find information on any topic if you have access to a library.",
			`{"skills":["Arcana","History"],"languages":["two of your choice"]}`},
		{"Sailor", "You spent time at sea on a ship.", "Ship's Passage", "You can arrange free passage for yourself and companions on a sailing ship.",
			`{"skills":["Athletics","Perception"],"tools":["Navigator's tools","Vehicles (water)"]}`},
		{"Soldier", "You served in an army or militia.", "Military Rank", "You can get access to military compounds and authority among soldiers.",
			`{"skills":["Athletics","Intimidation"],"tools":["One gaming set","Vehicles (land)"]}`},
		{"Urchin", "You grew up on the streets of a major city.", "City Secrets", "You can move through cities twice as fast and know secret passages.",
			`{"skills":["Sleight of Hand","Stealth"],"tools":["Disguise kit","Thieves' tools"]}`},
	}

	for _, b := range bgs {
		DB.Exec("INSERT INTO compendium_backgrounds(name,description,feature_name,feature_description,proficiencies) VALUES(?,?,?,?,?)",
			b.name, b.desc, b.featName, b.featDesc, b.profs)
	}
	log.Printf("Seeded %d backgrounds", len(bgs))
}

func seedSpells() {
	var count int
	DB.QueryRow("SELECT COUNT(*) FROM compendium_spells").Scan(&count)
	if count > 0 {
		return
	}

	spells := []struct {
		name, school, time, rng, comp, dur, desc, higher, classes string
		level                                                     int
	}{
		{"Acid Splash", "Conjuration", "1 action", "60 feet", "V, S", "Instantaneous", "You hurl a bubble of acid. Choose one creature within range, or two within 5 feet. Each target must succeed on a DEX save or take 1d6 acid damage.", "This spell's damage increases by 1d6 when you reach 5th level (2d6), 11th level (3d6), and 17th level (4d6).", `["Sorcerer","Wizard"]`, 0},
		{"Blade Ward", "Abjuration", "1 action", "Self", "V, S", "1 round", "You extend your hand and trace a sigil of warding. Until the end of your next turn, you have resistance against bludgeoning, piercing, and slashing damage from weapon attacks.", "", `["Bard","Sorcerer","Warlock","Wizard"]`, 0},
		{"Chill Touch", "Necromancy", "1 action", "120 feet", "V, S", "1 round", "You create a ghostly hand that makes a ranged spell attack against a creature. On hit, take 1d8 necrotic damage and can't regain HP until your next turn. Undead have disadvantage on attack rolls against you.", "Damage increases by 1d8 at 5th/11th/17th level.", `["Sorcerer","Warlock","Wizard"]`, 0},
		{"Dancing Lights", "Evocation", "1 action", "120 feet", "V, S, M", "Concentration, up to 1 minute", "You create up to four torch-sized lights that shed dim light in 10 feet.", "", `["Bard","Sorcerer","Wizard"]`, 0},
		{"Fire Bolt", "Evocation", "1 action", "120 feet", "V, S", "Instantaneous", "You hurl a mote of fire. Make a ranged spell attack. On hit, target takes 1d10 fire damage. Unattended objects ignite.", "Damage increases by 1d10 at 5th/11th/17th level.", `["Sorcerer","Wizard"]`, 0},
		{"Guidance", "Divination", "1 action", "Touch", "V, S", "Concentration, up to 1 minute", "You touch a willing creature. Once before the spell ends, it can add 1d4 to one ability check of its choice.", "", `["Cleric","Druid"]`, 0},
		{"Light", "Evocation", "1 action", "Touch", "V, M", "1 hour", "You touch an object no larger than 10 feet in any dimension. Until the spell ends, the object sheds bright light in a 20-foot radius and dim light for an additional 20 feet.", "", `["Bard","Cleric","Sorcerer","Wizard"]`, 0},
		{"Mage Hand", "Conjuration", "1 action", "30 feet", "V, S", "1 minute", "A spectral hand appears that you can use to manipulate objects, open doors, etc. It has 10 ft reach, can carry 10 lbs.", "", `["Bard","Sorcerer","Warlock","Wizard"]`, 0},
		{"Minor Illusion", "Illusion", "1 action", "30 feet", "S, M", "1 minute", "You create a sound or an image of an object no larger than a 5-foot cube.", "", `["Bard","Sorcerer","Warlock","Wizard"]`, 0},
		{"Poison Spray", "Conjuration", "1 action", "10 feet", "V, S", "Instantaneous", "You extend your hand toward a creature and spray poisonous gas. The target must succeed on a CON save or take 1d12 poison damage.", "Damage increases by 1d12 at 5th/11th/17th level.", `["Druid","Sorcerer","Warlock","Wizard"]`, 0},
		{"Prestidigitation", "Transmutation", "1 action", "10 feet", "V, S", "Up to 1 hour", "Minor magical trick. You can create harmless sensory effects, clean/soil items, light/snuff candles, etc.", "", `["Bard","Sorcerer","Warlock","Wizard"]`, 0},
		{"Produce Flame", "Conjuration", "1 action", "Self", "V, S", "10 minutes", "A flickering flame appears in your hand. You can attack with it (ranged spell attack, 30 ft, 1d8 fire) or use it to shed light.", "Damage increases at 5th/11th/17th level.", `["Druid"]`, 0},
		{"Ray of Frost", "Evocation", "1 action", "60 feet", "V, S", "Instantaneous", "A frigid beam strikes a creature. Make a ranged spell attack. On hit, take 1d8 cold damage and speed reduced by 10 ft until next turn.", "Damage increases by 1d8 at 5th/11th/17th level.", `["Sorcerer","Wizard"]`, 0},
		{"Sacred Flame", "Evocation", "1 action", "60 feet", "V, S", "Instantaneous", "Flame-like radiance descends. Target must succeed on DEX save or take 1d8 radiant damage. No cover bonus.", "Damage increases by 1d8 at 5th/11th/17th level.", `["Cleric"]`, 0},
		{"Shillelagh", "Transmutation", "1 bonus action", "Touch", "V, S, M", "1 minute", "The wood of a club or quarterstaff is imbued with nature's power. Use spellcasting ability instead of STR for attacks, damage die becomes d8.", "", `["Druid"]`, 0},
		{"Shocking Grasp", "Evocation", "1 action", "Touch", "V, S", "Instantaneous", "Lightning springs from your hand. Make a melee spell attack. On hit, target takes 1d8 lightning damage and can't take reactions until its next turn. Advantage against metal armor.", "Damage increases by 1d8 at 5th/11th/17th level.", `["Sorcerer","Wizard"]`, 0},
		{"Spare the Dying", "Necromancy", "1 action", "Touch", "V, S", "Instantaneous", "You touch a living creature that has 0 HP. The creature becomes stable.", "", `["Cleric"]`, 0},
		{"Thaumaturgy", "Transmutation", "1 action", "30 feet", "V", "Up to 1 minute", "You manifest a minor wonder. You can create sounds, alter flames, cause ground tremor, etc.", "", `["Cleric"]`, 0},
		{"True Strike", "Divination", "1 action", "30 feet", "S", "Concentration, up to 1 round", "You gain a brief insight. On your next turn, you gain advantage on your first attack roll against the target.", "", `["Bard","Sorcerer","Warlock","Wizard"]`, 0},
		{"Vicious Mockery", "Enchantment", "1 action", "60 feet", "V", "Instantaneous", "You unleash a string of insults. Target must succeed on WIS save or take 1d4 psychic damage and have disadvantage on next attack roll.", "Damage increases by 1d4 at 5th/11th/17th level.", `["Bard"]`, 0},

		// Level 1 spells
		{"Bless", "Enchantment", "1 action", "30 feet", "V, S, M", "Concentration, up to 1 minute", "You bless up to three creatures. Each add 1d4 to attack rolls and saving throws.", "When cast at 2nd+ level, you can target one additional creature per level.", `["Cleric","Paladin"]`, 1},
		{"Burning Hands", "Evocation", "1 action", "Self (15-ft cone)", "V, S", "Instantaneous", "A thin sheet of flames shoots from your outstretched hands. Creatures in the cone must succeed on a DEX save or take 3d6 fire damage.", "Damage increases by 1d6 per level above 1st.", `["Sorcerer","Wizard"]`, 1},
		{"Cure Wounds", "Evocation", "1 action", "Touch", "V, S", "Instantaneous", "A creature you touch regains 1d8 + spellcasting ability modifier HP.", "Healing increases by 1d8 per level above 1st.", `["Bard","Cleric","Druid","Paladin","Ranger"]`, 1},
		{"Detect Magic", "Divination", "1 action", "Self", "V, S", "Concentration, up to 10 minutes", "You sense the presence of magic within 30 feet of you. You can see a faint aura around visible creatures and objects.", "", `["Bard","Cleric","Druid","Paladin","Ranger","Sorcerer","Warlock","Wizard"]`, 1},
		{"Fireball", "Evocation", "1 action", "150 feet", "V, S, M", "Instantaneous", "A bright streak flashes from your finger to a point you choose within range and blossoms with a low roar into an explosion of flame. Each creature in a 20-foot-radius sphere must make a DEX save. On a failed save, target takes 8d6 fire damage.", "Damage increases by 1d6 per level above 3rd.", `["Sorcerer","Wizard"]`, 3},
		{"Magic Missile", "Evocation", "1 action", "120 feet", "V, S", "Instantaneous", "You create three glowing darts. Each hits a creature you can see, dealing 1d4+1 force damage. The darts strike simultaneously.", "One additional dart per level above 1st.", `["Sorcerer","Wizard"]`, 1},
		{"Mage Armor", "Abjuration", "1 action", "Touch", "V, S, M", "8 hours", "You touch a willing creature not wearing armor. Its base AC becomes 13 + DEX modifier.", "", `["Sorcerer","Wizard"]`, 1},
		{"Shield", "Abjuration", "1 reaction", "Self", "V, S", "1 round", "An invisible barrier appears. You gain +5 AC until the start of your next turn, including against the triggering attack.", "", `["Sorcerer","Wizard"]`, 1},
		{"Healing Word", "Evocation", "1 bonus action", "60 feet", "V", "Instantaneous", "A creature regains 1d4 + spellcasting ability modifier HP.", "Healing increases by 1d4 per level above 1st.", `["Bard","Cleric","Druid"]`, 1},
	}

	for _, s := range spells {
		DB.Exec("INSERT INTO compendium_spells(name,level,school,casting_time,range,components,duration,description,higher_levels,classes) VALUES(?,?,?,?,?,?,?,?,?,?)",
			s.name, s.level, s.school, s.time, s.rng, s.comp, s.dur, s.desc, s.higher, s.classes)
	}
	log.Printf("Seeded %d spells", len(spells))
}

func seedEquipment() {
	var count int
	DB.QueryRow("SELECT COUNT(*) FROM compendium_equipment").Scan(&count)
	if count > 0 {
		return
	}

	equipment := []struct {
		name, cat, cost, desc string
		weight                float64
	}{
		{"Battleaxe", "Weapon", `{"quantity":1,"unit":"gp","amount":10}`, "1d8 slashing (versatile 1d10)", 4},
		{"Chain Mail", "Armor", `{"quantity":1,"unit":"gp","amount":75}`, "AC 16, STR 13 required, disadvantage on Stealth", 55},
		{"Leather Armor", "Armor", `{"quantity":1,"unit":"gp","amount":10}`, "AC 11 + DEX mod", 10},
		{"Plate Armor", "Armor", `{"quantity":1,"unit":"gp","amount":1500}`, "AC 18, STR 15 required, disadvantage on Stealth", 65},
		{"Shield", "Armor", `{"quantity":1,"unit":"gp","amount":10}`, "+2 AC", 6},
		{"Longsword", "Weapon", `{"quantity":1,"unit":"gp","amount":15}`, "1d8 slashing (versatile 1d10)", 3},
		{"Shortsword", "Weapon", `{"quantity":1,"unit":"gp","amount":10}`, "1d6 piercing (light, finesse)", 2},
		{"Dagger", "Weapon", `{"quantity":1,"unit":"gp","amount":2}`, "1d4 piercing (finesse, light, thrown 20/60)", 1},
		{"Greatsword", "Weapon", `{"quantity":1,"unit":"gp","amount":50}`, "2d6 slashing (heavy, two-handed)", 6},
		{"Shortbow", "Weapon", `{"quantity":1,"unit":"gp","amount":25}`, "1d6 piercing (ammunition 80/320, two-handed)", 2},
		{"Longbow", "Weapon", `{"quantity":1,"unit":"gp","amount":50}`, "1d8 piercing (ammunition 150/600, heavy, two-handed)", 2},
		{"Crossbow, Light", "Weapon", `{"quantity":1,"unit":"gp","amount":25}`, "1d8 piercing (ammunition 80/320, loading, two-handed)", 5},
		{"Quarterstaff", "Weapon", `{"quantity":1,"unit":"gp","amount":0.2}`, "1d6 bludgeoning (versatile 1d8)", 4},
		{"Spear", "Weapon", `{"quantity":1,"unit":"gp","amount":1}`, "1d6 piercing (thrown 20/60, versatile 1d8)", 3},
		{"Handaxe", "Weapon", `{"quantity":1,"unit":"gp","amount":5}`, "1d6 slashing (light, thrown 20/60)", 2},
		{"Backpack", "Gear", `{"quantity":1,"unit":"gp","amount":2}`, "Holds 30 lbs of gear. Includes bedroll, mess kit, tinderbox, torch, rations, waterskin, 50 ft rope", 5},
		{"Candle", "Gear", `{"quantity":1,"unit":"cp","amount":1}`, "Sheds dim light in 5-foot radius for 1 hour", 0},
		{"Crowbar", "Gear", `{"quantity":1,"unit":"gp","amount":2}`, "Advantage on Strength checks to pry things open", 5},
		{"Healer's Kit", "Gear", `{"quantity":1,"unit":"gp","amount":5}`, "10 uses, stabilizes a creature without a check", 3},
		{"Potion of Healing", "Consumable", `{"quantity":1,"unit":"gp","amount":50}`, "Regain 2d4+2 HP", 0.5},
		{"Rations (1 day)", "Gear", `{"quantity":1,"unit":"sp","amount":5}`, "Nutritious field rations for one day", 2},
		{"Rope, Hempen (50 ft)", "Gear", `{"quantity":1,"unit":"gp","amount":1}`, "50 feet of hempen rope", 10},
		{"Torch", "Gear", `{"quantity":1,"unit":"cp","amount":1}`, "Sheds bright light in 20 feet, dim light 20 more. Burns for 1 hour", 1},
		{"Waterskin", "Gear", `{"quantity":1,"unit":"sp","amount":2}`, "Holds 4 pints of water", 5},
		{"Spellbook", "Gear", `{"quantity":1,"unit":"gp","amount":50}`, "100 pages of parchment for recording spells", 3},
	}

	for _, e := range equipment {
		DB.Exec("INSERT INTO compendium_equipment(name,category,cost,weight,description) VALUES(?,?,?,?,?)",
			e.name, e.cat, e.cost, e.weight, e.desc)
	}
	log.Printf("Seeded %d equipment items", len(equipment))
}
