package analyzers_test

import (
	"testing"

	"github.com/sailpoint-oss/barrelman/btesting"
)

func TestInfoDescription(t *testing.T) {
	rule := registeredRule("info-description")

	btesting.Run(t, rule,
		btesting.Case{
			Name: "info with description passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
  description: A test API
paths: {}`,
			Expect: nil,
		},
		btesting.Case{
			Name: "info without description triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}`,
			Expect: []btesting.Diag{
				{Code: "info-description", Severity: btesting.Warn},
			},
		},
	)
}

func TestOperationDescription(t *testing.T) {
	rule := registeredRule("operation-description")

	btesting.Run(t, rule,
		btesting.Case{
			Name: "operation with description passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /pets:
    get:
      description: List all pets
      responses:
        "200":
          description: ok`,
			Expect: nil,
		},
		btesting.Case{
			Name: "operation without description triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /pets:
    get:
      responses:
        "200":
          description: ok`,
			Expect: []btesting.Diag{
				{Code: "operation-description", Severity: btesting.Warn},
			},
		},
	)
}

func TestEnumDescription(t *testing.T) {
	rule := registeredRule("enum-description")

	btesting.Run(t, rule,
		btesting.Case{
			Name: "enum with description passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  schemas:
    Status:
      type: string
      description: The status of the resource
      enum:
        - active
        - inactive`,
			Expect: nil,
		},
		btesting.Case{
			Name: "enum without description triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  schemas:
    Status:
      type: string
      enum:
        - active
        - inactive`,
			Expect: []btesting.Diag{
				{Code: "enum-description", Severity: btesting.Warn},
			},
		},
	)
}
