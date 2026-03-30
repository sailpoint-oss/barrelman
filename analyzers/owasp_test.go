package analyzers_test

import (
	"testing"

	"github.com/sailpoint-oss/barrelman/btesting"
)

func TestOWASPNoHTTPBasic(t *testing.T) {
	rule := registeredRule("owasp-no-http-basic")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "bearer auth passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer`,
			Expect: nil,
		},
		btesting.Case{
			Name: "basic auth triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  securitySchemes:
    BasicAuth:
      type: http
      scheme: basic`,
			Expect: []btesting.Diag{
				{Code: "owasp-no-http-basic", Severity: btesting.Warn},
			},
		},
	)
}

func TestOWASPNoAPIKeysInURL(t *testing.T) {
	rule := registeredRule("owasp-no-api-keys-in-url")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "apiKey in header passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  securitySchemes:
    ApiKey:
      type: apiKey
      in: header
      name: X-API-Key`,
			Expect: nil,
		},
		btesting.Case{
			Name: "apiKey in query triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  securitySchemes:
    ApiKey:
      type: apiKey
      in: query
      name: api_key`,
			Expect: []btesting.Diag{
				{Code: "owasp-no-api-keys-in-url", Severity: btesting.Warn},
			},
		},
	)
}

func TestOWASPNoCredentialsInURL(t *testing.T) {
	rule := registeredRule("owasp-no-credentials-in-url")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "clean server URL passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
servers:
  - url: https://api.example.com`,
			Expect: nil,
		},
		btesting.Case{
			Name: "URL with credentials triggers error",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
servers:
  - url: https://user:pass@api.example.com`,
			Expect: []btesting.Diag{
				{Code: "owasp-no-credentials-in-url", Severity: btesting.Error},
			},
		},
	)
}

func TestOWASPAuthInsecureSchemes(t *testing.T) {
	rule := registeredRule("owasp-auth-insecure-schemes")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "bearer scheme passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer`,
			Expect: nil,
		},
		btesting.Case{
			Name: "negotiate scheme triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  securitySchemes:
    NegotiateAuth:
      type: http
      scheme: negotiate`,
			Expect: []btesting.Diag{
				{Code: "owasp-auth-insecure-schemes", Severity: btesting.Warn},
			},
		},
	)
}

func TestOWASPJWTBestPractices(t *testing.T) {
	rule := registeredRule("owasp-jwt-best-practices")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "non-bearer scheme passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  securitySchemes:
    ApiKey:
      type: apiKey
      in: header
      name: X-API-Key`,
			Expect: nil,
		},
		btesting.Case{
			Name: "bearer without bearerFormat triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer`,
			Expect: []btesting.Diag{
				{Code: "owasp-jwt-best-practices", Severity: btesting.Warn},
			},
		},
	)
}

func TestOWASPShortLivedAccessTokens(t *testing.T) {
	rule := registeredRule("owasp-short-lived-access-tokens")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "non-oauth2 scheme passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer`,
			Expect: nil,
		},
		btesting.Case{
			Name: "oauth2 implicit flow without refreshUrl warns",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  securitySchemes:
    OAuth:
      type: oauth2
      flows:
        implicit:
          authorizationUrl: https://example.com/auth
          scopes:
            read: read`,
			Expect: []btesting.Diag{
				{Code: "owasp-short-lived-access-tokens", Severity: btesting.Warn},
			},
		},
	)
}

func TestOWASPProtectionGlobalUnsafe(t *testing.T) {
	rule := registeredRule("owasp-protection-global-unsafe")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "global security covers unsafe op",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
security:
  - BearerAuth: []
paths:
  /pets:
    post:
      responses:
        "201":
          description: created
components:
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer`,
			Expect: nil,
		},
		btesting.Case{
			Name: "POST without security triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /pets:
    post:
      responses:
        "201":
          description: created`,
			Expect: []btesting.Diag{
				{Code: "owasp-protection-global-unsafe", Severity: btesting.Warn},
			},
		},
	)
}

func TestOWASPProtectionGlobalSafe(t *testing.T) {
	rule := registeredRule("owasp-protection-global-safe")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "global security defined passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
security:
  - BearerAuth: []
paths:
  /pets:
    get:
      responses:
        "200":
          description: ok
components:
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer`,
			Expect: nil,
		},
		btesting.Case{
			Name: "operation without any security triggers info",
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
				{Code: "owasp-protection-global-safe", Severity: btesting.Info},
			},
		},
	)
}

