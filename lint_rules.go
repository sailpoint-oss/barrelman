package barrelman

import (
	"fmt"
	"path/filepath"

	"github.com/sailpoint-oss/barrelman/config"
	"github.com/sailpoint-oss/barrelman/rulesets"
)

func resolveLintRules(opts LintOptions, workspaceRoot string) ([]Rule, map[string]rulesets.SeverityOverride, error) {
	allRules := opts.Rules
	if allRules == nil {
		allRules = DefaultRegistry.AllRules()
	}

	if opts.ConfigPath == "" && opts.RulesetPath == "" && opts.WorkspaceRoot == "" {
		return allRules, nil, nil
	}

	resolvedRuleset, err := loadResolvedRuleset(opts, workspaceRoot)
	if err != nil {
		return nil, nil, err
	}
	if resolvedRuleset == nil {
		return allRules, nil, nil
	}

	overrides := make(map[string]rulesets.SeverityOverride)
	for _, override := range rulesets.BuildSeverityOverrides(resolvedRuleset) {
		overrides[override.RuleID] = override
	}

	if opts.Rules != nil {
		filtered := make([]Rule, 0, len(allRules))
		for _, rule := range allRules {
			override, ok := overrides[rule.ID]
			if ok && override.Disabled {
				continue
			}
			filtered = append(filtered, rule)
		}
		return filtered, overrides, nil
	}

	enabled := rulesets.BuildEnabledMap(resolvedRuleset)
	filtered := make([]Rule, 0, len(allRules))
	for _, rule := range allRules {
		if enabled[rule.ID] {
			filtered = append(filtered, rule)
		}
	}
	return filtered, overrides, nil
}

func loadResolvedRuleset(opts LintOptions, workspaceRoot string) (*rulesets.RuleSet, error) {
	var (
		cfg           *config.Config
		err           error
		configBaseDir = workspaceRoot
	)
	switch {
	case opts.ConfigPath != "":
		cfg, err = config.LoadFile(opts.ConfigPath)
		configBaseDir = filepath.Dir(opts.ConfigPath)
	case workspaceRoot != "":
		cfg, err = config.Load(workspaceRoot)
	default:
		cfg = config.DefaultConfig()
	}
	if err != nil {
		return nil, fmt.Errorf("load barrelman config: %w", err)
	}

	baseRuleset := rulesets.GetBuiltin(cfg.Extends)
	if baseRuleset == nil {
		baseRuleset = &rulesets.RuleSet{Rules: make(map[string]rulesets.RuleDefinition)}
	}
	for id, sev := range cfg.Rules {
		baseRuleset.Rules[id] = rulesets.RuleDefinition{Severity: sev}
	}

	resolvedBase, err := rulesets.Resolve(baseRuleset, configBaseDir)
	if err != nil {
		return nil, fmt.Errorf("resolve built-in ruleset %q: %w", cfg.Extends, err)
	}
	if opts.RulesetPath == "" {
		return resolvedBase, nil
	}

	rs, err := rulesets.LoadFile(opts.RulesetPath)
	if err != nil {
		return nil, fmt.Errorf("load ruleset %s: %w", opts.RulesetPath, err)
	}
	resolvedExtra, err := rulesets.Resolve(rs, filepath.Dir(opts.RulesetPath))
	if err != nil {
		return nil, fmt.Errorf("resolve ruleset %s: %w", opts.RulesetPath, err)
	}
	return rulesets.Merge(resolvedBase, resolvedExtra), nil
}

func applyRuleSeverityOverride(diags []Diagnostic, ruleID string, overrides map[string]rulesets.SeverityOverride) []Diagnostic {
	if len(diags) == 0 || len(overrides) == 0 {
		return diags
	}
	override, ok := overrides[ruleID]
	if !ok {
		return diags
	}
	if override.Disabled {
		return nil
	}

	adjusted := make([]Diagnostic, 0, len(diags))
	for _, diag := range diags {
		diag.Severity = barrelmanSeverity(override.Severity)
		adjusted = append(adjusted, diag)
	}
	return adjusted
}

func barrelmanSeverity(sev rulesets.Severity) Severity {
	switch sev {
	case rulesets.SeverityError:
		return SeverityError
	case rulesets.SeverityWarning:
		return SeverityWarning
	case rulesets.SeverityInfo:
		return SeverityInfo
	case rulesets.SeverityHint:
		return SeverityHint
	default:
		return SeverityWarning
	}
}
