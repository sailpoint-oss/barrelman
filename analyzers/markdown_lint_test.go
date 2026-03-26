package analyzers_test

import (
	"testing"

	"github.com/sailpoint-oss/barrelman/btesting"
)

// TestDescriptionMarkdown verifies the description-markdown rule. The rule uses
// a Custom visitor to collect descriptions from the navigator index. Navigator
// may not fully preserve multiline YAML block scalars in Description.Text,
// which limits what can be tested via the btesting harness.
func TestDescriptionMarkdown(t *testing.T) {
	rule := registeredRule("description-markdown")

	btesting.Run(t, rule,
		btesting.Case{
			Name: "simple description passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
  description: A simple description without markdown issues.
paths: {}`,
			Expect: nil,
		},
	)
}

func TestDescriptionHTML(t *testing.T) {
	rule := registeredRule("description-html")

	btesting.Run(t, rule,
		btesting.Case{
			Name: "plain markdown passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
  description: |
    This is plain markdown without HTML.
paths: {}`,
			Expect: nil,
		},
		btesting.Case{
			Name: "HTML in code block passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
  description: |
    Example code:
    ` + "```" + `
    <div>This is in a code block</div>
    ` + "```" + `
paths: {}`,
			Expect: nil,
		},
		btesting.Case{
			Name: "raw HTML triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
  description: |
    This has <br> raw HTML.
paths: {}`,
			Expect: []btesting.Diag{
				{Code: "description-html", Severity: btesting.Warn},
			},
		},
	)
}
