package analyzers

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sailpoint-oss/barrelman"
	navigator "github.com/sailpoint-oss/navigator"
)

var (
	lowerCamelRe = regexp.MustCompile(`^[a-z][A-Za-z0-9]*$`)
	scopeNameRe  = regexp.MustCompile(`^[a-z][a-z0-9]*(?:-[a-z0-9]+)*:[a-z][a-z0-9]*(?:-[a-z0-9]+)*:[a-z][a-z0-9]*(?:-[a-z0-9]+)*$`)
	upperSnakeRe = regexp.MustCompile(`^[A-Z][A-Z0-9]*(?:_[A-Z0-9]+)*$`)
)

func registerSailPointAnalyzers(reg *barrelman.Registry) {
	registerSP104PropertyCamelCase(reg)
	registerSP107PathConventions(reg)
	registerSP108QueryParamCamelCase(reg)
	registerSP111ScopeNaming(reg)
	registerSP112EnumCase(reg)
	registerSP115Descriptions(reg)
	registerSP116Examples(reg)
	registerSP122OperationIDs(reg)
	registerSP123Tags(reg)
	registerSP124ResourceOperationID(reg)
	registerSP204SuccessBodies(reg)
	registerSP300OAuthSecurity(reg)
	registerSP301OAuthScopes(reg)
	registerSP304HTTPS(reg)
	registerSP403StatusCodes(reg)
	registerSP404ProblemDetails(reg)
	registerSP500NoAPIBasePath(reg)
	registerSP514NonNumericPathIDs(reg)
	registerSP602Pagination(reg)
	registerSP710RequiredFields(reg)
	registerSP701NoNullableBooleans(reg)
	registerSP702NoNullableArrays(reg)
	registerSP804NumericTypes(reg)
	registerSP903RequestIDHeader(reg)
}

func sailpointMeta(id int, description string, severity barrelman.Severity, category barrelman.Category, howToFix string) barrelman.RuleMeta {
	code := barrelman.NormalizeGuidelineCode(fmt.Sprintf("%d", id))
	return barrelman.RuleMeta{
		ID:          code,
		Description: description,
		Severity:    severity,
		Category:    category,
		Recommended: true,
		HowToFix:    howToFix,
		DocURL:      barrelman.GuidelineDocURL(code),
	}
}

func registerSP104PropertyCamelCase(reg *barrelman.Registry) {
	meta := sailpointMeta(104,
		"JSON object property names must use lowerCamelCase.",
		barrelman.SeverityError,
		barrelman.CategoryNaming,
		"Rename schema properties to lowerCamelCase.",
	)

	barrelman.Define(meta.ID, meta).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		walkAllSchemas(idx, func(name string, schema *navigator.Schema, pointer string) {
			for propName, prop := range schema.Properties {
				if isExtensionName(propName) || isLowerCamelCase(propName) {
					continue
				}
				loc := propertyLoc(prop, schema)
				r.At(loc, "[#104] Property '%s' at %s must use lowerCamelCase", propName, pointerForProperty(pointer, propName))
			}
		})
	}).Register(reg)
}

func registerSP107PathConventions(reg *barrelman.Registry) {
	meta := sailpointMeta(107,
		"Paths must use lowercase hyphenated segments and lowerCamelCase path parameters.",
		barrelman.SeverityError,
		barrelman.CategoryPaths,
		"Use lowercase hyphenated resource segments and lowerCamelCase path parameter names.",
	)

	barrelman.Define(meta.ID, meta).Paths(func(path string, item *navigator.PathItem, r *barrelman.Reporter) {
		for _, seg := range barrelman.NonParamSegments(path) {
			if barrelman.IsKebabCase(seg) {
				continue
			}
			r.At(item.PathLoc, "[#107] Path segment '%s' in %s must use lowercase hyphenated form", seg, path)
		}
		for _, param := range barrelman.ExtractPathParams(path) {
			if isLowerCamelCase(param) {
				continue
			}
			r.At(item.PathLoc, "[#107] Path parameter '{%s}' in %s must use lowerCamelCase", param, path)
		}
	}).Register(reg)
}

