package analyzers

import (
	"strings"

	"github.com/sailpoint-oss/barrelman"
	navigator "github.com/sailpoint-oss/navigator"
)

var (
	owaspNoHTTPBasicMeta                 = barrelman.RuleMeta{ID: "owasp-no-http-basic", Description: "Security scheme should not use HTTP basic auth.", Severity: barrelman.SeverityWarning, Category: barrelman.CategoryOWASP, Recommended: false, DocURL: barrelman.DocBaseURL + "owasp-no-http-basic"}
	owaspNoAPIKeysInURLMeta              = barrelman.RuleMeta{ID: "owasp-no-api-keys-in-url", Description: "API keys should not be in query or path.", Severity: barrelman.SeverityWarning, Category: barrelman.CategoryOWASP, Recommended: false, DocURL: barrelman.DocBaseURL + "owasp-no-api-keys-in-url"}
	owaspNoCredentialsInURLMeta          = barrelman.RuleMeta{ID: "owasp-no-credentials-in-url", Description: "URLs should not contain credentials.", Severity: barrelman.SeverityError, Category: barrelman.CategoryOWASP, Recommended: false, DocURL: barrelman.DocBaseURL + "owasp-no-credentials-in-url"}
	owaspAuthInsecureSchemesMeta         = barrelman.RuleMeta{ID: "owasp-auth-insecure-schemes", Description: "Should not use insecure auth schemes (negotiate, oauth).", Severity: barrelman.SeverityWarning, Category: barrelman.CategoryOWASP, Recommended: false, DocURL: barrelman.DocBaseURL + "owasp-auth-insecure-schemes"}
	owaspJWTBestPracticesMeta            = barrelman.RuleMeta{ID: "owasp-jwt-best-practices", Description: "JWT bearer tokens should follow best practices.", Severity: barrelman.SeverityWarning, Category: barrelman.CategoryOWASP, Recommended: false, DocURL: barrelman.DocBaseURL + "owasp-jwt-best-practices"}
	owaspShortLivedAccessTokensMeta      = barrelman.RuleMeta{ID: "owasp-short-lived-access-tokens", Description: "OAuth2 flows should define refreshUrl for short-lived tokens.", Severity: barrelman.SeverityWarning, Category: barrelman.CategoryOWASP, Recommended: false, DocURL: barrelman.DocBaseURL + "owasp-short-lived-access-tokens"}
	owaspProtectionGlobalUnsafeMeta      = barrelman.RuleMeta{ID: "owasp-protection-global-unsafe", Description: "Unsafe operations should have security defined.", Severity: barrelman.SeverityWarning, Category: barrelman.CategoryOWASP, Recommended: false, DocURL: barrelman.DocBaseURL + "owasp-protection-global-unsafe"}
	owaspProtectionGlobalSafeMeta        = barrelman.RuleMeta{ID: "owasp-protection-global-safe", Description: "All operations should have some security defined.", Severity: barrelman.SeverityInfo, Category: barrelman.CategoryOWASP, Recommended: false, DocURL: barrelman.DocBaseURL + "owasp-protection-global-safe"}
	owaspDefineErrorResponses401Meta     = barrelman.RuleMeta{ID: "owasp-define-error-responses-401", Description: "Operations should define 401 responses.", Severity: barrelman.SeverityWarning, Category: barrelman.CategoryOWASP, Recommended: false, DocURL: barrelman.DocBaseURL + "owasp-define-error-responses-401"}
	owaspDefineErrorResponses500Meta     = barrelman.RuleMeta{ID: "owasp-define-error-responses-500", Description: "Operations should define 500 responses.", Severity: barrelman.SeverityWarning, Category: barrelman.CategoryOWASP, Recommended: false, DocURL: barrelman.DocBaseURL + "owasp-define-error-responses-500"}
	owaspRateLimitMeta                   = barrelman.RuleMeta{ID: "owasp-rate-limit", Description: "Responses should define rate limit headers.", Severity: barrelman.SeverityWarning, Category: barrelman.CategoryOWASP, Recommended: false, DocURL: barrelman.DocBaseURL + "owasp-rate-limit"}
	owaspRateLimitRetryAfterMeta         = barrelman.RuleMeta{ID: "owasp-rate-limit-retry-after", Description: "429 responses should include Retry-After header.", Severity: barrelman.SeverityWarning, Category: barrelman.CategoryOWASP, Recommended: false, DocURL: barrelman.DocBaseURL + "owasp-rate-limit-retry-after"}
	owaspRateLimitResponses429Meta       = barrelman.RuleMeta{ID: "owasp-rate-limit-responses-429", Description: "Operations should define a 429 Too Many Requests response.", Severity: barrelman.SeverityWarning, Category: barrelman.CategoryOWASP, Recommended: false, DocURL: barrelman.DocBaseURL + "owasp-rate-limit-responses-429"}
	owaspDefineErrorValidationMeta       = barrelman.RuleMeta{ID: "owasp-define-error-validation", Description: "Operations should define 422/400 responses for input validation.", Severity: barrelman.SeverityWarning, Category: barrelman.CategoryOWASP, Recommended: false, DocURL: barrelman.DocBaseURL + "owasp-define-error-validation"}
	owaspDefineCORSOriginMeta            = barrelman.RuleMeta{ID: "owasp-define-cors-origin", Description: "Responses should define Access-Control-Allow-Origin header.", Severity: barrelman.SeverityWarning, Category: barrelman.CategoryOWASP, Recommended: false, DocURL: barrelman.DocBaseURL + "owasp-define-cors-origin"}
	owaspNoSchemeHTTPMeta                = barrelman.RuleMeta{ID: "owasp-no-scheme-http", Description: "OAS 2.0 schemes must not include http.", Severity: barrelman.SeverityError, Category: barrelman.CategoryOWASP, Recommended: false, DocURL: barrelman.DocBaseURL + "owasp-no-scheme-http"}
	owaspNoServerHTTPMeta                = barrelman.RuleMeta{ID: "owasp-no-server-http", Description: "Server URLs must use HTTPS.", Severity: barrelman.SeverityError, Category: barrelman.CategoryOWASP, Recommended: false, DocURL: barrelman.DocBaseURL + "owasp-no-server-http"}
	owaspNoNumericIDsMeta                = barrelman.RuleMeta{ID: "owasp-no-numeric-ids", Description: "Avoid integer IDs; use UUIDs or random strings.", Severity: barrelman.SeverityWarning, Category: barrelman.CategoryOWASP, Recommended: false, DocURL: barrelman.DocBaseURL + "owasp-no-numeric-ids"}
	owaspNoAdditionalPropertiesMeta      = barrelman.RuleMeta{ID: "owasp-no-additionalProperties", Description: "Object schemas should restrict additional properties.", Severity: barrelman.SeverityWarning, Category: barrelman.CategoryOWASP, Recommended: false, DocURL: barrelman.DocBaseURL + "owasp-no-additionalProperties"}
	owaspConstrainedAdditionalPropsMeta  = barrelman.RuleMeta{ID: "owasp-constrained-additionalProperties", Description: "Additional properties should have constraints.", Severity: barrelman.SeverityWarning, Category: barrelman.CategoryOWASP, Recommended: false, DocURL: barrelman.DocBaseURL + "owasp-constrained-additionalProperties"}
	owaspNoUnevaluatedPropertiesMeta     = barrelman.RuleMeta{ID: "owasp-no-unevaluatedProperties", Description: "Object schemas should set unevaluatedProperties to false (OAS 3.1+).", Severity: barrelman.SeverityWarning, Category: barrelman.CategoryOWASP, Recommended: false, DocURL: barrelman.DocBaseURL + "owasp-no-unevaluatedProperties"}
	owaspConstrainedUnevaluatedPropsMeta = barrelman.RuleMeta{ID: "owasp-constrained-unevaluatedProperties", Description: "Unevaluated properties schema should have maxProperties.", Severity: barrelman.SeverityWarning, Category: barrelman.CategoryOWASP, Recommended: false, DocURL: barrelman.DocBaseURL + "owasp-constrained-unevaluatedProperties"}
	owaspStringLimitMeta                 = barrelman.RuleMeta{ID: "owasp-string-limit", Description: "String schemas should define maxLength.", Severity: barrelman.SeverityWarning, Category: barrelman.CategoryOWASP, Recommended: false, DocURL: barrelman.DocBaseURL + "owasp-string-limit"}
	owaspStringRestrictedMeta            = barrelman.RuleMeta{ID: "owasp-string-restricted", Description: "String schemas should specify format, pattern, enum, or const.", Severity: barrelman.SeverityWarning, Category: barrelman.CategoryOWASP, Recommended: false, DocURL: barrelman.DocBaseURL + "owasp-string-restricted"}
	owaspArrayLimitMeta                  = barrelman.RuleMeta{ID: "owasp-array-limit", Description: "Array schemas should define maxItems.", Severity: barrelman.SeverityWarning, Category: barrelman.CategoryOWASP, Recommended: false, DocURL: barrelman.DocBaseURL + "owasp-array-limit"}
	owaspIntegerLimitMeta                = barrelman.RuleMeta{ID: "owasp-integer-limit", Description: "Integer schemas should define minimum and maximum bounds (OAS 3.1+).", Severity: barrelman.SeverityWarning, Category: barrelman.CategoryOWASP, Recommended: false, DocURL: barrelman.DocBaseURL + "owasp-integer-limit"}
	owaspIntegerLimitLegacyMeta          = barrelman.RuleMeta{ID: "owasp-integer-limit-legacy", Description: "Integer schemas should define minimum and maximum (OAS 2.0/3.0).", Severity: barrelman.SeverityWarning, Category: barrelman.CategoryOWASP, Recommended: false, DocURL: barrelman.DocBaseURL + "owasp-integer-limit-legacy"}
	owaspIntegerFormatMeta               = barrelman.RuleMeta{ID: "owasp-integer-format", Description: "Integer schemas should specify format (int32 or int64).", Severity: barrelman.SeverityWarning, Category: barrelman.CategoryOWASP, Recommended: false, DocURL: barrelman.DocBaseURL + "owasp-integer-format"}
	owaspAdminSecurityUniqueMeta         = barrelman.RuleMeta{ID: "owasp-admin-security-unique", Description: "Admin endpoints should use distinct security schemes.", Severity: barrelman.SeverityWarning, Category: barrelman.CategoryOWASP, Recommended: false, DocURL: barrelman.DocBaseURL + "owasp-admin-security-unique"}
	owaspConcerningURLParameterMeta      = barrelman.RuleMeta{ID: "owasp-concerning-url-parameter", Description: "Parameters with URL-like names may be vulnerable to SSRF.", Severity: barrelman.SeverityInfo, Category: barrelman.CategoryOWASP, Recommended: false, DocURL: barrelman.DocBaseURL + "owasp-concerning-url-parameter"}
	owaspInventoryAccessMeta             = barrelman.RuleMeta{ID: "owasp-inventory-access", Description: "Server objects should declare x-internal to indicate intended audience.", Severity: barrelman.SeverityWarning, Category: barrelman.CategoryOWASP, Recommended: false, DocURL: barrelman.DocBaseURL + "owasp-inventory-access"}
	owaspInventoryEnvironmentMeta        = barrelman.RuleMeta{ID: "owasp-inventory-environment", Description: "Server descriptions should state the environment (production, staging, etc.).", Severity: barrelman.SeverityWarning, Category: barrelman.CategoryOWASP, Recommended: false, DocURL: barrelman.DocBaseURL + "owasp-inventory-environment"}
)

