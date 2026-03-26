package barrelman

import (
	"regexp"
	"testing"
)

func TestValidatorRequired(t *testing.T) {
	v := V.Required()
	if r := v("hello", "name"); !r.Valid {
		t.Error("non-empty value should pass Required")
	}
	if r := v("", "name"); r.Valid {
		t.Error("empty value should fail Required")
	}
}

func TestValidatorMinLength(t *testing.T) {
	v := V.MinLength(3)
	if r := v("abcd", "f"); !r.Valid {
		t.Error("4 chars should pass MinLength(3)")
	}
	if r := v("ab", "f"); r.Valid {
		t.Error("2 chars should fail MinLength(3)")
	}
	if r := v("", "f"); !r.Valid {
		t.Error("empty should pass MinLength (skip)")
	}
}

func TestValidatorMaxLength(t *testing.T) {
	v := V.MaxLength(5)
	if r := v("abc", "f"); !r.Valid {
		t.Error("3 chars should pass MaxLength(5)")
	}
	if r := v("abcdef", "f"); r.Valid {
		t.Error("6 chars should fail MaxLength(5)")
	}
	if r := v("", "f"); !r.Valid {
		t.Error("empty should pass MaxLength (skip)")
	}
}

func TestValidatorPattern(t *testing.T) {
	v := V.Pattern(regexp.MustCompile(`^[a-z]+$`))
	if r := v("hello", "f"); !r.Valid {
		t.Error("lowercase should pass")
	}
	if r := v("Hello", "f"); r.Valid {
		t.Error("mixed case should fail")
	}
	if r := v("", "f"); !r.Valid {
		t.Error("empty should pass (skip)")
	}
}

func TestValidatorOneOf(t *testing.T) {
	v := V.OneOf([]string{"a", "b", "c"})
	if r := v("b", "f"); !r.Valid {
		t.Error("allowed value should pass")
	}
	if r := v("d", "f"); r.Valid {
		t.Error("disallowed value should fail")
	}
}

func TestValidatorTitleCase(t *testing.T) {
	v := V.TitleCase()
	if r := v("Hello", "f"); !r.Valid {
		t.Error("capitalized should pass")
	}
	if r := v("hello", "f"); r.Valid {
		t.Error("lowercase should fail")
	}
}

func TestValidatorCamelCase(t *testing.T) {
	v := V.CamelCase()
	if r := v("camelCase", "f"); !r.Valid {
		t.Error("camelCase should pass")
	}
	if r := v("PascalCase", "f"); r.Valid {
		t.Error("PascalCase should fail (starts uppercase)")
	}
	if r := v("snake_case", "f"); r.Valid {
		t.Error("snake_case should fail")
	}
}

func TestValidatorKebabCase(t *testing.T) {
	v := V.KebabCase()
	if r := v("kebab-case", "f"); !r.Valid {
		t.Error("kebab-case should pass")
	}
	if r := v("camelCase", "f"); r.Valid {
		t.Error("camelCase should fail")
	}
}

func TestValidatorCustom(t *testing.T) {
	v := V.Custom(func(s string) bool { return s == "yes" }, "must be yes")
	if r := v("yes", "f"); !r.Valid {
		t.Error("yes should pass")
	}
	if r := v("no", "f"); r.Valid {
		t.Error("no should fail")
	}
}

func TestValidatorAll(t *testing.T) {
	v := V.All(V.Required(), V.MinLength(3))
	if r := v("hello", "f"); !r.Valid {
		t.Error("should pass both validators")
	}
	if r := v("", "f"); r.Valid {
		t.Error("should fail Required")
	}
	if r := v("ab", "f"); r.Valid {
		t.Error("should fail MinLength")
	}
}

func TestValidatorAny(t *testing.T) {
	v := V.Any(
		V.Custom(func(s string) bool { return s == "a" }, "not a"),
		V.Custom(func(s string) bool { return s == "b" }, "not b"),
	)
	if r := v("a", "f"); !r.Valid {
		t.Error("a should pass (first validator)")
	}
	if r := v("b", "f"); !r.Valid {
		t.Error("b should pass (second validator)")
	}
	if r := v("c", "f"); r.Valid {
		t.Error("c should fail both")
	}
}

func TestValidatorOptional(t *testing.T) {
	v := V.Optional(V.MinLength(5))
	if r := v("", "f"); !r.Valid {
		t.Error("empty should pass Optional")
	}
	if r := v("hi", "f"); r.Valid {
		t.Error("short non-empty should fail inner")
	}
	if r := v("hello", "f"); !r.Valid {
		t.Error("long enough should pass inner")
	}
}
