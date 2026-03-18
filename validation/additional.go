// Package validation provides additional file validation capabilities beyond
// the built-in OpenAPI structural validation. It allows applying JSON Schema
// validation to arbitrary YAML/JSON files via pattern-based matching, and
// provides schema validation with error enrichment for OpenAPI documents.
package validation

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
)

// ValidationGroup defines a set of file patterns and associated schemas
// for validating non-OpenAPI files.
type ValidationGroup struct {
	Patterns []string               `yaml:"patterns" json:"patterns"`
	Schemas  []SchemaPatternMapping `yaml:"schemas,omitempty" json:"schemas,omitempty"`
}

// SchemaPatternMapping pairs a JSON Schema file with optional pattern overrides.
type SchemaPatternMapping struct {
	Schema   string   `yaml:"schema" json:"schema"`
	Patterns []string `yaml:"patterns,omitempty" json:"patterns,omitempty"`
}

// AdditionalValidator validates files against schemas based on pattern matching.
type AdditionalValidator struct {
	mu         sync.RWMutex
	groups     map[string]ValidationGroup
	rootDir    string
	schemasDir string
}

// NewAdditionalValidator creates a new validator.
func NewAdditionalValidator() *AdditionalValidator {
	return &AdditionalValidator{
		groups: make(map[string]ValidationGroup),
	}
}

// Configure sets up validation groups and their schema directory.
func (v *AdditionalValidator) Configure(rootDir string, groups map[string]ValidationGroup) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.rootDir = rootDir
	v.schemasDir = filepath.Join(rootDir, ".telescope", "schemas")
	v.groups = groups
}

// MatchesFilePatterns returns whether the given file path matches any additional
// validation group.
func (v *AdditionalValidator) MatchesFilePatterns(filePath string) (groupName string, matched bool) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	relPath := uriToRelPath(filePath, v.rootDir)
	if relPath == "" {
		return "", false
	}

	for name, group := range v.groups {
		if matchesPatterns(relPath, group.Patterns) {
			return name, true
		}
	}
	return "", false
}

// SchemaDir returns the configured schemas directory.
func (v *AdditionalValidator) SchemaDir() string {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.schemasDir
}

// Groups returns the configured validation groups.
func (v *AdditionalValidator) Groups() map[string]ValidationGroup {
	v.mu.RLock()
	defer v.mu.RUnlock()
	result := make(map[string]ValidationGroup, len(v.groups))
	for k, g := range v.groups {
		result[k] = g
	}
	return result
}

func matchesPatterns(path string, patterns []string) bool {
	for _, pattern := range patterns {
		matched, err := filepath.Match(pattern, path)
		if err == nil && matched {
			return true
		}
		if matched, err := doubleStarMatch(pattern, path); err == nil && matched {
			return true
		}
	}
	return false
}

// doubleStarMatch handles ** glob patterns by expanding them to match any path segment.
func doubleStarMatch(pattern, path string) (bool, error) {
	if !strings.Contains(pattern, "**") {
		return filepath.Match(pattern, path)
	}
	parts := strings.SplitN(pattern, "**", 2)
	prefix := parts[0]
	suffix := strings.TrimPrefix(parts[1], "/")
	if prefix != "" {
		if !strings.HasPrefix(path, prefix) {
			return false, nil
		}
		path = path[len(prefix):]
	}
	if suffix == "" {
		return true, nil
	}
	for i := 0; i <= len(path); i++ {
		if matched, _ := filepath.Match(suffix, path[i:]); matched {
			return true, nil
		}
	}
	return false, nil
}

func uriToRelPath(uri, rootDir string) string {
	path := uri
	if len(path) > 7 && path[:7] == "file://" {
		path = path[7:]
	}
	rel, err := filepath.Rel(rootDir, path)
	if err != nil {
		return ""
	}
	return filepath.ToSlash(rel)
}

// EnrichAdditionalDiagnostics adds group/schema context to diagnostic messages.
func EnrichAdditionalDiagnostics(messages []string, groupName, schemaFile string) []string {
	if len(messages) == 0 {
		return messages
	}
	enriched := make([]string, 0, len(messages))
	prefix := fmt.Sprintf("[schema:%s group:%s] ", schemaFile, groupName)
	for _, msg := range messages {
		if msg != "" {
			enriched = append(enriched, prefix+msg)
		} else {
			enriched = append(enriched, prefix+"Schema validation failed")
		}
	}
	return enriched
}
