package fixes

import (
	"fmt"

	"github.com/sailpoint-oss/barrelman"
	"github.com/sailpoint-oss/barrelman/codemod"
)

// OperationSingleTag implements the Fix for
// sailpoint-operation-single-tag. When the operation has no `tags:`
// key, inserts `tags: [TODO]`. When the operation has an empty list
// or a multi-entry list, the fix declines to patch (the correct
// single tag cannot be determined deterministically); the diagnostic
// remains for human attention.
func OperationSingleTag(ctx *codemod.FixContext, diag barrelman.Diagnostic) ([]codemod.Patch, error) {
	mapping := mappingForDiagnostic(ctx, diag)
	if mapping == nil {
		return nil, nil
	}
	// Idempotent: only fix the "no tags:" case. If a tags key already
	// exists (empty/multi-entry), leave it to humans.
	if codemod.MappingHasKey(mapping, ctx.Source, "tags") {
		return nil, nil
	}
	patch := codemod.InsertMappingPair(mapping, ctx.Source, "tags", "["+TodoDescription+"]")
	patch.URI = ctx.URI
	patch.RuleID = diag.Code
	patch.Description = fmt.Sprintf("insert tags: [%s] on operation", TodoDescription)
	return []codemod.Patch{patch}, nil
}
