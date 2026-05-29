package analyzers

import (
	"github.com/sailpoint-oss/barrelman"
	navigator "github.com/sailpoint-oss/navigator"
)

var serversDefinedMeta = barrelman.RuleMeta{
	ID:          "oas3-api-servers",
	Description: "OpenAPI document should define at least one server.",
	Severity:    barrelman.SeverityWarning,
	Category:    barrelman.CategoryServers,
	Recommended: true,
	HowToFix:    "Add a 'servers' section with at least one server URL.",
	DocURL:      barrelman.DocBaseURL + "oas3-api-servers",
}

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
}
