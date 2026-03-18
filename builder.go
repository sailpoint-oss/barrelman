package barrelman

import (
	navigator "github.com/LukasParke/navigator"
)

// RuleBuilder provides a fluent API for defining OpenAPI diagnostic rules.
// Use Define to start building, chain visitor methods, then call Build or
// Register to produce a Rule.
type RuleBuilder struct {
	id   string
	meta RuleMeta
	v    Visitors
}

// Define begins building a new rule with the given ID and metadata.
func Define(id string, meta RuleMeta) *RuleBuilder {
	meta.ID = id
	return &RuleBuilder{id: id, meta: meta}
}

// Document adds a visitor that receives the full document.
func (b *RuleBuilder) Document(fn func(doc *navigator.Document, r *Reporter)) *RuleBuilder {
	b.v.Document = fn
	return b
}

// Info adds a visitor that is called with the document's Info object.
func (b *RuleBuilder) Info(fn func(info *navigator.Info, r *Reporter)) *RuleBuilder {
	b.v.Info = fn
	return b
}

// Paths adds a visitor that is called for each path.
func (b *RuleBuilder) Paths(fn func(path string, item *navigator.PathItem, r *Reporter)) *RuleBuilder {
	b.v.Path = fn
	return b
}

// Operations adds a visitor that is called for each operation.
func (b *RuleBuilder) Operations(fn func(path string, method string, op *navigator.Operation, r *Reporter)) *RuleBuilder {
	b.v.Operation = fn
	return b
}

// Schemas adds a visitor for each schema (component and inline).
func (b *RuleBuilder) Schemas(fn func(name string, schema *navigator.Schema, pointer string, r *Reporter)) *RuleBuilder {
	b.v.Schema = fn
	return b
}

// RecursiveSchemas adds a visitor that recursively walks all schemas.
func (b *RuleBuilder) RecursiveSchemas(fn func(name string, schema *navigator.Schema, pointer string, r *Reporter)) *RuleBuilder {
	b.v.RecursiveSchema = fn
	return b
}

// Parameters adds a visitor for each parameter.
func (b *RuleBuilder) Parameters(fn func(param *navigator.Parameter, r *Reporter)) *RuleBuilder {
	b.v.Parameter = fn
	return b
}

// Responses adds a visitor for each response.
func (b *RuleBuilder) Responses(fn func(code string, resp *navigator.Response, r *Reporter)) *RuleBuilder {
	b.v.Response = fn
	return b
}

// Tags adds a visitor for each tag.
func (b *RuleBuilder) Tags(fn func(tag *navigator.Tag, r *Reporter)) *RuleBuilder {
	b.v.Tag = fn
	return b
}

// Servers adds a visitor for each server.
func (b *RuleBuilder) Servers(fn func(server *navigator.Server, r *Reporter)) *RuleBuilder {
	b.v.Server = fn
	return b
}

// RequestBodies adds a visitor for each request body.
func (b *RuleBuilder) RequestBodies(fn func(path string, method string, rb *navigator.RequestBody, r *Reporter)) *RuleBuilder {
	b.v.RequestBody = fn
	return b
}

// SecuritySchemes adds a visitor for each security scheme.
func (b *RuleBuilder) SecuritySchemes(fn func(name string, ss *navigator.SecurityScheme, r *Reporter)) *RuleBuilder {
	b.v.SecurityScheme = fn
	return b
}

// Examples adds a visitor for each component example.
func (b *RuleBuilder) Examples(fn func(name string, ex *navigator.Example, r *Reporter)) *RuleBuilder {
	b.v.Example = fn
	return b
}

// Custom adds a visitor that receives the full index for arbitrary logic.
func (b *RuleBuilder) Custom(fn func(idx *navigator.Index, r *Reporter)) *RuleBuilder {
	b.v.Custom = fn
	return b
}

// Meta returns the rule metadata.
func (b *RuleBuilder) Meta() RuleMeta {
	return b.meta
}

// Build returns a Rule ready for execution or registration.
func (b *RuleBuilder) Build() Rule {
	v := b.v
	meta := b.meta
	id := b.id
	return Rule{
		ID:   id,
		Meta: meta,
		Run: func(ctx *AnalysisContext) []Diagnostic {
			if ctx.Index == nil {
				return nil
			}
			r := NewReporter(id, meta.Severity)
			Walk(ctx.Index, v, r)
			return r.Diagnostics()
		},
	}
}

// Register builds the rule and adds it to the given registry.
func (b *RuleBuilder) Register(reg *Registry) Rule {
	rule := b.Build()
	reg.Register(rule)
	return rule
}
