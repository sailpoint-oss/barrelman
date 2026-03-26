package analyzers_test

import (
	"testing"

	"github.com/sailpoint-oss/barrelman/btesting"
)

func TestMigrationNullable(t *testing.T) {
	rule := registeredRule("migration-nullable")

	btesting.Run(t, rule,
		btesting.Case{
			Name: "non-nullable schema in 3.0 passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  schemas:
    Pet:
      type: object
      properties:
        name:
          type: string`,
			Expect: nil,
		},
		btesting.Case{
			Name: "nullable schema in 3.0 triggers info",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  schemas:
    Pet:
      type: object
      properties:
        name:
          type: string
          nullable: true`,
			Expect: []btesting.Diag{
				{Code: "migration-nullable", Severity: btesting.Info},
			},
		},
		btesting.Case{
			Name: "nullable schema in 3.1 does not fire",
			Spec: `openapi: "3.1.0"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  schemas:
    Pet:
      type: object
      properties:
        name:
          type: string
          nullable: true`,
			Expect: nil,
		},
	)
}