func registerSP108QueryParamCamelCase(reg *barrelman.Registry) {
	meta := sailpointMeta(108,
		"Query parameter names must use lowerCamelCase.",
		barrelman.SeverityError,
		barrelman.CategoryPaths,
		"Rename query parameters to lowerCamelCase.",
	)

	barrelman.Define(meta.ID, meta).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		for path, item := range idx.Document.Paths {
			for _, mo := range item.Operations() {
				for _, param := range operationParameters(item, mo.Operation) {
					if param == nil || param.Ref != "" || param.In != "query" || isLowerCamelCase(param.Name) {
						continue
					}
					r.At(param.NameLoc, "[#108] Query parameter '%s' in %s %s must use lowerCamelCase", param.Name, strings.ToUpper(mo.Method), path)
				}
			}
		}
	}).Register(reg)
}

func registerSP111ScopeNaming(reg *barrelman.Registry) {
	meta := sailpointMeta(111,
		"OAuth scopes must use lower-case <domain>:<resource>:<action> names.",
		barrelman.SeverityError,
		barrelman.CategorySecurity,
		"Rename scopes to lower-case colon-separated domain, resource, and action segments.",
	)

	barrelman.Define(meta.ID, meta).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		for name, ss := range idx.SecuritySchemes {
			for _, scope := range declaredOAuthScopes(ss) {
				if isValidScopeName(scope) {
					continue
				}
				r.At(navigator.LocOrFallback(ss.NameLoc, ss.Loc), "[#111] OAuth scope '%s' declared by security scheme '%s' must use lower-case <domain>:<resource>:<action> naming", scope, name)
			}
		}
		for path, item := range idx.Document.Paths {
			for _, mo := range item.Operations() {
				for _, req := range effectiveSecurity(idx.Document, mo.Operation) {
					for _, entry := range req.Entries {
						if ss, ok := idx.SecuritySchemes[entry.Name]; ok && ss != nil && ss.Type != "oauth2" {
							continue
						}
						for i, scope := range entry.Scopes {
							if isValidScopeName(scope) {
								continue
							}
							loc := entry.NameLoc
							if i < len(entry.ScopeLocs) {
								loc = navigator.LocOrFallback(entry.ScopeLocs[i], entry.NameLoc)
							}
							r.At(loc, "[#111] OAuth scope '%s' on %s %s must use lower-case <domain>:<resource>:<action> naming", scope, strings.ToUpper(mo.Method), path)
						}
					}
				}
			}
		}
	}).Register(reg)
}

func registerSP112EnumCase(reg *barrelman.Registry) {
	meta := sailpointMeta(112,
		"String enum values must use UPPER_SNAKE_CASE.",
		barrelman.SeverityError,
		barrelman.CategoryTypes,
		"Rename string enum values to UPPER_SNAKE_CASE.",
	)

	barrelman.Define(meta.ID, meta).RecursiveSchemas(func(name string, schema *navigator.Schema, pointer string, r *barrelman.Reporter) {
		if schema.Type != "string" {
			return
		}
		for _, value := range schema.Enum {
			if upperSnakeRe.MatchString(value) {
				continue
			}
			r.At(schema.Loc, "[#112] Enum value '%s' at %s must use UPPER_SNAKE_CASE", value, pointer)
			return
		}
	}).Register(reg)
}

func registerSP115Descriptions(reg *barrelman.Registry) {
	meta := sailpointMeta(115,
		"Parameters and schema properties must include descriptions.",
		barrelman.SeverityError,
		barrelman.CategoryDocumentation,
		"Add clear descriptions to every parameter and schema property.",
	)

	barrelman.Define(meta.ID, meta).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		for path, item := range idx.Document.Paths {
			for _, mo := range item.Operations() {
				for _, param := range operationParameters(item, mo.Operation) {
					if param == nil || param.Ref != "" || strings.TrimSpace(param.Description.Text) != "" {
						continue
					}
					r.At(param.Loc, "[#115] Parameter '%s' in %s %s must include a description", param.Name, strings.ToUpper(mo.Method), path)
				}
			}
		}
		walkAllSchemas(idx, func(name string, schema *navigator.Schema, pointer string) {
			for propName, prop := range schema.Properties {
				if prop == nil || prop.Ref != "" || strings.TrimSpace(prop.Description.Text) != "" {
					continue
				}
				loc := propertyLoc(prop, schema)
				r.At(loc, "[#115] Property '%s' at %s must include a description", propName, pointerForProperty(pointer, propName))
			}
		})
	}).Register(reg)
}

