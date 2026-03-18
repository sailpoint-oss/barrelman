package analyzers

import (
	navigator "github.com/LukasParke/navigator"
	"github.com/sailpoint-oss/barrelman"
)

var unusedComponentMeta = barrelman.RuleMeta{
	ID:          "unused-component",
	Description: "Components defined but never referenced are unnecessary.",
	Severity:    barrelman.SeverityWarning,
	Category:    barrelman.CategoryStructure,
	Recommended: true,
	HowToFix:    "Remove the unused component or add a $ref that references it.",
	DocURL:      barrelman.DocBaseURL + "unused-component",
}

func registerUnusedComponentAnalyzers(reg *barrelman.Registry) {
	barrelman.Define("unused-component", unusedComponentMeta).Custom(
		func(idx *navigator.Index, r *barrelman.Reporter) {
			if idx.Document.Components == nil {
				return
			}

			type componentInfo struct {
				kind string
				name string
				loc  navigator.Loc
			}

			var components []componentInfo

			for name, schema := range idx.Document.Components.Schemas {
				components = append(components, componentInfo{"schemas", name, schema.NameLoc})
			}
			for name, resp := range idx.Document.Components.Responses {
				components = append(components, componentInfo{"responses", name, navigator.LocOrFallback(resp.NameLoc, resp.Loc)})
			}
			for name, param := range idx.Document.Components.Parameters {
				components = append(components, componentInfo{"parameters", name, navigator.LocOrFallback(param.NameLoc, param.Loc)})
			}
			for name, ex := range idx.Document.Components.Examples {
				components = append(components, componentInfo{"examples", name, navigator.LocOrFallback(ex.NameLoc, ex.Loc)})
			}
			for name, rb := range idx.Document.Components.RequestBodies {
				components = append(components, componentInfo{"requestBodies", name, navigator.LocOrFallback(rb.NameLoc, rb.Loc)})
			}
			for name, h := range idx.Document.Components.Headers {
				components = append(components, componentInfo{"headers", name, navigator.LocOrFallback(h.NameLoc, h.Loc)})
			}
			for name, ss := range idx.Document.Components.SecuritySchemes {
				components = append(components, componentInfo{"securitySchemes", name, navigator.LocOrFallback(ss.NameLoc, ss.Loc)})
			}
			for name, l := range idx.Document.Components.Links {
				components = append(components, componentInfo{"links", name, navigator.LocOrFallback(l.NameLoc, l.Loc)})
			}

			for _, comp := range components {
				if isRefWrapper(comp.kind, comp.name, idx) {
					continue
				}

				refPath := navigator.ComponentRefPath(comp.kind, comp.name)
				refs := idx.RefsTo(refPath)

				if comp.kind == "securitySchemes" && isSecuritySchemeUsed(comp.name, idx) {
					continue
				}

				if len(refs) == 0 {
					loc := comp.loc
					if loc.Node == nil {
						loc = navigator.Loc{Range: barrelman.FileStartRange}
					}
					r.WithTags(barrelman.DiagnosticTagUnnecessary).
						At(loc, "Component '%s/%s' is defined but never referenced", comp.kind, comp.name)
				}
			}
		},
	).Register(reg)
}

func isSecuritySchemeUsed(name string, idx *navigator.Index) bool {
	for _, req := range idx.Document.Security {
		for _, entry := range req.Entries {
			if entry.Name == name {
				return true
			}
		}
	}
	for _, item := range idx.Document.Paths {
		for _, mo := range item.Operations() {
			for _, req := range mo.Operation.Security {
				for _, entry := range req.Entries {
					if entry.Name == name {
						return true
					}
				}
			}
		}
	}
	return false
}

func isRefWrapper(kind, name string, idx *navigator.Index) bool {
	if idx.Document.Components == nil {
		return false
	}
	switch kind {
	case "securitySchemes":
		if ss, ok := idx.Document.Components.SecuritySchemes[name]; ok {
			return ss.Ref != ""
		}
	case "schemas":
		if s, ok := idx.Document.Components.Schemas[name]; ok {
			return s.Ref != ""
		}
	case "responses":
		if r, ok := idx.Document.Components.Responses[name]; ok {
			return r.Ref != ""
		}
	case "parameters":
		if p, ok := idx.Document.Components.Parameters[name]; ok {
			return p.Ref != ""
		}
	case "requestBodies":
		if rb, ok := idx.Document.Components.RequestBodies[name]; ok {
			return rb.Ref != ""
		}
	case "headers":
		if h, ok := idx.Document.Components.Headers[name]; ok {
			return h.Ref != ""
		}
	case "links":
		if l, ok := idx.Document.Components.Links[name]; ok {
			return l.Ref != ""
		}
	case "examples":
		if e, ok := idx.Document.Components.Examples[name]; ok {
			return e.Ref != ""
		}
	}
	return false
}
