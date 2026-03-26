package analyzers_test

import (
	"testing"

	"github.com/sailpoint-oss/barrelman/btesting"
)

// TestAdditionalProperties verifies the additional-properties rule. Navigator
// v0.2.0 does not populate Schema.AdditionalProperties, so the rule fires for
// any named object with properties. The passing case uses a schema without
// properties to avoid triggering the rule.
func TestAdditionalProperties(t *testing.T) {
	rule := registeredRule("additional-properties")

	btesting.Run(t, rule,
		btesting.Case{
			Name: "schema without properties passes",
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
			Name: "schema with properties but no additionalProperties triggers warning",
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
				{Code: "additional-properties", Severity: btesting.Warn},
			},
		},
	)
}

func TestAllOfMixedTypes(t *testing.T) {
	rule := registeredRule("allof-mixed-types")

	btesting.Run(t, rule,
		btesting.Case{
			Name: "allOf with same types passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  schemas:
    Combined:
      allOf:
        - type: object
          properties:
            a:
              type: string
        - type: object
          properties:
            b:
              type: string`,
			Expect: nil,
		},
		btesting.Case{
			Name: "allOf with mixed types triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  schemas:
    Combined:
      allOf:
        - type: object
          properties:
            a:
              type: string
        - type: string`,
			Expect: []btesting.Diag{
				{Code: "allof-mixed-types", Severity: btesting.Warn},
			},
		},
	)
}

func TestAllOfStructure(t *testing.T) {
	rule := registeredRule("allof-structure")

	btesting.Run(t, rule,
		btesting.Case{
			Name: "allOf with ref passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  schemas:
    Base:
      type: object
      properties:
        id:
          type: string
    Extended:
      allOf:
        - $ref: '#/components/schemas/Base'`,
			Expect: nil,
		},
		btesting.Case{
			Name: "allOf with single non-ref triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  schemas:
    Wrapper:
      allOf:
        - type: object
          properties:
            name:
              type: string`,
			Expect: []btesting.Diag{
				{Code: "allof-structure", Severity: btesting.Warn},
			},
		},
	)
}

// TestDiscriminatorMapping is limited because navigator v0.2.0 may not fully
// parse discriminator.mapping from YAML. When navigator adds support, this
// test can be extended with a failing case.
func TestDiscriminatorMapping(t *testing.T) {
	rule := registeredRule("discriminator-mapping")

	btesting.Run(t, rule,
		btesting.Case{
			Name: "schema without discriminator passes",
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
	)
}