func registerSP116Examples(reg *barrelman.Registry) {
	meta := sailpointMeta(116,
		"Parameters, schema properties, and response payloads must include examples.",
		barrelman.SeverityError,
		barrelman.CategoryDocumentation,
		"Add representative examples to parameters, schema properties, and payload media types.",
	)

	barrelman.Define(meta.ID, meta).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		for path, item := range idx.Document.Paths {
			for _, mo := range item.Operations() {
				for _, param := range operationParameters(item, mo.Operation) {
					if param == nil || param.Ref != "" || hasParameterExample(param) {
						continue
					}
					r.At(param.Loc, "[#116] Parameter '%s' in %s %s must include an example", param.Name, strings.ToUpper(mo.Method), path)
				}
				for code, resp := range mo.Operation.Responses {
					if !isErrorOrSuccessCode(code) {
						continue
					}
					for mediaType, media := range resp.Content {
						if media == nil || hasMediaExample(media) || (media.Schema != nil && media.Schema.Example != nil) {
							continue
						}
						r.At(resp.Loc, "[#116] Response %s for %s %s must include an example for %s", code, strings.ToUpper(mo.Method), path, mediaType)
					}
				}
			}
		}
		walkAllSchemas(idx, func(name string, schema *navigator.Schema, pointer string) {
			for propName, prop := range schema.Properties {
				if prop == nil || prop.Ref != "" || prop.Example != nil {
					continue
				}
				loc := propertyLoc(prop, schema)
				r.At(loc, "[#116] Property '%s' at %s must include an example", propName, pointerForProperty(pointer, propName))
			}
		})
	}).Register(reg)
}

func registerSP122OperationIDs(reg *barrelman.Registry) {
	meta := sailpointMeta(122,
		"Operations must declare unique lowerCamelCase operationIds.",
		barrelman.SeverityError,
		barrelman.CategoryNaming,
		"Add a unique lowerCamelCase operationId to every operation.",
	)

	barrelman.Define(meta.ID, meta).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		type operationIDLocation struct {
			loc  navigator.Loc
			desc string
		}
		seen := make(map[string]operationIDLocation)
		for path, item := range idx.Document.Paths {
			for _, mo := range item.Operations() {
				op := mo.Operation
				method := strings.ToUpper(mo.Method)
				if strings.TrimSpace(op.OperationID) == "" {
					r.At(op.Loc, "[#122] Operation %s %s must declare an operationId", method, path)
					continue
				}
				if !isLowerCamelCase(op.OperationID) {
					r.At(op.OperationIDLoc, "[#122] operationId '%s' for %s %s must use lowerCamelCase", op.OperationID, method, path)
				}
				location := method + " " + path
				if first, ok := seen[op.OperationID]; ok {
					r.WithRelated(first.loc, "", "[#122] First defined here at %s", first.desc).
						At(op.OperationIDLoc, "[#122] operationId '%s' is already used at %s", op.OperationID, first.desc)
					continue
				}
				seen[op.OperationID] = operationIDLocation{
					loc:  op.OperationIDLoc,
					desc: location,
				}
			}
		}
	}).Register(reg)
}

func registerSP123Tags(reg *barrelman.Registry) {
	meta := sailpointMeta(123,
		"Each operation must declare exactly one documented tag.",
		barrelman.SeverityError,
		barrelman.CategoryDocumentation,
		"Declare exactly one root tag per operation and give each root tag a description.",
	)

	barrelman.Define(meta.ID, meta).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		rootTags := make(map[string]navigator.Tag, len(idx.Document.Tags))
		for _, tag := range idx.Document.Tags {
			rootTags[tag.Name] = tag
			if strings.TrimSpace(tag.Description.Text) == "" {
				r.At(tag.Loc, "[#123] Root tag '%s' must include a description", tag.Name)
			}
		}
		for path, item := range idx.Document.Paths {
			for _, mo := range item.Operations() {
				op := mo.Operation
				method := strings.ToUpper(mo.Method)
				if len(op.Tags) != 1 {
					r.At(navigator.LocOrFallback(op.TagsLoc, op.Loc), "[#123] Operation %s %s must declare exactly one tag", method, path)
					continue
				}
				tagName := op.Tags[0].Name
				rootTag, ok := rootTags[tagName]
				if !ok {
					r.At(op.Tags[0].Loc, "[#123] Operation tag '%s' on %s %s must be declared in the root tags section", tagName, method, path)
					continue
				}
				if strings.TrimSpace(rootTag.Description.Text) == "" {
					r.At(rootTag.Loc, "[#123] Root tag '%s' must include a description", tagName)
				}
			}
		}
	}).Register(reg)
}

