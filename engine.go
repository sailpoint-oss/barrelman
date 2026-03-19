package barrelman

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"

	navigator "github.com/sailpoint-oss/navigator"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

// LintResult holds the diagnostics for a single file.
type LintResult struct {
	File        string
	URI         string
	Diagnostics []Diagnostic
}

// LintOptions configures a lint run.
type LintOptions struct {
	MinSeverity   Severity
	ConfigPath    string
	RulesetPath   string
	WorkspaceRoot string
	Include       []string
	Exclude       []string
	TargetVersion string
	Rules         []Rule // if non-nil, overrides DefaultRegistry
}

// LintFiles runs the full barrelman rule suite against the given files.
// Pure Go, no gossip, no plugins, no external LSPs.
func LintFiles(files []string, opts LintOptions) ([]LintResult, error) {
	allRules := opts.Rules
	if allRules == nil {
		allRules = DefaultRegistry.AllRules()
	}

	var results []LintResult
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			results = append(results, LintResult{
				File: file,
				URI:  fileToURI(file),
				Diagnostics: []Diagnostic{{
					Range:    FileStartRange,
					Severity: SeverityError,
					Source:   Source,
					Code:     "file-read-error",
					Message:  err.Error(),
				}},
			})
			continue
		}

		uri := fileToURI(file)
		idx := navigator.ParseContent(content, uri)

		var tree *tree_sitter.Tree
		var lang *tree_sitter.Language
		if idx != nil {
			tree = idx.Tree()
		}
		if tree == nil {
			format := navigator.DetectFormat(uri, content)
			parser := tree_sitter.NewParser()
			switch format {
			case navigator.FormatJSON:
				lang = navigator.JSONLanguage()
			default:
				lang = navigator.YAMLLanguage()
			}
			if err := parser.SetLanguage(lang); err == nil {
				tree = parser.Parse(content, nil)
			}
			parser.Close()
		} else {
			format := navigator.DetectFormat(uri, content)
			switch format {
			case navigator.FormatJSON:
				lang = navigator.JSONLanguage()
			default:
				lang = navigator.YAMLLanguage()
			}
		}

		targetVersion := navigator.Version(opts.TargetVersion)
		if targetVersion == "" && idx != nil {
			targetVersion = idx.Version
		}

		ctx := &AnalysisContext{
			Index:         idx,
			Tree:          tree,
			Language:      lang,
			Content:       content,
			URI:           uri,
			TargetVersion: targetVersion,
		}

		var diags []Diagnostic
		for _, rule := range allRules {
			diags = append(diags, rule.Run(ctx)...)
		}

		if opts.MinSeverity > 0 {
			diags = filterBySeverity(diags, opts.MinSeverity)
		}

		results = append(results, LintResult{
			File:        file,
			URI:         uri,
			Diagnostics: diags,
		})
	}
	return results, nil
}

// LintContent runs all rules against in-memory content.
func LintContent(uri string, content []byte, opts LintOptions) ([]Diagnostic, error) {
	allRules := opts.Rules
	if allRules == nil {
		allRules = DefaultRegistry.AllRules()
	}

	idx := navigator.ParseContent(content, uri)

	var tree *tree_sitter.Tree
	var lang *tree_sitter.Language
	if idx != nil {
		tree = idx.Tree()
	}
	if tree == nil {
		format := navigator.DetectFormat(uri, content)
		parser := tree_sitter.NewParser()
		switch format {
		case navigator.FormatJSON:
			lang = navigator.JSONLanguage()
		default:
			lang = navigator.YAMLLanguage()
		}
		if err := parser.SetLanguage(lang); err == nil {
			tree = parser.Parse(content, nil)
		}
		parser.Close()
	} else {
		format := navigator.DetectFormat(uri, content)
		switch format {
		case navigator.FormatJSON:
			lang = navigator.JSONLanguage()
		default:
			lang = navigator.YAMLLanguage()
		}
	}

	targetVersion := navigator.Version(opts.TargetVersion)
	if targetVersion == "" && idx != nil {
		targetVersion = idx.Version
	}

	ctx := &AnalysisContext{
		Index:         idx,
		Tree:          tree,
		Language:      lang,
		Content:       content,
		URI:           uri,
		TargetVersion: targetVersion,
	}

	var diags []Diagnostic
	for _, rule := range allRules {
		diags = append(diags, rule.Run(ctx)...)
	}

	if opts.MinSeverity > 0 {
		diags = filterBySeverity(diags, opts.MinSeverity)
	}

	return diags, nil
}

func filterBySeverity(diags []Diagnostic, minSev Severity) []Diagnostic {
	var filtered []Diagnostic
	for _, d := range diags {
		if d.Severity <= minSev {
			filtered = append(filtered, d)
		}
	}
	return filtered
}

func fileToURI(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	abs = filepath.ToSlash(abs)
	if !strings.HasPrefix(abs, "/") {
		abs = "/" + abs
	}
	return (&url.URL{Scheme: "file", Path: abs}).String()
}
