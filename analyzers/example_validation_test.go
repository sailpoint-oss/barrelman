package analyzers_test

import (
	"testing"

	"github.com/sailpoint-oss/barrelman/btesting"
)

func TestOAS3ValidMediaExample_ValidExample(t *testing.T) {
	rule := registeredRule("oas3-valid-media-example")

	btesting.Run(t, rule, btesting.Case{
		Name: "valid media example passes",
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
                type: string
              example: hello`,
		Expect: nil,
	})
}

func TestOAS3ValidMediaExample_TypeMismatch(t *testing.T) {
	rule := registeredRule("oas3-valid-media-example")

	btesting.Run(t, rule, btesting.Case{
		Name: "media example type mismatch triggers error",
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
                type: integer
              example: not-a-number`,
		Expect: []btesting.Diag{
			{Code: "oas3-valid-media-example", Severity: btesting.Warn, Message: "type mismatch"},
		},
	})
}

func TestOAS3ValidMediaExample_EnumMismatch(t *testing.T) {
	rule := registeredRule("oas3-valid-media-example")

	btesting.Run(t, rule, btesting.Case{
		Name: "media example enum mismatch triggers error",
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
                type: string
                enum:
                  - cat
                  - dog
              example: fish`,
		Expect: []btesting.Diag{
			{Code: "oas3-valid-media-example", Severity: btesting.Warn, Message: "not in enum"},
		},
	})
}

func TestOAS3ValidMediaExample_NoSchema(t *testing.T) {
	rule := registeredRule("oas3-valid-media-example")

	btesting.Run(t, rule, btesting.Case{
		Name: "media type without schema produces no diagnostics",
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
              example: anything`,
		Expect: nil,
	})
}

func TestOAS3ValidMediaExample_RequestBody(t *testing.T) {
	rule := registeredRule("oas3-valid-media-example")

	btesting.Run(t, rule, btesting.Case{
		Name: "request body example type mismatch triggers error",
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
              type: boolean
            example: not-bool
      responses:
        "201":
          description: created`,
		Expect: []btesting.Diag{
			{Code: "oas3-valid-media-example", Severity: btesting.Warn, Message: "type mismatch"},
		},
	})
}

func TestOAS3ValidSchemaExample_ValidExample(t *testing.T) {
	rule := registeredRule("oas3-valid-schema-example")

	btesting.Run(t, rule, btesting.Case{
		Name: "valid schema example passes",
		Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  schemas:
    Pet:
      type: string
      example: Fido`,
		Expect: nil,
	})
}

func TestOAS3ValidSchemaExample_TypeMismatch(t *testing.T) {
	rule := registeredRule("oas3-valid-schema-example")

	btesting.Run(t, rule, btesting.Case{
		Name: "schema example type mismatch triggers error",
		Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  schemas:
    Age:
      type: integer
      example: not-a-number`,
		Expect: []btesting.Diag{
			{Code: "oas3-valid-schema-example", Severity: btesting.Warn, Message: "does not match type"},
		},
	})
}

func TestOAS3ValidSchemaExample_EnumMismatch(t *testing.T) {
	rule := registeredRule("oas3-valid-schema-example")

	btesting.Run(t, rule, btesting.Case{
		Name: "schema example enum mismatch triggers error",
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
        - inactive
      example: deleted`,
		Expect: []btesting.Diag{
			{Code: "oas3-valid-schema-example", Severity: btesting.Warn, Message: "not in enum"},
		},
	})
}

func TestOAS3ValidSchemaExample_NoExample(t *testing.T) {
	rule := registeredRule("oas3-valid-schema-example")

	btesting.Run(t, rule, btesting.Case{
		Name: "schema without example produces no diagnostics",
		Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  schemas:
    Pet:
      type: string`,
		Expect: nil,
	})
}

func TestOAS3ValidSchemaExample_BooleanValid(t *testing.T) {
	rule := registeredRule("oas3-valid-schema-example")

	btesting.Run(t, rule, btesting.Case{
		Name: "valid boolean example passes",
		Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  schemas:
    Active:
      type: boolean
      example: true`,
		Expect: nil,
	})
}