func registerSP124ResourceOperationID(reg *barrelman.Registry) {
	meta := sailpointMeta(124,
		"Path parameters must declare x-sailpoint-resource-operation-id values that reference existing operationIds.",
		barrelman.SeverityError,
		barrelman.CategoryDocumentation,
		"Add x-sailpoint-resource-operation-id to every path parameter using a valid lowerCamelCase operationId.",
	)

	barrelman.Define(meta.ID, meta).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		operationIDs := make(map[string]bool)
		for _, item := range idx.Document.Paths {
			for _, mo := range item.Operations() {
				if strings.TrimSpace(mo.Operation.OperationID) != "" {
					operationIDs[mo.Operation.OperationID] = true
				}
			}
		}

		seen := make(map[string]bool)
		for path, item := range idx.Document.Paths {
			for _, mo := range item.Operations() {
				for _, param := range operationParameters(item, mo.Operation) {
					resolved := resolveParameter(idx, param)
					if resolved == nil || resolved.In != "path" {
						continue
					}
					key := pathParameterRuleKey(param, resolved, path)
					if seen[key] {
						continue
					}
					seen[key] = true

					ext := resolved.Extensions["x-sailpoint-resource-operation-id"]
					if ext == nil || strings.TrimSpace(ext.Value) == "" {
						r.At(navigator.LocOrFallback(resolved.NameLoc, resolved.Loc), "[#124] Path parameter '%s' in %s must declare x-sailpoint-resource-operation-id", resolved.Name, path)
						continue
					}
					value := strings.TrimSpace(ext.Value)
					if !isLowerCamelCase(value) {
						r.At(ext.Loc, "[#124] x-sailpoint-resource-operation-id '%s' for path parameter '%s' in %s must use lowerCamelCase", value, resolved.Name, path)
					}
					if !operationIDs[value] {
						r.At(ext.Loc, "[#124] x-sailpoint-resource-operation-id '%s' for path parameter '%s' in %s must reference an existing operationId", value, resolved.Name, path)
					}
				}
			}
		}
	}).Register(reg)
}

func registerSP204SuccessBodies(reg *barrelman.Registry) {
	meta := sailpointMeta(204,
		"Successful JSON responses must return top-level objects, not arrays.",
		barrelman.SeverityError,
		barrelman.CategoryStructure,
		"Wrap collections in an object envelope instead of returning a top-level array.",
	)

	barrelman.Define(meta.ID, meta).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		for path, item := range idx.Document.Paths {
			for _, mo := range item.Operations() {
				for code, resp := range mo.Operation.Responses {
					if !strings.HasPrefix(code, "2") {
						continue
					}
					for mediaType, media := range resp.Content {
						if media == nil || !strings.Contains(strings.ToLower(mediaType), "json") {
							continue
						}
						schema := resolveSchema(idx, media.Schema)
						if schema == nil || schema.Type != "array" {
							continue
						}
						r.At(resp.Loc, "[#204] Response %s for %s %s must return a top-level object instead of an array", code, strings.ToUpper(mo.Method), path)
					}
				}
			}
		}
	}).Register(reg)
}