func TestOWASPDefineErrorResponses401(t *testing.T) {
	rule := registeredRule("owasp-define-error-responses-401")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "401 response defined passes",
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
        "401":
          description: unauthorized`,
			Expect: nil,
		},
		btesting.Case{
			Name: "missing 401 triggers warning",
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
				{Code: "owasp-define-error-responses-401", Severity: btesting.Warn},
			},
		},
	)
}

func TestOWASPDefineErrorResponses500(t *testing.T) {
	rule := registeredRule("owasp-define-error-responses-500")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "500 response defined passes",
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
          description: server error`,
			Expect: nil,
		},
		btesting.Case{
			Name: "missing 500 triggers warning",
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
				{Code: "owasp-define-error-responses-500", Severity: btesting.Warn},
			},
		},
	)
}

func TestOWASPRateLimit(t *testing.T) {
	rule := registeredRule("owasp-rate-limit")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "no 200 or 201 response passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /pets:
    delete:
      responses:
        "204":
          description: deleted`,
			Expect: nil,
		},
		btesting.Case{
			Name: "200 response without rate limit headers triggers warning",
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
				{Code: "owasp-rate-limit", Severity: btesting.Warn},
			},
		},
	)
}

func TestOWASPRateLimitRetryAfter(t *testing.T) {
	rule := registeredRule("owasp-rate-limit-retry-after")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "no 429 response passes",
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
			Name: "429 without Retry-After triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /pets:
    get:
      responses:
        "429":
          description: too many requests`,
			Expect: []btesting.Diag{
				{Code: "owasp-rate-limit-retry-after", Severity: btesting.Warn},
			},
		},
	)
}

func TestOWASPRateLimitResponses429(t *testing.T) {
	rule := registeredRule("owasp-rate-limit-responses-429")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "429 response defined passes",
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
        "429":
          description: too many requests`,
			Expect: nil,
		},
		btesting.Case{
			Name: "missing 429 triggers warning",
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
				{Code: "owasp-rate-limit-responses-429", Severity: btesting.Warn},
			},
		},
	)
}

func TestOWASPDefineErrorValidation(t *testing.T) {
	rule := registeredRule("owasp-define-error-validation")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "operation with requestBody and 400 passes",
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
              type: object
      responses:
        "201":
          description: created
        "400":
          description: bad request`,
			Expect: nil,
		},
		btesting.Case{
			Name: "requestBody without 400 or 422 triggers warning",
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
              type: object
      responses:
        "201":
          description: created`,
			Expect: []btesting.Diag{
				{Code: "owasp-define-error-validation", Severity: btesting.Warn},
			},
		},
	)
}

func TestOWASPDefineCORSOrigin(t *testing.T) {
	rule := registeredRule("owasp-define-cors-origin")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "no 2xx response passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /pets:
    delete:
      responses:
        "404":
          description: not found`,
			Expect: nil,
		},
		btesting.Case{
			Name: "200 response without CORS header triggers warning",
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
				{Code: "owasp-define-cors-origin", Severity: btesting.Warn},
			},
		},
	)
}

func TestOWASPNoSchemeHTTP(t *testing.T) {
	rule := registeredRule("owasp-no-scheme-http")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "OAS 3.0 doc is skipped",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}`,
			Expect: nil,
		},
		btesting.Case{
			Name: "OAS 2.0 with https only passes",
			Spec: `swagger: "2.0"
