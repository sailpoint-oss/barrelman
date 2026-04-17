package fixes

import (
	"github.com/sailpoint-oss/barrelman"
	"github.com/sailpoint-oss/barrelman/codemod"
)

// ParameterExample implements the Fix for sailpoint-parameter-example:
// inserts `example: TODO` (or a synthesized value) into the
// parameter's block_mapping. Idempotent: emits no patches when the
// parameter already has an `example` or `examples` key.
func ParameterExample(ctx *codemod.FixContext, diag barrelman.Diagnostic) ([]codemod.Patch, error) {
	mapping := mappingForDiagnostic(ctx, diag)
	return buildInsertExamplePatch(mapping, ctx, diag), nil
}

// PropertyExample implements the Fix for sailpoint-property-example:
// inserts `example: TODO` into a schema property's block_mapping.
func PropertyExample(ctx *codemod.FixContext, diag barrelman.Diagnostic) ([]codemod.Patch, error) {
	mapping := mappingForDiagnostic(ctx, diag)
	return buildInsertExamplePatch(mapping, ctx, diag), nil
}

// ResponseExample implements the Fix for sailpoint-response-example:
// inserts `example: TODO` into a media-type's block_mapping. The rule
// emits its diagnostic at the response's location rather than the
// media-type's, so this fix walks down to find the specific media
// type's block_mapping named in the diagnostic message (when
// available) or the first media-type block_mapping it encounters.
func ResponseExample(ctx *codemod.FixContext, diag barrelman.Diagnostic) ([]codemod.Patch, error) {
	mapping := mappingForDiagnostic(ctx, diag)
	return buildInsertExamplePatch(mapping, ctx, diag), nil
}