func registerSP300OAuthSecurity(reg *barrelman.Registry) {
	meta := sailpointMeta(300,
		"Security must use OAuth 2.0 and every operation must declare security requirements.",
		barrelman.SeverityError,
		barrelman.CategorySecurity,
		"Define OAuth 2.0 security schemes and apply security requirements to every operation.",
	)

	barrelman.Define(meta.ID, meta).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		if len(idx.SecuritySchemes) == 0 {
			r.AtRange(barrelman.FileStartRange, "[#300] The API must define at least one OAuth 2.0 security scheme")
		}
		for name, ss := range idx.SecuritySchemes {
			if ss == nil || ss.Ref != "" {
				continue
			}
			if ss.Type != "oauth2" {
				r.At(ss.Loc, "[#300] Security scheme '%s' must use type oauth2", name)
			}
		}
		for path, item := range idx.Document.Paths {
			for _, mo := range item.Operations() {
				reqs := effectiveSecurity(idx.Document, mo.Operation)
				if len(reqs) == 0 {
					r.At(mo.Operation.Loc, "[#300] Operation %s %s must declare security requirements", strings.ToUpper(mo.Method), path)
				}
			}
		}
	}).Register(reg)
}

func registerSP301OAuthScopes(reg *barrelman.Registry) {
	meta := sailpointMeta(301,
		"OAuth 2.0 security requirements must declare valid scopes.",
		barrelman.SeverityError,
		barrelman.CategorySecurity,
		"List one or more valid OAuth scopes on each security requirement entry.",
	)

	barrelman.Define(meta.ID, meta).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		for path, item := range idx.Document.Paths {
			for _, mo := range item.Operations() {
				for _, req := range effectiveSecurity(idx.Document, mo.Operation) {
					for _, entry := range req.Entries {
						ss, ok := idx.SecuritySchemes[entry.Name]
						if !ok || ss == nil || ss.Type != "oauth2" {
							continue
						}
						if len(entry.Scopes) == 0 {
							r.At(entry.NameLoc, "[#301] OAuth requirement '%s' on %s %s must declare at least one scope", entry.Name, strings.ToUpper(mo.Method), path)
							continue
						}
						allowed := oauthScopes(ss)
						for _, scope := range entry.Scopes {
							if allowed[scope] {
								continue
							}
							r.At(entry.NameLoc, "[#301] OAuth scope '%s' on %s %s is not defined by security scheme '%s'", scope, strings.ToUpper(mo.Method), path, entry.Name)
						}
					}
				}
			}
		}
	}).Register(reg)
}

func registerSP304HTTPS(reg *barrelman.Registry) {
	meta := sailpointMeta(304,
		"Server URLs must use HTTPS.",
		barrelman.SeverityError,
		barrelman.CategoryServers,
		"Use https:// for every published server URL.",
	)

	barrelman.Define(meta.ID, meta).Servers(func(server *navigator.Server, r *barrelman.Reporter) {
		if server.URL != "" && !barrelman.IsHTTPS(server.URL) {
			r.At(server.URLLoc, "[#304] Server URL '%s' must use HTTPS", server.URL)
		}
	}).Register(reg)
}

func registerSP403StatusCodes(reg *barrelman.Registry) {
	meta := sailpointMeta(403,
		"Operations must declare standard status codes and at least one 4xx response.",
		barrelman.SeverityError,
		barrelman.CategoryStructure,
		"Use IANA standard response codes and declare at least one client error response per operation.",
	)

	barrelman.Define(meta.ID, meta).Operations(func(path, method string, op *navigator.Operation, r *barrelman.Reporter) {
		has4xx := false
		has200 := false
		has201 := false
		has204 := false
		has401 := false
		has403 := false
		for code := range op.Responses {
			if code == "default" {
				continue
			}
			if !knownStatusCodes[code] {
				r.At(navigator.LocOrFallback(op.ResponsesLoc, op.Loc), "[#403] Response code '%s' for %s %s is not a standard HTTP status code", code, strings.ToUpper(method), path)
			}
			if strings.HasPrefix(code, "4") {
				has4xx = true
			}
			switch code {
			case "200":
				has200 = true
			case "201":
				has201 = true
			case "204":
				has204 = true
			case "401":
				has401 = true
			case "403":
				has403 = true
			}
		}
		responsesLoc := navigator.LocOrFallback(op.ResponsesLoc, op.Loc)
		switch strings.ToUpper(method) {
		case "GET":
			if !has200 {
				r.At(responsesLoc, "[#403] GET %s must declare a 200 response", path)
			}
		case "DELETE":
			if !has200 && !has204 {
				r.At(responsesLoc, "[#403] DELETE %s must declare a 200 or 204 response", path)
			}
		case "POST":
			if isLikelyCollectionCreate(path) && !has201 {
				r.At(responsesLoc, "[#403] POST %s should declare a 201 response for resource creation", path)
			}
		}
		if !has4xx {
			r.At(responsesLoc, "[#403] Operation %s %s must declare at least one 4xx response", strings.ToUpper(method), path)
		}
		if !has401 {
			r.At(responsesLoc, "[#403] Operation %s %s must declare a 401 response", strings.ToUpper(method), path)
		}
		if !has403 {
			r.At(responsesLoc, "[#403] Operation %s %s must declare a 403 response", strings.ToUpper(method), path)
		}
	}).Register(reg)
}

