package barrelman

import (
	"testing"

	navigator "github.com/sailpoint-oss/navigator"
)

func TestReporterAt(t *testing.T) {
	r := NewReporter("test-rule", SeverityWarning)
	loc := navigator.Loc{Range: Range{
		Start: Position{Line: 5, Character: 10},
		End:   Position{Line: 5, Character: 20},
	}}
	r.At(loc, "found issue at %s", "field")

	diags := r.Diagnostics()
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
	d := diags[0]
	if d.Code != "test-rule" {
		t.Errorf("Code = %q, want %q", d.Code, "test-rule")
	}
	if d.Severity != SeverityWarning {
		t.Errorf("Severity = %d, want %d", d.Severity, SeverityWarning)
	}
	if d.Message != "found issue at field" {
		t.Errorf("Message = %q, want %q", d.Message, "found issue at field")
	}
	if d.Source != Source {
		t.Errorf("Source = %q, want %q", d.Source, Source)
	}
}

func TestReporterAtRange(t *testing.T) {
	r := NewReporter("test-rule", SeverityError)
	rng := Range{
		Start: Position{Line: 1, Character: 0},
		End:   Position{Line: 1, Character: 5},
	}
	r.AtRange(rng, "range issue")

	diags := r.Diagnostics()
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
	if diags[0].Severity != SeverityError {
		t.Errorf("Severity = %d, want %d", diags[0].Severity, SeverityError)
	}
}

func TestReporterErrorWarn(t *testing.T) {
	r := NewReporter("test-rule", SeverityInfo)
	loc := navigator.Loc{Range: FileStartRange}

	r.Error(loc, "error msg")
	r.Warn(loc, "warn msg")

	diags := r.Diagnostics()
	if len(diags) != 2 {
		t.Fatalf("expected 2 diagnostics, got %d", len(diags))
	}
	if diags[0].Severity != SeverityError {
		t.Errorf("Error() severity = %d, want %d", diags[0].Severity, SeverityError)
	}
	if diags[1].Severity != SeverityWarning {
		t.Errorf("Warn() severity = %d, want %d", diags[1].Severity, SeverityWarning)
	}
}

func TestReporterWithTags(t *testing.T) {
	r := NewReporter("test-rule", SeverityHint)
	loc := navigator.Loc{Range: FileStartRange}

	r.WithTags(DiagnosticTagDeprecated).At(loc, "deprecated")
	r.At(loc, "no tags")

	diags := r.Diagnostics()
	if len(diags) != 2 {
		t.Fatalf("expected 2 diagnostics, got %d", len(diags))
	}
	if len(diags[0].Tags) != 1 || diags[0].Tags[0] != DiagnosticTagDeprecated {
		t.Errorf("first diagnostic should have deprecated tag")
	}
	if len(diags[1].Tags) != 0 {
		t.Errorf("second diagnostic should have no tags (tags should be consumed)")
	}
}

func TestReporterWithData(t *testing.T) {
	r := NewReporter("test-rule", SeverityInfo)
	loc := navigator.Loc{Range: FileStartRange}

	r.WithData(map[string]string{"key": "val"}).At(loc, "with data")
	r.At(loc, "no data")

	diags := r.Diagnostics()
	if len(diags) != 2 {
		t.Fatalf("expected 2 diagnostics, got %d", len(diags))
	}
	if diags[0].Data == nil {
		t.Error("first diagnostic should have data")
	}
	if diags[1].Data != nil {
		t.Error("second diagnostic should have nil data (consumed)")
	}
}

func TestReporterMultiLineRangeClamped(t *testing.T) {
	r := NewReporter("test-rule", SeverityWarning)
	rng := Range{
		Start: Position{Line: 1, Character: 5},
		End:   Position{Line: 3, Character: 10},
	}
	r.AtRange(rng, "multiline")

	diags := r.Diagnostics()
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
	// Multi-line ranges are clamped to single line.
	if diags[0].Range.End.Line != diags[0].Range.Start.Line {
		t.Errorf("End.Line = %d, want %d (clamped to start line)", diags[0].Range.End.Line, diags[0].Range.Start.Line)
	}
}
