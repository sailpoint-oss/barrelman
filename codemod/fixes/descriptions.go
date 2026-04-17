package fixes

import (
	"github.com/sailpoint-oss/barrelman"
	"github.com/sailpoint-oss/barrelman/codemod"
)

// ParameterDescription implements the Fix for
// sailpoint-parameter-description: inserts `description: TODO` (or a
// cartographer-sourced description when one is available) into the
// parameter's block_mapping.
//
// Idempotent: emits no patches when the parameter already has a
// description key.
func ParameterDescription(ctx *codemod.FixContext, diag barrelman.Diagnostic) ([]codemod.Patch, error) {
	mapping := mappingForDiagnostic(ctx, diag)
	return buildInsertDescriptionPatch(mapping, ctx, diag), nil
}

// PropertyDescription implements the Fix for
// sailpoint-property-description: inserts `description: TODO` into
// the schema property's block_mapping.
func PropertyDescription(ctx *codemod.FixContext, diag barrelman.Diagnostic) ([]codemod.Patch, error) {
	mapping := mappingForDiagnostic(ctx, diag)
	return buildInsertDescriptionPatch(mapping, ctx, diag), nil
}

// TagDocumented implements the Fix for sailpoint-tag-documented:
// inserts `description: TODO` into the root-level tag entry's
// block_mapping. Mechanically identical to the parameter case but
// named separately so CLI/LSP telemetry attributes fixes to the
// right rule.
func TagDocumented(ctx *codemod.FixContext, diag barrelman.Diagnostic) ([]codemod.Patch, error) {
	mapping := mappingForDiagnostic(ctx, diag)
	return buildInsertDescriptionPatch(mapping, ctx, diag), nil
}
