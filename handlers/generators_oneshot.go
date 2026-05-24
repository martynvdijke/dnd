package handlers

import (
	"math/rand"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ─── Adventure Hook Generator ───

var hookNames = []string{
	"The Missing Relic", "The Village in Peril", "The Shadow Council",
	"The Forgotten Tunnels", "The Cursed Heirloom", "The Bandit Lord's Prize",
	"The Lost Caravan", "The Necromancer's Lair", "The Stolen Child",
	"The Ancient Seal", "The Poisoned Well", "The Royal Summons",
	"The Smuggler's Cove", "The Haunted Lighthouse", "The Dragon's Trail",
}

var hookTypes = []string{"mystery", "rescue", "retrieval", "elimination", "escort", "exploration", "political"}

var villainTypes = []string{
	"Corrupt noble", "Cult leader", "Rogue wizard", "Bandit chieftain",
	"Revenant seeking vengeance", "Ambitious merchant guild", "Obsessed scholar",
	"Fey trickster", "Dragon in human form", "Beholder's minion",
	"Undead warlord", "Hag coven", "Rival adventuring party", "Criminal syndicate", "Possessed royal advisor",
}

var macguffins = []string{
	"Ancient artifact", "Sealed scroll", "Magical gemstone", "Royal seal",
	"Map to a hidden treasure", "Rare alchemical ingredient", "Enchanted weapon",
	"Bound elemental", "Cursed idol", "Prophecy tablet",
	"Key to a planar gate", "Dragon egg", "Heart of a fallen angel", "Seed of the World Tree", "Crown of forgotten king",
}

var stakes = []string{
	"The destruction of the village", "The safety of a royal heir",
	"The balance of power in the region", "The release of a sealed evil",
	"The fate of a noble house", "The control of a vital trade route",
	"The survival of a reclusive order", "The prevention of a war",
	"The awakening of an ancient terror", "The spread of a deadly plague",
}

var locationHints = []string{
	"Deep in the Whispering Woods", "Beneath the abandoned monastery",
	"Among the ruins of an elven city", "In the catacombs below the city",
	"Across the Scorched Wastes", "On a remote island in the Misty Sea",
	"Inside the volcanic mountain", "In the heart of the enchanted forest",
	"Beyond the Wall of Bones", "In the flooded dwarven mines",
	"At the forgotten crossroads", "Within the desert temple", "Through the faerie circle",
	"Beneath the frozen lake", "Along the treacherous cliffside trail",
}

var twists = []string{
	"The real villain is someone the party trusts",
	"The macguffin is a fake, planted as a trap",
	"The quest giver is actually the villain's pawn",
	"The treasure is cursed to cause a worse problem",
	"The person to be rescued doesn't want to leave",
	"The villain's motivation is completely justified",
	"The party is not the first - others have tried and failed",
	"There's a traitor among the quest givers",
	"The location has shifted to another plane",
	"The true threat is much larger than implied",
	"The alliance between enemies is fragile and exploitable",
	"The artifact must be destroyed, not retrieved",
	"Time is shorter than they were told",
	"The monsters are being controlled, not naturally aggressive",
	"The reward is counterfeit or worthless",
}

func HandleGenerateAdventureHook(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"hook_name":     hookNames[rand.Intn(len(hookNames))],
		"hook_type":     hookTypes[rand.Intn(len(hookTypes))],
		"villain":       villainTypes[rand.Intn(len(villainTypes))],
		"macguffin":     macguffins[rand.Intn(len(macguffins))],
		"stakes":        stakes[rand.Intn(len(stakes))],
		"location_hint": locationHints[rand.Intn(len(locationHints))],
		"twist":         twists[rand.Intn(len(twists))],
	})
}

// ─── Dungeon Dressing Generator ───

var roomTypes = []string{"chamber", "hallway", "stairwell", "cavern", "tomb", "workshop", "shrine", "prison", "vault", "pool"}

var roomSizes = []string{"10x10 feet", "15x20 feet", "20x30 feet", "30x40 feet", "10x50 foot hallway", "circular 20ft diameter", "irregular cavern", "split level 15x25"}

var roomShapes = []string{"square", "rectangle", "circle", "octagon", "L-shaped", "T-shaped", "cross-shaped", "irregular", "oval", "triangular"}

var roomFloors = []string{"flagstone", "packed earth", "wooden planks", "mosaic tile", "rough hewn stone", "cobblestone", "smooth marble", "rotting wood", "bloodstained stone", "bones fused into floor"}

var roomWalls = []string{"rough stone", "smooth plaster", "ancient brick", "carved bedrock", "damp masonry", "cracked tiles", "iron-reinforced", "moss-covered stone", "obsidian", "root-penetrated"}

