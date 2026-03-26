package barrelman

import (
	navigator "github.com/sailpoint-oss/navigator"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

// CrossRefResolver can resolve $ref values across files within a project.
type CrossRefResolver interface {
	CanResolve(fromURI, ref string) bool
}

// AnalysisContext is the gossip-free execution environment for every rule.
// It replaces treesitter.AnalysisContext as the universal context type.
type AnalysisContext struct {
	Index         *navigator.Index      // parsed OpenAPI document (nil for non-OpenAPI files)
	Tree          *tree_sitter.Tree     // tree-sitter CST (nil for content-only analysis)
	Language      *tree_sitter.Language // paired with Tree (YAML or JSON)
	Content       []byte                // raw source bytes
	URI           string                // file URI
	WorkspaceRoot string                // absolute workspace root, if known
	Resolver      CrossRefResolver      // cross-file $ref resolution (nil for single-file)
	TargetVersion navigator.Version     // from config; empty = auto-detect
}
