package analyzers

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sailpoint-oss/barrelman"
	"github.com/sailpoint-oss/barrelman/rulesets/bridge"
	navigator "github.com/sailpoint-oss/navigator"
)

// SailPoint analyzer registrations are grouped into per-domain files:
//
//   sailpoint_naming.go     - property, path segment, parameter, enum, tag, operationId casing
//   sailpoint_security.go   - OAuth/OAuth scope/HTTPS rules
//   sailpoint_errors.go     - #403 status codes and #404 Problem Details rules
//   sailpoint_payload.go    - nullable/required/numeric-format/identifier rules
//   sailpoint_operations.go - resource-operation-id linkage, pagination, X-Request-Id
//
// This file owns the central registration entry point, the metadata helper,
// and small shared utilities used by those per-domain files.

var (
	lowerCamelRe = regexp.MustCompile(`^[a-z][A-Za-z0-9]*$`)
	scopeNameRe  = regexp.MustCompile(`^[a-z][a-z0-9]*(?:-[a-z0-9]+)*:[a-z][a-z0-9]*(?:-[a-z0-9]+)*:[a-z][a-z0-9]*(?:-[a-z0-9]+)*$`)
	upperSnakeRe = regexp.MustCompile(`^[A-Z][A-Z0-9]*(?:_[A-Z0-9]+)*$`)
)

func registerSailPointAnalyzers(reg *barrelman.Registry) {
	// naming
	registerSailpointPropertyCamelCase(reg)
	registerSailpointPathKebabCase(reg)
	registerSailpointPathParamCamelCase(reg)
	registerSailpointQueryParamCamelCase(reg)
	registerSailpointEnumScreamingSnakeCase(reg)
	registerSailpointOperationIDCamelCase(reg)
	registerSailpointOperationIDUnique(reg)
	registerSailpointOperationSingleTag(reg)
	registerSailpointTagDocumented(reg)
	registerSailpointParameterDescription(reg)
	registerSailpointPropertyDescription(reg)
	registerSailpointParameterExample(reg)
	registerSailpointPropertyExample(reg)
	registerSailpointResponseExample(reg)

	// security
	registerSailpointOAuthScopeFormat(reg)
	registerSailpointSecurityOAuth2Required(reg)
	registerSailpointOperationSecurityRequired(reg)
	registerSailpointOAuthScopesDeclared(reg)
	registerSailpointServerURLHTTPS(reg)

	// errors (#403 + #404)
	registerSailpointOperationMethodStatusCodes(reg)
	registerSailpointOperation4xxResponse(reg)
	registerSailpointOperation401Response(reg)
	registerSailpointOperation403Response(reg)
	registerSailpointErrorProblemDetailsMediaType(reg)
	registerSailpointErrorProblemDetailsSchema(reg)
	registerSailpointErrorProblemDetailsSharedComponent(reg)
	registerSailpointErrorCorrelationID(reg)

	// payload
	registerSailpointResponseTopLevelObject(reg)
	registerSailpointPathNoAPIPrefix(reg)
	registerSailpointPathParamNoNumericID(reg)
	registerSailpointBooleanNotNullable(reg)
	registerSailpointArrayNotNullable(reg)
	registerSailpointSchemaRequiredFields(reg)
	registerSailpointPathParamRequired(reg)
	registerSailpointNumericFormatApproved(reg)
	registerSailpointIdentifierStringType(reg)

	// operations
	registerSailpointPathParamResourceOperationLink(reg)
	registerSailpointCollectionOffsetPagination(reg)
	registerSailpointCollectionWrappedResponse(reg)
	registerSailpointXRequestIDHeader(reg)
	registerSailpointXRequestIDSharedComponent(reg)
	registerSailpointXRequestIDUUID(reg)
}

// sailpointMeta builds a RuleMeta for a canonical SailPoint rule. The
// guideline number drives the doc URL; the rule slug is independent.
func sailpointMeta(slug string, severity barrelman.Severity, category barrelman.Category, description, howToFix string) barrelman.RuleMeta {
	entry, ok := bridge.FromCanonical(slug)
	if !ok {
		panic(fmt.Sprintf("sailpointMeta: canonical slug %q is not registered in the rule bridge", slug))
	}
	primary := entry.PrimaryGuideline()
	meta := barrelman.RuleMeta{
		ID:           slug,
		Description:  description,
		Severity:     severity,
		Category:     category,
		Recommended:  true,
		HowToFix:     howToFix,
		GuidelineID:  primary,
		GuidelineIDs: append([]int(nil), entry.GuidelineIDs...),
		VacuumID:     entry.Vacuum,
		SpectralID:   entry.Spectral,
	}
	if primary > 0 {
		meta.DocURL = barrelman.GuidelineDocURL(fmt.Sprintf("%d", primary))
	}
	return meta
}

