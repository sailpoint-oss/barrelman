package spectral

import (
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/sailpoint-oss/barrelman/validation/engine"
	"github.com/sailpoint-oss/barrelman/validation/engine/openapi"
)

// validateSchemaViaEngine validates a YAML node against a schema using
// the shared validation engine. The second return value is false when
// the node could not be decoded into an engine instance (so the caller
// can fall back to the legacy validator without emitting a bogus
// diagnostic).
func validateSchemaViaEngine(node *yaml.Node, schemaDef map[string]any) ([]Issue, bool) {
	if node == nil || schemaDef == nil {
		return nil, false
	}
	var instance any
	if err := node.Decode(&instance); err != nil {
		return nil, false
	}
	compiled, err := openapi.CompileSchema("mem://spectral-schema", schemaDef)
	if err != nil {
		return nil, false
	}
	result := compiled.Validate(instance)
	if len(result) == 0 {
		return nil, true
	}
	issues := make([]Issue, 0, len(result))
	for _, is := range result {
		issues = append(issues, Issue{Node: node, Message: messageForIssue(is)})
	}
	return issues, true
}

func messageForIssue(is engine.Issue) string {
	msg := is.Message
	if msg == "" {
		msg = is.Code
	}
	if is.Expected != "" && is.Received != "" {
		msg = fmt.Sprintf("%s (expected %s, received %s)", msg, is.Expected, is.Received)
	}
	return msg
}