func registerSP404ProblemDetails(reg *barrelman.Registry) {
	meta := sailpointMeta(404,
		"Error responses must use application/problem+json with Problem Details fields.",
		barrelman.SeverityError,
		barrelman.CategoryStructure,
		"Return RFC 7807 Problem Details payloads for 4xx and 5xx responses.",
	)

	barrelman.Define(meta.ID, meta).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		if !hasProblemDetailsComponent(idx.Document) {
			r.AtRange(barrelman.FileStartRange, "[#404] Components must define a shared ProblemDetails schema")
		}
		requiredProps := []string{"type", "title", "status", "detail", "instance"}
		for path, item := range idx.Document.Paths {
			for _, mo := range item.Operations() {
				for code, resp := range mo.Operation.Responses {
					if !strings.HasPrefix(code, "4") && !strings.HasPrefix(code, "5") {
						continue
					}
					media := resp.Content["application/problem+json"]
					if media == nil {
						r.At(resp.Loc, "[#404] Response %s for %s %s must use application/problem+json", code, strings.ToUpper(mo.Method), path)
						continue
					}
					if !hasMediaExample(media) {
						r.At(resp.Loc, "[#404] Problem Details response %s for %s %s must include an example", code, strings.ToUpper(mo.Method), path)
					}
					if media.Schema == nil || media.Schema.Ref != "#/components/schemas/ProblemDetails" {
						r.At(resp.Loc, "[#404] Problem Details response %s for %s %s must reference #/components/schemas/ProblemDetails", code, strings.ToUpper(mo.Method), path)
					}
					schema := resolveSchema(idx, media.Schema)
					if schema == nil {
						r.At(resp.Loc, "[#404] Problem Details response %s for %s %s must declare a schema", code, strings.ToUpper(mo.Method), path)
						continue
					}
					if schema.Type != "object" {
						r.At(resp.Loc, "[#404] Problem Details response %s for %s %s must use an object schema", code, strings.ToUpper(mo.Method), path)
						continue
					}
					for _, prop := range requiredProps {
						if _, ok := schema.Properties[prop]; !ok {
							r.At(resp.Loc, "[#404] Problem Details response %s for %s %s must define property '%s'", code, strings.ToUpper(mo.Method), path, prop)
						}
					}
					if _, ok := schema.Properties["correlationId"]; !ok {
						r.At(resp.Loc, "[#404] Problem Details response %s for %s %s must define property 'correlationId'", code, strings.ToUpper(mo.Method), path)
					}
				}
			}
		}
	}).Register(reg)
}

func registerSP500NoAPIBasePath(reg *barrelman.Registry) {
	meta := sailpointMeta(500,
		"Paths must not include an /api base prefix.",
		barrelman.SeverityError,
		barrelman.CategoryPaths,
		"Publish resource paths directly instead of nesting them under /api.",
	)

	barrelman.Define(meta.ID, meta).Paths(func(path string, item *navigator.PathItem, r *barrelman.Reporter) {
		if path == "/api" || strings.HasPrefix(path, "/api/") {
			r.At(item.PathLoc, "[#500] Path '%s' must not include an /api base prefix", path)
		}
	}).Register(reg)
}