info:
  title: T
  version: "1"
paths: {}
schemes:
  - https`,
			Expect: nil,
		},
	)
}

func TestOWASPNoServerHTTP(t *testing.T) {
	rule := registeredRule("owasp-no-server-http")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "https server URL passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
servers:
  - url: https://api.example.com`,
			Expect: nil,
		},
		btesting.Case{
			Name: "http server URL triggers error",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
servers:
  - url: http://api.example.com`,
			Expect: []btesting.Diag{
				{Code: "owasp-no-server-http", Severity: btesting.Error},
			},
		},
	)
}

func TestOWASPNoNumericIDs(t *testing.T) {
	rule := registeredRule("owasp-no-numeric-ids")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "string schema named with id suffix passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  schemas:
    UserId:
      type: string`,
			Expect: nil,
		},
		btesting.Case{
			Name: "integer schema named with id suffix triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  schemas:
    UserId:
      type: integer`,
			Expect: []btesting.Diag{
				{Code: "owasp-no-numeric-ids", Severity: btesting.Warn},
			},
		},
	)
}

func TestOWASPNoAdditionalProperties(t *testing.T) {
	rule := registeredRule("owasp-no-additionalProperties")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "object without properties passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  schemas:
    Empty:
      type: object`,
			Expect: nil,
		},
		btesting.Case{
			Name: "object with properties but no additionalProperties triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  schemas:
    Pet:
      type: object
      properties:
        name:
          type: string`,
			Expect: []btesting.Diag{
				{Code: "owasp-no-additionalProperties", Severity: btesting.Warn},
			},
		},
	)
}

func TestOWASPConstrainedAdditionalProperties(t *testing.T) {
	rule := registeredRule("owasp-constrained-additionalProperties")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "no additionalProperties schema passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
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
			Name: "additionalProperties with type but no constraints passes with tree-sitter",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  schemas:
    Pet:
      type: object
      properties:
        name:
          type: string
      additionalProperties:
        type: string`,
			Expect: nil,
		},
	)
}

func TestOWASPStringLimit(t *testing.T) {
	rule := registeredRule("owasp-string-limit")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "string with enum passes",
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
			Expect: nil,
		},
		btesting.Case{
			Name: "plain string without maxLength triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  schemas:
    Name:
      type: string`,
			Expect: []btesting.Diag{
				{Code: "owasp-string-limit", Severity: btesting.Warn},
			},
		},
	)
}

func TestOWASPStringRestricted(t *testing.T) {
	rule := registeredRule("owasp-string-restricted")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "string with format passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  schemas:
    Email:
      type: string
      format: email`,
			Expect: nil,
		},
		btesting.Case{
			Name: "unrestricted string triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  schemas:
    Name:
      type: string`,
			Expect: []btesting.Diag{
				{Code: "owasp-string-restricted", Severity: btesting.Warn},
			},
		},
	)
}

func TestOWASPArrayLimit(t *testing.T) {
	rule := registeredRule("owasp-array-limit")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "non-array schema passes",
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
			Name: "array without maxItems triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  schemas:
    Pets:
      type: array
      items:
        type: string`,
			Expect: []btesting.Diag{
				{Code: "owasp-array-limit", Severity: btesting.Warn},
			},
		},
	)
}

func TestOWASPIntegerFormat(t *testing.T) {
	rule := registeredRule("owasp-integer-format")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "integer with format passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  schemas:
    Count:
      type: integer
      format: int32`,
			Expect: nil,
		},
		btesting.Case{
			Name: "integer without format triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  schemas:
    Count:
      type: integer`,
			Expect: []btesting.Diag{
				{Code: "owasp-integer-format", Severity: btesting.Warn},
			},
		},
	)
}

func TestOWASPNoUnevaluatedProperties(t *testing.T) {
	rule := registeredRule("owasp-no-unevaluatedProperties")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "OAS 3.0 doc is skipped",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
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
			Name: "OAS 3.1 object with properties but no unevaluatedProperties triggers warning",
			Spec: `openapi: "3.1.0"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  schemas:
    Pet:
      type: object
      properties:
        name:
          type: string`,
			Expect: []btesting.Diag{
				{Code: "owasp-no-unevaluatedProperties", Severity: btesting.Warn},
			},
		},
	)
}

func TestOWASPConstrainedUnevaluatedProperties(t *testing.T) {
	rule := registeredRule("owasp-constrained-unevaluatedProperties")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "OAS 3.0 doc is skipped",
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
			Name: "OAS 3.1 schema without unevaluatedProperties passes",
			Spec: `openapi: "3.1.0"
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
	)
}

