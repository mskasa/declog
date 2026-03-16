package search

import (
	"path/filepath"
	"strings"

	"github.com/mskasa/declog/internal/decision"
)

// stopWords are common English words excluded from keyword matching.
var stopWords = map[string]struct{}{
	"a": {}, "an": {}, "the": {}, "to": {}, "for": {}, "of": {},
	"in": {}, "on": {}, "at": {}, "by": {}, "with": {}, "and": {},
	"or": {}, "is": {}, "as": {},
}

// Keywords splits title into lowercase words and removes stop words.
func Keywords(title string) []string {
	var result []string
	for _, word := range strings.Fields(title) {
		w := strings.ToLower(word)
		if _, skip := stopWords[w]; !skip && w != "" {
			result = append(result, w)
		}
	}
	return result
}

// Similar returns decisions in dir that contain any keyword derived from title.
// Results are deduplicated and sorted by ID ascending.
func Similar(dir, title string) ([]*decision.Decision, error) {
	keywords := Keywords(title)
	if len(keywords) == 0 {
		return nil, nil
	}

	seen := map[string]struct{}{}
	var decisions []*decision.Decision

	for _, kw := range keywords {
		results, err := RunCaseInsensitive(dir, kw)
		if err != nil {
			return nil, err
		}
		for _, r := range results {
			base := filepath.Base(r.File)
			if _, ok := seen[base]; ok {
				continue
			}
			seen[base] = struct{}{}
			d, err := decision.Parse(r.File)
			if err != nil {
				return nil, err
			}
			decisions = append(decisions, d)
		}
	}

	// Sort by ID ascending.
	for i := 0; i < len(decisions)-1; i++ {
		for j := i + 1; j < len(decisions); j++ {
			if decisions[i].ID > decisions[j].ID {
				decisions[i], decisions[j] = decisions[j], decisions[i]
			}
		}
	}
	return decisions, nil
}
