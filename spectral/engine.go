package spectral

import (
	"log/slog"
	"sync"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/sailpoint-oss/barrelman"
	"gopkg.in/yaml.v3"
)

// Engine evaluates Spectral custom rules against YAML/JSON documents.
type Engine struct {
	mu     sync.RWMutex
	rules  []Rule
	logger *slog.Logger
}

// NewEngine creates a Spectral rule engine with the given custom rules.
func NewEngine(rules []Rule, logger *slog.Logger) *Engine {
	return &Engine{rules: rules, logger: logger}
}

// SetRules replaces the engine's rule set. Safe for concurrent use.
func (e *Engine) SetRules(rules []Rule) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.rules = rules
}

// Rules returns a snapshot of the current rule set.
func (e *Engine) Rules() []Rule {
	e.mu.RLock()
	defer e.mu.RUnlock()
	out := make([]Rule, len(e.rules))
	copy(out, e.rules)
	return out
}

// Execute runs all custom rules against the given document content and returns
// diagnostics with source positions derived from the yaml.Node tree.
func (e *Engine) Execute(content []byte) []barrelman.Diagnostic {
	e.mu.RLock()
	rules := make([]Rule, len(e.rules))
	copy(rules, e.rules)
	e.mu.RUnlock()

	if len(rules) == 0 {
		return nil
	}

	root, err := parseYAML(content)
	if err != nil {
		if e.logger != nil {
			e.logger.Warn("spectral engine: failed to parse document", "error", err)
		}
		return nil
	}

	var diags []barrelman.Diagnostic
	for _, rule := range rules {
		diags = append(diags, e.evaluateRule(rule, root)...)
	}
	return diags
}

func (e *Engine) evaluateRule(rule Rule, root *yaml.Node) []barrelman.Diagnostic {
	var diags []barrelman.Diagnostic

	for _, givenExpr := range rule.Given {
		matches, err := EvaluateJSONPath(root, givenExpr)
		if err != nil {
			if e.logger != nil {
				e.logger.Warn("spectral: JSONPath evaluation failed",
					"rule", rule.ID, "given", givenExpr, "error", err)
			}
			continue
		}

		for _, matchNode := range matches {
			for _, then := range rule.Then {
				fn, ok := BuiltinFunctions[then.Function]
				if !ok {
					if e.logger != nil {
						e.logger.Warn("spectral: unknown function",
							"rule", rule.ID, "function", then.Function)
					}
					continue
				}

				funcOpts := then.FunctionOptions
				if then.Function == "unreferencedReusableObject" {
					funcOpts = make(map[string]interface{})
					for k, v := range then.FunctionOptions {
						funcOpts[k] = v
					}
					funcOpts["__root__"] = root
				}

				issues := fn(matchNode, then.Field, funcOpts)
				for _, iss := range issues {
					reportNode := iss.Node
					if reportNode == nil {
						reportNode = matchNode
					}

					msg := iss.Message
					if rule.Message != "" {
						expanded := ExpandMessage(rule.Message, map[string]string{
							"error":    iss.Message,
							"property": then.Field,
							"path":     givenExpr,
							"value":    scalarValueString(reportNode),
						})
						if expanded != "" {
							msg = expanded
						}
					}
					if msg == "" {
						msg = rule.Description
					}

					diags = append(diags, barrelman.Diagnostic{
						Range:    nodeToRange(reportNode),
						Severity: barrelman.Severity(rule.Severity),
						Source:   "spectral",
						Code:     rule.ID,
						Message:  msg,
					})
				}
			}
		}
	}

	return diags
}

func parseYAML(content []byte) (*yaml.Node, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(content, &doc); err != nil {
		return nil, err
	}
	return &doc, nil
}

func nodeToRange(node *yaml.Node) barrelman.Range {
	if node == nil {
		return barrelman.Range{}
	}
	// yaml.Node uses 1-based lines and 1-based columns; we use 0-based.
	line := node.Line - 1
	col := node.Column - 1
	if line < 0 {
		line = 0
	}
	if col < 0 {
		col = 0
	}

	endCol := col
	if node.Kind == yaml.ScalarNode {
		endCol = col + spectralUTF16Len(node.Value)
	}

	return barrelman.Range{
		Start: barrelman.Position{Line: uint32(line), Character: uint32(col)},    //nolint:gosec
		End:   barrelman.Position{Line: uint32(line), Character: uint32(endCol)}, //nolint:gosec
	}
}

func spectralUTF16Len(s string) int {
	n := 0
	for len(s) > 0 {
		r, size := utf8.DecodeRuneInString(s)
		s = s[size:]
		n += utf16.RuneLen(r)
	}
	return n
}

func scalarValueString(node *yaml.Node) string {
	if node != nil && node.Kind == yaml.ScalarNode {
		return node.Value
	}
	return ""
}
