package analyzers

import (
	"github.com/sailpoint-oss/barrelman"
	navigator "github.com/sailpoint-oss/navigator"
)

var (
	noAPIKeyInQueryMeta = barrelman.RuleMeta{
		ID:          "no-api-key-in-query",
		Description: "API keys should not be passed in query parameters.",
		Severity:    barrelman.SeverityWarning,
		Category:    barrelman.CategorySecurity,
		Recommended: true,
		HowToFix:    "Use header or cookie authentication instead of query parameters.",
		DocURL:      barrelman.DocBaseURL + "no-api-key-in-query",
	}

	oauthFlowURLsMeta = barrelman.RuleMeta{
		ID:          "oauth-flow-urls",
		Description: "OAuth flow URLs should be absolute and use HTTPS.",
		Severity:    barrelman.SeverityWarning,
		Category:    barrelman.CategorySecurity,
		Recommended: true,
		HowToFix:    "Use absolute HTTPS URLs for OAuth flow endpoints.",
		DocURL:      barrelman.DocBaseURL + "oauth-flow-urls",
	}

	securitySchemesDefinedMeta = barrelman.RuleMeta{
		ID:          "security-schemes-defined",
		Description: "Security requirements must reference defined security schemes.",
		Severity:    barrelman.SeverityError,
		Category:    barrelman.CategorySecurity,
		Recommended: true,
		HowToFix:    "Define the security scheme in components/securitySchemes.",
		DocURL:      barrelman.DocBaseURL + "security-schemes-defined",
	}
)

func registerSecurityAnalyzers(reg *barrelman.Registry) {
	barrelman.Define("no-api-key-in-query", noAPIKeyInQueryMeta).SecuritySchemes(
		func(name string, ss *navigator.SecurityScheme, r *barrelman.Reporter) {
			if ss.Type == "apiKey" && ss.In == "query" {
				r.At(ss.Loc, "Security scheme '%s' passes API key in query; use header instead", name)
			}
		},
	).Register(reg)

	barrelman.Define("oauth-flow-urls", oauthFlowURLsMeta).SecuritySchemes(
		func(name string, ss *navigator.SecurityScheme, r *barrelman.Reporter) {
			if ss.Flows == nil {
				return
			}
			flows := []*navigator.OAuthFlow{
				ss.Flows.Implicit,
				ss.Flows.Password,
				ss.Flows.ClientCredentials,
				ss.Flows.AuthorizationCode,
			}
			for _, flow := range flows {
				if flow == nil {
					continue
				}
				if flow.AuthorizationURL != "" && !isHTTPS(flow.AuthorizationURL) {
					r.At(flow.AuthorizationURLLoc, "OAuth authorizationUrl in '%s' should use HTTPS", name)
				}
				if flow.TokenURL != "" && !isHTTPS(flow.TokenURL) {
					r.At(flow.TokenURLLoc, "OAuth tokenUrl in '%s' should use HTTPS", name)
				}
			}
		},
	).Register(reg)

	barrelman.Define("security-schemes-defined", securitySchemesDefinedMeta).Custom(
		func(idx *navigator.Index, r *barrelman.Reporter) {
			allReqs := append([]navigator.SecurityRequirement{}, idx.Document.Security...)
			for _, item := range idx.Document.Paths {
				for _, mo := range item.Operations() {
					allReqs = append(allReqs, mo.Operation.Security...)
				}
			}

			var availableSchemes []string
			for name := range idx.SecuritySchemes {
				availableSchemes = append(availableSchemes, name)
			}

			for _, req := range allReqs {
				for _, entry := range req.Entries {
					if _, ok := idx.SecuritySchemes[entry.Name]; !ok {
						loc := entry.NameLoc
						if loc.Node == nil {
							loc = navigator.Loc{Range: barrelman.FileStartRange}
						}
						suggestion := closestString(entry.Name, availableSchemes)
						if suggestion != "" {
							r.At(loc, "Security requirement references undefined scheme '%s'. Did you mean '%s'?", entry.Name, suggestion)
						} else {
							r.At(loc, "Security requirement references undefined scheme '%s'", entry.Name)
						}
					}
				}
			}
		},
	).Register(reg)
}
