package handlers

import (
	"math/rand"
	"net/http"

	"github.com/gin-gonic/gin"
)

var npcNames = []string{"Elara", "Finn", "Rowan", "Seraphina", "Dorian", "Lyra", "Cassian", "Mira", "Thorne", "Vesper", "Gideon", "Ivy", "Orin", "Zara", "Kael", "Nova", "Asher", "Wren", "Soren", "Luna"}
var npcSurnames = []string{"Stormwind", "Ironheart", "Moonshadow", "Silverwood", "Blackthorn", "Goldweaver", "Frostbeard", "Dawnbringer", "Nightwalker", "Sunforge"}
var npcRaces = []string{"Human", "Elf", "Dwarf", "Halfling", "Dragonborn", "Gnome", "Half-Elf", "Half-Orc", "Tiefling"}
var npcClasses = []string{"Barbarian", "Bard", "Cleric", "Druid", "Fighter", "Monk", "Paladin", "Ranger", "Rogue", "Sorcerer", "Warlock", "Wizard"}
var npcPersonalities = []string{"Brave", "Shy", "Greedy", "Generous", "Suspicious", "Trusting", "Arrogant", "Humble", "Mysterious", "Cheerful", "Grumpy", "Wise", "Foolish", "Honest", "Deceitful"}
var npcQuirks = []string{"Speaks in riddles", "Forgets names", "Collects spoons", "Afraid of heights", "Talks to animals", "Perfect recall", "Laughs at everything", "Refuses magic", "Always whistling", "Bargains constantly"}

func HandleGenerateNPC(c *gin.Context) {
	name := npcNames[rand.Intn(len(npcNames))] + " " + npcSurnames[rand.Intn(len(npcSurnames))]
	race := npcRaces[rand.Intn(len(npcRaces))]
	cls := npcClasses[rand.Intn(len(npcClasses))]
	personality := npcPersonalities[rand.Intn(len(npcPersonalities))]
	quirk := npcQuirks[rand.Intn(len(npcQuirks))]

	c.JSON(http.StatusOK, gin.H{
		"name":        name,
		"race":        race,
		"class":       cls,
		"personality": personality,
		"quirk":       quirk,
		"appearance":  genAppearance(race),
	})
}

func genAppearance(race string) string {
	heights := map[string]string{
		"Dwarf": "4'5\"", "Elf": "5'10\"", "Halfling": "3'3\"", "Gnome": "3'6\"",
		"Human": "5'9\"", "Dragonborn": "6'5\"", "Half-Elf": "5'11\"", "Half-Orc": "6'2\"", "Tiefling": "5'10\"",
	}
	colors := []string{"black", "brown", "blonde", "red", "white", "gray", "auburn"}
	eyes := []string{"blue", "green", "brown", "hazel", "gray", "amber"}
	height := heights[race]
	if height == "" {
		height = "5'8\""
	}
	return height + ", " + colors[rand.Intn(len(colors))] + " hair, " + eyes[rand.Intn(len(eyes))] + " eyes"
}

func HandleGenerateName(c *gin.Context) {
	race := c.DefaultQuery("race", "human")
	var firsts, lasts []string
	switch race {
	case "dwarf":
		firsts = []string{"Adrik", "Baern", "Borin", "Dain", "Darrak", "Eberk", "Grimm", "Helja", "Kathra", "Mardred", "Riswynn", "Torbera", "Ursula", "Vistra"}
		lasts = []string{"Hornblower", "Ironfoot", "Stonehelm", "Battlehammer", "Deepdelve", "Goldfinder", "Redbeard", "Strongaxe", "Stoutshield"}
	case "elf":
		firsts = []string{"Aerendir", "Celeste", "Elowen", "Faelar", "Galad", "Ithil", "Laeron", "Luthien", "Maiele", "Nimloth", "Rael", "Silvan", "Thalas", "Variel"}
		lasts = []string{"Amakiir", "Caebrek", "Holimion", "Kinder", "Liadon", "Meliamne", "Naïlo", "Siannodel", "Xiloscient"}
	case "halfling":
		firsts = []string{"Andry", "Bree", "Callie", "Corrin", "Dannad", "Elden", "Garret", "Haldo", "Jillian", "Kithri", "Lavon", "Nedda", "Paela", "Rosalind"}
		lasts = []string{"Brushgather", "Goodbarrel", "Greenbottle", "High-hill", "Hornblower", "Roper", "Talbot", "Underbough"}
	default:
		firsts = []string{"Aric", "Bella", "Cade", "Dara", "Elara", "Finn", "Gwen", "Hale", "Iris", "Jace", "Kira", "Leo", "Maya", "Nash", "Ora", "Pace", "Quinn", "Rena", "Sage", "Tess", "Uma", "Vance", "Willa", "Xander", "Zoe"}
		lasts = []string{"Ashford", "Blackwood", "Davenport", "Farley", "Gallagher", "Hartwell", "Irving", "Kendall", "Mercer", "Nightingale", "Pendleton", "Sterling", "Thornfield", "Westbrook"}
	}

	first := firsts[rand.Intn(len(firsts))]
	last := lasts[rand.Intn(len(lasts))]
	c.JSON(http.StatusOK, gin.H{"name": first + " " + last, "first": first, "last": last, "race": race})
}

