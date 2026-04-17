package analyzers_test

import (
	"testing"

	"github.com/sailpoint-oss/barrelman/btesting"
)

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
      description: Returns pets.
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

func TestInfoContact(t *testing.T) {
	rule := registeredRule("info-contact")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "info without contact triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}`,
			Expect: []btesting.Diag{
				{Code: "info-contact", Severity: btesting.Warn},
			},
		},
	)
}

func TestInfoLicense(t *testing.T) {
	rule := registeredRule("info-license")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "info without license triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}`,
			Expect: []btesting.Diag{
				{Code: "info-license", Severity: btesting.Warn},
			},
		},
	)
}

func TestResponseDescription(t *testing.T) {
	rule := registeredRule("response-description")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "response with description passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /pets:
    get:
      responses:
        "200":
          description: A list of pets`,
			Expect: nil,
		},
		btesting.Case{
			Name: "response without description triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /pets:
    get:
      responses:
        "200":
          content:
            application/json:
              schema:
                type: array`,
			Expect: []btesting.Diag{
				{Code: "response-description", Severity: btesting.Warn},
			},
		},
	)
}

func TestSchemaDescription(t *testing.T) {
	rule := registeredRule("schema-description")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "schema with description passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  schemas:
    Pet:
      type: object
      description: A pet in the store`,
			Expect: nil,
		},
		btesting.Case{
			Name: "schema without description triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  schemas:
    Pet:
      type: object`,
			Expect: []btesting.Diag{
				{Code: "schema-description", Severity: btesting.Warn},
			},
		},
	)
}