// Shared helpers used by multiple per-domain SailPoint files.

func operationParameters(item *navigator.PathItem, op *navigator.Operation) []*navigator.Parameter {
	params := make([]*navigator.Parameter, 0, len(item.Parameters)+len(op.Parameters))
	params = append(params, item.Parameters...)
	params = append(params, op.Parameters...)
	return params
}

func effectiveSecurity(doc *navigator.Document, op *navigator.Operation) []navigator.SecurityRequirement {
	if len(op.Security) > 0 {
		return op.Security
	}
	return doc.Security
}

func oauthScopes(ss *navigator.SecurityScheme) map[string]bool {
	out := make(map[string]bool)
	if ss == nil || ss.Flows == nil {
		return out
	}
	flows := []*navigator.OAuthFlow{
		ss.Flows.Implicit,
		ss.Flows.Password,
		ss.Flows.ClientCredentials,
		ss.Flows.AuthorizationCode,
	}
	for _, flow := range flows {
		if flow == nil {
			continue
		}
		for scope := range flow.Scopes {
			out[scope] = true
		}
	}
	return out
}

func resolveSchema(idx *navigator.Index, schema *navigator.Schema) *navigator.Schema {
	if schema == nil {
		return nil
	}
	if schema.Ref == "" || idx == nil {
		return schema
	}
	resolved, err := idx.Resolve(schema.Ref)
	if err != nil {
		return schema
	}
	if actual, ok := resolved.(*navigator.Schema); ok {
		return actual
	}
	return schema
}

func resolveHeader(idx *navigator.Index, header *navigator.Header) *navigator.Header {
	if header == nil {
		return nil
	}
	if header.Ref == "" || idx == nil {
		return header
	}
	resolved, err := idx.Resolve(header.Ref)
	if err != nil {
		return header
	}
	if actual, ok := resolved.(*navigator.Header); ok {
		return actual
	}
	return header
}

func resolveParameter(idx *navigator.Index, param *navigator.Parameter) *navigator.Parameter {
	if param == nil {
		return nil
	}
	if param.Ref == "" || idx == nil {
		return param
	}
	resolved, err := idx.Resolve(param.Ref)
	if err != nil {
		return param
	}
	if actual, ok := resolved.(*navigator.Parameter); ok {
		return actual
	}
	return param
}

func firstSuccessSchema(idx *navigator.Index, op *navigator.Operation) *navigator.Schema {
	for code, resp := range op.Responses {
		if !strings.HasPrefix(code, "2") {
			continue
		}
		for _, media := range resp.Content {
			return resolveSchema(idx, media.Schema)
		}
	}
	return nil
}

func hasParam(params []*navigator.Parameter, in, name string) bool {
	for _, param := range params {
		if param != nil && param.In == in && strings.EqualFold(param.Name, name) {
			return true
		}
	}
	return false
}

func isValidScopeName(scope string) bool {
	return scopeNameRe.MatchString(strings.TrimSpace(scope))
}

func declaredOAuthScopes(ss *navigator.SecurityScheme) []string {
	seen := make(map[string]bool)
	var out []string
	if ss == nil || ss.Type != "oauth2" || ss.Flows == nil {
		return out
	}
	flows := []*navigator.OAuthFlow{
		ss.Flows.Implicit,
		ss.Flows.Password,
		ss.Flows.ClientCredentials,
		ss.Flows.AuthorizationCode,
	}
	for _, flow := range flows {
		if flow == nil {
			continue
		}
		for scope := range flow.Scopes {
			if seen[scope] {
				continue
			}
			seen[scope] = true
			out = append(out, scope)
		}
	}
	return out
}

func schemaNeedsRequiredArray(schema *navigator.Schema) bool {
	if schema == nil || schema.Type != "object" || len(schema.Properties) == 0 {
		return false
	}
	return !schema.HasRequired
}

func isLikelyCollectionCreate(path string) bool {
	if pathEndsWithParam(path) {
		return false
	}
	segment := lastPathSegment(path)
	return looksPluralCollection(segment)
}

func pathEndsWithParam(path string) bool {
	segment := lastPathSegment(path)
	return strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}")
}