func HandleGenerateEncounter(c *gin.Context) {
	terrain := c.DefaultQuery("terrain", "dungeon")
	level := c.DefaultQuery("level", "3")
	terrains := map[string][]gin.H{
		"dungeon": {
			{"monsters": "2d4 Giant Rats + 1 Giant Spider", "xp": 200, "difficulty": "Easy"},
			{"monsters": "3 Skeletons + 1 Minotaur Skeleton", "xp": 700, "difficulty": "Medium"},
			{"monsters": "1 Gelatinous Cube + 2d6 Stirges", "xp": 1100, "difficulty": "Hard"},
		},
		"forest": {
			{"monsters": "1d4 Wolves + 1 Dire Wolf", "xp": 350, "difficulty": "Easy"},
			{"monsters": "2 Druids + 1d4 Twig Blights", "xp": 600, "difficulty": "Medium"},
			{"monsters": "1 Treant + 2d4 Awakened Shrubs", "xp": 1800, "difficulty": "Hard"},
		},
		"mountain": {
			{"monsters": "2d4 Orcs", "xp": 200, "difficulty": "Easy"},
			{"monsters": "1 Griffon + 1d4 Orcs", "xp": 900, "difficulty": "Medium"},
			{"monsters": "1 Young Red Dragon", "xp": 3900, "difficulty": "Deadly"},
		},
		"coast": {
			{"monsters": "1d4 Merfolk + 1 Sea Hag", "xp": 500, "difficulty": "Easy"},
			{"monsters": "1 Water Elemental", "xp": 1800, "difficulty": "Hard"},
			{"monsters": "1 Kraken Priest + 2d4 Sahuagin", "xp": 2800, "difficulty": "Deadly"},
		},
	}
	encs, ok := terrains[terrain]
	if !ok {
		terrain = "dungeon"
		encs = terrains["dungeon"]
	}
	enc := encs[rand.Intn(len(encs))]
	enc["terrain"] = terrain
	enc["level"] = level
	c.JSON(http.StatusOK, enc)
}

func HandleGenerateLoot(c *gin.Context) {
	cr := c.DefaultQuery("cr", "1-4")
	var gp int
	var items []string
	switch cr {
	case "1-4":
		gp = 5 + rand.Intn(50)
		items = []string{"Silver holy symbol", "Potion of Healing", "Scroll of Bless", "Bag of Caltrops", "Silver ring"}
	case "5-10":
		gp = 100 + rand.Intn(400)
		items = []string{"Potion of Greater Healing", "+1 Weapon", "Spell Scroll (2nd)", "Bag of Holding", "Cloak of Protection"}
	case "11-16":
		gp = 1000 + rand.Intn(3000)
		items = []string{"Potion of Superior Healing", "+2 Weapon", "Wand of Magic Missiles", "Ring of Protection", "Boots of Speed"}
	default:
		gp = 5000 + rand.Intn(20000)
		items = []string{"Potion of Supreme Healing", "+3 Weapon", "Rod of Alertness", "Tome of Leadership", "Dragon Scale Mail"}
	}

	nItems := 1 + rand.Intn(3)
	result := []string{items[rand.Intn(len(items))]}
	for i := 1; i < nItems; i++ {
		result = append(result, items[rand.Intn(len(items))])
	}

	c.JSON(http.StatusOK, gin.H{
		"gp":    gp,
		"items": result,
		"cr":    cr,
	})
}

