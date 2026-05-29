# AGENTS.md

This file is the canonical agent context for Barrelman.

## Project overview

**Barrelman** is the public, org-neutral static linting and diagnostic-packaging layer of the OpenAPI toolchain. It runs rule logic against Navigator-parsed documents and packages the resulting issues into stable diagnostics.

Barrelman owns:

- the `Rule` and `RuleMeta` shape every analyzer compiles down to
- the `Registry` that holds the rule catalogue
- the ruleset model (`barrelman:recommended`, `barrelman:all`, `barrelman:owasp`, `barrelman:strict`) and Spectral bridge
- the **generic** rule families: OAS structural parity (`oas3-schema`), naming, documentation, structure, types, security, servers, paths, OWASP, references, syntax checks, validation engine, markdown
- the codemod / fix engine for auto-fixable rules
- the `RulePack` plug-in interface so downstream consumers can attach their own rule packs

Barrelman does **not** own:

- the parser or canonical structural validator (use Navigator)
- editor / LSP UX (use Telescope)
- vendor-specific guideline rules (downstream rule packs)

## Plug-in interface (`barrelman.RulePack`)

The plug-in surface lets a downstream consumer register additional rules at startup without forking Barrelman or telescope. It lives in `plugin.go`:

```go
type RulePack interface {
    Name() string
    Register(reg *Registry)
}

func RegisterPlugin(p RulePack)
func ApplyPlugins(reg *Registry)
func RegisteredPluginNames() []string
func ClearPlugins() // for tests
```

A downstream consumer ships their rule pack like this:

```go
package mybrand

import "github.com/sailpoint-oss/barrelman"

type Pack struct{}
func (Pack) Name() string                       { return "mybrand" }
func (Pack) Register(reg *barrelman.Registry)   { /* register rules */ }

func init() { barrelman.RegisterPlugin(Pack{}) }
```

Any binary that blank-imports `mybrand` and calls `analyzers.RegisterAll(reg)` will pick up the pack automatically — `RegisterAll` calls `ApplyPlugins` after loading the generic analyzers.

`RegisterGeneric(reg)` is the dependency-free entry point for callers that want to skip plug-in application.

## Exported helper functions

Downstream rule packs frequently need the same traversal helpers Barrelman uses internally. The following are exported for that purpose:

| Function | Location | Purpose |
|----------|----------|---------|
| `analyzers.SchemaNameFromPointer` | `analyzers/context.go` | Recover a schema name from a JSON-pointer-like path. |
| `analyzers.WalkAllSchemas` | `analyzers/owasp.go` | Iterate every schema reachable from the document index. |
| `analyzers.HeaderDiagLoc` | `analyzers/owasp.go` | Best diagnostic location for a header-related finding on a response. |

## Core API

- `LintContent(uri, content, opts)` runs every enabled rule against in-memory YAML/JSON.
- `LintFiles(files, opts)` runs the same rules against files on disk.
- `Registry`, `Rule`, and `RuleMeta` let you register custom rules in-process.
- `LintOptions` controls severity filtering, rulesets, target version hints, and workspace settings.

Navigator-owned structural and meta issues flow through Barrelman via the built-in `oas3-schema` rule. The rule ID is legacy OpenAPI-shaped for compatibility; diagnostics describe the underlying document family through Navigator issue metadata so OpenAPI and Arazzo can share the same structural bridge.

## Config and rulesets

Barrelman owns the shared static-analysis config and ruleset contract:

- Canonical workspace config files: `.barrelman.yaml`, `.barrelman.yml`, `.barrelman/config.yaml`, `.barrelman/config.yml`.
- Legacy `.telescope.yaml` and `.telescope/config.yaml` are still discovered for compatibility.
- Canonical built-in ruleset names: `barrelman:recommended`, `barrelman:all`, `barrelman:owasp`, `barrelman:strict`.
- Legacy `telescope:*` ruleset aliases still resolve to the same built-in sets.

## Build and test

```bash
go test ./...
```

For sibling-repo development, use a workspace `go.work` (gitignored):

```bash
go work init .
go work use ../navigator ../telescope/server ../barometer
```

## Working boundaries

- Do NOT add vendor-branded rules (`<vendor>-*` slugs) to this repo. Downstream consumers attach them via `RegisterPlugin`.
- Do NOT recognise vendor-specific `x-<vendor>-*` extension names by default. Generic `x-source-*` evidence extensions emitted by cartographer are fine; anything beyond that belongs in a downstream consumer.
- Do NOT depend on Telescope or any LSP-side abstractions.

## Leak guard

`.github/leak-guard/` ships a salted-bloom filter + shape-only patterns + a small standalone Go scanner. PR CI fails on any hit. To install the pre-push hook locally:

```bash
./scripts/install-leak-guard-hooks.sh
```

## Release coordination

- `.github/workflows/release.yml` auto-tags and publishes from pushes to `main`.
- If Barrelman changes diagnostic, ruleset, or config contracts, bump consumer integrations after `telescope/server`.
- Use `navigator/TOOLCHAIN_BOUNDARIES.md` for the shared bump order and `navigator/TOOLCHAIN_FIXTURE_MATRIX.md` for cross-repo smoke anchors.

## Related repositories

This repo is part of a six-repo OpenAPI toolchain:

- [tree-sitter-openapi](https://github.com/sailpoint-oss/tree-sitter-openapi) — grammar and tree-sitter bindings
- [navigator](https://github.com/sailpoint-oss/navigator) — parse, index, `$ref` resolution, document validation
- [cartographer](https://github.com/sailpoint-oss/cartographer) — source-to-OpenAPI extractor for Go, Java, TypeScript, Python, C#
- [telescope](https://github.com/sailpoint-oss/telescope) — VS Code extension, language server, and CLI built on the above
- [barometer](https://github.com/sailpoint-oss/barometer) — live HTTP contract testing and Arazzo runner