var rateLimitHeaders = []string{"x-ratelimit-limit", "x-rate-limit-limit", "ratelimit-limit", "ratelimit"}

func registerOWASPAnalyzers(reg *barrelman.Registry) {
	barrelman.Define("owasp-no-http-basic", owaspNoHTTPBasicMeta).SecuritySchemes(
		func(name string, ss *navigator.SecurityScheme, r *barrelman.Reporter) {
			if ss.Type == "http" && strings.EqualFold(ss.Scheme, "basic") {
				r.At(ss.Loc, "Security Scheme '%s' uses HTTP Basic; consider a stronger mechanism", name)
			}
		},
	).Register(reg)

	barrelman.Define("owasp-no-api-keys-in-url", owaspNoAPIKeysInURLMeta).SecuritySchemes(
		func(name string, ss *navigator.SecurityScheme, r *barrelman.Reporter) {
			if ss.Type == "apiKey" && (ss.In == "query" || ss.In == "path") {
				r.At(ss.Loc, "Security Scheme '%s' passes API key in %s; use header instead", name, ss.In)
			}
		},
	).Register(reg)

	barrelman.Define("owasp-no-credentials-in-url", owaspNoCredentialsInURLMeta).Servers(
		func(server *navigator.Server, r *barrelman.Reporter) {
			if containsCredentials(server.URL) {
				r.At(server.URLLoc, "Server Object URL should not contain credentials")
			}
		},
	).Register(reg)

	barrelman.Define("owasp-auth-insecure-schemes", owaspAuthInsecureSchemesMeta).SecuritySchemes(
		func(name string, ss *navigator.SecurityScheme, r *barrelman.Reporter) {
			insecure := map[string]bool{"negotiate": true, "oauth": true}
			if ss.Type == "http" && insecure[strings.ToLower(ss.Scheme)] {
				r.At(ss.Loc, "Security Scheme '%s' uses insecure scheme '%s'", name, ss.Scheme)
			}
		},
	).Register(reg)

	barrelman.Define("owasp-jwt-best-practices", owaspJWTBestPracticesMeta).SecuritySchemes(
		func(name string, ss *navigator.SecurityScheme, r *barrelman.Reporter) {
			if ss.Type == "http" && strings.EqualFold(ss.Scheme, "bearer") && ss.BearerFormat == "" {
				r.At(ss.Loc, "Security Scheme '%s' should specify bearerFormat (e.g., JWT)", name)
			}
		},
	).Register(reg)

	barrelman.Define("owasp-protection-global-unsafe", owaspProtectionGlobalUnsafeMeta).Custom(
		func(idx *navigator.Index, r *barrelman.Reporter) {
			unsafeMethods := map[string]bool{"post": true, "put": true, "patch": true, "delete": true}
			for path, item := range idx.Document.Paths {
				for _, mo := range item.Operations() {
					if unsafeMethods[mo.Method] && len(mo.Operation.Security) == 0 && len(idx.Document.Security) == 0 {
						r.At(mo.Operation.Loc, "Unsafe operation %s %s has no security", strings.ToUpper(mo.Method), path)
					}
				}
			}
		},
	).Register(reg)

	barrelman.Define("owasp-protection-global-safe", owaspProtectionGlobalSafeMeta).Custom(
		func(idx *navigator.Index, r *barrelman.Reporter) {
			if len(idx.Document.Security) > 0 {
				return
			}
			for path, item := range idx.Document.Paths {
				for _, mo := range item.Operations() {
					if len(mo.Operation.Security) == 0 {
						r.At(mo.Operation.Loc, "Operation %s %s has no security defined", strings.ToUpper(mo.Method), path)
					}
				}
			}
		},
	).Register(reg)

	registerResponseCheck := func(meta barrelman.RuleMeta, code, msg string) {
		barrelman.Define(meta.ID, meta).Operations(
			func(path, method string, op *navigator.Operation, r *barrelman.Reporter) {
				if _, ok := op.Responses[code]; !ok {
					r.At(navigator.LocOrFallback(op.ResponsesLoc, op.Loc), "%s for %s %s", msg, strings.ToUpper(method), path)
				}
			},
		).Register(reg)
	}
	registerResponseCheck(owaspDefineErrorResponses401Meta, "401", "Missing 401 Unauthorized response")
	registerResponseCheck(owaspDefineErrorResponses500Meta, "500", "Missing 500 Internal Server Error response")

	barrelman.Define("owasp-rate-limit", owaspRateLimitMeta).Operations(
		func(path, method string, op *navigator.Operation, r *barrelman.Reporter) {
			for code, resp := range op.Responses {
				if code != "200" && code != "201" {
					continue
				}
				hasRateLimit := false
				for header := range resp.Headers {
					hl := strings.ToLower(header)
					for _, rlh := range rateLimitHeaders {
						if hl == rlh {
							hasRateLimit = true
							break
						}
					}
				}
				if !hasRateLimit {
					r.At(HeaderDiagLoc(resp), "Response %s for %s %s should include rate limit headers", code, strings.ToUpper(method), path)
				}
			}
		},
	).Register(reg)

	barrelman.Define("owasp-define-error-validation", owaspDefineErrorValidationMeta).Operations(
		func(path, method string, op *navigator.Operation, r *barrelman.Reporter) {
			if op.RequestBody == nil {
				return
			}
			_, has400 := op.Responses["400"]
			_, has422 := op.Responses["422"]
			if !has400 && !has422 {
				r.At(navigator.LocOrFallback(op.ResponsesLoc, op.Loc), "Operation %s %s with request body should define 400 or 422 response", strings.ToUpper(method), path)
			}
		},
	).Register(reg)

	barrelman.Define("owasp-no-numeric-ids", owaspNoNumericIDsMeta).Schemas(
		func(name string, schema *navigator.Schema, pointer string, r *barrelman.Reporter) {
			nameLower := strings.ToLower(name)
			if (strings.HasSuffix(nameLower, "id") || strings.HasSuffix(nameLower, "_id")) &&
				(schema.Type == "integer" || schema.Type == "number") {
				r.At(schema.Loc, "%s uses numeric ID; consider UUID", InferContextFromPointer(pointer))
			}
		},
	).Register(reg)

	barrelman.Define("owasp-no-additionalProperties", owaspNoAdditionalPropertiesMeta).Schemas(
		func(name string, schema *navigator.Schema, pointer string, r *barrelman.Reporter) {
			if schema.Type == "object" && len(schema.Properties) > 0 &&
				schema.AdditionalProperties == nil && !schema.AdditionalPropertiesFalse {
				r.At(schema.Loc, "%s should restrict additionalProperties", InferContextFromPointer(pointer))
			}
		},
	).Register(reg)

	barrelman.Define("owasp-constrained-additionalProperties", owaspConstrainedAdditionalPropsMeta).Schemas(
		func(name string, schema *navigator.Schema, pointer string, r *barrelman.Reporter) {
			if schema.AdditionalProperties == nil || schema.AdditionalProperties.Type == "" {
				return
			}
			ap := schema.AdditionalProperties
			hasConstraints := ap.MaxLength != nil || ap.Maximum != nil || ap.MaxItems != nil ||
				len(ap.Enum) > 0 || ap.Pattern != ""
			if !hasConstraints {
				r.At(schema.Loc, "Additional properties in %s should have constraints (maxLength, maximum, etc.)", InferContextFromPointer(pointer))
			}
		},
	).Register(reg)

	barrelman.Define("owasp-string-limit", owaspStringLimitMeta).Schemas(
		func(name string, schema *navigator.Schema, pointer string, r *barrelman.Reporter) {
			if schema.Type == "string" && schema.MaxLength == nil && schema.Enum == nil &&
				!schema.HasConst &&
				schema.Format != "date" && schema.Format != "date-time" && schema.Format != "uuid" &&
				schema.Format != "email" && schema.Format != "uri" && schema.Format != "binary" {
				r.At(schema.Loc, "String schema in %s should define maxLength", InferContextFromPointer(pointer))
			}
		},
	).Register(reg)

	barrelman.Define("owasp-short-lived-access-tokens", owaspShortLivedAccessTokensMeta).SecuritySchemes(
		func(name string, ss *navigator.SecurityScheme, r *barrelman.Reporter) {
			if ss.Flows == nil {
				return
			}
			type namedFlow struct {
				name string
				flow *navigator.OAuthFlow
			}
			flows := []namedFlow{
				{"implicit", ss.Flows.Implicit},
				{"password", ss.Flows.Password},
				{"authorizationCode", ss.Flows.AuthorizationCode},
			}
			for _, nf := range flows {
				if nf.flow != nil && nf.flow.RefreshURL == "" {
					r.At(nf.flow.Loc, "OAuth2 %s flow in '%s' should define refreshUrl for short-lived tokens", nf.name, name)
				}
			}
		},
	).Register(reg)

	barrelman.Define("owasp-no-unevaluatedProperties", owaspNoUnevaluatedPropertiesMeta).Custom(
		func(idx *navigator.Index, r *barrelman.Reporter) {
			if idx.Document.ParsedVersion != navigator.Version31 && idx.Document.ParsedVersion != navigator.Version32 {
				return
			}
			WalkAllSchemas(idx, func(name string, schema *navigator.Schema, pointer string) {
				if schema.Type == "object" && len(schema.Properties) > 0 &&
					schema.UnevaluatedProperties == nil && !schema.UnevaluatedPropertiesFalse {
					r.At(schema.Loc, "%s should set unevaluatedProperties to false", InferContextFromPointer(pointer))
				}
			})
		},
	).Register(reg)

	barrelman.Define("owasp-constrained-unevaluatedProperties", owaspConstrainedUnevaluatedPropsMeta).Custom(
		func(idx *navigator.Index, r *barrelman.Reporter) {
			if idx.Document.ParsedVersion != navigator.Version31 && idx.Document.ParsedVersion != navigator.Version32 {
				return
			}
			WalkAllSchemas(idx, func(name string, schema *navigator.Schema, pointer string) {
				if schema.UnevaluatedProperties != nil && schema.MaxProperties == nil {
					r.At(schema.Loc, "%s with unevaluatedProperties schema should define maxProperties", InferContextFromPointer(pointer))
				}
			})
		},
	).Register(reg)

	barrelman.Define("owasp-rate-limit-retry-after", owaspRateLimitRetryAfterMeta).Operations(
		func(path, method string, op *navigator.Operation, r *barrelman.Reporter) {
			resp, ok := op.Responses["429"]
			if !ok {
				return
			}
			for header := range resp.Headers {
				if strings.EqualFold(header, "Retry-After") {
					return
				}
			}
			r.At(HeaderDiagLoc(resp), "429 response for %s %s should include Retry-After header", strings.ToUpper(method), path)
		},
	).Register(reg)

	registerResponseCheck(owaspRateLimitResponses429Meta, "429", "Missing 429 Too Many Requests response")

	barrelman.Define("owasp-array-limit", owaspArrayLimitMeta).Schemas(
		func(name string, schema *navigator.Schema, pointer string, r *barrelman.Reporter) {
			if schema.Type == "array" && schema.MaxItems == nil {
				r.At(schema.Loc, "Array schema in %s should define maxItems", InferContextFromPointer(pointer))
			}
		},
	).Register(reg)

	barrelman.Define("owasp-string-restricted", owaspStringRestrictedMeta).Schemas(
		func(name string, schema *navigator.Schema, pointer string, r *barrelman.Reporter) {
			if schema.Type == "string" && schema.Format == "" && schema.Pattern == "" &&
				schema.Enum == nil && !schema.HasConst {
				r.At(schema.Loc, "String schema in %s should specify format, pattern, enum, or const", InferContextFromPointer(pointer))
			}
		},
	).Register(reg)

	barrelman.Define("owasp-integer-limit", owaspIntegerLimitMeta).Custom(
		func(idx *navigator.Index, r *barrelman.Reporter) {
			if idx.Document.ParsedVersion != navigator.Version31 && idx.Document.ParsedVersion != navigator.Version32 {
				return
			}
			WalkAllSchemas(idx, func(name string, schema *navigator.Schema, pointer string) {
				if schema.Type != "integer" {
					return
				}
				hasLower := schema.Minimum != nil || schema.ExclusiveMinimum != nil
				hasUpper := schema.Maximum != nil || schema.ExclusiveMaximum != nil
				if !hasLower || !hasUpper {
					r.At(schema.Loc, "Integer schema in %s should define minimum/exclusiveMinimum and maximum/exclusiveMaximum", InferContextFromPointer(pointer))
				}
			})
		},
	).Register(reg)

	barrelman.Define("owasp-integer-limit-legacy", owaspIntegerLimitLegacyMeta).Custom(
		func(idx *navigator.Index, r *barrelman.Reporter) {
			if idx.Document.ParsedVersion != navigator.Version20 && idx.Document.ParsedVersion != navigator.Version30 {
				return
			}
			WalkAllSchemas(idx, func(name string, schema *navigator.Schema, pointer string) {
				if schema.Type == "integer" && (schema.Minimum == nil || schema.Maximum == nil) {
					r.At(schema.Loc, "Integer schema in %s should define minimum and maximum", InferContextFromPointer(pointer))
				}
			})
		},
	).Register(reg)

	barrelman.Define("owasp-integer-format", owaspIntegerFormatMeta).Schemas(
		func(name string, schema *navigator.Schema, pointer string, r *barrelman.Reporter) {
			if schema.Type == "integer" && schema.Format == "" {
				r.At(schema.Loc, "Integer schema in %s should specify format (int32 or int64)", InferContextFromPointer(pointer))
			}
		},
	).Register(reg)

	barrelman.Define("owasp-admin-security-unique", owaspAdminSecurityUniqueMeta).Custom(
		func(idx *navigator.Index, r *barrelman.Reporter) {
			globalSchemes := securitySchemeNames(idx.Document.Security)
			if len(globalSchemes) == 0 {
				return
			}
			for path, item := range idx.Document.Paths {
				if !isAdminPath(path) {
					continue
				}
				for _, mo := range item.Operations() {
					opSchemes := securitySchemeNames(mo.Operation.Security)
					if len(opSchemes) == 0 {
						continue
					}
					if sameSchemes(globalSchemes, opSchemes) {
						r.At(mo.Operation.Loc, "Admin operation %s %s uses the same security as non-admin endpoints", strings.ToUpper(mo.Method), path)
					}
				}
			}
		},
	).Register(reg)

	barrelman.Define("owasp-concerning-url-parameter", owaspConcerningURLParameterMeta).Parameters(
		func(param *navigator.Parameter, r *barrelman.Reporter) {
			nameLower := strings.ToLower(param.Name)
			for _, pattern := range urlParameterPatterns {
				if strings.Contains(nameLower, pattern) {
					r.At(param.NameLoc, "Parameter '%s' has a URL-like name; review for SSRF risk", param.Name)
					return
				}
			}
		},
	).Register(reg)

	barrelman.Define("owasp-define-cors-origin", owaspDefineCORSOriginMeta).Operations(
		func(path, method string, op *navigator.Operation, r *barrelman.Reporter) {
			for code, resp := range op.Responses {
				if !strings.HasPrefix(code, "2") {
					continue
				}
				hasCORS := false
				for header := range resp.Headers {
					if strings.EqualFold(header, "Access-Control-Allow-Origin") {
						hasCORS = true
						break
					}
				}
				if !hasCORS {
					r.At(HeaderDiagLoc(resp), "Response %s for %s %s should define Access-Control-Allow-Origin header", code, strings.ToUpper(method), path)
				}
			}
		},
	).Register(reg)

	barrelman.Define("owasp-no-scheme-http", owaspNoSchemeHTTPMeta).Custom(
		func(idx *navigator.Index, r *barrelman.Reporter) {
			if idx.Document.ParsedVersion != navigator.Version20 {
				return
			}
			for _, scheme := range idx.Document.Schemes {
				if strings.EqualFold(scheme, "http") {
					r.AtRange(barrelman.FileStartRange, "OAS 2.0 schemes must not include 'http'; use 'https'")
					return
				}
			}
		},
	).Register(reg)

	barrelman.Define("owasp-no-server-http", owaspNoServerHTTPMeta).Custom(
		func(idx *navigator.Index, r *barrelman.Reporter) {
			if !idx.IsOpenAPI() || idx.Document.ParsedVersion == navigator.Version20 {
				return
			}
			for _, srv := range idx.Document.Servers {
				if srv.URL == "" {
					continue
				}
				lower := strings.ToLower(srv.URL)
				if strings.HasPrefix(lower, "https://") || strings.HasPrefix(lower, "wss://") {
					continue
				}
				r.At(srv.URLLoc, "Server Object URL '%s' must use https:// or wss://", srv.URL)
			}
		},
	).Register(reg)

	barrelman.Define("owasp-inventory-access", owaspInventoryAccessMeta).Servers(
		func(server *navigator.Server, r *barrelman.Reporter) {
			if _, ok := server.Extensions["x-internal"]; !ok {
				r.At(server.Loc, "Server Object '%s' should declare x-internal: true or false", server.URL)
			}
		},
	).Register(reg)

	barrelman.Define("owasp-inventory-environment", owaspInventoryEnvironmentMeta).Servers(
		func(server *navigator.Server, r *barrelman.Reporter) {
			desc := strings.ToLower(server.Description.Text)
			for _, term := range environmentTerms {
				if strings.Contains(desc, term) {
					return
				}
			}
			r.At(server.Loc, "Server Object '%s' description should include environment (production, staging, etc.)", server.URL)
		},
	).Register(reg)
}

