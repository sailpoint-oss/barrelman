# Barrelman

Navigator-backed OpenAPI linting and diagnostic packaging for the workspace toolchain.

## Toolchain role

Barrelman sits one layer above `navigator`:

- `navigator` owns parsing, typed OpenAPI indexes, `$ref` indexes, workspace resolution primitives, and parse-time issues.
- `barrelman` owns rule execution, rulesets, severities, filtering, and diagnostic shaping.
- `telescope`, `cartographer`, and other consumers run Barrelman to surface lint output in editors, CI, and reports.

Barrelman does **not** own the OpenAPI parser or the canonical structural validator. The built-in `oas3-schema` rule maps Navigator issues into Barrelman diagnostics so downstream tools see one consistent structural-validation story.

## Core API

- `LintContent(uri, content, opts)` runs all enabled rules against in-memory YAML/JSON.
- `LintFiles(files, opts)` runs the same rules against files on disk.
- `Registry`, `Rule`, and `RuleMeta` let you register custom rules.
- `LintOptions` controls severity filtering, rulesets, target version hints, and workspace settings.

## Quick start

```go
package main

import (
	"fmt"
	"log"

	"github.com/sailpoint-oss/barrelman"
)

func main() {
	content := []byte(`openapi: "3.1.0"
info:
  title: Example
  version: "1.0.0"
paths: {}`)

	diags, err := barrelman.LintContent("file:///api.yaml", content, barrelman.LintOptions{})
	if err != nil {
		log.Fatal(err)
	}
	for _, d := range diags {
		fmt.Printf("%s: %s\n", d.Code, d.Message)
	}
}
```

## Local sibling development

When changing Barrelman alongside other toolchain repos, prefer a workspace `go.work` file:

```bash
go work init .
go work use ../navigator ../telescope/server ../cartographer/cartographer ../barometer
```

This keeps `go.mod` pins clean while you iterate on shared contracts.

## Release coordination

- `.github/workflows/release.yml` currently auto-tags and publishes from pushes to `main`.
- If Barrelman changes diagnostic, ruleset, or config contracts, bump consumers in this order: `telescope/server`, then `cartographer/cartographer`.
- Use `navigator/TOOLCHAIN_BOUNDARIES.md` for the shared bump order and `navigator/TOOLCHAIN_FIXTURE_MATRIX.md` for cross-repo smoke anchors.
- A minimal local smoke pass is:
  - `go test ./...`
  - `cd ../telescope/server && go test ./...`
  - `cd ../cartographer/cartographer && go test ./cmd ./spec ./unified ./telescope`

## Ownership boundaries

- Need CST or grammar work: use `tree-sitter-openapi`.
- Need parse/index/ref/model access: use `navigator`.
- Need rule execution or CI/editor diagnostics: use `barrelman`.
- Need LSP, code actions, or editor workflows: use `telescope`.

See `navigator/TOOLCHAIN_BOUNDARIES.md` in the sibling workspace for the full cross-repo contract.
