package analyzers

import (
	"encoding/json"
	"log/slog"
	"strings"
	"sync"

	navigator "github.com/LukasParke/navigator"
	"github.com/dlclark/regexp2"
	"github.com/sailpoint-oss/barrelman"
	"github.com/sailpoint-oss/barrelman/schemas"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"
)

var structuralMeta = barrelman.RuleMeta{
	ID:          "oas3-schema",
	Description: "Validates the document structure against the OpenAPI JSON Schema.",
	Severity:    barrelman.SeverityError,
	Category:    barrelman.CategoryStructure,
	Recommended: true,
	HowToFix:    "Fix the invalid or unknown properties flagged by the validator.",
	DocURL:      barrelman.DocBaseURL + "oas3-schema",
}

var (
	schemaOnce    sync.Once
	compiledCache map[navigator.Version]*jsonschema.Schema
	fragmentCache map[navigator.Version]map[navigator.FragmentType]*jsonschema.Schema
)

var fragmentVersions = []navigator.Version{navigator.Version30, navigator.Version31, navigator.Version32}

type pcreRegexp struct {
	re *regexp2.Regexp
}

func (r *pcreRegexp) MatchString(s string) bool {
	ok, _ := r.re.MatchString(s)
	return ok
}

func (r *pcreRegexp) String() string {
	return r.re.String()
}

func newSchemaCompiler() *jsonschema.Compiler {
	c := jsonschema.NewCompiler()
	c.UseRegexpEngine(func(pattern string) (jsonschema.Regexp, error) {
		re, err := regexp2.Compile(pattern, regexp2.None)
		if err != nil {
			return nil, err
		}
		return &pcreRegexp{re: re}, nil
	})
	return c
}

func loadSchemas() {
	compiledCache = make(map[navigator.Version]*jsonschema.Schema, 4)
	fragmentCache = make(map[navigator.Version]map[navigator.FragmentType]*jsonschema.Schema, 3)

	schemaFiles := map[navigator.Version]string{
		navigator.Version20: "generated/openapi-2.0-root.json",
		navigator.Version30: "generated/openapi-3.0-root.json",
		navigator.Version31: "generated/openapi-3.1-root.json",
		navigator.Version32: "generated/openapi-3.2-root.json",
	}

	compiler := newSchemaCompiler()

	for ver, path := range schemaFiles {
		data, err := schemas.FS.ReadFile(path)
		if err != nil {
			slog.Warn("failed to read OpenAPI schema file", "path", path, "version", ver, "error", err)
			continue
		}
		var raw interface{}
		if err := json.Unmarshal(data, &raw); err != nil {
			slog.Warn("failed to parse schema JSON", "path", path, "error", err)
			continue
		}
		patchMissingDefs(raw)
		url := "file:///" + path
		if err := compiler.AddResource(url, raw); err != nil {
			slog.Warn("failed to add schema resource", "path", path, "error", err)
			continue
		}
		compiled, err := compiler.Compile(url)
		if err != nil {
			slog.Warn("failed to compile OpenAPI schema", "path", path, "version", ver, "error", err)
			continue
		}
		compiledCache[ver] = compiled
	}

	fragmentTypes := map[navigator.FragmentType]string{
		navigator.FragmentSchema:         "schema",
		navigator.FragmentPathItem:       "path-item",
		navigator.FragmentOperation:      "operation",
		navigator.FragmentParameter:      "parameter",
		navigator.FragmentRequestBody:    "request-body",
		navigator.FragmentResponse:       "response",
		navigator.FragmentHeader:         "header",
		navigator.FragmentSecurityScheme: "security-scheme",
		navigator.FragmentComponents:     "components",
		navigator.FragmentServer:         "server",
	}

	for _, ver := range fragmentVersions {
		versionPrefix := string(ver)
		vmap := make(map[navigator.FragmentType]*jsonschema.Schema, len(fragmentTypes))
		for fragType, suffix := range fragmentTypes {
			path := "generated/openapi-" + versionPrefix + "-" + suffix + ".json"
			data, err := schemas.FS.ReadFile(path)
			if err != nil {
				slog.Warn("failed to read fragment schema file", "path", path, "version", ver, "error", err)
				continue
			}
			var raw interface{}
			if err := json.Unmarshal(data, &raw); err != nil {
				slog.Warn("failed to parse fragment schema JSON", "path", path, "error", err)
				continue
			}
			patchMissingDefs(raw)
			url := "file:///" + path
			fragCompiler := newSchemaCompiler()
			if err := fragCompiler.AddResource(url, raw); err != nil {
				slog.Warn("failed to add fragment schema resource", "path", path, "error", err)
				continue
			}
			compiled, err := fragCompiler.Compile(url)
			if err != nil {
				slog.Warn("failed to compile fragment schema", "path", path, "version", ver, "error", err)
				continue
			}
			vmap[fragType] = compiled
		}
		fragmentCache[ver] = vmap
	}

	if len(compiledCache) == 0 {
		slog.Error("no OpenAPI schemas loaded — structural validation disabled")
	}
}

func getSchema(ver navigator.Version) *jsonschema.Schema {
	schemaOnce.Do(loadSchemas)
	return compiledCache[ver]
}

func getFragmentSchema(ver navigator.Version, ft navigator.FragmentType) *jsonschema.Schema {
	schemaOnce.Do(loadSchemas)
	vmap := fragmentCache[ver]
	if vmap == nil {
		return nil
	}
	return vmap[ft]
}

const defaultFragmentVersion = navigator.Version31

