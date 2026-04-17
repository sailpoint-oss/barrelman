package analyzers

import (
	"strings"

	"github.com/sailpoint-oss/barrelman"
	navigator "github.com/sailpoint-oss/navigator"
)

// #111 - OAuth scopes must use lower-case <domain>:<resource>:<action> naming.
func registerSailpointOAuthScopeFormat(reg *barrelman.Registry) {
	meta := sailpointMeta(
		"sailpoint-oauth-scope-format",
		barrelman.SeverityError,
		barrelman.CategorySecurity,
		"OAuth scopes must use lower-case <domain>:<resource>:<action> names.",
		"Rename scopes to lower-case colon-separated domain, resource, and action segments.",
	)

	barrelman.Define(meta.ID, meta).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		for name, ss := range idx.SecuritySchemes {
			for _, scope := range declaredOAuthScopes(ss) {
				if isValidScopeName(scope) {
					continue
				}
				r.At(navigator.LocOrFallback(ss.NameLoc, ss.Loc),
					"OAuth scope '%s' declared by security scheme '%s' must use lower-case <domain>:<resource>:<action> naming",
					scope, name)
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
							r.At(loc,
								"OAuth scope '%s' on %s %s must use lower-case <domain>:<resource>:<action> naming",
								scope, strings.ToUpper(mo.Method), path)
						}
					}
				}
			}
		}
	}).Register(reg)
}

// #300 (split 1/2) - The API must define at least one OAuth 2.0 security scheme.
func registerSailpointSecurityOAuth2Required(reg *barrelman.Registry) {
	meta := sailpointMeta(
		"sailpoint-security-oauth2-required",
		barrelman.SeverityError,
		barrelman.CategorySecurity,
		"The API must define at least one OAuth 2.0 security scheme.",
		"Define at least one OAuth 2.0 security scheme in components.securitySchemes.",
	)

	barrelman.Define(meta.ID, meta).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		if len(idx.SecuritySchemes) == 0 {
			r.AtRange(barrelman.FileStartRange, "The API must define at least one OAuth 2.0 security scheme")
			return
		}
		for name, ss := range idx.SecuritySchemes {
			if ss == nil || ss.Ref != "" {
				continue
			}
			if ss.Type != "oauth2" {
				r.At(ss.Loc, "Security scheme '%s' must use type oauth2", name)
			}
		}
	}).Register(reg)
}

// #300 (split 2/2) - Every operation must declare security requirements.
func registerSailpointOperationSecurityRequired(reg *barrelman.Registry) {
	meta := sailpointMeta(
		"sailpoint-operation-security-required",
		barrelman.SeverityError,
		barrelman.CategorySecurity,
		"Every operation must declare security requirements.",
		"Add a security requirement to the operation or to the root of the document.",
	)

	barrelman.Define(meta.ID, meta).Operations(func(path, method string, op *navigator.Operation, r *barrelman.Reporter) {
		// The Operations visitor does not provide access to document-level
		// security, so we re-check via the custom accessor below.
	}).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		for path, item := range idx.Document.Paths {
			for _, mo := range item.Operations() {
				reqs := effectiveSecurity(idx.Document, mo.Operation)
				if len(reqs) == 0 {
					r.At(mo.Operation.Loc, "Operation %s %s must declare security requirements", strings.ToUpper(mo.Method), path)
				}
			}
		}
	}).Register(reg)
}

// #301 - OAuth 2.0 security requirements must declare valid scopes.
func registerSailpointOAuthScopesDeclared(reg *barrelman.Registry) {
	meta := sailpointMeta(
		"sailpoint-oauth-scopes-declared",
		barrelman.SeverityError,
		barrelman.CategorySecurity,
		"OAuth 2.0 security requirements must declare valid scopes.",
		"List at least one scope defined by the OAuth security scheme on every security requirement entry.",
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
							r.At(entry.NameLoc, "OAuth requirement '%s' on %s %s must declare at least one scope", entry.Name, strings.ToUpper(mo.Method), path)
							continue
						}
						allowed := oauthScopes(ss)
						for _, scope := range entry.Scopes {
							if allowed[scope] {
								continue
							}
							r.At(entry.NameLoc, "OAuth scope '%s' on %s %s is not defined by security scheme '%s'", scope, strings.ToUpper(mo.Method), path, entry.Name)
						}
					}
				}
			}
		}
	}).Register(reg)
}

// #304 - Server URLs must use HTTPS.
func registerSailpointServerURLHTTPS(reg *barrelman.Registry) {
	meta := sailpointMeta(
		"sailpoint-server-url-https",
		barrelman.SeverityError,
		barrelman.CategoryServers,
		"Server URLs must use HTTPS.",
		"Use https:// for every published server URL.",
	)

	barrelman.Define(meta.ID, meta).Servers(func(server *navigator.Server, r *barrelman.Reporter) {
		if server.URL != "" && !barrelman.IsHTTPS(server.URL) {
			r.At(server.URLLoc, "Server URL '%s' must use HTTPS", server.URL)
		}
	}).Register(reg)
}
