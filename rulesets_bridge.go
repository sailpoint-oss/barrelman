package barrelman

import "github.com/sailpoint-oss/barrelman/rulesets"

func init() {
	rulesets.SetBuiltinCatalogProvider(func() []rulesets.CatalogRule {
		metas := DefaultRegistry.All()
		out := make([]rulesets.CatalogRule, 0, len(metas))
		for _, meta := range metas {
			out = append(out, rulesets.CatalogRule{
				ID:          meta.ID,
				Severity:    rulesets.Severity(meta.Severity),
				Category:    string(meta.Category),
				Recommended: meta.Recommended,
			})
		}
		return out
	})
}
