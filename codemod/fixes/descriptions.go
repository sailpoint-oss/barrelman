package fixes

import (
	"github.com/sailpoint-oss/barrelman"
	"github.com/sailpoint-oss/barrelman/codemod"
)

// ParameterDescription inserts `description: TODO` (or a source-aware
// description when one is available) into the parameter's block_mapping.
//
// Idempotent: emits no patches when the parameter already has a
// description key.
func ParameterDescription(ctx *codemod.FixContext, diag barrelman.Diagnostic) ([]codemod.Patch, error) {
	mapping := mappingForDiagnostic(ctx, diag)
	return buildInsertDescriptionPatch(mapping, ctx, diag), nil
}

// PropertyDescription inserts `description: TODO` into the schema property's
// block_mapping.
func PropertyDescription(ctx *codemod.FixContext, diag barrelman.Diagnostic) ([]codemod.Patch, error) {
	mapping := mappingForDiagnostic(ctx, diag)
	return buildInsertDescriptionPatch(mapping, ctx, diag), nil
}

// TagDocumented inserts `description: TODO` into the root-level tag entry's
// block_mapping. Mechanically identical to the parameter case but named
// separately so CLI/LSP telemetry attributes fixes to the right rule.
func TagDocumented(ctx *codemod.FixContext, diag barrelman.Diagnostic) ([]codemod.Patch, error) {
	mapping := mappingForDiagnostic(ctx, diag)
	return buildInsertDescriptionPatch(mapping, ctx, diag), nil
}
