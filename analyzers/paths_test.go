package analyzers_test

import (
	"testing"

	"github.com/sailpoint-oss/barrelman/btesting"
)

func TestKebabCase(t *testing.T) {
	rule := registeredRule("kebab-case")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "kebab-case path passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /user-items:
    get:
      responses:
        "200":
          description: ok`,
			Expect: nil,
		},
		btesting.Case{
			Name: "camelCase segment triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /userItems:
    get:
      responses:
        "200":
          description: ok`,
			Expect: []btesting.Diag{
				{Code: "kebab-case", Severity: btesting.Warn},
			},
		},
		btesting.Case{
			Name: "snake_case segment triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /user_items:
    get:
      responses:
        "200":
          description: ok`,
			Expect: []btesting.Diag{
				{Code: "kebab-case", Severity: btesting.Warn},
			},
		},
	)
}

func TestNoTrailingSlash(t *testing.T) {
	rule := registeredRule("path-keys-no-trailing-slash")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "path without trailing slash passes",
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
			Name: "root path passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /:
    get:
      responses:
        "200":
          description: ok`,
			Expect: nil,
		},
		btesting.Case{
			Name: "trailing slash triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /pets/:
    get:
      responses:
        "200":
          description: ok`,
			Expect: []btesting.Diag{
				{Code: "path-keys-no-trailing-slash", Severity: btesting.Warn},
			},
		},
	)
}

func TestNoHTTPVerbs(t *testing.T) {
	rule := registeredRule("no-http-verbs")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "path without verbs passes",
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
			Name: "standalone get segment triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /pets/get:
    get:
      responses:
        "200":
          description: ok`,
			Expect: []btesting.Diag{
				{Code: "no-http-verbs", Severity: btesting.Warn},
			},
		},
		btesting.Case{
			Name: "standalone delete segment triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /pets/delete:
    get:
      responses:
        "200":
          description: ok`,
			Expect: []btesting.Diag{
				{Code: "no-http-verbs", Severity: btesting.Warn},
			},
		},
	)
}

func TestPathParams(t *testing.T) {
	rule := registeredRule("path-params")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "declared path param passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /pets/{petId}:
    get:
      parameters:
        - name: petId
          in: path
          required: true
          schema:
            type: string
      responses:
        "200":
          description: ok`,
			Expect: nil,
		},
		btesting.Case{
			Name: "undeclared path param triggers error",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /pets/{petId}:
    get:
      responses:
        "200":
          description: ok`,
			Expect: []btesting.Diag{
				{Code: "path-params", Severity: btesting.Error},
			},
		},
	)
}

func TestPathDeclarationsMustExist(t *testing.T) {
	rule := registeredRule("path-declarations-must-exist")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "balanced braces pass",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /pets/{id}:
    get:
      responses:
        "200":
          description: ok`,
			Expect: nil,
		},
		btesting.Case{
			Name: "mismatched braces trigger error",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /pets/{id:
    get:
      responses:
        "200":
          description: ok`,
			Expect: []btesting.Diag{
				{Code: "path-declarations-must-exist", Severity: btesting.Error},
			},
		},
	)
}

func TestIdUniqueInPath(t *testing.T) {
	rule := registeredRule("id-unique-in-path")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "unique param names pass",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /pets/{id}/toys/{toyId}:
    get:
      responses:
        "200":
          description: ok`,
			Expect: nil,
		},
		btesting.Case{
			Name: "duplicate param name triggers error",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /pets/{id}/toys/{id}:
    get:
      responses:
        "200":
          description: ok`,
			Expect: []btesting.Diag{
				{Code: "id-unique-in-path", Severity: btesting.Error},
			},
		},
	)
}

func TestCasingConsistency(t *testing.T) {
	rule := registeredRule("casing-consistency")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "consistent kebab-case passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /user-items:
    get:
      responses:
        "200":
          description: ok
  /pet-names:
    get:
      responses:
        "200":
          description: ok`,
			Expect: nil,
		},
		btesting.Case{
			Name: "mixed kebab and snake triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /user-items:
    get:
      responses:
        "200":
          description: ok
  /pet-names:
    get:
      responses:
        "200":
          description: ok
  /order_history:
    get:
      responses:
        "200":
          description: ok`,
			Expect: []btesting.Diag{
				{Code: "casing-consistency", Severity: btesting.Warn},
			},
		},
	)
}

func TestPathParamValuesNoGenericSyntax(t *testing.T) {
	rule := registeredRule("path-param-values-no-generic-syntax")
	btesting.Run(t, rule,
		btesting.Case{
			Name: "OpenAPI brace syntax passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /pets/{id}:
    get:
      responses:
        "200":
          description: ok`,
			Expect: nil,
		},
		btesting.Case{
			Name: "colon syntax triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /pets/:id:
    get:
      responses:
        "200":
          description: ok`,
			Expect: []btesting.Diag{
				{Code: "path-param-values-no-generic-syntax", Severity: btesting.Warn},
			},
		},
		btesting.Case{
			Name: "angle bracket syntax triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /pets/<id>:
    get:
      responses:
        "200":
          description: ok`,
			Expect: []btesting.Diag{
				{Code: "path-param-values-no-generic-syntax", Severity: btesting.Warn},
			},
		},
	)
}
