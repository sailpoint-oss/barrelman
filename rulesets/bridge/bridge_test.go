package bridge

import (
	"testing"
)

func TestCanonical_RoundTripsEveryEntry(t *testing.T) {
	for _, e := range All() {
		if got := Canonical(e.Canonical); got != e.Canonical {
			t.Errorf("Canonical(%q) = %q, want %q", e.Canonical, got, e.Canonical)
		}
		if e.Vacuum != "" {
			if got := Canonical(e.Vacuum); got != e.Canonical {
				t.Errorf("Canonical(%q vacuum) = %q, want %q", e.Vacuum, got, e.Canonical)
			}
		}
		if e.Spectral != "" {
			if got := Canonical(e.Spectral); got != e.Canonical {
				t.Errorf("Canonical(%q spectral) = %q, want %q", e.Spectral, got, e.Canonical)
			}
		}
	}
}

func TestCanonical_NumericAndUnknownFormatsPassThrough(t *testing.T) {
	cases := map[string]string{
		"304":  "304",
		"#304": "#304",
	}
	for in, want := range cases {
		if got := Canonical(in); got != want {
			t.Errorf("Canonical(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCanonical_UnknownPassThrough(t *testing.T) {
	cases := []string{"custom-rule", "info-contact", "totally-made-up"}
	for _, id := range cases {
		if got := Canonical(id); got != id {
			t.Errorf("Canonical(%q) = %q, want pass-through", id, got)
		}
	}
}

func TestCanonical_Empty(t *testing.T) {
	if got := Canonical(""); got != "" {
		t.Errorf("Canonical(\"\") = %q, want empty", got)
	}
}

func TestVacuumAndSpectralLookups(t *testing.T) {
	for _, e := range All() {
		if e.Vacuum != "" {
			if got := Vacuum(e.Canonical); got != e.Vacuum {
				t.Errorf("Vacuum(%q) = %q, want %q", e.Canonical, got, e.Vacuum)
			}
		}
		if e.Spectral != "" {
			if got := Spectral(e.Canonical); got != e.Spectral {
				t.Errorf("Spectral(%q) = %q, want %q", e.Canonical, got, e.Spectral)
			}
		}
	}
}

func TestFromCanonical_UnknownReturnsFalse(t *testing.T) {
	if _, ok := FromCanonical("not-a-real-rule"); ok {
		t.Error("FromCanonical for unknown id returned ok=true")
	}
}

func TestForGuideline_UnknownReturnsEmpty(t *testing.T) {
	if entries := ForGuideline(403); len(entries) != 0 {
		t.Fatalf("ForGuideline(403) = %d entries, want 0", len(entries))
	}
}

func TestEntryPrimaryGuideline(t *testing.T) {
	for _, e := range All() {
		if len(e.GuidelineIDs) == 0 {
			if e.PrimaryGuideline() != 0 {
				t.Errorf("entry %q with no guideline ids returned primary %d", e.Canonical, e.PrimaryGuideline())
			}
			continue
		}
		if e.PrimaryGuideline() != e.GuidelineIDs[0] {
			t.Errorf("entry %q primary = %d, want %d", e.Canonical, e.PrimaryGuideline(), e.GuidelineIDs[0])
		}
	}
}
