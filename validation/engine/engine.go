package engine

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
	jsonschemakind "github.com/santhosh-tekuri/jsonschema/v6/kind"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var _ = jsonschemakind.AllOf{} // keep the kind package referenced at package level.

// messagePrinter is the go x/text printer used for every santhosh
// LocalizedString call. Santhosh panics if the printer is nil, so we
// never pass a nil value.
var messagePrinter = message.NewPrinter(language.English)

// Schema is a compiled schema ready to validate arbitrary instances.
// Compilation is expensive; callers are expected to cache Schema values
// keyed by the source document pointer (navigator.Index does this for
// OpenAPI documents).
type Schema struct {
	compiled *jsonschema.Schema
	compiler *jsonschema.Compiler
	// url is the identifier the compiler registered the schema under;
	// preserved for error messages and for ValidateBytes.
	url string
	// openapi tracks whether this schema participates in OpenAPI
	// semantic handling (nullable rewrite, discriminator, format lax mode).
	openapi OpenAPIOptions
}

// CompileOpts controls Compile behaviour.
type CompileOpts struct {
	// URL is the identity the compiler uses for $ref resolution. When
	// empty a stable placeholder is synthesized so the compile still
	// succeeds for self-contained schemas.
	URL string

	// OpenAPI enables OpenAPI-specific rewrites and format handling. Use
	// this when compiling schemas that came from an OpenAPI document
	// (components.schemas, parameter schemas, media-type schemas).
	OpenAPI OpenAPIOptions

	// Resolver, when non-nil, is called by the santhosh compiler to
	// resolve external $ref URLs. The canonical navigator-backed
	// resolver lives in validation/engine/openapi; this field is an
	// escape hatch for tests and non-OpenAPI callers.
	Resolver jsonschema.URLLoader
}

// Compile parses a schema document and returns a reusable *Schema.
//
// The schema argument may be raw bytes, a string, a map[string]any, or
// any value that encoding/json can round-trip through.
func Compile(schema any, opts CompileOpts) (*Schema, error) {
	loader, err := normalizeSchemaForCompile(schema)
	if err != nil {
		return nil, fmt.Errorf("engine: normalize schema: %w", err)
	}
	if opts.OpenAPI.NullableRewrite {
		loader = rewriteOpenAPINullable(loader)
	}
	url := opts.URL
	if url == "" {
		url = "mem://schema"
	}

	compiler := jsonschema.NewCompiler()
	if opts.Resolver != nil {
		compiler.UseLoader(opts.Resolver)
	}
	if err := compiler.AddResource(url, loader); err != nil {
		return nil, fmt.Errorf("engine: add resource %s: %w", url, err)
	}
	compiled, err := compiler.Compile(url)
	if err != nil {
		return nil, fmt.Errorf("engine: compile %s: %w", url, err)
	}
	return &Schema{compiled: compiled, compiler: compiler, url: url, openapi: opts.OpenAPI}, nil
}

// Validate validates a Go-level instance (map[string]any, []any, strings,
// numbers, bool, nil) against the compiled schema. Instances are the
// representation produced by `json.Unmarshal(data, &iface)` — the most
// common shape for both JSON and YAML pipelines.
//
// The returned slice is empty when the instance conforms; otherwise it
// contains one or more Issues ordered by the compiler's traversal order
// (stable within a single call).
func (s *Schema) Validate(instance any) []Issue {
	if s == nil || s.compiled == nil {
		return nil
	}
	err := s.compiled.Validate(instance)
	if err == nil {
		return nil
	}
	var ve *jsonschema.ValidationError
	if !errors.As(err, &ve) {
		return []Issue{{
			Code:     "validate",
			Severity: SeverityError,
			Source:   SourceEngine,
			Message:  err.Error(),
		}}
	}
	return collectIssues(ve, instance)
}

// ValidateBytes unmarshals raw JSON bytes and validates the result. Use
// Validate when the caller already holds a decoded instance.
func (s *Schema) ValidateBytes(raw []byte) ([]Issue, error) {
	var instance any
	if err := json.Unmarshal(raw, &instance); err != nil {
		return nil, fmt.Errorf("engine: decode instance: %w", err)
	}
	return s.Validate(instance), nil
}

// URL returns the compiler identity the schema was registered under.
func (s *Schema) URL() string {
	if s == nil {
		return ""
	}
	return s.url
}

