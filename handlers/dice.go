package handlers

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/dice"
	"villum/models"
)

type DiceRequest struct {
	Expression  string `json:"expression"`
	CharacterID *int64 `json:"character_id,omitempty"`
	Advantage   string `json:"advantage,omitempty"`
}

// HandlerResult is the response format sent to the frontend.
type HandlerResult struct {
	Expression string              `json:"expression"`
	Total      int                 `json:"total"`
	Breakdown  []dice.BreakdownGroup `json:"breakdown"`
	Text       string              `json:"text"`
}

var (
	dicePool *dice.Pool
	diceOnce sync.Once
)

func getDicePool() *dice.Pool {
	diceOnce.Do(func() {
		var err error
		dicePool, err = dice.NewPool(2)
		if err != nil {
			// Will be caught at handler time
			panic(fmt.Sprintf("failed to init dice engine: %v", err))
		}
	})
	return dicePool
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

	// Map advantage/disadvantage to RPG notation
	expression := req.Expression
	adv := strings.ToLower(req.Advantage)
	if adv == "advantage" {
		expression = mapToAdvantageNotation(expression)
		if expression == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "advantage only supported for single die expressions"})
			return
		}
	} else if adv == "disadvantage" {
		expression = mapToDisadvantageNotation(expression)
		if expression == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "disadvantage only supported for single die expressions"})
			return
		}
	}

	pool := getDicePool()
	if pool == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "dice engine not available"})
		return
	}

	rollResult, err := pool.Roll(expression)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Transform to handler result format
	hr, ok := rollResultToHandler(req.Expression, rollResult)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse dice result"})
		return
	}

	// Save to history
	userID, _ := c.Get("user_id")
	db.DB.Exec("INSERT INTO dice_rolls(user_id,character_id,expression,result,total) VALUES(?,?,?,?,?)",
		userID, req.CharacterID, hr.Expression, hr.Text, hr.Total)

	c.JSON(http.StatusOK, hr)
}

// mapToAdvantageNotation maps a simple expression like "1d20" to "2d20kh1" for advantage.
// Returns empty string if not applicable.
func mapToAdvantageNotation(expr string) string {
	diePart, modPart, ok := splitDieAndMod(expr)
	if !ok {
		return ""
	}
	if strings.HasPrefix(diePart, "-") {
		return "" // negative dice not supported
	}
	diePart = strings.TrimPrefix(diePart, "+")

	parts := strings.SplitN(diePart, "d", 2)
	if len(parts) != 2 {
		return ""
	}
	count := parts[0]
	if count == "" {
		count = "1"
	}
	if count != "1" {
		return "" // only single die advantage
	}

	notation := "2d" + parts[1] + "kh1"
	if modPart != "" {
		notation += modPart
	}
	return notation
}

// mapToDisadvantageNotation maps a simple expression like "1d20" to "2d20kl1" for disadvantage.
func mapToDisadvantageNotation(expr string) string {
	diePart, modPart, ok := splitDieAndMod(expr)
	if !ok {
		return ""
	}
	if strings.HasPrefix(diePart, "-") {
		return ""
	}
	diePart = strings.TrimPrefix(diePart, "+")

	parts := strings.SplitN(diePart, "d", 2)
	if len(parts) != 2 {
		return ""
	}
	count := parts[0]
	if count == "" {
		count = "1"
	}
	if count != "1" {
		return ""
	}

	notation := "2d" + parts[1] + "kl1"
	if modPart != "" {
		notation += modPart
	}
	return notation
}

// splitDieAndMod splits an expression like "1d20+5" into die part "1d20" and mod part "+5".
func splitDieAndMod(expr string) (diePart, modPart string, ok bool) {
	expr = strings.ReplaceAll(expr, " ", "")
	expr = strings.ToLower(expr)

	if !strings.Contains(expr, "d") {
		return "", "", false
	}

	// Split on first sign after a die expression
	for i := 0; i < len(expr); i++ {
		if (expr[i] == '+' || expr[i] == '-') && i > 0 && expr[i-1] != 'd' && expr[i-1] != 'k' && expr[i-1] != 'r' && expr[i-1] != '!' {
			diePart = expr[:i]
			modPart = expr[i:]
			return diePart, modPart, true
		}
	}

	return expr, "", true
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

// rollResultToHandler transforms a dice engine RollResult to the HandlerResult format.
func rollResultToHandler(originalExpression string, result *dice.RollResult) (*HandlerResult, bool) {
	// Parse total from json.Number
	total := 0
	totalStr := string(result.Total)
	if totalStr != "" {
		fmt.Sscanf(totalStr, "%d", &total)
	}

	// Use the dice package's transform
	transformed := dice.ToHandlerResult(originalExpression, result)

	return &HandlerResult{
		Expression: originalExpression,
		Total:      transformed.Total,
		Breakdown:  transformed.Breakdown,
		Text:       transformed.Text,
	}, true
}

// randInt generates a random integer in [min, max] using crypto/rand.
// Kept here for use by other handlers (campaign, check, combat, downtime).
func randInt(min, max int) (int, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	if err != nil {
		return 0, err
	}
	return int(n.Int64()) + min, nil
}
