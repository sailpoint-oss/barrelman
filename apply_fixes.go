package barrelman

import (
	"fmt"
	"os"

	"github.com/sailpoint-oss/barrelman/codemod"
	navigator "github.com/sailpoint-oss/navigator"
)

// FixResult captures the outcome of running auto-fixes against a
// single file. Patched is the new source bytes after every fix in
// Patches has been applied; Diagnostics is the original lint result
// (useful for CLI/LSP reporting).
type FixResult struct {
	File        string
	URI         string
	Original    []byte
	Patched     []byte
	Patches     []codemod.Patch
	Diagnostics []Diagnostic
	// Unfixable holds diagnostics whose rule has no Fix attached.
	// Callers that run with --fail-on-unfixable raise an error when
	// this list is non-empty.
	Unfixable []Diagnostic
}

// Changed reports whether applying the patches modified the source.
func (r FixResult) Changed() bool {
	return len(r.Patches) > 0 && string(r.Original) != string(r.Patched)
}

// FixOptions controls auto-fix application.
type FixOptions struct {
	// Lint passes through to barrelman's lint engine for rule
	// selection, ruleset resolution, severity filtering, and so on.
	Lint LintOptions

	// Rules filters fixes by rule ID. When nil every rule with an
	// attached Fix is considered.
	Rules []string

	// Hints is an optional source-aware hint provider (cartographer,
	// schema-synth). When nil the sentinels TODO are used.
	Hints codemod.SourceHintProvider

	// Waivers, when non-nil, filters out patches whose rule is
	// actively waived. Pass the result of codemod.LoadWaivers here.
	Waivers *codemod.WaiverSet
}

// ApplyFixes runs barrelman's lint engine against files, invokes the
// Fix callback of each rule that has one, and returns the patched
// source per file. The files on disk are NOT modified; callers (CLI
// `telescope fix --write`, LSP code actions, tests) decide whether
// to persist the result.
func ApplyFixes(files []string, opts FixOptions) ([]FixResult, error) {
	lintResults, err := LintFiles(files, opts.Lint)
	if err != nil {
		return nil, err
	}
	ruleByID := indexRulesByID(opts.Lint.Rules)
	if len(ruleByID) == 0 {
		ruleByID = indexRulesByID(DefaultRegistry.AllRules())
	}
	whitelist := setOfStrings(opts.Rules)

	out := make([]FixResult, 0, len(lintResults))
	for _, lr := range lintResults {
		content := lr.content
		if content == nil {
			var err error
			content, err = os.ReadFile(lr.File)
			if err != nil {
				return nil, fmt.Errorf("read %s: %w", lr.File, err)
			}
		}
		idx := lr.index
		if idx == nil {
			idx = navigator.ParseContent(content, lr.URI)
		}
		fixCtx := &codemod.FixContext{
			Index:  idx,
			Source: content,
			URI:    lr.URI,
			Hints:  opts.Hints,
		}
		var patches []codemod.Patch
		var unfixable []Diagnostic
		for _, diag := range lr.Diagnostics {
			if len(whitelist) > 0 && !whitelist[diag.Code] {
				continue
			}
			rule, ok := ruleByID[diag.Code]
			if !ok || rule.Fix == nil {
				unfixable = append(unfixable, diag)
				continue
			}
			ps, ferr := rule.Fix(fixCtx, diag)
			if ferr != nil {
				return nil, fmt.Errorf("rule %s fix: %w", rule.ID, ferr)
			}
			for i := range ps {
				if ps[i].URI == "" {
					ps[i].URI = lr.URI
				}
				if ps[i].RuleID == "" {
					ps[i].RuleID = rule.ID
				}
			}
			patches = append(patches, ps...)
		}
		if opts.Waivers != nil {
			patches = opts.Waivers.Filter(patches)
		}
		driver := &codemod.Driver{}
		patched, applyErr := driver.Apply(content, patches)
		if applyErr != nil {
			return nil, fmt.Errorf("apply fixes to %s: %w", lr.File, applyErr)
		}
		out = append(out, FixResult{
			File:        lr.File,
			URI:         lr.URI,
			Original:    content,
			Patched:     patched,
			Patches:     patches,
			Diagnostics: lr.Diagnostics,
			Unfixable:   unfixable,
		})
	}
	return out, nil
}

func indexRulesByID(rules []Rule) map[string]Rule {
	out := make(map[string]Rule, len(rules))
	for _, r := range rules {
		out[r.ID] = r
	}
	return out
}

func setOfStrings(xs []string) map[string]bool {
	if len(xs) == 0 {
		return nil
	}
	m := make(map[string]bool, len(xs))
	for _, x := range xs {
		m[x] = true
	}
	return m
}
