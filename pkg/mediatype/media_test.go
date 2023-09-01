package mediatype

import (
	"testing"
)

func TestParse(t *testing.T) {
	cases := map[string]struct {
		v        string
		expected MediaType
		wantErr  bool
	}{
		"simple type":      {"text/plain", TextPlain, false},
		"partial wildcard": {"text/*", TextAny, false},
		"full wildcard":    {"*/*", Any, false},
		"no type":          {"/plain", Nil, true},
		"no subtype":       {"text/", Nil, true},
		"no slash":         {"text", Nil, true},
		"parameter":        {"text/plain;q=0.5", TextPlain.WithQuality(0.5), false},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			actual, err := Parse(c.v)
			if err != nil && !c.wantErr {
				t.Errorf("Parse(%q) returned an unexpected error: %s", c.v, err)
			} else if err == nil && c.wantErr {
				t.Errorf("Parse(%q) did not return an error, but an error was expeced", c.v)
			}
			if !actual.Equals(c.expected) {
				t.Errorf("Parse(%q) = %q, expected %q", c.v, actual, c.expected)
			}
		})
	}
}

func TestMediaType_IsWildcardSubtype(t *testing.T) {
	cases := map[string]struct {
		v    string
		want bool
	}{
		"concrete type":    {"text/plain", false},
		"full wildcard":    {"*/*", true},
		"full wildcard 2":  {"*/plain", true},
		"partial wildcard": {"text/*", true},
		"suffix wildcard":  {"text/*+json", true},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			parsed := MustParse(c.v)
			actual := parsed.IsWildcardSubtype()
			if actual != c.want {
				t.Errorf("%q.IsWildcardSubtype() = %t, expected %t", parsed, actual, c.want)
			}
		})
	}
}

func TestMediaType_IsConcrete(t *testing.T) {
	cases := map[string]struct {
		v    string
		want bool
	}{
		"concrete type":    {"text/plain", true},
		"type wildcard":    {"*/*", false},
		"subtype wildcard": {"text/*", false},
		"concrete suffix":  {"application/problem+json", true},
		"wildcard suffix":  {"application/*+json", false},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			parsed := MustParse(c.v)
			actual := parsed.IsConcrete()
			if actual != c.want {
				t.Errorf("%q.IsConcrete() = %t, expected %t", parsed, actual, c.want)
			}
		})
	}
}

func TestMediaType_SubtypeSuffix(t *testing.T) {
	cases := map[string]struct {
		v    string
		want string
	}{
		"concrete type":    {"text/plain", ""},
		"type wildcard":    {"*/*", ""},
		"subtype wildcard": {"text/*", ""},
		"concrete suffix":  {"application/problem+json", "json"},
		"wildcard suffix":  {"application/*+json", "json"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			parsed := MustParse(c.v)
			actual := parsed.SubtypeSuffix()
			if actual != c.want {
				t.Errorf("%q.SubtypeSuffix() = %q, expected %q", parsed, actual, c.want)
			}
		})
	}
}

func TestMediaType_Quality(t *testing.T) {
	cases := map[string]struct {
		v    string
		want float32
	}{
		"default":       {"text/plain", 1},
		"explicit":      {"application/json;q=1", 1},
		"decimal value": {"text/*; q=0.3", 0.3},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			parsed := MustParse(c.v)
			actual := parsed.Quality()
			if actual != c.want {
				t.Errorf("%q.Quality() = %f, expected %f", parsed, actual, c.want)
			}
		})
	}
}

func TestMediaType_Includes(t *testing.T) {
	cases := map[string]struct {
		one   string
		other string
		want  bool
	}{
		"same":                       {"text/plain", "text/plain", true},
		"different":                  {"text/plain", "text/xml", false},
		"same subtype":               {"text/xml", "application/xml", false},
		"full wildcard":              {"*/*", "text/plain", true},
		"full wildcard reverse":      {"text/plain", "*/*", false},
		"subtype wildcard":           {"text/*", "text/plain", true},
		"subtype wildcard reverse":   {"text/plain", "text/*", false},
		"partial wildcard":           {"application/*+json", "application/json", true},
		"partial wildcard reverse":   {"application/json", "application/*+json", false},
		"partial wildcard 2":         {"application/*+json", "application/problem+json", true},
		"partial wildcard 2 reverse": {"application/problem+json", "application/*+json", false},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			one := MustParse(c.one)
			other := MustParse(c.other)
			actual := one.Includes(other)
			if actual != c.want {
				t.Errorf("%q.Includes(%q) = %t, expected %t", one, other, actual, c.want)
			}
		})
	}
}

func TestMediaType_IsCompatibleWith(t *testing.T) {
	cases := map[string]struct {
		one   string
		other string
		want  bool
	}{
		"same":                       {"text/plain", "text/plain", true},
		"different":                  {"text/plain", "text/xml", false},
		"same subtype":               {"text/xml", "application/xml", false},
		"full wildcard":              {"*/*", "text/plain", true},
		"full wildcard reverse":      {"text/plain", "*/*", true},
		"subtype wildcard":           {"text/*", "text/plain", true},
		"subtype wildcard reverse":   {"text/plain", "text/*", true},
		"partial wildcard":           {"application/*+json", "application/json", true},
		"partial wildcard reverse":   {"application/json", "application/*+json", true},
		"partial wildcard 2":         {"application/*+json", "application/problem+json", true},
		"partial wildcard 2 reverse": {"application/problem+json", "application/*+json", true},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			one := MustParse(c.one)
			other := MustParse(c.other)
			actual := one.IsCompatibleWith(other)
			if actual != c.want {
				t.Errorf("%q.IsCompatibleWith(%q) = %t, expected %t", one, other, actual, c.want)
			}
		})
	}
}

func TestMediaType_IsMoreSpecific(t *testing.T) {
	cases := map[string]struct {
		one   string
		other string
		want  bool
	}{
		"same":                    {"text/plain", "text/plain", false},
		"different":               {"text/plain", "text/xml", false},
		"wildcard subtype":        {"text/xml", "application/*", true},
		"full wildcard":           {"text/plain", "*/*", true},
		"full wildcard 2":         {"text/*", "*/*", true},
		"partial wildcard":        {"application/*+xml", "application/*", true},
		"priority":                {"text/plain", "text/xml;q=0.3", false},
		"priority with wildcards": {"text/plain;q=0.4", "*/*;q=0.5", true},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			one := MustParse(c.one)
			other := MustParse(c.other)
			actual := one.IsMoreSpecific(other)
			if actual != c.want {
				t.Errorf("%q.IsMoreSpecific(%q) = %t, expected %t", one, other, actual, c.want)
			}
		})
	}
}

func TestMediaType_HasHigherPriority(t *testing.T) {
	cases := map[string]struct {
		one   string
		other string
		want  bool
	}{
		"same":                   {"text/plain", "text/plain", false},
		"different":              {"text/plain", "text/xml", false},
		"wildcard subtype":       {"text/xml", "application/*", true},
		"full wildcard":          {"text/plain", "*/*", true},
		"full wildcard 2":        {"text/*", "*/*", true},
		"priority":               {"text/plain", "text/xml;q=0.3", true},
		"priority over concrete": {"text/plain;q=0.4", "*/*;q=0.5", false},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			one := MustParse(c.one)
			other := MustParse(c.other)
			actual := one.HasHigherPriority(other)
			if actual != c.want {
				t.Errorf("%q.HasHigherPriority(%q) = %t, expected %t", one, other, actual, c.want)
			}
		})
	}
}