func registerSP514NonNumericPathIDs(reg *barrelman.Registry) {
	meta := sailpointMeta(514,
		"Path parameters must not use numeric identifier schemas.",
		barrelman.SeverityError,
		barrelman.CategoryPaths,
		"Model path identifiers as strings rather than integers or numbers.",
	)

	barrelman.Define(meta.ID, meta).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		for path, item := range idx.Document.Paths {
			for _, mo := range item.Operations() {
				for _, param := range operationParameters(item, mo.Operation) {
					if param == nil || param.Ref != "" || param.In != "path" || param.Schema == nil {
						continue
					}
					schema := resolveSchema(idx, param.Schema)
					if schema == nil {
						continue
					}
					if schema.Type == "integer" || schema.Type == "number" {
						r.At(param.NameLoc, "[#514] Path parameter '%s' in %s %s must use a string identifier schema", param.Name, strings.ToUpper(mo.Method), path)
					}
				}
			}
		}
	}).Register(reg)
}

func registerSP602Pagination(reg *barrelman.Registry) {
	meta := sailpointMeta(602,
		"Collection GET operations must use limit/offset pagination and a wrapped response object.",
		barrelman.SeverityError,
		barrelman.CategoryStructure,
		"Add limit and offset query parameters and wrap collection responses in an object with items metadata.",
	)

	barrelman.Define(meta.ID, meta).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		for path, item := range idx.Document.Paths {
			for _, mo := range item.Operations() {
				if strings.ToUpper(mo.Method) != "GET" {
					continue
				}
				respSchema := firstSuccessSchema(idx, mo.Operation)
				if respSchema == nil {
					continue
				}
				isCollection := respSchema.Type == "array" || (respSchema.Type == "object" && respSchema.Properties["items"] != nil)
				if !isCollection {
					continue
				}
				params := operationParameters(item, mo.Operation)
				if !hasParam(params, "query", "limit") || !hasParam(params, "query", "offset") {
					r.At(navigator.LocOrFallback(mo.Operation.ParametersLoc, mo.Operation.Loc), "[#602] Collection operation %s %s must declare query parameters 'limit' and 'offset'", strings.ToUpper(mo.Method), path)
				}
				if respSchema.Type == "array" {
					r.At(navigator.LocOrFallback(mo.Operation.ResponsesLoc, mo.Operation.Loc), "[#602] Collection operation %s %s must wrap items in a top-level object response", strings.ToUpper(mo.Method), path)
					continue
				}
				for _, prop := range []string{"items", "limit", "offset", "count"} {
					if _, ok := respSchema.Properties[prop]; !ok {
						r.At(navigator.LocOrFallback(mo.Operation.ResponsesLoc, mo.Operation.Loc), "[#602] Collection operation %s %s response schema must define property '%s'", strings.ToUpper(mo.Method), path, prop)
					}
				}
			}
		}
	}).Register(reg)
}

func registerSP710RequiredFields(reg *barrelman.Registry) {
	meta := sailpointMeta(710,
		"Request and response schemas must declare required fields, and path parameters must be required.",
		barrelman.SeverityError,
		barrelman.CategoryTypes,
		"Declare required arrays on object schemas and mark every path parameter as required.",
	)

	barrelman.Define(meta.ID, meta).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		for path, item := range idx.Document.Paths {
			for _, mo := range item.Operations() {
				for _, param := range operationParameters(item, mo.Operation) {
					if param == nil || param.Ref != "" || param.In != "path" {
						continue
					}
					if !param.Required {
						r.At(param.NameLoc, "[#710] Path parameter '%s' in %s %s must set required: true", param.Name, strings.ToUpper(mo.Method), path)
					}
				}
				if mo.Operation.RequestBody != nil {
					for mediaType, media := range mo.Operation.RequestBody.Content {
						schema := resolveSchema(idx, media.Schema)
						if !schemaNeedsRequiredArray(schema) {
							continue
						}
						r.At(mo.Operation.RequestBody.Loc, "[#710] Request body schema for %s %s (%s) must declare a required array", strings.ToUpper(mo.Method), path, mediaType)
					}
				}
				for code, resp := range mo.Operation.Responses {
					for mediaType, media := range resp.Content {
						schema := resolveSchema(idx, media.Schema)
						if !schemaNeedsRequiredArray(schema) {
							continue
						}
						r.At(resp.Loc, "[#710] Response schema for %s %s response %s (%s) must declare a required array", strings.ToUpper(mo.Method), path, code, mediaType)
					}
				}
			}
		}
	}).Register(reg)
}

