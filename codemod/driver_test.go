package codemod

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestDriver_Apply_PureInsertions(t *testing.T) {
	src := []byte("hello world")
	patches := []Patch{
		{StartByte: 5, EndByte: 5, Replacement: []byte(","), Description: "comma"},
		{StartByte: 11, EndByte: 11, Replacement: []byte("!"), Description: "bang"},
	}
	got, err := (&Driver{}).Apply(src, patches)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if string(got) != "hello, world!" {
		t.Fatalf("got %q, want %q", got, "hello, world!")
	}
}

func TestDriver_Apply_ReverseOrderIndependence(t *testing.T) {
	src := []byte("aaaa")
	// Two insertions at different offsets; applying in the order given
	// here would shift offsets if not reverse-sorted.
	patches := []Patch{
		{StartByte: 1, EndByte: 1, Replacement: []byte("X")},
		{StartByte: 3, EndByte: 3, Replacement: []byte("Y")},
	}
	forward, err := (&Driver{}).Apply(src, patches)
	if err != nil {
		t.Fatalf("Apply forward: %v", err)
	}
	reversed, err := (&Driver{}).Apply(src, []Patch{patches[1], patches[0]})
	if err != nil {
		t.Fatalf("Apply reversed: %v", err)
	}
	if string(forward) != "aXaaYa" {
		t.Fatalf("forward = %q, want aXaaYa", forward)
	}
	if string(forward) != string(reversed) {
		t.Fatalf("result depends on input order: forward=%q reversed=%q", forward, reversed)
	}
}

func TestDriver_Apply_OverlapRejected(t *testing.T) {
	src := []byte("abcdef")
	patches := []Patch{
		{StartByte: 1, EndByte: 4, Replacement: []byte("XXX"), RuleID: "r1"},
		{StartByte: 2, EndByte: 5, Replacement: []byte("YYY"), RuleID: "r2"},
	}
	_, err := (&Driver{AllowShrink: true}).Apply(src, patches)
	if !errors.Is(err, ErrPatchConflict) {
		t.Fatalf("expected ErrPatchConflict, got %v", err)
	}
}

func TestDriver_Apply_AbuttingReplacementsRejected(t *testing.T) {
	src := []byte("abcdef")
	patches := []Patch{
		{StartByte: 1, EndByte: 3, Replacement: []byte("XX"), RuleID: "r1"},
		{StartByte: 3, EndByte: 5, Replacement: []byte("YY"), RuleID: "r2"},
	}
	// Patches that share a boundary must both be pure insertions;
	// replacements that abut are treated as a conflict.
	_, err := (&Driver{AllowShrink: true}).Apply(src, patches)
	if !errors.Is(err, ErrPatchConflict) {
		t.Fatalf("expected ErrPatchConflict for abutting replacements, got %v", err)
	}
}

func TestDriver_Apply_AbuttingInsertionsAllowed(t *testing.T) {
	src := []byte("ab")
	patches := []Patch{
		{StartByte: 1, EndByte: 1, Replacement: []byte("X")},
		{StartByte: 1, EndByte: 1, Replacement: []byte("Y")},
	}
	got, err := (&Driver{}).Apply(src, patches)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if string(got) != "aXYb" && string(got) != "aYXb" {
		t.Fatalf("unexpected result %q (want aXYb or aYXb)", got)
	}
}

func TestDriver_Apply_ShrinkDisallowedByDefault(t *testing.T) {
	src := []byte("hello")
	patches := []Patch{
		{StartByte: 0, EndByte: 5, Replacement: []byte("hi"), RuleID: "shrink"},
	}
	_, err := (&Driver{}).Apply(src, patches)
	if !errors.Is(err, ErrShrinkDisallowed) {
		t.Fatalf("expected ErrShrinkDisallowed, got %v", err)
	}
}

func TestDriver_Apply_ShrinkAllowedWhenOptedIn(t *testing.T) {
	src := []byte("hello")
	patches := []Patch{
		{StartByte: 0, EndByte: 5, Replacement: []byte("hi"), RuleID: "shrink"},
	}
	got, err := (&Driver{AllowShrink: true}).Apply(src, patches)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if string(got) != "hi" {
		t.Fatalf("got %q, want hi", got)
	}
}

