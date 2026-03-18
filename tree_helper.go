package barrelman

import (
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

// Capture represents a single tree-sitter query capture (pattern match).
type Capture struct {
	Name string
	Node tree_sitter.Node
}

// TreeHelper provides gossip-free tree-sitter utilities for syntax checks.
type TreeHelper struct {
	tree    *tree_sitter.Tree
	content []byte
}

// NewTreeHelper creates a TreeHelper from a tree-sitter tree and its source.
func NewTreeHelper(tree *tree_sitter.Tree, content []byte) *TreeHelper {
	return &TreeHelper{tree: tree, content: content}
}

// RootNode returns the root node of the tree.
func (h *TreeHelper) RootNode() *tree_sitter.Node {
	if h.tree == nil {
		return nil
	}
	return h.tree.RootNode()
}

// NodeText returns the source text of a node.
func (h *TreeHelper) NodeText(node *tree_sitter.Node) string {
	if node == nil || h.content == nil {
		return ""
	}
	start := node.StartByte()
	end := node.EndByte()
	if int(end) > len(h.content) {
		end = uint(len(h.content))
	}
	return string(h.content[start:end])
}

// NodeRange converts a tree-sitter node's range to a barrelman Range.
func (h *TreeHelper) NodeRange(node *tree_sitter.Node) Range {
	if node == nil {
		return Range{}
	}
	sp := node.StartPosition()
	ep := node.EndPosition()
	return Range{
		Start: Position{Line: uint32(sp.Row), Character: uint32(sp.Column)},
		End:   Position{Line: uint32(ep.Row), Character: uint32(ep.Column)},
	}
}

// QueryCaptures runs a tree-sitter query pattern and returns all captures.
func (h *TreeHelper) QueryCaptures(lang *tree_sitter.Language, pattern string) ([]Capture, error) {
	q, err := tree_sitter.NewQuery(lang, pattern)
	if err != nil {
		return nil, err
	}
	defer q.Close()

	root := h.RootNode()
	if root == nil {
		return nil, nil
	}

	cursor := tree_sitter.NewQueryCursor()
	defer cursor.Close()

	matches := cursor.Matches(q, root, h.content)
	var captures []Capture
	for {
		m := matches.Next()
		if m == nil {
			break
		}
		for _, c := range m.Captures {
			captures = append(captures, Capture{
				Name: q.CaptureNames()[c.Index],
				Node: c.Node,
			})
		}
	}
	return captures, nil
}
