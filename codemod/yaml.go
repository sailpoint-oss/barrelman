package codemod

import (
	"bytes"
	"fmt"

	ts "github.com/tree-sitter/go-tree-sitter"
)

// QuoteStyle describes how a YAML scalar is quoted in the source. Used
// by fix authors who want their generated values to match the local
// style (plain text, double-quoted, single-quoted).
type QuoteStyle int

const (
	// QuoteStylePlain is an unquoted scalar (most common for simple values).
	QuoteStylePlain QuoteStyle = iota
	// QuoteStyleDouble is a "..."-quoted scalar.
	QuoteStyleDouble
	// QuoteStyleSingle is a '...'-quoted scalar.
	QuoteStyleSingle
)

// QuoteStyleOf inspects the YAML scalar node (or a node containing
// one) and returns the detected quoting. Returns QuoteStylePlain for
// unquoted scalars and for any node the function cannot classify.
func QuoteStyleOf(node *ts.Node, _ []byte) QuoteStyle {
	if node == nil {
		return QuoteStylePlain
	}
	switch node.Kind() {
	case "double_quote_scalar":
		return QuoteStyleDouble
	case "single_quote_scalar":
		return QuoteStyleSingle
	}
	return QuoteStylePlain
}

// IndentOf returns the number of spaces preceding the first byte of
// the node on its line. For a node whose first byte is at the start
// of a line, returns 0. The function walks backward from the node's
// start byte to the previous newline (or BOF) and counts spaces; any
// non-space byte encountered before the newline produces a best-effort
// result (number of leading spaces on that line).
func IndentOf(node *ts.Node, src []byte) int {
	if node == nil {
		return 0
	}
	start := int(node.StartByte())
	// Walk backward to the start of the line.
	lineStart := start
	for lineStart > 0 && src[lineStart-1] != '\n' {
		lineStart--
	}
	// Count leading spaces.
	indent := 0
	for i := lineStart; i < start; i++ {
		if src[i] != ' ' {
			break
		}
		indent++
	}
	return indent
}

// ChildIndent estimates the indent width a child of the given block
// mapping or block sequence should use. It prefers the indent of a
// later child over the first because the first child of a mapping
// that sits inside a block_sequence is preceded by `- `, which
// throws off a raw column measurement. When every child shares the
// same indent (the normal case), the returned value is unchanged.
func ChildIndent(parent *ts.Node, src []byte) int {
	if parent == nil {
		return 0
	}
	// Prefer a child other than the first: skips past the `- ` prefix
	// case. When no later child exists, adjust the first child's
	// indent by scanning forward past any leading dash-space sequence.
	var first *ts.Node
	for i := uint(0); i < parent.NamedChildCount(); i++ {
		child := parent.NamedChild(i)
		if child == nil {
			continue
		}
		if first == nil {
			first = child
			continue
		}
		return IndentOf(child, src)
	}
	if first != nil {
		return effectiveChildIndent(first, src)
	}
	return IndentOf(parent, src) + 2
}

// effectiveChildIndent returns the column of the first non-space,
// non-dash character on the child's line. Handles the `- key:` case
// by advancing past a leading `- ` once.
func effectiveChildIndent(child *ts.Node, src []byte) int {
	start := int(child.StartByte())
	lineStart := start
	for lineStart > 0 && src[lineStart-1] != '\n' {
		lineStart--
	}
	col := 0
	i := lineStart
	for i < start && src[i] == ' ' {
		col++
		i++
	}
	// If the next characters are `- ` (sequence item indicator),
	// count them and any following spaces so col reflects where the
	// key itself sits.
	if i < start && src[i] == '-' {
		col++
		i++
		for i < start && src[i] == ' ' {
			col++
			i++
		}
	}
	return col
}

// MappingHasKey reports whether a block_mapping contains a pair with
// the given key name. Used as an idempotency guard so Fix
// implementations emit zero patches when the fix has already been
// applied.
func MappingHasKey(mapping *ts.Node, src []byte, key string) bool {
	return FindMappingPair(mapping, src, key) != nil
}