func TestOWASPIntegerLimit(t *testing.T) {
	rule := registeredRule("owasp-integer-limit")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "OAS 3.0 doc is skipped",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  schemas:
    Count:
      type: integer`,
			Expect: nil,
		},
		btesting.Case{
			Name: "OAS 3.1 integer without bounds triggers warning",
			Spec: `openapi: "3.1.0"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  schemas:
    Count:
      type: integer`,
			Expect: []btesting.Diag{
				{Code: "owasp-integer-limit", Severity: btesting.Warn},
			},
		},
	)
}

func TestOWASPIntegerLimitLegacy(t *testing.T) {
	rule := registeredRule("owasp-integer-limit-legacy")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "OAS 3.1 doc is skipped",
			Spec: `openapi: "3.1.0"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  schemas:
    Count:
      type: integer`,
			Expect: nil,
		},
		btesting.Case{
			Name: "OAS 3.0 integer without bounds triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  schemas:
    Count:
      type: integer`,
			Expect: []btesting.Diag{
				{Code: "owasp-integer-limit-legacy", Severity: btesting.Warn},
			},
		},
	)
}

func TestOWASPAdminSecurityUnique(t *testing.T) {
	rule := registeredRule("owasp-admin-security-unique")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "no global security passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /admin/users:
    get:
      security:
        - BearerAuth: []
      responses:
        "200":
          description: ok
components:
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer`,
			Expect: nil,
		},
		btesting.Case{
			Name: "admin path with same security as global triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
security:
  - BearerAuth: []
paths:
  /admin/users:
    get:
      security:
        - BearerAuth: []
      responses:
        "200":
          description: ok
components:
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer`,
			Expect: []btesting.Diag{
				{Code: "owasp-admin-security-unique", Severity: btesting.Warn},
			},
		},
	)
}

func TestOWASPConcerningURLParameter(t *testing.T) {
	rule := registeredRule("owasp-concerning-url-parameter")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "normal parameter name passes",
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
			Expect: nil,
		},
		btesting.Case{
			Name: "redirect parameter triggers info",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /login:
    get:
      parameters:
        - name: redirect
          in: query
          schema:
            type: string
      responses:
        "302":
          description: redirect`,
			Expect: []btesting.Diag{
				{Code: "owasp-concerning-url-parameter", Severity: btesting.Info},
			},
		},
	)
}

func TestOWASPInventoryAccess(t *testing.T) {
	rule := registeredRule("owasp-inventory-access")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "no servers passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}`,
			Expect: nil,
		},
		btesting.Case{
			Name: "server without x-internal triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
servers:
  - url: https://api.example.com`,
			Expect: []btesting.Diag{
				{Code: "owasp-inventory-access", Severity: btesting.Warn},
			},
		},
	)
}

func TestOWASPInventoryEnvironment(t *testing.T) {
	rule := registeredRule("owasp-inventory-environment")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "no servers passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}`,
			Expect: nil,
		},
		btesting.Case{
			Name: "server without environment in description triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
servers:
  - url: https://api.example.com
    description: Main API server`,
			Expect: []btesting.Diag{
				{Code: "owasp-inventory-environment", Severity: btesting.Warn},
			},
		},
	)
}