var roomCeilings = []string{"vaulted 15ft", "flat 10ft", "arched 20ft", "caved in, open to sky", "low 7ft", "domed 25ft", "sloped", "beamed 12ft", "natural rock 30ft+", "covered in hanging roots"}

var roomSounds = []string{
	"dripping water echoing", "distant metallic clanking", "faint whispers",
	"scuttling of tiny claws", "low moaning wind", "steady dripping from stalactites",
	"skittering of rats", "creaking of old wood", "hum of arcane energy",
	"distant howling", "stone grinding on stone", "silence—unnaturally still",
}

var roomSmells = []string{
	"musty and damp", "coppery scent of blood", "sweet incense",
	"rotting organic matter", "ozone and brimstone", "earthy mushroom smell",
	"stale and dusty", "acrid chemical fumes", "perfumed with dried flowers",
	"metallic and cold", "smoke and ash", "sickly sweet decay",
}

var roomLights = []string{
	"dim torchlight", "complete darkness", "faint magical glow (blue)",
	"flickering candlelight", "pale moonlight from ceiling cracks", "glowing fungus (green)",
	"no light source needed—everything visible in dim grey", "bright lantern light",
	"pulsing red runes", "swirling motes of faerie fire",
}

var roomTemperatures = []string{"cool (50°F)", "cold (40°F)", "freezing (30°F)", "warm (70°F)", "hot (85°F)", "sweltering (100°F)", "comfortable (65°F)", "variable draft"}

var roomDebris = []string{
	"overturned furniture", "scattered bones", "broken pottery shards",
	"rotting tapestries", "collapsed pillars", "piles of rubble",
	"abandoned mining equipment", "empty crates and barrels", "scattered pages from a journal",
	"lichen-covered statues", "broken weapons and shields", "a single preserved coffin",
}

func HandleGenerateDungeonDressing(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"room_type":    roomTypes[rand.Intn(len(roomTypes))],
		"size":         roomSizes[rand.Intn(len(roomSizes))],
		"shape":        roomShapes[rand.Intn(len(roomShapes))],
		"floor":        roomFloors[rand.Intn(len(roomFloors))],
		"walls":        roomWalls[rand.Intn(len(roomWalls))],
		"ceiling":      roomCeilings[rand.Intn(len(roomCeilings))],
		"sound":        roomSounds[rand.Intn(len(roomSounds))],
		"smell":        roomSmells[rand.Intn(len(roomSmells))],
		"light":        roomLights[rand.Intn(len(roomLights))],
		"temperature":  roomTemperatures[rand.Intn(len(roomTemperatures))],
		"debris":       roomDebris[rand.Intn(len(roomDebris))],
	})
}

// ─── Tavern Generator ───

var tavernPrefixes = []string{"The", "The Golden", "The Silver", "The Crimson", "The Rusty", "The Tipsy", "The Drunken", "The Wobbly", "The Howling", "The Squeaky", "The Lazy", "The Jolly", "The Salty", "The Burnt", "The Broken"}
var tavernAdjectives = []string{"Dancing", "Drunken", "Wandering", "Laughing", "Sleeping", "Singing", "Flying", "Swimming", "Crawling", "Galloping", "Brawling", "Floating", "Roaring", "Whispering", "Leaping"}
var tavernNouns = []string{"Dragon", "Unicorn", "Griffin", "Basilisk", "Kraken", "Phoenix", "Wyvern", "Satyr", "Golem", "Chimera", "Manticore", "Hydra", "Pegasus", "Dire Wolf", "Owlbear"}

var proprietorNames = []string{"Hannah Hops", "Borin Stonebrew", "Seraphina Swift", "Gideon Grey", "Mira Goldleaf", "Thorne Ironfist", "Lyra Silver", "Corbin Malt", "Elara Stout", "Finnigan Brew"}
var proprietorTraits = []string{"friendly but forgetful", "gruff and silent", "tells terrible jokes", "knows everyone's secrets", "always wiping the same glass", "suspicious of outsiders", "motherly and warm", "shady dealings in back", "former adventurer", "obsessive organizer"}

var tavernClientele = []string{
	"locals nursing their drinks", "whispering merchants", "boisterous farmers celebrating a good harvest",
	"nervous travelers keeping to themselves", "off-duty guards", "a bard playing for copper pieces",
	"gamblers around a dice table", "miners fresh from the quarries", "hooded figures in a dark corner",
	"off-duty sailors swapping stories", "a trio of halflings eating heartily", "religious pilgrims resting",
}

var specialtyDrinkNames = []string{
	"Lava Breath", "Siren's Kiss", "Dwarven Stout of Holding", "Elven Dew",
	"Goblin's Grog", "Witch's Brew", "Dragon's Milk", "Silvertongue Ale",
	"Shadowfell Porter", "Feywild Fizz", "Sleeping Giant IPA", "Harpy's Tears",
}

