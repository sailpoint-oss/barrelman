package analyzers_test

import (
	"testing"

	"github.com/sailpoint-oss/barrelman/btesting"
)

func TestInfoDescription(t *testing.T) {
	rule := registeredRule("info-description")

	btesting.Run(t, rule,
		btesting.Case{
			Name: "info with description passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
  description: A test API
paths: {}`,
			Expect: nil,
		},
		btesting.Case{
			Name: "info without description triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}`,
			Expect: []btesting.Diag{
				{Code: "info-description", Severity: btesting.Warn},
			},
		},
	)
}

func TestEnumDescription(t *testing.T) {
	rule := registeredRule("enum-description")

	btesting.Run(t, rule,
		btesting.Case{
			Name: "enum with description passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  schemas:
    Status:
      type: string
      description: The status of the resource
      enum:
        - active
        - inactive`,
			Expect: nil,
		},
		btesting.Case{
			Name: "enum without description triggers warning",
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
			Expect: []btesting.Diag{
				{Code: "enum-description", Severity: btesting.Warn},
			},
		},
	)
}

func TestDeprecatedDescription(t *testing.T) {
	rule := registeredRule("deprecated-description")

	btesting.Run(t, rule,
		btesting.Case{
			Name: "deprecated operation with description passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /pets:
    get:
      deprecated: true
      description: Use /animals instead.
      responses:
        "200":
          description: ok`,
			Expect: nil,
		},
		btesting.Case{
			Name: "deprecated operation without description triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /pets:
    get:
      deprecated: true
      responses:
        "200":
          description: ok`,
			Expect: []btesting.Diag{
				{Code: "deprecated-description", Severity: btesting.Warn},
			},
		},
		btesting.Case{
			Name: "deprecated schema without description triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  schemas:
    OldPet:
      type: object
      deprecated: true`,
			Expect: []btesting.Diag{
				{Code: "deprecated-description", Severity: btesting.Warn},
			},
		},
	)
}

func TestDeprecatedOperation(t *testing.T) {
	rule := registeredRule("deprecated-operation")

	btesting.Run(t, rule,
		btesting.Case{
			Name: "non-deprecated operation passes",
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
			Name: "deprecated operation triggers hint",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /pets:
    get:
      deprecated: true
      responses:
        "200":
          description: ok`,
			Expect: []btesting.Diag{
				{Code: "deprecated-operation", Severity: btesting.Hint},
			},
		},
	)
}

func TestDeprecatedSchema(t *testing.T) {
	rule := registeredRule("deprecated-schema")

	btesting.Run(t, rule,
		btesting.Case{
			Name: "non-deprecated schema passes",
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
			Name: "deprecated schema triggers hint",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  schemas:
    OldPet:
      type: object
      deprecated: true`,
			Expect: []btesting.Diag{
				{Code: "deprecated-schema", Severity: btesting.Hint},
			},
		},
	)
}
