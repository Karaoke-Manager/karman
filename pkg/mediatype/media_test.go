package mediatype

import (
	"github.com/stretchr/testify/assert"
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
		"no slash":         {"text", NewMediaType("text", ""), false},
		"parameter":        {"text/plain;q=0.5", TextPlain.WithQuality(0.5), false},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			result, err := Parse(c.v)
			if c.wantErr {
				assert.Error(t, err)
			}
			assert.Equal(t, c.expected, result)
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
			assert.Equalf(t, c.want, parsed.IsWildcardSubtype(), "IsWildcardSubtype(%q)", parsed)
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
			assert.Equalf(t, c.want, parsed.IsConcrete(), "IsConcrete()")
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
			assert.Equalf(t, c.want, parsed.SubtypeSuffix(), "SubtypeSuffix()")
		})
	}
}

func TestMediaType_Quality(t *testing.T) {
	cases := map[string]struct {
		v    string
		want float64
	}{
		"default":       {"text/plain", 1},
		"explicit":      {"application/json;q=1", 1},
		"decimal value": {"text/*; q=0.3", 0.3},
		"invalid value": {"application/problem+json; q=abc", 0},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			parsed := MustParse(c.v)
			assert.Equalf(t, c.want, parsed.Quality(), "Quality()")
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
			assert.Equalf(t, c.want, one.Includes(other), "Includes(%#v, %#v)", c.one, c.other)
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
			assert.Equalf(t, c.want, one.IsCompatibleWith(other), "IsCompatibleWith(%#v, %#v)", c.one, c.other)
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
			assert.Equalf(t, c.want, one.IsMoreSpecific(other), "IsMoreSpecific(%#v, %#v)", c.one, c.other)
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
			assert.Equalf(t, c.want, one.HasHigherPriority(other), "HasHigherPriority(%#v, %#v)", c.one, c.other)
		})
	}
}
