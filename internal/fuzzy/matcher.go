package fuzzy

import (
	"strings"

	"github.com/sahilm/fuzzy"
)

// MatchResult represents a fuzzy match result
type MatchResult struct {
	SessionID string
	Score     int
	Matched   string
}

// Match performs fuzzy matching on sessions
func Match(query string, sessions []string) []MatchResult {
	matches := fuzzy.Find(query, sessions)

	results := make([]MatchResult, len(matches))
	for i, match := range matches {
		results[i] = MatchResult{
			SessionID: match.Str,
			Score:     match.Score,
			Matched:   match.Str,
		}
	}

	return results
}

// NormalizeQuery normalizes search query
func NormalizeQuery(query string) string {
	return strings.ToLower(strings.TrimSpace(query))
}
