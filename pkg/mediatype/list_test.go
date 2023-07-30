package mediatype

import (
	"github.com/stretchr/testify/assert"
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
		"match wildcard":       {"image/*, text/*, application/json", "video/mp4, font/ttf, image/*;q=0.001", ImageAny},
		"mixed priorities":     {"application/json, application/problem+json;q=0.5", "application/problem+json, application/json;q=0.999", ApplicationJSON},
		"empty":                {"", "text/plain", TextPlain},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			l := ParseList(c.list)
			a := ParseList(c.available)
			m := l.BestMatch(a...)
			assert.Equal(t, c.expected, m, "BestMatch()")
		})
	}
}