func registerSP701NoNullableBooleans(reg *barrelman.Registry) {
	meta := sailpointMeta(701,
		"Boolean schemas must not be nullable.",
		barrelman.SeverityError,
		barrelman.CategoryTypes,
		"Represent booleans as true/false only; remove nullable from boolean schemas.",
	)

	barrelman.Define(meta.ID, meta).RecursiveSchemas(func(name string, schema *navigator.Schema, pointer string, r *barrelman.Reporter) {
		if schema.Type == "boolean" && schema.Nullable {
			r.At(schema.Loc, "[#701] Boolean schema at %s must not be nullable", pointer)
		}
	}).Register(reg)
}

func registerSP702NoNullableArrays(reg *barrelman.Registry) {
	meta := sailpointMeta(702,
		"Array schemas must not be nullable.",
		barrelman.SeverityError,
		barrelman.CategoryTypes,
		"Use empty arrays instead of nullable arrays.",
	)

	barrelman.Define(meta.ID, meta).RecursiveSchemas(func(name string, schema *navigator.Schema, pointer string, r *barrelman.Reporter) {
		if schema.Type == "array" && schema.Nullable {
			r.At(schema.Loc, "[#702] Array schema at %s must not be nullable", pointer)
		}
	}).Register(reg)
}

func registerSP804NumericTypes(reg *barrelman.Registry) {
	meta := sailpointMeta(804,
		"Numeric schemas must use approved numeric formats, and identifier fields must use strings.",
		barrelman.SeverityError,
		barrelman.CategoryTypes,
		"Use int32/int64 for integers, float/double for numbers, and string types for identifiers.",
	)

	barrelman.Define(meta.ID, meta).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		walkAllSchemas(idx, func(name string, schema *navigator.Schema, pointer string) {
			schemaName := nameFromPointer(pointer, name)
			if looksLikeID(schemaName) && (schema.Type == "integer" || schema.Type == "number") {
				r.At(schema.Loc, "[#804] Identifier schema at %s must use type string instead of %s", pointer, schema.Type)
			}
			switch schema.Type {
			case "integer":
				if schema.Format != "int32" && schema.Format != "int64" {
					r.At(schema.Loc, "[#804] Integer schema at %s must declare format int32 or int64", pointer)
				}
			case "number":
				if schema.Format != "float" && schema.Format != "double" {
					r.At(schema.Loc, "[#804] Number schema at %s must declare format float or double", pointer)
				}
			}
			for propName, prop := range schema.Properties {
				if prop == nil {
					continue
				}
				if looksLikeID(propName) && (prop.Type == "integer" || prop.Type == "number") {
					r.At(propertyLoc(prop, schema), "[#804] Identifier property '%s' at %s must use type string", propName, pointerForProperty(pointer, propName))
				}
			}
		})
	}).Register(reg)
}

func registerSP903RequestIDHeader(reg *barrelman.Registry) {
	meta := sailpointMeta(903,
		"Responses must declare the X-Request-Id header.",
		barrelman.SeverityError,
		barrelman.CategoryStructure,
		"Add an X-Request-Id response header to every response definition.",
	)

	barrelman.Define(meta.ID, meta).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		if !hasSharedRequestIDHeaderComponent(idx.Document) {
			r.AtRange(barrelman.FileStartRange, "[#903] Components must define a shared X-Request-Id header with type string and format uuid")
		}
		for path, item := range idx.Document.Paths {
			for _, mo := range item.Operations() {
				for code, resp := range mo.Operation.Responses {
					header := responseHeader(resp, "X-Request-Id")
					if header == nil {
						r.At(headerDiagLoc(resp), "[#903] Response %s for %s %s must declare the X-Request-Id header", code, strings.ToUpper(mo.Method), path)
						continue
					}
					if !requestIDHeaderHasUUIDFormat(idx, header) {
						r.At(headerDiagLoc(resp), "[#903] X-Request-Id header on response %s for %s %s must use type string with format uuid", code, strings.ToUpper(mo.Method), path)
					}
				}
			}
		}
	}).Register(reg)
}

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
