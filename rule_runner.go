package barrelman

import (
	"github.com/sailpoint-oss/barrelman/rulesets"
	navigator "github.com/sailpoint-oss/navigator"
)

// runRules executes rules while batching adjacent RuleBuilder-produced rules
// into one model traversal. Manually authored rules keep their existing Run
// behavior, which preserves compatibility for structural, reference, syntax,
// spectral, and downstream custom rules.
func runRules(ctx *AnalysisContext, rules []Rule, severityOverrides map[string]rulesets.SeverityOverride) []Diagnostic {
	var diags []Diagnostic
	for i := 0; i < len(rules); {
		if !hasVisitors(rules[i].visitors) {
			rule := rules[i]
			diags = append(diags, applyRuleSeverityOverride(rule.Run(ctx), rule.ID, severityOverrides)...)
			i++
			continue
		}

		start := i
		for i < len(rules) && hasVisitors(rules[i].visitors) {
			i++
		}
		diags = append(diags, runVisitorRules(ctx, rules[start:i], severityOverrides)...)
	}
	return diags
}

func runVisitorRules(ctx *AnalysisContext, rules []Rule, severityOverrides map[string]rulesets.SeverityOverride) []Diagnostic {
	if ctx.Index == nil || len(rules) == 0 {
		return nil
	}

	reporters := make([]*Reporter, len(rules))
	for i, rule := range rules {
		reporters[i] = NewReporter(rule.ID, rule.Meta.Severity)
	}

	combined := combineVisitors(rules, reporters)
	Walk(ctx.Index, combined, nil)

	var diags []Diagnostic
	for i, rule := range rules {
		diags = append(diags, applyRuleSeverityOverride(reporters[i].Diagnostics(), rule.ID, severityOverrides)...)
	}
	return diags
}

func combineVisitors(rules []Rule, reporters []*Reporter) Visitors {
	var combined Visitors
	if anyVisitor(rules, func(v Visitors) bool { return v.Document != nil }) {
		combined.Document = func(doc *navigator.Document, _ *Reporter) {
			for i, rule := range rules {
				if fn := rule.visitors.Document; fn != nil {
					fn(doc, reporters[i])
				}
			}
		}
	}
	if anyVisitor(rules, func(v Visitors) bool { return v.Info != nil }) {
		combined.Info = func(info *navigator.Info, _ *Reporter) {
			for i, rule := range rules {
				if fn := rule.visitors.Info; fn != nil {
					fn(info, reporters[i])
				}
			}
		}
	}
	if anyVisitor(rules, func(v Visitors) bool { return v.Path != nil }) {
		combined.Path = func(path string, item *navigator.PathItem, _ *Reporter) {
			for i, rule := range rules {
				if fn := rule.visitors.Path; fn != nil {
					fn(path, item, reporters[i])
				}
			}
		}
	}
	if anyVisitor(rules, func(v Visitors) bool { return v.Operation != nil }) {
		combined.Operation = func(path string, method string, op *navigator.Operation, _ *Reporter) {
			for i, rule := range rules {
				if fn := rule.visitors.Operation; fn != nil {
					fn(path, method, op, reporters[i])
				}
			}
		}
	}
	if anyVisitor(rules, func(v Visitors) bool { return v.Schema != nil }) {
		combined.Schema = func(name string, schema *navigator.Schema, pointer string, _ *Reporter) {
			for i, rule := range rules {
				if fn := rule.visitors.Schema; fn != nil {
					fn(name, schema, pointer, reporters[i])
				}
			}
		}
	}
	if anyVisitor(rules, func(v Visitors) bool { return v.RecursiveSchema != nil }) {
		combined.RecursiveSchema = func(name string, schema *navigator.Schema, pointer string, _ *Reporter) {
			for i, rule := range rules {
				if fn := rule.visitors.RecursiveSchema; fn != nil {
					fn(name, schema, pointer, reporters[i])
				}
			}
		}
	}
	if anyVisitor(rules, func(v Visitors) bool { return v.Parameter != nil }) {
		combined.Parameter = func(param *navigator.Parameter, _ *Reporter) {
			for i, rule := range rules {
				if fn := rule.visitors.Parameter; fn != nil {
					fn(param, reporters[i])
				}
			}
		}
	}
	if anyVisitor(rules, func(v Visitors) bool { return v.Response != nil }) {
		combined.Response = func(code string, resp *navigator.Response, _ *Reporter) {
			for i, rule := range rules {
				if fn := rule.visitors.Response; fn != nil {
					fn(code, resp, reporters[i])
				}
			}
		}
	}
	if anyVisitor(rules, func(v Visitors) bool { return v.Tag != nil }) {
		combined.Tag = func(tag *navigator.Tag, _ *Reporter) {
			for i, rule := range rules {
				if fn := rule.visitors.Tag; fn != nil {
					fn(tag, reporters[i])
				}
			}
		}
	}
	if anyVisitor(rules, func(v Visitors) bool { return v.Server != nil }) {
		combined.Server = func(server *navigator.Server, _ *Reporter) {
			for i, rule := range rules {
				if fn := rule.visitors.Server; fn != nil {
					fn(server, reporters[i])
				}
			}
		}
	}
	if anyVisitor(rules, func(v Visitors) bool { return v.RequestBody != nil }) {
		combined.RequestBody = func(path string, method string, rb *navigator.RequestBody, _ *Reporter) {
			for i, rule := range rules {
				if fn := rule.visitors.RequestBody; fn != nil {
					fn(path, method, rb, reporters[i])
				}
			}
		}
	}
	if anyVisitor(rules, func(v Visitors) bool { return v.SecurityScheme != nil }) {
		combined.SecurityScheme = func(name string, ss *navigator.SecurityScheme, _ *Reporter) {
			for i, rule := range rules {
				if fn := rule.visitors.SecurityScheme; fn != nil {
					fn(name, ss, reporters[i])
				}
			}
		}
	}
	if anyVisitor(rules, func(v Visitors) bool { return v.Example != nil }) {
		combined.Example = func(name string, ex *navigator.Example, _ *Reporter) {
			for i, rule := range rules {
				if fn := rule.visitors.Example; fn != nil {
					fn(name, ex, reporters[i])
				}
			}
		}
	}
	if anyVisitor(rules, func(v Visitors) bool { return v.Custom != nil }) {
		combined.Custom = func(idx *navigator.Index, _ *Reporter) {
			for i, rule := range rules {
				if fn := rule.visitors.Custom; fn != nil {
					fn(idx, reporters[i])
				}
			}
		}
	}
	return combined
}

func anyVisitor(rules []Rule, check func(Visitors) bool) bool {
	for _, rule := range rules {
		if check(rule.visitors) {
			return true
		}
	}
	return false
}

func hasVisitors(v Visitors) bool {
	return v.Document != nil ||
		v.Info != nil ||
		v.Path != nil ||
		v.Operation != nil ||
		v.Schema != nil ||
		v.RecursiveSchema != nil ||
		v.Parameter != nil ||
		v.Response != nil ||
		v.Tag != nil ||
		v.Server != nil ||
		v.RequestBody != nil ||
		v.SecurityScheme != nil ||
		v.Example != nil ||
		v.Custom != nil
}
