package analyzers

import (
	"strings"

	navigator "github.com/sailpoint-oss/navigator"
	"github.com/sailpoint-oss/barrelman"
)

var unresolvedRefMeta = barrelman.RuleMeta{
	ID:          "unresolved-ref",
	Description: "Reports $ref values that cannot be resolved.",
	Severity:    barrelman.SeverityError,
	Category:    barrelman.CategoryReferences,
	Recommended: true,
	HowToFix:    "Check the $ref path and ensure the target component exists.",
	DocURL:      barrelman.DocBaseURL + "unresolved-ref",
}

func registerUnresolvedRef(reg *barrelman.Registry) {
	reg.Register(barrelman.Rule{
		ID:   "unresolved-ref",
		Meta: unresolvedRefMeta,
		Run: func(ctx *barrelman.AnalysisContext) []barrelman.Diagnostic {
			if ctx.Index == nil {
				return nil
			}
			idx := ctx.Index

			r := barrelman.NewReporter("unresolved-ref", unresolvedRefMeta.Severity)

			for target, usages := range idx.Refs {
				if _, err := idx.Resolve(target); err == nil {
					continue
				}

				if strings.HasPrefix(target, "#") {
					suggestion := findClosestRef(target, idx)
					for _, usage := range usages {
						if suggestion != "" {
							r.At(usage.Loc, "Cannot resolve $ref: %s. Did you mean '%s'?", target, suggestion)
						} else {
							r.At(usage.Loc, "Cannot resolve $ref: %s", target)
						}
					}
					continue
				}

				if ctx.Resolver != nil && ctx.URI != "" {
					if ctx.Resolver.CanResolve(ctx.URI, target) {
						continue
					}
				}

				for _, usage := range usages {
					r.At(usage.Loc, "Cannot resolve $ref: %s", target)
				}
			}

			return r.Diagnostics()
		},
	})
}

func findClosestRef(target string, idx *navigator.Index) string {
	if idx.Document == nil || idx.Document.Components == nil {
		return ""
	}

	parts := strings.Split(strings.TrimPrefix(target, "#/"), "/")
	if len(parts) < 2 {
		return ""
	}
	kind := parts[len(parts)-2]
	name := parts[len(parts)-1]

	var available []string
	switch kind {
	case "schemas":
		for n := range idx.Schemas {
			available = append(available, n)
		}
	case "parameters":
		for n := range idx.Parameters {
			available = append(available, n)
		}
	case "responses":
		for n := range idx.Responses {
			available = append(available, n)
		}
	case "securitySchemes":
		for n := range idx.SecuritySchemes {
			available = append(available, n)
		}
	default:
		return ""
	}

	bestDist := len(name)/2 + 1
	bestMatch := ""
	for _, a := range available {
		d := levenshtein(strings.ToLower(name), strings.ToLower(a))
		if d < bestDist {
			bestDist = d
			bestMatch = navigator.ComponentRefPath(kind, a)
		}
	}
	return bestMatch
}

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
			curr[j] = min3(curr[j-1]+1, prev[j]+1, prev[j-1]+cost)
		}
		prev, curr = curr, prev
	}
	return prev[lb]
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
