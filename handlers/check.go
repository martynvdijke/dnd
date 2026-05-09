package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"villum/db"
)

var skillsMap = map[string]string{
	"athletics":         "str",
	"acrobatics":        "dex",
	"sleight_of_hand":   "dex",
	"stealth":           "dex",
	"arcana":            "int",
	"history":           "int",
	"investigation":     "int",
	"nature":            "int",
	"religion":          "int",
	"animal_handling":   "wis",
	"insight":           "wis",
	"medicine":          "wis",
	"perception":        "wis",
	"survival":          "wis",
	"deception":         "cha",
	"intimidation":      "cha",
	"performance":       "cha",
	"persuasion":        "cha",
}

var savesMap = map[string]string{
	"str": "str", "dex": "dex", "con": "con",
	"int": "int", "wis": "wis", "cha": "cha",
}

type CheckRollRequest struct {
	CharacterID int64  `json:"character_id"`
	Type        string `json:"type"`       // "skill", "save", "check"
	Name        string `json:"name"`        // skill name or save ability
	Advantage   string `json:"advantage"`   // "normal", "advantage", "disadvantage"
	Modifier    int    `json:"modifier"`    // optional extra modifier
}

type CheckRollResult struct {
	Rolls     []int  `json:"rolls"`
	Raw       int    `json:"raw"`
	Total     int    `json:"total"`
	Modifier  int    `json:"modifier"`
	Ability   string `json:"ability"`
	Proficient bool   `json:"proficient"`
	Advantage string `json:"advantage"`
	Text      string `json:"text"`
}

func HandleCheckRoll(c *gin.Context) {
	var req CheckRollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	abils := make(map[string]int)
	profBonus := 0
	var str, dex, con, intel, wis, cha int
	err := db.DB.QueryRow("SELECT str, dex, con, int, wis, cha, proficiency_bonus FROM characters WHERE id=?", req.CharacterID).
		Scan(&str, &dex, &con, &intel, &wis, &cha, &profBonus)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}
	abils["str"] = str
	abils["dex"] = dex
	abils["con"] = con
	abils["int"] = intel
	abils["wis"] = wis
	abils["cha"] = cha
	for k, v := range abils {
		abils[k+"_mod"] = abilityMod(v)
	}

	ability := ""
	isProficient := false
	name := strings.ToLower(req.Name)

	switch req.Type {
	case "skill":
		abilKey, ok := skillsMap[name]
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "unknown skill: " + name})
			return
		}
		ability = abilKey
		// Check if character has this proficiency
		var count int
		db.DB.QueryRow("SELECT COUNT(*) FROM character_proficiencies WHERE character_id=? AND type='skill' AND LOWER(name)=?", req.CharacterID, name).Scan(&count)
		isProficient = count > 0
	case "save":
		abilKey, ok := savesMap[name]
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "unknown save: " + name})
			return
		}
		ability = abilKey
		// Saving throw proficiencies are stored in character_proficiencies with type='save'
		var count int
		db.DB.QueryRow("SELECT COUNT(*) FROM character_proficiencies WHERE character_id=? AND type='save' AND LOWER(name)=?", req.CharacterID, name).Scan(&count)
		isProficient = count > 0
	case "check":
		abilKey, ok := savesMap[name]
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "unknown ability: " + name})
			return
		}
		ability = abilKey
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid type: must be skill/save/check"})
		return
	}

	abilMod := abils[ability+"_mod"]
	totalMod := abilMod + req.Modifier
	if isProficient {
		totalMod += profBonus
	}

	adv := strings.ToLower(req.Advantage)
	if adv == "" {
		adv = "normal"
	}

	rolls := make([]int, 1)
	d20, _ := randInt(1, 20)
	rolls[0] = d20

	if adv == "advantage" {
		d20b, _ := randInt(1, 20)
		rolls = append(rolls, d20b)
		if d20b > d20 {
			d20 = d20b
		}
	} else if adv == "disadvantage" {
		d20b, _ := randInt(1, 20)
		rolls = append(rolls, d20b)
		if d20b < d20 {
			d20 = d20b
		}
	}

	total := d20 + totalMod

	var label string
	switch req.Type {
	case "skill":
		label = cases.Title(language.English).String(name)
	case "save":
		label = cases.Title(language.English).String(ability) + " Save"
	default:
		label = cases.Title(language.English).String(ability) + " Check"
	}
	label = strings.ReplaceAll(label, "_", " ")

	parts := []string{fmt.Sprintf("%s: %d", label, total)}
	rollStr := strings.Trim(strings.Replace(fmt.Sprint(rolls), " ", ", ", -1), "[]")
	parts = append(parts, fmt.Sprintf("(d20: [%s]", rollStr))
	if abilMod >= 0 {
		parts = append(parts, fmt.Sprintf("+%d %s", abilMod, strings.ToUpper(ability)))
	} else {
		parts = append(parts, fmt.Sprintf("%d %s", abilMod, strings.ToUpper(ability)))
	}
	if isProficient {
		parts = append(parts, fmt.Sprintf("+%d Prof", profBonus))
	}
	if req.Modifier != 0 {
		if req.Modifier > 0 {
			parts = append(parts, fmt.Sprintf("+%d Bonus", req.Modifier))
		} else {
			parts = append(parts, fmt.Sprintf("%d Bonus", req.Modifier))
		}
	}
	parts = append(parts, ")")
	text := strings.Join(parts, " ")

	userID, _ := c.Get("user_id")
	db.DB.Exec("INSERT INTO dice_rolls(user_id,character_id,expression,result,total) VALUES(?,?,?,?,?)",
		userID, req.CharacterID, label, text, total)

	c.JSON(http.StatusOK, CheckRollResult{
		Rolls:     rolls,
		Raw:       d20,
		Total:     total,
		Modifier:  totalMod,
		Ability:   ability,
		Proficient: isProficient,
		Advantage: adv,
		Text:      text,
	})
}
