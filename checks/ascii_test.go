package checks_test

import (
	"testing"

	"github.com/sailpoint-oss/barrelman"
	"github.com/sailpoint-oss/barrelman/btesting"
	"github.com/sailpoint-oss/barrelman/checks"
)

func registeredCheck(id string) barrelman.Rule {
	reg := barrelman.NewRegistry()
	checks.RegisterAll(reg)
	for _, r := range reg.AllRules() {
		if r.ID == id {
			return r
		}
	}
	panic("check not found: " + id)
}

func TestASCII(t *testing.T) {
	rule := registeredCheck("ascii")

	btesting.Run(t, rule,
		btesting.Case{
			Name: "pure ASCII passes",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}`,
			Expect: nil,
		},
		btesting.Case{
			Name: "non-ASCII character triggers warning",
			Spec: "openapi: \"3.0.3\"\ninfo:\n  title: Caf\u00e9\n  version: \"1.0\"\npaths: {}",
			Expect: []btesting.Diag{
				{Line: 2, Code: "ascii", Severity: btesting.Warn},
			},
		},
		btesting.Case{
			Name: "emoji triggers warning",
			Spec: "openapi: \"3.0.3\"\ninfo:\n  title: Test \xf0\x9f\x90\xb1\n  version: \"1.0\"\npaths: {}",
			Expect: []btesting.Diag{
				{Line: 2, Code: "ascii", Severity: btesting.Warn},
			},
		},
	)
}
