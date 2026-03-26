package analyzers_test

import (
	"testing"

	"github.com/sailpoint-oss/barrelman/btesting"
)

func TestMissingErrorResponses(t *testing.T) {
	rule := registeredRule("missing-error-responses")

	btesting.Run(t, rule,
		btesting.Case{
			Name: "operation with error response passes",
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
        "500":
          description: error`,
			Expect: nil,
		},
		btesting.Case{
			Name: "operation with default response passes",
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
        default:
          description: error`,
			Expect: nil,
		},
		btesting.Case{
			Name: "operation with only success response triggers warning",
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
				{Code: "missing-error-responses", Severity: btesting.Warn},
			},
		},
	)
}

func TestResponseBodyOnDelete(t *testing.T) {
	rule := registeredRule("response-body-on-delete")

	btesting.Run(t, rule,
		btesting.Case{
			Name: "DELETE with 204 passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /pets/{id}:
    delete:
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        "204":
          description: deleted`,
			Expect: nil,
		},
		btesting.Case{
			Name: "DELETE with 200 body triggers info",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /pets/{id}:
    delete:
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        "200":
          description: deleted
          content:
            application/json:
              schema:
                type: object`,
			Expect: []btesting.Diag{
				{Code: "response-body-on-delete", Severity: btesting.Info},
			},
		},
	)
}

func TestNoRequestBodyOnGet(t *testing.T) {
	rule := registeredRule("no-request-body-on-get")

	btesting.Run(t, rule,
		btesting.Case{
			Name: "GET without request body passes",
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
			Expect: nil,
		},
		btesting.Case{
			Name: "GET with request body triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /pets:
    get:
      requestBody:
        content:
          application/json:
            schema:
              type: object
      responses:
        "200":
          description: ok`,
			Expect: []btesting.Diag{
				{Code: "no-request-body-on-get", Severity: btesting.Warn},
			},
		},
	)
}

func TestMissingPagination(t *testing.T) {
	rule := registeredRule("missing-pagination")

	btesting.Run(t, rule,
		btesting.Case{
			Name: "GET returning array with pagination passes",
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
        - name: offset
          in: query
          schema:
            type: integer
      responses:
        "200":
          description: ok
          content:
            application/json:
              schema:
                type: array
                items:
                  type: object`,
			Expect: nil,
		},
		btesting.Case{
			Name: "GET returning array without pagination triggers info",
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
                type: array
                items:
                  type: object`,
			Expect: []btesting.Diag{
				{Code: "missing-pagination", Severity: btesting.Info},
			},
		},
		btesting.Case{
			Name: "GET returning object without pagination passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /pets/{id}:
    get:
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        "200":
          description: ok
          content:
            application/json:
              schema:
                type: object`,
			Expect: nil,
		},
	)
}

func TestInconsistentErrorShape(t *testing.T) {
	rule := registeredRule("inconsistent-error-shape")

	btesting.Run(t, rule,
		btesting.Case{
			Name: "consistent error refs pass",
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
        "400":
          description: bad request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
    post:
      requestBody:
        content:
          application/json:
            schema:
              type: object
      responses:
        "201":
          description: created
        "400":
          description: bad request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
components:
  schemas:
    Error:
      type: object
      properties:
        message:
          type: string`,
			Expect: nil,
		},
		btesting.Case{
			Name: "inconsistent error schemas triggers info",
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
        "400":
          description: bad request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorA'
    post:
      requestBody:
        content:
          application/json:
            schema:
              type: object
      responses:
        "201":
          description: created
        "400":
          description: bad request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorB'
components:
  schemas:
    ErrorA:
      type: object
      properties:
        message:
          type: string
    ErrorB:
      type: object
      properties:
        error:
          type: string`,
			Expect: []btesting.Diag{
				{Code: "inconsistent-error-shape", Severity: btesting.Info},
			},
		},
	)
}
