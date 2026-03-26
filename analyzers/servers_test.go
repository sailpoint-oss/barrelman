package analyzers_test

import (
	"testing"

	"github.com/sailpoint-oss/barrelman/btesting"
)

func TestOAS3APIServers(t *testing.T) {
	rule := registeredRule("oas3-api-servers")

	btesting.Run(t, rule,
		btesting.Case{
			Name: "spec with servers passes",
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
			Name: "spec without servers triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}`,
			Expect: []btesting.Diag{
				{Code: "oas3-api-servers", Severity: btesting.Warn},
			},
		},
	)
}

func TestServerURLHTTPS(t *testing.T) {
	rule := registeredRule("server-url-https")

	btesting.Run(t, rule,
		btesting.Case{
			Name: "https server passes",
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
			Name: "http server triggers warning",
			Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
servers:
  - url: http://api.example.com`,
			Expect: []btesting.Diag{
				{Code: "server-url-https", Severity: btesting.Warn},
			},
		},
	)
}