func lastPathSegment(path string) string {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return ""
	}
	parts := strings.Split(trimmed, "/")
	return parts[len(parts)-1]
}

func looksPluralCollection(segment string) bool {
	switch strings.ToLower(segment) {
	case "", "data", "metadata":
		return segment != ""
	}
	return strings.HasSuffix(segment, "s") || strings.HasSuffix(segment, "ies") || strings.HasSuffix(segment, "ren")
}

func hasParameterExample(param *navigator.Parameter) bool {
	return param.Example != nil || len(param.Examples) > 0
}

func hasMediaExample(media *navigator.MediaType) bool {
	return media != nil && (media.Example != nil || len(media.Examples) > 0)
}

func hasResponseHeader(resp *navigator.Response, name string) bool {
	for headerName := range resp.Headers {
		if strings.EqualFold(headerName, name) {
			return true
		}
	}
	return false
}

func pathParameterRuleKey(param, resolved *navigator.Parameter, path string) string {
	if param != nil && param.Ref != "" {
		return "ref:" + param.Ref
	}
	target := resolved
	if target == nil {
		target = param
	}
	if target == nil {
		return "path:" + path
	}
	return fmt.Sprintf("%s:%s:%d:%d", path, target.Name, target.Loc.Range.Start.Line, target.Loc.Range.Start.Character)
}

func responseHeader(resp *navigator.Response, name string) *navigator.Header {
	for headerName, header := range resp.Headers {
		if strings.EqualFold(headerName, name) {
			return header
		}
	}
	return nil
}

func requestIDHeaderHasUUIDFormat(idx *navigator.Index, header *navigator.Header) bool {
	resolved := resolveHeader(idx, header)
	if resolved == nil {
		return false
	}
	schema := resolveSchema(idx, resolved.Schema)
	return schema != nil && schema.Type == "string" && schema.Format == "uuid"
}

func hasSharedRequestIDHeaderComponent(doc *navigator.Document) bool {
	if doc == nil || doc.Components == nil {
		return false
	}
	header, ok := doc.Components.Headers["X-Request-Id"]
	if !ok || header == nil || header.Schema == nil {
		return false
	}
	return header.Schema.Type == "string" && header.Schema.Format == "uuid"
}

func hasProblemDetailsComponent(doc *navigator.Document) bool {
	if doc == nil || doc.Components == nil {
		return false
	}
	_, ok := doc.Components.Schemas["ProblemDetails"]
	return ok
}

func propertyLoc(prop, parent *navigator.Schema) navigator.Loc {
	if prop != nil && prop.NameLoc.Range != (barrelman.Range{}) {
		return prop.NameLoc
	}
	if prop != nil && prop.Loc.Range != (barrelman.Range{}) {
		return prop.Loc
	}
	if parent != nil {
		return parent.Loc
	}
	return navigator.Loc{Range: barrelman.FileStartRange}
}

func pointerForProperty(pointer, prop string) string {
	base := strings.TrimPrefix(pointer, "/")
	if base == "" {
		return prop
	}
	return strings.TrimLeft(base, "/") + "/properties/" + prop
}

func isExtensionName(name string) bool {
	return strings.HasPrefix(name, "x-")
}

func isLowerCamelCase(value string) bool {
	return lowerCamelRe.MatchString(value)
}

func looksLikeID(value string) bool {
	lower := strings.ToLower(value)
	return lower == "id" || strings.HasSuffix(lower, "id") || strings.HasSuffix(lower, "_id")
}

func nameFromPointer(pointer, fallback string) string {
	if fallback != "" {
		return fallback
	}
	parts := strings.Split(strings.Trim(pointer, "/"), "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

func isErrorOrSuccessCode(code string) bool {
	return strings.HasPrefix(code, "2") || strings.HasPrefix(code, "4") || strings.HasPrefix(code, "5")
}

var knownStatusCodes = func() map[string]bool {
	codes := []string{
		"100", "101", "102", "103",
		"200", "201", "202", "203", "204", "205", "206", "207", "208", "226",
		"300", "301", "302", "303", "304", "305", "307", "308",
		"400", "401", "402", "403", "404", "405", "406", "407", "408", "409",
		"410", "411", "412", "413", "414", "415", "416", "417", "418", "421",
		"422", "423", "424", "425", "426", "428", "429", "431", "451",
		"500", "501", "502", "503", "504", "505", "506", "507", "508", "510", "511",
	}
	out := make(map[string]bool, len(codes))
	for _, code := range codes {
		out[code] = true
	}
	return out
}()