// FindMappingPair returns the block_mapping_pair child whose key
// matches name, or nil. The search only inspects direct children (not
// grandchildren); callers can recurse themselves when needed.
func FindMappingPair(mapping *ts.Node, src []byte, key string) *ts.Node {
	if mapping == nil {
		return nil
	}
	for i := uint(0); i < mapping.NamedChildCount(); i++ {
		pair := mapping.NamedChild(i)
		if pair == nil || pair.Kind() != "block_mapping_pair" {
			continue
		}
		if keyOfPair(pair, src) == key {
			return pair
		}
	}
	return nil
}

// MappingValueNode returns the value node of the pair whose key
// matches name (or nil when no such pair exists).
func MappingValueNode(mapping *ts.Node, src []byte, key string) *ts.Node {
	pair := FindMappingPair(mapping, src, key)
	if pair == nil {
		return nil
	}
	return pair.ChildByFieldName("value")
}

// keyOfPair returns the textual key of a block_mapping_pair. Strips
// surrounding quotes for quoted scalar keys so callers can compare
// against plain identifiers.
func keyOfPair(pair *ts.Node, src []byte) string {
	keyNode := pair.ChildByFieldName("key")
	if keyNode == nil {
		return ""
	}
	text := string(keyNode.Utf8Text(src))
	// Strip surrounding quotes if present.
	if len(text) >= 2 {
		if (text[0] == '"' && text[len(text)-1] == '"') ||
			(text[0] == '\'' && text[len(text)-1] == '\'') {
			return text[1 : len(text)-1]
		}
	}
	return text
}

// InsertMappingPair emits a Patch that appends `key: value` as a new
// entry at the end of parent (a block_mapping). The indent is
// detected from parent's existing children so the inserted pair
// aligns with siblings.
//
// value is a YAML fragment inserted verbatim after `key:`. For
// multi-line values, the fragment must already be indented relative
// to the parent (use IndentFragment as a helper).
func InsertMappingPair(parent *ts.Node, src []byte, key, value string) Patch {
	indent := ChildIndent(parent, src)
	pad := bytes.Repeat([]byte(" "), indent)

	// Find the last non-whitespace byte the mapping spans. We insert
	// after that position, preceded by a newline so the new entry
	// starts on its own line.
	insertAt := endOfMappingContent(parent, src)

	fragment := fmt.Sprintf("\n%s%s: %s", pad, key, value)
	return Patch{
		StartByte:   uint(insertAt),
		EndByte:     uint(insertAt),
		Replacement: []byte(fragment),
		Description: fmt.Sprintf("insert %s: %s", key, truncate(value, 40)),
	}
}

// InsertMappingPairAfter emits a Patch that inserts a new pair
// immediately after the given sibling pair in the same mapping. Indent
// is taken from the sibling.
func InsertMappingPairAfter(sibling *ts.Node, src []byte, key, value string) Patch {
	if sibling == nil {
		return Patch{}
	}
	indent := IndentOf(sibling, src)
	pad := bytes.Repeat([]byte(" "), indent)

	// Insert immediately after the sibling's newline terminator. The
	// sibling's EndByte is the end of its value; we scan forward to
	// the next newline and insert just after it.
	insertAt := int(sibling.EndByte())
	for insertAt < len(src) && src[insertAt] != '\n' {
		insertAt++
	}
	if insertAt < len(src) && src[insertAt] == '\n' {
		insertAt++
	}

	fragment := fmt.Sprintf("%s%s: %s\n", pad, key, value)
	return Patch{
		StartByte:   uint(insertAt),
		EndByte:     uint(insertAt),
		Replacement: []byte(fragment),
		Description: fmt.Sprintf("insert %s: %s", key, truncate(value, 40)),
	}
}

