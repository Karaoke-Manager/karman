package mediatype

import (
	"testing"
)

func TestMediaTypes_BestMatch(t *testing.T) {
	cases := map[string]struct {
		list      string
		available string
		expected  MediaType
	}{
		"single":               {"text/plain", "text/plain", TextPlain},
		"one of two":           {"text/plain", "text/plain, text/xml", TextPlain},
		"candidate priorities": {"application/json, application/problem+json", "application/problem+json, application/json;q=0.9", ApplicationProblemJSON},
		"no match":             {"image/png", "application/json", Nil},
		"wildcard in list":     {"image/*", "image/*, image/png", ImagePNG},
		"match wildcard":       {"image/*, text/*, application/json", "video/mp4, font/ttf, image/*;q=0.001", ImageAny.WithQuality(0.001)},
		"mixed priorities":     {"application/json, application/problem+json;q=0.5", "application/problem+json, application/json;q=0.999", ApplicationJSON.WithQuality(0.999)},
		"empty":                {"", "text/plain", TextPlain},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			l := ParseList(c.list)
			a := ParseList(c.available)
			m := l.BestMatch(a...)
			if !m.Equals(c.expected) {
				t.Errorf("BestMatch(%v) = %q, expected %q", a, m, c.expected)
			}
		})
	}
}
