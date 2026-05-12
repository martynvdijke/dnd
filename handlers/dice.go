package handlers

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

type DiceRequest struct {
	Expression  string `json:"expression"`
	CharacterID *int64 `json:"character_id,omitempty"`
	Advantage   string `json:"advantage,omitempty"`
}

type DiceResult struct {
	Expression string      `json:"expression"`
	Total      int         `json:"total"`
	Breakdown  []DieResult `json:"breakdown"`
	Text       string      `json:"text"`
}

type DieResult struct {
	Die    string `json:"die"`
	Rolls  []int  `json:"rolls"`
	Total  int    `json:"total"`
	Mod    int    `json:"mod"`
	Signed string `json:"signed"`
}

func HandleRoll(c *gin.Context) {
	var req DiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.Expression == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "expression required"})
		return
	}

	adv := strings.ToLower(req.Advantage)
	if adv == "advantage" || adv == "disadvantage" {
		handleAdvantageRoll(c, req, adv)
		return
	}

	result, err := rollDice(req.Expression)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Save to history
	userID, _ := c.Get("user_id")
	db.DB.Exec("INSERT INTO dice_rolls(user_id,character_id,expression,result,total) VALUES(?,?,?,?,?)",
		userID, req.CharacterID, result.Expression, result.Text, result.Total)

	c.JSON(http.StatusOK, result)
}

func handleAdvantageRoll(c *gin.Context, req DiceRequest, adv string) {
	expr := strings.ReplaceAll(req.Expression, " ", "")
	expr = strings.ToLower(expr)

	var diePart string
	var modPart int
	if idx := strings.IndexAny(expr, "+-"); idx >= 0 {
		diePart = expr[:idx]
		modPart, _ = strconv.Atoi(expr[idx:])
	} else {
		diePart = expr
	}

	if !strings.Contains(diePart, "d") || strings.Count(diePart, "d") > 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "advantage/disadvantage only supported for single die expressions"})
		return
	}

	parts := strings.SplitN(diePart, "d", 2)
	count, _ := strconv.Atoi(parts[0])
	sides, _ := strconv.Atoi(parts[1])
	if count != 1 || sides <= 0 || sides > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "advantage/disadvantage only supported for single die"})
		return
	}

	roll1, _ := randInt(1, sides)
	roll2, _ := randInt(1, sides)

	chosen := roll1
	if adv == "advantage" && roll2 > chosen {
		chosen = roll2
	}
	if adv == "disadvantage" && roll2 < chosen {
		chosen = roll2
	}

	total := chosen + modPart

	breakdown := []DieResult{
		{
			Die:   diePart,
			Rolls: []int{roll1, roll2},
			Total: chosen,
		},
	}
	if modPart != 0 {
		signed := fmt.Sprintf("%+d", modPart)
		breakdown = append(breakdown, DieResult{Die: signed, Total: modPart, Mod: modPart, Signed: signed})
	}

	keepLabel := "higher"
	if adv == "disadvantage" {
		keepLabel = "lower"
	}
	text := fmt.Sprintf("%s (%s) = %d  (%s: [%d, %d], keeping %s)", expr, adv, total, diePart, roll1, roll2, keepLabel)
	if modPart != 0 {
		text += fmt.Sprintf(" %+d", modPart)
	}

	result := &DiceResult{
		Expression: req.Expression,
		Total:      total,
		Breakdown:  breakdown,
		Text:       text,
	}

	userID, _ := c.Get("user_id")
	db.DB.Exec("INSERT INTO dice_rolls(user_id,character_id,expression,result,total) VALUES(?,?,?,?,?)",
		userID, req.CharacterID, result.Expression, result.Text, result.Total)

	c.JSON(http.StatusOK, result)
}