// collectIssues turns a santhosh validation error tree into flat Issues.
// Composition keyword failures are flattened in document order so
// downstream formatters can present them as "oneOf variant N failed:
// ...".
func collectIssues(root *jsonschema.ValidationError, instance any) []Issue {
	var out []Issue
	walk(root, instance, &out)
	return out
}

func walk(ve *jsonschema.ValidationError, instance any, out *[]Issue) {
	issue := issueFromValidationError(ve, instance)
	if issue.Message != "" {
		*out = append(*out, issue)
	}
	for _, cause := range ve.Causes {
		walk(cause, instance, out)
	}
}

// issueFromValidationError converts one santhosh ValidationError node
// into an Issue. It picks a stable Code, formats a friendly message, and
// derives Expected / Received snapshots when the error kind provides
// enough context.
func issueFromValidationError(ve *jsonschema.ValidationError, instance any) Issue {
	if ve == nil {
		return Issue{}
	}
	ptr := instancePointer(ve)
	issue := Issue{
		Code:     errorCode(ve),
		Severity: SeverityError,
		Source:   SourceEngine,
		Pointer:  ptr,
		Path:     humanPath(ptr),
	}
	issue.Message = ve.ErrorKind.LocalizedString(messagePrinter)
	if issue.Message == "" {
		issue.Message = fmt.Sprintf("%v", ve.ErrorKind)
	}
	if expected, received, ok := expectedReceived(ve, instance); ok {
		issue.Expected = expected
		issue.Received = received
	}
	return issue
}

// errorCode derives a short, stable rule code from a santhosh error kind.
// The KeywordPath slice is guaranteed non-empty by the santhosh contract;
// we prefer the first element (keyword name) and fall back to a generic
// label if the contract is violated.
func errorCode(ve *jsonschema.ValidationError) string {
	if ve == nil || ve.ErrorKind == nil {
		return "validate"
	}
	path := ve.ErrorKind.KeywordPath()
	if len(path) == 0 {
		return "validate"
	}
	return path[0]
}

// instancePointer renders santhosh's InstanceLocation as an RFC 6901
// pointer. santhosh returns the instance path as a slice of segments; we
// escape and join them here so downstream code can treat the pointer as
// opaque.
func instancePointer(ve *jsonschema.ValidationError) string {
	if len(ve.InstanceLocation) == 0 {
		return ""
	}
	var b strings.Builder
	for _, seg := range ve.InstanceLocation {
		b.WriteByte('/')
		esc := strings.ReplaceAll(seg, "~", "~0")
		esc = strings.ReplaceAll(esc, "/", "~1")
		b.WriteString(esc)
	}
	return b.String()
}

// humanPath mirrors Issue.HumanPath for use at issue-construction time.
func humanPath(ptr string) string {
	if ptr == "" {
		return ""
	}
	segs := strings.Split(strings.TrimPrefix(ptr, "/"), "/")
	for i, seg := range segs {
		seg = strings.ReplaceAll(seg, "~1", "/")
		seg = strings.ReplaceAll(seg, "~0", "~")
		segs[i] = seg
	}
	return strings.Join(segs, ".")
}

// normalizeSchemaForCompile accepts the flexible schema input types and
// returns a value the santhosh compiler can consume.
func normalizeSchemaForCompile(schema any) (any, error) {
	switch v := schema.(type) {
	case []byte:
		if len(bytes.TrimSpace(v)) == 0 {
			return nil, fmt.Errorf("empty schema bytes")
		}
		var out any
		if err := json.Unmarshal(v, &out); err != nil {
			return nil, fmt.Errorf("unmarshal schema: %w", err)
		}
		return out, nil
	case string:
		var out any
		if err := json.Unmarshal([]byte(v), &out); err != nil {
			return nil, fmt.Errorf("unmarshal schema: %w", err)
		}
		return out, nil
	case map[string]any, []any, bool:
		return v, nil
	default:
		// Fall through to a generic JSON round-trip so callers can pass
		// typed Go structs that marshal to JSON.
		raw, err := json.Marshal(schema)
		if err != nil {
			return nil, fmt.Errorf("marshal schema: %w", err)
		}
		var out any
		if err := json.Unmarshal(raw, &out); err != nil {
			return nil, fmt.Errorf("unmarshal schema: %w", err)
		}
		return out, nil
	}
}
