package analyzers_test

import (
	"testing"

	"github.com/sailpoint-oss/barrelman/btesting"
)

func TestUnresolvedRef(t *testing.T) {
	rule := registeredRule("unresolved-ref")

	btesting.Run(t, rule,
		btesting.Case{
			Name: "valid ref passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /pets:
    get:
      responses:
        "200":
          description: ok
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Pet'
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
			Name: "broken ref triggers error",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /pets:
    get:
      responses:
        "200":
          description: ok
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Nonexistent'
components:
  schemas:
    Pet:
      type: object`,
			Expect: []btesting.Diag{
				{Code: "unresolved-ref", Severity: btesting.Error},
			},
		},
	)
}

func TestUnusedComponent(t *testing.T) {
	rule := registeredRule("unused-component")

	btesting.Run(t, rule,
		btesting.Case{
			Name: "referenced schema passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /pets:
    get:
      responses:
        "200":
          description: ok
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Pet'
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
			Name: "unreferenced schema triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /pets:
    get:
      responses:
        "200":
          description: ok
components:
  schemas:
    Unused:
      type: object
      properties:
        name:
          type: string`,
			Expect: []btesting.Diag{
				{Code: "unused-component", Severity: btesting.Warn},
			},
		},
	)
}
