package checks_test

import (
	"testing"

	"github.com/sailpoint-oss/barrelman/btesting"
)

func TestDuplicateKeys(t *testing.T) {
	rule := registeredCheck("duplicate-keys")

	btesting.Run(t, rule,
		btesting.Case{
			Name: "unique keys pass",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}`,
			Expect: nil,
		},
		btesting.Case{
			Name: "duplicate top-level key triggers error",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
paths: {}`,
			Expect: []btesting.Diag{
				{Line: 5, Code: "duplicate-keys", Severity: btesting.Error},
			},
		},
		btesting.Case{
			Name: "duplicate nested key triggers error",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
  title: Duplicate`,
			Expect: []btesting.Diag{
				{Line: 4, Code: "duplicate-keys", Severity: btesting.Error},
			},
		},
	)
}
