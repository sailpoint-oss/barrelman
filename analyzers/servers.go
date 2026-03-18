package analyzers

import (
	navigator "github.com/LukasParke/navigator"
	"github.com/sailpoint-oss/barrelman"
)

var (
	serversDefinedMeta = barrelman.RuleMeta{
		ID:          "oas3-api-servers",
		Description: "OpenAPI document should define at least one server.",
		Severity:    barrelman.SeverityWarning,
		Category:    barrelman.CategoryServers,
		Recommended: true,
		HowToFix:    "Add a 'servers' section with at least one server URL.",
		DocURL:      barrelman.DocBaseURL + "oas3-api-servers",
	}

	serverURLHTTPSMeta = barrelman.RuleMeta{
		ID:          "server-url-https",
		Description: "Server URLs should use HTTPS.",
		Severity:    barrelman.SeverityWarning,
		Category:    barrelman.CategoryServers,
		Recommended: true,
		HowToFix:    "Change the server URL to use https:// instead of http://.",
		DocURL:      barrelman.DocBaseURL + "server-url-https",
	}
)

func registerServersAnalyzers(reg *barrelman.Registry) {
	barrelman.Define("oas3-api-servers", serversDefinedMeta).Custom(
		func(idx *navigator.Index, r *barrelman.Reporter) {
			if !idx.IsOpenAPI() {
				return
			}
			if len(idx.Document.Servers) == 0 {
				r.AtRange(barrelman.FileStartRange, "No servers defined; add a 'servers' section")
			}
		},
	).Register(reg)

	barrelman.Define("server-url-https", serverURLHTTPSMeta).Custom(
		func(idx *navigator.Index, r *barrelman.Reporter) {
			if !idx.IsOpenAPI() {
				return
			}
			for _, srv := range idx.Document.Servers {
				if srv.URL != "" && !isHTTPS(srv.URL) {
					r.At(srv.URLLoc, "Server URL '%s' should use HTTPS", srv.URL)
				}
			}
		},
	).Register(reg)
}