func TestDriver_Apply_OutOfBoundsRejected(t *testing.T) {
	src := []byte("short")
	patches := []Patch{
		{StartByte: 0, EndByte: 99, Replacement: []byte("x"), RuleID: "oob"},
	}
	_, err := (&Driver{AllowShrink: true}).Apply(src, patches)
	if err == nil || !strings.Contains(err.Error(), "exceeds source length") {
		t.Fatalf("expected out-of-bounds error, got %v", err)
	}
}

func TestDriver_Apply_Idempotent_InsertionGuardedByPrecondition(t *testing.T) {
	// Model the real-world pattern: the fix author checks the
	// precondition (key already present) and emits an empty patch
	// list, so running twice stabilizes.
	src := []byte("a: 1\n")
	precondition := func(hasKey bool) []Patch {
		if hasKey {
			return nil
		}
		return []Patch{{StartByte: 5, EndByte: 5, Replacement: []byte("b: 2\n")}}
	}
	out, err := (&Driver{}).Apply(src, precondition(false))
	if err != nil {
		t.Fatalf("Apply 1: %v", err)
	}
	if string(out) != "a: 1\nb: 2\n" {
		t.Fatalf("first apply: got %q", out)
	}
	// Second run with precondition=true returns no patches; result
	// is unchanged.
	out2, err := (&Driver{}).Apply(out, precondition(true))
	if err != nil {
		t.Fatalf("Apply 2: %v", err)
	}
	if string(out2) != string(out) {
		t.Fatalf("idempotence broken: %q vs %q", out, out2)
	}
}

func TestDriver_ApplyAndVerify_RollbackOnParseRegression(t *testing.T) {
	src := []byte("valid")
	patches := []Patch{
		{StartByte: 5, EndByte: 5, Replacement: []byte("!!!"), RuleID: "bad"},
	}
	d := &Driver{
		ParseFn: func(b []byte) error { return fmt.Errorf("parse failed on %q", b) },
	}
	got, err := d.ApplyAndVerify(src, patches, nil)
	if !errors.Is(err, ErrParseRegression) {
		t.Fatalf("expected ErrParseRegression, got %v", err)
	}
	if string(got) != string(src) {
		t.Fatalf("rollback failed: got %q, want original %q", got, src)
	}
}

func TestDriver_ApplyAndVerify_ParseFnOverrideWins(t *testing.T) {
	src := []byte("abc")
	patches := []Patch{{StartByte: 3, EndByte: 3, Replacement: []byte("d")}}
	called := false
	d := &Driver{
		ParseFn: func(b []byte) error {
			t.Errorf("base ParseFn should not run when override is passed")
			return nil
		},
	}
	override := func(b []byte) error {
		called = true
		return nil
	}
	got, err := d.ApplyAndVerify(src, patches, override)
	if err != nil {
		t.Fatalf("ApplyAndVerify: %v", err)
	}
	if !called {
		t.Fatal("override ParseFn was not invoked")
	}
	if string(got) != "abcd" {
		t.Fatalf("got %q, want abcd", got)
	}
}

func TestDriver_GroupByURI(t *testing.T) {
	patches := []Patch{
		{URI: "a", StartByte: 0, EndByte: 0, Replacement: []byte("x")},
		{URI: "b", StartByte: 0, EndByte: 0, Replacement: []byte("y")},
		{URI: "a", StartByte: 1, EndByte: 1, Replacement: []byte("z")},
	}
	groups := GroupByURI(patches)
	if len(groups) != 2 {
		t.Fatalf("groups = %d, want 2", len(groups))
	}
	if len(groups["a"]) != 2 {
		t.Fatalf("groups[a] has %d, want 2", len(groups["a"]))
	}
	if len(groups["b"]) != 1 {
		t.Fatalf("groups[b] has %d, want 1", len(groups["b"]))
	}
}

func TestPatch_Helpers(t *testing.T) {
	insert := Patch{StartByte: 10, EndByte: 10, Replacement: []byte("xyz")}
	if !insert.IsInsertion() {
		t.Error("IsInsertion should be true for zero-width range")
	}
	if insert.Size() != 3 {
		t.Errorf("Size = %d, want 3", insert.Size())
	}

	replace := Patch{StartByte: 0, EndByte: 5, Replacement: []byte("xx")}
	if replace.IsInsertion() {
		t.Error("IsInsertion should be false for non-zero range")
	}
	if replace.Size() != -3 {
		t.Errorf("Size = %d, want -3", replace.Size())
	}
}