var specialtyDesc = []string{
	"Smoke rises from its surface—it's served warm with cinnamon",
	"A shimmering blue liquid that leaves sparkles on your tongue",
	"So thick you can chew it—served in a hollowed horn",
	"Clear as morning dew with a floral bouquet",
	"A murky green concoction that smells of earth and mystery",
	"Purple and bubbling—changes flavor based on the drinker's mood",
	"Creamy white with a honey aftertaste—rumored to restore vitality",
	"Silvery liquid that coats the mouth—makes you more persuasive",
	"Dark as ink with a heavy body—tastes of mushrooms and stone",
	"Golden with tiny sparkles—refreshing with a hint of citrus and magic",
}

var tavernAtmospheres = []string{
	"warm and smoky with a crackling hearth", "dim and mysterious with shaded lanterns",
	"bright and raucous with a bard on stage", "quiet and somber with sparse patrons",
	"busy and chaotic at peak dinner hour", "clean and orderly with uniformed staff",
	"rustic and cozy with animal heads on walls", "opulent and gilded with crystal chandeliers",
}

var tavernPrices = []string{"cheap (1 cp meal, 2 cp ale)", "moderate (5 cp meal, 4 cp ale)", "expensive (2 sp meal, 1 sp ale)", "luxury (5 sp meal, 3 sp ale)"}

var tavernRumors = []string{
	"The old mill outside town has been seeing strange lights at night.",
	"Merchants have been disappearing on the south road.",
	"The baron is secretly selling weapons to the orc tribes.",
	"A treasure map was found in a book at the local temple.",
	"Strange fish with human eyes have been caught in the river.",
	"The well in the central square is said to grant wishes—for a price.",
	"A dragon was spotted flying north three days ago.",
	"The neighboring kingdom is preparing for war.",
	"Ghostly figures have been seen in the ruined keep.",
	"A thieves' guild is recruiting new members in the shadows.",
	"The local apothecary knows more than she lets on.",
	"There's a hidden passage somewhere beneath the tavern itself.",
}

func HandleGenerateTavern(c *gin.Context) {
	prefix := tavernPrefixes[rand.Intn(len(tavernPrefixes))]
	adj := tavernAdjectives[rand.Intn(len(tavernAdjectives))]
	noun := tavernNouns[rand.Intn(len(tavernNouns))]
	name := prefix + " " + adj + " " + noun

	nRumors := 1 + rand.Intn(2)
	selected := make([]string, nRumors)
	copy(selected, tavernRumors[0:nRumors])
	// shuffle index for variety
	start := rand.Intn(len(tavernRumors) - nRumors)
	for i := 0; i < nRumors; i++ {
		selected[i] = tavernRumors[(start+i)%len(tavernRumors)]
	}

	c.JSON(http.StatusOK, gin.H{
		"name":          name,
		"proprietor":    proprietorNames[rand.Intn(len(proprietorNames))],
		"proprietor_trait": proprietorTraits[rand.Intn(len(proprietorTraits))],
		"clientele": []string{
			tavernClientele[rand.Intn(len(tavernClientele))],
			tavernClientele[rand.Intn(len(tavernClientele))],
			tavernClientele[rand.Intn(len(tavernClientele))],
		},
		"specialty_drink": specialtyDrinkNames[rand.Intn(len(specialtyDrinkNames))],
		"drink_description": specialtyDesc[rand.Intn(len(specialtyDesc))],
		"atmosphere":    tavernAtmospheres[rand.Intn(len(tavernAtmospheres))],
		"prices":        tavernPrices[rand.Intn(len(tavernPrices))],
		"rumors":        selected,
	})
}

// ─── Urban Encounter Generator ───

var urbanThemes = []string{"commerce", "crime", "gossip", "disaster", "festival", "mystery", "politics", "supernatural", "romance", "competition"}

var urbanNPCs = []string{
	"A frantic messenger looking for help", "A street urchin picking pockets",
	"A merchant closing his shop early", "A city watchman questioning bystanders",
	"A noble in a sedan chair being carried through", "A priest handing out pamphlets",
	"A crazed inventor demonstrating a device", "A bard reciting satirical poetry",
	"A foreign dignitary with armed escort", "An old woman feeding stray cats",
	"A known fence negotiating in the shadows", "A drunkard telling tall tales to anyone who listens",
}

var urbanDescriptions = []string{
	"A crowd has gathered around a street performer",
	"The market square is bustling with vendors",
	"Smoke rises from a building two blocks away",
	"A procession of robed figures marches through the streets",
	"The town crier announces new decrees at the square",
	"A pickpocket darts through the crowd, a purse in hand",
	"Guards are questioning everyone near the city gate",
	"An impromptu duel is brewing between two heated nobles",
	"Street preachers from different faiths argue on opposite corners",
	"A cart has overturned, spilling fruit and vegetables everywhere",
	"Children chase a runaway goat through the market",
	"A mysterious fog rolls in, carrying faint whispers",
}

