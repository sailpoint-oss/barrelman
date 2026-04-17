package analyzers_test

import (
	"testing"

	"github.com/sailpoint-oss/barrelman/btesting"
)

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
