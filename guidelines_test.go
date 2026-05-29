package barrelman

import "testing"

func TestNormalizeGuidelineCode(t *testing.T) {
	tests := map[string]string{
		"104":      "rule-104",
		"rule-104": "rule-104",
		"#304":     "rule-304",
		"oops":     "",
	}
	for input, want := range tests {
		if got := NormalizeGuidelineCode(input); got != want {
			t.Fatalf("NormalizeGuidelineCode(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestGuidelineDocURL(t *testing.T) {
	SetGuidelinesBaseURL("https://example.com/guidelines")
	t.Cleanup(func() {
		SetGuidelinesBaseURL("")
	})

	got := GuidelineDocURL("rule-104")
	want := "https://example.com/guidelines/docs/rules/api-contract-and-documentation#104"
	if got != want {
		t.Fatalf("GuidelineDocURL(rule-104) = %q, want %q", got, want)
	}
}
