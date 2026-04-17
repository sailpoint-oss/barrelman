// Package message renders engine Issues into three human-facing tiers:
//
//   - Terse: a single-line CI-friendly string (`at path: expected X, received Y`).
//   - Structured: a JSON object matching the Issue shape plus an optional
//     humanReadable field.
//   - LSP: an adapter to the github.com/LukasParke/gossip/protocol.Diagnostic
//     shape is provided by the caller since this package must not pull in
//     gossip.
//
// "Did you mean" suggestions are produced here (rather than in the engine
// core) so that callers can opt in to Levenshtein-based suggestion cost.
package message

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sailpoint-oss/barrelman/validation/engine"
)

// Terse renders a single-line human string for CI log lines.
func Terse(issue engine.Issue) string {
	var b strings.Builder
	b.WriteString("at ")
	b.WriteString(issue.HumanPath())
	b.WriteString(": ")
	b.WriteString(issue.Message)
	if issue.Expected != "" && issue.Received != "" {
		fmt.Fprintf(&b, " (expected %s, received %s)", issue.Expected, issue.Received)
	}
	if issue.Suggestion != "" {
		fmt.Fprintf(&b, " — did you mean %q?", issue.Suggestion)
	}
	return b.String()
}

// StructuredBytes returns the issue encoded as JSON with a humanReadable
// field populated for consumer UIs that want both machine and display
// forms.
func StructuredBytes(issue engine.Issue) ([]byte, error) {
	wrapper := struct {
		engine.Issue
		HumanReadable string `json:"humanReadable"`
	}{
		Issue:         issue,
		HumanReadable: Terse(issue),
	}
	return json.Marshal(wrapper)
}

// DidYouMean augments issues with a Suggestion field when the Code is one
// the suggester can help with (`additional-properties` / `enum`). known
// is the list of valid alternatives (property names or enum values).
func DidYouMean(issue engine.Issue, received string, known []string) engine.Issue {
	if received == "" || len(known) == 0 {
		return issue
	}
	best, dist := closest(received, known)
	if best != "" && dist <= maxEditsFor(received) {
		issue.Suggestion = best
	}
	return issue
}

func maxEditsFor(s string) int {
	// Allow 1 edit for short strings, 2 for medium, 3 for long.
	switch {
	case len(s) <= 4:
		return 1
	case len(s) <= 8:
		return 2
	default:
		return 3
	}
}

// closest returns the candidate with the smallest Levenshtein distance
// to target, or "" when the pool is empty.
func closest(target string, candidates []string) (string, int) {
	best := ""
	bestDist := -1
	for _, c := range candidates {
		d := levenshtein(target, c)
		if bestDist < 0 || d < bestDist {
			best = c
			bestDist = d
		}
	}
	return best, bestDist
}

// levenshtein computes the edit distance between a and b using dynamic
// programming. O(len(a) * len(b)) time and O(len(b)) space — suitable
// for the small strings we compare (property names, enum values).
func levenshtein(a, b string) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}
	prev := make([]int, lb+1)
	curr := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}
	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			del := prev[j] + 1
			ins := curr[j-1] + 1
			sub := prev[j-1] + cost
			curr[j] = min3(del, ins, sub)
		}
		prev, curr = curr, prev
	}
	return prev[lb]
}

func min3(a, b, c int) int {
	m := a
	if b < m {
		m = b
	}
	if c < m {
		m = c
	}
	return m
}