// InsertSequenceItem emits a Patch that appends a new item to a
// block_sequence. item is a YAML fragment (commonly a plain scalar,
// but may be an inline flow collection).
func InsertSequenceItem(parent *ts.Node, src []byte, item string) Patch {
	indent := IndentOf(parent, src)
	pad := bytes.Repeat([]byte(" "), indent)

	insertAt := endOfMappingContent(parent, src)
	fragment := fmt.Sprintf("\n%s- %s", pad, item)
	return Patch{
		StartByte:   uint(insertAt),
		EndByte:     uint(insertAt),
		Replacement: []byte(fragment),
		Description: fmt.Sprintf("append sequence item: %s", truncate(item, 40)),
	}
}

// IndentFragment rewrites every line in fragment so its indentation
// is at least `indent` spaces. The first line is indented too; callers
// that want the first line to appear on an existing line should strip
// it themselves before calling.
func IndentFragment(fragment string, indent int) string {
	pad := string(bytes.Repeat([]byte(" "), indent))
	var out bytes.Buffer
	start := 0
	for i := 0; i <= len(fragment); i++ {
		if i == len(fragment) || fragment[i] == '\n' {
			line := fragment[start:i]
			if len(line) > 0 {
				out.WriteString(pad)
				out.WriteString(line)
			}
			if i < len(fragment) {
				out.WriteByte('\n')
			}
			start = i + 1
		}
	}
	return out.String()
}

// NodeAtByteRange returns the smallest tree-sitter node whose byte
// range covers [start, end). Equivalent to
// root.DescendantForByteRange(start, end) but defensive against nil
// trees / roots, which simplifies Fix implementations that may be
// handed a partially-initialized FixContext.
func NodeAtByteRange(tree *ts.Tree, start, end uint) *ts.Node {
	if tree == nil {
		return nil
	}
	root := tree.RootNode()
	if root == nil {
		return nil
	}
	return root.DescendantForByteRange(start, end)
}

// EnclosingBlockMapping walks from node to find the nearest
// block_mapping: first checks node itself, then its descendants
// (following a block_mapping_pair -> value -> block_node -> block_mapping
// chain), then its ancestors. Returns nil when no block_mapping is
// reachable. Fix implementations use this to convert "the node my
// diagnostic refers to" into "the block_mapping I should insert into".
func EnclosingBlockMapping(node *ts.Node) *ts.Node {
	if node == nil {
		return nil
	}
	if node.Kind() == "block_mapping" {
		return node
	}
	if m := findDescendantBlockMapping(node); m != nil {
		return m
	}
	for cur := node.Parent(); cur != nil; cur = cur.Parent() {
		if cur.Kind() == "block_mapping" {
			return cur
		}
	}
	return nil
}

// findDescendantBlockMapping returns the first block_mapping found in
// a breadth-first traversal of a small set of structural parent kinds
// (block_node, block_mapping_pair, flow_node). Bounded to prevent
// recursion into unrelated subtrees.
func findDescendantBlockMapping(node *ts.Node) *ts.Node {
	if node == nil {
		return nil
	}
	queue := []*ts.Node{node}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if cur == nil {
			continue
		}
		if cur.Kind() == "block_mapping" && cur != node {
			return cur
		}
		if cur == node || cur.Kind() == "block_node" || cur.Kind() == "block_mapping_pair" || cur.Kind() == "flow_node" {
			for i := uint(0); i < cur.NamedChildCount(); i++ {
				queue = append(queue, cur.NamedChild(i))
			}
		}
	}
	return nil
}

// endOfMappingContent returns the byte offset at which new content
// should be inserted to appear as the mapping's last entry. It
// trims trailing whitespace / newlines so the inserted fragment sits
// immediately after the last real byte.
func endOfMappingContent(parent *ts.Node, src []byte) int {
	if parent == nil {
		return 0
	}
	end := int(parent.EndByte())
	// Trim trailing whitespace and newlines within the parent's span
	// so inserts land right after the last real byte.
	for end > int(parent.StartByte()) {
		c := src[end-1]
		if c == '\n' || c == ' ' || c == '\t' || c == '\r' {
			end--
			continue
		}
		break
	}
	return end
}

// truncate shortens a string for logging.
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