func GetDiceRolls(c *gin.Context) {
	userID, _ := c.Get("user_id")
	query := "SELECT id, user_id, character_id, expression, result, total, timestamp FROM dice_rolls WHERE user_id=? ORDER BY timestamp DESC LIMIT 50"
	args := []interface{}{userID}

	if charID := c.Query("character_id"); charID != "" {
		query = "SELECT id, user_id, character_id, expression, result, total, timestamp FROM dice_rolls WHERE character_id=? ORDER BY timestamp DESC LIMIT 50"
		args = []interface{}{charID}
	}

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var rolls []models.DiceRoll
	for rows.Next() {
		var r models.DiceRoll
		rows.Scan(&r.ID, &r.UserID, &r.CharacterID, &r.Expression, &r.Result, &r.Total, &r.Timestamp)
		rolls = append(rolls, r)
	}
	c.JSON(http.StatusOK, rolls)
}

func rollDice(expr string) (*DiceResult, error) {
	expr = strings.ReplaceAll(expr, " ", "")
	expr = strings.ToLower(expr)

	result := &DiceResult{
		Expression: expr,
	}

	// Parse expression like "2d6+3", "1d20+5", "3d8+2d6+4"
	total := 0
	parts := splitExpression(expr)

	for _, part := range parts {
		if strings.Contains(part, "d") {
			// Dice roll
			parts := strings.SplitN(part, "d", 2)
			count, _ := strconv.Atoi(parts[0])
			sides, _ := strconv.Atoi(parts[1])

			if count <= 0 || sides <= 0 {
				return nil, fmt.Errorf("invalid dice: %s", part)
			}
			if count > 100 {
				return nil, fmt.Errorf("too many dice (max 100)")
			}
			if sides > 1000 {
				return nil, fmt.Errorf("too many sides (max 1000)")
			}

			dr := DieResult{Die: part}
			rolls := make([]int, count)
			sum := 0
			for i := 0; i < count; i++ {
				n, err := randInt(1, sides)
				if err != nil {
					return nil, err
				}
				rolls[i] = n
				sum += n
			}
			dr.Rolls = rolls
			dr.Total = sum
			dr.Signed = fmt.Sprintf("+%d", sum)
			result.Breakdown = append(result.Breakdown, dr)
			total += sum
		} else {
			// Modifier
			n, _ := strconv.Atoi(part)
			dr := DieResult{Die: part, Total: n, Mod: n}
			if n >= 0 {
				dr.Signed = fmt.Sprintf("+%d", n)
			} else {
				dr.Signed = fmt.Sprintf("%d", n)
			}
			result.Breakdown = append(result.Breakdown, dr)
			total += n
		}
	}

	result.Total = total

	// Build text
	parts2 := []string{}
	for _, dr := range result.Breakdown {
		if strings.Contains(dr.Die, "d") {
			rollStrs := []string{}
			for _, r := range dr.Rolls {
				rollStrs = append(rollStrs, strconv.Itoa(r))
			}
			parts2 = append(parts2, fmt.Sprintf("%s: [%s]", dr.Die, strings.Join(rollStrs, ", ")))
		} else {
			parts2 = append(parts2, fmt.Sprintf("%+d", dr.Mod))
		}
	}
	result.Text = fmt.Sprintf("%s = %d  (%s)", expr, total, strings.Join(parts2, " "))

	return result, nil
}

func splitExpression(expr string) []string {
	var parts []string
	current := ""
	sign := byte('+')

	for i := 0; i < len(expr); i++ {
		ch := expr[i]
		if ch == '+' || ch == '-' {
			if current != "" {
				if sign == '-' {
					parts = append(parts, "-"+current)
				} else {
					parts = append(parts, current)
				}
			}
			current = ""
			sign = ch
		} else {
			current += string(ch)
		}
	}

	if current != "" {
		if sign == '-' {
			parts = append(parts, "-"+current)
		} else {
			parts = append(parts, current)
		}
	}

	return parts
}

func randInt(min, max int) (int, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	if err != nil {
		return 0, err
	}
	return int(n.Int64()) + min, nil
}
