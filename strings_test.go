package barrelman

import "testing"

func TestIsCapitalized(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"Pet", true},
		{"pet", false},
		{"", false},
		{"A", true},
		{"a", false},
		{"123", false},
	}
	for _, tc := range tests {
		if got := IsCapitalized(tc.input); got != tc.want {
			t.Errorf("IsCapitalized(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestIsKebabCase(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"user-items", true},
		{"users", true},
		{"userItems", false},
		{"user_items", false},
		{"User", false},
		{"", true},
		{"a-b-c", true},
	}
	for _, tc := range tests {
		if got := IsKebabCase(tc.input); got != tc.want {
			t.Errorf("IsKebabCase(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestContainsHTTPVerb(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"/pets", false},
		{"/get-pets", false},
		{"/pets/get", true},
		{"/delete/items", true},
		{"/pets/post", true},
		{"/items/{id}", false},
		{"/getaway", false},
	}
	for _, tc := range tests {
		if got := ContainsHTTPVerb(tc.input); got != tc.want {
			t.Errorf("ContainsHTTPVerb(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestHasTrailingSlash(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"/pets", false},
		{"/pets/", true},
		{"/", false},
		{"/a/b/", true},
	}
	for _, tc := range tests {
		if got := HasTrailingSlash(tc.input); got != tc.want {
			t.Errorf("HasTrailingSlash(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestIsHTTPS(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"https://example.com", true},
		{"HTTPS://example.com", true},
		{"http://example.com", false},
		{"ftp://example.com", false},
		{"", false},
	}
	for _, tc := range tests {
		if got := IsHTTPS(tc.input); got != tc.want {
			t.Errorf("IsHTTPS(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestContainsCredentials(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"https://example.com", false},
		{"https://user@example.com", true},
		{"https://user:pass@example.com/path", true},
		{"http://example.com/path", false},
		{"noscheme", false},
	}
	for _, tc := range tests {
		if got := ContainsCredentials(tc.input); got != tc.want {
			t.Errorf("ContainsCredentials(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestExtractPathParams(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"/pets/{petId}", []string{"petId"}},
		{"/pets/{petId}/toys/{toyId}", []string{"petId", "toyId"}},
		{"/pets", nil},
		{"/{a}/{b}/{c}", []string{"a", "b", "c"}},
	}
	for _, tc := range tests {
		got := ExtractPathParams(tc.input)
		if len(got) != len(tc.want) {
			t.Errorf("ExtractPathParams(%q) = %v, want %v", tc.input, got, tc.want)
			continue
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("ExtractPathParams(%q)[%d] = %q, want %q", tc.input, i, got[i], tc.want[i])
			}
		}
	}
}

func TestNonParamSegments(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"/pets/{id}", []string{"pets"}},
		{"/user-items/{id}/toys", []string{"user-items", "toys"}},
		{"/pets", []string{"pets"}},
		{"/{id}", nil},
	}
	for _, tc := range tests {
		got := NonParamSegments(tc.input)
		if len(got) != len(tc.want) {
			t.Errorf("NonParamSegments(%q) = %v, want %v", tc.input, got, tc.want)
			continue
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("NonParamSegments(%q)[%d] = %q, want %q", tc.input, i, got[i], tc.want[i])
			}
		}
	}
}