// WalkAllSchemas iterates every schema reachable from the document index
// (components, parameters, request bodies, responses) and invokes fn with a
// human-readable name, the schema node, and a JSON-pointer-like location
// string. Exposed so downstream rule packs can share the same traversal
// without forking it.
func WalkAllSchemas(idx *navigator.Index, fn func(name string, schema *navigator.Schema, pointer string)) {
	doc := idx.Document
	if doc.Components != nil {
		for name, schema := range doc.Components.Schemas {
			fn(name, schema, "components/schemas/"+name)
		}
	}
	for path, item := range doc.Paths {
		for _, mo := range item.Operations() {
			for _, p := range mo.Operation.Parameters {
				if p.Schema != nil {
					fn("", p.Schema, "paths"+path+"/"+mo.Method+"/parameters/"+p.Name)
				}
			}
			if mo.Operation.RequestBody != nil {
				for mt, media := range mo.Operation.RequestBody.Content {
					if media.Schema != nil {
						fn("", media.Schema, "paths"+path+"/"+mo.Method+"/requestBody/"+mt)
					}
				}
			}
			for code, resp := range mo.Operation.Responses {
				for mt, media := range resp.Content {
					if media.Schema != nil {
						fn("", media.Schema, "paths"+path+"/"+mo.Method+"/responses/"+code+"/"+mt)
					}
				}
			}
		}
	}
}

