package analyzers_test

import (
	"testing"

	"github.com/sailpoint-oss/barrelman/btesting"
)

func TestSailpointOAuthScopeFormat(t *testing.T) {
	rule := registeredRule("sailpoint-oauth-scope-format")

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
				{Code: "sailpoint-oauth-scope-format", Severity: btesting.Error, Message: "declared by security scheme"},
				{Code: "sailpoint-oauth-scope-format", Severity: btesting.Error, Message: "must use lower-case <domain>:<resource>:<action> naming"},
			},
		},
	)
}

func TestSailpointOperation403Response(t *testing.T) {
	rule := registeredRule("sailpoint-operation-403-response")

	btesting.Run(t, rule,
		btesting.Case{
			Name: "operation with 403 passes",
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
				{Code: "sailpoint-operation-403-response", Severity: btesting.Error, Message: "must declare a 403 response"},
			},
		},
	)
}

func TestSailpointErrorProblemDetailsSharedComponent(t *testing.T) {
	rule := registeredRule("sailpoint-error-problem-details-shared-component")

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
				{Code: "sailpoint-error-problem-details-shared-component", Severity: btesting.Error, Message: "shared ProblemDetails schema"},
				{Code: "sailpoint-error-problem-details-shared-component", Severity: btesting.Error, Message: "must reference #/components/schemas/ProblemDetails"},
			},
		},
	)
}

func TestSailpointSchemaRequiredFields(t *testing.T) {
	rule := registeredRule("sailpoint-schema-required-fields")

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
			Name: "missing required arrays in request/response trigger errors",
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
				{Code: "sailpoint-schema-required-fields", Severity: btesting.Error, Message: "Request body schema"},
				{Code: "sailpoint-schema-required-fields", Severity: btesting.Error, Message: "Response schema"},
			},
		},
	)
}

func TestSailpointPathParamRequired(t *testing.T) {
	rule := registeredRule("sailpoint-path-param-required")

	btesting.Run(t, rule,
		btesting.Case{
			Name: "required path param passes",
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
      responses:
        "200":
          description: OK`,
			Expect: nil,
		},
		btesting.Case{
			Name: "path parameter missing required flag triggers error",
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
      responses:
        "200":
          description: OK`,
			Expect: []btesting.Diag{
				{Code: "sailpoint-path-param-required", Severity: btesting.Error, Message: "must set required: true"},
			},
		},
	)
}

func TestSailpointXRequestIDSharedComponent(t *testing.T) {
	rule := registeredRule("sailpoint-x-request-id-shared-component")

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
				{Code: "sailpoint-x-request-id-shared-component", Severity: btesting.Error, Message: "shared X-Request-Id header"},
			},
		},
	)
}

func TestSailpointPathParamResourceOperationLink(t *testing.T) {
	rule := registeredRule("sailpoint-path-param-resource-operation-link")

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
				{Code: "sailpoint-path-param-resource-operation-link", Severity: btesting.Error, Message: "must declare x-sailpoint-resource-operation-id"},
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
				{Code: "sailpoint-path-param-resource-operation-link", Severity: btesting.Error, Message: "must use lowerCamelCase"},
				{Code: "sailpoint-path-param-resource-operation-link", Severity: btesting.Error, Message: "must reference an existing operationId"},
			},
		},
	)
}
