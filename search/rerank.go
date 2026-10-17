package search

import (
	"sort"
	"strings"

	"github.com/agext/levenshtein"
)

// DefaultRerank scores candidates by normalized Levenshtein distance to name
// with boosts for exact and prefix matches, stable sorted by (rerankScore, bm25).
func DefaultRerank(query string, in []CompendiumResult) []CompendiumResult {
	if len(in) == 0 {
		return in
	}
	q := strings.ToLower(strings.TrimSpace(query))
	out := make([]CompendiumResult, len(in))
	copy(out, in)
	type scored struct {
		idx   int
		score float64
	}
	scores := make([]scored, len(out))
	for i, c := range out {
		name := strings.ToLower(c.Name)
		var s float64
		if name == q {
			s = -2
		} else if strings.HasPrefix(name, q) {
			s = -1
		} else {
			dist := levenshtein.Distance(q, name, nil)
			maxLen := len(q)
			if len(name) > maxLen {
				maxLen = len(name)
			}
			if maxLen == 0 {
				s = 0
			} else {
				s = float64(dist) / float64(maxLen)
			}
		}
		// tie-breaker: use negative bm25 contribution slightly
		// lower score = better, so subtract small fraction for high text relevance
		s -= c.Score * 0.0001
		scores[i] = scored{idx: i, score: s}
	}
	sort.SliceStable(scores, func(a, b int) bool {
		return scores[a].score < scores[b].score
	})
	result := make([]CompendiumResult, len(out))
	for i, sc := range scores {
		result[i] = out[sc.idx]
	}
	return result
}
