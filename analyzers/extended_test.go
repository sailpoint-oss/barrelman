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

func TestOperationTags(t *testing.T) {
	rule := registeredRule("operation-tags")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "operation with tags passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /pets:
    get:
      tags:
        - pets
      responses:
        "200":
          description: ok`,
			Expect: nil,
		},
		btesting.Case{
			Name: "operation without tags triggers warning",
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
				{Code: "operation-tags", Severity: btesting.Warn},
			},
		},
	)
}

func TestOperationOperationID(t *testing.T) {
	rule := registeredRule("operation-operationId")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "operation with operationId passes",
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
          description: ok`,
			Expect: nil,
		},
		btesting.Case{
			Name: "operation without operationId triggers warning",
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
				{Code: "operation-operationId", Severity: btesting.Warn},
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

func TestTagDescription(t *testing.T) {
	rule := registeredRule("tag-description")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "tag with description passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
tags:
  - name: pets
    description: Everything about your pets`,
			Expect: nil,
		},
		btesting.Case{
			Name: "tag without description triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
tags:
  - name: pets`,
			Expect: []btesting.Diag{
				{Code: "tag-description", Severity: btesting.Warn},
			},
		},
	)
}

func TestParameterDescription(t *testing.T) {
	rule := registeredRule("parameter-description")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "parameter with description passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /pets:
    get:
      parameters:
        - name: limit
          in: query
          description: How many items to return
          schema:
            type: integer
      responses:
        "200":
          description: ok`,
			Expect: nil,
		},
		btesting.Case{
			Name: "parameter without description triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /pets:
    get:
      parameters:
        - name: limit
          in: query
          schema:
            type: integer
      responses:
        "200":
          description: ok`,
			Expect: []btesting.Diag{
				{Code: "parameter-description", Severity: btesting.Warn},
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
