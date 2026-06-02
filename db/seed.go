package db

import (
	"log"
)

func Seed() {
	dataDir := "data"

	// Try JSON first for each category; fall back to Go structs
	type seedTask struct {
		name string
		json func(string) bool
		goFn func()
	}
	tasks := []seedTask{
		{"races", seedJSONCategory("races"), seedRaces},
		{"classes", seedJSONCategory("classes"), seedClasses},
		{"backgrounds", seedJSONCategory("backgrounds"), seedBackgrounds},
		{"spells", seedJSONCategory("spells"), seedSpells},
		{"equipment", seedJSONCategory("equipment"), seedEquipment},
		{"monsters", seedJSONCategory("monsters"), seedMonsters},
	}

	anyLoaded := false
	for _, t := range tasks {
		if t.json(dataDir) {
			anyLoaded = true
		} else {
			t.goFn()
		}
	}
	if anyLoaded {
		log.Println("Seed data loaded (some from JSON)")
	} else {
		log.Println("Seed data applied from built-in data")
	}
}

// seedJSONCategory returns a function that seeds a single category from JSON if the file exists.
func seedJSONCategory(category string) func(string) bool {
	return func(dataDir string) bool {
		return SeedJSONCategory(dataDir, category)
	}
}

func seedRaces() {
	var count int
	DB.QueryRow("SELECT COUNT(*) FROM compendium_races").Scan(&count)
	if count > 0 {
		return
	}

	races := []struct {
		name, desc, size         string
		speed                    int
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
		DB.Exec("INSERT INTO compendium_races(name,description,speed,size,ability_bonuses,traits,languages,system,source) VALUES(?,?,?,?,?,?,?,?,?)",
			r.name, r.desc, r.speed, r.size, r.abilities, r.traits, r.langs, "dnd5e", "srd")
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
		DB.Exec("INSERT INTO compendium_classes(name,description,hit_die,primary_ability,saving_throws,proficiencies,spellcasting_ability,system,source) VALUES(?,?,?,?,?,?,?,?,?)",
			c.name, c.desc, c.hitDie, c.primary, c.saves, c.profs, c.spellcasting, "dnd5e", "srd")
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
		DB.Exec("INSERT INTO compendium_backgrounds(name,description,feature_name,feature_description,proficiencies,system,source) VALUES(?,?,?,?,?,?,?)",
			b.name, b.desc, b.featName, b.featDesc, b.profs, "dnd5e", "srd")
	}
	log.Printf("Seeded %d backgrounds", len(bgs))
}

func seedSpells() {
	var count int
	DB.QueryRow("SELECT COUNT(*) FROM compendium_spells").Scan(&count)
	if count > 0 {
		return
	}

	for _, s := range SRDSpells {
		DB.Exec("INSERT INTO compendium_spells(name,level,school,casting_time,range,components,duration,description,higher_levels,classes,system,source) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)",
			s.name, s.level, s.school, s.time, s.rng, s.comp, s.dur, s.desc, s.higher, s.classes, "dnd5e", "srd")
	}
	log.Printf("Seeded %d spells", len(SRDSpells))
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
		DB.Exec("INSERT INTO compendium_equipment(name,category,cost,weight,description,system,source) VALUES(?,?,?,?,?,?,?)",
			e.name, e.cat, e.cost, e.weight, e.desc, "dnd5e", "srd")
	}
	log.Printf("Seeded %d equipment items", len(equipment))
}
