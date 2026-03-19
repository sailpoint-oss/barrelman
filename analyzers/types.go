package analyzers

import (
	navigator "github.com/sailpoint-oss/navigator"
	"github.com/sailpoint-oss/barrelman"
)

var noUnknownFormatsMeta = barrelman.RuleMeta{
	ID:          "no-unknown-formats",
	Description: "Schema format values should be known/standard formats.",
	Severity:    barrelman.SeverityWarning,
	Category:    barrelman.CategoryTypes,
	Recommended: true,
	HowToFix:    "Use a known format (e.g., date-time, email, uri, uuid, int32, int64, float, double).",
	DocURL:      barrelman.DocBaseURL + "no-unknown-formats",
}

var knownFormats = map[string]bool{
	"int32": true, "int64": true,
	"float": true, "double": true,
	"byte": true, "binary": true,
	"date": true, "date-time": true,
	"password": true, "email": true,
	"hostname": true, "uri": true,
	"uri-reference": true, "uuid": true,
	"ipv4": true, "ipv6": true,
	"uri-template": true, "json-pointer": true,
	"relative-json-pointer": true, "regex": true,
	"duration": true, "time": true,
}

func registerTypesAnalyzers(reg *barrelman.Registry) {
	barrelman.Define("no-unknown-formats", noUnknownFormatsMeta).Schemas(
		func(name string, schema *navigator.Schema, pointer string, r *barrelman.Reporter) {
			if schema.Format != "" && !knownFormats[schema.Format] {
				r.At(schema.Loc, "Unknown format '%s' at %s", schema.Format, pointer)
			}
		},
	).Register(reg)
}
