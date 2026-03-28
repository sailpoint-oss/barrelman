// Package config handles loading and parsing of Barrelman's static-analysis
// configuration files.
//
// # Loading Configuration
//
// [Load] searches for configuration files in a workspace directory,
// trying each path in [ConfigFiles] order:
//
//	cfg, err := config.Load("/path/to/project")
//
// If no configuration file is found, [DefaultConfig] is returned with
// sensible defaults (extends barrelman:recommended, standard include
// patterns for YAML/JSON files).
//
// [LoadFile] loads from a specific path:
//
//	cfg, err := config.LoadFile(".barrelman.yaml")
//
// # Configuration Fields
//
// The [Config] struct supports:
//
//   - Extends: base ruleset name (e.g., "barrelman:recommended")
//   - Rules: per-rule severity overrides
//   - Plugins: paths to Spectral YAML rulesets or Go plugin binaries
//   - Include/Exclude: glob patterns for file discovery
//   - OpenAPI: version targeting and extension schemas
//   - Output: CLI format and color preferences
//   - LSP: debounce and file size limits
package config
