package analyzers

import (
	"strings"

	"github.com/sailpoint-oss/barrelman"
)

func isCapitalized(s string) bool     { return barrelman.IsCapitalized(s) }
func isKebabCase(s string) bool       { return barrelman.IsKebabCase(s) }
func containsHTTPVerb(s string) bool   { return barrelman.ContainsHTTPVerb(s) }
func hasTrailingSlash(path string) bool { return barrelman.HasTrailingSlash(path) }
func isHTTPS(url string) bool          { return barrelman.IsHTTPS(url) }
func containsCredentials(url string) bool { return barrelman.ContainsCredentials(url) }

func closestString(target string, candidates []string) string {
	bestDist := len(target)/2 + 1
	bestMatch := ""
	tLower := strings.ToLower(target)
	for _, c := range candidates {
		d := levenshtein(tLower, strings.ToLower(c))
		if d < bestDist {
			bestDist = d
			bestMatch = c
		}
	}
	return bestMatch
}
