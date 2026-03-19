package analyzers_test

import (
	"testing"

	"github.com/sailpoint-oss/barrelman/btesting"
)

func TestContactProperties_AllPresent(t *testing.T) {
	rule := registeredRule("contact-properties")

	btesting.Run(t, rule, btesting.Case{
		Name: "contact with all properties passes",
		Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
  contact:
    name: API Support
    url: https://example.com
    email: support@example.com
paths: {}`,
		Expect: nil,
	})
}

func TestContactProperties_MissingFields(t *testing.T) {
	t.Skip("requires navigator to populate Info.Contact")
	rule := registeredRule("contact-properties")

	btesting.Run(t, rule, btesting.Case{
		Name: "contact missing all fields triggers 3 warnings",
		Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
  contact: {}
paths: {}`,
		Expect: []btesting.Diag{
			{Code: "contact-properties", Severity: btesting.Warn, Message: "name"},
			{Code: "contact-properties", Severity: btesting.Warn, Message: "url"},
			{Code: "contact-properties", Severity: btesting.Warn, Message: "email"},
		},
	})
}

func TestContactProperties_NoContact(t *testing.T) {
	rule := registeredRule("contact-properties")

	btesting.Run(t, rule, btesting.Case{
		Name: "no contact object produces no diagnostics",
		Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}`,
		Expect: nil,
	})
}

func TestContactProperties_PartialFields(t *testing.T) {
	t.Skip("requires navigator to populate Info.Contact")
	rule := registeredRule("contact-properties")

	btesting.Run(t, rule, btesting.Case{
		Name: "contact with only name triggers url and email warnings",
		Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
  contact:
    name: Support
paths: {}`,
		Expect: []btesting.Diag{
			{Code: "contact-properties", Severity: btesting.Warn, Message: "url"},
			{Code: "contact-properties", Severity: btesting.Warn, Message: "email"},
		},
	})
}

func TestLicenseURL_WithURL(t *testing.T) {
	rule := registeredRule("license-url")

	btesting.Run(t, rule, btesting.Case{
		Name: "license with url passes",
		Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
  license:
    name: MIT
    url: https://opensource.org/licenses/MIT
paths: {}`,
		Expect: nil,
	})
}

func TestLicenseURL_MissingURL(t *testing.T) {
	t.Skip("requires navigator to populate Info.License")
	rule := registeredRule("license-url")

	btesting.Run(t, rule, btesting.Case{
		Name: "license without url triggers warning",
		Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
  license:
    name: MIT
paths: {}`,
		Expect: []btesting.Diag{
			{Code: "license-url", Severity: btesting.Warn, Message: "url"},
		},
	})
}

func TestLicenseURL_NoLicense(t *testing.T) {
	rule := registeredRule("license-url")

	btesting.Run(t, rule, btesting.Case{
		Name: "no license object produces no diagnostics",
		Spec: `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}`,
		Expect: nil,
	})
}