func registerStructuralValidation(reg *barrelman.Registry) {
	reg.Register(barrelman.Rule{
		ID:   "oas3-schema",
		Meta: structuralMeta,
		Run: func(ctx *barrelman.AnalysisContext) []barrelman.Diagnostic {
			if ctx.Content == nil || len(ctx.Content) == 0 {
				return nil
			}

			var parsed interface{}
			if err := yaml.Unmarshal(ctx.Content, &parsed); err != nil {
				return nil
			}
			normalized := normalizeYAML(parsed)

			if ctx.Index != nil && ctx.Index.Version != navigator.VersionUnknown {
				schema := getSchema(ctx.Index.Version)
				if schema == nil {
					return nil
				}
				return validateAgainstSchema(normalized, schema)
			}

			rootKeys := extractRootKeys(normalized)
			fragType := navigator.DetectFragmentType(rootKeys)
			if fragType == navigator.FragmentUnknown {
				return nil
			}

			ver := ctx.TargetVersion
			if ver == "" {
				ver = defaultFragmentVersion
			}
			fragSchema := getFragmentSchema(ver, fragType)
			if fragSchema == nil {
				return nil
			}
			return validateAgainstSchema(normalized, fragSchema)
		},
	})
}

func validateAgainstSchema(data interface{}, schema *jsonschema.Schema) []barrelman.Diagnostic {
	err := schema.Validate(data)
	if err == nil {
		return nil
	}

	var diags []barrelman.Diagnostic
	ve, ok := err.(*jsonschema.ValidationError)
	if !ok {
		diags = append(diags, barrelman.Diagnostic{
			Range:    barrelman.FileStartRange,
			Severity: barrelman.SeverityError,
			Source:   barrelman.Source,
			Code:     "oas3-schema",
			Message:  err.Error(),
		})
		return diags
	}

	for _, cause := range flattenValidationErrors(ve) {
		if isExtensionKeyError(cause) {
			continue
		}
		msg := cause.Error()
		pointer := strings.Join(cause.InstanceLocation, "/")
		if pointer != "" {
			msg = "/" + pointer + ": " + msg
		}
		diags = append(diags, barrelman.Diagnostic{
			Range:    barrelman.FileStartRange,
			Severity: barrelman.SeverityError,
			Source:   barrelman.Source,
			Code:     "oas3-schema",
			Message:  msg,
		})
		if len(diags) >= 100 {
			break
		}
	}
	return diags
}

// isExtensionKeyError returns true if the validation error is about an x-*
// extension key being rejected as an additional property. The OpenAPI spec
// allows x-* keys everywhere so these are false positives.
func isExtensionKeyError(ve *jsonschema.ValidationError) bool {
	msg := ve.Error()
	if !strings.Contains(msg, "additional properties") {
		return false
	}
	for _, seg := range ve.InstanceLocation {
		if strings.HasPrefix(seg, "x-") {
			return true
		}
	}
	if strings.Contains(msg, "'x-") {
		return true
	}
	return false
}

func flattenValidationErrors(ve *jsonschema.ValidationError) []*jsonschema.ValidationError {
	if len(ve.Causes) == 0 {
		return []*jsonschema.ValidationError{ve}
	}
	var flat []*jsonschema.ValidationError
	for _, cause := range ve.Causes {
		flat = append(flat, flattenValidationErrors(cause)...)
	}
	return flat
}

// normalizeYAML converts yaml.v3 output (map[string]interface{}) so
// that it uses map[string]interface{} consistently (the jsonschema lib
// doesn't accept map[interface{}]interface{}).
func normalizeYAML(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		m := make(map[string]interface{}, len(val))
		for k, v := range val {
			m[k] = normalizeYAML(v)
		}
		return m
	case map[interface{}]interface{}:
		m := make(map[string]interface{}, len(val))
		for k, v := range val {
			key, ok := k.(string)
			if !ok {
				key = strings.TrimSpace(strings.Trim(strings.Trim(
					strings.Trim(jsonMarshal(k), "\""), "'"), " "))
			}
			m[key] = normalizeYAML(v)
		}
		return m
	case []interface{}:
		a := make([]interface{}, len(val))
		for i, v := range val {
			a[i] = normalizeYAML(v)
		}
		return a
	default:
		return v
	}
}

func jsonMarshal(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// patchMissingDefs scans the schema document for $ref targets pointing to
// $defs entries that don't exist, and fills them in with permissive `true`
// schemas. This works around generated OpenAPI meta-schemas that contain
// dangling internal references.
func patchMissingDefs(doc interface{}) {
	root, ok := doc.(map[string]interface{})
	if !ok {
		return
	}
	defsRaw, ok := root["$defs"]
	if !ok {
		return
	}
	defs, ok := defsRaw.(map[string]interface{})
	if !ok {
		return
	}

	referenced := make(map[string]bool)
	collectDefRefs(doc, referenced)

	for name := range referenced {
		if _, exists := defs[name]; !exists {
			defs[name] = true
		}
	}
}

func collectDefRefs(v interface{}, out map[string]bool) {
	switch val := v.(type) {
	case map[string]interface{}:
		if ref, ok := val["$ref"]; ok {
			if s, ok := ref.(string); ok && strings.HasPrefix(s, "#/$defs/") {
				out[strings.TrimPrefix(s, "#/$defs/")] = true
			}
		}
		for _, child := range val {
			collectDefRefs(child, out)
		}
	case []interface{}:
		for _, child := range val {
			collectDefRefs(child, out)
		}
	}
}

func extractRootKeys(data interface{}) []string {
	m, ok := data.(map[string]interface{})
	if !ok {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
