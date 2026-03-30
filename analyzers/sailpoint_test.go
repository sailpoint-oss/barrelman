package analyzers_test

import (
	"testing"

	"github.com/sailpoint-oss/barrelman/btesting"
)

func TestSP111ScopeNaming(t *testing.T) {
	rule := registeredRule("sp-111")

	btesting.Run(t, rule,
		btesting.Case{
			Name: "three segment lowercase scopes pass",
			Spec: `openapi: "3.1.0"
info:
  title: Test
  version: "1.0"
paths:
  /users:
    get:
      operationId: listUsers
      security:
        - oauth2:
            - identity:users:read
      responses:
        "200":
          description: OK
components:
  securitySchemes:
    oauth2:
      type: oauth2
      flows:
        clientCredentials:
          tokenUrl: https://auth.example.com/oauth/token
          scopes:
            identity:users:read: Read users`,
			Expect: nil,
		},
		btesting.Case{
			Name: "invalid declared and required scope names trigger errors",
			Spec: `openapi: "3.1.0"
info:
  title: Test
  version: "1.0"
paths:
  /users:
    get:
      operationId: listUsers
      security:
        - oauth2:
            - identity.users.read
      responses:
        "200":
          description: OK
components:
  securitySchemes:
    oauth2:
      type: oauth2
      flows:
        clientCredentials:
          tokenUrl: https://auth.example.com/oauth/token
          scopes:
            identity.users.read: Read users`,
			Expect: []btesting.Diag{
				{Code: "sp-111", Severity: btesting.Error, Message: "declared by security scheme"},
				{Code: "sp-111", Severity: btesting.Error, Message: "must use lower-case <domain>:<resource>:<action> naming"},
			},
		},
	)
}

func TestSP403StatusCodes(t *testing.T) {
	rule := registeredRule("sp-403")

	btesting.Run(t, rule,
		btesting.Case{
			Name: "get with standard auth responses passes",
			Spec: `openapi: "3.1.0"
info:
  title: Test
  version: "1.0"
paths:
  /users:
    get:
      operationId: listUsers
      responses:
        "200":
          description: OK
        "401":
          description: Unauthorized
        "403":
          description: Forbidden
        "404":
          description: Not found`,
			Expect: nil,
		},
		btesting.Case{
			Name: "missing 403 triggers error",
			Spec: `openapi: "3.1.0"
info:
  title: Test
  version: "1.0"
paths:
  /users:
    get:
      operationId: listUsers
      responses:
        "200":
          description: OK
        "401":
          description: Unauthorized
        "404":
          description: Not found`,
			Expect: []btesting.Diag{
				{Code: "sp-403", Severity: btesting.Error, Message: "must declare a 403 response"},
			},
		},
	)
}

func TestSP404ProblemDetails(t *testing.T) {
	rule := registeredRule("sp-404")

	btesting.Run(t, rule,
		btesting.Case{
			Name: "shared problem details response passes",
			Spec: `openapi: "3.1.0"
info:
  title: Test
  version: "1.0"
paths:
  /users:
    get:
      operationId: listUsers
      responses:
        "400":
          description: Bad request
          content:
            application/problem+json:
              schema:
                $ref: '#/components/schemas/ProblemDetails'
              example:
                type: https://example.com/problems/invalid-request
                title: Invalid Request
                status: 400
                detail: The request is invalid.
                instance: /users
                correlationId: 123e4567-e89b-12d3-a456-426614174000
components:
  schemas:
    ProblemDetails:
      type: object
      properties:
        type:
          type: string
        title:
          type: string
        status:
          type: integer
        detail:
          type: string
        instance:
          type: string
        correlationId:
          type: string`,
			Expect: nil,
		},
		btesting.Case{
			Name: "inline problem details triggers shared component requirement",
			Spec: `openapi: "3.1.0"
info:
  title: Test
  version: "1.0"
paths:
  /users:
    get:
      operationId: listUsers
      responses:
        "400":
          description: Bad request
          content:
            application/problem+json:
              schema:
                type: object
                properties:
                  type:
                    type: string
                  title:
                    type: string
                  status:
                    type: integer
                  detail:
                    type: string
                  instance:
                    type: string
                  correlationId:
                    type: string
              example:
                type: https://example.com/problems/invalid-request
                title: Invalid Request
                status: 400
                detail: The request is invalid.
                instance: /users
                correlationId: 123e4567-e89b-12d3-a456-426614174000`,
			Expect: []btesting.Diag{
				{Code: "sp-404", Severity: btesting.Error, Message: "shared ProblemDetails schema"},
				{Code: "sp-404", Severity: btesting.Error, Message: "must reference #/components/schemas/ProblemDetails"},
			},
		},
	)
}

