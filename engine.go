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

	content []byte
	index   *navigator.Index
}

// LintOptions configures a lint run.
type LintOptions struct {
	MinSeverity   Severity
	ConfigPath    string // optional Barrelman config file; legacy .telescope.* configs also parse
	RulesetPath   string // optional Barrelman ruleset file merged on top of config/built-ins
	WorkspaceRoot string
	Include       []string
	Exclude       []string
	TargetVersion string
	Rules         []Rule // if non-nil, supplies the executable rule set before config/ruleset overrides apply
}

// LintFiles runs the full barrelman rule suite against the given files.
// Pure Go, no gossip, no plugins, no external LSPs.
func LintFiles(files []string, opts LintOptions) ([]LintResult, error) {
	workspaceRoot := resolveLintWorkspaceRoot(files, opts.WorkspaceRoot)
	allRules, severityOverrides, err := resolveLintRules(opts, workspaceRoot)
	if err != nil {
		return nil, err
	}
	workspace := navigator.NewWorkspace()

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
		var resolver CrossRefResolver
		if idx != nil {
			project := navigator.OpenProjectFromIndex(uri, idx,
				navigator.WithCache(workspace.Cache),
				navigator.WithGraph(workspace.Graph),
				navigator.WithDiscovery(workspace.Discovery),
			)
			resolver = project.Resolver()
		}

		var lang = languageForContent(uri, content)
		var tree *tree_sitter.Tree
		if idx != nil {
			tree = idx.Tree()
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
			WorkspaceRoot: workspaceRoot,
			Resolver:      resolver,
			TargetVersion: targetVersion,
		}

		diags := runRules(ctx, allRules, severityOverrides)

		if opts.MinSeverity > 0 {
			diags = filterBySeverity(diags, opts.MinSeverity)
		}

		results = append(results, LintResult{
			File:        file,
			URI:         uri,
			Diagnostics: diags,
			content:     content,
			index:       idx,
		})
	}
	return results, nil
}

// LintContent runs all rules against in-memory content.
func LintContent(uri string, content []byte, opts LintOptions) ([]Diagnostic, error) {
	workspaceRoot := resolveLintWorkspaceRoot(nil, opts.WorkspaceRoot)
	allRules, severityOverrides, err := resolveLintRules(opts, workspaceRoot)
	if err != nil {
		return nil, err
	}

	idx := navigator.ParseContent(content, uri)

	var tree *tree_sitter.Tree
	var lang = languageForContent(uri, content)
	if idx != nil {
		tree = idx.Tree()
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
		WorkspaceRoot: workspaceRoot,
		TargetVersion: targetVersion,
	}

	diags := runRules(ctx, allRules, severityOverrides)

	if opts.MinSeverity > 0 {
		diags = filterBySeverity(diags, opts.MinSeverity)
	}

	return diags, nil
}

func languageForContent(uri string, content []byte) *tree_sitter.Language {
	switch navigator.DetectFormat(uri, content) {
	case navigator.FormatJSON:
		return navigator.JSONLanguage()
	default:
		return navigator.YAMLLanguage()
	}
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
