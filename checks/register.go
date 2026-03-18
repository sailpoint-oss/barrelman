// Package checks implements syntactic diagnostic rules. These operate on
// raw source content and tree-sitter trees without needing the OpenAPI index.
package checks

import (
	"github.com/sailpoint-oss/barrelman"
)

// RegisterAll registers all syntactic checks with the given registry.
func RegisterAll(reg *barrelman.Registry) {
	registerDuplicateKeys(reg)
	registerASCII(reg)
}