func TestSP710RequiredFields(t *testing.T) {
	rule := registeredRule("sp-710")

	btesting.Run(t, rule,
		btesting.Case{
			Name: "explicit required arrays pass",
			Spec: `openapi: "3.1.0"
info:
  title: Test
  version: "1.0"
paths:
  /users/{userId}:
    put:
      operationId: updateUser
      parameters:
        - name: userId
          in: path
          required: true
          schema:
            type: string
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: []
              properties:
                name:
                  type: string
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: object
                required:
                  - id
                properties:
                  id:
                    type: string`,
			Expect: nil,
		},
		btesting.Case{
			Name: "missing required array and required path flag trigger errors",
			Spec: `openapi: "3.1.0"
info:
  title: Test
  version: "1.0"
paths:
  /users/{userId}:
    put:
      operationId: updateUser
      parameters:
        - name: userId
          in: path
          schema:
            type: string
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                name:
                  type: string
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: string`,
			Expect: []btesting.Diag{
				{Code: "sp-710", Severity: btesting.Error, Message: "must set required: true"},
				{Code: "sp-710", Severity: btesting.Error, Message: "Request body schema"},
				{Code: "sp-710", Severity: btesting.Error, Message: "Response schema"},
			},
		},
	)
}

func TestSP903RequestIDHeader(t *testing.T) {
	rule := registeredRule("sp-903")

	btesting.Run(t, rule,
		btesting.Case{
			Name: "shared request id header passes",
			Spec: `openapi: "3.1.0"
info:
  title: Test
  version: "1.0"
paths:
  /users:
    get:
      operationId: listUsers
      responses:
        "200":
          description: OK
          headers:
            X-Request-Id:
              $ref: '#/components/headers/X-Request-Id'
components:
  headers:
    X-Request-Id:
      description: Request correlation id.
      schema:
        type: string
        format: uuid`,
			Expect: nil,
		},
		btesting.Case{
			Name: "missing shared header component triggers error",
			Spec: `openapi: "3.1.0"
info:
  title: Test
  version: "1.0"
paths:
  /users:
    get:
      operationId: listUsers
      responses:
        "200":
          description: OK
          headers:
            X-Request-Id:
              description: Request correlation id.
              schema:
                type: string
                format: uuid`,
			Expect: []btesting.Diag{
				{Code: "sp-903", Severity: btesting.Error, Message: "shared X-Request-Id header"},
			},
		},
	)
}

func TestSP124ResourceOperationID(t *testing.T) {
	rule := registeredRule("sp-124")

	btesting.Run(t, rule,
		btesting.Case{
			Name: "path parameter extension may use referenced component",
			Spec: `openapi: "3.1.0"
info:
  title: Test
  version: "1.0"
paths:
  /users/{userId}:
    get:
      operationId: getUser
      parameters:
        - $ref: '#/components/parameters/UserId'
      responses:
        "200":
          description: OK
components:
  parameters:
    UserId:
      name: userId
      in: path
      required: true
      x-sailpoint-resource-operation-id: getUser
      schema:
        type: string`,
			Expect: nil,
		},
		btesting.Case{
			Name: "missing extension triggers error",
			Spec: `openapi: "3.1.0"
info:
  title: Test
  version: "1.0"
paths:
  /users/{userId}:
    get:
      operationId: getUser
      parameters:
        - name: userId
          in: path
          required: true
          schema:
            type: string
      responses:
        "200":
          description: OK`,
			Expect: []btesting.Diag{
				{Code: "sp-124", Severity: btesting.Error, Message: "must declare x-sailpoint-resource-operation-id"},
			},
		},
		btesting.Case{
			Name: "non camel case unknown operation id triggers both checks",
			Spec: `openapi: "3.1.0"
info:
  title: Test
  version: "1.0"
paths:
  /users/{userId}:
    get:
      operationId: getUser
      parameters:
        - name: userId
          in: path
          required: true
          x-sailpoint-resource-operation-id: GetUserResource
          schema:
            type: string
      responses:
        "200":
          description: OK`,
			Expect: []btesting.Diag{
				{Code: "sp-124", Severity: btesting.Error, Message: "must use lowerCamelCase"},
				{Code: "sp-124", Severity: btesting.Error, Message: "must reference an existing operationId"},
			},
		},
	)
}
