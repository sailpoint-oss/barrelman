package rulesets

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadBytes_ParsesRuleSet(t *testing.T) {
	rs, err := LoadBytes([]byte(`
name: custom
rules:
  sp-123: error
`))
	if err != nil {
		t.Fatalf("LoadBytes: %v", err)
	}
	if rs.Name != "custom" {
		t.Fatalf("Name = %q, want custom", rs.Name)
	}
	if rs.Rules["sp-123"].Severity != "error" {
		t.Fatalf("unexpected rules: %+v", rs.Rules)
	}
}

func TestLoadBytes_NormalizesLegacyRuleIDs(t *testing.T) {
	rs, err := LoadBytes([]byte(`
rules:
  operation-tags: error
`))
	if err != nil {
		t.Fatalf("LoadBytes: %v", err)
	}
	if _, ok := rs.Rules["operation-tags"]; ok {
		t.Fatalf("expected legacy rule ID to be normalized, got %+v", rs.Rules)
	}
	if rs.Rules["sp-123"].Severity != "error" {
		t.Fatalf("unexpected rules: %+v", rs.Rules)
	}
}

func TestResolve_LoadsRelativeExtendsChain(t *testing.T) {
	dir := t.TempDir()
	basePath := filepath.Join(dir, "base.yaml")
	childPath := filepath.Join(dir, "child.yaml")

	if err := os.WriteFile(basePath, []byte(`
rules:
  sp-123: error
`), 0o644); err != nil {
		t.Fatalf("write base ruleset: %v", err)
	}
	if err := os.WriteFile(childPath, []byte(`
extends: ./base.yaml
rules:
  info-description: warn
`), 0o644); err != nil {
		t.Fatalf("write child ruleset: %v", err)
	}

	rs, err := LoadFile(childPath)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}
	resolved, err := Resolve(rs, dir)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if resolved.Rules["sp-123"].Severity != "error" {
		t.Fatalf("expected inherited rule, got %+v", resolved.Rules)
	}
	if resolved.Rules["info-description"].Severity != "warn" {
		t.Fatalf("expected child rule override, got %+v", resolved.Rules)
	}
}

func TestResolve_RejectsCircularExtends(t *testing.T) {
	dir := t.TempDir()
	aPath := filepath.Join(dir, "a.yaml")
	bPath := filepath.Join(dir, "b.yaml")

	if err := os.WriteFile(aPath, []byte("extends: ./b.yaml\n"), 0o644); err != nil {
		t.Fatalf("write a.yaml: %v", err)
	}
	if err := os.WriteFile(bPath, []byte("extends: ./a.yaml\n"), 0o644); err != nil {
		t.Fatalf("write b.yaml: %v", err)
	}

	rs, err := LoadFile(aPath)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}
	_, err = Resolve(rs, dir)
	if err == nil || !strings.Contains(err.Error(), "circular extends") {
		t.Fatalf("expected circular extends error, got %v", err)
	}
}

func TestResolve_LoadsBarrelmanBuiltinRuleset(t *testing.T) {
	rs, err := LoadBytes([]byte(`
extends: barrelman:recommended
rules:
  sp-123: error
`))
	if err != nil {
		t.Fatalf("LoadBytes: %v", err)
	}

	resolved, err := Resolve(rs, t.TempDir())
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if len(resolved.Rules) == 0 {
		t.Fatal("expected resolved builtin rules")
	}
	if resolved.Rules["sp-123"].Severity != "error" {
		t.Fatalf("sp-123 severity = %q, want error", resolved.Rules["sp-123"].Severity)
	}
}
