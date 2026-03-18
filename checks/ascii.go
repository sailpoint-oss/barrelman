package checks

import (
	"fmt"
	"unicode/utf8"

	"github.com/sailpoint-oss/barrelman"
)

var asciiMeta = barrelman.RuleMeta{
	ID:          "ascii",
	Description: "Reports non-ASCII characters in the document that may cause interoperability issues.",
	Severity:    barrelman.SeverityWarning,
	Category:    barrelman.CategorySyntax,
	Recommended: true,
	HowToFix:    "Replace non-ASCII characters with their ASCII equivalents or escape sequences.",
	DocURL:      barrelman.DocBaseURL + "ascii",
}

func registerASCII(reg *barrelman.Registry) {
	reg.Register(barrelman.Rule{
		ID:   "ascii",
		Meta: asciiMeta,
		Run: func(ctx *barrelman.AnalysisContext) []barrelman.Diagnostic {
			if ctx.Content == nil {
				return nil
			}
			text := string(ctx.Content)
			var diags []barrelman.Diagnostic

			line := uint32(0)
			col := uint32(0)
			for i := 0; i < len(text); {
				b := text[i]
				if b == '\n' {
					line++
					col = 0
					i++
					continue
				}
				if b <= 127 {
					col++
					i++
					continue
				}
				r, size := utf8.DecodeRuneInString(text[i:])
				runeWidth := uint32(1)
				if r > 0xFFFF {
					runeWidth = 2
				}
				diags = append(diags, barrelman.Diagnostic{
					Range: barrelman.Range{
						Start: barrelman.Position{Line: line, Character: col},
						End:   barrelman.Position{Line: line, Character: col + runeWidth},
					},
					Severity:        barrelman.SeverityWarning,
					Source:          barrelman.Source,
					Code:            "ascii",
					CodeDescription: asciiMeta.DocURL,
					Message:         fmt.Sprintf("Non-ASCII character (U+%04X) at line %d, column %d", r, line+1, col+1),
				})
				col += runeWidth
				i += size
			}
			return diags
		},
	})
}
