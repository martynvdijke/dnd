package search

import "strings"

// BuildFTS5Query converts a plain-text query into a safe FTS5 MATCH string.
// It escapes double-quotes, splits on whitespace, and appends * to each term
// for prefix matching.
func BuildFTS5Query(q string) string {
	terms := strings.Fields(q)
	if len(terms) == 0 {
		return ""
	}
	escaped := make([]string, len(terms))
	for i, t := range terms {
		s := strings.ReplaceAll(t, `"`, `""`)
		if len(s) > 0 {
			escaped[i] = `"` + s + `"*`
		}
	}
	return strings.Join(escaped, " AND ")
}
