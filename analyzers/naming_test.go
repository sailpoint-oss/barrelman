package analyzers_test

import (
	"testing"

	"github.com/sailpoint-oss/barrelman"
	"github.com/sailpoint-oss/barrelman/analyzers"
	"github.com/sailpoint-oss/barrelman/btesting"
)

func registeredRule(id string) barrelman.Rule {
	reg := barrelman.NewRegistry()
	analyzers.RegisterAll(reg)
	for _, r := range reg.AllRules() {
		if r.ID == id {
			return r
		}
	}
	panic("rule not found: " + id)
}

func TestSchemaNameCapital(t *testing.T) {
	rule := registeredRule("schema-name-capital")

	btesting.Run(t, rule,
		btesting.Case{
			Name: "uppercase schema passes",
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
			Name: "lowercase schema triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  schemas:
    pet:
      type: object`,
			Expect: []btesting.Diag{
				{Code: "schema-name-capital", Severity: btesting.Warn},
			},
		},
	)
}

func TestOperationIDUnique(t *testing.T) {
	rule := registeredRule("operation-operationId-unique")

	btesting.Run(t, rule,
		btesting.Case{
			Name: "unique operationIds pass",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /pets:
    get:
      operationId: listPets
      responses:
        "200":
          description: ok
  /users:
    get:
      operationId: listUsers
      responses:
        "200":
          description: ok`,
			Expect: nil,
		},
		btesting.Case{
			Name: "duplicate operationId triggers error",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /pets:
    get:
      operationId: listItems
      responses:
        "200":
          description: ok
  /users:
    get:
      operationId: listItems
      responses:
        "200":
          description: ok`,
			Expect: []btesting.Diag{
				{
					Code:     "operation-operationId-unique",
					Severity: btesting.Error,
					Message:  "GET /pets",
				},
			},
		},
	)
}

func TestTagsFormat(t *testing.T) {
	rule := registeredRule("tags-format")

	btesting.Run(t, rule,
		btesting.Case{
			Name: "named tag passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
tags:
  - name: pets`,
			Expect: nil,
		},
		btesting.Case{
			Name: "empty tag name triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
tags:
  - name: ""`,
			Expect: []btesting.Diag{
				{Code: "tags-format", Severity: btesting.Warn},
			},
		},
	)
}
