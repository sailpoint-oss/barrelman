package analyzers_test

import (
	"testing"

	"github.com/sailpoint-oss/barrelman/btesting"
)

func TestArrayItems(t *testing.T) {
	rule := registeredRule("array-items")

	btesting.Run(t, rule,
		btesting.Case{
			Name: "array with items passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  schemas:
    PetList:
      type: array
      items:
        type: string`,
			Expect: nil,
		},
		btesting.Case{
			Name: "array without items triggers error",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  schemas:
    PetList:
      type: array`,
			Expect: []btesting.Diag{
				{Code: "array-items", Severity: btesting.Error},
			},
		},
	)
}

func TestTypeRequired(t *testing.T) {
	rule := registeredRule("type-required")

	btesting.Run(t, rule,
		btesting.Case{
			Name: "schema with type passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  schemas:
    Pet:
      type: object`,
			Expect: nil,
		},
		btesting.Case{
			Name: "schema without type triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  schemas:
    Pet:
      description: A pet`,
			Expect: []btesting.Diag{
				{Code: "type-required", Severity: btesting.Warn},
			},
		},
	)
}

func TestRequestBodyContent(t *testing.T) {
	rule := registeredRule("request-body-content")

	btesting.Run(t, rule,
		btesting.Case{
			Name: "request body with content passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /pets:
    post:
      requestBody:
        content:
          application/json:
            schema:
              type: object
      responses:
        "201":
          description: created`,
			Expect: nil,
		},
		btesting.Case{
			Name: "request body without content triggers error",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /pets:
    post:
      requestBody:
        description: A pet
      responses:
        "201":
          description: created`,
			Expect: []btesting.Diag{
				{Code: "request-body-content", Severity: btesting.Error},
			},
		},
	)
}
