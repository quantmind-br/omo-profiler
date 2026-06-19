package views

import "strings"

// rankModelMatch scores how well a model matches a search term so that exact and
// substring matches rank above loose fuzzy subsequence matches. The fuzzy matcher
// (used for membership) scores subsequences across the combined provider/id/name
// string, which lets weak matches outrank obvious ones; callers re-sort their
// fuzzy results with this score (descending, stable) to fix the ordering.
//
// Higher is better:
//
//	4 — exact match on the model id or display name
//	3 — model id or display name starts with the term
//	2 — term is a substring of the model id or display name
//	1 — term is a substring of the provider
//	0 — fuzzy subsequence only
func rankModelMatch(term, provider, modelID, displayName string) int {
	t := strings.ToLower(strings.TrimSpace(term))
	if t == "" {
		return 0
	}
	id := strings.ToLower(modelID)
	name := strings.ToLower(displayName)
	prov := strings.ToLower(provider)

	switch {
	case id == t || name == t:
		return 4
	case strings.HasPrefix(id, t) || strings.HasPrefix(name, t):
		return 3
	case strings.Contains(id, t) || strings.Contains(name, t):
		return 2
	case strings.Contains(prov, t):
		return 1
	default:
		return 0
	}
}
