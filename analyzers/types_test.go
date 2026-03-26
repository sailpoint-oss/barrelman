package analyzers_test

import (
	"testing"

	"github.com/sailpoint-oss/barrelman/btesting"
)

func TestNoUnknownFormats(t *testing.T) {
	rule := registeredRule("no-unknown-formats")

	btesting.Run(t, rule,
		btesting.Case{
			Name: "known format passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: string
          format: uuid`,
			Expect: nil,
		},
		btesting.Case{
			Name: "unknown format triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
components:
  schemas:
    Data:
      type: string
      format: custom-thing`,
			Expect: []btesting.Diag{
				{Code: "no-unknown-formats", Severity: btesting.Warn},
			},
		},
	)
}
