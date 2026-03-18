package checks

import (
	"fmt"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"

	"github.com/sailpoint-oss/barrelman"
)

var duplicateKeysMeta = barrelman.RuleMeta{
	ID:          "duplicate-keys",
	Description: "Reports duplicate mapping keys in YAML/JSON objects.",
	Severity:    barrelman.SeverityError,
	Category:    barrelman.CategorySyntax,
	Recommended: true,
	HowToFix:    "Remove or rename the duplicate key.",
	DocURL:      barrelman.DocBaseURL + "duplicate-keys",
}

func registerDuplicateKeys(reg *barrelman.Registry) {
	reg.Register(barrelman.Rule{
		ID:   "duplicate-keys",
		Meta: duplicateKeysMeta,
		Run: func(ctx *barrelman.AnalysisContext) []barrelman.Diagnostic {
			if ctx.Tree == nil {
				return nil
			}
			helper := barrelman.NewTreeHelper(ctx.Tree, ctx.Content)
			root := helper.RootNode()
			if root == nil {
				return nil
			}
			var diags []barrelman.Diagnostic
			walkForDuplicates(root, helper, &diags)
			return diags
		},
	})
}

func walkForDuplicates(node *tree_sitter.Node, helper *barrelman.TreeHelper, diags *[]barrelman.Diagnostic) {
	if node == nil {
		return
	}

	kind := node.Kind()
	if kind == "block_mapping" || kind == "flow_mapping" || kind == "object" {
		seen := make(map[string]barrelman.Range)
		for i := uint(0); i < node.ChildCount(); i++ {
			child := node.Child(i)
			if child == nil {
				continue
			}
			ck := child.Kind()
			if ck == "block_mapping_pair" || ck == "flow_pair" || ck == "pair" {
				keyNode := child.ChildByFieldName("key")
				if keyNode == nil {
					continue
				}
				keyText := unquoteKey(helper.NodeText(keyNode))
				keyRange := helper.NodeRange(keyNode)

				if firstRange, exists := seen[keyText]; exists {
					*diags = append(*diags, barrelman.Diagnostic{
						Range:           keyRange,
						Severity:        barrelman.SeverityError,
						Source:          barrelman.Source,
						Code:            "duplicate-keys",
						CodeDescription: duplicateKeysMeta.DocURL,
						Message: fmt.Sprintf(
							"Duplicate key '%s' (first defined at line %d)",
							keyText, firstRange.Start.Line+1,
						),
					})
				} else {
					seen[keyText] = keyRange
				}
			}
		}
	}

	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child != nil {
			walkForDuplicates(child, helper, diags)
		}
	}
}

func unquoteKey(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