var urbanComplications = []string{
	"The guards mistakenly think the party is involved",
	"A child pickpocket steals from a party member during the commotion",
	"The crowd turns hostile toward outsiders",
	"The city watch arrives and starts arresting everyone",
	"A second unrelated event happens simultaneously",
	"The key witness refuses to speak to strangers",
	"The noble involved recognizes a party member from their past",
	"Magical mishap causes a small explosion nearby",
	"A sudden downpour scatters the crowd",
	"The situation is a trap set for the party specifically",
}

var urbanResolutions = []string{
	"Bribe the right people (10-50 gp)", "Prove innocence through investigation",
	"Help resolve the underlying issue", "Fight through the opposition",
	"Diplomatic negotiation with the authorities", "Find the real culprit",
	"Make a deal with a shady contact", "Use a creative distraction to slip away",
	"Offer a service in exchange for freedom", "Wait it out until things calm down",
}

func HandleGenerateUrbanEncounter(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"theme":              urbanThemes[rand.Intn(len(urbanThemes))],
		"npc":                urbanNPCs[rand.Intn(len(urbanNPCs))],
		"description":        urbanDescriptions[rand.Intn(len(urbanDescriptions))],
		"complication":       urbanComplications[rand.Intn(len(urbanComplications))],
		"possible_resolution": urbanResolutions[rand.Intn(len(urbanResolutions))],
	})
}

// ─── Road Encounter Generator ───

var roadTerrainTypes = []string{"plains", "forest", "mountain pass", "swamp", "coastal road", "desert", "river crossing", "hills", "farmland", "valley"}

var roadEncounterTypes = []string{"combat", "social", "exploration", "hazard", "treasure", "weather", "mystical", "caravan"}

var roadDescriptions = []string{
	"The party spots smoke on the horizon—a campsite or wreckage?",
	"A broken-down wagon blocks the path, its owner nowhere in sight",
	"Strange tracks cross the road—something dragged a heavy load",
	"An overturned cart spills goods across the trail",
	"A lone traveler approaches from the opposite direction",
	"The bridge ahead is collapsed, requiring a detour",
	"A group of figures waits by the roadside, watching",
	"Vultures circle overhead, marking something ahead",
	"A heavy fog rolls in, reducing visibility to 20 feet",
	"Scavenged bones and broken gear litter the roadside",
	"A small shrine sits at a crossroads, offerings left at its base",
	"The sound of rushing water grows louder—a flash flood risk",
}

var roadCreatures = []string{
	"2d4 goblins with a hobgoblin leader", "1d4+1 wolves hunting for food",
	"A traveling merchant with hired guards", "1d6 bandits demanding toll",
	"A wounded griffin unable to fly", "1d4 orcs returning from a raid",
	"A lost child from a nearby settlement", "A wandering knight seeking jousting opponent",
	"2d6 giant wasps protecting a nest", "A traveling circus with exotic animals",
	"Animated scarecrows in an adjacent field", "A solitary hill giant looking for food",
}

var roadLootHints = []string{
	"Gleaming metal visible in the wreckage", "A locked chest half-buried in mud",
	"The corpse of a previous traveler still has its pack", "A satchel hanging from a tree branch",
	"Glittering coins scattered among the rocks", "An abandoned camp with supplies",
	"A dead horse with intact saddlebags", "A hidden cache under a marked stone",
	"Jewelry on a corpse half-eaten by scavengers", "A merchant willing to trade information for escort",
}

var roadComplications = []string{
	"Reinforcements arrive in 1d4 rounds", "The terrain makes retreat difficult",
	"A storm is approaching rapidly", "The encounter site is a monster's lair",
	"Another party is also interested in the same prize", "The supposed victim is the actual threat",
	"A territorial beast joins the fray", "The encounter attracts scavengers",
	"Magical exhaustion affects spellcasters nearby", "The noise alerts nearby creatures",
	"One of the creatures has a disease the party can catch", "The ground gives way to a hidden pit or burrow",
}

func HandleGenerateRoadEncounter(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"terrain":       roadTerrainTypes[rand.Intn(len(roadTerrainTypes))],
		"encounter_type": roadEncounterTypes[rand.Intn(len(roadEncounterTypes))],
		"description":   roadDescriptions[rand.Intn(len(roadDescriptions))],
		"creatures":     roadCreatures[rand.Intn(len(roadCreatures))],
		"loot_hint":     roadLootHints[rand.Intn(len(roadLootHints))],
		"complication":  roadComplications[rand.Intn(len(roadComplications))],
	})
}
