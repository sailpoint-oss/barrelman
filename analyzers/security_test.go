package analyzers_test

import (
	"testing"

	"github.com/sailpoint-oss/barrelman/btesting"
)

func TestNoAPIKeyInQuery(t *testing.T) {
	rule := registeredRule("no-api-key-in-query")

	btesting.Run(t, rule,
		btesting.Case{
			Name: "header apiKey passes",
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
			Name: "query apiKey triggers warning",
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
				{Code: "no-api-key-in-query", Severity: btesting.Warn},
			},
		},
	)
}

func TestSecurityGlobalOrOperation(t *testing.T) {
	rule := registeredRule("security-global-or-operation")

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
			Name: "no security anywhere triggers warning",
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
				{Code: "security-global-or-operation", Severity: btesting.Warn},
			},
		},
	)
}

func TestSecuritySchemesDefined(t *testing.T) {
	rule := registeredRule("security-schemes-defined")

	btesting.Run(t, rule,
		btesting.Case{
			Name: "defined scheme passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
security:
  - BearerAuth: []
paths: {}
components:
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer`,
			Expect: nil,
		},
		btesting.Case{
			Name: "undefined scheme triggers error",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
security:
  - NonExistent: []
paths: {}`,
			Expect: []btesting.Diag{
				{Code: "security-schemes-defined", Severity: btesting.Error},
			},
		},
	)
}