var charNames = []string{"Aldric", "Brienne", "Cedric", "Drizzt", "Elowen", "Fenris", "Gareth", "Helga", "Isolde", "Jasper", "Kaelen", "Lyanna", "Magnus", "Nyx", "Orion", "Petra", "Quinn", "Raven", "Sylas", "Talia", "Ulric", "Vanya", "Wren", "Xander", "Yara", "Zephyr"}
var backgrounds = []string{"Acolyte", "Criminal", "Folk Hero", "Noble", "Sage", "Soldier", "Urchin", "Entertainer", "Guild Artisan", "Hermit", "Outlander", "Sailor"}
var alignments = []string{"Lawful Good", "Neutral Good", "Chaotic Good", "Lawful Neutral", "True Neutral", "Chaotic Neutral", "Lawful Evil", "Neutral Evil", "Chaotic Evil"}
var backstoryHooks = []string{
	"Seeking revenge for a fallen mentor",
	"On a quest to find a lost artifact",
	"Fleeing a dark past in another realm",
	"Bound by an ancient oath",
	"Searching for a missing family member",
	"Carrying a cursed heirloom",
	"Hunted by a powerful organization",
	"Promised a great fortune for a dangerous task",
	"The last of a destroyed lineage",
	"Awakened with strange powers they don't understand",
	"Made a pact with a mysterious entity",
	"Chosen by prophecy for a coming war",
}

func abilityScoreArray() []int {
	// 4d6 drop lowest method, 6 times
	scores := make([]int, 6)
	for i := range 6 {
		rolls := []int{rand.Intn(6) + 1, rand.Intn(6) + 1, rand.Intn(6) + 1, rand.Intn(6) + 1}
		// Find min and remove it
		minIdx := 0
		for j := 1; j < 4; j++ {
			if rolls[j] < rolls[minIdx] {
				minIdx = j
			}
		}
		sum := 0
		for j := range 4 {
			if j != minIdx {
				sum += rolls[j]
			}
		}
		scores[i] = sum
	}
	return scores
}

func HandleGenerateRandomCharacter(c *gin.Context) {
	race := npcRaces[rand.Intn(len(npcRaces))]
	cls := npcClasses[rand.Intn(len(npcClasses))]
	name := charNames[rand.Intn(len(charNames))]
	bg := backgrounds[rand.Intn(len(backgrounds))]
	alignment := alignments[rand.Intn(len(alignments))]
	hook := backstoryHooks[rand.Intn(len(backstoryHooks))]
	personality := npcPersonalities[rand.Intn(len(npcPersonalities))]
	quirk := npcQuirks[rand.Intn(len(npcQuirks))]
	level := rand.Intn(20) + 1
	scores := abilityScoreArray()
	// Assign scores to abilities: assign highest to primary stat based on class
	primaryStats := map[string]string{
		"Barbarian": "str", "Bard": "cha", "Cleric": "wis", "Druid": "wis",
		"Fighter": "str", "Monk": "dex", "Paladin": "str", "Ranger": "dex",
		"Rogue": "dex", "Sorcerer": "cha", "Warlock": "cha", "Wizard": "int",
	}
	primary := primaryStats[cls]
	if primary == "" {
		primary = "str"
	}
	statOrder := []string{"str", "dex", "con", "int", "wis", "cha"}
	// Sort scores descending, but put primary highest
	statValues := make(map[string]int)
	maxScore := 0
	maxIdx := 0
	for i, s := range scores {
		if s > maxScore {
			maxScore = s
			maxIdx = i
		}
	}
	// Swap highest with primary position
	primaryPos := 0
	for i, s := range statOrder {
		if s == primary {
			primaryPos = i
			break
		}
	}
	scores[maxIdx], scores[primaryPos] = scores[primaryPos], scores[maxIdx]

	for i, s := range statOrder {
		statValues[s] = scores[i]
	}

	c.JSON(http.StatusOK, gin.H{
		"name":           name,
		"race":           race,
		"class":          cls,
		"level":          level,
		"background":     bg,
		"alignment":      alignment,
		"personality":    personality,
		"quirk":          quirk,
		"backstory_hook": hook,
		"str":            statValues["str"],
		"dex":            statValues["dex"],
		"con":            statValues["con"],
		"int":            statValues["int"],
		"wis":            statValues["wis"],
		"cha":            statValues["cha"],
	})
}