// HeaderDiagLoc returns the best diagnostic location for a header-related
// finding on resp, falling back from HeadersLoc to CodeLoc to Loc. Exposed
// for downstream rule packs.
func HeaderDiagLoc(resp *navigator.Response) navigator.Loc {
	if resp.HeadersLoc.Range != (barrelman.Range{}) {
		return resp.HeadersLoc
	}
	return navigator.LocOrFallback(resp.CodeLoc, resp.Loc)
}

var urlParameterPatterns = []string{"callback", "redirect", "_url", "-url", "returnurl", "next_url", "target"}

var environmentTerms = []string{"production", "staging", "development", "sandbox", "local", "test", "qa", "dev", "prod", "uat"}

func isAdminPath(path string) bool {
	lower := strings.ToLower(path)
	return strings.Contains(lower, "/admin") || strings.Contains(lower, "/internal")
}

func securitySchemeNames(reqs []navigator.SecurityRequirement) []string {
	seen := make(map[string]bool)
	for _, req := range reqs {
		for _, e := range req.Entries {
			seen[e.Name] = true
		}
	}
	names := make([]string, 0, len(seen))
	for n := range seen {
		names = append(names, n)
	}
	return names
}

func sameSchemes(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	set := make(map[string]bool, len(a))
	for _, s := range a {
		set[s] = true
	}
	for _, s := range b {
		if !set[s] {
			return false
		}
	}
	return true
}
