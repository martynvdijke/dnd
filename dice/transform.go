package dice

import (
	"encoding/json"
	"fmt"
	"strings"
)

// DieRollDetail represents a single die roll with metadata for frontend animation.
type DieRollDetail struct {
	Value         int      `json:"value"`
	InitialValue  int      `json:"initialValue,omitempty"`
	UseInTotal    bool     `json:"useInTotal"`
	ModifierFlags string   `json:"modifierFlags,omitempty"`
	Modifiers     []string `json:"modifiers,omitempty"`
}

// BreakdownGroup is a group of dice rolls for one die type.
type BreakdownGroup struct {
	Die   string          `json:"die"`   // e.g. "d6", "d20"
	Rolls []DieRollDetail `json:"rolls"` // individual die results
	Total int             `json:"total"` // sum of used rolls
	Mod   int             `json:"mod,omitempty"`
}

// HandlerResult is the response format sent to the frontend.
type HandlerResult struct {
	Expression string           `json:"expression"`
	Total      int              `json:"total"`
	Breakdown  []BreakdownGroup `json:"breakdown"`
	Text       string           `json:"text"`
}

// ToHandlerResult transforms a dice-roller RollResult into the handler response format.
func ToHandlerResult(expression string, result *RollResult) *HandlerResult {
	hr := &HandlerResult{
		Expression: expression,
	}

	if result == nil {
		return hr
	}

	// Parse total
	totalStr := string(result.Total)
	if totalStr != "" {
		fmt.Sscanf(totalStr, "%d", &hr.Total)
	}

	// Parse the rolls JSON — it's a mixed array of objects and primitives
	var rawRolls []json.RawMessage
	if err := json.Unmarshal(result.Rolls, &rawRolls); err != nil {
		hr.Text = result.Output
		return hr
	}

	total := 0
	textParts := []string{}
	i := 0

	for i < len(rawRolls) {
		raw := rawRolls[i]

		// Check if this element is an object (die roll group) or primitive (modifier)
		rawStr := strings.TrimSpace(string(raw))

		if strings.HasPrefix(rawStr, "{") {
			// Die roll group
			var rollGroup struct {
				Rolls []json.RawMessage `json:"rolls"`
				Value float64           `json:"value"`
				Type  string            `json:"type"`
			}
			if err := json.Unmarshal(raw, &rollGroup); err != nil {
				i++
				continue
			}

			details := make([]DieRollDetail, 0, len(rollGroup.Rolls))
			groupTotal := 0
			rollVals := []string{}

			for _, rRaw := range rollGroup.Rolls {
				var rd struct {
					Value         float64  `json:"value"`
					InitialValue  float64  `json:"initialValue,omitempty"`
					UseInTotal    bool     `json:"useInTotal"`
					ModifierFlags string   `json:"modifierFlags,omitempty"`
					Modifiers     []string `json:"modifiers,omitempty"`
					Type          string   `json:"type"`
				}
				if err := json.Unmarshal(rRaw, &rd); err != nil {
					continue
				}

				detail := DieRollDetail{
					Value:         int(rd.Value),
					InitialValue:  int(rd.InitialValue),
					UseInTotal:    rd.UseInTotal,
					ModifierFlags: rd.ModifierFlags,
					Modifiers:     rd.Modifiers,
				}
				details = append(details, detail)

				if rd.UseInTotal {
					groupTotal += int(rd.Value)
				}

				vStr := fmt.Sprintf("%d", int(rd.Value))
				if !rd.UseInTotal && rd.ModifierFlags != "" {
					vStr += rd.ModifierFlags
				}
				rollVals = append(rollVals, vStr)
			}

			total += groupTotal

			// Infer die label from the expression
			dieLabel := extractDieLabel(expression, len(details))

			// If the roll group has explicit modifier flags indicating dropped dice,
			// the die label wouldn't have changed — use the last known die label
			if dieLabel == "" {
				dieLabel = "d?" // fallback
			}

			hr.Breakdown = append(hr.Breakdown, BreakdownGroup{
				Die:   dieLabel,
				Rolls: details,
				Total: groupTotal,
			})

			textParts = append(textParts, fmt.Sprintf("%s: [%s] = %d", dieLabel, strings.Join(rollVals, ", "), groupTotal))
			i++
		} else if strings.HasPrefix(rawStr, `"`) {
			// String — operator like "+" or "-". Skip, the next element is the value.
			i++
			// Check if next element is a number
			if i < len(rawRolls) {
				nextStr := strings.TrimSpace(string(rawRolls[i]))
				var modVal int
				if _, err := fmt.Sscanf(nextStr, "%d", &modVal); err == nil {
					hr.Breakdown = append(hr.Breakdown, BreakdownGroup{
						Die:   fmt.Sprintf("%+d", modVal),
						Mod:   modVal,
						Total: modVal,
					})
					total += modVal
					if modVal >= 0 {
						textParts = append(textParts, fmt.Sprintf("+%d", modVal))
					} else {
						textParts = append(textParts, fmt.Sprintf("%d", modVal))
					}
					i++
				}
			}
		} else {
			// Bare number — standalone modifier (edge case)
			var modVal float64
			if err := json.Unmarshal(raw, &modVal); err == nil {
				m := int(modVal)
				hr.Breakdown = append(hr.Breakdown, BreakdownGroup{
					Die:   fmt.Sprintf("%+d", m),
					Mod:   m,
					Total: m,
				})
				total += m
				if m >= 0 {
					textParts = append(textParts, fmt.Sprintf("+%d", m))
				} else {
					textParts = append(textParts, fmt.Sprintf("%d", m))
				}
			}
			i++
		}
	}

	// Fallback: if parsing produced nothing but we have total from the engine
	if len(hr.Breakdown) == 0 && hr.Total > 0 {
		total = hr.Total
	}

	hr.Total = total
	if hr.Text == "" && len(textParts) > 0 {
		hr.Text = fmt.Sprintf("%s = %d  (%s)", expression, total, strings.Join(textParts, " "))
	} else if hr.Text == "" {
		hr.Text = result.Output
	}

	return hr
}

// extractDieLabel extracts the die label (e.g. "d6", "d20") from the expression.
func extractDieLabel(expression string, rollCount int) string {
	expr := strings.ReplaceAll(expression, " ", "")
	expr = strings.ToLower(expr)

	parts := splitOnSign(expr)
	for _, part := range parts {
		if strings.Contains(part, "d") {
			dParts := strings.SplitN(part, "d", 2)
			count := 0
			cStr := dParts[0]
			if cStr == "" {
				count = 1
			} else {
				fmt.Sscanf(cStr, "%d", &count)
			}
			if count == 0 || count == rollCount || count == 1 {
				suffix := part[strings.Index(part, "d"):]
				return suffix
			}
		}
	}

	// Fallback: try to find any d-something
	for _, part := range parts {
		if idx := strings.Index(part, "d"); idx >= 0 {
			return part[idx:]
		}
	}

	return ""
}

// splitOnSign splits an expression on + and - signs, preserving the sign.
func splitOnSign(expr string) []string {
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
